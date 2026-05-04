package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ory/fosite"
)

const sessionShmPath = "/dev/shm/sampatti-sessions.enc"

// SessionCache encrypts OAuth session state and persists to /dev/shm.
type SessionCache struct {
	gcm cipher.AEAD
	mu  sync.Mutex
}

// serializedState holds all OAuth state that needs to survive restarts.
type serializedState struct {
	Clients       map[string]*serializedClient `json:"clients"`
	AccessTokens  map[string]*serializedToken  `json:"access_tokens"`
	RefreshTokens map[string]*serializedToken  `json:"refresh_tokens"`
}

type serializedClient struct {
	ID            string   `json:"id"`
	RedirectURIs  []string `json:"redirect_uris"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types"`
	Scopes        []string `json:"scopes"`
	Public        bool     `json:"public"`
}

type serializedToken struct {
	RequestID       string    `json:"request_id"`
	Subject         string    `json:"subject"`
	ClientID        string    `json:"client_id"`
	RequestedScopes []string  `json:"requested_scopes"`
	GrantedScopes   []string  `json:"granted_scopes"`
	RequestedAt     time.Time `json:"requested_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// NewSessionCache creates a session cache. Returns nil if no secret.
func NewSessionCache(secret []byte) *SessionCache {
	if len(secret) == 0 {
		return nil
	}

	// Use a different key than the cred cache by hashing with a domain separator
	h := sha256.New()
	h.Write(secret)
	h.Write([]byte("session-cache"))
	key := h.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("Failed to create AES cipher for session cache: %v", err)
		return nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Failed to create GCM for session cache: %v", err)
		return nil
	}

	return &SessionCache{gcm: gcm}
}

// Save encrypts and writes OAuth state to /dev/shm.
func (c *SessionCache) Save(
	clients map[string]*fosite.DefaultClient,
	accessTokens map[string]StoreData,
	refreshTokens map[string]StoreData,
) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	state := serializedState{
		Clients:       make(map[string]*serializedClient, len(clients)),
		AccessTokens:  make(map[string]*serializedToken, len(accessTokens)),
		RefreshTokens: make(map[string]*serializedToken, len(refreshTokens)),
	}

	for id, client := range clients {
		state.Clients[id] = &serializedClient{
			ID:            client.ID,
			RedirectURIs:  client.RedirectURIs,
			GrantTypes:    client.GrantTypes,
			ResponseTypes: client.ResponseTypes,
			Scopes:        client.Scopes,
			Public:        client.Public,
		}
	}

	for sig, data := range accessTokens {
		state.AccessTokens[sig] = tokenToSerialized(data)
	}

	for sig, data := range refreshTokens {
		state.RefreshTokens[sig] = tokenToSerialized(data)
	}

	plaintext, err := json.Marshal(state)
	if err != nil {
		log.Printf("Failed to marshal session state: %v", err)
		return
	}

	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("Failed to generate nonce for session cache: %v", err)
		return
	}

	encrypted := c.gcm.Seal(nonce, nonce, plaintext, nil)

	if err := os.WriteFile(sessionShmPath, encrypted, 0600); err != nil {
		log.Printf("Failed to write session cache to %s: %v", sessionShmPath, err)
	}
}

// Load decrypts and restores OAuth state from /dev/shm.
func (c *SessionCache) Load() (
	clients map[string]*fosite.DefaultClient,
	accessTokens map[string]StoreData,
	refreshTokens map[string]StoreData,
) {
	if c == nil {
		return nil, nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	encrypted, err := os.ReadFile(sessionShmPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read session cache: %v", err)
		}
		return nil, nil, nil
	}

	nonceSize := c.gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, nil, nil
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("Failed to decrypt session cache (wrong secret?): %v", err)
		return nil, nil, nil
	}

	var state serializedState
	if err := json.Unmarshal(plaintext, &state); err != nil {
		log.Printf("Failed to unmarshal session cache: %v", err)
		return nil, nil, nil
	}

	// Reconstruct fosite objects
	clients = make(map[string]*fosite.DefaultClient, len(state.Clients))
	for id, sc := range state.Clients {
		clients[id] = &fosite.DefaultClient{
			ID:            sc.ID,
			RedirectURIs:  sc.RedirectURIs,
			GrantTypes:    sc.GrantTypes,
			ResponseTypes: sc.ResponseTypes,
			Scopes:        sc.Scopes,
			Public:        sc.Public,
		}
	}

	accessTokens = make(map[string]StoreData, len(state.AccessTokens))
	for sig, st := range state.AccessTokens {
		accessTokens[sig] = serializedToStoreData(st, clients)
	}

	refreshTokens = make(map[string]StoreData, len(state.RefreshTokens))
	for sig, st := range state.RefreshTokens {
		refreshTokens[sig] = serializedToStoreData(st, clients)
	}

	log.Printf("Restored %d client(s), %d access token(s), %d refresh token(s) from %s",
		len(clients), len(accessTokens), len(refreshTokens), sessionShmPath)

	return clients, accessTokens, refreshTokens
}

func tokenToSerialized(data StoreData) *serializedToken {
	st := &serializedToken{
		RequestID:       data.Requester.GetID(),
		RequestedScopes: []string(data.Requester.GetRequestedScopes()),
		GrantedScopes:   []string(data.Requester.GetGrantedScopes()),
		RequestedAt:     data.Requester.GetRequestedAt(),
	}

	if session, ok := data.Requester.GetSession().(*Session); ok {
		st.Subject = session.Subject
		st.ExpiresAt = session.GetExpiresAt(fosite.AccessToken)
	}

	if client := data.Requester.GetClient(); client != nil {
		st.ClientID = client.GetID()
	}

	return st
}

func serializedToStoreData(st *serializedToken, clients map[string]*fosite.DefaultClient) StoreData {
	session := NewSession(st.Subject)
	session.SetExpiresAt(fosite.AccessToken, st.ExpiresAt)

	var client fosite.Client
	if c, ok := clients[st.ClientID]; ok {
		client = c
	}

	req := &fosite.Request{
		ID:                st.RequestID,
		RequestedAt:       st.RequestedAt,
		Client:            client,
		RequestedScope:    fosite.Arguments(st.RequestedScopes),
		GrantedScope:      fosite.Arguments(st.GrantedScopes),
		Session:           session,
		Form:              url.Values{},
	}

	return StoreData{
		Requester: req,
		CreatedAt: st.RequestedAt,
	}
}

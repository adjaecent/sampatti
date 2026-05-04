package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ory/fosite"
)

// MemoryStore implements fosite's storage interfaces with in-memory maps.
// OAuth artifacts are lost on restart. Credentials can survive restarts
// via an encrypted cache on /dev/shm (tmpfs).
type MemoryStore struct {
	mu sync.RWMutex

	// OAuth artifacts
	clients       map[string]*fosite.DefaultClient
	authCodes     map[string]StoreData
	accessTokens  map[string]StoreData
	refreshTokens map[string]StoreData
	pkceRequests  map[string]fosite.Requester

	// User credentials keyed by subject ID
	credentials map[string]*Credentials

	// Encrypted tmpfs caches
	credCache    *CredCache
	sessionCache *SessionCache
}

// StoreData wraps a fosite.Requester with a creation timestamp for expiry tracking.
type StoreData struct {
	Requester fosite.Requester
	CreatedAt time.Time
}

// NewMemoryStore creates a new in-memory store.
// If secret is provided, credentials are cached encrypted on /dev/shm
// and restored on startup.
func NewMemoryStore(secret []byte) *MemoryStore {
	credCache := NewCredCache(secret)
	sessionCache := NewSessionCache(secret)

	creds := make(map[string]*Credentials)
	if restored := credCache.Load(); restored != nil {
		creds = restored
	}

	clients := make(map[string]*fosite.DefaultClient)
	accessTokens := make(map[string]StoreData)
	refreshTokens := make(map[string]StoreData)

	if rc, rat, rrt := sessionCache.Load(); rc != nil {
		clients = rc
		accessTokens = rat
		refreshTokens = rrt
	}

	return &MemoryStore{
		clients:       clients,
		authCodes:     make(map[string]StoreData),
		accessTokens:  accessTokens,
		refreshTokens: refreshTokens,
		pkceRequests:  make(map[string]fosite.Requester),
		credentials:   creds,
		credCache:     credCache,
		sessionCache:  sessionCache,
	}
}

// --- Credential management ---

// StoreCredentials stores credentials keyed by subject ID.
// Also persists encrypted to /dev/shm if a cache is configured.
func (s *MemoryStore) StoreCredentials(subject string, creds *Credentials) {
	s.mu.Lock()
	s.credentials[subject] = creds
	snapshot := make(map[string]*Credentials, len(s.credentials))
	for k, v := range s.credentials {
		snapshot[k] = v
	}
	s.mu.Unlock()

	s.credCache.Save(snapshot)
}

// GetCredentials retrieves credentials by subject ID.
func (s *MemoryStore) GetCredentials(subject string) *Credentials {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.credentials[subject]
}

// saveSessionState snapshots and persists OAuth state to /dev/shm.
// Must be called WITHOUT the lock held (it acquires a read lock).
func (s *MemoryStore) saveSessionState() {
	s.mu.RLock()
	clients := make(map[string]*fosite.DefaultClient, len(s.clients))
	for k, v := range s.clients {
		clients[k] = v
	}
	accessTokens := make(map[string]StoreData, len(s.accessTokens))
	for k, v := range s.accessTokens {
		accessTokens[k] = v
	}
	refreshTokens := make(map[string]StoreData, len(s.refreshTokens))
	for k, v := range s.refreshTokens {
		refreshTokens[k] = v
	}
	s.mu.RUnlock()

	s.sessionCache.Save(clients, accessTokens, refreshTokens)
}

// --- Client management (fosite.ClientManager) ---

func (s *MemoryStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return client, nil
}

func (s *MemoryStore) ClientAssertionJWTValid(_ context.Context, _ string) error {
	return nil // Not using JWT client assertions
}

func (s *MemoryStore) SetClientAssertionJWT(_ context.Context, _ string, _ time.Time) error {
	return nil
}

// RegisterClient creates a new OAuth client (for dynamic registration).
func (s *MemoryStore) RegisterClient(redirectURIs []string) (*fosite.DefaultClient, error) {
	clientID, err := generateID("client_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate client ID: %w", err)
	}

	client := &fosite.DefaultClient{
		ID:            clientID,
		Secret:        nil, // Public client, no secret
		RedirectURIs:  redirectURIs,
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
		Scopes:        []string{"portfolio", "offline_access"},
		Public:        true,
	}

	s.mu.Lock()
	s.clients[clientID] = client
	s.mu.Unlock()

	s.saveSessionState()
	return client, nil
}

// --- Authorization code storage ---

func (s *MemoryStore) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authCodes[code] = StoreData{Requester: req, CreatedAt: time.Now()}
	return nil
}

func (s *MemoryStore) GetAuthorizeCodeSession(_ context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.authCodes[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return data.Requester, nil
}

func (s *MemoryStore) InvalidateAuthorizeCodeSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.authCodes, code)
	return nil
}

// --- Access token storage ---

func (s *MemoryStore) CreateAccessTokenSession(_ context.Context, signature string, req fosite.Requester) error {
	s.mu.Lock()
	s.accessTokens[signature] = StoreData{Requester: req, CreatedAt: time.Now()}
	s.mu.Unlock()

	s.saveSessionState()
	return nil
}

func (s *MemoryStore) GetAccessTokenSession(_ context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.accessTokens[signature]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return data.Requester, nil
}

func (s *MemoryStore) DeleteAccessTokenSession(_ context.Context, signature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.accessTokens, signature)
	return nil
}

// --- Refresh token storage ---

func (s *MemoryStore) CreateRefreshTokenSession(_ context.Context, signature string, accessSignature string, req fosite.Requester) error {
	s.mu.Lock()
	s.refreshTokens[signature] = StoreData{Requester: req, CreatedAt: time.Now()}
	s.mu.Unlock()

	s.saveSessionState()
	return nil
}

func (s *MemoryStore) GetRefreshTokenSession(_ context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.refreshTokens[signature]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return data.Requester, nil
}

func (s *MemoryStore) DeleteRefreshTokenSession(_ context.Context, signature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, signature)
	return nil
}

func (s *MemoryStore) RotateRefreshToken(_ context.Context, requestID string, refreshTokenSignature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, refreshTokenSignature)
	return nil
}

func (s *MemoryStore) RevokeRefreshToken(_ context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sig, data := range s.refreshTokens {
		if data.Requester.GetID() == requestID {
			delete(s.refreshTokens, sig)
		}
	}
	return nil
}

func (s *MemoryStore) RevokeAccessToken(_ context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sig, data := range s.accessTokens {
		if data.Requester.GetID() == requestID {
			delete(s.accessTokens, sig)
		}
	}
	return nil
}

// --- PKCE storage ---

func (s *MemoryStore) CreatePKCERequestSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pkceRequests[code] = req
	return nil
}

func (s *MemoryStore) GetPKCERequestSession(_ context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.pkceRequests[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return req, nil
}

func (s *MemoryStore) DeletePKCERequestSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pkceRequests, code)
	return nil
}

// --- Helpers ---

func generateID(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

const shmPath = "/dev/shm/sampatti-creds.enc"

// CredCache encrypts credentials and persists them to /dev/shm (tmpfs).
// Survives process restarts but not machine reboots.
type CredCache struct {
	gcm    cipher.AEAD
	mu     sync.Mutex
}

// serializedCreds is the JSON-friendly form of the credential map.
type serializedCreds struct {
	Subjects map[string]*Credentials `json:"subjects"`
}

// NewCredCache creates a cache using the given secret for AES-256-GCM encryption.
// If secret is nil or empty, caching is disabled (returns nil).
func NewCredCache(secret []byte) *CredCache {
	if len(secret) == 0 {
		log.Println("No OAUTH_SECRET set, credential caching to /dev/shm disabled")
		return nil
	}

	// Derive a 32-byte key from the secret
	key := sha256.Sum256(secret)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		log.Printf("Failed to create AES cipher for cred cache: %v", err)
		return nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Failed to create GCM for cred cache: %v", err)
		return nil
	}

	return &CredCache{gcm: gcm}
}

// Save encrypts and writes the credential map to /dev/shm.
func (c *CredCache) Save(creds map[string]*Credentials) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(serializedCreds{Subjects: creds})
	if err != nil {
		log.Printf("Failed to marshal credentials for cache: %v", err)
		return
	}

	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("Failed to generate nonce for cred cache: %v", err)
		return
	}

	encrypted := c.gcm.Seal(nonce, nonce, data, nil)

	if err := os.WriteFile(shmPath, encrypted, 0600); err != nil {
		log.Printf("Failed to write cred cache to %s: %v", shmPath, err)
		return
	}
}

// Load decrypts and returns the credential map from /dev/shm.
// Returns nil if the file doesn't exist or decryption fails.
func (c *CredCache) Load() map[string]*Credentials {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	encrypted, err := os.ReadFile(shmPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read cred cache from %s: %v", shmPath, err)
		}
		return nil
	}

	nonceSize := c.gcm.NonceSize()
	if len(encrypted) < nonceSize {
		log.Printf("Cred cache file too small, ignoring")
		return nil
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("Failed to decrypt cred cache (wrong secret?): %v", err)
		return nil
	}

	var sc serializedCreds
	if err := json.Unmarshal(plaintext, &sc); err != nil {
		log.Printf("Failed to unmarshal cred cache: %v", err)
		return nil
	}

	log.Printf("Restored %d credential(s) from %s", len(sc.Subjects), shmPath)
	return sc.Subjects
}

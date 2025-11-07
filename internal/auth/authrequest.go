package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type KuveraCredentials struct {
	Username string
	Password string
}

type StockalCredentials struct {
	Username string
	Password string
}

type Credentials struct {
	Kuvera  *KuveraCredentials
	Stockal *StockalCredentials
}

type AuthRequest struct {
	ID          string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Used        bool
	Credentials *Credentials
	Token       string // Short-lived token generated after auth
	TokenExpiry time.Time
}

type AuthRequestManager struct {
	requests map[string]*AuthRequest
	tokens   map[string]*AuthRequest // token -> request mapping
	mutex    sync.RWMutex
}

func NewAuthRequestManager() *AuthRequestManager {
	arm := &AuthRequestManager{
		requests: make(map[string]*AuthRequest),
		tokens:   make(map[string]*AuthRequest),
	}

	// Start cleanup goroutine
	go arm.cleanup()

	return arm
}

func (arm *AuthRequestManager) CreateAuthRequest() (*AuthRequest, error) {
	requestID, err := generateSecureID("auth_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}

	authRequest := &AuthRequest{
		ID:        requestID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(3 * time.Minute), // 3 minute window
		Used:      false,
	}

	arm.mutex.Lock()
	arm.requests[requestID] = authRequest
	arm.mutex.Unlock()

	return authRequest, nil
}

func (arm *AuthRequestManager) CompleteAuthRequest(requestID string, creds *Credentials) (string, error) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()

	authRequest, exists := arm.requests[requestID]
	if !exists {
		return "", fmt.Errorf("auth request not found")
	}

	if time.Now().After(authRequest.ExpiresAt) {
		delete(arm.requests, requestID)
		return "", fmt.Errorf("auth request expired")
	}

	if authRequest.Used {
		return "", fmt.Errorf("auth request already used")
	}

	// Generate short-lived token
	token, err := generateSecureID("tok_")
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Mark as used and store credentials
	authRequest.Used = true
	authRequest.Credentials = creds
	authRequest.Token = token
	authRequest.TokenExpiry = time.Now().Add(30 * 24 * time.Hour) // 1 month token

	// Add to token mapping
	arm.tokens[token] = authRequest

	return token, nil
}

func (arm *AuthRequestManager) GetAuthToken(requestID string) (string, time.Time, error) {
	arm.mutex.RLock()
	defer arm.mutex.RUnlock()

	authRequest, exists := arm.requests[requestID]
	if !exists {
		return "", time.Time{}, fmt.Errorf("auth request not found")
	}

	if !authRequest.Used || authRequest.Token == "" {
		return "", time.Time{}, fmt.Errorf("auth request not completed yet")
	}

	if time.Now().After(authRequest.TokenExpiry) {
		return "", time.Time{}, fmt.Errorf("token expired")
	}

	return authRequest.Token, authRequest.TokenExpiry, nil
}

func (arm *AuthRequestManager) ValidateToken(token string) (*Credentials, error) {
	arm.mutex.RLock()
	defer arm.mutex.RUnlock()

	authRequest, exists := arm.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	if time.Now().After(authRequest.TokenExpiry) {
		return nil, fmt.Errorf("token expired")
	}

	return authRequest.Credentials, nil
}

func (arm *AuthRequestManager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		arm.mutex.Lock()
		now := time.Now()

		// Clean up expired requests
		for id, request := range arm.requests {
			if now.After(request.ExpiresAt) || (request.Used && now.After(request.TokenExpiry)) {
				delete(arm.requests, id)
				if request.Token != "" {
					delete(arm.tokens, request.Token)
				}
			}
		}

		// Clean up expired tokens
		for token, request := range arm.tokens {
			if now.After(request.TokenExpiry) {
				delete(arm.tokens, token)
			}
		}

		arm.mutex.Unlock()
	}
}

func generateSecureID(prefix string) (string, error) {
	bytes := make([]byte, 32) // 256-bit entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}

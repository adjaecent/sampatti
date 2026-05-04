package oauth

import (
	"crypto/rand"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
)

// NewOAuthProvider creates and configures a fosite OAuth2 provider with PKCE support.
func NewOAuthProvider(store *MemoryStore, secret []byte) fosite.OAuth2Provider {
	if secret == nil {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			panic("failed to generate OAuth secret: " + err.Error())
		}
	}

	config := &fosite.Config{
		AccessTokenLifespan:           24 * time.Hour,
		RefreshTokenLifespan:          30 * 24 * time.Hour,
		AuthorizeCodeLifespan:         10 * time.Minute,
		EnforcePKCE:                   true,
		EnablePKCEPlainChallengeMethod: false, // Require S256
		GlobalSecret:                  secret,
	}

	return compose.Compose(
		config,
		store,
		&compose.CommonStrategy{
			CoreStrategy: compose.NewOAuth2HMACStrategy(config),
		},
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2PKCEFactory,
		compose.OAuth2TokenIntrospectionFactory,
	)
}

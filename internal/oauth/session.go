package oauth

import (
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

// Session is a custom fosite session that links an OAuth token to a subject.
// The subject is used to look up credentials in the in-memory store.
type Session struct {
	*openid.DefaultSession
}

// NewSession creates a new session with the given subject identifier.
func NewSession(subject string) *Session {
	return &Session{
		DefaultSession: &openid.DefaultSession{
			Subject: subject,
			Claims: &jwt.IDTokenClaims{
				Subject:   subject,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			Headers: &jwt.Headers{},
		},
	}
}

// Ensure Session implements fosite.Session.
var _ fosite.Session = (*Session)(nil)

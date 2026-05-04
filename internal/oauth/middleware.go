package oauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ory/fosite"
)

type contextKey string

const credentialsContextKey contextKey = "sampatti_credentials"

// CredentialsFromContext retrieves the authenticated user's credentials from context.
// Returns nil if not authenticated.
func CredentialsFromContext(ctx context.Context) *Credentials {
	creds, _ := ctx.Value(credentialsContextKey).(*Credentials)
	return creds
}

// AuthMiddleware validates the Bearer token on incoming requests and injects
// credentials into the request context. For GET/DELETE (MCP session management),
// it allows unauthenticated access so clients can discover the server.
// For POST (tool calls), it requires a valid Bearer token.
func AuthMiddleware(provider fosite.OAuth2Provider, store *MemoryStore, baseURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always set CORS headers
		setCORSHeaders(w, r)

		// Handle CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Try to extract and validate Bearer token if present
		token := extractBearerToken(r)
		if token != "" {
			_, ar, err := provider.IntrospectToken(r.Context(), token, fosite.AccessToken, NewSession(""))
			if err == nil {
				if session, ok := ar.GetSession().(*Session); ok {
					if creds := store.GetCredentials(session.Subject); creds != nil {
						ctx := context.WithValue(r.Context(), credentialsContextKey, creds)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		// For GET and DELETE, allow through without auth (MCP session setup/teardown)
		// Tool calls happen via POST and will fail gracefully if no credentials in context
		if r.Method == http.MethodGet || r.Method == http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		// For POST without valid token, return 401
		if token == "" {
			w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+baseURL+`/.well-known/oauth-protected-resource"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Token was present but invalid
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", resource_metadata="`+baseURL+`/.well-known/oauth-protected-resource"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Mcp-Protocol-Version")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// DevMiddleware injects pre-configured credentials into every request (no auth required).
// Only use for local development.
func DevMiddleware(creds *Credentials, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		ctx := context.WithValue(r.Context(), credentialsContextKey, creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

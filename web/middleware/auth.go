package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

// SetSessionStore sets the session store to be used by the middleware
func SetSessionStore(sessionStore *sessions.CookieStore) {
	store = sessionStore
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "Session store not configured", http.StatusInternalServerError)
			return
		}
		
		session, _ := store.Get(r, "session")
		
		userEmail, ok := session.Values["user_email"].(string)
		log.Printf("Middleware checking auth for %s: userEmail=%s, ok=%v", r.URL.Path, userEmail, ok)
		
		if !ok || userEmail == "" {
			log.Printf("User not authenticated, redirecting to home")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Add user email, user ID, and plan ID to context
		ctx := context.WithValue(r.Context(), "user_email", userEmail)
		
		if userID, ok := session.Values["user_id"].(int); ok {
			ctx = context.WithValue(ctx, "user_id", userID)
		}
		
		if planID, ok := session.Values["plan_id"].(int); ok {
			ctx = context.WithValue(ctx, "plan_id", planID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserEmail(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}

func GetUserID(r *http.Request) int {
	if userID, ok := r.Context().Value("user_id").(int); ok {
		return userID
	}
	return 0
}

func GetPlanID(r *http.Request) int {
	if planID, ok := r.Context().Value("plan_id").(int); ok {
		return planID
	}
	return 0
}
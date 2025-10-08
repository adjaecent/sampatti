package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/adjaecent/sampatti/internal/service"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authService  *service.AuthService
	oauth2Config *oauth2.Config
	store        *sessions.CookieStore
	templates    *template.Template
}

func NewAuthHandler(authService *service.AuthService, oauth2Config *oauth2.Config, store *sessions.CookieStore, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		oauth2Config: oauth2Config,
		store:        store,
		templates:    templates,
	}
}

func (h *AuthHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session")
	
	if userEmail, ok := session.Values["user_email"].(string); ok && userEmail != "" {
		// Add some debug logging
		log.Printf("User already logged in: %s, redirecting to dashboard", userEmail)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	h.templates.ExecuteTemplate(w, "home.html", nil)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "random-state-string" // In production, use a secure random string
	url := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := h.oauth2Config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Get or create user plan
	user, plan, err := h.authService.GetOrCreateUserPlan(userInfo.Email, userInfo.Name)
	if err != nil {
		http.Error(w, "Failed to get or create plan", http.StatusInternalServerError)
		return
	}

	session, _ := h.store.Get(r, "session")
	session.Values["user_email"] = userInfo.Email
	session.Values["user_id"] = user.ID
	session.Values["plan_id"] = plan.ID
	
	log.Printf("Saving session for user: %s, plan: %d", userInfo.Email, plan.ID)
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Printf("Redirecting to dashboard after successful login")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session")
	session.Values["user_email"] = ""
	session.Values["user_id"] = 0
	session.Values["plan_id"] = 0
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
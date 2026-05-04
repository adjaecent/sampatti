package oauth

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/ory/fosite"
)

//go:embed templates/authorize.html
var authorizeTemplate string

// Handlers holds the HTTP handlers for OAuth endpoints.
type Handlers struct {
	provider  fosite.OAuth2Provider
	store     *MemoryStore
	baseURL   string
	templates *template.Template
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(provider fosite.OAuth2Provider, store *MemoryStore, baseURL string) *Handlers {
	templates := template.New("oauth")
	template.Must(templates.New("authorize.html").Parse(authorizeTemplate))

	return &Handlers{
		provider:  provider,
		store:     store,
		baseURL:   baseURL,
		templates: templates,
	}
}

// HandleProtectedResourceMetadata serves the OAuth 2.0 Protected Resource Metadata (RFC 9728).
// This tells clients where to find the authorization server for this resource.
func (h *Handlers) HandleProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := map[string]interface{}{
		"resource":              h.baseURL + "/mcp",
		"authorization_servers": []string{h.baseURL},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// HandleMetadata serves the OAuth 2.0 Authorization Server Metadata (RFC 8414).
func (h *Handlers) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := map[string]interface{}{
		"issuer":                                h.baseURL,
		"authorization_endpoint":                h.baseURL + "/authorize",
		"token_endpoint":                        h.baseURL + "/token",
		"registration_endpoint":                 h.baseURL + "/register",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      []string{"portfolio", "offline_access"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// HandleRegister implements dynamic client registration (RFC 7591).
func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RedirectURIs []string `json:"redirect_uris"`
		ClientName   string   `json:"client_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.RedirectURIs) == 0 {
		http.Error(w, "redirect_uris is required", http.StatusBadRequest)
		return
	}

	client, err := h.store.RegisterClient(req.RedirectURIs)
	if err != nil {
		http.Error(w, fmt.Sprintf("registration failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"client_id":                client.ID,
		"client_name":              req.ClientName,
		"redirect_uris":           client.RedirectURIs,
		"grant_types":             client.GrantTypes,
		"response_types":          client.ResponseTypes,
		"token_endpoint_auth_method": "none",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleAuthorize handles both GET (show form) and POST (process form) for /authorize.
func (h *Handlers) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleAuthorizeGET(w, r)
	case http.MethodPost:
		h.handleAuthorizePOST(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handlers) handleAuthorizeGET(w http.ResponseWriter, r *http.Request) {
	// Parse the OAuth authorize request to validate parameters
	ar, err := h.provider.NewAuthorizeRequest(r.Context(), r)
	if err != nil {
		log.Printf("Authorize request error: %v", err)
		h.provider.WriteAuthorizeError(r.Context(), w, ar, err)
		return
	}

	// Render the authorization form, passing through OAuth params as hidden fields
	data := map[string]string{
		"ClientID":            ar.GetClient().GetID(),
		"RedirectURI":         r.URL.Query().Get("redirect_uri"),
		"State":               r.URL.Query().Get("state"),
		"Scope":               r.URL.Query().Get("scope"),
		"ResponseType":        r.URL.Query().Get("response_type"),
		"CodeChallenge":       r.URL.Query().Get("code_challenge"),
		"CodeChallengeMethod": r.URL.Query().Get("code_challenge_method"),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "authorize.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handlers) handleAuthorizePOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	// Reconstruct the OAuth query parameters from the hidden form fields
	// so fosite can validate the authorize request again
	q := url.Values{}
	q.Set("client_id", r.FormValue("client_id"))
	q.Set("redirect_uri", r.FormValue("redirect_uri"))
	q.Set("state", r.FormValue("state"))
	q.Set("scope", r.FormValue("scope"))
	q.Set("response_type", r.FormValue("response_type"))
	q.Set("code_challenge", r.FormValue("code_challenge"))
	q.Set("code_challenge_method", r.FormValue("code_challenge_method"))
	r.URL.RawQuery = q.Encode()

	ar, err := h.provider.NewAuthorizeRequest(r.Context(), r)
	if err != nil {
		log.Printf("Authorize POST request error: %v", err)
		h.provider.WriteAuthorizeError(r.Context(), w, ar, err)
		return
	}

	// Extract platform credentials from form
	creds := &Credentials{}

	kuveraUsername := r.FormValue("kuvera_username")
	kuveraPassword := r.FormValue("kuvera_password")
	if kuveraUsername != "" && kuveraPassword != "" {
		creds.Kuvera = &KuveraCredentials{
			Username: kuveraUsername,
			Password: kuveraPassword,
		}
	}

	stockalUsername := r.FormValue("stockal_username")
	stockalPassword := r.FormValue("stockal_password")
	if stockalUsername != "" && stockalPassword != "" {
		creds.Stockal = &StockalCredentials{
			Username: stockalUsername,
			Password: stockalPassword,
		}
	}

	if creds.Kuvera == nil && creds.Stockal == nil {
		// Re-render the form with an error
		data := map[string]string{
			"ClientID":            r.FormValue("client_id"),
			"RedirectURI":         r.FormValue("redirect_uri"),
			"State":               r.FormValue("state"),
			"Scope":               r.FormValue("scope"),
			"ResponseType":        r.FormValue("response_type"),
			"CodeChallenge":       r.FormValue("code_challenge"),
			"CodeChallengeMethod": r.FormValue("code_challenge_method"),
			"Error":               "Please provide credentials for at least one platform",
		}
		w.Header().Set("Content-Type", "text/html")
		h.templates.ExecuteTemplate(w, "authorize.html", data)
		return
	}

	// Generate a subject ID and store credentials in memory
	subject, err := generateID("user_")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.store.StoreCredentials(subject, creds)

	// Create session and authorize
	session := NewSession(subject)

	// Grant requested scopes
	for _, scope := range ar.GetRequestedScopes() {
		ar.GrantScope(scope)
	}

	// Create the authorize response (generates auth code)
	response, err := h.provider.NewAuthorizeResponse(r.Context(), ar, session)
	if err != nil {
		log.Printf("Authorize response error: %v", err)
		h.provider.WriteAuthorizeError(r.Context(), w, ar, err)
		return
	}

	// Write the redirect with the authorization code
	h.provider.WriteAuthorizeResponse(r.Context(), w, ar, response)
}

// HandleToken handles the token exchange endpoint (POST /token).
func (h *Handlers) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := NewSession("")

	// Let fosite handle the token request (validates auth code + PKCE)
	ar, err := h.provider.NewAccessRequest(r.Context(), r, session)
	if err != nil {
		log.Printf("Token request error: %v", err)
		h.provider.WriteAccessError(r.Context(), w, ar, err)
		return
	}

	// Grant requested scopes
	for _, scope := range ar.GetRequestedScopes() {
		ar.GrantScope(scope)
	}

	// Generate the access token response
	response, err := h.provider.NewAccessResponse(r.Context(), ar)
	if err != nil {
		log.Printf("Token response error: %v", err)
		h.provider.WriteAccessError(r.Context(), w, ar, err)
		return
	}

	// Write the token response
	h.provider.WriteAccessResponse(r.Context(), w, ar, response)
}

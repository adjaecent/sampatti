package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/adjaecent/sampatti/internal/auth"
)

//go:embed templates/home.html
var homeTemplate string

//go:embed templates/success.html
var successTemplate string

//go:embed static/css/picnic.min.css
var picnicCSS []byte

type AuthServer struct {
	authRequestMgr *auth.AuthRequestManager
	port           string
	templates      *template.Template
}

func NewAuthServer(authRequestMgr *auth.AuthRequestManager, port string) *AuthServer {
	templates := template.New("embedded")
	template.Must(templates.New("home.html").Parse(homeTemplate))
	template.Must(templates.New("success.html").Parse(successTemplate))

	return &AuthServer{
		authRequestMgr: authRequestMgr,
		port:           port,
		templates:      templates,
	}
}

func (s *AuthServer) Start() {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/auth/", s.handleAuthRequest)
	http.HandleFunc("/success", s.handleSuccess)
	http.HandleFunc("/static/css/picnic.min.css", s.servePicnicCSS)

	log.Printf("Auth request server starting on http://localhost:%s", s.port)
	if err := http.ListenAndServe(":"+s.port, nil); err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
	}
}

func (s *AuthServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "home.html", nil); err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *AuthServer) handleAuthRequest(w http.ResponseWriter, r *http.Request) {
	// Extract request ID from URL path /auth/{requestID}
	path := r.URL.Path
	if len(path) < 7 { // "/auth/" is 6 chars
		http.Error(w, "Invalid auth request URL", http.StatusBadRequest)
		return
	}
	requestID := path[6:] // Remove "/auth/" prefix

	if r.Method == "GET" {
		// Show the auth form for this specific request
		if err := s.templates.ExecuteTemplate(w, "home.html", map[string]string{
			"RequestID": requestID,
		}); err != nil {
			log.Printf("Error executing home template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		creds := &auth.Credentials{}

		// Handle Kuvera credentials if provided
		kuveraUsername := r.FormValue("kuvera_username")
		kuveraPassword := r.FormValue("kuvera_password")
		if kuveraUsername != "" && kuveraPassword != "" {
			creds.Kuvera = &auth.KuveraCredentials{
				Username: kuveraUsername,
				Password: kuveraPassword,
			}
		}

		// Handle Stockal credentials if provided
		stockalUsername := r.FormValue("stockal_username")
		stockalPassword := r.FormValue("stockal_password")
		if stockalUsername != "" && stockalPassword != "" {
			creds.Stockal = &auth.StockalCredentials{
				Username: stockalUsername,
				Password: stockalPassword,
			}
		}

		// Validate at least one platform is configured
		if creds.Kuvera == nil && creds.Stockal == nil {
			http.Error(w, "Please provide credentials for at least one platform", http.StatusBadRequest)
			return
		}

		// Complete the auth request
		token, err := s.authRequestMgr.CompleteAuthRequest(requestID, creds)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to complete auth request: %v", err), http.StatusBadRequest)
			return
		}

		// Redirect to success page
		http.Redirect(w, r, fmt.Sprintf("/success?token=%s", token), http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *AuthServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Verify token exists and get credentials
	credentials, err := s.authRequestMgr.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	data := struct {
		Token           string
		HasKuvera       bool
		HasStockal      bool
		KuveraUsername  string
		StockalUsername string
	}{
		Token:      token,
		HasKuvera:  credentials.Kuvera != nil,
		HasStockal: credentials.Stockal != nil,
		KuveraUsername: func() string {
			if credentials.Kuvera != nil {
				return credentials.Kuvera.Username
			}
			return ""
		}(),
		StockalUsername: func() string {
			if credentials.Stockal != nil {
				return credentials.Stockal.Username
			}
			return ""
		}(),
	}

	if err := s.templates.ExecuteTemplate(w, "success.html", data); err != nil {
		log.Printf("Error executing success template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// CSS serving handlers
func (s *AuthServer) servePicnicCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(picnicCSS)
}

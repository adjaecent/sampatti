package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/adjaecent/sampatti/config"
	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/internal/service"
	"github.com/adjaecent/sampatti/web/handlers"
	"github.com/adjaecent/sampatti/web/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func (tr *TemplateRenderer) Render(w http.ResponseWriter, templateName string, data interface{}) error {
	tmpl, ok := tr.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s does not exist", templateName)
	}
	
	// For home template, execute directly
	if templateName == "home.html" {
		return tmpl.ExecuteTemplate(w, "home.html", data)
	}
	
	// For other templates, execute the "base" template
	return tmpl.ExecuteTemplate(w, "base", data)
}

type WebHandler struct {
	authHandler    *handlers.AuthHandler
	dashHandler    *handlers.DashboardHandler
	accountHandler *handlers.AccountHandler
}

func New(db *sql.DB, cfg *config.Config) *WebHandler {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	planRepo := repository.NewPlanRepository(db)
	investmentAccountRepo := repository.NewInvestmentAccountRepository(db)
	fetchRepo := repository.NewFetchRepository(db)
	preciousMetalsRepo := repository.NewPreciousMetalsRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, planRepo)
	investmentAccountService := service.NewInvestmentAccountService(investmentAccountRepo, authService)
	dashboardService := service.NewDashboardService(planRepo, userRepo, investmentAccountRepo, fetchRepo, preciousMetalsRepo)
	fetchService := service.NewFetchService(investmentAccountRepo, fetchRepo, authService)
	preciousMetalsService := service.NewPreciousMetalsService(preciousMetalsRepo)

	// Initialize OAuth config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleSecret,
		RedirectURL:  cfg.BaseURL + "/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// Initialize session store
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	// Configure middleware with the same session store
	middleware.SetSessionStore(store)

	// Initialize templates with custom functions
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	
	// Create a map to store pre-loaded templates
	templateMap := make(map[string]*template.Template)
	
	// List of page templates (excluding base and home)
	pageTemplates := []string{
		"dashboard.html",
		"accounts.html", 
		"add_account.html",
		"add_precious_metal.html",
		"edit_precious_metal.html",
	}
	
	// Pre-load each page template with base
	for _, page := range pageTemplates {
		tmpl := template.New("").Funcs(funcMap)
		tmpl = template.Must(tmpl.ParseFiles("templates/base.html", "templates/"+page))
		templateMap[page] = tmpl
	}
	
	// Load home template separately (standalone)
	homeTemplate := template.New("").Funcs(funcMap)
	homeTemplate = template.Must(homeTemplate.ParseFiles("templates/home.html"))
	templateMap["home.html"] = homeTemplate

	// Create template renderer
	templateRenderer := &TemplateRenderer{templates: templateMap}
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, oauth2Config, store, templateRenderer)
	dashHandler := handlers.NewDashboardHandler(dashboardService, fetchService, preciousMetalsService, store, templateRenderer)
	accountHandler := handlers.NewAccountHandler(investmentAccountService, store, templateRenderer)

	return &WebHandler{
		authHandler:    authHandler,
		dashHandler:    dashHandler,
		accountHandler: accountHandler,
	}
}

func (w *WebHandler) SetupRoutes(r *mux.Router) {
	// Public routes
	r.HandleFunc("/", w.authHandler.Home).Methods("GET")
	r.HandleFunc("/auth/google", w.authHandler.GoogleLogin).Methods("GET")
	r.HandleFunc("/auth/google/callback", w.authHandler.GoogleCallback).Methods("GET")
	r.HandleFunc("/logout", w.authHandler.Logout).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireAuth)
	protected.HandleFunc("/dashboard", w.dashHandler.Dashboard).Methods("GET")
	protected.HandleFunc("/accounts", w.accountHandler.Accounts).Methods("GET")
	protected.HandleFunc("/accounts/add", w.accountHandler.AddAccount).Methods("GET", "POST")
	protected.HandleFunc("/accounts/{id}/delete", w.accountHandler.DeleteAccount).Methods("POST")
	protected.HandleFunc("/fetch", w.dashHandler.FetchData).Methods("POST")
	protected.HandleFunc("/precious-metals/add", w.dashHandler.AddPreciousMetal).Methods("GET", "POST")
	protected.HandleFunc("/precious-metals/{id}/edit", w.dashHandler.EditPreciousMetal).Methods("GET", "POST")
	protected.HandleFunc("/precious-metals/{id}/delete", w.dashHandler.DeletePreciousMetal).Methods("POST")
}

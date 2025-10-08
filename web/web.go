package web

import (
	"database/sql"
	"html/template"
	"net/http"

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

	// Initialize services
	authService := service.NewAuthService(userRepo, planRepo)
	investmentAccountService := service.NewInvestmentAccountService(investmentAccountRepo, authService)
	dashboardService := service.NewDashboardService(planRepo, userRepo, investmentAccountRepo, fetchRepo)
	fetchService := service.NewFetchService(investmentAccountRepo, fetchRepo, authService)

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

	// Initialize templates - Parse each template separately to avoid conflicts
	templates := template.New("")
	templates = template.Must(templates.ParseFiles(
		"templates/base.html",
		"templates/home.html",
		"templates/dashboard.html",
		"templates/accounts.html",
		"templates/add_account.html",
	))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, oauth2Config, store, templates)
	dashHandler := handlers.NewDashboardHandler(dashboardService, fetchService, store, templates)
	accountHandler := handlers.NewAccountHandler(investmentAccountService, store, templates)

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
}

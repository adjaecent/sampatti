package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/adjaecent/sampatti/internal/service"
	"github.com/adjaecent/sampatti/web/middleware"
	"github.com/gorilla/sessions"
)

type DashboardHandler struct {
	dashService  *service.DashboardService
	fetchService *service.FetchService
	store        *sessions.CookieStore
	templates    *template.Template
}

func NewDashboardHandler(dashService *service.DashboardService, fetchService *service.FetchService, store *sessions.CookieStore, templates *template.Template) *DashboardHandler {
	return &DashboardHandler{
		dashService:  dashService,
		fetchService: fetchService,
		store:        store,
		templates:    templates,
	}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	planID := middleware.GetPlanID(r)
	userID := middleware.GetUserID(r)
	log.Printf("Dashboard request - planID: %d, userID: %d", planID, userID)
	
	if planID == 0 || userID == 0 {
		log.Printf("Invalid planID or userID, clearing session")
		// Clear the session and redirect to home
		session, _ := h.store.Get(r, "session")
		session.Values["user_email"] = ""
		session.Values["user_id"] = 0
		session.Values["plan_id"] = 0
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data, err := h.dashService.GetDashboardData(planID, userID)
	if err != nil {
		log.Printf("Failed to load dashboard data: %v", err)
		// If the plan/user doesn't exist in DB, clear session and redirect
		log.Printf("Clearing stale session and redirecting to home")
		session, _ := h.store.Get(r, "session")
		session.Values["user_email"] = ""
		session.Values["user_id"] = 0
		session.Values["plan_id"] = 0
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.templates.ExecuteTemplate(w, "dashboard.html", data)
}

func (h *DashboardHandler) FetchData(w http.ResponseWriter, r *http.Request) {
	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	err := h.fetchService.FetchAllData(planID)
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
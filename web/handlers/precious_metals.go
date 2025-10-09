package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/adjaecent/sampatti/internal/service"
	"github.com/adjaecent/sampatti/web/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type PreciousMetalsHandler struct {
	preciousMetalsService *service.PreciousMetalsService
	store                 *sessions.CookieStore
	templates             *template.Template
}

func NewPreciousMetalsHandler(preciousMetalsService *service.PreciousMetalsService, store *sessions.CookieStore, templates *template.Template) *PreciousMetalsHandler {
	return &PreciousMetalsHandler{
		preciousMetalsService: preciousMetalsService,
		store:                 store,
		templates:             templates,
	}
}

func (h *PreciousMetalsHandler) Holdings(w http.ResponseWriter, r *http.Request) {
	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	holdings, err := h.preciousMetalsService.GetHoldingsByPlan(planID)
	if err != nil {
		http.Error(w, "Failed to load holdings", http.StatusInternalServerError)
		return
	}

	summary, err := h.preciousMetalsService.GetSummaryByPlan(planID)
	if err != nil {
		http.Error(w, "Failed to load summary", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Holdings": holdings,
		"Summary":  summary,
	}

	h.templates.ExecuteTemplate(w, "precious_metals.html", data)
}

func (h *PreciousMetalsHandler) AddHolding(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "add_precious_metal.html", nil)
		return
	}

	// POST: Add new holding
	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	metalType := r.FormValue("metal_type")
	quantityStr := r.FormValue("quantity")
	notes := r.FormValue("notes")

	if metalType == "" || quantityStr == "" {
		http.Error(w, "Metal type and quantity are required", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	err = h.preciousMetalsService.AddHolding(planID, metalType, quantity, notes)
	if err != nil {
		http.Error(w, "Failed to add holding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/precious-metals", http.StatusSeeOther)
}

func (h *PreciousMetalsHandler) EditHolding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	holdingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid holding ID", http.StatusBadRequest)
		return
	}

	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		holding, err := h.preciousMetalsService.GetHoldingByID(holdingID)
		if err != nil || holding.PlanID != planID {
			http.Error(w, "Holding not found", http.StatusNotFound)
			return
		}

		h.templates.ExecuteTemplate(w, "edit_precious_metal.html", holding)
		return
	}

	// POST: Update holding
	quantityStr := r.FormValue("quantity")
	notes := r.FormValue("notes")

	if quantityStr == "" {
		http.Error(w, "Quantity is required", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	err = h.preciousMetalsService.UpdateHolding(holdingID, planID, quantity, notes)
	if err != nil {
		http.Error(w, "Failed to update holding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/precious-metals", http.StatusSeeOther)
}

func (h *PreciousMetalsHandler) DeleteHolding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	holdingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid holding ID", http.StatusBadRequest)
		return
	}

	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	err = h.preciousMetalsService.DeleteHolding(holdingID, planID)
	if err != nil {
		http.Error(w, "Failed to delete holding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/precious-metals", http.StatusSeeOther)
}
package handlers

import (
	"net/http"
	"strconv"
	"log"

	"github.com/adjaecent/sampatti/internal/service"
	"github.com/adjaecent/sampatti/web/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type AccountHandler struct {
	investmentAccountService *service.InvestmentAccountService
	store                    *sessions.CookieStore
	templates                TemplateRenderer
}

func NewAccountHandler(investmentAccountService *service.InvestmentAccountService, store *sessions.CookieStore, templates TemplateRenderer) *AccountHandler {
	return &AccountHandler{
		investmentAccountService: investmentAccountService,
		store:                    store,
		templates:                templates,
	}
}

func (h *AccountHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	accounts, err := h.investmentAccountService.GetAccountsByPlan(planID)
	if err != nil {
		http.Error(w, "Failed to load accounts", http.StatusInternalServerError)
		return
	}

	h.templates.Render(w, "accounts.html", map[string]interface{}{
		"InvestmentAccounts": accounts,
	})
}

func (h *AccountHandler) AddAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.Render(w, "add_account.html", nil)
		return
	}

	log.Printf("Adding account...")


	// POST: Add new account
	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	log.Printf("Adding account... %s", planID)

	platform := r.FormValue("platform")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if platform == "" || username == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	err := h.investmentAccountService.CreateAccount(planID, platform, username, password)
	if err != nil {
		http.Error(w, "Failed to add account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/accounts", http.StatusSeeOther)
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	planID := middleware.GetPlanID(r)
	if planID == 0 {
		http.Error(w, "Plan not found", http.StatusBadRequest)
		return
	}

	err = h.investmentAccountService.DeleteAccount(accountID, planID)
	if err != nil {
		http.Error(w, "Failed to delete account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/accounts", http.StatusSeeOther)
}

package models

import (
	"time"
)

type KuveraFetch struct {
	ID                    int       `json:"id" db:"id"`
	InvestmentAccountID   int       `json:"investment_account_id" db:"investment_account_id"`
	PortfolioData         *string   `json:"portfolio_data" db:"portfolio_data"`
	HoldingsData          *string   `json:"holdings_data" db:"holdings_data"`
	GoldPriceData         *string   `json:"gold_price_data" db:"gold_price_data"`
	FetchedAt             time.Time `json:"fetched_at" db:"fetched_at"`
	Success               bool      `json:"success" db:"success"`
	ErrorMessage          *string   `json:"error_message" db:"error_message"`
}

type StockalFetch struct {
	ID                    int       `json:"id" db:"id"`
	InvestmentAccountID   int       `json:"investment_account_id" db:"investment_account_id"`
	AccountSummaryData    *string   `json:"account_summary_data" db:"account_summary_data"`
	PortfolioData         *string   `json:"portfolio_data" db:"portfolio_data"`
	FetchedAt             time.Time `json:"fetched_at" db:"fetched_at"`
	Success               bool      `json:"success" db:"success"`
	ErrorMessage          *string   `json:"error_message" db:"error_message"`
}

type DashboardData struct {
	Plan                  *Plan                 `json:"plan"`
	User                  *User                 `json:"user"`
	InvestmentAccounts    []InvestmentAccount   `json:"investment_accounts"`
	LatestKuvera          []KuveraFetch         `json:"latest_kuvera"`
	LatestStockal         []StockalFetch        `json:"latest_stockal"`
	TotalValue            float64               `json:"total_value"`
	LastFetchTime         *time.Time            `json:"last_fetch_time"`
	PreciousMetalsHoldings []PreciousMetalHolding `json:"precious_metals_holdings"`
	PreciousMetalsSummary  []PreciousMetalSummary `json:"precious_metals_summary"`
}
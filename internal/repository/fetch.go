package repository

import (
	"database/sql"

	"github.com/adjaecent/sampatti/models"
)

type FetchRepository struct {
	db *sql.DB
}

func NewFetchRepository(db *sql.DB) *FetchRepository {
	return &FetchRepository{db: db}
}

// Kuvera Fetch Methods
func (r *FetchRepository) CreateKuveraFetch(investmentAccountID int, portfolioData, holdingsData, goldPriceData *string, success bool, errorMessage string) error {
	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	_, err := r.db.Exec(`
		INSERT INTO kuvera_fetches (investment_account_id, portfolio_data, holdings_data, gold_price_data, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?)
	`, investmentAccountID, portfolioData, holdingsData, goldPriceData, success, errMsg)
	
	return err
}

func (r *FetchRepository) GetLatestKuveraFetches(planID int, limit int) ([]models.KuveraFetch, error) {
	query := `
		SELECT kf.id, kf.investment_account_id, kf.portfolio_data, kf.holdings_data, 
			   kf.gold_price_data, kf.fetched_at, kf.success, kf.error_message
		FROM kuvera_fetches kf
		JOIN investment_accounts ia ON kf.investment_account_id = ia.id
		WHERE ia.plan_id = ? AND kf.success = true
		ORDER BY kf.fetched_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, planID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fetches []models.KuveraFetch
	for rows.Next() {
		var fetch models.KuveraFetch
		err := rows.Scan(&fetch.ID, &fetch.InvestmentAccountID, &fetch.PortfolioData,
			&fetch.HoldingsData, &fetch.GoldPriceData, &fetch.FetchedAt,
			&fetch.Success, &fetch.ErrorMessage)
		if err != nil {
			return nil, err
		}
		fetches = append(fetches, fetch)
	}

	return fetches, nil
}

func (r *FetchRepository) GetLatestKuveraFetchPerAccount(planID int) ([]models.KuveraFetch, error) {
	query := `
		SELECT kf.id, kf.investment_account_id, kf.portfolio_data, kf.holdings_data, 
			   kf.gold_price_data, kf.fetched_at, kf.success, kf.error_message
		FROM kuvera_fetches kf
		JOIN investment_accounts ia ON kf.investment_account_id = ia.id
		WHERE ia.plan_id = ? AND kf.success = true
		  AND kf.fetched_at = (
			  SELECT MAX(kf2.fetched_at) 
			  FROM kuvera_fetches kf2 
			  WHERE kf2.investment_account_id = kf.investment_account_id 
			    AND kf2.success = true
		  )
		ORDER BY kf.fetched_at DESC
	`

	rows, err := r.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fetches []models.KuveraFetch
	for rows.Next() {
		var fetch models.KuveraFetch
		err := rows.Scan(
			&fetch.ID,
			&fetch.InvestmentAccountID,
			&fetch.PortfolioData,
			&fetch.HoldingsData,
			&fetch.GoldPriceData,
			&fetch.FetchedAt,
			&fetch.Success,
			&fetch.ErrorMessage,
		)
		if err != nil {
			return nil, err
		}
		fetches = append(fetches, fetch)
	}

	return fetches, nil
}

func (r *FetchRepository) GetLatestStockalFetchPerAccount(planID int) ([]models.StockalFetch, error) {
	query := `
		SELECT sf.id, sf.investment_account_id, sf.account_summary_data, sf.portfolio_data,
			   sf.fetched_at, sf.success, sf.error_message
		FROM stockal_fetches sf
		JOIN investment_accounts ia ON sf.investment_account_id = ia.id
		WHERE ia.plan_id = ? AND sf.success = true
		  AND sf.fetched_at = (
			  SELECT MAX(sf2.fetched_at) 
			  FROM stockal_fetches sf2 
			  WHERE sf2.investment_account_id = sf.investment_account_id 
			    AND sf2.success = true
		  )
		ORDER BY sf.fetched_at DESC
	`

	rows, err := r.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fetches []models.StockalFetch
	for rows.Next() {
		var fetch models.StockalFetch
		err := rows.Scan(&fetch.ID, &fetch.InvestmentAccountID, &fetch.AccountSummaryData,
			&fetch.PortfolioData, &fetch.FetchedAt, &fetch.Success, &fetch.ErrorMessage)
		if err != nil {
			return nil, err
		}
		fetches = append(fetches, fetch)
	}

	return fetches, nil
}

// Stockal Fetch Methods
func (r *FetchRepository) CreateStockalFetch(investmentAccountID int, accountSummaryData, portfolioData *string, success bool, errorMessage string) error {
	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	_, err := r.db.Exec(`
		INSERT INTO stockal_fetches (investment_account_id, account_summary_data, portfolio_data, success, error_message)
		VALUES (?, ?, ?, ?, ?)
	`, investmentAccountID, accountSummaryData, portfolioData, success, errMsg)
	
	return err
}

func (r *FetchRepository) GetLatestStockalFetches(planID int, limit int) ([]models.StockalFetch, error) {
	query := `
		SELECT sf.id, sf.investment_account_id, sf.account_summary_data, sf.portfolio_data,
			   sf.fetched_at, sf.success, sf.error_message
		FROM stockal_fetches sf
		JOIN investment_accounts ia ON sf.investment_account_id = ia.id
		WHERE ia.plan_id = ? AND sf.success = true
		ORDER BY sf.fetched_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, planID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fetches []models.StockalFetch
	for rows.Next() {
		var fetch models.StockalFetch
		err := rows.Scan(&fetch.ID, &fetch.InvestmentAccountID, &fetch.AccountSummaryData,
			&fetch.PortfolioData, &fetch.FetchedAt, &fetch.Success, &fetch.ErrorMessage)
		if err != nil {
			return nil, err
		}
		fetches = append(fetches, fetch)
	}

	return fetches, nil
}

package repository

import (
	"database/sql"

	"github.com/adjaecent/sampatti/models"
)

type InvestmentAccountRepository struct {
	db *sql.DB
}

func NewInvestmentAccountRepository(db *sql.DB) *InvestmentAccountRepository {
	return &InvestmentAccountRepository{db: db}
}

func (r *InvestmentAccountRepository) GetByPlanID(planID int) ([]models.InvestmentAccount, error) {
	rows, err := r.db.Query("SELECT id, plan_id, platform, username, password_hash, created_at FROM investment_accounts WHERE plan_id = ?", planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.InvestmentAccount
	for rows.Next() {
		var account models.InvestmentAccount
		err := rows.Scan(&account.ID, &account.PlanID, &account.Platform, &account.Username, &account.PasswordHash, &account.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *InvestmentAccountRepository) GetByID(id int) (*models.InvestmentAccount, error) {
	var account models.InvestmentAccount
	err := r.db.QueryRow("SELECT id, plan_id, platform, username, password_hash, created_at FROM investment_accounts WHERE id = ?", id).
		Scan(&account.ID, &account.PlanID, &account.Platform, &account.Username, &account.PasswordHash, &account.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *InvestmentAccountRepository) Create(planID int, platform models.Platform, username, passwordHash string) error {
	_, err := r.db.Exec(
		"INSERT INTO investment_accounts (plan_id, platform, username, password_hash) VALUES (?, ?, ?, ?)",
		planID, platform, username, passwordHash,
	)
	return err
}

func (r *InvestmentAccountRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM investment_accounts WHERE id = ?", id)
	return err
}

func (r *InvestmentAccountRepository) VerifyOwnership(accountID, planID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM investment_accounts WHERE id = ? AND plan_id = ?", accountID, planID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
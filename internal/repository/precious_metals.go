package repository

import (
	"database/sql"

	"github.com/adjaecent/sampatti/models"
)

type PreciousMetalsRepository struct {
	db *sql.DB
}

func NewPreciousMetalsRepository(db *sql.DB) *PreciousMetalsRepository {
	return &PreciousMetalsRepository{db: db}
}

func (r *PreciousMetalsRepository) Create(planID int, metalType string, quantity float64, notes string) error {
	_, err := r.db.Exec(`
		INSERT INTO precious_metal_holdings (plan_id, metal_type, quantity, notes)
		VALUES (?, ?, ?, ?)
	`, planID, metalType, quantity, notes)
	return err
}

func (r *PreciousMetalsRepository) GetByPlanID(planID int) ([]models.PreciousMetalHolding, error) {
	query := `
		SELECT id, plan_id, metal_type, quantity, notes, created_at, updated_at
		FROM precious_metal_holdings
		WHERE plan_id = ?
		ORDER BY metal_type, created_at DESC
	`

	rows, err := r.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.PreciousMetalHolding
	for rows.Next() {
		var holding models.PreciousMetalHolding
		err := rows.Scan(
			&holding.ID,
			&holding.PlanID,
			&holding.MetalType,
			&holding.Quantity,
			&holding.Notes,
			&holding.CreatedAt,
			&holding.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}

	return holdings, nil
}

func (r *PreciousMetalsRepository) GetByID(id int) (*models.PreciousMetalHolding, error) {
	query := `
		SELECT id, plan_id, metal_type, quantity, notes, created_at, updated_at
		FROM precious_metal_holdings
		WHERE id = ?
	`

	var holding models.PreciousMetalHolding
	err := r.db.QueryRow(query, id).Scan(
		&holding.ID,
		&holding.PlanID,
		&holding.MetalType,
		&holding.Quantity,
		&holding.Notes,
		&holding.CreatedAt,
		&holding.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &holding, nil
}

func (r *PreciousMetalsRepository) Update(id int, quantity float64, notes string) error {
	_, err := r.db.Exec(`
		UPDATE precious_metal_holdings
		SET quantity = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, quantity, notes, id)
	return err
}

func (r *PreciousMetalsRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM precious_metal_holdings WHERE id = ?", id)
	return err
}

func (r *PreciousMetalsRepository) GetSummaryByPlanID(planID int) ([]models.PreciousMetalSummary, error) {
	query := `
		SELECT metal_type, SUM(quantity) as total_grams
		FROM precious_metal_holdings
		WHERE plan_id = ?
		GROUP BY metal_type
		ORDER BY metal_type
	`

	rows, err := r.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.PreciousMetalSummary
	for rows.Next() {
		var summary models.PreciousMetalSummary
		err := rows.Scan(&summary.MetalType, &summary.TotalGrams)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r *PreciousMetalsRepository) VerifyOwnership(holdingID, planID int) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM precious_metal_holdings WHERE id = ? AND plan_id = ?",
		holdingID, planID,
	).Scan(&count)

	return count > 0, err
}

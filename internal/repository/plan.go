package repository

import (
	"database/sql"

	"github.com/adjaecent/sampatti/models"
)

type PlanRepository struct {
	db *sql.DB
}

func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) GetByID(id int) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.QueryRow("SELECT id, name, owner_user_id, created_at FROM plans WHERE id = ?", id).
		Scan(&plan.ID, &plan.Name, &plan.OwnerUserID, &plan.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) GetByOwnerUserID(userID int) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.QueryRow("SELECT id, name, owner_user_id, created_at FROM plans WHERE owner_user_id = ?", userID).
		Scan(&plan.ID, &plan.Name, &plan.OwnerUserID, &plan.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) GetByMemberUserID(userID int) (*models.Plan, error) {
	query := `
		SELECT p.id, p.name, p.owner_user_id, p.created_at 
		FROM plans p 
		JOIN plan_members pm ON p.id = pm.plan_id 
		WHERE pm.user_id = ?
	`
	var plan models.Plan
	err := r.db.QueryRow(query, userID).
		Scan(&plan.ID, &plan.Name, &plan.OwnerUserID, &plan.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) Create(name string, ownerUserID int) (*models.Plan, error) {
	result, err := r.db.Exec("INSERT INTO plans (name, owner_user_id) VALUES (?, ?)", name, ownerUserID)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}

func (r *PlanRepository) GetAllPlanIDs() ([]int, error) {
	rows, err := r.db.Query("SELECT id FROM plans")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []int
	for rows.Next() {
		var planID int
		if err := rows.Scan(&planID); err != nil {
			return nil, err
		}
		plans = append(plans, planID)
	}

	return plans, nil
}
package repository

import (
	"database/sql"

	"github.com/adjaecent/sampatti/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow("SELECT id, email, name, google_id, created_at FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Email, &user.Name, &user.GoogleID, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow("SELECT id, email, name, google_id, created_at FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.Name, &user.GoogleID, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(email, name string, googleID *string) (*models.User, error) {
	result, err := r.db.Exec("INSERT INTO users (email, name, google_id) VALUES (?, ?, ?)", email, name, googleID)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}
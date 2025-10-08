package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	GoogleID  *string   `json:"google_id" db:"google_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
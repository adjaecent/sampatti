package models

import (
	"time"
)

type Plan struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	OwnerUserID int       `json:"owner_user_id" db:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type PlanMember struct {
	ID              int       `json:"id" db:"id"`
	PlanID          int       `json:"plan_id" db:"plan_id"`
	UserID          int       `json:"user_id" db:"user_id"`
	InvitedByUserID int       `json:"invited_by_user_id" db:"invited_by_user_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
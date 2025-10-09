package models

import (
	"time"
)

type PreciousMetalHolding struct {
	ID        int       `json:"id" db:"id"`
	PlanID    int       `json:"plan_id" db:"plan_id"`
	MetalType string    `json:"metal_type" db:"metal_type"` // "gold" or "silver"
	Quantity  float64   `json:"quantity" db:"quantity"`     // in grams
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type PreciousMetalSummary struct {
	MetalType    string  `json:"metal_type"`
	TotalGrams   float64 `json:"total_grams"`
	CurrentPrice float64 `json:"current_price"` // per gram
	TotalValue   float64 `json:"total_value"`
}
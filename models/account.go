package models

import (
	"time"
)

type Platform string

const (
	PlatformKuvera  Platform = "kuvera"
	PlatformStockal Platform = "stockal"
)

// Renamed to InvestmentAccount to clarify these are external platform credentials
type InvestmentAccount struct {
	ID           int       `json:"id" db:"id"`
	PlanID       int       `json:"plan_id" db:"plan_id"`
	Platform     Platform  `json:"platform" db:"platform"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // Don't expose in JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

func (a InvestmentAccount) IsPlatform(platform Platform) bool {
	return a.Platform == platform
}

func (a InvestmentAccount) IsKuvera() bool {
	return a.Platform == PlatformKuvera
}

func (a InvestmentAccount) IsStockal() bool {
	return a.Platform == PlatformStockal
}

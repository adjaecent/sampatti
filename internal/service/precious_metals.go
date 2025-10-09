package service

import (
	"errors"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/models"
)

type PreciousMetalsService struct {
	preciousMetalsRepo *repository.PreciousMetalsRepository
}

func NewPreciousMetalsService(preciousMetalsRepo *repository.PreciousMetalsRepository) *PreciousMetalsService {
	return &PreciousMetalsService{
		preciousMetalsRepo: preciousMetalsRepo,
	}
}

func (s *PreciousMetalsService) AddHolding(planID int, metalType string, quantity float64, notes string) error {
	// Validate metal type
	if metalType != "gold" && metalType != "silver" {
		return errors.New("metal type must be 'gold' or 'silver'")
	}

	// Validate quantity
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	return s.preciousMetalsRepo.Create(planID, metalType, quantity, notes)
}

func (s *PreciousMetalsService) GetHoldingsByPlan(planID int) ([]models.PreciousMetalHolding, error) {
	return s.preciousMetalsRepo.GetByPlanID(planID)
}

func (s *PreciousMetalsService) GetSummaryByPlan(planID int) ([]models.PreciousMetalSummary, error) {
	return s.preciousMetalsRepo.GetSummaryByPlanID(planID)
}

func (s *PreciousMetalsService) UpdateHolding(holdingID, planID int, quantity float64, notes string) error {
	// Verify ownership
	owned, err := s.preciousMetalsRepo.VerifyOwnership(holdingID, planID)
	if err != nil {
		return err
	}
	if !owned {
		return errors.New("holding not found or access denied")
	}

	// Validate quantity
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	return s.preciousMetalsRepo.Update(holdingID, quantity, notes)
}

func (s *PreciousMetalsService) DeleteHolding(holdingID, planID int) error {
	// Verify ownership
	owned, err := s.preciousMetalsRepo.VerifyOwnership(holdingID, planID)
	if err != nil {
		return err
	}
	if !owned {
		return errors.New("holding not found or access denied")
	}

	return s.preciousMetalsRepo.Delete(holdingID)
}

func (s *PreciousMetalsService) GetHoldingByID(holdingID int) (*models.PreciousMetalHolding, error) {
	return s.preciousMetalsRepo.GetByID(holdingID)
}
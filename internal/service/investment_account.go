package service

import (
	"errors"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/models"
)

type InvestmentAccountService struct {
	investmentAccountRepo *repository.InvestmentAccountRepository
	authService           *AuthService
}

func NewInvestmentAccountService(investmentAccountRepo *repository.InvestmentAccountRepository, authService *AuthService) *InvestmentAccountService {
	return &InvestmentAccountService{
		investmentAccountRepo: investmentAccountRepo,
		authService:           authService,
	}
}

func (s *InvestmentAccountService) GetAccountsByPlan(planID int) ([]models.InvestmentAccount, error) {
	return s.investmentAccountRepo.GetByPlanID(planID)
}

func (s *InvestmentAccountService) CreateAccount(planID int, platform, username, password string) error {
	// Validate platform
	if err := s.authService.ValidatePlatform(platform); err != nil {
		return err
	}

	// Validate input
	if username == "" || password == "" {
		return errors.New("username and password are required")
	}

	// Encrypt password
	encryptedPassword, err := s.authService.EncryptPassword(password)
	if err != nil {
		return err
	}

	return s.investmentAccountRepo.Create(planID, models.Platform(platform), username, encryptedPassword)
}

func (s *InvestmentAccountService) DeleteAccount(accountID, planID int) error {
	// Verify ownership
	owned, err := s.investmentAccountRepo.VerifyOwnership(accountID, planID)
	if err != nil {
		return err
	}
	
	if !owned {
		return errors.New("account not found or access denied")
	}

	return s.investmentAccountRepo.Delete(accountID)
}

func (s *InvestmentAccountService) GetAccountByID(id int) (*models.InvestmentAccount, error) {
	return s.investmentAccountRepo.GetByID(id)
}
package service

import (
	"encoding/json"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/models"
)

type DashboardService struct {
	planRepo              *repository.PlanRepository
	userRepo              *repository.UserRepository
	investmentAccountRepo *repository.InvestmentAccountRepository
	fetchRepo             *repository.FetchRepository
}

func NewDashboardService(planRepo *repository.PlanRepository, userRepo *repository.UserRepository, investmentAccountRepo *repository.InvestmentAccountRepository, fetchRepo *repository.FetchRepository) *DashboardService {
	return &DashboardService{
		planRepo:              planRepo,
		userRepo:              userRepo,
		investmentAccountRepo: investmentAccountRepo,
		fetchRepo:             fetchRepo,
	}
}

func (s *DashboardService) GetDashboardData(planID int, userID int) (*models.DashboardData, error) {
	data := &models.DashboardData{}

	// Get plan info
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, err
	}
	data.Plan = plan

	// Get user info
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	data.User = user

	// Get investment accounts
	accounts, err := s.investmentAccountRepo.GetByPlanID(planID)
	if err != nil {
		return nil, err
	}
	data.InvestmentAccounts = accounts

	// Get latest fetches (still get recent ones for display)
	data.LatestKuvera, _ = s.fetchRepo.GetLatestKuveraFetches(planID, 10)
	data.LatestStockal, _ = s.fetchRepo.GetLatestStockalFetches(planID, 10)

	// Get latest fetch per account for accurate portfolio calculation
	latestKuveraPerAccount, _ := s.fetchRepo.GetLatestKuveraFetchPerAccount(planID)
	latestStockalPerAccount, _ := s.fetchRepo.GetLatestStockalFetchPerAccount(planID)

	// Calculate total value using only latest fetch per account
	data.TotalValue = s.calculateTotalValue(latestKuveraPerAccount, latestStockalPerAccount)

	return data, nil
}

func (s *DashboardService) calculateTotalValue(kuveraFetches []models.KuveraFetch, stockalFetches []models.StockalFetch) float64 {
	total := 0.0

	// Sum Kuvera portfolio values (one per account)
	for _, fetch := range kuveraFetches {
		if fetch.PortfolioData != nil {
			var portfolio struct {
				Data struct {
					CurrentValue float64 `json:"current_value"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(*fetch.PortfolioData), &portfolio); err == nil {
				total += portfolio.Data.CurrentValue
			}
		}
	}

	// Sum Stockal portfolio values (one per account)
	for _, fetch := range stockalFetches {
		if fetch.PortfolioData != nil {
			var portfolio struct {
				Data struct {
					PortfolioSummary struct {
						TotalCurrentValue float64 `json:"totalCurrentValue"`
					} `json:"portfolioSummary"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(*fetch.PortfolioData), &portfolio); err == nil {
				total += portfolio.Data.PortfolioSummary.TotalCurrentValue
			}
		}
	}

	return total
}
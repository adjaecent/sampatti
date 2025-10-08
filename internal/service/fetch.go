package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/models"
	kuvera "github.com/adjaecent/unofficial-kuvera-api"
	stockal "github.com/adjaecent/unofficial-stockal-api"
)

type FetchService struct {
	investmentAccountRepo *repository.InvestmentAccountRepository
	fetchRepo             *repository.FetchRepository
	authService           *AuthService
}

func NewFetchService(investmentAccountRepo *repository.InvestmentAccountRepository, fetchRepo *repository.FetchRepository, authService *AuthService) *FetchService {
	return &FetchService{
		investmentAccountRepo: investmentAccountRepo,
		fetchRepo:             fetchRepo,
		authService:           authService,
	}
}

func (s *FetchService) FetchAllData(planID int) error {
	accounts, err := s.investmentAccountRepo.GetByPlanID(planID)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if account.IsKuvera() {
			go s.fetchKuveraData(account)
		} else if account.IsStockal() {
			go s.fetchStockalData(account)
		}
	}

	return nil
}

func (s *FetchService) fetchKuveraData(account models.InvestmentAccount) {
	client := kuvera.NewClient()
	ctx := context.Background()

	// Decrypt the password for API authentication
	password, err := s.authService.DecryptPassword(account.PasswordHash)
	if err != nil {
		log.Printf("Failed to decrypt password for account %d: %v", account.ID, err)
		s.fetchRepo.CreateKuveraFetch(account.ID, nil, nil, nil, false, "Failed to decrypt password")
		return
	}

	var portfolioData, holdingsData, goldPriceData *string
	var errorMessage string
	success := false

	// Login
	_, err = client.Login(ctx, account.Username, password)
	if err != nil {
		errorMessage = err.Error()
		s.fetchRepo.CreateKuveraFetch(account.ID, nil, nil, nil, false, errorMessage)
		return
	}

	// Get portfolio data
	portfolioResp, err := client.GetPortfolio(ctx)
	if err != nil {
		log.Printf("Failed to get Kuvera portfolio for account %d: %v", account.ID, err)
	} else {
		portfolioJSON, _ := json.Marshal(portfolioResp)
		portfolioStr := string(portfolioJSON)
		portfolioData = &portfolioStr
	}

	// Get holdings data
	holdingsResp, err := client.GetHoldings(ctx)
	if err != nil {
		log.Printf("Failed to get Kuvera holdings for account %d: %v", account.ID, err)
	} else {
		holdingsJSON, _ := json.Marshal(holdingsResp)
		holdingsStr := string(holdingsJSON)
		holdingsData = &holdingsStr
	}

	// Get gold price data
	goldResp, err := client.GetGoldPrice(ctx)
	if err != nil {
		log.Printf("Failed to get Kuvera gold price for account %d: %v", account.ID, err)
	} else {
		goldJSON, _ := json.Marshal(goldResp)
		goldStr := string(goldJSON)
		goldPriceData = &goldStr
	}

	success = true
	s.fetchRepo.CreateKuveraFetch(account.ID, portfolioData, holdingsData, goldPriceData, success, "")
}

func (s *FetchService) fetchStockalData(account models.InvestmentAccount) {
	client := stockal.NewClient()
	ctx := context.Background()

	// Decrypt the password for API authentication
	password, err := s.authService.DecryptPassword(account.PasswordHash)
	if err != nil {
		log.Printf("Failed to decrypt password for account %d: %v", account.ID, err)
		s.fetchRepo.CreateStockalFetch(account.ID, nil, nil, false, "Failed to decrypt password")
		return
	}

	var accountSummaryData, portfolioData *string
	var errorMessage string
	success := false

	// Login
	_, err = client.Login(ctx, account.Username, password)
	if err != nil {
		errorMessage = err.Error()
		s.fetchRepo.CreateStockalFetch(account.ID, nil, nil, false, errorMessage)
		return
	}

	// Get account summary
	summaryResp, err := client.GetAccountSummary(ctx)
	if err != nil {
		log.Printf("Failed to get Stockal account summary for account %d: %v", account.ID, err)
	} else {
		summaryJSON, _ := json.Marshal(summaryResp)
		summaryStr := string(summaryJSON)
		accountSummaryData = &summaryStr
	}

	// Get portfolio details
	portfolioResp, err := client.GetPortfolioDetail(ctx)
	if err != nil {
		log.Printf("Failed to get Stockal portfolio for account %d: %v", account.ID, err)
	} else {
		portfolioJSON, _ := json.Marshal(portfolioResp)
		portfolioStr := string(portfolioJSON)
		portfolioData = &portfolioStr
	}

	success = true
	s.fetchRepo.CreateStockalFetch(account.ID, accountSummaryData, portfolioData, success, "")
}
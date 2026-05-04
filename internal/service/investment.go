package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	kuveraAPI "github.com/adjaecent/unofficial-kuvera-api"
	stockalAPI "github.com/adjaecent/unofficial-stockal-api"
)

// InvestmentService handles fetching data from investment platforms.
type InvestmentService struct{}

// NewInvestmentService creates a new investment service.
func NewInvestmentService() *InvestmentService {
	return &InvestmentService{}
}

// FetchKuveraPortfolio fetches and formats portfolio data from Kuvera.
func (s *InvestmentService) FetchKuveraPortfolio(username, password string) (string, error) {
	client := kuveraAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Kuvera for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("kuvera login failed: %w", err)
	}

	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get portfolio: %w", err)
	}

	data, err := json.MarshalIndent(portfolio, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format portfolio: %w", err)
	}

	return string(data), nil
}

// FetchKuveraHoldings fetches and formats holdings data from Kuvera.
func (s *InvestmentService) FetchKuveraHoldings(username, password string) (string, error) {
	client := kuveraAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Kuvera for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("kuvera login failed: %w", err)
	}

	holdings, err := client.GetHoldings(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get holdings: %w", err)
	}

	data, err := json.MarshalIndent(holdings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format holdings: %w", err)
	}

	return string(data), nil
}

// FetchGoldPrice fetches current gold prices from Kuvera.
func (s *InvestmentService) FetchGoldPrice(username, password string) (string, error) {
	client := kuveraAPI.NewClient()
	ctx := context.Background()

	_, err := client.Login(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("kuvera login failed: %w", err)
	}

	goldPrice, err := client.GetGoldPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gold price: %w", err)
	}

	data, err := json.MarshalIndent(goldPrice, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format gold price: %w", err)
	}

	return string(data), nil
}

// FetchStockalAccount fetches account summary from Stockal.
func (s *InvestmentService) FetchStockalAccount(username, password string) (string, error) {
	client := stockalAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Stockal for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("stockal login failed: %w", err)
	}

	summary, err := client.GetAccountSummary(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get account summary: %w", err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format account summary: %w", err)
	}

	return string(data), nil
}

// FetchStockalHoldings fetches portfolio details from Stockal.
func (s *InvestmentService) FetchStockalHoldings(username, password string) (string, error) {
	client := stockalAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Stockal for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("stockal login failed: %w", err)
	}

	portfolio, err := client.GetPortfolioDetail(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get portfolio detail: %w", err)
	}

	data, err := json.MarshalIndent(portfolio, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format portfolio detail: %w", err)
	}

	return string(data), nil
}

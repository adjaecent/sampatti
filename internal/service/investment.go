package service

import (
	"context"
	"fmt"
	"log"

	"github.com/adjaecent/sampatti/internal/auth"
	kuveraAPI "github.com/adjaecent/unofficial-kuvera-api"
	stockalAPI "github.com/adjaecent/unofficial-stockal-api"
)

type InvestmentService struct {
}

func NewInvestmentService(authRequestMgr *auth.AuthRequestManager) *InvestmentService {
	return &InvestmentService{}
}

func (s *InvestmentService) FetchKuveraData(username, password string) (map[string]interface{}, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("kuvera credentials not provided")
	}

	client := kuveraAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Kuvera for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login to Kuvera: %w", err)
	}

	// Fetch portfolio data
	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kuvera portfolio: %w", err)
	}

	// Fetch holdings data
	holdings, err := client.GetHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kuvera holdings: %w", err)
	}

	// Fetch gold price data (if available)
	goldPrice, err := client.GetGoldPrice(ctx)
	if err != nil {
		log.Printf("Failed to get gold price from Kuvera: %v", err)
		goldPrice = nil
	}

	return map[string]interface{}{
		"username":   username,
		"portfolio":  portfolio,
		"holdings":   holdings,
		"gold_price": goldPrice,
	}, nil
}

func (s *InvestmentService) FetchStockalData(username, password string) (map[string]interface{}, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("stockal credentials not provided")
	}

	client := stockalAPI.NewClient()
	ctx := context.Background()

	log.Printf("Logging into Stockal for user: %s", username)
	_, err := client.Login(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login to Stockal: %w", err)
	}

	// Fetch account summary
	accountSummary, err := client.GetAccountSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stockal account summary: %w", err)
	}

	// Fetch portfolio details
	portfolioDetail, err := client.GetPortfolioDetail(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stockal portfolio detail: %w", err)
	}

	return map[string]interface{}{
		"username":         username,
		"account_summary":  accountSummary,
		"portfolio_detail": portfolioDetail,
	}, nil
}

func (s *InvestmentService) GetGoldSilverRates(kuveraUsername, kuveraPassword string) (map[string]interface{}, error) {
	// Try to get gold price from Kuvera first if credentials are available
	if kuveraUsername != "" && kuveraPassword != "" {
		client := kuveraAPI.NewClient()
		ctx := context.Background()

		_, err := client.Login(ctx, kuveraUsername, kuveraPassword)
		if err == nil {
			goldPrice, err := client.GetGoldPrice(ctx)
			if err == nil {
				return map[string]interface{}{
					"source":     "kuvera",
					"gold_price": goldPrice,
				}, nil
			}
		}
	}

	// TODO: Add other sources for precious metal rates
	// Could integrate with APIs like:
	// - metals-api.com
	// - precious-metals-api.com
	// - or scrape public sources

	return map[string]interface{}{
		"source": "unavailable",
		"error":  "No precious metals price source available",
	}, nil
}

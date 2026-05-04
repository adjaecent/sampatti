package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/adjaecent/sampatti/internal/oauth"
	"github.com/adjaecent/sampatti/internal/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server is the unified MCP server exposing tools for all platforms.
type Server struct {
	mcpServer         *server.MCPServer
	investmentService *service.InvestmentService
}

// NewServer creates a new unified MCP server with all tools registered.
func NewServer() *Server {
	mcpServer := server.NewMCPServer("sampatti", "2.0.0")
	investmentService := service.NewInvestmentService()

	s := &Server{
		mcpServer:         mcpServer,
		investmentService: investmentService,
	}

	s.registerTools()
	return s
}

// GetMCPServer returns the underlying mcp-go server for HTTP handler creation.
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

func (s *Server) registerTools() {
	// Kuvera tools
	s.mcpServer.AddTool(
		mcp.NewTool("get_kuvera_portfolio",
			mcp.WithDescription("Get Kuvera portfolio data including mutual fund holdings, current value, gains, and XIRR. Returns complete portfolio breakdown by asset class."),
		),
		s.handleGetKuveraPortfolio,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_kuvera_holdings",
			mcp.WithDescription("Get detailed Kuvera mutual fund holdings including individual funds, SIPs, folio numbers, and order history."),
		),
		s.handleGetKuveraHoldings,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_gold_price",
			mcp.WithDescription("Get current gold buy/sell prices and tax rates from Kuvera."),
		),
		s.handleGetGoldPrice,
	)

	// Stockal tools
	s.mcpServer.AddTool(
		mcp.NewTool("get_stockal_account",
			mcp.WithDescription("Get Stockal account summary including cash balances, trading restrictions, and portfolio value totals."),
		),
		s.handleGetStockalAccount,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_stockal_holdings",
			mcp.WithDescription("Get detailed Stockal US stock holdings including individual positions, current prices, units, investment amounts, and gain/loss."),
		),
		s.handleGetStockalHoldings,
	)
}

// --- Kuvera tool handlers ---

func (s *Server) handleGetKuveraPortfolio(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	creds := oauth.CredentialsFromContext(ctx)
	if creds == nil || creds.Kuvera == nil {
		return mcp.NewToolResultError("Kuvera credentials not configured. Please re-authorize with Kuvera credentials."), nil
	}

	data, err := s.investmentService.FetchKuveraPortfolio(creds.Kuvera.Username, creds.Kuvera.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Kuvera portfolio: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func (s *Server) handleGetKuveraHoldings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	creds := oauth.CredentialsFromContext(ctx)
	if creds == nil || creds.Kuvera == nil {
		return mcp.NewToolResultError("Kuvera credentials not configured. Please re-authorize with Kuvera credentials."), nil
	}

	data, err := s.investmentService.FetchKuveraHoldings(creds.Kuvera.Username, creds.Kuvera.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Kuvera holdings: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func (s *Server) handleGetGoldPrice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	creds := oauth.CredentialsFromContext(ctx)
	if creds == nil || creds.Kuvera == nil {
		return mcp.NewToolResultError("Kuvera credentials not configured. Please re-authorize with Kuvera credentials."), nil
	}

	data, err := s.investmentService.FetchGoldPrice(creds.Kuvera.Username, creds.Kuvera.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch gold price: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

// --- Stockal tool handlers ---

func (s *Server) handleGetStockalAccount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	creds := oauth.CredentialsFromContext(ctx)
	if creds == nil || creds.Stockal == nil {
		return mcp.NewToolResultError("Stockal credentials not configured. Please re-authorize with Stockal credentials."), nil
	}

	data, err := s.investmentService.FetchStockalAccount(creds.Stockal.Username, creds.Stockal.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Stockal account: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func (s *Server) handleGetStockalHoldings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	creds := oauth.CredentialsFromContext(ctx)
	if creds == nil || creds.Stockal == nil {
		return mcp.NewToolResultError("Stockal credentials not configured. Please re-authorize with Stockal credentials."), nil
	}

	data, err := s.investmentService.FetchStockalHoldings(creds.Stockal.Username, creds.Stockal.Password)
	if err != nil {
		log.Printf("Failed to fetch Stockal holdings: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Stockal holdings: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

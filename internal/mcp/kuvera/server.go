package kuvera

import (
	"context"
	"fmt"

	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/adjaecent/sampatti/internal/mcp/shared"
	"github.com/adjaecent/sampatti/internal/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// KuveraMCPServer represents the Kuvera-specific MCP server
type KuveraMCPServer struct {
	mcpServer         *server.MCPServer
	investmentService *service.InvestmentService
	authHandler       *shared.AuthToolHandler
}

// NewKuveraMCPServer creates a new Kuvera MCP server instance
func NewKuveraMCPServer(authRequestMgr *auth.AuthRequestManager, authPort string) *KuveraMCPServer {
	// Initialize services
	investmentService := service.NewInvestmentService(authRequestMgr)
	authHandler := shared.NewAuthToolHandler(authRequestMgr, authPort)

	// Create MCP server
	mcpServer := server.NewMCPServer("kuvera-mcp-server", "1.0.0")

	s := &KuveraMCPServer{
		mcpServer:         mcpServer,
		investmentService: investmentService,
		authHandler:       authHandler,
	}

	// Register tools
	s.registerTools()

	return s
}

// registerTools registers all Kuvera-specific tools with the MCP server
func (s *KuveraMCPServer) registerTools() {
	// 1. Authentication request tool
	s.mcpServer.AddTool(s.authHandler.GetAuthRequestTool(), s.authHandler.HandleAuthRequest)

	// 2. Get authentication token tool
	s.mcpServer.AddTool(s.authHandler.GetAuthTokenTool(), s.authHandler.HandleGetAuthToken)

	// 3. Kuvera data tool
	kuveraDataTool := mcp.NewTool("get_kuvera_data",
		mcp.WithDescription("STEP 3: Get detailed live data from Kuvera including portfolio and mutual fund holdings. Requires authentication token from get_auth_token."),
		mcp.WithString("token",
			mcp.Required(),
			mcp.Description("Authentication token from get_auth_token"),
		),
	)
	s.mcpServer.AddTool(kuveraDataTool, s.handleGetKuveraData)

	// 4. Gold/Silver rates tool
	goldSilverRatesTool := mcp.NewTool("get_gold_silver_rates",
		mcp.WithDescription("Get current gold and silver prices from Kuvera (requires authentication token)"),
		mcp.WithString("token",
			mcp.Required(),
			mcp.Description("Authentication token from get_auth_token"),
		),
	)
	s.mcpServer.AddTool(goldSilverRatesTool, s.handleGetGoldSilverRates)
}

// GetMCPServer returns the underlying MCP server
func (s *KuveraMCPServer) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}


// Tool handler implementations

func (s *KuveraMCPServer) handleGetKuveraData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	token, ok := args["token"].(string)
	if !ok {
		token = ""
	}

	credentials, err := s.authHandler.ValidateToken(token)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if credentials.Kuvera == nil {
		return mcp.NewToolResultError("Kuvera credentials not provided for this token"), nil
	}

	data, err := s.investmentService.FetchKuveraData(credentials.Kuvera.Username, credentials.Kuvera.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Kuvera data: %v", err)), nil
	}

	result := fmt.Sprintf("Kuvera Data for %s:\n", data["username"])
	result += "- Portfolio data fetched ✅\n"
	result += "- Holdings data fetched ✅\n"
	
	if goldPrice := data["gold_price"]; goldPrice != nil {
		result += "- Gold price data available ✅\n"
	}

	return mcp.NewToolResultText(result), nil
}

func (s *KuveraMCPServer) handleGetGoldSilverRates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	token, ok := args["token"].(string)
	if !ok {
		token = ""
	}

	credentials, err := s.authHandler.ValidateToken(token)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Use Kuvera credentials for precious metals data
	var kuveraUsername, kuveraPassword string
	if credentials.Kuvera != nil {
		kuveraUsername = credentials.Kuvera.Username
		kuveraPassword = credentials.Kuvera.Password
	}

	rates, err := s.investmentService.GetGoldSilverRates(kuveraUsername, kuveraPassword)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get precious metal rates: %v", err)), nil
	}

	source := rates["source"].(string)
	result := fmt.Sprintf("Precious Metal Rates (Source: %s):\n", source)

	if source == "unavailable" {
		result += "❌ " + rates["error"].(string)
	} else {
		if goldPrice := rates["gold_price"]; goldPrice != nil {
			result += "✅ Gold price data available\n"
		}
		// TODO: Add silver price when available
	}

	return mcp.NewToolResultText(result), nil
}
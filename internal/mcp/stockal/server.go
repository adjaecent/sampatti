package stockal

import (
	"context"
	"fmt"

	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/adjaecent/sampatti/internal/mcp/shared"
	"github.com/adjaecent/sampatti/internal/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StockalMCPServer represents the Stockal-specific MCP server
type StockalMCPServer struct {
	mcpServer         *server.MCPServer
	investmentService *service.InvestmentService
	authHandler       *shared.AuthToolHandler
}

// NewStockalMCPServer creates a new Stockal MCP server instance
func NewStockalMCPServer(authRequestMgr *auth.AuthRequestManager, authPort string) *StockalMCPServer {
	investmentService := service.NewInvestmentService(authRequestMgr)
	authHandler := shared.NewAuthToolHandler(authRequestMgr, authPort)
	mcpServer := server.NewMCPServer("stockal-mcp-server", "1.0.0")

	s := &StockalMCPServer{
		mcpServer:         mcpServer,
		investmentService: investmentService,
		authHandler:       authHandler,
	}

	s.registerTools()

	return s
}

// registerTools registers all Stockal-specific tools with the MCP server
func (s *StockalMCPServer) registerTools() {
	s.mcpServer.AddTool(s.authHandler.GetAuthRequestTool(), s.authHandler.HandleAuthRequest)
	s.mcpServer.AddTool(s.authHandler.GetAuthTokenTool(), s.authHandler.HandleGetAuthToken)
	stockalDataTool := mcp.NewTool("get_stockal_data",
		mcp.WithDescription("Get detailed live data from Stockal (a.k.a Global Investing) including portfolio and US stock positions (requires authentication token)"),
		mcp.WithString("token",
			mcp.Required(),
			mcp.Description("Authentication token from get_auth_token"),
		),
	)
	s.mcpServer.AddTool(stockalDataTool, s.handleGetStockalData)
}

// GetMCPServer returns the underlying MCP server
func (s *StockalMCPServer) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

// Tool handler implementations

func (s *StockalMCPServer) handleGetStockalData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	token, ok := args["token"].(string)
	if !ok {
		token = ""
	}

	credentials, err := s.authHandler.ValidateToken(token)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if credentials.Stockal == nil {
		return mcp.NewToolResultError("Stockal credentials not provided for this token"), nil
	}

	data, err := s.investmentService.FetchStockalData(credentials.Stockal.Username, credentials.Stockal.Password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch Stockal data: %v", err)), nil
	}

	result := fmt.Sprintf("Stockal Data for %s:\n", data["username"])
	result += "- Account summary fetched ✅\n"
	result += "- Portfolio detail fetched ✅\n"

	return mcp.NewToolResultText(result), nil
}

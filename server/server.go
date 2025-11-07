package server

import (
	"context"
	"log"

	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/adjaecent/sampatti/internal/service"
	"github.com/mark3labs/mcp-go/server"
)

// SampattiServer represents the main MCP server for investment data
type SampattiServer struct {
	mcpServer         *server.MCPServer
	investmentService *service.InvestmentService
	authRequestMgr    *auth.AuthRequestManager
	authPort          string
}

// NewSampattiServer creates a new Sampatti MCP server instance
func NewSampattiServer(authPort string) *SampattiServer {
	// Create auth request manager
	authRequestMgr := auth.NewAuthRequestManager()

	// Initialize services
	investmentService := service.NewInvestmentService(authRequestMgr)

	// Create MCP server
	mcpServer := server.NewMCPServer("sampatti-investment-server", "1.0.0")

	s := &SampattiServer{
		mcpServer:         mcpServer,
		investmentService: investmentService,
		authRequestMgr:    authRequestMgr,
		authPort:          authPort,
	}

	// Register tools
	s.registerTools()

	return s
}

// Run starts the MCP server
func (s *SampattiServer) Run(ctx context.Context) error {
	log.Println("Starting Sampatti MCP server...")
	return server.ServeStdio(ctx, s.mcpServer)
}

// GetAuthRequestManager returns the auth request manager for the web server
func (s *SampattiServer) GetAuthRequestManager() *auth.AuthRequestManager {
	return s.authRequestMgr
}

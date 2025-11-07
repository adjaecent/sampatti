package mcp

import (
	"context"
	"log"
	"net/http"

	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/adjaecent/sampatti/internal/mcp/kuvera"
	"github.com/adjaecent/sampatti/internal/mcp/stockal"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer manages the HTTP MCP server with streaming endpoints
type MCPServer struct {
	authRequestMgr *auth.AuthRequestManager
	authPort       string

	// Individual servers
	kuveraServer  *kuvera.KuveraMCPServer
	stockalServer *stockal.StockalMCPServer
}

// NewMCPServer creates a new MCP server
func NewMCPServer(authRequestMgr *auth.AuthRequestManager, authPort string) *MCPServer {
	return &MCPServer{
		authRequestMgr: authRequestMgr,
		authPort:       authPort,
	}
}

// Run starts the HTTP MCP server with streaming endpoints
func (m *MCPServer) Run(ctx context.Context, port string) error {
	// Initialize both servers
	m.kuveraServer = kuvera.NewKuveraMCPServer(m.authRequestMgr, m.authPort)
	m.stockalServer = stockal.NewStockalMCPServer(m.authRequestMgr, m.authPort)

	// Create HTTP handlers
	mux := http.NewServeMux()

	// Create streaming HTTP handlers
	kuveraHTTPServer := server.NewStreamableHTTPServer(m.kuveraServer.GetMCPServer())
	stockalHTTPServer := server.NewStreamableHTTPServer(m.stockalServer.GetMCPServer())

	// Route /mcp-kuvera to Kuvera server
	mux.Handle("/mcp-kuvera", kuveraHTTPServer)

	// Route /mcp-stockal to Stockal server
	mux.Handle("/mcp-stockal", stockalHTTPServer)

	log.Printf("Starting streaming HTTP MCP server on port %s", port)
	log.Printf("Kuvera endpoint: http://localhost:%s/mcp-kuvera", port)
	log.Printf("Stockal endpoint: http://localhost:%s/mcp-stockal", port)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	return httpServer.Shutdown(context.Background())
}

// GetAuthRequestManager returns the auth request manager
func (m *MCPServer) GetAuthRequestManager() *auth.AuthRequestManager {
	return m.authRequestMgr
}

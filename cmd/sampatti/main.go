package main

import (
	"context"
	"fmt"
	"github.com/adjaecent/sampatti/config"
	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/adjaecent/sampatti/internal/mcp"
	"github.com/adjaecent/sampatti/internal/web"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Fprintf(os.Stderr, "Sampatti starting...\n")
	config.Load()

	authRequestMgr := auth.NewAuthRequestManager()
	mcpServer := mcp.NewMCPServer(authRequestMgr, config.C.MCPPort)
	authHTTPServer := web.NewAuthServer(authRequestMgr, config.C.AuthHTTPPort)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down servers...")
		cancel()
	}()

	// Start auth server in background
	go func() {
		authHTTPServer.Start()
	}()

	if err := mcpServer.Run(ctx, config.C.MCPPort); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server failed: %v\n", err)
		log.Printf("MCP server failed: %v", err)
	}

	fmt.Fprintf(os.Stderr, "MCP server stopped\n")
	log.Println("MCP server stopped")
}

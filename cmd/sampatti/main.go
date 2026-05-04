package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adjaecent/sampatti/config"
	"github.com/adjaecent/sampatti/internal/mcp"
	"github.com/adjaecent/sampatti/internal/oauth"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	fmt.Fprintf(os.Stderr, "Sampatti starting...\n")
	config.Load()

	// Initialize OAuth infrastructure
	store := oauth.NewMemoryStore()
	provider := oauth.NewOAuthProvider(store, config.C.Secret)
	oauthHandlers := oauth.NewHandlers(provider, store, config.C.BaseURL)

	// Initialize MCP server
	mcpServer := mcp.NewServer()
	mcpHTTPServer := server.NewStreamableHTTPServer(mcpServer.GetMCPServer(), server.WithStateLess(true))

	// Set up HTTP routes
	mux := http.NewServeMux()

	// OAuth endpoints
	mux.HandleFunc("/.well-known/oauth-authorization-server", oauthHandlers.HandleMetadata)
	mux.HandleFunc("/.well-known/oauth-protected-resource", oauthHandlers.HandleProtectedResourceMetadata)
	mux.HandleFunc("/register", oauthHandlers.HandleRegister)
	mux.HandleFunc("/authorize", oauthHandlers.HandleAuthorize)
	mux.HandleFunc("/token", oauthHandlers.HandleToken)

	// MCP endpoint
	if config.C.DevMode {
		// Dev mode: inject credentials from env, no OAuth required
		log.Println("DEV MODE: OAuth disabled, using credentials from environment")
		mux.Handle("/mcp", oauth.DevMiddleware(config.C.DevCredentials, mcpHTTPServer))
	} else {
		// Production: protected by OAuth middleware
		mux.Handle("/mcp", oauth.AuthMiddleware(provider, store, config.C.BaseURL, mcpHTTPServer))
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down...")
		cancel()
	}()

	// Wrap with request logging
	logged := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s (Origin: %s)", r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Origin"))
		mux.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: logged,
	}

	log.Printf("Sampatti MCP server starting on port %s", config.C.Port)
	log.Printf("MCP endpoint: %s/mcp", config.C.BaseURL)
	log.Printf("OAuth metadata: %s/.well-known/oauth-authorization-server", config.C.BaseURL)

	// Start server
	go func() {
		var err error
		// Try TLS first, fall back to plain HTTP
		if _, certErr := os.Stat("certs/cert.pem"); certErr == nil {
			log.Println("Starting with TLS (certs/cert.pem, certs/key.pem)")
			err = httpServer.ListenAndServeTLS("certs/cert.pem", "certs/key.pem")
		} else {
			log.Println("Starting without TLS (no certs found)")
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	httpServer.Shutdown(context.Background())
	log.Println("Sampatti stopped")
}

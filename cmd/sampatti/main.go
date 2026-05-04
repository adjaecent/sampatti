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
	"github.com/adjaecent/sampatti/static"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	fmt.Fprintf(os.Stderr, "Sampatti starting...\n")
	config.Load()

	// Initialize OAuth infrastructure
	store := oauth.NewMemoryStore(config.C.Secret)
	provider := oauth.NewOAuthProvider(store, config.C.Secret)
	oauthHandlers := oauth.NewHandlers(provider, store, config.C.BaseURL)

	// Initialize MCP server
	mcpServer := mcp.NewServer()
	mcpHTTPServer := server.NewStreamableHTTPServer(mcpServer.GetMCPServer(), server.WithStateLess(true))

	// Set up HTTP routes
	mux := http.NewServeMux()

	// Static assets
	mux.HandleFunc("/favicon.ico", serveFavicon)
	mux.HandleFunc("/favicon.png", serveFavicon)

	// OAuth endpoints
	mux.HandleFunc("/.well-known/oauth-authorization-server", oauthHandlers.HandleMetadata)
	mux.HandleFunc("/.well-known/oauth-protected-resource", oauthHandlers.HandleProtectedResourceMetadata)
	mux.HandleFunc("/register", oauthHandlers.HandleRegister)
	mux.HandleFunc("/authorize", oauthHandlers.HandleAuthorize)
	mux.HandleFunc("/token", oauthHandlers.HandleToken)

	// MCP endpoint
	if config.C.DevMode {
		log.Println("DEV MODE: OAuth disabled, using credentials from environment")
		mux.Handle("/mcp", oauth.DevMiddleware(config.C.DevCredentials, mcpHTTPServer))
	} else {
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
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.URL.RawQuery)
		mux.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: logged,
	}

	log.Printf("Sampatti MCP server starting on port %s", config.C.Port)
	log.Printf("MCP endpoint: %s/mcp", config.C.BaseURL)

	// Plain HTTP — TLS termination handled by reverse proxy (Caddy)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	httpServer.Shutdown(context.Background())
	log.Println("Sampatti stopped")
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(static.FaviconPNG)
}

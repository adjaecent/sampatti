package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adjaecent/sampatti/config"
	"github.com/adjaecent/sampatti/db"
	"github.com/adjaecent/sampatti/internal/scheduler"
	"github.com/adjaecent/sampatti/web"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Start scheduler
	sched := scheduler.New(database)
	go sched.Start()

	// Setup web server
	r := mux.NewRouter()
	webHandler := web.New(database, cfg)
	webHandler.SetupRoutes(r)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	sched.Stop()
}
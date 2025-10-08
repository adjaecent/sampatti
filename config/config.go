package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabasePath     string
	GoogleClientID   string
	GoogleSecret     string
	SessionSecret    string
	BaseURL          string
	Port             string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	cfg := &Config{
		DatabasePath:   getEnv("DATABASE_PATH", "./sampatti.db"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		SessionSecret:  getEnv("SESSION_SECRET", "your-session-secret-change-me"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		Port:           getEnv("PORT", "8080"),
	}

	// Validate required OAuth credentials
	if cfg.GoogleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}
	if cfg.GoogleSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
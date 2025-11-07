package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MCPPort      string
	AuthHTTPPort string
}

var C *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	C = &Config{
		MCPPort:      getEnv("MCP_PORT", "8081"),
		AuthHTTPPort: getEnv("AUTH_HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

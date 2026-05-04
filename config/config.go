package config

import (
	"encoding/hex"
	"log"
	"os"
	"strings"

	"github.com/adjaecent/sampatti/internal/oauth"
	"github.com/joho/godotenv"
)

type Config struct {
	Port           string // Single port for everything (OAuth + MCP)
	BaseURL        string // Public base URL (e.g. https://sampatti.example.com)
	Secret         []byte // HMAC secret for token signing
	DevMode        bool   // Skip OAuth, use env credentials directly
	DevCredentials *oauth.Credentials
}

var C *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	var secret []byte
	if s := os.Getenv("OAUTH_SECRET"); s != "" {
		var err error
		secret, err = hex.DecodeString(s)
		if err != nil {
			log.Fatalf("OAUTH_SECRET must be a hex-encoded string: %v", err)
		}
	}

	devMode := strings.ToLower(os.Getenv("DEV_MODE")) == "true"

	var devCreds *oauth.Credentials
	if devMode {
		devCreds = &oauth.Credentials{}
		if u, p := os.Getenv("KUVERA_USERNAME"), os.Getenv("KUVERA_PASSWORD"); u != "" && p != "" {
			devCreds.Kuvera = &oauth.KuveraCredentials{Username: u, Password: p}
		}
		if u, p := os.Getenv("STOCKAL_USERNAME"), os.Getenv("STOCKAL_PASSWORD"); u != "" && p != "" {
			devCreds.Stockal = &oauth.StockalCredentials{Username: u, Password: p}
		}
	}

	C = &Config{
		Port:           getEnv("PORT", "8081"),
		BaseURL:        getEnv("BASE_URL", "https://localhost:8081"),
		Secret:         secret,
		DevMode:        devMode,
		DevCredentials: devCreds,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

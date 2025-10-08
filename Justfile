# Development commands for Sampatti

# Install dependencies
install:
    go mod tidy

# Run the application
run:
    go run cmd/sampatti/main.go

# Build the application
build:
    go build -o bin/sampatti cmd/sampatti/main.go

# Run tests
test:
    go test ./...

# Format code
fmt:
    go fmt ./...

# Lint code
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -rf bin/

# Setup development environment
setup:
    cp .env.example .env
    @echo "Please update .env with your Google OAuth credentials"

# Database operations
db-reset:
    rm -f sampatti.db
    @echo "Database reset complete"

# Run with live reload (installs air if needed)
dev:
    @which air > /dev/null || go install github.com/cosmtrek/air@latest
    air

# Show help
help:
    @just --list
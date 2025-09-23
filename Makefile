.PHONY: run gen db-up db-down migrate-up migrate-down test clean build docker-build docker-run

# Default target
all: gen db-up migrate-up run

# Run the application
run:
	go run ./cmd/devhive-api

# Generate sqlc code
gen:
	sqlc generate

# Database operations
db-up:
	docker compose up -d postgres

db-down:
	docker compose down postgres

# Migration operations
migrate-up:
	go run ./cmd/devhive-api migrate

migrate-down:
	go run ./cmd/devhive-api migrate down

# Test
test:
	go test ./...

# Clean
clean:
	go clean
	docker compose down -v

# Build
build:
	go build -o bin/devhive-api ./cmd/devhive-api

# Docker operations
docker-build:
	docker compose build

docker-run:
	docker compose up

# Development setup
dev-setup: gen db-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Development environment ready!"

# Production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/devhive-api ./cmd/devhive-api

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Security scan
security:
	gosec ./...

# Help
help:
	@echo "Available targets:"
	@echo "  run          - Run the application"
	@echo "  gen          - Generate sqlc code"
	@echo "  db-up        - Start PostgreSQL database"
	@echo "  db-down      - Stop PostgreSQL database"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts and containers"
	@echo "  build        - Build the application"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  dev-setup    - Setup development environment"
	@echo "  prod-build   - Build for production"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  security     - Run security scan"
	@echo "  help         - Show this help message"
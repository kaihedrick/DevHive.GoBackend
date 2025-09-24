.PHONY: help build run test clean deps fmt fmt-check lint security docker-build docker-run dev-setup prod-build gen db-up db-down ci vet test-coverage

# Variables
BINARY_NAME=devhive-api
BINARY_PATH=bin/$(BINARY_NAME)
DOCKER_IMAGE=devhive-backend
LDFLAGS=-ldflags="-w -s"

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "DevHive Go Backend - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev-setup    - Setup development environment"
	@echo "  run          - Run the application locally"
	@echo "  gen          - Generate sqlc code"
	@echo "  fmt          - Format code"
	@echo "  fmt-check    - Check code formatting"
	@echo "  lint         - Lint code"
	@echo "  vet          - Run go vet"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  ci           - Run CI pipeline locally"
	@echo ""
	@echo "Database:"
	@echo "  db-up        - Start PostgreSQL database"
	@echo "  db-down      - Stop PostgreSQL database"
	@echo "  db-reset     - Reset database (WARNING: destroys data)"
	@echo ""
	@echo "Build & Deploy:"
	@echo "  build        - Build the application"
	@echo "  prod-build    - Build for production"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo ""
	@echo "Utilities:"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  security     - Run security scan"
	@echo "  clean        - Clean build artifacts and containers"
	@echo "  help         - Show this help message"

## dev-setup: Setup development environment
dev-setup: deps gen gen-grpc db-up
	@echo "🚀 Setting up development environment..."
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "✅ Development environment ready!"
	@echo "Run 'make run' to start the application"

## run: Run the application
run:
	@echo "🚀 Starting DevHive API server..."
	go run ./cmd/devhive-api

## gen: Generate sqlc code
gen:
	@echo "🔧 Generating sqlc code..."
	sqlc generate

## gen-grpc: Generate gRPC code
gen-grpc:
	@echo "🔧 Generating gRPC code..."
	@if exist scripts\generate-grpc.bat (scripts\generate-grpc.bat) else (scripts/generate-grpc.sh)

## db-up: Start PostgreSQL database
db-up:
	@echo "🐘 Starting PostgreSQL database..."
	docker compose up -d postgres
	@echo "✅ Database started"

## db-down: Stop PostgreSQL database
db-down:
	@echo "🛑 Stopping PostgreSQL database..."
	docker compose down postgres
	@echo "✅ Database stopped"

## db-reset: Reset database (WARNING: destroys data)
db-reset: db-down
	@echo "⚠️  Resetting database - all data will be lost!"
	docker compose down -v
	docker compose up -d postgres
	@echo "✅ Database reset complete"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	go test -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

## build: Build the application
build: gen
	@echo "🔨 Building application..."
	mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY_PATH) ./cmd/devhive-api
	@echo "✅ Build complete: $(BINARY_PATH)"

## prod-build: Build for production
prod-build: gen
	@echo "🔨 Building for production..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/devhive-api
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/devhive-api
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/devhive-api
	@echo "✅ Production builds complete"

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE)"

## docker-run: Run with Docker Compose
docker-run:
	@echo "🐳 Starting with Docker Compose..."
	docker compose up --build

## deps: Download and tidy dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	go mod verify
	@echo "✅ Dependencies updated"

## fmt: Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

## fmt-check: Check code formatting
fmt-check:
	@echo "🔍 Checking code formatting..."
	@gofmt -s -l . | findstr /R "." >nul && (echo "❌ Code is not properly formatted. Run 'gofmt -s -w .'" && gofmt -s -l . && exit 1) || echo "✅ Code formatting is correct"


## security: Run security scan
security:
	@echo "🔒 Running security scan..."
	gosec ./...
	@echo "✅ Security scan complete"

## clean: Clean build artifacts and containers
clean:
	@echo "🧹 Cleaning up..."
	go clean
	docker compose down -v
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"

## install-tools: Install development tools
install-tools:
	@echo "🛠️  Installing development tools..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development tools installed"

## check: Run all checks (format, lint, test, security)
check: fmt lint test security
	@echo "✅ All checks passed!"

## ci: Run CI pipeline locally
ci: vet lint test-coverage
	@echo "✅ CI pipeline completed successfully!"

## vet: Run go vet
vet:
	@echo "🔍 Running go vet..."
	go vet ./...
	@echo "✅ Go vet complete"

## lint: Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run --timeout=5m --build-tags=postgres,prod
	@echo "✅ Linting complete"
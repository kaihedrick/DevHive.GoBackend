# ---- Config ----
PORT ?= 8080

# ---- Targets ----
.PHONY: run build clean deps test test-coverage fmt lint dev

# OpenAPI codegen not configured - skipping
oapi:
	@echo "⚠️ OpenAPI codegen not configured - skipping"
	@echo "✅ Continuing with build process"

# CI-friendly: skip OpenAPI check
check-oapi:
	@echo "⚠️ OpenAPI check skipped - not configured"
	@echo "✅ Continuing with build process"

# Run the server locally
run:
	PORT=$(PORT) go run ./cmd/server

# Build the application
build:
	go build -o bin/devhive ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Format code
fmt:
	go fmt ./...

# Lint code (auto-installs golangci-lint if missing)
lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
	  echo "Installing golangci-lint..."; \
	  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3; \
	fi
	golangci-lint run

# Generate and run
dev: oapi run

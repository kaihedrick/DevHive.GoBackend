.PHONY: oapi run build clean

# Generate Go code from OpenAPI specification
oapi:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

# Run the server locally
run:
	PORT=8080 go run ./cmd/server

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

# Lint code
lint:
	golangci-lint run

# Generate and run
dev: oapi run

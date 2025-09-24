# DevHive Go Backend - Setup Guide

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** - [Download](https://golang.org/dl/)
- **Docker & Docker Compose** - [Download](https://docs.docker.com/get-docker/)
- **Git** - [Download](https://git-scm.com/downloads)

### 1. Clone and Setup
```bash
# Clone the repository
git clone <your-repo-url>
cd DevHive.GoBackend

# Install development tools
make install-tools

# Setup development environment
make dev-setup
```

### 2. Environment Configuration
```bash
# Copy environment template
cp env.example .env

# Edit .env with your configuration
nano .env
```

**Required Environment Variables:**
```env
# Database
DATABASE_URL=postgres://devhive:devhive@localhost:5432/devhive?sslmode=disable

# JWT Configuration
JWT_SIGNING_KEY=your-super-secret-jwt-key-change-in-production
JWT_ISSUER=https://api.devhive.it.com
JWT_AUDIENCE=devhive-clients

# Application
PORT=8080
GIN_MODE=debug
```

### 3. Start the Application
```bash
# Start the application
make run

# Or run with Docker
make docker-run
```

The API will be available at `http://localhost:8080`

## 🛠️ Development Commands

### Essential Commands
```bash
# Show all available commands
make help

# Setup development environment
make dev-setup

# Run the application
make run

# Generate sqlc code
make gen

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Run security scan
make security

# Run all checks
make check
```

### Database Commands
```bash
# Start database
make db-up

# Stop database
make db-down

# Reset database (WARNING: destroys data)
make db-reset
```

### Build Commands
```bash
# Build for development
make build

# Build for production (multi-platform)
make prod-build

# Build Docker image
make docker-build

# Run with Docker Compose
make docker-run
```

## 📁 Project Structure

```
DevHive.GoBackend/
├── cmd/devhive-api/          # Main application
├── internal/                 # Internal packages
│   ├── config/               # Configuration
│   ├── http/                 # HTTP handlers and middleware
│   │   ├── handlers/         # Request handlers
│   │   ├── middleware/       # HTTP middleware
│   │   └── router/           # Route configuration
│   ├── repo/                 # Database repository (sqlc generated)
│   └── ws/                   # WebSocket handlers
├── db/                       # Database utilities
├── config/                   # Configuration files
├── storage/                   # File storage utilities
├── .github/workflows/        # CI/CD pipelines
├── docker-compose.yml        # Local development
├── Dockerfile               # Production container
├── Makefile                 # Development commands
└── README.md               # This file
```

## 🔧 Development Workflow

### 1. Daily Development
```bash
# Start your day
make dev-setup

# Make changes to code
# ...

# Run tests
make test

# Format and lint
make fmt lint

# Run the application
make run
```

### 2. Before Committing
```bash
# Run all checks
make check

# This runs: fmt, lint, test, security
```

### 3. Database Changes
```bash
# If you modify SQL queries
make gen

# If you need to reset database
make db-reset
```

## 🐳 Docker Development

### Local Development with Docker
```bash
# Start all services (database + API)
make docker-run

# Build and run specific service
docker compose up --build api

# View logs
docker compose logs -f api

# Stop all services
docker compose down
```

### Database Access
```bash
# Connect to database
docker compose exec postgres psql -U devhive -d devhive

# View database logs
docker compose logs postgres
```

## 🧪 Testing

### Run Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/http/handlers/...

# Run tests with verbose output
go test -v ./...
```

### Test Coverage
```bash
# Generate coverage report
make test-coverage

# Open coverage report in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

## 🔒 Security

### Security Scanning
```bash
# Run security scan
make security

# Run Trivy vulnerability scan
trivy fs .

# Check for dependency vulnerabilities
go list -json -m all | nancy sleuth
```

## 🚀 Deployment

### Local Production Build
```bash
# Build production binaries
make prod-build

# Build Docker image
make docker-build

# Test production build
./bin/devhive-api-linux-amd64
```

### Fly.io Deployment
```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Login to Fly.io
fly auth login

# Deploy to Fly.io
fly deploy
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Database Connection Failed
```bash
# Check if database is running
make db-up

# Check database logs
docker compose logs postgres

# Reset database
make db-reset
```

#### 2. Build Failures
```bash
# Clean and rebuild
make clean
make deps
make build
```

#### 3. Import Errors
```bash
# Regenerate sqlc code
make gen

# Download dependencies
make deps
```

#### 4. Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port
PORT=8081 make run
```

### Debug Commands
```bash
# Run with debug logging
GIN_MODE=debug make run

# Run with race detection
go run -race ./cmd/devhive-api

# Profile the application
go run ./cmd/devhive-api &
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 📚 Additional Resources

- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [Chi Router](https://github.com/go-chi/chi)
- [Docker Compose](https://docs.docker.com/compose/)
- [Fly.io Documentation](https://fly.io/docs/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make check`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 📞 Support

If you encounter any issues:

1. Check the troubleshooting section above
2. Search existing issues in the repository
3. Create a new issue with detailed information
4. Join our community discussions

---

**Happy Coding! 🚀**

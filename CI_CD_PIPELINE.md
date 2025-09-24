# DevHive CI/CD Pipeline Documentation

## 🚀 Overview

This document describes the optimized CI/CD pipeline for the DevHive Go Backend project. The pipeline has been redesigned for efficiency, security, and maintainability.

## 📋 Pipeline Structure

### 1. CI Pipeline (`.github/workflows/ci.yml`)
**Triggers:** Push/PR to `main` or `develop` branches

**Jobs:**
- **Quality Check**: Code quality, formatting, linting, testing, security scanning
- **Build**: Multi-platform builds with artifact storage

**Features:**
- Go 1.23 with module caching
- sqlc code generation
- Comprehensive security scanning with Trivy
- Test coverage reporting (minimum 50%)
- Multi-platform builds (Linux, Windows, macOS)

### 2. Deploy Pipeline (`.github/workflows/deploy.yml`)
**Triggers:** Push to `main` branch, manual dispatch

**Jobs:**
- **Deploy**: Production deployment to Fly.io

**Features:**
- Docker image building and pushing to GHCR
- Fly.io deployment with verification
- Production environment protection
- Deployment status notifications

### 3. Release Pipeline (`.github/workflows/release.yml`)
**Triggers:** Git tags starting with `v*`, manual dispatch

**Jobs:**
- **Release**: Create GitHub release with binaries
- **Deploy Release**: Deploy tagged releases to production

**Features:**
- Multi-platform binary releases
- Semantic versioning support
- Production deployment of releases
- Release notes generation

## 🔐 Environment Secrets

### Required Secrets
The following secrets must be configured in your GitHub repository:

#### GitHub Secrets
- `FLY_API_TOKEN` - Fly.io API token for deployment
- `GITHUB_TOKEN` - Automatically provided for GHCR access

#### Fly.io Secrets (Already Configured)
Based on your existing configuration:
- `ConnectionStringsDbConnection` - Database connection string
- `FIREBASE_JSON_BASE64` - Firebase service account (base64 encoded)
- `FIREBASE_PROJECT_ID` - Firebase project ID
- `JwtKey` - JWT signing key
- `JwtIssuer` - JWT issuer
- `JwtAudience` - JWT audience
- `MailgunApiKey` - Mailgun API key
- `MailgunDomain` - Mailgun domain
- `Mailgun__SenderEmail` - Mailgun sender email

## 🛠️ Build Process

### Docker Build Optimization
The Dockerfile has been optimized with:
- Multi-stage build for smaller final image
- sqlc code generation during build
- Binary optimization with `-ldflags="-w -s"`
- Non-root user for security
- Health checks for container orchestration

### Build Artifacts
- **CI Pipeline**: Builds for Linux, Windows, macOS
- **Release Pipeline**: Creates GitHub releases with all platform binaries
- **Docker**: Optimized container images pushed to GHCR

## 🔍 Security Features

### Security Scanning
- **Trivy**: Vulnerability scanning for dependencies and OS packages
- **SARIF**: Security results uploaded to GitHub Security tab
- **Configuration**: Custom Trivy config with false positive filtering

### Security Headers
Fly.io configuration includes security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`

## 📊 Quality Gates

### Code Quality
- **Formatting**: `gofmt` formatting check
- **Linting**: `golangci-lint` with comprehensive rules
- **Vet**: `go vet` for common issues
- **Coverage**: Minimum 50% test coverage requirement

### Testing
- **Unit Tests**: `go test` with race detection
- **Coverage**: Coverage reporting and threshold enforcement
- **Build Verification**: Application builds successfully

## 🚀 Deployment Strategy

### Development Flow
1. **Feature Branch** → **Pull Request** → **CI Pipeline**
2. **Merge to Main** → **Deploy Pipeline** → **Production**

### Release Flow
1. **Create Tag** → **Release Pipeline** → **GitHub Release**
2. **Release Deployment** → **Production with Tagged Version**

### Rollback Strategy
- **Immediate**: Use Fly.io rollback commands
- **Versioned**: Deploy specific tagged releases
- **Database**: Migrations are reversible

## 🔧 Local Development

### Prerequisites
```bash
# Install required tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Development Commands
```bash
# Generate sqlc code
make gen

# Run tests
make test

# Build application
make build

# Run with Docker
make docker-run
```

## 📈 Monitoring & Observability

### Health Checks
- **Endpoint**: `/health` - Application health status
- **Metrics**: `/metrics` - Prometheus-compatible metrics
- **Fly.io**: Built-in health checks every 15 seconds

### Logging
- **Structured Logging**: JSON format in production
- **Log Levels**: Configurable via environment variables
- **Fly.io Logs**: Accessible via `fly logs` command

## 🚨 Troubleshooting

### Common Issues

#### 1. Build Failures
```bash
# Check Go version compatibility
go version

# Verify dependencies
go mod verify

# Clean and rebuild
go clean -cache
go mod download
```

#### 2. Deployment Issues
```bash
# Check Fly.io status
fly status

# View deployment logs
fly logs

# SSH into container
fly ssh console
```

#### 3. Security Scan Failures
- Check `.trivyignore` for false positives
- Update dependencies to fix vulnerabilities
- Review security scan results in GitHub Security tab

### Debug Commands
```bash
# Local testing
go run ./cmd/devhive-api

# Docker testing
docker build -t devhive-backend .
docker run -p 8080:8080 devhive-backend

# Fly.io debugging
fly logs --app devhive-go-backend
fly ssh console --app devhive-go-backend
```

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Fly.io Documentation](https://fly.io/docs/)
- [Trivy Security Scanner](https://trivy.dev/)
- [sqlc Documentation](https://docs.sqlc.dev/)

## 🔄 Pipeline Maintenance

### Regular Tasks
- **Monthly**: Update Go version and dependencies
- **Quarterly**: Review and update security configurations
- **As Needed**: Update Trivy ignore list for false positives

### Monitoring
- **Build Times**: Monitor and optimize slow builds
- **Test Coverage**: Maintain and improve test coverage
- **Security**: Regular security scan reviews
- **Performance**: Monitor deployment and application performance

# GitHub Actions Pipeline Testing

This directory contains tools and configurations for testing GitHub Actions workflows locally using [Act](https://github.com/nektos/act).

## Prerequisites

1. **Install Act**: 
   - **Windows**: See [INSTALL_ACT.md](INSTALL_ACT.md) for multiple installation methods
     - Quick: Download from [releases](https://github.com/nektos/act/releases/latest)
     - Or: `choco install act-cli` (requires admin)
   - **macOS**: `brew install act`
   - **Linux**: See [Act installation guide](https://github.com/nektos/act#installation)

2. **Docker**: Act requires Docker Desktop or Docker Engine to be running

3. **Secrets**: Create `.github/tests/.secrets` file (see `.secrets.example`)

## Quick Start

```bash
# Test the deploy workflow (build only, no actual deployment)
cd .github/tests
.\test-deploy.ps1

# Or test specific workflow
act -W ../workflows/deploy.yml --workflows deploy.yml
```

## Testing Structure

```
.github/tests/
├── README.md              # This file
├── .actrc                  # Act configuration
├── .secrets.example        # Example secrets file
├── test-deploy.ps1        # PowerShell test script
├── test-deploy.sh         # Bash test script
├── test-build-only.yml    # Test workflow (build without deploy)
└── fixtures/              # Test fixtures and mock data
    └── .gitkeep
```

## Test Workflows

### Build-Only Test

The `test-build-only.yml` workflow tests the build steps without actually deploying to Fly.io. This is useful for:
- Validating Go build process
- Testing dependency caching
- Verifying workflow syntax
- Fast feedback during development

### Full Deploy Test

The `deploy.yml` workflow can be tested with Act, but will require:
- Valid `FLY_API_TOKEN` secret
- Fly.io CLI configured
- Network access to Fly.io

**Note**: Actual deployment is skipped in test mode by default.

## Running Tests

### PowerShell (Windows)

```powershell
cd .github/tests
.\test-deploy.ps1
```

### Bash (Linux/macOS)

```bash
cd .github/tests
chmod +x test-deploy.sh
./test-deploy.sh
```

### Manual Act Commands

```bash
# List available workflows
act -l

# Run specific workflow
act -W ../workflows/deploy.yml

# Run with secrets
act -W ../workflows/deploy.yml --secret-file .secrets

# Dry run (show what would run)
act -W ../workflows/deploy.yml --dryrun

# Run specific job
act -j build-and-deploy -W ../workflows/deploy.yml
```

## Configuration

### Act Configuration (`.actrc`)

The `.actrc` file contains Act-specific settings:
- Platform images to use
- Container options
- Environment variables

### Secrets

Copy `.secrets.example` to `.secrets` and fill in your test values:

```bash
cp .secrets.example .secrets
# Edit .secrets with your test credentials
```

**Important**: `.secrets` is in `.gitignore` - never commit real secrets!

## Troubleshooting

### Docker Issues

If Act fails to start containers:
```bash
# Check Docker is running
docker ps

# Pull required images manually
docker pull node:16-buster-slim
docker pull ubuntu-latest
```

### Platform Image Issues

If you see platform image errors, update `.actrc` with correct image tags:
```bash
# List available images
act -l --platform ubuntu-latest=ubuntu:22.04
```

### Secret Issues

If secrets aren't being passed:
```bash
# Verify secrets file format
cat .secrets

# Test with explicit secret
act -W ../workflows/deploy.yml -s FLY_API_TOKEN=test-token
```

## Best Practices

1. **Test Build Steps First**: Use `test-build-only.yml` for fast iteration
2. **Use Dry Run**: Always test with `--dryrun` first to see what will execute
3. **Mock External Services**: Don't hit real Fly.io during development
4. **Keep Secrets Safe**: Never commit `.secrets` file
5. **Test Locally Before Pushing**: Catch workflow errors before CI runs

## Limitations

- Act doesn't perfectly replicate GitHub Actions environment
- Some actions may behave differently locally
- Fly.io deployment should be tested in actual CI/CD environment
- Secrets management differs from GitHub's secret system

## Resources

- [Act Documentation](https://github.com/nektos/act)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Fly.io CLI Documentation](https://fly.io/docs/flyctl/)


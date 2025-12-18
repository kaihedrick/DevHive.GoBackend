# Quick Start Guide - Testing GitHub Actions Locally

## Installation

### Windows
```powershell
# Option 1: Direct Download (Recommended - No Admin needed)
# Download from: https://github.com/nektos/act/releases/latest
# Extract and add to PATH

# Option 2: Chocolatey (Requires Admin - Run PowerShell as Administrator)
choco install act-cli

# Option 3: Scoop (No Admin needed)
scoop install act

# See INSTALL_ACT.md for detailed instructions
```

### macOS
```bash
brew install act
```

### Linux
```bash
# See: https://github.com/nektos/act#installation
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

## Setup

1. **Copy secrets template**:
   ```bash
   cd .github/tests
   cp .secrets.example .secrets
   ```

2. **Edit `.secrets`** with your test values (optional for build-only tests)

3. **Ensure Docker is running**

## Running Tests

### Quick Test (Build Only)
```powershell
# Windows
cd .github/tests
.\test-deploy.ps1 --BuildOnly
```

```bash
# Linux/macOS
cd .github/tests
./test-deploy.sh --build-only
```

### Test Full Deploy Workflow
```powershell
# Windows
.\test-deploy.ps1
```

```bash
# Linux/macOS
./test-deploy.sh
```

### Dry Run (See What Would Execute)
```powershell
.\test-deploy.ps1 --DryRun
```

```bash
./test-deploy.sh --dryrun
```

### Test Specific Job
```powershell
.\test-deploy.ps1 --Job build-and-deploy
```

```bash
./test-deploy.sh --job build-and-deploy
```

## Common Commands

```bash
# List all workflows
act -l

# Run specific workflow
act -W ../workflows/deploy.yml

# Run with secrets
act -W ../workflows/deploy.yml --secret-file .secrets

# Verbose output
act -W ../workflows/deploy.yml -v

# List available jobs
act -l -W ../workflows/deploy.yml
```

## Troubleshooting

### Act not found
- Install Act (see Installation above)
- Restart terminal after installation

### Docker not running
- Start Docker Desktop (Windows/macOS)
- Start Docker service (Linux): `sudo systemctl start docker`

### Platform image errors
- Act will pull images automatically on first run
- Or manually: `docker pull catthehacker/ubuntu:act-latest`

### Secrets not working
- Ensure `.secrets` file exists in `.github/tests/`
- Check file format: `SECRET_NAME=value` (one per line)
- Use `--secret-file .secrets` flag explicitly

## Next Steps

- Read full [README.md](README.md) for detailed documentation
- Customize `.actrc` for your needs
- Add test fixtures in `fixtures/` directory


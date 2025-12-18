# Troubleshooting Guide

## File Lock Issues

### PowerShell Script Locked

If you see an error like:
```
The process cannot access the file because it is being used by another process
```

**Solutions:**

1. **Close VS Code/Cursor**: The file might be locked by your editor
   - Close all editor windows
   - Try again

2. **Close PowerShell Windows**: Any PowerShell session might have the file locked
   - Close all PowerShell windows
   - Open a new PowerShell window
   - Try again

3. **Use Alternative Script**: Use `run-tests.ps1` instead:
   ```powershell
   .\run-tests.ps1 --BuildOnly
   ```

4. **Use Bash Script**: If you have Git Bash or WSL:
   ```bash
   ./test-deploy.sh --build-only
   ```

5. **Manual Creation**: If all else fails, manually create the file:
   - Copy the content from the error message or documentation
   - Create `test-deploy.ps1` manually in `.github/tests/`
   - Save and close the file
   - Try running again

## Act Installation Issues

### Act Not Found

**Windows:**

See [INSTALL_ACT.md](INSTALL_ACT.md) for detailed installation instructions.

**Quick options:**

1. **Direct Download (Recommended - No Admin needed):**
   - Download from: https://github.com/nektos/act/releases/latest
   - Extract and add to PATH

2. **Chocolatey (Requires Admin):**
   ```powershell
   # Run PowerShell as Administrator
   choco install act-cli
   ```

3. **Scoop (No Admin needed):**
   ```powershell
   scoop install act
   ```

4. **winget:**
   ```powershell
   winget install nektos.act
   ```

**If Chocolatey fails with lock/permission errors:**
- Run PowerShell as Administrator
- See [INSTALL_ACT.md](INSTALL_ACT.md) for troubleshooting steps

**macOS:**
```bash
brew install act
```

**Linux:**
```bash
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

After installation, restart your terminal.

## Docker Issues

### Docker Not Running

**Windows/macOS:**
- Start Docker Desktop
- Wait for it to fully start (whale icon in system tray)

**Linux:**
```bash
sudo systemctl start docker
sudo systemctl enable docker  # Enable on boot
```

### Docker Permission Denied (Linux)

```bash
# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or:
newgrp docker
```

## Workflow File Not Found

If you see:
```
❌ Workflow file not found
```

**Check:**
1. You're in the `.github/tests/` directory
2. The workflow file exists in `.github/workflows/`
3. Use correct path format:
   - Windows: `..\workflows\deploy.yml`
   - Linux/macOS: `../workflows/deploy.yml`

## Secrets Not Working

**Symptoms:**
- Secrets not being passed to workflow
- Environment variables empty

**Solutions:**

1. **Check `.secrets` file exists:**
   ```bash
   ls .github/tests/.secrets
   ```

2. **Verify file format:**
   ```
   SECRET_NAME=value
   ```
   - One secret per line
   - No spaces around `=`
   - No quotes needed

3. **Use explicit secret flag:**
   ```bash
   act -W ../workflows/deploy.yml --secret-file .secrets
   ```

4. **Pass secrets directly:**
   ```bash
   act -W ../workflows/deploy.yml -s FLY_API_TOKEN=your-token
   ```

## Platform Image Errors

**Error:**
```
Error: unable to find image
```

**Solution:**
```bash
# Pull the image manually
docker pull catthehacker/ubuntu:act-latest

# Or update .actrc with correct image
```

## Build Failures

### Go Build Fails

**Check:**
1. Go version matches workflow (`go version`)
2. Dependencies are downloaded (`go mod download`)
3. Build works locally (`go build ./cmd/devhive-api`)

### sqlc Not Found

If your workflow uses sqlc:
```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Or ensure it's in PATH
```

## Still Having Issues?

1. **Check Act logs:**
   ```bash
   act -W ../workflows/deploy.yml -v
   ```

2. **Dry run first:**
   ```bash
   act -W ../workflows/deploy.yml --dryrun
   ```

3. **Test with minimal workflow:**
   - Use `test-build-only.yml` first
   - Verify Act works before testing full deploy

4. **Check Act version:**
   ```bash
   act --version
   ```

5. **Update Act:**
   ```bash
   # Windows
   choco upgrade act-cli
   
   # macOS
   brew upgrade act
   ```


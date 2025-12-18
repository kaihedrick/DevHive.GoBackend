# Installing Act on Windows

## Method 1: Chocolatey (Requires Admin)

**Run PowerShell as Administrator**, then:

```powershell
choco install act-cli
```

### If you get lock/permission errors:

1. **Close all Chocolatey processes:**
   ```powershell
   Get-Process choco* | Stop-Process -Force
   ```

2. **Remove lock file** (if it exists):
   ```powershell
   # Run as Administrator
   Remove-Item "C:\ProgramData\chocolatey\lib\5a292bfd398f73c60065c65362910a64c8254386" -Force -ErrorAction SilentlyContinue
   ```

3. **Try installation again:**
   ```powershell
   choco install act-cli
   ```

## Method 2: Scoop (No Admin Required)

If you have [Scoop](https://scoop.sh/) installed:

```powershell
scoop install act
```

If you don't have Scoop:
```powershell
# Install Scoop first
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# Then install Act
scoop install act
```

## Method 3: Direct Download (Recommended)

1. **Download the latest release:**
   - Go to: https://github.com/nektos/act/releases/latest
   - Download `act_windows_amd64.zip` (or appropriate for your architecture)

2. **Extract and add to PATH:**
   ```powershell
   # Extract to a folder (e.g., C:\tools\act)
   Expand-Archive act_windows_amd64.zip -DestinationPath C:\tools\act

   # Add to PATH (current session)
   $env:Path += ";C:\tools\act"

   # Add to PATH permanently (run as Admin)
   [Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\tools\act", [EnvironmentVariableTarget]::Machine)
   ```

3. **Verify installation:**
   ```powershell
   act --version
   ```

## Method 4: Using Go (If you have Go installed)

```powershell
go install github.com/nektos/act@latest
```

Then add Go bin directory to PATH:
```powershell
$env:Path += ";$env:USERPROFILE\go\bin"
```

## Method 5: Using winget (Windows Package Manager)

```powershell
winget install nektos.act
```

## Verification

After installation, verify Act works:

```powershell
act --version
```

You should see something like:
```
act version 0.2.83
```

## Troubleshooting

### "act: command not found"

- **Restart PowerShell** after installation
- **Check PATH**: `$env:Path -split ';' | Select-String act`
- **Add manually**: Add Act's installation directory to your PATH

### Still having issues?

Try the **Direct Download** method (Method 3) - it's the most reliable and doesn't require admin rights or package managers.



# Refresh PATH environment variable in current PowerShell session
# Run this if Act is not recognized in a new terminal

# Reload PATH from environment variables
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

Write-Host "✅ PATH refreshed" -ForegroundColor Green
Write-Host ""

# Test Act
if (Get-Command act -ErrorAction SilentlyContinue) {
    Write-Host "✅ Act is now available:" -ForegroundColor Green
    act --version
} else {
    Write-Host "❌ Act still not found. Checking..." -ForegroundColor Red
    Write-Host ""
    Write-Host "Checking if C:\Nektos\act.exe exists:" -ForegroundColor Yellow
    if (Test-Path "C:\Nektos\act.exe") {
        Write-Host "✅ File exists at C:\Nektos\act.exe" -ForegroundColor Green
        Write-Host ""
        Write-Host "C:\Nektos in PATH:" -ForegroundColor Yellow
        $env:Path -split ';' | Select-String -Pattern "Nektos"
        Write-Host ""
        Write-Host "Try running: C:\Nektos\act.exe --version" -ForegroundColor Yellow
    } else {
        Write-Host "❌ File not found at C:\Nektos\act.exe" -ForegroundColor Red
    }
}


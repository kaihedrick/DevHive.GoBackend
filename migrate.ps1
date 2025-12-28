# DevHive Database Migration Script
# This script runs all migrations in order against your Neon database

param(
    [Parameter(Mandatory=$true)]
    [string]$DatabaseURL
)

Write-Host "DevHive Database Migration Script" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check if psql is available
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "[ERROR] psql not found. Please install PostgreSQL client tools:" -ForegroundColor Red
    Write-Host "   winget install PostgreSQL.PostgreSQL" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Or use Neon SQL Editor at: https://console.neon.tech" -ForegroundColor Yellow
    exit 1
}

# Get all migration files in order
$migrationFiles = Get-ChildItem "cmd\devhive-api\migrations\*.sql" |
    Where-Object { $_.Name -match '^\d{3}_' } |
    Sort-Object Name

Write-Host "Found $($migrationFiles.Count) migration files" -ForegroundColor Green
Write-Host ""

foreach ($file in $migrationFiles) {
    Write-Host "[RUNNING] $($file.Name)" -ForegroundColor Cyan

    $env:PGPASSWORD = ""  # Will use password from connection string
    $result = & psql "$DatabaseURL" -f $file.FullName 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Host "   [SUCCESS]" -ForegroundColor Green
    } else {
        # Check if error is idempotent (already exists)
        if ($result -match "already exists|duplicate key") {
            Write-Host "   [SKIPPED] Already applied" -ForegroundColor Yellow
        } else {
            Write-Host "   [FAILED] $result" -ForegroundColor Red
            exit 1
        }
    }
}

Write-Host ""
Write-Host "[COMPLETE] All migrations completed successfully!" -ForegroundColor Green
Write-Host ""

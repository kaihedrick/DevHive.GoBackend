# PowerShell script to test GitHub Actions deploy workflow locally using Act
# Usage: .\run-tests.ps1 [--workflow WORKFLOW] [--dryrun] [--job JOB]

param(
    [string]$Workflow = "deploy.yml",
    [switch]$DryRun = $false,
    [string]$Job = "",
    [switch]$BuildOnly = $false,
    [switch]$Help = $false
)

if ($Help) {
    Write-Host @"
GitHub Actions Local Testing Script

Usage:
    .\run-tests.ps1 [OPTIONS]

Options:
    --Workflow WORKFLOW    Workflow file to test (default: deploy.yml)
    --DryRun              Show what would run without executing
    --Job JOB             Run specific job only
    --BuildOnly           Test build-only workflow instead
    --Help                Show this help message

Examples:
    .\run-tests.ps1
    .\run-tests.ps1 --Workflow deploy.yml --DryRun
    .\run-tests.ps1 --Job build-and-deploy
    .\run-tests.ps1 --BuildOnly

"@
    exit 0
}

# Check if Act is installed
$actInstalled = Get-Command act -ErrorAction SilentlyContinue
if (-not $actInstalled) {
    Write-Host "❌ Act is not installed. Please install it first:" -ForegroundColor Red
    Write-Host "   choco install act-cli" -ForegroundColor Yellow
    Write-Host "   Or download from: https://github.com/nektos/act/releases" -ForegroundColor Yellow
    exit 1
}

# Check if Docker is running
try {
    docker ps | Out-Null
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host "🚀 Starting GitHub Actions local test..." -ForegroundColor Green
Write-Host ""

# Set working directory to tests folder
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

# Determine workflow file
if ($BuildOnly) {
    $workflowFile = "test-build-only.yml"
    Write-Host "📋 Testing build-only workflow..." -ForegroundColor Cyan
} else {
    $workflowFile = "..\workflows\$Workflow"
    Write-Host "📋 Testing workflow: $Workflow" -ForegroundColor Cyan
}

# Check if workflow file exists
if (-not (Test-Path $workflowFile)) {
    Write-Host "❌ Workflow file not found: $workflowFile" -ForegroundColor Red
    exit 1
}

# Build Act command
$actCmd = "act -W $workflowFile"

# Add secrets file if it exists
if (Test-Path ".secrets") {
    $actCmd += " --secret-file .secrets"
    Write-Host "🔐 Using secrets from .secrets file" -ForegroundColor Cyan
} else {
    Write-Host "⚠️  No .secrets file found. Using environment variables only." -ForegroundColor Yellow
    Write-Host "   Copy .secrets.example to .secrets and add your test secrets." -ForegroundColor Yellow
}

# Add dry run flag
if ($DryRun) {
    $actCmd += " --dryrun"
    Write-Host "🔍 Dry run mode - no actions will be executed" -ForegroundColor Cyan
}

# Add job filter
if ($Job) {
    $actCmd += " -j $Job"
    Write-Host "🎯 Running specific job: $Job" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Command: $actCmd" -ForegroundColor Gray
Write-Host ""

# Execute Act
Invoke-Expression $actCmd

$exitCode = $LASTEXITCODE

if ($exitCode -eq 0) {
    Write-Host ""
    Write-Host "✅ Test completed successfully!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "❌ Test failed with exit code: $exitCode" -ForegroundColor Red
}

exit $exitCode


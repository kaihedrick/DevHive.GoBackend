# Script to check if NOTIFY listener starts correctly on Fly.io
# Usage: .\check_notify_listener.ps1

Write-Host "Checking NOTIFY listener startup logs..." -ForegroundColor Cyan
Write-Host ""

# Get recent logs and filter for NOTIFY listener messages
$logs = flyctl logs --app devhive-go-backend --no-tail 2>&1 | Select-String -Pattern "NOTIFY listener|PostgreSQL NOTIFY|listener started" 

if ($logs) {
    Write-Host "✅ NOTIFY Listener Status:" -ForegroundColor Green
    Write-Host ""
    $logs | ForEach-Object {
        Write-Host $_ -ForegroundColor Green
    }
    Write-Host ""
    Write-Host "✅ NOTIFY listener is starting correctly!" -ForegroundColor Green
} else {
    Write-Host "⚠️  No NOTIFY listener startup logs found" -ForegroundColor Yellow
    Write-Host "This might mean the app hasn't restarted recently." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Checking for any NOTIFY received messages..." -ForegroundColor Cyan
$notifyMessages = flyctl logs --app devhive-go-backend --no-tail 2>&1 | Select-String -Pattern "NOTIFY received|Project member notification" | Select-Object -Last 5

if ($notifyMessages) {
    Write-Host "Recent NOTIFY messages:" -ForegroundColor Cyan
    $notifyMessages | ForEach-Object {
        Write-Host $_ -ForegroundColor White
    }
} else {
    Write-Host "No recent NOTIFY messages found (this is normal if no changes occurred)" -ForegroundColor Gray
}






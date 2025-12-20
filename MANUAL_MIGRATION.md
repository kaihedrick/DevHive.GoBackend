# Manual Migration Guide - NOTIFY Triggers

## Option 1: Run Migration via API (Recommended)

### Run the migration (Linux/Mac/Git Bash):
```bash
curl -X POST https://devhive-go-backend.fly.dev/api/v1/migrations/run \
  -H "Content-Type: application/json" \
  -d '{"scriptName": "007_ensure_notify_triggers.sql"}'
```

### Run the migration (PowerShell - Windows):
```powershell
$body = @{
    scriptName = "007_ensure_notify_triggers.sql"
} | ConvertTo-Json

Invoke-RestMethod -Uri "https://devhive-go-backend.fly.dev/api/v1/migrations/run" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

### Run the migration (PowerShell - One-liner):
```powershell
Invoke-RestMethod -Uri "https://devhive-go-backend.fly.dev/api/v1/migrations/run" -Method POST -ContentType "application/json" -Body '{"scriptName":"007_ensure_notify_triggers.sql"}'
```

### For local development:
```bash
curl -X POST http://localhost:8080/api/v1/migrations/run \
  -H "Content-Type: application/json" \
  -d '{"scriptName": "007_ensure_notify_triggers.sql"}'
```

### Check the error response (if you get 400):
```powershell
try {
    Invoke-RestMethod -Uri "https://devhive-go-backend.fly.dev/api/v1/migrations/run" -Method POST -ContentType "application/json" -Body '{"scriptName":"007_ensure_notify_triggers.sql"}'
} catch {
    $_.Exception.Response | ConvertTo-Json
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $reader.BaseStream.Position = 0
    $reader.DiscardBufferedData()
    $responseBody = $reader.ReadToEnd()
    Write-Host "Error response: $responseBody"
}
```

### Verify it worked:
```bash
curl https://devhive-go-backend.fly.dev/api/v1/migrations/verify?migrationName=004_add_cache_invalidation_triggers.sql
```

### Test the NOTIFY system:
```bash
curl -X POST https://devhive-go-backend.fly.dev/api/v1/migrations/test-notify \
  -H "Content-Type: application/json"
```

## Option 2: Run SQL Directly in Database

If you have direct database access (via `flyctl postgres connect` or psql):

### Connect to your database:
```bash
flyctl postgres connect -a devhive-go-backend
```

### Then run the SQL:
```sql
-- Copy and paste the entire contents of:
-- cmd/devhive-api/migrations/007_ensure_notify_triggers.sql
```

Or run it directly:
```bash
flyctl postgres connect -a devhive-go-backend < cmd/devhive-api/migrations/007_ensure_notify_triggers.sql
```

### Verify triggers exist:
```sql
-- Check if function exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_proc p
            JOIN pg_namespace n ON p.pronamespace = n.oid
            WHERE n.nspname = 'public' AND p.proname = 'notify_cache_invalidation'
        ) THEN '✅ Function EXISTS'
        ELSE '❌ Function MISSING'
    END as function_status;

-- List all triggers
SELECT 
    c.relname AS table_name,
    t.tgname AS trigger_name
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE n.nspname = 'public'
AND t.tgname LIKE '%cache_invalidate%'
AND NOT t.tgisinternal
ORDER BY c.relname;
```

## Option 3: PowerShell (Windows)

```powershell
# Run migration
Invoke-RestMethod -Uri "https://devhive-go-backend.fly.dev/api/v1/migrations/run" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"scriptName": "007_ensure_notify_triggers.sql"}'

# Test NOTIFY
Invoke-RestMethod -Uri "https://devhive-go-backend.fly.dev/api/v1/migrations/test-notify" `
  -Method POST `
  -ContentType "application/json"
```

## Expected Response

### Successful migration:
```json
{
  "success": true,
  "message": "Successfully executed migration script: 007_ensure_notify_triggers.sql"
}
```

### Successful test:
```json
{
  "success": true,
  "message": "Test NOTIFY trigger executed. Check logs for '🔔 RAW NOTIFY received' messages.",
  "project_id": "...",
  "user_id": "...",
  "note": "If you see this but no NOTIFY logs, the trigger may not be installed or firing."
}
```

## Troubleshooting

### If migration fails with "file not found":
The migration file needs to be in the deployed container. Make sure:
1. The file exists in `cmd/devhive-api/migrations/007_ensure_notify_triggers.sql`
2. You've deployed the latest code to Fly.io

### If you see "already exists" errors:
This is OK! The migration is idempotent. The function/triggers already exist.

### If test-notify works but no NOTIFY logs appear:
1. Check that the NOTIFY listener is running (should see startup logs)
2. Check that triggers are actually installed (use verification SQL above)
3. Check backend logs for `🔔 RAW NOTIFY received` messages


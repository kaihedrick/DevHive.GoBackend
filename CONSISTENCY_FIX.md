# Project Owner Consistency Fix

## Problem Identified

Project owners were **not** being inserted into the `project_members` table, causing inconsistent behavior:

- **Owners** see empty member lists (because they query `project_members` which doesn't include them)
- **Members** see everyone (because they use `ListProjectMembers` which includes owner via UNION)
- **Owners** can't assign tasks to other members (empty assignee list)
- **Owners** see empty sprints/tasks context

## Root Cause

1. `CreateProject` only created the project, didn't insert owner into `project_members`
2. `GetProjectMembers` query only looked at `project_members` table (missed owner)
3. `ListProjectMembers` query correctly included owner via UNION (this is why members saw everyone)

## Fixes Applied

### 1. Fixed `CreateProject` Handler
- Now inserts owner into `project_members` with role 'owner' immediately after project creation
- Ensures all new projects have owners in the members table

### 2. Fixed `GetProjectMembers` Query
- Updated to match `ListProjectMembers` structure
- Now includes owner via UNION (same pattern as `ListProjectMembers`)
- Both queries now return consistent results

### 3. Created Backfill Migration
- `008_backfill_owners_in_project_members.sql` 
- Inserts all existing project owners into `project_members` table
- Idempotent - safe to run multiple times

## Next Steps

### 1. Regenerate SQLC Code
After changing the SQL queries, you need to regenerate the Go code:

```bash
sqlc generate
```

This will update `internal/repo/queries.sql.go` with the new `GetProjectMembers` structure.

### 2. Run Migrations
Deploy and run the backfill migration:

```bash
# Via API
curl -X POST https://devhive-go-backend.fly.dev/api/v1/migrations/run \
  -H "Content-Type: application/json" \
  -d '{"scriptName": "008_backfill_owners_in_project_members.sql"}'

# Or it will run automatically on next deployment
```

### 3. Verify Fix
After deployment:
1. Create a new project as owner → should see yourself in members list
2. Check existing projects → owners should now appear in members
3. Assign tasks → should see all members including owner
4. Create sprints → should work correctly

## Files Changed

1. `internal/http/handlers/project.go` - CreateProject now inserts owner
2. `internal/projects/queries.sql` - GetProjectMembers now includes owner
3. `cmd/devhive-api/migrations/008_backfill_owners_in_project_members.sql` - Backfill migration

## Consistency Rules Going Forward

✅ **ALWAYS** insert owner into `project_members` when creating a project
✅ **ALWAYS** use queries that include owner (via UNION or explicit join)
✅ **NEVER** query `project_members` alone without including owner
✅ **ALWAYS** use `ListProjectMembers` or updated `GetProjectMembers` for member lists




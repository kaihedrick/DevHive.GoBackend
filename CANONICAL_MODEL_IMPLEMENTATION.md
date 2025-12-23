# Canonical Project Members Model - Implementation Complete

## ✅ Changes Applied

### 1. SQL Queries Updated

#### GetProjectMembers
- **Before**: Used UNION ALL to include owner from `projects` table
- **After**: Only queries `project_members` table (canonical source)
- **File**: `internal/projects/queries.sql`

#### ListProjectMembers  
- **Before**: Used UNION ALL to include owner from `projects` table
- **After**: Only queries `project_members` table (canonical source)
- **File**: `internal/projects/queries.sql`

#### ListProjectsByUser
- **Before**: Used LEFT JOIN which could cause duplicate rows
- **After**: Uses EXISTS subquery to avoid duplicates
- **File**: `internal/projects/queries.sql`

#### AddProjectMember
- **Before**: `ON CONFLICT DO UPDATE SET role = $3`
- **After**: `ON CONFLICT DO UPDATE SET role = EXCLUDED.role` (more explicit)
- **File**: `internal/projects/queries.sql`

### 2. Database Migrations

#### Migration 009: Canonical Project Members
- **File**: `cmd/devhive-api/migrations/009_canonical_project_members.sql`
- **Actions**:
  1. Adds unique constraint `(project_id, user_id)` if not exists
  2. Backfills all existing project owners into `project_members`
  3. Logs migration statistics

### 3. Handler Updates

#### CreateProject Handler
- Already inserts owner into `project_members` (idempotent via ON CONFLICT)
- Comment updated to reflect canonical model

## 📋 Next Steps

### 1. Regenerate SQLC Code
```bash
sqlc generate
```

This will update `internal/repo/queries.sql.go` with:
- Simplified `GetProjectMembers` query (no UNION)
- Simplified `ListProjectMembers` query (no UNION)  
- Updated `ListProjectsByUser` query (EXISTS instead of LEFT JOIN)
- Updated `AddProjectMember` with `EXCLUDED.role`

### 2. Run Migration
```bash
# Via API
curl -X POST https://devhive-go-backend.fly.dev/api/v1/migrations/run \
  -H "Content-Type: application/json" \
  -d '{"scriptName": "009_canonical_project_members.sql"}'

# Or it will run automatically on deployment
```

### 3. Verify Changes

After `sqlc generate`, check `internal/repo/queries.sql.go`:

**GetProjectMembers should be:**
```go
const getProjectMembers = `-- name: GetProjectMembers :many
SELECT pm.project_id, pm.user_id, pm.role, pm.joined_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at
`
```

**ListProjectMembers should be:**
```go
const listProjectMembers = `-- name: ListProjectMembers :many
SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url,
       pm.joined_at, pm.role
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at
`
```

## ✅ Benefits

1. **No Duplicates**: Owner appears exactly once in member lists
2. **Consistent Queries**: All member queries use same source (`project_members`)
3. **Simpler SQL**: No UNION ALL needed
4. **Simpler SQLC Structs**: Single query structure
5. **Database Integrity**: Unique constraint prevents duplicates
6. **Idempotent**: Safe to run migrations and handlers multiple times

## 🧪 Test Checklist

- [ ] Create new project → owner appears in members list exactly once
- [ ] Existing projects → owners appear after migration runs
- [ ] ListProjectsByUser → no duplicate project rows
- [ ] JoinProject → adds member, no conflicts
- [ ] AcceptInvite → adds member, no conflicts
- [ ] CheckProjectAccess → owner + member both return true
- [ ] GetProjectBundle → members list includes owner

## 📝 Notes

- Migration 008 (`008_backfill_owners_in_project_members.sql`) is superseded by 009
- Migration 009 is idempotent and can be run multiple times safely
- The unique constraint ensures no duplicates even if code tries to insert twice
- All existing handlers already use `AddProjectMember` which is now idempotent





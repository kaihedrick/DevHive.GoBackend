# Project Creation Invite Fix

## Problem

**Symptom**: New projects don't allow invite creation/viewing, but old projects work fine.

**Root Cause**: 
1. `CreateProject` handler was silently failing when adding owner to `project_members` table
2. `CreateProject` response didn't include `userRole` and `permissions`, so frontend didn't know user had invite permissions
3. `ListInvites` uses `CheckProjectAccess` which only checks `project_members` - if owner isn't there, access is denied

## What Was Fixed

### 1. Made `AddProjectMember` failure fatal
**Before**:
```go
err = h.queries.AddProjectMember(...)
if err != nil {
    log.Printf("Warning: Failed to add owner...") // Only logged, didn't fail
}
```

**After**:
```go
err = h.queries.AddProjectMember(...)
if err != nil {
    log.Printf("ERROR: Failed to add owner...")
    response.InternalServerError(w, "Failed to initialize project membership")
    return // Now fails the request
}
```

### 2. Added `userRole` and `permissions` to `CreateProject` response
**Before**:
```go
response.JSON(w, http.StatusCreated, ProjectResponse{
    ID: project.ID.String(),
    // ... basic fields only
    // No userRole or permissions
})
```

**After**:
```go
userRole, permissions := h.getUserRoleAndPermissions(r.Context(), project.ID, userUUID)
response.JSON(w, http.StatusCreated, ProjectResponse{
    ID: project.ID.String(),
    // ... all fields
    Owner: {...}, // Owner details
    UserRole: userRole,        // ✅ Now included
    Permissions: permissions,   // ✅ Now included
})
```

## Why Old Projects Work

Old projects likely have owners in `project_members` table (either from migration `008_backfill_owners_in_project_members.sql` or they were added manually). New projects created before this fix might not have owners in `project_members` if `AddProjectMember` failed silently.

## Fixing Existing Projects

If you have existing projects where owners can't access invites, run this SQL to backfill:

```sql
-- Backfill owners into project_members for projects missing them
INSERT INTO project_members (project_id, user_id, role, joined_at)
SELECT
    p.id as project_id,
    p.owner_id as user_id,
    'owner' as role,
    p.created_at as joined_at
FROM projects p
WHERE NOT EXISTS (
    SELECT 1
    FROM project_members pm
    WHERE pm.project_id = p.id
    AND pm.user_id = p.owner_id
)
ON CONFLICT (project_id, user_id) DO UPDATE
SET role = 'owner'
WHERE project_members.role != 'owner';
```

## Testing

After deploying the fix:

1. **Create a new project**:
   ```bash
   POST /api/v1/projects
   {
     "name": "Test Project",
     "description": "Test"
   }
   ```

2. **Verify response includes permissions**:
   ```json
   {
     "id": "...",
     "userRole": "owner",
     "permissions": {
       "canViewInvites": true,
       "canCreateInvites": true,
       "canRevokeInvites": true,
       "canManageMembers": true
     }
   }
   ```

3. **Try to list invites**:
   ```bash
   GET /api/v1/projects/{projectId}/invites
   ```
   Should return `200 OK` with invites array (even if empty)

4. **Try to create invite**:
   ```bash
   POST /api/v1/projects/{projectId}/invites
   ```
   Should return `201 Created` with invite details

## Authorization Flow

### ListInvites (View Invites)
- Uses: `CheckProjectAccess` (checks `project_members` table)
- Requires: User must be in `project_members` for the project
- **Fix**: Owner is now guaranteed to be in `project_members` on creation

### CreateInvite (Create Invites)
- Uses: `CheckProjectOwnerOrAdmin` (checks `projects.owner_id` OR `project_members` with admin role)
- Requires: User is owner (from `projects.owner_id`) OR admin (from `project_members`)
- **Note**: This should work even if owner isn't in `project_members` (checks `projects.owner_id`), but having owner in `project_members` ensures consistency

## Summary

✅ **Fixed**: `CreateProject` now fails if owner can't be added to `project_members`  
✅ **Fixed**: `CreateProject` response now includes `userRole` and `permissions`  
✅ **Result**: New projects will have owners in `project_members` and frontend will know about permissions





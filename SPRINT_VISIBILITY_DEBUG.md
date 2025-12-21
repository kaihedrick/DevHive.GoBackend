# Sprint Visibility Issue - Debug Guide

## Problem

When a non-owner (member) creates a sprint, the owner doesn't see it until it becomes ACTIVE. The cache invalidation is working (NOTIFY is firing), but the owner's UI isn't updating.

## Backend Analysis

### ✅ Backend Query is Correct

The `ListSprintsByProject` query does **NOT** filter by status or creator:

```sql
SELECT s.id, s.project_id, s.name, ...
FROM sprints s
WHERE s.project_id = $1
ORDER BY s.start_date DESC
LIMIT $2 OFFSET $3;
```

**No filters on:**
- ❌ `status` (no status column exists - uses `is_started` and `is_completed`)
- ❌ `created_by` (no creator filtering)
- ❌ `is_started` or `is_completed` (returns all sprints regardless of state)

### ✅ Cache Invalidation is Working

Logs confirm:
- NOTIFY is firing: `resource=sprint`, `action=INSERT`
- WebSocket broadcasts are reaching both users
- Backend is correctly sending cache invalidation messages

## Most Likely Causes

### 1. Frontend Filtering by Status (Most Likely)

The frontend is likely filtering sprints to show only **active** ones (`is_started = true && is_completed = false`), so newly created sprints (which default to `is_started = false`) are hidden.

**Check**: Look at the frontend sprint list component - does it filter by `is_started` or `status`?

### 2. Frontend Cache Invalidation Query Key Mismatch

The frontend might be using a different query key than what's being invalidated.

**Backend sends**: `resource: "sprint"` (singular, normalized)

**Frontend should invalidate**:
```typescript
case 'sprint':
  queryClient.invalidateQueries({ queryKey: ['sprints', 'list', project_id] });
  // OR
  queryClient.invalidateQueries({ queryKey: ['sprints', project_id] });
```

**Check**: Does the frontend query key match what's being invalidated?

### 3. Frontend Not Refetching After Invalidation

TanStack Query `invalidateQueries` marks queries as stale but doesn't automatically refetch unless:
- The query is currently being observed (mounted component)
- `refetchOnMount` or `refetchOnWindowFocus` is enabled

**Check**: Is the sprint list component mounted when the cache invalidation arrives?

## Debugging Steps

### Step 1: Check Backend Response

After a member creates a sprint (while it's still inactive), have the owner call:

```bash
GET /api/v1/projects/{projectId}/sprints
```

**Expected**: Should return the newly created sprint (even if `is_started = false`)

**If sprint is NOT in response** → Backend issue (but query looks correct, so unlikely)

**If sprint IS in response** → Frontend issue (filtering or cache invalidation)

### Step 2: Check Backend Logs

Added logging to `ListSprintsByProject` handler. After deploying, logs will show:

```
📋 ListSprintsByProject: user_id=..., project_id=..., sprints_returned=3, limit=20, offset=0
📋 Sprint status breakdown: active=1, inactive/planned=2
```

This confirms:
- How many sprints the backend is returning
- Status breakdown (active vs inactive)

### Step 3: Check Frontend Cache Invalidation

In browser DevTools, check:
1. Is the WebSocket message received? (should see `cache_invalidate` with `resource: "sprint"`)
2. Is the query being invalidated? (check TanStack Query DevTools)
3. Is the component refetching? (check Network tab for new GET request)

### Step 4: Check Frontend Query Key

Verify the frontend query key matches what's being invalidated:

```typescript
// Frontend query
const { data } = useQuery({
  queryKey: ['sprints', 'list', projectId], // ← Check this key
  queryFn: () => fetchSprints(projectId)
});

// Cache invalidation handler
case 'sprint':
  queryClient.invalidateQueries({ 
    queryKey: ['sprints', 'list', project_id] // ← Must match exactly
  });
```

## Fixes

### Fix 1: Ensure Frontend Shows All Sprints

If the frontend is filtering by status, ensure it shows:
- **Active sprints** (`is_started = true && is_completed = false`)
- **Planned/Inactive sprints** (`is_started = false`) ← This is what's missing

Example:
```typescript
// ❌ Wrong - only shows active
const activeSprints = sprints.filter(s => s.isStarted && !s.isCompleted);

// ✅ Correct - shows all sprints, or at least planned + active
const visibleSprints = sprints.filter(s => 
  !s.isCompleted // Show all non-completed sprints
);
```

### Fix 2: Use Refetch Instead of Invalidate

If the component isn't mounted when invalidation arrives, use `refetchQueries`:

```typescript
case 'sprint':
  // Immediately refetch if query is active
  queryClient.refetchQueries({ 
    queryKey: ['sprints', 'list', project_id],
    exact: false
  });
  // Also invalidate for when component mounts later
  queryClient.invalidateQueries({ 
    queryKey: ['sprints', 'list', project_id] 
  });
```

### Fix 3: Add Status Filter to Backend (If Needed)

If you want to support filtering by status, add query parameters:

```go
// In ListSprintsByProject handler
statusFilter := r.URL.Query().Get("status") // "active", "planned", "completed", or empty for all

// Then modify SQL query to filter if needed
```

But this should NOT be the default - all sprints should be returned by default.

## Testing

1. **Member creates sprint** (should be `is_started = false`)
2. **Owner immediately calls** `GET /api/v1/projects/{projectId}/sprints`
3. **Check response** - should include the new sprint
4. **Check logs** - should show sprint count includes the new one
5. **Check frontend** - does it show the sprint?

## Summary

- ✅ **Backend query is correct** - returns all sprints
- ✅ **Cache invalidation is working** - NOTIFY and WebSocket broadcasts are correct
- ❓ **Frontend likely filtering** - probably only showing active sprints
- ❓ **Frontend cache invalidation** - might not be refetching or query key mismatch

**Next Steps**: Check frontend sprint list component for status filtering and verify cache invalidation query keys match.


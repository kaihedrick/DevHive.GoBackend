# Sprint Visibility Fix - Owner Not Seeing New Sprints

## Problem Summary

When a **member** creates a sprint, the **owner** doesn't see it until it becomes ACTIVE. Cache invalidation is working (NOTIFY fires, WebSocket broadcasts), but the owner's UI doesn't update.

## Root Cause Analysis

### ✅ Backend is Correct

1. **Query doesn't filter by status**: `ListSprintsByProject` returns ALL sprints for a project, regardless of `is_started` or `is_completed` status.

2. **Cache invalidation works**: NOTIFY triggers correctly, WebSocket broadcasts reach all users.

3. **No creator filtering**: Query doesn't filter by who created the sprint.

### ❌ Frontend Issue (Most Likely)

The frontend is likely:
1. **Filtering by status** - Only showing sprints where `is_started = true`
2. **Not refetching after invalidation** - Query is invalidated but component doesn't refetch
3. **Query key mismatch** - Cache invalidation uses different key than the query

## Backend Changes Made

### Added Logging to `ListSprintsByProject`

```go
// Log sprint list query for debugging
log.Printf("📋 ListSprintsByProject: user_id=%s, project_id=%s, sprints_returned=%d, limit=%d, offset=%d",
    userUUID.String(), projectUUID.String(), len(sprints), limit, offset)
if len(sprints) > 0 {
    // Log status breakdown
    activeCount := 0
    inactiveCount := 0
    for _, s := range sprints {
        if s.IsStarted && !s.IsCompleted {
            activeCount++
        } else {
            inactiveCount++
        }
    }
    log.Printf("📋 Sprint status breakdown: active=%d, inactive/planned=%d", activeCount, inactiveCount)
}
```

**This will show in logs**:
- How many sprints the backend is returning
- Breakdown of active vs inactive sprints
- Which user is requesting the list

## Frontend Fixes Needed

### Fix 1: Show All Non-Completed Sprints

**Problem**: Frontend might be filtering to only show active sprints.

**Solution**: Show all sprints that aren't completed, or at least show planned + active:

```typescript
// ❌ Wrong - only shows active sprints
const visibleSprints = sprints.filter(s => 
  s.isStarted && !s.isCompleted
);

// ✅ Correct - show all non-completed sprints
const visibleSprints = sprints.filter(s => 
  !s.isCompleted // Shows both planned (isStarted=false) and active (isStarted=true)
);

// OR show all sprints with sections
const activeSprints = sprints.filter(s => s.isStarted && !s.isCompleted);
const plannedSprints = sprints.filter(s => !s.isStarted && !s.isCompleted);
const completedSprints = sprints.filter(s => s.isCompleted);
```

### Fix 2: Use Refetch Instead of Just Invalidate

**Problem**: `invalidateQueries` marks queries as stale but doesn't refetch if component isn't actively observing.

**Solution**: Use `refetchQueries` for immediate updates:

```typescript
// In cache invalidation handler
case 'sprint':
  if (id) {
    queryClient.invalidateQueries({ queryKey: ['sprints', 'detail', id] });
  }
  
  // Immediately refetch if query is active (component is mounted)
  queryClient.refetchQueries({ 
    queryKey: ['sprints', 'list', project_id],
    exact: false // Refetch all queries that start with this key
  });
  
  // Also invalidate for when component mounts later
  queryClient.invalidateQueries({ 
    queryKey: ['sprints', 'list', project_id] 
  });
  
  // Invalidate project bundle
  queryClient.invalidateQueries({ 
    queryKey: ['projects', 'bundle', project_id] 
  });
  break;
```

### Fix 3: Verify Query Key Matches

**Problem**: Query key mismatch between query and invalidation.

**Check**:
```typescript
// Your sprint list query
const { data } = useQuery({
  queryKey: ['sprints', 'list', projectId], // ← Must match
  queryFn: () => fetchSprints(projectId)
});

// Cache invalidation
case 'sprint':
  queryClient.invalidateQueries({ 
    queryKey: ['sprints', 'list', project_id] // ← Must match exactly
  });
```

### Fix 4: Ensure Component is Observing Query

**Problem**: Component might not be mounted when cache invalidation arrives.

**Solution**: Ensure the sprint list component is mounted/active when viewing the project:

```typescript
// Make sure the component using the query is mounted
function SprintList({ projectId }: { projectId: string }) {
  const { data: sprints, isLoading } = useQuery({
    queryKey: ['sprints', 'list', projectId],
    queryFn: () => fetchSprints(projectId),
    // Ensure it refetches when component mounts
    refetchOnMount: true,
    // Ensure it refetches when window regains focus
    refetchOnWindowFocus: true,
  });
  
  // ... render sprints
}
```

## Testing Steps

### 1. Test Backend Response

After a member creates a sprint (while inactive), have the owner call:

```bash
curl -H "Authorization: Bearer {owner_token}" \
  "https://devhive-go-backend.fly.dev/api/v1/projects/{projectId}/sprints"
```

**Expected**: Response should include the newly created sprint (even if `isStarted: false`)

**Check logs**: Should see:
```
📋 ListSprintsByProject: user_id=..., project_id=..., sprints_returned=3, ...
📋 Sprint status breakdown: active=1, inactive/planned=2
```

### 2. Test Frontend Cache Invalidation

1. Open browser DevTools → Network tab
2. Member creates a sprint
3. Check if owner's browser receives WebSocket message
4. Check if owner's browser makes a new GET request to `/sprints`
5. Check if the sprint appears in the UI

### 3. Test Frontend Filtering

1. Check the sprint list component code
2. Look for `.filter()` calls on the sprints array
3. Verify it's not filtering out `isStarted: false` sprints

## Expected Behavior

### When Member Creates Sprint

1. ✅ Sprint is created with `isStarted: false`, `isCompleted: false`
2. ✅ NOTIFY fires: `resource: "sprint"`, `action: "INSERT"`
3. ✅ WebSocket broadcasts to all connected users
4. ✅ Frontend receives `cache_invalidate` message
5. ✅ Frontend invalidates/refetches sprint list query
6. ✅ **Owner sees the new sprint immediately** (even though it's not started)

### When Sprint Becomes Active

1. ✅ Sprint status updated: `isStarted: true`
2. ✅ NOTIFY fires: `resource: "sprint"`, `action: "UPDATE"`
3. ✅ Frontend updates sprint in list
4. ✅ Sprint moves to "Active" section (if UI has sections)

## Summary

- ✅ **Backend is correct** - Returns all sprints, no filtering
- ✅ **Cache invalidation works** - NOTIFY and WebSocket are functioning
- ❌ **Frontend likely filtering** - Probably only showing active sprints
- ❌ **Frontend not refetching** - Might need `refetchQueries` instead of just `invalidateQueries`

**Next Steps**: 
1. Check frontend sprint list component for status filtering
2. Update cache invalidation to use `refetchQueries`
3. Verify query keys match between query and invalidation
4. Test with backend logs to confirm sprints are being returned





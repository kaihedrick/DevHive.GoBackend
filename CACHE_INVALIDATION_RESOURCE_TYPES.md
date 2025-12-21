# Cache Invalidation Resource Types

## Backend → Frontend Resource Name Mapping

The backend normalizes database table names to singular resource names for frontend consistency:

| Database Table | Resource Name (Sent to Frontend) | Frontend Case Statement |
|---------------|-----------------------------------|-------------------------|
| `projects` | `project` | `case 'project':` |
| `sprints` | `sprint` | `case 'sprint':` |
| `tasks` | `task` | `case 'task':` |
| `project_members` | `project_members` (kept plural) | `case 'project_members':` |

## Frontend Handling

Your frontend should handle these resource types in the cache invalidation switch statement:

```typescript
switch (resource) {
  case 'project':
    // Handle project changes
    if (id) {
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
    }
    queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
    break;
  
  case 'sprint':
    // Handle sprint changes
    if (id) {
      queryClient.invalidateQueries({ queryKey: ['sprints', id] });
    }
    queryClient.invalidateQueries({ queryKey: ['sprints', 'list', project_id] });
    break;
  
  case 'task':
    // Handle task changes
    if (id) {
      queryClient.invalidateQueries({ queryKey: ['tasks', id] });
    }
    queryClient.invalidateQueries({ queryKey: ['tasks', 'list', project_id] });
    break;
  
  case 'project_members':
    // Handle member changes (immediately refetch for real-time updates)
    queryClient.refetchQueries({ 
      queryKey: ['projectMembers', project_id],
      exact: false 
    });
    break;
  
  default:
    console.warn('Unknown resource type for cache invalidation:', resource);
}
```

**See `CACHE_INVALIDATION_PAYLOAD_REFERENCE.md` for complete payload examples and implementation guide.**

## Migration Status

✅ **Fixed in migration `007_ensure_notify_triggers.sql`**

The `notify_cache_invalidation()` PostgreSQL function now normalizes resource names:
- `projects` → `project`
- `sprints` → `sprint`
- `tasks` → `task`
- `project_members` → `project_members` (unchanged)

## Before vs After

### Before (Incorrect)
```json
{
  "resource": "tasks",  // ❌ Plural from table name
  "action": "INSERT",
  "project_id": "..."
}
```

### After (Correct)
```json
{
  "resource": "task",  // ✅ Singular, normalized
  "action": "INSERT",
  "project_id": "..."
}
```

## Testing

After deploying the migration, verify:
1. Creating a task sends `resource: "task"` (not `"tasks"`)
2. Creating a sprint sends `resource: "sprint"` (not `"sprints"`)
3. Creating a project sends `resource: "project"` (not `"projects"`)
4. Adding a member sends `resource: "project_members"` (unchanged)


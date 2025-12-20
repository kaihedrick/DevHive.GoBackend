# Cache Invalidation Resource Types

## Backend → Frontend Resource Name Mapping

The backend normalizes database table names to singular resource names for frontend consistency:

| Database Table | Resource Name (Sent to Frontend) |
|---------------|-----------------------------------|
| `projects` | `project` |
| `sprints` | `sprint` |
| `tasks` | `task` |
| `project_members` | `project_members` (kept plural) |

## Frontend Handling

Your frontend should handle these resource types in the cache invalidation switch statement:

```typescript
switch (resource) {
  case 'project':
    // Handle project changes
    break;
  
  case 'sprint':
    // Handle sprint changes
    break;
  
  case 'task':
    // Handle task changes
    break;
  
  case 'project_members':
    // Handle member changes
    break;
  
  default:
    console.warn('Unknown resource type for cache invalidation:', resource);
}
```

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


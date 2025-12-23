# Cache Invalidation Payload Reference

## Backend → Frontend Resource Name Mapping

The backend normalizes database table names to singular resource names (except `project_members`):

| Database Table | Resource Name (in payload) | Frontend Case |
|---------------|---------------------------|--------------|
| `projects` | `project` | `'project'` |
| `sprints` | `sprint` | `'sprint'` |
| `tasks` | `task` | `'task'` |
| `project_members` | `project_members` | `'project_members'` |

**Important**: Always use the **singular** form in your frontend switch statements, except for `project_members` which stays plural.

---

## WebSocket Message Structure

### Complete Message Format

```typescript
interface WebSocketMessage {
  type: 'cache_invalidate' | 'reconnect';
  data: CacheInvalidationPayload;
  project_id?: string;  // Optional, for convenience
  resource?: string;    // Optional, duplicate of data.resource
  action?: string;      // Optional, duplicate of data.action
}

interface CacheInvalidationPayload {
  resource: 'project' | 'sprint' | 'task' | 'project_members';
  id: string;           // Record ID (UUID or composite key)
  action: 'INSERT' | 'UPDATE' | 'DELETE';
  project_id: string;  // UUID of the project
  timestamp: string;    // ISO 8601 format
}
```

---

## Example Payloads by Resource Type

### 1. Project (`project`)

**When**: Project is created, updated, or deleted

```json
{
  "type": "cache_invalidate",
  "data": {
    "resource": "project",
    "id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268",
    "action": "INSERT",
    "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268",
    "timestamp": "2025-01-20T15:30:00Z"
  },
  "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268"
}
```

**Frontend Handling**:
```typescript
case 'project':
  if (id) {
    queryClient.invalidateQueries({ queryKey: ['projects', id] });
  }
  queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
  break;
```

---

### 2. Sprint (`sprint`)

**When**: Sprint is created, updated, or deleted

```json
{
  "type": "cache_invalidate",
  "data": {
    "resource": "sprint",
    "id": "b8b8f1dg-b6e8-5b9c-b8g5-c6c1904e0379",
    "action": "INSERT",
    "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268",
    "timestamp": "2025-01-20T15:30:00Z"
  },
  "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268"
}
```

**Frontend Handling**:
```typescript
case 'sprint':
  if (id) {
    queryClient.invalidateQueries({ queryKey: ['sprints', id] });
  }
  queryClient.invalidateQueries({ queryKey: ['sprints', 'list', project_id] });
  queryClient.invalidateQueries({ queryKey: ['projects', 'bundle', project_id] });
  break;
```

---

### 3. Task (`task`)

**When**: Task is created, updated, or deleted

```json
{
  "type": "cache_invalidate",
  "data": {
    "resource": "task",
    "id": "c9c9g2eh-c7f9-6c0d-c9h6-d7d2015f1480",
    "action": "INSERT",
    "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268",
    "timestamp": "2025-01-20T15:30:00Z"
  },
  "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268"
}
```

**Frontend Handling**:
```typescript
case 'task':
  if (id) {
    queryClient.invalidateQueries({ queryKey: ['tasks', id] });
  }
  queryClient.invalidateQueries({ queryKey: ['tasks', 'list', project_id] });
  queryClient.invalidateQueries({ queryKey: ['tasks', 'sprint'] }); // If task belongs to sprint
  break;
```

---

### 4. Project Members (`project_members`)

**When**: Member joins, leaves, or role changes

```json
{
  "type": "cache_invalidate",
  "data": {
    "resource": "project_members",
    "id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268:cbc19378-0ca1-445f-902d-2fbd135b3ed4",
    "action": "INSERT",
    "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268",
    "timestamp": "2025-01-20T15:30:00Z"
  },
  "project_id": "a7a7e0cf-a5d7-4a8b-a7f4-b5b0893d9268"
}
```

**Note**: The `id` field is a composite key: `project_id:user_id` (no UUID column exists for `project_members`)

**Frontend Handling**:
```typescript
case 'project_members':
  // Immediately refetch to update UI
  queryClient.refetchQueries({ 
    queryKey: ['projectMembers', project_id],
    exact: false 
  });
  queryClient.refetchQueries({ 
    queryKey: ['projects', 'bundle', project_id],
    exact: false 
  });
  queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
  break;
```

---

## Action Types

| Action | Description | When It Fires |
|--------|-------------|---------------|
| `INSERT` | New record created | `CREATE` operations |
| `UPDATE` | Record modified | `UPDATE` operations |
| `DELETE` | Record removed | `DELETE` operations |

---

## Complete Frontend Implementation

### TypeScript Interfaces

```typescript
// Cache invalidation payload structure
interface CacheInvalidationPayload {
  resource: 'project' | 'sprint' | 'task' | 'project_members';
  id: string;  // UUID for projects/sprints/tasks, "project_id:user_id" for project_members
  action: 'INSERT' | 'UPDATE' | 'DELETE';
  project_id: string;  // Always present
  timestamp: string;   // ISO 8601 format
}

// WebSocket message wrapper
interface WebSocketMessage {
  type: 'cache_invalidate' | 'reconnect';
  data: CacheInvalidationPayload;
  project_id?: string;  // Convenience field, same as data.project_id
  resource?: string;    // Convenience field, same as data.resource
  action?: string;      // Convenience field, same as data.action
}
```

### Handler Function

```typescript
import { queryClient } from '../lib/queryClient';

function handleCacheInvalidation(message: WebSocketMessage) {
  const { data } = message;
  const { resource, id, action, project_id } = data;

  console.log(`Cache invalidation: ${resource} ${action}`, { id, project_id });

  switch (resource) {
    case 'project':
      // Invalidate specific project if ID provided
      if (id) {
        queryClient.invalidateQueries({ queryKey: ['projects', id] });
      }
      // Always invalidate project list
      queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
      break;

    case 'sprint':
      // Invalidate specific sprint if ID provided
      if (id) {
        queryClient.invalidateQueries({ queryKey: ['sprints', id] });
      }
      // Invalidate sprints list for the project
      queryClient.invalidateQueries({ queryKey: ['sprints', 'list', project_id] });
      // Invalidate project bundle (may include sprints)
      queryClient.invalidateQueries({ queryKey: ['projects', 'bundle', project_id] });
      break;

    case 'task':
      // Invalidate specific task if ID provided
      if (id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', id] });
      }
      // Invalidate tasks list for the project
      queryClient.invalidateQueries({ queryKey: ['tasks', 'list', project_id] });
      // Invalidate sprint tasks (if task belongs to a sprint)
      queryClient.invalidateQueries({ queryKey: ['tasks', 'sprint'] });
      break;

    case 'project_members':
      // For member changes, immediately refetch to update UI
      queryClient.refetchQueries({ 
        queryKey: ['projectMembers', project_id],
        exact: false // Refetch all queries that start with this key
      });
      queryClient.refetchQueries({ 
        queryKey: ['projects', 'bundle', project_id],
        exact: false 
      });
      // Invalidate project list (will refetch when user navigates to it)
      queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
      console.log(`✅ Refetched project members for project ${project_id}`);
      break;

    default:
      console.warn('Unknown resource type for cache invalidation:', resource);
  }
}
```

### WebSocket Message Handler

```typescript
function handleWebSocketMessage(event: MessageEvent) {
  try {
    const message: WebSocketMessage = JSON.parse(event.data);

    if (message.type === 'cache_invalidate') {
      handleCacheInvalidation(message);
    } else if (message.type === 'reconnect') {
      console.log('Reconnect requested:', message.data);
      // Handle reconnection logic
    }
  } catch (error) {
    console.error('Failed to parse WebSocket message:', error);
  }
}
```

---

## Common Mistakes to Avoid

### ❌ Wrong: Using Plural Resource Names

```typescript
// ❌ DON'T DO THIS
case 'tasks':  // Wrong - backend sends 'task' (singular)
case 'sprints': // Wrong - backend sends 'sprint' (singular)
case 'projects': // Wrong - backend sends 'project' (singular)
```

### ✅ Correct: Using Singular Resource Names

```typescript
// ✅ DO THIS
case 'task':     // Correct - matches backend
case 'sprint':   // Correct - matches backend
case 'project':  // Correct - matches backend
case 'project_members': // Correct - kept plural (exception)
```

---

## Query Key Conventions

For consistency, use these query key patterns:

| Resource | Query Key Pattern | Example |
|----------|------------------|---------|
| Projects | `['projects', id]` or `['projects', 'list']` | `['projects', 'a7a7e0cf...']` |
| Sprints | `['sprints', id]` or `['sprints', 'list', projectId]` | `['sprints', 'list', 'a7a7e0cf...']` |
| Tasks | `['tasks', id]` or `['tasks', 'list', projectId]` | `['tasks', 'list', 'a7a7e0cf...']` |
| Members | `['projectMembers', projectId]` | `['projectMembers', 'a7a7e0cf...']` |

---

## Testing

### Test Each Resource Type

1. **Create a project** → Should receive `resource: "project"`, `action: "INSERT"`
2. **Create a sprint** → Should receive `resource: "sprint"`, `action: "INSERT"`
3. **Create a task** → Should receive `resource: "task"`, `action: "INSERT"`
4. **Add a member** → Should receive `resource: "project_members"`, `action: "INSERT"`

### Verify Payload Structure

```typescript
// Add logging to verify structure
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('WebSocket message:', JSON.stringify(message, null, 2));
  
  if (message.type === 'cache_invalidate') {
    console.log('Resource:', message.data.resource);
    console.log('Action:', message.data.action);
    console.log('Project ID:', message.data.project_id);
  }
};
```

---

## Summary

- ✅ Use **singular** resource names: `'project'`, `'sprint'`, `'task'`
- ✅ Exception: `'project_members'` stays plural
- ✅ Always check `data.resource` (not `message.resource`)
- ✅ `project_id` is always present in payload
- ✅ `id` format: UUID for most resources, `"project_id:user_id"` for `project_members`
- ✅ Actions: `'INSERT'`, `'UPDATE'`, `'DELETE'`




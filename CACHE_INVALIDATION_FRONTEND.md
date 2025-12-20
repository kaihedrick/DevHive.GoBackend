# Real-Time Cache Invalidation - Frontend Integration Guide

This document describes how to integrate real-time cache invalidation into your React frontend application.

## Overview

The backend now supports real-time cache invalidation using PostgreSQL NOTIFY and WebSockets. When data changes in the database (projects, sprints, tasks, project_members), the backend automatically sends cache invalidation notifications to connected WebSocket clients.

**Important**: Resource names in cache invalidation notifications are normalized to singular (except `project_members`):
- `projects` table → `project` resource
- `sprints` table → `sprint` resource  
- `tasks` table → `task` resource
- `project_members` table → `project_members` resource (kept plural for consistency)

## Architecture

```
PostgreSQL Table Change
  ↓
PostgreSQL Trigger → NOTIFY 'cache_invalidate'
  ↓
Go NOTIFY Listener → WebSocket Hub
  ↓
Frontend WebSocket Client
  ↓
TanStack Query Cache Invalidation
```

## Frontend Implementation

### 1. WebSocket Connection Service

Create a service to manage WebSocket connections for cache invalidation:

**File: `src/services/cacheInvalidationService.ts`**

```typescript
import { queryClient } from '../lib/queryClient';

interface CacheInvalidationPayload {
  resource: 'project' | 'sprint' | 'task' | 'project_members';
  id?: string;
  action: 'INSERT' | 'UPDATE' | 'DELETE';
  project_id: string;
  timestamp: string;
}

interface WebSocketMessage {
  type: 'cache_invalidate' | 'reconnect';
  resource?: string;
  action?: string;
  data?: CacheInvalidationPayload | { reason: string };
  project_id?: string;
}

class CacheInvalidationService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isConnecting = false;

  connect(projectId: string, accessToken: string) {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    this.isConnecting = true;
    const wsUrl = `${process.env.REACT_APP_WS_URL || 'wss://devhive-go-backend.fly.dev'}/api/v1/messages/ws?project_id=${projectId}`;
    
    try {
      this.ws = new WebSocket(wsUrl, [], {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      } as any);

      this.ws.onopen = () => {
        console.log('Cache invalidation WebSocket connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
      };

      this.ws.onclose = () => {
        console.log('Cache invalidation WebSocket disconnected');
        this.isConnecting = false;
        this.ws = null;
        this.scheduleReconnect(projectId, accessToken);
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.isConnecting = false;
      this.scheduleReconnect(projectId, accessToken);
    }
  }

  private scheduleReconnect(projectId: string, accessToken: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached. Cache invalidation unavailable.');
      return;
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.reconnectDelay = Math.min(
        this.reconnectDelay * 2,
        this.maxReconnectDelay
      );
      console.log(`Reconnecting cache invalidation WebSocket (attempt ${this.reconnectAttempts})...`);
      this.connect(projectId, accessToken);
    }, this.reconnectDelay);
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case 'cache_invalidate':
        this.handleCacheInvalidation(message.data as CacheInvalidationPayload);
        break;
      case 'reconnect':
        // On reconnect, invalidate all caches and refetch
        console.log('Cache invalidation reconnected, invalidating all caches');
        queryClient.invalidateQueries();
        break;
      default:
        console.warn('Unknown WebSocket message type:', message.type);
    }
  }

  private handleCacheInvalidation(payload: CacheInvalidationPayload) {
    const { resource, id, action, project_id } = payload;

    console.log(`Cache invalidation: ${resource} ${action}`, { id, project_id });

    // Invalidate based on resource type
    switch (resource) {
      case 'project':
        if (id) {
          // Invalidate specific project
          queryClient.invalidateQueries({ queryKey: ['projects', 'detail', id] });
        }
        // Always invalidate project list
        queryClient.invalidateQueries({ queryKey: ['projects', 'list'] });
        break;

      case 'sprint':
        if (id) {
          queryClient.invalidateQueries({ queryKey: ['sprints', 'detail', id] });
        }
        // Invalidate sprints for the project
        queryClient.invalidateQueries({ queryKey: ['sprints', 'list', project_id] });
        // Also invalidate project bundle (which includes sprints)
        queryClient.invalidateQueries({ queryKey: ['projects', 'bundle', project_id] });
        break;

      case 'task':
        if (id) {
          queryClient.invalidateQueries({ queryKey: ['tasks', 'detail', id] });
        }
        // Invalidate tasks for the project
        queryClient.invalidateQueries({ queryKey: ['tasks', 'list', project_id] });
        // Invalidate sprint tasks if task belongs to a sprint
        queryClient.invalidateQueries({ queryKey: ['tasks', 'sprint'] });
        break;

      case 'project_members':
        // For member changes, immediately refetch active queries to update UI
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

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
    this.reconnectDelay = 1000;
  }
}

export const cacheInvalidationService = new CacheInvalidationService();
```

### 2. Update Query Client Configuration

Update your TanStack Query client to use `staleTime: Infinity` for indefinite cache persistence:

**File: `src/lib/queryClient.ts`**

```typescript
import { QueryClient } from '@tanstack/react-query';
import { persistQueryClient } from '@tanstack/react-query-persist-client';
import { createSyncStoragePersister } from '@tanstack/query-sync-storage-persister';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: Infinity, // Cache never goes stale - invalidate only via WebSocket
      gcTime: 24 * 60 * 60 * 1000, // 24 hours garbage collection
      refetchOnWindowFocus: false, // Don't refetch on window focus
      refetchOnReconnect: false, // Don't refetch on reconnect (WebSocket handles this)
      retry: 1,
    },
  },
});

// Optional: Persist cache to localStorage
if (typeof window !== 'undefined') {
  const persister = createSyncStoragePersister({
    storage: window.localStorage,
    key: 'REACT_QUERY_OFFLINE_CACHE',
  });

  persistQueryClient({
    queryClient,
    persister,
  });
}
```

### 3. Integrate with Auth Context

Connect the cache invalidation service to your authentication context:

**File: `src/contexts/AuthContext.tsx`** (additions)

```typescript
import { cacheInvalidationService } from '../services/cacheInvalidationService';
import { useEffect } from 'react';

// Inside your AuthContext component or hook:
useEffect(() => {
  if (isAuthenticated && user && currentProjectId) {
    // Connect to cache invalidation WebSocket when authenticated
    const accessToken = localStorage.getItem('accessToken'); // Or from your auth state
    if (accessToken) {
      cacheInvalidationService.connect(currentProjectId, accessToken);
    }

    return () => {
      // Disconnect on unmount or logout
      cacheInvalidationService.disconnect();
    };
  } else {
    // Disconnect if not authenticated
    cacheInvalidationService.disconnect();
  }
}, [isAuthenticated, user, currentProjectId]);
```

### 4. Update Query Hooks

Remove any `staleTime` overrides from your query hooks to use the `Infinity` default:

**File: `src/hooks/useProjects.ts`** (example)

```typescript
// Remove staleTime overrides - use Infinity default from queryClient
export const useProjects = () => {
  return useQuery({
    queryKey: ['projects', 'list'],
    queryFn: () => projectService.getProjects(),
    // staleTime removed - uses Infinity from queryClient
  });
};

export const useProject = (projectId: string) => {
  return useQuery({
    queryKey: ['projects', 'detail', projectId],
    queryFn: () => projectService.getProject(projectId),
    enabled: !!projectId,
    // staleTime removed - uses Infinity from queryClient
  });
};
```

## Fallback Strategy

If WebSocket is unavailable or disconnected:

1. **Cache Persistence**: Cache remains valid indefinitely until WebSocket reconnects
2. **Manual Refresh**: Users can manually refresh to get latest data
3. **Reconnection**: On reconnect, all caches are invalidated and refetched automatically
4. **Graceful Degradation**: Application continues to work normally, just without real-time updates

## Testing

1. **Test Cache Invalidation**:
   - Open your app in two browser tabs
   - Update a project in one tab
   - Verify the other tab automatically updates

2. **Test Reconnection**:
   - Disconnect network temporarily
   - Reconnect network
   - Verify WebSocket reconnects and caches are invalidated

3. **Test Multiple Resources**:
   - Update projects, sprints, tasks, and members
   - Verify each triggers appropriate cache invalidation

## Environment Variables

Add to your `.env`:

```env
REACT_APP_WS_URL=wss://devhive-go-backend.fly.dev
```

For local development:

```env
REACT_APP_WS_URL=ws://localhost:8080
```

## Security Notes

- WebSocket connections require JWT authentication via `Authorization` header
- Project access is validated on the backend before allowing WebSocket connection
- Only users with project access receive cache invalidation notifications for that project

## Troubleshooting

1. **WebSocket not connecting**:
   - Check browser console for connection errors
   - Verify JWT token is valid and not expired
   - Check CORS settings on backend

2. **Cache not invalidating**:
   - Verify WebSocket is connected (check browser DevTools → Network → WS)
   - Check browser console for WebSocket messages
   - Verify query keys match invalidation patterns

3. **Too many reconnections**:
   - Check network stability
   - Verify backend WebSocket endpoint is accessible
   - Check for firewall/proxy issues


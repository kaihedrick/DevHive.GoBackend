# WebSocket Connection Guide - Frontend Best Practices

## Problem: Members Disconnecting While Owners Stay Connected

This guide addresses common issues where project members can't connect to WebSockets while owners can, and provides solutions for proper connection handling.

---

## Root Causes & Solutions

### 1. Project List Not Including Member Projects

**Problem**: Backend only returns projects where user is owner, not member.

**Solution**: ✅ **FIXED** - Backend now returns all projects (owned + member projects) via `ListProjectsByUser` query.

**Frontend Check**:
```typescript
const { data: projects } = useProjects();

// After login, verify projects list includes member projects
console.log('User projects:', projects?.projects);
// Should include projects where user is owner OR member
```

---

### 2. Missing User Role & Permissions in Project List

**Problem**: Frontend can't validate if `selectedProjectId` belongs to current user.

**Solution**: ✅ **FIXED** - `ListProjects` now includes `userRole` and `permissions` for each project.

**Frontend Implementation**:
```typescript
import { useProjects } from './hooks/useProjects';

function useValidateProjectAccess() {
  const { data: projects } = useProjects();
  
  const validateProjectId = (projectId: string | null): boolean => {
    if (!projectId || !projects?.projects) return false;
    
    // Check if projectId exists in user's accessible projects
    return projects.projects.some(p => p.id === projectId);
  };
  
  const getProjectRole = (projectId: string | null): string | null => {
    if (!projectId || !projects?.projects) return null;
    
    const project = projects.projects.find(p => p.id === projectId);
    return project?.userRole || null;
  };
  
  return { validateProjectId, getProjectRole };
}
```

---

### 3. Stale selectedProjectId in localStorage

**Problem**: Switching accounts reuses old `projectId` that new user can't access.

**Solution**: Always validate `selectedProjectId` against current user's projects.

**Frontend Implementation**:
```typescript
import { useEffect } from 'react';
import { useProjects } from './hooks/useProjects';

function useProjectSelection() {
  const { data: projects, isLoading } = useProjects();
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  
  useEffect(() => {
    if (isLoading || !projects?.projects) return;
    
    // Get stored projectId
    const stored = localStorage.getItem('selectedProjectId');
    
    if (stored) {
      // Validate it belongs to current user
      const isValid = projects.projects.some(p => p.id === stored);
      
      if (!isValid) {
        // Stale projectId - clear it
        console.warn('Stored projectId is not accessible, clearing...');
        localStorage.removeItem('selectedProjectId');
        setSelectedProjectId(null);
        return;
      }
      
      setSelectedProjectId(stored);
    } else if (projects.projects.length === 1) {
      // Auto-select if only one project
      const projectId = projects.projects[0].id;
      setSelectedProjectId(projectId);
      localStorage.setItem('selectedProjectId', projectId);
    }
  }, [projects, isLoading]);
  
  return { selectedProjectId, setSelectedProjectId };
}
```

---

### 4. WebSocket Authentication Issues

**Problem**: WebSocket handshake fails with 401/403, but frontend treats it as "not authenticated".

**Solution**: Distinguish between 401 (unauthenticated) and 403 (unauthorized for project).

**Backend Error Messages**:
- **401 Unauthorized**: "Authentication token has expired. Please refresh your token and reconnect."
- **403 Forbidden**: "You are not a member of this project. Please select a project you have access to."

**Frontend Implementation**:
```typescript
function connectWebSocket(projectId: string, token: string): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(
      `wss://devhive-go-backend.fly.dev/api/v1/messages/ws?project_id=${projectId}`,
      [],
      {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      }
    );
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      resolve(ws);
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      reject(new Error('WebSocket connection failed'));
    };
    
    ws.onclose = (event) => {
      const { code, reason } = event;
      
      if (code === 1008) {
        // Policy violation - likely 403 Forbidden
        const message = reason || 'Access denied';
        if (message.includes('not a member')) {
          // User is authenticated but not authorized for this project
          console.warn('Not authorized for project:', projectId);
          // Clear selectedProjectId and show message
          localStorage.removeItem('selectedProjectId');
          reject(new Error('NOT_AUTHORIZED_FOR_PROJECT'));
        } else {
          reject(new Error('WebSocket closed: ' + message));
        }
      } else if (code === 1002) {
        // Protocol error - might be expired token
        console.warn('WebSocket protocol error, token may be expired');
        reject(new Error('TOKEN_EXPIRED'));
      } else if (code === 1006) {
        // Abnormal closure - connection lost
        console.warn('WebSocket connection lost');
        reject(new Error('CONNECTION_LOST'));
      } else {
        reject(new Error(`WebSocket closed: ${code} ${reason}`));
      }
    };
  });
}
```

---

### 5. Auto-Select Project After Login

**Best Practice**: Auto-select first project if user has exactly one project.

**Frontend Implementation**:
```typescript
function useAutoSelectProject() {
  const { data: projects, isLoading } = useProjects();
  const { validateProjectId, setSelectedProjectId } = useProjectSelection();
  
  useEffect(() => {
    if (isLoading || !projects?.projects) return;
    
    const stored = localStorage.getItem('selectedProjectId');
    
    // If stored projectId is valid, use it
    if (stored && validateProjectId(stored)) {
      setSelectedProjectId(stored);
      return;
    }
    
    // If exactly one project, auto-select it
    if (projects.projects.length === 1) {
      const projectId = projects.projects[0].id;
      setSelectedProjectId(projectId);
      localStorage.setItem('selectedProjectId', projectId);
      console.log('Auto-selected project:', projectId);
    } else if (projects.projects.length > 1) {
      // Multiple projects - let user choose
      // Or auto-select most recent (first in list)
      const projectId = projects.projects[0].id;
      setSelectedProjectId(projectId);
      localStorage.setItem('selectedProjectId', projectId);
    }
  }, [projects, isLoading, validateProjectId, setSelectedProjectId]);
}
```

---

## Complete WebSocket Manager Example

```typescript
import { useEffect, useRef, useState } from 'react';
import { useProjectSelection } from './useProjectSelection';
import { ensureValidToken } from './useAuth';

class WebSocketManager {
  private ws: WebSocket | null = null;
  private projectId: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  
  async connect(projectId: string): Promise<void> {
    // Disconnect existing connection if different project
    if (this.ws && this.projectId !== projectId) {
      this.disconnect();
    }
    
    // Don't reconnect if already connected to same project
    if (this.ws && this.projectId === projectId && this.ws.readyState === WebSocket.OPEN) {
      return;
    }
    
    try {
      // Ensure token is valid before connecting
      const token = await ensureValidToken();
      
      const wsUrl = `wss://devhive-go-backend.fly.dev/api/v1/messages/ws?project_id=${projectId}`;
      
      this.ws = new WebSocket(wsUrl, [], {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      this.projectId = projectId;
      
      this.ws.onopen = () => {
        console.log(`✅ WebSocket connected to project: ${projectId}`);
        this.reconnectAttempts = 0; // Reset on successful connection
      };
      
      this.ws.onclose = async (event) => {
        console.log(`WebSocket closed: ${event.code} ${event.reason}`);
        
        // Handle different close codes
        if (event.code === 1008) {
          // Policy violation - likely 403 Forbidden
          if (event.reason?.includes('not a member')) {
            console.error('❌ Not authorized for project:', projectId);
            // Don't reconnect - user needs to select different project
            this.projectId = null;
            return;
          }
        }
        
        // Attempt to reconnect (unless it was a 403)
        if (this.reconnectAttempts < this.maxReconnectAttempts && this.projectId) {
          this.reconnectAttempts++;
          const delay = Math.min(1000 * this.reconnectAttempts, 10000); // Max 10 seconds
          console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})...`);
          
          this.reconnectTimeout = setTimeout(() => {
            if (this.projectId) {
              this.connect(this.projectId);
            }
          }, delay);
        } else {
          console.error('Max reconnection attempts reached');
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
      
      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      };
      
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      throw error;
    }
  }
  
  disconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.projectId = null;
    this.reconnectAttempts = 0;
  }
  
  private handleMessage(message: any): void {
    // Handle cache invalidation messages
    if (message.type === 'cache_invalidate') {
      // Your cache invalidation logic
      console.log('Cache invalidation:', message.data);
    }
  }
  
  send(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }
}

// React hook
export function useWebSocketConnection() {
  const managerRef = useRef<WebSocketManager | null>(null);
  const { selectedProjectId, validateProjectId } = useProjectSelection();
  
  useEffect(() => {
    if (!managerRef.current) {
      managerRef.current = new WebSocketManager();
    }
    
    const manager = managerRef.current;
    
    if (selectedProjectId && validateProjectId(selectedProjectId)) {
      manager.connect(selectedProjectId).catch((error) => {
        if (error.message === 'NOT_AUTHORIZED_FOR_PROJECT') {
          // Clear invalid projectId
          localStorage.removeItem('selectedProjectId');
        } else if (error.message === 'TOKEN_EXPIRED') {
          // Token expired - will auto-refresh on next attempt
          console.log('Token expired, will retry after refresh');
        }
      });
    } else {
      manager.disconnect();
    }
    
    return () => {
      // Don't disconnect on unmount - keep connection alive
      // manager.disconnect();
    };
  }, [selectedProjectId, validateProjectId]);
  
  return managerRef.current;
}
```

---

## Error Handling Checklist

### ✅ 401 Unauthorized (Token Expired)
- **Action**: Refresh token and reconnect
- **Don't**: Log user out or clear auth state

### ✅ 403 Forbidden (Not Authorized for Project)
- **Action**: Clear `selectedProjectId`, show message, let user select different project
- **Don't**: Log user out or treat as authentication failure

### ✅ 404 Not Found (Project Doesn't Exist)
- **Action**: Clear `selectedProjectId`, show error message
- **Don't**: Attempt to reconnect

### ✅ 1006 Abnormal Closure (Connection Lost)
- **Action**: Attempt to reconnect with exponential backoff
- **Don't**: Immediately retry (may cause connection spam)

---

## Testing Checklist

1. ✅ **Member Login**: Verify projects list includes member projects
2. ✅ **Project Selection**: Verify `selectedProjectId` is validated against user's projects
3. ✅ **WebSocket Connection**: Verify members can connect to their projects
4. ✅ **Token Expiry**: Verify automatic refresh and reconnection
5. ✅ **Invalid ProjectId**: Verify graceful handling (clear selection, show message)
6. ✅ **Account Switch**: Verify stale `projectId` is cleared when switching accounts

---

## Backend Changes Made

1. ✅ **ListProjects** now includes `userRole` and `permissions` for each project
2. ✅ **WebSocket handler** provides clear error messages distinguishing 401 vs 403
3. ✅ **CheckProjectAccess** uses canonical model (only checks `project_members`)

---

## Frontend Next Steps

1. Copy `useProjectSelection` hook to validate project access
2. Update WebSocket connection logic to handle 401 vs 403 differently
3. Implement auto-project selection after login
4. Add validation to clear stale `selectedProjectId` on account switch
5. Use `userRole` and `permissions` from project list to conditionally show UI

See `WEBSOCKET_TOKEN_REFRESH_GUIDE.md` for token refresh implementation.





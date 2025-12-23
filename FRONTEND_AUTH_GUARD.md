# Frontend Authentication Guard Guide

## Problem

Your frontend is fetching projects even when the user is logged out. This happens because:

1. **React Query auto-fetches** on component mount
2. **No auth state check** before making API calls
3. **Stale tokens** in localStorage triggering requests

## Solution: Guard API Calls with Auth State

### 1. Check Auth State Before Fetching

Your React Query hooks should only run when the user is authenticated:

```typescript
// useProjects.ts or useProjects.js
import { useQuery } from '@tanstack/react-query';
import { useAuth } from './useAuth'; // Your auth hook
import { fetchUserProjects } from './projectService';

export function useProjects() {
  const { isAuthenticated, user } = useAuth();

  return useQuery({
    queryKey: ['projects'],
    queryFn: fetchUserProjects,
    enabled: isAuthenticated && !!user, // ✅ Only fetch when authenticated
    staleTime: 2 * 60 * 1000, // 2 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on 401 - token refresh should handle it
      if (error?.response?.status === 401) {
        return false;
      }
      return failureCount < 3;
    },
  });
}
```

### 2. Auth Hook Implementation

Your auth hook should properly track authentication state:

```typescript
// useAuth.ts or useAuth.js
import { useState, useEffect } from 'react';

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated
    const token = localStorage.getItem('access_token');
    
    if (!token) {
      setIsAuthenticated(false);
      setUser(null);
      setIsLoading(false);
      return;
    }

    // Optional: Validate token format (basic check)
    try {
      // Decode token to check expiration (without verifying signature)
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000);
      
      // If token is expired, clear auth state
      if (payload.exp < now) {
        localStorage.removeItem('access_token');
        setIsAuthenticated(false);
        setUser(null);
        setIsLoading(false);
        return;
      }

      // Token exists and is not expired
      setIsAuthenticated(true);
      // Optionally fetch user data here
    } catch (error) {
      // Invalid token format
      localStorage.removeItem('access_token');
      setIsAuthenticated(false);
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = () => {
    localStorage.removeItem('access_token');
    setIsAuthenticated(false);
    setUser(null);
  };

  return {
    isAuthenticated,
    user,
    isLoading,
    logout,
  };
}
```

### 3. Guard Components with Auth Check

Components that fetch data should check auth state:

```typescript
// ProjectsPage.tsx or ProjectsPage.jsx
import { useProjects } from './hooks/useProjects';
import { useAuth } from './hooks/useAuth';
import { Navigate } from 'react-router-dom';

export function ProjectsPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { data: projects, isLoading: projectsLoading } = useProjects();

  // Show loading while checking auth
  if (authLoading) {
    return <div>Loading...</div>;
  }

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Show loading while fetching projects
  if (projectsLoading) {
    return <div>Loading projects...</div>;
  }

  return (
    <div>
      <h1>My Projects</h1>
      {projects?.map(project => (
        <div key={project.id}>{project.name}</div>
      ))}
    </div>
  );
}
```

### 4. Update API Client to Handle 401 Properly

Your API client should clear auth state on 401 (when refresh fails):

```typescript
// apiClient.ts
import axios, { AxiosError } from 'axios';

// ... existing code ...

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;

    // Skip refresh for auth routes
    if (isAuthRoute(originalRequest?.url)) {
      return Promise.reject(error);
    }

    // Handle 401 errors
    if (error.response?.status === 401) {
      // If this is already a retry after refresh, clear auth
      if (originalRequest._retry) {
        // Refresh failed - user needs to log in again
        localStorage.removeItem('access_token');
        // Optionally trigger auth state update
        window.dispatchEvent(new Event('auth-state-changed'));
        return Promise.reject(error);
      }

      // Attempt token refresh
      try {
        const response = await axios.post(
          `${API_BASE_URL}/auth/refresh`,
          {},
          { withCredentials: true }
        );

        const { token } = response.data;
        localStorage.setItem('access_token', token);
        originalRequest.headers.Authorization = `Bearer ${token}`;
        originalRequest._retry = true;

        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed - clear auth
        localStorage.removeItem('access_token');
        window.dispatchEvent(new Event('auth-state-changed'));
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### 5. Listen for Auth State Changes

Update your auth hook to listen for auth state changes:

```typescript
// useAuth.ts
import { useState, useEffect } from 'react';

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  const checkAuth = () => {
    const token = localStorage.getItem('access_token');
    
    if (!token) {
      setIsAuthenticated(false);
      setUser(null);
      setIsLoading(false);
      return;
    }

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000);
      
      if (payload.exp < now) {
        localStorage.removeItem('access_token');
        setIsAuthenticated(false);
        setUser(null);
        setIsLoading(false);
        return;
      }

      setIsAuthenticated(true);
    } catch (error) {
      localStorage.removeItem('access_token');
      setIsAuthenticated(false);
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    // Check auth on mount
    checkAuth();

    // Listen for auth state changes (from API client)
    const handleAuthChange = () => {
      checkAuth();
    };

    window.addEventListener('auth-state-changed', handleAuthChange);

    return () => {
      window.removeEventListener('auth-state-changed', handleAuthChange);
    };
  }, []);

  const logout = () => {
    localStorage.removeItem('access_token');
    setIsAuthenticated(false);
    setUser(null);
    window.dispatchEvent(new Event('auth-state-changed'));
  };

  return {
    isAuthenticated,
    user,
    isLoading,
    logout,
  };
}
```

### 6. Clear Queries on Logout

When user logs out, clear all cached queries:

```typescript
// In your logout function or component
import { useQueryClient } from '@tanstack/react-query';

function LogoutButton() {
  const queryClient = useQueryClient();
  const { logout } = useAuth();

  const handleLogout = () => {
    logout();
    // Clear all cached queries
    queryClient.clear();
    // Optionally redirect to login
    window.location.href = '/login';
  };

  return <button onClick={handleLogout}>Logout</button>;
}
```

## Complete Example: Protected Projects Component

```typescript
// ProjectsPage.tsx
import { useQuery } from '@tanstack/react-query';
import { useAuth } from './hooks/useAuth';
import { Navigate } from 'react-router-dom';
import { fetchUserProjects } from './services/projectService';

export function ProjectsPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  // Only fetch when authenticated
  const { data, isLoading, error } = useQuery({
    queryKey: ['projects'],
    queryFn: fetchUserProjects,
    enabled: isAuthenticated, // ✅ Key fix: only fetch when authenticated
    staleTime: 2 * 60 * 1000,
  });

  // Show loading while checking auth
  if (authLoading) {
    return <div>Checking authentication...</div>;
  }

  // Redirect if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Show loading while fetching
  if (isLoading) {
    return <div>Loading projects...</div>;
  }

  // Show error
  if (error) {
    return <div>Error loading projects: {error.message}</div>;
  }

  // Render projects
  return (
    <div>
      <h1>My Projects</h1>
      {data?.projects?.length === 0 ? (
        <p>No projects found</p>
      ) : (
        <ul>
          {data?.projects?.map(project => (
            <li key={project.id}>{project.name}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

## Key Points

1. **Always check `isAuthenticated`** before enabling queries
2. **Use `enabled` option** in React Query to conditionally fetch
3. **Clear auth state** when tokens are removed
4. **Listen for auth changes** to update UI reactively
5. **Clear queries on logout** to prevent stale data
6. **Redirect to login** when not authenticated

## Testing

1. **Logout** and navigate to projects page
   - Should redirect to login
   - No API calls should be made

2. **Login** and navigate to projects page
   - Should fetch projects
   - Should display projects

3. **Token expires** while on projects page
   - Should attempt refresh
   - If refresh fails, should redirect to login
   - No more API calls should be made

## Common Issues

### Issue: Queries still run after logout

**Fix**: Make sure `enabled: isAuthenticated` is set in all queries, and clear queries on logout.

### Issue: Stale auth state

**Fix**: Listen for `auth-state-changed` event and update state accordingly.

### Issue: Multiple redirects

**Fix**: Use `replace` in Navigate component: `<Navigate to="/login" replace />`




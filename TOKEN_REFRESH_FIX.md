# Token Refresh Fix for Frontend

## Problem

Your frontend is getting "Token expired" errors when fetching projects because:

1. The API client isn't automatically refreshing expired access tokens
2. **CRITICAL**: Your frontend has a "24-hour window" check that's clearing auth state instead of attempting refresh

## Backend Token System

- **Access Token**: Expires after 15 minutes
- **Refresh Token**: Stored in httpOnly cookie, expires after **7 days** (not 24 hours!)
- **Refresh Endpoint**: `POST /api/v1/auth/refresh`
  - Automatically reads `refresh_token` from cookies
  - Returns new access token: `{ "token": "...", "userID": "..." }`
  - No request body needed

## ⚠️ Important: Remove 24-Hour Window Check

**Your frontend should NEVER check token age before attempting refresh.** The refresh token is valid for 7 days, so even if the access token is old, you should always attempt refresh on 401 errors. The backend will tell you if the refresh token is expired.

**Remove any code like this:**
```typescript
// ❌ WRONG - Don't do this
if (tokenAge > 24 * 60 * 60 * 1000) {
  console.log('⚠️ Token expired beyond 24-hour window, clearing auth state');
  clearAuth();
  return;
}
```

**Instead, always attempt refresh on 401:**
```typescript
// ✅ CORRECT - Always try refresh on 401
if (error.response?.status === 401) {
  // Attempt refresh regardless of token age
  await refreshToken();
}
```

## Frontend Fix: API Client with Token Refresh

### Complete API Client Implementation

```typescript
// apiClient.ts
import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://devhive-go-backend.fly.dev/api/v1';

// Routes that should NOT have auth headers or token refresh
const AUTH_ROUTES = ['/auth/login', '/auth/register', '/auth/refresh', '/auth/logout'];
const PUBLIC_ROUTES = ['/users/validate-email', '/users/validate-username'];

// Token storage helpers
function getAccessToken(): string | null {
  return localStorage.getItem('access_token');
}

function setAccessToken(token: string): void {
  localStorage.setItem('access_token', token);
}

function clearAuth(): void {
  localStorage.removeItem('access_token');
  // Optionally clear other auth-related data
}

// Check if route should skip auth
function isAuthRoute(url: string | undefined): boolean {
  if (!url) return false;
  return AUTH_ROUTES.some(route => url.includes(route));
}

function isPublicRoute(url: string | undefined): boolean {
  if (!url) return false;
  return PUBLIC_ROUTES.some(route => url.includes(route));
}

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Important: sends cookies (for refresh_token)
});

// Track if we're currently refreshing to avoid multiple refresh calls
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: any) => void;
  reject: (error?: any) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  
  failedQueue = [];
};

// Request interceptor: Add auth token to requests
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Skip auth for auth routes and public routes
    if (isAuthRoute(config.url) || isPublicRoute(config.url)) {
      return config;
    }

    // Add access token to Authorization header
    const token = getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor: Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Skip refresh logic for auth routes
    if (isAuthRoute(originalRequest?.url)) {
      return Promise.reject(error);
    }

    // Skip refresh logic for public routes
    if (isPublicRoute(originalRequest?.url)) {
      return Promise.reject(error);
    }

    // If error is not 401, or request was already retried, reject
    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }

    // If we're already refreshing, queue this request
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      })
        .then((token) => {
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
          }
          return apiClient(originalRequest);
        })
        .catch((err) => {
          return Promise.reject(err);
        });
    }

    // Mark that we're refreshing
    originalRequest._retry = true;
    isRefreshing = true;

    try {
      // Call refresh endpoint (uses refresh_token cookie automatically)
      const response = await axios.post(
        `${API_BASE_URL}/auth/refresh`,
        {},
        {
          withCredentials: true, // Important: sends cookies
        }
      );

      const { token: newToken } = response.data;

      // Store new access token
      setAccessToken(newToken);

      // Process queued requests
      processQueue(null, newToken);

      // Update original request with new token
      if (originalRequest.headers) {
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
      }

      // Retry original request
      return apiClient(originalRequest);
    } catch (refreshError) {
      // Refresh failed - user needs to log in again
      processQueue(refreshError as AxiosError, null);
      clearAuth();
      
      // Redirect to login or handle re-authentication
      // window.location.href = '/login';
      
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);

export default apiClient;
```

### Usage in Services

```typescript
// projectService.js or projectService.ts
import apiClient from './apiClient';

export async function fetchUserProjects() {
  try {
    const response = await apiClient.get('/projects');
    return response.data;
  } catch (error) {
    console.error('❌ Error fetching user projects:', error);
    throw error;
  }
}
```

### React Hook Example

```typescript
// useProjects.ts or useProjects.js
import { useQuery } from '@tanstack/react-query';
import { fetchUserProjects } from './projectService';

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: fetchUserProjects,
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

## 🔴 CRITICAL FIX: Remove 24-Hour Window Check

If you're seeing this error:
```
⚠️ Token expired beyond 24-hour window, clearing auth state
```

**You MUST remove any code that checks token age before attempting refresh.** Here's what to look for and remove:

### ❌ Remove This Code

```typescript
// ❌ WRONG - Remove this entire block
const tokenAge = Date.now() - tokenTimestamp;
if (tokenAge > 24 * 60 * 60 * 1000) { // 24 hours
  console.log('⚠️ Token expired beyond 24-hour window, clearing auth state');
  clearAuth();
  throw new Error('Token expired');
}
```

### ✅ Correct Behavior

**Always attempt refresh on 401, regardless of token age.** The refresh token is valid for 7 days, so even if your access token is days old, you should still try to refresh it. The backend will tell you if the refresh token is expired.

```typescript
// ✅ CORRECT - Always try refresh on 401
if (error.response?.status === 401) {
  // Don't check token age - just attempt refresh
  // Backend will tell us if refresh token is expired
  try {
    await refreshToken();
    // Retry original request
  } catch (refreshError) {
    // Only clear auth if refresh actually fails
    clearAuth();
    throw refreshError;
  }
}
```

### Why This Matters

- **Access tokens expire in 15 minutes** (short-lived for security)
- **Refresh tokens expire in 7 days** (long-lived for convenience)
- Even if your access token is 6 days old, the refresh token might still be valid
- Only the backend knows if the refresh token is expired - let it tell you!

## Key Points

1. **`withCredentials: true`**: Required to send cookies (refresh_token) with requests
2. **Queue failed requests**: When refreshing, queue all failed requests and retry them after refresh
3. **Skip auth routes**: Don't add auth headers or refresh tokens for login/register/refresh endpoints
4. **Single refresh**: Use `isRefreshing` flag to prevent multiple simultaneous refresh calls
5. **Clear auth on refresh failure**: If refresh fails, clear tokens and redirect to login
6. **NEVER check token age**: Always attempt refresh on 401 - let the backend decide if refresh is possible

## Testing

1. **Login** and get an access token
2. **Wait 16 minutes** (or manually expire token)
3. **Make an API call** (e.g., fetch projects)
4. **Expected behavior**:
   - API call gets 401
   - Client automatically calls `/auth/refresh`
   - Gets new access token
   - Retries original request
   - Request succeeds

## Error Handling

### Refresh Token Expired (7 days)

If the refresh token is also expired, the refresh endpoint will return 401. In this case:

```typescript
catch (refreshError) {
  // Refresh failed - user needs to log in again
  clearAuth();
  window.location.href = '/login';
  return Promise.reject(refreshError);
}
```

### Network Errors

Handle network errors separately:

```typescript
if (error.code === 'ERR_NETWORK') {
  // Show network error message
  console.error('Network error - please check your connection');
}
```

## Alternative: Proactive Token Refresh

Instead of waiting for 401, you can proactively refresh tokens before they expire:

```typescript
import { jwtDecode } from 'jwt-decode';

function isTokenExpired(token: string): boolean {
  try {
    const decoded = jwtDecode<{ exp: number }>(token);
    const now = Math.floor(Date.now() / 1000);
    // Consider expired if it expires within 30 seconds
    return decoded.exp < (now + 30);
  } catch {
    return true;
  }
}

// In request interceptor
apiClient.interceptors.request.use(async (config) => {
  const token = getAccessToken();
  
  if (token && isTokenExpired(token)) {
    // Proactively refresh before making request
    try {
      const response = await axios.post(
        `${API_BASE_URL}/auth/refresh`,
        {},
        { withCredentials: true }
      );
      setAccessToken(response.data.token);
      config.headers.Authorization = `Bearer ${response.data.token}`;
    } catch (error) {
      clearAuth();
      window.location.href = '/login';
      return Promise.reject(error);
    }
  } else if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  return config;
});
```

## Common Issues

### Issue: "Token expired beyond 24-hour window"

**Cause**: Frontend code is checking token age and clearing auth before attempting refresh

**Fix**: Remove any code that checks token age. Always attempt refresh on 401 errors. The refresh token is valid for 7 days, so even old access tokens can be refreshed.

**Find and remove code like:**
```typescript
// Search for this pattern in your codebase:
if (tokenAge > 24 * 60 * 60 * 1000) {
  // Remove this entire block
}
```

### Issue: "Refresh token not found"

**Cause**: Cookies aren't being sent (missing `withCredentials: true`)

**Fix**: Ensure `withCredentials: true` is set on both the axios instance and the refresh call.

### Issue: CORS errors

**Cause**: Backend needs to allow credentials in CORS

**Backend fix** (if needed):
```go
// In CORS middleware
w.Header().Set("Access-Control-Allow-Credentials", "true")
```

### Issue: Multiple refresh calls

**Cause**: Multiple requests failing simultaneously

**Fix**: Use the `isRefreshing` flag and queue system (already in the code above).

### Issue: Infinite refresh loop

**Cause**: Refresh endpoint itself returns 401

**Fix**: Skip refresh logic for `/auth/refresh` route (already handled in code above).


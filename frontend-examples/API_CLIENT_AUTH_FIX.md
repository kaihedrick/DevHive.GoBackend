# API Client Auth Route Fix

## Problem

The API client was incorrectly handling authentication for auth routes (`/auth/login`, `/auth/register`, etc.):

1. **Adding Authorization headers** to auth routes (they should be unauthenticated)
2. **Attempting token refresh** when auth routes return 401 (which is expected and correct)
3. **Clearing auth state** when auth routes fail (treating it as token expiration)

This caused:
- Login/register requests to fail with 401 (backend correctly rejects authenticated requests to these endpoints)
- Interceptor to mislabel this as "token expired" and try to refresh
- Auth state to be cleared incorrectly
- Redirect loops or incorrect error messages

## Fix Applied

### 1. Added Auth Route Detection

```typescript
// Auth routes that should NEVER have auth headers or token refresh
const AUTH_ROUTES = [
  '/auth/login',
  '/auth/register',
  '/auth/refresh',
  '/auth/logout',
  '/auth/password/reset-request',
  '/auth/password/reset',
];

// Check if a URL is an auth route
function isAuthRoute(url: string | undefined): boolean {
  if (!url) return false;
  return AUTH_ROUTES.some(route => url.includes(route));
}
```

### 2. Skip Auth Headers for Auth Routes

**Before**:
```typescript
apiClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`; // ❌ Added to ALL routes
  }
  return config;
});
```

**After**:
```typescript
apiClient.interceptors.request.use((config) => {
  // Skip auth handling for auth routes
  if (isAuthRoute(config.url)) {
    return config; // ✅ No auth header for auth routes
  }

  const token = getAccessToken();
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### 3. Skip Token Refresh for Auth Routes

**Before**:
```typescript
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401 && !originalRequest._retry) {
      // ❌ Always tries to refresh, even for /auth/login
      await refreshToken();
    }
  }
);
```

**After**:
```typescript
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    // Skip token refresh for auth routes - they should return 401 directly
    if (isAuthRoute(originalRequest.url)) {
      return Promise.reject(error); // ✅ Return error directly, no refresh
    }

    if (error.response?.status === 401 && !originalRequest._retry) {
      // Only refresh for non-auth routes
      await refreshToken();
    }
  }
);
```

## Behavior After Fix

### Auth Routes (Login, Register, etc.)

✅ **No Authorization header** attached  
✅ **No token refresh** attempted on 401  
✅ **Error returned directly** to caller  
✅ **No auth state clearing**  

### Protected Routes (Projects, Users, etc.)

✅ **Authorization header** attached if token exists  
✅ **Token refresh** attempted on 401  
✅ **Auth state cleared** if refresh fails  
✅ **Redirect to login** if refresh fails  

## Testing

### Test Login Flow

1. **User is on login page** (not authenticated)
2. **User enters credentials and submits**
3. **Request to `/auth/login`**:
   - ✅ No `Authorization` header attached
   - ✅ Backend accepts request (no 401 from auth middleware)
   - ✅ Login succeeds
   - ✅ Token stored

### Test Register Flow

1. **User is on registration page** (not authenticated)
2. **User submits registration form**
3. **Request to `/auth/register`** (or `/users`):
   - ✅ No `Authorization` header attached
   - ✅ Registration succeeds

### Test Token Expiration

1. **User is authenticated** (has token)
2. **Token expires** (15 minutes)
3. **User makes request to `/api/v1/projects`**:
   - ✅ Request includes `Authorization: Bearer {expired_token}`
   - ✅ Backend returns 401
   - ✅ Interceptor attempts refresh
   - ✅ New token obtained
   - ✅ Original request retried with new token

### Test Invalid Login

1. **User enters wrong password**
2. **Request to `/auth/login`**:
   - ✅ No `Authorization` header attached
   - ✅ Backend returns 401 (invalid credentials)
   - ✅ **No token refresh attempted** (correct!)
   - ✅ Error returned to login form
   - ✅ User sees "Invalid credentials" message

## Summary

✅ **Fixed**: Auth routes no longer get Authorization headers  
✅ **Fixed**: Auth routes no longer trigger token refresh on 401  
✅ **Fixed**: Auth routes return errors directly to caller  
✅ **Result**: Login/register work correctly, token refresh only happens for protected routes




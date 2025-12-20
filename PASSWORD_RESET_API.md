# Account & Authentication API - Complete Frontend Guide

## Table of Contents
1. [Password Reset](#password-reset)
2. [User Account Endpoints](#user-account-endpoints)
3. [Authentication Endpoints](#authentication-endpoints)
4. [Common Errors (400/405)](#common-errors)
5. [Caching Guide](#caching-guide)

---

# Password Reset

## Endpoints

### 1. Request Password Reset
**Endpoint**: `POST /api/v1/auth/password/reset-request`

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Reset token created",
  "token": "abc123..." // ⚠️ REMOVE IN PRODUCTION - only for development
}
```

**Note**: In production, the `token` field should be removed. The backend will send an email instead.

**Error Responses**:

- **400 Bad Request** (Invalid JSON):
```json
{
  "type": "invalid_json",
  "title": "Bad Request",
  "status": 400,
  "detail": "error message here"
}
```

- **500 Internal Server Error** (Failed to create token):
```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "Failed to create reset token"
}
```

**Security Note**: If the email doesn't exist, the backend still returns 200 OK with a generic message to prevent email enumeration:
```json
{
  "message": "If the email exists, a reset link has been sent"
}
```

---

### 2. Reset Password
**Endpoint**: `POST /api/v1/auth/password/reset`

**Request Body**:
```json
{
  "token": "abc123...",
  "password": "newSecurePassword123"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password updated successfully"
}
```

**Error Responses**:

- **400 Bad Request** (Invalid JSON):
```json
{
  "type": "invalid_json",
  "title": "Bad Request",
  "status": 400,
  "detail": "error message here"
}
```

- **400 Bad Request** (Invalid/Expired Token):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid or expired reset token"
}
```

OR

```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Reset token has expired"
}
```

- **500 Internal Server Error** (Failed to hash/update password):
```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "Failed to hash password"
}
```

OR

```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "Failed to update password"
}
```

---

## Frontend Implementation Example

### TypeScript Interfaces

```typescript
// Request interfaces
interface PasswordResetRequest {
  email: string;
}

interface PasswordReset {
  token: string;
  password: string;
}

// Success response interfaces
interface PasswordResetRequestResponse {
  message: string;
  token?: string; // Only in development
}

interface PasswordResetResponse {
  message: string;
}

// Error response interface (RFC 7807 Problem Details)
interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
}
```

### React Hook Example

```typescript
import { useMutation } from '@tanstack/react-query';
import { apiClient } from './apiClient';

// Request password reset
export const useRequestPasswordReset = () => {
  return useMutation({
    mutationFn: async (email: string) => {
      const response = await apiClient.post<PasswordResetRequestResponse>(
        '/auth/password/reset-request',
        { email }
      );
      return response.data;
    },
    onSuccess: (data) => {
      // Show success message
      // In production, don't show the token
      console.log('Password reset email sent:', data.message);
    },
    onError: (error: any) => {
      // Handle error
      const problem = error.response?.data as ProblemDetails;
      console.error('Password reset request failed:', problem.detail);
    },
  });
};

// Reset password with token
export const useResetPassword = () => {
  return useMutation({
    mutationFn: async ({ token, password }: PasswordReset) => {
      const response = await apiClient.post<PasswordResetResponse>(
        '/auth/password/reset',
        { token, password }
      );
      return response.data;
    },
    onSuccess: (data) => {
      // Show success message and redirect to login
      console.log('Password reset successful:', data.message);
      // Redirect to login page
    },
    onError: (error: any) => {
      // Handle error
      const problem = error.response?.data as ProblemDetails;
      
      if (problem.status === 400) {
        // Invalid or expired token
        console.error('Reset failed:', problem.detail);
        // Show error message, maybe redirect to request new reset
      } else {
        // Server error
        console.error('Server error:', problem.detail);
      }
    },
  });
};
```

### Usage Example

```typescript
function PasswordResetRequestForm() {
  const requestReset = useRequestPasswordReset();
  
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const email = formData.get('email') as string;
    
    requestReset.mutate(email);
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input name="email" type="email" required />
      <button type="submit" disabled={requestReset.isPending}>
        {requestReset.isPending ? 'Sending...' : 'Send Reset Link'}
      </button>
      {requestReset.isSuccess && (
        <p>{requestReset.data.message}</p>
      )}
      {requestReset.isError && (
        <p>Error: {requestReset.error.response?.data.detail}</p>
      )}
    </form>
  );
}

function PasswordResetForm({ token }: { token: string }) {
  const resetPassword = useResetPassword();
  
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const password = formData.get('password') as string;
    
    resetPassword.mutate({ token, password });
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input name="password" type="password" required />
      <button type="submit" disabled={resetPassword.isPending}>
        {resetPassword.isPending ? 'Resetting...' : 'Reset Password'}
      </button>
      {resetPassword.isSuccess && (
        <p>Password reset successful! Redirecting to login...</p>
      )}
      {resetPassword.isError && (
        <p>Error: {resetPassword.error.response?.data.detail}</p>
      )}
    </form>
  );
}
```

---

## Response Summary

| Endpoint | Status | Response Type | Response Body |
|----------|--------|---------------|---------------|
| `POST /auth/password/reset-request` | 200 | Success | `{ "message": "...", "token": "..." }` |
| `POST /auth/password/reset-request` | 400 | Error | RFC 7807 Problem Details |
| `POST /auth/password/reset-request` | 500 | Error | RFC 7807 Problem Details |
| `POST /auth/password/reset` | 200 | Success | `{ "message": "Password updated successfully" }` |
| `POST /auth/password/reset` | 400 | Error | RFC 7807 Problem Details |
| `POST /auth/password/reset` | 500 | Error | RFC 7807 Problem Details |

---

## Important Notes

1. **Error Format**: All errors use RFC 7807 Problem Details format with `type`, `title`, `status`, and `detail` fields.

2. **Security**: The request reset endpoint always returns 200 OK even if email doesn't exist (prevents email enumeration).

3. **Token in Response**: The `token` field in the reset request response should be removed in production. It's only included for development/testing.

4. **Token Expiration**: Reset tokens expire after 24 hours.

5. **Token Usage**: Each reset token can only be used once. After successful password reset, the token is deleted.

---

# User Account Endpoints

## 1. Create User (Register)
**Endpoint**: `POST /api/v1/users/`

**Authentication**: ❌ None (Public)

**Request Body**:
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePassword123!",
  "firstName": "John",
  "lastName": "Doe"
}
```

**Success Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "active": true,
  "avatarUrl": "",
  "createdAt": "2025-01-20T10:30:00Z",
  "updatedAt": "2025-01-20T10:30:00Z"
}
```

**Error Responses**:
- **400 Bad Request** (Invalid JSON or validation error):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Failed to create user: [error details]"
}
```

- **400 Bad Request** (Invalid JSON format):
```json
{
  "type": "invalid_json",
  "title": "Bad Request",
  "status": 400,
  "detail": "error parsing JSON"
}
```

---

## 2. Get Current User (Me)
**Endpoint**: `GET /api/v1/users/me`

**Authentication**: ✅ Required (Bearer Token)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "active": true,
  "avatarUrl": "https://example.com/avatar.jpg",
  "createdAt": "2025-01-20T10:30:00Z",
  "updatedAt": "2025-01-20T10:30:00Z"
}
```

**Error Responses**:
- **401 Unauthorized** (Missing/Invalid token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "User ID not found in context"
}
```

- **400 Bad Request** (Invalid user ID):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid user ID"
}
```

- **404 Not Found** (User not found):
```json
{
  "type": "not_found",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

---

## 2.5. Update Current User (Me)
**Endpoint**: `PATCH /api/v1/users/me`

**Authentication**: ✅ Required (Bearer Token)

**Headers**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body** (All fields optional - partial update):
```json
{
  "username": "newusername",
  "email": "newemail@example.com",
  "firstName": "NewFirst",
  "lastName": "NewLast",
  "avatarUrl": "https://example.com/new-avatar.jpg"
}
```

**Note**: 
- All fields are optional - only include fields you want to update
- To clear `avatarUrl`, send an empty string: `"avatarUrl": ""`
- Fields not included in the request will remain unchanged
- **⚠️ DO NOT send `password` field here** - use `POST /auth/password/change` instead

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "newusername",
  "email": "newemail@example.com",
  "firstName": "NewFirst",
  "lastName": "NewLast",
  "active": true,
  "avatarUrl": "https://example.com/new-avatar.jpg",
  "createdAt": "2025-01-20T10:30:00Z",
  "updatedAt": "2025-01-20T11:45:00Z"
}
```

**Error Responses**:
- **400 Bad Request** (Invalid JSON, validation error, or unknown field):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "json: unknown field \"password\""
}
```

**Note**: If you receive `"json: unknown field \"password\""`, you are trying to update the password through this endpoint. Use `POST /auth/password/change` instead.

OR

```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Failed to update user: <error details>"
}
```

- **401 Unauthorized** (Missing/Invalid token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "User ID not found in context"
}
```

- **404 Not Found** (User not found):
```json
{
  "type": "not_found",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

**Example Request** (Update only username and email):
```json
{
  "username": "updatedusername",
  "email": "updated@example.com"
}
```

**Example Request** (Clear avatar):
```json
{
  "avatarUrl": ""
}
```

---

## 3. Get User by ID
**Endpoint**: `GET /api/v1/users/{userId}`

**Authentication**: ✅ Required (Bearer Token)

**URL Parameters**:
- `userId` (string, UUID) - Required

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "active": true,
  "avatarUrl": "https://example.com/avatar.jpg",
  "createdAt": "2025-01-20T10:30:00Z",
  "updatedAt": "2025-01-20T10:30:00Z"
}
```

**Error Responses**:
- **400 Bad Request** (Missing/Invalid userId):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "User ID is required"
}
```

OR

```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid user ID"
}
```

- **401 Unauthorized** (Missing/Invalid token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Authorization header required"
}
```

- **404 Not Found** (User not found):
```json
{
  "type": "not_found",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

---

## 4. Validate Email
**Endpoint**: `GET /api/v1/users/validate-email?email={email}` OR `POST /api/v1/users/validate-email`

**Authentication**: ❌ None (Public)

**GET Request** (Query Parameter):
```
GET /api/v1/users/validate-email?email=john@example.com
```

**POST Request Body**:
```json
{
  "email": "john@example.com"
}
```

**Success Response** (200 OK):
```json
{
  "available": true
}
```

OR (if email is taken):
```json
{
  "available": false
}
```

**Error Responses**:
- **400 Bad Request** (Missing email):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "email is required"
}
```

- **400 Bad Request** (Invalid email format):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "invalid email format"
}
```

- **500 Internal Server Error** (Database error):
```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "database error"
}
```

**Note**: Email validation is lenient for live typing. It allows incomplete emails during typing but validates format.

---

## 5. Validate Username
**Endpoint**: `GET /api/v1/users/validate-username?username={username}` OR `POST /api/v1/users/validate-username`

**Authentication**: ❌ None (Public)

**GET Request** (Query Parameter):
```
GET /api/v1/users/validate-username?username=johndoe
```

**POST Request Body**:
```json
{
  "username": "johndoe"
}
```

**Success Response** (200 OK):
```json
{
  "available": true
}
```

OR (if username is taken):
```json
{
  "available": false
}
```

**Error Responses**:
- **400 Bad Request** (Missing username):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "username is required"
}
```

- **400 Bad Request** (Invalid username format):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "invalid username format"
}
```

- **500 Internal Server Error** (Database error):
```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "database error"
}
```

**Username Rules**:
- 3-30 characters
- Alphanumeric, underscores, and hyphens only
- Must start and end with alphanumeric character
- Case-insensitive (stored as lowercase)

---

# Authentication Endpoints

## 1. Login
**Endpoint**: `POST /api/v1/auth/login`

**Authentication**: ❌ None (Public)

**Request Body**:
```json
{
  "username": "johndoe",
  "password": "SecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Cookies Set**:
- `refresh_token` (HttpOnly, Secure, SameSite=None)
  - Used for token refresh
  - Not accessible via JavaScript

**Error Responses**:
- **401 Unauthorized** (Invalid credentials):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Invalid credentials"
}
```

- **401 Unauthorized** (Account deactivated):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Account is deactivated"
}
```

- **500 Internal Server Error** (Token generation failed):
```json
{
  "type": "internal_server_error",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "Failed to generate token"
}
```

---

## 2. Refresh Token
**Endpoint**: `POST /api/v1/auth/refresh`

**Authentication**: ❌ None (Uses refresh token cookie)

**Request Body**: None

**Cookies Required**:
- `refresh_token` (HttpOnly cookie set during login)

**Success Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Cookies Set**:
- `refresh_token` (New refresh token, HttpOnly, Secure, SameSite=None)

**Error Responses**:
- **401 Unauthorized** (No refresh token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Refresh token not found"
}
```

- **401 Unauthorized** (Invalid/Expired refresh token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Invalid refresh token"
}
```

OR

```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Refresh token has expired"
}
```

- **401 Unauthorized** (User not found/deactivated):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "User not found"
}
```

OR

```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Account is deactivated"
}
```

---

## 3. Logout
**Endpoint**: `POST /api/v1/auth/logout`

**Authentication**: ❌ None (Uses refresh token cookie)

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

**Cookies Cleared**:
- `refresh_token` (Deleted)

**Note**: This endpoint deletes the refresh token from the database and clears the cookie.

---

## 4. Change Password (Authenticated)
**Endpoint**: `POST /api/v1/auth/password/change`

**Authentication**: ✅ Required (Bearer Token)

**Headers**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "currentPassword": "oldPassword123",
  "newPassword": "newSecurePassword456"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password updated successfully"
}
```

**Error Responses**:
- **400 Bad Request** (Missing/Invalid fields):
```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "Current password is required"
}
```

OR

```json
{
  "type": "bad_request",
  "title": "Bad Request",
  "status": 400,
  "detail": "New password must be at least 8 characters"
}
```

- **401 Unauthorized** (Incorrect current password):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "Current password is incorrect"
}
```

- **401 Unauthorized** (Missing/Invalid token):
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "User ID not found in context"
}
```

- **404 Not Found** (User not found):
```json
{
  "type": "not_found",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

**Note**: 
- This endpoint is for **authenticated users** to change their own password
- Requires the current password for security
- Different from password reset (forgot password flow) which uses a reset token
- New password must be at least 8 characters

---

# Common Errors (400/405)

## 400 Bad Request - Common Causes

### 1. Invalid JSON Format
**Symptom**: `400 Bad Request` with `type: "invalid_json"`

**Causes**:
- Missing `Content-Type: application/json` header
- Malformed JSON in request body
- Extra fields in JSON (if `DisallowUnknownFields` is enabled)
- Missing required fields

**Fix**:
```typescript
// ✅ Correct
const response = await fetch('/api/v1/users/', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    username: 'johndoe',
    email: 'john@example.com',
    password: 'password123',
    firstName: 'John',
    lastName: 'Doe'
  })
});

// ❌ Wrong - Missing Content-Type
const response = await fetch('/api/v1/users/', {
  method: 'POST',
  body: JSON.stringify({ username: 'johndoe' })
});

// ❌ Wrong - Missing required fields
const response = await fetch('/api/v1/users/', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'johndoe' }) // Missing email, password, etc.
});
```

### 2. Missing Required Fields
**Symptom**: `400 Bad Request` with specific field error

**Fix**: Ensure all required fields are present:
```typescript
// CreateUser requires ALL of these:
{
  username: string;    // Required
  email: string;       // Required
  password: string;    // Required
  firstName: string;   // Required
  lastName: string;    // Required
}
```

### 3. Invalid Field Values
**Symptom**: `400 Bad Request` with validation error

**Examples**:
- Invalid email format
- Invalid username format (must be 3-30 chars, alphanumeric + _ -)
- Invalid UUID format for userId

---

## 405 Method Not Allowed - Common Causes

### 1. Wrong HTTP Method
**Symptom**: `405 Method Not Allowed`

**Common Mistakes**:
- Using `GET` instead of `POST` for mutations
- Using `POST` instead of `GET` for queries
- Using `PUT`/`PATCH` when endpoint doesn't support it

**Endpoint Methods**:
```
POST   /api/v1/users/                    ✅ Create user
GET    /api/v1/users/me                 ✅ Get current user
GET    /api/v1/users/{userId}           ✅ Get user by ID
GET    /api/v1/users/validate-email     ✅ Validate email (GET)
POST   /api/v1/users/validate-email     ✅ Validate email (POST)
GET    /api/v1/users/validate-username  ✅ Validate username (GET)
POST   /api/v1/users/validate-username  ✅ Validate username (POST)

POST   /api/v1/auth/login               ✅ Login
POST   /api/v1/auth/refresh             ✅ Refresh token
POST   /api/v1/auth/logout              ✅ Logout
POST   /api/v1/auth/password/reset-request ✅ Request password reset (forgot password)
POST   /api/v1/auth/password/reset      ✅ Reset password (forgot password)
POST   /api/v1/auth/password/change     ✅ Change password (authenticated user)
```

**Fix**:
```typescript
// ✅ Correct - POST for create
await apiClient.post('/users/', userData);

// ❌ Wrong - GET for create
await apiClient.get('/users/', { params: userData }); // 405 Error!

// ✅ Correct - GET for query
await apiClient.get('/users/me');

// ❌ Wrong - POST for query
await apiClient.post('/users/me'); // 405 Error!
```

### 2. Route Not Found
**Symptom**: `404 Not Found` (not 405, but related)

**Common Mistakes**:
- Missing `/api/v1` prefix
- Wrong endpoint path
- Trailing slash issues

**Fix**:
```typescript
// ✅ Correct
/api/v1/users/me

// ❌ Wrong
/users/me              // Missing /api/v1
/api/v1/user/me        // Wrong: "user" instead of "users"
/api/v1/users/me/      // Trailing slash (may cause issues)
```

---

# Caching Guide

## Current Caching Status

**⚠️ Note**: User endpoints currently do NOT have explicit HTTP caching headers set. However, TanStack Query on the frontend provides client-side caching.

## Frontend Caching Strategy

### Using TanStack Query

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from './apiClient';

// Get current user with caching
export const useGetMe = () => {
  return useQuery({
    queryKey: ['user', 'me'],
    queryFn: async () => {
      const response = await apiClient.get('/users/me');
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
};

// Get user by ID with caching
export const useGetUser = (userId: string | null) => {
  return useQuery({
    queryKey: ['user', userId],
    queryFn: async () => {
      if (!userId) return null;
      const response = await apiClient.get(`/users/${userId}`);
      return response.data;
    },
    enabled: !!userId,
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
};

// Update current user and invalidate cache
export const useUpdateMe = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (userData: UpdateUserRequest) => {
      const response = await apiClient.patch('/users/me', userData);
      return response.data;
    },
    onSuccess: (data) => {
      // Update cache with new data
      queryClient.setQueryData(['user', 'me'], data);
      queryClient.setQueryData(['user', data.id], data);
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
  });
};

// Change password (authenticated user)
export const useChangePassword = () => {
  return useMutation({
    mutationFn: async ({ currentPassword, newPassword }: { currentPassword: string; newPassword: string }) => {
      const response = await apiClient.post('/auth/password/change', {
        currentPassword,
        newPassword,
      });
      return response.data;
    },
    onSuccess: () => {
      // Password change doesn't affect user cache, but you might want to show a success message
      console.log('Password changed successfully');
    },
  });
};
```

**Usage Example**:
```typescript
function ProfileEditForm() {
  const { data: currentUser } = useGetMe();
  const updateMe = useUpdateMe();
  
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    // Only send fields that changed (partial update)
    const updates: UpdateUserRequest = {};
    if (formData.get('username') !== currentUser?.username) {
      updates.username = formData.get('username') as string;
    }
    if (formData.get('email') !== currentUser?.email) {
      updates.email = formData.get('email') as string;
    }
    if (formData.get('firstName') !== currentUser?.firstName) {
      updates.firstName = formData.get('firstName') as string;
    }
    if (formData.get('lastName') !== currentUser?.lastName) {
      updates.lastName = formData.get('lastName') as string;
    }
    
    updateMe.mutate(updates);
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input name="username" defaultValue={currentUser?.username} />
      <input name="email" type="email" defaultValue={currentUser?.email} />
      <input name="firstName" defaultValue={currentUser?.firstName} />
      <input name="lastName" defaultValue={currentUser?.lastName} />
      <button type="submit" disabled={updateMe.isPending}>
        {updateMe.isPending ? 'Updating...' : 'Update Profile'}
      </button>
      {updateMe.isSuccess && <p>Profile updated successfully!</p>}
      {updateMe.isError && (
        <p>Error: {updateMe.error.response?.data.detail}</p>
      )}
    </form>
  );
}

// Example: Clear avatar
function ClearAvatarButton() {
  const updateMe = useUpdateMe();
  
  const handleClear = () => {
    updateMe.mutate({ avatarUrl: '' });
  };
  
  return (
    <button onClick={handleClear} disabled={updateMe.isPending}>
      Remove Avatar
    </button>
  );
}

// Example: Change password
function ChangePasswordForm() {
  const changePassword = useChangePassword();
  
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    changePassword.mutate({
      currentPassword: formData.get('currentPassword') as string,
      newPassword: formData.get('newPassword') as string,
    });
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input name="currentPassword" type="password" placeholder="Current Password" required />
      <input name="newPassword" type="password" placeholder="New Password" required minLength={8} />
      <button type="submit" disabled={changePassword.isPending}>
        {changePassword.isPending ? 'Changing...' : 'Change Password'}
      </button>
      {changePassword.isSuccess && <p>Password changed successfully!</p>}
      {changePassword.isError && (
        <p>Error: {changePassword.error.response?.data.detail}</p>
      )}
    </form>
  );
}
```

### Cache Invalidation on User Updates

When user data changes (via WebSocket cache invalidation or direct updates):

```typescript
// In your cache invalidation handler
case 'users':
  // Invalidate all user queries
  queryClient.invalidateQueries({ queryKey: ['user'] });
  // Or more specific:
  queryClient.invalidateQueries({ queryKey: ['user', 'me'] });
  queryClient.invalidateQueries({ queryKey: ['user', user_id] });
  break;
```

---

## Complete TypeScript Interfaces

```typescript
// Request Interfaces
interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

interface LoginRequest {
  username: string;
  password: string;
}

interface ValidateEmailRequest {
  email: string;
}

interface ValidateUsernameRequest {
  username: string;
}

interface UpdateUserRequest {
  username?: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  avatarUrl?: string; // Empty string to clear avatar
  // ⚠️ DO NOT include password here - use ChangePasswordRequest instead
}

interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

// Response Interfaces
interface UserResponse {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  active: boolean;
  avatarUrl: string;
  createdAt: string; // ISO 8601 format
  updatedAt: string; // ISO 8601 format
}

interface LoginResponse {
  token: string;
  userId: string;
}

interface ValidateResponse {
  available: boolean;
}

// Error Interface (RFC 7807 Problem Details)
interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
}
```

---

## Complete Endpoint Reference Table

| Endpoint | Method | Auth | Request Body | Success Response | Error Codes |
|----------|--------|------|--------------|------------------|-------------|
| `/users/` | POST | ❌ | `CreateUserRequest` | `UserResponse` (201) | 400, 500 |
| `/users/me` | GET | ✅ | None | `UserResponse` (200) | 400, 401, 404 |
| `/users/me` | PATCH | ✅ | `UpdateUserRequest` | `UserResponse` (200) | 400, 401, 404 |
| `/users/{userId}` | GET | ✅ | None | `UserResponse` (200) | 400, 401, 404 |
| `/users/validate-email` | GET/POST | ❌ | Query/`ValidateEmailRequest` | `ValidateResponse` (200) | 400, 500 |
| `/users/validate-username` | GET/POST | ❌ | Query/`ValidateUsernameRequest` | `ValidateResponse` (200) | 400, 500 |
| `/auth/login` | POST | ❌ | `LoginRequest` | `LoginResponse` (200) | 401, 500 |
| `/auth/refresh` | POST | ❌* | None | `LoginResponse` (200) | 401, 500 |
| `/auth/logout` | POST | ❌* | None | `{message}` (200) | 401 |
| `/auth/password/reset-request` | POST | ❌ | `{email}` | `{message, token?}` (200) | 400, 500 |
| `/auth/password/reset` | POST | ❌ | `{token, password}` | `{message}` (200) | 400, 500 |
| `/auth/password/change` | POST | ✅ | `{currentPassword, newPassword}` | `{message}` (200) | 400, 401, 404, 500 |

*Uses refresh token cookie instead of Bearer token

---

## Quick Debugging Checklist

### 400 Bad Request
- [ ] Check `Content-Type: application/json` header is set
- [ ] Verify JSON is valid (no syntax errors)
- [ ] Ensure all required fields are present
- [ ] Check field types match (string, not number, etc.)
- [ ] Verify no extra fields if `DisallowUnknownFields` is enabled
- [ ] Check UUID format for userId parameters

### 405 Method Not Allowed
- [ ] Verify HTTP method matches endpoint (GET vs POST)
- [ ] Check endpoint path is correct (`/api/v1/users/me` not `/users/me`)
- [ ] Ensure no trailing slashes on GET endpoints
- [ ] Verify route exists in router configuration

### 401 Unauthorized
- [ ] Check `Authorization: Bearer <token>` header is present
- [ ] Verify token is not expired
- [ ] Ensure token is valid JWT format
- [ ] Check token includes required claims (sub, exp, etc.)

### 404 Not Found
- [ ] Verify endpoint path is correct
- [ ] Check userId exists in database
- [ ] Ensure `/api/v1` prefix is included
- [ ] Verify route is registered in router

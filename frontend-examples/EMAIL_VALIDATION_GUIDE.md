# Email Validation Guide - User Creation

## Overview

The backend provides **two levels** of email validation:

1. **Live Validation** (`GET/POST /api/v1/users/validate-email`) - For real-time feedback as user types
2. **Strict Validation** (in `CreateUser`) - For final submission (currently relies on database constraints)

## Email Validation Endpoints

### 1. Validate Email Availability

**Endpoint**: `GET /api/v1/users/validate-email?email={email}`  
**OR**: `POST /api/v1/users/validate-email`

**Purpose**: Check if email is available and has valid format (for live typing feedback)

**Request (GET)**:
```
GET /api/v1/users/validate-email?email=user@example.com
```

**Request (POST)**:
```json
{
  "email": "user@example.com"
}
```

**Response (200 OK)**:
```json
{
  "available": true
}
```

**Response (200 OK - Email Taken)**:
```json
{
  "available": false
}
```

**Error Responses**:
- `400 Bad Request` - Missing email or invalid email format

---

## Email Validation Rules

### Live Validation (for typing feedback)

The `validate-email` endpoint uses **lenient validation** to allow users to type without premature errors:

✅ **Allowed during typing**:
- Empty string
- Incomplete emails (e.g., `user@`, `user@ex`)
- Basic format checks only

❌ **Rejected**:
- Whitespace characters
- Multiple `@` symbols
- Empty local part (before `@`)
- Whitespace in domain

**Example**:
```typescript
// ✅ Valid during typing
"user"           // No @ yet, OK
"user@"          // Incomplete, OK
"user@ex"        // Incomplete domain, OK
"user@example"    // No TLD yet, OK

// ❌ Invalid
"user @example"  // Whitespace
"user@@example"   // Multiple @
"@example"       // Empty local part
```

### Strict Validation (for final submission)

When creating a user, the email must pass **strict validation**:

✅ **Requirements**:
- Length: 5-254 characters
- Exactly one `@` symbol
- Non-empty local part (before `@`)
- Non-empty domain (after `@`)
- Domain must contain at least one `.` (e.g., `example.com`)

❌ **Rejected**:
- Too short (< 5 chars) or too long (> 254 chars)
- No `@` or multiple `@`
- Empty local or domain parts
- Domain without `.` (e.g., `user@example`)

**Example**:
```typescript
// ✅ Valid for submission
"user@example.com"
"test.user@subdomain.example.com"
"user+tag@example.co.uk"

// ❌ Invalid for submission
"user@example"        // No TLD
"user@"              // Incomplete
"@example.com"       // Empty local part
"user@.com"          // Empty domain
"user@example"       // No dot in domain
```

---

## Frontend Implementation

### Step 1: Live Validation (as user types)

```typescript
import { useState, useEffect } from 'react';
import { useDebounce } from './useDebounce'; // Or use a debounce hook
import apiClient from '../lib/apiClient';

interface EmailValidationResult {
  available: boolean;
  isValid: boolean;
  error?: string;
}

export function useEmailValidation(email: string) {
  const [result, setResult] = useState<EmailValidationResult>({
    available: false,
    isValid: false,
  });
  const [isValidating, setIsValidating] = useState(false);

  // Debounce email input to avoid excessive API calls
  const debouncedEmail = useDebounce(email, 500);

  useEffect(() => {
    if (!debouncedEmail) {
      setResult({ available: false, isValid: false });
      return;
    }

    setIsValidating(true);

    // Call validation endpoint
    apiClient
      .get('/users/validate-email', {
        params: { email: debouncedEmail },
      })
      .then((response) => {
        setResult({
          available: response.data.available,
          isValid: true,
        });
      })
      .catch((error) => {
        if (error.response?.status === 400) {
          // Invalid email format
          setResult({
            available: false,
            isValid: false,
            error: 'Invalid email format',
          });
        } else {
          setResult({
            available: false,
            isValid: false,
            error: 'Failed to validate email',
          });
        }
      })
      .finally(() => {
        setIsValidating(false);
      });
  }, [debouncedEmail]);

  return { ...result, isValidating };
}
```

### Step 2: Use in Registration Form

```typescript
import { useState } from 'react';
import { useEmailValidation } from './useEmailValidation';
import apiClient from '../lib/apiClient';

function RegistrationForm() {
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const { available, isValid, isValidating, error: validationError } = useEmailValidation(email);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Final validation before submission
    if (!isValid || !available) {
      setError('Please enter a valid, available email address');
      return;
    }

    // Additional strict validation on frontend
    if (!isStrictEmailValid(email)) {
      setError('Email format is invalid');
      return;
    }

    setIsSubmitting(true);

    try {
      const response = await apiClient.post('/users', {
        username,
        email,
        password,
        firstName,
        lastName,
      });

      // Success - redirect or show success message
      console.log('User created:', response.data);
    } catch (err: any) {
      if (err.response?.status === 400) {
        setError(err.response.data.message || 'Failed to create user');
      } else {
        setError('An error occurred. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label>Email</label>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="user@example.com"
        />
        {isValidating && <span>Checking...</span>}
        {!isValidating && email && (
          <>
            {!isValid && (
              <span className="error">Invalid email format</span>
            )}
            {isValid && !available && (
              <span className="error">Email is already taken</span>
            )}
            {isValid && available && (
              <span className="success">✓ Email is available</span>
            )}
          </>
        )}
      </div>

      <div>
        <label>Username</label>
        <input
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>

      <div>
        <label>Password</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>

      <div>
        <label>First Name</label>
        <input
          type="text"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
        />
      </div>

      <div>
        <label>Last Name</label>
        <input
          type="text"
          value={lastName}
          onChange={(e) => setLastName(e.target.value)}
        />
      </div>

      {error && <div className="error">{error}</div>}

      <button type="submit" disabled={isSubmitting || !isValid || !available}>
        {isSubmitting ? 'Creating Account...' : 'Create Account'}
      </button>
    </form>
  );
}

// Strict email validation function (matches backend)
function isStrictEmailValid(email: string): boolean {
  if (email.length < 5 || email.length > 254) {
    return false;
  }

  const atCount = (email.match(/@/g) || []).length;
  if (atCount !== 1) {
    return false;
  }

  const parts = email.split('@');
  if (parts.length !== 2) {
    return false;
  }

  const [local, domain] = parts;
  if (!local || !domain) {
    return false;
  }

  // Domain must contain at least one dot
  if (!domain.includes('.')) {
    return false;
  }

  return true;
}
```

---

## Backend Validation Flow

### Current Implementation

1. **Validate Email Endpoint** (`/users/validate-email`):
   - Uses `isValidEmailForLiveValidation()` - lenient validation
   - Checks database for existing email
   - Returns `{ available: true/false }`

2. **Create User Endpoint** (`POST /users`):
   - **Currently**: Relies on database constraints (unique email)
   - **No explicit format validation** in handler
   - Database will reject duplicate emails

### Recommended: Add Strict Validation to CreateUser

The `CreateUser` handler should validate email format before attempting to create the user:

```go
// In CreateUser handler (should be added)
if !isValidEmail(req.Email) {
    response.BadRequest(w, "Invalid email format")
    return
}
```

---

## Complete Request/Response Examples

### Validate Email (GET)

**Request**:
```bash
curl "https://devhive-go-backend.fly.dev/api/v1/users/validate-email?email=user@example.com"
```

**Response (Available)**:
```json
{
  "available": true
}
```

**Response (Taken)**:
```json
{
  "available": false
}
```

**Error (Invalid Format)**:
```json
{
  "type": "https://tools.ietf.org/html/rfc7807",
  "title": "Bad Request",
  "status": 400,
  "detail": "invalid email format"
}
```

### Validate Email (POST)

**Request**:
```bash
curl -X POST "https://devhive-go-backend.fly.dev/api/v1/users/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

**Response**: Same as GET

### Create User

**Request**:
```bash
curl -X POST "https://devhive-go-backend.fly.dev/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword123",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

**Response (201 Created)**:
```json
{
  "id": "uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "active": true,
  "avatarUrl": "",
  "createdAt": "2025-01-20T15:30:00Z",
  "updatedAt": "2025-01-20T15:30:00Z"
}
```

**Error (Duplicate Email)**:
```json
{
  "type": "https://tools.ietf.org/html/rfc7807",
  "title": "Bad Request",
  "status": 400,
  "detail": "Failed to create user: duplicate key value violates unique constraint \"users_email_key\""
}
```

---

## Best Practices

1. **Validate on blur** - Don't validate on every keystroke, use debouncing
2. **Show clear feedback** - Display whether email is available/invalid
3. **Disable submit** - Disable submit button if email is invalid or unavailable
4. **Final validation** - Always validate again before submission (don't trust client-side only)
5. **Handle errors gracefully** - Show user-friendly error messages

---

## TypeScript Interfaces

```typescript
// Validation request
interface ValidateEmailRequest {
  email: string;
}

// Validation response
interface ValidateEmailResponse {
  available: boolean;
}

// Create user request
interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

// User response
interface UserResponse {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  active: boolean;
  avatarUrl: string;
  createdAt: string;
  updatedAt: string;
}
```

---

## Summary

- ✅ Use `/users/validate-email` for **live validation** as user types
- ✅ Use **debouncing** to avoid excessive API calls
- ✅ Show **real-time feedback** (available/invalid)
- ✅ Perform **strict validation** before final submission
- ⚠️ **Note**: `CreateUser` currently relies on database constraints; consider adding explicit validation




# Email Validation Guide

This guide explains how email validation works on the backend and how to implement real-time validation on the frontend.

## Backend Email Validation

### Endpoint

**URL:** `GET /api/v1/users/validate-email?email={email}` or `POST /api/v1/users/validate-email`

**Method:** `GET` or `POST`

**Authentication:** Not required (public endpoint)

**Request (GET):**
```
GET /api/v1/users/validate-email?email=user@example.com
```

**Request (POST):**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "available": true
}
```

- `available: true` - Email is available (not in use)
- `available: false` - Email is already taken

**Error Responses:**

- `400 Bad Request` - Invalid email format or missing email
- `500 Internal Server Error` - Database error

### Two-Tier Validation System

The backend uses **two different validation functions** depending on context:

#### 1. Lenient Validation (Live Typing)
Used by: `GET/POST /api/v1/users/validate-email`

**Function:** `isValidEmailForLiveValidation(email string)`

**Rules:**
- Allows empty strings (user is still typing)
- Rejects strings with whitespace (` `, `\t`, `\n`, `\r`)
- Allows incomplete emails (e.g., `user@` or `user@example`)
- Only rejects if:
  - Multiple `@` symbols
  - `@` symbol with empty local part
  - Domain contains whitespace

**Purpose:** Don't annoy users while they're typing. Only reject obviously invalid input.

#### 2. Strict Validation (Final Submission)
Used by: `POST /api/v1/users` (user creation)

**Function:** `isValidEmail(email string)`

**Rules:**
- Length: 5-254 characters
- Exactly one `@` symbol
- Non-empty local part (before `@`)
- Non-empty domain part (after `@`)
- Domain must contain at least one `.` (e.g., `example.com`)

**Purpose:** Ensure data integrity before creating a user account.

### Backend Code Reference

**File:** `internal/http/handlers/user_validate.go`

```go
// ValidateEmail handles email availability validation
func (h *UserHandler) ValidateEmail(w http.ResponseWriter, r *http.Request) {
    var email string

    // Support both GET (query param) and POST (JSON body) requests
    if r.Method == "GET" {
        email = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("email")))
    } else {
        var req ValidateEmailRequest
        if !response.Decode(w, r, &req) {
            return
        }
        email = strings.TrimSpace(strings.ToLower(req.Email))
    }

    if email == "" {
        response.BadRequest(w, "email is required")
        return
    }

    // Validate email format (basic validation for live typing)
    if !isValidEmailForLiveValidation(email) {
        response.BadRequest(w, "invalid email format")
        return
    }

    _, err := h.queries.GetUserByEmail(r.Context(), email)
    available := false
    if err != nil {
        if err == pgx.ErrNoRows {
            available = true
        } else {
            response.InternalServerError(w, "database error")
            return
        }
    }

    response.JSON(w, http.StatusOK, ValidateResponse{
        Available: available,
    })
}
```

## Frontend Real-Time Validation

### Implementation Pattern

The frontend should implement **debounced validation** to avoid excessive API calls while the user is typing.

### Key Requirements

1. **Debounce:** Wait 300-500ms after user stops typing before validating
2. **Skip empty:** Don't validate empty strings
3. **Skip obviously invalid:** Don't validate if format is clearly wrong (client-side check)
4. **Show feedback:** Display validation state (checking, available, taken, invalid)
5. **Final validation:** Always validate again on form submission with strict rules

### React Hook Example

```typescript
import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';

interface EmailValidationState {
  isValid: boolean | null; // null = not validated yet, true = available, false = taken/invalid
  isChecking: boolean;
  error: string | null;
}

export function useEmailValidation(email: string, debounceMs: number = 400) {
  const [debouncedEmail, setDebouncedEmail] = useState(email);
  const [validationState, setValidationState] = useState<EmailValidationState>({
    isValid: null,
    isChecking: false,
    error: null,
  });

  // Debounce the email input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedEmail(email);
    }, debounceMs);

    return () => clearTimeout(timer);
  }, [email, debounceMs]);

  // Client-side format check (lenient, like backend)
  const hasBasicFormat = (email: string): boolean => {
    if (email.length === 0) return true; // Allow empty during typing
    if (/\s/.test(email)) return false; // Reject whitespace
    if ((email.match(/@/g) || []).length > 1) return false; // Reject multiple @
    
    if (email.includes('@')) {
      const parts = email.split('@');
      if (parts.length !== 2 || parts[0].length === 0) return false;
      if (/\s/.test(parts[1])) return false; // Reject whitespace in domain
    }
    
    return true;
  };

  // Query for email availability
  const { data, isLoading, error } = useQuery({
    queryKey: ['validate-email', debouncedEmail],
    queryFn: async () => {
      if (!debouncedEmail || debouncedEmail.length === 0) {
        return { available: null };
      }

      // Skip API call if format is obviously invalid
      if (!hasBasicFormat(debouncedEmail)) {
        return { available: false, reason: 'invalid_format' };
      }

      const response = await fetch(
        `/api/v1/users/validate-email?email=${encodeURIComponent(debouncedEmail)}`
      );

      if (!response.ok) {
        if (response.status === 400) {
          return { available: false, reason: 'invalid_format' };
        }
        throw new Error('Validation failed');
      }

      const result = await response.json();
      return { available: result.available };
    },
    enabled: debouncedEmail.length > 0 && hasBasicFormat(debouncedEmail),
    staleTime: 30000, // Cache for 30 seconds
  });

  // Update validation state
  useEffect(() => {
    if (debouncedEmail.length === 0) {
      setValidationState({
        isValid: null,
        isChecking: false,
        error: null,
      });
      return;
    }

    if (!hasBasicFormat(debouncedEmail)) {
      setValidationState({
        isValid: false,
        isChecking: false,
        error: 'Invalid email format',
      });
      return;
    }

    if (isLoading) {
      setValidationState({
        isValid: null,
        isChecking: true,
        error: null,
      });
      return;
    }

    if (error) {
      setValidationState({
        isValid: false,
        isChecking: false,
        error: 'Failed to validate email',
      });
      return;
    }

    if (data) {
      if (data.reason === 'invalid_format') {
        setValidationState({
          isValid: false,
          isChecking: false,
          error: 'Invalid email format',
        });
      } else if (data.available === false) {
        setValidationState({
          isValid: false,
          isChecking: false,
          error: 'This email is already registered',
        });
      } else {
        setValidationState({
          isValid: true,
          isChecking: false,
          error: null,
        });
      }
    }
  }, [debouncedEmail, data, isLoading, error]);

  return validationState;
}
```

### React Component Example

```typescript
import React, { useState } from 'react';
import { useEmailValidation } from './hooks/useEmailValidation';

export function RegistrationForm() {
  const [email, setEmail] = useState('');
  const validation = useEmailValidation(email);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Final strict validation before submission
    const strictEmailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!strictEmailRegex.test(email) || email.length < 5 || email.length > 254) {
      alert('Please enter a valid email address');
      return;
    }

    // Check if email is available one more time
    if (validation.isValid === false) {
      alert('This email is already registered');
      return;
    }

    // Proceed with registration
    // ... submit form
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className={validation.isValid === false ? 'error' : validation.isValid === true ? 'success' : ''}
        />
        
        {validation.isChecking && (
          <span className="checking">Checking availability...</span>
        )}
        
        {validation.isValid === true && (
          <span className="success">✓ Email is available</span>
        )}
        
        {validation.isValid === false && validation.error && (
          <span className="error">✗ {validation.error}</span>
        )}
      </div>

      <button type="submit" disabled={validation.isValid !== true}>
        Register
      </button>
    </form>
  );
}
```

### Vanilla JavaScript Example

```javascript
class EmailValidator {
  constructor(apiBaseUrl, debounceMs = 400) {
    this.apiBaseUrl = apiBaseUrl;
    this.debounceMs = debounceMs;
    this.debounceTimer = null;
    this.currentValidation = null;
  }

  // Client-side format check (lenient)
  hasBasicFormat(email) {
    if (email.length === 0) return true;
    if (/\s/.test(email)) return false;
    if ((email.match(/@/g) || []).length > 1) return false;
    
    if (email.includes('@')) {
      const parts = email.split('@');
      if (parts.length !== 2 || parts[0].length === 0) return false;
      if (/\s/.test(parts[1])) return false;
    }
    
    return true;
  }

  // Validate email availability
  async validate(email) {
    if (!email || email.length === 0) {
      return { available: null, checking: false };
    }

    if (!this.hasBasicFormat(email)) {
      return { available: false, checking: false, error: 'Invalid email format' };
    }

    try {
      const response = await fetch(
        `${this.apiBaseUrl}/api/v1/users/validate-email?email=${encodeURIComponent(email)}`
      );

      if (!response.ok) {
        if (response.status === 400) {
          return { available: false, checking: false, error: 'Invalid email format' };
        }
        throw new Error('Validation failed');
      }

      const result = await response.json();
      return {
        available: result.available,
        checking: false,
        error: result.available ? null : 'This email is already registered',
      };
    } catch (error) {
      return { available: false, checking: false, error: 'Failed to validate email' };
    }
  }

  // Debounced validation
  validateDebounced(email, callback) {
    clearTimeout(this.debounceTimer);

    // Immediately show "checking" state if email has basic format
    if (this.hasBasicFormat(email) && email.length > 0) {
      callback({ available: null, checking: true });
    }

    this.debounceTimer = setTimeout(async () => {
      const result = await this.validate(email);
      callback(result);
    }, this.debounceMs);
  }
}

// Usage
const validator = new EmailValidator('https://api.example.com');

const emailInput = document.getElementById('email');
const feedbackElement = document.getElementById('email-feedback');

emailInput.addEventListener('input', (e) => {
  const email = e.target.value;
  
  validator.validateDebounced(email, (result) => {
    if (result.checking) {
      feedbackElement.textContent = 'Checking availability...';
      feedbackElement.className = 'checking';
    } else if (result.available === true) {
      feedbackElement.textContent = '✓ Email is available';
      feedbackElement.className = 'success';
    } else if (result.available === false) {
      feedbackElement.textContent = `✗ ${result.error}`;
      feedbackElement.className = 'error';
    } else {
      feedbackElement.textContent = '';
      feedbackElement.className = '';
    }
  });
});
```

## Validation Flow Diagram

```
User Types Email
    ↓
Client-side format check (lenient)
    ↓
If obviously invalid → Show error immediately
    ↓
If valid format → Wait 400ms (debounce)
    ↓
Call GET /api/v1/users/validate-email?email={email}
    ↓
Backend: isValidEmailForLiveValidation() → Check format
    ↓
Backend: Check database for existing email
    ↓
Response: { "available": true/false }
    ↓
Frontend: Update UI (available/taken/invalid)
    ↓
User Submits Form
    ↓
Frontend: Strict validation (regex + length)
    ↓
If invalid → Block submission
    ↓
If valid → Submit to POST /api/v1/users
    ↓
Backend: isValidEmail() → Strict validation
    ↓
Backend: Create user or return error
```

## Best Practices

1. **Always debounce:** Don't validate on every keystroke
2. **Client-side first:** Do basic format checks before API calls
3. **Show feedback:** Let users know validation is happening
4. **Final validation:** Always validate again on submission
5. **Handle errors gracefully:** Network errors shouldn't break the form
6. **Cache results:** Use TanStack Query or similar to cache validation results
7. **Don't block typing:** Validation errors shouldn't prevent users from continuing to type

## Error Handling

### Backend Errors

- **400 Bad Request:** Invalid email format or missing email
  - Frontend should show format error message
- **500 Internal Server Error:** Database error
  - Frontend should show generic error, allow retry

### Network Errors

- **Timeout:** Show "Validation unavailable, please try again"
- **Offline:** Show "Please check your connection"
- **CORS:** Check API configuration

## Testing

### Test Cases for Frontend

1. Empty email → No validation call
2. Typing "user@" → Wait for debounce, then validate
3. Typing "user@example" → Validate (incomplete domain allowed)
4. Typing "user@example.com" → Validate, show available/taken
5. Typing "user @example.com" (with space) → Show format error immediately
6. Network error → Show error message, allow retry
7. Form submission with invalid email → Block submission
8. Form submission with valid but taken email → Show error

### Test Cases for Backend

1. Empty email → 400 Bad Request
2. "user@" → 200 OK (lenient validation allows)
3. "user@example.com" → 200 OK, check availability
4. "user @example.com" → 400 Bad Request (whitespace)
5. "user@@example.com" → 400 Bad Request (multiple @)
6. Valid available email → 200 OK, `{"available": true}`
7. Valid taken email → 200 OK, `{"available": false}`




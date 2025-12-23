# Admin Certificates Page Password Protection

## Backend Implementation

The backend now has password protection endpoints configured with password: **`jtAppmine2021`**

### API Endpoints

#### 1. Verify Password
**POST** `/api/v1/verify-password`

**Request Body:**
```json
{
  "password": "jtAppmine2021"
}
```

**Success Response (200):**
```json
{
  "success": true
}
```
- Sets HttpOnly cookie `admin_certificates_verified=true` (expires in 30 days)

**Error Response (401):**
```json
{
  "success": false,
  "message": "Nesprávné heslo"
}
```

#### 2. Check Authentication Status
**GET** `/api/v1/check-auth`

**Success Response (200):**
```json
{
  "authenticated": true
}
```

**Unauthorized Response (401):**
```json
{
  "authenticated": false
}
```

## Frontend Implementation Guide

Your frontend admin page (`/admin/certifikaty`) should:

1. **On page load**, check authentication:
   ```typescript
   const checkAuth = async () => {
     const response = await fetch('/api/v1/check-auth', {
       credentials: 'include' // Important: sends cookies
     });
     const data = await response.json();
     return data.authenticated;
   };
   ```

2. **If not authenticated**, show password form:
   ```typescript
   const handlePasswordSubmit = async (password: string) => {
     const response = await fetch('/api/v1/verify-password', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       credentials: 'include', // Important: receives cookies
       body: JSON.stringify({ password })
     });
     const data = await response.json();
     if (data.success) {
       // Redirect to admin page or reload
       window.location.reload();
     } else {
       // Show error message
       alert(data.message || 'Nesprávné heslo');
     }
   };
   ```

3. **Important**: Use `credentials: 'include'` in all fetch requests to send/receive cookies

## Configuration

The password can be configured via environment variable:
- `ADMIN_CERTIFICATES_PASSWORD` (default: `jtAppmine2021`)

To change it on Fly.io:
```bash
fly secrets set ADMIN_CERTIFICATES_PASSWORD="your-new-password"
```

## Cookie Details

- **Name**: `admin_certificates_verified`
- **Value**: `true`
- **Expires**: 30 days
- **HttpOnly**: Yes (not accessible via JavaScript)
- **Secure**: Yes (HTTPS only in production)
- **SameSite**: None (for cross-origin support)

## Security Notes

- Password is stored in backend config/environment variable
- Cookie is HttpOnly (prevents XSS attacks)
- Cookie expires after 30 days
- Use HTTPS in production (Secure flag)







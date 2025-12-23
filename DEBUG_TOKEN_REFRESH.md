# Debugging Token Refresh Issue

## Quick Diagnosis Steps

### 1. Check if Cookie Exists After Login

**In Browser DevTools:**
1. Open DevTools (F12)
2. Go to **Application** tab → **Cookies**
3. Select your backend domain: `https://devhive-go-backend.fly.dev`
4. Look for `refresh_token` cookie

**What to Check:**
- ✅ Cookie exists
- ✅ `HttpOnly` = checked
- ✅ `Secure` = checked  
- ✅ `SameSite` = `None`
- ✅ `Path` = `/`
- ✅ `Expires` = future date (or "Session" if rememberMe=false)

**If cookie is missing:** Cookie wasn't set/accepted by browser → **This is your problem**

---

### 2. Check if Refresh Endpoint is Called

**In Browser DevTools:**
1. Open **Network** tab
2. Wait 15+ minutes (or manually expire token)
3. Make any API request (e.g., get projects)
4. Look for `/api/v1/auth/refresh` request

**What to Check:**
- ✅ Request exists
- ✅ Request headers include: `Cookie: refresh_token=...`
- ✅ Response status = 200
- ✅ Response body has: `{"token": "..."}`

**If refresh request doesn't exist:** Interceptor not working → Check `apiClient.ts` import

**If refresh request fails:** Cookie not being sent or invalid → Check cookie settings

---

### 3. Check CORS Headers

**In Browser DevTools:**
1. Open **Network** tab
2. Click any API request
3. Check **Response Headers**:
   - `Access-Control-Allow-Origin: https://devhive.it.com` (or your frontend domain)
   - `Access-Control-Allow-Credentials: true`

**If missing or wrong:** CORS not configured → Check Fly.io secrets

---

### 4. Check Access Token Storage

**In Browser Console:**
```javascript
// If you're using the tokenManager from apiClient.ts
// Check if token is stored in memory
// (This requires exposing it or checking your auth context)
```

**Or check your login handler:**
```typescript
// After login, verify token is stored:
const response = await apiClient.post('/auth/login', {...});
console.log('Token received:', response.data.token);
// Then check if it's stored:
// tokenManager.setAccessToken(response.data.token); // Make sure this is called!
```

---

## Most Common Issues

### Issue A: Cookie Not Set After Login

**Symptoms:**
- Login succeeds, but no `refresh_token` cookie in Application → Cookies
- Refresh fails immediately

**Causes:**
1. **CORS not allowing credentials**
   - Backend must return: `Access-Control-Allow-Credentials: true`
   - Backend must return: `Access-Control-Allow-Origin: <exact-domain>` (NOT `*`)

2. **Cookie rejected by browser**
   - SameSite policy violation
   - Domain mismatch

**Fix:**
```bash
# Check Fly.io secrets
fly secrets list | grep CORS

# Set if missing:
fly secrets set CORS_ALLOW_CREDENTIALS="true"
fly secrets set CORS_ORIGINS="https://devhive.it.com,http://localhost:3000"
```

---

### Issue B: Cookie Expired

**Symptoms:**
- Cookie exists but refresh fails
- Cookie expires date is in the past

**Causes:**
1. **rememberMe=false** → Session cookie (expires when browser closes)
2. **Browser was closed/reopened** → Session cookie gone

**Fix:**
- Use `rememberMe=true` for persistent cookies
- Or handle session expiry gracefully (redirect to login)

---

### Issue C: Access Token Not Stored

**Symptoms:**
- Login succeeds, but subsequent requests fail immediately
- Token not in memory

**Causes:**
- Frontend not calling `tokenManager.setAccessToken()` after login

**Fix:**
```typescript
// In your login handler:
const response = await apiClient.post('/auth/login', { username, password });
tokenManager.setAccessToken(response.data.token); // ← Make sure this is called!
```

---

### Issue D: Interceptor Not Working

**Symptoms:**
- 401 errors occur, but `/auth/refresh` is never called
- No automatic retry

**Causes:**
1. `apiClient.ts` not imported/used
2. Route excluded from interceptor (in `AUTH_ROUTES` array)
3. Interceptor code has bug

**Fix:**
- Verify `apiClient` is used for all API calls (not raw axios)
- Check that route is not in `AUTH_ROUTES` array
- Add console.log to interceptor to debug

---

## Quick Test Script

Add this to your frontend to test:

```typescript
// Test function to check auth state
export function debugAuth() {
  console.log('=== Auth Debug ===');
  
  // Check cookie (can't read HttpOnly cookie from JS, but check in DevTools)
  console.log('1. Check Application → Cookies → refresh_token');
  
  // Check token in memory
  const token = tokenManager.getAccessToken();
  console.log('2. Access token in memory:', token ? 'EXISTS' : 'MISSING');
  
  // Test refresh endpoint
  apiClient.post('/auth/refresh', {})
    .then(res => {
      console.log('3. Refresh endpoint works:', res.data);
    })
    .catch(err => {
      console.error('3. Refresh endpoint failed:', err.response?.status, err.response?.data);
    });
}
```

---

## Expected Behavior

### After Login:
1. ✅ `refresh_token` cookie exists (check Application → Cookies)
2. ✅ Access token stored in memory (check your auth context)
3. ✅ API requests include `Authorization: Bearer <token>` header

### After 15 Minutes (Token Expires):
1. ✅ API request returns 401
2. ✅ Interceptor catches 401
3. ✅ `/auth/refresh` is called automatically
4. ✅ Cookie is sent with refresh request
5. ✅ New access token received
6. ✅ Original request retried with new token
7. ✅ Request succeeds

### If Any Step Fails:
- Check the corresponding section above
- Most likely: Cookie not being sent → Check CORS and cookie settings


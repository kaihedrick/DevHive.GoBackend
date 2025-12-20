# How to Verify CORS Configuration on Fly.io

## 1. Check Fly.io Secrets/Environment Variables

### List all secrets:
```bash
fly secrets list
```

### Check specific CORS secrets:
```bash
fly secrets list | grep CORS
```

### Expected values:
- `CORS_ORIGINS` should include your frontend domain (e.g., `https://devhive.it.com`)
- `CORS_ALLOW_CREDENTIALS` should be set to `true`

### Set/Update secrets if needed:
```bash
# Set CORS_ORIGINS (comma-separated list)
fly secrets set CORS_ORIGINS="https://devhive.it.com,https://d35scdhidypl44.cloudfront.net"

# Set CORS_ALLOW_CREDENTIALS
fly secrets set CORS_ALLOW_CREDENTIALS="true"
```

## 2. Test CORS Headers with curl

### Test OPTIONS (preflight) request:
```bash
curl -X OPTIONS https://devhive-go-backend.fly.dev/api/v1/auth/login \
  -H "Origin: https://devhive.it.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v
```

### Test actual POST request with CORS headers:
```bash
curl -X POST https://devhive-go-backend.fly.dev/api/v1/auth/login \
  -H "Origin: https://devhive.it.com" \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}' \
  -v
```

### What to look for in the response headers:
```
< HTTP/1.1 200 OK
< Access-Control-Allow-Origin: https://devhive.it.com
< Access-Control-Allow-Credentials: true
< Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
< Access-Control-Allow-Headers: *
< Access-Control-Max-Age: 300
```

## 3. Test Refresh Token Endpoint (with cookies)

### Test refresh endpoint with credentials:
```bash
curl -X POST https://devhive-go-backend.fly.dev/api/v1/auth/refresh \
  -H "Origin: https://devhive.it.com" \
  -H "Cookie: refresh_token=your_refresh_token_here" \
  -v \
  --cookie-jar cookies.txt \
  --cookie cookies.txt
```

## 4. Check Application Logs

### View recent logs:
```bash
fly logs
```

### Filter for CORS-related errors:
```bash
fly logs | grep -i cors
```

## 5. Browser DevTools Check

1. Open your frontend in the browser
2. Open DevTools → Network tab
3. Make a request to your backend
4. Check the Response Headers for:
   - `Access-Control-Allow-Origin: https://devhive.it.com`
   - `Access-Control-Allow-Credentials: true`
   - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS`

## 6. Quick Verification Script

Save this as `test-cors.sh`:

```bash
#!/bin/bash

BACKEND_URL="https://devhive-go-backend.fly.dev"
FRONTEND_ORIGIN="https://devhive.it.com"

echo "Testing CORS configuration..."
echo ""

echo "1. Testing OPTIONS (preflight) request:"
curl -X OPTIONS "${BACKEND_URL}/api/v1/auth/login" \
  -H "Origin: ${FRONTEND_ORIGIN}" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v 2>&1 | grep -i "access-control"

echo ""
echo "2. Testing POST request:"
curl -X POST "${BACKEND_URL}/api/v1/auth/login" \
  -H "Origin: ${FRONTEND_ORIGIN}" \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}' \
  -v 2>&1 | grep -i "access-control"

echo ""
echo "Done!"
```

Run it:
```bash
chmod +x test-cors.sh
./test-cors.sh
```

## Common Issues

### Issue: `Access-Control-Allow-Credentials` header missing
**Solution**: Ensure `CORS_ALLOW_CREDENTIALS=true` is set in Fly.io secrets

### Issue: `Access-Control-Allow-Origin` is `*` instead of your domain
**Solution**: When `AllowCredentials: true`, you cannot use `*` for origins. Set `CORS_ORIGINS` to your specific domain(s)

### Issue: Cookie not being sent
**Solution**: 
1. Verify `SameSite=None` and `Secure=true` in cookie settings (already fixed)
2. Ensure frontend uses `withCredentials: true` in axios config
3. Verify CORS allows credentials

## Expected Configuration

Your backend should have:
- `CORS_ORIGINS` = `https://devhive.it.com` (or comma-separated list)
- `CORS_ALLOW_CREDENTIALS` = `true`
- Cookies set with `SameSite=None` and `Secure=true` (already fixed)




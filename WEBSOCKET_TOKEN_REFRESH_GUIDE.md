# WebSocket Token Refresh Guide

## Problem

WebSocket connections are failing with **401 Unauthorized** errors because access tokens expire after **15 minutes**. The NOTIFY listener and WebSocket infrastructure are working correctly - the issue is expired tokens.

## Solution: Token Refresh Before WebSocket Connection

### Current Token Configuration

- **Access Token**: 15 minutes (short-lived for security)
- **Refresh Token**: 7 days (long-lived, stored in httpOnly cookie)

### Frontend Implementation Pattern

#### 1. Check Token Before Connecting

Before opening a WebSocket connection, check if the token is expired or near expiry:

```typescript
import { jwtDecode } from 'jwt-decode'; // or use your JWT library

interface TokenPayload {
  sub: string; // user ID
  exp: number; // expiration timestamp (Unix)
  iat: number; // issued at timestamp
}

function isTokenExpired(token: string): boolean {
  try {
    const decoded = jwtDecode<TokenPayload>(token);
    const now = Math.floor(Date.now() / 1000);
    // Consider token expired if it expires within 30 seconds
    return decoded.exp < (now + 30);
  } catch (error) {
    return true; // If we can't decode, consider it invalid
  }
}

async function ensureValidToken(): Promise<string> {
  const currentToken = getAccessToken(); // Your token storage method
  
  if (!currentToken || isTokenExpired(currentToken)) {
    // Token is expired or missing, refresh it
    const newToken = await refreshAccessToken();
    return newToken;
  }
  
  return currentToken;
}
```

#### 2. Refresh Token Function

```typescript
async function refreshAccessToken(): Promise<string> {
  try {
    // Refresh endpoint uses refresh_token cookie automatically
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      credentials: 'include', // Important: sends cookies
    });
    
    if (!response.ok) {
      throw new Error('Failed to refresh token');
    }
    
    const data = await response.json();
    const newToken = data.token;
    
    // Store the new access token
    setAccessToken(newToken);
    
    return newToken;
  } catch (error) {
    // Refresh failed - user needs to log in again
    console.error('Token refresh failed:', error);
    // Redirect to login or handle re-authentication
    redirectToLogin();
    throw error;
  }
}
```

#### 3. WebSocket Connection with Token Refresh

```typescript
async function connectWebSocket(projectId: string): Promise<WebSocket> {
  // Ensure we have a valid token before connecting
  const token = await ensureValidToken();
  
  // Use Authorization header (most secure) instead of query parameter
  const wsUrl = `wss://devhive-go-backend.fly.dev/api/v1/messages/ws?project_id=${projectId}`;
  
  const ws = new WebSocket(wsUrl, [], {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  // Handle connection errors
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
  
  ws.onclose = (event) => {
    if (event.code === 1008 || event.code === 1002) {
      // Policy violation or protocol error - might be expired token
      console.log('WebSocket closed, token may be expired. Refreshing...');
      // Retry connection after refresh
      setTimeout(async () => {
        await connectWebSocket(projectId);
      }, 1000);
    }
  };
  
  return ws;
}
```

#### 4. Automatic Reconnection on Token Expiry

```typescript
class WebSocketManager {
  private ws: WebSocket | null = null;
  private projectId: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  
  constructor(projectId: string) {
    this.projectId = projectId;
  }
  
  async connect(): Promise<void> {
    try {
      // Always refresh token before connecting
      const token = await ensureValidToken();
      
      const wsUrl = `wss://devhive-go-backend.fly.dev/api/v1/messages/ws?project_id=${this.projectId}`;
      
      this.ws = new WebSocket(wsUrl, [], {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0; // Reset on successful connection
      };
      
      this.ws.onclose = async (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        
        // If closed due to authentication error, refresh and reconnect
        if (event.code === 1008 || event.code === 1002) {
          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Reconnecting (attempt ${this.reconnectAttempts})...`);
            await new Promise(resolve => setTimeout(resolve, 1000 * this.reconnectAttempts));
            await this.connect();
          } else {
            console.error('Max reconnection attempts reached');
          }
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
      
      this.ws.onmessage = (event) => {
        // Handle cache invalidation messages
        const message = JSON.parse(event.data);
        handleCacheInvalidation(message);
      };
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      throw error;
    }
  }
  
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
```

## Security Improvements

### ✅ Recommended: Use Authorization Header

**Before (Less Secure)**:
```typescript
// Token in query string - appears in logs, browser history
const ws = new WebSocket(`wss://.../ws?project_id=${id}&token=${token}`);
```

**After (More Secure)**:
```typescript
// Token in Authorization header - more secure
const ws = new WebSocket(wsUrl, [], {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

### ✅ Alternative: Cookie-Based Authentication

The backend now supports cookie-based authentication. You can set an `access_token` cookie (httpOnly, secure) and the WebSocket handler will use it automatically.

## Backend Changes Made

1. **Improved Error Messages**: WebSocket handler now provides specific error messages for expired tokens
2. **Multiple Auth Methods**: Supports Authorization header, cookie, or query parameter (in order of preference)
3. **Better Logging**: Warnings when tokens are provided via query parameter (less secure)

## Testing

### Test Token Expiration

1. Log in and get an access token
2. Wait 16 minutes (or manually expire the token)
3. Try to connect WebSocket - should get clear error message
4. Refresh token and reconnect - should work

### Test Automatic Refresh

1. Connect WebSocket with a token that expires in 1 minute
2. Wait for token to expire
3. WebSocket should automatically refresh and reconnect

## Error Messages

The backend now returns specific error messages:

- **"Authentication token has expired. Please refresh your token and reconnect."** - Token is expired
- **"Invalid authentication token format"** - Token is malformed
- **"Invalid authentication token"** - Other validation errors

## Next Steps

1. ✅ Backend: Improved error messages and multiple auth methods
2. ⏳ Frontend: Implement token refresh before WebSocket connection
3. ⏳ Frontend: Use Authorization header instead of query parameter
4. ⏳ Frontend: Implement automatic reconnection on token expiry

## Configuration

Token expiration can be configured via environment variables:

- `JWT_EXPIRATION_MINUTES` - Access token lifetime (default: 15 minutes)
- `JWT_REFRESH_EXPIRATION_DAYS` - Refresh token lifetime (default: 7 days)

To change access token to 24 hours (less secure, but simpler):

```bash
export JWT_EXPIRATION_MINUTES=1440  # 24 hours
```

However, **recommended approach** is to keep 15-minute tokens and implement refresh logic.




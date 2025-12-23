import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

// Create axios instance with base configuration
const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'https://devhive-go-backend.fly.dev/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Important: allows cookies (refresh token) to be sent
});

// Auth routes that should NEVER have auth headers or token refresh
// These routes are public and should not require authentication
const AUTH_ROUTES = [
  '/auth/login',
  '/auth/register', // If you have a register endpoint
  '/auth/refresh',
  '/auth/logout',
  '/auth/password/reset-request',
  '/auth/password/reset',
  '/users/validate-email', // Public email validation
  '/users/validate-username', // Public username validation
  '/invites/', // GET /invites/{token} - public invite details
];

// Public routes that need special handling (check method)
const PUBLIC_ROUTES = [
  { path: '/users', methods: ['POST'] }, // POST /users (registration) is public
];

// Check if a URL is an auth route
function isAuthRoute(url: string | undefined, method?: string): boolean {
  if (!url) return false;
  
  // Check exact auth routes
  if (AUTH_ROUTES.some(route => url.includes(route))) {
    return true;
  }
  
  // Check public routes with method
  if (method) {
    const upperMethod = method.toUpperCase();
    return PUBLIC_ROUTES.some(route => 
      url.includes(route.path) && route.methods.includes(upperMethod)
    );
  }
  
  return false;
}

// Store for tracking refresh attempts to prevent loops
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

// Request interceptor: Add access token to requests (skip for auth routes)
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Skip auth handling for auth routes
    if (isAuthRoute(config.url, config.method)) {
      return config;
    }

    // Get access token from memory (stored by auth context)
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

// Response interceptor: Handle token refresh on 401 (skip for auth routes)
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Skip token refresh for auth routes - they should return 401 directly
    if (isAuthRoute(originalRequest.url, originalRequest.method)) {
      return Promise.reject(error);
    }

    // If error is 401 and we haven't tried refreshing yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue this request
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

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Attempt to refresh token
        const response = await axios.post(
          `${process.env.REACT_APP_API_URL || 'https://devhive-go-backend.fly.dev/api/v1'}/auth/refresh`,
          {},
          { withCredentials: true } // Send refresh token cookie
        );

        // Backend returns: { token: "...", userId: "..." }
        const { token } = response.data;
        
        if (!token) {
          throw new Error('Refresh response missing token');
        }
        
        // Store new access token (both in-memory and localStorage)
        setAccessToken(token);
        
        // Store userId if provided
        if (response.data.userId) {
          localStorage.setItem(USER_ID_KEY, response.data.userId);
        }
        
        // Process queued requests
        processQueue(null, token);
        
        // Retry original request with new token
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${token}`;
        }
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed - log error for debugging
        const error = refreshError as AxiosError;
        console.error('Token refresh failed:', {
          status: error.response?.status,
          data: error.response?.data,
          message: error.message,
        });
        
        // Clear auth and redirect to login
        processQueue(error, null);
        clearAuth();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// Token management functions (these will be implemented by auth context)
// Primary: In-memory storage (security best practice - reduces XSS exposure)
let accessToken: string | null = null;

// Secondary: localStorage keys for persistence across page reloads
const TOKEN_KEY = 'token';
const TOKEN_EXPIRATION_KEY = 'tokenExpiration';
const USER_ID_KEY = 'userId';
// Token validity duration: 24 hours (access tokens expire in 15 min, but we store for 24h for persistence)
const TOKEN_VALIDITY_DURATION = 24 * 60 * 60 * 1000; // 24 hours in milliseconds

function getAccessToken(): string | null {
  // First check in-memory (primary)
  if (accessToken) {
    return accessToken;
  }
  
  // Fallback to localStorage (for persistence across page reloads)
  const storedToken = localStorage.getItem(TOKEN_KEY);
  if (storedToken) {
    // Check if token is still valid (not expired)
    const expirationTime = localStorage.getItem(TOKEN_EXPIRATION_KEY);
    if (expirationTime) {
      const expirationTimestamp = parseInt(expirationTime, 10);
      if (Date.now() < expirationTimestamp) {
        // Restore to memory and return
        accessToken = storedToken;
        return storedToken;
      } else {
        // Token expired, clear it
        clearAuth();
        return null;
      }
    }
    // No expiration stored, assume valid and restore to memory
    accessToken = storedToken;
    return storedToken;
  }
  
  return null;
}

function setAccessToken(token: string | null): void {
  // Update in-memory (primary)
  accessToken = token;
  
  // Also update localStorage for backward compatibility and persistence
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
    // Store expiration timestamp (24 hours from now)
    const expirationTime = Date.now() + TOKEN_VALIDITY_DURATION;
    localStorage.setItem(TOKEN_EXPIRATION_KEY, expirationTime.toString());
  } else {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(TOKEN_EXPIRATION_KEY);
  }
}

function clearAuth(): void {
  // Clear in-memory
  accessToken = null;
  
  // Clear localStorage
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(TOKEN_EXPIRATION_KEY);
  localStorage.removeItem(USER_ID_KEY);
  localStorage.removeItem('auth_state');
}

// Initialize token from localStorage on module load (for persistence across page reloads)
// This restores the token to memory if it exists and is still valid
if (typeof window !== 'undefined') {
  const storedToken = localStorage.getItem(TOKEN_KEY);
  const expirationTime = localStorage.getItem(TOKEN_EXPIRATION_KEY);
  
  if (storedToken && expirationTime) {
    const expirationTimestamp = parseInt(expirationTime, 10);
    if (Date.now() < expirationTimestamp) {
      // Token is still valid, restore to memory
      accessToken = storedToken;
    } else {
      // Token expired, clear it
      clearAuth();
    }
  }
}

// Export functions to be used by auth context
export const tokenManager = {
  getAccessToken,
  setAccessToken,
  clearAuth,
};

export default apiClient;




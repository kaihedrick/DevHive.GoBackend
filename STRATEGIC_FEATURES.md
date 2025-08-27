# DevHive Go Backend - Strategic MVP Features

## Overview
This document outlines the strategic MVP features implemented in the DevHive Go Backend, designed to provide a robust foundation for real-time collaboration, feature management, and mobile optimization.

## 🚀 Core Strategic Features

### 1. WebSocket Real-Time Communication
**Location**: `internal/ws/ws.go`

#### Features:
- **Hub Pattern**: Scalable WebSocket management with centralized message broadcasting
- **Project-Specific Broadcasting**: Messages are routed to specific project channels
- **Client Lifecycle Management**: Automatic connection cleanup and heartbeat monitoring
- **Dual Endpoints**:
  - `/ws` - Basic WebSocket connection (for public updates)
  - `/ws/auth` - Authenticated WebSocket with JWT validation

#### Usage:
```javascript
// Basic connection
const ws = new WebSocket('ws://localhost:8080/ws?user_id=123&project_id=456');

// Authenticated connection
const ws = new WebSocket('ws://localhost:8080/ws/auth?token=JWT_TOKEN&project_id=456');
```

#### Broadcasting Functions:
```go
// Broadcast sprint updates
ws.BroadcastSprintUpdate(projectID, sprintData)

// Broadcast project updates
ws.BroadcastProjectUpdate(projectID, projectData)

// Broadcast message updates
ws.BroadcastMessageUpdate(projectID, messageData)
```

### 2. Feature Flags System
**Location**: `internal/flags/flags.go`

#### Features:
- **Database Persistence**: Feature flags stored in PostgreSQL with caching
- **Global Manager**: Singleton instance for application-wide access
- **Background Synchronization**: Automatic cache refresh every 5 minutes
- **Admin Management**: Full CRUD operations for feature flags

#### Default Flags:
- `new_sprint_ui` - New sprint user interface
- `enable_websockets` - WebSocket functionality
- `mobile_v2_api` - Mobile API endpoints
- `rate_limiting` - API rate limiting
- `gzip_compression` - Response compression

#### Usage:
```go
// Check if feature is enabled
if flags.IsEnabledGlobal("mobile_v2_api") {
    // Enable mobile-specific functionality
}

// Update feature flag
flags.GlobalManager.SetFlag("new_sprint_ui", true)
```

#### API Endpoints:
```
GET    /api/v1/admin/feature-flags          - List all flags
GET    /api/v1/admin/feature-flags/:key     - Get specific flag
POST   /api/v1/admin/feature-flags          - Create new flag
PUT    /api/v1/admin/feature-flags/:key     - Update flag
DELETE /api/v1/admin/feature-flags/:key     - Delete flag
POST   /api/v1/admin/feature-flags/bulk-update - Bulk update
```

### 3. Mobile-Optimized APIs
**Location**: `controllers/mobile_api.go`

#### Features:
- **Minimal Payloads**: Optimized data structures for mobile consumption
- **Feature Flag Gating**: Mobile APIs can be toggled on/off
- **Pagination Support**: Efficient data retrieval with limits and offsets
- **Progress Calculations**: Real-time sprint progress and days remaining

#### Mobile-Specific Endpoints:
```
GET /api/v1/mobile/v2/projects                    - Mobile-optimized projects
GET /api/v1/mobile/v2/projects/:id                - Mobile project details
GET /api/v1/mobile/v2/projects/:projectId/sprints - Mobile sprints
GET /api/v1/mobile/v2/projects/:projectId/messages - Mobile messages
```

#### Data Structures:
```go
type MobileSprint struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Status      string    `json:"status"`
    DaysLeft    int       `json:"days_left"`
    Progress    float64   `json:"progress"`
}
```

### 4. Advanced Middleware Stack
**Location**: `internal/middleware/rate_limit.go`

#### Features:
- **Multi-Strategy Rate Limiting**: IP-based, user-based, and mobile-specific limits
- **Feature Flag Integration**: Rate limiting can be toggled globally
- **HTTP Headers**: Standard rate limit headers for client awareness
- **Configurable Limits**: Different limits for different use cases

#### Rate Limit Configurations:
```go
DefaultRateLimit:     100 requests/minute
StrictRateLimit:      20 requests/minute
MobileRateLimit:      200 requests/minute
```

#### Usage:
```go
// Apply to all routes
router.Use(middleware.RateLimiter(middleware.DefaultRateLimit))

// Apply to specific routes
router.Use(middleware.StrictRateLimit())
router.Use(middleware.MobileRateLimit())
```

### 5. Database Schema Enhancements
**Location**: `db/schema.sql`

#### Features:
- **Feature Flags Table**: Dedicated table for feature management
- **Automatic Timestamps**: Created/updated triggers for all tables
- **Performance Indexes**: Optimized queries for common operations
- **Default Data**: Pre-populated feature flags and sample data

## 🔧 Configuration & Environment

### Environment Variables:
```bash
# Database (Fly.io)
ConnectionStringsDbConnection=postgresql://user:pass@host:5432/devhive

# Server
PORT=8080
GIN_MODE=release

# Firebase (Fly.io)
FIREBASE_JSON_BASE64=base64_encoded_service_account_json
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_SERVICE_ACCOUNT_KEY_PATH=/path/to/service-account.json

# JWT (Fly.io)
JwtKey=your-jwt-secret-key
JwtIssuer=your-jwt-issuer
JwtAudience=your-jwt-audience

# Email (Fly.io)
MailgunApiKey=your-mailgun-api-key
MailgunDomain=your-mailgun-domain
Mailgun__SenderEmail=your-sender-email
```

### Feature Flag Management:
```bash
# Enable mobile API
curl -X PUT http://localhost:8080/api/v1/admin/feature-flags/mobile_v2_api \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'

# Check feature status
curl http://localhost:8080/api/v1/admin/feature-flags/mobile_v2_api \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🚀 Getting Started

### 1. Database Setup:
```sql
-- Run the schema
psql -d devhive -f db/schema.sql

-- Verify feature flags
SELECT * FROM feature_flags;
```

### 2. Start the Server:
```bash
go run cmd/main.go
```

### 3. Test WebSocket Connection:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?user_id=test&project_id=test');
ws.onmessage = (event) => console.log('Received:', event.data);
```

### 4. Test Feature Flags:
```bash
curl http://localhost:8080/api/v1/admin/feature-flags \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📊 Performance & Scalability

### WebSocket Performance:
- **Connection Pooling**: Efficient client management
- **Message Buffering**: 256-byte message channels per client
- **Heartbeat Monitoring**: 54-second ping intervals
- **Automatic Cleanup**: Dead connection detection and removal

### Feature Flag Performance:
- **5-Minute Caching**: Reduces database queries
- **Background Sync**: Non-blocking flag updates
- **Memory Storage**: In-memory cache for fast lookups

### Rate Limiting Performance:
- **Memory Store**: Fast in-memory rate limiting
- **User-Based Limiting**: Personalized limits per user
- **Mobile Optimization**: Higher limits for mobile clients

## 🔒 Security Features

### Authentication:
- **JWT Validation**: Secure WebSocket connections
- **Project Access Control**: User authorization per project
- **Firebase Integration**: Enterprise-grade authentication

### Rate Limiting:
- **IP-Based Limiting**: Protection against abuse
- **User-Based Limiting**: Personalized limits
- **Configurable Thresholds**: Adjustable per endpoint

## 🧪 Testing & Development

### Feature Flag Testing:
```go
// Enable feature for testing
flags.GlobalManager.SetFlag("test_feature", true)

// Verify feature is enabled
if flags.IsEnabledGlobal("test_feature") {
    // Run test code
}
```

### WebSocket Testing:
```go
// Test broadcasting
ws.BroadcastToProject("test-project", "test_message", "Hello World")

// Verify client receives message
// (Use WebSocket client to test)
```

## 📈 Monitoring & Observability

### Logging:
- **WebSocket Events**: Connection, disconnection, and message logs
- **Feature Flag Changes**: Flag updates and cache refreshes
- **Rate Limit Violations**: Exceeded limit attempts

### Metrics:
- **Active Connections**: Real-time WebSocket client count
- **Feature Flag Usage**: Flag check frequency
- **Rate Limit Hits**: Limit violation counts

## 🚀 Future Enhancements

### Planned Features:
1. **Redis Integration**: Distributed WebSocket hubs
2. **Metrics Dashboard**: Real-time system monitoring
3. **A/B Testing**: Feature flag-based experimentation
4. **Webhook System**: External system notifications
5. **GraphQL Support**: Flexible data querying

### Scalability Improvements:
1. **Horizontal Scaling**: Multiple WebSocket servers
2. **Message Queuing**: Reliable message delivery
3. **Load Balancing**: Distributed request handling
4. **Caching Layer**: Redis-based caching

## 📚 API Reference

### WebSocket Events:
```json
{
  "type": "sprint_update",
  "data": {...},
  "project_id": "uuid",
  "user_id": "uuid"
}
```

### Feature Flag Response:
```json
{
  "key": "mobile_v2_api",
  "enabled": true,
  "description": "Enable mobile v2 API endpoints",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Mobile API Response:
```json
{
  "sprints": [...],
  "count": 5,
  "pagination": {
    "total": 25,
    "limit": 10,
    "offset": 0,
    "has_more": true
  }
}
```

---

**DevHive Go Backend** - Enterprise-grade collaboration platform with real-time capabilities, feature management, and mobile optimization.

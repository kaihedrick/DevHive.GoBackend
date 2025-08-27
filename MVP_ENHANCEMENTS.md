# DevHive MVP Enhancements

This document outlines the strategic enhancements implemented for the DevHive Go backend, focusing on real-time capabilities, clean architecture, feature flags, and mobile optimization.

## 🚀 Overview

The MVP enhancements introduce four key architectural improvements:

1. **WebSocket Layer** - Real-time sprint updates and project collaboration
2. **Service Layer** - Clean architecture with business logic separation
3. **Feature Flags** - Runtime feature toggling and A/B testing capabilities
4. **Mobile APIs** - Optimized endpoints for mobile clients with rate limiting

## 🔌 WebSocket Layer (`internal/ws/`)

### Features
- Real-time bidirectional communication
- Project-scoped message broadcasting
- Automatic connection management with heartbeat
- Scalable hub architecture for multiple clients

### Usage
```go
// Start WebSocket hub in main.go
ws.StartWebSocketHub()

// Broadcast updates to project members
ws.BroadcastSprintUpdate(projectID, sprintData)
ws.BroadcastProjectUpdate(projectID, projectData)
ws.BroadcastMessageUpdate(projectID, messageData)
```

### WebSocket Endpoint
```
GET /ws?user_id={uuid}&project_id={uuid}
```

### Message Types
- `sprint_update` - Sprint status changes, creation, updates
- `project_update` - Project modifications, member changes
- `message_update` - New messages, edits, deletions

## 🏗️ Service Layer (`internal/service/`)

### Architecture Benefits
- **Separation of Concerns** - Business logic separated from HTTP handlers
- **Testability** - Services can be easily unit tested
- **Reusability** - Services can be used by multiple controllers
- **WebSocket Integration** - Automatic real-time updates on data changes

### Sprint Service Example
```go
type SprintService interface {
    GetSprintsForProject(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
    CreateSprint(ctx context.Context, req models.SprintCreateRequest, projectID, userID uuid.UUID) (*models.Sprint, error)
    StartSprint(ctx context.Context, sprintID, userID uuid.UUID) error
    CompleteSprint(ctx context.Context, sprintID, userID uuid.UUID) error
}

// Usage
sprintService := service.NewSprintService(db.GetDB())
sprint, err := sprintService.CreateSprint(ctx, req, projectID, userID)
```

### Automatic WebSocket Broadcasting
All service operations automatically broadcast updates to connected clients:
- Sprint creation → `sprint_update` with `action: "created"`
- Sprint updates → `sprint_update` with `action: "updated"`
- Sprint deletion → `sprint_update` with `action: "deleted"`

## 🚩 Feature Flags (`internal/flags/`)

### Database Schema
```sql
CREATE TABLE feature_flags (
    key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Default Flags
- `new_sprint_ui` - New sprint user interface (disabled by default)
- `enable_websockets` - WebSocket real-time updates (enabled by default)
- `mobile_v2_api` - Mobile v2 API endpoints (disabled by default)
- `rate_limiting` - API rate limiting (enabled by default)
- `gzip_compression` - Response compression (enabled by default)

### Usage
```go
// Check if feature is enabled
if flags.IsEnabledGlobal("mobile_v2_api") {
    // Enable mobile-specific functionality
}

// Initialize global manager
flags.InitGlobalManager(db.GetDB())

// Runtime flag management
manager := flags.NewManager(db.GetDB())
manager.SetFlag("new_sprint_ui", true)
```

### Benefits
- **Runtime Configuration** - No server restarts required
- **A/B Testing** - Gradual feature rollouts
- **Emergency Rollbacks** - Instant feature disabling
- **Environment-Specific** - Different flags per environment

## 📱 Mobile APIs (`controllers/mobile_api.go`)

### Optimized Data Structures
- **MobileSprint** - Simplified sprint with progress calculation
- **MobileProject** - Project with member count and active sprint
- **MobileMessage** - Message with sender name instead of full user object

### Endpoints
```
GET /api/v1/mobile/v2/projects          - User's projects
GET /api/v1/mobile/v2/projects/:id      - Specific project with active sprint
GET /api/v1/mobile/v2/projects/:projectId/sprints  - Project sprints
GET /api/v1/mobile/v2/projects/:projectId/messages - Project messages
```

### Features
- **Feature Flag Protected** - Only available when `mobile_v2_api` is enabled
- **Rate Limited** - Mobile-specific rate limiting (200 req/min)
- **Optimized Responses** - Smaller payloads, calculated fields
- **Progress Calculation** - Automatic sprint progress based on time

### Example Response
```json
{
  "sprints": [
    {
      "id": "uuid",
      "name": "Sprint 1",
      "status": "active",
      "days_left": 5,
      "progress": 65.5
    }
  ],
  "count": 1
}
```

## 🛡️ Rate Limiting (`internal/middleware/rate_limit.go`)

### Configurations
- **Default** - 100 requests per minute
- **Strict** - 20 requests per minute (sensitive operations)
- **Mobile** - 200 requests per minute (mobile APIs)

### Features
- **Feature Flag Controlled** - Can be disabled via `rate_limiting` flag
- **Multiple Strategies** - IP-based and user-based limiting
- **Standard Headers** - `X-RateLimit-*` headers for client awareness
- **Graceful Degradation** - Returns 429 with retry information

### Usage
```go
// Global rate limiting
router.Use(middleware.RateLimiter(middleware.DefaultRateLimit))

// Route-specific rate limiting
mobile.Use(middleware.MobileRateLimit())

// User-based rate limiting
sensitive.Use(middleware.RateLimitByUser(middleware.StrictRateLimitConfig))
```

## 🔧 Integration in Main Application

### Initialization
```go
func main() {
    // Initialize feature flags
    flags.InitGlobalManager(db.GetDB())
    
    // Start WebSocket hub
    ws.StartWebSocketHub()
    
    // Add middleware based on feature flags
    if flags.IsEnabledGlobal("gzip_compression") {
        router.Use(gzip.Gzip(gzip.DefaultCompression))
    }
    
    if flags.IsEnabledGlobal("rate_limiting") {
        router.Use(middleware.RateLimiter(middleware.DefaultRateLimit))
    }
}
```

### WebSocket Integration
```go
if flags.IsEnabledGlobal("enable_websockets") {
    router.GET("/ws", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
        ws.HandleConnections(ws.GlobalHub, w, r)
    }))
}
```

### Mobile API Routes
```go
if flags.IsEnabledGlobal("mobile_v2_api") {
    mobile := protected.Group("/mobile/v2")
    mobile.Use(middleware.MobileRateLimit())
    // ... mobile endpoints
}
```

## 🧪 Testing and Development

### Feature Flag Testing
```go
// Enable feature for testing
manager := flags.NewManager(testDB)
manager.SetFlag("mobile_v2_api", true)

// Test mobile API endpoints
// ... test code
```

### WebSocket Testing
```javascript
// Client-side WebSocket connection
const ws = new WebSocket('ws://localhost:8080/ws?user_id=123&project_id=456');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'sprint_update') {
        console.log('Sprint updated:', data.data);
    }
};
```

## 📊 Performance Considerations

### WebSocket
- **Connection Limits** - Configurable per client
- **Message Size** - 512 bytes maximum per message
- **Heartbeat** - 54-second ping/pong for connection health
- **Memory Management** - Automatic cleanup of disconnected clients

### Feature Flags
- **Caching** - 5-minute cache with background sync
- **Database Queries** - Minimal impact with indexed lookups
- **Memory Usage** - Lightweight in-memory cache

### Rate Limiting
- **Memory Store** - Fast in-memory rate limiting
- **Configurable** - Different limits for different use cases
- **Headers** - Client can adapt based on rate limit information

## 🚀 Deployment and Configuration

### Environment Variables
```bash
# Feature flag overrides (optional)
FEATURE_FLAG_MOBILE_V2_API=true
FEATURE_FLAG_NEW_SPRINT_UI=false
```

### Database Migration
```bash
# Run schema updates
psql -d devhive -f db/schema.sql
```

### Monitoring
- WebSocket connection count
- Feature flag usage statistics
- Rate limiting metrics
- API response times

## 🔮 Future Enhancements

### Planned Features
- **Redis Backend** - Distributed WebSocket and rate limiting
- **Analytics** - Feature flag usage tracking
- **Admin UI** - Feature flag management interface
- **WebSocket Clustering** - Multi-server WebSocket support
- **Advanced Rate Limiting** - Token bucket algorithms
- **Mobile Push Notifications** - WebSocket + push integration

### Scalability Considerations
- **Horizontal Scaling** - Load balancer with sticky sessions
- **Database Sharding** - Feature flags and user data distribution
- **CDN Integration** - Static asset optimization
- **Microservices** - Service layer extraction to separate services

## 📚 Additional Resources

- [Gorilla WebSocket Documentation](https://github.com/gorilla/websocket)
- [Gin Framework](https://gin-gonic.com/)
- [Feature Flags Best Practices](https://martinfowler.com/articles/feature-toggles.html)
- [Rate Limiting Strategies](https://cloud.google.com/architecture/rate-limiting-strategies-techniques)

---

**Note**: This implementation follows Go best practices and provides a solid foundation for scaling the DevHive platform. All enhancements are feature-flag protected and can be enabled/disabled at runtime without code changes.

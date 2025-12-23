# DevHive Backend - Project Architecture

## Related Documentation
- [Database Schema](./database_schema.md) - Complete database schema and relationships
- [Authentication Flow](./authentication_flow.md) - JWT authentication and refresh token mechanism
- [Real-time System](./realtime_system.md) - WebSocket and cache invalidation architecture
- [SOP: Adding Migrations](../SOP/adding_migrations.md)
- [SOP: Adding API Endpoints](../SOP/adding_api_endpoints.md)

## Project Overview

**DevHive Backend** is a modern, scalable Go backend for project management and team collaboration. It provides comprehensive APIs for user management, project organization, sprint planning, task tracking, and real-time messaging.

### Core Purpose
- Enable teams to collaborate on projects with role-based access control
- Support agile sprint planning and task management
- Provide real-time updates through WebSocket connections
- Offer secure authentication with JWT tokens and refresh mechanism

### Tech Stack

#### Backend Framework
- **Go 1.25** - Primary language
- **Chi v5** - HTTP router (lightweight, idiomatic)
- **SQLC** - Type-safe SQL code generation
- **PostgreSQL 12+** - Primary database with UUID support

#### Key Libraries
- `jackc/pgx/v5` - PostgreSQL driver and connection pooling
- `golang-jwt/jwt/v5` - JWT token generation and validation
- `gorilla/websocket` - WebSocket support for real-time features
- `go-chi/cors` - CORS middleware
- `go-chi/httprate` - Rate limiting (100 req/min per IP)
- `golang.org/x/crypto` - bcrypt password hashing
- `joho/godotenv` - Environment variable management

#### Optional Integrations
- Firebase Auth (v4.18.0) - Alternative authentication provider
- Firebase Storage - File storage (with Fly.io volume fallback)
- Mailgun - Email delivery
- gRPC - Optional gRPC API endpoints

## Project Structure

```
DevHive.GoBackend/
├── cmd/
│   └── devhive-api/
│       ├── main.go                 # Application entry point
│       └── migrations/             # Database migration files (SQL)
├── internal/                       # Private application code
│   ├── config/                     # Configuration management
│   │   └── config.go              # Environment-based config loading
│   ├── repo/                       # SQLC-generated database layer
│   │   ├── db.go                  # Database interface
│   │   ├── models.go              # Auto-generated models
│   │   └── queries.sql.go         # Auto-generated query functions
│   ├── http/
│   │   ├── handlers/              # HTTP request handlers
│   │   │   ├── auth.go            # Authentication endpoints
│   │   │   ├── user.go            # User management
│   │   │   ├── project.go         # Project CRUD and invites
│   │   │   ├── sprint.go          # Sprint management
│   │   │   ├── task.go            # Task management
│   │   │   ├── message.go         # Messaging and WebSocket
│   │   │   ├── mail.go            # Email sending
│   │   │   └── migration.go       # Migration utilities (dev/admin)
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT authentication middleware
│   │   │   └── cache.go           # Cache invalidation middleware
│   │   ├── response/
│   │   │   └── response.go        # Standardized HTTP responses
│   │   └── router/
│   │       └── router.go          # Route definitions and setup
│   ├── ws/                         # WebSocket system
│   │   ├── hub.go                 # WebSocket hub and client management
│   │   └── handlers.go            # WebSocket message handlers
│   ├── db/                         # Database utilities
│   │   ├── notify_listener.go     # PostgreSQL NOTIFY listener
│   │   └── postgres.go            # Database connection and migration runner
│   ├── grpc/                       # gRPC server (optional)
│   └── auth/                       # Authentication utilities
├── db/                             # Database layer
│   ├── migrate.go                 # Migration runner
│   └── postgres.go                # DB connection utilities
├── config/                         # Legacy config (deprecated)
├── api/v1/                         # gRPC proto definitions
├── frontend-examples/              # Frontend integration examples
│   └── src/
│       ├── lib/
│       │   └── apiClient.ts       # Axios client with token refresh
│       ├── hooks/
│       │   └── useInvites.ts      # Example React hooks
│       └── components/
│           ├── ProjectInvites.tsx  # Invite management UI
│           └── ProjectPage.example.tsx
├── docs/                           # Additional documentation
├── .agent/                         # AI assistant documentation (this folder)
│   ├── System/                    # Architecture and design docs
│   ├── Tasks/                     # Feature PRDs and implementation plans
│   ├── SOP/                       # Standard operating procedures
│   └── README.md                  # Documentation index
├── go.mod                          # Go module dependencies
├── sqlc.yaml                       # SQLC configuration
├── Dockerfile                      # Container image definition
├── fly.toml                        # Fly.io deployment config
└── Makefile                        # Build and development tasks
```

## Architecture Layers

### 1. HTTP Layer (Entry Point)
- **Router** (`internal/http/router/router.go`): Chi-based HTTP router with middleware chain
- **Handlers** (`internal/http/handlers/`): Request processing and business logic
- **Middleware**:
  - Request ID, logging, recovery (Chi built-in)
  - Rate limiting (100 req/min per IP)
  - CORS (configurable origins)
  - JWT authentication (`middleware.RequireAuth`)

### 2. Data Layer
- **SQLC Repository** (`internal/repo/`): Type-safe SQL queries auto-generated from SQL files
- **Database Connection**: pgx v5 connection pool (25 max open, 5 max idle, 5min lifetime)
- **Migrations**: Sequential SQL files in `cmd/devhive-api/migrations/`

### 3. Real-time Layer
- **WebSocket Hub** (`internal/ws/hub.go`): Central hub managing all WebSocket connections
- **PostgreSQL NOTIFY** (`internal/db/notify_listener.go`): Database triggers broadcast changes
- **Cache Invalidation**: Triggers on INSERT/UPDATE/DELETE notify connected clients

### 4. gRPC Layer (Optional)
- Separate gRPC server on port 8081
- Proto definitions in `api/v1/`
- Parallel to HTTP API (not widely used currently)

## Core Features

### Authentication & Authorization
- **JWT-based authentication** with access tokens (15min) and refresh tokens (7 days)
- **Refresh token rotation**: Stored in PostgreSQL, validated on refresh
- **Password hashing**: bcrypt with automatic salt
- **Role-based access control**: Project-level roles (owner, admin, member, viewer)
- **Public endpoints**: Email/username validation, invite details, password reset

### User Management
- User registration with email/username validation
- Profile management (first name, last name, bio, avatar)
- Password reset flow with time-limited tokens
- Active/inactive user status

### Project Management
- CRUD operations for projects
- Project ownership and member management
- Role-based permissions per project
- Time-limited invite links with optional max uses
- Public invite acceptance flow

### Sprint Planning
- Sprint CRUD with start/end dates
- Sprint status tracking (is_started, is_completed)
- Association with projects
- Task assignment to sprints

### Task Tracking
- Task CRUD operations
- Status management (integer-based status codes)
- Sprint and assignee association
- Project-level task listing

### Messaging & Real-time
- Project-based threaded messaging
- WebSocket connections for real-time updates
- Message types: text, image, file
- Parent-child message relationships (threading)

## Configuration

Configuration is loaded from environment variables with fallback defaults (see `internal/config/config.go`):

```go
type Config struct {
    Port          string           // HTTP port (default: 8080)
    GRPCPort      string           // gRPC port (default: 8081)
    DatabaseURL   string           // PostgreSQL connection string
    JWT           JWTConfig        // JWT signing key, expiration
    CORS          CORSConfig       // Allowed origins, credentials
    Mail          MailConfig       // Mailgun API key, domain, sender
    AdminPassword string           // Admin verification password
}
```

### Environment Variables
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SIGNING_KEY` - Secret for JWT token signing
- `JWT_EXPIRATION_MINUTES` - Access token lifetime (default: 15)
- `JWT_REFRESH_EXPIRATION_DAYS` - Refresh token lifetime (default: 7)
- `CORS_ORIGINS` - Comma-separated allowed origins
- `ADMIN_CERTIFICATES_PASSWORD` - Admin verification password
- `MAILGUN_API_KEY`, `MAILGUN_DOMAIN`, `MAILGUN_SENDER` - Email config

## Database Design Principles

### Core Tables
- **users** - User accounts with bcrypt password hashing
- **projects** - Project entities with owner reference
- **project_members** - Many-to-many with role (owner, admin, member, viewer)
- **project_invites** - Time-limited invite links with usage tracking
- **sprints** - Sprint planning with dates and status
- **tasks** - Task entities with sprint/assignee associations
- **messages** - Threaded messaging with parent reference
- **refresh_tokens** - Persistent refresh token storage
- **password_resets** - Time-limited password reset tokens

### Key Design Patterns
- **UUID primary keys** across all tables for distributed system compatibility
- **Soft deletes** not implemented (hard deletes with CASCADE)
- **Timestamp tracking** (created_at, updated_at with triggers)
- **Indexed foreign keys** for query performance
- **CITEXT for emails** - Case-insensitive email storage
- **JSON payloads in NOTIFY** - Minimal, targeted cache invalidation

## API Design

### RESTful Conventions
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{projectId}` - Get project details
- `PATCH /api/v1/projects/{projectId}` - Update project
- `DELETE /api/v1/projects/{projectId}` - Delete project

### Nested Resources
- `GET /api/v1/projects/{projectId}/sprints` - List project sprints
- `POST /api/v1/projects/{projectId}/sprints` - Create sprint
- `GET /api/v1/projects/{projectId}/tasks` - List project tasks
- `GET /api/v1/projects/{projectId}/messages` - List project messages

### Authentication
- **Public endpoints**: `/auth/login`, `/auth/refresh`, `/users` (POST), `/invites/{token}` (GET)
- **Protected endpoints**: All others require `Authorization: Bearer <token>` header
- **Token refresh**: Automatic retry with refresh token on 401 responses (see frontend apiClient.ts)

### Response Format
Standardized JSON responses:
```json
{
  "message": "Success message",
  "data": { ... }
}
```

Error responses:
```json
{
  "error": "Error message"
}
```

## Deployment Architecture

### Fly.io Deployment
- **Platform**: Fly.io with Docker containerization
- **Database**: Fly.io PostgreSQL cluster
- **Storage**: Fly.io persistent volumes for file uploads
- **Regions**: Configurable (default: DFW)
- **Scaling**: Horizontal scaling supported (stateless design)

### Health Checks
- `GET /health` - Basic health check
- `GET /healthz` - Liveness probe
- `GET /readyz` - Readiness probe (checks DB connection)

### CI/CD Pipeline
- GitHub Actions workflows in `.github/workflows/`
- Automated testing on push
- Container builds and deployments
- Migration verification

## Security Features

### Authentication Security
- bcrypt password hashing (automatic salt)
- Short-lived access tokens (15min)
- Refresh token rotation (7 days)
- Token invalidation on logout
- Secure cookie support (withCredentials)

### API Security
- Rate limiting (100 req/min per IP)
- CORS with configurable origins
- Request ID tracking for audit logs
- SQL injection prevention (parameterized queries via SQLC)
- Input validation on all endpoints

### Database Security
- Foreign key constraints with CASCADE deletes
- Unique constraints on email/username
- Password reset token expiration
- Invite token expiration and usage limits

## Performance Optimizations

### Database
- Connection pooling (pgx v5 pool)
- Indexed foreign keys
- Efficient query patterns via SQLC
- NOTIFY/LISTEN for cache invalidation (vs polling)

### HTTP
- Rate limiting to prevent abuse
- Compression support
- Keep-alive connections
- Request/response timeouts (15s read/write, 60s idle)

### WebSocket
- Centralized hub architecture
- Project-based message filtering
- Automatic ping/pong health checks (54s interval)
- Graceful connection cleanup

## Integration Points

### Frontend Integration
- Example React components in `frontend-examples/`
- Axios client with automatic token refresh
- WebSocket connection management
- TypeScript type definitions (can be generated)

### External Services
- **Mailgun** - Transactional email delivery
- **Firebase** (optional) - Alternative auth and storage
- **Fly.io** - Hosting and persistent storage

## Development Workflow

### Local Development
1. Set up PostgreSQL database
2. Copy `.env.example` to `.env` and configure
3. Run `go mod download` to install dependencies
4. Run `make migrate` or start server (auto-migrates)
5. Access API at `http://localhost:8080`

### Code Generation
- **SQLC**: `sqlc generate` to regenerate database code
- **gRPC**: `protoc` to generate proto code (if using gRPC)

### Testing
- Unit tests: `go test ./...`
- Coverage: `go test -cover ./...`
- Integration tests in `.github/tests/`

## Monitoring and Observability

### Logging
- Structured logging with Chi middleware
- Request ID tracking
- WebSocket connection tracking
- Database query logging (via pgx)

### Metrics
- Health check endpoints
- Connection pool statistics
- WebSocket client counts
- Migration status verification

## Known Limitations & Future Improvements

### Current Limitations
- No WebSocket authentication on initial connection (handled via join_project message)
- Admin endpoints (`/migrations`) not production-secured
- Firebase integration optional but not fully utilized
- gRPC API exists but not actively used

### Planned Improvements
- Enhanced observability (Prometheus metrics)
- WebSocket authentication on connection handshake
- Admin role and RBAC for system-level operations
- Automated database backup and restore
- Multi-region deployment support

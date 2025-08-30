# DevHive Backend Refactoring

This document outlines the refactoring of the DevHive Go backend from a monolithic `main.go` to a clean, modular architecture.

## What Was Refactored

### Before (Monolithic Structure)
- **Single `cmd/main.go` file** (~1080 lines) with all routes, middleware, and business logic
- **Swagger comments** scattered throughout the code
- **Inline route handlers** mixed with business logic
- **Tight coupling** between different concerns

### After (Clean Architecture)
- **Modular structure** following Go best practices
- **OpenAPI specification** as the single source of truth
- **Generated Go types** and server stubs (OpenAPI codegen not currently configured)
- **Separation of concerns** with clear layers

## New Directory Structure

```
DevHive.GoBackend/
├─ api/                           # API contract & codegen
│  ├─ openapi.yaml               # OpenAPI specification
│  └─ (OpenAPI codegen not currently configured)
├─ cmd/
│  └─ server/
│     └─ main.go                 # Clean, minimal main (30 lines)
├─ internal/                      # Private application code
│  ├─ gen/                       # Generated OpenAPI code
│  ├─ router/                    # Route registration
│  ├─ middleware/                # HTTP middleware
│  ├─ ws/                        # WebSocket handling
│  ├─ auth/                      # Authentication
│  ├─ users/                     # User management
│  ├─ projects/                  # Project management
│  ├─ sprints/                   # Sprint management
│  ├─ tasks/                     # Task management
│  ├─ messages/                  # Messaging
│  └─ database/                  # Database operations
├─ pkg/                          # Public packages
│  ├─ config/                    # Configuration
│  ├─ db/                        # Database connection
│  └─ models/                    # Data models
├─ Makefile                      # Build & codegen commands
└─ fly.toml                      # Fly.io deployment
```

## Key Benefits

### 1. **Contract-First API Design**
- OpenAPI specification (not currently configured)
- All endpoints, schemas, and security defined declaratively
- Easy to generate client SDKs and documentation

### 2. **Generated Code**
- `make oapi` generates Go types and server stubs
- Ensures API contract compliance
- Reduces manual boilerplate code

### 3. **Clean Separation of Concerns**
- **Controllers**: Handle HTTP requests/responses
- **Services**: Business logic and orchestration
- **Repositories**: Data access and persistence
- **Middleware**: Cross-cutting concerns (CORS, auth, rate limiting)

### 4. **Maintainable Codebase**
- Each feature has its own package
- Clear dependencies between layers
- Easy to test individual components
- Simple to add new features

## Migration Steps

### 1. Generate OpenAPI Code
```bash
make oapi
```

### 2. Run the Server
```bash
make run
# or
PORT=8080 go run ./cmd/server
```

### 3. Build for Production
```bash
make build
```

## API Documentation

- **OpenAPI Spec**: `/swagger/openapi.yaml`
- **JSON Format**: `/swagger/doc.json`
- **Root Redirect**: `/` → OpenAPI specification

## WebSocket Support

- **Public WS**: `/ws` - General connections
- **Auth WS**: `/ws/auth` - Authenticated connections
- **Real-time updates** for projects, tasks, and messages

## Middleware Stack

1. **CORS**: Cross-origin resource sharing
2. **Rate Limiting**: Request throttling
3. **Authentication**: JWT token validation
4. **Logging & Recovery**: Request logging and panic recovery
5. **Compression**: Gzip compression

## Next Steps

### Immediate
- [ ] Implement remaining controllers (users, projects, sprints, tasks, messages)
- [ ] Add service layer with business logic
- [ ] Create repository layer for data access
- [ ] Add comprehensive error handling

### Future Enhancements
- [ ] Add OpenAPI validation middleware
- [ ] Implement API versioning
- [ ] Add metrics and monitoring
- [ ] Create automated tests
- [ ] Add database migrations

## Dependencies

The refactoring maintains all existing dependencies:
- **Gin**: HTTP framework
- **GORM**: Database ORM
- **JWT**: Authentication
- **WebSocket**: Real-time communication
- **Firebase**: External services

## Deployment

The refactored structure is fully compatible with:
- **Fly.io**: Primary deployment platform
- **Docker**: Containerization
- **GitHub Actions**: CI/CD pipeline

## Contributing

When adding new features:
1. **Update** OpenAPI specification (not currently configured)
2. **Generate** code with `make oapi`
3. **Implement** controllers, services, and repositories
4. **Register** routes in `internal/router/router.go`
5. **Test** locally with `make run`

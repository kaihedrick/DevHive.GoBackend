# DevHive API v1

This document describes the new versioned REST API for DevHive, built with idiomatic Go and PostgreSQL.

## Architecture

- **Router**: chi v5 with middleware
- **Database**: PostgreSQL with pgx v5 driver
- **Code Generation**: sqlc for type-safe database queries
- **Migrations**: goose with embedded SQL files
- **Authentication**: JWT with golang-jwt v5
- **Error Handling**: RFC 7807 problem details

## API Structure

### Base URL
- **Development**: `http://localhost:8080/api/v1`
- **Production**: `https://api.devhive.it.com/api/v1`

### Authentication
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

## Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh token (optional)
- `POST /api/v1/auth/password/reset-request` - Request password reset
- `POST /api/v1/auth/password/reset` - Reset password

### Users
- `POST /api/v1/users` - Create user (public)
- `GET /api/v1/users/me` - Get current user
- `GET /api/v1/users/{userId}` - Get user by ID

### Projects
- `GET /api/v1/projects` - List user's projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{projectId}` - Get project
- `PATCH /api/v1/projects/{projectId}` - Update project
- `DELETE /api/v1/projects/{projectId}` - Delete project

### Project Members
- `PUT /api/v1/projects/{projectId}/members/{userId}` - Add member
- `DELETE /api/v1/projects/{projectId}/members/{userId}` - Remove member

### Sprints
- `GET /api/v1/projects/{projectId}/sprints` - List project sprints
- `POST /api/v1/projects/{projectId}/sprints` - Create sprint
- `GET /api/v1/sprints/{sprintId}` - Get sprint
- `PATCH /api/v1/sprints/{sprintId}` - Update sprint
- `DELETE /api/v1/sprints/{sprintId}` - Delete sprint

### Tasks
- `GET /api/v1/projects/{projectId}/tasks` - List project tasks
- `POST /api/v1/projects/{projectId}/tasks` - Create task
- `GET /api/v1/tasks/{taskId}` - Get task
- `PATCH /api/v1/tasks/{taskId}` - Update task
- `PATCH /api/v1/tasks/{taskId}/status` - Update task status
- `DELETE /api/v1/tasks/{taskId}` - Delete task

### Messages
- `POST /api/v1/messages` - Create message
- `GET /api/v1/messages` - List messages with filters
- `GET /api/v1/messages/ws` - WebSocket endpoint

### Mail
- `POST /api/v1/mail/send` - Send email

## Legacy API Support

The API also provides legacy route shims for backward compatibility:

- `/api/User/ProcessLogin` → `POST /api/v1/auth/login`
- `/api/User/Register` → `POST /api/v1/users`
- `/api/User/{id}` → `GET /api/v1/users/{id}`
- `/api/Scrum/Project` → `POST /api/v1/projects`
- And more...

## Error Handling

All errors follow RFC 7807 problem details format:

```json
{
  "type": "https://api.devhive.it.com/problems/validation-error",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid input data",
  "instance": "/api/v1/users"
}
```

## Response Format

### Success Responses
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "johndoe",
  "email": "john@example.com",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

### Paginated Responses
```json
{
  "projects": [...],
  "limit": 20,
  "offset": 0
}
```

## Development

### Prerequisites
- Go 1.23+
- PostgreSQL 15+
- Docker & Docker Compose
- sqlc (for code generation)

### Quick Start
```bash
# Setup development environment
make dev-setup

# Run the application
make run
```

### Database Migrations
```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Code Generation
```bash
# Generate sqlc code
make gen
```

## Configuration

Environment variables:

```bash
PORT=8080
DATABASE_URL=postgres://devhive:devhive@localhost:5432/devhive?sslmode=disable
JWT_SIGNING_KEY=your-super-secret-jwt-key
JWT_ISSUER=https://api.devhive.it.com
JWT_AUDIENCE=devhive-clients
CORS_ORIGINS=https://d35scdhidypl44.cloudfront.net,https://devhive.it.com
CORS_ALLOW_CREDENTIALS=true
```

## Health Checks

- `GET /healthz` - Basic health check
- `GET /readyz` - Readiness check (includes database connectivity)


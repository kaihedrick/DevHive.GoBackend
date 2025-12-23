# DevHive Backend - Documentation Index

Welcome to the DevHive Backend documentation! This directory contains comprehensive documentation for engineers to understand and work with the codebase.

## Quick Start

New to the project? Start here:

1. **[Project Architecture](./System/project_architecture.md)** - Understand the overall system design, tech stack, and project structure
2. **[Database Schema](./System/database_schema.md)** - Learn about the data models and relationships
3. **[Authentication Flow](./System/authentication_flow.md)** - Understand how auth works (critical for API integration)

## Documentation Structure

```
.agent/
├── README.md           # This file - documentation index
├── System/             # Architecture and system design docs
│   ├── project_architecture.md
│   ├── database_schema.md
│   ├── authentication_flow.md
│   └── realtime_system.md
├── Tasks/              # Feature PRDs and implementation plans
└── SOP/                # Standard operating procedures
    ├── adding_migrations.md
    └── adding_api_endpoints.md
```

---

## System Documentation

These documents describe the current state of the system.

### 📐 [Project Architecture](./System/project_architecture.md)

**Read this if:** You're new to the project or need to understand the big picture

**Contents:**
- Project overview and core purpose
- Tech stack (Go 1.25, Chi, PostgreSQL, SQLC, WebSocket)
- Project structure and folder organization
- Architecture layers (HTTP, Data, Real-time, gRPC)
- Core features (auth, user management, projects, sprints, tasks, messaging)
- Configuration and environment variables
- Deployment architecture (Fly.io)
- Security features and performance optimizations
- Integration points (frontend, external services)

**Key sections:**
- Tech Stack - What libraries and frameworks we use
- Project Structure - Where to find code
- Core Features - What the system does
- Configuration - Environment variables and settings

---

### 🗄️ [Database Schema](./System/database_schema.md)

**Read this if:** You need to understand the data model or write database queries

**Contents:**
- Complete schema documentation for all tables
- Entity relationships diagram
- Database functions and triggers
- Migration history and versioning
- Common query patterns
- Performance considerations
- Security considerations (cascade deletes, constraints)

**Tables documented:**
- `users` - User accounts and profiles
- `projects` - Project entities
- `project_members` - Project membership with roles
- `project_invites` - Time-limited invite links
- `sprints` - Sprint planning
- `tasks` - Task tracking
- `messages` - Threaded messaging
- `refresh_tokens` - Persistent auth tokens
- `password_resets` - Password reset tokens

**Key sections:**
- Core Tables - Detailed field descriptions
- Entity Relationships - How tables connect
- Common Query Patterns - Example SQL queries
- Database Functions - Triggers and stored procedures

---

### 🔐 [Authentication Flow](./System/authentication_flow.md)

**Read this if:** You're integrating with the API or working on auth features

**Contents:**
- Dual-token JWT system (access + refresh tokens)
- Complete authentication flow (registration, login, refresh, logout)
- JWT token structure and validation
- Project-level authorization and RBAC
- Password reset flow
- Admin password verification
- Security best practices
- Frontend integration guide (with code examples)

**Flows documented:**
- User Registration - Email/username validation, password hashing
- User Login - JWT generation, refresh token creation
- Token Refresh - Automatic token renewal on 401
- Password Reset - Email-based reset flow
- Change Password - Authenticated password change

**Key sections:**
- Authentication Flow - Step-by-step auth process
- JWT Token Structure - What's inside the tokens
- Authorization & Access Control - Role-based permissions
- Frontend Integration Guide - How to use the auth API

---

### ⚡ [Real-time System](./System/realtime_system.md)

**Read this if:** You're working on WebSocket features or cache invalidation

**Contents:**
- PostgreSQL NOTIFY/LISTEN architecture
- WebSocket hub and client management
- Cache invalidation triggers
- Message flow examples
- WebSocket protocol specification
- Frontend integration (React/WebSocket)
- Performance characteristics
- Monitoring and debugging

**Components documented:**
1. Database NOTIFY Triggers - Auto-notify on data changes
2. NOTIFY Listener - Dedicated listener connection
3. WebSocket Hub - Central message broker
4. WebSocket Clients - Client lifecycle and message handling

**Key sections:**
- Architecture Overview - How real-time notifications work
- Message Flow Example - Step-by-step notification delivery
- WebSocket Protocol - Client-server communication spec
- Frontend Integration - How to connect and handle messages

---

## SOP Documentation

Standard operating procedures for common development tasks.

### 🔧 [Adding Database Migrations](./SOP/adding_migrations.md)

**Use this when:** You need to change the database schema

**Contents:**
- Sequential migration system
- Step-by-step migration process
- Common migration patterns (add column, add index, add trigger, etc.)
- Migration checklist
- Rollback strategy
- Zero-downtime schema changes
- Troubleshooting

**Common patterns covered:**
- Adding/removing columns
- Adding indexes (simple, composite, unique, partial)
- Adding foreign keys
- Adding triggers
- Data backfill

**Key sections:**
- Step-by-Step Process - How to create a migration
- Common Migration Patterns - Copy-paste examples
- Migration Checklist - What to check before deploying
- Rollback Strategy - How to undo changes

---

### 🌐 [Adding API Endpoints](./SOP/adding_api_endpoints.md)

**Use this when:** You need to add new REST API endpoints

**Contents:**
- File structure for handlers and routes
- Step-by-step process for creating endpoints
- SQLC query definition
- Handler implementation patterns (CRUD operations)
- Route registration
- Testing strategies
- Response helper functions
- Common patterns (authorization, pagination, validation, transactions)

**Patterns covered:**
- List Resource (GET /resources)
- Get Resource by ID (GET /resources/{id})
- Create Resource (POST /resources)
- Update Resource (PATCH /resources/{id})
- Delete Resource (DELETE /resources/{id})
- Nested Resources (GET /projects/{id}/tasks)

**Key sections:**
- Step-by-Step Process - How to add an endpoint
- Implement Handler Methods - Code examples for CRUD
- Common Patterns - Authorization, pagination, validation
- API Endpoint Checklist - What to check before deploying

---

## Tasks Documentation

This folder contains PRD (Product Requirement Documents) and implementation plans for specific features.

**When to create a Task doc:**
- Planning a new feature
- Documenting an implementation approach
- Recording decisions made during development

**Naming convention:** `{feature_name}_prd.md` or `{feature_name}_implementation.md`

**Current tasks:**

### [Google OAuth 2.0 Authentication](./Tasks/google_oauth.md)
**Status:** Planning - Ready for Implementation
**Description:** Comprehensive implementation plan for adding Google OAuth 2.0 authentication alongside existing username/password auth, with persistent login controlled by "Remember Me" checkbox

**Key Features:**
- Google OAuth 2.0 sign-in integration
- Persistent login (30 days) vs session-based login
- Frontend "Remember Me" checkbox control
- Unified user management for both auth methods
- CSRF protection with state parameter
- Secure token storage and refresh handling

---

## How to Use This Documentation

### For New Engineers

1. **Start with Project Architecture** - Understand what DevHive is and how it's built
2. **Read Database Schema** - Learn the data model
3. **Read Authentication Flow** - Critical for API integration
4. **Skim the SOPs** - Know where to look when you need to do specific tasks

### For Experienced Engineers

- **Use SOPs as reference** when doing specific tasks (adding migrations, endpoints)
- **Read System docs** when working on specific subsystems (auth, real-time, etc.)
- **Update docs** when making architectural changes or adding new patterns

### For AI Assistants

Before implementing features:
1. **Read README.md first** to understand what documentation exists
2. **Read relevant System docs** to understand current architecture
3. **Follow SOPs** for standard tasks (migrations, endpoints)
4. **Update docs** after implementing features to reflect changes

---

## Documentation Maintenance

### When to Update Documentation

- **System docs:** After architectural changes, new features, or significant refactoring
- **SOP docs:** When adding new common patterns or changing existing procedures
- **Task docs:** Create for new features, update after implementation

### How to Update

1. **Find the relevant doc** in .agent/System/ or .agent/SOP/
2. **Make changes** directly in the markdown file
3. **Update "Related Documentation" links** if adding new docs
4. **Update this README.md** if adding new documentation files
5. **Commit changes** with clear commit message

### Documentation Standards

- **Use markdown** for all documentation
- **Include code examples** where relevant
- **Link between related docs** using relative paths
- **Keep docs up-to-date** with code changes
- **Use clear headings** for easy scanning
- **Include "Related Documentation" section** at the top of each doc

---

## External Documentation

For reference beyond this codebase:

- **Main README:** `../README.md` - Project overview, quick start, API reference
- **Frontend Examples:** `../frontend-examples/` - Integration examples for React
- **Migration Files:** `../cmd/devhive-api/migrations/` - Actual migration SQL
- **Root-level Docs:** `../docs/` - Additional documentation and guides

### Key Root-level Docs

- `CACHE_INVALIDATION_FRONTEND.md` - Frontend cache invalidation guide
- `PROJECT_INVITES_COMPLETE_GUIDE.md` - Invite system deep dive
- `EMAIL_VALIDATION_GUIDE.md` - Email validation implementation
- `TOKEN_REFRESH_FIX.md` - Token refresh troubleshooting
- `WEBSOCKET_TOKEN_REFRESH_GUIDE.md` - WebSocket auth and token refresh

---

## Quick Reference

### Common File Locations

| What | Where |
|------|-------|
| Main entry point | `cmd/devhive-api/main.go` |
| Route definitions | `internal/http/router/router.go` |
| Handlers | `internal/http/handlers/{resource}.go` |
| Database queries (SQLC) | `internal/repo/queries.sql.go` |
| Database models | `internal/repo/models.go` |
| Migrations | `cmd/devhive-api/migrations/` |
| WebSocket hub | `internal/ws/hub.go` |
| NOTIFY listener | `internal/db/notify_listener.go` |
| Config | `internal/config/config.go` |

### Common Commands

```bash
# Run server locally
go run cmd/devhive-api/main.go

# Generate SQLC code
sqlc generate

# Run tests
go test ./...

# Run migrations manually
psql -d devhive -f cmd/devhive-api/migrations/{number}_{name}.sql

# Deploy to Fly.io
fly deploy

# Check deployment logs
fly logs
```

### Common API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/auth/login` | POST | User login |
| `/api/v1/auth/refresh` | POST | Refresh access token |
| `/api/v1/users` | POST | Register user |
| `/api/v1/projects` | GET | List projects |
| `/api/v1/projects/{id}` | GET | Get project details |
| `/api/v1/projects/{id}/sprints` | GET | List sprints |
| `/api/v1/projects/{id}/tasks` | GET | List tasks |
| `/api/v1/messages/ws` | GET | WebSocket connection |

---

## Contributing to Documentation

Found something unclear? Documentation out of date? Please update it!

1. **Identify the issue** - What's missing or wrong?
2. **Find the relevant doc** - Which file needs updating?
3. **Make the change** - Edit the markdown directly
4. **Test your change** - Ensure links work, formatting is correct
5. **Commit** - Clear commit message describing the change

Example commit message:
```
docs: Update authentication flow to include new MFA feature

- Add MFA setup and verification flow
- Update JWT claims structure with mfa_enabled field
- Add frontend integration example for MFA
```

---

## Feedback

If you have suggestions for improving the documentation structure or content:

- Create an issue in the repository
- Discuss with the team in Slack/Discord
- Submit a PR with proposed changes

---

**Last Updated:** 2025-12-22

**Documentation Version:** 1.0

**Maintained by:** DevHive Team

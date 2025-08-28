# Swagger API Documentation Setup

Your Go backend is now set up with Swagger/OpenAPI documentation! Here's how to use it:

## What's Been Added

1. **Swagger Dependencies**: Added `swaggo/swag`, `swaggo/gin-swagger`, and `swaggo/files`
2. **Swagger CLI Tool**: Installed `swag` command-line tool for generating documentation
3. **API Annotations**: Added Swagger annotations to your models and controllers
4. **Swagger Route**: Added `/swagger/*any` route to serve the Swagger UI

## API Architecture

**Unified REST API**: Your backend now uses a single, unified REST API for both web and mobile clients. This eliminates code duplication and ensures consistency across platforms.

### Available Endpoints

- **Authentication**: `/api/v1/auth/*` - Login, register, token refresh
- **Users**: `/api/v1/users/*` - Profile management, avatar uploads
- **Projects**: `/api/v1/projects/*` - Project CRUD operations, member management
- **Sprints**: `/api/v1/projects/:id/sprints/*` - Sprint management within projects
- **Tasks**: `/api/v1/projects/:id/sprints/:sprintId/tasks/*` - Task management within sprints
- **Messages**: `/api/v1/projects/:id/messages/*` - Project messaging system
- **Feature Flags**: `/api/v1/admin/feature-flags/*` - Admin feature flag management

## How to Use

### 1. View API Documentation

Once your server is running, you can access the Swagger UI at:
```
http://localhost:8080/swagger/
```

This will show you an interactive API documentation interface where you can:
- Browse all available endpoints
- See request/response schemas
- Test API calls directly from the browser
- View authentication requirements
- Understand error responses

### 2. Generate Documentation

To regenerate the Swagger documentation after making changes:
```bash
swag init -g cmd/main.go
```

### 3. Test Your Endpoints

Use the Swagger UI to:
- **Authenticate**: Use the "Authorize" button with your JWT token
- **Test Endpoints**: Click "Try it out" on any endpoint
- **View Responses**: See actual API responses and error codes
- **Understand Schemas**: View detailed request/response models

## Benefits of Unified API

✅ **Single Source of Truth**: One set of endpoints for all clients
✅ **Easier Maintenance**: No duplicate code or logic
✅ **Consistent Behavior**: Same validation, authentication, and business logic
✅ **Better Testing**: Test once, works everywhere
✅ **Simplified Development**: Frontend teams can use the same API documentation

## Frontend Integration

Both your React web app and mobile app can now:
- Use the same API endpoints
- Follow the same authentication flow
- Handle responses consistently
- Share the same data models

The frontend can handle any platform-specific data transformation or UI requirements while using the same underlying API.

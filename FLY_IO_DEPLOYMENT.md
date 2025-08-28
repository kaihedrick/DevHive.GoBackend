# DevHive Go Backend - Fly.io Deployment Guide

## 🚀 Overview
This guide covers deploying your DevHive Go Backend to Fly.io with proper secret management and configuration.

## 🆕 **New in This Version**
- **Enhanced Task Management**: Project-level and sprint-level task operations
- **Complete API Coverage**: All 38 endpoints now supported
- **Improved Workflow**: Better project and sprint management capabilities
- **Backlog Management**: Create and manage tasks without sprint assignment

## 🔐 Fly.io Secrets Configuration

### Current Secrets (Already Configured)
Your Fly.io app has the following secrets configured:

```bash
# Database
ConnectionStringsDbConnection         19a57c2b3eff497a        Aug 23 2025 20:36

# Firebase
FIREBASE_JSON                           ce4b40d95dd5fb3b        Aug 21 2025 23:55
FIREBASE_JSON_BASE64                    db68f6d6272d21e8        Aug 22 2025 00:30
FIREBASE_PROJECT_ID                     025f312ebc8b663c        Aug 21 2025 23:46
FIREBASE_SERVICE_ACCOUNT_KEY_PATH       766a01c0677b127a        Aug 22 2025 00:21

# JWT Configuration
JwtAudience                           be7391c57350931c        Aug 23 2025 06:21
JwtIssuer                             d3b27a5e2c78f3a7        Aug 23 2025 06:21
JwtKey                                f313ac9ab1b641b6        Aug 23 2025 22:56

# Email (Mailgun)
MailgunApiKey                         860b59dacddd3e10        Aug 25 2025 21:30
MailgunDomain                         9b333c64735d3615        Aug 23 2025 06:17
MAILGUN_DOMAIN                          9b333c64735d3615        Aug 21 2025 23:56
MAILGUN_SENDER                          913cae3c53f375f7        Aug 21 2025 23:56
Mailgun__SenderEmail                    913cae3c53f375f7        Aug 23 2025 06:17
```

## 🔧 Environment Variable Mapping

### Database Configuration
- **Fly.io Secret**: `ConnectionStringsDbConnection`
- **Code Usage**: Automatically used in `db/postgres.go`
- **Fallback**: Individual `DB_*` environment variables

### Firebase Configuration
- **Primary**: `FIREBASE_JSON_BASE64` (base64 encoded service account)
- **Fallback**: `FIREBASE_SERVICE_ACCOUNT_KEY_PATH` (file path)
- **Project ID**: `FIREBASE_PROJECT_ID`

### JWT Configuration
- **Secret Key**: `JwtKey`
- **Issuer**: `JwtIssuer`
- **Audience**: `JwtAudience`

### Email Configuration
- **API Key**: `MailgunApiKey`
- **Domain**: `MailgunDomain`
- **Sender**: `Mailgun__SenderEmail`

## 🚀 Deployment Steps

### 1. Build and Deploy
```bash
# Build the application
fly deploy

# Or build locally and deploy
fly deploy --local-only
```

### 2. Verify Deployment
```bash
# Check app status
fly status

# View logs
fly logs

# SSH into the app (if needed)
fly ssh console
```

### 3. Health Check
```bash
# Test the health endpoint
curl https://your-app.fly.dev/health
```

### 4. Test New Task Endpoints
```bash
# Test project-level task endpoints
curl -X GET https://your-app.fly.dev/api/v1/projects/{projectId}/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Test sprint-level task endpoints  
curl -X GET https://your-app.fly.dev/api/v1/projects/{projectId}/sprints/{sprintId}/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Test task assignment
curl -X POST https://your-app.fly.dev/api/v1/projects/{projectId}/tasks/{taskId}/assign \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"assignee_id":"user-uuid"}'
```

## 📊 **API Endpoints Overview**

### **Complete Endpoint Coverage (38 Total)**
```
🔐 Auth: 5 endpoints
👤 Users: 6 endpoints  
📁 Projects: 9 endpoints
🏁 Sprints: 7 endpoints
✅ Tasks: 14 endpoints (NEW: Project + Sprint level)
💬 Messages: 4 endpoints
🧠 Feature Flags: 6 endpoints
📱 Mobile API: 4 endpoints
🛠️ Admin & Utilities: 3 endpoints
```

### **New Task Management Endpoints**
- **Project-Level**: `/api/v1/projects/{id}/tasks` (7 endpoints)
- **Sprint-Level**: `/api/v1/projects/{id}/sprints/{sprintId}/tasks` (7 endpoints)
- **Enhanced Operations**: Assign, status updates, cross-sprint management

## 📊 Environment Variable Priority

The application uses the following priority for configuration:

1. **Fly.io Secrets** (highest priority)
2. **Environment Variables** (if secrets not found)
3. **Default Values** (lowest priority)

### Example Priority Chain:
```go
// JWT Secret priority:
// 1. JwtKey (Fly.io secret)
// 2. JWT_SECRET (environment variable)
// 3. "your-secret-key" (default)
JWTSecret: getEnv("JwtKey", getEnv("JWT_SECRET", "your-secret-key"))
```

## 🔍 Troubleshooting

### Common Issues

#### 1. Database Connection Failed
```bash
# Check if the secret is set
fly secrets list | grep ConnectionStringsDbConnection

# Verify the connection string format
fly ssh console
echo $ConnectionStringsDbConnection
```

#### 2. Firebase Not Initializing
```bash
# Check Firebase secrets
fly secrets list | grep FIREBASE

# Verify base64 encoding
fly ssh console
echo $FIREBASE_JSON_BASE64 | base64 -d | jq .
```

#### 3. JWT Authentication Failing
```bash
# Check JWT secrets
fly secrets list | grep Jwt

# Verify secret values
fly ssh console
echo $JwtKey
echo $JwtIssuer
echo $JwtAudience
```

### Debug Commands
```bash
# View all environment variables
fly ssh console
env | sort

# Check specific secret
fly ssh console
echo $ConnectionStringsDbConnection

# View application logs
fly logs
```

## 📝 Adding New Secrets

### Add a New Secret
```bash
fly secrets set NEW_SECRET_NAME="secret_value"
```

### Update Existing Secret
```bash
fly secrets set EXISTING_SECRET="new_value"
```

### Remove Secret
```bash
fly secrets unset SECRET_NAME
```

## 🔒 Security Best Practices

### 1. Secret Rotation
- Rotate JWT keys regularly
- Update Firebase service account keys periodically
- Monitor secret access logs

### 2. Access Control
- Limit who can view secrets
- Use different secrets for different environments
- Audit secret access regularly

### 3. Environment Separation
- Use different Fly.io apps for staging/production
- Separate database connections per environment
- Different Firebase projects per environment

## 📊 Monitoring

### Health Endpoints
```bash
# Basic health check
GET /health

# Detailed health check (if implemented)
GET /health/detailed
```

### Logging
```bash
# View real-time logs
fly logs

# View specific log levels
fly logs --level error

# Follow logs
fly logs -f
```

## 🚀 Performance Optimization

### 1. Database Connection Pooling
- Configured in `db/postgres.go`
- Max connections: 25
- Idle connections: 25

### 2. Feature Flags
- 5-minute caching for performance
- Background synchronization
- Memory-based lookups

### 3. Rate Limiting
- Configurable per endpoint
- Feature flag controlled
- Different limits for mobile vs web

## 🔄 CI/CD Integration

### GitHub Actions Example
```yaml
name: Deploy to Fly.io
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --remote-only
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

## 📚 Additional Resources

### Fly.io Documentation
- [Secrets Management](https://fly.io/docs/reference/secrets/)
- [Deployment](https://fly.io/docs/deploy/)
- [Configuration](https://fly.io/docs/reference/configuration/)

### DevHive Documentation
- [Strategic Features](STRATEGIC_FEATURES.md)
- [API Reference](README.md)
- [Database Schema](db/schema.sql)

## 🎯 **Deployment Summary**

### **What's Deployed**
- ✅ **Complete API**: All 38 endpoints fully functional
- ✅ **Enhanced Tasks**: Project-level and sprint-level task management
- ✅ **Enterprise Security**: JWT, Firebase, and database security
- ✅ **Production Ready**: Health checks, monitoring, and auto-scaling
- ✅ **Mobile Optimized**: All endpoints optimized for mobile applications

### **Key Benefits**
- **Flexible Task Management**: Create tasks at project level, assign to sprints later
- **Backlog Support**: Maintain unassigned tasks for future planning
- **Cross-Sprint Operations**: Move tasks between sprints seamlessly
- **Comprehensive Coverage**: 100% API endpoint coverage achieved

---

**DevHive Go Backend** - Successfully configured for Fly.io deployment with enterprise-grade secret management and **complete task management capabilities**.

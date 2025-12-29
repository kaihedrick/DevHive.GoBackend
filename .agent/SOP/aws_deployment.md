# SOP: AWS Deployment

## Related Documentation
- [Project Architecture](../System/project_architecture.md) - Overall system architecture
- [Real-time System](../System/realtime_system.md) - WebSocket architecture details
- [Adding Migrations](./adding_migrations.md) - Database schema changes

## Overview

This guide covers deploying the DevHive backend to AWS using SAM (Serverless Application Model). The deployment creates:

- **HTTP API Gateway** - REST API endpoints
- **WebSocket API Gateway** - Real-time bidirectional communication
- **3 Lambda Functions** - HTTP handler, WebSocket handler, Broadcaster
- **DynamoDB Table** - WebSocket connection state

---

## Prerequisites

### 1. AWS CLI and SAM CLI

```bash
# Install AWS CLI
# Windows (via winget)
winget install Amazon.AWSCLI

# Install SAM CLI
# Windows (via MSI installer from AWS)
# Download from: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html
```

### 2. AWS SSO Configuration

```bash
# Configure SSO profile
aws configure sso --profile devhive

# Verify access
aws sts get-caller-identity --profile devhive
```

### 3. Neon Database

1. Create account at [console.neon.tech](https://console.neon.tech)
2. Create a new project in `us-west-2` region
3. Copy the connection string (pooler URL with `?sslmode=require`)

---

## Deployment Configuration

### template.yaml

The SAM template (`template.yaml`) defines all AWS resources:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Parameters:
  DatabaseURL:           # Neon PostgreSQL connection string
  JWTSigningKey:         # Secret for JWT signing
  GoogleClientID:        # Google OAuth client ID
  GoogleClientSecret:    # Google OAuth client secret
  CORSOrigins:           # Comma-separated allowed origins
  AdminPassword:         # Admin certificates password
  ResendAPIKey:          # Resend email API key
  ResendFromEmail:       # Sender email address
  FrontendURL:           # Frontend URL for redirects
  GoogleRedirectURL:     # OAuth callback URL
```

### samconfig.toml

Stores default deployment settings:

```toml
[default.deploy.parameters]
stack_name = "devhive-api"
region = "us-west-2"
capabilities = "CAPABILITY_IAM"
confirm_changeset = false
resolve_s3 = true
profile = "devhive"
```

---

## Step-by-Step Deployment

### 1. Build Lambda Functions

```bash
# Build all Lambda functions
sam build --profile devhive
```

This compiles Go code for:
- `cmd/lambda/` → `devhive-api` function
- `cmd/websocket/` → `devhive-websocket` function
- `cmd/broadcaster/` → `devhive-broadcaster` function

### 2. Deploy to AWS

```bash
sam deploy --profile devhive --parameter-overrides \
  "DatabaseURL=postgresql://user:pass@host/db?sslmode=require" \
  "JWTSigningKey=your-secret-key" \
  "GoogleClientID=your-client-id.apps.googleusercontent.com" \
  "GoogleClientSecret=GOCSPX-xxx" \
  "AdminPassword=your-admin-password" \
  "ResendAPIKey=re_xxx" \
  "ResendFromEmail=noreply@devhive.it.com" \
  "CORSOrigins=http://localhost:3000,https://devhive.it.com" \
  "FrontendURL=https://devhive.it.com" \
  "GoogleRedirectURL=https://xxx.execute-api.us-west-2.amazonaws.com/api/v1/auth/google/callback"
```

### 3. Get Deployment Outputs

After deployment, SAM outputs the endpoints:

```
Outputs:
  HttpApiEndpoint:    https://xxx.execute-api.us-west-2.amazonaws.com
  WebSocketEndpoint:  wss://xxx.execute-api.us-west-2.amazonaws.com/prod
```

### 4. Update Google OAuth Redirect URL

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Navigate to APIs & Services → Credentials
3. Edit your OAuth 2.0 Client ID
4. Add the HTTP API endpoint + `/api/v1/auth/google/callback` to Authorized Redirect URIs
5. Redeploy with updated `GoogleRedirectURL` parameter

---

## Database Migrations

### Option 1: Run via Neon Console

1. Go to [console.neon.tech](https://console.neon.tech)
2. Select your project
3. Open SQL Editor
4. Paste migration SQL from `cmd/devhive-api/migrations/` files in order

### Option 2: Run via psql

```bash
# Install psql if needed
winget install PostgreSQL.psql

# Run migrations
psql "postgresql://user:pass@host/db?sslmode=require" \
  -f cmd/devhive-api/migrations/001_initial_schema.sql
psql "postgresql://user:pass@host/db?sslmode=require" \
  -f cmd/devhive-api/migrations/002_remove_title_from_tasks.sql
# ... continue with all migration files
```

### Option 3: Auto-migrate on Lambda start

The Lambda function automatically runs migrations on cold start (see `cmd/lambda/main.go`):

```go
if err := db.RunMigrations(database); err != nil {
    log.Printf("Warning: Migration failed: %v", err)
}
```

---

## Monitoring & Debugging

### View Lambda Logs

```bash
# Tail logs for HTTP API Lambda
aws logs tail /aws/lambda/devhive-api --follow --profile devhive

# Tail logs for WebSocket Lambda
aws logs tail /aws/lambda/devhive-websocket --follow --profile devhive

# Tail logs for Broadcaster Lambda
aws logs tail /aws/lambda/devhive-broadcaster --follow --profile devhive
```

### Test HTTP API

```bash
# Health check
curl https://xxx.execute-api.us-west-2.amazonaws.com/health

# Login
curl -X POST https://xxx.execute-api.us-west-2.amazonaws.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test"}'
```

### Test WebSocket

```bash
# Using wscat
npm install -g wscat
wscat -c "wss://xxx.execute-api.us-west-2.amazonaws.com/prod?token=YOUR_JWT"
```

---

## Common Issues & Solutions

### Issue: 404 on all routes

**Cause:** Wrong API Gateway stage name

**Solution:** Ensure `StageName: $default` in template.yaml for HTTP API (removes `/prod` prefix)

### Issue: Lambda timeout on cold start

**Cause:** Database connection taking too long

**Solution:**
- Use Neon pooler URL (not direct connection)
- Increase Lambda timeout in template.yaml

### Issue: OPTIONS returns 405 or 500 (CORS preflight failing)

**Cause:** Lambda is receiving OPTIONS requests instead of API Gateway handling them

**Root Issue:** Using `Method: ANY` in Lambda event configuration causes Lambda to handle OPTIONS requests, which results in 405/500 errors and breaks CORS preflight.

**Solution:**

**In `template.yaml`**, replace `Method: ANY` with explicit method events:

```yaml
# ❌ WRONG - Lambda handles OPTIONS
Events:
  ApiProxy:
    Type: HttpApi
    Properties:
      ApiId: !Ref DevHiveHttpApi
      Path: /{proxy+}
      Method: ANY  # This includes OPTIONS!

# ✅ CORRECT - Only send actual HTTP methods to Lambda
Events:
  ApiProxyGet:
    Type: HttpApi
    Properties:
      ApiId: !Ref DevHiveHttpApi
      Path: /{proxy+}
      Method: GET
  ApiProxyPost:
    Type: HttpApi
    Properties:
      ApiId: !Ref DevHiveHttpApi
      Path: /{proxy+}
      Method: POST
  # ... repeat for PUT, DELETE, PATCH
```

**Router Conditional CORS** (`internal/http/router/router.go`):

```go
// Enable CORS if: (1) not in Lambda at all, OR (2) running in SAM local
enableAppCors := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" || os.Getenv("AWS_SAM_LOCAL") == "true"
if enableAppCors {
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   cfg.CORS.AllowedOrigins,
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "Origin", "Cookie"},
        ExposedHeaders:   []string{"Set-Cookie", "X-Request-Id", "Link"},
        AllowCredentials: cfg.CORS.AllowCredentials,
        MaxAge:           300,
    }))
}
```

**Verification:**
```bash
# Test CORS preflight - should return 204 with NO rate-limit headers
# (No rate-limit headers = API Gateway handled it, not Lambda)
curl -i -X OPTIONS https://go.devhive.it.com/api/v1/auth/login \
  -H "Origin: https://devhive.it.com" \
  -H "Access-Control-Request-Method: POST"

# Expected response:
# HTTP/1.1 204 No Content
# access-control-allow-origin: https://devhive.it.com
# access-control-allow-methods: DELETE,GET,OPTIONS,PATCH,POST,PUT
# access-control-allow-credentials: true
# (NO X-Ratelimit-* headers)

# Test actual endpoint - should work normally with rate-limit headers
curl -i -X POST https://go.devhive.it.com/api/v1/auth/login \
  -H "Origin: https://devhive.it.com" \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Should have X-Ratelimit-* headers (Lambda processed it)
```

### Issue: CORS errors (duplicate headers)

**Cause:** Both Lambda and API Gateway adding CORS headers

**Solution:**
1. Check `CORSOrigins` parameter includes your frontend domain
2. Verify `AllowCredentials: true` in HTTP API CORS config
3. Ensure frontend uses `withCredentials: true`
4. **IMPORTANT**: Ensure Lambda event config uses explicit methods (not `ANY`)
   - See "OPTIONS returns 405 or 500" above
   - API Gateway handles CORS automatically via `CorsConfiguration` in template.yaml
   - Router conditionally applies CORS middleware (only local dev + SAM local)
   - This prevents duplicate/conflicting CORS headers

### Issue: WebSocket 401 on connect

**Cause:** Invalid or expired JWT token

**Solution:**
1. Verify `JWT_SIGNING_KEY` matches between Lambda and token issuer
2. Check token expiration
3. Ensure token is passed in `?token=` query parameter

### Issue: WebSocket subscribe not working

**Cause:** Incorrect field name in subscribe message

**Solution:**
The WebSocket Lambda accepts both `projectId` (camelCase) and `project_id` (snake_case) for backward compatibility:

```json
// Both formats work:
{"action": "subscribe", "projectId": "uuid-here"}
{"action": "subscribe", "project_id": "uuid-here"}
```

**Recommended**: Use `projectId` (camelCase) for consistency with the rest of the API.

**Implementation**: The `WebSocketMessage` struct in `cmd/websocket/main.go` has both fields:
```go
type WebSocketMessage struct {
    Action       string `json:"action"`
    ProjectID    string `json:"projectId,omitempty"`
    ProjectIdAlt string `json:"project_id,omitempty"` // Backward compatibility
}
```

### Issue: Lambda receives paths with stage prefix (e.g., `/prod/api/v1/auth/login`)

**Cause:** Using `StageName: prod` adds stage prefix to paths

**Symptoms:**
- All routes return 404
- Lambda logs show paths like `POST /prod/api/v1/auth/login`
- Routes work with direct API Gateway URL but not custom domain

**Solution:**

Use `StageName: $default` in `template.yaml`:

```yaml
DevHiveHttpApi:
  Type: AWS::Serverless::HttpApi
  Properties:
    StageName: $default  # No stage prefix in paths
```

**Why this matters:**
- `StageName: prod` → Lambda receives `/prod/api/v1/...`
- `StageName: $default` → Lambda receives `/api/v1/...`
- Your router expects `/api/v1/...` without stage prefix

**Custom Domain Mapping:**

Ensure custom domain is mapped to `$default` stage:

```bash
# Update API mapping
aws apigatewayv2 update-api-mapping \
  --api-id <api-id> \
  --api-mapping-id <mapping-id> \
  --domain-name go.devhive.it.com \
  --stage '$default' \
  --profile devhive

# Verify
aws apigatewayv2 get-api-mappings \
  --domain-name go.devhive.it.com \
  --profile devhive
```

**CloudFormation Stage Conflicts:**

If CloudFormation tries to create a stage that conflicts with custom domain mapping:

```bash
# First update custom domain to new stage
aws apigatewayv2 update-api-mapping --stage '$default' ...

# Then delete old stage
aws apigatewayv2 delete-stage --api-id <api-id> --stage-name prod --profile devhive
```

### Issue: Circular dependency in CloudFormation

**Cause:** Lambda environment variables referencing the HTTP API URL

**Solution:** Use a parameter for `GoogleRedirectURL` instead of `!Sub` reference:
```yaml
GoogleRedirectURL:
  Type: String
  Default: 'http://localhost:8080/api/v1/auth/google/callback'
```

---

## Updating the Deployment

### Update code only (no infra changes)

```bash
sam build --profile devhive && sam deploy --profile devhive
```

### Update parameters

```bash
sam deploy --profile devhive --parameter-overrides \
  "GoogleRedirectURL=https://new-url.execute-api.us-west-2.amazonaws.com/api/v1/auth/google/callback"
```

### Full redeploy

```bash
sam build --profile devhive
sam deploy --profile devhive --parameter-overrides "..." --force-upload
```

---

## Cleanup

### Delete the stack

```bash
aws cloudformation delete-stack --stack-name devhive-api --profile devhive
```

**Warning:** This deletes all resources including the DynamoDB table (WebSocket connections).

---

## Custom Domain Setup

Custom domains provide cleaner URLs and allow consistent endpoints even if the API Gateway ID changes.

### HTTP API Custom Domain (go.devhive.it.com)

1. **Request ACM Certificate:**
   ```bash
   aws acm request-certificate --domain-name go.devhive.it.com \
     --validation-method DNS --profile devhive
   ```

2. **Add DNS Validation Record** (provided by ACM)

3. **Create Custom Domain:**
   ```bash
   aws apigatewayv2 create-domain-name --domain-name go.devhive.it.com \
     --domain-name-configurations CertificateArn=arn:aws:acm:... \
     --profile devhive
   ```

4. **Create API Mapping:**
   ```bash
   aws apigatewayv2 create-api-mapping --domain-name go.devhive.it.com \
     --api-id <HTTP_API_ID> --stage '$default' --profile devhive
   ```

5. **Add DNS CNAME Record:**
   ```
   go.devhive.it.com → d-xxx.execute-api.us-west-2.amazonaws.com
   ```

### WebSocket API Custom Domain (ws.devhive.it.com)

1. **Request ACM Certificate:**
   ```bash
   aws acm request-certificate --domain-name ws.devhive.it.com \
     --validation-method DNS --profile devhive
   ```

2. **Add DNS Validation Record** (provided by ACM)

3. **Create Custom Domain:**
   ```bash
   aws apigatewayv2 create-domain-name --domain-name ws.devhive.it.com \
     --domain-name-configurations CertificateArn=arn:aws:acm:us-west-2:783764580656:certificate/44da95fb-b74e-4254-b558-9065efc83041 \
     --profile devhive
   ```

4. **Create API Mapping:**
   ```bash
   aws apigatewayv2 create-api-mapping --domain-name ws.devhive.it.com \
     --api-id er7oc4a3o5 --stage prod --profile devhive
   ```

5. **Add DNS CNAME Record:**
   ```
   ws.devhive.it.com → d-pgayuxdrz2.execute-api.us-west-2.amazonaws.com
   ```

### Verify Custom Domains

```bash
# List custom domains
aws apigatewayv2 get-domain-names --profile devhive

# Get specific domain details
aws apigatewayv2 get-domain-name --domain-name ws.devhive.it.com --profile devhive
```

---

## Production URLs

### Custom Domains (Recommended)

| Resource | URL |
|----------|-----|
| HTTP API | `https://go.devhive.it.com` |
| WebSocket API | `wss://ws.devhive.it.com` |
| Google OAuth Callback | `https://go.devhive.it.com/api/v1/auth/google/callback` |

### Direct API Gateway URLs

| Resource | URL |
|----------|-----|
| HTTP API | `https://7x1vij0u6k.execute-api.us-west-2.amazonaws.com` |
| WebSocket API | `wss://er7oc4a3o5.execute-api.us-west-2.amazonaws.com/prod` |
| Neon Database | `us-west-2` region |

---

## Checklist

### Before First Deploy
- [ ] AWS CLI and SAM CLI installed
- [ ] AWS SSO profile configured (`devhive`)
- [ ] Neon database created in `us-west-2`
- [ ] Google OAuth credentials ready
- [ ] Resend API key obtained
- [ ] Generate JWT signing key: `openssl rand -base64 32`

### After Deploy
- [ ] Health check returns 200
- [ ] Database migrations completed
- [ ] Google OAuth redirect URL updated in Google Console
- [ ] Frontend CORS origins match deployed API
- [ ] WebSocket connection test successful

### Before Production
- [ ] All secrets stored securely (not in code/commits)
- [ ] CloudWatch alarms configured
- [ ] Error monitoring set up
- [ ] Backup strategy confirmed (Neon auto-backups)

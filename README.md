# DevHive Backend - Go MVP

A modern, scalable Go backend for DevHive, featuring Firebase authentication, PostgreSQL database, and comprehensive project management capabilities.

## 🚀 Features

- **Authentication**: Firebase-based user authentication with JWT tokens
- **User Management**: User profiles, avatars, and role-based access control
- **Project Management**: Create, manage, and collaborate on projects
- **Sprint Planning**: Agile sprint management with start/end dates
- **Team Collaboration**: Role-based project membership (owner, admin, member, viewer)
- **Real-time Messaging**: Threaded project discussions with file support
- **File Storage**: Firebase Storage integration with Fly.io volume fallback
- **API-First Design**: RESTful API with comprehensive documentation
- **Production Ready**: Docker containerization, health checks, and monitoring

## 🏗️ Architecture

```
DevHive Backend
├── cmd/main.go          # Application entry point
├── config/              # Configuration management
├── db/                  # Database connection and schema
├── models/              # Data models and business logic
├── controllers/         # HTTP request handlers
├── storage/             # File storage abstraction
├── Dockerfile           # Container configuration
├── fly.toml            # Fly.io deployment config
└── .github/workflows/   # CI/CD automation
```

## 🛠️ Tech Stack

- **Language**: Go 1.22
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL with UUID support
- **Authentication**: Firebase Auth + JWT
- **Storage**: Firebase Storage + Fly.io volumes
- **Deployment**: Fly.io with Docker
- **CI/CD**: GitHub Actions

## 📋 Prerequisites

- Go 1.22 or later
- PostgreSQL 12 or later
- Firebase project with Auth and Storage enabled
- Fly.io account and CLI
- Docker (for local development)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd DevHive.GoBackend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Environment Configuration

Create a `.env` file in the root directory:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=devhive
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key

# Firebase Configuration
FIREBASE_SERVICE_ACCOUNT_KEY_PATH=./firebase-service-account.json
FIREBASE_STORAGE_BUCKET=your-project-id.appspot.com

# Application Configuration
PORT=8080
GIN_MODE=debug
```

### 4. Database Setup

```bash
# Create PostgreSQL database
createdb devhive

# Run schema migrations
psql -d devhive -f db/schema.sql
```

### 5. Firebase Setup

1. Download your Firebase service account key
2. Place it in the project root as `firebase-service-account.json`
3. Enable Firebase Auth and Storage in your Firebase console

### 6. Run the Application

```bash
# Development mode
go run cmd/main.go

# Production build
go build -o devhive cmd/main.go
./devhive
```

The API will be available at `http://localhost:8080`

## 🐳 Docker Deployment

### Build the Image

```bash
docker build -t devhive-backend .
```

### Run the Container

```bash
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  devhive-backend
```

## ☁️ Fly.io Deployment

### 1. Install Fly CLI

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh
```

### 2. Login to Fly.io

```bash
fly auth login
```

### 3. Create Volume for File Storage

```bash
fly volumes create devhive_data --size 10 --region dfw
```

### 4. Deploy

```bash
fly deploy
```

### 5. Set Secrets

```bash
fly secrets set DB_PASSWORD=your-db-password
fly secrets set JWT_SECRET=your-jwt-secret
fly secrets set FIREBASE_SERVICE_ACCOUNT_KEY="$(cat firebase-service-account.json)"
```

## 📚 API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "firebase_id_token": "firebase_token_here",
  "email": "user@example.com",
  "username": "username",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login User
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "firebase_id_token": "firebase_token_here"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

### User Endpoints

#### Get User Profile
```http
GET /api/v1/users/profile
Authorization: Bearer <access_token>
```

#### Update User Profile
```http
PUT /api/v1/users/profile
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "first_name": "Jane",
  "bio": "Updated bio"
}
```

#### Upload Avatar
```http
POST /api/v1/users/avatar
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

avatar: <file>
```

### Project Endpoints

#### Get Projects
```http
GET /api/v1/projects?limit=20&offset=0
Authorization: Bearer <access_token>
```

#### Create Project
```http
POST /api/v1/projects
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My Project",
  "description": "Project description"
}
```

#### Get Project
```http
GET /api/v1/projects/{project_id}
Authorization: Bearer <access_token>
```

#### Update Project
```http
PUT /api/v1/projects/{project_id}
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Project Name",
  "status": "active"
}
```

#### Delete Project
```http
DELETE /api/v1/projects/{project_id}
Authorization: Bearer <access_token>
```

### Sprint Endpoints

#### Get Sprints
```http
GET /api/v1/projects/{project_id}/sprints?limit=20&offset=0
Authorization: Bearer <access_token>
```

#### Create Sprint
```http
POST /api/v1/projects/{project_id}/sprints
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Sprint 1",
  "description": "First sprint",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-15T00:00:00Z"
}
```

### Message Endpoints

#### Get Messages
```http
GET /api/v1/projects/{project_id}/messages?limit=20&offset=0
Authorization: Bearer <access_token>
```

#### Create Message
```http
POST /api/v1/projects/{project_id}/messages
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "Hello team!",
  "message_type": "text"
}
```

#### Reply to Message
```http
POST /api/v1/projects/{project_id}/messages
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "Great idea!",
  "message_type": "text",
  "parent_message_id": "message-uuid-here"
}
```

## 🔐 Authentication

All protected endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <access_token>
```

Tokens are obtained through the authentication endpoints and expire after 1 hour. Use the refresh token endpoint to obtain new access tokens.

## 🗄️ Database Schema

The application uses PostgreSQL with the following main tables:

- **users**: User accounts and profiles
- **projects**: Project information and metadata
- **project_members**: Many-to-many relationship for project access
- **sprints**: Sprint planning and management
- **messages**: Project communication with threading support

## 📁 File Storage

Files are stored using a hybrid approach:

1. **Firebase Storage**: Primary storage for production
2. **Fly.io Volumes**: Fallback storage for local files
3. **Supported Types**: Images (JPEG, PNG, GIF, WebP), documents, and other files
4. **Size Limits**: 5MB for avatars, configurable for other uploads

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./controllers/...

# Run tests with verbose output
go test -v ./...
```

## 🔍 Monitoring

### Health Check
```http
GET /health
```

Returns application status and uptime information.

### Metrics (Optional)
```http
GET /metrics
```

Prometheus-compatible metrics endpoint.

## 🚀 Production Deployment

### Environment Variables

Ensure all required environment variables are set in production:

- Database credentials
- JWT secret (use a strong, random key)
- Firebase configuration
- CORS origins
- Logging level

### Security Considerations

- Use HTTPS in production
- Implement rate limiting
- Set secure headers
- Use environment-specific JWT secrets
- Regularly rotate secrets
- Monitor application logs

### Scaling

The application is designed to scale horizontally:

- Stateless design
- Database connection pooling
- Configurable concurrency limits
- Health checks for load balancers

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

- Create an issue in the GitHub repository
- Check the API documentation
- Review the troubleshooting guide

## 🔄 Changelog

### v1.0.0 (MVP)
- Initial release with core functionality
- User authentication and management
- Project and sprint management
- Team collaboration features
- File upload support
- RESTful API design

# Development Guide

This guide covers development practices, coding standards, and contribution guidelines for the MS-AI Backend project.

## 🏗️ Architecture Overview

The project follows **Hexagonal Architecture (Ports & Adapters)** pattern:

```
┌─────────────────────────────────────┐
│             API Layer               │
│  (HTTP Handlers, DTOs, Routing)    │
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│           Core Layer                │
│  (Business Logic, Use Cases)       │
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│           Data Layer                │
│  (Repositories, Database)          │
└─────────────────────────────────────┘
```

## 🧪 Testing Strategy

### Unit Tests
- Test business logic in isolation
- Mock external dependencies
- Focus on core domain logic

```go
// Example: internal/core/auth/auth_service_test.go
func TestSignUp_ValidCredentials_ReturnsToken(t *testing.T) {
    // Arrange
    mockRepo := &mocks.MockUserRepository{}
    service := auth.NewAuthService(mockRepo, cfg, log)

    // Act
    result, err := service.SignUp(ctx, "testuser", "password123", "invite123")

    // Assert
    require.NoError(t, err)
    require.NotNil(t, result.AccessToken)
}
```

### Integration Tests
- Test with real database
- Test API endpoints end-to-end
- Located in `test/integration/`

```go
// Example: test/integration/auth_test.go
func TestAuthAPI_SignUpAndLogin(t *testing.T) {
    // Setup test server with real DB
    // Test complete user journey
}
```

### Test Coverage
- Aim for >80% coverage
- Critical paths: 100% coverage
- Run `make test-coverage` to generate reports

## 📝 Coding Standards

### Go Best Practices

1. **Error Handling**
```go
// ✅ Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Bad: Lose error context
if err != nil {
    return errors.New("database error")
}
```

2. **Structured Logging**
```go
// ✅ Good: Structured fields
log.Info("user created",
    "user_id", userID,
    "username", username,
    "timestamp", time.Now(),
)

// ❌ Bad: String concatenation
log.Info(fmt.Sprintf("User %s created with ID %s", username, userID))
```

3. **Context Usage**
```go
// Always pass context for cancellation and tracing
func (s *Service) CreateUser(ctx context.Context, user *User) error {
    // Use ctx for database operations
    return s.repo.Create(ctx, user)
}
```

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `auth`, `user`, `manga`)
- **Interfaces**: end with `er` (e.g., `Repository`, `Service`)
- **Structs**: PascalCase (e.g., `UserService`, `AuthHandler`)
- **Methods**: PascalCase, descriptive (e.g., `CreateUser`, `ValidatePassword`)

### File Organization

```
internal/core/auth/
├── auth_service.go      # Business logic
├── auth_service_test.go # Unit tests
└── models.go           # Domain models

internal/api/handler/
├── auth_handler.go     # HTTP handlers
└── auth_handler_test.go # Handler tests
```

## 🔄 Development Workflow

### 1. Branching Strategy

```bash
# Feature branch
git checkout -b feature/user-profile-page

# Bug fix
git checkout -b bugfix/auth-validation-error

# Hotfix
git checkout -b hotfix/critical-security-patch
```

### 2. Commit Messages

Follow [Conventional Commits](https://conventionalcommits.org/):

```bash
feat: add user profile API endpoint
fix: resolve auth token validation bug
docs: update API documentation
refactor: simplify user service logic
test: add integration tests for auth flow
```

### 3. Pull Request Process

1. **Create PR**: Use descriptive title and detailed description
2. **Code Review**: At least one approval required
3. **Tests**: All CI checks must pass
4. **Merge**: Squash merge with descriptive commit message

### 4. Pre-commit Checks

```bash
# Run all checks before committing
make check

# Or individually
make fmt      # Format code
make lint     # Run linter
make security # Security checks
make test-unit # Unit tests
```

## 🐳 Docker Development

### Local Development with Docker

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Run tests in container
docker-compose exec app make test

# Clean up
make docker-clean
```

### Multi-stage Builds

The Dockerfile uses multi-stage builds for optimization:

```dockerfile
# Builder stage
FROM golang:1.24-alpine AS builder
# Build optimized binary

# Runtime stage
FROM alpine:3.19
# Minimal runtime image
```

## 🔧 Development Tools

### Required Tools

```bash
# Install development tools
make install-tools

# Tools installed:
# - air: Hot reload for Go
# - golangci-lint: Linter
# - goimports: Import formatter
# - gosec: Security linter
# - mockgen: Mock generator
```

### IDE Setup

#### VS Code
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v", "-race"],
    "go.testTimeout": "30s"
}
```

#### GoLand
- Enable Go modules
- Set GOPROXY=https://proxy.golang.org
- Configure golangci-lint as external tool

## 📊 Performance Guidelines

### Database Optimization

1. **Indexes**: Ensure proper indexes on frequently queried fields
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Query Optimization**: Use aggregation pipelines for complex queries

### Memory Management

1. **Avoid Memory Leaks**: Always close resources (connections, files)
2. **Use sync.Pool**: For frequently allocated objects
3. **Profile Memory**: Use `go tool pprof` for memory analysis

### Concurrent Programming

1. **Use Context**: For cancellation and timeouts
2. **Avoid Goroutine Leaks**: Always clean up goroutines
3. **Use Channels Wisely**: Prefer buffered channels for performance

## 🔒 Security Practices

### Input Validation

```go
// Always validate input
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
    Password string `json:"password" binding:"required,min=8"`
}

// Use domain validation
if err := validator.ValidateUsername(req.Username); err != nil {
    return err
}
```

### Authentication & Authorization

- JWT tokens with appropriate expiry
- Role-based access control
- Rate limiting on sensitive endpoints
- Secure password hashing with bcrypt

### Data Protection

- Encrypt sensitive data at rest
- Use HTTPS in production
- Implement proper CORS policies
- Sanitize user inputs

## 📈 Monitoring & Observability

### Logging

```go
// Structured logging with correlation IDs
log := logger.With().
    Str("request_id", requestID).
    Str("user_id", userID).
    Logger()

log.Info("user authenticated",
    "method", "login",
    "ip", clientIP,
)
```

### Health Checks

```go
// Application health
GET /health  // Basic health check
GET /ready   // Database connectivity check
GET /metrics // Prometheus metrics (future)
```

### Error Tracking

- Log errors with full context
- Use error codes for categorization
- Implement proper error responses

## 🚀 Deployment

### Environment Variables

```bash
# Production environment
export ENVIRONMENT=production
export MONGO_URI=mongodb://prod-server:27017
export JWT_SECRET=$(openssl rand -hex 32)
export CORS_ORIGINS=https://yourdomain.com
```

### Health Checks

```bash
# Docker health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

### Scaling Considerations

1. **Horizontal Scaling**: Stateless application design
2. **Database Scaling**: Use MongoDB replica sets
3. **Caching**: Implement Redis for session storage
4. **Load Balancing**: Use reverse proxy (nginx, traefik)

## 📚 Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Twelve-Factor App](https://12factor.net/)

## 🤝 Contributing

1. Read this development guide
2. Follow the coding standards
3. Write tests for new features
4. Update documentation
5. Create meaningful commit messages
6. Submit a well-documented PR
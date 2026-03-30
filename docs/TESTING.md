# Testing Guide

Comprehensive testing strategy for the MS-AI Backend project.

## 🧪 Testing Pyramid

```
          E2E Tests (Slow, High Value)
                 │
        Integration Tests (Medium, High Value)
                 │
       Unit Tests (Fast, High Coverage)
```

## 🏃 Unit Tests

### Location
- `*_test.go` files alongside source code
- Example: `internal/core/auth/auth_service_test.go`

### Principles
- Test business logic in isolation
- Mock external dependencies
- Fast execution (< 100ms per test)
- High coverage (>80%)

### Example

```go
// internal/core/auth/auth_service_test.go
package auth_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/your-org/ms-ai-backend/internal/core/auth"
    "github.com/your-org/ms-ai-backend/test/mocks"
)

func TestSignUp_ValidCredentials_ReturnsToken(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := &mocks.MockUserRepository{}
    mockRepo.On("FindByUsername", ctx, "testuser").Return(nil, nil)
    mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User"), mock.AnythingOfType("*user.UserDetails")).Return(nil)

    cfg := &config.Config{
        JWTSecret: "test-secret-key-at-least-32-chars-long",
        JWTAccessExpiry: time.Hour,
    }

    service := auth.NewDefaultAuthService(mockRepo, nil, cfg, nil)

    // Act
    result, err := service.SignUp(ctx, "testuser", "password123", "invite123")

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.AccessToken)
    mockRepo.AssertExpectations(t)
}

func TestSignUp_UserExists_ReturnsError(t *testing.T) {
    // Arrange
    ctx := context.Background()
    existingUser := &user.User{Username: "testuser"}

    mockRepo := &mocks.MockUserRepository{}
    mockRepo.On("FindByUsername", ctx, "testuser").Return(existingUser, nil)

    service := auth.NewDefaultAuthService(mockRepo, nil, &config.Config{}, nil)

    // Act
    result, err := service.SignUp(ctx, "testuser", "password123", "invite123")

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "user already exists")
}
```

### Mock Generation

```bash
# Generate mocks
make gen-mocks

# Or manually
mockgen -source=internal/core/auth/service.go \
        -destination=test/mocks/auth_service_mock.go \
        -package=mocks
```

## 🔗 Integration Tests

### Location
- `test/integration/` directory
- Example: `test/integration/auth_test.go`

### Setup
- Real database connection
- Full application stack
- Slower execution (integration tests)

### Example

```go
// test/integration/auth_test.go
package integration_test

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/your-org/ms-ai-backend/internal/api/router"
    "github.com/your-org/ms-ai-backend/internal/container"
    "github.com/your-org/ms-ai-backend/pkg/config"
)

func TestAuthAPI_SignUpAndLogin(t *testing.T) {
    // Setup test database
    mongoClient, cleanup := setupTestDatabase(t)
    defer cleanup()

    // Setup application
    cfg := &config.Config{
        JWTSecret: "test-secret-key-at-least-32-chars-long",
        JWTAccessExpiry: time.Hour,
        MongoURI: "mongodb://localhost:27017/testdb",
    }

    container, err := container.NewContainer(cfg)
    require.NoError(t, err)
    defer container.Close(context.Background())

    // Setup router
    r := gin.New()
    v1.SetupRoutes(r, container.Handlers, container.Middlewares)

    // Test signup
    signupReq := map[string]string{
        "username": "testuser",
        "password": "password123",
        "invite_code": "TEST1234",
    }
    signupBody, _ := json.Marshal(signupReq)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(signupBody))
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var signupResp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &signupResp)
    assert.NotEmpty(t, signupResp["data"].(map[string]interface{})["access_token"])

    // Test login
    loginReq := map[string]string{
        "username": "testuser",
        "password": "password123",
    }
    loginBody, _ := json.Marshal(loginReq)

    w = httptest.NewRecorder()
    req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func setupTestDatabase(t *testing.T) (*mongo.Client, func()) {
    ctx := context.Background()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    require.NoError(t, err)

    // Create test database
    db := client.Database("testdb")

    // Clean up function
    cleanup := func() {
        db.Drop(ctx)
        client.Disconnect(ctx)
    }

    return client, cleanup
}
```

## 🌐 API Contract Tests

### Location
- `test/integration/api_test.go`

### Purpose
- Test API contracts
- Ensure backward compatibility
- Validate response schemas

### Example

```go
func TestAuthAPI_Contract(t *testing.T) {
    tests := []struct {
        name           string
        method         string
        path           string
        body           interface{}
        expectedStatus int
        expectedFields []string
    }{
        {
            name:           "signup success",
            method:         "POST",
            path:           "/api/v1/auth/signup",
            body:           map[string]string{"username": "user", "password": "pass", "invite_code": "code"},
            expectedStatus: 201,
            expectedFields: []string{"access_token", "refresh_token", "user"},
        },
        {
            name:           "signup validation error",
            method:         "POST",
            path:           "/api/v1/auth/signup",
            body:           map[string]string{"username": "", "password": "pass"},
            expectedStatus: 400,
            expectedFields: []string{"error"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 🏃 Running Tests

### Local Development

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run specific test
go test ./internal/core/auth -v -run TestSignUp

# Run tests in watch mode
make dev-test
```

### CI/CD

```yaml
# .github/workflows/ci.yml
- name: Run unit tests
  run: make test-unit

- name: Run integration tests
  run: make test-integration
  env:
    MONGO_URI: mongodb://localhost:27017
```

### Docker Testing

```bash
# Run tests in Docker
docker-compose exec app make test

# Run integration tests with real DB
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 📊 Test Coverage

### Requirements
- **Overall Coverage**: >80%
- **Critical Paths**: >90%
- **New Features**: >85%

### Coverage Report

```bash
# Generate HTML coverage report
make test-coverage

# View coverage by package
go tool cover -func=coverage.out

# Coverage by function
go tool cover -func=coverage.out | grep -E "(total|auth|user|manga)"
```

### Coverage Badges

```markdown
[![codecov](https://codecov.io/gh/your-org/ms-ai-backend/branch/main/graph/badge.svg)](https://codecov.io/gh/your-org/ms-ai-backend)
```

## 🛠️ Testing Tools

### Required Dependencies

```go
// go.mod
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/testcontainers/testcontainers-go v0.20.0
)
```

### Test Utilities

```go
// test/helpers.go
package test

import (
    "context"
    "testing"
    "time"

    "github.com/your-org/ms-ai-backend/internal/container"
    "github.com/your-org/ms-ai-backend/pkg/config"
)

// SetupTestContainer creates a test container with mocked dependencies
func SetupTestContainer(t *testing.T) (*container.Container, func()) {
    cfg := &config.Config{
        JWTSecret: "test-secret-key-at-least-32-chars-long",
        JWTAccessExpiry: time.Hour,
        Environment: "test",
    }

    container, err := container.NewContainer(cfg)
    require.NoError(t, err)

    cleanup := func() {
        container.Close(context.Background())
    }

    return container, cleanup
}

// SetupTestDB creates a test database
func SetupTestDB(t *testing.T) (*mongo.Database, func()) {
    // Implementation
}
```

## 📋 Test Categories

### 1. Unit Tests
- **Scope**: Single function/method
- **Dependencies**: Mocked
- **Speed**: Fast (<100ms)
- **Coverage**: High

### 2. Integration Tests
- **Scope**: Multiple components
- **Dependencies**: Real (database, external APIs)
- **Speed**: Medium (1-10s)
- **Coverage**: Medium

### 3. End-to-End Tests
- **Scope**: Full application
- **Dependencies**: All real services
- **Speed**: Slow (10s+)
- **Coverage**: Low (critical paths only)

### 4. Performance Tests
- **Scope**: Load and stress testing
- **Tools**: Apache Bench, hey, k6
- **Metrics**: Response time, throughput, resource usage

## 🔧 Test Best Practices

### Naming Conventions

```go
// Test[Function]_[Condition]_[ExpectedResult]
func TestSignUp_ValidCredentials_ReturnsToken(t *testing.T)
func TestSignUp_InvalidPassword_ReturnsError(t *testing.T)
func TestSignUp_UserExists_ReturnsConflictError(t *testing.T)
```

### Test Structure

```go
func TestFunction(t *testing.T) {
    t.Run("subtest name", func(t *testing.T) {
        // Arrange
        // Act
        // Assert
    })
}
```

### Table-Driven Tests

```go
func TestValidatePassword(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"valid password", "ValidPass123!", false},
        {"too short", "Short1!", true},
        {"no uppercase", "nouppercase123!", true},
        {"no special char", "NoSpecial123", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidatePassword(tt.password)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Parallel Tests

```go
func TestParallel(t *testing.T) {
    t.Parallel() // Run tests in parallel

    // Test implementation
}
```

## 🚨 Test Flakiness

### Common Causes
- Race conditions
- Time-dependent logic
- External dependencies
- Shared state

### Solutions

```go
// Use t.Parallel() for independent tests
func TestIndependent(t *testing.T) {
    t.Parallel()
    // Test that doesn't depend on shared state
}

// Use proper cleanup
func TestWithCleanup(t *testing.T) {
    // Setup
    resource := setupResource(t)
    t.Cleanup(func() {
        cleanupResource(resource)
    })

    // Test
}

// Avoid time.Sleep in tests
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Use ctx for operations
}
```

## 📈 Test Metrics

### Coverage Goals
- **Unit Tests**: >80% coverage
- **Integration Tests**: Cover critical user journeys
- **API Tests**: Cover all endpoints

### Performance Benchmarks

```go
func BenchmarkSignUp(b *testing.B) {
    // Setup
    service := setupBenchmarkService()

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            service.SignUp(context.Background(), "user", "pass", "code")
        }
    })
}
```

## 🔍 Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./internal/core/auth

# Run specific test with verbose
go test -v -run TestSignUp ./internal/core/auth
```

### Debug Flags

```bash
# Enable race detection
go test -race ./...

# Enable CPU profiling
go test -cpuprofile cpu.prof ./...

# View profiling data
go tool pprof cpu.prof
```

### Test Debugging Tips

1. **Use t.Log()** for debug output
2. **Check error messages** carefully
3. **Verify mock expectations** are met
4. **Use testify's require** for early failures
5. **Check test isolation** - no shared state

## 📚 Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://godoc.org/github.com/stretchr/testify)
- [GoMock Documentation](https://github.com/golang/mock)
- [Test-Driven Development](https://en.wikipedia.org/wiki/Test-driven_development)
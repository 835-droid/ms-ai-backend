# MS-AI Backend

[![CI](https://github.com/your-org/ms-ai-backend/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/ms-ai-backend/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/ms-ai-backend)](https://goreportcard.com/report/github.com/your-org/ms-ai-backend)
[![codecov](https://codecov.io/gh/your-org/ms-ai-backend/branch/main/graph/badge.svg)](https://codecov.io/gh/your-org/ms-ai-backend)

A production-ready, enterprise-grade manga management platform built with Go, featuring hexagonal architecture, comprehensive testing, and cloud-native deployment capabilities.

## 🏗️ Architecture

This project follows **Hexagonal Architecture (Ports & Adapters)** with clear separation of concerns:

```
├── cmd/                    # Application entrypoints
│   ├── server/            # HTTP server
│   ├── cli/               # CLI tools
│   └── migrations/        # Database migrations
├── internal/              # Private application code
│   ├── api/               # HTTP layer (controllers, DTOs, routing)
│   │   ├── v1/           # API versioning
│   │   ├── dto/          # Data Transfer Objects
│   │   ├── handler/      # HTTP handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── router/       # Route definitions
│   ├── core/             # Business logic (use cases, domain models)
│   ├── data/             # Data access layer (repositories)
│   └── infra/            # Infrastructure concerns
├── pkg/                  # Public packages
│   ├── config/           # Configuration management
│   ├── errors/           # Domain errors
│   ├── jwt/              # JWT utilities
│   ├── logger/           # Structured logging
│   └── response/         # HTTP response helpers
├── test/                 # Testing utilities
│   ├── integration/      # Integration tests
│   └── fixtures/         # Test data
└── docs/                 # Documentation
```

## ✨ Features

### 🔐 Authentication & Authorization
- JWT-based authentication with refresh tokens
- Invite-code registration system
- Role-based access control (User, Admin)
- Secure password hashing with bcrypt
- Rate limiting and brute-force protection

### 📚 Manga Management
- Complete manga metadata management
- Chapter organization and storage
- Tagging and categorization system
- Search and filtering capabilities
- Bulk operations support

### 👥 User Management
- User profiles and preferences
- Admin dashboard for content management
- User activity tracking
- Account status management

### 🛡️ Security
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CORS configuration
- Security headers

### 📊 Monitoring & Observability
- Structured logging with correlation IDs
- Health checks and readiness probes
- Metrics collection (future)
- Distributed tracing support (future)

### 🧪 Testing
- Unit tests with mocks
- Integration tests with real database
- API contract testing
- Performance testing utilities

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- MongoDB 6.0+
- Docker & Docker Compose (recommended)

### 1. Clone and Setup

```bash
git clone https://github.com/your-org/ms-ai-backend.git
cd ms-ai-backend

# Copy environment configuration
cp .env.example .env

# Edit .env with your settings
nano .env
```

### 2. Environment Configuration

```env
# Database
MONGO_URI=mongodb://localhost:27017
DB_NAME=MSAIDB

# Server
SERVER_PORT=8080
ENVIRONMENT=development

# Security (generate secure values)
JWT_SECRET=$(openssl rand -hex 32)

# CORS (update for your frontend)
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

### 3. Start Services

```bash
# Start with Docker Compose (recommended)
docker-compose up -d

# Or start manually
make build
./bin/server
```

### 4. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# API documentation
open http://localhost:8080/docs
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run specific test
go test ./internal/core/auth -v -run TestSignUp
```

## 📦 Build & Deployment

### Local Development

```bash
# Build
make build

# Run
./bin/server

# Development with hot reload
make dev
```

### Docker

```bash
# Build image
docker build -t ms-ai-backend .

# Run with Docker Compose
docker-compose up -d

# Run specific services
docker-compose up -d mongodb app
```

### Production Deployment

```bash
# Build optimized binary
make build-prod

# Deploy with docker-compose.prod.yml
docker-compose -f docker-compose.prod.yml up -d
```

## 🔧 Development

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Run security checks
make security

# Pre-commit checks
make check
```

### Database Operations

```bash
# Create invite code
go run cmd/utils/gen_invite.go

# Run migrations
go run cmd/migrations/migrate.go up

# Seed database
go run cmd/utils/seed.go
```

### API Documentation

- **Swagger/OpenAPI**: `http://localhost:8080/swagger/`
- **API Reference**: See [docs/API.md](docs/API.md)
- **Architecture**: See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## 📊 Monitoring

### Health Checks

```bash
# Application health
GET /health

# Database connectivity
GET /ready

# Metrics (future)
GET /metrics
```

### Logs

```bash
# Structured JSON logs
tail -f server.log

# Filter by level
jq 'select(.level == "error")' server.log
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Write tests for your changes
4. Ensure all tests pass: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Development Guidelines

- Follow Go best practices and effective Go guidelines
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure code passes all CI checks
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Gin](https://gin-gonic.com/) web framework
- Database operations with [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- Authentication with [JWT](https://github.com/golang-jwt/jwt)
- Logging with [Zerolog](https://github.com/rs/zerolog)

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-org/ms-ai-backend/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/ms-ai-backend/discussions)
- **Documentation**: [docs/](docs/)
   mongod
   ```

4. **Build and run**:
   ```bash
   make build
   make run
   ```

5. **Access the application**:
   - Open `http://localhost:8080` in your browser
   - Register using an invite code (generate from admin panel)

## API Documentation

Complete API documentation is available in [`docs/API.md`](docs/API.md).

### Key Endpoints

- `POST /api/auth/signup` - User registration
- `POST /api/auth/login` - User authentication
- `GET /api/mangas` - List manga
- `POST /api/mangas` - Create manga
- `GET /api/mangas/{id}/chapters` - Get chapters
- `POST /api/mangas/{id}/chapters` - Add chapter

## Project Structure

```
├── cmd/
│   ├── server/          # Main application entry point
│   └── web/            # Static web files
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── core/           # Business logic services
│   ├── data/           # Data repositories
│   └── pkg/            # Shared utilities
├── docs/               # Documentation
└── scripts/            # Database seeding and utilities
```

## Development

### Available Commands

```bash
make build      # Build the application
make run        # Run the application
make test       # Run tests
make clean      # Clean build artifacts
make docker     # Build Docker image
```

### Database Seeding

To populate the database with sample data:

```bash
go run scripts/seed.go
```

### Testing

```bash
go test ./...
```

## Deployment

See [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) for detailed deployment instructions.

### Docker Deployment

```bash
docker build -t ms-ai .
docker run -p 8080:8080 --env-file .env ms-ai
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `CORS_ORIGINS` | Allowed CORS origins | `http://localhost:8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/msai` |
| `JWT_SECRET` | JWT signing secret | Required |
| `JWT_ACCESS_EXPIRY` | Access token expiry | `15m` |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `168h` |

## Troubleshooting

### Common Issues

- **CORS errors**: Check `CORS_ORIGINS` in your `.env` file
- **Database connection**: Ensure MongoDB is running and accessible
- **JWT errors**: Verify `JWT_SECRET` is set and consistent
- **WebSocket issues**: Check CORS settings for WebSocket connections

### Logs

The application uses structured logging. Check the console output for detailed error information.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

API Endpoints
- POST /api/auth/signup
- POST /api/auth/login
- POST /api/auth/refresh
- POST /api/auth/logout
- POST /api/admin/invite (admin only)
- GET /api/admin/invites (admin only)
- DELETE /api/admin/invite/:id (admin only)
- **GET /api/chat/ws (auth required, WebSocket)**
- **GET /api/chat/history (auth required)**
- **GET /api/chat/users (auth required)**
- GET /livez
- GET /readyz

See `docs/API.md` for full API details and `docs/DEPLOYMENT.md` for deployment instructions.
# MS-AI Backend

A modern, scalable manga reading platform backend built with Go, following Clean Architecture and Hexagonal Architecture principles.

## 🚀 Features

- **User Authentication & Authorization** - JWT-based auth with refresh tokens
- **Manga Management** - CRUD operations for manga and chapters
- **Reading Progress Tracking** - Save and track reading progress
- **Viewing History** - Track and manage viewing history
- **Favorite Lists** - Create custom manga collections
- **Ratings & Reviews** - Rate manga and chapters
- **Comments System** - Comment on manga and chapters
- **Admin Dashboard** - User and content management
- **Hybrid Database** - MongoDB + PostgreSQL for optimal performance
- **RESTful API** - Well-documented API endpoints
- **Swagger Documentation** - Interactive API documentation

## 🏗️ Architecture

This project follows **Clean Architecture** and **Hexagonal Architecture** principles:

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                       │
│  (HTTP Handlers, Routes, Middleware, Request/Response DTOs) │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│        (Use Cases, Service Interfaces, Application DTOs)     │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                            │
│         (Entities, Business Rules, Repository Interfaces)    │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                       │
│     (Repository Implementations, Database, External APIs)    │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
MS-AI/
├── cmd/                        # Application entry points
│   ├── server/                 # Main server
│   ├── create_admin/           # Admin user creation utility
│   ├── utils/                  # Utility commands
│   └── web/                    # Static web assets
│
├── internal/                   # Private application code
│   ├── domain/                 # Enterprise business rules
│   │   ├── manga/              # Manga domain entities & interfaces
│   │   └── user/               # User domain entities & interfaces
│   │
│   ├── application/            # Application business rules
│   │   ├── interfaces/         # Service interfaces (ports)
│   │   ├── dtos/               # Data transfer objects
│   │   └── services/           # (Future: Use case implementations)
│   │
│   ├── core/                   # Business logic implementations
│   │   ├── content/manga/      # Manga services
│   │   ├── auth/               # Authentication service
│   │   ├── admin/              # Admin service
│   │   └── user/               # User domain logic
│   │
│   ├── infrastructure/         # Framework & driver implementations
│   │   └── persistence/        # (Future: Database connections)
│   │
│   ├── data/                   # Data access layer
│   │   ├── content/manga/      # Manga repositories
│   │   ├── user/               # User repositories
│   │   ├── admin/              # Admin repositories
│   │   └── infrastructure/     # Database connections
│   │       ├── mongo/          # MongoDB setup
│   │       └── postgres/       # PostgreSQL setup
│   │
│   ├── api/                    # HTTP presentation layer
│   │   ├── handler/            # HTTP request handlers
│   │   ├── router/             # Route definitions
│   │   ├── middleware/         # HTTP middleware
│   │   └── dto/                # API-specific DTOs
│   │
│   ├── container/              # Dependency injection
│   │   ├── container.go        # Main container
│   │   ├── types.go            # Container types
│   │   └── *_initializers.go   # Component initializers
│   │
│   └── delivery/               # (Future: Additional delivery mechanisms)
│
├── pkg/                        # Public utilities
│   ├── config/                 # Configuration management
│   ├── errors/                 # Common error types
│   ├── i18n/                   # Internationalization
│   ├── jwt/                    # JWT utilities
│   ├── logger/                 # Logging utilities
│   ├── response/               # Response helpers
│   ├── utils/                  # General utilities
│   └── validator/              # Input validation
│
├── docs/                       # Documentation
│   ├── API.md                  # API reference
│   ├── ARCHITECTURE.md         # Architecture documentation
│   ├── DATABASE.md             # Database documentation
│   ├── DEPLOYMENT.md           # Deployment guide
│   ├── DEVELOPMENT.md          # Development guide
│   └── TESTING.md              # Testing guide
│
├── scripts/                    # Utility scripts
│   ├── init-mongo.js/          # MongoDB initialization
│   ├── migrate_postgres.sql    # Database migrations
│   └── seed.go                 # Data seeding
│
├── test/                       # Test utilities
├── tools/                      # Development tools
├── uploads/                    # File upload directory
│
├── docker-compose.yml          # Docker Compose configuration
├── Dockerfile                  # Docker image definition
├── go.mod                      # Go module definition
├── Makefile                    # Build automation
└── .env                        # Environment configuration
```

## 🛠️ Tech Stack

- **Language:** Go 1.24
- **Framework:** Gin Web Framework
- **Databases:**
  - MongoDB (Document storage)
  - PostgreSQL (Relational data)
- **Authentication:** JWT (JSON Web Tokens)
- **Validation:** go-playground/validator
- **Logging:** rs/zerolog
- **Configuration:** godotenv

## 📋 Prerequisites

- Go 1.24 or higher
- MongoDB 6.0 or higher
- PostgreSQL 15 or higher
- Docker & Docker Compose (optional)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/835-droid/ms-ai-backend.git
cd ms-ai-backend/MS-AI
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start Databases (Using Docker)

```bash
docker-compose up -d
```

### 4. Run Migrations

```bash
# MongoDB indexes are created automatically
# PostgreSQL migrations
psql -U postgres -d ms_ai -f scripts/migrate_postgres.sql
```

### 5. Build and Run

```bash
# Build
go build -o ms-ai-server ./cmd/server

# Run
./ms-ai-server
```

Or use the Makefile:

```bash
make run
```

### 6. Create Admin User

```bash
go run ./cmd/create_admin
```

## 📚 API Documentation

### Swagger Documentation

The API is documented using Swagger annotations. To generate and view the Swagger UI:

```bash
# Install swag (if not installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
swag init -g cmd/server/main.go -o ./docs/swagger

# Start server and visit http://localhost:8080/swagger/index.html
```

### API Reference

See [docs/API_REFERENCE.md](docs/API_REFERENCE.md) for complete API documentation.

### Base URL

```
http://localhost:8080/api/v1
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test file
go test ./internal/domain/manga/...
```

## 📦 Deployment

### Docker Deployment

```bash
# Build Docker image
docker build -t ms-ai-backend .

# Run with Docker Compose
docker-compose up -d
```

### Production Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

## 🔧 Configuration

Configuration is managed through environment variables. See `.env.example` for all available options:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `POSTGRES_URI` | PostgreSQL connection string | `postgres://localhost:5432/ms_ai` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `LOG_LEVEL` | Logging level | `info` |

## 🏛️ Architecture Decisions

### Why Clean Architecture?

1. **Testability** - Business logic can be tested without UI, DB, or external services
2. **Maintainability** - Clear separation of concerns makes code easier to understand
3. **Flexibility** - Easy to swap implementations (e.g., change database)
4. **Scalability** - Well-organized codebase scales better with team size

### Why Hybrid Database?

- **MongoDB** - Flexible schema for manga content, chapters, and user-generated content
- **PostgreSQL** - ACID compliance for transactions, favorites lists, and relational data

### Dependency Injection

The project uses a custom DI container (`internal/container/`) for:
- Loose coupling between components
- Easy testing with mock implementations
- Centralized configuration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices
- Write tests for new features
- Update documentation
- Use meaningful commit messages

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- **835-droid** - Initial work

## 🙏 Acknowledgments

- Gin Web Framework team
- MongoDB team
- PostgreSQL team
- All open-source contributors

## 📞 Support

- **Documentation:** [docs/](docs/)
- **API Reference:** [docs/API_REFERENCE.md](docs/API_REFERENCE.md)
- **Issues:** [GitHub Issues](https://github.com/835-droid/ms-ai-backend/issues)

---

**Built with ❤️ using Go and Clean Architecture**
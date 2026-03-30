# Architecture

## Overview

MS-AI is a specialized manga management platform built with Go and MongoDB. The application follows a clean layered architecture designed for scalability and maintainability.

## Architecture Layers

```
cmd/           # Application entry points
├── server/    # Main server application
└── web/       # Static web files (HTML/CSS/JS)

internal/
├── api/       # HTTP API layer
│   ├── handler/     # HTTP request handlers
│   ├── middleware/  # HTTP middleware (auth, cors, logging, etc.)
│   └── router/      # Route definitions and setup
├── core/      # Business logic layer
│   ├── common/      # Shared business logic
│   ├── content/     # Content-specific business logic
│   └── user/        # User management business logic
├── data/      # Data access layer
│   ├── common/      # Shared data models
│   ├── content/     # Content data repositories
│   ├── mongo/       # MongoDB connection and utilities
│   └── user/        # User data repositories
└── pkg/       # Shared utilities and libraries
    ├── cache/       # Caching utilities
    ├── config/      # Configuration management
    ├── jwt/         # JWT token handling
    ├── logger/      # Structured logging
    ├── response/    # HTTP response utilities
    ├── utils/       # General utilities (slugify, ID generation)
    └── validator/   # Data validation utilities
```

## Key Design Decisions

### Technology Stack
- **Backend**: Go with Gin web framework
- **Database**: MongoDB with official Go driver
- **Authentication**: JWT tokens (access + refresh tokens)
- **Logging**: Zerolog for structured logging
- **Validation**: Custom validation with reflection
- **Frontend**: Vanilla JavaScript with responsive design

### Architecture Principles
- **Layered Architecture**: Clear separation of concerns between HTTP handling, business logic, and data access
- **Dependency Injection**: Services and repositories are injected for testability
- **Stateless Authentication**: JWT tokens eliminate server-side session storage
- **Rate Limiting**: Protection against abuse with configurable limits
- **Structured Logging**: Consistent logging format for debugging and monitoring

### Security Features
- JWT-based authentication with refresh token rotation
- Password hashing with bcrypt
- CORS protection
- Rate limiting on sensitive endpoints
- Input validation and sanitization
- Authorization checks (owner/admin permissions)

### Database Design
- **MongoDB Collections**:
  - `users`: User accounts and authentication data
  - `invites`: Invitation codes for registration
  - `mangas`: Manga metadata and content information
  - `manga_chapters`: Individual chapters with page references
- **Indexing**: Optimized indexes for common query patterns
- **Transactions**: Atomic operations for data consistency

### API Design
- RESTful endpoints with consistent response formats
- JSON request/response bodies
- Proper HTTP status codes
- Pagination for list endpoints
- Comprehensive error handling

### Content Management
- **Manga Structure**:
  - Title, description, cover image, tags
  - Author attribution and ownership
  - Publication status and timestamps
- **Chapter Management**:
  - Ordered chapters with numbering
  - Page storage (URLs to images)
  - Creation and update timestamps

### Frontend Architecture
- **Vanilla JavaScript**: No heavy frameworks for better performance
- **Responsive Design**: Mobile-first approach with CSS Grid/Flexbox
- **Component-based**: Reusable UI components
- **State Management**: Local storage for authentication tokens
- **Progressive Enhancement**: Works without JavaScript (basic functionality)

## Deployment Considerations

- **Containerization**: Docker support for consistent deployment
- **Environment Configuration**: Configurable via environment variables
- **Health Checks**: Built-in health endpoint for monitoring
- **Graceful Shutdown**: Proper cleanup on termination
- **Static File Serving**: Efficient serving of web assets

## Performance Optimizations

- **Database Indexing**: Optimized queries with proper indexes
- **Connection Pooling**: MongoDB connection reuse
- **Caching**: In-memory caching for frequently accessed data
- **Rate Limiting**: Prevents abuse and ensures fair usage
- **Pagination**: Efficient handling of large datasets
- **Lazy Loading**: On-demand loading of related data

## Monitoring and Observability

- **Structured Logging**: Comprehensive logging with context
- **Health Endpoints**: Service availability monitoring
- **Error Tracking**: Detailed error responses and logging
- **Performance Metrics**: Response times and throughput tracking

# Refactoring Plan - MS-AI Backend

## Overview

This document outlines the comprehensive refactoring plan to transform the MS-AI backend into a clean, maintainable architecture following **Clean Code** and **Hexagonal Architecture** principles.

## Current Status: Phase 1 Completed ✅

We have successfully created the new architectural foundation:

### ✅ Completed Work

#### 1. Domain Layer (`internal/domain/`)
- **`domain/manga/`**: Core manga entities
  - `manga.go` - Main Manga entity
  - `chapter.go` - Chapter entity
  - `rating.go` - Rating value objects
  - `reaction.go` - Reaction types and entities
  - `favorite.go` - Favorite and FavoriteList entities
  - `comment.go` - Comment entities
  - `viewing_history.go` - Viewing history and reading progress
  - `errors.go` - Domain-specific errors

- **`domain/user/`**: User entities
  - `user.go` - User entity with role management
  - `invite_code.go` - Invite code entity

#### 2. Application Layer (`internal/application/`)
- **`interfaces/repositories/`**: Repository ports
  - `user_repository.go` - User, InviteCode, UserToken repositories
  - `manga_repository.go` - All manga-related repositories

- **`interfaces/services/`**: Service ports
  - `auth_service.go` - Authentication operations
  - `manga_service.go` - Manga business logic
  - `admin_service.go` - Admin operations

- **`dtos/`**: Data Transfer Objects
  - `auth_dtos.go` - Auth-related DTOs
  - `manga_dtos.go` - Manga-related DTOs

## 🔄 Remaining Work

### Phase 2: Move Infrastructure Layer

Move existing data access code to `internal/infrastructure/persistence/`:

```bash
# Move MongoDB repositories
internal/data/content/manga/mongo_*.go → internal/infrastructure/persistence/mongodb/
internal/data/user/mongo_*.go → internal/infrastructure/persistence/mongodb/

# Move PostgreSQL repositories  
internal/data/content/manga/postgres_*.go → internal/infrastructure/persistence/postgres/
internal/data/user/postgres_*.go → internal/infrastructure/persistence/postgres/

# Move hybrid repositories
internal/data/content/manga/hybrid_*.go → internal/infrastructure/persistence/hybrid/
internal/data/user/hybrid_*.go → internal/infrastructure/persistence/hybrid/

# Move database setup
internal/data/infrastructure/mongo/ → internal/infrastructure/database/
internal/data/infrastructure/postgres/ → internal/infrastructure/database/
```

### Phase 3: Move Delivery Layer

Move API handlers and routers to `internal/delivery/http/`:

```bash
# Move handlers
internal/api/handler/ → internal/delivery/http/handlers/
internal/api/middleware/ → internal/delivery/http/middleware/
internal/api/router/ → internal/delivery/http/routers/

# Move response utilities
internal/api/dto/ → internal/application/dtos/ (merge with existing)
pkg/response/ → internal/delivery/http/responses/
```

### Phase 4: Create Shared Kernel

Move common utilities to `internal/shared/`:

```bash
pkg/config/ → internal/infrastructure/config/
pkg/logger/ → internal/shared/logger/
pkg/utils/ → internal/shared/utils/
pkg/jwt/ → internal/shared/jwt/
pkg/i18n/ → internal/shared/i18n/
pkg/validator/ → internal/shared/utils/validator.go
pkg/errors/ → internal/shared/errors/
```

### Phase 5: Update Container/DI

Create a new dependency injection container:

```go
// internal/delivery/http/container.go
package http

import (
    "github.com/835-droid/ms-ai-backend/internal/application/interfaces/repositories"
    "github.com/835-droid/ms-ai-backend/internal/application/interfaces/services"
    "github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/mongodb"
    "github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/postgres"
)

type Container struct {
    Config     *config.Config
    Logger     *logger.Logger
    MongoDB    *mongodb.MongoStore
    PostgresDB *postgres.PostgresStore
    
    // Repositories
    UserRepository      repositories.UserRepository
    MangaRepository     repositories.MangaRepository
    ChapterRepository   repositories.MangaChapterRepository
    FavoriteListRepo    repositories.FavoriteListRepository
    ViewingHistoryRepo  repositories.ViewingHistoryRepository
    
    // Services
    AuthService      services.AuthService
    MangaService     services.MangaService
    ChapterService   services.MangaChapterService
    FavListService   services.FavoriteListService
    AdminService     services.AdminService
    ViewingHistService services.ViewingHistoryService
    
    // Handlers
    AuthHandler           *handlers.AuthHandler
    MangaHandler          *handlers.MangaHandler
    ChapterHandler        *handlers.MangaChapterHandler
    FavListHandler        *handlers.FavoriteListHandler
    ViewingHistoryHandler *handlers.ViewingHistoryHandler
    AdminHandler          *handlers.AdminHandler
}

func NewContainer(cfg *config.Config, log *logger.Logger) (*Container, error) {
    // Initialize databases
    // Initialize repositories
    // Initialize services
    // Initialize handlers
    // Return container
}
```

### Phase 6: Update Main Entry Point

Update `cmd/server/main.go`:

```go
package main

import (
    "log"
    "github.com/835-droid/ms-ai-backend/internal/delivery/http"
    "github.com/835-droid/ms-ai-backend/internal/infrastructure/config"
    "github.com/835-droid/ms-ai-backend/internal/shared/logger"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    log := logger.NewLogger(cfg)

    // Initialize container (DI)
    container, err := http.NewContainer(cfg, log)
    if err != nil {
        log.Fatalf("Failed to initialize container: %v", err)
    }
    defer container.Close()

    // Setup router
    router := http.SetupRouter(cfg, log, container)

    // Start server
    log.Info("Starting server", map[string]interface{}{
        "port": cfg.ServerPort,
    })
    
    if err := router.Run(":" + cfg.ServerPort); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### Phase 7: Split Large Files

Break down large handler files into smaller, focused files:

#### Split `manga_handler.go` (484 lines):
- `manga_handler.go` - CRUD operations only (~150 lines)
- `manga_interaction_handler.go` - Reactions, favorites (~200 lines)
- `manga_rating_handler.go` - Rating operations (~100 lines)

#### Split `manga_service.go` (494 lines):
- `manga_crud_service.go` - Create, Read, Update, Delete
- `manga_interaction_service.go` - Favorites, reactions
- `manga_rating_service.go` - Ratings

## 📋 Implementation Checklist

### Phase 2: Infrastructure Layer
- [ ] Create `internal/infrastructure/persistence/mongodb/` directory structure
- [ ] Create `internal/infrastructure/persistence/postgres/` directory structure
- [ ] Create `internal/infrastructure/persistence/hybrid/` directory structure
- [ ] Move MongoDB repository files
- [ ] Move PostgreSQL repository files
- [ ] Move hybrid repository files
- [ ] Update all import paths
- [ ] Test database connections

### Phase 3: Delivery Layer
- [ ] Create `internal/delivery/http/handlers/` subdirectories
- [ ] Move handler files
- [ ] Move middleware files
- [ ] Move router files
- [ ] Update all import paths
- [ ] Test API endpoints

### Phase 4: Shared Kernel
- [ ] Move config to infrastructure
- [ ] Move shared utilities
- [ ] Update all import paths
- [ ] Test utility functions

### Phase 5: Dependency Injection
- [ ] Create new container structure
- [ ] Implement repository initialization
- [ ] Implement service initialization
- [ ] Implement handler initialization
- [ ] Update main.go to use new container

### Phase 6: Main Entry Point
- [ ] Update `cmd/server/main.go`
- [ ] Test application startup
- [ ] Test graceful shutdown

### Phase 7: File Splitting
- [ ] Split manga_handler.go
- [ ] Split manga_service.go
- [ ] Update imports
- [ ] Test all functionality

## 🎯 Benefits of This Refactoring

1. **Clear Separation of Concerns**: Each layer has a specific responsibility
2. **Testability**: Easy to mock dependencies and test in isolation
3. **Maintainability**: Code is organized logically and easy to navigate
4. **Flexibility**: Easy to swap implementations (e.g., MongoDB ↔ PostgreSQL)
5. **Scalability**: Easy to add new features without breaking existing code
6. **Professional Naming**: Clear, descriptive names following Go conventions

## ⚠️ Important Notes

- **No Functional Changes**: The business logic remains exactly the same
- **Backward Compatibility**: All existing API endpoints will work the same
- **Gradual Migration**: Can be done incrementally to minimize risk
- **Testing Required**: Each phase should be tested before moving to the next

## 🚀 Next Steps

1. Review this plan and provide feedback
2. Proceed with Phase 2 (Infrastructure Layer migration)
3. Continue through remaining phases sequentially
4. Perform comprehensive testing after each phase
5. Deploy to staging environment for validation
6. Deploy to production after full validation

---

**Status**: Phase 1 Complete ✅ | **Next**: Phase 2 - Infrastructure Layer Migration
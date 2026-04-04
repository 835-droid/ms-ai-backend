# MS-AI Backend Refactoring Plan - Phase 2

## Overview
This document outlines the detailed plan for refactoring the MS-AI backend to fully align with Clean Architecture and Hexagonal Architecture principles.

## Current Status Analysis

### What's Already Good ✅
1. **Basic separation exists**: Domain, Application, Infrastructure layers are separated
2. **Dependency Injection**: Container pattern is implemented
3. **Repository pattern**: Data access is abstracted
4. **Domain entities**: Core entities are defined in `/domain`
5. **Interface segregation**: Repository and Service interfaces exist

### What Needs Improvement 🔧
1. **Naming inconsistencies**: Mixed naming conventions across layers
2. **File organization**: Some files are too large and need splitting
3. **Unused directories**: Empty/redundant directories need cleanup
4. **Layer confusion**: Some responsibilities are mixed between layers
5. **Import paths**: Need to be updated to reflect new structure

## Target Architecture

### New Directory Structure
```
/internal
├── /domain                    # Enterprise Business Rules (Entities)
│   ├── /manga
│   │   ├── manga.go          # Main manga entity
│   │   ├── chapter.go        # Chapter entity
│   │   ├── rating.go         # Rating entity
│   │   ├── reaction.go       # Reaction entity
│   │   ├── favorite.go       # Favorite entity
│   │   ├── comment.go        # Comment entity
│   │   ├── viewing_history.go# ViewingHistory entity
│   │   ├── errors.go         # Domain-specific errors
│   │   └── repository.go     # Repository interfaces (moved from application)
│   └── /user
│       ├── user.go           # User entity
│       ├── invite_code.go    # InviteCode entity
│       ├── errors.go         # User domain errors
│       └── repository.go     # User repository interfaces
│
├── /application               # Application Business Rules (Use Cases)
│   ├── /ports                # Interfaces (Inbound & Outbound)
│   │   ├── /repository       # Repository interfaces (will be moved to domain)
│   │   └── /service          # Service interfaces
│   ├── /services             # Use Case implementations
│   │   ├── /manga
│   │   │   ├── manga_service.go
│   │   │   ├── chapter_service.go
│   │   │   ├── favorite_list_service.go
│   │   │   └── viewing_history_service.go
│   │   ├── /auth
│   │   │   └── auth_service.go
│   │   └── /admin
│   │       └── admin_service.go
│   └── /dto                  # Data Transfer Objects
│       ├── auth_dto.go
│       ├── manga_dto.go
│       └── user_dto.go
│
├── /infrastructure            # Frameworks & Drivers
│   ├── /persistence
│   │   ├── /mongo
│   │   │   ├── connection.go
│   │   │   ├── indexes.go
│   │   │   ├── monitor.go
│   │   │   └── transaction.go
│   │   └── /postgres
│   │       ├── connection.go
│   │       └── monitor.go
│   ├── /repositories         # Repository implementations
│   │   ├── /manga
│   │   │   ├── mongo_manga_repository.go
│   │   │   ├── postgres_manga_repository.go
│   │   │   ├── hybrid_manga_repository.go
│   │   │   ├── mongo_chapter_repository.go
│   │   │   ├── postgres_chapter_repository.go
│   │   │   ├── hybrid_chapter_repository.go
│   │   │   └── [other repositories...]
│   │   └── /user
│   │       ├── mongo_user_repository.go
│   │       ├── postgres_user_repository.go
│   │       └── [other repositories...]
│   └── /config               # Configuration management
│
├── /presentation              # Interface Adapters
│   ├── /http
│   │   ├── /handlers         # HTTP request handlers
│   │   │   ├── /manga
│   │   │   │   ├── manga_handler.go
│   │   │   │   ├── chapter_handler.go
│   │   │   │   ├── interaction_handler.go
│   │   │   │   └── [other handlers...]
│   │   │   ├── /auth
│   │   │   │   └── auth_handler.go
│   │   │   ├── /admin
│   │   │   │   └── admin_handler.go
│   │   │   └── /health
│   │   │       └── health_handler.go
│   │   ├── /routes           # Route definitions
│   │   │   ├── /manga
│   │   │   │   ├── manga_routes.go
│   │   │   │   └── history_routes.go
│   │   │   ├── /auth
│   │   │   │   └── auth_routes.go
│   │   │   ├── /admin
│   │   │   │   └── admin_routes.go
│   │   │   └── router.go     # Main router setup
│   │   ├── /middleware       # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   ├── ratelimit.go
│   │   │   ├── recovery.go
│   │   │   └── requestid.go
│   │   ├── /dto              # HTTP-specific DTOs
│   │   └── /response         # Response helpers
│   └── /ws                   # WebSocket handlers (if any)
│
├── /container                 # Dependency Injection
│   ├── container.go           # Main container
│   ├── types.go               # Container types
│   ├── initializers.go        # Database initializers
│   ├── repo_initializers.go   # Repository initializers
│   ├── service_initializers.go# Service initializers
│   ├── handler_initializers.go# Handler initializers
│   └── seed_initializers.go   # Seed data initializers
│
└── /shared                    # Shared utilities (if any)

/pkg                           # Public utilities (unchanged)
├── /config
├── /errors
├── /i18n
├── /jwt
├── /logger
├── /response
├── /utils
└── /validator

/cmd                           # Entry points (unchanged)
├── /server
├── /create_admin
├── /utils
└── /web                        # Static web files
```

## Migration Strategy

### Phase 1: Preparation (Non-breaking changes)
1. ✅ Create new directory structure
2. ✅ Move files to new locations
3. ✅ Update import paths
4. ✅ Verify build still works

### Phase 2: Cleanup
1. Remove unused/empty directories
2. Update documentation
3. Clean up redundant code

### Phase 3: Optimization
1. Split large files
2. Improve naming consistency
3. Add missing interfaces

## File Movement Map

### From → To

#### Domain Layer
- `internal/domain/manga/*` → `internal/domain/manga/*` (stays, but add repository.go)
- `internal/domain/user/*` → `internal/domain/user/*` (stays, but add repository.go)
- `internal/core/content/manga/manga.go` → Merge into `internal/domain/manga/manga.go`
- `internal/core/content/manga/chapter.go` → Merge into `internal/domain/manga/chapter.go`
- `internal/core/content/manga/rating.go` → Merge into `internal/domain/manga/rating.go`
- `internal/core/content/manga/reaction.go` → Merge into `internal/domain/manga/reaction.go`
- `internal/core/content/manga/favorite.go` → Merge into `internal/domain/manga/favorite.go`
- `internal/core/content/manga/comment.go` → Merge into `internal/domain/manga/comment.go`
- `internal/core/content/manga/viewing_history.go` → Merge into `internal/domain/manga/viewing_history.go`
- `internal/core/common/errors.go` → Merge into domain error files

#### Application Layer
- `internal/application/interfaces/repositories/*` → `internal/domain/manga/repository.go` & `internal/domain/user/repository.go`
- `internal/application/interfaces/services/*` → Keep as is (service interfaces)
- `internal/application/dtos/*` → `internal/application/dto/*`
- `internal/core/content/manga/*_service.go` → `internal/application/services/manga/*_service.go`
- `internal/core/auth/auth_service.go` → `internal/application/services/auth/auth_service.go`
- `internal/core/admin/*` → `internal/application/services/admin/*`
- `internal/core/user/*` → `internal/domain/user/*` (merge)

#### Infrastructure Layer
- `internal/data/infrastructure/mongo/*` → `internal/infrastructure/persistence/mongo/*`
- `internal/data/infrastructure/postgres/*` → `internal/infrastructure/persistence/postgres/*`
- `internal/data/content/manga/*` → `internal/infrastructure/repositories/manga/*`
- `internal/data/user/*` → `internal/infrastructure/repositories/user/*`
- `internal/data/admin/*` → `internal/infrastructure/repositories/admin/*`
- `internal/data/common/*` → `internal/infrastructure/common/*`

#### Presentation Layer
- `internal/api/handler/*` → `internal/presentation/http/handlers/*`
- `internal/api/router/*` → `internal/presentation/http/routes/*`
- `internal/api/middleware/*` → `internal/presentation/http/middleware/*`
- `internal/api/dto/*` → `internal/presentation/http/dto/*`

#### Container (DI)
- `internal/container/*` → `internal/container/*` (stays, but update imports)

## Import Path Updates

### Old → New
```
github.com/835-droid/ms-ai-backend/internal/domain/manga
  → github.com/835-droid/ms-ai-backend/internal/domain/manga (unchanged)

github.com/835-droid/ms-ai-backend/internal/application/interfaces/repositories
  → github.com/835-droid/ms-ai-backend/internal/domain/manga (for manga repos)
  → github.com/835-droid/ms-ai-backend/internal/domain/user (for user repos)

github.com/835-droid/ms-ai-backend/internal/core/content/manga
  → github.com/835-droid/ms-ai-backend/internal/application/services/manga

github.com/835-droid/ms-ai-backend/internal/core/auth
  → github.com/835-droid/ms-ai-backend/internal/application/services/auth

github.com/835-droid/ms-ai-backend/internal/core/admin
  → github.com/835-droid/ms-ai-backend/internal/application/services/admin

github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo
  → github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/mongo

github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres
  → github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/postgres

github.com/835-droid/ms-ai-backend/internal/data/content/manga
  → github.com/835-droid/ms-ai-backend/internal/infrastructure/repositories/manga

github.com/835-droid/ms-ai-backend/internal/data/user
  → github.com/835-droid/ms-ai-backend/internal/infrastructure/repositories/user

github.com/835-droid/ms-ai-backend/internal/api/handler
  → github.com/835-droid/ms-ai-backend/internal/presentation/http/handlers

github.com/835-droid/ms-ai-backend/internal/api/router
  → github.com/835-droid/ms-ai-backend/internal/presentation/http/routes

github.com/835-droid/ms-ai-backend/internal/api/middleware
  → github.com/835-droid/ms-ai-backend/internal/presentation/http/middleware
```

## Validation Checklist

### After Each Phase
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] No import cycles
- [ ] All tests pass
- [ ] Application starts successfully

### Final Validation
- [ ] Clean Architecture principles followed
- [ ] Separation of concerns achieved
- [ ] Testability improved
- [ ] Code is more maintainable
- [ ] Documentation updated

## Rollback Plan
If any issues arise, we can:
1. Use git to revert changes
2. Keep backup of original structure
3. Test each phase before proceeding to next

## Timeline Estimate
- Phase 1: 2-3 hours (file movement and import updates)
- Phase 2: 1 hour (cleanup and documentation)
- Phase 3: 2-3 hours (optimization and splitting)
- **Total: 5-7 hours**

## Success Criteria
1. ✅ All files moved to appropriate layers
2. ✅ All imports updated and working
3. ✅ Build succeeds without errors
4. ✅ Tests pass
5. ✅ No circular dependencies
6. ✅ Clear separation of concerns
7. ✅ Improved testability
8. ✅ Better code organization
9. ✅ Consistent naming conventions
10. ✅ Documentation updated
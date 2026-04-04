# Novel System Implementation Guide

## Overview

This document describes the complete implementation of the novel (روايات) content system in the MS-AI backend. The novel system follows the same architecture as the manga system, providing a full-featured platform for managing novels, chapters, ratings, comments, and user engagement.

## Architecture

The novel system follows Clean Architecture principles with the following layers:

### 1. Domain Layer (`internal/domain/novel/`)

Contains the core business entities and repository interfaces:

- **novel.go**: Main Novel entity with all novel properties
- **chapter.go**: NovelChapter entity for chapter content
- **rating.go**: NovelRating and ChapterRating entities
- **reaction.go**: NovelReaction entity with reaction types (upvote, funny, love, surprised, angry, sad)
- **favorite.go**: UserFavorite and FavoriteList entities
- **comment.go**: NovelComment and ChapterComment entities
- **viewing_history.go**: ReadingProgress and ViewingHistory entities
- **errors.go**: Domain-specific error definitions
- **repository.go**: Repository interfaces for data access

### 2. Application Layer (`internal/application/`)

#### DTOs (`internal/application/dtos/novel_dtos.go`)

Data Transfer Objects for API requests and responses:
- NovelRequest/NovelResponse
- NovelChapterRequest/NovelChapterResponse
- NovelRatingRequest
- NovelCommentRequest/NovelCommentResponse
- NovelFavoriteListRequest/NovelFavoriteListResponse
- NovelReadingProgressRequest/NovelReadingProgressResponse
- NovelViewingHistoryRequest/NovelViewingHistoryResponse

### 3. Core/Service Layer (`internal/core/content/novel/`)

Business logic implementation:

- **novel.go**: Type aliases re-exporting domain types
- **novel_service.go**: NovelService interface and DefaultNovelService implementation

Key service methods:
- CRUD operations for novels
- List operations (most viewed, recently updated, most followed, top rated)
- Engagement operations (reactions, ratings, favorites, comments)
- Authorization checks for update/delete operations

### 4. Data Layer (`internal/data/content/novel/`)

MongoDB repository implementations:

- **mongo_novel_repository.go**: Core novel CRUD and listing operations
- **mongo_novel_engagement_repository.go**: Reactions, ratings, favorites, comments

### 5. API Layer (`internal/api/`)

#### Handlers (`internal/api/handler/content/novel/`)

- **novel_handler.go**: Main novel CRUD and listing handlers
- **novel_interaction_handler.go**: Engagement handlers (views, reactions, favorites, comments)

#### Routes (`internal/api/router/content/novel/`)

- **novel_routes.go**: Route definitions for all novel endpoints

## API Endpoints

### Novel Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/novels` | List all novels (paginated) | No |
| GET | `/api/novels/:novelID` | Get novel by ID | No |
| POST | `/api/novels` | Create new novel | Admin |
| PUT | `/api/novels/:novelID` | Update novel | Admin |
| DELETE | `/api/novels/:novelID` | Delete novel | Admin |

### Novel Lists

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/novels/most-viewed` | Most viewed novels | No |
| GET | `/api/novels/recently-updated` | Recently updated novels | No |
| GET | `/api/novels/most-followed` | Most followed novels | No |
| GET | `/api/novels/top-rated` | Top rated novels | No |

### Engagement

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/novels/:novelID/view` | Increment view count | No (Optional Auth) |
| POST | `/api/novels/:novelID/react` | Set reaction | Yes |
| GET | `/api/novels/:novelID/my-reaction` | Get user's reaction | Yes |
| POST | `/api/novels/:novelID/rate` | Rate novel (1-10) | Yes |
| POST | `/api/novels/:novelID/favorite` | Add to favorites | Yes |
| DELETE | `/api/novels/:novelID/favorite` | Remove from favorites | Yes |
| GET | `/api/novels/:novelID/favorite` | Check if favorited | Yes |
| GET | `/api/novels/favorites` | List user's favorites | Yes |
| POST | `/api/novels/:novelID/comments` | Add comment | Yes |
| GET | `/api/novels/:novelID/comments` | List comments | No |
| DELETE | `/api/novels/:novelID/comments/:comment_id` | Delete comment | Yes |

## Database Schema

### PostgreSQL Tables

The system uses PostgreSQL for structured data:

- `novels`: Main novel data
- `novel_chapters`: Chapter content
- `novel_view_logs`: View tracking
- `novel_ratings`: User ratings
- `chapter_ratings`: Chapter ratings
- `novel_reactions`: User reactions
- `novel_favorites`: Favorite tracking
- `novel_comments`: Novel comments
- `chapter_comments`: Chapter comments
- `chapter_comment_reactions`: Comment reactions
- `novel_favorite_lists`: User favorite lists
- `novel_favorite_list_items`: List items
- `novel_reading_progress`: Reading progress
- `novel_viewing_history`: View history

### MongoDB Collections

MongoDB is used for document storage:

- `novel`: Novel documents
- `novel_view_logs`: View logs
- `novel_reactions`: Reaction documents
- `novel_ratings`: Rating documents
- `novel_favorites`: Favorite documents
- `novel_comments`: Comment documents

## Setup Instructions

### 1. Run Database Migration

```bash
# Apply PostgreSQL schema
psql -U your_user -d your_database -f MS-AI/scripts/novel_tables.sql
```

### 2. MongoDB Collections

MongoDB collections will be created automatically when data is inserted.

### 3. Register Routes

In your main server file, add the novel routes:

```go
import (
    novelhandler "github.com/835-droid/ms-ai-backend/internal/api/handler/content/novel"
    novelrouter "github.com/835-droid/ms-ai-backend/internal/api/router/content/novel"
)

// Initialize novel handler
novelHandler := novelhandler.NewNovelHandler(novelService)

// Register routes
novelrouter.SetupNovelRoutes(engine, novelHandler, cfg, userRepo)
```

### 4. Initialize Services

```go
// Create MongoDB repository
novelRepo := noveldata.NewMongoNovelRepository(mongoStore)

// Create service
novelService := novelservice.NewNovelService(novelRepo, logger)
```

## Configuration

The novel system uses the same configuration as the manga system:

- MongoDB connection string
- PostgreSQL connection string
- Rate limiting configuration
- Authentication middleware

## Rate Limiting

Different endpoints have different rate limits:

- Write operations (create, update, delete): 0.25 req/sec
- Engagement operations (react, favorite, comment): 0.15 req/sec
- Read operations: No strict limit

## Error Handling

The system uses standard HTTP status codes:

- `200 OK`: Successful request
- `201 Created`: Resource created
- `204 No Content`: Successful delete
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Testing

### Manual Testing

```bash
# List novels
curl http://localhost:8080/api/novels

# Get novel by ID
curl http://localhost:8080/api/novels/{novelID}

# Create novel (requires admin token)
curl -X POST http://localhost:8080/api/novels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Novel","description":"A test novel","tags":["test"]}'

# Rate novel
curl -X POST http://localhost:8080/api/novels/{novelID}/rate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"score":8}'
```

## Future Enhancements

1. **Chapter Management**: Full CRUD for novel chapters
2. **Reading Progress**: Track user reading progress
3. **Viewing History**: Track reading history
4. **Favorite Lists**: Custom user lists for organizing novels
5. **Search**: Full-text search for novels
6. **Recommendations**: Novel recommendations based on user preferences
7. **Notifications**: Notify users of new chapters for followed novels

## Migration from Manga System

The novel system is designed to mirror the manga system architecture, making it easy to:

1. Reuse existing middleware and utilities
2. Share authentication and authorization logic
3. Use the same database patterns
4. Maintain consistent API design

## Support

For issues or questions:
1. Check the API documentation
2. Review the manga system implementation as reference
3. Consult the error logs for debugging
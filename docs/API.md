# API Documentation

See endpoints and request/response shapes used by clients.

## Authentication

POST /api/auth/signup
- Request: {"username":"...","password":"...","invite_code":"..."}
- Response: 201 {"access_token":"...","refresh_token":"..."}

POST /api/auth/login
- Request: {"username":"...","password":"..."}
- Response: 200 {"access_token":"...","refresh_token":"..."}

POST /api/auth/refresh
- Request: {"refresh_token":"..."}
- Response: 200 {"access_token":"..."}

POST /api/auth/logout
- Auth: Bearer
- Request: {"refresh_token":"..."}
- Response: 200 {"message":"logged out"}

## Admin

POST /api/admin/invite
- Auth: Bearer (admin)
- Request: {"length":12, "expires_in":"72h"}
- Response: 201 {"code":"ABC123...","expires_at":"..."}

GET /api/admin/invites
- Auth: Bearer (admin)
- Query: page, limit
- Response: 200 {"data":[...], "meta":{}}

DELETE /api/admin/invite/:id
- Auth: Bearer (admin)
- Response: 200 {"message":"deleted"}

## Manga

**All manga routes require Bearer token authentication.**

### Manga Management

GET /api/mangas
- **Purpose**: Get paginated list of all manga
- Auth: Bearer
- Query: page (default: 1), limit (default: 20, max: 100)
- Response: 200 {"total": 100, "total_pages": 5, "current_page": 1, "per_page": 20, "items": [...]}

POST /api/mangas
- **Purpose**: Create a new manga
- Auth: Bearer
- Request: {"title":"...", "description":"...", "cover_image":"...", "tags":["tag1", "tag2"]}
- Response: 201 {"id":"...", "title":"...", "description":"...", "cover_image":"...", "tags":[...], "author_id":"...", "is_published":false, "created_at":"...", "updated_at":"..."}

GET /api/mangas/{manga_id}
- **Purpose**: Get details of a specific manga
- Auth: Bearer
- Response: 200 {"id":"...", "title":"...", "description":"...", "cover_image":"...", "tags":[...], "author_id":"...", "is_published":false, "created_at":"...", "updated_at":"..."}

PUT /api/mangas/{manga_id}
- **Purpose**: Update an existing manga (owner or admin only)
- Auth: Bearer
- Request: {"title":"...", "description":"...", "cover_image":"...", "tags":["tag1", "tag2"]}
- Response: 200 {"id":"...", "title":"...", "description":"...", "cover_image":"...", "tags":[...], "author_id":"...", "is_published":false, "created_at":"...", "updated_at":"..."}

DELETE /api/mangas/{manga_id}
- **Purpose**: Delete a manga (owner or admin only)
- Auth: Bearer
- Response: 204

### Chapter Management

GET /api/mangas/{manga_id}/chapters
- **Purpose**: Get all chapters for a specific manga
- Auth: Bearer
- Query: page (default: 1), limit (default: 20)
- Response: 200 {"total": 10, "page": 1, "limit": 20, "chapters": [{"id":"...", "manga_id":"...", "title":"...", "number":1, "created_at":"...", "updated_at":"..."}]}

POST /api/mangas/{manga_id}/chapters
- **Purpose**: Create a new chapter for a manga (owner or admin only)
- Auth: Bearer
- Request: {"title":"...", "number":1}
- Response: 201 {"id":"...", "manga_id":"...", "title":"...", "number":1, "created_at":"...", "updated_at":"..."}

GET /api/mangas/{manga_id}/chapters/{chapter_number}
- **Purpose**: Get details of a specific chapter
- Auth: Bearer
- Response: 200 {"id":"...", "manga_id":"...", "title":"...", "number":1, "pages":[], "created_at":"...", "updated_at":"..."}

PUT /api/mangas/{manga_id}/chapters/{chapter_number}
- **Purpose**: Update a chapter (owner or admin only)
- Auth: Bearer
- Request: {"title":"...", "number":1}
- Response: 200 {"id":"...", "manga_id":"...", "title":"...", "number":1, "pages":[], "created_at":"...", "updated_at":"..."}

DELETE /api/mangas/{manga_id}/chapters/{chapter_number}
- **Purpose**: Delete a chapter (owner or admin only)
- Auth: Bearer
- Response: 204

## Health Check

GET /api/health
- **Purpose**: Check if the service is running
- Response: 200 {"status":"ok", "timestamp":"..."}

## Error Responses

All endpoints may return the following error responses:

- 400 Bad Request: Invalid request data
- 401 Unauthorized: Missing or invalid authentication
- 403 Forbidden: Insufficient permissions
- 404 Not Found: Resource not found
- 500 Internal Server Error: Server error

Error response format:
```json
{
  "error": "error message",
  "code": "ERROR_CODE"
}
```
- Message Format (Example from `main.go`): `hello from ws client at 2025-11-09T11:16:55+03:00`
- Response: 101 Switching Protocols on success.

GET /api/chat/history
- **Purpose**: Retrieves a limited number of the most recent chat messages.
- Auth: Bearer
- Query: `limit` (optional, default is 50, as seen in `chat_handler.go`)
- Response: 200 [{"_id":"...","user_id":"...","username":"...","content":"...","created_at":"..."}]

GET /api/chat/users
- **Purpose**: Retrieves a list of currently active (connected) users.
- Auth: Bearer
- Response: 200 [{"id":"...","username":"...","state":"..."}]

GET /api/chat/messages/since
- **Purpose**: Retrieves chat messages sent since a specific timestamp.
- Auth: Bearer
- Query: `since` (required, RFC3339 format, e.g., `2023-01-01T15:04:05Z`)
- Response: 200 [{"_id":"...","user_id":"...","username":"...","content":"...","created_at":"..."}]

## Health

GET /livez
- Response: 200 {"status":"alive"}

GET /readyz
- Response: 200|503 depending on DB readiness
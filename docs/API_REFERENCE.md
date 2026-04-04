# MS-AI Backend API Reference

## Overview

This document provides a comprehensive reference for the MS-AI Backend API. The API follows RESTful principles and uses JSON for request/response payloads.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require authentication using JWT tokens. Include the token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

## Response Format

All responses follow a consistent format:

### Success Response
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": { /* response data */ }
}
```

### Error Response
```json
{
  "status": "error",
  "message": "Error description",
  "errors": [ /* validation errors */ ]
}
```

### Paginated Response
```json
{
  "status": "success",
  "data": {
    "items": [ /* array of items */ ],
    "total": 100,
    "page": 1,
    "per_page": 20,
    "total_pages": 5
  }
}
```

## API Endpoints

### Authentication

#### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "username": "username",
  "invite_code": "optional_code"
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "user": { /* user object */ },
    "token": "jwt_token"
  }
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "access_token": "jwt_token",
    "refresh_token": "refresh_token",
    "user": { /* user object */ }
  }
}
```

#### Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "valid_refresh_token"
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "access_token": "new_jwt_token"
  }
}
```

#### Logout
```http
POST /auth/logout
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Delete Account
```http
DELETE /auth/delete-account
Authorization: Bearer <token>
```

**Response:** `200 OK`

---

### Manga

#### List All Manga
```http
GET /manga?page=1&limit=20
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "items": [ /* manga array */ ],
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

#### Get Manga by ID
```http
GET /manga/{mangaID}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": { /* manga object */ }
}
```

#### Create Manga
```http
POST /manga
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Manga Title",
  "description": "Description",
  "tags": ["tag1", "tag2"],
  "cover_image": "url_to_cover"
}
```

**Response:** `201 Created`

#### Update Manga
```http
PUT /manga/{mangaID}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "tags": ["new_tag"],
  "cover_image": "new_url"
}
```

**Response:** `200 OK`

#### Delete Manga
```http
DELETE /manga/{mangaID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Rate Manga
```http
POST /manga/{mangaID}/rate
Authorization: Bearer <token>
Content-Type: application/json

{
  "score": 8.5
}
```

**Response:** `200 OK`

#### List Most Viewed
```http
GET /api/mangas/most-viewed?period=day&limit=10
```

**Response:** `200 OK`

#### List Recently Updated
```http
GET /api/mangas/recently-updated?limit=10
```

**Response:** `200 OK`

#### List Most Followed
```http
GET /api/mangas/most-followed?limit=10
```

**Response:** `200 OK`

#### List Top Rated
```http
GET /api/mangas/top-rated?limit=10
```

**Response:** `200 OK`

---

### Manga Chapters

#### List Chapters
```http
GET /manga/{mangaID}/chapters?page=1&limit=50
```

**Response:** `200 OK`

#### Get Chapter
```http
GET /manga/{mangaID}/chapters/{chapterID}
```

**Response:** `200 OK`

#### Create Chapter
```http
POST /manga/{mangaID}/chapters
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Chapter 1",
  "chapter_number": 1,
  "volume": 1,
  "pages": 20
}
```

**Response:** `201 Created`

#### Update Chapter
```http
PUT /manga/{mangaID}/chapters/{chapterID}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Chapter Title",
  "pages": 25
}
```

**Response:** `200 OK`

#### Delete Chapter
```http
DELETE /manga/{mangaID}/chapters/{chapterID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Rate Chapter
```http
POST /manga/{mangaID}/chapters/{chapterID}/rate
Authorization: Bearer <token>
Content-Type: application/json

{
  "score": 9.0
}
```

**Response:** `200 OK`

---

### Manga Interactions

#### Add to Favorites
```http
POST /manga/{mangaID}/favorites
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Remove from Favorites
```http
DELETE /manga/{mangaID}/favorites
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Check Favorite Status
```http
GET /manga/{mangaID}/favorites
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "is_favorite": true
  }
}
```

#### List User Favorites
```http
GET /users/me/favorites?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Set Reaction
```http
POST /manga/{mangaID}/reactions
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "upvote" // "upvote", "funny", "love", "surprised", "angry", "sad"
}
```

**Response:** `200 OK`

#### Get User Reaction
```http
GET /manga/{mangaID}/reactions
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### List Liked Manga
```http
GET /users/me/liked?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

---

### Favorite Lists

#### Create List
```http
POST /favorites/lists
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Favorite Manga",
  "description": "Collection of my favorites",
  "is_public": false
}
```

**Response:** `201 Created`

#### Get List
```http
GET /favorites/lists/{listID}
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### List User Lists
```http
GET /favorites/lists?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Update List
```http
PUT /favorites/lists/{listID}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated List Name",
  "description": "Updated description",
  "is_public": true
}
```

**Response:** `200 OK`

#### Delete List
```http
DELETE /favorites/lists/{listID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Add Manga to List
```http
POST /favorites/lists/{listID}/manga
Authorization: Bearer <token>
Content-Type: application/json

{
  "manga_id": "manga_object_id"
}
```

**Response:** `200 OK`

#### Remove Manga from List
```http
DELETE /favorites/lists/{listID}/manga/{mangaID}
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### List Manga in List
```http
GET /favorites/lists/{listID}/manga?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

---

### Reading Progress

#### Save Progress
```http
POST /reading/progress
Authorization: Bearer <token>
Content-Type: application/json

{
  "manga_id": "manga_object_id",
  "chapter_id": "chapter_object_id",
  "page": 5,
  "completed": false
}
```

**Response:** `200 OK`

#### Get Progress
```http
GET /reading/progress/{mangaID}
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Delete Progress
```http
DELETE /reading/progress/{mangaID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

---

### Viewing History

#### Track View
```http
POST /history
Authorization: Bearer <token>
Content-Type: application/json

{
  "manga_id": "manga_object_id",
  "chapter_id": "chapter_object_id",
  "page": 1
}
```

**Response:** `200 OK`

#### Get User History
```http
GET /history?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Get Recent History
```http
GET /history/recent?limit=10
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Delete History Item
```http
DELETE /history/{historyID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Delete History by Manga
```http
DELETE /history/manga/{mangaID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Clean Old History
```http
POST /history/clean
Authorization: Bearer <token>
Content-Type: application/json

{
  "days": 30
}
```

**Response:** `200 OK`

---

### Comments

#### Add Manga Comment
```http
POST /manga/{mangaID}/comments
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Great manga!",
  "parent_id": null // optional, for replies
}
```

**Response:** `201 Created`

#### List Manga Comments
```http
GET /manga/{mangaID}/comments?page=1&limit=20&sort=desc
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Delete Manga Comment
```http
DELETE /manga/comments/{commentID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Add Chapter Comment
```http
POST /manga/{mangaID}/chapters/{chapterID}/comments
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Great chapter!",
  "parent_id": null
}
```

**Response:** `201 Created`

#### List Chapter Comments
```http
GET /manga/{mangaID}/chapters/{chapterID}/comments?page=1&limit=20&sort=desc
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Delete Chapter Comment
```http
DELETE /chapters/comments/{commentID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

---

### Admin

#### List All Users
```http
GET /admin/users?page=1&limit=20
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Get User Details
```http
GET /admin/users/{userID}
Authorization: Bearer <token>
```

**Response:** `200 OK`

#### Update User Role
```http
PUT /admin/users/{userID}/role
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "admin" // "admin", "moderator", "user"
}
```

**Response:** `200 OK`

#### Delete User
```http
DELETE /admin/users/{userID}
Authorization: Bearer <token>
```

**Response:** `204 No Content`

#### Get System Stats
```http
GET /admin/stats
Authorization: Bearer <token>
```

**Response:** `200 OK`

---

### Health Check

#### Health Status
```http
GET /health
```

**Response:** `200 OK`
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-04-04T12:00:00Z"
}
```

#### Liveness Probe
```http
GET /livez
```

**Response:** `200 OK`

#### Readiness Probe
```http
GET /readyz
```

**Response:** `200 OK`

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

## Rate Limiting

API requests are rate limited to ensure fair usage:

- **Standard endpoints:** 100 requests per minute
- **Authentication endpoints:** 20 requests per minute
- **Admin endpoints:** 50 requests per minute

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1649078400
```

## Pagination

All list endpoints support pagination with the following parameters:

- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

## Sorting

Some endpoints support sorting with the `sort` parameter:
- `sort=asc` - Ascending order
- `sort=desc` - Descending order (default)

## Versioning

The API is versioned through the URL path:
- Current version: `/api/v1`
- Example: `https://api.ms-ai.com/api/v1/manga`

## Webhooks

(To be implemented)

## SDKs & Libraries

(To be implemented)

## Support

For API support, please contact:
- Email: support@ms-ai.com
- Documentation: https://docs.ms-ai.com
- Status: https://status.ms-ai.com
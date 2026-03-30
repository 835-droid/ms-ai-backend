# Database Documentation

This document describes the MongoDB collections, indexes, and data models for the MS-AI Manga Platform.

## Database Overview

The application uses MongoDB as its primary database with the following key collections:

- **users**: User accounts and authentication data
- **invite_codes**: Invitation system for user registration
- **mangas**: Manga series metadata
- **manga_chapters**: Individual manga chapters

## Collections

### Users Collection

Stores user account information and authentication data.

```javascript
{
  _id: ObjectId,
  username: String,     // Unique username (indexed)
  password_hash: String, // Bcrypt hashed password
  roles: [String],      // User roles (e.g., ["user"], ["admin"])
  created_at: Date,
  updated_at: Date,
  last_login: Date
}
```

**Indexes:**
- `username`: Unique index for login lookups
- Compound index on `roles` for admin queries

### Invite Codes Collection

Manages invitation codes for new user registration.

```javascript
{
  _id: ObjectId,
  code: String,         // Unique invitation code (indexed)
  created_by: ObjectId, // User who created the invite
  expires_at: Date,     // Expiration timestamp
  used_by: ObjectId,    // User who used the invite (null if unused)
  used_at: Date,        // Usage timestamp (null if unused)
  created_at: Date
}
```

**Indexes:**
- `code`: Unique index for code validation
- `expires_at`: TTL index for automatic cleanup
- `used_by`: Index for usage tracking

### Mangas Collection

Stores manga series metadata and information.

```javascript
{
  _id: ObjectId,
  title: String,           // Manga title (max 200 chars)
  slug: String,            // URL-friendly slug (indexed)
  description: String,     // Manga description (max 2000 chars)
  author_id: ObjectId,     // Reference to author user (indexed)
  tags: [String],          // Array of tags for categorization
  cover_image: String,     // URL to cover image
  is_published: Boolean,   // Publication status (default: false)
  published_at: Date,      // Publication timestamp
  created_at: Date,
  updated_at: Date
}
```

**Indexes:**
- `slug`: Unique index for URL routing
- `author_id`: Index for author's manga queries
- `is_published`: Index for published content filtering
- `tags`: Index for tag-based searches
- Text index on `title` and `description` for search functionality

### Manga Chapters Collection

Stores individual manga chapters with page references.

```javascript
{
  _id: ObjectId,
  manga_id: ObjectId,      // Reference to parent manga (indexed)
  title: String,           // Chapter title (max 200 chars)
  number: Number,          // Chapter number (indexed)
  pages: [String],         // Array of page image URLs
  created_at: Date,
  updated_at: Date
}
```

**Indexes:**
- `manga_id`: Index for manga-specific chapter queries
- `number`: Index for chapter ordering
- Compound unique index on `manga_id + number` to prevent duplicate chapter numbers
- Compound index on `manga_id + number` for efficient chapter listing

## Indexing Strategy

### Performance Indexes

1. **User Authentication**
   - Unique index on `users.username` for fast login lookups

2. **Content Discovery**
   - Text index on `mangas.title` and `mangas.description` for search
   - Index on `mangas.tags` for tag filtering
   - Index on `mangas.is_published` for published content queries

3. **Author Management**
   - Index on `mangas.author_id` for author-specific queries
   - Index on `invite_codes.created_by` for admin invite tracking

4. **Chapter Navigation**
   - Compound index on `manga_chapters.manga_id + manga_chapters.number` for ordered chapter lists
   - Unique compound index to prevent duplicate chapter numbers per manga

### Maintenance Indexes

- TTL index on `invite_codes.expires_at` for automatic cleanup of expired invites
- Standard indexes on `created_at` and `updated_at` fields for sorting and filtering

## Data Relationships

```
users (1) ──── (many) invite_codes.created_by
users (1) ──── (many) invite_codes.used_by
users (1) ──── (many) mangas.author_id
mangas (1) ─── (many) manga_chapters.manga_id
```

## Transactions

The application uses MongoDB transactions for atomic operations:

1. **User Registration**: Atomic invite code usage + user creation
2. **Content Publishing**: Atomic manga publication updates
3. **Bulk Operations**: Chapter batch operations

## Data Validation

### Constraints

- **Username**: 3-50 characters, alphanumeric + underscores
- **Manga Title**: 1-200 characters
- **Manga Description**: 0-2000 characters
- **Chapter Title**: 1-200 characters
- **Tags**: Maximum 10 tags per manga, 50 chars each
- **Chapter Numbers**: Positive integers, unique per manga

### Business Rules

- Users can only modify their own manga (unless admin)
- Chapter numbers must be unique within a manga
- Invite codes can only be used once
- Published content cannot be deleted (soft delete pattern)

## Monitoring

### Collection Sizes

Monitor collection growth:
```javascript
db.mangas.stats()
db.manga_chapters.stats()
db.users.stats()
```

### Query Performance

Key queries to monitor:
- Manga listing with pagination
- Chapter listing for a specific manga
- User authentication lookups
- Search queries on manga content

### Index Usage

Check index effectiveness:
```javascript
db.mangas.aggregate([{$indexStats: {}}])
```

## Backup Strategy

### Collections to Backup

- `users`: Critical user data
- `mangas`: Content metadata
- `manga_chapters`: Chapter data
- `invite_codes`: Active invitation codes

### Backup Frequency

- **Full Backup**: Daily
- **Incremental**: Hourly for active collections
- **Point-in-Time**: For disaster recovery

## Migration Notes

When upgrading the database schema:

1. Create new indexes before dropping old ones
2. Use background index creation for production
3. Test migrations on staging environment first
4. Have rollback plan for failed migrations
5. Update application code after schema changes
- Updates: Point updates by _id
- Deletes: Point deletes by _id

Best practices:
- Always use the compound (anime_id, episode) index for uniqueness checks
- Perform uniqueness checks in transactions
- Use skip/limit with proper sorting for pagination
- Consider adding field-level validation before writes

## Indexes created by the application

The application ensures the following indexes at startup:

- users: username (unique), email (unique, sparse), is_active+created_at
- invite_codes: code (unique), is_used+expires_at
- novels: slug (unique), author_id, status+created_at
- novel_chapters: compound (novel_id, number) unique; (novel_id, created_at)
- anime_episodes: compound (anime_id, episode) unique; (anime_id, created_at)
- manga_chapters: compound (manga_id, number) unique; (manga_id, created_at)
- chat_messages: created_at (desc); (user_id, created_at)

## Transactions

Use transactions for operations that must be atomic across multiple collections or when uniqueness checks and writes must be done together (e.g. create chapter with uniqueness check).

Note: The codebase prefers checking existence via matched counts (MatchedCount) and performing uniqueness checks inside a transaction where possible to avoid race conditions (for example: creating novels, chapters, anime episodes).

## Connection pooling and timeouts

The application reads pool sizes and timeouts from environment variables. Tune `MONGO_MAX_POOL_SIZE` and `MONGO_MIN_POOL_SIZE` based on traffic.

## Monitoring

The internal MongoMonitor collects:

- TotalOperations
- FailedOperations
- SlowQueries
- ActiveConnections
- PoolSize
- AverageLatencyMs
- MaxLatencyMs
- SuccessRate

There is a `/metrics` health endpoint that exposes mongo stats and a `GetDetailedHealth` helper for richer diagnostics.

## Collection and Index Migrations

### Collection Rename: novel_chapters

The chapters collection has been renamed to `novel_chapters`. For existing clusters with data in old collections:

1. Check if old collection exists:
```js
db.getCollection('novel_shapters').count()  // or other old names: 'chapters', 'novel_shards'
```

2. If data exists, run this migration in a maintenance window:
```js
// Copy documents to the canonical collection
db.novel_shapters.find().forEach(function(doc) {
    db.novel_chapters.insert(doc);
});

// Verify document counts match
var oldCount = db.novel_shapters.count();
var newCount = db.novel_chapters.count();
if (oldCount === newCount) {
    print("Migration successful. Old collection:", oldCount, "New collection:", newCount);
    // Optional: rename old collection as backup
    db.novel_shapters.renameCollection("novel_shapters_backup");
} else {
    print("Warning: Count mismatch. Old:", oldCount, "New:", newCount);
}
```

3. After verifying the migration:
```js
db.novel_shapters_backup.drop()  // Drop the backup if confident
```

Note: The application creates all required indexes on startup in `novel_chapters`.

### General Index Guidelines

To add new indexes in production:
1. Add the index creation to `internal/data/mongo/mongodb.go` EnsureIndexes
2. Deploy and let the app create the index in the background
3. For large collections consider creating indexes manually with proper maintenance windows


## Notes about recent refactors

- Episodes were moved out of the main `anime` document into a dedicated `anime_episodes` collection to improve query patterns and reduce document growth.
- Repository update checks now use `MatchedCount` instead of `ModifiedCount` to detect whether a document existed (Modify count is 0 when the new document equals the old one).
- Create operations that require uniqueness (novel slug, chapter number, anime episode) use transactions and CountDocuments inside the transaction to ensure atomic behavior.

Note: Modern MongoDB (4.2+) creates indexes in the background by default for most cases. The application will ensure indexes at startup using the driver, and will retry EnsureIndexes a few times before failing. For very large collections consider creating indexes offline or in a maintenance window.

````


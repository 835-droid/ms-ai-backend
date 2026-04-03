-- Sequences table for counters
CREATE TABLE IF NOT EXISTS sequences (
    name TEXT PRIMARY KEY,
    value INTEGER NOT NULL DEFAULT 0
);

-- Invite codes table
CREATE TABLE IF NOT EXISTS invite_codes (
    id TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    is_used BOOLEAN DEFAULT false,
    used_by TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS mangas (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    author_id TEXT,
    tags JSONB DEFAULT '[]',
    cover_image TEXT,
    is_published BOOLEAN DEFAULT false,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    views_count BIGINT DEFAULT 0,
    likes_count BIGINT DEFAULT 0,
    favorites_count BIGINT DEFAULT 0,
    rating_sum DOUBLE PRECISION DEFAULT 0,
    rating_count BIGINT DEFAULT 0,
    average_rating DOUBLE PRECISION DEFAULT 0
);

CREATE TABLE IF NOT EXISTS manga_likes (
    manga_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    reaction_type VARCHAR(50) DEFAULT 'upvote',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (manga_id, user_id)
);

CREATE TABLE IF NOT EXISTS manga_ratings (
    manga_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (manga_id, user_id)
);

-- Manga view logs table for time-based tracking
CREATE TABLE IF NOT EXISTS manga_view_logs (
    id SERIAL PRIMARY KEY,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for manga_view_logs
CREATE INDEX IF NOT EXISTS idx_manga_view_logs_manga_id ON manga_view_logs(manga_id);
CREATE INDEX IF NOT EXISTS idx_manga_view_logs_viewed_at ON manga_view_logs(viewed_at);
CREATE INDEX IF NOT EXISTS idx_manga_view_logs_manga_id_viewed_at ON manga_view_logs(manga_id, viewed_at);

CREATE TABLE IF NOT EXISTS manga_chapters (
    id TEXT PRIMARY KEY,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    pages JSONB DEFAULT '[]',
    number INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    views_count BIGINT DEFAULT 0,
    rating_sum DOUBLE PRECISION DEFAULT 0,
    rating_count BIGINT DEFAULT 0,
    average_rating DOUBLE PRECISION DEFAULT 0,
    UNIQUE (manga_id, number)
);

-- جدول المفضلة المستقل
CREATE TABLE IF NOT EXISTS user_favorites (
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (manga_id, user_id)
);

-- تقييم الفصول (1-10)
CREATE TABLE IF NOT EXISTS chapter_ratings (
    chapter_id TEXT NOT NULL REFERENCES manga_chapters(id) ON DELETE CASCADE,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL CHECK (score >= 1 AND score <= 10),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (chapter_id, user_id)
);

-- تعليقات المانجا
CREATE TABLE IF NOT EXISTS manga_comments (
    id TEXT PRIMARY KEY,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- تعليقات الفصول
CREATE TABLE IF NOT EXISTS chapter_comments (
    id TEXT PRIMARY KEY,
    chapter_id TEXT NOT NULL REFERENCES manga_chapters(id) ON DELETE CASCADE,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_chapter_ratings_chapter_id ON chapter_ratings(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_ratings_user_id ON chapter_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_manga_comments_manga_id ON manga_comments(manga_id);
CREATE INDEX IF NOT EXISTS idx_manga_comments_user_id ON manga_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_manga_comments_created_at ON manga_comments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chapter_comments_chapter_id ON chapter_comments(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_comments_manga_id ON chapter_comments(manga_id);
CREATE INDEX IF NOT EXISTS idx_chapter_comments_created_at ON chapter_comments(created_at DESC);

CREATE TABLE IF NOT EXISTS chapter_views (
    chapter_id TEXT NOT NULL REFERENCES manga_chapters(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chapter_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_chapter_views_chapter_id ON chapter_views(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_views_user_id ON chapter_views(user_id);
CREATE INDEX IF NOT EXISTS idx_chapter_views_manga_id ON chapter_views(manga_id);

-- Forward-compatible migrations for existing databases
ALTER TABLE IF EXISTS manga_likes ADD COLUMN IF NOT EXISTS reaction_type VARCHAR(50) DEFAULT 'upvote';
ALTER TABLE IF EXISTS manga_likes ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE IF EXISTS mangas ADD COLUMN IF NOT EXISTS reactions_count JSONB DEFAULT '{}';
ALTER TABLE IF EXISTS mangas ADD COLUMN IF NOT EXISTS favorites_count BIGINT DEFAULT 0;
ALTER TABLE IF EXISTS manga_chapters ADD COLUMN IF NOT EXISTS views_count BIGINT DEFAULT 0;
ALTER TABLE IF EXISTS manga_chapters ADD COLUMN IF NOT EXISTS rating_sum DOUBLE PRECISION DEFAULT 0;
ALTER TABLE IF EXISTS manga_chapters ADD COLUMN IF NOT EXISTS rating_count BIGINT DEFAULT 0;
ALTER TABLE IF EXISTS manga_chapters ADD COLUMN IF NOT EXISTS average_rating DOUBLE PRECISION DEFAULT 0;
ALTER TABLE IF EXISTS chapter_ratings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

-- Remove duplicate columns from user_details
ALTER TABLE IF EXISTS user_details DROP COLUMN IF EXISTS roles;
ALTER TABLE IF EXISTS user_details DROP COLUMN IF EXISTS is_active;
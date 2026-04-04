-- Novel Tables Migration Script
-- This script creates all necessary tables for the novel content system

-- Novel main table
CREATE TABLE IF NOT EXISTS novels (
    id VARCHAR(24) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    author_id VARCHAR(24) NOT NULL,
    tags TEXT[], -- Array of tags
    cover_image VARCHAR(500),
    is_published BOOLEAN DEFAULT FALSE,
    views_count BIGINT DEFAULT 0,
    favorites_count BIGINT DEFAULT 0,
    rating_sum FLOAT DEFAULT 0,
    rating_count BIGINT DEFAULT 0,
    average_rating FLOAT DEFAULT 0,
    reactions_count JSONB DEFAULT '{}', -- Stores count for each reaction type
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Novel chapters table
CREATE TABLE IF NOT EXISTS novel_chapters (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    number INTEGER NOT NULL,
    content TEXT NOT NULL,
    word_count INTEGER DEFAULT 0,
    views_count BIGINT DEFAULT 0,
    rating_sum FLOAT DEFAULT 0,
    rating_count BIGINT DEFAULT 0,
    average_rating FLOAT DEFAULT 0,
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(novel_id, number)
);

-- Novel view logs (for tracking views over time periods)
CREATE TABLE IF NOT EXISTS novel_view_logs (
    id SERIAL PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Novel ratings
CREATE TABLE IF NOT EXISTS novel_ratings (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    score FLOAT NOT NULL CHECK (score >= 1 AND score <= 10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(novel_id, user_id)
);

-- Chapter ratings
CREATE TABLE IF NOT EXISTS chapter_ratings (
    id VARCHAR(24) PRIMARY KEY,
    chapter_id VARCHAR(24) NOT NULL,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    score FLOAT NOT NULL CHECK (score >= 1 AND score <= 10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chapter_id, user_id)
);

-- Novel reactions (like, love, funny, etc.)
CREATE TABLE IF NOT EXISTS novel_reactions (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('upvote', 'funny', 'love', 'surprised', 'angry', 'sad')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(novel_id, user_id)
);

-- Novel favorites
CREATE TABLE IF NOT EXISTS novel_favorites (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(novel_id, user_id)
);

-- Novel comments
CREATE TABLE IF NOT EXISTS novel_comments (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    username VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chapter comments
CREATE TABLE IF NOT EXISTS chapter_comments (
    id VARCHAR(24) PRIMARY KEY,
    chapter_id VARCHAR(24) NOT NULL,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    username VARCHAR(100) NOT NULL,
    user_avatar VARCHAR(500),
    content TEXT NOT NULL,
    like_count BIGINT DEFAULT 0,
    dislike_count BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chapter comment reactions
CREATE TABLE IF NOT EXISTS chapter_comment_reactions (
    id VARCHAR(24) PRIMARY KEY,
    comment_id VARCHAR(24) NOT NULL,
    chapter_id VARCHAR(24) NOT NULL,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('like', 'dislike')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(comment_id, user_id)
);

-- Novel favorite lists
CREATE TABLE IF NOT EXISTS novel_favorite_lists (
    id VARCHAR(24) PRIMARY KEY,
    user_id VARCHAR(24) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

-- Novel favorite list items
CREATE TABLE IF NOT EXISTS novel_favorite_list_items (
    id VARCHAR(24) PRIMARY KEY,
    list_id VARCHAR(24) NOT NULL REFERENCES novel_favorite_lists(id) ON DELETE CASCADE,
    novel_id VARCHAR(24) NOT NULL,
    notes TEXT,
    sort_order INTEGER DEFAULT 0,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(list_id, novel_id)
);

-- Reading progress
CREATE TABLE IF NOT EXISTS novel_reading_progress (
    id VARCHAR(24) PRIMARY KEY,
    novel_id VARCHAR(24) NOT NULL,
    user_id VARCHAR(24) NOT NULL,
    last_read_chapter VARCHAR(24),
    last_read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(novel_id, user_id)
);

-- Viewing history
CREATE TABLE IF NOT EXISTS novel_viewing_history (
    id VARCHAR(24) PRIMARY KEY,
    user_id VARCHAR(24) NOT NULL,
    novel_id VARCHAR(24) NOT NULL,
    chapter_id VARCHAR(24),
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_novels_slug ON novels(slug);
CREATE INDEX IF NOT EXISTS idx_novels_author_id ON novels(author_id);
CREATE INDEX IF NOT EXISTS idx_novels_is_published ON novels(is_published);
CREATE INDEX IF NOT EXISTS idx_novels_views_count ON novels(views_count DESC);
CREATE INDEX IF NOT EXISTS idx_novels_favorites_count ON novels(favorites_count DESC);
CREATE INDEX IF NOT EXISTS idx_novels_average_rating ON novels(average_rating DESC, rating_count DESC);
CREATE INDEX IF NOT EXISTS idx_novels_created_at ON novels(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_novels_updated_at ON novels(updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_novel_chapters_novel_id ON novel_chapters(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_chapters_number ON novel_chapters(number);
CREATE INDEX IF NOT EXISTS idx_novel_chapters_novel_number ON novel_chapters(novel_id, number);

CREATE INDEX IF NOT EXISTS idx_novel_view_logs_novel_id ON novel_view_logs(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_view_logs_viewed_at ON novel_view_logs(viewed_at);

CREATE INDEX IF NOT EXISTS idx_novel_ratings_novel_id ON novel_ratings(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_ratings_user_id ON novel_ratings(user_id);

CREATE INDEX IF NOT EXISTS idx_chapter_ratings_chapter_id ON chapter_ratings(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_ratings_user_id ON chapter_ratings(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_reactions_novel_id ON novel_reactions(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_reactions_user_id ON novel_reactions(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_favorites_novel_id ON novel_favorites(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_favorites_user_id ON novel_favorites(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_comments_novel_id ON novel_comments(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_comments_created_at ON novel_comments(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_chapter_comments_chapter_id ON chapter_comments(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_comments_created_at ON chapter_comments(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_chapter_comment_reactions_comment_id ON chapter_comment_reactions(comment_id);
CREATE INDEX IF NOT EXISTS idx_chapter_comment_reactions_user_id ON chapter_comment_reactions(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_favorite_lists_user_id ON novel_favorite_lists(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_favorite_list_items_list_id ON novel_favorite_list_items(list_id);
CREATE INDEX IF NOT EXISTS idx_novel_favorite_list_items_novel_id ON novel_favorite_list_items(novel_id);

CREATE INDEX IF NOT EXISTS idx_novel_reading_progress_novel_id ON novel_reading_progress(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_reading_progress_user_id ON novel_reading_progress(user_id);

CREATE INDEX IF NOT EXISTS idx_novel_viewing_history_user_id ON novel_viewing_history(user_id);
CREATE INDEX IF NOT EXISTS idx_novel_viewing_history_novel_id ON novel_viewing_history(novel_id);
CREATE INDEX IF NOT EXISTS idx_novel_viewing_history_viewed_at ON novel_viewing_history(viewed_at DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_novels_updated_at BEFORE UPDATE ON novels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_chapters_updated_at BEFORE UPDATE ON novel_chapters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_comments_updated_at BEFORE UPDATE ON novel_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chapter_comments_updated_at BEFORE UPDATE ON chapter_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_favorite_lists_updated_at BEFORE UPDATE ON novel_favorite_lists
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_reading_progress_updated_at BEFORE UPDATE ON novel_reading_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
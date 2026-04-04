package dtos

// NovelRequest represents a novel creation/update request.
type NovelRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"required,max=2000"`
	Tags        []string `json:"tags" validate:"dive,max=50"`
	CoverImage  string   `json:"cover_image,omitempty"`
}

// NovelResponse represents a novel response.
type NovelResponse struct {
	ID             string           `json:"id"`
	Title          string           `json:"title"`
	Slug           string           `json:"slug"`
	Description    string           `json:"description"`
	AuthorID       string           `json:"author_id"`
	Tags           []string         `json:"tags"`
	CoverImage     string           `json:"cover_image"`
	IsPublished    bool             `json:"is_published"`
	PublishedAt    string           `json:"published_at,omitempty"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	ViewsCount     int64            `json:"views_count"`
	LikesCount     int64            `json:"likes_count"`
	FavoritesCount int64            `json:"favorites_count"`
	ReactionsCount map[string]int64 `json:"reactions_count"`
	RatingSum      float64          `json:"rating_sum"`
	RatingCount    int64            `json:"rating_count"`
	AverageRating  float64          `json:"average_rating"`
}

// ChapterRequest represents a chapter creation/update request.
type NovelChapterRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=256"`
	Number  int    `json:"number" validate:"required,min=1"`
	Content string `json:"content" validate:"required,min=1"`
	NovelID string `json:"novel_id" validate:"required"`
}

// ChapterResponse represents a chapter response.
type NovelChapterResponse struct {
	ID            string  `json:"id"`
	NovelID       string  `json:"novel_id"`
	Title         string  `json:"title"`
	Content       string  `json:"content,omitempty"`
	WordCount     int64   `json:"word_count"`
	Number        int     `json:"number"`
	ViewsCount    int64   `json:"views_count"`
	RatingSum     float64 `json:"rating_sum"`
	RatingCount   int64   `json:"rating_count"`
	AverageRating float64 `json:"average_rating"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	HasUserViewed *bool   `json:"has_user_viewed,omitempty"`
}

// ReactionRequest represents a reaction request.
type NovelReactionRequest struct {
	Type string `json:"type" validate:"required,oneof=upvote funny love surprised angry sad"`
}

// ReactionResponse represents a reaction response.
type NovelReactionResponse struct {
	ReactionType string         `json:"reaction_type,omitempty"`
	Novel        *NovelResponse `json:"novel,omitempty"`
	Removed      bool           `json:"removed"`
}

// RatingRequest represents a rating request.
type NovelRatingRequest struct {
	Score float64 `json:"score" validate:"required,min=1,max=10"`
}

// CommentRequest represents a comment request.
type NovelCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// CommentResponse represents a comment response.
type NovelCommentResponse struct {
	ID         string `json:"id"`
	NovelID    string `json:"novel_id"`
	UserID     string `json:"user_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	CanDelete  bool   `json:"can_delete"`
}

// FavoriteListRequest represents a favorite list request.
type NovelFavoriteListRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	IsPublic    bool   `json:"is_public"`
}

// FavoriteListResponse represents a favorite list response.
type NovelFavoriteListResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	NovelCount  int64  `json:"novel_count,omitempty"`
}

// FavoriteListItemRequest represents a favorite list item request.
type NovelFavoriteListItemRequest struct {
	NovelID string `json:"novel_id" validate:"required"`
	Notes   string `json:"notes,omitempty" validate:"max=500"`
}

// FavoriteListItemResponse represents a favorite list item response.
type NovelFavoriteListItemResponse struct {
	ListID    string         `json:"list_id"`
	NovelID   string         `json:"novel_id"`
	Notes     string         `json:"notes,omitempty"`
	AddedAt   string         `json:"added_at"`
	SortOrder int            `json:"sort_order"`
	Novel     *NovelResponse `json:"novel,omitempty"`
}

// ReadingProgressRequest represents a reading progress request.
type NovelReadingProgressRequest struct {
	NovelID   string `json:"novel_id" validate:"required"`
	ChapterID string `json:"chapter_id,omitempty"`
}

// ReadingProgressResponse represents a reading progress response.
type NovelReadingProgressResponse struct {
	ID              string `json:"id"`
	NovelID         string `json:"novel_id"`
	UserID          string `json:"user_id"`
	LastReadChapter string `json:"last_read_chapter"`
	LastReadAt      string `json:"last_read_at"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ViewingHistoryRequest represents a viewing history request.
type NovelViewingHistoryRequest struct {
	NovelID   string `json:"novel_id" validate:"required"`
	ChapterID string `json:"chapter_id,omitempty"`
}

// ViewingHistoryResponse represents a viewing history response.
type NovelViewingHistoryResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	NovelID   string `json:"novel_id"`
	ChapterID string `json:"chapter_id,omitempty"`
	ViewedAt  string `json:"viewed_at"`
	CreatedAt string `json:"created_at"`
}

// RankedNovelResponse represents a ranked novel response.
type RankedNovelResponse struct {
	Novel     *NovelResponse `json:"novel"`
	ViewCount int64          `json:"view_count"`
}

// NovelListResponse represents a paginated novel list response.
type NovelListResponse struct {
	Items       []*NovelResponse `json:"items"`
	Total       int64            `json:"total"`
	TotalPages  int64            `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	PerPage     int              `json:"per_page"`
}

package dtos

// MangaRequest represents a manga creation/update request.
type MangaRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"required,max=2000"`
	Tags        []string `json:"tags" validate:"dive,max=50"`
	CoverImage  string   `json:"cover_image,omitempty"`
}

// MangaResponse represents a manga response.
type MangaResponse struct {
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
type ChapterRequest struct {
	Title   string   `json:"title" validate:"required,min=1,max=256"`
	Number  int      `json:"number" validate:"required,min=1"`
	Pages   []string `json:"pages" validate:"required,min=1"`
	MangaID string   `json:"manga_id" validate:"required"`
}

// ChapterResponse represents a chapter response.
type ChapterResponse struct {
	ID            string   `json:"id"`
	MangaID       string   `json:"manga_id"`
	Title         string   `json:"title"`
	Pages         []string `json:"pages"`
	Number        int      `json:"number"`
	ViewsCount    int64    `json:"views_count"`
	RatingSum     float64  `json:"rating_sum"`
	RatingCount   int64    `json:"rating_count"`
	AverageRating float64  `json:"average_rating"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	HasUserViewed *bool    `json:"has_user_viewed,omitempty"`
}

// ReactionRequest represents a reaction request.
type ReactionRequest struct {
	Type string `json:"type" validate:"required,oneof=upvote funny love surprised angry sad"`
}

// ReactionResponse represents a reaction response.
type ReactionResponse struct {
	ReactionType string         `json:"reaction_type,omitempty"`
	Manga        *MangaResponse `json:"manga,omitempty"`
	Removed      bool           `json:"removed"`
}

// RatingRequest represents a rating request.
type RatingRequest struct {
	Score float64 `json:"score" validate:"required,min=1,max=10"`
}

// CommentRequest represents a comment request.
type CommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// CommentResponse represents a comment response.
type CommentResponse struct {
	ID         string `json:"id"`
	MangaID    string `json:"manga_id"`
	UserID     string `json:"user_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	CanDelete  bool   `json:"can_delete"`
}

// FavoriteListRequest represents a favorite list request.
type FavoriteListRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	IsPublic    bool   `json:"is_public"`
}

// FavoriteListResponse represents a favorite list response.
type FavoriteListResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	MangaCount  int64  `json:"manga_count,omitempty"`
}

// FavoriteListItemRequest represents a favorite list item request.
type FavoriteListItemRequest struct {
	MangaID string `json:"manga_id" validate:"required"`
	Notes   string `json:"notes,omitempty" validate:"max=500"`
}

// FavoriteListItemResponse represents a favorite list item response.
type FavoriteListItemResponse struct {
	ListID    string         `json:"list_id"`
	MangaID   string         `json:"manga_id"`
	Notes     string         `json:"notes,omitempty"`
	AddedAt   string         `json:"added_at"`
	SortOrder int            `json:"sort_order"`
	Manga     *MangaResponse `json:"manga,omitempty"`
}

// ReadingProgressRequest represents a reading progress request.
type ReadingProgressRequest struct {
	MangaID   string `json:"manga_id" validate:"required"`
	ChapterID string `json:"chapter_id,omitempty"`
	Page      int    `json:"page"`
}

// ReadingProgressResponse represents a reading progress response.
type ReadingProgressResponse struct {
	ID              string `json:"id"`
	MangaID         string `json:"manga_id"`
	UserID          string `json:"user_id"`
	LastReadChapter string `json:"last_read_chapter"`
	LastReadPage    int    `json:"last_read_page"`
	LastReadAt      string `json:"last_read_at"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ViewingHistoryRequest represents a viewing history request.
type ViewingHistoryRequest struct {
	MangaID   string `json:"manga_id" validate:"required"`
	ChapterID string `json:"chapter_id,omitempty"`
	Page      int    `json:"page"`
}

// ViewingHistoryResponse represents a viewing history response.
type ViewingHistoryResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	ChapterID string `json:"chapter_id,omitempty"`
	Page      int    `json:"page"`
	ViewedAt  string `json:"viewed_at"`
	CreatedAt string `json:"created_at"`
}

// RankedMangaResponse represents a ranked manga response.
type RankedMangaResponse struct {
	Manga     *MangaResponse `json:"manga"`
	ViewCount int64          `json:"view_count"`
}

// MangaListResponse represents a paginated manga list response.
type MangaListResponse struct {
	Items       []*MangaResponse `json:"items"`
	Total       int64            `json:"total"`
	TotalPages  int64            `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	PerPage     int              `json:"per_page"`
}

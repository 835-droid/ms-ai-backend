package manga

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Manga represents a manga in the system.
type Manga struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title" validate:"required,min=1,max=256"`
	Slug        string             `bson:"slug" json:"slug"`
	Description string             `bson:"description" json:"description" validate:"required,min=1"`
	AuthorID    primitive.ObjectID `bson:"author_id" json:"author_id"`
	Tags        []string           `bson:"tags" json:"tags"`
	CoverImage  string             `bson:"cover_image" json:"cover_image"`
	IsPublished bool               `bson:"is_published" json:"is_published"`
	PublishedAt *time.Time         `bson:"published_at,omitempty" json:"published_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// MangaChapter represents a chapter in a manga.
type MangaChapter struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	Title     string             `bson:"title" json:"title" validate:"required,min=1,max=256"`
	Pages     []string           `bson:"pages" json:"pages" validate:"required,min=1"`
	Number    int                `bson:"number" json:"number" validate:"required,min=1"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MangaRepository defines the manga data operations.
type MangaRepository interface {
	CreateManga(ctx context.Context, manga *Manga) error
	GetMangaByID(ctx context.Context, id primitive.ObjectID) (*Manga, error)
	GetMangaBySlug(ctx context.Context, slug string) (*Manga, error)
	ListMangas(ctx context.Context, skip, limit int64) ([]*Manga, int64, error)
	UpdateManga(ctx context.Context, manga *Manga) error
	DeleteManga(ctx context.Context, id primitive.ObjectID) error
}

// MangaChapterRepository defines the manga chapter data operations.
type MangaChapterRepository interface {
	CreateMangaChapter(ctx context.Context, chapter *MangaChapter) error
	GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*MangaChapter, error)
	ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*MangaChapter, int64, error)
	UpdateMangaChapter(ctx context.Context, chapter *MangaChapter) error
	DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error
}

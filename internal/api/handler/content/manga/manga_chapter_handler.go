package manga

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaChapterHandler handles chapter-related requests for manga
type MangaChapterHandler struct {
	service manga.MangaChapterService
}

// NewMangaChapterHandler creates a new manga chapter handler
func NewMangaChapterHandler(s manga.MangaChapterService) *MangaChapterHandler {
	return &MangaChapterHandler{service: s}
}

type mangaCreateChapterRequest struct {
	Title  string   `json:"title" binding:"required"`
	Images []string `json:"images" binding:"required,min=1"`
	Number float64  `json:"number"`
}

var uploadFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func normalizeChapterImages(images []string) []string {
	normalized := make([]string, 0, len(images))
	for _, image := range images {
		value := strings.TrimSpace(image)
		value = strings.Trim(value, "\"'")
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeChapterForResponse(chapter *manga.MangaChapter) *manga.MangaChapter {
	if chapter == nil {
		return nil
	}
	chapter.Pages = normalizeChapterImages(chapter.Pages)
	return chapter
}

func normalizeChaptersForResponse(chapters []*manga.MangaChapter) []*manga.MangaChapter {
	for i := range chapters {
		chapters[i] = normalizeChapterForResponse(chapters[i])
	}
	return chapters
}

func sanitizeUploadFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "image"
	}
	base = uploadFilenameSanitizer.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._")
	if base == "" {
		return "image"
	}
	return base
}

func chapterUploadPublicPath(mangaID primitive.ObjectID, filename string) string {
	return fmt.Sprintf("/uploads/chapters/%s/%s", mangaID.Hex(), filename)
}

func chapterUploadDiskPath(mangaID primitive.ObjectID, filename string) string {
	return filepath.Join("uploads", "chapters", mangaID.Hex(), filename)
}

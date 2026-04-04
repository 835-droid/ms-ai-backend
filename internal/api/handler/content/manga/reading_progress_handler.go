// ----- START OF FILE: backend/MS-AI/internal/api/handler/content/manga/reading_progress_handler.go -----
package manga

import (
	"net/http"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadingProgressHandler handles reading progress requests
type ReadingProgressHandler struct {
	repo manga.ReadingProgressRepository
}

// NewReadingProgressHandler creates a new reading progress handler
func NewReadingProgressHandler(repo manga.ReadingProgressRepository) *ReadingProgressHandler {
	return &ReadingProgressHandler{repo: repo}
}

// SaveProgressRequest represents the request body for saving reading progress
type SaveProgressRequest struct {
	ChapterID string `json:"chapter_id" binding:"required"`
	Page      int    `json:"page"`
}

// SaveProgress saves a user's reading progress for a manga
// POST /api/mangas/:mangaID/progress
func (h *ReadingProgressHandler) SaveProgress(c *gin.Context) {
	// Get manga ID from URL
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid manga ID")
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		response.Unauthorized(c, "authentication required")
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req SaveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	chapterID, err := primitive.ObjectIDFromHex(req.ChapterID)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid chapter ID")
		return
	}

	progress := &manga.ReadingProgress{
		MangaID:         mangaID,
		UserID:          userID,
		LastReadChapter: chapterID,
		LastReadPage:    req.Page,
		LastReadAt:      time.Now(),
	}

	if err := h.repo.SaveProgress(c.Request.Context(), progress); err != nil {
		response.InternalError(c, "failed to save progress")
		return
	}

	response.SuccessResp(c, http.StatusOK, nil)
}

// GetProgress gets a user's reading progress for a specific manga
// GET /api/mangas/:mangaID/progress
func (h *ReadingProgressHandler) GetProgress(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid manga ID")
		return
	}

	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		response.Unauthorized(c, "authentication required")
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	progress, err := h.repo.GetProgress(c.Request.Context(), mangaID, userID)
	if err != nil {
		response.InternalError(c, "failed to get progress")
		return
	}

	response.SuccessResp(c, http.StatusOK, progress)
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/content/manga/reading_progress_handler.go -----

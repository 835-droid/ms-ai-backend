package manga

import (
	"net/http"
	"strconv"

	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewingHistoryHandler handles viewing history related requests
type ViewingHistoryHandler struct {
	service      manga.ViewingHistoryService
	mangaService manga.MangaService
}

// NewViewingHistoryHandler creates a new ViewingHistoryHandler
func NewViewingHistoryHandler(service manga.ViewingHistoryService, mangaService manga.MangaService) *ViewingHistoryHandler {
	return &ViewingHistoryHandler{service: service, mangaService: mangaService}
}

// TrackView tracks a manga view
// POST /api/mangas/history/track
func (h *ViewingHistoryHandler) TrackView(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		MangaID   string `json:"manga_id" binding:"required"`
		ChapterID string `json:"chapter_id"`
		Page      int    `json:"page"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body")
		return
	}

	mangaID, err := primitive.ObjectIDFromHex(req.MangaID)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
		return
	}

	var chapterID primitive.ObjectID
	if req.ChapterID != "" {
		chapterID, err = primitive.ObjectIDFromHex(req.ChapterID)
		if err != nil {
			response.ErrorResp(c, http.StatusBadRequest, "Invalid chapter ID")
			return
		}
	}

	history, err := h.service.TrackView(c.Request.Context(), userID, mangaID, chapterID, req.Page)
	if err != nil {
		response.InternalError(c, "Failed to track view")
		return
	}

	response.SuccessResp(c, http.StatusOK, history)
}

// GetUserHistory retrieves the user's viewing history
// GET /api/mangas/history
func (h *ViewingHistoryHandler) GetUserHistory(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	histories, total, err := h.service.GetUserHistory(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "Failed to get viewing history")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"items":        histories,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// GetRecentHistory retrieves recent viewing history
// GET /api/mangas/history/recent
func (h *ViewingHistoryHandler) GetRecentHistory(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 50 {
			limit = int64(v)
		}
	}

	histories, err := h.service.GetRecentHistory(c.Request.Context(), userID, limit)
	if err != nil {
		response.InternalError(c, "Failed to get recent viewing history")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"items": histories})
}

// GetUserStats retrieves statistics about the user's viewing history
// GET /api/mangas/history/stats
func (h *ViewingHistoryHandler) GetUserStats(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	stats, err := h.service.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get viewing stats")
		return
	}

	response.SuccessResp(c, http.StatusOK, stats)
}

// DeleteHistoryItem deletes a specific history item
// DELETE /api/mangas/history/:id
func (h *ViewingHistoryHandler) DeleteHistoryItem(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		response.ValidationError(c, "History ID is required")
		return
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid history ID")
		return
	}

	if err := h.service.DeleteHistoryItem(c.Request.Context(), id, userID); err != nil {
		response.InternalError(c, "Failed to delete history item")
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteHistoryByManga deletes all history for a specific manga
// DELETE /api/mangas/history/manga/:mangaID
func (h *ViewingHistoryHandler) DeleteHistoryByManga(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	mangaIDStr := c.Param("mangaID")
	if mangaIDStr == "" {
		response.ValidationError(c, "Manga ID is required")
		return
	}

	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
		return
	}

	if err := h.service.DeleteHistoryByManga(c.Request.Context(), userID, mangaID); err != nil {
		response.InternalError(c, "Failed to delete history for manga")
		return
	}

	c.Status(http.StatusNoContent)
}

// CleanOldHistory deletes old history entries
// DELETE /api/mangas/history/clean
func (h *ViewingHistoryHandler) CleanOldHistory(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	days := 90 // Default to 90 days
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 365 {
			days = v
		}
	}

	deleted, err := h.service.CleanOldHistory(c.Request.Context(), userID, days)
	if err != nil {
		response.InternalError(c, "Failed to clean old history")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"deleted_count": deleted,
		"message":       "Old history cleaned successfully",
	})
}

// Middleware to track views automatically
// This can be used as middleware for manga detail/chapter endpoints
func (h *ViewingHistoryHandler) TrackViewMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID if authenticated (optional auth)
		userIDStr, exists := c.Get(middleware.ContextUserIDKey)
		if !exists {
			c.Next()
			return
		}

		userID, ok := userIDStr.(string)
		if !ok {
			c.Next()
			return
		}

		userOID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.Next()
			return
		}

		mangaIDStr := c.Param("mangaID")
		if mangaIDStr == "" {
			c.Next()
			return
		}

		mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
		if err != nil {
			c.Next()
			return
		}

		// Get chapter ID if present
		chapterIDStr := c.Param("chapterID")
		var chapterID primitive.ObjectID
		if chapterIDStr != "" {
			chapterID, _ = primitive.ObjectIDFromHex(chapterIDStr)
		}

		// Track the view (non-blocking)
		go func() {
			_, _ = h.service.TrackView(c.Request.Context(), userOID, mangaID, chapterID, 0)
		}()

		c.Next()
	}
}

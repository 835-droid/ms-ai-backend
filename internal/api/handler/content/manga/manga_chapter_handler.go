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

// MangaChapterHandler handles chapter-related requests for manga
type MangaChapterHandler struct {
	service manga.MangaChapterService
}

// NewMangaChapterHandler creates a new manga chapter handler
func NewMangaChapterHandler(s manga.MangaChapterService) *MangaChapterHandler {
	return &MangaChapterHandler{service: s}
}

type mangaCreateChapterRequest struct {
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Number  float64 `json:"number"`
}

func (h *MangaChapterHandler) CreateChapter(c *gin.Context) {
	mangaIDStr := c.Param("manga_id")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	var req mangaCreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid payload")
		return
	}

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	chapter := &manga.MangaChapter{
		Number:  int(req.Number),
		Title:   req.Title,
		MangaID: mangaID,
	}

	if _, err := h.service.CreateMangaChapter(c.Request.Context(), chapter, callerID, roles); err != nil {
		response.InternalError(c, "failed to create chapter")
		return
	}

	response.SuccessResp(c, http.StatusCreated, chapter)
}

func (h *MangaChapterHandler) ListChapters(c *gin.Context) {
	mangaIDStr := c.Param("manga_id")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
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
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	var chapters []*manga.MangaChapter
	var total int64
	chapters, total, err = h.service.ListMangaChapters(c.Request.Context(), mangaID, skip, lmt)

	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"limit":    limit,
		"chapters": chapters,
	})
}

func (h *MangaChapterHandler) GetChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapter_id"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		response.InternalError(c, "failed to get chapter")
		return
	}
	if chapter == nil {
		response.NotFound(c, "chapter not found")
		return
	}

	response.SuccessResp(c, http.StatusOK, chapter)
}

func (h *MangaChapterHandler) DeleteChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapter_id"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	if err := h.service.DeleteMangaChapter(c.Request.Context(), chapterID, callerID, roles); err != nil {
		response.InternalError(c, "failed to delete chapter")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MangaChapterHandler) UpdateChapter(c *gin.Context) {
	var req mangaCreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid payload")
		return
	}

	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapter_id"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		response.InternalError(c, "failed to get chapter")
		return
	}
	if chapter == nil {
		response.NotFound(c, "chapter not found")
		return
	}

	chapter.Title = req.Title
	if req.Number != 0 {
		chapter.Number = int(req.Number)
	}

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	if err := h.service.UpdateMangaChapter(c.Request.Context(), chapter, callerID, roles); err != nil {
		response.InternalError(c, "failed to update chapter")
		return
	}

	response.SuccessResp(c, http.StatusOK, chapter)
}

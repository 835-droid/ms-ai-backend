package manga

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	common "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *MangaChapterHandler) CreateChapter(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
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
	req.Images = normalizeChapterImages(req.Images)
	if len(req.Images) == 0 {
		response.ValidationError(c, "at least one valid image url is required")
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
		Pages:   req.Images,
	}

	if _, err := h.service.CreateMangaChapter(c.Request.Context(), chapter, callerID, roles); err != nil {
		response.InternalError(c, "failed to create chapter")
		return
	}
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusCreated, chapter)
}

func (h *MangaChapterHandler) ListChapters(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
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
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	chapters, total, err := h.service.ListMangaChapters(c.Request.Context(), mangaID, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list chapters")
		return
	}
	chapters = normalizeChaptersForResponse(chapters)

	// Set has_user_viewed for authenticated users
	if uid, ok := c.Get(middleware.ContextUserIDKey); ok {
		if userID, ok := uid.(string); ok {
			if objID, err := primitive.ObjectIDFromHex(userID); err == nil {
				for _, chapter := range chapters {
					viewed, err := h.service.HasUserViewedChapter(c.Request.Context(), chapter.ID, objID)
					if err == nil {
						chapter.HasUserViewed = &viewed
					}
				}
			}
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"limit":    limit,
		"chapters": chapters,
	})
}

func (h *MangaChapterHandler) GetChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		if errors.Is(err, common.ErrMangaChapterNotFound) || errors.Is(err, common.ErrNotFound) {
			response.NotFound(c, "chapter not found")
			return
		}
		response.InternalError(c, "failed to get chapter")
		return
	}
	if chapter == nil {
		response.NotFound(c, "chapter not found")
		return
	}
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusOK, chapter)
}

func (h *MangaChapterHandler) DeleteChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
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
		if errors.Is(err, common.ErrMangaChapterNotFound) || errors.Is(err, common.ErrNotFound) {
			response.NotFound(c, "chapter not found")
			return
		}
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
	req.Images = normalizeChapterImages(req.Images)
	if len(req.Images) == 0 {
		response.ValidationError(c, "at least one valid image url is required")
		return
	}

	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		if errors.Is(err, common.ErrMangaChapterNotFound) || errors.Is(err, common.ErrNotFound) {
			response.NotFound(c, "chapter not found")
			return
		}
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
	chapter.Pages = req.Images

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
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusOK, chapter)
}

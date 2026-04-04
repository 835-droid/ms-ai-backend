// Package novel handles novel-related HTTP requests.
package novel

import (
	"errors"
	"net/http"
	"strconv"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/content/novel"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NovelHandler handles novel-related requests
type NovelHandler struct {
	service novel.NovelService
}

// NewNovelHandler creates a new NovelHandler
func NewNovelHandler(s novel.NovelService) *NovelHandler {
	return &NovelHandler{service: s}
}

// getPaginationParams extracts and validates pagination parameters
func getPaginationParams(c *gin.Context) (page, limit int, skip, lmt int64) {
	page = 1
	limit = 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > maxPageSize {
				v = maxPageSize
			}
			limit = v
		}
	}

	skip = int64((page - 1) * limit)
	lmt = int64(limit)
	return
}

const (
	maxTitleLength       = 200
	maxDescriptionLength = 2000
	maxTagLength         = 50
	maxTagsCount         = 10
	maxPageSize          = 100
)

type novelRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Description string   `json:"description" binding:"max=2000"`
	Tags        []string `json:"tags" binding:"dive,max=50"`
	CoverImage  string   `json:"cover_image"`
}

type novelRatingRequest struct {
	Score float64 `json:"score" binding:"required"`
}

// getCallerInfo extracts user ID and roles from context
func getCallerInfo(c *gin.Context) (primitive.ObjectID, []string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, nil, errors.New("user not authenticated")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return primitive.NilObjectID, nil, errors.New("invalid user ID")
	}

	oid, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID, nil, err
	}

	roles, _ := c.Get("roles")
	roleSlice, _ := roles.([]string)

	return oid, roleSlice, nil
}

// CreateNovel handles the creation of a new novel.
// @Summary Create a new novel
// @Description Create a new novel with the given details
// @Accept json
// @Produce json
// @Param novel body novelRequest true "Novel details"
// @Success 201 {object} novel.Novel
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Router /novels [post]
func (h *NovelHandler) CreateNovel(c *gin.Context) {
	var req novelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Validate tags count
	if len(req.Tags) > maxTagsCount {
		response.ValidationError(c, "too many tags")
		return
	}

	authorID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// Ensure tags is not nil
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	n := &novel.Novel{
		Title:       req.Title,
		Description: req.Description,
		Tags:        tags,
		CoverImage:  req.CoverImage,
		AuthorID:    authorID,
	}

	created, err := h.service.CreateNovel(c.Request.Context(), n)
	if err != nil {
		if err == core.ErrForbidden {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to create novel")
		return
	}

	response.SuccessResp(c, http.StatusCreated, created)
}

// ListNovels retrieves a paginated list of novels.
// @Summary List all novels
// @Description Get a paginated list of all novels
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.PaginatedResponse
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /novels [get]
func (h *NovelHandler) ListNovels(c *gin.Context) {
	page, limit, skip, lmt := getPaginationParams(c)

	list, total, err := h.service.ListNovels(c.Request.Context(), skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list novels")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
		"items":        list,
	})
}

// GetNovel retrieves a specific novel by ID.
// @Summary Get a novel by ID
// @Description Get details of a specific novel
// @Produce json
// @Param novel_id path string true "Novel ID"
// @Success 200 {object} novel.Novel
// @Failure 400 {object} response.ErrorResponse "Invalid novel ID"
// @Failure 404 {object} response.ErrorResponse "Novel not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /novels/{novelID} [get]
func (h *NovelHandler) GetNovel(c *gin.Context) {
	id := c.Param("novelID")
	if id == "" {
		response.ValidationError(c, "novel id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid novel id")
		return
	}

	n, err := h.service.GetNovel(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "novel not found")
			return
		}
		response.InternalError(c, "failed to get novel")
		return
	}
	if n == nil {
		response.NotFound(c, "novel not found")
		return
	}

	response.SuccessResp(c, http.StatusOK, n)
}

// UpdateNovel updates an existing novel.
// @Summary Update a novel
// @Description Update details of an existing novel
// @Accept json
// @Produce json
// @Param novel_id path string true "Novel ID"
// @Param novel body novelRequest true "Updated novel details"
// @Success 200 {object} novel.Novel
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Novel not found"
// @Router /novels/{novelID} [put]
func (h *NovelHandler) UpdateNovel(c *gin.Context) {
	id := c.Param("novelID")
	if id == "" {
		response.ValidationError(c, "novel id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid novel id")
		return
	}

	var req novelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Validate tags count
	if len(req.Tags) > maxTagsCount {
		response.ValidationError(c, "too many tags")
		return
	}

	n, err := h.service.GetNovel(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "novel not found")
			return
		}
		response.InternalError(c, "failed to get novel")
		return
	}
	if n == nil {
		response.NotFound(c, "novel not found")
		return
	}

	// Update fields
	n.Title = req.Title
	n.Description = req.Description
	if req.Tags != nil {
		n.Tags = req.Tags
	}
	n.CoverImage = req.CoverImage

	callerID, roles, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.UpdateNovel(c.Request.Context(), n, callerID, roles); err != nil {
		if err == core.ErrForbidden || err == core.ErrUnauthorized {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to update novel")
		return
	}

	response.SuccessResp(c, http.StatusOK, n)
}

// DeleteNovel removes an existing novel.
// @Summary Delete a novel
// @Description Delete an existing novel
// @Produce json
// @Param novelID path string true "Novel ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid novel ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Router /novels/{novelID} [delete]
func (h *NovelHandler) DeleteNovel(c *gin.Context) {
	id := c.Param("novelID")
	if id == "" {
		response.ValidationError(c, "novel id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid novel id")
		return
	}

	callerID, roles, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteNovel(c.Request.Context(), oid, callerID, roles); err != nil {
		if err == core.ErrForbidden || err == core.ErrUnauthorized {
			response.Forbidden(c, "forbidden")
			return
		}
		if errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "novel not found")
			return
		}
		response.InternalError(c, "failed to delete novel")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMostViewed lists the most viewed novels for a given period.
// @Summary List most viewed novels
// @Description Get the most viewed novels for day/week/month/all time
// @Tags novels
// @Accept json
// @Produce json
// @Param period query string false "Period: day, week, month, all" default(day)
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} novel.RankedNovel
// @Router /api/novels/most-viewed [get]
func (h *NovelHandler) ListMostViewed(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	rankedNovels, err := h.service.ListMostViewed(c.Request.Context(), period, 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list most viewed novels")
		return
	}

	response.SuccessResp(c, http.StatusOK, rankedNovels)
}

// ListRecentlyUpdated lists the recently updated novels.
// @Summary List recently updated novels
// @Description Get the most recently updated novels
// @Tags novels
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} novel.Novel
// @Router /api/novels/recently-updated [get]
func (h *NovelHandler) ListRecentlyUpdated(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	novels, err := h.service.ListRecentlyUpdated(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list recently updated novels")
		return
	}

	response.SuccessResp(c, http.StatusOK, novels)
}

// ListMostFollowed lists the most followed novels.
// @Summary List most followed novels
// @Description Get the most followed novels ordered by favorites count
// @Tags novels
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} novel.Novel
// @Router /api/novels/most-followed [get]
func (h *NovelHandler) ListMostFollowed(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	novels, err := h.service.ListMostFollowed(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list most followed novels")
		return
	}

	response.SuccessResp(c, http.StatusOK, novels)
}

// ListTopRated lists the top rated novels.
// @Summary List top rated novels
// @Description Get the top rated novels ordered by average rating
// @Tags novels
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} novel.Novel
// @Router /api/novels/top-rated [get]
func (h *NovelHandler) ListTopRated(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	novels, err := h.service.ListTopRated(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list top rated novels")
		return
	}

	response.SuccessResp(c, http.StatusOK, novels)
}

// RateNovel adds a rating to a novel
func (h *NovelHandler) RateNovel(c *gin.Context) {
	novelIDStr := c.Param("novelID")
	novelID, err := primitive.ObjectIDFromHex(novelIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid novel ID")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Score float64 `json:"score" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body")
		return
	}

	_, err = h.service.AddRating(c.Request.Context(), novelID, userID, req.Score)
	if err != nil {
		response.InternalError(c, "Failed to rate novel")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Novel rated successfully"})
}

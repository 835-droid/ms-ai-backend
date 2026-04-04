// ----- START OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_handler.go -----
package manga

import (
	"errors"
	"net/http"
	"strconv"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaHandler handles manga-related requests
type MangaHandler struct {
	service        manga.MangaService
	favListService manga.FavoriteListService
}

// reactionRequest represents a reaction request payload.
type reactionRequest struct {
	Type string `json:"type" binding:"required,oneof=upvote funny love surprised angry sad"`
}

func NewMangaHandler(s manga.MangaService, favListService manga.FavoriteListService) *MangaHandler {
	return &MangaHandler{service: s, favListService: favListService}
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

type mangaRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Description string   `json:"description" binding:"max=2000"`
	Tags        []string `json:"tags" binding:"dive,max=50"`
	CoverImage  string   `json:"cover_image"`
}

type mangaRatingRequest struct {
	Score float64 `json:"score" binding:"required"`
}

// CreateManga handles the creation of a new manga.
// @Summary Create a new manga
// @Description Create a new manga with the given details
// @Accept json
// @Produce json
// @Param manga body mangaRequest true "Manga details"
// @Success 201 {object} manga.Manga
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Router /manga [post]
func (h *MangaHandler) CreateManga(c *gin.Context) {
	var req mangaRequest
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

	m := &manga.Manga{
		Title:       req.Title,
		Description: req.Description,
		Tags:        tags,
		CoverImage:  req.CoverImage,
		AuthorID:    authorID,
	}

	created, err := h.service.CreateManga(c.Request.Context(), m)
	if err != nil {
		if err == core.ErrForbidden {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to create manga")
		return
	}

	response.SuccessResp(c, http.StatusCreated, created)
}

// ListMangas retrieves a paginated list of manga.
// @Summary List all manga
// @Description Get a paginated list of all manga
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.PaginatedResponse
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /manga [get]
func (h *MangaHandler) ListMangas(c *gin.Context) {
	page, limit, skip, lmt := getPaginationParams(c)

	list, total, err := h.service.ListMangas(c.Request.Context(), skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list manga")
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

// GetManga retrieves a specific manga by ID.
// @Summary Get a manga by ID
// @Description Get details of a specific manga
// @Produce json
// @Param manga_id path string true "Manga ID"
// @Success 200 {object} manga.Manga
// @Failure 400 {object} response.ErrorResponse "Invalid manga ID"
// @Failure 404 {object} response.ErrorResponse "Manga not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /manga/{mangaID} [get]
func (h *MangaHandler) GetManga(c *gin.Context) {
	id := c.Param("mangaID")
	if id == "" {
		response.ValidationError(c, "manga id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	m, err := h.service.GetManga(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrMangaNotFound) {
			response.NotFound(c, "manga not found")
			return
		}
		response.InternalError(c, "failed to get manga")
		return
	}
	if m == nil {
		response.NotFound(c, "manga not found")
		return
	}

	response.SuccessResp(c, http.StatusOK, m)
}

// IncrementViews increases the view count for a manga.
// getMangaID extracts and validates manga ID from URL parameters
func getMangaID(c *gin.Context) (primitive.ObjectID, error) {
	id := c.Param("mangaID")
	if id == "" {
		return primitive.NilObjectID, errors.New("manga id is required")
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid manga id")
	}
	return oid, nil
}

// RateManga adds a rating to a manga
func (h *MangaHandler) RateManga(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
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

	_, err = h.service.AddRating(c.Request.Context(), mangaID, userID, req.Score)
	if err != nil {
		response.InternalError(c, "Failed to rate manga")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Manga rated successfully"})
}

// UpdateManga updates an existing manga.
// @Summary Update a manga
// @Description Update details of an existing manga
// @Accept json
// @Produce json
// @Param manga_id path string true "Manga ID"
// @Param manga body mangaRequest true "Updated manga details"
// @Success 200 {object} manga.Manga
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Manga not found"
// @Router /manga/{mangaID} [put]
func (h *MangaHandler) UpdateManga(c *gin.Context) {
	id := c.Param("mangaID")
	if id == "" {
		response.ValidationError(c, "manga id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	var req mangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Validate tags count
	if len(req.Tags) > maxTagsCount {
		response.ValidationError(c, "too many tags")
		return
	}

	m, err := h.service.GetManga(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, core.ErrMangaNotFound) || errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "manga not found")
			return
		}
		response.InternalError(c, "failed to get manga")
		return
	}
	if m == nil {
		response.NotFound(c, "manga not found")
		return
	}

	// Update fields
	m.Title = req.Title
	m.Description = req.Description
	if req.Tags != nil {
		m.Tags = req.Tags
	}
	m.CoverImage = req.CoverImage

	callerID, roles, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.UpdateManga(c.Request.Context(), m, callerID, roles); err != nil {
		if err == core.ErrForbidden || err == core.ErrUnauthorized {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to update manga")
		return
	}

	response.SuccessResp(c, http.StatusOK, m)
}

// DeleteManga removes an existing manga.
// @Summary Delete a manga
// @Description Delete an existing manga
// @Produce json
// @Param mangaID path string true "Manga ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid manga ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Router /manga/{mangaID} [delete]
func (h *MangaHandler) DeleteManga(c *gin.Context) {
	id := c.Param("mangaID")
	if id == "" {
		response.ValidationError(c, "manga id is required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	callerID, roles, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteManga(c.Request.Context(), oid, callerID, roles); err != nil {
		if err == core.ErrForbidden || err == core.ErrUnauthorized {
			response.Forbidden(c, "forbidden")
			return
		}
		if errors.Is(err, core.ErrMangaNotFound) || errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "manga not found")
			return
		}
		response.InternalError(c, "failed to delete manga")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMostViewed lists the most viewed mangas for a given period.
// @Summary List most viewed mangas
// @Description Get the most viewed mangas for day/week/month/all time
// @Tags manga
// @Accept json
// @Produce json
// @Param period query string false "Period: day, week, month, all" default(day)
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} manga.RankedManga
// @Router /api/mangas/most-viewed [get]
func (h *MangaHandler) ListMostViewed(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	rankedMangas, err := h.service.ListMostViewed(c.Request.Context(), period, 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list most viewed mangas")
		return
	}

	response.SuccessResp(c, http.StatusOK, rankedMangas)
}

// ListRecentlyUpdated lists the recently updated mangas.
// @Summary List recently updated mangas
// @Description Get the most recently updated mangas
// @Tags manga
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} manga.Manga
// @Router /api/mangas/recently-updated [get]
func (h *MangaHandler) ListRecentlyUpdated(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	mangas, err := h.service.ListRecentlyUpdated(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list recently updated mangas")
		return
	}

	response.SuccessResp(c, http.StatusOK, mangas)
}

// ListMostFollowed lists the most followed mangas.
// @Summary List most followed mangas
// @Description Get the most followed mangas ordered by favorites count
// @Tags manga
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} manga.Manga
// @Router /api/mangas/most-followed [get]
func (h *MangaHandler) ListMostFollowed(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	mangas, err := h.service.ListMostFollowed(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list most followed mangas")
		return
	}

	response.SuccessResp(c, http.StatusOK, mangas)
}

// ListTopRated lists the top rated mangas.
// @Summary List top rated mangas
// @Description Get the top rated mangas ordered by average rating
// @Tags manga
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} manga.Manga
// @Router /api/mangas/top-rated [get]
func (h *MangaHandler) ListTopRated(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	mangas, err := h.service.ListTopRated(c.Request.Context(), 0, int64(limit))
	if err != nil {
		response.InternalError(c, "failed to list top rated mangas")
		return
	}

	response.SuccessResp(c, http.StatusOK, mangas)
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_handler.go -----

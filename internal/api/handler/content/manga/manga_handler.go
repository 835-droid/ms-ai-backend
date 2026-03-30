package manga

import (
	"net/http"
	"strconv"

	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaHandler handles manga-related requests
type MangaHandler struct {
	service manga.MangaService
}

func NewMangaHandler(s manga.MangaService) *MangaHandler {
	return &MangaHandler{service: s}
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

// getCallerInfo extracts the caller ID and roles from the gin context
func getCallerInfo(c *gin.Context) (primitive.ObjectID, []string, error) {
	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return primitive.NilObjectID, nil, core.ErrUnauthorized
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		return primitive.NilObjectID, nil, core.ErrUnauthorized
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	return callerID, roles, nil
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
// @Router /manga/{manga_id} [get]
func (h *MangaHandler) GetManga(c *gin.Context) {
	id := c.Param("manga_id")
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
		response.InternalError(c, "failed to get manga")
		return
	}
	if m == nil {
		response.NotFound(c, "manga not found")
		return
	}

	response.SuccessResp(c, http.StatusOK, m)
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
// @Router /manga/{manga_id} [put]
func (h *MangaHandler) UpdateManga(c *gin.Context) {
	id := c.Param("manga_id")
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
		if err == core.ErrForbidden {
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
// @Param manga_id path string true "Manga ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid manga ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Router /manga/{manga_id} [delete]
func (h *MangaHandler) DeleteManga(c *gin.Context) {
	id := c.Param("manga_id")
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
		if err == core.ErrForbidden {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to delete manga")
		return
	}

	c.Status(http.StatusNoContent)
}

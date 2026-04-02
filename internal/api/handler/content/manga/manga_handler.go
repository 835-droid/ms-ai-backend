// ----- START OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_handler.go -----
package manga

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

// reactionRequest represents a reaction request payload.
type reactionRequest struct {
	Type string `json:"type" binding:"required,oneof=upvote funny love surprised angry sad"`
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
func (h *MangaHandler) IncrementViews(c *gin.Context) {
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

	m, err := h.service.IncrementViews(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, core.ErrMangaNotFound) || errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "manga not found")
			return
		}
		response.InternalError(c, "failed to increment views")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"views_count":    m.ViewsCount,
		"likes_count":    m.LikesCount,
		"average_rating": m.AverageRating,
		"rating_count":   m.RatingCount,
		"manga":          m,
	})
}

// SetReaction sets or toggles a reaction for a manga.
func (h *MangaHandler) SetReaction(c *gin.Context) {
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

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req reactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid reaction type")
		return
	}

	reactionType := manga.ReactionType(req.Type)
	m, reaction, err := h.service.SetReaction(c.Request.Context(), oid, userID, reactionType)
	if err != nil {
		if errors.Is(err, core.ErrMangaNotFound) || errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "manga not found")
			return
		}
		if strings.Contains(err.Error(), "reaction request already in progress") {
			response.ErrorResp(c, http.StatusTooManyRequests, "reaction request already in progress, please wait")
			return
		}
		response.InternalError(c, "failed to set reaction")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"reaction_type":   reaction,
		"reactions_count": m.ReactionsCount,
		"likes_count":     m.LikesCount,
		"views_count":     m.ViewsCount,
		"average_rating":  m.AverageRating,
		"manga":           m,
	})
}

// GetUserReaction gets the current reaction for a user on a manga.
func (h *MangaHandler) GetUserReaction(c *gin.Context) {
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

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reaction, err := h.service.GetUserReaction(c.Request.Context(), oid, userID)
	if err != nil {
		response.InternalError(c, "failed to get user reaction")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"reaction_type": reaction,
	})
}

func (h *MangaHandler) ListLikedMangas(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	page, limit, skip, lmt := getPaginationParams(c)
	mangas, total, err := h.service.ListLikedMangas(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list liked mangas")
		return
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
		"items":        mangas,
	})
}

// RateManga stores a user rating for a manga.
func (h *MangaHandler) RateManga(c *gin.Context) {
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

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req mangaRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if req.Score < 1 || req.Score > 5 {
		response.ValidationError(c, "score must be between 1 and 5")
		return
	}

	m, err := h.service.AddRating(c.Request.Context(), oid, userID, req.Score)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrMangaNotFound), errors.Is(err, core.ErrNotFound):
			response.NotFound(c, "manga not found")
			return
		default:
			response.InternalError(c, "failed to rate manga")
			return
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"average_rating": m.AverageRating,
		"rating_count":   m.RatingCount,
		"likes_count":    m.LikesCount,
		"views_count":    m.ViewsCount,
		"manga":          m,
	})
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
		if err == core.ErrForbidden {
			response.Forbidden(c, "forbidden")
			return
		}
		response.InternalError(c, "failed to delete manga")
		return
	}

	c.Status(http.StatusNoContent)
}

// ========== ENGAGEMENT METHODS ==========

// AddFavorite adds a manga to user's favorites
func (h *MangaHandler) AddFavorite(c *gin.Context) {
	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.AddFavorite(c.Request.Context(), oid, userID); err != nil {
		response.InternalError(c, "failed to add favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "favorite added"})
}

// RemoveFavorite removes a manga from user's favorites
func (h *MangaHandler) RemoveFavorite(c *gin.Context) {
	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.RemoveFavorite(c.Request.Context(), oid, userID); err != nil {
		response.InternalError(c, "failed to remove favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "favorite removed"})
}

// IsFavorite checks if a manga is favorited by the user
func (h *MangaHandler) IsFavorite(c *gin.Context) {
	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		// For anonymous users, return false
		response.SuccessResp(c, http.StatusOK, gin.H{"is_favorite": false})
		return
	}

	isFav, err := h.service.IsFavorite(c.Request.Context(), oid, userID)
	if err != nil {
		response.InternalError(c, "failed to check favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"is_favorite": isFav})
}

// ListFavorites retrieves user's favorite mangas
func (h *MangaHandler) ListFavorites(c *gin.Context) {
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

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	mangas, total, err := h.service.ListFavorites(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list favorites")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"items":        mangas,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// AddMangaComment adds a comment to a manga
func (h *MangaHandler) AddMangaComment(c *gin.Context) {
	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	oid, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	username, _ := c.Get("username")
	usernameStr, _ := username.(string)

	comment := &manga.MangaComment{
		MangaID:  oid,
		UserID:   userID,
		Username: usernameStr,
		Content:  req.Content,
	}

	if err := h.service.AddMangaComment(c.Request.Context(), comment); err != nil {
		response.InternalError(c, "failed to add comment")
		return
	}

	response.SuccessResp(c, http.StatusCreated, comment)
}

// ListMangaComments retrieves comments for a manga
func (h *MangaHandler) ListMangaComments(c *gin.Context) {
	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(mangaID)
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

	comments, total, err := h.service.ListMangaComments(c.Request.Context(), oid, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list comments")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"data":         comments,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// DeleteMangaComment deletes a manga comment
func (h *MangaHandler) DeleteMangaComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		response.ValidationError(c, "comment id required")
		return
	}

	oid, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		response.ValidationError(c, "invalid comment id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteMangaComment(c.Request.Context(), oid, userID); err != nil {
		response.InternalError(c, "failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_handler.go -----

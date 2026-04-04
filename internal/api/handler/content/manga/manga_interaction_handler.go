// internal/api/handler/content/manga/manga_interaction_handler.go
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

// IncrementViews increments the view count for a manga
func (h *MangaHandler) IncrementViews(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
		return
	}

	_, err = h.service.IncrementViews(c.Request.Context(), mangaID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "Failed to increment views")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Views incremented"})
}

// SetReaction sets a reaction for a manga
func (h *MangaHandler) SetReaction(c *gin.Context) {
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

	var req reactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body")
		return
	}

	reactionType, err := manga.ReactionTypeFromString(req.Type)
	if err != nil {
		response.ValidationError(c, "Invalid reaction type")
		return
	}

	manga, reaction, err := h.service.SetReaction(c.Request.Context(), mangaID, userID, reactionType)
	if err != nil {
		response.InternalError(c, "Failed to set reaction")
		return
	}

	removed := reaction == ""
	var reactionTypeResp interface{} = reaction
	if removed {
		reactionTypeResp = nil
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"reaction_type": reactionTypeResp,
		"manga":         manga,
		"removed":       removed,
	})
}

// GetUserReaction gets the user's reaction for a manga
func (h *MangaHandler) GetUserReaction(c *gin.Context) {
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

	reaction, err := h.service.GetUserReaction(c.Request.Context(), mangaID, userID)
	if err != nil {
		response.InternalError(c, "Failed to get user reaction")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"reaction_type": reaction})
}

// ListLikedMangas lists mangas liked by the user
func (h *MangaHandler) ListLikedMangas(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)

	mangas, total, err := h.service.ListLikedMangas(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "Failed to list liked mangas")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"mangas": mangas,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// AddFavorite adds a manga to user's favorites
func (h *MangaHandler) AddFavorite(c *gin.Context) {
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

	if err := h.service.AddFavorite(c.Request.Context(), mangaID, userID); err != nil {
		response.InternalError(c, "Failed to add favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Manga added to favorites"})
}

// RemoveFavorite removes a manga from user's favorites
func (h *MangaHandler) RemoveFavorite(c *gin.Context) {
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

	if err := h.service.RemoveFavorite(c.Request.Context(), mangaID, userID); err != nil {
		response.InternalError(c, "Failed to remove favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Manga removed from favorites"})
}

// IsFavorite checks if a manga is in user's favorites
func (h *MangaHandler) IsFavorite(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.SuccessResp(c, http.StatusOK, gin.H{"is_favorite": false})
		return
	}

	isFav, err := h.service.IsFavorite(c.Request.Context(), mangaID, userID)
	if err != nil {
		response.InternalError(c, "Failed to check favorite status")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"is_favorite": isFav})
}

// ListFavorites lists user's favorite mangas
func (h *MangaHandler) ListFavorites(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)

	mangas, total, err := h.service.ListFavorites(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "Failed to list favorites")
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

	username, _ := c.Get(middleware.ContextUsernameKey)
	usernameStr, _ := username.(string)
	if usernameStr == "" {
		usernameStr = userID.Hex() // fallback to ID if username wasn't provided
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body")
		return
	}

	comment := &manga.MangaComment{
		MangaID:  mangaID,
		UserID:   userID,
		Username: usernameStr,
		Content:  req.Content,
	}

	if err := h.service.AddMangaComment(c.Request.Context(), comment); err != nil {
		response.InternalError(c, "Failed to add comment")
		return
	}

	response.SuccessResp(c, http.StatusCreated, comment)
}

// ListMangaComments lists comments for a manga
func (h *MangaHandler) ListMangaComments(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid manga ID")
		return
	}

	page := 1
	limit := 20
	sort := "newest" // default sort

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
	if s := c.Query("sort"); s != "" {
		if s == "newest" || s == "oldest" {
			sort = s
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	comments, total, err := h.service.ListMangaComments(c.Request.Context(), mangaID, skip, lmt, sort)
	if err != nil {
		response.InternalError(c, "Failed to list comments")
		return
	}

	var callerID primitive.ObjectID
	if uid, ok := c.Get(middleware.ContextUserIDKey); ok {
		if uidStr, ok := uid.(string); ok {
			if oid, err := primitive.ObjectIDFromHex(uidStr); err == nil {
				callerID = oid
			}
		}
	}

	mapped := make([]gin.H, 0, len(comments))
	for _, comment := range comments {
		mapped = append(mapped, gin.H{
			"id":          comment.ID,
			"manga_id":    comment.MangaID,
			"user_id":     comment.UserID,
			"author_name": comment.Username,
			"content":     comment.Content,
			"created_at":  comment.CreatedAt,
			"updated_at":  comment.UpdatedAt,
			"can_delete":  !callerID.IsZero() && comment.UserID == callerID,
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"data":         mapped,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// DeleteMangaComment deletes a comment from a manga
func (h *MangaHandler) DeleteMangaComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteMangaComment(c.Request.Context(), commentID, userID); err != nil {
		response.InternalError(c, "Failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}

// ========== FAVORITE LIST HANDLERS ==========

// CreateFavoriteList creates a new favorite list for the user
func (h *MangaHandler) CreateFavoriteList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" max:"500"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	list, err := h.favListService.CreateList(c.Request.Context(), userID.Hex(), req.Name, req.Description, req.IsPublic)
	if err != nil {
		if err == manga.ErrDuplicateList {
			response.ErrorResp(c, http.StatusConflict, "List with same name already exists")
			return
		}
		response.InternalError(c, "Failed to create favorite list")
		return
	}

	response.SuccessResp(c, http.StatusCreated, list)
}

// GetFavoriteList retrieves a specific favorite list
func (h *MangaHandler) GetFavoriteList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	if listID == "" {
		response.ValidationError(c, "List ID is required")
		return
	}

	list, err := h.favListService.GetList(c.Request.Context(), listID, userID.Hex())
	if err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		response.InternalError(c, "Failed to get favorite list")
		return
	}

	response.SuccessResp(c, http.StatusOK, list)
}

// ListMyFavoriteLists retrieves all favorite lists for the current user
func (h *MangaHandler) ListMyFavoriteLists(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)

	lists, total, err := h.favListService.ListUserLists(c.Request.Context(), userID.Hex(), skip, lmt)
	if err != nil {
		response.InternalError(c, "Failed to list favorite lists")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"items":        lists,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// UpdateFavoriteList updates a favorite list
func (h *MangaHandler) UpdateFavoriteList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	if listID == "" {
		response.ValidationError(c, "List ID is required")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" max:"500"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	list, err := h.favListService.UpdateList(c.Request.Context(), listID, userID.Hex(), req.Name, req.Description, req.IsPublic)
	if err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		if err == manga.ErrDuplicateList {
			response.ErrorResp(c, http.StatusConflict, "List with same name already exists")
			return
		}
		response.InternalError(c, "Failed to update favorite list")
		return
	}

	response.SuccessResp(c, http.StatusOK, list)
}

// DeleteFavoriteList deletes a favorite list
func (h *MangaHandler) DeleteFavoriteList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	if listID == "" {
		response.ValidationError(c, "List ID is required")
		return
	}

	if err := h.favListService.DeleteList(c.Request.Context(), listID, userID.Hex()); err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		response.InternalError(c, "Failed to delete favorite list")
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMangaToList adds a manga to a favorite list
func (h *MangaHandler) AddMangaToList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	if listID == "" {
		response.ValidationError(c, "List ID is required")
		return
	}

	var req struct {
		MangaID string `json:"manga_id" binding:"required"`
		Notes   string `json:"notes" max:"500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.favListService.AddMangaToList(c.Request.Context(), listID, userID.Hex(), req.MangaID, req.Notes); err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		response.InternalError(c, "Failed to add manga to list")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Manga added to list successfully"})
}

// RemoveMangaFromList removes a manga from a favorite list
func (h *MangaHandler) RemoveMangaFromList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	mangaID := c.Param("mangaID")
	if listID == "" || mangaID == "" {
		response.ValidationError(c, "List ID and Manga ID are required")
		return
	}

	if err := h.favListService.RemoveMangaFromList(c.Request.Context(), listID, userID.Hex(), mangaID); err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		if err == manga.ErrMangaNotInList {
			response.NotFound(c, "Manga not in list")
			return
		}
		response.InternalError(c, "Failed to remove manga from list")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMangaInList retrieves all manga in a specific list
func (h *MangaHandler) ListMangaInList(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	listID := c.Param("listID")
	if listID == "" {
		response.ValidationError(c, "List ID is required")
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)

	items, total, err := h.favListService.ListMangaInList(c.Request.Context(), listID, userID.Hex(), skip, lmt)
	if err != nil {
		if err == manga.ErrListNotFound {
			response.NotFound(c, "List not found")
			return
		}
		if err == manga.ErrListNotOwned {
			response.Forbidden(c, "You don't have access to this list")
			return
		}
		response.InternalError(c, "Failed to list manga in list")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"items":        items,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// GetUserMangaLists retrieves all lists that contain a specific manga
func (h *MangaHandler) GetUserMangaLists(c *gin.Context) {
	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	mangaID := c.Param("mangaID")
	if mangaID == "" {
		response.ValidationError(c, "Manga ID is required")
		return
	}

	lists, err := h.favListService.GetUserMangaLists(c.Request.Context(), userID.Hex(), mangaID)
	if err != nil {
		response.InternalError(c, "Failed to get manga lists")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"lists": lists})
}

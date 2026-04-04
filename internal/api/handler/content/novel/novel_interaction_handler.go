// Package novel handles novel interaction-related HTTP requests.
package novel

import (
	"errors"
	"net/http"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/content/novel"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// reactionRequest represents a reaction request payload.
type reactionRequest struct {
	Type string `json:"type" binding:"required,oneof=upvote funny love surprised angry sad"`
}

// IncrementViews increases the view count for a novel.
func (h *NovelHandler) IncrementViews(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	_, err = h.service.IncrementViews(c.Request.Context(), novelID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "novel not found")
			return
		}
		response.InternalError(c, "failed to increment views")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "View counted"})
}

// SetReaction sets a reaction for a novel.
func (h *NovelHandler) SetReaction(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req reactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	n, reaction, err := h.service.SetReaction(c.Request.Context(), novelID, userID, novel.ReactionType(req.Type))
	if err != nil {
		response.InternalError(c, "failed to set reaction")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"reaction_type": reaction,
		"novel":         n,
		"removed":       reaction == "",
	})
}

// GetUserReaction gets the current reaction type for a user on a novel.
func (h *NovelHandler) GetUserReaction(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reaction, err := h.service.GetUserReaction(c.Request.Context(), novelID, userID)
	if err != nil {
		response.InternalError(c, "failed to get reaction")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"reaction": reaction})
}

// AddFavorite adds a novel to user's favorites.
func (h *NovelHandler) AddFavorite(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.AddFavorite(c.Request.Context(), novelID, userID); err != nil {
		response.InternalError(c, "failed to add favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Added to favorites"})
}

// RemoveFavorite removes a novel from user's favorites.
func (h *NovelHandler) RemoveFavorite(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.RemoveFavorite(c.Request.Context(), novelID, userID); err != nil {
		response.InternalError(c, "failed to remove favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "Removed from favorites"})
}

// IsFavorite checks if a novel is in user's favorites.
func (h *NovelHandler) IsFavorite(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	isFav, err := h.service.IsFavorite(c.Request.Context(), novelID, userID)
	if err != nil {
		response.InternalError(c, "failed to check favorite")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"is_favorite": isFav})
}

// ListFavorites retrieves a user's favorite novels.
func (h *NovelHandler) ListFavorites(c *gin.Context) {
	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)

	novels, total, err := h.service.ListFavorites(c.Request.Context(), userID, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list favorites")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
		"items":        novels,
	})
}

// AddNovelComment adds a comment to a novel.
func (h *NovelHandler) AddNovelComment(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	comment := &novel.NovelComment{
		NovelID: novelID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := h.service.AddNovelComment(c.Request.Context(), comment); err != nil {
		response.InternalError(c, "failed to add comment")
		return
	}

	response.SuccessResp(c, http.StatusCreated, comment)
}

// ListNovelComments retrieves comments for a novel.
func (h *NovelHandler) ListNovelComments(c *gin.Context) {
	novelID, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	page, limit, skip, lmt := getPaginationParams(c)
	sortOrder := c.DefaultQuery("sort", "newest")

	comments, total, err := h.service.ListNovelComments(c.Request.Context(), novelID, skip, lmt, sortOrder)
	if err != nil {
		response.InternalError(c, "failed to list comments")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
		"items":        comments,
	})
}

// DeleteNovelComment deletes a novel comment.
func (h *NovelHandler) DeleteNovelComment(c *gin.Context) {
	_, err := getNovelID(c)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	commentIDStr := c.Param("comment_id")
	if commentIDStr == "" {
		response.ValidationError(c, "comment id is required")
		return
	}

	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		response.ValidationError(c, "invalid comment id")
		return
	}

	userID, _, err := getCallerObjectID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteNovelComment(c.Request.Context(), commentID, userID); err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.NotFound(c, "comment not found")
			return
		}
		response.InternalError(c, "failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}

// getNovelID extracts and validates novel ID from URL parameters
func getNovelID(c *gin.Context) (primitive.ObjectID, error) {
	id := c.Param("novelID")
	if id == "" {
		return primitive.NilObjectID, errors.New("novel id is required")
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid novel id")
	}
	return oid, nil
}

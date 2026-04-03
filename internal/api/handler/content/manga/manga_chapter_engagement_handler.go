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

func (h *MangaChapterHandler) IncrementChapterViews(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	// For anonymous users, always increment views
	userID, hasUser := c.Get(middleware.ContextUserIDKey)
	incremented := true

	if hasUser {
		if uidStr, ok := userID.(string); ok && uidStr != "" {
			uID, err := primitive.ObjectIDFromHex(uidStr)
			if err == nil {
				// Use atomic operation for logged-in users
				isNewView, err := h.service.TrackAndIncrementChapterView(c.Request.Context(), chID, mID, uID)
				if err != nil {
					response.InternalError(c, "failed to track view")
					return
				}
				incremented = isNewView
			}
		}
	} else {
		// Anonymous users always increment
		if err := h.service.IncrementChapterViews(c.Request.Context(), chID, mID); err != nil {
			response.InternalError(c, "failed to increment views")
			return
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "views processed", "incremented": incremented})
}

func (h *MangaChapterHandler) AddChapterRating(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	var req struct {
		Score float64 `json:"score" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	rating := &manga.ChapterRating{ChapterID: chID, MangaID: mID, UserID: userID, Score: req.Score}

	avgRating, ratingCount, userScore, err := h.service.AddChapterRating(c.Request.Context(), rating)
	if err != nil {
		if errors.Is(err, common.ErrForbidden) {
			response.Forbidden(c, "must view chapter before rating")
			return
		}
		response.InternalError(c, "failed to add rating")
		return
	}

	response.SuccessResp(c, http.StatusCreated, gin.H{
		"average_rating": avgRating,
		"rating_count":   ratingCount,
		"user_score":     userScore,
	})
}

func (h *MangaChapterHandler) AddChapterComment(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
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

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
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
		usernameStr = userID.Hex()
	}

	comment := &manga.ChapterComment{ChapterID: chID, MangaID: mID, UserID: userID, Username: usernameStr, Content: req.Content}
	if err := h.service.AddChapterComment(c.Request.Context(), comment); err != nil {
		response.InternalError(c, "failed to add comment")
		return
	}

	response.SuccessResp(c, http.StatusCreated, comment)
}

func (h *MangaChapterHandler) ListChapterComments(c *gin.Context) {
	chapterID := c.Param("chapterID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
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

	comments, total, err := h.service.ListChapterComments(c.Request.Context(), chID, skip, lmt, sort)
	if err != nil {
		response.InternalError(c, "failed to list comments")
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
			"chapter_id":  comment.ChapterID,
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
	response.SuccessResp(c, http.StatusOK, gin.H{"data": mapped, "total": total, "total_pages": totalPages, "current_page": page, "limit": limit})
}

func (h *MangaChapterHandler) DeleteChapterComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		response.ValidationError(c, "comment id required")
		return
	}

	cID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		response.ValidationError(c, "invalid comment id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteChapterComment(c.Request.Context(), cID, userID); err != nil {
		response.InternalError(c, "failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserChapterRating retrieves the user's rating for a specific chapter
func (h *MangaChapterHandler) GetUserChapterRating(c *gin.Context) {
	chapterID := c.Param("chapterID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id is required")
		return
	}

	cID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	rating, hasRated, err := h.service.GetUserChapterRating(c.Request.Context(), cID, userID)
	if err != nil {
		response.InternalError(c, "failed to get user rating")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"score": rating, "has_rated": hasRated})
}

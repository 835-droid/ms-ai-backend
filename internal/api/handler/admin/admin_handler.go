// Package admin provides admin-related handlers
package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/835-droid/ms-ai-backend/internal/core/admin"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler handles admin endpoints
type Handler struct {
	service admin.Service
}

func NewHandler(svc admin.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) CreateInvite(c *gin.Context) {
	var req struct {
		Length int `json:"length"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request")
		return
	}
	if req.Length <= 0 {
		req.Length = 12
	}
	invite, err := h.service.CreateInviteCode(c.Request.Context(), req.Length)
	if err != nil {
		response.InternalError(c, "failed to create invite code")
		return
	}
	response.SuccessResp(c, http.StatusCreated, gin.H{
		"code":       invite.Code,
		"created_at": invite.CreatedAt,
		"expires_at": invite.ExpiresAt,
	})
}

func (h *Handler) ListInvites(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	invs, total, err := h.service.ListInviteCodes(c.Request.Context(), skip, int64(limit))
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to list invite codes: %v", err))
		return
	}
	response.SuccessResp(c, http.StatusOK, gin.H{"invites": invs, "total": total})
}

func (h *Handler) DeleteInvite(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.ValidationError(c, "missing id")
		return
	}
	if err := h.service.DeleteInviteCode(c.Request.Context(), id); err != nil {
		response.InternalError(c, fmt.Sprintf("failed to delete invite code: %v", err))
		return
	}
	response.SuccessResp(c, http.StatusOK, gin.H{"message": "invite code deleted successfully"})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	// This should be implemented to return system metrics
	metrics := h.service.GetMetrics(c.Request.Context())
	response.SuccessResp(c, http.StatusOK, metrics)
}

func (h *Handler) GetDBMetrics(c *gin.Context) {
	// This should be implemented to return database metrics
	metrics := h.service.GetDBMetrics(c.Request.Context())
	response.SuccessResp(c, http.StatusOK, metrics)
}

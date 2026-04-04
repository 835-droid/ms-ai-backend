// ----- START OF FILE: backend/MS-AI/internal/api/handler/admin/admin_handler.go -----
package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/835-droid/ms-ai-backend/internal/core/admin"
	"github.com/835-droid/ms-ai-backend/pkg/i18n"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

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
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}
	if req.Length <= 0 {
		req.Length = 12
	}
	invite, err := h.service.CreateInviteCode(c.Request.Context(), req.Length)
	if err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
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
		response.InternalError(c, fmt.Sprintf("%s: %v", i18n.TContext(c, i18n.MsgSystemInternalError), err))
		return
	}
	response.SuccessResp(c, http.StatusOK, gin.H{"invites": invs, "total": total})
}

func (h *Handler) DeleteInvite(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}
	if err := h.service.DeleteInviteCode(c.Request.Context(), id); err != nil {
		response.InternalError(c, fmt.Sprintf("%s: %v", i18n.TContext(c, i18n.MsgSystemInternalError), err))
		return
	}
	response.SuccessResp(c, http.StatusOK, gin.H{"message": i18n.TContext(c, i18n.MsgInviteCodeGenerated)})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := h.service.GetMetrics(c.Request.Context())
	response.SuccessResp(c, http.StatusOK, metrics)
}

func (h *Handler) GetDBMetrics(c *gin.Context) {
	metrics := h.service.GetDBMetrics(c.Request.Context())
	response.SuccessResp(c, http.StatusOK, metrics)
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, err := h.service.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) PromoteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationRequired))
		return
	}

	if err := h.service.PromoteToAdmin(c.Request.Context(), userID); err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": i18n.TContext(c, i18n.MsgUserProfileUpdated)})
}

func (h *Handler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationRequired))
		return
	}

	if err := h.service.DeactivateUser(c.Request.Context(), userID); err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": i18n.TContext(c, i18n.MsgUserProfileUpdated)})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationRequired))
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": i18n.TContext(c, i18n.MsgUserProfileUpdated)})
}

func (h *Handler) DemoteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationRequired))
		return
	}

	if err := h.service.DemoteToUser(c.Request.Context(), userID); err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": i18n.TContext(c, i18n.MsgUserProfileUpdated)})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationRequired))
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationMinLength))
		return
	}

	if err := h.service.ChangeUserPassword(c.Request.Context(), userID, req.Password); err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/admin/admin_handler.go -----

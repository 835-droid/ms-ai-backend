// ----- START OF FILE: backend/MS-AI/internal/api/handler/auth/auth_handler.go -----
// Package auth provides authentication related handlerspackage auth

package auth

import (
	"errors"
	"net/http"

	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/pkg/i18n"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Handler handles HTTP requests related to authentication.
type Handler struct {
	Service    coreauth.AuthService
	Translator *i18n.Translator
}

func NewHandler(authSvc coreauth.AuthService) *Handler {
	return &Handler{
		Service:    authSvc,
		Translator: i18n.GetTranslator(),
	}
}

type SignUpRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=30,alphanum"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
	InviteCode string `json:"invite_code" binding:"required,min=8,max=32"`
}

func (h *Handler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}

	res, err := h.Service.SignUp(c.Request.Context(), req.Username, req.Password, req.InviteCode)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUserExists):
			response.ErrorResp(c, http.StatusConflict, i18n.TContext(c, i18n.MsgAuthUserExists))
			return
		case errors.Is(err, core.ErrInvalidInviteCode):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgInviteCodeInvalid))
			return
		case errors.Is(err, core.ErrInviteCodeUsed):
			response.ErrorResp(c, http.StatusConflict, i18n.TContext(c, i18n.MsgInviteCodeUsed))
			return
		case errors.Is(err, core.ErrInviteCodeExpired):
			response.ErrorResp(c, http.StatusBadRequest, i18n.TContext(c, i18n.MsgInviteCodeExpired))
			return
		case errors.Is(err, core.ErrInvalidInput):
			response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
			return
		default:
			response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
			return
		}
	}

	response.SuccessResp(c, http.StatusCreated, gin.H{
		"message":       i18n.TContext(c, i18n.MsgAuthSignupSuccess),
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"user": gin.H{
			"id":       res.User.ID.Hex(),
			"username": res.User.Username,
			"roles":    res.User.Roles,
		},
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}

	res, err := h.Service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidCredentials):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgAuthInvalidCredentials))
			return
		case errors.Is(err, core.ErrUserNotFound):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgUserNotFound))
			return
		case errors.Is(err, core.ErrInternalServer):
			response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
			return
		default:
			response.InternalError(c, i18n.TContext(c, i18n.MsgAuthLoginFailed))
			return
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"message":       i18n.TContext(c, i18n.MsgAuthLoginSuccess),
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"user": gin.H{
			"id":       res.User.ID.Hex(),
			"username": res.User.Username,
			"roles":    res.User.Roles,
		},
	})
}

// RefreshTokenRequest holds refresh token payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles access token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid refresh token format")
		return
	}

	res, err := h.Service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidToken):
			response.Unauthorized(c, "Invalid refresh token")
			return
		case errors.Is(err, core.ErrTokenExpired):
			response.Unauthorized(c, "Refresh token expired")
			return
		default:
			response.InternalError(c, "Failed to refresh token")
			return
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"token_type":    "Bearer",
	})
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidate refresh token and logout user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/logout [post]
// @Security BearerAuth
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from auth middleware
	v, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, i18n.TContext(c, i18n.MsgAuthAccessDenied))
		return
	}

	uidStr, ok := v.(string)
	if !ok {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	id, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.InternalError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}

	if err := h.Service.Logout(c.Request.Context(), id); err != nil {
		switch {
		case errors.Is(err, core.ErrUserNotFound):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgUserNotFound))
			return
		default:
			response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
			return
		}
	}

	// Clear any session cookies if used
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	response.SuccessResp(c, http.StatusOK, gin.H{
		"message": i18n.TContext(c, i18n.MsgAuthLogoutSuccess),
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, i18n.TContext(c, i18n.MsgAuthAccessDenied))
		return
	}
	uidStr, ok := v.(string)
	if !ok {
		response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, i18n.TContext(c, i18n.MsgValidationInvalidFormat))
		return
	}

	if err := h.Service.ChangePassword(c.Request.Context(), uidStr, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidCredentials):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgAuthInvalidCredentials))
			return
		case errors.Is(err, core.ErrUserNotFound):
			response.Unauthorized(c, i18n.TContext(c, i18n.MsgUserNotFound))
			return
		default:
			response.InternalError(c, i18n.TContext(c, i18n.MsgSystemInternalError))
			return
		}
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/auth/auth_handler.go -----

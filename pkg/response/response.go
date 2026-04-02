// ----- START OF FILE: backend/MS-AI/pkg/response/response.go -----
package response

import (
	"net/http"

	"github.com/835-droid/ms-ai-backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func write(c *gin.Context, status int, resp Response) {
	if rid, ok := c.Get("request_id"); ok {
		if s, ok := rid.(string); ok {
			resp.RequestID = s
		}
	}
	c.JSON(status, resp)
}

func SuccessResp(c *gin.Context, status int, data interface{}) {
	write(c, status, Response{Success: true, Data: data})
}

func ErrorResp(c *gin.Context, status int, message string) {
	write(c, status, Response{Success: false, Error: message})
}

func ValidationError(c *gin.Context, message string) {
	ErrorResp(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	ErrorResp(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	ErrorResp(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	ErrorResp(c, http.StatusNotFound, message)
}

func InternalError(c *gin.Context, message string) {
	ErrorResp(c, http.StatusInternalServerError, message)
}

// DomainError translates a domain error into a standard HTTP response
func DomainError(c *gin.Context, err *errors.DomainError) {
	if err == nil {
		return
	}
	write(c, err.HTTPStatus, Response{Success: false, Error: err.Message, Code: string(err.Code)})
}

// ----- END OF FILE: backend/MS-AI/pkg/response/response.go -----

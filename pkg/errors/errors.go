// ----- START OF FILE: backend/MS-AI/pkg/errors/errors.go -----
// pkg/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Authentication errors
	ErrCodeAuthenticationFailed ErrorCode = "AUTH_FAILED"
	ErrCodeInvalidCredentials   ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired         ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid         ErrorCode = "TOKEN_INVALID"

	// Authorization errors
	ErrCodeAccessDenied     ErrorCode = "ACCESS_DENIED"
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// User errors
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists       ErrorCode = "USER_EXISTS"
	ErrCodeInvalidUserInput ErrorCode = "INVALID_USER_INPUT"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"

	// Resource errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// System errors
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError  ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPI    ErrorCode = "EXTERNAL_API_ERROR"
)

// DomainError represents a structured application error
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// New creates a new domain error
func New(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
	}
}

// NewWithCause creates a new domain error with a cause
func NewWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Cause:      cause,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeValidationFailed,
		Message: "Validation failed",
		Details: map[string]interface{}{
			"field":   field,
			"message": message,
		},
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeAuthenticationFailed,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeAccessDenied,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *DomainError {
	return &DomainError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{
			"resource": resource,
		},
		HTTPStatus: http.StatusNotFound,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, cause error) *DomainError {
	return &DomainError{
		Code:       ErrCodeInternalServer,
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// getHTTPStatus maps error codes to HTTP status codes
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeAuthenticationFailed, ErrCodeInvalidCredentials, ErrCodeTokenExpired, ErrCodeTokenInvalid:
		return http.StatusUnauthorized
	case ErrCodeAccessDenied, ErrCodeInsufficientRole:
		return http.StatusForbidden
	case ErrCodeUserNotFound, ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUserExists, ErrCodeAlreadyExists:
		return http.StatusConflict
	case ErrCodeValidationFailed, ErrCodeInvalidUserInput:
		return http.StatusBadRequest
	case ErrCodeInternalServer, ErrCodeDatabaseError, ErrCodeExternalAPI:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ----- END OF FILE: backend/MS-AI/pkg/errors/errors.go -----

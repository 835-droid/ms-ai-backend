package core

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInviteCode  = errors.New("invalid invite code")
	ErrInviteCodeExists   = errors.New("invite code already exists")
	ErrInviteCodeUsed     = errors.New("invite code already used")
	ErrInviteCodeExpired  = errors.New("invite code expired")
	ErrInviteCodeNotFound = errors.New("invite code not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotFound           = errors.New("not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrSlugExists         = errors.New("slug already exists")
	// Manga domain errors
	ErrMangaExists          = errors.New("manga already exists")
	ErrMangaNotFound        = errors.New("manga not found")
	ErrMangaChapterExists   = errors.New("manga chapter already exists")
	ErrMangaChapterNotFound = errors.New("manga chapter not found")
)

// AppError is a rich application error with HTTP mapping.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func NewAppError(code, message string, httpStatus int, err error) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, Err: err}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// IsNotFoundError checks whether an error is a domain not-found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrMangaNotFound) || errors.Is(err, ErrMangaChapterNotFound)
}

// IsExistsError checks whether an error indicates an already-exists condition
func IsExistsError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrUserExists) || errors.Is(err, ErrMangaExists) || errors.Is(err, ErrMangaChapterExists)
}

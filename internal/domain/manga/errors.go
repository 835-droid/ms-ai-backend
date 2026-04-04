package manga

import "errors"

// Domain errors for manga operations.
var (
	ErrMangaNotFound      = errors.New("manga not found")
	ErrChapterNotFound    = errors.New("chapter not found")
	ErrListNotFound       = errors.New("favorite list not found")
	ErrListNotOwned       = errors.New("you don't have access to this list")
	ErrDuplicateList      = errors.New("list with same name already exists")
	ErrMangaNotInList     = errors.New("manga not in list")
	ErrInvalidReaction    = errors.New("invalid reaction type")
	ErrAlreadyRated       = errors.New("user has already rated this item")
	ErrInvalidRatingScore = errors.New("rating score must be between 1 and 10")
)

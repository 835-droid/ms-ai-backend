package novel

import "errors"

// Domain errors for novel operations.
var (
	ErrNovelNotFound      = errors.New("novel not found")
	ErrChapterNotFound    = errors.New("chapter not found")
	ErrListNotFound       = errors.New("favorite list not found")
	ErrListNotOwned       = errors.New("you don't have access to this list")
	ErrDuplicateList      = errors.New("list with same name already exists")
	ErrNovelNotInList     = errors.New("novel not in list")
	ErrInvalidReaction    = errors.New("invalid reaction type")
	ErrAlreadyRated       = errors.New("user has already rated this item")
	ErrInvalidRatingScore = errors.New("rating score must be between 1 and 10")
)

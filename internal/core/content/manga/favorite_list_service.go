package manga

import (
	"context"
	"errors"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

var (
	ErrListNotFound   = errors.New("favorite list not found")
	ErrListNotOwned   = errors.New("user does not own this list")
	ErrMangaNotFound  = errors.New("manga not found")
	ErrMangaNotInList = errors.New("manga not in list")
	ErrDuplicateList  = errors.New("list with same name already exists")
)

// FavoriteListService defines favorite list business logic operations.
type FavoriteListService interface {
	// List operations
	CreateList(ctx context.Context, userID, name, description string, isPublic bool) (*FavoriteList, error)
	GetList(ctx context.Context, listID, userID string) (*FavoriteList, error)
	ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*FavoriteList, int64, error)
	UpdateList(ctx context.Context, listID, userID, name, description string, isPublic bool) (*FavoriteList, error)
	DeleteList(ctx context.Context, listID, userID string) error

	// List item operations
	AddMangaToList(ctx context.Context, listID, userID, mangaID, notes string) error
	RemoveMangaFromList(ctx context.Context, listID, userID, mangaID string) error
	ListMangaInList(ctx context.Context, listID, userID string, skip, limit int64) ([]*FavoriteListItem, int64, error)
	UpdateListItemNotes(ctx context.Context, listID, userID, mangaID, notes string) error
	MoveMangaToList(ctx context.Context, fromListID, toListID, userID, mangaID string) error

	// Query operations
	GetUserMangaLists(ctx context.Context, userID, mangaID string) ([]*FavoriteList, error)
	GetPublicListManga(ctx context.Context, listID string, skip, limit int64) ([]*Manga, int64, error)
}

// DefaultFavoriteListService implements FavoriteListService.
type DefaultFavoriteListService struct {
	repo      FavoriteListRepository
	mangaRepo MangaRepository
	log       *logger.Logger
}

// NewFavoriteListService creates a new DefaultFavoriteListService.
func NewFavoriteListService(repo FavoriteListRepository, mangaRepo MangaRepository, log *logger.Logger) *DefaultFavoriteListService {
	return &DefaultFavoriteListService{repo: repo, mangaRepo: mangaRepo, log: log}
}

// CreateList creates a new favorite list for a user.
func (s *DefaultFavoriteListService) CreateList(ctx context.Context, userID, name, description string, isPublic bool) (*FavoriteList, error) {
	s.log.Debug("creating favorite list", map[string]interface{}{
		"user_id":   userID,
		"name":      name,
		"is_public": isPublic,
	})

	// Check if list with same name exists
	existing, err := s.repo.GetListByName(ctx, userID, name)
	if err != nil {
		s.log.Error("failed to check existing list", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateList
	}

	list := &FavoriteList{
		UserID:      userID,
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateList(ctx, list); err != nil {
		s.log.Error("failed to create favorite list", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	s.log.Info("favorite list created", map[string]interface{}{
		"list_id": list.ID,
		"name":    list.Name,
	})
	return list, nil
}

// GetList retrieves a favorite list by ID, verifying ownership.
func (s *DefaultFavoriteListService) GetList(ctx context.Context, listID, userID string) (*FavoriteList, error) {
	s.log.Debug("getting favorite list", map[string]interface{}{
		"list_id": listID,
		"user_id": userID,
	})

	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		s.log.Error("failed to get favorite list", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return nil, err
	}

	if list == nil {
		return nil, ErrListNotFound
	}

	if list.UserID != userID {
		return nil, ErrListNotOwned
	}

	// Get manga count
	count, err := s.repo.GetListMangaCount(ctx, listID)
	if err != nil {
		s.log.Warn("failed to get list manga count", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
	} else {
		list.MangaCount = count
	}

	return list, nil
}

// ListUserLists retrieves all favorite lists for a user.
func (s *DefaultFavoriteListService) ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*FavoriteList, int64, error) {
	s.log.Debug("listing user favorite lists", map[string]interface{}{
		"user_id": userID,
		"skip":    skip,
		"limit":   limit,
	})

	lists, total, err := s.repo.ListUserLists(ctx, userID, skip, limit)
	if err != nil {
		s.log.Error("failed to list favorite lists", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, 0, err
	}

	// Get manga count for each list
	for _, list := range lists {
		count, err := s.repo.GetListMangaCount(ctx, list.ID)
		if err != nil {
			s.log.Warn("failed to get list manga count", map[string]interface{}{
				"error":   err.Error(),
				"list_id": list.ID,
			})
		} else {
			list.MangaCount = count
		}
	}

	return lists, total, nil
}

// UpdateList updates a favorite list.
func (s *DefaultFavoriteListService) UpdateList(ctx context.Context, listID, userID, name, description string, isPublic bool) (*FavoriteList, error) {
	s.log.Debug("updating favorite list", map[string]interface{}{
		"list_id": listID,
		"user_id": userID,
		"name":    name,
	})

	// Get existing list
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		s.log.Error("failed to get list for update", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return nil, err
	}

	if list == nil {
		return nil, ErrListNotFound
	}

	if list.UserID != userID {
		return nil, ErrListNotOwned
	}

	// Check if new name conflicts with existing
	if name != list.Name {
		existing, err := s.repo.GetListByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != listID {
			return nil, ErrDuplicateList
		}
	}

	list.Name = name
	list.Description = description
	list.IsPublic = isPublic

	if err := s.repo.UpdateList(ctx, list); err != nil {
		s.log.Error("failed to update favorite list", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return nil, err
	}

	s.log.Info("favorite list updated", map[string]interface{}{
		"list_id": listID,
		"name":    name,
	})
	return list, nil
}

// DeleteList deletes a favorite list.
func (s *DefaultFavoriteListService) DeleteList(ctx context.Context, listID, userID string) error {
	s.log.Debug("deleting favorite list", map[string]interface{}{
		"list_id": listID,
		"user_id": userID,
	})

	// Verify ownership
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list == nil {
		return ErrListNotFound
	}

	if list.UserID != userID {
		return ErrListNotOwned
	}

	if err := s.repo.DeleteList(ctx, listID, userID); err != nil {
		s.log.Error("failed to delete favorite list", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return err
	}

	s.log.Info("favorite list deleted", map[string]interface{}{
		"list_id": listID,
	})
	return nil
}

// AddMangaToList adds a manga to a favorite list.
func (s *DefaultFavoriteListService) AddMangaToList(ctx context.Context, listID, userID, mangaID, notes string) error {
	s.log.Debug("adding manga to list", map[string]interface{}{
		"list_id":  listID,
		"user_id":  userID,
		"manga_id": mangaID,
	})

	// Verify list ownership
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list == nil {
		return ErrListNotFound
	}

	if list.UserID != userID {
		return ErrListNotOwned
	}

	// Note: We skip manga existence check as it will fail gracefully if manga doesn't exist
	// The repository handles this with proper error messages

	item := &FavoriteListItem{
		ListID:    listID,
		MangaID:   mangaID,
		Notes:     notes,
		AddedAt:   time.Now(),
		SortOrder: 0,
	}

	if err := s.repo.AddMangaToList(ctx, item); err != nil {
		s.log.Error("failed to add manga to list", map[string]interface{}{
			"error":    err.Error(),
			"list_id":  listID,
			"manga_id": mangaID,
		})
		return err
	}

	s.log.Info("manga added to list", map[string]interface{}{
		"list_id":  listID,
		"manga_id": mangaID,
	})
	return nil
}

// RemoveMangaFromList removes a manga from a favorite list.
func (s *DefaultFavoriteListService) RemoveMangaFromList(ctx context.Context, listID, userID, mangaID string) error {
	s.log.Debug("removing manga from list", map[string]interface{}{
		"list_id":  listID,
		"user_id":  userID,
		"manga_id": mangaID,
	})

	// Verify list ownership
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list == nil {
		return ErrListNotFound
	}

	if list.UserID != userID {
		return ErrListNotOwned
	}

	// Check if manga is in list
	inList, err := s.repo.IsMangaInList(ctx, listID, mangaID)
	if err != nil {
		return err
	}

	if !inList {
		return ErrMangaNotInList
	}

	if err := s.repo.RemoveMangaFromList(ctx, listID, mangaID); err != nil {
		s.log.Error("failed to remove manga from list", map[string]interface{}{
			"error":    err.Error(),
			"list_id":  listID,
			"manga_id": mangaID,
		})
		return err
	}

	s.log.Info("manga removed from list", map[string]interface{}{
		"list_id":  listID,
		"manga_id": mangaID,
	})
	return nil
}

// ListMangaInList retrieves all manga in a specific list.
func (s *DefaultFavoriteListService) ListMangaInList(ctx context.Context, listID, userID string, skip, limit int64) ([]*FavoriteListItem, int64, error) {
	s.log.Debug("listing manga in list", map[string]interface{}{
		"list_id": listID,
		"user_id": userID,
		"skip":    skip,
		"limit":   limit,
	})

	// Verify list ownership or check if public
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		return nil, 0, err
	}

	if list == nil {
		return nil, 0, ErrListNotFound
	}

	if list.UserID != userID && !list.IsPublic {
		return nil, 0, ErrListNotOwned
	}

	items, total, err := s.repo.ListMangaInList(ctx, listID, skip, limit)
	if err != nil {
		s.log.Error("failed to list manga in list", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateListItemNotes updates notes for a manga in a list.
func (s *DefaultFavoriteListService) UpdateListItemNotes(ctx context.Context, listID, userID, mangaID, notes string) error {
	s.log.Debug("updating list item notes", map[string]interface{}{
		"list_id":  listID,
		"user_id":  userID,
		"manga_id": mangaID,
	})

	// Verify list ownership
	list, err := s.repo.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list == nil {
		return ErrListNotFound
	}

	if list.UserID != userID {
		return ErrListNotOwned
	}

	if err := s.repo.UpdateListItemNotes(ctx, listID, mangaID, notes); err != nil {
		s.log.Error("failed to update list item notes", map[string]interface{}{
			"error":    err.Error(),
			"list_id":  listID,
			"manga_id": mangaID,
		})
		return err
	}

	s.log.Info("list item notes updated", map[string]interface{}{
		"list_id":  listID,
		"manga_id": mangaID,
	})
	return nil
}

// MoveMangaToList moves a manga from one list to another.
func (s *DefaultFavoriteListService) MoveMangaToList(ctx context.Context, fromListID, toListID, userID, mangaID string) error {
	s.log.Debug("moving manga between lists", map[string]interface{}{
		"from_list_id": fromListID,
		"to_list_id":   toListID,
		"user_id":      userID,
		"manga_id":     mangaID,
	})

	// Verify ownership of both lists
	fromList, err := s.repo.GetListByID(ctx, fromListID)
	if err != nil {
		return err
	}

	if fromList == nil || fromList.UserID != userID {
		return ErrListNotOwned
	}

	toList, err := s.repo.GetListByID(ctx, toListID)
	if err != nil {
		return err
	}

	if toList == nil || toList.UserID != userID {
		return ErrListNotOwned
	}

	if err := s.repo.MoveMangaToList(ctx, fromListID, toListID, mangaID); err != nil {
		s.log.Error("failed to move manga between lists", map[string]interface{}{
			"error":        err.Error(),
			"from_list_id": fromListID,
			"to_list_id":   toListID,
			"manga_id":     mangaID,
		})
		return err
	}

	s.log.Info("manga moved between lists", map[string]interface{}{
		"from_list_id": fromListID,
		"to_list_id":   toListID,
		"manga_id":     mangaID,
	})
	return nil
}

// GetUserMangaLists retrieves all lists that contain a specific manga for a user.
func (s *DefaultFavoriteListService) GetUserMangaLists(ctx context.Context, userID, mangaID string) ([]*FavoriteList, error) {
	s.log.Debug("getting user manga lists", map[string]interface{}{
		"user_id":  userID,
		"manga_id": mangaID,
	})

	lists, err := s.repo.GetUserMangaLists(ctx, userID, mangaID)
	if err != nil {
		s.log.Error("failed to get user manga lists", map[string]interface{}{
			"error":    err.Error(),
			"user_id":  userID,
			"manga_id": mangaID,
		})
		return nil, err
	}

	// Get manga count for each list
	for _, list := range lists {
		count, err := s.repo.GetListMangaCount(ctx, list.ID)
		if err != nil {
			s.log.Warn("failed to get list manga count", map[string]interface{}{
				"error":   err.Error(),
				"list_id": list.ID,
			})
		} else {
			list.MangaCount = count
		}
	}

	return lists, nil
}

// GetPublicListManga retrieves manga from a public list.
func (s *DefaultFavoriteListService) GetPublicListManga(ctx context.Context, listID string, skip, limit int64) ([]*Manga, int64, error) {
	s.log.Debug("getting public list manga", map[string]interface{}{
		"list_id": listID,
		"skip":    skip,
		"limit":   limit,
	})

	mangas, total, err := s.repo.GetPublicListManga(ctx, listID, skip, limit)
	if err != nil {
		s.log.Error("failed to get public list manga", map[string]interface{}{
			"error":   err.Error(),
			"list_id": listID,
		})
		return nil, 0, err
	}

	return mangas, total, nil
}

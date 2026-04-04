package manga

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresFavoriteListRepository struct {
	db *pginfra.PostgresStore
}

func NewPostgresFavoriteListRepository(p *pginfra.PostgresStore) manga.FavoriteListRepository {
	return &postgresFavoriteListRepository{db: p}
}

// helper methods to access sqlx.DB methods through PostgresStore
func (r *postgresFavoriteListRepository) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.db.DB.ExecContext(ctx, query, args...)
}

func (r *postgresFavoriteListRepository) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.db.DB.QueryRowContext(ctx, query, args...)
}

func (r *postgresFavoriteListRepository) queryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return r.db.DB.QueryxContext(ctx, query, args...)
}

func (r *postgresFavoriteListRepository) beginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.DB.BeginTxx(ctx, nil)
}

// CreateList creates a new favorite list
func (r *postgresFavoriteListRepository) CreateList(ctx context.Context, list *manga.FavoriteList) error {
	if list.ID == "" {
		list.ID = uuid.New().String()
	}
	query := `INSERT INTO favorite_lists (id, user_id, name, description, is_public, sort_order, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.execContext(ctx, query, list.ID, list.UserID, list.Name, list.Description, list.IsPublic, list.SortOrder, list.CreatedAt, list.UpdatedAt)
	return err
}

// GetListByID retrieves a favorite list by ID
func (r *postgresFavoriteListRepository) GetListByID(ctx context.Context, listID string) (*manga.FavoriteList, error) {
	query := `SELECT id, user_id, name, description, is_public, sort_order, created_at, updated_at 
			  FROM favorite_lists WHERE id = $1`
	list := &manga.FavoriteList{}
	err := r.queryRowContext(ctx, query, listID).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic, &list.SortOrder, &list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

// GetListByName retrieves a favorite list by user ID and name
func (r *postgresFavoriteListRepository) GetListByName(ctx context.Context, userID, name string) (*manga.FavoriteList, error) {
	query := `SELECT id, user_id, name, description, is_public, sort_order, created_at, updated_at 
			  FROM favorite_lists WHERE user_id = $1 AND name = $2`
	list := &manga.FavoriteList{}
	err := r.queryRowContext(ctx, query, userID, name).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic, &list.SortOrder, &list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

// ListUserLists retrieves all favorite lists for a user
func (r *postgresFavoriteListRepository) ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*manga.FavoriteList, int64, error) {
	countQuery := `SELECT COUNT(*) FROM favorite_lists WHERE user_id = $1`
	var total int64
	err := r.queryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, name, description, is_public, sort_order, created_at, updated_at 
			  FROM favorite_lists WHERE user_id = $1 ORDER BY sort_order, created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.queryContext(ctx, query, userID, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lists []*manga.FavoriteList
	for rows.Next() {
		list := &manga.FavoriteList{}
		err := rows.Scan(&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic, &list.SortOrder, &list.CreatedAt, &list.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		lists = append(lists, list)
	}
	return lists, total, rows.Err()
}

// UpdateList updates a favorite list
func (r *postgresFavoriteListRepository) UpdateList(ctx context.Context, list *manga.FavoriteList) error {
	list.UpdatedAt = time.Now()
	query := `UPDATE favorite_lists SET name = $1, description = $2, is_public = $3, sort_order = $4, updated_at = $5 
			  WHERE id = $6 AND user_id = $7`
	_, err := r.execContext(ctx, query, list.Name, list.Description, list.IsPublic, list.SortOrder, list.UpdatedAt, list.ID, list.UserID)
	return err
}

// DeleteList deletes a favorite list
func (r *postgresFavoriteListRepository) DeleteList(ctx context.Context, listID, userID string) error {
	query := `DELETE FROM favorite_lists WHERE id = $1 AND user_id = $2`
	_, err := r.execContext(ctx, query, listID, userID)
	return err
}

// GetListMangaCount returns the number of manga in a list
func (r *postgresFavoriteListRepository) GetListMangaCount(ctx context.Context, listID string) (int64, error) {
	query := `SELECT COUNT(*) FROM favorite_list_items WHERE list_id = $1`
	var count int64
	err := r.queryRowContext(ctx, query, listID).Scan(&count)
	return count, err
}

// AddMangaToList adds a manga to a favorite list
func (r *postgresFavoriteListRepository) AddMangaToList(ctx context.Context, item *manga.FavoriteListItem) error {
	query := `INSERT INTO favorite_list_items (list_id, manga_id, notes, added_at, sort_order) 
			  VALUES ($1, $2, $3, $4, $5) ON CONFLICT (list_id, manga_id) DO UPDATE SET notes = EXCLUDED.notes`
	_, err := r.execContext(ctx, query, item.ListID, item.MangaID, item.Notes, item.AddedAt, item.SortOrder)
	return err
}

// RemoveMangaFromList removes a manga from a favorite list
func (r *postgresFavoriteListRepository) RemoveMangaFromList(ctx context.Context, listID, mangaID string) error {
	query := `DELETE FROM favorite_list_items WHERE list_id = $1 AND manga_id = $2`
	_, err := r.execContext(ctx, query, listID, mangaID)
	return err
}

// IsMangaInList checks if a manga is in a specific list
func (r *postgresFavoriteListRepository) IsMangaInList(ctx context.Context, listID, mangaID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorite_list_items WHERE list_id = $1 AND manga_id = $2)`
	var exists bool
	err := r.queryRowContext(ctx, query, listID, mangaID).Scan(&exists)
	return exists, err
}

// ListMangaInList retrieves all manga in a specific list
func (r *postgresFavoriteListRepository) ListMangaInList(ctx context.Context, listID string, skip, limit int64) ([]*manga.FavoriteListItem, int64, error) {
	countQuery := `SELECT COUNT(*) FROM favorite_list_items WHERE list_id = $1`
	var total int64
	err := r.queryRowContext(ctx, countQuery, listID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT list_id, manga_id, notes, added_at, sort_order 
			  FROM favorite_list_items WHERE list_id = $1 ORDER BY sort_order, added_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.queryContext(ctx, query, listID, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*manga.FavoriteListItem
	for rows.Next() {
		item := &manga.FavoriteListItem{}
		err := rows.Scan(&item.ListID, &item.MangaID, &item.Notes, &item.AddedAt, &item.SortOrder)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// UpdateListItemNotes updates notes for a manga in a list
func (r *postgresFavoriteListRepository) UpdateListItemNotes(ctx context.Context, listID, mangaID, notes string) error {
	query := `UPDATE favorite_list_items SET notes = $1 WHERE list_id = $2 AND manga_id = $3`
	_, err := r.execContext(ctx, query, notes, listID, mangaID)
	return err
}

// MoveMangaToList moves a manga from one list to another
func (r *postgresFavoriteListRepository) MoveMangaToList(ctx context.Context, fromListID, toListID, mangaID string) error {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the item from the source list
	getQuery := `SELECT notes, sort_order FROM favorite_list_items WHERE list_id = $1 AND manga_id = $2`
	var notes string
	var sortOrder int
	err = tx.QueryRowContext(ctx, getQuery, fromListID, mangaID).Scan(&notes, &sortOrder)
	if err != nil {
		return err
	}

	// Insert into the target list (or update if exists)
	upsertQuery := `INSERT INTO favorite_list_items (list_id, manga_id, notes, added_at, sort_order) 
					VALUES ($1, $2, $3, $4, $5) 
					ON CONFLICT (list_id, manga_id) DO UPDATE SET notes = EXCLUDED.notes, sort_order = EXCLUDED.sort_order`
	_, err = tx.ExecContext(ctx, upsertQuery, toListID, mangaID, notes, time.Now(), sortOrder)
	if err != nil {
		return err
	}

	// Delete from source list
	deleteQuery := `DELETE FROM favorite_list_items WHERE list_id = $1 AND manga_id = $2`
	_, err = tx.ExecContext(ctx, deleteQuery, fromListID, mangaID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateItemSortOrder updates the sort order of a manga in a list
func (r *postgresFavoriteListRepository) UpdateItemSortOrder(ctx context.Context, listID, mangaID string, sortOrder int) error {
	query := `UPDATE favorite_list_items SET sort_order = $1 WHERE list_id = $2 AND manga_id = $3`
	_, err := r.execContext(ctx, query, sortOrder, listID, mangaID)
	return err
}

// GetUserMangaLists retrieves all lists that contain a specific manga for a user
func (r *postgresFavoriteListRepository) GetUserMangaLists(ctx context.Context, userID, mangaID string) ([]*manga.FavoriteList, error) {
	query := `SELECT fl.id, fl.user_id, fl.name, fl.description, fl.is_public, fl.sort_order, fl.created_at, fl.updated_at 
			  FROM favorite_lists fl
			  INNER JOIN favorite_list_items fli ON fl.id = fli.list_id
			  WHERE fl.user_id = $1 AND fli.manga_id = $2
			  ORDER BY fl.sort_order, fl.created_at DESC`
	rows, err := r.queryContext(ctx, query, userID, mangaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*manga.FavoriteList
	for rows.Next() {
		list := &manga.FavoriteList{}
		err := rows.Scan(&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic, &list.SortOrder, &list.CreatedAt, &list.UpdatedAt)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, rows.Err()
}

// GetPublicListManga retrieves manga from a public list
func (r *postgresFavoriteListRepository) GetPublicListManga(ctx context.Context, listID string, skip, limit int64) ([]*manga.Manga, int64, error) {
	// First check if list is public
	var isPublic bool
	err := r.queryRowContext(ctx, `SELECT is_public FROM favorite_lists WHERE id = $1`, listID).Scan(&isPublic)
	if err != nil || !isPublic {
		return nil, 0, errors.New("list not found or not public")
	}

	countQuery := `SELECT COUNT(*) FROM favorite_list_items WHERE list_id = $1`
	var total int64
	err = r.queryRowContext(ctx, countQuery, listID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT m.id, m.title, m.slug, m.description, m.cover_image, m.views_count, m.likes_count, m.favorites_count, m.average_rating
			  FROM mangas m
			  INNER JOIN favorite_list_items fli ON m.id = fli.manga_id
			  WHERE fli.list_id = $1
			  ORDER BY fli.sort_order, fli.added_at DESC
			  LIMIT $2 OFFSET $3`
	rows, err := r.queryContext(ctx, query, listID, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var mangas []*manga.Manga
	for rows.Next() {
		manga := &manga.Manga{}
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Slug, &manga.Description, &manga.CoverImage, &manga.ViewsCount, &manga.LikesCount, &manga.FavoritesCount, &manga.AverageRating)
		if err != nil {
			return nil, 0, err
		}
		mangas = append(mangas, manga)
	}
	return mangas, total, rows.Err()
}

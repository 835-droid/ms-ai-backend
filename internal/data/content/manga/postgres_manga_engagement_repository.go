package manga

import (
	"context"
	"fmt"
	"time"

	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ========== ENGAGEMENT METHODS ==========

func (r *PostgresMangaRepository) AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO user_favorites (manga_id, user_id, created_at) VALUES ($1, $2, $3)
		 ON CONFLICT (manga_id, user_id) DO NOTHING`,
		mangaID.Hex(), userID.Hex(), time.Now())
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		_, err = tx.ExecContext(ctx,
			`UPDATE mangas SET favorites_count = favorites_count + 1 WHERE id = $1`,
			mangaID.Hex())
		if err != nil {
			return fmt.Errorf("update favorites count: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresMangaRepository) RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`DELETE FROM user_favorites WHERE manga_id=$1 AND user_id=$2`,
		mangaID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		_, err = tx.ExecContext(ctx,
			`UPDATE mangas SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = $1`,
			mangaID.Hex())
		if err != nil {
			return fmt.Errorf("update favorites count: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresMangaRepository) IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga repo not initialized")
	}
	var exists bool
	err := r.store.DB.GetContext(ctx,
		&exists,
		`SELECT EXISTS(SELECT 1 FROM user_favorites WHERE manga_id=$1 AND user_id=$2)`,
		mangaID.Hex(), userID.Hex())
	if err != nil {
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return exists, nil
}

func (r *PostgresMangaRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga repo not initialized")
	}

	// Get total count
	var total int64
	err := r.store.DB.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM user_favorites WHERE user_id=$1`,
		userID.Hex())
	if err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	// Get paginated favorites
	rows, err := r.store.DB.QueryxContext(ctx,
		`SELECT m.id, m.title, m.slug, m.description, m.author_id, m.tags, m.cover_image,
				m.is_published, m.published_at, m.created_at, m.updated_at, m.views_count, m.likes_count, m.favorites_count,
				m.rating_sum, m.rating_count, m.average_rating, m.reactions_count
		 FROM mangas m
		 INNER JOIN user_favorites uf ON m.id = uf.manga_id
		 WHERE uf.user_id=$1
		 ORDER BY uf.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		manga, err := r.scanManga(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan favorite manga: %w", err)
		}
		mangas = append(mangas, manga)
	}

	return mangas, total, nil
}

func (r *PostgresMangaRepository) AddMangaComment(ctx context.Context, comment *coremanga.MangaComment) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO manga_comments (id, manga_id, user_id, username, content, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		comment.ID.Hex(), comment.MangaID.Hex(), comment.UserID.Hex(), comment.Username, comment.Content, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add manga comment: %w", err)
	}
	return nil
}

func (r *PostgresMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.MangaComment, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga repo not initialized")
	}

	// Get total count
	var total int64
	err := r.store.DB.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM manga_comments WHERE manga_id=$1`,
		mangaID.Hex())
	if err != nil {
		return nil, 0, fmt.Errorf("count manga comments: %w", err)
	}

	// Determine sort order
	orderBy := "created_at DESC"
	if sortOrder == "oldest" {
		orderBy = "created_at ASC"
	}

	// Get paginated comments
	rows, err := r.store.DB.QueryxContext(ctx,
		fmt.Sprintf(`SELECT id, manga_id, user_id, username, content, created_at, updated_at
		 FROM manga_comments
		 WHERE manga_id=$1
		 ORDER BY %s
		 LIMIT $2 OFFSET $3`, orderBy),
		mangaID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list manga comments: %w", err)
	}
	defer rows.Close()

	var comments []*coremanga.MangaComment
	for rows.Next() {
		comment := &coremanga.MangaComment{}
		var id, mangaID, userID string
		err := rows.Scan(&id, &mangaID, &userID, &comment.Username, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan manga comment: %w", err)
		}
		comment.ID, _ = primitive.ObjectIDFromHex(id)
		comment.MangaID, _ = primitive.ObjectIDFromHex(mangaID)
		comment.UserID, _ = primitive.ObjectIDFromHex(userID)
		comments = append(comments, comment)
	}

	return comments, total, nil
}

func (r *PostgresMangaRepository) DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	result, err := r.store.DB.ExecContext(ctx,
		`DELETE FROM manga_comments WHERE id=$1 AND user_id=$2`,
		commentID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("delete manga comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("manga comment not found or unauthorized")
	}

	return nil
}

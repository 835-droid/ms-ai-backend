package manga

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ========== REACTION METHODS ==========

func (r *PostgresMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (string, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return "", fmt.Errorf("postgres manga repo not initialized")
	}

	// Per-(manga,user) in-memory guard for anti-spam burst protections
	lockKey := fmt.Sprintf("%s_%s", mangaID.Hex(), userID.Hex())
	if _, loaded := r.reactionLock.LoadOrStore(lockKey, true); loaded {
		return "", fmt.Errorf("reaction request already in progress")
	}
	defer r.reactionLock.Delete(lockKey)

	// Use transaction for atomic updates
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingType string
	queryCheck := `SELECT reaction_type FROM manga_likes WHERE manga_id=$1 AND user_id=$2`
	row := tx.QueryRowContext(ctx, queryCheck, mangaID.Hex(), userID.Hex())
	err = row.Scan(&existingType)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("check reaction exists: %w", err)
	}

	if existingType == string(reactionType) {
		// Same reaction, remove it (toggle off)
		_, err := tx.ExecContext(ctx, `DELETE FROM manga_likes WHERE manga_id=$1 AND user_id=$2`, mangaID.Hex(), userID.Hex())
		if err != nil {
			return "", fmt.Errorf("remove reaction: %w", err)
		}
		// Decrement reactions_count for this type
		_, err = tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(reactions_count, ARRAY[$2], GREATEST((COALESCE((reactions_count->$2)::text, '0')::int - 1), 0)::text::jsonb) WHERE id=$1`, mangaID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("decrement reactions count: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return "", fmt.Errorf("commit reaction removal: %w", err)
		}
		return "", nil // Removed
	} else if existingType != "" {
		// Different reaction, update it
		_, err := tx.ExecContext(ctx, `UPDATE manga_likes SET reaction_type=$3 WHERE manga_id=$1 AND user_id=$2`, mangaID.Hex(), userID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("update reaction: %w", err)
		}
		// Update reactions_count: decrement old, increment new
		result, err := tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(jsonb_set(reactions_count, ARRAY[$2], (COALESCE((reactions_count->$2)::text, '0')::int - 1)::text::jsonb), ARRAY[$3], (COALESCE((reactions_count->$3)::text, '0')::int + 1)::text::jsonb) WHERE id=$1`, mangaID.Hex(), existingType, string(reactionType))
		if err != nil {
			return "", fmt.Errorf("update reactions count: %w", err)
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			return "", corecommon.ErrNotFound
		}
		if err := tx.Commit(); err != nil {
			return "", fmt.Errorf("commit reaction update: %w", err)
		}
		return string(reactionType), nil
	} else {
		// No reaction, add new one
		_, err := tx.ExecContext(ctx, `INSERT INTO manga_likes (manga_id, user_id, reaction_type) VALUES ($1, $2, $3)`, mangaID.Hex(), userID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("insert reaction: %w", err)
		}
		// Increment reactions_count for this type
		result, err := tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(reactions_count, ARRAY[$2], (COALESCE((reactions_count->$2)::text, '0')::int + 1)::text::jsonb) WHERE id=$1`, mangaID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("increment reactions count: %w", err)
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			return "", corecommon.ErrNotFound
		}
		if err := tx.Commit(); err != nil {
			return "", fmt.Errorf("commit reaction insert: %w", err)
		}
		return string(reactionType), nil
	}
}

func (r *PostgresMangaRepository) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return "", fmt.Errorf("postgres manga repo not initialized")
	}

	var reactionType string
	query := `SELECT reaction_type FROM manga_likes WHERE manga_id=$1 AND user_id=$2`
	if err := r.store.DB.GetContext(ctx, &reactionType, query, mangaID.Hex(), userID.Hex()); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", nil
		}
		return "", fmt.Errorf("get user reaction: %w", err)
	}

	return reactionType, nil
}

func (r *PostgresMangaRepository) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga repo not initialized")
	}

	var total int64
	if err := r.store.DB.GetContext(ctx, &total, `SELECT COUNT(*) FROM manga_likes WHERE user_id=$1`, userID.Hex()); err != nil {
		return nil, 0, fmt.Errorf("count liked mangas: %w", err)
	}

	query := `SELECT m.id, m.title, m.slug, m.description, m.author_id, m.tags, m.cover_image, m.is_published, m.published_at, m.created_at, m.updated_at, m.views_count, m.likes_count, m.rating_sum, m.rating_count, m.average_rating FROM mangas m INNER JOIN manga_likes l ON m.id=l.manga_id WHERE l.user_id=$1 ORDER BY m.updated_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.store.DB.QueryxContext(ctx, query, userID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("query liked mangas: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		var m coremanga.Manga
		var idStr, authorID string
		var tagsString sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
			&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount,
			&m.RatingSum, &m.RatingCount, &m.AverageRating); err != nil {
			return nil, 0, fmt.Errorf("scan liked manga row: %w", err)
		}
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			m.ID = oid
		}
		if authorID != "" {
			if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
				m.AuthorID = aid
			}
		}
		if tagsString.Valid {
			_ = json.Unmarshal([]byte(tagsString.String), &m.Tags)
		}
		if publishedAt.Valid {
			m.PublishedAt = &publishedAt.Time
		}
		mangas = append(mangas, &m)
	}

	return mangas, total, nil
}

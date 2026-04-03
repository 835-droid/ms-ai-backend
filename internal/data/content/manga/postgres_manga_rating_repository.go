package manga

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
)

// ========== RATING METHODS ==========

func (r *PostgresMangaRepository) AddRating(ctx context.Context, rating *coremanga.MangaRating) (float64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return 0, fmt.Errorf("postgres manga repo not initialized")
	}
	if rating == nil {
		return 0, corecommon.ErrInvalidInput
	}

	var existingScore sql.NullFloat64
	err := r.store.DB.GetContext(ctx, &existingScore, `SELECT score FROM manga_ratings WHERE manga_id=$1 AND user_id=$2`, rating.MangaID.Hex(), rating.UserID.Hex())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("check existing rating: %w", err)
	}

	var currentSum float64
	var currentCount int64
	err = r.store.DB.GetContext(ctx, &currentSum, `SELECT COALESCE(rating_sum,0) FROM mangas WHERE id=$1`, rating.MangaID.Hex())
	if err != nil {
		return 0, fmt.Errorf("fetch current rating_sum: %w", err)
	}
	err = r.store.DB.GetContext(ctx, &currentCount, `SELECT COALESCE(rating_count,0) FROM mangas WHERE id=$1`, rating.MangaID.Hex())
	if err != nil {
		return 0, fmt.Errorf("fetch current rating_count: %w", err)
	}

	if !existingScore.Valid {
		_, err = r.store.DB.ExecContext(ctx, `INSERT INTO manga_ratings (manga_id, user_id, score, created_at) VALUES ($1, $2, $3, $4)`, rating.MangaID.Hex(), rating.UserID.Hex(), rating.Score, time.Now())
		if err != nil {
			return 0, fmt.Errorf("insert rating: %w", err)
		}
		newSum := currentSum + rating.Score
		newCount := currentCount + 1
		average := 0.0
		if newCount > 0 {
			average = newSum / float64(newCount)
		}
		_, err = r.store.DB.ExecContext(ctx, `UPDATE mangas SET rating_sum=$1, rating_count=$2, average_rating=$3 WHERE id=$4`, newSum, newCount, average, rating.MangaID.Hex())
		if err != nil {
			return 0, fmt.Errorf("update manga rating stats: %w", err)
		}
		return average, nil
	}

	oldScore := existingScore.Float64
	_, err = r.store.DB.ExecContext(ctx, `UPDATE manga_ratings SET score=$1 WHERE manga_id=$2 AND user_id=$3`, rating.Score, rating.MangaID.Hex(), rating.UserID.Hex())
	if err != nil {
		return 0, fmt.Errorf("update rating: %w", err)
	}

	newSum := currentSum - oldScore + rating.Score
	average := 0.0
	if currentCount > 0 {
		average = newSum / float64(currentCount)
	}
	_, err = r.store.DB.ExecContext(ctx, `UPDATE mangas SET rating_sum=$1, average_rating=$2 WHERE id=$3`, newSum, average, rating.MangaID.Hex())
	if err != nil {
		return 0, fmt.Errorf("update manga rating stats: %w", err)
	}
	return average, nil
}

func (r *PostgresMangaRepository) HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga repo not initialized")
	}
	var exists bool
	if err := r.store.DB.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM manga_ratings WHERE manga_id=$1 AND user_id=$2)`, mangaID.Hex(), userID.Hex()); err != nil {
		return false, fmt.Errorf("check user rated: %w", err)
	}
	return exists, nil
}

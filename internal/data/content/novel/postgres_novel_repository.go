// Package novel implements novel data access operations backed by PostgreSQL.
package novel

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	corenovel "github.com/835-droid/ms-ai-backend/internal/core/content/novel"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostgresNovelRepository implements corenovel.NovelRepository backed by PostgreSQL.
type PostgresNovelRepository struct {
	store        *pginfra.PostgresStore
	reactionLock sync.Map
	ratingLock   sync.Map
}

// NewPostgresNovelRepository creates a new PostgresNovelRepository.
func NewPostgresNovelRepository(s *pginfra.PostgresStore) *PostgresNovelRepository {
	return &PostgresNovelRepository{store: s}
}

func (r *PostgresNovelRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres novel repository not initialized")
	}
	return nil
}

// generateID generates a 24-char hex ID compatible with MongoDB ObjectID format
func generateID() string {
	return uuid.New().String()[:24]
}

// scanNovel scans a novel from a database row
func (r *PostgresNovelRepository) scanNovel(row *sql.Row) (*corenovel.Novel, error) {
	var n corenovel.Novel
	var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
	var publishedAt sql.NullTime

	err := row.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
		&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
		&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, corecommon.ErrNotFound
		}
		return nil, fmt.Errorf("scan novel: %w", err)
	}

	if idStr.Valid {
		if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
			n.ID = oid
		} else {
			return nil, fmt.Errorf("invalid novel id: %s", idStr.String)
		}
	}
	if authorIDStr.Valid {
		if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
			n.AuthorID = oid
		}
	}
	if tagsStr.Valid {
		_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
	}
	if reactionsCountStr.Valid {
		_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
	}
	if publishedAt.Valid {
		n.PublishedAt = &publishedAt.Time
	}

	return &n, nil
}

// CreateNovel creates a new novel.
func (r *PostgresNovelRepository) CreateNovel(ctx context.Context, novel *corenovel.Novel) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	if novel.ID.IsZero() {
		novel.ID = primitive.NewObjectID()
	}

	tagsJSON, _ := json.Marshal(novel.Tags)
	reactionsJSON, _ := json.Marshal(novel.ReactionsCount)

	query := `INSERT INTO novels (id, title, slug, description, author_id, tags, cover_image, 
		is_published, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`

	_, err := r.store.DB.ExecContext(ctx, query,
		novel.ID.Hex(), novel.Title, novel.Slug, novel.Description, novel.AuthorID.Hex(),
		string(tagsJSON), novel.CoverImage, novel.IsPublished, novel.CreatedAt, novel.UpdatedAt,
		novel.ViewsCount, novel.FavoritesCount, novel.RatingSum, novel.RatingCount,
		novel.AverageRating, string(reactionsJSON),
	)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"novels_slug_key\"" {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("insert novel: %w", err)
	}
	return nil
}

// GetNovelByID retrieves a novel by ID.
func (r *PostgresNovelRepository) GetNovelByID(ctx context.Context, id primitive.ObjectID) (*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels WHERE id = $1`

	row := r.store.DB.QueryRowContext(ctx, query, id.Hex())
	return r.scanNovel(row)
}

// GetNovelBySlug retrieves a novel by slug.
func (r *PostgresNovelRepository) GetNovelBySlug(ctx context.Context, slug string) (*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels WHERE slug = $1`

	row := r.store.DB.QueryRowContext(ctx, query, slug)
	return r.scanNovel(row)
}

// ListNovels retrieves a paginated list of novels.
func (r *PostgresNovelRepository) ListNovels(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	// Count total
	var total int64
	err := r.store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM novels").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count novels: %w", err)
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels 
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list novels: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, 0, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, total, nil
}

// UpdateNovel updates an existing novel.
func (r *PostgresNovelRepository) UpdateNovel(ctx context.Context, novel *corenovel.Novel) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	tagsJSON, _ := json.Marshal(novel.Tags)
	reactionsJSON, _ := json.Marshal(novel.ReactionsCount)

	query := `UPDATE novels SET title=$1, slug=$2, description=$3, author_id=$4, tags=$5, 
		cover_image=$6, is_published=$7, updated_at=$8, views_count=$9, favorites_count=$10,
		rating_sum=$11, rating_count=$12, average_rating=$13, reactions_count=$14
		WHERE id=$15`

	result, err := r.store.DB.ExecContext(ctx, query,
		novel.Title, novel.Slug, novel.Description, novel.AuthorID.Hex(), string(tagsJSON),
		novel.CoverImage, novel.IsPublished, novel.UpdatedAt, novel.ViewsCount, novel.FavoritesCount,
		novel.RatingSum, novel.RatingCount, novel.AverageRating, string(reactionsJSON), novel.ID.Hex(),
	)
	if err != nil {
		return fmt.Errorf("update novel: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return corecommon.ErrNotFound
	}

	return nil
}

// DeleteNovel deletes a novel.
func (r *PostgresNovelRepository) DeleteNovel(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	result, err := r.store.DB.ExecContext(ctx, "DELETE FROM novels WHERE id = $1", id.Hex())
	if err != nil {
		return fmt.Errorf("delete novel: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return corecommon.ErrNotFound
	}

	return nil
}

// IncrementViews increments the view count for a novel.
func (r *PostgresNovelRepository) IncrementViews(ctx context.Context, novelID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	_, err := r.store.DB.ExecContext(ctx,
		"UPDATE novels SET views_count = views_count + 1 WHERE id = $1", novelID.Hex())
	return err
}

// LogView logs a view for a novel.
func (r *PostgresNovelRepository) LogView(ctx context.Context, novelID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	_, err := r.store.DB.ExecContext(ctx,
		"INSERT INTO novel_view_logs (novel_id) VALUES ($1)", novelID.Hex())
	if err != nil {
		return err
	}
	return r.IncrementViews(ctx, novelID)
}

// ListMostViewed returns the most viewed novels for a given time period.
func (r *PostgresNovelRepository) ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*corenovel.RankedNovel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	var query string
	if since.IsZero() {
		query = `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
			published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
			rating_count, average_rating, reactions_count 
			FROM novels ORDER BY views_count DESC LIMIT $1 OFFSET $2`
	} else {
		query = `SELECT n.id, n.title, n.slug, n.description, n.author_id, n.tags, n.cover_image, 
			n.is_published, n.published_at, n.created_at, n.updated_at, n.views_count, 
			n.favorites_count, n.rating_sum, n.rating_count, n.average_rating, n.reactions_count,
			COUNT(vl.id) as view_count
			FROM novels n
			LEFT JOIN novel_view_logs vl ON n.id = vl.novel_id AND vl.viewed_at >= $3
			GROUP BY n.id ORDER BY view_count DESC LIMIT $1 OFFSET $2`
	}

	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip, since)
	if err != nil {
		return nil, fmt.Errorf("list most viewed: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.RankedNovel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime
		var viewCount int64

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.FavoritesCount,
			&n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr, &viewCount)
		if err != nil {
			return nil, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &corenovel.RankedNovel{
			Novel:     &n,
			ViewCount: viewCount,
		})
	}

	return novels, nil
}

// ListRecentlyUpdated returns recently updated novels.
func (r *PostgresNovelRepository) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels 
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list recently updated: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, nil
}

// ListMostFollowed returns the most followed novels.
func (r *PostgresNovelRepository) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels 
		ORDER BY favorites_count DESC LIMIT $1 OFFSET $2`

	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list most followed: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, nil
}

// ListTopRated returns the top rated novels.
func (r *PostgresNovelRepository) ListTopRated(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, 
		published_at, created_at, updated_at, views_count, favorites_count, rating_sum, 
		rating_count, average_rating, reactions_count FROM novels 
		WHERE rating_count > 0
		ORDER BY average_rating DESC, rating_count DESC LIMIT $1 OFFSET $2`

	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list top rated: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, nil
}

// SetReaction sets or toggles a reaction for a novel.
func (r *PostgresNovelRepository) SetReaction(ctx context.Context, novelID, userID primitive.ObjectID, reactionType corenovel.ReactionType) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}

	// Check if reaction exists
	var existingType string
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT type FROM novel_reactions WHERE novel_id = $1 AND user_id = $2",
		novelID.Hex(), userID.Hex()).Scan(&existingType)

	if err == nil {
		// Reaction exists, toggle it off
		_, err = r.store.DB.ExecContext(ctx,
			"DELETE FROM novel_reactions WHERE novel_id = $1 AND user_id = $2",
			novelID.Hex(), userID.Hex())
		return "", err
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("check reaction: %w", err)
	}

	// No existing reaction, create new one
	id := generateID()
	_, err = r.store.DB.ExecContext(ctx,
		"INSERT INTO novel_reactions (id, novel_id, user_id, type) VALUES ($1, $2, $3, $4)",
		id, novelID.Hex(), userID.Hex(), string(reactionType))
	return string(reactionType), err
}

// GetUserReaction gets the current reaction type for a user on a novel.
func (r *PostgresNovelRepository) GetUserReaction(ctx context.Context, novelID, userID primitive.ObjectID) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}

	var reactionType string
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT type FROM novel_reactions WHERE novel_id = $1 AND user_id = $2",
		novelID.Hex(), userID.Hex()).Scan(&reactionType)

	if err == sql.ErrNoRows {
		return "", nil
	}
	return reactionType, err
}

// ListLikedNovels returns novels liked by a user.
func (r *PostgresNovelRepository) ListLikedNovels(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	// Count total
	var total int64
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM novel_reactions WHERE user_id = $1 AND type = 'upvote'",
		userID.Hex()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count liked novels: %w", err)
	}

	query := `SELECT n.id, n.title, n.slug, n.description, n.author_id, n.tags, n.cover_image, 
		n.is_published, n.published_at, n.created_at, n.updated_at, n.views_count, 
		n.favorites_count, n.rating_sum, n.rating_count, n.average_rating, n.reactions_count
		FROM novels n
		JOIN novel_reactions nr ON n.id = nr.novel_id
		WHERE nr.user_id = $1 AND nr.type = 'upvote'
		ORDER BY nr.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.store.DB.QueryContext(ctx, query, userID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list liked novels: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, 0, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, total, nil
}

// AddRating stores a user rating.
func (r *PostgresNovelRepository) AddRating(ctx context.Context, rating *corenovel.NovelRating) (float64, error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, err
	}

	id := generateID()

	// Use INSERT ... ON CONFLICT for upsert
	query := `INSERT INTO novel_ratings (id, novel_id, user_id, score) VALUES ($1, $2, $3, $4)
		ON CONFLICT (novel_id, user_id) DO UPDATE SET score = EXCLUDED.score`

	_, err := r.store.DB.ExecContext(ctx, query, id, rating.NovelID.Hex(), rating.UserID.Hex(), rating.Score)
	if err != nil {
		return 0, fmt.Errorf("add rating: %w", err)
	}

	// Update novel rating aggregates
	updateQuery := `UPDATE novels SET 
		rating_sum = (SELECT COALESCE(SUM(score), 0) FROM novel_ratings WHERE novel_id = $1),
		rating_count = (SELECT COUNT(*) FROM novel_ratings WHERE novel_id = $1),
		average_rating = (SELECT COALESCE(AVG(score), 0) FROM novel_ratings WHERE novel_id = $1)
		WHERE id = $1`

	_, err = r.store.DB.ExecContext(ctx, updateQuery, rating.NovelID.Hex())
	if err != nil {
		return 0, fmt.Errorf("update rating aggregates: %w", err)
	}

	// Get new average
	var avgRating float64
	err = r.store.DB.QueryRowContext(ctx,
		"SELECT average_rating FROM novels WHERE id = $1", rating.NovelID.Hex()).Scan(&avgRating)

	return avgRating, err
}

// HasUserRated checks if a user has rated a novel.
func (r *PostgresNovelRepository) HasUserRated(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	var count int
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM novel_ratings WHERE novel_id = $1 AND user_id = $2",
		novelID.Hex(), userID.Hex()).Scan(&count)

	return count > 0, err
}

// AddFavorite adds a novel to user's favorites.
func (r *PostgresNovelRepository) AddFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	id := generateID()
	query := `INSERT INTO novel_favorites (id, novel_id, user_id) VALUES ($1, $2, $3)
		ON CONFLICT (novel_id, user_id) DO NOTHING`

	_, err := r.store.DB.ExecContext(ctx, query, id, novelID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}

	// Update favorites count
	_, err = r.store.DB.ExecContext(ctx,
		"UPDATE novels SET favorites_count = favorites_count + 1 WHERE id = $1", novelID.Hex())

	return err
}

// RemoveFavorite removes a novel from user's favorites.
func (r *PostgresNovelRepository) RemoveFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	_, err := r.store.DB.ExecContext(ctx,
		"DELETE FROM novel_favorites WHERE novel_id = $1 AND user_id = $2",
		novelID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}

	// Update favorites count
	_, err = r.store.DB.ExecContext(ctx,
		"UPDATE novels SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = $1",
		novelID.Hex())

	return err
}

// IsFavorite checks if a novel is in user's favorites.
func (r *PostgresNovelRepository) IsFavorite(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	var count int
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM novel_favorites WHERE novel_id = $1 AND user_id = $2",
		novelID.Hex(), userID.Hex()).Scan(&count)

	return count > 0, err
}

// ListFavorites retrieves a user's favorite novels.
func (r *PostgresNovelRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	// Count total
	var total int64
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM novel_favorites WHERE user_id = $1",
		userID.Hex()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	query := `SELECT n.id, n.title, n.slug, n.description, n.author_id, n.tags, n.cover_image, 
		n.is_published, n.published_at, n.created_at, n.updated_at, n.views_count, 
		n.favorites_count, n.rating_sum, n.rating_count, n.average_rating, n.reactions_count
		FROM novels n
		JOIN novel_favorites nf ON n.id = nf.novel_id
		WHERE nf.user_id = $1
		ORDER BY nf.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.store.DB.QueryContext(ctx, query, userID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	var novels []*corenovel.Novel
	for rows.Next() {
		var n corenovel.Novel
		var idStr, authorIDStr, tagsStr, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&idStr, &n.Title, &n.Slug, &n.Description, &authorIDStr, &tagsStr, &n.CoverImage,
			&n.IsPublished, &publishedAt, &n.CreatedAt, &n.UpdatedAt, &n.ViewsCount, &n.LikesCount,
			&n.FavoritesCount, &n.RatingSum, &n.RatingCount, &n.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, 0, fmt.Errorf("scan novel: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				n.ID = oid
			}
		}
		if authorIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(authorIDStr.String); err == nil {
				n.AuthorID = oid
			}
		}
		if tagsStr.Valid {
			_ = json.Unmarshal([]byte(tagsStr.String), &n.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &n.ReactionsCount)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}

		novels = append(novels, &n)
	}

	return novels, total, nil
}

// AddNovelComment adds a comment to a novel.
func (r *PostgresNovelRepository) AddNovelComment(ctx context.Context, comment *corenovel.NovelComment) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}

	query := `INSERT INTO novel_comments (id, novel_id, user_id, username, content) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.store.DB.ExecContext(ctx, query,
		comment.ID.Hex(), comment.NovelID.Hex(), comment.UserID.Hex(), comment.Username, comment.Content)

	return err
}

// ListNovelComments retrieves comments for a novel.
func (r *PostgresNovelRepository) ListNovelComments(ctx context.Context, novelID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*corenovel.NovelComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	// Count total
	var total int64
	err := r.store.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM novel_comments WHERE novel_id = $1",
		novelID.Hex()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	order := "created_at DESC"
	if sortOrder == "oldest" {
		order = "created_at ASC"
	}

	query := fmt.Sprintf(`SELECT id, novel_id, user_id, username, content, created_at, updated_at 
		FROM novel_comments WHERE novel_id = $1 ORDER BY %s LIMIT $2 OFFSET $3`, order)

	rows, err := r.store.DB.QueryContext(ctx, query, novelID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []*corenovel.NovelComment
	for rows.Next() {
		var c corenovel.NovelComment
		var idStr, novelIDStr, userIDStr sql.NullString

		err := rows.Scan(&idStr, &novelIDStr, &userIDStr, &c.Username, &c.Content, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan comment: %w", err)
		}

		if idStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
				c.ID = oid
			}
		}
		if novelIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(novelIDStr.String); err == nil {
				c.NovelID = oid
			}
		}
		if userIDStr.Valid {
			if oid, err := primitive.ObjectIDFromHex(userIDStr.String); err == nil {
				c.UserID = oid
			}
		}

		comments = append(comments, &c)
	}

	return comments, total, nil
}

// DeleteNovelComment deletes a novel comment.
func (r *PostgresNovelRepository) DeleteNovelComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	result, err := r.store.DB.ExecContext(ctx,
		"DELETE FROM novel_comments WHERE id = $1 AND user_id = $2",
		commentID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return corecommon.ErrNotFound
	}

	return nil
}

// ----- START OF FILE: backend/MS-AI/internal/data/postgres/manga_repo.go -----
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type postgresMangaRepository struct {
	store *PostgresStore
}

func NewPostgresMangaRepository(store *PostgresStore) coremanga.MangaRepository {
	return &postgresMangaRepository{store: store}
}

func (r *postgresMangaRepository) CreateManga(ctx context.Context, manga *coremanga.Manga) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	if manga.ID.IsZero() {
		manga.ID = primitive.NewObjectID()
	}

	tagsJSON, err := json.Marshal(manga.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	query := `INSERT INTO mangas (id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, rating_sum, rating_count, average_rating)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err = r.store.DB.ExecContext(ctx, query,
		manga.ID.Hex(), manga.Title, manga.Slug, manga.Description, manga.AuthorID.Hex(), string(tagsJSON), manga.CoverImage,
		manga.IsPublished, manga.PublishedAt, manga.CreatedAt, manga.UpdatedAt, manga.ViewsCount, manga.LikesCount,
		manga.RatingSum, manga.RatingCount, manga.AverageRating,
	)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"mangas_slug_key\"" {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("insert manga: %w", err)
	}
	return nil
}

func (r *postgresMangaRepository) scanManga(row scannable) (*coremanga.Manga, error) {
	var m coremanga.Manga
	var idStr, authorIdStr, tagsStr sql.NullString
	var publishedAt sql.NullTime
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorIdStr, &tagsStr, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating); err != nil {
		return nil, err
	}
	if idStr.Valid {
		if oid, err := primitive.ObjectIDFromHex(idStr.String); err == nil {
			m.ID = oid
		} else {
			return nil, fmt.Errorf("invalid manga id value: %s", idStr.String)
		}
	}
	if authorIdStr.Valid {
		if authorID, err := primitive.ObjectIDFromHex(authorIdStr.String); err == nil {
			m.AuthorID = authorID
		}
	}
	if tagsStr.Valid {
		_ = json.Unmarshal([]byte(tagsStr.String), &m.Tags)
	}
	if publishedAt.Valid {
		m.PublishedAt = &publishedAt.Time
	}
	return &m, nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *postgresMangaRepository) GetMangaByID(ctx context.Context, id primitive.ObjectID) (*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}

	var m coremanga.Manga
	var idStr, authorID string
	var tagsString sql.NullString
	var publishedAt sql.NullTime
	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, rating_sum, rating_count, average_rating FROM mangas WHERE id = $1`
	row := r.store.DB.QueryRowContext(ctx, query, id.Hex())
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, corecommon.ErrNotFound
		}
		return nil, fmt.Errorf("get manga by id: %w", err)
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
	return &m, nil
}

func (r *postgresMangaRepository) GetMangaBySlug(ctx context.Context, slug string) (*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}

	var m coremanga.Manga
	var idStr, authorID string
	var tagsString sql.NullString
	var publishedAt sql.NullTime
	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, rating_sum, rating_count, average_rating FROM mangas WHERE slug = $1`
	row := r.store.DB.QueryRowContext(ctx, query, slug)
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, corecommon.ErrNotFound
		}
		return nil, fmt.Errorf("get manga by slug: %w", err)
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
	return &m, nil
}

func (r *postgresMangaRepository) ListMangas(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga repo not initialized")
	}

	countQuery := `SELECT COUNT(*) FROM mangas`
	var total int64
	if err := r.store.DB.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("count mangas: %w", err)
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, rating_sum, rating_count, average_rating FROM mangas ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryxContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("query mangas: %w", err)
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
			return nil, 0, fmt.Errorf("scan manga row: %w", err)
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

func (r *postgresMangaRepository) UpdateManga(ctx context.Context, manga *coremanga.Manga) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}

	if manga.ID.IsZero() {
		return corecommon.ErrInvalidInput
	}

	tagsJSON, err := json.Marshal(manga.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	query := `UPDATE mangas SET title=$1, slug=$2, description=$3, author_id=$4, tags=$5, cover_image=$6, is_published=$7, published_at=$8, updated_at=$9, views_count=$10, likes_count=$11, rating_sum=$12, rating_count=$13, average_rating=$14 WHERE id=$15`
	res, err := r.store.DB.ExecContext(ctx, query,
		manga.Title, manga.Slug, manga.Description, manga.AuthorID.Hex(), string(tagsJSON), manga.CoverImage,
		manga.IsPublished, manga.PublishedAt, manga.UpdatedAt, manga.ViewsCount, manga.LikesCount,
		manga.RatingSum, manga.RatingCount, manga.AverageRating, manga.ID.Hex(),
	)
	if err != nil {
		return fmt.Errorf("update manga: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

func (r *postgresMangaRepository) DeleteManga(ctx context.Context, id primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}
	res, err := r.store.DB.ExecContext(ctx, `DELETE FROM mangas WHERE id=$1`, id.Hex())
	if err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

func (r *postgresMangaRepository) IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}
	res, err := r.store.DB.ExecContext(ctx, `UPDATE mangas SET views_count = views_count + 1 WHERE id=$1`, mangaID.Hex())
	if err != nil {
		return fmt.Errorf("increment views: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

func (r *postgresMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (string, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return "", fmt.Errorf("postgres manga repo not initialized")
	}

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
		// Same reaction, remove it
		_, err := tx.ExecContext(ctx, `DELETE FROM manga_likes WHERE manga_id=$1 AND user_id=$2`, mangaID.Hex(), userID.Hex())
		if err != nil {
			return "", fmt.Errorf("delete reaction: %w", err)
		}
		// Decrement reactions_count for this type
		_, err = tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(reactions_count, ARRAY[$2], (COALESCE(reactions_count->$2, '0')::int - 1)::text::jsonb) WHERE id=$1`, mangaID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("decrement reactions count: %w", err)
		}
		return "", tx.Commit()
	} else if existingType != "" {
		// Different reaction, update it
		_, err := tx.ExecContext(ctx, `UPDATE manga_likes SET reaction_type=$3 WHERE manga_id=$1 AND user_id=$2`, mangaID.Hex(), userID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("update reaction: %w", err)
		}
		// Update reactions_count: decrement old, increment new
		_, err = tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(jsonb_set(reactions_count, ARRAY[$2], (COALESCE(reactions_count->$2, '0')::int - 1)::text::jsonb), ARRAY[$3], (COALESCE(reactions_count->$3, '0')::int + 1)::text::jsonb) WHERE id=$1`, mangaID.Hex(), existingType, string(reactionType))
		if err != nil {
			return "", fmt.Errorf("update reactions count: %w", err)
		}
		return string(reactionType), tx.Commit()
	} else {
		// No reaction, add new one
		_, err := tx.ExecContext(ctx, `INSERT INTO manga_likes (manga_id, user_id, reaction_type) VALUES ($1, $2, $3)`, mangaID.Hex(), userID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("insert reaction: %w", err)
		}
		// Increment reactions_count for this type
		_, err = tx.ExecContext(ctx, `UPDATE mangas SET reactions_count = jsonb_set(reactions_count, ARRAY[$2], (COALESCE(reactions_count->$2, '0')::int + 1)::text::jsonb) WHERE id=$1`, mangaID.Hex(), string(reactionType))
		if err != nil {
			return "", fmt.Errorf("increment reactions count: %w", err)
		}
		return string(reactionType), tx.Commit()
	}
}

func (r *postgresMangaRepository) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
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

func (r *postgresMangaRepository) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
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

func (r *postgresMangaRepository) AddRating(ctx context.Context, rating *coremanga.MangaRating) (float64, error) {
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

func (r *postgresMangaRepository) HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga repo not initialized")
	}
	var exists bool
	if err := r.store.DB.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM manga_ratings WHERE manga_id=$1 AND user_id=$2)`, mangaID.Hex(), userID.Hex()); err != nil {
		return false, fmt.Errorf("check user rated: %w", err)
	}
	return exists, nil
}

// ========== ENGAGEMENT METHODS ==========

func (r *postgresMangaRepository) AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}
	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO user_favorites (manga_id, user_id, created_at) VALUES ($1, $2, $3)
		 ON CONFLICT (manga_id, user_id) DO NOTHING`,
		mangaID.Hex(), userID.Hex(), time.Now())
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (r *postgresMangaRepository) RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}
	_, err := r.store.DB.ExecContext(ctx,
		`DELETE FROM user_favorites WHERE manga_id=$1 AND user_id=$2`,
		mangaID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

func (r *postgresMangaRepository) IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
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

func (r *postgresMangaRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
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
				m.is_published, m.published_at, m.created_at, m.updated_at, m.views_count, m.likes_count,
				m.rating_sum, m.rating_count, m.average_rating
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

func (r *postgresMangaRepository) AddMangaComment(ctx context.Context, comment *coremanga.MangaComment) error {
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

func (r *postgresMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaComment, int64, error) {
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

	// Get paginated comments
	rows, err := r.store.DB.QueryxContext(ctx,
		`SELECT id, manga_id, user_id, username, content, created_at, updated_at
		 FROM manga_comments
		 WHERE manga_id=$1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
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

func (r *postgresMangaRepository) DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
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

// ----- END OF FILE: backend/MS-AI/internal/data/postgres/manga_repo.go -----

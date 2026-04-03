// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/postgres_manga_repository.go -----
package manga

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Specialized manga repository methods (reactions, ratings, favorites, comments)

// PostgresMangaRepository is the PostgreSQL backend implementation for manga repository.
type PostgresMangaRepository struct {
	store        *pginfra.PostgresStore
	reactionLock sync.Map
}

// NewPostgresMangaRepository creates a new PostgresMangaRepository instance.
func NewPostgresMangaRepository(store *pginfra.PostgresStore) *PostgresMangaRepository {
	return &PostgresMangaRepository{store: store}
}

func (r *PostgresMangaRepository) CreateManga(ctx context.Context, manga *coremanga.Manga) error {
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

func (r *PostgresMangaRepository) scanManga(row scannable) (*coremanga.Manga, error) {
	var m coremanga.Manga
	var idStr, authorIdStr, tagsStr, reactionsCountStr sql.NullString
	var publishedAt sql.NullTime
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorIdStr, &tagsStr, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount, &m.FavoritesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating, &reactionsCountStr); err != nil {
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
	if reactionsCountStr.Valid {
		_ = json.Unmarshal([]byte(reactionsCountStr.String), &m.ReactionsCount)
	}
	if publishedAt.Valid {
		m.PublishedAt = &publishedAt.Time
	}
	return &m, nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *PostgresMangaRepository) GetMangaByID(ctx context.Context, id primitive.ObjectID) (*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}

	var m coremanga.Manga
	var idStr, authorID string
	var tagsString, reactionsCountStr sql.NullString
	var publishedAt sql.NullTime
	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, favorites_count, rating_sum, rating_count, average_rating, reactions_count FROM mangas WHERE id = $1`
	row := r.store.DB.QueryRowContext(ctx, query, id.Hex())
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount, &m.FavoritesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating, &reactionsCountStr); err != nil {
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
	if reactionsCountStr.Valid {
		_ = json.Unmarshal([]byte(reactionsCountStr.String), &m.ReactionsCount)
	}
	if publishedAt.Valid {
		m.PublishedAt = &publishedAt.Time
	}
	return &m, nil
}

func (r *PostgresMangaRepository) GetMangaBySlug(ctx context.Context, slug string) (*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}

	var m coremanga.Manga
	var idStr, authorID string
	var tagsString, reactionsCountStr sql.NullString
	var publishedAt sql.NullTime
	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, favorites_count, rating_sum, rating_count, average_rating, reactions_count FROM mangas WHERE slug = $1`
	row := r.store.DB.QueryRowContext(ctx, query, slug)
	if err := row.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
		&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount, &m.FavoritesCount,
		&m.RatingSum, &m.RatingCount, &m.AverageRating, &reactionsCountStr); err != nil {
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
	if reactionsCountStr.Valid {
		_ = json.Unmarshal([]byte(reactionsCountStr.String), &m.ReactionsCount)
	}
	if publishedAt.Valid {
		m.PublishedAt = &publishedAt.Time
	}
	return &m, nil
}

func (r *PostgresMangaRepository) ListMangas(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga repo not initialized")
	}

	countQuery := `SELECT COUNT(*) FROM mangas`
	var total int64
	if err := r.store.DB.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("count mangas: %w", err)
	}

	query := `SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, favorites_count, rating_sum, rating_count, average_rating, reactions_count FROM mangas ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryxContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("query mangas: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		var m coremanga.Manga
		var idStr, authorID string
		var tagsString, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&idStr, &m.Title, &m.Slug, &m.Description, &authorID, &tagsString, &m.CoverImage,
			&m.IsPublished, &publishedAt, &m.CreatedAt, &m.UpdatedAt, &m.ViewsCount, &m.LikesCount,
			&m.FavoritesCount, &m.RatingSum, &m.RatingCount, &m.AverageRating, &reactionsCountStr); err != nil {
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
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &m.ReactionsCount)
		}
		if publishedAt.Valid {
			m.PublishedAt = &publishedAt.Time
		}
		mangas = append(mangas, &m)
	}

	return mangas, total, nil
}

func (r *PostgresMangaRepository) UpdateManga(ctx context.Context, manga *coremanga.Manga) error {
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

func (r *PostgresMangaRepository) DeleteManga(ctx context.Context, id primitive.ObjectID) error {
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

func (r *PostgresMangaRepository) IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error {
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

func (r *PostgresMangaRepository) LogView(ctx context.Context, mangaID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga repo not initialized")
	}
	_, err := r.store.DB.ExecContext(ctx, `INSERT INTO manga_view_logs (manga_id) VALUES ($1)`, mangaID.Hex())
	if err != nil {
		return fmt.Errorf("log view: %w", err)
	}
	return r.IncrementViews(ctx, mangaID)
}

func (r *PostgresMangaRepository) ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*coremanga.RankedManga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}

	// Special case for "all time" - use the persisted views_count instead of aggregating logs
	// This ensures mangas with historical views but no post-deployment logs still appear
	if since.IsZero() {
		// Remove is_published filter to show all manga (including unpublished)
		query := `
			SELECT m.id, m.title, m.slug, m.description, m.author_id, m.tags, m.cover_image, m.is_published, m.published_at, m.created_at, m.updated_at, m.views_count, m.likes_count, m.rating_sum, m.rating_count, m.average_rating
			FROM mangas m
			ORDER BY m.views_count DESC
			LIMIT $1 OFFSET $2
		`
		rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
		if err != nil {
			return nil, fmt.Errorf("list most viewed (all time): %w", err)
		}
		defer rows.Close()

		var rankedMangas []*coremanga.RankedManga
		for rows.Next() {
			var idStr, authorID string
			var tagsString sql.NullString
			var publishedAt sql.NullTime
			manga := &coremanga.Manga{}

			err := rows.Scan(&idStr, &manga.Title, &manga.Slug, &manga.Description, &authorID, &tagsString, &manga.CoverImage,
				&manga.IsPublished, &publishedAt, &manga.CreatedAt, &manga.UpdatedAt, &manga.ViewsCount, &manga.LikesCount,
				&manga.RatingSum, &manga.RatingCount, &manga.AverageRating)
			if err != nil {
				return nil, fmt.Errorf("scan manga: %w", err)
			}
			if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
				manga.ID = oid
			}
			if authorID != "" {
				if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
					manga.AuthorID = aid
				}
			}
			if tagsString.Valid {
				_ = json.Unmarshal([]byte(tagsString.String), &manga.Tags)
			}
			if publishedAt.Valid {
				manga.PublishedAt = &publishedAt.Time
			}
			rankedMangas = append(rankedMangas, &coremanga.RankedManga{
				Manga:     manga,
				ViewCount: manga.ViewsCount,
			})
		}
		return rankedMangas, nil
	}

	// For time-bounded periods, aggregate from manga_view_logs
	// Remove is_published filter to show all manga (including unpublished)
	query := `
		SELECT m.id, m.title, m.slug, m.description, m.author_id, m.tags, m.cover_image, m.is_published, m.published_at, m.created_at, m.updated_at, m.views_count, m.likes_count, m.rating_sum, m.rating_count, m.average_rating, v.view_count
		FROM mangas m
		INNER JOIN (
			SELECT l.manga_id, COUNT(*) as view_count
			FROM manga_view_logs l
			WHERE l.viewed_at >= $1
			GROUP BY l.manga_id
			ORDER BY view_count DESC
			LIMIT $2 OFFSET $3
		) v ON m.id = v.manga_id
		ORDER BY v.view_count DESC
	`
	rows, err := r.store.DB.QueryContext(ctx, query, since, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list most viewed: %w", err)
	}
	defer rows.Close()

	var rankedMangas []*coremanga.RankedManga
	for rows.Next() {
		var idStr, authorID string
		var tagsString sql.NullString
		var publishedAt sql.NullTime
		var viewCount int64
		manga := &coremanga.Manga{}

		err := rows.Scan(&idStr, &manga.Title, &manga.Slug, &manga.Description, &authorID, &tagsString, &manga.CoverImage,
			&manga.IsPublished, &publishedAt, &manga.CreatedAt, &manga.UpdatedAt, &manga.ViewsCount, &manga.LikesCount,
			&manga.RatingSum, &manga.RatingCount, &manga.AverageRating, &viewCount)
		if err != nil {
			return nil, fmt.Errorf("scan manga: %w", err)
		}
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			manga.ID = oid
		}
		if authorID != "" {
			if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
				manga.AuthorID = aid
			}
		}
		if tagsString.Valid {
			_ = json.Unmarshal([]byte(tagsString.String), &manga.Tags)
		}
		if publishedAt.Valid {
			manga.PublishedAt = &publishedAt.Time
		}
		rankedMangas = append(rankedMangas, &coremanga.RankedManga{
			Manga:     manga,
			ViewCount: viewCount,
		})
	}
	return rankedMangas, nil
}

func (r *PostgresMangaRepository) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}
	// Remove is_published filter to show all manga (including unpublished)
	query := `
		SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, rating_sum, rating_count, average_rating, reactions_count
		FROM mangas
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list recently updated: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		var idStr, authorID string
		var tagsString, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime
		manga := &coremanga.Manga{}

		err := rows.Scan(&idStr, &manga.Title, &manga.Slug, &manga.Description, &authorID, &tagsString, &manga.CoverImage,
			&manga.IsPublished, &publishedAt, &manga.CreatedAt, &manga.UpdatedAt, &manga.ViewsCount, &manga.LikesCount,
			&manga.RatingSum, &manga.RatingCount, &manga.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan manga: %w", err)
		}
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			manga.ID = oid
		}
		if authorID != "" {
			if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
				manga.AuthorID = aid
			}
		}
		if tagsString.Valid {
			_ = json.Unmarshal([]byte(tagsString.String), &manga.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &manga.ReactionsCount)
		}
		if publishedAt.Valid {
			manga.PublishedAt = &publishedAt.Time
		}
		mangas = append(mangas, manga)
	}
	return mangas, nil
}

func (r *PostgresMangaRepository) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}
	// Remove is_published filter to show all manga (including unpublished)
	query := `
		SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, favorites_count, rating_sum, rating_count, average_rating, reactions_count
		FROM mangas
		ORDER BY favorites_count DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list most followed: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		var idStr, authorID string
		var tagsString, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime
		manga := &coremanga.Manga{}

		err := rows.Scan(&idStr, &manga.Title, &manga.Slug, &manga.Description, &authorID, &tagsString, &manga.CoverImage,
			&manga.IsPublished, &publishedAt, &manga.CreatedAt, &manga.UpdatedAt, &manga.ViewsCount, &manga.LikesCount,
			&manga.FavoritesCount, &manga.RatingSum, &manga.RatingCount, &manga.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan manga: %w", err)
		}
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			manga.ID = oid
		}
		if authorID != "" {
			if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
				manga.AuthorID = aid
			}
		}
		if tagsString.Valid {
			_ = json.Unmarshal([]byte(tagsString.String), &manga.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &manga.ReactionsCount)
		}
		if publishedAt.Valid {
			manga.PublishedAt = &publishedAt.Time
		}
		mangas = append(mangas, manga)
	}
	return mangas, nil
}

func (r *PostgresMangaRepository) ListTopRated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga repo not initialized")
	}
	// Remove is_published filter to show all manga with ratings (including unpublished)
	query := `
		SELECT id, title, slug, description, author_id, tags, cover_image, is_published, published_at, created_at, updated_at, views_count, likes_count, favorites_count, rating_sum, rating_count, average_rating, reactions_count
		FROM mangas
		WHERE rating_count > 0
		ORDER BY average_rating DESC, rating_count DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("list top rated: %w", err)
	}
	defer rows.Close()

	var mangas []*coremanga.Manga
	for rows.Next() {
		var idStr, authorID string
		var tagsString, reactionsCountStr sql.NullString
		var publishedAt sql.NullTime
		manga := &coremanga.Manga{}

		err := rows.Scan(&idStr, &manga.Title, &manga.Slug, &manga.Description, &authorID, &tagsString, &manga.CoverImage,
			&manga.IsPublished, &publishedAt, &manga.CreatedAt, &manga.UpdatedAt, &manga.ViewsCount, &manga.LikesCount,
			&manga.FavoritesCount, &manga.RatingSum, &manga.RatingCount, &manga.AverageRating, &reactionsCountStr)
		if err != nil {
			return nil, fmt.Errorf("scan manga: %w", err)
		}
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			manga.ID = oid
		}
		if authorID != "" {
			if aid, err := primitive.ObjectIDFromHex(authorID); err == nil {
				manga.AuthorID = aid
			}
		}
		if tagsString.Valid {
			_ = json.Unmarshal([]byte(tagsString.String), &manga.Tags)
		}
		if reactionsCountStr.Valid {
			_ = json.Unmarshal([]byte(reactionsCountStr.String), &manga.ReactionsCount)
		}
		if publishedAt.Valid {
			manga.PublishedAt = &publishedAt.Time
		}
		mangas = append(mangas, manga)
	}
	return mangas, nil
}

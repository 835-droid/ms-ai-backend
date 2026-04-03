// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/postgres_chapter_repository.go -----
package manga

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostgresMangaChapterRepository struct {
	store *pginfra.PostgresStore
}

func NewPostgresChapterRepository(store *pginfra.PostgresStore) coremanga.MangaChapterRepository {
	return &PostgresMangaChapterRepository{store: store}
}

func (r *PostgresMangaChapterRepository) CreateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}
	if chapter.ID.IsZero() {
		chapter.ID = primitive.NewObjectID()
	}

	pagesJSON, err := json.Marshal(chapter.Pages)
	if err != nil {
		return fmt.Errorf("marshal pages: %w", err)
	}

	_, err = r.store.DB.ExecContext(ctx, `INSERT INTO manga_chapters (id, manga_id, title, pages, number, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		chapter.ID.Hex(), chapter.MangaID.Hex(), chapter.Title, string(pagesJSON), chapter.Number, chapter.CreatedAt, chapter.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert manga chapter: %w", err)
	}
	return nil
}

func (r *PostgresMangaChapterRepository) GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*coremanga.MangaChapter, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	var c coremanga.MangaChapter
	var idStr, mangaIdStr string
	var pagesStr sql.NullString
	query := `SELECT id, manga_id, title, pages, number, views_count, rating_sum, rating_count, average_rating, created_at, updated_at FROM manga_chapters WHERE id=$1`
	row := r.store.DB.QueryRowContext(ctx, query, id.Hex())
	if err := row.Scan(&idStr, &mangaIdStr, &c.Title, &pagesStr, &c.Number, &c.ViewsCount, &c.RatingSum, &c.RatingCount, &c.AverageRating, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, corecommon.ErrNotFound
		}
		return nil, fmt.Errorf("get manga chapter by id: %w", err)
	}
	c.ID, _ = primitive.ObjectIDFromHex(idStr)
	c.MangaID, _ = primitive.ObjectIDFromHex(mangaIdStr)
	if pagesStr.Valid {
		_ = json.Unmarshal([]byte(pagesStr.String), &c.Pages)
	}
	return &c, nil
}

func (r *PostgresMangaChapterRepository) ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaChapter, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	var total int64
	if err := r.store.DB.GetContext(ctx, &total, `SELECT COUNT(*) FROM manga_chapters WHERE manga_id=$1`, mangaID.Hex()); err != nil {
		return nil, 0, fmt.Errorf("count manga chapters: %w", err)
	}

	query := `SELECT id, manga_id, title, pages, number, views_count, rating_sum, rating_count, average_rating, created_at, updated_at FROM manga_chapters WHERE manga_id=$1 ORDER BY number ASC LIMIT $2 OFFSET $3`
	rows, err := r.store.DB.QueryxContext(ctx, query, mangaID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("query manga chapters: %w", err)
	}
	defer rows.Close()

	var chapters []*coremanga.MangaChapter
	for rows.Next() {
		var c coremanga.MangaChapter
		var idStr, mangaIdStr string
		var pagesStr sql.NullString
		if err := rows.Scan(&idStr, &mangaIdStr, &c.Title, &pagesStr, &c.Number, &c.ViewsCount, &c.RatingSum, &c.RatingCount, &c.AverageRating, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan manga chapter: %w", err)
		}
		c.ID, _ = primitive.ObjectIDFromHex(idStr)
		c.MangaID, _ = primitive.ObjectIDFromHex(mangaIdStr)
		if pagesStr.Valid {
			_ = json.Unmarshal([]byte(pagesStr.String), &c.Pages)
		}
		chapters = append(chapters, &c)
	}

	return chapters, total, nil
}

func (r *PostgresMangaChapterRepository) UpdateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}
	if chapter.ID.IsZero() {
		return corecommon.ErrInvalidInput
	}
	pagesJSON, err := json.Marshal(chapter.Pages)
	if err != nil {
		return fmt.Errorf("marshal pages: %w", err)
	}
	chapter.UpdatedAt = time.Now()
	res, err := r.store.DB.ExecContext(ctx, `UPDATE manga_chapters SET title=$1, pages=$2, number=$3, updated_at=$4 WHERE id=$5`,
		chapter.Title, string(pagesJSON), chapter.Number, chapter.UpdatedAt, chapter.ID.Hex())
	if err != nil {
		return fmt.Errorf("update manga chapter: %w", err)
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

func (r *PostgresMangaChapterRepository) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}
	res, err := r.store.DB.ExecContext(ctx, `DELETE FROM manga_chapters WHERE id=$1`, id.Hex())
	if err != nil {
		return fmt.Errorf("delete manga chapter: %w", err)
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// ========== ENGAGEMENT METHODS ==========

func (r *PostgresMangaChapterRepository) IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}

	// Increment chapter views
	_, err := r.store.DB.ExecContext(ctx,
		`UPDATE manga_chapters SET views_count = views_count + 1 WHERE id=$1`,
		chapterID.Hex())
	if err != nil {
		return fmt.Errorf("increment chapter views: %w", err)
	}

	// Increment manga total views
	_, err = r.store.DB.ExecContext(ctx,
		`UPDATE mangas SET views_count = views_count + 1 WHERE id=$1`,
		mangaID.Hex())
	if err != nil {
		return fmt.Errorf("increment manga views: %w", err)
	}

	// Log view to manga_view_logs for period-based analytics
	_, err = r.store.DB.ExecContext(ctx,
		`INSERT INTO manga_view_logs (manga_id) VALUES ($1)`,
		mangaID.Hex())
	if err != nil {
		return fmt.Errorf("log manga view: %w", err)
	}

	return nil
}

func (r *PostgresMangaChapterRepository) TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}

	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO chapter_views (chapter_id, manga_id, user_id, created_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (chapter_id, user_id) DO NOTHING`,
		chapterID.Hex(), mangaID.Hex(), userID.Hex(), time.Now())
	if err != nil {
		return fmt.Errorf("track chapter view: %w", err)
	}
	return nil
}

func (r *PostgresMangaChapterRepository) HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	var exists bool
	err := r.store.DB.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM chapter_views WHERE chapter_id=$1 AND user_id=$2)`,
		chapterID.Hex(), userID.Hex())
	if err != nil {
		return false, fmt.Errorf("check chapter view: %w", err)
	}
	return exists, nil
}

func (r *PostgresMangaChapterRepository) TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	// Use transaction for atomic operation
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Try to insert view record - use RETURNING to detect if it was inserted
	var inserted bool
	err = tx.QueryRowContext(ctx,
		`INSERT INTO chapter_views (chapter_id, manga_id, user_id, created_at) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (chapter_id, user_id) DO NOTHING
		 RETURNING (xmax = 0) as inserted`,
		chapterID.Hex(), mangaID.Hex(), userID.Hex(), time.Now()).Scan(&inserted)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("upsert chapter view: %w", err)
	}

	// If no row was returned, it means the record already existed (conflict)
	isNewView := err != sql.ErrNoRows

	// Only increment counters if this was a new view
	if isNewView {
		// Increment chapter views
		_, err = tx.ExecContext(ctx,
			`UPDATE manga_chapters SET views_count = views_count + 1 WHERE id=$1`,
			chapterID.Hex())
		if err != nil {
			return false, fmt.Errorf("increment chapter views: %w", err)
		}

		// Increment manga total views
		_, err = tx.ExecContext(ctx,
			`UPDATE mangas SET views_count = views_count + 1 WHERE id=$1`,
			mangaID.Hex())
		if err != nil {
			return false, fmt.Errorf("increment manga views: %w", err)
		}

		// Log view to manga_view_logs for period-based analytics
		_, err = tx.ExecContext(ctx,
			`INSERT INTO manga_view_logs (manga_id) VALUES ($1)`,
			mangaID.Hex())
		if err != nil {
			return false, fmt.Errorf("log manga view: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}

	return isNewView, nil
}

func (r *PostgresMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (float64, int64, float64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return 0, 0, 0, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	// Insert or update rating (allow re-rating)
	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO chapter_ratings (chapter_id, manga_id, user_id, score, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5)
		 ON CONFLICT (chapter_id, user_id) DO UPDATE SET
		   score = EXCLUDED.score,
		   updated_at = EXCLUDED.updated_at`,
		rating.ChapterID.Hex(), rating.MangaID.Hex(), rating.UserID.Hex(), rating.Score, time.Now())
	if err != nil {
		return 0, 0, 0, fmt.Errorf("upsert chapter rating: %w", err)
	}

	// Calculate average rating for chapter
	var chapterAverage float64
	err = r.store.DB.GetContext(ctx, &chapterAverage,
		`SELECT AVG(score) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		chapterAverage = 0
	}

	// Update chapter with new average
	var chapterCount, chapterSum float64
	err = r.store.DB.GetContext(ctx, &chapterCount,
		`SELECT COUNT(*) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		chapterCount = 0
	}
	err = r.store.DB.GetContext(ctx, &chapterSum,
		`SELECT COALESCE(SUM(score), 0) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		chapterSum = 0
	}

	_, err = r.store.DB.ExecContext(ctx,
		`UPDATE manga_chapters SET average_rating=$1, rating_count=$2, rating_sum=$3 WHERE id=$4`,
		chapterAverage, chapterCount, chapterSum, rating.ChapterID.Hex())
	if err != nil {
		return 0, 0, 0, fmt.Errorf("update chapter averages: %w", err)
	}

	// Recalculate manga aggregates from all its chapters
	var mangaAverage float64
	var mangaTotalCount int64
	var mangaTotalSum float64

	err = r.store.DB.GetContext(ctx, &mangaAverage,
		`SELECT AVG(average_rating) FROM manga_chapters WHERE manga_id=$1 AND average_rating > 0`,
		rating.MangaID.Hex())
	if err != nil {
		mangaAverage = 0
	}

	err = r.store.DB.GetContext(ctx, &mangaTotalCount,
		`SELECT COALESCE(SUM(rating_count), 0) FROM manga_chapters WHERE manga_id=$1`,
		rating.MangaID.Hex())
	if err != nil {
		mangaTotalCount = 0
	}

	err = r.store.DB.GetContext(ctx, &mangaTotalSum,
		`SELECT COALESCE(SUM(rating_sum), 0) FROM manga_chapters WHERE manga_id=$1`,
		rating.MangaID.Hex())
	if err != nil {
		mangaTotalSum = 0
	}

	_, err = r.store.DB.ExecContext(ctx,
		`UPDATE mangas SET average_rating=$1, rating_count=$2, rating_sum=$3 WHERE id=$4`,
		mangaAverage, mangaTotalCount, mangaTotalSum, rating.MangaID.Hex())
	if err != nil {
		return 0, 0, 0, fmt.Errorf("update manga aggregates: %w", err)
	}

	return chapterAverage, int64(chapterCount), rating.Score, nil
}

func (r *PostgresMangaChapterRepository) HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return false, fmt.Errorf("postgres manga chapter repo not initialized")
	}
	var exists bool
	err := r.store.DB.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM chapter_ratings WHERE chapter_id=$1 AND user_id=$2)`,
		chapterID.Hex(), userID.Hex())
	if err != nil {
		return false, fmt.Errorf("check chapter rating: %w", err)
	}
	return exists, nil
}

func (r *PostgresMangaChapterRepository) GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return 0, false, fmt.Errorf("postgres manga chapter repo not initialized")
	}
	var score float64
	err := r.store.DB.GetContext(ctx, &score,
		`SELECT score::float8 FROM chapter_ratings WHERE chapter_id=$1 AND user_id=$2`,
		chapterID.Hex(), userID.Hex())
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0, false, nil // No rating found
		}
		return 0, false, fmt.Errorf("get chapter rating: %w", err)
	}
	return score, true, nil
}

func (r *PostgresMangaChapterRepository) AddChapterComment(ctx context.Context, comment *coremanga.ChapterComment) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}

	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO chapter_comments (id, chapter_id, manga_id, user_id, username, content, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		comment.ID.Hex(), comment.ChapterID.Hex(), comment.MangaID.Hex(), comment.UserID.Hex(), comment.Username, comment.Content, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add chapter comment: %w", err)
	}
	return nil
}

func (r *PostgresMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.ChapterComment, int64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, 0, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	// Get total count
	var total int64
	err := r.store.DB.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM chapter_comments WHERE chapter_id=$1`,
		chapterID.Hex())
	if err != nil {
		return nil, 0, fmt.Errorf("count chapter comments: %w", err)
	}

	// Determine sort order
	orderBy := "created_at DESC"
	if sortOrder == "oldest" {
		orderBy = "created_at ASC"
	}

	// Get paginated comments
	rows, err := r.store.DB.QueryxContext(ctx,
		fmt.Sprintf(`SELECT id, chapter_id, manga_id, user_id, username, content, created_at, updated_at
		 FROM chapter_comments
		 WHERE chapter_id=$1
		 ORDER BY %s
		 LIMIT $2 OFFSET $3`, orderBy),
		chapterID.Hex(), limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("list chapter comments: %w", err)
	}
	defer rows.Close()

	var comments []*coremanga.ChapterComment
	for rows.Next() {
		comment := &coremanga.ChapterComment{}
		var id, chapterID, mangaID, userID string
		err := rows.Scan(&id, &chapterID, &mangaID, &userID, &comment.Username, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan chapter comment: %w", err)
		}
		comment.ID, _ = primitive.ObjectIDFromHex(id)
		comment.ChapterID, _ = primitive.ObjectIDFromHex(chapterID)
		comment.MangaID, _ = primitive.ObjectIDFromHex(mangaID)
		comment.UserID, _ = primitive.ObjectIDFromHex(userID)
		comments = append(comments, comment)
	}

	return comments, total, nil
}

func (r *PostgresMangaChapterRepository) DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if r == nil || r.store == nil || r.store.DB == nil {
		return fmt.Errorf("postgres manga chapter repo not initialized")
	}

	result, err := r.store.DB.ExecContext(ctx,
		`DELETE FROM chapter_comments WHERE id=$1 AND user_id=$2`,
		commentID.Hex(), userID.Hex())
	if err != nil {
		return fmt.Errorf("delete chapter comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("chapter comment not found or unauthorized")
	}

	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/postgres/manga_chapter_repo.go -----

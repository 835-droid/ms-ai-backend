// ----- START OF FILE: backend/MS-AI/internal/data/postgres/manga_chapter_repo.go -----
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

type postgresMangaChapterRepository struct {
	store *PostgresStore
}

func NewPostgresChapterRepository(store *PostgresStore) coremanga.MangaChapterRepository {
	return &postgresMangaChapterRepository{store: store}
}

func (r *postgresMangaChapterRepository) CreateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
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

func (r *postgresMangaChapterRepository) GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*coremanga.MangaChapter, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return nil, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	var c coremanga.MangaChapter
	var idStr, mangaIdStr string
	var pagesStr sql.NullString
	query := `SELECT id, manga_id, title, pages, number, created_at, updated_at FROM manga_chapters WHERE id=$1`
	row := r.store.DB.QueryRowContext(ctx, query, id.Hex())
	if err := row.Scan(&idStr, &mangaIdStr, &c.Title, &pagesStr, &c.Number, &c.CreatedAt, &c.UpdatedAt); err != nil {
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

func (r *postgresMangaChapterRepository) ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaChapter, int64, error) {
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

func (r *postgresMangaChapterRepository) UpdateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
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

func (r *postgresMangaChapterRepository) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error {
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

func (r *postgresMangaChapterRepository) IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error {
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

	return nil
}

func (r *postgresMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (float64, error) {
	if r == nil || r.store == nil || r.store.DB == nil {
		return 0, fmt.Errorf("postgres manga chapter repo not initialized")
	}

	// Upsert rating
	_, err := r.store.DB.ExecContext(ctx,
		`INSERT INTO chapter_ratings (chapter_id, manga_id, user_id, score, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (chapter_id, user_id) DO UPDATE SET score=$4, created_at=$5`,
		rating.ChapterID.Hex(), rating.MangaID.Hex(), rating.UserID.Hex(), rating.Score, time.Now())
	if err != nil {
		return 0, fmt.Errorf("upsert chapter rating: %w", err)
	}

	// Calculate average rating for chapter
	var avgRating float64
	err = r.store.DB.GetContext(ctx, &avgRating,
		`SELECT AVG(score) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		avgRating = 0
	}

	// Update chapter with new average
	var count, sum float64
	err = r.store.DB.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		count = 0
	}
	err = r.store.DB.GetContext(ctx, &sum,
		`SELECT COALESCE(SUM(score), 0) FROM chapter_ratings WHERE chapter_id=$1`,
		rating.ChapterID.Hex())
	if err != nil {
		sum = 0
	}

	_, err = r.store.DB.ExecContext(ctx,
		`UPDATE manga_chapters SET average_rating=$1, rating_count=$2, rating_sum=$3 WHERE id=$4`,
		avgRating, count, sum, rating.ChapterID.Hex())
	if err != nil {
		return avgRating, fmt.Errorf("update chapter averages: %w", err)
	}

	// Recalculate manga average from all its chapters
	err = r.store.DB.GetContext(ctx, &avgRating,
		`SELECT AVG(average_rating) FROM manga_chapters WHERE manga_id=$1 AND average_rating > 0`,
		rating.MangaID.Hex())
	if err != nil {
		avgRating = 0
	}

	_, err = r.store.DB.ExecContext(ctx,
		`UPDATE mangas SET average_rating=$1 WHERE id=$2`,
		avgRating, rating.MangaID.Hex())
	if err != nil {
		return avgRating, fmt.Errorf("update manga average: %w", err)
	}

	return avgRating, nil
}

func (r *postgresMangaChapterRepository) HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
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

func (r *postgresMangaChapterRepository) AddChapterComment(ctx context.Context, comment *coremanga.ChapterComment) error {
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

func (r *postgresMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64) ([]*coremanga.ChapterComment, int64, error) {
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

	// Get paginated comments
	rows, err := r.store.DB.QueryxContext(ctx,
		`SELECT id, chapter_id, manga_id, user_id, username, content, created_at, updated_at
		 FROM chapter_comments
		 WHERE chapter_id=$1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
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

func (r *postgresMangaChapterRepository) DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
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

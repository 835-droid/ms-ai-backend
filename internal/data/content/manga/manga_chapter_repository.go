// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/manga_chapter_repository.go -----
package manga

import (
	"context"
	"errors"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	datamongo "github.com/835-droid/ms-ai-backend/internal/data/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMangaChapterRepository implements coremanga.MangaChapterRepository backed by MongoDB.
type MongoMangaChapterRepository struct {
	store *datamongo.MongoStore
	coll  *mongo.Collection
}

// NewMongoMangaChapterRepository creates a new MongoMangaChapterRepository.
func NewMongoMangaChapterRepository(s *datamongo.MongoStore) *MongoMangaChapterRepository {
	return &MongoMangaChapterRepository{
		store: s,
		coll:  s.GetCollection("manga_chapters"),
	}
}

func (r *MongoMangaChapterRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.coll == nil {
		return fmt.Errorf("mongo manga chapter repository not initialized")
	}
	return nil
}

// CreateMangaChapter creates a new manga chapter.
func (r *MongoMangaChapterRepository) CreateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	r.store.Log.Debug("creating manga chapter", map[string]interface{}{
		"manga_id": chapter.MangaID.Hex(),
		"title":    chapter.Title,
		"number":   chapter.Number,
	})

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.createMangaChapterInSession(sessCtx, chapter)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.createMangaChapterWithoutTransaction(ctx, chapter)
	}

	if err != nil {
		r.store.Log.Error("create chapter failed", map[string]interface{}{
			"manga_id":    chapter.MangaID.Hex(),
			"number":      chapter.Number,
			"replica_set": isReplicaSet,
			"error":       err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga chapter created", map[string]interface{}{
		"id":          chapter.ID.Hex(),
		"manga_id":    chapter.MangaID.Hex(),
		"title":       chapter.Title,
		"number":      chapter.Number,
		"replica_set": isReplicaSet,
	})
	return nil
}

// createMangaChapterInSession creates manga chapter within a transaction session
func (r *MongoMangaChapterRepository) createMangaChapterInSession(sessCtx mongo.SessionContext, chapter *coremanga.MangaChapter) error {
	// Check if chapter number exists for this manga
	count, err := r.coll.CountDocuments(sessCtx, bson.M{
		"manga_id": chapter.MangaID,
		"number":   chapter.Number,
	})
	if err != nil {
		return fmt.Errorf("check chapter number: %w", err)
	}
	if count > 0 {
		return corecommon.ErrInvalidInput
	}

	res, err := r.coll.InsertOne(sessCtx, chapter)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrInvalidInput
		}
		return fmt.Errorf("insert chapter: %w", err)
	}
	chapter.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// createMangaChapterWithoutTransaction creates manga chapter without transaction (for standalone MongoDB)
func (r *MongoMangaChapterRepository) createMangaChapterWithoutTransaction(ctx context.Context, chapter *coremanga.MangaChapter) error {
	// Check if chapter number exists for this manga
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"manga_id": chapter.MangaID,
		"number":   chapter.Number,
	})
	if err != nil {
		return fmt.Errorf("check chapter number: %w", err)
	}
	if count > 0 {
		return corecommon.ErrInvalidInput
	}

	res, err := r.coll.InsertOne(ctx, chapter)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrInvalidInput
		}
		return fmt.Errorf("insert chapter: %w", err)
	}
	chapter.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetMangaChapterByID retrieves a manga chapter by ID.
func (r *MongoMangaChapterRepository) GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*coremanga.MangaChapter, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "read")
	defer cancel()

	var chapter coremanga.MangaChapter
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&chapter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		return nil, fmt.Errorf("find chapter by id: %w", err)
	}
	return &chapter, nil
}

// ListMangaChaptersByManga retrieves a paginated list of chapters for a manga.
func (r *MongoMangaChapterRepository) ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaChapter, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "read")
	defer cancel()

	filter := bson.M{"manga_id": mangaID}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "number", Value: 1}})

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count chapters: %w", err)
	}

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find chapters: %w", err)
	}
	defer cur.Close(ctx)

	var chapters []*coremanga.MangaChapter
	if err := cur.All(ctx, &chapters); err != nil {
		return nil, 0, fmt.Errorf("decode chapters: %w", err)
	}
	return chapters, total, nil
}

// UpdateMangaChapter updates an existing manga chapter.
func (r *MongoMangaChapterRepository) UpdateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	chapter.UpdatedAt = time.Now()

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.updateMangaChapterInSession(sessCtx, chapter)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.updateMangaChapterWithoutTransaction(ctx, chapter)
	}

	if err != nil {
		r.store.Log.Error("update chapter failed", map[string]interface{}{
			"id":          chapter.ID.Hex(),
			"replica_set": isReplicaSet,
			"error":       err.Error(),
		})
		return err
	}
	return nil
}

// updateMangaChapterInSession updates manga chapter within a transaction session
func (r *MongoMangaChapterRepository) updateMangaChapterInSession(sessCtx mongo.SessionContext, chapter *coremanga.MangaChapter) error {
	count, err := r.coll.CountDocuments(sessCtx, bson.M{
		"_id":      bson.M{"$ne": chapter.ID},
		"manga_id": chapter.MangaID,
		"number":   chapter.Number,
	})
	if err != nil {
		return fmt.Errorf("check chapter number: %w", err)
	}
	if count > 0 {
		return corecommon.ErrInvalidInput
	}

	result, err := r.coll.ReplaceOne(sessCtx, bson.M{"_id": chapter.ID}, chapter)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrInvalidInput
		}
		return fmt.Errorf("update chapter: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// updateMangaChapterWithoutTransaction updates manga chapter without transaction (for standalone MongoDB)
func (r *MongoMangaChapterRepository) updateMangaChapterWithoutTransaction(ctx context.Context, chapter *coremanga.MangaChapter) error {
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"_id":      bson.M{"$ne": chapter.ID},
		"manga_id": chapter.MangaID,
		"number":   chapter.Number,
	})
	if err != nil {
		return fmt.Errorf("check chapter number: %w", err)
	}
	if count > 0 {
		return corecommon.ErrInvalidInput
	}

	result, err := r.coll.ReplaceOne(ctx, bson.M{"_id": chapter.ID}, chapter)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrInvalidInput
		}
		return fmt.Errorf("update chapter: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// DeleteMangaChapter deletes a manga chapter.
func (r *MongoMangaChapterRepository) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	if result.DeletedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/content/manga/manga_chapter_repository.go -----

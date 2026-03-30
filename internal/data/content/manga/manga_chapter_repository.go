package data

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

// MongoMangaChapterRepository implements core.MangaChapterRepository backed by MongoDB.
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

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
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

		// Insert the chapter
		res, err := r.coll.InsertOne(sessCtx, chapter)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return corecommon.ErrInvalidInput
			}
			return fmt.Errorf("insert chapter: %w", err)
		}
		chapter.ID = res.InsertedID.(primitive.ObjectID)
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("create chapter failed", map[string]interface{}{
			"manga_id": chapter.MangaID.Hex(),
			"number":   chapter.Number,
			"error":    err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga chapter created", map[string]interface{}{
		"id":       chapter.ID.Hex(),
		"manga_id": chapter.MangaID.Hex(),
		"title":    chapter.Title,
		"number":   chapter.Number,
	})
	return nil
}

// GetMangaChapterByID retrieves a manga chapter by ID.
func (r *MongoMangaChapterRepository) GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*coremanga.MangaChapter, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "read")
	defer cancel()

	r.store.Log.Debug("getting manga chapter", map[string]interface{}{
		"id": id.Hex(),
	})

	var chapter coremanga.MangaChapter
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&chapter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		r.store.Log.Error("get chapter failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
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

	r.store.Log.Debug("listing manga chapters", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"skip":     skip,
		"limit":    limit,
	})

	filter := bson.M{"manga_id": mangaID}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "number", Value: 1}})

	// Get total count
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
		r.store.Log.Error("decode chapter list failed", map[string]interface{}{"error": err.Error()})
		return nil, 0, fmt.Errorf("decode chapters: %w", err)
	}

	r.store.Log.Info("manga chapters listed", map[string]interface{}{
		"count":    len(chapters),
		"total":    total,
		"manga_id": mangaID.Hex(),
		"skip":     skip,
		"limit":    limit,
	})
	return chapters, total, nil
}

// UpdateMangaChapter updates an existing manga chapter.
func (r *MongoMangaChapterRepository) UpdateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	r.store.Log.Debug("updating manga chapter", map[string]interface{}{
		"id":       chapter.ID.Hex(),
		"manga_id": chapter.MangaID.Hex(),
		"title":    chapter.Title,
		"number":   chapter.Number,
	})

	chapter.UpdatedAt = time.Now()

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Check if chapter number exists for this manga (except this chapter)
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

		// Update the chapter
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
	}, nil)

	if err != nil {
		r.store.Log.Error("update chapter failed", map[string]interface{}{
			"id":    chapter.ID.Hex(),
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga chapter updated", map[string]interface{}{
		"id":       chapter.ID.Hex(),
		"manga_id": chapter.MangaID.Hex(),
		"title":    chapter.Title,
		"number":   chapter.Number,
	})
	return nil
}

// DeleteMangaChapter deletes a manga chapter.
func (r *MongoMangaChapterRepository) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	r.store.Log.Debug("deleting manga chapter", map[string]interface{}{
		"id": id.Hex(),
	})

	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.store.Log.Error("delete chapter failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return fmt.Errorf("delete chapter: %w", err)
	}

	if result.DeletedCount == 0 {
		return corecommon.ErrNotFound
	}

	r.store.Log.Info("manga chapter deleted", map[string]interface{}{
		"id": id.Hex(),
	})
	return nil
}

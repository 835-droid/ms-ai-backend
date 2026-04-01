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

// MongoMangaRepository implements coremanga.MangaRepository backed by MongoDB.
type MongoMangaRepository struct {
	store *datamongo.MongoStore
	coll  *mongo.Collection
}

// NewMongoMangaRepository creates a new MongoMangaRepository.
func NewMongoMangaRepository(s *datamongo.MongoStore) *MongoMangaRepository {
	return &MongoMangaRepository{
		store: s,
		coll:  s.GetCollection("manga"),
	}
}

func (r *MongoMangaRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.coll == nil {
		return fmt.Errorf("mongo manga repository not initialized")
	}
	return nil
}

// CreateManga creates a new manga.
func (r *MongoMangaRepository) CreateManga(ctx context.Context, manga *coremanga.Manga) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	r.store.Log.Debug("creating manga", map[string]interface{}{
		"title":     manga.Title,
		"author_id": manga.AuthorID.Hex(),
	})

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Check if slug exists
		count, err := r.coll.CountDocuments(sessCtx, bson.M{"slug": manga.Slug})
		if err != nil {
			return fmt.Errorf("check slug: %w", err)
		}
		if count > 0 {
			return corecommon.ErrSlugExists
		}

		// Insert the manga
		res, err := r.coll.InsertOne(sessCtx, manga)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return corecommon.ErrSlugExists
			}
			return fmt.Errorf("insert manga: %w", err)
		}
		manga.ID = res.InsertedID.(primitive.ObjectID)
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("create manga failed", map[string]interface{}{
			"title": manga.Title,
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga created", map[string]interface{}{
		"title": manga.Title,
		"id":    manga.ID.Hex(),
	})
	return nil
}

// GetMangaByID retrieves a manga by ID.
func (r *MongoMangaRepository) GetMangaByID(ctx context.Context, id primitive.ObjectID) (*coremanga.Manga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	var manga coremanga.Manga
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&manga)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		r.store.Log.Error("get manga failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, fmt.Errorf("find manga by id: %w", err)
	}
	return &manga, nil
}

// GetMangaBySlug retrieves a manga by slug.
func (r *MongoMangaRepository) GetMangaBySlug(ctx context.Context, slug string) (*coremanga.Manga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	var manga coremanga.Manga
	err := r.coll.FindOne(ctx, bson.M{"slug": slug}).Decode(&manga)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		r.store.Log.Error("find manga by slug failed", map[string]interface{}{
			"slug":  slug,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("find manga by slug: %w", err)
	}
	return &manga, nil
}

// ListMangas retrieves a paginated list of manga.
func (r *MongoMangaRepository) ListMangas(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	filter := bson.M{}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count mangas: %w", err)
	}

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find mangas: %w", err)
	}
	defer cur.Close(ctx)

	var mangas []*coremanga.Manga
	if err := cur.All(ctx, &mangas); err != nil {
		return nil, 0, fmt.Errorf("decode mangas: %w", err)
	}

	return mangas, total, nil
}

// UpdateManga updates an existing manga.
func (r *MongoMangaRepository) UpdateManga(ctx context.Context, manga *coremanga.Manga) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	manga.UpdatedAt = time.Now()

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Check if slug exists (except for this manga)
		count, err := r.coll.CountDocuments(sessCtx, bson.M{
			"_id":  bson.M{"$ne": manga.ID},
			"slug": manga.Slug,
		})
		if err != nil {
			return fmt.Errorf("check slug: %w", err)
		}
		if count > 0 {
			return corecommon.ErrSlugExists
		}

		result, err := r.coll.ReplaceOne(sessCtx, bson.M{"_id": manga.ID}, manga)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return corecommon.ErrSlugExists
			}
			return fmt.Errorf("update manga: %w", err)
		}
		if result.MatchedCount == 0 {
			return corecommon.ErrNotFound
		}
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("update manga failed", map[string]interface{}{
			"id":    manga.ID.Hex(),
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga updated", map[string]interface{}{
		"id":    manga.ID.Hex(),
		"title": manga.Title,
	})
	return nil
}

// DeleteManga deletes a manga.
func (r *MongoMangaRepository) DeleteManga(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	if result.DeletedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

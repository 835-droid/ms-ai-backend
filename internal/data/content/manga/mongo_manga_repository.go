// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/mongo_manga_repository.go -----
package manga

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMangaRepository implements coremanga.MangaRepository backed by MongoDB.
type MongoMangaRepository struct {
	store        *mongoinfra.MongoStore
	coll         *mongo.Collection
	reactionLock sync.Map // key: "mangaID_userID" for per-user reaction throttling
	ratingLock   sync.Map // key: "mangaID_userID" for per-user rating throttling
}

// NewMongoMangaRepository creates a new MongoMangaRepository.
func NewMongoMangaRepository(s *mongoinfra.MongoStore) *MongoMangaRepository {
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

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.createMangaInSession(sessCtx, manga)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.createMangaWithoutTransaction(ctx, manga)
	}

	if err != nil {
		r.store.Log.Error("create manga failed", map[string]interface{}{
			"title": manga.Title,
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga created", map[string]interface{}{
		"title":       manga.Title,
		"id":          manga.ID.Hex(),
		"replica_set": isReplicaSet,
	})
	return nil
}

// createMangaInSession creates manga within a transaction session
func (r *MongoMangaRepository) createMangaInSession(sessCtx mongo.SessionContext, manga *coremanga.Manga) error {
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
}

// createMangaWithoutTransaction creates manga without transaction (for standalone MongoDB)
func (r *MongoMangaRepository) createMangaWithoutTransaction(ctx context.Context, manga *coremanga.Manga) error {
	// Check if slug exists
	count, err := r.coll.CountDocuments(ctx, bson.M{"slug": manga.Slug})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	// Insert the manga
	res, err := r.coll.InsertOne(ctx, manga)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("insert manga: %w", err)
	}
	manga.ID = res.InsertedID.(primitive.ObjectID)
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

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.updateMangaInSession(sessCtx, manga)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.updateMangaWithoutTransaction(ctx, manga)
	}

	if err != nil {
		r.store.Log.Error("update manga failed", map[string]interface{}{
			"id":    manga.ID.Hex(),
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("manga updated", map[string]interface{}{
		"id":          manga.ID.Hex(),
		"title":       manga.Title,
		"replica_set": isReplicaSet,
	})
	return nil
}

// updateMangaInSession updates manga within a transaction session
func (r *MongoMangaRepository) updateMangaInSession(sessCtx mongo.SessionContext, manga *coremanga.Manga) error {
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
}

// updateMangaWithoutTransaction updates manga without transaction (for standalone MongoDB)
func (r *MongoMangaRepository) updateMangaWithoutTransaction(ctx context.Context, manga *coremanga.Manga) error {
	// Check if slug exists (except for this manga)
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"_id":  bson.M{"$ne": manga.ID},
		"slug": manga.Slug,
	})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	result, err := r.coll.ReplaceOne(ctx, bson.M{"_id": manga.ID}, manga)
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

// IncrementViews increments the view count for a manga.
func (r *MongoMangaRepository) IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	result, err := r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{"$inc": bson.M{"views_count": 1}})
	if err != nil {
		r.store.Log.Error("increment views failed", map[string]interface{}{
			"manga_id": mangaID.Hex(),
			"error":    err.Error(),
		})
		return fmt.Errorf("increment views: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

func (r *MongoMangaRepository) LogView(ctx context.Context, mangaID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_view_logs", "write")
	defer cancel()

	// Store manga_id as primitive.ObjectID (not string) to match manga._id type
	// This ensures $lookup joins work correctly in ListMostViewed
	viewLog := bson.M{
		"manga_id":  mangaID,
		"viewed_at": time.Now(),
	}
	_, err := r.store.Client.Database(r.store.DBName).Collection("manga_view_logs").InsertOne(ctx, viewLog)
	if err != nil {
		return fmt.Errorf("log view: %w", err)
	}
	return r.IncrementViews(ctx, mangaID)
}

func (r *MongoMangaRepository) ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*coremanga.RankedManga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	// Special case for "all time" - use the persisted views_count instead of aggregating logs
	// This ensures mangas with historical views but no post-deployment logs still appear
	if since.IsZero() {
		ctx2, cancel2 := r.store.WithCollectionTimeout(ctx, "manga", "read")
		defer cancel2()

		findOptions := options.Find()
		findOptions.SetSort(bson.M{"views_count": -1})
		findOptions.SetSkip(skip)
		findOptions.SetLimit(limit)

		cursor, err := r.coll.Find(ctx2, bson.M{"is_published": true}, findOptions)
		if err != nil {
			return nil, fmt.Errorf("find most viewed (all time): %w", err)
		}
		defer cursor.Close(ctx2)

		var mangas []*coremanga.Manga
		if err := cursor.All(ctx2, &mangas); err != nil {
			return nil, fmt.Errorf("decode mangas: %w", err)
		}

		rankedMangas := make([]*coremanga.RankedManga, 0, len(mangas))
		for _, m := range mangas {
			rankedMangas = append(rankedMangas, &coremanga.RankedManga{
				Manga:     m,
				ViewCount: m.ViewsCount,
			})
		}
		return rankedMangas, nil
	}

	// For time-bounded periods, aggregate from manga_view_logs
	// Handle both ObjectID (new) and string (legacy) manga_id types for backward compatibility
	// Use $addFields to normalize manga_id to string, then $lookup using string comparison
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_view_logs", "read")
	defer cancel()

	// Pipeline that handles both ObjectID and string manga_id types:
	// 1. Match by date range
	// 2. Add manga_id_str field that converts ObjectID to string or keeps existing string
	// 3. Group by the normalized string ID
	// 4. Lookup manga using $toString on manga._id to match our string IDs
	// 5. Filter for published manga
	// 6. Sort, skip, limit
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"viewed_at": bson.M{"$gte": since}}}},
		{{Key: "$addFields", Value: bson.M{
			"manga_id_str": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$type": "$manga_id"},
					"then": bson.M{"$toString": "$manga_id"},
					"else": "$manga_id",
				},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$manga_id_str",
			"view_count": bson.M{"$sum": 1},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "manga",
			"let":  bson.M{"mangaIdStr": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{bson.M{"$toString": "$_id"}, "$$mangaIdStr"}},
							bson.M{"$eq": bson.A{"$is_published", true}},
						},
					},
				}}},
			},
			"as": "manga",
		}}},
		{{Key: "$match", Value: bson.M{"manga": bson.M{"$ne": bson.A{}}}}},
		{{Key: "$sort", Value: bson.M{"view_count": -1}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.store.Client.Database(r.store.DBName).Collection("manga_view_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate most viewed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode aggregation results: %w", err)
	}

	if len(results) == 0 {
		return []*coremanga.RankedManga{}, nil
	}

	// Build ordered list of manga IDs and view counts
	// Handle both string IDs (from aggregation) and convert to ObjectID for fetching
	type mangaRank struct {
		IDString  string
		ID        primitive.ObjectID
		ViewCount int64
	}
	var ranks []mangaRank
	for _, result := range results {
		if mangaIDStr, ok := result["_id"].(string); ok {
			oid, err := primitive.ObjectIDFromHex(mangaIDStr)
			if err != nil {
				// Skip invalid ObjectID strings
				continue
			}
			viewCount := int64(0)
			if vc, ok := result["view_count"].(int); ok {
				viewCount = int64(vc)
			} else if vc, ok := result["view_count"].(int64); ok {
				viewCount = vc
			} else if vc, ok := result["view_count"].(int32); ok {
				viewCount = int64(vc)
			}
			ranks = append(ranks, mangaRank{IDString: mangaIDStr, ID: oid, ViewCount: viewCount})
		}
	}

	if len(ranks) == 0 {
		return []*coremanga.RankedManga{}, nil
	}

	// Build separate []primitive.ObjectID slice for the $in query
	ids := make([]primitive.ObjectID, 0, len(ranks))
	for _, rank := range ranks {
		ids = append(ids, rank.ID)
	}

	// Fetch manga details (only published ones)
	ctx2, cancel2 := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel2()

	cursor2, err := r.coll.Find(ctx2, bson.M{"_id": bson.M{"$in": ids}, "is_published": true})
	if err != nil {
		return nil, fmt.Errorf("find mangas: %w", err)
	}
	defer cursor2.Close(ctx2)

	var mangas []*coremanga.Manga
	if err := cursor2.All(ctx2, &mangas); err != nil {
		return nil, fmt.Errorf("decode mangas: %w", err)
	}

	// Reorder mangas to match the aggregation order
	mangaMap := make(map[string]*coremanga.Manga)
	for _, m := range mangas {
		mangaMap[m.ID.Hex()] = m
	}

	orderedRanked := make([]*coremanga.RankedManga, 0, len(ranks))
	for _, rank := range ranks {
		if m, ok := mangaMap[rank.IDString]; ok {
			orderedRanked = append(orderedRanked, &coremanga.RankedManga{
				Manga:     m,
				ViewCount: rank.ViewCount,
			})
		}
	}

	return orderedRanked, nil
}

func (r *MongoMangaRepository) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"updated_at": -1})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{"is_published": true}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find recently updated: %w", err)
	}
	defer cursor.Close(ctx)

	var mangas []*coremanga.Manga
	if err := cursor.All(ctx, &mangas); err != nil {
		return nil, fmt.Errorf("decode mangas: %w", err)
	}

	return mangas, nil
}

func (r *MongoMangaRepository) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"favorites_count": -1})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{"is_published": true}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find most followed: %w", err)
	}
	defer cursor.Close(ctx)

	var mangas []*coremanga.Manga
	if err := cursor.All(ctx, &mangas); err != nil {
		return nil, fmt.Errorf("decode mangas: %w", err)
	}

	return mangas, nil
}

func (r *MongoMangaRepository) ListTopRated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"average_rating": -1, "rating_count": -1})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{"is_published": true, "rating_count": bson.M{"$gt": 0}}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find top rated: %w", err)
	}
	defer cursor.Close(ctx)

	var mangas []*coremanga.Manga
	if err := cursor.All(ctx, &mangas); err != nil {
		return nil, fmt.Errorf("decode mangas: %w", err)
	}

	return mangas, nil
}

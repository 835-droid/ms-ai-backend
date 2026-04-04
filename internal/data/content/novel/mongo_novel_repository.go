// Package novel implements novel data access operations backed by MongoDB.
package novel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	corenovel "github.com/835-droid/ms-ai-backend/internal/core/content/novel"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoNovelRepository implements corenovel.NovelRepository backed by MongoDB.
type MongoNovelRepository struct {
	store        *mongoinfra.MongoStore
	coll         *mongo.Collection
	reactionLock sync.Map // key: "novelID_userID" for per-user reaction throttling
	ratingLock   sync.Map // key: "novelID_userID" for per-user rating throttling
}

// NewMongoNovelRepository creates a new MongoNovelRepository.
func NewMongoNovelRepository(s *mongoinfra.MongoStore) *MongoNovelRepository {
	return &MongoNovelRepository{
		store: s,
		coll:  s.GetCollection("novel"),
	}
}

func (r *MongoNovelRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.coll == nil {
		return fmt.Errorf("mongo novel repository not initialized")
	}
	return nil
}

// CreateNovel creates a new novel.
func (r *MongoNovelRepository) CreateNovel(ctx context.Context, novel *corenovel.Novel) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "write")
	defer cancel()

	r.store.Log.Debug("creating novel", map[string]interface{}{
		"title":     novel.Title,
		"author_id": novel.AuthorID.Hex(),
	})

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.createNovelInSession(sessCtx, novel)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.createNovelWithoutTransaction(ctx, novel)
	}

	if err != nil {
		r.store.Log.Error("create novel failed", map[string]interface{}{
			"title": novel.Title,
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("novel created", map[string]interface{}{
		"title":       novel.Title,
		"id":          novel.ID.Hex(),
		"replica_set": isReplicaSet,
	})
	return nil
}

// createNovelInSession creates novel within a transaction session
func (r *MongoNovelRepository) createNovelInSession(sessCtx mongo.SessionContext, novel *corenovel.Novel) error {
	// Check if slug exists
	count, err := r.coll.CountDocuments(sessCtx, bson.M{"slug": novel.Slug})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	// Insert the novel
	res, err := r.coll.InsertOne(sessCtx, novel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("insert novel: %w", err)
	}
	novel.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// createNovelWithoutTransaction creates novel without transaction (for standalone MongoDB)
func (r *MongoNovelRepository) createNovelWithoutTransaction(ctx context.Context, novel *corenovel.Novel) error {
	// Check if slug exists
	count, err := r.coll.CountDocuments(ctx, bson.M{"slug": novel.Slug})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	// Insert the novel
	res, err := r.coll.InsertOne(ctx, novel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("insert novel: %w", err)
	}
	novel.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetNovelByID retrieves a novel by ID.
func (r *MongoNovelRepository) GetNovelByID(ctx context.Context, id primitive.ObjectID) (*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	var novel corenovel.Novel
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&novel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		r.store.Log.Error("get novel failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, fmt.Errorf("find novel by id: %w", err)
	}
	return &novel, nil
}

// GetNovelBySlug retrieves a novel by slug.
func (r *MongoNovelRepository) GetNovelBySlug(ctx context.Context, slug string) (*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	var novel corenovel.Novel
	err := r.coll.FindOne(ctx, bson.M{"slug": slug}).Decode(&novel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, corecommon.ErrNotFound
		}
		r.store.Log.Error("find novel by slug failed", map[string]interface{}{
			"slug":  slug,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("find novel by slug: %w", err)
	}
	return &novel, nil
}

// ListNovels retrieves a paginated list of novels.
func (r *MongoNovelRepository) ListNovels(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	filter := bson.M{}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count novels: %w", err)
	}

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find novels: %w", err)
	}
	defer cur.Close(ctx)

	var novels []*corenovel.Novel
	if err := cur.All(ctx, &novels); err != nil {
		return nil, 0, fmt.Errorf("decode novels: %w", err)
	}

	return novels, total, nil
}

// UpdateNovel updates an existing novel.
func (r *MongoNovelRepository) UpdateNovel(ctx context.Context, novel *corenovel.Novel) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "write")
	defer cancel()

	novel.UpdatedAt = time.Now()

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.updateNovelInSession(sessCtx, novel)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.updateNovelWithoutTransaction(ctx, novel)
	}

	if err != nil {
		r.store.Log.Error("update novel failed", map[string]interface{}{
			"id":    novel.ID.Hex(),
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("novel updated", map[string]interface{}{
		"id":          novel.ID.Hex(),
		"title":       novel.Title,
		"replica_set": isReplicaSet,
	})
	return nil
}

// updateNovelInSession updates novel within a transaction session
func (r *MongoNovelRepository) updateNovelInSession(sessCtx mongo.SessionContext, novel *corenovel.Novel) error {
	// Check if slug exists (except for this novel)
	count, err := r.coll.CountDocuments(sessCtx, bson.M{
		"_id":  bson.M{"$ne": novel.ID},
		"slug": novel.Slug,
	})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	result, err := r.coll.ReplaceOne(sessCtx, bson.M{"_id": novel.ID}, novel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("update novel: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// updateNovelWithoutTransaction updates novel without transaction (for standalone MongoDB)
func (r *MongoNovelRepository) updateNovelWithoutTransaction(ctx context.Context, novel *corenovel.Novel) error {
	// Check if slug exists (except for this novel)
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"_id":  bson.M{"$ne": novel.ID},
		"slug": novel.Slug,
	})
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if count > 0 {
		return corecommon.ErrSlugExists
	}

	result, err := r.coll.ReplaceOne(ctx, bson.M{"_id": novel.ID}, novel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return corecommon.ErrSlugExists
		}
		return fmt.Errorf("update novel: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// DeleteNovel deletes a novel.
func (r *MongoNovelRepository) DeleteNovel(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "write")
	defer cancel()

	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete novel: %w", err)
	}
	if result.DeletedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// IncrementViews increments the view count for a novel.
func (r *MongoNovelRepository) IncrementViews(ctx context.Context, novelID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "write")
	defer cancel()

	result, err := r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{"$inc": bson.M{"views_count": 1}})
	if err != nil {
		r.store.Log.Error("increment views failed", map[string]interface{}{
			"novel_id": novelID.Hex(),
			"error":    err.Error(),
		})
		return fmt.Errorf("increment views: %w", err)
	}
	if result.MatchedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

// LogView logs a view for a novel.
func (r *MongoNovelRepository) LogView(ctx context.Context, novelID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_view_logs", "write")
	defer cancel()

	viewLog := bson.M{
		"novel_id":  novelID,
		"viewed_at": time.Now(),
	}
	_, err := r.store.Client.Database(r.store.DBName).Collection("novel_view_logs").InsertOne(ctx, viewLog)
	if err != nil {
		return fmt.Errorf("log view: %w", err)
	}
	return r.IncrementViews(ctx, novelID)
}

// ListMostViewed returns the most viewed novels for a given time period.
func (r *MongoNovelRepository) ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*corenovel.RankedNovel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	// Special case for "all time" - use the persisted views_count instead of aggregating logs
	if since.IsZero() {
		ctx2, cancel2 := r.store.WithCollectionTimeout(ctx, "novel", "read")
		defer cancel2()

		findOptions := options.Find()
		findOptions.SetSort(bson.M{"views_count": -1})
		findOptions.SetSkip(skip)
		findOptions.SetLimit(limit)

		cursor, err := r.coll.Find(ctx2, bson.M{}, findOptions)
		if err != nil {
			return nil, fmt.Errorf("find most viewed (all time): %w", err)
		}
		defer cursor.Close(ctx2)

		var novels []*corenovel.Novel
		if err := cursor.All(ctx2, &novels); err != nil {
			return nil, fmt.Errorf("decode novels: %w", err)
		}

		rankedNovels := make([]*corenovel.RankedNovel, 0, len(novels))
		for _, n := range novels {
			rankedNovels = append(rankedNovels, &corenovel.RankedNovel{
				Novel:     n,
				ViewCount: n.ViewsCount,
			})
		}
		return rankedNovels, nil
	}

	// For time-bounded periods, aggregate from novel_view_logs
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_view_logs", "read")
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"viewed_at": bson.M{"$gte": since}}}},
		{{Key: "$addFields", Value: bson.M{
			"novel_id_str": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$type": "$novel_id"},
					"then": bson.M{"$toString": "$novel_id"},
					"else": "$novel_id",
				},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$novel_id_str",
			"view_count": bson.M{"$sum": 1},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "novel",
			"let":  bson.M{"novelIdStr": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"$expr": bson.M{
						"$eq": bson.A{bson.M{"$toString": "$_id"}, "$$novelIdStr"},
					},
				}}},
			},
			"as": "novel",
		}}},
		{{Key: "$match", Value: bson.M{"novel": bson.M{"$ne": bson.A{}}}}},
		{{Key: "$sort", Value: bson.M{"view_count": -1}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.store.Client.Database(r.store.DBName).Collection("novel_view_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate most viewed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode aggregation results: %w", err)
	}

	if len(results) == 0 {
		return []*corenovel.RankedNovel{}, nil
	}

	type novelRank struct {
		IDString  string
		ID        primitive.ObjectID
		ViewCount int64
	}
	var ranks []novelRank
	for _, result := range results {
		if novelIDStr, ok := result["_id"].(string); ok {
			oid, err := primitive.ObjectIDFromHex(novelIDStr)
			if err != nil {
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
			ranks = append(ranks, novelRank{IDString: novelIDStr, ID: oid, ViewCount: viewCount})
		}
	}

	if len(ranks) == 0 {
		return []*corenovel.RankedNovel{}, nil
	}

	ids := make([]primitive.ObjectID, 0, len(ranks))
	for _, rank := range ranks {
		ids = append(ids, rank.ID)
	}

	ctx2, cancel2 := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel2()

	cursor2, err := r.coll.Find(ctx2, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("find novels: %w", err)
	}
	defer cursor2.Close(ctx2)

	var novels []*corenovel.Novel
	if err := cursor2.All(ctx2, &novels); err != nil {
		return nil, fmt.Errorf("decode novels: %w", err)
	}

	novelMap := make(map[string]*corenovel.Novel)
	for _, n := range novels {
		novelMap[n.ID.Hex()] = n
	}

	orderedRanked := make([]*corenovel.RankedNovel, 0, len(ranks))
	for _, rank := range ranks {
		if n, ok := novelMap[rank.IDString]; ok {
			orderedRanked = append(orderedRanked, &corenovel.RankedNovel{
				Novel:     n,
				ViewCount: rank.ViewCount,
			})
		}
	}

	return orderedRanked, nil
}

// ListRecentlyUpdated returns recently updated novels.
func (r *MongoNovelRepository) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"updated_at": -1})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find recently updated: %w", err)
	}
	defer cursor.Close(ctx)

	var novels []*corenovel.Novel
	if err := cursor.All(ctx, &novels); err != nil {
		return nil, fmt.Errorf("decode novels: %w", err)
	}

	return novels, nil
}

// ListMostFollowed returns the most followed novels.
func (r *MongoNovelRepository) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"favorites_count": -1})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find most followed: %w", err)
	}
	defer cursor.Close(ctx)

	var novels []*corenovel.Novel
	if err := cursor.All(ctx, &novels); err != nil {
		return nil, fmt.Errorf("decode novels: %w", err)
	}

	return novels, nil
}

// ListTopRated returns the top rated novels.
func (r *MongoNovelRepository) ListTopRated(ctx context.Context, skip, limit int64) ([]*corenovel.Novel, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel", "read")
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "average_rating", Value: -1}, {Key: "rating_count", Value: -1}})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := r.coll.Find(ctx, bson.M{"rating_count": bson.M{"$gt": 0}}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find top rated: %w", err)
	}
	defer cursor.Close(ctx)

	var novels []*corenovel.Novel
	if err := cursor.All(ctx, &novels); err != nil {
		return nil, fmt.Errorf("decode novels: %w", err)
	}

	return novels, nil
}

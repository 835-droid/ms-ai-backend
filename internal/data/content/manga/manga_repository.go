// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/manga_repository.go -----
package manga

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	store        *datamongo.MongoStore
	coll         *mongo.Collection
	reactionLock sync.Map // key: "mangaID_userID" for per-user reaction throttling
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

// SetReaction sets or toggles a reaction for a manga by a user.
func (r *MongoMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (reaction string, err error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	// Implement per-(manga,user) locking to prevent concurrent requests
	lockKey := fmt.Sprintf("%s_%s", mangaID.Hex(), userID.Hex())
	if _, loaded := r.reactionLock.LoadOrStore(lockKey, true); loaded {
		// Another request is already processing for this user-manga pair
		return "", fmt.Errorf("reaction request already in progress")
	}
	defer r.reactionLock.Delete(lockKey)

	reactionsColl := r.store.GetCollection("manga_reactions")

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			reaction, err = r.setReactionInSession(sessCtx, mangaID, userID, reactionType, reactionsColl)
			return err
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		reaction, err = r.setReactionWithoutTransaction(ctx, mangaID, userID, reactionType, reactionsColl)
	}

	if err != nil {
		r.store.Log.Error("set reaction failed", map[string]interface{}{
			"manga_id":      mangaID.Hex(),
			"user_id":       userID.Hex(),
			"reaction_type": string(reactionType),
			"replica_set":   isReplicaSet,
			"error":         err.Error(),
		})
		return "", err
	}
	return reaction, nil
}

// GetUserReaction gets the current reaction type for a user on a manga.
func (r *MongoMangaRepository) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_reactions", "read")
	defer cancel()

	reactionsColl := r.store.GetCollection("manga_reactions")

	var reaction coremanga.MangaReaction
	err := reactionsColl.FindOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID}).Decode(&reaction)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", fmt.Errorf("find user reaction: %w", err)
	}

	return string(reaction.Type), nil
}

// toggleLikeInSession toggles like within a transaction session
func (r *MongoMangaRepository) toggleLikeInSession(sessCtx mongo.SessionContext, mangaID, userID primitive.ObjectID, likesColl *mongo.Collection) (bool, error) {
	// First check if manga exists
	mangaCount, err := r.coll.CountDocuments(sessCtx, bson.M{"_id": mangaID})
	if err != nil {
		return false, fmt.Errorf("check manga exists: %w", err)
	}
	if mangaCount == 0 {
		return false, corecommon.ErrNotFound
	}

	// Check if like exists
	count, err := likesColl.CountDocuments(sessCtx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("check like exists: %w", err)
	}

	if count > 0 {
		// Remove like
		_, err = likesColl.DeleteOne(sessCtx, bson.M{"manga_id": mangaID, "user_id": userID})
		if err != nil {
			return false, fmt.Errorf("remove like: %w", err)
		}
		// Decrement likes count
		result, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": mangaID}, bson.M{"$inc": bson.M{"likes_count": -1}})
		if err != nil {
			return false, fmt.Errorf("decrement likes count: %w", err)
		}
		if result.MatchedCount == 0 {
			return false, corecommon.ErrNotFound
		}
		return false, nil
	} else {
		// Add like
		_, err = likesColl.InsertOne(sessCtx, bson.M{"manga_id": mangaID, "user_id": userID})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Race condition, treat as already liked
				return true, nil
			}
			return false, fmt.Errorf("add like: %w", err)
		}
		// Increment likes count
		result, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": mangaID}, bson.M{"$inc": bson.M{"likes_count": 1}})
		if err != nil {
			return false, fmt.Errorf("increment likes count: %w", err)
		}
		if result.MatchedCount == 0 {
			return false, corecommon.ErrNotFound
		}
		return true, nil
	}
}

// toggleLikeWithoutTransaction toggles like without transaction (for standalone MongoDB)
func (r *MongoMangaRepository) toggleLikeWithoutTransaction(ctx context.Context, mangaID, userID primitive.ObjectID, likesColl *mongo.Collection) (bool, error) {
	// First check if manga exists
	mangaCount, err := r.coll.CountDocuments(ctx, bson.M{"_id": mangaID})
	if err != nil {
		return false, fmt.Errorf("check manga exists: %w", err)
	}
	if mangaCount == 0 {
		return false, corecommon.ErrNotFound
	}

	// Check if like exists
	count, err := likesColl.CountDocuments(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("check like exists: %w", err)
	}

	if count > 0 {
		// Remove like
		_, err = likesColl.DeleteOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
		if err != nil {
			return false, fmt.Errorf("remove like: %w", err)
		}
		// Decrement likes count
		result, err := r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{"$inc": bson.M{"likes_count": -1}})
		if err != nil {
			return false, fmt.Errorf("decrement likes count: %w", err)
		}
		if result.MatchedCount == 0 {
			return false, corecommon.ErrNotFound
		}
		return false, nil
	} else {
		// Add like
		_, err = likesColl.InsertOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Race condition, treat as already liked
				return true, nil
			}
			return false, fmt.Errorf("add like: %w", err)
		}
		// Increment likes count
		result, err := r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{"$inc": bson.M{"likes_count": 1}})
		if err != nil {
			return false, fmt.Errorf("increment likes count: %w", err)
		}
		if result.MatchedCount == 0 {
			return false, corecommon.ErrNotFound
		}
		return true, nil
	}
}

// ListLikedMangas returns mangas that a user has reacted to (favorites).
func (r *MongoMangaRepository) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	reactionsColl := r.store.GetCollection("manga_reactions")
	total, err := reactionsColl.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, 0, fmt.Errorf("count user reactions: %w", err)
	}

	findOpts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := reactionsColl.Find(ctx, bson.M{"user_id": userID}, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find user reactions: %w", err)
	}
	defer cursor.Close(ctx)

	var mangaIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var reaction struct {
			MangaID primitive.ObjectID `bson:"manga_id"`
		}
		if err := cursor.Decode(&reaction); err != nil {
			return nil, 0, fmt.Errorf("decode reaction manga id: %w", err)
		}
		mangaIDs = append(mangaIDs, reaction.MangaID)
	}

	if len(mangaIDs) == 0 {
		return []*coremanga.Manga{}, total, nil
	}

	mangasCursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": mangaIDs}})
	if err != nil {
		return nil, 0, fmt.Errorf("find manga records for favorites: %w", err)
	}
	defer mangasCursor.Close(ctx)

	var mangas []*coremanga.Manga
	for mangasCursor.Next(ctx) {
		var m coremanga.Manga
		if err := mangasCursor.Decode(&m); err != nil {
			return nil, 0, fmt.Errorf("decode manga record: %w", err)
		}
		mangas = append(mangas, &m)
	}

	return mangas, total, nil
}

// AddRating creates a new rating or updates the existing user rating for a manga.
func (r *MongoMangaRepository) AddRating(ctx context.Context, rating *coremanga.MangaRating) (newAverage float64, err error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "write")
	defer cancel()

	ratingsColl := r.store.GetCollection("manga_ratings")

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			newAverage, err = r.addRatingInSession(sessCtx, rating, ratingsColl)
			return err
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		newAverage, err = r.addRatingWithoutTransaction(ctx, rating, ratingsColl)
	}

	if err != nil {
		r.store.Log.Error("add rating failed", map[string]interface{}{
			"manga_id":    rating.MangaID.Hex(),
			"user_id":     rating.UserID.Hex(),
			"score":       rating.Score,
			"replica_set": isReplicaSet,
			"error":       err.Error(),
		})
		return 0, err
	}
	return newAverage, nil
}

// addRatingInSession creates or updates a rating within a transaction session.
func (r *MongoMangaRepository) addRatingInSession(sessCtx mongo.SessionContext, rating *coremanga.MangaRating, ratingsColl *mongo.Collection) (float64, error) {
	var existing coremanga.MangaRating
	existingErr := ratingsColl.FindOne(sessCtx, bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID}).Decode(&existing)
	if existingErr != nil && !errors.Is(existingErr, mongo.ErrNoDocuments) {
		return 0, fmt.Errorf("find existing rating: %w", existingErr)
	}

	type MangaStats struct {
		RatingSum   float64 `bson:"rating_sum"`
		RatingCount int64   `bson:"rating_count"`
	}
	var current MangaStats
	var err error
	err = r.coll.FindOne(sessCtx, bson.M{"_id": rating.MangaID}, options.FindOne().SetProjection(bson.M{"rating_sum": 1, "rating_count": 1})).Decode(&current)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, corecommon.ErrNotFound
		}
		return 0, fmt.Errorf("get current stats: %w", err)
	}

	update := bson.M{}
	average := 0.0
	if errors.Is(existingErr, mongo.ErrNoDocuments) {
		rating.CreatedAt = time.Now()
		_, err := ratingsColl.InsertOne(sessCtx, rating)
		if err != nil {
			return 0, fmt.Errorf("insert rating: %w", err)
		}
		average = (current.RatingSum + rating.Score) / float64(current.RatingCount+1)
		update["$inc"] = bson.M{
			"rating_sum":   rating.Score,
			"rating_count": 1,
		}
	} else {
		_, err = ratingsColl.UpdateOne(
			sessCtx,
			bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID},
			bson.M{"$set": bson.M{"score": rating.Score}},
		)
		if err != nil {
			return 0, fmt.Errorf("update rating: %w", err)
		}
		newSum := current.RatingSum - existing.Score + rating.Score
		if current.RatingCount > 0 {
			average = newSum / float64(current.RatingCount)
		}
		update["$inc"] = bson.M{"rating_sum": rating.Score - existing.Score}
	}
	update["$set"] = bson.M{"average_rating": average}

	result, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": rating.MangaID}, update)
	if err != nil {
		return 0, fmt.Errorf("update manga stats: %w", err)
	}
	if result.MatchedCount == 0 {
		return 0, corecommon.ErrNotFound
	}

	return average, nil
}

// addRatingWithoutTransaction creates or updates a rating without transaction.
func (r *MongoMangaRepository) addRatingWithoutTransaction(ctx context.Context, rating *coremanga.MangaRating, ratingsColl *mongo.Collection) (float64, error) {
	type MangaStats struct {
		RatingSum   float64 `bson:"rating_sum"`
		RatingCount int64   `bson:"rating_count"`
	}
	var current MangaStats
	var err error
	err = r.coll.FindOne(ctx, bson.M{"_id": rating.MangaID}, options.FindOne().SetProjection(bson.M{"rating_sum": 1, "rating_count": 1})).Decode(&current)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, corecommon.ErrNotFound
		}
		return 0, fmt.Errorf("get current stats: %w", err)
	}

	var existing coremanga.MangaRating
	existingErr := ratingsColl.FindOne(ctx, bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID}).Decode(&existing)
	if existingErr != nil && !errors.Is(existingErr, mongo.ErrNoDocuments) {
		return 0, fmt.Errorf("find existing rating: %w", existingErr)
	}

	update := bson.M{
		"$set": bson.M{},
	}
	if errors.Is(existingErr, mongo.ErrNoDocuments) {
		rating.CreatedAt = time.Now()
		_, err = ratingsColl.InsertOne(ctx, rating)
		if err != nil {
			return 0, fmt.Errorf("insert rating: %w", err)
		}
		average := (current.RatingSum + rating.Score) / float64(current.RatingCount+1)
		update["$inc"] = bson.M{
			"rating_sum":   rating.Score,
			"rating_count": 1,
		}
		update["$set"] = bson.M{"average_rating": average}
		result, err := r.coll.UpdateOne(ctx, bson.M{"_id": rating.MangaID}, update)
		if err != nil {
			_, _ = ratingsColl.DeleteOne(ctx, bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID})
			return 0, fmt.Errorf("update manga stats: %w", err)
		}
		if result.MatchedCount == 0 {
			_, _ = ratingsColl.DeleteOne(ctx, bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID})
			return 0, corecommon.ErrNotFound
		}
		return average, nil
	}

	_, err = ratingsColl.UpdateOne(
		ctx,
		bson.M{"manga_id": rating.MangaID, "user_id": rating.UserID},
		bson.M{"$set": bson.M{"score": rating.Score}},
	)
	if err != nil {
		return 0, fmt.Errorf("update rating: %w", err)
	}

	newSum := current.RatingSum - existing.Score + rating.Score
	average := 0.0
	if current.RatingCount > 0 {
		average = newSum / float64(current.RatingCount)
	}
	update["$inc"] = bson.M{"rating_sum": rating.Score - existing.Score}
	update["$set"] = bson.M{"average_rating": average}

	result, err := r.coll.UpdateOne(ctx, bson.M{"_id": rating.MangaID}, update)
	if err != nil {
		return 0, fmt.Errorf("update manga stats: %w", err)
	}
	if result.MatchedCount == 0 {
		return 0, corecommon.ErrNotFound
	}

	return average, nil
}

// hasUserRatedInSession checks if user rated within a session
func (r *MongoMangaRepository) hasUserRatedInSession(sessCtx mongo.SessionContext, mangaID, userID primitive.ObjectID, ratingsColl *mongo.Collection) (bool, error) {
	count, err := ratingsColl.CountDocuments(sessCtx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("count ratings: %w", err)
	}
	return count > 0, nil
}

// HasUserRated checks if a user has already rated a manga.
func (r *MongoMangaRepository) HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga", "read")
	defer cancel()

	ratingsColl := r.store.GetCollection("manga_ratings")

	count, err := ratingsColl.CountDocuments(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("count ratings: %w", err)
	}
	return count > 0, nil
}

// setReactionInSession sets or updates a reaction within a transaction session
func (r *MongoMangaRepository) setReactionInSession(sessCtx mongo.SessionContext, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType, reactionsColl *mongo.Collection) (string, error) {
	// First check if manga exists
	mangaCount, err := r.coll.CountDocuments(sessCtx, bson.M{"_id": mangaID})
	if err != nil {
		return "", fmt.Errorf("check manga exists: %w", err)
	}
	if mangaCount == 0 {
		return "", corecommon.ErrNotFound
	}

	// Check if reaction exists
	var existingReaction coremanga.MangaReaction
	err = reactionsColl.FindOne(sessCtx, bson.M{"manga_id": mangaID, "user_id": userID}).Decode(&existingReaction)

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("find existing reaction: %w", err)
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		// No existing reaction, create new one
		reaction := coremanga.MangaReaction{
			MangaID:   mangaID,
			UserID:    userID,
			Type:      reactionType,
			CreatedAt: time.Now(),
		}
		_, err = reactionsColl.InsertOne(sessCtx, reaction)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Race condition, just return the reaction type
				return string(reactionType), nil
			}
			return "", fmt.Errorf("insert reaction: %w", err)
		}
		// Increment reactions_count for this reaction type
		result, err := r.coll.UpdateOne(
			sessCtx,
			bson.M{"_id": mangaID},
			bson.M{
				"$inc": bson.M{fmt.Sprintf("reactions_count.%s", string(reactionType)): 1},
			},
		)
		if err != nil {
			return "", fmt.Errorf("increment reaction count: %w", err)
		}
		if result.MatchedCount == 0 {
			return "", corecommon.ErrNotFound
		}
		return string(reactionType), nil
	}

	// Existing reaction - update it
	if existingReaction.Type == reactionType {
		// Same reaction, just return it
		return string(reactionType), nil
	}

	// Different reaction, update it
	_, err = reactionsColl.UpdateOne(
		sessCtx,
		bson.M{"manga_id": mangaID, "user_id": userID},
		bson.M{"$set": bson.M{"type": reactionType}},
	)
	if err != nil {
		return "", fmt.Errorf("update reaction: %w", err)
	}

	// Update reactions_count: decrement old, increment new
	result, err := r.coll.UpdateOne(
		sessCtx,
		bson.M{"_id": mangaID},
		bson.M{
			"$inc": bson.M{
				fmt.Sprintf("reactions_count.%s", string(existingReaction.Type)): -1,
				fmt.Sprintf("reactions_count.%s", string(reactionType)):          1,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("update reaction counts: %w", err)
	}
	if result.MatchedCount == 0 {
		return "", corecommon.ErrNotFound
	}

	return string(reactionType), nil
}

// setReactionWithoutTransaction sets or updates a reaction without transaction (for standalone MongoDB)
func (r *MongoMangaRepository) setReactionWithoutTransaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType, reactionsColl *mongo.Collection) (string, error) {
	// First check if manga exists
	mangaCount, err := r.coll.CountDocuments(ctx, bson.M{"_id": mangaID})
	if err != nil {
		return "", fmt.Errorf("check manga exists: %w", err)
	}
	if mangaCount == 0 {
		return "", corecommon.ErrNotFound
	}

	// Check if reaction exists
	var existingReaction coremanga.MangaReaction
	err = reactionsColl.FindOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID}).Decode(&existingReaction)

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("find existing reaction: %w", err)
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		// No existing reaction, create new one
		reaction := coremanga.MangaReaction{
			MangaID:   mangaID,
			UserID:    userID,
			Type:      reactionType,
			CreatedAt: time.Now(),
		}
		_, err = reactionsColl.InsertOne(ctx, reaction)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Race condition, just return the reaction type
				return string(reactionType), nil
			}
			return "", fmt.Errorf("insert reaction: %w", err)
		}
		// Increment reactions_count for this reaction type
		result, err := r.coll.UpdateOne(
			ctx,
			bson.M{"_id": mangaID},
			bson.M{
				"$inc": bson.M{fmt.Sprintf("reactions_count.%s", string(reactionType)): 1},
			},
		)
		if err != nil {
			return "", fmt.Errorf("increment reaction count: %w", err)
		}
		if result.MatchedCount == 0 {
			return "", corecommon.ErrNotFound
		}
		return string(reactionType), nil
	}

	// Existing reaction - update it
	if existingReaction.Type == reactionType {
		// Same reaction, just return it
		return string(reactionType), nil
	}

	// Different reaction, update it
	_, err = reactionsColl.UpdateOne(
		ctx,
		bson.M{"manga_id": mangaID, "user_id": userID},
		bson.M{"$set": bson.M{"type": reactionType}},
	)
	if err != nil {
		return "", fmt.Errorf("update reaction: %w", err)
	}

	// Update reactions_count: decrement old, increment new
	result, err := r.coll.UpdateOne(
		ctx,
		bson.M{"_id": mangaID},
		bson.M{
			"$inc": bson.M{
				fmt.Sprintf("reactions_count.%s", string(existingReaction.Type)): -1,
				fmt.Sprintf("reactions_count.%s", string(reactionType)):          1,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("update reaction counts: %w", err)
	}
	if result.MatchedCount == 0 {
		return "", corecommon.ErrNotFound
	}

	return string(reactionType), nil
}

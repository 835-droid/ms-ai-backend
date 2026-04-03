// internal/data/content/manga/manga_rating_repository.go
package manga

import (
	"context"
	"fmt"
	"time"

	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// addRatingInSession adds rating within a transaction session and returns the new average
func (r *MongoMangaRepository) addRatingInSession(sessCtx mongo.SessionContext, rating *coremanga.MangaRating, ratingsColl *mongo.Collection) (float64, error) {
	// Upsert the user rating
	_, err := ratingsColl.UpdateOne(sessCtx, bson.M{
		"manga_id": rating.MangaID,
		"user_id":  rating.UserID,
	}, bson.M{
		"$set": bson.M{
			"rating":     rating.Score,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return 0, fmt.Errorf("upsert rating: %w", err)
	}

	// Calculate new average rating
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"manga_id": rating.MangaID}}},
		{{"$group", bson.M{
			"_id":   nil,
			"avg":   bson.M{"$avg": "$rating"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := ratingsColl.Aggregate(sessCtx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("aggregate ratings: %w", err)
	}
	defer cursor.Close(sessCtx)

	var result struct {
		Avg   float64 `bson:"avg"`
		Count int     `bson:"count"`
	}
	if cursor.Next(sessCtx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, fmt.Errorf("decode aggregate result: %w", err)
		}
	}

	return result.Avg, nil
}

// addRatingWithoutTransaction adds rating without transaction (for standalone MongoDB) and returns the new average
func (r *MongoMangaRepository) addRatingWithoutTransaction(ctx context.Context, rating *coremanga.MangaRating, ratingsColl *mongo.Collection) (float64, error) {
	// Upsert the user rating
	_, err := ratingsColl.UpdateOne(ctx, bson.M{
		"manga_id": rating.MangaID,
		"user_id":  rating.UserID,
	}, bson.M{
		"$set": bson.M{
			"rating":     rating.Score,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return 0, fmt.Errorf("upsert rating: %w", err)
	}

	// Calculate new average rating
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"manga_id": rating.MangaID}}},
		{{"$group", bson.M{
			"_id":   nil,
			"avg":   bson.M{"$avg": "$rating"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := ratingsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("aggregate ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Avg   float64 `bson:"avg"`
		Count int     `bson:"count"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, fmt.Errorf("decode aggregate result: %w", err)
		}
	}

	return result.Avg, nil
}

// HasUserRated checks if a user has already rated a manga.
func (r *MongoMangaRepository) HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	ratingsColl := r.store.GetCollection("manga_ratings")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_ratings", "read")
	defer cancel()

	count, err := ratingsColl.CountDocuments(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("count ratings: %w", err)
	}
	return count > 0, nil
}

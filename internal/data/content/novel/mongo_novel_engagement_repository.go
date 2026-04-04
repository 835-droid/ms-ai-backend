// Package novel implements novel engagement data access operations backed by MongoDB.
package novel

import (
	"context"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	corenovel "github.com/835-droid/ms-ai-backend/internal/core/content/novel"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetReaction sets or toggles a reaction for a novel.
func (r *MongoNovelRepository) SetReaction(ctx context.Context, novelID, userID primitive.ObjectID, reactionType corenovel.ReactionType) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_reactions", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_reactions")

	// Check if user already has a reaction
	existing := bson.M{}
	err := coll.FindOne(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	}).Decode(&existing)

	if err == nil {
		// User has existing reaction - check if same type (toggle off)
		if existing["type"].(string) == string(reactionType) {
			// Remove reaction
			_, err := coll.DeleteOne(ctx, bson.M{
				"novel_id": novelID,
				"user_id":  userID,
			})
			if err != nil {
				return "", fmt.Errorf("remove reaction: %w", err)
			}

			// Decrement reaction count
			r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{
				"$inc": bson.M{
					fmt.Sprintf("reactions_count.%s", reactionType): -1,
				},
			})
			return "", nil
		}

		// Change reaction type
		oldType := existing["type"].(string)
		_, err := coll.UpdateOne(ctx, bson.M{
			"novel_id": novelID,
			"user_id":  userID,
		}, bson.M{
			"$set": bson.M{"type": string(reactionType)},
		})
		if err != nil {
			return "", fmt.Errorf("update reaction: %w", err)
		}

		// Update reaction counts
		r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{
			"$inc": bson.M{
				fmt.Sprintf("reactions_count.%s", oldType):      -1,
				fmt.Sprintf("reactions_count.%s", reactionType): 1,
			},
		})
		return string(reactionType), nil
	}

	// Add new reaction
	reaction := bson.M{
		"novel_id":   novelID,
		"user_id":    userID,
		"type":       string(reactionType),
		"created_at": time.Now(),
	}
	_, err = coll.InsertOne(ctx, reaction)
	if err != nil {
		return "", fmt.Errorf("insert reaction: %w", err)
	}

	// Increment reaction count
	r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{
		"$inc": bson.M{
			fmt.Sprintf("reactions_count.%s", reactionType): 1,
		},
	})

	return string(reactionType), nil
}

// GetUserReaction gets the current reaction type for a user on a novel.
func (r *MongoNovelRepository) GetUserReaction(ctx context.Context, novelID, userID primitive.ObjectID) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_reactions", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_reactions")

	var result bson.M
	err := coll.FindOne(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", fmt.Errorf("find reaction: %w", err)
	}

	return result["type"].(string), nil
}

// ListLikedNovels returns novels liked by a user.
func (r *MongoNovelRepository) ListLikedNovels(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_reactions", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_reactions")

	// Find all reactions by user
	cursor, err := coll.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSkip(skip).SetLimit(limit))
	if err != nil {
		return nil, 0, fmt.Errorf("find reactions: %w", err)
	}
	defer cursor.Close(ctx)

	var novelIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, fmt.Errorf("decode reaction: %w", err)
		}
		if id, ok := result["novel_id"].(primitive.ObjectID); ok {
			novelIDs = append(novelIDs, id)
		}
	}

	if len(novelIDs) == 0 {
		return []*corenovel.Novel{}, 0, nil
	}

	// Get novels
	novels, total, err := r.ListNovels(ctx, 0, int64(len(novelIDs)))
	if err != nil {
		return nil, 0, fmt.Errorf("find novels: %w", err)
	}

	// Filter to only include novels in our list
	novelMap := make(map[primitive.ObjectID]*corenovel.Novel)
	for _, n := range novels {
		novelMap[n.ID] = n
	}

	var result []*corenovel.Novel
	for _, id := range novelIDs {
		if n, ok := novelMap[id]; ok {
			result = append(result, n)
		}
	}

	return result, total, nil
}

// AddRating adds a rating for a novel.
func (r *MongoNovelRepository) AddRating(ctx context.Context, rating *corenovel.NovelRating) (float64, error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_ratings", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_ratings")

	// Check if user already rated
	var existing bson.M
	err := coll.FindOne(ctx, bson.M{
		"novel_id": rating.NovelID,
		"user_id":  rating.UserID,
	}).Decode(&existing)

	if err == nil {
		// Update existing rating
		_, err := coll.UpdateOne(ctx, bson.M{
			"novel_id": rating.NovelID,
			"user_id":  rating.UserID,
		}, bson.M{
			"$set": bson.M{
				"score":      rating.Score,
				"created_at": rating.CreatedAt,
			},
		})
		if err != nil {
			return 0, fmt.Errorf("update rating: %w", err)
		}
	} else {
		// Insert new rating
		_, err := coll.InsertOne(ctx, rating)
		if err != nil {
			return 0, fmt.Errorf("insert rating: %w", err)
		}
	}

	// Calculate new average
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"novel_id": rating.NovelID}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$novel_id",
			"sum":   bson.M{"$sum": "$score"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("calculate average: %w", err)
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, fmt.Errorf("decode average: %w", err)
	}

	var newAverage float64
	if len(result) > 0 {
		sum := result[0]["sum"].(int32)
		count := result[0]["count"].(int32)
		newAverage = float64(sum) / float64(count)

		// Update novel rating stats
		r.coll.UpdateOne(ctx, bson.M{"_id": rating.NovelID}, bson.M{
			"$set": bson.M{
				"average_rating": newAverage,
				"rating_count":   int64(count),
				"rating_sum":     float64(sum),
			},
		})
	}

	return newAverage, nil
}

// HasUserRated checks if a user has rated a novel.
func (r *MongoNovelRepository) HasUserRated(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_ratings", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_ratings")

	count, err := coll.CountDocuments(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	})
	if err != nil {
		return false, fmt.Errorf("count ratings: %w", err)
	}

	return count > 0, nil
}

// AddFavorite adds a novel to user's favorites.
func (r *MongoNovelRepository) AddFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_favorites", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_favorites")

	// Check if already favorited
	count, err := coll.CountDocuments(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	})
	if err != nil {
		return fmt.Errorf("check favorite: %w", err)
	}
	if count > 0 {
		return nil // Already favorited
	}

	// Add favorite
	favorite := bson.M{
		"novel_id":   novelID,
		"user_id":    userID,
		"created_at": time.Now(),
	}
	_, err = coll.InsertOne(ctx, favorite)
	if err != nil {
		return fmt.Errorf("insert favorite: %w", err)
	}

	// Increment favorites count
	r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{
		"$inc": bson.M{"favorites_count": 1},
	})

	return nil
}

// RemoveFavorite removes a novel from user's favorites.
func (r *MongoNovelRepository) RemoveFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_favorites", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_favorites")

	_, err := coll.DeleteOne(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	})
	if err != nil {
		return fmt.Errorf("delete favorite: %w", err)
	}

	// Decrement favorites count
	r.coll.UpdateOne(ctx, bson.M{"_id": novelID}, bson.M{
		"$inc": bson.M{"favorites_count": -1},
	})

	return nil
}

// IsFavorite checks if a novel is in user's favorites.
func (r *MongoNovelRepository) IsFavorite(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_favorites", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_favorites")

	count, err := coll.CountDocuments(ctx, bson.M{
		"novel_id": novelID,
		"user_id":  userID,
	})
	if err != nil {
		return false, fmt.Errorf("count favorites: %w", err)
	}

	return count > 0, nil
}

// ListFavorites retrieves a user's favorite novels.
func (r *MongoNovelRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*corenovel.Novel, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_favorites", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_favorites")

	// Find all favorites by user
	cursor, err := coll.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, 0, fmt.Errorf("find favorites: %w", err)
	}
	defer cursor.Close(ctx)

	var novelIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, fmt.Errorf("decode favorite: %w", err)
		}
		if id, ok := result["novel_id"].(primitive.ObjectID); ok {
			novelIDs = append(novelIDs, id)
		}
	}

	if len(novelIDs) == 0 {
		return []*corenovel.Novel{}, 0, nil
	}

	// Get novels
	novels, total, err := r.ListNovels(ctx, 0, int64(len(novelIDs)))
	if err != nil {
		return nil, 0, fmt.Errorf("find novels: %w", err)
	}

	// Filter to only include novels in our list
	novelMap := make(map[primitive.ObjectID]*corenovel.Novel)
	for _, n := range novels {
		novelMap[n.ID] = n
	}

	var result []*corenovel.Novel
	for _, id := range novelIDs {
		if n, ok := novelMap[id]; ok {
			result = append(result, n)
		}
	}

	return result, total, nil
}

// AddNovelComment adds a comment to a novel.
func (r *MongoNovelRepository) AddNovelComment(ctx context.Context, comment *corenovel.NovelComment) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_comments", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_comments")

	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	res, err := coll.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}
	comment.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// ListNovelComments retrieves comments for a novel.
func (r *MongoNovelRepository) ListNovelComments(ctx context.Context, novelID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*corenovel.NovelComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_comments", "read")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_comments")

	filter := bson.M{"novel_id": novelID}

	sortDir := -1
	if sortOrder == "oldest" {
		sortDir = 1
	}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": sortDir})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []*corenovel.NovelComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("decode comments: %w", err)
	}

	return comments, total, nil
}

// DeleteNovelComment deletes a novel comment.
func (r *MongoNovelRepository) DeleteNovelComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "novel_comments", "write")
	defer cancel()

	coll := r.store.Client.Database(r.store.DBName).Collection("novel_comments")

	result, err := coll.DeleteOne(ctx, bson.M{
		"_id":     commentID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if result.DeletedCount == 0 {
		return corecommon.ErrNotFound
	}
	return nil
}

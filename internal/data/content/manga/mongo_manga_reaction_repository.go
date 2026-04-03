// internal/data/content/manga/manga_reaction_repository.go
package manga

import (
	"context"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetReaction sets a reaction for a manga by a user.
func (r *MongoMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (reaction string, err error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}

	reactionsColl := r.store.GetCollection("manga_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_reactions", "write")
	defer cancel()

	// Check throttling
	key := fmt.Sprintf("%s_%s", mangaID.Hex(), userID.Hex())
	if _, exists := r.reactionLock.LoadOrStore(key, true); exists {
		return "", corecommon.ErrTooManyRequests
	}
	defer r.reactionLock.Delete(key)

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			var e error
			reaction, e = r.setReactionInSession(sessCtx, mangaID, userID, reactionType, reactionsColl)
			return e
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		reaction, err = r.setReactionWithoutTransaction(ctx, mangaID, userID, reactionType, reactionsColl)
	}

	if err != nil {
		r.store.Log.Error("set reaction failed", map[string]interface{}{
			"manga_id": mangaID.Hex(),
			"user_id":  userID.Hex(),
			"reaction": reactionType,
			"error":    err.Error(),
		})
		return "", err
	}

	r.store.Log.Info("reaction set", map[string]interface{}{
		"manga_id":    mangaID.Hex(),
		"user_id":     userID.Hex(),
		"reaction":    reaction,
		"replica_set": isReplicaSet,
	})
	return reaction, nil
}

// GetUserReaction gets the user's reaction for a manga.
func (r *MongoMangaRepository) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}

	reactionsColl := r.store.GetCollection("manga_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_reactions", "read")
	defer cancel()

	var reactionDoc struct {
		Reaction string `bson:"reaction"`
	}
	err := reactionsColl.FindOne(ctx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}).Decode(&reactionDoc)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // No reaction
		}
		return "", fmt.Errorf("get user reaction: %w", err)
	}

	return reactionDoc.Reaction, nil
}

// ListLikedMangas retrieves a paginated list of mangas liked by a user.
func (r *MongoMangaRepository) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	reactionsColl := r.store.GetCollection("manga_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_reactions", "read")
	defer cancel()

	if limit == 0 {
		limit = mongoinfra.DefaultLimit
	}
	if limit > 100 {
		limit = 100
	}

	// Count total liked mangas
	total, err := reactionsColl.CountDocuments(ctx, bson.M{
		"user_id":  userID,
		"reaction": "like",
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count liked mangas: %w", err)
	}

	// Get manga IDs
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1}).
		SetProjection(bson.M{"manga_id": 1})

	cursor, err := reactionsColl.Find(ctx, bson.M{
		"user_id":  userID,
		"reaction": "like",
	}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find liked manga ids: %w", err)
	}
	defer cursor.Close(ctx)

	var mangaIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var doc struct {
			MangaID primitive.ObjectID `bson:"manga_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, fmt.Errorf("decode manga id: %w", err)
		}
		mangaIDs = append(mangaIDs, doc.MangaID)
	}

	if len(mangaIDs) == 0 {
		return []*coremanga.Manga{}, total, nil
	}

	// Fetch manga details
	mangaOpts := options.Find().
		SetSort(bson.M{"created_at": -1})

	mangaCursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": mangaIDs}}, mangaOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find mangas: %w", err)
	}
	defer mangaCursor.Close(ctx)

	var mangas []*coremanga.Manga
	if err := mangaCursor.All(ctx, &mangas); err != nil {
		return nil, 0, fmt.Errorf("decode mangas: %w", err)
	}

	return mangas, total, nil
}

// setReactionInSession sets reaction within a transaction session
func (r *MongoMangaRepository) setReactionInSession(sessCtx mongo.SessionContext, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType, reactionsColl *mongo.Collection) (string, error) {
	reaction := string(reactionType)

	// Check if reaction already exists
	var existing struct {
		Reaction string `bson:"reaction"`
	}
	err := reactionsColl.FindOne(sessCtx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}).Decode(&existing)

	if err != nil && err != mongo.ErrNoDocuments {
		return "", fmt.Errorf("check existing reaction: %w", err)
	}

	// If same reaction, remove it (toggle off)
	if err == nil && existing.Reaction == reaction {
		_, err := reactionsColl.DeleteOne(sessCtx, bson.M{
			"manga_id": mangaID,
			"user_id":  userID,
		})
		if err != nil {
			return "", fmt.Errorf("remove reaction: %w", err)
		}
		// Decrement reactions_count
		_, err = r.coll.UpdateOne(sessCtx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{fmt.Sprintf("reactions_count.%s", reaction): -1},
		})
		if err != nil {
			return "", fmt.Errorf("decrement reactions count: %w", err)
		}
		// Ensure non-negative count
		_, err = r.coll.UpdateOne(sessCtx, bson.M{
			"_id": mangaID,
			fmt.Sprintf("reactions_count.%s", reaction): bson.M{"$lt": 0},
		}, bson.M{
			"$set": bson.M{fmt.Sprintf("reactions_count.%s", reaction): 0},
		})
		if err != nil {
			return "", fmt.Errorf("fix negative reactions count: %w", err)
		}
		return "", nil // Removed
	}

	// Handle changing reaction type or adding new reaction
	var oldReaction string
	if err == nil {
		oldReaction = existing.Reaction
	}

	// Upsert the reaction
	_, err = reactionsColl.UpdateOne(sessCtx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}, bson.M{
		"$set": bson.M{
			"reaction":   reaction,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return "", fmt.Errorf("upsert reaction: %w", err)
	}

	// Update reactions_count
	if oldReaction != "" && oldReaction != reaction {
		// Changing from one type to another - first increment/decrement
		_, err = r.coll.UpdateOne(sessCtx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{
				fmt.Sprintf("reactions_count.%s", oldReaction): -1,
				fmt.Sprintf("reactions_count.%s", reaction):    1,
			},
		})
		if err != nil {
			return "", fmt.Errorf("update reactions count: %w", err)
		}
		// Ensure old reaction count is non-negative
		_, err = r.coll.UpdateOne(sessCtx, bson.M{
			"_id": mangaID,
			fmt.Sprintf("reactions_count.%s", oldReaction): bson.M{"$lt": 0},
		}, bson.M{
			"$set": bson.M{fmt.Sprintf("reactions_count.%s", oldReaction): 0},
		})
		if err != nil {
			return "", fmt.Errorf("fix negative old reaction count: %w", err)
		}
	} else {
		// Adding new reaction
		_, err = r.coll.UpdateOne(sessCtx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{fmt.Sprintf("reactions_count.%s", reaction): 1},
		})
	}
	if err != nil {
		return "", fmt.Errorf("update reactions count: %w", err)
	}

	return reaction, nil
}

// setReactionWithoutTransaction sets reaction without transaction (for standalone MongoDB)
func (r *MongoMangaRepository) setReactionWithoutTransaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType, reactionsColl *mongo.Collection) (string, error) {
	reaction := string(reactionType)

	// Check if reaction already exists
	var existing struct {
		Reaction string `bson:"reaction"`
	}
	err := reactionsColl.FindOne(ctx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}).Decode(&existing)

	if err != nil && err != mongo.ErrNoDocuments {
		return "", fmt.Errorf("check existing reaction: %w", err)
	}

	// If same reaction, remove it (toggle off)
	if err == nil && existing.Reaction == reaction {
		_, err := reactionsColl.DeleteOne(ctx, bson.M{
			"manga_id": mangaID,
			"user_id":  userID,
		})
		if err != nil {
			return "", fmt.Errorf("remove reaction: %w", err)
		}
		// Decrement reactions_count
		_, err = r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{fmt.Sprintf("reactions_count.%s", reaction): -1},
		})
		if err != nil {
			return "", fmt.Errorf("decrement reactions count: %w", err)
		}
		// Ensure non-negative count
		_, err = r.coll.UpdateOne(ctx, bson.M{
			"_id": mangaID,
			fmt.Sprintf("reactions_count.%s", reaction): bson.M{"$lt": 0},
		}, bson.M{
			"$set": bson.M{fmt.Sprintf("reactions_count.%s", reaction): 0},
		})
		if err != nil {
			return "", fmt.Errorf("fix negative reactions count: %w", err)
		}
		return "", nil // Removed
	}

	// Handle changing reaction type or adding new reaction
	var oldReaction string
	if err == nil {
		oldReaction = existing.Reaction
	}

	// Upsert the reaction
	_, err = reactionsColl.UpdateOne(ctx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}, bson.M{
		"$set": bson.M{
			"reaction":   reaction,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return "", fmt.Errorf("upsert reaction: %w", err)
	}

	// Update reactions_count
	if oldReaction != "" && oldReaction != reaction {
		// Changing from one type to another - first increment/decrement
		_, err = r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{
				fmt.Sprintf("reactions_count.%s", oldReaction): -1,
				fmt.Sprintf("reactions_count.%s", reaction):    1,
			},
		})
		if err != nil {
			return "", fmt.Errorf("update reactions count: %w", err)
		}
		// Ensure old reaction count is non-negative
		_, err = r.coll.UpdateOne(ctx, bson.M{
			"_id": mangaID,
			fmt.Sprintf("reactions_count.%s", oldReaction): bson.M{"$lt": 0},
		}, bson.M{
			"$set": bson.M{fmt.Sprintf("reactions_count.%s", oldReaction): 0},
		})
		if err != nil {
			return "", fmt.Errorf("fix negative old reaction count: %w", err)
		}
	} else {
		// Adding new reaction
		_, err = r.coll.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$inc": bson.M{fmt.Sprintf("reactions_count.%s", reaction): 1},
		})
	}
	if err != nil {
		return "", fmt.Errorf("update reactions count: %w", err)
	}

	return reaction, nil
}

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

// ========== FAVORITES ==========

// AddFavorite adds a manga to user's favorites
func (r *MongoMangaRepository) AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	favColl := r.store.GetCollection("user_favorites")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "user_favorites", "write")
	defer cancel()

	opts := options.Update().SetUpsert(true)
	result, err := favColl.UpdateOne(ctx,
		bson.M{"manga_id": mangaID, "user_id": userID},
		bson.M{
			"$setOnInsert": bson.M{"manga_id": mangaID, "user_id": userID},
			"$set":         bson.M{"created_at": time.Now()},
		},
		opts,
	)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}

	// Update favorites count if a new favorite was added
	if result.UpsertedCount > 0 {
		mangaColl := r.store.GetCollection("manga")
		_, err = mangaColl.UpdateOne(ctx,
			bson.M{"_id": mangaID},
			bson.M{"$inc": bson.M{"favorites_count": 1}},
		)
		if err != nil {
			return fmt.Errorf("update favorites count: %w", err)
		}
	}

	return nil
}

// RemoveFavorite removes a manga from user's favorites
func (r *MongoMangaRepository) RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	favColl := r.store.GetCollection("user_favorites")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "user_favorites", "write")
	defer cancel()

	result, err := favColl.DeleteOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}

	// Update favorites count if a favorite was removed
	// Use aggregation pipeline to safely decrement and clamp to 0
	// Note: $max is not a valid update operator in a standard update document,
	// so we use an aggregation pipeline with $max to ensure the count doesn't go below 0
	if result.DeletedCount > 0 {
		mangaColl := r.store.GetCollection("manga")
		// Use aggregation pipeline update (MongoDB 4.2+) to clamp the value
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$set", Value: bson.M{
				"favorites_count": bson.M{
					"$max": bson.A{
						bson.M{"$subtract": bson.A{"$favorites_count", 1}},
						0,
					},
				},
			}}},
		}
		_, err = mangaColl.UpdateOne(ctx, bson.M{"_id": mangaID}, pipeline)
		if err != nil {
			return fmt.Errorf("update favorites count: %w", err)
		}
	}

	return nil
}

// IsFavorite checks if a manga is in user's favorites
func (r *MongoMangaRepository) IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	favColl := r.store.GetCollection("user_favorites")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "user_favorites", "read")
	defer cancel()

	count, err := favColl.CountDocuments(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("is favorite check: %w", err)
	}
	return count > 0, nil
}

// ListFavorites lists all favorites for a user
func (r *MongoMangaRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	favColl := r.store.GetCollection("user_favorites")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "user_favorites", "read")
	defer cancel()

	// Get user's favorite manga IDs with pagination
	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := favColl.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	defer cursor.Close(ctx)

	var favorites []coremanga.UserFavorite
	if err := cursor.All(ctx, &favorites); err != nil {
		return nil, 0, fmt.Errorf("decode favorites: %w", err)
	}

	// Get each manga
	mangas := make([]*coremanga.Manga, 0)
	for _, fav := range favorites {
		manga, err := r.GetMangaByID(ctx, fav.MangaID)
		if err == nil && manga != nil {
			mangas = append(mangas, manga)
		}
	}

	// Get total count
	total, err := favColl.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	return mangas, total, nil
}

// ========== MANGA COMMENTS ==========

// AddMangaComment adds a comment to a manga
func (r *MongoMangaRepository) AddMangaComment(ctx context.Context, comment *coremanga.MangaComment) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	commentColl := r.store.GetCollection("manga_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_comments", "write")
	defer cancel()

	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}
	comment.UpdatedAt = time.Now()

	_, err := commentColl.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("add manga comment: %w", err)
	}
	return nil
}

// ListMangaComments lists comments for a manga
func (r *MongoMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.MangaComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	commentColl := r.store.GetCollection("manga_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_comments", "read")
	defer cancel()

	sort := bson.M{"created_at": -1} // default: newest first
	if sortOrder == "oldest" {
		sort = bson.M{"created_at": 1}
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
	cursor, err := commentColl.Find(ctx, bson.M{"manga_id": mangaID}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list manga comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []*coremanga.MangaComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("decode manga comments: %w", err)
	}

	total, err := commentColl.CountDocuments(ctx, bson.M{"manga_id": mangaID})
	if err != nil {
		return nil, 0, fmt.Errorf("count manga comments: %w", err)
	}

	return comments, total, nil
}

// DeleteMangaComment deletes a comment (user or admin only)
func (r *MongoMangaRepository) DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	commentColl := r.store.GetCollection("manga_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_comments", "write")
	defer cancel()

	// Only allow deletion by comment creator
	result, err := commentColl.DeleteOne(ctx, bson.M{"_id": commentID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("delete manga comment: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("comment not found or unauthorized")
	}
	return nil
}

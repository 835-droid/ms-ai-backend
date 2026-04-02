package manga

import (
	"context"
	"fmt"
	"time"

	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	_, err := favColl.UpdateOne(ctx,
		bson.M{"manga_id": mangaID, "user_id": userID},
		bson.M{"$set": bson.M{"created_at": time.Now()}},
		opts,
	)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
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

	_, err := favColl.DeleteOne(ctx, bson.M{"manga_id": mangaID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
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

	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := commentColl.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("add manga comment: %w", err)
	}
	return nil
}

// ListMangaComments lists comments for a manga
func (r *MongoMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	commentColl := r.store.GetCollection("manga_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_comments", "read")
	defer cancel()

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
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

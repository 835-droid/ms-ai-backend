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

// ========== CHAPTER VIEWS ==========

// IncrementChapterViews increments view count for a chapter and the manga total
func (r *MongoMangaChapterRepository) IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "manga_chapters", "write")
	defer cancel()

	// Increment chapter views
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": chapterID},
		bson.M{"$inc": bson.M{"views_count": int64(1)}},
	)
	if err != nil {
		return fmt.Errorf("increment chapter views: %w", err)
	}

	// Increment manga total views
	mangaColl := r.store.GetCollection("manga")
	_, err = mangaColl.UpdateOne(ctx,
		bson.M{"_id": mangaID},
		bson.M{"$inc": bson.M{"views_count": int64(1)}},
	)
	if err != nil {
		return fmt.Errorf("increment manga views: %w", err)
	}

	return nil
}

// ========== CHAPTER RATINGS ==========

// AddChapterRating adds a rating to a chapter and recalculates averages
func (r *MongoMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (newAverage float64, err error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, err
	}

	ratingColl := r.store.GetCollection("chapter_ratings")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_ratings", "write")
	defer cancel()

	rating.CreatedAt = time.Now()

	// Upsert the rating
	opts := options.Update().SetUpsert(true)
	result, err := ratingColl.UpdateOne(ctx,
		bson.M{"chapter_id": rating.ChapterID, "user_id": rating.UserID},
		bson.M{"$set": rating},
		opts,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert chapter rating: %w", err)
	}

	// If it's a new rating (not an update), increment rating_count
	if result.UpsertedID != nil {
		// New rating
		_, err = r.coll.UpdateOne(ctx,
			bson.M{"_id": rating.ChapterID},
			bson.M{"$inc": bson.M{"rating_count": 1}},
		)
		if err != nil {
			return 0, fmt.Errorf("increment rating_count: %w", err)
		}
	}

	// Fetch all ratings for this chapter
	cursor, err := ratingColl.Find(ctx, bson.M{"chapter_id": rating.ChapterID})
	if err != nil {
		return 0, fmt.Errorf("fetch chapter ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*coremanga.ChapterRating
	if err := cursor.All(ctx, &ratings); err != nil {
		return 0, fmt.Errorf("decode ratings: %w", err)
	}

	// Calculate average
	var sum float64
	count := int64(len(ratings))
	for _, r := range ratings {
		sum += r.Score
	}

	avgRating := 0.0
	if count > 0 {
		avgRating = sum / float64(count)

		// Update chapter with new average
		_, err = r.coll.UpdateOne(ctx,
			bson.M{"_id": rating.ChapterID},
			bson.M{"$set": bson.M{
				"average_rating": avgRating,
				"rating_sum":     sum,
				"rating_count":   count,
			}},
		)
		if err != nil {
			return 0, fmt.Errorf("update chapter averages: %w", err)
		}
	}

	// Recalculate manga average from all its chapters
	chapColl := r.store.GetCollection("manga_chapters")
	cursor, err = chapColl.Find(ctx, bson.M{"manga_id": rating.MangaID})
	if err != nil {
		return avgRating, fmt.Errorf("fetch manga chapters: %w", err)
	}
	defer cursor.Close(ctx)

	var chapters []*coremanga.MangaChapter
	if err := cursor.All(ctx, &chapters); err != nil {
		return avgRating, fmt.Errorf("decode chapters: %w", err)
	}

	mangaAvgSum := 0.0
	validChapters := 0
	for _, ch := range chapters {
		if ch.AverageRating > 0 {
			mangaAvgSum += ch.AverageRating
			validChapters++
		}
	}

	if validChapters > 0 {
		mangaAvg := mangaAvgSum / float64(validChapters)
		mangaColl := r.store.GetCollection("manga")
		_, err = mangaColl.UpdateOne(ctx,
			bson.M{"_id": rating.MangaID},
			bson.M{"$set": bson.M{"average_rating": mangaAvg}},
		)
		if err != nil {
			return avgRating, fmt.Errorf("update manga average: %w", err)
		}
	}

	return avgRating, nil
}

// HasUserRatedChapter checks if user has already rated a chapter
func (r *MongoMangaChapterRepository) HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	ratingColl := r.store.GetCollection("chapter_ratings")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_ratings", "read")
	defer cancel()

	count, err := ratingColl.CountDocuments(ctx, bson.M{"chapter_id": chapterID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("check chapter rating: %w", err)
	}
	return count > 0, nil
}

// ========== CHAPTER COMMENTS ==========

// AddChapterComment adds a comment to a chapter
func (r *MongoMangaChapterRepository) AddChapterComment(ctx context.Context, comment *coremanga.ChapterComment) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	commentColl := r.store.GetCollection("chapter_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comments", "write")
	defer cancel()

	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := commentColl.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("add chapter comment: %w", err)
	}
	return nil
}

// ListChapterComments lists comments for a chapter
func (r *MongoMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64) ([]*coremanga.ChapterComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	commentColl := r.store.GetCollection("chapter_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comments", "read")
	defer cancel()

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := commentColl.Find(ctx, bson.M{"chapter_id": chapterID}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list chapter comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []*coremanga.ChapterComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("decode chapter comments: %w", err)
	}

	total, err := commentColl.CountDocuments(ctx, bson.M{"chapter_id": chapterID})
	if err != nil {
		return nil, 0, fmt.Errorf("count chapter comments: %w", err)
	}

	return comments, total, nil
}

// DeleteChapterComment deletes a comment (user or admin only)
func (r *MongoMangaChapterRepository) DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	commentColl := r.store.GetCollection("chapter_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comments", "write")
	defer cancel()

	// Only allow deletion by comment creator
	result, err := commentColl.DeleteOne(ctx, bson.M{"_id": commentID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("delete chapter comment: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("comment not found or unauthorized")
	}
	return nil
}

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

	// Log view to manga_view_logs for period-based analytics
	mangaViewLogColl := r.store.Client.Database(r.store.DBName).Collection("manga_view_logs")
	_, err = mangaViewLogColl.InsertOne(ctx, bson.M{
		"manga_id":  mangaID.Hex(),
		"viewed_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("log manga view: %w", err)
	}

	return nil
}

func (r *MongoMangaChapterRepository) TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	viewColl := r.store.GetCollection("chapter_views")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_views", "write")
	defer cancel()

	_, err := viewColl.UpdateOne(ctx,
		bson.M{"chapter_id": chapterID, "user_id": userID},
		bson.M{"$setOnInsert": bson.M{"chapter_id": chapterID, "manga_id": mangaID, "user_id": userID, "created_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("track chapter view: %w", err)
	}
	return nil
}

func (r *MongoMangaChapterRepository) HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}
	viewColl := r.store.GetCollection("chapter_views")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_views", "read")
	defer cancel()

	count, err := viewColl.CountDocuments(ctx, bson.M{"chapter_id": chapterID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("check chapter view: %w", err)
	}
	return count > 0, nil
}

func (r *MongoMangaChapterRepository) TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return false, err
	}

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var isNewView bool
	var err error
	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			var err error
			isNewView, err = r.trackAndIncrementChapterViewInSession(sessCtx, chapterID, mangaID, userID)
			return err
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		isNewView, err = r.trackAndIncrementChapterViewWithoutTransaction(ctx, chapterID, mangaID, userID)
	}

	if err != nil {
		r.store.Log.Error("track chapter view failed", map[string]interface{}{
			"chapter_id":  chapterID.Hex(),
			"manga_id":    mangaID.Hex(),
			"user_id":     userID.Hex(),
			"replica_set": isReplicaSet,
			"error":       err.Error(),
		})
		return false, err
	}

	r.store.Log.Info("chapter view tracked", map[string]interface{}{
		"chapter_id":  chapterID.Hex(),
		"manga_id":    mangaID.Hex(),
		"user_id":     userID.Hex(),
		"is_new_view": isNewView,
		"replica_set": isReplicaSet,
	})
	return isNewView, nil
}

// trackAndIncrementChapterViewInSession performs atomic view tracking within a transaction session
func (r *MongoMangaChapterRepository) trackAndIncrementChapterViewInSession(sessCtx mongo.SessionContext, chapterID, mangaID, userID primitive.ObjectID) (bool, error) {
	viewColl := r.store.GetCollection("chapter_views")
	chapterColl := r.store.GetCollection("manga_chapters")
	mangaColl := r.store.GetCollection("manga")
	mangaViewLogColl := r.store.Client.Database(r.store.DBName).Collection("manga_view_logs")

	// Try to insert view record - if it already exists, this will not insert
	insertResult, err := viewColl.UpdateOne(sessCtx,
		bson.M{"chapter_id": chapterID, "user_id": userID},
		bson.M{"$setOnInsert": bson.M{
			"chapter_id": chapterID,
			"manga_id":   mangaID,
			"user_id":    userID,
			"created_at": time.Now(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return false, fmt.Errorf("upsert chapter view: %w", err)
	}

	// Check if this was a new insertion (UpsertedCount > 0 means new document)
	isNewView := insertResult.UpsertedCount > 0

	// Only increment counters if this was a new view
	if isNewView {
		// Increment chapter views
		_, err = chapterColl.UpdateOne(sessCtx,
			bson.M{"_id": chapterID},
			bson.M{"$inc": bson.M{"views_count": int64(1)}},
		)
		if err != nil {
			return false, fmt.Errorf("increment chapter views: %w", err)
		}

		// Increment manga total views
		_, err = mangaColl.UpdateOne(sessCtx,
			bson.M{"_id": mangaID},
			bson.M{"$inc": bson.M{"views_count": int64(1)}},
		)
		if err != nil {
			return false, fmt.Errorf("increment manga views: %w", err)
		}

		// Log view to manga_view_logs for period-based analytics
		_, err = mangaViewLogColl.InsertOne(sessCtx, bson.M{
			"manga_id":  mangaID.Hex(),
			"viewed_at": time.Now(),
		})
		if err != nil {
			return false, fmt.Errorf("log manga view: %w", err)
		}
	}

	return isNewView, nil
}

// trackAndIncrementChapterViewWithoutTransaction performs view tracking without transaction (for standalone MongoDB)
func (r *MongoMangaChapterRepository) trackAndIncrementChapterViewWithoutTransaction(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error) {
	viewColl := r.store.GetCollection("chapter_views")
	chapterColl := r.store.GetCollection("manga_chapters")
	mangaColl := r.store.GetCollection("manga")
	mangaViewLogColl := r.store.Client.Database(r.store.DBName).Collection("manga_view_logs")

	// Try to insert view record - if it already exists, this will not insert
	insertResult, err := viewColl.UpdateOne(ctx,
		bson.M{"chapter_id": chapterID, "user_id": userID},
		bson.M{"$setOnInsert": bson.M{
			"chapter_id": chapterID,
			"manga_id":   mangaID,
			"user_id":    userID,
			"created_at": time.Now(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return false, fmt.Errorf("upsert chapter view: %w", err)
	}

	// Check if this was a new insertion (UpsertedCount > 0 means new document)
	isNewView := insertResult.UpsertedCount > 0

	// Only increment counters if this was a new view
	if isNewView {
		// Increment chapter views
		_, err = chapterColl.UpdateOne(ctx,
			bson.M{"_id": chapterID},
			bson.M{"$inc": bson.M{"views_count": int64(1)}},
		)
		if err != nil {
			return false, fmt.Errorf("increment chapter views: %w", err)
		}

		// Increment manga total views
		_, err = mangaColl.UpdateOne(ctx,
			bson.M{"_id": mangaID},
			bson.M{"$inc": bson.M{"views_count": int64(1)}},
		)
		if err != nil {
			return false, fmt.Errorf("increment manga views: %w", err)
		}

		// Log view to manga_view_logs for period-based analytics
		_, err = mangaViewLogColl.InsertOne(ctx, bson.M{
			"manga_id":  mangaID.Hex(),
			"viewed_at": time.Now(),
		})
		if err != nil {
			return false, fmt.Errorf("log manga view: %w", err)
		}
	}

	return isNewView, nil
}

// ========== CHAPTER RATINGS ==========

// AddChapterRating adds a rating to a chapter and recalculates averages
func (r *MongoMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (newAverage float64, count int64, userScore float64, err error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, 0, 0, err
	}

	ratingColl := r.store.GetCollection("chapter_ratings")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_ratings", "write")
	defer cancel()

	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()

	// Upsert the rating (allow updates)
	_, err = ratingColl.UpdateOne(ctx,
		bson.M{"chapter_id": rating.ChapterID, "user_id": rating.UserID},
		bson.M{
			"$set": bson.M{
				"score":      rating.Score,
				"created_at": rating.CreatedAt,
				"updated_at": rating.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"chapter_id": rating.ChapterID,
				"user_id":    rating.UserID,
				"manga_id":   rating.MangaID,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("upsert chapter rating: %w", err)
	}

	// Fetch all ratings for this chapter to calculate actual count
	cursor, err := ratingColl.Find(ctx, bson.M{"chapter_id": rating.ChapterID})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("fetch chapter ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*coremanga.ChapterRating
	if err := cursor.All(ctx, &ratings); err != nil {
		return 0, 0, 0, fmt.Errorf("decode ratings: %w", err)
	}

	// Calculate average
	var sum float64
	count = int64(len(ratings))
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
			return 0, 0, 0, fmt.Errorf("update chapter averages: %w", err)
		}
	}

	// Recalculate manga aggregates from all its chapters
	chapColl := r.store.GetCollection("manga_chapters")
	cursor, err = chapColl.Find(ctx, bson.M{"manga_id": rating.MangaID})
	if err != nil {
		return avgRating, 0, rating.Score, fmt.Errorf("fetch manga chapters: %w", err)
	}
	defer cursor.Close(ctx)

	var chapters []*coremanga.MangaChapter
	if err := cursor.All(ctx, &chapters); err != nil {
		return avgRating, 0, rating.Score, fmt.Errorf("decode chapters: %w", err)
	}

	mangaAvgSum := 0.0
	validChapters := 0
	var mangaTotalCount int64
	var mangaTotalSum float64

	for _, ch := range chapters {
		if ch.AverageRating > 0 {
			mangaAvgSum += ch.AverageRating
			validChapters++
		}
		mangaTotalCount += ch.RatingCount
		mangaTotalSum += ch.RatingSum
	}

	if validChapters > 0 {
		mangaAvg := mangaAvgSum / float64(validChapters)
		mangaColl := r.store.GetCollection("manga")
		_, err = mangaColl.UpdateOne(ctx,
			bson.M{"_id": rating.MangaID},
			bson.M{"$set": bson.M{
				"average_rating": mangaAvg,
				"rating_count":   mangaTotalCount,
				"rating_sum":     mangaTotalSum,
			}},
		)
		if err != nil {
			return avgRating, 0, rating.Score, fmt.Errorf("update manga aggregates: %w", err)
		}
	} else {
		// Even if no chapters have ratings yet, update count and sum to ensure consistency
		mangaColl := r.store.GetCollection("manga")
		_, err = mangaColl.UpdateOne(ctx,
			bson.M{"_id": rating.MangaID},
			bson.M{"$set": bson.M{
				"rating_count": mangaTotalCount,
				"rating_sum":   mangaTotalSum,
			}},
		)
		if err != nil {
			return avgRating, 0, rating.Score, fmt.Errorf("update manga rating counts: %w", err)
		}
	}

	return avgRating, count, rating.Score, nil
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

// GetUserChapterRating retrieves the user's rating for a specific chapter
func (r *MongoMangaChapterRepository) GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error) {
	if err := r.ensureInitialized(); err != nil {
		return 0, false, err
	}

	ratingColl := r.store.GetCollection("chapter_ratings")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_ratings", "read")
	defer cancel()

	var rating coremanga.ChapterRating
	err := ratingColl.FindOne(ctx, bson.M{"chapter_id": chapterID, "user_id": userID}).Decode(&rating)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, false, nil // No rating found
		}
		return 0, false, fmt.Errorf("get chapter rating: %w", err)
	}
	return rating.Score, true, nil
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

	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}
	comment.UpdatedAt = time.Now()

	_, err := commentColl.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("add chapter comment: %w", err)
	}
	return nil
}

// ListChapterComments lists comments for a chapter
func (r *MongoMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.ChapterComment, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}

	commentColl := r.store.GetCollection("chapter_comments")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comments", "read")
	defer cancel()

	sort := bson.M{"created_at": -1} // default: newest first
	if sortOrder == "oldest" {
		sort = bson.M{"created_at": 1}
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
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

// ========== CHAPTER COMMENT REACTIONS (Like/Dislike) ==========

// AddChapterCommentReaction adds a like/dislike to a chapter comment
func (r *MongoMangaChapterRepository) AddChapterCommentReaction(ctx context.Context, reaction *coremanga.ChapterCommentReaction) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	reactionColl := r.store.GetCollection("chapter_comment_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comment_reactions", "write")
	defer cancel()

	if reaction.ID.IsZero() {
		reaction.ID = primitive.NewObjectID()
	}
	if reaction.CreatedAt.IsZero() {
		reaction.CreatedAt = time.Now()
	}

	// Check if user already reacted
	existingCtx, existingCancel := r.store.WithCollectionTimeout(ctx, "chapter_comment_reactions", "read")
	defer existingCancel()

	var existingReaction struct {
		Type string `bson:"type"`
	}
	err := reactionColl.FindOne(existingCtx, bson.M{
		"comment_id": reaction.CommentID,
		"user_id":    reaction.UserID,
	}).Decode(&existingReaction)

	if err == nil {
		// User already reacted - update if different type
		if existingReaction.Type != reaction.Type {
			_, err = reactionColl.UpdateOne(ctx,
				bson.M{"comment_id": reaction.CommentID, "user_id": reaction.UserID},
				bson.M{"$set": bson.M{"type": reaction.Type}},
			)
			if err != nil {
				return fmt.Errorf("update comment reaction: %w", err)
			}

			// Update counts
			return r.updateCommentReactionCounts(ctx, reaction.CommentID)
		}
		return nil // Same reaction, do nothing
	}

	// Insert new reaction
	_, err = reactionColl.InsertOne(ctx, reaction)
	if err != nil {
		return fmt.Errorf("insert comment reaction: %w", err)
	}

	// Update comment counts
	return r.updateCommentReactionCounts(ctx, reaction.CommentID)
}

// RemoveChapterCommentReaction removes a like/dislike from a chapter comment
func (r *MongoMangaChapterRepository) RemoveChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	reactionColl := r.store.GetCollection("chapter_comment_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comment_reactions", "write")
	defer cancel()

	result, err := reactionColl.DeleteOne(ctx, bson.M{
		"comment_id": commentID,
		"user_id":    userID,
	})
	if err != nil {
		return fmt.Errorf("delete comment reaction: %w", err)
	}
	if result.DeletedCount == 0 {
		return nil // Already removed
	}

	// Update comment counts
	return r.updateCommentReactionCounts(ctx, commentID)
}

// GetUserChapterCommentReaction gets user's reaction to a comment
func (r *MongoMangaChapterRepository) GetUserChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error) {
	if err := r.ensureInitialized(); err != nil {
		return "", err
	}

	reactionColl := r.store.GetCollection("chapter_comment_reactions")
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "chapter_comment_reactions", "read")
	defer cancel()

	var reaction struct {
		Type string `bson:"type"`
	}
	err := reactionColl.FindOne(ctx, bson.M{
		"comment_id": commentID,
		"user_id":    userID,
	}).Decode(&reaction)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", fmt.Errorf("get comment reaction: %w", err)
	}

	return reaction.Type, nil
}

// updateCommentReactionCounts updates the like/dislike counts for a comment
func (r *MongoMangaChapterRepository) updateCommentReactionCounts(ctx context.Context, commentID primitive.ObjectID) error {
	reactionColl := r.store.GetCollection("chapter_comment_reactions")
	commentColl := r.store.GetCollection("chapter_comments")

	// Count likes
	likeCount, err := reactionColl.CountDocuments(ctx, bson.M{
		"comment_id": commentID,
		"type":       "like",
	})
	if err != nil {
		return fmt.Errorf("count likes: %w", err)
	}

	// Count dislikes
	dislikeCount, err := reactionColl.CountDocuments(ctx, bson.M{
		"comment_id": commentID,
		"type":       "dislike",
	})
	if err != nil {
		return fmt.Errorf("count dislikes: %w", err)
	}

	// Update comment
	_, err = commentColl.UpdateOne(ctx,
		bson.M{"_id": commentID},
		bson.M{"$set": bson.M{
			"like_count":    likeCount,
			"dislike_count": dislikeCount,
		}},
	)
	if err != nil {
		return fmt.Errorf("update comment counts: %w", err)
	}

	return nil
}

package manga

import (
	"context"
	"time"

	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoViewingHistoryRepository struct {
	store      *mongoinfra.MongoStore
	collection *mongo.Collection
	log        *logger.Logger
}

func NewMongoViewingHistoryRepository(s *mongoinfra.MongoStore, log *logger.Logger) coremanga.ViewingHistoryRepository {
	return &mongoViewingHistoryRepository{
		store:      s,
		collection: s.GetCollection("viewing_history"),
		log:        log,
	}
}

func (r *mongoViewingHistoryRepository) CreateHistory(ctx context.Context, history *coremanga.ViewingHistory) error {
	r.log.Debug("creating viewing history", map[string]interface{}{
		"user_id":    history.UserID,
		"manga_id":   history.MangaID,
		"chapter_id": history.ChapterID,
	})

	history.ID = primitive.NewObjectID()
	history.CreatedAt = time.Now()
	history.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, history)
	if err != nil {
		r.log.Error("failed to create viewing history", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (r *mongoViewingHistoryRepository) UpdateHistory(ctx context.Context, history *coremanga.ViewingHistory) error {
	r.log.Debug("updating viewing history", map[string]interface{}{
		"id":      history.ID,
		"user_id": history.UserID,
	})

	history.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"chapter_id": history.ChapterID,
			"page":       history.Page,
			"viewed_at":  history.ViewedAt,
			"duration":   history.Duration,
			"updated_at": history.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": history.ID}, update)
	if err != nil {
		r.log.Error("failed to update viewing history", map[string]interface{}{
			"error": err.Error(),
			"id":    history.ID,
		})
		return err
	}

	return nil
}

func (r *mongoViewingHistoryRepository) GetHistoryByUser(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.ViewingHistory, int64, error) {
	r.log.Debug("getting history by user", map[string]interface{}{
		"user_id": userID,
		"skip":    skip,
		"limit":   limit,
	})

	// Count total
	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		r.log.Error("failed to count viewing history", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, 0, err
	}

	// Get history
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "viewed_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		r.log.Error("failed to get viewing history", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var histories []*coremanga.ViewingHistory
	if err := cursor.All(ctx, &histories); err != nil {
		r.log.Error("failed to decode viewing history", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, err
	}

	return histories, count, nil
}

func (r *mongoViewingHistoryRepository) GetHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) (*coremanga.ViewingHistory, error) {
	r.log.Debug("getting history by manga", map[string]interface{}{
		"user_id":  userID,
		"manga_id": mangaID,
	})

	var history coremanga.ViewingHistory
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":  userID,
		"manga_id": mangaID,
	}).Decode(&history)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		r.log.Error("failed to get viewing history by manga", map[string]interface{}{
			"error":    err.Error(),
			"user_id":  userID,
			"manga_id": mangaID,
		})
		return nil, err
	}

	return &history, nil
}

func (r *mongoViewingHistoryRepository) GetHistoryByChapter(ctx context.Context, userID, chapterID primitive.ObjectID) (*coremanga.ViewingHistory, error) {
	r.log.Debug("getting history by chapter", map[string]interface{}{
		"user_id":    userID,
		"chapter_id": chapterID,
	})

	var history coremanga.ViewingHistory
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":    userID,
		"chapter_id": chapterID,
	}).Decode(&history)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		r.log.Error("failed to get viewing history by chapter", map[string]interface{}{
			"error":      err.Error(),
			"user_id":    userID,
			"chapter_id": chapterID,
		})
		return nil, err
	}

	return &history, nil
}

func (r *mongoViewingHistoryRepository) GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*coremanga.ViewingHistory, error) {
	r.log.Debug("getting recent history", map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
	})

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "viewed_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		r.log.Error("failed to get recent viewing history", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}
	defer cursor.Close(ctx)

	var histories []*coremanga.ViewingHistory
	if err := cursor.All(ctx, &histories); err != nil {
		r.log.Error("failed to decode recent viewing history", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return histories, nil
}

func (r *mongoViewingHistoryRepository) GetUserStats(ctx context.Context, userID primitive.ObjectID) (*coremanga.HistoryStats, error) {
	r.log.Debug("getting user stats", map[string]interface{}{
		"user_id": userID,
	})

	// Total views
	totalViews, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Unique manga count
	uniqueManga, err := r.collection.Distinct(ctx, "manga_id", bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Unique chapters count
	uniqueChapters, err := r.collection.Distinct(ctx, "chapter_id", bson.M{
		"user_id":    userID,
		"chapter_id": bson.M{"$ne": primitive.NilObjectID},
	})
	if err != nil {
		return nil, err
	}

	// Total duration
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total_duration": bson.M{"$sum": "$duration"}}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var totalDuration int64
	if cursor.Next(ctx) {
		var result struct {
			TotalDuration int64 `bson:"total_duration"`
		}
		if err := cursor.Decode(&result); err == nil {
			totalDuration = result.TotalDuration
		}
	}

	stats := &coremanga.HistoryStats{
		TotalViews:     totalViews,
		UniqueManga:    int64(len(uniqueManga)),
		UniqueChapters: int64(len(uniqueChapters)),
		TotalDuration:  totalDuration,
	}

	return stats, nil
}

func (r *mongoViewingHistoryRepository) DeleteHistory(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	r.log.Debug("deleting viewing history", map[string]interface{}{
		"id":      id,
		"user_id": userID,
	})

	_, err := r.collection.DeleteOne(ctx, bson.M{
		"_id":     id,
		"user_id": userID,
	})
	if err != nil {
		r.log.Error("failed to delete viewing history", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	return nil
}

func (r *mongoViewingHistoryRepository) DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error {
	r.log.Debug("deleting history by manga", map[string]interface{}{
		"user_id":  userID,
		"manga_id": mangaID,
	})

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"user_id":  userID,
		"manga_id": mangaID,
	})
	if err != nil {
		r.log.Error("failed to delete history by manga", map[string]interface{}{
			"error":    err.Error(),
			"user_id":  userID,
			"manga_id": mangaID,
		})
		return err
	}

	return nil
}

func (r *mongoViewingHistoryRepository) DeleteHistoryByChapter(ctx context.Context, userID, chapterID primitive.ObjectID) error {
	r.log.Debug("deleting history by chapter", map[string]interface{}{
		"user_id":    userID,
		"chapter_id": chapterID,
	})

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"user_id":    userID,
		"chapter_id": chapterID,
	})
	if err != nil {
		r.log.Error("failed to delete history by chapter", map[string]interface{}{
			"error":      err.Error(),
			"user_id":    userID,
			"chapter_id": chapterID,
		})
		return err
	}

	return nil
}

func (r *mongoViewingHistoryRepository) DeleteOlderThan(ctx context.Context, userID primitive.ObjectID, before time.Time) (int64, error) {
	r.log.Debug("deleting old history", map[string]interface{}{
		"user_id": userID,
		"before":  before,
	})

	result, err := r.collection.DeleteMany(ctx, bson.M{
		"user_id":   userID,
		"viewed_at": bson.M{"$lt": before},
	})
	if err != nil {
		r.log.Error("failed to delete old history", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return 0, err
	}

	return result.DeletedCount, nil
}

// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/mongo_reading_progress_repository.go -----
package manga

import (
	"context"
	"fmt"
	"time"

	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoReadingProgressRepository struct {
	store *mongoinfra.MongoStore
	coll  *mongo.Collection
}

func NewMongoReadingProgressRepository(store *mongoinfra.MongoStore) coremanga.ReadingProgressRepository {
	return &MongoReadingProgressRepository{
		store: store,
		coll:  store.GetCollection("reading_progress"),
	}
}

func (r *MongoReadingProgressRepository) SaveProgress(ctx context.Context, progress *coremanga.ReadingProgress) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "reading_progress", "write")
	defer cancel()

	if progress.ID.IsZero() {
		progress.ID = primitive.NewObjectID()
	}
	if progress.CreatedAt.IsZero() {
		progress.CreatedAt = time.Now()
	}
	progress.UpdatedAt = time.Now()

	// Upsert the progress
	_, err := r.coll.UpdateOne(ctx,
		bson.M{
			"manga_id": progress.MangaID,
			"user_id":  progress.UserID,
		},
		bson.M{
			"$set": bson.M{
				"last_read_chapter": progress.LastReadChapter,
				"last_read_page":    progress.LastReadPage,
				"last_read_at":      progress.LastReadAt,
				"updated_at":        progress.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"_id":        progress.ID,
				"manga_id":   progress.MangaID,
				"user_id":    progress.UserID,
				"created_at": progress.CreatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert reading progress: %w", err)
	}

	return nil
}

func (r *MongoReadingProgressRepository) GetProgress(ctx context.Context, mangaID, userID primitive.ObjectID) (*coremanga.ReadingProgress, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "reading_progress", "read")
	defer cancel()

	var progress coremanga.ReadingProgress
	err := r.coll.FindOne(ctx, bson.M{
		"manga_id": mangaID,
		"user_id":  userID,
	}).Decode(&progress)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("get reading progress: %w", err)
	}

	return &progress, nil
}

func (r *MongoReadingProgressRepository) GetProgressForMangas(ctx context.Context, mangaIDs []primitive.ObjectID, userID primitive.ObjectID) (map[string]*coremanga.ReadingProgress, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}

	if len(mangaIDs) == 0 {
		return make(map[string]*coremanga.ReadingProgress), nil
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "reading_progress", "read")
	defer cancel()

	// Convert ObjectIDs to strings for the query
	mangaIDStrings := make([]string, len(mangaIDs))
	for i, id := range mangaIDs {
		mangaIDStrings[i] = id.Hex()
	}

	cursor, err := r.coll.Find(ctx, bson.M{
		"manga_id": bson.M{"$in": mangaIDStrings},
		"user_id":  userID.Hex(),
	})
	if err != nil {
		return nil, fmt.Errorf("find reading progress: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]*coremanga.ReadingProgress)
	for cursor.Next(ctx) {
		var progress coremanga.ReadingProgress
		if err := cursor.Decode(&progress); err != nil {
			continue
		}
		result[progress.MangaID.Hex()] = &progress
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return result, nil
}

func (r *MongoReadingProgressRepository) DeleteProgress(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "reading_progress", "write")
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{
		"manga_id": mangaID.Hex(),
		"user_id":  userID.Hex(),
	})
	if err != nil {
		return fmt.Errorf("delete reading progress: %w", err)
	}

	return nil
}

func (r *MongoReadingProgressRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.coll == nil {
		return fmt.Errorf("reading progress repository not initialized")
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/content/manga/mongo_reading_progress_repository.go -----

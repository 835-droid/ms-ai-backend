package mongo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexConfig defines a collection's indexes
type IndexConfig struct {
	Collection string
	Indexes    []mongo.IndexModel
}

// GetAllIndexConfigs returns the index configurations for all collections
func GetAllIndexConfigs() []IndexConfig {
	return []IndexConfig{
		{
			Collection: "users",
			Indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "username", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "email", Value: 1}},
					Options: options.Index().SetUnique(true).SetSparse(true),
				},
				{
					Keys: bson.D{
						{Key: "is_active", Value: 1},
						{Key: "created_at", Value: -1},
					},
					Options: options.Index(),
				},
				{
					Keys:    bson.D{{Key: "refresh_token", Value: 1}},
					Options: options.Index().SetSparse(true), // SetUnique removed to match existing index
				},
				// add commonly queried fields for permissions/roles
				{
					Keys:    bson.D{{Key: "is_active", Value: 1}, {Key: "roles", Value: 1}},
					Options: options.Index(),
				},
			},
		},
		{
			Collection: "invite_codes",
			Indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "code", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "is_used", Value: 1}, {Key: "expires_at", Value: 1}},
					Options: options.Index(),
				},
			},
		},
		{
			Collection: "manga",
			Indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "author_id", Value: 1}},
					Options: options.Index(),
				},
				{
					Keys:    bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
					Options: options.Index(),
				},
				// text index for title/description
				{
					Keys:    bson.D{{Key: "$**", Value: "text"}},
					Options: options.Index().SetName("manga_text_idx"),
				},
			},
		},
		{
			Collection: "manga_chapters",
			Indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "manga_id", Value: 1}, {Key: "number", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "manga_id", Value: 1}, {Key: "created_at", Value: -1}},
					Options: options.Index(),
				},
			},
		},
	}
}

// createManyIndexes creates multiple indexes for a collection with proper error handling
func createManyIndexes(ctx context.Context, coll *mongo.Collection, models []mongo.IndexModel, name string) error {
	if coll == nil {
		return fmt.Errorf("collection %s is nil", name)
	}
	if len(models) == 0 {
		return nil
	}

	// Create indexes individually with simple retry and jitter to improve robustness and logging
	for idx, m := range models {
		// Try a few times for each index
		var lastErr error
		for attempt := 1; attempt <= 3; attempt++ {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			nameOpt := m.Options
			// Compose an index name for logging
			idxName := ""
			if nameOpt != nil && nameOpt.Name != nil {
				idxName = *nameOpt.Name
			}
			if idxName == "" {
				idxName = fmt.Sprintf("%s_idx_%d", name, idx)
			}

			_, err := coll.Indexes().CreateOne(ctx, m)
			if err == nil {
				// created
				break
			}

			lastErr = err
			// log and retry with jitter
			jitter := time.Duration(rand.Intn(200)+50) * time.Millisecond
			time.Sleep(time.Duration(attempt)*100*time.Millisecond + jitter)
		}
		if lastErr != nil {
			return fmt.Errorf("create %s index #%d: %w", name, idx, lastErr)
		}
	}

	return nil
}

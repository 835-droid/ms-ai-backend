// ----- START OF FILE: backend/MS-AI/internal/container/container.go -----
package container

import (
	"context"
	"fmt"

	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

func NewContainer(cfg *config.Config, log *logger.Logger) (*Container, error) {
	ctx := context.Background()

	mStore, pStore, err := initializeDatabases(cfg, log)
	if err != nil {
		return nil, err
	}

	repos := initializeRepositories(cfg, log, mStore, pStore)
	svcs := initializeServices(cfg, log, repos, mStore, pStore)
	initializeInitialData(ctx, cfg, repos, log)

	return &Container{
		Config:     cfg,
		Logger:     log,
		MongoDB:    mStore,
		PostgresDB: pStore,
		UserRepo:   repos.User,
		Handlers:   initializeHandlers(svcs, mStore, pStore),
	}, nil
}

func InitializeRepositories(cfg *config.Config, log *logger.Logger, m *mongo.MongoStore, p *postgres.PostgresStore) *RepoBundle {
	return initializeRepositories(cfg, log, m, p)
}

func (c *Container) Close(ctx context.Context) error {
	var errEncountered error
	if c.MongoDB != nil {
		if err := c.MongoDB.Close(ctx); err != nil {
			errEncountered = fmt.Errorf("mongo close: %w", err)
		}
	}
	if c.PostgresDB != nil {
		if err := c.PostgresDB.Close(); err != nil {
			if errEncountered != nil {
				errEncountered = fmt.Errorf("%v; postgres close: %w", errEncountered, err)
			} else {
				errEncountered = fmt.Errorf("postgres close: %w", err)
			}
		}
	}
	return errEncountered
}

// ----- END OF FILE: backend/MS-AI/internal/container/container.go -----

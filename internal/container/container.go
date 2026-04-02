// ----- START OF FILE: backend/MS-AI/internal/container/container.go -----
package container

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
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
		Config:   cfg,
		Logger:   log,
		MongoDB:  mStore,
		UserRepo: repos.User,
		Handlers: initializeHandlers(svcs, mStore, pStore),
	}, nil
}

func InitializeDatabases(cfg *config.Config, log *logger.Logger) (*mongo.MongoStore, *postgres.PostgresStore, error) {
	return initializeDatabases(cfg, log)
}

func InitializeRepositories(cfg *config.Config, log *logger.Logger, m *mongo.MongoStore, p *postgres.PostgresStore) *RepoBundle {
	return initializeRepositories(cfg, log, m, p)
}

func (c *Container) Close(ctx context.Context) error {
	if c.MongoDB != nil {
		return c.MongoDB.Close(ctx)
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/container/container.go -----

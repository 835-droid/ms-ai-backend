// internal/container/handler_initializers.go
package container

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
)

func initializeHandlers(svcs *serviceBundle, m *mongoinfra.MongoStore, p *pginfra.PostgresStore) *handler.Container {
	return handler.NewContainer(svcs.Auth, svcs.Manga, svcs.Chapter, svcs.Admin, m, p)
}

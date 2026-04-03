// internal/container/initializers.go
package container

import (
	"fmt"

	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

// initializeDatabases attempts to connect to databases according to DB_TYPE.
// If the primary database fails, it attempts to failover to the other if available.
// It may modify cfg.DBType to reflect the actual working database.
func initializeDatabases(cfg *config.Config, log *logger.Logger) (*mongoinfra.MongoStore, *pginfra.PostgresStore, error) {
	var m *mongoinfra.MongoStore
	var p *pginfra.PostgresStore
	var err error

	switch cfg.DBType {
	case "mongo":
		// Try MongoDB
		m, err = mongoinfra.NewMongoStore(cfg, log)
		if err != nil {
			if cfg.DBFailoverEnabled {
				log.Warn("MongoDB connection failed and failover is enabled, attempting PostgreSQL",
					map[string]interface{}{"error": err.Error()})
				if cfg.PostgresDSN != "" {
					p, err2 := pginfra.NewPostgresStore(cfg, log)
					if err2 == nil {
						log.Info("Failover to PostgreSQL successful. Switching DB_TYPE to postgres", nil)
						cfg.DBType = "postgres"
						return nil, p, nil
					}
					log.Warn("PostgreSQL failover also failed",
						map[string]interface{}{"mongo_error": err.Error(), "postgres_error": err2.Error()})
				}
			}
			// Strict DB_TYPE behavior: fail fast when configured as mongo.
			return nil, nil, err
		}
		return m, p, nil

	case "postgres":
		// Try PostgreSQL
		p, err = pginfra.NewPostgresStore(cfg, log)
		if err != nil {
			if cfg.DBFailoverEnabled {
				log.Warn("PostgreSQL connection failed and failover is enabled, attempting MongoDB",
					map[string]interface{}{"error": err.Error()})
				if cfg.MongoURI != "" {
					m, err2 := mongoinfra.NewMongoStore(cfg, log)
					if err2 == nil {
						log.Info("Failover to MongoDB successful. Switching DB_TYPE to mongo", nil)
						cfg.DBType = "mongo"
						return m, p, nil
					}
					log.Warn("MongoDB failover also failed",
						map[string]interface{}{"postgres_error": err.Error(), "mongo_error": err2.Error()})
				}
			}
			// Strict DB_TYPE behavior: fail fast when configured as postgres.
			return nil, nil, err
		}
		return m, p, nil

	case "hybrid":
		// Try both databases for hybrid mode
		m, mongoErr := mongoinfra.NewMongoStore(cfg, log)
		p, postgresErr := pginfra.NewPostgresStore(cfg, log)

		if mongoErr != nil && postgresErr != nil {
			return nil, nil, fmt.Errorf("both databases failed: mongo=%v, postgres=%v", mongoErr, postgresErr)
		}

		if mongoErr != nil {
			log.Warn("MongoDB connection failed in hybrid mode", map[string]interface{}{"error": mongoErr.Error()})
		}
		if postgresErr != nil {
			log.Warn("PostgreSQL connection failed in hybrid mode", map[string]interface{}{"error": postgresErr.Error()})
		}

		return m, p, nil

	default:
		return nil, nil, fmt.Errorf("unknown DB_TYPE: %s", cfg.DBType)
	}
}

func InitializeDatabases(cfg *config.Config, log *logger.Logger) (*mongo.MongoStore, *postgres.PostgresStore, error) {
	return initializeDatabases(cfg, log)
}

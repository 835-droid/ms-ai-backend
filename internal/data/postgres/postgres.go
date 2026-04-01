package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func NewPostgresStore(cfg *config.Config, log *logger.Logger) (*PostgresStore, error) {
	// First, connect to the default postgres database to create our database if needed
	defaultDSN := strings.Replace(cfg.PostgresDSN, "/msai", "/postgres", 1)
	defaultDB, err := sqlx.Connect("postgres", defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default postgres db: %w", err)
	}
	defer defaultDB.Close()

	// Create the database if it doesn't exist
	_, err = defaultDB.Exec("CREATE DATABASE msai WITH OWNER = admin_user ENCODING = 'UTF8'")
	if err != nil {
		// Ignore error if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			log.Warn("failed to create database", map[string]interface{}{"error": err.Error()})
		}
	}

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	if err := createTables(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Info("postgresql initialized successfully", nil)
	return &PostgresStore{DB: db, Log: log}, nil
}

func createTables(ctx context.Context, db *sqlx.DB) error {
	// Create sequence for user IDs
	_, err := db.ExecContext(ctx, `CREATE SEQUENCE IF NOT EXISTS user_id_counter`)
	if err != nil {
		return err
	}

	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(24) PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		roles TEXT DEFAULT '{}',
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		refresh_token TEXT,
		refresh_token_expires_at TIMESTAMP
	);`

	userDetailsTable := `
	CREATE TABLE IF NOT EXISTS user_details (
		uuid VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(24) REFERENCES users(id) ON DELETE CASCADE,
		roles TEXT DEFAULT '{}',
		is_active BOOLEAN DEFAULT true,
		status VARCHAR(50) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	inviteCodesTable := `
	CREATE TABLE IF NOT EXISTS invite_codes (
		id VARCHAR(24) PRIMARY KEY,
		code VARCHAR(255) UNIQUE NOT NULL,
		is_used BOOLEAN DEFAULT false,
		used_by VARCHAR(24),
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	tables := []string{usersTable, userDetailsTable, inviteCodesTable}
	for _, query := range tables {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

// Healthy checks if the PostgreSQL connection is healthy
func (s *PostgresStore) Healthy(ctx context.Context) error {
	if s == nil || s.DB == nil {
		return fmt.Errorf("postgres store is nil")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

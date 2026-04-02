// ----- START OF FILE: backend/MS-AI/internal/data/postgres/postgres.go -----
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
	defaultDSN := strings.Replace(cfg.PostgresDSN, "/msai", "/postgres", 1)
	defaultDB, err := sqlx.Connect("postgres", defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default postgres db: %w", err)
	}
	defer defaultDB.Close()

	var dbExists bool
	var dbName string
	parts := strings.Split(cfg.PostgresDSN, "/")
	if len(parts) > 3 {
		dbName = strings.Split(parts[3], "?")[0]
	} else {
		dbName = "msai"
	}
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err = defaultDB.QueryRow(query, dbName).Scan(&dbExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !dbExists {
		_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s WITH OWNER = %s ENCODING = 'UTF8'", dbName, "admin_user"))
		if err != nil {
			log.Warn("failed to create database (permission issue?), assuming it exists",
				map[string]interface{}{"error": err.Error(), "database": dbName})
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

	if err := ensureTables(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to ensure tables: %w", err)
	}

	if err := RunMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("postgresql initialized successfully", nil)
	return &PostgresStore{DB: db, Log: log}, nil
}

func ensureTables(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `CREATE SEQUENCE IF NOT EXISTS user_id_counter`)
	if err != nil {
		return fmt.Errorf("create sequence: %w", err)
	}

	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(24) PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		roles TEXT NOT NULL DEFAULT '["user"]',
		is_active BOOLEAN DEFAULT true,
		last_login_at TIMESTAMP NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		refresh_token TEXT,
		refresh_token_expires_at TIMESTAMP
	);`

	userDetailsTable := `
	CREATE TABLE IF NOT EXISTS user_details (
		uuid VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(24) REFERENCES users(id) ON DELETE CASCADE,
		roles TEXT NOT NULL DEFAULT '["user"]',
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

	createQueries := []string{usersTable, userDetailsTable, inviteCodesTable}
	for _, query := range createQueries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	return nil
}

func (s *PostgresStore) Healthy(ctx context.Context) error {
	if s == nil || s.DB == nil {
		return fmt.Errorf("postgres store is nil")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

func RunMigrations(ctx context.Context, db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("postgres db is nil")
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS mangas (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    author_id TEXT,
    tags JSONB DEFAULT '[]',
    cover_image TEXT,
    is_published BOOLEAN DEFAULT false,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    views_count BIGINT DEFAULT 0,
    likes_count BIGINT DEFAULT 0,
    rating_sum DOUBLE PRECISION DEFAULT 0,
    rating_count BIGINT DEFAULT 0,
    average_rating DOUBLE PRECISION DEFAULT 0
);`,
		`CREATE TABLE IF NOT EXISTS manga_likes (
    manga_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    reaction_type VARCHAR(50) DEFAULT 'upvote',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (manga_id, user_id)
);`,
		`CREATE TABLE IF NOT EXISTS manga_ratings (
    manga_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (manga_id, user_id)
);`,
		`CREATE TABLE IF NOT EXISTS manga_chapters (
    id TEXT PRIMARY KEY,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    pages JSONB DEFAULT '[]',
    number INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (manga_id, number)
);`,
		// Forward-compatible migration: add reaction_type column if it doesn't exist
		`ALTER TABLE IF EXISTS manga_likes ADD COLUMN IF NOT EXISTS reaction_type VARCHAR(50) DEFAULT 'upvote';`,
		// Add created_at column if it doesn't exist
		`ALTER TABLE IF EXISTS manga_likes ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/postgres/postgres.go -----

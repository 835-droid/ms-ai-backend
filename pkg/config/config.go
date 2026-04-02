// ----- START OF FILE: backend/MS-AI/pkg/config/config.go -----
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI                 string
	DBName                   string
	ServerPort               string
	JWTSecret                string
	JWTAccessExpiry          time.Duration
	JWTRefreshExpiry         time.Duration
	RateLimitRequests        int
	RateLimitWindow          time.Duration
	LogLevel                 string
	Environment              string
	CORSOrigins              string
	AllowedOriginsSet        map[string]struct{}
	AllowWebSocketQueryToken bool
	DBTimeout                time.Duration

	MongoUsername           string
	MongoPassword           string
	MongoAuthSource         string
	MongoMaxPoolSize        uint64
	MongoMinPoolSize        uint64
	MongoEnableMonitoring   bool
	MongoSlowQueryThreshold time.Duration

	PostgresDSN string
	DBType      string
}

func LoadConfig() (*Config, error) {
	fmt.Fprintf(os.Stderr, "DB_TYPE before load: %s\n", os.Getenv("DB_TYPE"))
	// Always try to load .env for development
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load .env file: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, ".env loaded successfully\n")
	}
	fmt.Fprintf(os.Stderr, "DB_TYPE after load: %s\n", os.Getenv("DB_TYPE"))

	cfg := &Config{
		MongoURI:                 getenv("MONGO_URI", ""),
		DBName:                   getenv("DB_NAME", "MSAIDB"),
		ServerPort:               getenv("SERVER_PORT", "8080"),
		JWTSecret:                getenv("JWT_SECRET", "defaultsecretthatislongenoughforvalidationpurposes1234567890123456789012345678901234567890"),
		LogLevel:                 getenv("LOG_LEVEL", "info"),
		Environment:              getenv("ENVIRONMENT", "development"),
		CORSOrigins:              getenv("CORS_ORIGINS", "*"),
		AllowWebSocketQueryToken: strings.ToLower(getenv("WEBSOCKET_ALLOW_QUERY_TOKEN", "false")) == "true",
	}

	cfg.DBType = strings.ToLower(getenv("DB_TYPE", "postgres"))
	cfg.PostgresDSN = getenv("POSTGRES_DSN", "")

	// MongoDB config only needed if DBType is mongo or hybrid
	if cfg.DBType == "mongo" || cfg.DBType == "hybrid" {
		cfg.MongoUsername = getenv("MONGO_USERNAME", "")
		cfg.MongoPassword = getenv("MONGO_PASSWORD", "")
		cfg.MongoAuthSource = getenv("MONGO_AUTH_SOURCE", "admin")

		if v := getenv("MONGO_MAX_POOL_SIZE", "100"); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				cfg.MongoMaxPoolSize = n
			} else {
				return nil, fmt.Errorf("invalid MONGO_MAX_POOL_SIZE: %w", err)
			}
		}
		if v := getenv("MONGO_MIN_POOL_SIZE", "10"); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				cfg.MongoMinPoolSize = n
			} else {
				return nil, fmt.Errorf("invalid MONGO_MIN_POOL_SIZE: %w", err)
			}
		}
		if v := getenv("MONGO_ENABLE_MONITORING", "true"); strings.ToLower(v) == "true" {
			cfg.MongoEnableMonitoring = true
		}
		if v := getenv("MONGO_SLOW_QUERY_THRESHOLD", "1s"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				cfg.MongoSlowQueryThreshold = d
			} else {
				return nil, fmt.Errorf("invalid MONGO_SLOW_QUERY_THRESHOLD: %w", err)
			}
		}
	}

	// JWT durations
	access := getenv("JWT_ACCESS_EXPIRY", "15m")
	if d, err := time.ParseDuration(access); err == nil {
		cfg.JWTAccessExpiry = d
	} else {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}
	refresh := getenv("JWT_REFRESH_EXPIRY", "168h")
	if d, err := time.ParseDuration(refresh); err == nil {
		cfg.JWTRefreshExpiry = d
	} else {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}
	rlw := getenv("RATE_LIMIT_WINDOW", "1m")
	if d, err := time.ParseDuration(rlw); err == nil {
		cfg.RateLimitWindow = d
	} else {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %w", err)
	}
	if v := getenv("RATE_LIMIT_REQUESTS", "100"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RateLimitRequests = n
		} else {
			return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
		}
	}
	if dt := getenv("DB_TIMEOUT", "10s"); dt != "" {
		if d, err := time.ParseDuration(dt); err == nil {
			cfg.DBTimeout = d
		} else {
			return nil, fmt.Errorf("invalid DB_TIMEOUT: %w", err)
		}
	}

	if cfg.CORSOrigins != "*" {
		cfg.AllowedOriginsSet = make(map[string]struct{})
		for _, o := range strings.Split(cfg.CORSOrigins, ",") {
			if origin := strings.ToLower(strings.TrimSpace(o)); origin != "" {
				cfg.AllowedOriginsSet[origin] = struct{}{}
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	// Check DBType first
	switch c.DBType {
	case "mongo", "hybrid":
		if c.MongoURI == "" {
			return errors.New("MONGO_URI is required when DB_TYPE is mongo or hybrid")
		}
		if strings.ToLower(c.Environment) == "production" {
			if strings.TrimSpace(c.MongoUsername) == "" || strings.TrimSpace(c.MongoPassword) == "" {
				return errors.New("MONGO_USERNAME and MONGO_PASSWORD are required in production")
			}
		}
	case "postgres":
		if c.PostgresDSN == "" {
			return errors.New("POSTGRES_DSN is required when DB_TYPE is postgres")
		}
	default:
		return errors.New("DB_TYPE must be one of: mongo, postgres, hybrid")
	}

	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	if c.ServerPort == "" {
		return errors.New("SERVER_PORT is required")
	}
	if strings.ToLower(c.Environment) == "production" && strings.TrimSpace(c.CORSOrigins) == "*" {
		return errors.New("CORS_ORIGINS must not be '*' in production; provide an allowlist")
	}
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ----- END OF FILE: backend/MS-AI/pkg/config/config.go -----

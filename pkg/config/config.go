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

// Config contains application configuration loaded from environment variables.
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
	AllowedOriginsSet        map[string]struct{} // Precomputed set of allowed origins
	AllowWebSocketQueryToken bool                // Feature flag for legacy token query param support
	DBTimeout                time.Duration
	// MongoDB auth and pool
	MongoUsername           string
	MongoPassword           string
	MongoAuthSource         string
	MongoMaxPoolSize        uint64
	MongoMinPoolSize        uint64
	MongoEnableMonitoring   bool
	MongoSlowQueryThreshold time.Duration
}

// LoadConfig loads configuration from environment variables. It will load .env file
// only when ENVIRONMENT is not "production".
func LoadConfig() (*Config, error) {
	// Load .env in non-production environments to help local development
	if os.Getenv("ENVIRONMENT") != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		MongoURI:                 getenv("MONGO_URI", ""),
		DBName:                   getenv("DB_NAME", "MSAIDB"),
		ServerPort:               getenv("SERVER_PORT", "8080"),
		JWTSecret:                getenv("JWT_SECRET", ""),
		LogLevel:                 getenv("LOG_LEVEL", "info"),
		Environment:              getenv("ENVIRONMENT", "development"),
		CORSOrigins:              getenv("CORS_ORIGINS", "*"),
		AllowWebSocketQueryToken: strings.ToLower(getenv("WEBSOCKET_ALLOW_QUERY_TOKEN", "false")) == "true",
	}

	// Mongo auth and pools
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

	// Monitoring
	if v := getenv("MONGO_ENABLE_MONITORING", "true"); strings.ToLower(v) == "true" {
		cfg.MongoEnableMonitoring = true
	} else {
		cfg.MongoEnableMonitoring = false
	}

	if v := getenv("MONGO_SLOW_QUERY_THRESHOLD", "1s"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.MongoSlowQueryThreshold = d
		} else {
			return nil, fmt.Errorf("invalid MONGO_SLOW_QUERY_THRESHOLD: %w", err)
		}
	}

	// durations
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

	// Precompute allowed origins set
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

// Validate checks critical configuration values.
func (c *Config) Validate() error {
	if c.MongoURI == "" {
		return errors.New("MONGO_URI is required")
	}
	// If in production, require explicit Mongo credentials (avoid anonymous access)
	if strings.ToLower(c.Environment) == "production" {
		if strings.TrimSpace(c.MongoUsername) == "" || strings.TrimSpace(c.MongoPassword) == "" {
			return errors.New("MONGO_USERNAME and MONGO_PASSWORD are required in production")
		}
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	if c.ServerPort == "" {
		return errors.New("SERVER_PORT is required")
	}
	// Disallow wildcard CORS in production to reduce attack surface.
	if strings.ToLower(c.Environment) == "production" {
		if strings.TrimSpace(c.CORSOrigins) == "*" {
			return errors.New("CORS_ORIGINS must not be '*' in production; provide an allowlist (e.g., https://yourdomain.com,https://www.yourdomain.com)")
		}
	}
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

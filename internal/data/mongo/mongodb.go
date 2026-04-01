package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const (
	defaultTimeout         = 16 * time.Second
	DefaultLimit           = 50
	connectTimeout         = 22 * time.Second
	serverSelectionTimeout = 8 * time.Second
)

var (
	// conservative defaults
	maxRetryWrites int32 = 3
	maxRetryReads  int32 = 3
)

// MongoStats holds MongoDB operation statistics
type MongoStats struct {
	ActiveConnections    int32
	AvailableConnections int32
	TotalPoolSize        uint64
	InUseConnections     int32
	TotalOperations      uint64
	FailedOperations     uint64
	SlowQueries          uint64
	TransactionLatency   time.Duration
}

// MongoStore holds the MongoDB client and helpers.
type MongoStore struct {
	Client  *mongo.Client
	Log     *logger.Logger
	DBName  string
	stats   *MongoStats
	slowOps sync.Map // tracks slow operation timings
	monitor *MongoMonitor
}

// NewMongoStore creates and returns a connected MongoStore.
func NewMongoStore(cfg *config.Config, log *logger.Logger) (*MongoStore, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if log == nil {
		return nil, errors.New("logger is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	log.Info("connecting to mongo", map[string]interface{}{"uri": cfg.MongoURI})

	retryWrites := writeconcern.New(
		writeconcern.WMajority(),
		writeconcern.J(true),
	)

	clientOptions := options.Client().
		ApplyURI(cfg.MongoURI).
		SetWriteConcern(retryWrites).
		SetReadPreference(readpref.Primary()).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetMaxConnIdleTime(5 * time.Minute).
		SetServerSelectionTimeout(serverSelectionTimeout).
		SetConnectTimeout(connectTimeout).
		SetMaxConnecting(2).
		SetHeartbeatInterval(10 * time.Second)

	// Apply pool sizes from config if provided
	if cfg.MongoMaxPoolSize == 0 {
		cfg.MongoMaxPoolSize = 100
	}
	if cfg.MongoMinPoolSize == 0 {
		cfg.MongoMinPoolSize = 10
	}
	clientOptions = clientOptions.SetMaxPoolSize(cfg.MongoMaxPoolSize)
	clientOptions = clientOptions.SetMinPoolSize(cfg.MongoMinPoolSize)

	// Apply credentials when provided
	if strings.TrimSpace(cfg.MongoUsername) != "" {
		cred := options.Credential{
			Username:   cfg.MongoUsername,
			Password:   cfg.MongoPassword,
			AuthSource: cfg.MongoAuthSource,
		}
		clientOptions = clientOptions.SetAuth(cred)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	store := &MongoStore{
		Client: client,
		Log:    log,
		DBName: cfg.DBName,
		stats:  &MongoStats{},
	}

	// Initialize connection monitoring if enabled
	if cfg.MongoEnableMonitoring {
		if err := store.startMonitoring(); err != nil {
			log.Warn("failed to start monitoring, continuing without it",
				map[string]interface{}{
					"error": err.Error(),
				})
		}
	}

	// Ensure indexes with retry and backoff
	backoff := 1 * time.Second
	maxBackoff := 5 * time.Second
	for i := 0; i < 3; i++ {
		if err := store.EnsureIndexes(ctx); err != nil {
			log.Warn("failed to ensure indexes, retrying...",
				map[string]interface{}{
					"attempt": i + 1,
					"error":   err.Error(),
					"backoff": backoff,
				})
			if i == 2 {
				return nil, fmt.Errorf("failed to ensure indexes after multiple retries: %w", err)
			}
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		break
	}

	log.Info("mongodb store initialized successfully",
		map[string]interface{}{
			"database":   cfg.DBName,
			"monitoring": cfg.MongoEnableMonitoring,
			"min_pool":   cfg.MongoMinPoolSize,
			"max_pool":   cfg.MongoMaxPoolSize,
		})

	return store, nil
}

// Close disconnects the MongoDB client and stops monitor (safe to call multiple times).
func (s *MongoStore) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.monitor != nil {
		s.monitor.Stop()
	}
	if s.Client == nil {
		return nil
	}
	if err := s.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("mongo disconnect: %w", err)
	}
	if s.Log != nil {
		s.Log.Info("disconnected from mongo", nil)
	}
	return nil
}

// GetCollection returns a collection handle from the configured database.
func (s *MongoStore) GetCollection(name string) *mongo.Collection {
	if s == nil || s.Client == nil || name == "" {
		return nil
	}
	return s.Client.Database(s.DBName).Collection(name)
}

// ValidOperationType validates operation types
func ValidOperationType(opType string) bool {
	switch strings.ToLower(opType) {
	case "read", "write", "query":
		return true
	default:
		return false
	}
}

// WithCollectionTimeout applies the appropriate timeout for a given collection and operation type (read/write/query).
// If collection or operation type is invalid, defaults are used.
func (s *MongoStore) WithCollectionTimeout(ctx context.Context, collName, opType string) (context.Context, context.CancelFunc) {
	if s == nil {
		// fallback to a background context when store is not available
		return context.WithTimeout(context.Background(), defaultTimeout)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// If incoming context is already canceled or expired, create a fresh one to avoid immediate cancellation
	if ctx.Err() != nil {
		if s != nil && s.Log != nil {
			s.Log.Warn("incoming context already canceled/expired, creating fresh context", map[string]interface{}{"collection": collName, "op": opType})
		}
		ctx = context.Background()
	}

	// Validate inputs
	if collName == "" {
		if s != nil && s.Log != nil {
			s.Log.Warn("empty collection name, using default timeout", nil)
		}
		return context.WithTimeout(ctx, defaultTimeout)
	}

	if !ValidOperationType(opType) {
		if s != nil && s.Log != nil {
			s.Log.Warn("invalid operation type, using default timeout",
				map[string]interface{}{
					"collection": collName,
					"op_type":    opType,
				})
		}
		return context.WithTimeout(ctx, defaultTimeout)
	}

	// Get collection config with fallback to default
	cfg, ok := collectionsConfig[collName]
	if !ok {
		s.Log.Debug("no config for collection, using default",
			map[string]interface{}{"collection": collName})
		cfg = DefaultCollectionConfig()
	}

	// Get operation-specific timeout
	var timeout time.Duration
	switch strings.ToLower(opType) {
	case "read":
		timeout = cfg.ReadTimeout
	case "write":
		timeout = cfg.WriteTimeout
	case "query":
		timeout = cfg.QueryTimeout
	}

	// Ensure minimum timeout
	if timeout < 100*time.Millisecond {
		if s != nil && s.Log != nil {
			s.Log.Warn("timeout too short, using minimum",
				map[string]interface{}{
					"collection": collName,
					"op_type":    opType,
					"timeout":    timeout,
				})
		}
		timeout = 100 * time.Millisecond
	}

	return context.WithTimeout(ctx, timeout)
}

// EnsureIndexes creates necessary indexes for all database collections.
// Uses CreateMany per-collection to reduce roundtrips.
func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	if s == nil {
		return errors.New("mongo store is nil")
	}
	if s.Log == nil {
		// If no logger is present, we still attempt to create indexes but we won't log details.
		// This keeps behavior safe for tests and environments without a logger.
	} else {
		s.Log.Info("ensuring mongo indexes", nil)
	}

	// Get all index configurations
	configs := GetAllIndexConfigs()

	// Create indexes for each collection
	for _, cfg := range configs {
		// Get collection
		coll := s.GetCollection(cfg.Collection)
		if coll == nil {
			if s.Log != nil {
				s.Log.Warn("collection not available, skipping index creation", map[string]interface{}{"collection": cfg.Collection})
			}
			// continue with other collections instead of failing fast
			continue
		}

		// Create indexes (log expected count beforehand)
		expected := len(cfg.Indexes)
		s.Log.Debug("creating indexes for collection", map[string]interface{}{"collection": cfg.Collection, "expected_count": expected})
		if err := createManyIndexes(ctx, coll, cfg.Indexes, cfg.Collection); err != nil {
			// Log the error but continue with other collections
			s.Log.Error(fmt.Sprintf("failed to create indexes for %s", cfg.Collection),
				map[string]interface{}{"error": err.Error(), "collection": cfg.Collection})
			continue
		}

		if s.Log != nil {
			s.Log.Debug(fmt.Sprintf("ensure indexes completed for %s", cfg.Collection), map[string]interface{}{"created_count": expected})
		}
	}

	s.Log.Info("mongo indexes ensured successfully", nil)
	return nil
}

// CollectionConfig holds timeouts for specific collections.
type CollectionConfig struct {
	Name         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	QueryTimeout time.Duration
	MaxDocs      int32
	IndexVersion int32
}

// DefaultCollectionConfig returns default collection timeouts
func DefaultCollectionConfig() CollectionConfig {
	return CollectionConfig{ReadTimeout: defaultTimeout, WriteTimeout: defaultTimeout, QueryTimeout: defaultTimeout}
}

var (
	// Collection-specific configurations
	collectionsConfig = map[string]CollectionConfig{
		"users":          {ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, QueryTimeout: 15 * time.Second},
		"manga":          {ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, QueryTimeout: 15 * time.Second},
		"invite_codes":   {ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, QueryTimeout: 10 * time.Second},
		"manga_chapters": {ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, QueryTimeout: 20 * time.Second},
	}
)

// WithTransaction executes the given function within a transaction using ExecuteTransaction defined in transaction.go
func (s *MongoStore) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error, cfg *TransactionConfig) error {
	return ExecuteTransaction(ctx, s, fn, cfg.ToOptions())
}

// ToOptions converts TransactionConfig to TransactionOptions
func (c *TransactionConfig) ToOptions() *TransactionOptions {
	if c == nil {
		return DefaultTransactionOptions()
	}
	return &TransactionOptions{
		MaxRetries:     c.MaxRetries,
		InitialDelay:   c.InitialDelay,
		MaxDelay:       c.MaxDelay,
		Timeout:        c.Timeout,
		ReadPreference: c.ReadPreference,
		ReadConcern:    c.ReadConcern,
		WriteConcern:   c.WriteConcern,
	}
}

// TransactionConfig holds configuration for a transaction
type TransactionConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	Timeout        time.Duration
	ReadPreference *readpref.ReadPref
	ReadConcern    *readconcern.ReadConcern
	WriteConcern   *writeconcern.WriteConcern
}

// DefaultTransactionConfig returns default transaction settings
func DefaultTransactionConfig() *TransactionConfig {
	return &TransactionConfig{
		MaxRetries:     3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		Timeout:        30 * time.Second,
		ReadPreference: readpref.Primary(),
		ReadConcern:    readconcern.Majority(),
		WriteConcern:   writeconcern.New(writeconcern.WMajority()),
	}
}

// GetStats returns current MongoDB connection statistics
func (s *MongoStore) GetStats() MongoStats {
	if s == nil || s.monitor == nil {
		return MongoStats{}
	}
	metrics := s.monitor.GetMetrics()
	return MongoStats{
		ActiveConnections:    metrics.ActiveConnections,
		AvailableConnections: int32(metrics.PoolSize),
		TotalPoolSize:        uint64(metrics.PoolSize),
		InUseConnections:     metrics.ActiveConnections,
		TotalOperations:      metrics.TotalOperations,
		FailedOperations:     metrics.FailedOperations,
		SlowQueries:          metrics.SlowQueries,
		TransactionLatency:   time.Duration(metrics.LatencyMs) * time.Millisecond,
	}
}

// GetDetailedHealth returns a detailed health snapshot including monitor metrics and basic pool info.
func (s *MongoStore) GetDetailedHealth(ctx context.Context) (map[string]interface{}, error) {
	metrics := map[string]interface{}{}
	if s != nil && s.monitor != nil {
		m := s.monitor.GetMetrics()
		metrics["metrics"] = m
	}

	// Attempt to get serverStatus for additional fields (best-effort)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var ss bson.M
	if s != nil && s.Client != nil {
		if err := s.Client.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&ss); err == nil {
			metrics["serverStatus"] = ss
		} else {
			metrics["serverStatus_error"] = err.Error()
		}
	}

	return metrics, nil
}

// Healthy checks if the MongoDB connection is healthy
func (s *MongoStore) Healthy(ctx context.Context) error {
	if s == nil || s.Client == nil {
		return errors.New("mongo store or client is nil")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.Client.Ping(ctx, readpref.Primary()); err != nil {
		if s.Log != nil {
			s.Log.Error("mongodb health check failed", map[string]interface{}{"error": err.Error()})
		}
		return fmt.Errorf("mongodb ping failed: %w", err)
	}
	return nil
}

// trackSlowOperation records the start time of an operation (safe no-op if monitor nil)
func (s *MongoStore) trackSlowOperation(start time.Time) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(start, nil)
}

// logOperationFailure records a failed operation (safe no-op if monitor nil)
func (s *MongoStore) logOperationFailure(err error) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(time.Now(), err)
}

// recordOperationLatency updates operation latency stats (safe no-op if monitor nil)
func (s *MongoStore) recordOperationLatency(latency time.Duration) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(time.Now().Add(-latency), nil)
}

// startMonitoring initializes and starts the connection monitor.
// Returns an error if monitoring cannot be started.
func (s *MongoStore) startMonitoring() error {
	if s == nil {
		return errors.New("store is nil")
	}
	if s.Client == nil {
		return errors.New("client is nil")
	}
	if s.Log == nil {
		return errors.New("logger is nil")
	}

	// Create monitor with validation
	monitor := NewMongoMonitor(s.Client, s.Log)
	if monitor == nil {
		return errors.New("failed to create monitor")
	}

	// Start monitoring with panic recovery
	defer func() {
		if r := recover(); r != nil {
			s.Log.Error("panic in monitor start",
				map[string]interface{}{"error": fmt.Sprintf("%v", r)})
		}
	}()

	// Start monitor
	monitor.Start()

	// Only set monitor if start succeeded
	s.monitor = monitor
	s.Log.Info("mongodb monitoring started", nil)

	return nil
}

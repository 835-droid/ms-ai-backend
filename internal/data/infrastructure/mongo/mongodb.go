// ----- START OF FILE: backend/MS-AI/internal/data/mongo/mongodb.go -----
package mongo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	connectTimeout         = 15 * time.Second
	serverSelectionTimeout = 12 * time.Second
)

var (
	maxRetryWrites int32 = 3
	maxRetryReads  int32 = 3
)

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

type MongoStore struct {
	Client       *mongo.Client
	Log          *logger.Logger
	DBName       string
	stats        *MongoStats
	slowOps      sync.Map
	monitor      *MongoMonitor
	replicaSet   bool
	replicaSetMu sync.RWMutex
}

func NewMongoStore(cfg *config.Config, log *logger.Logger) (*MongoStore, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if log == nil {
		return nil, errors.New("logger is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	// Do not log credentials; log sanitized host/<cluster> only.
	sanitizedURI := sanitizeMongoURI(cfg.MongoURI)
	log.Info("connecting to mongo", map[string]interface{}{"uri": sanitizedURI})

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

	if cfg.MongoMaxPoolSize == 0 {
		cfg.MongoMaxPoolSize = 100
	}
	if cfg.MongoMinPoolSize == 0 {
		cfg.MongoMinPoolSize = 10
	}
	clientOptions = clientOptions.SetMaxPoolSize(cfg.MongoMaxPoolSize)
	clientOptions = clientOptions.SetMinPoolSize(cfg.MongoMinPoolSize)

	// Determine authentication source with precedence: URI creds unless override
	uriHasCreds := mongoURIHasCredentials(cfg.MongoURI)
	useEnvAuth := false
	if cfg.DBMongoAuthOverride {
		useEnvAuth = strings.TrimSpace(cfg.MongoUsername) != ""
	} else {
		useEnvAuth = !uriHasCreds && strings.TrimSpace(cfg.MongoUsername) != ""
	}

	if useEnvAuth {
		cred := options.Credential{
			Username:   cfg.MongoUsername,
			Password:   cfg.MongoPassword,
			AuthSource: cfg.MongoAuthSource,
		}
		clientOptions = clientOptions.SetAuth(cred)
		log.Info("using env-based MongoDB authentication", nil)
	} else if uriHasCreds {
		log.Info("using URI-based MongoDB authentication", nil)
	} else {
		log.Info("connecting to MongoDB without authentication", nil)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect (URI=%s): %w", sanitizeMongoURI(cfg.MongoURI), err)
	}

	var pingErr error
	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 8*time.Second)
		pingErr = client.Ping(pingCtx, readpref.Primary())
		pingCancel()

		if pingErr == nil {
			break
		}

		if strings.Contains(pingErr.Error(), "server selection timeout") && i < maxRetries-1 {
			log.Warn("MongoDB ping failed, retrying (Atlas cluster might be paused)",
				map[string]interface{}{
					"attempt":     i + 1,
					"max_retries": maxRetries,
					"error":       pingErr.Error(),
				})
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		break
	}

	if pingErr != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping failed after %d retries: %w", maxRetries, pingErr)
	}

	store := &MongoStore{
		Client: client,
		Log:    log,
		DBName: cfg.DBName,
		stats:  &MongoStats{},
	}

	if cfg.MongoEnableMonitoring {
		if err := store.startMonitoring(); err != nil {
			log.Warn("failed to start monitoring, continuing without it",
				map[string]interface{}{"error": err.Error()})
		}
	}

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

	store.replicaSet = store.checkReplicaSet(ctx)

	return store, nil
}

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

func (s *MongoStore) GetCollection(name string) *mongo.Collection {
	if s == nil || s.Client == nil || name == "" {
		return nil
	}
	return s.Client.Database(s.DBName).Collection(name)
}

// IsReplicaSet now uses the cached value
func (s *MongoStore) IsReplicaSet(ctx context.Context) bool {
	if s == nil || s.Client == nil {
		return false
	}
	s.replicaSetMu.RLock()
	defer s.replicaSetMu.RUnlock()
	return s.replicaSet
}

func ValidOperationType(opType string) bool {
	switch strings.ToLower(opType) {
	case "read", "write", "query":
		return true
	default:
		return false
	}
}

func (s *MongoStore) WithCollectionTimeout(ctx context.Context, collName, opType string) (context.Context, context.CancelFunc) {
	if s == nil {
		return context.WithTimeout(context.Background(), defaultTimeout)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		if s.Log != nil {
			s.Log.Warn("incoming context already canceled/expired, creating fresh context", map[string]interface{}{"collection": collName, "op": opType})
		}
		ctx = context.Background()
	}

	if collName == "" {
		if s.Log != nil {
			s.Log.Warn("empty collection name, using default timeout", nil)
		}
		return context.WithTimeout(ctx, defaultTimeout)
	}

	if !ValidOperationType(opType) {
		if s.Log != nil {
			s.Log.Warn("invalid operation type, using default timeout", map[string]interface{}{"collection": collName, "op": opType})
		}
		return context.WithTimeout(ctx, defaultTimeout)
	}

	cfg, ok := collectionsConfig[collName]
	if !ok {
		if s.Log != nil {
			s.Log.Debug("no config for collection, using default", map[string]interface{}{"collection": collName})
		}
		cfg = DefaultCollectionConfig()
	}

	var timeout time.Duration
	switch strings.ToLower(opType) {
	case "read":
		timeout = cfg.ReadTimeout
	case "write":
		timeout = cfg.WriteTimeout
	case "query":
		timeout = cfg.QueryTimeout
	}

	if timeout < 100*time.Millisecond {
		if s.Log != nil {
			s.Log.Warn("timeout too short, using minimum", map[string]interface{}{"collection": collName, "op_type": opType, "timeout": timeout})
		}
		timeout = 100 * time.Millisecond
	}

	return context.WithTimeout(ctx, timeout)
}

func sanitizeMongoURI(uri string) string {
	if uri == "" {
		return ""
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return "<invalid-mongo-uri>"
	}
	if parsed.User != nil {
		parsed.User = nil
	}
	return parsed.String()
}

func mongoURIHasCredentials(uri string) bool {
	if strings.TrimSpace(uri) == "" {
		return false
	}
	parsed, err := url.Parse(uri)
	if err != nil || parsed.User == nil {
		return false
	}
	username := strings.TrimSpace(parsed.User.Username())
	password, hasPassword := parsed.User.Password()
	if username == "" || !hasPassword || strings.TrimSpace(password) == "" {
		return false
	}
	return true
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	if s == nil {
		return errors.New("mongo store is nil")
	}
	if s.Log != nil {
		s.Log.Info("ensuring mongo indexes", nil)
	}

	configs := GetAllIndexConfigs()
	for _, cfg := range configs {
		coll := s.GetCollection(cfg.Collection)
		if coll == nil {
			if s.Log != nil {
				s.Log.Warn("collection not available, skipping index creation", map[string]interface{}{"collection": cfg.Collection})
			}
			continue
		}

		expected := len(cfg.Indexes)
		s.Log.Debug("creating indexes for collection", map[string]interface{}{"collection": cfg.Collection, "expected_count": expected})
		if err := createManyIndexes(ctx, coll, cfg.Indexes, cfg.Collection); err != nil {
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

type CollectionConfig struct {
	Name         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	QueryTimeout time.Duration
	MaxDocs      int32
	IndexVersion int32
}

func DefaultCollectionConfig() CollectionConfig {
	return CollectionConfig{ReadTimeout: defaultTimeout, WriteTimeout: defaultTimeout, QueryTimeout: defaultTimeout}
}

var (
	collectionsConfig = map[string]CollectionConfig{
		"users":          {ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, QueryTimeout: 15 * time.Second},
		"manga":          {ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, QueryTimeout: 15 * time.Second},
		"invite_codes":   {ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, QueryTimeout: 10 * time.Second},
		"manga_chapters": {ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, QueryTimeout: 20 * time.Second},
	}
)

func (s *MongoStore) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error, cfg *TransactionConfig) error {
	if !s.IsReplicaSet(ctx) {
		session, err := s.Client.StartSession()
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}
		defer session.EndSession(ctx)
		sessionCtx := mongo.NewSessionContext(ctx, session)
		return fn(sessionCtx)
	}
	return ExecuteTransaction(ctx, s, fn, cfg.ToOptions())
}

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

type TransactionConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	Timeout        time.Duration
	ReadPreference *readpref.ReadPref
	ReadConcern    *readconcern.ReadConcern
	WriteConcern   *writeconcern.WriteConcern
}

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

func (s *MongoStore) GetDetailedHealth(ctx context.Context) (map[string]interface{}, error) {
	metrics := map[string]interface{}{}
	if s != nil && s.monitor != nil {
		m := s.monitor.GetMetrics()
		metrics["metrics"] = m
	}
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

func (s *MongoStore) trackSlowOperation(start time.Time) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(start, nil)
}

func (s *MongoStore) logOperationFailure(err error) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(time.Now(), err)
}

func (s *MongoStore) recordOperationLatency(latency time.Duration) {
	if s == nil || s.monitor == nil {
		return
	}
	s.monitor.RecordOperation(time.Now().Add(-latency), nil)
}

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
	monitor := NewMongoMonitor(s.Client, s.Log)
	if monitor == nil {
		return errors.New("failed to create monitor")
	}
	defer func() {
		if r := recover(); r != nil {
			s.Log.Error("panic in monitor start",
				map[string]interface{}{"error": fmt.Sprintf("%v", r)})
		}
	}()
	monitor.Start()
	s.monitor = monitor
	s.Log.Info("mongodb monitoring started", nil)
	return nil
}

func (s *MongoStore) checkReplicaSet(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var result bson.M
	err := s.Client.Database("admin").RunCommand(ctx, bson.M{"isMaster": 1}).Decode(&result)
	if err != nil {
		s.Log.Warn("failed to check replica set status", map[string]interface{}{"error": err.Error()})
		return false
	}
	if setName, ok := result["setName"]; ok && setName != nil {
		return true
	}
	return false
}

// ----- END OF FILE: backend/MS-AI/internal/data/mongo/mongodb.go -----

package postgres

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"

	"github.com/jmoiron/sqlx"
)

// PostgresMetrics holds PostgreSQL performance metrics
type PostgresMetrics struct {
	TotalQueries     uint64
	FailedQueries    uint64
	SlowQueries      uint64
	OpenConnections  int32
	InUseConnections int32
	IdleConnections  int32
	WaitCount        int64
	WaitDuration     int64
	AverageLatencyMs int64
	MaxLatencyMs     int64
	SuccessRate      float64
}

// PostgresMonitor monitors PostgreSQL database performance
type PostgresMonitor struct {
	db      *sqlx.DB
	log     *logger.Logger
	metrics *PostgresMetrics
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	once    sync.Once
}

// NewPostgresMonitor creates a new PostgreSQL monitor
func NewPostgresMonitor(db *sqlx.DB, log *logger.Logger) *PostgresMonitor {
	return &PostgresMonitor{
		db:      db,
		log:     log,
		metrics: &PostgresMetrics{},
		done:    make(chan struct{}),
	}
}

// Start begins monitoring PostgreSQL metrics
func (m *PostgresMonitor) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				m.collectStats()
			}
		}
	}()
}

// Stop stops the monitoring
func (m *PostgresMonitor) Stop() {
	m.once.Do(func() {
		close(m.done)
	})
	m.wg.Wait()
}

// collectStats collects current database statistics
func (m *PostgresMonitor) collectStats() {
	stats := m.db.Stats()

	atomic.StoreInt32(&m.metrics.OpenConnections, int32(stats.OpenConnections))
	atomic.StoreInt32(&m.metrics.InUseConnections, int32(stats.InUse))
	atomic.StoreInt32(&m.metrics.IdleConnections, int32(stats.Idle))
	atomic.StoreInt64(&m.metrics.WaitCount, stats.WaitCount)
	atomic.StoreInt64(&m.metrics.WaitDuration, stats.WaitDuration.Nanoseconds())

	total := atomic.LoadUint64(&m.metrics.TotalQueries)
	failed := atomic.LoadUint64(&m.metrics.FailedQueries)

	var successRate float64
	if total > 0 {
		successRate = float64(total-failed) / float64(total) * 100
	}

	m.mu.Lock()
	m.metrics.SuccessRate = successRate
	m.mu.Unlock()

	m.log.Info("PostgreSQL Stats", map[string]interface{}{
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
		"wait_count":       stats.WaitCount,
		"wait_duration_ms": stats.WaitDuration.Milliseconds(),
		"success_rate":     successRate,
	})
}

// RecordQuery records a query execution
func (m *PostgresMonitor) RecordQuery(start time.Time, err error) {
	duration := time.Since(start)
	latencyMs := duration.Milliseconds()

	atomic.AddUint64(&m.metrics.TotalQueries, 1)

	if err != nil {
		atomic.AddUint64(&m.metrics.FailedQueries, 1)
	}

	if duration > time.Second {
		atomic.AddUint64(&m.metrics.SlowQueries, 1)
	}

	// Update average latency (simple moving average)
	currentAvg := atomic.LoadInt64(&m.metrics.AverageLatencyMs)
	newAvg := (currentAvg + latencyMs) / 2
	atomic.StoreInt64(&m.metrics.AverageLatencyMs, newAvg)

	// Update max latency
	for {
		currentMax := atomic.LoadInt64(&m.metrics.MaxLatencyMs)
		if latencyMs <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&m.metrics.MaxLatencyMs, currentMax, latencyMs) {
			break
		}
	}
}

// GetMetrics returns a snapshot of current metrics
func (m *PostgresMonitor) GetMetrics() PostgresMetrics {
	return PostgresMetrics{
		TotalQueries:     atomic.LoadUint64(&m.metrics.TotalQueries),
		FailedQueries:    atomic.LoadUint64(&m.metrics.FailedQueries),
		SlowQueries:      atomic.LoadUint64(&m.metrics.SlowQueries),
		OpenConnections:  atomic.LoadInt32(&m.metrics.OpenConnections),
		InUseConnections: atomic.LoadInt32(&m.metrics.InUseConnections),
		IdleConnections:  atomic.LoadInt32(&m.metrics.IdleConnections),
		WaitCount:        atomic.LoadInt64(&m.metrics.WaitCount),
		WaitDuration:     atomic.LoadInt64(&m.metrics.WaitDuration),
		AverageLatencyMs: atomic.LoadInt64(&m.metrics.AverageLatencyMs),
		MaxLatencyMs:     atomic.LoadInt64(&m.metrics.MaxLatencyMs),
		SuccessRate:      m.getSuccessRate(),
	}
}

func (m *PostgresMonitor) getSuccessRate() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics.SuccessRate
}

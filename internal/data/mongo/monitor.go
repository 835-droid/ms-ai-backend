// ----- START OF FILE: backend/MS-AI/internal/data/mongo/monitor.go -----
package mongo

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MonitoringMetrics holds MongoDB operation metrics
type MonitoringMetrics struct {
	TotalOperations   uint64
	FailedOperations  uint64
	SlowQueries       uint64
	ActiveConnections int32
	PoolSize          int32
	LatencyMs         int64
	AverageLatencyMs  int64
	MaxLatencyMs      int64
	SuccessRate       float64
}

// MongoMonitor tracks MongoDB operation metrics
type MongoMonitor struct {
	client  *mongo.Client
	log     *logger.Logger
	metrics *MonitoringMetrics
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	once    sync.Once
}

// NewMongoMonitor creates a new MongoDB monitor
func NewMongoMonitor(client *mongo.Client, log *logger.Logger) *MongoMonitor {
	return &MongoMonitor{
		client:  client,
		log:     log,
		metrics: &MonitoringMetrics{},
		done:    make(chan struct{}),
	}
}

// Start begins monitoring MongoDB metrics
func (m *MongoMonitor) Start() {
	m.wg.Add(2)

	// Monitor connection stats
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if m.log != nil {
					m.log.Error("mongo monitor connection goroutine panic", map[string]interface{}{"panic": r})
				}
			}
			m.wg.Done()
		}()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.collectConnectionStats()
			case <-m.done:
				return
			}
		}
	}()

	// Monitor operation stats
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if m.log != nil {
					m.log.Error("mongo monitor operation goroutine panic", map[string]interface{}{"panic": r})
				}
			}
			m.wg.Done()
		}()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.collectOperationStats()
			case <-m.done:
				return
			}
		}
	}()
}

// Stop terminates monitoring
func (m *MongoMonitor) Stop() {
	// ensure stop is only executed once
	m.once.Do(func() {
		if m.done != nil {
			close(m.done)
		}
		m.wg.Wait()
	})
}

// RecordOperation records metrics for a single operation
func (m *MongoMonitor) RecordOperation(start time.Time, err error) {
	if m == nil || m.metrics == nil {
		return
	}

	atomic.AddUint64(&m.metrics.TotalOperations, 1)

	duration := time.Since(start)
	latency := duration.Milliseconds()
	atomic.StoreInt64(&m.metrics.LatencyMs, latency)

	// Update rolling average (simple exponential moving average)
	alpha := 0.2
	for {
		oldAvg := atomic.LoadInt64(&m.metrics.AverageLatencyMs)
		var newAvg int64
		if oldAvg == 0 {
			newAvg = latency
		} else {
			newAvg = int64(float64(oldAvg)*(1-alpha) + float64(latency)*alpha)
		}
		if atomic.CompareAndSwapInt64(&m.metrics.AverageLatencyMs, oldAvg, newAvg) {
			break
		}
	}

	// Update max latency
	for {
		oldMax := atomic.LoadInt64(&m.metrics.MaxLatencyMs)
		if latency <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&m.metrics.MaxLatencyMs, oldMax, latency) {
			break
		}
	}

	// Update success rate
	total := atomic.LoadUint64(&m.metrics.TotalOperations)
	failed := atomic.LoadUint64(&m.metrics.FailedOperations)
	var successRate float64
	if total > 0 {
		successRate = float64(total-failed) / float64(total)
	}
	m.mu.Lock()
	if m.metrics != nil {
		m.metrics.SuccessRate = successRate
	}
	m.mu.Unlock()

	if err != nil {
		atomic.AddUint64(&m.metrics.FailedOperations, 1)
	}

	if duration > time.Second {
		atomic.AddUint64(&m.metrics.SlowQueries, 1)
		if m.log != nil {
			m.log.Warn("slow mongodb operation", map[string]interface{}{"duration_ms": duration.Milliseconds()})
		}
	}
}

// collectConnectionStats gathers connection metrics
func (m *MongoMonitor) collectConnectionStats() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var serverStatus bson.M
	err := m.client.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&serverStatus)
	if err != nil {
		if m.log != nil {
			m.log.Error("failed to collect connection stats", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	if connRaw, ok := serverStatus["connections"]; ok {
		if conn, ok := connRaw.(bson.M); ok {
			// fields may be numeric types (int32, int64, float64)
			if currentRaw, ok := conn["current"]; ok {
				switch v := currentRaw.(type) {
				case int32:
					atomic.StoreInt32(&m.metrics.ActiveConnections, v)
				case int64:
					atomic.StoreInt32(&m.metrics.ActiveConnections, int32(v))
				case float64:
					atomic.StoreInt32(&m.metrics.ActiveConnections, int32(v))
				}
			}
			if availableRaw, ok := conn["available"]; ok {
				switch v := availableRaw.(type) {
				case int32:
					atomic.StoreInt32(&m.metrics.PoolSize, v)
				case int64:
					atomic.StoreInt32(&m.metrics.PoolSize, int32(v))
				case float64:
					atomic.StoreInt32(&m.metrics.PoolSize, int32(v))
				}
			}
		}
	}
}

// collectOperationStats gathers operation metrics
func (m *MongoMonitor) collectOperationStats() {
	metrics := m.GetMetrics()
	if m.log != nil {
		m.log.Info("mongodb metrics", map[string]interface{}{
			"total_operations":   metrics.TotalOperations,
			"failed_operations":  metrics.FailedOperations,
			"slow_queries":       metrics.SlowQueries,
			"active_connections": metrics.ActiveConnections,
			"pool_size":          metrics.PoolSize,
			"latency_ms":         metrics.LatencyMs,
		})
	}
}

// GetMetrics returns current monitoring metrics
func (m *MongoMonitor) GetMetrics() MonitoringMetrics {
	if m == nil || m.metrics == nil {
		return MonitoringMetrics{}
	}
	return MonitoringMetrics{
		TotalOperations:   atomic.LoadUint64(&m.metrics.TotalOperations),
		FailedOperations:  atomic.LoadUint64(&m.metrics.FailedOperations),
		SlowQueries:       atomic.LoadUint64(&m.metrics.SlowQueries),
		ActiveConnections: atomic.LoadInt32(&m.metrics.ActiveConnections),
		PoolSize:          atomic.LoadInt32(&m.metrics.PoolSize),
		LatencyMs:         atomic.LoadInt64(&m.metrics.LatencyMs),
	}
}

// ----- END OF FILE: backend/MS-AI/internal/data/mongo/monitor.go -----

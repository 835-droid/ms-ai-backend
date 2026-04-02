// ----- START OF FILE: backend/MS-AI/internal/data/mongo/transaction.go -----
package mongo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// TransactionOptions defines options for MongoDB transactions
type TransactionOptions struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	Timeout        time.Duration
	ReadPreference *readpref.ReadPref
	ReadConcern    *readconcern.ReadConcern
	WriteConcern   *writeconcern.WriteConcern
}

// DefaultTransactionOptions returns default transaction settings
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		MaxRetries:     3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		Timeout:        30 * time.Second,
		ReadPreference: readpref.Primary(),
		ReadConcern:    readconcern.Majority(),
		WriteConcern:   writeconcern.New(writeconcern.WMajority()),
	}
}

// ExecuteTransaction runs a function within a transaction with retry logic
func ExecuteTransaction(ctx context.Context, store *MongoStore, fn func(sessCtx mongo.SessionContext) error, opts *TransactionOptions) error {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	// Validate store parameter
	if store == nil {
		return fmt.Errorf("store is required for ExecuteTransaction")
	}

	start := time.Now()
	var lastErr error

	// seed jitter source
	rand.Seed(time.Now().UnixNano())

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		// respect context cancellation early
		if ctx.Err() != nil {
			return ctx.Err()
		}
		session, err := store.Client.StartSession()
		if err != nil {
			if store.Log != nil {
				store.Log.Error("failed to start session", map[string]interface{}{"error": err.Error()})
			}
			return fmt.Errorf("start session: %w", err)
		}
		sessionCtx, cancel := context.WithTimeout(ctx, opts.Timeout)

		// Configure transaction options
		txnOpts := options.Transaction().
			SetReadPreference(opts.ReadPreference).
			SetReadConcern(opts.ReadConcern).
			SetWriteConcern(opts.WriteConcern)

		// Execute transaction
		_, err = session.WithTransaction(sessionCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			return nil, fn(sessCtx)
		}, txnOpts)

		session.EndSession(sessionCtx)
		cancel()

		if err == nil {
			// record successful operation safely
			store.trackSlowOperation(start)
			if store.Log != nil {
				store.Log.Info("transaction succeeded", map[string]interface{}{"attempt": attempt, "elapsed_ms": time.Since(start).Milliseconds()})
			}
			return nil
		}

		lastErr = err
		if store.Log != nil {
			store.Log.Warn("transaction failed, will retry if eligible", map[string]interface{}{
				"attempt":    attempt,
				"error":      err.Error(),
				"retryable":  IsRetryableError(err),
				"elapsed_ms": time.Since(start).Milliseconds(),
			})
		}

		// Break if error is not retryable
		if !IsRetryableError(err) {
			break
		}

		// Exponential backoff with jitter
		// base * 2^(attempt-1)
		multiplier := 1 << uint(attempt-1)
		delay := time.Duration(multiplier) * opts.InitialDelay
		if delay > opts.MaxDelay {
			delay = opts.MaxDelay
		}
		// jitter up to 30% of delay
		jitter := time.Duration(rand.Int63n(int64(delay) / 3))
		time.Sleep(delay + jitter)
	}

	if store != nil && store.Log != nil {
		if lastErr != nil {
			store.Log.Error("transaction failed after retries", map[string]interface{}{
				"max_retries": opts.MaxRetries,
				"elapsed_ms":  time.Since(start).Milliseconds(),
				"error":       lastErr.Error(),
			})
		} else {
			store.Log.Error("transaction failed after retries", map[string]interface{}{"max_retries": opts.MaxRetries, "elapsed_ms": time.Since(start).Milliseconds()})
		}
	}
	if store != nil {
		store.logOperationFailure(lastErr)
	}

	return lastErr
}

// IsRetryableError checks if the error is a transient error that warrants a retry.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Fall back to message checks for common retryable markers
	msg := err.Error()
	return strings.Contains(msg, "TransientTransactionError") ||
		strings.Contains(msg, "InterruptedByTimeout") ||
		strings.Contains(msg, "ShutdownInProgress") ||
		strings.Contains(msg, "NotMaster") ||
		strings.Contains(msg, "NodeIsRecovering")
}

// Contains is a simple helper function (يجب أن يكون موجوداً في مكان ما داخل الحزمة)
// NOTE: helper functions removed as they were unused; expand IsRetryableError if driver-specific checks are needed.
// ----- END OF FILE: backend/MS-AI/internal/data/mongo/transaction.go -----

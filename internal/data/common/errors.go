// ----- START OF FILE: backend/MS-AI/internal/data/common/errors.go -----
package data

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

// IsRetryableError checks if the error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a MongoDB error
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		// Error codes that indicate retryable writes
		retryableCodes := map[int32]bool{
			11600: true, // InterruptedAtShutdown
			11602: true, // InterruptedDueToReplStateChange
			10107: true, // NotWritablePrimary
			13435: true, // NotPrimaryNoSecondaryOk
			13436: true, // NotPrimaryOrSecondary
			189:   true, // PrimarySteppedDown
			91:    true, // ShutdownInProgress
			7:     true, // HostNotFound
			6:     true, // HostUnreachable
			89:    true, // NetworkTimeout
			9001:  true, // SocketException
			50:    true, // ExceededTimeLimit
		}
		return retryableCodes[cmdErr.Code]
	}

	// Check for driver timeout helper
	if mongo.IsTimeout(err) {
		return true
	}

	// Check error message for common retryable cases
	msg := strings.ToLower(err.Error())
	retryableStrings := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"no servers available",
		"server selection timeout",
		"connection timed out",
		"broken pipe",
		"context deadline exceeded",
	}

	for _, s := range retryableStrings {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}

// IsDuplicateKeyError returns true if the error is a Mongo duplicate key error
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}

// IsTimeoutError returns true when an error represents a timeout
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if mongo.IsTimeout(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded")
}

// IsNetworkError returns true when an error looks like a transient network error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	networkSigns := []string{"connection refused", "connection reset", "no servers available", "server selection timeout", "connection closed", "broken pipe"}
	for _, s := range networkSigns {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// ----- END OF FILE: backend/MS-AI/internal/data/common/errors.go -----

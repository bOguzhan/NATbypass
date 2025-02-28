// internal/utils/retry.go
package utils

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig defines parameters for retry operations
type RetryConfig struct {
	MaxAttempts       int           // Maximum number of attempts
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffFactor     float64       // Multiplier for backoff after each attempt
	TimeoutPerAttempt time.Duration // Timeout for each attempt
}

// DefaultRetryConfig provides sensible defaults for retry operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        2 * time.Second,
		BackoffFactor:     2.0,
		TimeoutPerAttempt: 5 * time.Second,
	}
}

// RetryWithBackoff executes the provided function with exponential backoff until successful or max attempts reached
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func(ctx context.Context) error) error {
	var lastError error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// If we're not on the first attempt, apply backoff
		if attempt > 0 {
			// Calculate backoff duration with exponential increase
			backoffDuration := time.Duration(float64(config.InitialBackoff) * math.Pow(config.BackoffFactor, float64(attempt-1)))

			// Cap to max backoff if needed
			if backoffDuration > config.MaxBackoff {
				backoffDuration = config.MaxBackoff
			}

			// Create a timer for backoff
			select {
			case <-time.After(backoffDuration):
				// Backoff period completed, continue with retry
			case <-ctx.Done():
				// Context canceled during backoff
				return fmt.Errorf("operation canceled during backoff: %w", ctx.Err())
			}
		}

		// Create a context with timeout for this attempt
		attemptCtx := ctx
		var cancel context.CancelFunc

		if config.TimeoutPerAttempt > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, config.TimeoutPerAttempt)
		}

		// Try the operation
		err := operation(attemptCtx)

		// Clean up the timeout if we created one
		if cancel != nil {
			cancel()
		}

		// If successful, return nil
		if err == nil {
			return nil
		}

		// Update the last error
		lastError = err

		// If context was canceled, stop retrying
		if ctx.Err() != nil {
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		}
	}

	// Return the last error after all attempts failed
	return fmt.Errorf("all %d attempts failed, last error: %w", config.MaxAttempts, lastError)
}

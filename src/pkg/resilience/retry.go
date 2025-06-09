package resilience

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// RetryableError represents an error that can be retried
type RetryableError interface {
	error
	IsRetryable() bool
}

// retryableError implements RetryableError
type retryableError struct {
	err        error
	retryable  bool
}

func (r retryableError) Error() string {
	return r.err.Error()
}

func (r retryableError) IsRetryable() bool {
	return r.retryable
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error, retryable bool) RetryableError {
	return retryableError{err: err, retryable: retryable}
}

// WithRetry executes a function with retry logic
func WithRetry(config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateDelay(config, attempt-1)
			time.Sleep(delay)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.IsRetryable() {
			return err
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// WithRetryContext executes a function with retry logic and context cancellation
func WithRetryContext(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt > 0 {
			delay := calculateDelay(config, attempt-1)
			
			// Sleep with context cancellation
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.IsRetryable() {
			return err
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateDelay calculates the delay for the given attempt
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
	
	// Apply max delay limit
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Apply jitter if enabled
	if config.Jitter {
		jitter := delay * 0.1 * (2.0*math.Abs(0.5) - 1.0) // Simple jitter
		delay += jitter
		if delay < 0 {
			delay = float64(config.BaseDelay)
		}
	}

	return time.Duration(delay)
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond * 100,
		MaxDelay:   time.Second * 30,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// ExponentialBackoffConfig returns an exponential backoff retry configuration
func ExponentialBackoffConfig(maxRetries int) RetryConfig {
	return RetryConfig{
		MaxRetries: maxRetries,
		BaseDelay:  time.Millisecond * 200,
		MaxDelay:   time.Minute * 5,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// LinearBackoffConfig returns a linear backoff retry configuration
func LinearBackoffConfig(maxRetries int, delay time.Duration) RetryConfig {
	return RetryConfig{
		MaxRetries: maxRetries,
		BaseDelay:  delay,
		MaxDelay:   delay * time.Duration(maxRetries),
		Multiplier: 1.0,
		Jitter:     false,
	}
}

// RetryMetrics holds metrics for retry operations
type RetryMetrics struct {
	TotalAttempts     int
	SuccessfulRetries int
	FailedRetries     int
	TotalDelay        time.Duration
}

// RetryWithMetrics executes a function with retry logic and collects metrics
func RetryWithMetrics(config RetryConfig, fn RetryableFunc) (error, RetryMetrics) {
	metrics := RetryMetrics{}
	var lastErr error
	startTime := time.Now()

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		metrics.TotalAttempts++

		if attempt > 0 {
			delay := calculateDelay(config, attempt-1)
			time.Sleep(delay)
			metrics.TotalDelay += delay
		}

		err := fn()
		if err == nil {
			if attempt > 0 {
				metrics.SuccessfulRetries++
			}
			metrics.TotalDelay = time.Since(startTime)
			return nil, metrics
		}

		lastErr = err

		// Check if error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.IsRetryable() {
			metrics.FailedRetries++
			metrics.TotalDelay = time.Since(startTime)
			return err, metrics
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}
	}

	metrics.FailedRetries++
	metrics.TotalDelay = time.Since(startTime)
	return fmt.Errorf("max retries exceeded: %w", lastErr), metrics
}
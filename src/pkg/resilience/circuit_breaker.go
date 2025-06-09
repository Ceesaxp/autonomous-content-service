package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// Closed allows requests to pass through
	Closed CircuitBreakerState = iota
	// Open rejects all requests
	Open
	// HalfOpen allows limited requests to test if service has recovered
	HalfOpen
)

// String returns string representation of circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

var (
	// ErrCircuitOpen is returned when circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTimeout is returned when operation times out
	ErrTimeout = errors.New("operation timeout")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name            string
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenTimeout time.Duration
	maxRequests     int // max requests allowed in half-open state

	mutex        sync.RWMutex
	state        CircuitBreakerState
	failures     int
	lastFailTime time.Time
	requests     int // current requests in half-open state
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name            string
	MaxFailures     int
	ResetTimeout    time.Duration
	HalfOpenTimeout time.Duration
	MaxRequests     int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:            config.Name,
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenTimeout: config.HalfOpenTimeout,
		maxRequests:     config.MaxRequests,
		state:           Closed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// ExecuteWithTimeout executes a function with circuit breaker and timeout protection
func (cb *CircuitBreaker) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Channel to receive result
	resultChan := make(chan error, 1)

	// Execute function in goroutine
	go func() {
		resultChan <- fn()
	}()

	// Wait for result or timeout
	select {
	case err := <-resultChan:
		cb.recordResult(err)
		return err
	case <-timeoutCtx.Done():
		cb.recordResult(ErrTimeout)
		return ErrTimeout
	}
}

// allowRequest checks if request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		// Check if reset timeout has passed
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.setState(HalfOpen)
			cb.requests = 0
			return true
		}
		return false
	case HalfOpen:
		// Allow limited requests in half-open state
		if cb.requests < cb.maxRequests {
			cb.requests++
			return true
		}
		return false
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case Closed:
		if cb.failures >= cb.maxFailures {
			cb.setState(Open)
		}
	case HalfOpen:
		cb.setState(Open)
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case HalfOpen:
		// If enough successful requests in half-open, move to closed
		if cb.requests >= cb.maxRequests {
			cb.setState(Closed)
			cb.failures = 0
		}
	case Closed:
		// Reset failure count on success
		cb.failures = 0
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
	cb.state = state
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetName returns the name of the circuit breaker
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failures
}

// GetMetrics returns current metrics for the circuit breaker
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return CircuitBreakerMetrics{
		Name:         cb.name,
		State:        cb.state,
		Failures:     cb.failures,
		LastFailTime: cb.lastFailTime,
		Requests:     cb.requests,
	}
}

// CircuitBreakerMetrics holds metrics for a circuit breaker
type CircuitBreakerMetrics struct {
	Name         string
	State        CircuitBreakerState
	Failures     int
	LastFailTime time.Time
	Requests     int
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	config.Name = name
	cb := NewCircuitBreaker(config)
	cbm.breakers[name] = cb
	return cb
}

// Get gets an existing circuit breaker
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// GetAllMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllMetrics() []CircuitBreakerMetrics {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	metrics := make([]CircuitBreakerMetrics, 0, len(cbm.breakers))
	for _, cb := range cbm.breakers {
		metrics = append(metrics, cb.GetMetrics())
	}

	return metrics
}

// DefaultCircuitBreakerConfig returns a default configuration
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:            name,
		MaxFailures:     5,
		ResetTimeout:    time.Minute * 1,
		HalfOpenTimeout: time.Second * 30,
		MaxRequests:     3,
	}
}

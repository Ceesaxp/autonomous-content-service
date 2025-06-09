package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all service metrics
type Metrics struct {
	// HTTP metrics
	RequestDuration *prometheus.HistogramVec
	RequestsTotal   *prometheus.CounterVec
	RequestsActive  prometheus.Gauge

	// Service metrics
	ServiceHealth   *prometheus.GaugeVec
	ServiceUptime   prometheus.Gauge
	ServiceVersion  *prometheus.GaugeVec

	// Business metrics
	EventsProcessed    *prometheus.CounterVec
	BusinessOperations *prometheus.HistogramVec
	ErrorsTotal        *prometheus.CounterVec

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec
	CircuitBreakerTrips *prometheus.CounterVec

	// Custom metrics
	CustomGauges     map[string]prometheus.Gauge
	CustomCounters   map[string]prometheus.Counter
	CustomHistograms map[string]prometheus.Histogram
}

// NewMetrics creates a new metrics instance
func NewMetrics(serviceName string) *Metrics {
	metrics := &Metrics{
		// HTTP metrics
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Subsystem: serviceName,
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
				Subsystem: serviceName,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		RequestsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "http_requests_active",
				Help:      "Number of active HTTP requests",
				Subsystem: serviceName,
			},
		),

		// Service metrics
		ServiceHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "service_health",
				Help:      "Service health status (1 = healthy, 0 = unhealthy)",
				Subsystem: serviceName,
			},
			[]string{"component"},
		),
		ServiceUptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "service_uptime_seconds",
				Help:      "Service uptime in seconds",
				Subsystem: serviceName,
			},
		),
		ServiceVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "service_version_info",
				Help:      "Service version information",
				Subsystem: serviceName,
			},
			[]string{"version", "commit", "build_time"},
		),

		// Business metrics
		EventsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "events_processed_total",
				Help:      "Total number of events processed",
				Subsystem: serviceName,
			},
			[]string{"event_type", "status"},
		),
		BusinessOperations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:      "business_operation_duration_seconds",
				Help:      "Business operation duration in seconds",
				Subsystem: serviceName,
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "status"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "errors_total",
				Help:      "Total number of errors",
				Subsystem: serviceName,
			},
			[]string{"type", "severity"},
		),

		// Circuit breaker metrics
		CircuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0 = closed, 1 = open, 2 = half-open)",
				Subsystem: serviceName,
			},
			[]string{"circuit_name"},
		),
		CircuitBreakerTrips: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "circuit_breaker_trips_total",
				Help:      "Total number of circuit breaker trips",
				Subsystem: serviceName,
			},
			[]string{"circuit_name"},
		),

		// Custom metrics
		CustomGauges:     make(map[string]prometheus.Gauge),
		CustomCounters:   make(map[string]prometheus.Counter),
		CustomHistograms: make(map[string]prometheus.Histogram),
	}

	// Register all metrics
	prometheus.MustRegister(
		metrics.RequestDuration,
		metrics.RequestsTotal,
		metrics.RequestsActive,
		metrics.ServiceHealth,
		metrics.ServiceUptime,
		metrics.ServiceVersion,
		metrics.EventsProcessed,
		metrics.BusinessOperations,
		metrics.ErrorsTotal,
		metrics.CircuitBreakerState,
		metrics.CircuitBreakerTrips,
	)

	return metrics
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
	m.RequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// IncActiveRequests increments active request count
func (m *Metrics) IncActiveRequests() {
	m.RequestsActive.Inc()
}

// DecActiveRequests decrements active request count
func (m *Metrics) DecActiveRequests() {
	m.RequestsActive.Dec()
}

// SetServiceHealth sets service health status
func (m *Metrics) SetServiceHealth(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.ServiceHealth.WithLabelValues(component).Set(value)
}

// SetServiceUptime sets service uptime
func (m *Metrics) SetServiceUptime(uptime time.Duration) {
	m.ServiceUptime.Set(uptime.Seconds())
}

// SetServiceVersion sets service version information
func (m *Metrics) SetServiceVersion(version, commit, buildTime string) {
	m.ServiceVersion.WithLabelValues(version, commit, buildTime).Set(1)
}

// RecordEvent records an event processing metric
func (m *Metrics) RecordEvent(eventType, status string) {
	m.EventsProcessed.WithLabelValues(eventType, status).Inc()
}

// RecordBusinessOperation records a business operation metric
func (m *Metrics) RecordBusinessOperation(operation, status string, duration time.Duration) {
	m.BusinessOperations.WithLabelValues(operation, status).Observe(duration.Seconds())
}

// RecordError records an error metric
func (m *Metrics) RecordError(errorType, severity string) {
	m.ErrorsTotal.WithLabelValues(errorType, severity).Inc()
}

// SetCircuitBreakerState sets circuit breaker state
func (m *Metrics) SetCircuitBreakerState(circuitName string, state int) {
	m.CircuitBreakerState.WithLabelValues(circuitName).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip(circuitName string) {
	m.CircuitBreakerTrips.WithLabelValues(circuitName).Inc()
}

// AddCustomGauge adds a custom gauge metric
func (m *Metrics) AddCustomGauge(name, help string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(gauge)
	m.CustomGauges[name] = gauge
	return gauge
}

// AddCustomCounter adds a custom counter metric
func (m *Metrics) AddCustomCounter(name, help string) prometheus.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(counter)
	m.CustomCounters[name] = counter
	return counter
}

// AddCustomHistogram adds a custom histogram metric
func (m *Metrics) AddCustomHistogram(name, help string, buckets []float64) prometheus.Histogram {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	})
	prometheus.MustRegister(histogram)
	m.CustomHistograms[name] = histogram
	return histogram
}

// HTTPMiddleware returns HTTP middleware for automatic metrics collection
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.IncActiveRequests()
		defer m.DecActiveRequests()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		m.RecordHTTPRequest(
			r.Method,
			r.URL.Path,
			strconv.Itoa(wrapped.statusCode),
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsHandler returns HTTP handler for Prometheus metrics endpoint
func (m *Metrics) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// Timer provides timing functionality for operations
type Timer struct {
	start   time.Time
	metrics *Metrics
}

// NewTimer creates a new timer
func (m *Metrics) NewTimer() *Timer {
	return &Timer{
		start:   time.Now(),
		metrics: m,
	}
}

// ObserveBusinessOperation observes a business operation with the timer duration
func (t *Timer) ObserveBusinessOperation(operation, status string) {
	duration := time.Since(t.start)
	t.metrics.RecordBusinessOperation(operation, status, duration)
}

// HealthChecker provides health check functionality with metrics
type HealthChecker struct {
	metrics *Metrics
	checks  map[string]func() bool
}

// NewHealthChecker creates a new health checker
func (m *Metrics) NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		metrics: m,
		checks:  make(map[string]func() bool),
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(component string, checkFunc func() bool) {
	hc.checks[component] = checkFunc
}

// RunChecks runs all health checks and updates metrics
func (hc *HealthChecker) RunChecks() map[string]bool {
	results := make(map[string]bool)
	for component, checkFunc := range hc.checks {
		healthy := checkFunc()
		results[component] = healthy
		hc.metrics.SetServiceHealth(component, healthy)
	}
	return results
}
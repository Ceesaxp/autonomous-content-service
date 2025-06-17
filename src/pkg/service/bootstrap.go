package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/pkg/auth"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/config"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/discovery"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/events"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/middleware"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/monitoring"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/resilience"
	"github.com/gorilla/mux"
)

// ServiceBootstrap provides a complete service bootstrap framework
type ServiceBootstrap struct {
	ServiceName     string
	Port            int
	Config          *config.ServiceConfig
	Metrics         *monitoring.Metrics
	Discovery       discovery.ServiceDiscovery
	EventPublisher  events.EventPublisher
	EventConsumer   events.EventConsumer
	Auth            *auth.AuthMiddleware
	CircuitBreakers *resilience.CircuitBreakerManager
	Router          *mux.Router
	Server          *http.Server
	
	// Internal state
	startTime    time.Time
	shutdownChan chan os.Signal
}

// NewServiceBootstrap creates a new service bootstrap instance
func NewServiceBootstrap(serviceName string) (*ServiceBootstrap, error) {
	log.Printf("Bootstrapping service: %s", serviceName)

	// Load service configuration
	serviceConfig, err := config.NewServiceConfig(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to load service config: %w", err)
	}

	// Initialize metrics
	metrics := monitoring.NewMetrics(serviceName)

	// Initialize service discovery
	consulConfig := &discovery.ConsulConfig{
		Address:    serviceConfig.ConsulAddr,
		Datacenter: "dc1",
		TTL:        time.Minute * 5,
	}
	
	serviceDiscovery, err := discovery.NewConsulClient(consulConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize service discovery: %v", err)
		// Continue without service discovery in development
	}

	// Initialize event system
	eventClient := events.NewRedisStreamsClient(serviceConfig.RedisAddr, "", 0)

	// Initialize authentication
	jwtSecret := serviceConfig.ConfigMgr.GetString("JWT_SECRET", "default-jwt-secret")
	serviceSecret := serviceConfig.ConfigMgr.GetString("SERVICE_SECRET", "default-service-secret")
	authMiddleware := auth.NewAuthMiddleware(jwtSecret, serviceSecret)

	// Initialize circuit breaker manager
	circuitBreakers := resilience.NewCircuitBreakerManager()

	// Create router with common middleware
	router := mux.NewRouter()
	
	// Set up common middleware stack
	router.Use(middleware.Recovery)
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging)
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	router.Use(metrics.HTTPMiddleware)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serviceConfig.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &ServiceBootstrap{
		ServiceName:     serviceName,
		Port:            serviceConfig.Port,
		Config:          serviceConfig,
		Metrics:         metrics,
		Discovery:       serviceDiscovery,
		EventPublisher:  eventClient,
		EventConsumer:   eventClient,
		Auth:            authMiddleware,
		CircuitBreakers: circuitBreakers,
		Router:          router,
		Server:          server,
		startTime:       time.Now(),
		shutdownChan:    make(chan os.Signal, 1),
	}, nil
}

// RegisterRoutes allows services to register their specific routes
func (sb *ServiceBootstrap) RegisterRoutes(registerFunc func(*mux.Router)) {
	// Create subrouter for API routes with authentication
	apiRouter := sb.Router.PathPrefix("/api").Subrouter()
	apiRouter.Use(sb.Auth.ServiceToServiceMiddleware)
	
	// Register service-specific routes
	registerFunc(apiRouter)
}

// AddHealthChecks adds standard health check endpoints
func (sb *ServiceBootstrap) AddHealthChecks(customChecks map[string]func() bool) {
	healthChecker := sb.Metrics.NewHealthChecker()
	
	// Add custom health checks
	for name, checkFunc := range customChecks {
		healthChecker.AddCheck(name, checkFunc)
	}

	// Health check endpoint
	sb.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		results := healthChecker.RunChecks()
		
		allHealthy := true
		for _, healthy := range results {
			if !healthy {
				allHealthy = false
				break
			}
		}

		status := http.StatusOK
		if !allHealthy {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		
		response := map[string]interface{}{
			"status":     getStatusString(allHealthy),
			"timestamp":  time.Now().UTC(),
			"uptime":     time.Since(sb.startTime).String(),
			"service":    sb.ServiceName,
			"version":    "1.0.0",
			"checks":     results,
		}
		
		writeJSONResponse(w, response)
	}).Methods("GET")

	// Readiness probe
	sb.Router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSONResponse(w, map[string]string{
			"status": "ready",
			"service": sb.ServiceName,
		})
	}).Methods("GET")

	// Liveness probe
	sb.Router.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSONResponse(w, map[string]string{
			"status": "alive",
			"service": sb.ServiceName,
		})
	}).Methods("GET")

	// Metrics endpoint
	sb.Router.Handle("/metrics", sb.Metrics.MetricsHandler()).Methods("GET")
}

// Start starts the service with all bootstrap components
func (sb *ServiceBootstrap) Start() error {
	// Set up signal handling for graceful shutdown
	signal.Notify(sb.shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Register with service discovery
	if sb.Discovery != nil {
		address := fmt.Sprintf("localhost:%d", sb.Port)
		err := sb.Discovery.RegisterService(sb.ServiceName, address, "/health")
		if err != nil {
			log.Printf("Warning: Failed to register with service discovery: %v", err)
		} else {
			log.Printf("Service registered with discovery at %s", address)
		}
	}

	// Set initial metrics
	sb.Metrics.SetServiceVersion("1.0.0", "unknown", time.Now().Format(time.RFC3339))
	sb.Metrics.SetServiceHealth("bootstrap", true)

	// Start HTTP server
	log.Printf("Starting %s service on port %d", sb.ServiceName, sb.Port)
	
	go func() {
		if err := sb.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Update uptime metric periodically
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				sb.Metrics.SetServiceUptime(time.Since(sb.startTime))
			case <-sb.shutdownChan:
				return
			}
		}
	}()

	log.Printf("Service %s started successfully on port %d", sb.ServiceName, sb.Port)
	return nil
}

// WaitForShutdown waits for shutdown signal and performs graceful shutdown
func (sb *ServiceBootstrap) WaitForShutdown() {
	// Wait for shutdown signal
	<-sb.shutdownChan
	sb.Shutdown()
}

// Shutdown performs graceful shutdown of all service components
func (sb *ServiceBootstrap) Shutdown() {
	log.Printf("Shutting down service %s...", sb.ServiceName)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := sb.Server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// Deregister from service discovery
	if sb.Discovery != nil {
		if consulClient, ok := sb.Discovery.(*discovery.ConsulClient); ok {
			consulClient.DeregisterAll()
		}
	}

	// Close event client
	if sb.EventPublisher != nil {
		sb.EventPublisher.Close()
	}

	log.Printf("Service %s shutdown complete", sb.ServiceName)
}

// CreateCircuitBreaker creates a circuit breaker for a specific operation
func (sb *ServiceBootstrap) CreateCircuitBreaker(name string, config resilience.CircuitBreakerConfig) *resilience.CircuitBreaker {
	return sb.CircuitBreakers.GetOrCreate(name, config)
}

// PublishEvent publishes an event to the event stream
func (sb *ServiceBootstrap) PublishEvent(ctx context.Context, stream string, eventType string, payload map[string]interface{}) error {
	if sb.EventPublisher == nil {
		return fmt.Errorf("event publisher not available")
	}

	event := events.CreateEvent(eventType, sb.ServiceName, payload)
	return sb.EventPublisher.Publish(ctx, stream, event)
}

// SubscribeToEvents subscribes to events from a stream
func (sb *ServiceBootstrap) SubscribeToEvents(ctx context.Context, stream, consumerGroup string, handler events.EventHandler) error {
	if sb.EventConsumer == nil {
		return fmt.Errorf("event consumer not available")
	}

	return sb.EventConsumer.Subscribe(ctx, stream, consumerGroup, handler)
}

// Helper functions

func getStatusString(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	// Simple JSON marshaling - in production you'd use a proper JSON library
	w.Write([]byte(fmt.Sprintf(`{"data": %v}`, data)))
}

// ServiceInfo provides information about the service
type ServiceInfo struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	BuildTime string    `json:"build_time"`
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`
}

// GetServiceInfo returns service information
func (sb *ServiceBootstrap) GetServiceInfo() ServiceInfo {
	return ServiceInfo{
		Name:      sb.ServiceName,
		Version:   "1.0.0",
		BuildTime: "unknown",
		StartTime: sb.startTime,
		Uptime:    time.Since(sb.startTime).String(),
	}
}
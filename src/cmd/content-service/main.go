package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/api/handlers"
	"github.com/Ceesaxp/autonomous-content-service/src/config"
	"github.com/Ceesaxp/autonomous-content-service/src/infrastructure/database"
	"github.com/Ceesaxp/autonomous-content-service/src/services/content_creation"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/events"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/service"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Content Service...")

	// Initialize service bootstrap
	bootstrap, err := service.NewServiceBootstrap("content-service")
	if err != nil {
		log.Fatalf("Failed to initialize service bootstrap: %v", err)
	}

	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up database connection
	db, err := database.NewPostgresDB(config.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	clientRepo := database.NewClientRepository(db)
	projectRepo := database.NewProjectRepository(db)
	contentRepo := database.NewContentRepository(db)
	contentVersionRepo := database.NewContentVersionRepository(db)
	feedbackRepo := database.NewFeedbackRepository(db)
	eventRepo := database.NewEventRepository(db)

	// Initialize event system
	eventClient := events.NewRedisStreamsClient(
		bootstrap.Config.RedisAddr,
		"",
		0,
	)

	// Create event bus factory
	eventBusFactory := events.NewServiceEventBusFactory(
		eventClient,
		eventClient,
		bootstrap.Discovery,
	)

	// Create service event bus
	eventBus := eventBusFactory.CreateEventBus("content-service")

	// Initialize content creation services
	llmClient := content_creation.NewOpenAIClient(
		config.LLMAPIKey,
		config.LLMModel,
		config.LLMMaxTokens,
		config.LLMTemperature,
	)

	searchService := content_creation.NewWebSearchService(
		config.SearchAPIKey,
		config.SearchURL,
	)

	plagiarismAPI := content_creation.NewSimplePlagiarismAPI()
	readabilityScorer := content_creation.NewBasicReadabilityScorer()
	seoAnalyzer := content_creation.NewBasicSEOAnalyzer()

	contextManager := content_creation.NewInMemoryContextManager(
		clientRepo,
		config.ContextWindowSize,
	)

	qualityChecker := content_creation.NewLLMQualityChecker(
		llmClient,
		plagiarismAPI,
		readabilityScorer,
		seoAnalyzer,
	)

	pipelineConfig := content_creation.PipelineConfig{
		MaxRetries:           3,
		ContextWindowSize:    config.ContextWindowSize,
		EnableFactChecking:   config.EnableFactChecking,
		EnablePlagiarismCheck: config.EnablePlagiarism,
		SEOOptimization:      config.EnableSEO,
	}

	researcher := content_creation.NewLLMResearcher(
		llmClient,
		searchService,
	)

	contentPipeline := content_creation.NewContentPipeline(
		contentRepo,
		contentVersionRepo,
		projectRepo,
		eventRepo,
		llmClient,
		contextManager,
		researcher,
		qualityChecker,
		pipelineConfig,
	)

	// Initialize event-integrated content service
	eventIntegratedContentService := content_creation.NewEventIntegratedContentService(
		contentPipeline,
		eventBus,
		contentRepo,
		projectRepo,
	)

	// Initialize handlers
	contentHandler := handlers.NewContentHandler(
		contentRepo,
		projectRepo,
		feedbackRepo,
		contentPipeline,
	)

	// Add health checks
	healthChecks := map[string]func() bool{
		"database": func() bool {
			return db.Ping() == nil
		},
		"events": func() bool {
			return eventClient != nil
		},
		"content_pipeline": func() bool {
			return contentPipeline != nil
		},
	}
	bootstrap.AddHealthChecks(healthChecks)

	// Register routes with the bootstrap router
	bootstrap.RegisterRoutes(func(router *mux.Router) {
		setupContentRoutes(router, contentHandler)
	})

	// Start event bus
	if err := eventBus.Start(context.Background()); err != nil {
		log.Printf("Warning: Failed to start event bus: %v", err)
	}
	defer eventBus.Stop()

	// Start event listeners
	if err := eventIntegratedContentService.StartEventListeners(context.Background()); err != nil {
		log.Printf("Warning: Failed to start event listeners: %v", err)
	}

	// Start the service
	if err := bootstrap.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Wait for shutdown
	bootstrap.WaitForShutdown()

	log.Println("Content Service exited gracefully")
}

// setupContentRoutes configures content-specific routes
func setupContentRoutes(router *mux.Router, contentHandler *handlers.ContentHandler) {
	// Content creation and management
	router.HandleFunc("/projects/{project_id}/content", contentHandler.CreateContent).Methods("POST")
	router.HandleFunc("/content/{content_id}", contentHandler.GetContent).Methods("GET")
	router.HandleFunc("/content/{content_id}", contentHandler.UpdateContent).Methods("PUT")
	router.HandleFunc("/content/{content_id}/approve", contentHandler.ApproveContent).Methods("POST")
	router.HandleFunc("/content/{content_id}/versions", contentHandler.GetContentVersions).Methods("GET")
}

// healthCheckHandler provides health status for the Content Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"content-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Content Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// getServicePort gets port from environment or uses default
func getServicePort(envVar string, defaultPort int) int {
	if portStr := os.Getenv(envVar); portStr != "" {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
			return port
		}
	}
	return defaultPort
}
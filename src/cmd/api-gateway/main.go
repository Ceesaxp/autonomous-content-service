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

	"github.com/Ceesaxp/autonomous-content-service/src/api"
	"github.com/Ceesaxp/autonomous-content-service/src/api/handlers"
	"github.com/Ceesaxp/autonomous-content-service/src/config"
	"github.com/Ceesaxp/autonomous-content-service/src/infrastructure/database"
	"github.com/Ceesaxp/autonomous-content-service/src/services/content_creation"
	"github.com/Ceesaxp/autonomous-content-service/src/services/risk_management"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting API Gateway Service...")

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

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repositories
	clientRepo := database.NewClientRepository(db)
	projectRepo := database.NewProjectRepository(db)
	contentRepo := database.NewContentRepository(db)
	contentVersionRepo := database.NewContentVersionRepository(db)
	feedbackRepo := database.NewFeedbackRepository(db)
	eventRepo := database.NewEventRepository(db)
	riskRepo := database.NewRiskRepository(db)

	// Initialize services (centralized for now, will be distributed later)
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

	// Initialize the researcher
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

	// Initialize handlers
	contentHandler := handlers.NewContentHandler(
		contentRepo,
		projectRepo,
		feedbackRepo,
		contentPipeline,
	)

	projectHandler := handlers.NewProjectHandler(
		projectRepo,
		contentRepo,
		clientRepo,
	)

	// Initialize risk management service
	riskService := risk_management.NewRiskManagementService(
		riskRepo,
		nil, // payment repo - will implement stub methods for now
		clientRepo,
		eventRepo,
	)

	// Initialize risk handlers
	riskHandler := handlers.NewRiskHandlers(riskService)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// Add health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Dashboard handler can be nil for now since we don't have a complete dashboard service implementation
	var dashboardHandler *handlers.DashboardHandlers = nil
	
	// Self-improvement handler can be nil for now since we don't have a complete implementation
	var selfImprovementHandler *handlers.SelfImprovementHandler = nil

	// Set up API routes
	api.SetupRoutes(router, contentHandler, projectHandler, nil, dashboardHandler, selfImprovementHandler, riskHandler) // nil for onboarding handler until we initialize it

	// Set up server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API Gateway listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Gateway...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway exited gracefully")
}

// healthCheckHandler provides health status for the API Gateway
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"api-gateway","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
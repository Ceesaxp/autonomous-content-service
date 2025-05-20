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
	"github.com/gorilla/mux"
)

func main() {
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
	clientProfileRepo := database.NewClientProfileRepository(db)
	projectRepo := database.NewProjectRepository(db)
	contentRepo := database.NewContentRepository(db)
	contentVersionRepo := database.NewContentVersionRepository(db)
	feedbackRepo := database.NewFeedbackRepository(db)
	systemCapabilityRepo := database.NewSystemCapabilityRepository(db)
	eventRepo := database.NewEventRepository(db)

	// Initialize services
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

	researcher := content_creation.NewLLMResearcher(
		llmClient,
		searchService,
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

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Set up API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	api.SetupRoutes(apiRouter, contentHandler, projectHandler)

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
		log.Printf("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

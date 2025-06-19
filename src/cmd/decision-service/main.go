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
	"github.com/Ceesaxp/autonomous-content-service/src/services/decision_making"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Decision Service...")

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
	decisionRepo := database.NewDecisionRepository(db)
	eventRepo := database.NewEventRepository(db)

	// Set up router with basic handlers for initial microservices implementation
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Initialize LLM client
	llmClient := content_creation.NewOpenAIClient(
		config.LLMAPIKey,
		config.LLMModel,
		config.LLMMaxTokens,
		config.LLMTemperature,
	)

	// Initialize decision service
	decisionService := decision_making.NewService(
		decisionRepo,
		eventRepo,
		llmClient,
	)

	// Initialize handlers
	decisionHandler := handlers.NewDecisionHandlers(decisionService)

	// Decision service routes
	decisionRouter := router.PathPrefix("/api/v1").Subrouter()
	setupDecisionRoutes(decisionRouter, decisionHandler)

	// Set up server
	port := getServicePort("DECISION_SERVICE_PORT", 8082)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Decision Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Decision Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Decision Service exited gracefully")
}

// setupDecisionRoutes configures decision-specific routes with real implementations
func setupDecisionRoutes(router *mux.Router, handler *handlers.DecisionHandlers) {
	// Core decision endpoints
	router.HandleFunc("/decisions", handler.CreateDecision).Methods("POST")
	router.HandleFunc("/decisions/{id}", handler.GetDecision).Methods("GET")
	// Note: UpdateDecision method not implemented in handlers
	router.HandleFunc("/decisions/{id}/execute", handler.ExecuteDecision).Methods("POST")
	router.HandleFunc("/decisions/{id}/override", handler.OverrideDecision).Methods("POST")
	router.HandleFunc("/decisions", handler.ListDecisions).Methods("GET")
	
	// Quality and metrics
	router.HandleFunc("/decisions/{id}/quality", handler.AssessDecisionQuality).Methods("GET")
	router.HandleFunc("/decisions/metrics", handler.GetDecisionMetrics).Methods("GET")
	
	// Emergency and system health
	router.HandleFunc("/system/emergency", handler.ActivateEmergencyMode).Methods("POST")
}

// healthCheckHandler provides health status for the Decision Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"decision-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Decision Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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


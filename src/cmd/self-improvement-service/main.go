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
	"github.com/Ceesaxp/autonomous-content-service/src/services/self_improvement"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Self-Improvement Service...")

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
	learningRepo := database.NewLearningRepository(db)

	// Initialize self-improvement service
	selfImprovementService := self_improvement.NewSelfImprovementService(learningRepo)

	// Initialize handlers
	selfImprovementHandler := handlers.NewSelfImprovementHandler(selfImprovementService)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Self-improvement service routes
	siRouter := router.PathPrefix("/api/v1").Subrouter()
	setupSelfImprovementRoutes(siRouter, selfImprovementHandler)

	// Set up server
	port := getServicePort("SELF_IMPROVEMENT_SERVICE_PORT", 8088)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Self-Improvement Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Self-Improvement Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Self-Improvement Service exited gracefully")
}

// setupSelfImprovementRoutes configures self-improvement-specific routes
func setupSelfImprovementRoutes(router *mux.Router, siHandler *handlers.SelfImprovementHandler) {
	// Learning and knowledge management
	router.HandleFunc("/learning/artifacts", siHandler.CreateLearningArtifact).Methods("POST")
	router.HandleFunc("/learning/artifacts", siHandler.GetLearningArtifacts).Methods("GET")
	router.HandleFunc("/learning/artifacts/{id}", siHandler.GetLearningArtifact).Methods("GET")
	router.HandleFunc("/learning/learn", siHandler.LearnFromProject).Methods("POST")
	router.HandleFunc("/learning/feedback", siHandler.LearnFromFeedback).Methods("POST")
	
	// Performance metrics and monitoring
	router.HandleFunc("/metrics/collect", siHandler.CollectMetrics).Methods("POST")
	router.HandleFunc("/metrics/performance", siHandler.GetPerformanceMetrics).Methods("GET")
	router.HandleFunc("/metrics/system", siHandler.GetSystemMetrics).Methods("GET")
	
	// Experimentation and A/B testing
	router.HandleFunc("/experiments", siHandler.CreateExperiment).Methods("POST")
	router.HandleFunc("/experiments", siHandler.GetExperiments).Methods("GET")
	router.HandleFunc("/experiments/{id}/results", siHandler.GetExperimentResults).Methods("GET")
	router.HandleFunc("/experiments/{id}/conclude", siHandler.ConcludeExperiment).Methods("POST")
	
	// Capability acquisition and optimization
	router.HandleFunc("/capabilities/gaps", siHandler.IdentifyCapabilityGaps).Methods("GET")
	router.HandleFunc("/capabilities/acquire", siHandler.AcquireCapability).Methods("POST")
	router.HandleFunc("/optimization/prompts", siHandler.OptimizePrompts).Methods("POST")
	router.HandleFunc("/optimization/workflows", siHandler.OptimizeWorkflows).Methods("POST")
	
	// Knowledge graph and insights
	router.HandleFunc("/knowledge/query", siHandler.QueryKnowledge).Methods("POST")
	router.HandleFunc("/knowledge/insights", siHandler.GetInsights).Methods("GET")
	router.HandleFunc("/knowledge/recommendations", siHandler.GetRecommendations).Methods("GET")
}

// healthCheckHandler provides health status for the Self-Improvement Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"self-improvement-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Self-Improvement Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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
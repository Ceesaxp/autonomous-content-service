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

	"github.com/Ceesaxp/autonomous-content-service/src/config"
	"github.com/Ceesaxp/autonomous-content-service/src/infrastructure/database"
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

	// Initialize repositories (using existing event repo for now)
	_ = database.NewEventRepository(db) // eventRepo for future use

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Self-improvement service routes (basic implementations for testing)
	siRouter := router.PathPrefix("/api/v1").Subrouter()
	setupSelfImprovementRoutes(siRouter)

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

// setupSelfImprovementRoutes configures self-improvement-specific routes with basic implementations
func setupSelfImprovementRoutes(router *mux.Router) {
	// Learning and knowledge management
	router.HandleFunc("/learning/artifacts", basicHandler("learning", "learning artifact created")).Methods("POST")
	router.HandleFunc("/learning/artifacts", basicHandler("learning", "learning artifacts retrieved")).Methods("GET")
	router.HandleFunc("/learning/artifacts/{id}", basicHandler("learning", "learning artifact retrieved")).Methods("GET")
	router.HandleFunc("/learning/learn", basicHandler("learning", "learned from project")).Methods("POST")
	router.HandleFunc("/learning/feedback", basicHandler("learning", "learned from feedback")).Methods("POST")
	
	// Performance metrics and monitoring
	router.HandleFunc("/metrics/collect", basicHandler("metrics", "metrics collected")).Methods("POST")
	router.HandleFunc("/metrics/performance", basicHandler("metrics", "performance metrics retrieved")).Methods("GET")
	router.HandleFunc("/metrics/system", basicHandler("metrics", "system metrics retrieved")).Methods("GET")
	
	// Experimentation and A/B testing
	router.HandleFunc("/experiments", basicHandler("experiments", "experiment created")).Methods("POST")
	router.HandleFunc("/experiments", basicHandler("experiments", "experiments retrieved")).Methods("GET")
	router.HandleFunc("/experiments/{id}/results", basicHandler("experiments", "experiment results retrieved")).Methods("GET")
	router.HandleFunc("/experiments/{id}/conclude", basicHandler("experiments", "experiment concluded")).Methods("POST")
	
	// Capability acquisition and optimization
	router.HandleFunc("/capabilities/gaps", basicHandler("capabilities", "capability gaps identified")).Methods("GET")
	router.HandleFunc("/capabilities/acquire", basicHandler("capabilities", "capability acquired")).Methods("POST")
	router.HandleFunc("/optimization/prompts", basicHandler("capabilities", "prompts optimized")).Methods("POST")
	router.HandleFunc("/optimization/workflows", basicHandler("capabilities", "workflows optimized")).Methods("POST")
	
	// Knowledge graph and insights
	router.HandleFunc("/knowledge/query", basicHandler("knowledge", "knowledge queried")).Methods("POST")
	router.HandleFunc("/knowledge/insights", basicHandler("knowledge", "insights retrieved")).Methods("GET")
	router.HandleFunc("/knowledge/recommendations", basicHandler("knowledge", "recommendations retrieved")).Methods("GET")
}

// basicHandler creates a basic JSON response handler
func basicHandler(action, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","action":"%s","message":"Self-improvement service endpoint - implementation pending: %s"}`, action, message)))
	}
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
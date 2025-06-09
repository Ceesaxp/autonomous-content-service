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

	// Initialize repositories (using existing event repo for now)
	_ = database.NewEventRepository(db) // eventRepo for future use

	// Set up router with basic handlers for initial microservices implementation
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Decision service routes (basic implementations for testing)
	decisionRouter := router.PathPrefix("/api/v1").Subrouter()
	setupDecisionRoutes(decisionRouter)

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

// setupDecisionRoutes configures decision-specific routes with basic implementations
func setupDecisionRoutes(router *mux.Router) {
	router.HandleFunc("/decisions", createDecisionHandler).Methods("POST")
	router.HandleFunc("/decisions/{id}", getDecisionHandler).Methods("GET")
	router.HandleFunc("/decisions/{id}/execute", executeDecisionHandler).Methods("POST")
	router.HandleFunc("/decisions/{id}/override", overrideDecisionHandler).Methods("POST")
	router.HandleFunc("/decisions/{id}/quality", assessDecisionQualityHandler).Methods("GET")
	router.HandleFunc("/decisions/metrics", getDecisionMetricsHandler).Methods("GET")
	router.HandleFunc("/policies", createPolicyHandler).Methods("POST")
	router.HandleFunc("/policies", getPoliciesHandler).Methods("GET")
	router.HandleFunc("/ethical-guidelines", createEthicalGuidelineHandler).Methods("POST")
	router.HandleFunc("/system/health", checkSystemHealthHandler).Methods("GET")
	router.HandleFunc("/system/emergency", activateEmergencyModeHandler).Methods("POST")
}

// Basic handler implementations for testing microservices architecture
func createDecisionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"decision created","id":"mock-decision-123","message":"Decision service endpoint - implementation pending"}`))
}

func getDecisionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","status":"pending","type":"mock","message":"Decision service endpoint - implementation pending"}`, id)))
}

func executeDecisionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"executed","message":"Decision service endpoint - implementation pending"}`))
}

func overrideDecisionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"overridden","message":"Decision service endpoint - implementation pending"}`))
}

func assessDecisionQualityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"quality_score":0.85,"message":"Decision service endpoint - implementation pending"}`))
}

func getDecisionMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"metrics":{"total_decisions":42,"success_rate":0.91},"message":"Decision service endpoint - implementation pending"}`))
}

func createPolicyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"policy created","message":"Decision service endpoint - implementation pending"}`))
}

func getPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"policies":[],"message":"Decision service endpoint - implementation pending"}`))
}

func createEthicalGuidelineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"guideline created","message":"Decision service endpoint - implementation pending"}`))
}

func checkSystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","components":{"decision_engine":"healthy"},"message":"Decision service endpoint - implementation pending"}`))
}

func activateEmergencyModeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"emergency mode activated","message":"Decision service endpoint - implementation pending"}`))
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


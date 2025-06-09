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
	log.Println("Starting Legal Service...")

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

	// Legal service routes (basic implementations for testing)
	legalRouter := router.PathPrefix("/api/v1").Subrouter()
	setupLegalRoutes(legalRouter)

	// Set up server
	port := getServicePort("LEGAL_SERVICE_PORT", 8086)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Legal Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Legal Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Legal Service exited gracefully")
}

// setupLegalRoutes configures legal-specific routes with basic implementations
func setupLegalRoutes(router *mux.Router) {
	// Contract management
	router.HandleFunc("/contracts", basicHandler("contracts", "contract generated")).Methods("POST")
	router.HandleFunc("/contracts", basicHandler("contracts", "contracts retrieved")).Methods("GET")
	router.HandleFunc("/contracts/{id}", basicHandler("contracts", "contract retrieved")).Methods("GET")
	router.HandleFunc("/contracts/{id}/review", basicHandler("contracts", "contract reviewed")).Methods("POST")
	router.HandleFunc("/contracts/{id}/sign", basicHandler("contracts", "signature requested")).Methods("POST")
	
	// Compliance monitoring
	router.HandleFunc("/compliance/check", basicHandler("compliance", "compliance checked")).Methods("POST")
	router.HandleFunc("/compliance/requirements", basicHandler("compliance", "requirements retrieved")).Methods("GET")
	router.HandleFunc("/compliance/reports", basicHandler("compliance", "report generated")).Methods("POST")
	
	// Data privacy (GDPR)
	router.HandleFunc("/privacy/data-subject-request", basicHandler("privacy", "data subject request processed")).Methods("POST")
	router.HandleFunc("/privacy/detect-pii", basicHandler("privacy", "PII detected")).Methods("POST")
	router.HandleFunc("/privacy/consent", basicHandler("privacy", "consent managed")).Methods("POST")
	
	// IP management
	router.HandleFunc("/ip/licenses", basicHandler("ip", "IP license registered")).Methods("POST")
	router.HandleFunc("/ip/licenses", basicHandler("ip", "IP licenses retrieved")).Methods("GET")
	router.HandleFunc("/ip/usage/validate", basicHandler("ip", "IP usage validated")).Methods("POST")
	
	// Insurance management
	router.HandleFunc("/insurance/policies", basicHandler("insurance", "policies retrieved")).Methods("GET")
	router.HandleFunc("/insurance/requirements", basicHandler("insurance", "requirements calculated")).Methods("POST")
	router.HandleFunc("/insurance/claims", basicHandler("insurance", "claim processed")).Methods("POST")
	
	// Dispute management
	router.HandleFunc("/disputes", basicHandler("disputes", "dispute created")).Methods("POST")
	router.HandleFunc("/disputes", basicHandler("disputes", "disputes retrieved")).Methods("GET")
	router.HandleFunc("/disputes/{id}/resolve", basicHandler("disputes", "dispute resolved")).Methods("POST")
}

// basicHandler creates a basic JSON response handler
func basicHandler(action, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","action":"%s","message":"Legal service endpoint - implementation pending: %s"}`, action, message)))
	}
}

// healthCheckHandler provides health status for the Legal Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"legal-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Legal Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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
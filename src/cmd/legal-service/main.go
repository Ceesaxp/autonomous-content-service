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
	"github.com/Ceesaxp/autonomous-content-service/src/services/legal_compliance"
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

	// Initialize repositories
	legalRepo := database.NewLegalRepository(db)

	// Initialize legal service
	legalService := legal_compliance.NewLegalService(legalRepo)

	// Initialize handlers
	legalHandler := handlers.NewLegalHandlers(legalService)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Legal service routes
	legalRouter := router.PathPrefix("/api/v1").Subrouter()
	setupLegalRoutes(legalRouter, legalHandler)

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

// setupLegalRoutes configures legal-specific routes
func setupLegalRoutes(router *mux.Router, legalHandler *handlers.LegalHandlers) {
	// Contract management
	router.HandleFunc("/contracts", legalHandler.GenerateContract).Methods("POST")
	router.HandleFunc("/contracts", legalHandler.GetContracts).Methods("GET")
	router.HandleFunc("/contracts/{id}", legalHandler.GetContract).Methods("GET")
	router.HandleFunc("/contracts/{id}/review", legalHandler.ReviewContract).Methods("POST")
	router.HandleFunc("/contracts/{id}/sign", legalHandler.RequestSignature).Methods("POST")
	
	// Compliance monitoring
	router.HandleFunc("/compliance/check", legalHandler.CheckCompliance).Methods("POST")
	router.HandleFunc("/compliance/requirements", legalHandler.GetComplianceRequirements).Methods("GET")
	router.HandleFunc("/compliance/reports", legalHandler.GenerateComplianceReport).Methods("POST")
	
	// Data privacy (GDPR)
	router.HandleFunc("/privacy/data-subject-request", legalHandler.ProcessDataSubjectRequest).Methods("POST")
	router.HandleFunc("/privacy/detect-pii", legalHandler.DetectPII).Methods("POST")
	router.HandleFunc("/privacy/consent", legalHandler.ManageConsent).Methods("POST")
	
	// IP management
	router.HandleFunc("/ip/licenses", legalHandler.RegisterIPLicense).Methods("POST")
	router.HandleFunc("/ip/licenses", legalHandler.GetIPLicenses).Methods("GET")
	router.HandleFunc("/ip/usage/validate", legalHandler.ValidateIPUsage).Methods("POST")
	
	// Insurance management
	router.HandleFunc("/insurance/policies", legalHandler.GetInsurancePolicies).Methods("GET")
	router.HandleFunc("/insurance/requirements", legalHandler.CalculateInsuranceRequirements).Methods("POST")
	router.HandleFunc("/insurance/claims", legalHandler.ProcessInsuranceClaim).Methods("POST")
	
	// Dispute management
	router.HandleFunc("/disputes", legalHandler.CreateDispute).Methods("POST")
	router.HandleFunc("/disputes", legalHandler.GetDisputes).Methods("GET")
	router.HandleFunc("/disputes/{id}/resolve", legalHandler.ResolveDispute).Methods("POST")
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
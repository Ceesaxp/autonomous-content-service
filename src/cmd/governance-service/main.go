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
	log.Println("Starting Governance Service...")

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

	// Initialize repositories for future use
	_ = database.NewEventRepository(db)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Governance service routes
	govRouter := router.PathPrefix("/api/v1").Subrouter()
	setupGovernanceRoutes(govRouter)

	// Set up server
	port := getServicePort("GOVERNANCE_SERVICE_PORT", 8085)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Governance Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Governance Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Governance Service exited gracefully")
}

// setupGovernanceRoutes configures governance-specific routes
func setupGovernanceRoutes(router *mux.Router) {
	// Proposal management
	router.HandleFunc("/proposals", basicHandler("proposal created")).Methods("POST")
	router.HandleFunc("/proposals", basicHandler("proposals")).Methods("GET")
	router.HandleFunc("/proposals/{id}", basicHandler("proposal")).Methods("GET")
	router.HandleFunc("/proposals/{id}/vote", basicHandler("vote cast")).Methods("POST")
	router.HandleFunc("/proposals/{id}/execute", basicHandler("proposal executed")).Methods("POST")
	
	// Member management
	router.HandleFunc("/members", basicHandler("member registered")).Methods("POST")
	router.HandleFunc("/members", basicHandler("members")).Methods("GET")
	router.HandleFunc("/members/{id}", basicHandler("member")).Methods("GET")
	router.HandleFunc("/members/{id}/delegate", basicHandler("vote delegated")).Methods("POST")
	
	// Treasury allocations
	router.HandleFunc("/treasury/allocations", basicHandler("allocation created")).Methods("POST")
	router.HandleFunc("/treasury/allocations", basicHandler("treasury allocations")).Methods("GET")
	router.HandleFunc("/treasury/allocations/{id}/release", basicHandler("funds released")).Methods("POST")
	
	// Governance metrics
	router.HandleFunc("/governance/metrics", basicHandler("governance metrics")).Methods("GET")
	router.HandleFunc("/governance/reports", basicHandler("governance report")).Methods("POST")
}

// basicHandler returns a simple JSON response for testing
func basicHandler(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","action":"%s","message":"Governance service endpoint - implementation pending"}`, action)))
	}
}

// healthCheckHandler provides health status for the Governance Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"governance-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Governance Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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
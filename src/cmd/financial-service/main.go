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
	log.Println("Starting Financial Service...")

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
	_ = database.NewTransactionRepository(db)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Financial service routes
	financialRouter := router.PathPrefix("/api/v1").Subrouter()
	setupFinancialRoutes(financialRouter)

	// Set up server
	port := getServicePort("FINANCIAL_SERVICE_PORT", 8084)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Financial Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Financial Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Financial Service exited gracefully")
}

// setupFinancialRoutes configures financial-specific routes
func setupFinancialRoutes(router *mux.Router) {
	// Payment processing
	router.HandleFunc("/payments/process", basicHandler("payment processed")).Methods("POST")
	router.HandleFunc("/payments/refund", basicHandler("payment refunded")).Methods("POST")
	router.HandleFunc("/payments/{id}/status", basicHandler("payment status")).Methods("GET")
	router.HandleFunc("/payments/methods", basicHandler("payment methods")).Methods("GET")
	
	// Pricing and quotes
	router.HandleFunc("/pricing/quote", basicHandler("price quote")).Methods("POST")
	router.HandleFunc("/pricing/models", basicHandler("pricing models")).Methods("GET")
	router.HandleFunc("/pricing/experiments", basicHandler("pricing experiment")).Methods("POST")
	router.HandleFunc("/pricing/market-data", basicHandler("market data")).Methods("GET")
	
	// Treasury operations
	router.HandleFunc("/treasury/balance", basicHandler("treasury balance")).Methods("GET")
	router.HandleFunc("/treasury/allocate", basicHandler("funds allocated")).Methods("POST")
	router.HandleFunc("/treasury/transactions", basicHandler("treasury transactions")).Methods("GET")
}

// basicHandler returns a simple JSON response for testing
func basicHandler(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","action":"%s","message":"Financial service endpoint - implementation pending"}`, action)))
	}
}

// healthCheckHandler provides health status for the Financial Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"financial-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Financial Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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
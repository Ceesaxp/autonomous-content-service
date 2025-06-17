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

	// Initialize repositories
	transactionRepo := database.NewTransactionRepository(db)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Initialize handlers
	financialHandler := handlers.NewFinancialHandlers(transactionRepo)

	// Financial service routes
	financialRouter := router.PathPrefix("/api/v1").Subrouter()
	setupFinancialRoutes(financialRouter, financialHandler)

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

// setupFinancialRoutes configures financial-specific routes with real implementations
func setupFinancialRoutes(router *mux.Router, handler *handlers.FinancialHandlers) {
	// Payment processing
	router.HandleFunc("/payments/process", handler.ProcessPayment).Methods("POST")
	router.HandleFunc("/payments/refund", handler.RefundPayment).Methods("POST")
	router.HandleFunc("/payments/{id}/status", handler.GetPaymentStatus).Methods("GET")
	router.HandleFunc("/payments/methods", handler.GetPaymentMethods).Methods("GET")
	
	// Pricing and quotes
	router.HandleFunc("/pricing/quote", handler.GetPriceQuote).Methods("POST")
	router.HandleFunc("/pricing/models", handler.GetPricingModels).Methods("GET")
	router.HandleFunc("/pricing/experiments", handler.CreatePricingExperiment).Methods("POST")
	router.HandleFunc("/pricing/market-data", handler.GetMarketData).Methods("GET")
	
	// Treasury operations
	router.HandleFunc("/treasury/balance", handler.GetTreasuryBalance).Methods("GET")
	router.HandleFunc("/treasury/allocate", handler.AllocateFunds).Methods("POST")
	router.HandleFunc("/treasury/transactions", handler.GetTreasuryTransactions).Methods("GET")
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
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
	"github.com/Ceesaxp/autonomous-content-service/src/services/payment"
	"github.com/Ceesaxp/autonomous-content-service/src/services/pricing"
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
	paymentRepo := database.NewPaymentRepository(db)
	pricingRepo := database.NewPricingRepository(db)
	transactionRepo := database.NewTransactionRepository(db)

	// Initialize financial services
	paymentService := payment.NewPaymentService(paymentRepo, transactionRepo)
	pricingService := pricing.NewPricingService(pricingRepo)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Financial service routes
	financialRouter := router.PathPrefix("/api/v1").Subrouter()
	setupFinancialRoutes(financialRouter, paymentService, pricingService)

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
func setupFinancialRoutes(router *mux.Router, paymentService payment.PaymentService, pricingService pricing.PricingService) {
	// Payment processing
	router.HandleFunc("/payments/process", handleProcessPayment(paymentService)).Methods("POST")
	router.HandleFunc("/payments/refund", handleRefundPayment(paymentService)).Methods("POST")
	router.HandleFunc("/payments/{id}/status", handleGetPaymentStatus(paymentService)).Methods("GET")
	router.HandleFunc("/payments/methods", handleGetPaymentMethods(paymentService)).Methods("GET")
	
	// Pricing and quotes
	router.HandleFunc("/pricing/quote", handleGetPriceQuote(pricingService)).Methods("POST")
	router.HandleFunc("/pricing/models", handleGetPricingModels(pricingService)).Methods("GET")
	router.HandleFunc("/pricing/experiments", handleCreatePricingExperiment(pricingService)).Methods("POST")
	router.HandleFunc("/pricing/market-data", handleGetMarketData(pricingService)).Methods("GET")
	
	// Treasury operations (placeholder for smart contract integration)
	router.HandleFunc("/treasury/balance", handleGetTreasuryBalance).Methods("GET")
	router.HandleFunc("/treasury/allocate", handleAllocateFunds).Methods("POST")
	router.HandleFunc("/treasury/transactions", handleGetTreasuryTransactions).Methods("GET")
}

// Payment handlers (placeholder implementations)
func handleProcessPayment(service payment.PaymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Payment processing endpoint - implementation pending"}`))
	}
}

func handleRefundPayment(service payment.PaymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Payment refund endpoint - implementation pending"}`))
	}
}

func handleGetPaymentStatus(service payment.PaymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Payment status endpoint - implementation pending"}`))
	}
}

func handleGetPaymentMethods(service payment.PaymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Payment methods endpoint - implementation pending"}`))
	}
}

// Pricing handlers (placeholder implementations)
func handleGetPriceQuote(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Price quote endpoint - implementation pending"}`))
	}
}

func handleGetPricingModels(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Pricing models endpoint - implementation pending"}`))
	}
}

func handleCreatePricingExperiment(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Pricing experiment endpoint - implementation pending"}`))
	}
}

func handleGetMarketData(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Market data endpoint - implementation pending"}`))
	}
}

// Treasury handlers (placeholder implementations)
func handleGetTreasuryBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Treasury balance endpoint - implementation pending"}`))
}

func handleAllocateFunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Fund allocation endpoint - implementation pending"}`))
}

func handleGetTreasuryTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Treasury transactions endpoint - implementation pending"}`))
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
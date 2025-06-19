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
	"github.com/Ceesaxp/autonomous-content-service/src/services/hr_management"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting HR Service...")

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
	talentRepo := database.NewTalentRepository(db)
	engagementRepo := database.NewEngagementRepository(db)
	workAssignmentRepo := database.NewWorkAssignmentRepository(db)
	performanceRepo := database.NewPerformanceRepository(db)
	compensationRepo := database.NewCompensationRepository(db)
	trainingRepo := database.NewTrainingRepository(db)
	talentApplicationRepo := database.NewTalentApplicationRepository(db)
	complianceRepo := database.NewComplianceRepository(db)
	offboardingRepo := database.NewOffboardingRepository(db)
	eventRepo := database.NewEventRepository(db)
	
	// Initialize HR service with real repositories
	hrService := hr_management.NewHRService(
		talentRepo,
		engagementRepo,
		workAssignmentRepo,
		performanceRepo,
		compensationRepo,
		trainingRepo,
		talentApplicationRepo,
		complianceRepo,
		offboardingRepo,
		eventRepo,
	)
	
	// Initialize handlers
	hrHandlers := handlers.NewHRHandlers(hrService)

	// Set up router 
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Register HR handlers
	hrHandlers.RegisterRoutes(router)

	// Set up server
	port := getServicePort("HR_SERVICE_PORT", 8083)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("HR Service listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HR Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("HR Service exited gracefully")
}

// healthCheckHandler provides health status for the HR Service
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"hr-service","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HR Service] %s %s %s", r.RemoteAddr, r.Method, r.URL)
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

// Note: Mock implementations removed - now using real database repositories
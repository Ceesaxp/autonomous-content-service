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
	hrRepo := database.NewHRRepository(db)

	// Initialize HR service
	hrService := hr_management.NewHRService(hrRepo)

	// Initialize handlers
	hrHandler := handlers.NewHRHandlers(hrService)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// HR service routes
	hrRouter := router.PathPrefix("/api/v1").Subrouter()
	setupHRRoutes(hrRouter, hrHandler)

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

// setupHRRoutes configures HR-specific routes
func setupHRRoutes(router *mux.Router, hrHandler *handlers.HRHandlers) {
	// Talent management
	router.HandleFunc("/talent", hrHandler.CreateTalent).Methods("POST")
	router.HandleFunc("/talent", hrHandler.GetTalentProfiles).Methods("GET")
	router.HandleFunc("/talent/{id}", hrHandler.GetTalentProfile).Methods("GET")
	router.HandleFunc("/talent/{id}", hrHandler.UpdateTalentProfile).Methods("PUT")
	router.HandleFunc("/talent/{id}/status", hrHandler.UpdateTalentStatus).Methods("PUT")
	
	// Job postings and applications
	router.HandleFunc("/job-postings", hrHandler.CreateJobPosting).Methods("POST")
	router.HandleFunc("/job-postings", hrHandler.GetJobPostings).Methods("GET")
	router.HandleFunc("/job-postings/{id}/applications", hrHandler.GetApplications).Methods("GET")
	router.HandleFunc("/applications/{id}/screen", hrHandler.ScreenApplication).Methods("POST")
	
	// Engagements and assignments
	router.HandleFunc("/engagements", hrHandler.CreateEngagement).Methods("POST")
	router.HandleFunc("/engagements", hrHandler.GetEngagements).Methods("GET")
	router.HandleFunc("/engagements/{id}/assignments", hrHandler.CreateWorkAssignment).Methods("POST")
	router.HandleFunc("/assignments/{id}/complete", hrHandler.CompleteAssignment).Methods("POST")
	
	// Performance and reviews
	router.HandleFunc("/talent/{id}/performance", hrHandler.GetPerformanceMetrics).Methods("GET")
	router.HandleFunc("/talent/{id}/reviews", hrHandler.CreatePerformanceReview).Methods("POST")
	router.HandleFunc("/talent/{id}/compensation", hrHandler.GetCompensationPlan).Methods("GET")
	
	// Training and development
	router.HandleFunc("/training-programs", hrHandler.GetTrainingPrograms).Methods("GET")
	router.HandleFunc("/talent/{id}/training", hrHandler.AssignTraining).Methods("POST")
	router.HandleFunc("/talent/{id}/progress", hrHandler.GetTrainingProgress).Methods("GET")
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
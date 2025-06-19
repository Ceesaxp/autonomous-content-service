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
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
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
	clientRepo := database.NewClientRepository(db)
	projectRepo := database.NewProjectRepository(db)
	contentRepo := database.NewContentRepository(db)

	// Initialize legal compliance service
	legalService := legal_compliance.NewService(
		legalRepo,
		clientRepo,
		projectRepo,
		contentRepo,
		&mockSignatureProvider{},
		&mockComplianceEngine{},
		&mockIPAnalyzer{},
		&mockRegulatoryAPI{},
	)

	// Initialize handlers
	legalHandlers := handlers.NewLegalHandlers(legalService)

	// Set up router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Register legal handlers
	legalHandlers.RegisterRoutes(router)

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

// Mock implementations for external service dependencies (kept for now)
// Note: Repository mocks have been replaced with real database implementations

type mockSignatureProvider struct{}

func (m *mockSignatureProvider) CreateSignature(ctx context.Context, request legal_compliance.SignatureCreationRequest) (*legal_compliance.SignatureData, error) {
	return &legal_compliance.SignatureData{
		Data:      "mock-signature",
		Hash:      "mock-hash",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla",
		Timestamp: time.Now(),
	}, nil
}

type mockComplianceEngine struct{}

func (m *mockComplianceEngine) CheckCompliance(ctx context.Context, regulation string, data interface{}) (*legal_compliance.ComplianceResult, error) {
	return &legal_compliance.ComplianceResult{
		IsCompliant: true,
		Issues:      []string{},
		Score:       0.99,
	}, nil
}

type mockIPAnalyzer struct{}

func (m *mockIPAnalyzer) AnalyzeContent(ctx context.Context, content string) (*legal_compliance.IPAnalysisResult, error) {
	return &legal_compliance.IPAnalysisResult{
		NoViolations:    true,
		UsagePermitted:  true,
		Restrictions:    []string{},
		RequiredActions: []string{},
		License:         &entities.IPLicense{},
	}, nil
}

type mockRegulatoryAPI struct{}

func (m *mockRegulatoryAPI) SubmitReport(ctx context.Context, report *entities.RegulatoryReport) (*legal_compliance.SubmissionResult, error) {
	return &legal_compliance.SubmissionResult{
		Success:        true,
		ConfirmationID: "mock-submission-123",
		Message:        "submitted",
	}, nil
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

package api

import (
	"net/http"

	"github.com/Ceesaxp/autonomous-content-service/src/api/handlers"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes for the service
func SetupRoutes(router *mux.Router, contentHandler *handlers.ContentHandler, projectHandler *handlers.ProjectHandler, onboardingHandler *handlers.OnboardingHandler) {
	// Create web handler
	webHandler := handlers.NewWebHandler(projectHandler, contentHandler)

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()
	
	// Project endpoints
	apiV1.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST")
	apiV1.HandleFunc("/projects", projectHandler.ListProjects).Methods("GET")
	apiV1.HandleFunc("/projects/{projectId}", projectHandler.GetProject).Methods("GET")
	apiV1.HandleFunc("/projects/{projectId}", projectHandler.UpdateProject).Methods("PUT")
	apiV1.HandleFunc("/projects/{projectId}", projectHandler.CancelProject).Methods("DELETE")

	// Content endpoints
	apiV1.HandleFunc("/projects/{projectId}/content", contentHandler.CreateContent).Methods("POST")
	apiV1.HandleFunc("/content/{contentId}", contentHandler.GetContent).Methods("GET")
	apiV1.HandleFunc("/content/{contentId}", contentHandler.UpdateContent).Methods("PUT")
	apiV1.HandleFunc("/content/{contentId}/versions", contentHandler.GetContentVersions).Methods("GET")
	apiV1.HandleFunc("/content/{contentId}/approve", contentHandler.ApproveContent).Methods("POST")

	// Web interface endpoints
	apiV1.HandleFunc("/quote", webHandler.RequestQuote).Methods("POST")
	apiV1.HandleFunc("/chat", webHandler.HandleChat).Methods("POST")
	apiV1.HandleFunc("/analytics", webHandler.TrackAnalytics).Methods("POST")
	apiV1.HandleFunc("/portfolio", webHandler.GetPortfolio).Methods("GET")
	apiV1.HandleFunc("/pricing", webHandler.GetPricing).Methods("GET")
	apiV1.HandleFunc("/status", webHandler.GetSystemStatus).Methods("GET")

	// Onboarding endpoints
	if onboardingHandler != nil {
		apiV1.HandleFunc("/onboarding/start", onboardingHandler.StartOnboarding).Methods("POST")
		apiV1.HandleFunc("/onboarding/message", onboardingHandler.ProcessMessage).Methods("POST")
		apiV1.HandleFunc("/onboarding/responses", onboardingHandler.SubmitResponses).Methods("POST")
		apiV1.HandleFunc("/onboarding/session/{sessionId}", onboardingHandler.GetSession).Methods("GET")
		apiV1.HandleFunc("/onboarding/questions/{sessionId}", onboardingHandler.GetQuestions).Methods("GET")
		apiV1.HandleFunc("/onboarding/complete/{sessionId}", onboardingHandler.CompleteOnboarding).Methods("POST")
		apiV1.HandleFunc("/onboarding/progress/{sessionId}", onboardingHandler.GetOnboardingProgress).Methods("GET")
		apiV1.HandleFunc("/onboarding/analytics", onboardingHandler.GetOnboardingAnalytics).Methods("GET")
		apiV1.HandleFunc("/onboarding/health", onboardingHandler.HealthCheck).Methods("GET")
	}

	// CORS middleware for web requests
	router.Use(corsMiddleware)
}

// corsMiddleware adds CORS headers for web interface
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

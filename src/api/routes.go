package api

import (
	"net/http"

	"github.com/Ceesaxp/autonomous-content-service/src/api/handlers"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes for the service
func SetupRoutes(router *mux.Router, contentHandler *handlers.ContentHandler, projectHandler *handlers.ProjectHandler, onboardingHandler *handlers.OnboardingHandler, dashboardHandler *handlers.DashboardHandlers) {
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

	// Dashboard endpoints
	if dashboardHandler != nil {
		// Dashboard summary
		apiV1.HandleFunc("/dashboard/summary/{clientId}", dashboardHandler.GetDashboardSummary).Methods("GET")
		
		// Projects
		apiV1.HandleFunc("/dashboard/projects/{clientId}", dashboardHandler.GetProjectsOverview).Methods("GET")
		apiV1.HandleFunc("/dashboard/projects/details/{projectId}", dashboardHandler.GetProjectDetails).Methods("GET")
		apiV1.HandleFunc("/dashboard/projects/{projectId}/status", dashboardHandler.UpdateProjectStatus).Methods("PUT")
		
		// Content approvals
		apiV1.HandleFunc("/dashboard/approvals/{clientId}", dashboardHandler.GetContentApprovals).Methods("GET")
		apiV1.HandleFunc("/dashboard/approvals/{approvalId}/approve", dashboardHandler.ApproveContent).Methods("PUT")
		apiV1.HandleFunc("/dashboard/approvals/{approvalId}/reject", dashboardHandler.RejectContent).Methods("PUT")
		apiV1.HandleFunc("/dashboard/approvals/{approvalId}/revision", dashboardHandler.RequestContentRevision).Methods("PUT")
		
		// Messages
		apiV1.HandleFunc("/dashboard/messages/{clientId}", dashboardHandler.GetMessageThreads).Methods("GET")
		apiV1.HandleFunc("/dashboard/messages/threads", dashboardHandler.CreateMessageThread).Methods("POST")
		apiV1.HandleFunc("/dashboard/messages/{threadId}/send", dashboardHandler.SendMessage).Methods("POST")
		apiV1.HandleFunc("/dashboard/messages/{threadId}/messages", dashboardHandler.GetThreadMessages).Methods("GET")
		apiV1.HandleFunc("/dashboard/messages/{threadId}/read", dashboardHandler.MarkMessagesAsRead).Methods("PUT")
		
		// Notifications
		apiV1.HandleFunc("/dashboard/notifications/{clientId}", dashboardHandler.GetNotifications).Methods("GET")
		apiV1.HandleFunc("/dashboard/notifications/{notificationId}/read", dashboardHandler.MarkNotificationAsRead).Methods("PUT")
		
		// Analytics
		apiV1.HandleFunc("/dashboard/analytics/{clientId}", dashboardHandler.GetClientAnalytics).Methods("GET")
		apiV1.HandleFunc("/dashboard/reports/{clientId}", dashboardHandler.GenerateReport).Methods("POST")
		
		// Billing
		apiV1.HandleFunc("/dashboard/billing/{clientId}", dashboardHandler.GetBillingHistory).Methods("GET")
		apiV1.HandleFunc("/dashboard/billing/{clientId}/outstanding", dashboardHandler.GetOutstandingInvoices).Methods("GET")
	}

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

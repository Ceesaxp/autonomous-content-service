package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	
	"github.com/Ceesaxp/autonomous-content-service/src/services/onboarding"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// OnboardingHandler handles HTTP requests for client onboarding
type OnboardingHandler struct {
	onboardingService onboarding.OnboardingService
}

// NewOnboardingHandler creates a new onboarding handler
func NewOnboardingHandler(service onboarding.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingService: service,
	}
}

// StartOnboardingRequest represents the request to start onboarding
type StartOnboardingRequest struct {
	ClientID string `json:"clientId,omitempty"`
}

// StartOnboardingResponse represents the response when starting onboarding
type StartOnboardingResponse struct {
	SessionID string                       `json:"sessionId"`
	Session   *entities.OnboardingSession  `json:"session"`
	Questions []onboarding.Question        `json:"questions"`
	Message   string                       `json:"message"`
}

// MessageRequest represents a client message in the onboarding flow
type MessageRequest struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// QuestionResponse represents a structured response to onboarding questions
type QuestionResponse struct {
	SessionID string                 `json:"sessionId"`
	Responses map[string]interface{} `json:"responses"`
}

// StartOnboarding initiates a new onboarding session
func (h *OnboardingHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	var req StartOnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Generate or use provided client ID
	var clientID uuid.UUID
	var err error
	
	if req.ClientID != "" {
		clientID, err = uuid.Parse(req.ClientID)
		if err != nil {
			// If invalid UUID, generate a new one
			clientID = uuid.New()
		}
	} else {
		clientID = uuid.New()
	}
	
	// Start onboarding session
	session, err := h.onboardingService.StartOnboarding(r.Context(), clientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start onboarding: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get initial questions
	questions, err := h.onboardingService.GetNextQuestions(r.Context(), session.SessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get questions: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := StartOnboardingResponse{
		SessionID: session.SessionID.String(),
		Session:   session,
		Questions: questions,
		Message:   "Welcome! Let's get started with your personalized content strategy.",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProcessMessage handles client messages in the onboarding conversation
func (h *OnboardingHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	// Process the message
	response, err := h.onboardingService.ProcessMessage(r.Context(), sessionID, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process message: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SubmitResponses handles structured responses to onboarding questions
func (h *OnboardingHandler) SubmitResponses(w http.ResponseWriter, r *http.Request) {
	var req QuestionResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	// Validate and process each response
	for key, value := range req.Responses {
		err := h.onboardingService.ValidateResponse(r.Context(), sessionID, key, value)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid response for %s: %v", key, err), http.StatusBadRequest)
			return
		}
	}
	
	// Get updated session
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get session: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Store responses
	for key, value := range req.Responses {
		session.AddResponse(key, value)
	}
	
	// Update session
	if err := h.onboardingService.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get next questions
	questions, err := h.onboardingService.GetNextQuestions(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get next questions: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success":   true,
		"stage":     session.Stage,
		"questions": questions,
		"message":   "Responses received successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSession retrieves an onboarding session
func (h *OnboardingHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// GetQuestions retrieves questions for the current stage
func (h *OnboardingHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	questions, err := h.onboardingService.GetNextQuestions(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get questions: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// CompleteOnboarding finalizes the onboarding process
func (h *OnboardingHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	profile, err := h.onboardingService.CompleteOnboarding(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to complete onboarding: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Onboarding completed successfully",
		"profile": profile,
		"nextSteps": []string{
			"Review your personalized content strategy",
			"Start your first project",
			"Access your client dashboard",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOnboardingProgress returns the progress of an onboarding session
func (h *OnboardingHandler) GetOnboardingProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
		return
	}
	
	// Calculate progress
	stages := []entities.OnboardingStage{
		entities.StageInitial,
		entities.StageIndustry,
		entities.StageGoals,
		entities.StageAudience,
		entities.StageStyle,
		entities.StageBrand,
		entities.StageCompetitors,
		entities.StageWelcome,
		entities.StageComplete,
	}
	
	currentIndex := 0
	for i, stage := range stages {
		if stage == session.Stage {
			currentIndex = i
			break
		}
	}
	
	progress := float64(currentIndex) / float64(len(stages)-1) * 100
	
	response := map[string]interface{}{
		"sessionId":      session.SessionID,
		"currentStage":   session.Stage,
		"progress":       progress,
		"completedAt":    session.CompletedAt,
		"totalStages":    len(stages),
		"currentStageIndex": currentIndex + 1,
		"isComplete":     session.Stage == entities.StageComplete,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOnboardingAnalytics returns analytics for onboarding sessions
func (h *OnboardingHandler) GetOnboardingAnalytics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	days := 30 // default
	if d := query.Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}
	
	// In a real implementation, this would query the repository for analytics
	// For now, return mock data
	analytics := map[string]interface{}{
		"period": fmt.Sprintf("Last %d days", days),
		"totalSessions": 150,
		"completedSessions": 120,
		"completionRate": 80.0,
		"averageCompletionTime": "12 minutes",
		"stageDropoff": map[string]int{
			"initial":     5,
			"industry":    3,
			"goals":       2,
			"audience":    4,
			"style":       2,
			"brand":       3,
			"competitors": 1,
			"welcome":     0,
		},
		"topIndustries": []map[string]interface{}{
			{"name": "Technology", "count": 45},
			{"name": "Healthcare", "count": 23},
			{"name": "Finance", "count": 18},
			{"name": "E-commerce", "count": 15},
			{"name": "Education", "count": 12},
		},
		"commonGoals": []map[string]interface{}{
			{"name": "Brand Awareness", "count": 89},
			{"name": "Lead Generation", "count": 76},
			{"name": "SEO Traffic", "count": 65},
			{"name": "Customer Retention", "count": 54},
			{"name": "Thought Leadership", "count": 43},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// Health check endpoint for onboarding service
func (h *OnboardingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"service": "onboarding",
		"status":  "healthy",
		"version": "1.0.0",
		"features": []string{
			"conversational_flow",
			"industry_analysis",
			"competitive_analysis",
			"brand_voice_extraction",
			"progress_tracking",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
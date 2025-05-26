package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/services/decision_making"
	"github.com/gorilla/mux"
)

// DecisionHandlers handles HTTP requests for decision-making operations
type DecisionHandlers struct {
	decisionService *decision_making.Service
}

// NewDecisionHandlers creates a new decision handlers instance
func NewDecisionHandlers(decisionService *decision_making.Service) *DecisionHandlers {
	return &DecisionHandlers{
		decisionService: decisionService,
	}
}

// CreateDecision handles POST /api/v1/decisions
func (h *DecisionHandlers) CreateDecision(w http.ResponseWriter, r *http.Request) {
	var request decision_making.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Type == "" || request.Title == "" {
		http.Error(w, "Type and title are required", http.StatusBadRequest)
		return
	}

	// Set default priority if not provided
	if request.Priority == "" {
		request.Priority = entities.PriorityMedium
	}

	// Create decision
	decision, err := h.decisionService.MakeDecision(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(decision)
}

// GetDecision handles GET /api/v1/decisions/{id}
func (h *DecisionHandlers) GetDecision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decisionID := vars["id"]

	decision, err := h.decisionService.GetDecision(r.Context(), decisionID)
	if err != nil {
		http.Error(w, "Decision not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// ListDecisions handles GET /api/v1/decisions
func (h *DecisionHandlers) ListDecisions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := repositories.DecisionFilter{}

	if decisionType := r.URL.Query().Get("type"); decisionType != "" {
		dt := entities.DecisionType(decisionType)
		filter.Type = &dt
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := entities.DecisionStatus(status)
		filter.Status = &s
	}

	if priority := r.URL.Query().Get("priority"); priority != "" {
		p := entities.DecisionPriority(priority)
		filter.Priority = &p
	}

	decisions, err := h.decisionService.ListDecisions(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decisions)
}

// ExecuteDecision handles POST /api/v1/decisions/{id}/execute
func (h *DecisionHandlers) ExecuteDecision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decisionID := vars["id"]

	result, err := h.decisionService.ExecuteDecision(r.Context(), decisionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// OverrideDecision handles POST /api/v1/decisions/{id}/override
func (h *DecisionHandlers) OverrideDecision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decisionID := vars["id"]

	var request struct {
		Reason       string `json:"reason"`
		AuthorizedBy string `json:"authorized_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Reason == "" || request.AuthorizedBy == "" {
		http.Error(w, "Reason and authorized_by are required", http.StatusBadRequest)
		return
	}

	err := h.decisionService.OverrideDecision(r.Context(), decisionID, request.Reason, request.AuthorizedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Decision overridden successfully",
	})
}

// AssessDecisionQuality handles GET /api/v1/decisions/{id}/quality
func (h *DecisionHandlers) AssessDecisionQuality(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decisionID := vars["id"]

	report, err := h.decisionService.AssessDecisionQuality(r.Context(), decisionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetDecisionLogs handles GET /api/v1/decisions/{id}/logs
func (h *DecisionHandlers) GetDecisionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decisionID := vars["id"]

	logs, err := h.decisionService.GetDecisionLogs(r.Context(), decisionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetDecisionMetrics handles GET /api/v1/decisions/metrics
func (h *DecisionHandlers) GetDecisionMetrics(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d" // Default to 7 days
	}

	metrics, err := h.decisionService.GetDecisionMetrics(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// CreatePolicy handles POST /api/v1/policies
func (h *DecisionHandlers) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var policy entities.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.decisionService.RegisterPolicy(r.Context(), &policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

// GetPolicies handles GET /api/v1/policies
func (h *DecisionHandlers) GetPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.decisionService.GetActivePolicies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// CreateEthicalGuideline handles POST /api/v1/ethical-guidelines
func (h *DecisionHandlers) CreateEthicalGuideline(w http.ResponseWriter, r *http.Request) {
	var guideline entities.EthicalGuideline
	if err := json.NewDecoder(r.Body).Decode(&guideline); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.decisionService.RegisterEthicalGuideline(r.Context(), &guideline); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(guideline)
}

// GetEthicalGuidelines handles GET /api/v1/ethical-guidelines
func (h *DecisionHandlers) GetEthicalGuidelines(w http.ResponseWriter, r *http.Request) {
	guidelines, err := h.decisionService.GetEthicalGuidelines(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guidelines)
}

// GetSystemHealth handles GET /api/v1/system/health
func (h *DecisionHandlers) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.decisionService.CheckSystemHealth(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// ActivateEmergencyMode handles POST /api/v1/system/emergency
func (h *DecisionHandlers) ActivateEmergencyMode(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Reason == "" {
		http.Error(w, "Reason is required", http.StatusBadRequest)
		return
	}

	err := h.decisionService.ActivateEmergencyMode(r.Context(), request.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Emergency mode activated",
		"reason":  request.Reason,
	})
}

// GetAuditTrail handles GET /api/v1/audit
func (h *DecisionHandlers) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	// Parse time range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startTime, endTime time.Time
	var err error

	if startStr != "" {
		startTime, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7) // Default to 7 days ago
	}

	if endStr != "" {
		endTime, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "Invalid end time format", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now()
	}

	logs, err := h.decisionService.GetAuditTrail(r.Context(), startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// RegisterRoutes registers all decision-related routes
func (h *DecisionHandlers) RegisterRoutes(router *mux.Router) {
	// Decision endpoints
	router.HandleFunc("/api/v1/decisions", h.CreateDecision).Methods("POST")
	router.HandleFunc("/api/v1/decisions", h.ListDecisions).Methods("GET")
	router.HandleFunc("/api/v1/decisions/{id}", h.GetDecision).Methods("GET")
	router.HandleFunc("/api/v1/decisions/{id}/execute", h.ExecuteDecision).Methods("POST")
	router.HandleFunc("/api/v1/decisions/{id}/override", h.OverrideDecision).Methods("POST")
	router.HandleFunc("/api/v1/decisions/{id}/quality", h.AssessDecisionQuality).Methods("GET")
	router.HandleFunc("/api/v1/decisions/{id}/logs", h.GetDecisionLogs).Methods("GET")
	router.HandleFunc("/api/v1/decisions/metrics", h.GetDecisionMetrics).Methods("GET")

	// Policy endpoints
	router.HandleFunc("/api/v1/policies", h.CreatePolicy).Methods("POST")
	router.HandleFunc("/api/v1/policies", h.GetPolicies).Methods("GET")

	// Ethical guidelines endpoints
	router.HandleFunc("/api/v1/ethical-guidelines", h.CreateEthicalGuideline).Methods("POST")
	router.HandleFunc("/api/v1/ethical-guidelines", h.GetEthicalGuidelines).Methods("GET")

	// System endpoints
	router.HandleFunc("/api/v1/system/health", h.GetSystemHealth).Methods("GET")
	router.HandleFunc("/api/v1/system/emergency", h.ActivateEmergencyMode).Methods("POST")

	// Audit endpoints
	router.HandleFunc("/api/v1/audit", h.GetAuditTrail).Methods("GET")
}

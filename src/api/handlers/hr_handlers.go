package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/services/hr_management"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HRHandlers handles HTTP requests for HR operations
type HRHandlers struct {
	hrService *hr_management.HRService
}

// NewHRHandlers creates a new HR handlers instance
func NewHRHandlers(hrService *hr_management.HRService) *HRHandlers {
	return &HRHandlers{
		hrService: hrService,
	}
}

// RegisterRoutes registers all HR routes
func (h *HRHandlers) RegisterRoutes(router *mux.Router) {
	// Talent management
	router.HandleFunc("/api/v1/talent", h.CreateTalent).Methods("POST")
	router.HandleFunc("/api/v1/talent", h.ListTalent).Methods("GET")
	router.HandleFunc("/api/v1/talent/{id}", h.GetTalent).Methods("GET")
	router.HandleFunc("/api/v1/talent/{id}", h.UpdateTalent).Methods("PUT")
	router.HandleFunc("/api/v1/talent/{id}/status", h.UpdateTalentStatus).Methods("PUT")
	
	// Job postings and applications
	router.HandleFunc("/api/v1/job-postings", h.CreateJobPosting).Methods("POST")
	router.HandleFunc("/api/v1/job-postings", h.ListJobPostings).Methods("GET")
	router.HandleFunc("/api/v1/job-postings/{id}/applications", h.GetApplications).Methods("GET")
	router.HandleFunc("/api/v1/applications/{id}/screen", h.ScreenApplication).Methods("POST")
	
	// Engagements and assignments
	router.HandleFunc("/api/v1/engagements", h.CreateEngagement).Methods("POST")
	router.HandleFunc("/api/v1/engagements", h.ListEngagements).Methods("GET")
	router.HandleFunc("/api/v1/engagements/{id}/assignments", h.CreateAssignment).Methods("POST")
	router.HandleFunc("/api/v1/assignments/{id}/complete", h.CompleteAssignment).Methods("POST")
	
	// Performance and reviews
	router.HandleFunc("/api/v1/talent/{id}/performance", h.GetPerformanceMetrics).Methods("GET")
	router.HandleFunc("/api/v1/talent/{id}/reviews", h.CreateReview).Methods("POST")
	router.HandleFunc("/api/v1/talent/{id}/compensation", h.GetCompensationPlan).Methods("GET")
	
	// Training and development
	router.HandleFunc("/api/v1/training-programs", h.ListTrainingPrograms).Methods("GET")
	router.HandleFunc("/api/v1/talent/{id}/training", h.AssignTraining).Methods("POST")
	router.HandleFunc("/api/v1/talent/{id}/progress", h.GetTrainingProgress).Methods("GET")
}

// CreateTalent creates a new talent profile
func (h *HRHandlers) CreateTalent(w http.ResponseWriter, r *http.Request) {
	var request hr_management.TalentCreationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	talent, err := h.hrService.CreateTalent(context.Background(), request)
	if err != nil {
		http.Error(w, "Failed to create talent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(talent)
}

// ListTalent lists all available talent
func (h *HRHandlers) ListTalent(w http.ResponseWriter, r *http.Request) {
	talentType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	talents := []map[string]interface{}{
		{
			"id":               uuid.New().String(),
			"type":             "Human",
			"name":             "Alice Johnson",
			"email":            "alice@example.com",
			"status":           "Available",
			"reputation_score": 4.8,
			"hourly_rate":      75.00,
			"currency":         "USD",
			"location":         "New York, NY",
			"skills":           []string{"Content Writing", "SEO", "Research"},
			"created_at":       time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			"id":               uuid.New().String(),
			"type":             "AI",
			"name":             "GPT-4 Writer",
			"status":           "Available",
			"reputation_score": 4.5,
			"cost_per_request": 0.02,
			"currency":         "USD",
			"capabilities":     []string{"Content Generation", "Proofreading"},
			"created_at":       time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	// Apply filters
	if talentType != "" || status != "" {
		filtered := []map[string]interface{}{}
		for _, talent := range talents {
			if talentType != "" && talent["type"] != talentType {
				continue
			}
			if status != "" && talent["status"] != status {
				continue
			}
			filtered = append(filtered, talent)
		}
		talents = filtered
	}

	if limit < len(talents) {
		talents = talents[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"talents": talents,
		"total":   len(talents),
		"limit":   limit,
	})
}

// GetTalent gets a specific talent profile
func (h *HRHandlers) GetTalent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid talent ID", http.StatusBadRequest)
		return
	}

	talent, err := h.hrService.GetTalent(context.Background(), id)
	if err != nil {
		http.Error(w, "Failed to get talent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(talent)
}

// UpdateTalent updates talent information
func (h *HRHandlers) UpdateTalent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid talent ID", http.StatusBadRequest)
		return
	}

	var updates hr_management.TalentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	talent, err := h.hrService.UpdateTalent(context.Background(), id, updates)
	if err != nil {
		http.Error(w, "Failed to update talent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(talent)
}

// UpdateTalentStatus updates talent status
func (h *HRHandlers) UpdateTalentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var request struct {
		Status entities.TalentStatus `json:"status"`
		Reason string                `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"talent_id":  id,
		"old_status": "Available",
		"new_status": request.Status,
		"reason":     request.Reason,
		"updated_at": time.Now(),
	})
}

// CreateJobPosting creates a new job posting
func (h *HRHandlers) CreateJobPosting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title       string                 `json:"title"`
		Description string                 `json:"description"`
		Skills      []string               `json:"skills"`
		Budget      float64                `json:"budget"`
		Duration    string                 `json:"duration"`
		Requirements map[string]interface{} `json:"requirements"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	jobPosting := map[string]interface{}{
		"job_id":      uuid.New().String(),
		"title":       request.Title,
		"description": request.Description,
		"skills":      request.Skills,
		"budget":      request.Budget,
		"duration":    request.Duration,
		"requirements": request.Requirements,
		"status":      "Active",
		"created_at":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(jobPosting)
}

// ListJobPostings lists job postings
func (h *HRHandlers) ListJobPostings(w http.ResponseWriter, r *http.Request) {
	jobPostings := []map[string]interface{}{
		{
			"job_id":             uuid.New().String(),
			"title":              "Senior Content Writer",
			"description":        "We need an experienced content writer",
			"skills":             []string{"Technical Writing", "Documentation"},
			"budget":             5000,
			"duration":           "2 weeks",
			"status":             "Active",
			"applications_count": 12,
			"created_at":         time.Now().Add(-3 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_postings": jobPostings,
		"total":        len(jobPostings),
	})
}

// GetApplications gets applications for a job posting
func (h *HRHandlers) GetApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	applications := []map[string]interface{}{
		{
			"application_id": uuid.New().String(),
			"job_id":         jobID,
			"talent_id":      uuid.New().String(),
			"talent_name":    "Alice Johnson",
			"status":         "New",
			"proposed_rate":  75.00,
			"submitted_at":   time.Now().Add(-2 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applications": applications,
		"total":        len(applications),
		"job_id":       jobID,
	})
}

// ScreenApplication screens an application
func (h *HRHandlers) ScreenApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	var request struct {
		Status   string  `json:"status"`
		Notes    string  `json:"notes"`
		Score    float64 `json:"score,omitempty"`
		Feedback string  `json:"feedback,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"application_id": applicationID,
		"old_status":     "New",
		"new_status":     request.Status,
		"notes":          request.Notes,
		"score":          request.Score,
		"screened_at":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateEngagement creates a new engagement
func (h *HRHandlers) CreateEngagement(w http.ResponseWriter, r *http.Request) {
	var request hr_management.EngagementCreationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	engagement, err := h.hrService.CreateEngagement(context.Background(), request)
	if err != nil {
		http.Error(w, "Failed to create engagement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(engagement)
}

// ListEngagements lists engagements
func (h *HRHandlers) ListEngagements(w http.ResponseWriter, r *http.Request) {
	engagements := []map[string]interface{}{
		{
			"engagement_id": uuid.New().String(),
			"talent_name":   "Alice Johnson",
			"project_name":  "Website Content Refresh",
			"type":          "Contract",
			"status":        "Active",
			"rate":          75.00,
			"start_date":    time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"engagements": engagements,
		"total":       len(engagements),
	})
}

// CreateAssignment creates a new assignment
func (h *HRHandlers) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	engagementIDStr := vars["id"]

	engagementID, err := uuid.Parse(engagementIDStr)
	if err != nil {
		http.Error(w, "Invalid engagement ID", http.StatusBadRequest)
		return
	}

	var request hr_management.AssignmentCreationRequest
	request.EngagementID = engagementID

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	assignment, err := h.hrService.CreateWorkAssignment(context.Background(), request)
	if err != nil {
		http.Error(w, "Failed to create assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

// CompleteAssignment marks an assignment as complete
func (h *HRHandlers) CompleteAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assignmentIDStr := vars["id"]

	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	var request struct {
		ActualHours  float64 `json:"actual_hours"`
		QualityScore float64 `json:"quality_score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.hrService.CompleteAssignment(context.Background(), assignmentID, request.ActualHours, request.QualityScore)
	if err != nil {
		http.Error(w, "Failed to complete assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assignment_id": assignmentID,
		"status":        "completed",
		"completed_at":  time.Now(),
	})
}

// GetPerformanceMetrics gets performance metrics for a talent
func (h *HRHandlers) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	talentID := vars["id"]

	metrics := map[string]interface{}{
		"talent_id": talentID,
		"performance_summary": map[string]interface{}{
			"projects_completed":   42,
			"total_revenue":        31500.00,
			"average_rating":       4.8,
			"on_time_delivery":     95.2,
			"client_satisfaction": 4.7,
		},
		"recent_performance": []map[string]interface{}{
			{
				"project_name":     "Blog Content Series",
				"completion_date":  time.Now().Add(-5 * 24 * time.Hour),
				"rating":           5.0,
				"revenue":          1200.00,
				"time_to_complete": "3 days",
			},
		},
		"calculated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// CreateReview creates a performance review
func (h *HRHandlers) CreateReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	talentID := vars["id"]

	var request struct {
		ReviewerID      string                 `json:"reviewer_id"`
		ProjectID       string                 `json:"project_id"`
		OverallRating   string                 `json:"overall_rating"`
		SkillRatings    map[string]float64     `json:"skill_ratings"`
		Feedback        string                 `json:"feedback"`
		Recommendations []string               `json:"recommendations"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	review := map[string]interface{}{
		"review_id":        uuid.New().String(),
		"talent_id":        talentID,
		"reviewer_id":      request.ReviewerID,
		"project_id":       request.ProjectID,
		"overall_rating":   request.OverallRating,
		"skill_ratings":    request.SkillRatings,
		"feedback":         request.Feedback,
		"recommendations":  request.Recommendations,
		"review_date":      time.Now(),
		"status":           "Completed",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

// GetCompensationPlan gets compensation plan for a talent
func (h *HRHandlers) GetCompensationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	talentID := vars["id"]

	plan := map[string]interface{}{
		"talent_id": talentID,
		"base_compensation": map[string]interface{}{
			"hourly_rate":    75.00,
			"currency":       "USD",
			"rate_type":      "hourly",
			"effective_date": time.Now().Add(-30 * 24 * time.Hour),
		},
		"performance_bonuses": []map[string]interface{}{
			{
				"trigger":      "Monthly high performance",
				"bonus_amount": 500.00,
				"frequency":    "monthly",
			},
		},
		"benefits": []string{
			"Professional development budget",
			"Flexible schedule",
			"Remote work",
		},
		"payment_terms": map[string]interface{}{
			"payment_frequency": "bi-weekly",
			"payment_method":    "direct_deposit",
			"invoice_terms":     "Net 15",
		},
		"updated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// ListTrainingPrograms lists available training programs
func (h *HRHandlers) ListTrainingPrograms(w http.ResponseWriter, r *http.Request) {
	programs := []map[string]interface{}{
		{
			"program_id":      uuid.New().String(),
			"title":           "Advanced SEO Techniques",
			"description":     "Learn cutting-edge SEO strategies",
			"category":        "Marketing",
			"duration":        "4 weeks",
			"difficulty":      "Advanced",
			"cost":            299.00,
			"provider":        "SEO Institute",
			"rating":          4.7,
			"enrollments":     245,
		},
		{
			"program_id":      uuid.New().String(),
			"title":           "Technical Writing Fundamentals",
			"description":     "Master clear, concise technical documentation",
			"category":        "Technical",
			"duration":        "6 weeks",
			"difficulty":      "Intermediate",
			"cost":            199.00,
			"provider":        "TechWrite Academy",
			"rating":          4.9,
			"enrollments":     532,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"training_programs": programs,
		"total":             len(programs),
		"categories":        []string{"Marketing", "Technical", "Technology"},
	})
}

// AssignTraining assigns training to a talent
func (h *HRHandlers) AssignTraining(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	talentID := vars["id"]

	var request struct {
		ProgramID   string     `json:"program_id"`
		Priority    string     `json:"priority"`
		Deadline    *time.Time `json:"deadline,omitempty"`
		CompanyPaid bool       `json:"company_paid"`
		Notes       string     `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	assignment := map[string]interface{}{
		"assignment_id": uuid.New().String(),
		"talent_id":     talentID,
		"program_id":    request.ProgramID,
		"status":        "Assigned",
		"priority":      request.Priority,
		"deadline":      request.Deadline,
		"company_paid":  request.CompanyPaid,
		"notes":         request.Notes,
		"assigned_at":   time.Now(),
		"progress":      0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

// GetTrainingProgress gets training progress for a talent
func (h *HRHandlers) GetTrainingProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	talentID := vars["id"]

	progress := map[string]interface{}{
		"talent_id": talentID,
		"active_trainings": []map[string]interface{}{
			{
				"assignment_id":       uuid.New().String(),
				"program_title":       "Advanced SEO Techniques",
				"status":              "InProgress",
				"progress":            65.0,
				"modules_completed":   5,
				"modules_total":       8,
				"deadline":            time.Now().Add(14 * 24 * time.Hour),
				"last_activity":       time.Now().Add(-2 * 24 * time.Hour),
			},
		},
		"completed_trainings": []map[string]interface{}{
			{
				"assignment_id":   uuid.New().String(),
				"program_title":   "Content Writing Essentials",
				"status":          "Completed",
				"completion_date": time.Now().Add(-30 * 24 * time.Hour),
				"final_score":     92.5,
				"certificate_url": "https://certificates.example.com/cert123",
			},
		},
		"summary": map[string]interface{}{
			"total_hours_completed":    31.5,
			"programs_completed":       3,
			"programs_in_progress":     1,
			"average_completion_score": 89.2,
			"certification_count":      2,
		},
		"updated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}
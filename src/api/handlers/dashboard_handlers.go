package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/services/dashboard"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DashboardHandlers handles dashboard-related HTTP requests
type DashboardHandlers struct {
	dashboardService dashboard.DashboardService
}

// NewDashboardHandlers creates a new dashboard handlers instance
func NewDashboardHandlers(dashboardService dashboard.DashboardService) *DashboardHandlers {
	return &DashboardHandlers{
		dashboardService: dashboardService,
	}
}

// GetDashboardSummary handles GET /api/v1/dashboard/summary/{clientId}
func (h *DashboardHandlers) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	summary, err := h.dashboardService.GetDashboardSummary(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Failed to get dashboard summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetProjectsOverview handles GET /api/v1/dashboard/projects/{clientId}
func (h *DashboardHandlers) GetProjectsOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	projects, err := h.dashboardService.GetProjectsOverview(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Failed to get projects overview", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// GetProjectDetails handles GET /api/v1/dashboard/projects/details/{projectId}
func (h *DashboardHandlers) GetProjectDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	details, err := h.dashboardService.GetProjectDetails(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Failed to get project details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// UpdateProjectStatus handles PUT /api/v1/dashboard/projects/{projectId}/status
func (h *DashboardHandlers) UpdateProjectStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Status entities.ProjectStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.UpdateProjectStatus(r.Context(), projectID, request.Status); err != nil {
		http.Error(w, "Failed to update project status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetContentApprovals handles GET /api/v1/dashboard/approvals/{clientId}
func (h *DashboardHandlers) GetContentApprovals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	limit, offset := h.getPaginationParams(r)
	approvals, err := h.dashboardService.GetContentApprovals(r.Context(), clientID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get content approvals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(approvals)
}

// ApproveContent handles PUT /api/v1/dashboard/approvals/{approvalId}/approve
func (h *DashboardHandlers) ApproveContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["approvalId"])
	if err != nil {
		http.Error(w, "Invalid approval ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Feedback string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.ApproveContent(r.Context(), approvalID, request.Feedback); err != nil {
		http.Error(w, "Failed to approve content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectContent handles PUT /api/v1/dashboard/approvals/{approvalId}/reject
func (h *DashboardHandlers) RejectContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["approvalId"])
	if err != nil {
		http.Error(w, "Invalid approval ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Feedback string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.RejectContent(r.Context(), approvalID, request.Feedback); err != nil {
		http.Error(w, "Failed to reject content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RequestContentRevision handles PUT /api/v1/dashboard/approvals/{approvalId}/revision
func (h *DashboardHandlers) RequestContentRevision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["approvalId"])
	if err != nil {
		http.Error(w, "Invalid approval ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Feedback string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.RequestContentRevision(r.Context(), approvalID, request.Feedback); err != nil {
		http.Error(w, "Failed to request content revision", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMessageThreads handles GET /api/v1/dashboard/messages/{clientId}
func (h *DashboardHandlers) GetMessageThreads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	limit, offset := h.getPaginationParams(r)
	threads, err := h.dashboardService.GetMessageThreads(r.Context(), clientID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get message threads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(threads)
}

// CreateMessageThread handles POST /api/v1/dashboard/messages/threads
func (h *DashboardHandlers) CreateMessageThread(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProjectID uuid.UUID `json:"projectId"`
		ClientID  uuid.UUID `json:"clientId"`
		Subject   string    `json:"subject"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	thread, err := h.dashboardService.CreateMessageThread(r.Context(), request.ProjectID, request.ClientID, request.Subject)
	if err != nil {
		http.Error(w, "Failed to create message thread", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thread)
}

// SendMessage handles POST /api/v1/dashboard/messages/{threadId}/send
func (h *DashboardHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["threadId"])
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Type    entities.MessageType `json:"type"`
		Content string               `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message, err := h.dashboardService.SendMessage(r.Context(), threadID, request.Type, request.Content)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetThreadMessages handles GET /api/v1/dashboard/messages/{threadId}/messages
func (h *DashboardHandlers) GetThreadMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["threadId"])
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	limit, offset := h.getPaginationParams(r)
	messages, err := h.dashboardService.GetThreadMessages(r.Context(), threadID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get thread messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// MarkMessagesAsRead handles PUT /api/v1/dashboard/messages/{threadId}/read
func (h *DashboardHandlers) MarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["threadId"])
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.MarkMessagesAsRead(r.Context(), threadID); err != nil {
		http.Error(w, "Failed to mark messages as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetNotifications handles GET /api/v1/dashboard/notifications/{clientId}
func (h *DashboardHandlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	limit, offset := h.getPaginationParams(r)
	notifications, err := h.dashboardService.GetNotifications(r.Context(), clientID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// MarkNotificationAsRead handles PUT /api/v1/dashboard/notifications/{notificationId}/read
func (h *DashboardHandlers) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["notificationId"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.dashboardService.MarkNotificationAsRead(r.Context(), notificationID); err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetClientAnalytics handles GET /api/v1/dashboard/analytics/{clientId}
func (h *DashboardHandlers) GetClientAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Parse date range parameters
	fromDateStr := r.URL.Query().Get("from")
	toDateStr := r.URL.Query().Get("to")
	
	fromDate := time.Now().AddDate(0, -6, 0) // Default to 6 months ago
	toDate := time.Now()

	if fromDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			fromDate = parsed
		}
	}

	if toDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", toDateStr); err == nil {
			toDate = parsed
		}
	}

	analytics, err := h.dashboardService.GetClientAnalytics(r.Context(), clientID, fromDate, toDate)
	if err != nil {
		http.Error(w, "Failed to get client analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GenerateReport handles POST /api/v1/dashboard/reports/{clientId}
func (h *DashboardHandlers) GenerateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Type dashboard.ReportType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.dashboardService.GeneratePerformanceReport(r.Context(), clientID, request.Type)
	if err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// GetBillingHistory handles GET /api/v1/dashboard/billing/{clientId}
func (h *DashboardHandlers) GetBillingHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	limit, offset := h.getPaginationParams(r)
	history, err := h.dashboardService.GetBillingHistory(r.Context(), clientID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get billing history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetOutstandingInvoices handles GET /api/v1/dashboard/billing/{clientId}/outstanding
func (h *DashboardHandlers) GetOutstandingInvoices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	invoices, err := h.dashboardService.GetOutstandingInvoices(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Failed to get outstanding invoices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

// Helper function to get pagination parameters
func (h *DashboardHandlers) getPaginationParams(r *http.Request) (limit, offset int) {
	limit = 20 // default limit
	offset = 0 // default offset

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
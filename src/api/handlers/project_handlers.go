package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	ProjectRepository repositories.ProjectRepository
	ContentRepository repositories.ContentRepository
	ClientRepository  repositories.ClientRepository
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectRepo repositories.ProjectRepository, contentRepo repositories.ContentRepository, clientRepo repositories.ClientRepository) *ProjectHandler {
	return &ProjectHandler{
		ProjectRepository: projectRepo,
		ContentRepository: contentRepo,
		ClientRepository:  clientRepo,
	}
}

// ProjectRequest represents a request to create a new project
type ProjectRequest struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	ContentType entities.ContentType `json:"contentType"`
	Deadline    string                `json:"deadline"`
	Budget      MoneyRequest          `json:"budget"`
	Priority    entities.Priority     `json:"priority,omitempty"`
}

// MoneyRequest represents a monetary amount in API requests
type MoneyRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ProjectID   string                 `json:"projectId"`
	ClientID    string                 `json:"clientId"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	ContentType entities.ContentType   `json:"contentType"`
	Deadline    string                 `json:"deadline"`
	Budget      MoneyResponse          `json:"budget"`
	Priority    entities.Priority      `json:"priority"`
	Status      entities.ProjectStatus `json:"status"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	Content     []ContentSummary       `json:"content,omitempty"`
}

// MoneyResponse represents a monetary amount in API responses
type MoneyResponse struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ContentSummary represents a summary of content in API responses
type ContentSummary struct {
	ContentID string                 `json:"contentId"`
	Title     string                 `json:"title"`
	Type      entities.ContentType   `json:"type"`
	Status    entities.ContentStatus `json:"status"`
	Version   int                    `json:"version"`
	WordCount int                    `json:"wordCount"`
	CreatedAt string                 `json:"createdAt"`
}

// ProjectUpdateRequest represents a request to update a project
type ProjectUpdateRequest struct {
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Deadline    string                 `json:"deadline,omitempty"`
	Budget      *MoneyRequest          `json:"budget,omitempty"`
	Priority    entities.Priority      `json:"priority,omitempty"`
	Status      entities.ProjectStatus `json:"status,omitempty"`
}

// ProjectSummaryResponse represents a project summary in list responses
type ProjectSummaryResponse struct {
	ProjectID   string                 `json:"projectId"`
	ClientID    string                 `json:"clientId"`
	Title       string                 `json:"title"`
	ContentType entities.ContentType   `json:"contentType"`
	Deadline    string                 `json:"deadline"`
	Status      entities.ProjectStatus `json:"status"`
	CreatedAt   string                 `json:"createdAt"`
}

// PaginationResponse represents pagination information in list responses
type PaginationResponse struct {
	TotalItems   int `json:"totalItems"`
	TotalPages   int `json:"totalPages"`
	CurrentPage  int `json:"currentPage"`
	ItemsPerPage int `json:"itemsPerPage"`
}

// ListProjectsResponse represents the response for listing projects
type ListProjectsResponse struct {
	Data       []ProjectSummaryResponse `json:"data"`
	Pagination PaginationResponse      `json:"pagination"`
}

// CreateProject handles requests to create a new project
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Get client ID from authentication context (in a real implementation)
	// For simplicity, we'll use a query parameter
	clientIDStr := r.URL.Query().Get("clientId")
	if clientIDStr == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Check if client exists
	_, err = h.ClientRepository.FindByID(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Decode request body
	var req ProjectRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	if req.Deadline == "" {
		http.Error(w, "Deadline is required", http.StatusBadRequest)
		return
	}

	// Parse deadline
	deadline, err := time.Parse("2006-01-02T15:04:05Z07:00", req.Deadline)
	if err != nil {
		http.Error(w, "Invalid deadline format", http.StatusBadRequest)
		return
	}

	// Create budget
	budget := entities.Money{
		Amount:   req.Budget.Amount,
		Currency: req.Budget.Currency,
	}

	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = entities.PriorityMedium
	}

	// Create project
	project, err := entities.NewProject(clientID, req.Title, req.Description, req.ContentType, deadline, budget)
	if err != nil {
		http.Error(w, "Failed to create project: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set priority
	project.UpdatePriority(priority)

	// Save project
	err = h.ProjectRepository.Create(r.Context(), project)
	if err != nil {
		http.Error(w, "Failed to save project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	res := ProjectResponse{
		ProjectID:   project.ProjectID.String(),
		ClientID:    project.ClientID.String(),
		Title:       project.Title,
		Description: project.Description,
		ContentType: project.ContentType,
		Deadline:    project.Deadline.Format("2006-01-02T15:04:05Z07:00"),
		Budget: MoneyResponse{
			Amount:   project.Budget.Amount,
			Currency: project.Budget.Currency,
		},
		Priority:  project.Priority,
		Status:    project.Status,
		CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:   []ContentSummary{},
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// GetProject handles requests to retrieve project details
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Retrieve project
	project, err := h.ProjectRepository.FindByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Retrieve project content
	content, err := h.ContentRepository.FindByProjectID(r.Context(), projectID)
	if err != nil {
		// Log error but continue
		content = []*entities.Content{}
	}

	// Prepare content summaries
	contentSummaries := []ContentSummary{}
	for _, item := range content {
		contentSummaries = append(contentSummaries, ContentSummary{
			ContentID: item.ContentID.String(),
			Title:     item.Title,
			Type:      item.Type,
			Status:    item.Status,
			Version:   item.Version,
			WordCount: item.WordCount,
			CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Prepare response
	res := ProjectResponse{
		ProjectID:   project.ProjectID.String(),
		ClientID:    project.ClientID.String(),
		Title:       project.Title,
		Description: project.Description,
		ContentType: project.ContentType,
		Deadline:    project.Deadline.Format("2006-01-02T15:04:05Z07:00"),
		Budget: MoneyResponse{
			Amount:   project.Budget.Amount,
			Currency: project.Budget.Currency,
		},
		Priority:  project.Priority,
		Status:    project.Status,
		CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:   contentSummaries,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// UpdateProject handles requests to update a project
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req ProjectUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve project
	project, err := h.ProjectRepository.FindByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Apply updates
	if req.Title != "" {
		project.Title = req.Title
	}

	if req.Description != "" {
		project.Description = req.Description
	}

	if req.Deadline != "" {
		deadline, err := time.Parse("2006-01-02T15:04:05Z07:00", req.Deadline)
		if err != nil {
			http.Error(w, "Invalid deadline format", http.StatusBadRequest)
			return
		}
		err = project.UpdateDeadline(deadline)
		if err != nil {
			http.Error(w, "Invalid deadline: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Budget != nil {
		project.Budget = entities.Money{
			Amount:   req.Budget.Amount,
			Currency: req.Budget.Currency,
		}
	}

	if req.Priority != "" {
		project.UpdatePriority(req.Priority)
	}

	if req.Status != "" {
		project.UpdateStatus(req.Status)
	}

	// Save updates
	err = h.ProjectRepository.Update(r.Context(), project)
	if err != nil {
		http.Error(w, "Failed to update project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve project content
	content, err := h.ContentRepository.FindByProjectID(r.Context(), projectID)
	if err != nil {
		// Log error but continue
		content = []*entities.Content{}
	}

	// Prepare content summaries
	contentSummaries := []ContentSummary{}
	for _, item := range content {
		contentSummaries = append(contentSummaries, ContentSummary{
			ContentID: item.ContentID.String(),
			Title:     item.Title,
			Type:      item.Type,
			Status:    item.Status,
			Version:   item.Version,
			WordCount: item.WordCount,
			CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Prepare response
	res := ProjectResponse{
		ProjectID:   project.ProjectID.String(),
		ClientID:    project.ClientID.String(),
		Title:       project.Title,
		Description: project.Description,
		ContentType: project.ContentType,
		Deadline:    project.Deadline.Format("2006-01-02T15:04:05Z07:00"),
		Budget: MoneyResponse{
			Amount:   project.Budget.Amount,
			Currency: project.Budget.Currency,
		},
		Priority:  project.Priority,
		Status:    project.Status,
		CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:   contentSummaries,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// ListProjects handles requests to list projects with filtering
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	clientIDStr := r.URL.Query().Get("clientId")
	statusStr := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Set defaults for pagination
	page := 1
	limit := 20

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Query projects based on filters
	var projects []*entities.Project
	var total int
	var err error

	if clientIDStr != "" {
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}
		projects, total, err = h.ProjectRepository.FindByClientID(r.Context(), clientID, offset, limit)
	} else if statusStr != "" {
		status := entities.ProjectStatus(statusStr)
		projects, total, err = h.ProjectRepository.FindByStatus(r.Context(), status, offset, limit)
	} else {
		projects, total, err = h.ProjectRepository.FindAll(r.Context(), offset, limit)
	}

	if err != nil {
		http.Error(w, "Failed to retrieve projects: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	data := []ProjectSummaryResponse{}
	for _, project := range projects {
		data = append(data, ProjectSummaryResponse{
			ProjectID:   project.ProjectID.String(),
			ClientID:    project.ClientID.String(),
			Title:       project.Title,
			ContentType: project.ContentType,
			Deadline:    project.Deadline.Format("2006-01-02T15:04:05Z07:00"),
			Status:      project.Status,
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	response := ListProjectsResponse{
		Data: data,
		Pagination: PaginationResponse{
			TotalItems:   total,
			TotalPages:   totalPages,
			CurrentPage:  page,
			ItemsPerPage: limit,
		},
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelProject handles requests to cancel a project
func (h *ProjectHandler) CancelProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Retrieve project
	project, err := h.ProjectRepository.FindByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Check if project can be cancelled
	if project.Status == entities.ProjectStatusCompleted || project.Status == entities.ProjectStatusCancelled {
		http.Error(w, "Project cannot be cancelled in its current state", http.StatusConflict)
		return
	}

	// Update project status
	project.UpdateStatus(entities.ProjectStatusCancelled)

	// Save updates
	err = h.ProjectRepository.Update(r.Context(), project)
	if err != nil {
		http.Error(w, "Failed to cancel project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}

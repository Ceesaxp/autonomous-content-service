package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/services/content_creation"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	ContentRepository  repositories.ContentRepository
	ProjectRepository  repositories.ProjectRepository
	FeedbackRepository repositories.FeedbackRepository
	ContentPipeline    *content_creation.ContentPipeline
}

// NewContentHandler creates a new content handler
func NewContentHandler(contentRepo repositories.ContentRepository, projectRepo repositories.ProjectRepository, feedbackRepo repositories.FeedbackRepository, contentPipeline *content_creation.ContentPipeline) *ContentHandler {
	return &ContentHandler{
		ContentRepository:  contentRepo,
		ProjectRepository:  projectRepo,
		FeedbackRepository: feedbackRepo,
		ContentPipeline:    contentPipeline,
	}
}

// ContentRequest represents a request to create new content
type ContentRequest struct {
	Title    string                 `json:"title"`
	Type     entities.ContentType   `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ContentResponse represents a content response
type ContentResponse struct {
	ContentID  string                 `json:"contentId"`
	ProjectID  string                 `json:"projectId"`
	Title      string                 `json:"title"`
	Type       entities.ContentType   `json:"type"`
	Status     entities.ContentStatus `json:"status"`
	Data       string                 `json:"data,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Version    int                    `json:"version"`
	WordCount  int                    `json:"wordCount"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
	Statistics *StatisticsResponse    `json:"statistics,omitempty"`
}

// StatisticsResponse represents content statistics in API responses
type StatisticsResponse struct {
	ReadabilityScore float64 `json:"readabilityScore"`
	SEOScore         float64 `json:"seoScore"`
	EngagementScore  float64 `json:"engagementScore"`
	PlagiarismScore  float64 `json:"plagiarismScore"`
}

// ContentVersionResponse represents a content version response
type ContentVersionResponse struct {
	VersionID     string                 `json:"versionId"`
	ContentID     string                 `json:"contentId"`
	VersionNumber int                    `json:"versionNumber"`
	Data          string                 `json:"data,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     string                 `json:"createdAt"`
	CreatedBy     string                 `json:"createdBy"`
}

// ContentUpdateRequest represents a request to update content
type ContentUpdateRequest struct {
	Title    string                 `json:"title,omitempty"`
	Data     string                 `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Status   entities.ContentStatus `json:"status,omitempty"`
}

// CreateContent handles content creation requests
func (h *ContentHandler) CreateContent(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req ContentRequest
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

	// Check if project exists
	_, err = h.ProjectRepository.FindByID(r.Context(), projectID) // project
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Create content through the pipeline
	content, err := h.ContentPipeline.CreateContent(r.Context(), projectID, req.Title, req.Type)
	if err != nil {
		http.Error(w, "Failed to create content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add any additional metadata from the request
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			content.UpdateMetadata(k, v)
		}
		h.ContentRepository.Update(r.Context(), content)
	}

	// Prepare response
	res := ContentResponse{
		ContentID: content.ContentID.String(),
		ProjectID: content.ProjectID.String(),
		Title:     content.Title,
		Type:      content.Type,
		Status:    content.Status,
		Data:      content.Data,
		Metadata:  content.Metadata,
		Version:   content.Version,
		WordCount: content.WordCount,
		CreatedAt: content.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: content.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add statistics if available
	if content.Statistics != nil {
		res.Statistics = &StatisticsResponse{
			ReadabilityScore: content.Statistics.ReadabilityScore,
			SEOScore:         content.Statistics.SEOScore,
			EngagementScore:  content.Statistics.EngagementScore,
			PlagiarismScore:  content.Statistics.PlagiarismScore,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// GetContent handles requests to retrieve content details
func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	// Extract content ID from URL
	vars := mux.Vars(r)
	contentID, err := uuid.Parse(vars["contentId"])
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	// Retrieve content
	content, err := h.ContentRepository.FindByID(r.Context(), contentID)
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Prepare response
	res := ContentResponse{
		ContentID: content.ContentID.String(),
		ProjectID: content.ProjectID.String(),
		Title:     content.Title,
		Type:      content.Type,
		Status:    content.Status,
		Data:      content.Data,
		Metadata:  content.Metadata,
		Version:   content.Version,
		WordCount: content.WordCount,
		CreatedAt: content.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: content.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add statistics if available
	if content.Statistics != nil {
		res.Statistics = &StatisticsResponse{
			ReadabilityScore: content.Statistics.ReadabilityScore,
			SEOScore:         content.Statistics.SEOScore,
			EngagementScore:  content.Statistics.EngagementScore,
			PlagiarismScore:  content.Statistics.PlagiarismScore,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// UpdateContent handles requests to update content
func (h *ContentHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	// Extract content ID from URL
	vars := mux.Vars(r)
	contentID, err := uuid.Parse(vars["contentId"])
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req ContentUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve content
	content, err := h.ContentRepository.FindByID(r.Context(), contentID)
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Apply updates
	if req.Title != "" {
		content.Title = req.Title
	}

	if req.Data != "" {
		err = content.UpdateContent(req.Data, "Client")
		if err != nil {
			http.Error(w, "Failed to update content: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Status != "" {
		content.UpdateStatus(req.Status)
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			content.UpdateMetadata(k, v)
		}
	}

	// Save updates
	err = h.ContentRepository.Update(r.Context(), content)
	if err != nil {
		http.Error(w, "Failed to save content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	res := ContentResponse{
		ContentID: content.ContentID.String(),
		ProjectID: content.ProjectID.String(),
		Title:     content.Title,
		Type:      content.Type,
		Status:    content.Status,
		Data:      content.Data,
		Metadata:  content.Metadata,
		Version:   content.Version,
		WordCount: content.WordCount,
		CreatedAt: content.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: content.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if content.Statistics != nil {
		res.Statistics = &StatisticsResponse{
			ReadabilityScore: content.Statistics.ReadabilityScore,
			SEOScore:         content.Statistics.SEOScore,
			EngagementScore:  content.Statistics.EngagementScore,
			PlagiarismScore:  content.Statistics.PlagiarismScore,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// GetContentVersions handles requests to retrieve content version history
func (h *ContentHandler) GetContentVersions(w http.ResponseWriter, r *http.Request) {
	// Extract content ID from URL
	vars := mux.Vars(r)
	contentID, err := uuid.Parse(vars["contentId"])
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	// Check if content exists
	_, err = h.ContentRepository.FindByID(r.Context(), contentID)
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Retrieve content versions
	versions, err := h.ContentPipeline.ContentVersionRepo.FindByContentID(r.Context(), contentID)
	if err != nil {
		http.Error(w, "Failed to retrieve content versions", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := []ContentVersionResponse{}
	for _, version := range versions {
		response = append(response, ContentVersionResponse{
			VersionID:     version.VersionID.String(),
			ContentID:     version.ContentID.String(),
			VersionNumber: version.VersionNumber,
			Data:          version.Data,
			Metadata:      version.Metadata,
			CreatedAt:     version.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:     version.CreatedBy,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApproveContent handles requests to approve content
func (h *ContentHandler) ApproveContent(w http.ResponseWriter, r *http.Request) {
	// Extract content ID from URL
	vars := mux.Vars(r)
	contentID, err := uuid.Parse(vars["contentId"])
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	// Retrieve content
	content, err := h.ContentRepository.FindByID(r.Context(), contentID)
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Check if content can be approved
	if content.Status != entities.ContentStatusReview {
		http.Error(w, "Content must be in Review status to be approved", http.StatusConflict)
		return
	}

	// Update content status
	originalStatus := content.Status
	log.Printf("Updating content status from %v to %v", originalStatus, content.Status)
	content.UpdateStatus(entities.ContentStatusApproved)

	// Save updates
	err = h.ContentRepository.Update(r.Context(), content)
	if err != nil {
		http.Error(w, "Failed to approve content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	res := ContentResponse{
		ContentID: content.ContentID.String(),
		ProjectID: content.ProjectID.String(),
		Title:     content.Title,
		Type:      content.Type,
		Status:    content.Status,
		Data:      content.Data,
		Metadata:  content.Metadata,
		Version:   content.Version,
		WordCount: content.WordCount,
		CreatedAt: content.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: content.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if content.Statistics != nil {
		res.Statistics = &StatisticsResponse{
			ReadabilityScore: content.Statistics.ReadabilityScore,
			SEOScore:         content.Statistics.SEOScore,
			EngagementScore:  content.Statistics.EngagementScore,
			PlagiarismScore:  content.Statistics.PlagiarismScore,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

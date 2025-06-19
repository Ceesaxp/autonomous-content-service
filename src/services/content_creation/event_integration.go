package content_creation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/events"
	"github.com/google/uuid"
)

// EventIntegratedContentService wraps the content pipeline with event-driven capabilities
type EventIntegratedContentService struct {
	pipeline     *ContentPipeline
	eventBus     *events.ServiceEventBus
	contentRepo  repositories.ContentRepository
	projectRepo  repositories.ProjectRepository
}

// NewEventIntegratedContentService creates a new event-integrated content service
func NewEventIntegratedContentService(
	pipeline *ContentPipeline,
	eventBus *events.ServiceEventBus,
	contentRepo repositories.ContentRepository,
	projectRepo repositories.ProjectRepository,
) *EventIntegratedContentService {
	return &EventIntegratedContentService{
		pipeline:    pipeline,
		eventBus:    eventBus,
		contentRepo: contentRepo,
		projectRepo: projectRepo,
	}
}

// HandleContentCreationRequest handles incoming content creation requests via events
func (s *EventIntegratedContentService) HandleContentCreationRequest(ctx context.Context, event events.Event) error {
	log.Printf("[EventIntegratedContentService] Handling content creation request")

	// Extract request data from event payload
	projectID, ok := event.Payload["project_id"].(string)
	if !ok {
		return fmt.Errorf("missing project_id in event payload")
	}

	contentType, ok := event.Payload["content_type"].(string)
	if !ok {
		contentType = "article" // Default
	}

	// Load project details
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Convert content type string to ContentType enum
	var contentTypeEnum entities.ContentType
	switch contentType {
	case "article":
		contentTypeEnum = entities.ContentTypeTechnicalArticle
	case "blog_post":
		contentTypeEnum = entities.ContentTypeBlogPost
	case "newsletter":
		contentTypeEnum = entities.ContentTypeEmailNewsletter
	case "case_study":
		contentTypeEnum = entities.ContentTypeTechnicalArticle
	default:
		contentTypeEnum = entities.ContentTypeBlogPost // Default
	}

	// Execute content creation pipeline
	content, err := s.pipeline.CreateContent(ctx, projectUUID, project.Title, contentTypeEnum)
	if err != nil {
		// Publish failure event
		s.publishContentEvent(ctx, events.EventContentRejected, projectID, "", err.Error())
		return fmt.Errorf("content creation failed: %w", err)
	}

	// Content is already saved by the pipeline, just publish success event
	s.publishContentCreatedEvent(ctx, content)

	return nil
}

// HandleContentApprovalRequest handles content approval requests
func (s *EventIntegratedContentService) HandleContentApprovalRequest(ctx context.Context, event events.Event) error {
	contentID, ok := event.Payload["content_id"].(string)
	if !ok {
		return fmt.Errorf("missing content_id in event payload")
	}

	contentUUID, err := uuid.Parse(contentID)
	if err != nil {
		return fmt.Errorf("invalid content_id: %w", err)
	}

	// Load content
	content, err := s.contentRepo.FindByID(ctx, contentUUID)
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}

	// Simulate quality check (in real implementation, this would use the quality pipeline)
	qualityScore := 0.85 // Simulated quality score
	
	if qualityScore < 0.7 {
		// Publish rejection event
		s.publishContentEvent(ctx, events.EventContentRejected, content.ProjectID.String(), contentID, 
			fmt.Sprintf("Quality score too low: %.2f", qualityScore))
		return nil
	}

	// Update content status
	content.Status = entities.ContentStatusApproved
	content.UpdatedAt = time.Now()

	err = s.contentRepo.Update(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	// Publish approval event
	s.publishContentApprovedEvent(ctx, content)

	return nil
}

// HandleProjectUpdate handles project update events
func (s *EventIntegratedContentService) HandleProjectUpdate(ctx context.Context, event events.Event) error {
	projectID, ok := event.Payload["project_id"].(string)
	if !ok {
		return fmt.Errorf("missing project_id in event payload")
	}

	status, ok := event.Payload["status"].(string)
	if !ok {
		return nil // No status change
	}

	// If project is cancelled, cancel all pending content
	if status == "cancelled" {
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return fmt.Errorf("invalid project_id: %w", err)
		}

		contents, err := s.contentRepo.FindByProjectID(ctx, projectUUID)
		if err != nil {
			return fmt.Errorf("failed to list project content: %w", err)
		}

		for _, content := range contents {
			if content.Status == entities.ContentStatusDrafting || content.Status == entities.ContentStatusReview {
				content.Status = entities.ContentStatusArchived
				content.UpdatedAt = time.Now()
				_ = s.contentRepo.Update(ctx, content) // Best effort update
				
				// Publish archival event
				s.publishContentEvent(ctx, events.EventContentArchived, projectID, content.ContentID.String(), "Project cancelled")
			}
		}
	}

	return nil
}

// Publishing helper methods

func (s *EventIntegratedContentService) publishContentCreatedEvent(ctx context.Context, content *entities.Content) {
	eventData := events.ContentEventData{
		ContentID:    content.ContentID.String(),
		ProjectID:    content.ProjectID.String(),
		ClientID:     "", // Should be loaded from project
		ContentType:  string(content.Type),
		Title:        content.Title,
		Status:       string(content.Status),
		WordCount:    content.WordCount,
		QualityScore: 0.0, // Will be set during quality evaluation
		Metadata:     content.Metadata,
	}

	event := events.CreateContentEvent(events.EventContentCreated, "content-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[EventIntegratedContentService] Failed to publish content created event: %v", err)
	}
}

func (s *EventIntegratedContentService) publishContentApprovedEvent(ctx context.Context, content *entities.Content) {
	eventData := events.ContentEventData{
		ContentID:    content.ContentID.String(),
		ProjectID:    content.ProjectID.String(),
		ClientID:     "", // Should be loaded from project
		ContentType:  string(content.Type),
		Title:        content.Title,
		Status:       string(content.Status),
		QualityScore: 0.85, // Simulated quality score
		Metadata:     content.Metadata,
	}

	event := events.CreateContentEvent(events.EventContentApproved, "content-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[EventIntegratedContentService] Failed to publish content approved event: %v", err)
	}
}

func (s *EventIntegratedContentService) publishContentEvent(ctx context.Context, eventType, projectID, contentID, message string) {
	eventData := map[string]interface{}{
		"project_id": projectID,
		"content_id": contentID,
		"message":    message,
		"timestamp":  time.Now().UTC(),
	}

	if err := s.eventBus.PublishEvent(ctx, eventType, eventData); err != nil {
		log.Printf("[EventIntegratedContentService] Failed to publish event %s: %v", eventType, err)
	}
}

// StartEventListeners starts listening for relevant events
func (s *EventIntegratedContentService) StartEventListeners(ctx context.Context) error {
	// The event bus will handle subscription based on registered handlers
	log.Printf("[EventIntegratedContentService] Event listeners started")
	return nil
}

// HandleClientFeedback processes client feedback to improve content quality
func (s *EventIntegratedContentService) HandleClientFeedback(ctx context.Context, event events.Event) error {
	clientID, ok := event.Payload["client_id"].(string)
	if !ok {
		return fmt.Errorf("missing client_id in event payload")
	}

	satisfaction, ok := event.Payload["satisfaction"].(float64)
	if !ok {
		return nil // No satisfaction score
	}

	log.Printf("[EventIntegratedContentService] Processing client feedback for %s (satisfaction: %.2f)", clientID, satisfaction)

	// TODO: Update content creation parameters based on feedback
	// - Adjust quality thresholds
	// - Update style preferences
	// - Modify tone and voice settings

	return nil
}

// HandleDecisionExecuted processes executed decisions that affect content creation
func (s *EventIntegratedContentService) HandleDecisionExecuted(ctx context.Context, event events.Event) error {
	decisionType, ok := event.Payload["type"].(string)
	if !ok {
		return nil
	}

	switch decisionType {
	case "content_policy":
		// Update content creation policies
		log.Printf("[EventIntegratedContentService] Updating content policies based on decision")
		// TODO: Implement policy updates
		
	case "quality_threshold":
		// Adjust quality thresholds
		log.Printf("[EventIntegratedContentService] Adjusting quality thresholds based on decision")
		// TODO: Implement threshold adjustments
		
	case "content_strategy":
		// Update content strategy
		log.Printf("[EventIntegratedContentService] Updating content strategy based on decision")
		// TODO: Implement strategy updates
	}

	return nil
}

// Pipeline stage progress monitoring methods would go here if needed

// Workflow integration

// StartContentWorkflow starts a content creation workflow
func (s *EventIntegratedContentService) StartContentWorkflow(ctx context.Context, projectID, contentType string) error {
	workflow, err := s.eventBus.StartWorkflow(ctx, "content_creation", map[string]interface{}{
		"project_id":   projectID,
		"content_type": contentType,
	})
	
	if err != nil {
		return fmt.Errorf("failed to start content workflow: %w", err)
	}

	log.Printf("[EventIntegratedContentService] Started content workflow %s for project %s", workflow.ID, projectID)
	return nil
}

// GetContentMetrics returns metrics about content creation
func (s *EventIntegratedContentService) GetContentMetrics() map[string]interface{} {
	// TODO: Implement actual metrics collection
	return map[string]interface{}{
		"total_content_created": 0,
		"average_quality_score": 0.0,
		"average_creation_time": 0,
		"approval_rate":         0.0,
	}
}
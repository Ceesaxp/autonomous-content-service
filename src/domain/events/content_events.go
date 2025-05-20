package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// Content event types
const (
	EventTypeContentRequested     EventType = "ContentRequested"
	EventTypeContentStageAdvanced EventType = "ContentStageAdvanced"
	EventTypeContentUpdated       EventType = "ContentUpdated"
	EventTypeContentApproved      EventType = "ContentApproved"
)

// ContentRequestedEvent is triggered when new content is requested
type ContentRequestedEvent struct {
	BaseEvent
	ContentID   uuid.UUID `json:"contentId"`
	ProjectID   uuid.UUID `json:"projectId"`
	Title       string    `json:"title"`
	ContentType entities.ContentType `json:"contentType"`
}

// NewContentRequestedEvent creates a new ContentRequestedEvent
func NewContentRequestedEvent(content *entities.Content) ContentRequestedEvent {
	return ContentRequestedEvent{
		BaseEvent:   NewBaseEvent(EventTypeContentRequested, content.ContentID),
		ContentID:   content.ContentID,
		ProjectID:   content.ProjectID,
		Title:       content.Title,
		ContentType: content.Type,
	}
}

// ContentStageAdvancedEvent is triggered when content moves to the next stage
type ContentStageAdvancedEvent struct {
	BaseEvent
	ContentID uuid.UUID              `json:"contentId"`
	OldStage  entities.ContentStatus `json:"oldStage"`
	NewStage  entities.ContentStatus `json:"newStage"`
}

// NewContentStageAdvancedEvent creates a new ContentStageAdvancedEvent
func NewContentStageAdvancedEvent(contentID uuid.UUID, oldStage, newStage entities.ContentStatus) ContentStageAdvancedEvent {
	return ContentStageAdvancedEvent{
		BaseEvent: NewBaseEvent(EventTypeContentStageAdvanced, contentID),
		ContentID: contentID,
		OldStage:  oldStage,
		NewStage:  newStage,
	}
}

// ContentUpdatedEvent is triggered when content is updated significantly
type ContentUpdatedEvent struct {
	BaseEvent
	ContentID     uuid.UUID `json:"contentId"`
	VersionNumber int       `json:"versionNumber"`
	WordCount     int       `json:"wordCount"`
}

// NewContentUpdatedEvent creates a new ContentUpdatedEvent
func NewContentUpdatedEvent(content *entities.Content) ContentUpdatedEvent {
	return ContentUpdatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeContentUpdated, content.ContentID),
		ContentID:     content.ContentID,
		VersionNumber: content.Version,
		WordCount:     content.WordCount,
	}
}

// ContentApprovedEvent is triggered when content is approved
type ContentApprovedEvent struct {
	BaseEvent
	ContentID       uuid.UUID `json:"contentId"`
	ProjectID       uuid.UUID `json:"projectId"`
	ApprovalTime    time.Time `json:"approvalTime"`
}

// NewContentApprovedEvent creates a new ContentApprovedEvent
func NewContentApprovedEvent(content *entities.Content) ContentApprovedEvent {
	return ContentApprovedEvent{
		BaseEvent:    NewBaseEvent(EventTypeContentApproved, content.ContentID),
		ContentID:    content.ContentID,
		ProjectID:    content.ProjectID,
		ApprovalTime: time.Now(),
	}
}

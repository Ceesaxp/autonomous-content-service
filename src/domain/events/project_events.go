package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// ProjectCreatedEvent is triggered when a new project is created
type ProjectCreatedEvent struct {
	BaseEvent
	ProjectID   uuid.UUID            `json:"projectId"`
	ClientID    uuid.UUID            `json:"clientId"`
	Title       string               `json:"title"`
	ContentType entities.ContentType `json:"contentType"`
}

// NewProjectCreatedEvent creates a new ProjectCreatedEvent
func NewProjectCreatedEvent(project *entities.Project) ProjectCreatedEvent {
	return ProjectCreatedEvent{
		BaseEvent:   *NewBaseEventWithID(EventTypeProjectCreated, project.ProjectID),
		ProjectID:   project.ProjectID,
		ClientID:    project.ClientID,
		Title:       project.Title,
		ContentType: project.ContentType,
	}
}

// ProjectStatusChangedEvent is triggered when project status changes
type ProjectStatusChangedEvent struct {
	BaseEvent
	ProjectID uuid.UUID              `json:"projectId"`
	OldStatus entities.ProjectStatus `json:"oldStatus"`
	NewStatus entities.ProjectStatus `json:"newStatus"`
}

// NewProjectStatusChangedEvent creates a new ProjectStatusChangedEvent
func NewProjectStatusChangedEvent(projectID uuid.UUID, oldStatus, newStatus entities.ProjectStatus) ProjectStatusChangedEvent {
	return ProjectStatusChangedEvent{
		BaseEvent: *NewBaseEventWithID(EventTypeProjectStatusChanged, projectID),
		ProjectID: projectID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}
}

// ProjectDeadlineApproachingEvent is triggered when a project's deadline is approaching
type ProjectDeadlineApproachingEvent struct {
	BaseEvent
	ProjectID     uuid.UUID `json:"projectId"`
	Deadline      time.Time `json:"deadline"`
	DaysRemaining int       `json:"daysRemaining"`
}

// NewProjectDeadlineApproachingEvent creates a new ProjectDeadlineApproachingEvent
func NewProjectDeadlineApproachingEvent(projectID uuid.UUID, deadline time.Time, daysRemaining int) ProjectDeadlineApproachingEvent {
	return ProjectDeadlineApproachingEvent{
		BaseEvent:     *NewBaseEventWithID(EventTypeProjectDeadlineApproaching, projectID),
		ProjectID:     projectID,
		Deadline:      deadline,
		DaysRemaining: daysRemaining,
	}
}

// ProjectCompletedEvent is triggered when a project is completed
type ProjectCompletedEvent struct {
	BaseEvent
	ProjectID      uuid.UUID `json:"projectId"`
	ClientID       uuid.UUID `json:"clientId"`
	CompletionTime time.Time `json:"completionTime"`
	ContentCount   int       `json:"contentCount"`
}

// NewProjectCompletedEvent creates a new ProjectCompletedEvent
func NewProjectCompletedEvent(project *entities.Project) ProjectCompletedEvent {
	return ProjectCompletedEvent{
		BaseEvent:      *NewBaseEventWithID(EventTypeProjectCompleted, project.ProjectID),
		ProjectID:      project.ProjectID,
		ClientID:       project.ClientID,
		CompletionTime: time.Now(),
		ContentCount:   len(project.Contents),
	}
}

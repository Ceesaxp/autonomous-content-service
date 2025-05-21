package events

import (
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// FeedbackReceivedEvent is triggered when new feedback is submitted
type FeedbackReceivedEvent struct {
	BaseEvent
	FeedbackID   uuid.UUID               `json:"feedbackId"`
	ContentID    *uuid.UUID              `json:"contentId,omitempty"`
	ProjectID    *uuid.UUID              `json:"projectId,omitempty"`
	Source       entities.FeedbackSource `json:"source"`
	FeedbackType entities.FeedbackType   `json:"feedbackType"`
	Score        *float64                `json:"score,omitempty"`
}

// NewFeedbackReceivedEvent creates a new FeedbackReceivedEvent
func NewFeedbackReceivedEvent(feedback *entities.Feedback) FeedbackReceivedEvent {
	return FeedbackReceivedEvent{
		BaseEvent:    NewBaseEvent(EventTypeFeedbackReceived, feedback.FeedbackID),
		FeedbackID:   feedback.FeedbackID,
		ContentID:    feedback.ContentID,
		ProjectID:    feedback.ProjectID,
		Source:       feedback.Source,
		FeedbackType: feedback.Type,
		Score:        feedback.Score,
	}
}

// RevisionRequestedEvent is triggered when a client requests a revision to content
type RevisionRequestedEvent struct {
	BaseEvent
	ContentID  uuid.UUID `json:"contentId"`
	FeedbackID uuid.UUID `json:"feedbackId"`
	ClientID   uuid.UUID `json:"clientId"`
	Details    string    `json:"details"`
}

// NewRevisionRequestedEvent creates a new RevisionRequestedEvent
func NewRevisionRequestedEvent(contentID, feedbackID, clientID uuid.UUID, details string) RevisionRequestedEvent {
	return RevisionRequestedEvent{
		BaseEvent:  NewBaseEvent(EventTypeRevisionRequested, contentID),
		ContentID:  contentID,
		FeedbackID: feedbackID,
		ClientID:   clientID,
		Details:    details,
	}
}

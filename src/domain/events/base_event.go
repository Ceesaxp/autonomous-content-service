package events

import (
	"time"

	"github.com/google/uuid"
)

// Event type constants
const (
	EventTypeClientRegistered              = "client.registered"
	EventTypeClientUpdated                 = "client.updated"
	EventTypeClientDeleted                 = "client.deleted"
	EventTypeClientProfileUpdated          = "client.profileUpdated"
	EventTypeClientStatusChanged           = "client.statusChanged"
	EventTypeContentCreated                = "content.created"
	EventTypeContentUpdated                = "content.updated"
	EventTypeContentPublished              = "content.published"
	EventTypeContentArchived               = "content.archived"
	EventTypeContentRequested              = "content.requested"
	EventTypeContentStageAdvanced          = "content.stageAdvanced"
	EventTypeContentApproved               = "content.approved"
	EventTypeFeedbackSubmitted             = "feedback.submitted"
	EventTypeFeedbackReceived              = "feedback.received"
	EventTypeRevisionRequested             = "revision.requested"
	EventTypePaymentProcessed              = "payment.processed"
	EventTypePaymentReceived               = "payment.received"
	EventTypeInvoiceGenerated              = "invoice,generated"
	EventTypePaymentFailed                 = "paymentf.ailed"
	EventTypeSystemStarted                 = "system.started"
	EventTypeCapabilityUpdated             = "capability.updated"
	EventTypeProjectCreated                = "project.created"
	EventTypeProjectStatusChanged          = "project.statusChanged"
	EventTypeProjectDeadlineApproaching    = "project.deadlineApproaching"
	EventTypeProjectCompleted              = "project.completed"
	EventTypeCapabilityPerformanceDeclined = "capability.performanceDeclined"
	EventTypeCapabilityUpgraded            = "capability.upgraded"
	EventTypeAnomalyDetected               = "anomaly.detected"
)

// Event represents the base event interface
type Event interface {
	GetID() uuid.UUID
	GetType() string
	GetTimestamp() time.Time
	GetData() map[string]interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	EventID   uuid.UUID              `json:"eventId"`
	EventType string                 `json:"eventType"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// GetID returns the event ID
func (e *BaseEvent) GetID() uuid.UUID {
	return e.EventID
}

// GetType returns the event type
func (e *BaseEvent) GetType() string {
	return e.EventType
}

// GetTimestamp returns the event timestamp
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetData returns the event data
func (e *BaseEvent) GetData() map[string]interface{} {
	return e.Data
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType string, data map[string]interface{}) *BaseEvent {
	return &BaseEvent{
		EventID:   uuid.New(),
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewBaseEventWithID creates a new base event with entity ID
func NewBaseEventWithID(eventType string, entityID uuid.UUID) *BaseEvent {
	return &BaseEvent{
		EventID:   uuid.New(),
		EventType: eventType,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"entityId": entityID,
		},
	}
}

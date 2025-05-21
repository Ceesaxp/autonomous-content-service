package events

import (
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// ClientRegisteredEvent is triggered when a new client registers
type ClientRegisteredEvent struct {
	BaseEvent
	ClientID uuid.UUID `json:"clientId"`
}

// NewClientRegisteredEvent creates a new ClientRegisteredEvent
func NewClientRegisteredEvent(clientID uuid.UUID) ClientRegisteredEvent {
	return ClientRegisteredEvent{
		BaseEvent: *NewBaseEventWithID(EventTypeClientRegistered, clientID),
		ClientID:  clientID,
	}
}

// ClientProfileUpdatedEvent is triggered when client profile information is modified
type ClientProfileUpdatedEvent struct {
	BaseEvent
	ClientID      uuid.UUID              `json:"clientId"`
	UpdatedFields map[string]interface{} `json:"updatedFields"`
}

// NewClientProfileUpdatedEvent creates a new ClientProfileUpdatedEvent
func NewClientProfileUpdatedEvent(clientID uuid.UUID, updatedFields map[string]interface{}) ClientProfileUpdatedEvent {
	return ClientProfileUpdatedEvent{
		BaseEvent:     *NewBaseEventWithID(EventTypeClientProfileUpdated, clientID),
		ClientID:      clientID,
		UpdatedFields: updatedFields,
	}
}

// ClientStatusChangedEvent is triggered when client status changes
type ClientStatusChangedEvent struct {
	BaseEvent
	ClientID  uuid.UUID             `json:"clientId"`
	OldStatus entities.ClientStatus `json:"oldStatus"`
	NewStatus entities.ClientStatus `json:"newStatus"`
	Reason    string                `json:"reason"`
}

// NewClientStatusChangedEvent creates a new ClientStatusChangedEvent
func NewClientStatusChangedEvent(clientID uuid.UUID, oldStatus, newStatus entities.ClientStatus, reason string) ClientStatusChangedEvent {
	return ClientStatusChangedEvent{
		BaseEvent: *NewBaseEventWithID(EventTypeClientStatusChanged, clientID),
		ClientID:  clientID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Reason:    reason,
	}
}

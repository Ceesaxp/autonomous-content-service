package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// PaymentReceivedEvent is triggered when payment is received from a client
type PaymentReceivedEvent struct {
	BaseEvent
	TransactionID uuid.UUID                  `json:"transactionId"`
	ClientID      uuid.UUID                  `json:"clientId"`
	Amount        entities.Money             `json:"amount"`
	PaymentMethod entities.PaymentMethodType `json:"paymentMethod"`
}

// NewPaymentReceivedEvent creates a new PaymentReceivedEvent
func NewPaymentReceivedEvent(transaction *entities.Transaction) PaymentReceivedEvent {
	return PaymentReceivedEvent{
		BaseEvent:     *NewBaseEventWithID(string(EventTypePaymentReceived), transaction.TransactionID),
		TransactionID: transaction.TransactionID,
		ClientID:      transaction.ClientID,
		Amount:        transaction.Amount,
		PaymentMethod: transaction.PaymentMethod,
	}
}

// InvoiceGeneratedEvent is triggered when a new invoice is created
type InvoiceGeneratedEvent struct {
	BaseEvent
	InvoiceID uuid.UUID      `json:"invoiceId"`
	ClientID  uuid.UUID      `json:"clientId"`
	ProjectID *uuid.UUID     `json:"projectId,omitempty"`
	Amount    entities.Money `json:"amount"`
	DueDate   time.Time      `json:"dueDate"`
}

// NewInvoiceGeneratedEvent creates a new InvoiceGeneratedEvent
func NewInvoiceGeneratedEvent(invoiceID, clientID uuid.UUID, projectID *uuid.UUID, amount entities.Money, dueDate time.Time) InvoiceGeneratedEvent {
	return InvoiceGeneratedEvent{
		BaseEvent: *NewBaseEventWithID(string(EventTypeInvoiceGenerated), invoiceID),
		InvoiceID: invoiceID,
		ClientID:  clientID,
		ProjectID: projectID,
		Amount:    amount,
		DueDate:   dueDate,
	}
}

// PaymentFailedEvent is triggered when a payment attempt fails
type PaymentFailedEvent struct {
	BaseEvent
	TransactionID uuid.UUID `json:"transactionId"`
	ClientID      uuid.UUID `json:"clientId"`
	Reason        string    `json:"reason"`
	RetryCount    int       `json:"retryCount"`
}

// NewPaymentFailedEvent creates a new PaymentFailedEvent
func NewPaymentFailedEvent(transaction *entities.Transaction, reason string, retryCount int) PaymentFailedEvent {
	return PaymentFailedEvent{
		BaseEvent:     *NewBaseEventWithID(string(EventTypePaymentFailed), transaction.TransactionID),
		TransactionID: transaction.TransactionID,
		ClientID:      transaction.ClientID,
		Reason:        reason,
		RetryCount:    retryCount,
	}
}

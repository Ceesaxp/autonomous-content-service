package entities

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethodType represents the type of payment method
type PaymentMethodType string

const (
	PaymentMethodCreditCard     PaymentMethodType = "CreditCard"
	PaymentMethodBankTransfer   PaymentMethodType = "BankTransfer"
	PaymentMethodPayPal         PaymentMethodType = "PayPal"
	PaymentMethodCryptocurrency PaymentMethodType = "Cryptocurrency"
	PaymentMethodStripe         PaymentMethodType = "Stripe"
	PaymentMethodSquare         PaymentMethodType = "Square"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "Pending"
	TransactionStatusCompleted TransactionStatus = "Completed"
	TransactionStatusFailed    TransactionStatus = "Failed"
	TransactionStatusCancelled TransactionStatus = "Cancelled"
	TransactionStatusRefunded  TransactionStatus = "Refunded"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypePayment TransactionType = "Payment"
	TransactionTypeRefund  TransactionType = "Refund"
	TransactionTypeFee     TransactionType = "Fee"
	TransactionTypeCredit  TransactionType = "Credit"
)

// Transaction represents a financial transaction
type Transaction struct {
	TransactionID   uuid.UUID         `json:"transactionId"`
	ClientID        uuid.UUID         `json:"clientId"`
	ProjectID       *uuid.UUID        `json:"projectId,omitempty"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	Amount          Money             `json:"amount"`
	PaymentMethod   PaymentMethodType `json:"paymentMethod"`
	PaymentReference string           `json:"paymentReference,omitempty"`
	Description     string            `json:"description"`
	ProcessedAt     *time.Time        `json:"processedAt,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewTransaction creates a new transaction
func NewTransaction(
	clientID uuid.UUID,
	transactionType TransactionType,
	amount Money,
	paymentMethod PaymentMethodType,
	description string,
) *Transaction {
	return &Transaction{
		TransactionID: uuid.New(),
		ClientID:      clientID,
		Type:          transactionType,
		Status:        TransactionStatusPending,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Description:   description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}
}

// UpdateStatus updates the transaction status
func (t *Transaction) UpdateStatus(status TransactionStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
	
	if status == TransactionStatusCompleted {
		now := time.Now()
		t.ProcessedAt = &now
	}
}

// SetPaymentReference sets the payment reference from the payment processor
func (t *Transaction) SetPaymentReference(reference string) {
	t.PaymentReference = reference
	t.UpdatedAt = time.Now()
}

// AttachToProject associates the transaction with a project
func (t *Transaction) AttachToProject(projectID uuid.UUID) {
	t.ProjectID = &projectID
	t.UpdatedAt = time.Now()
}

// AddMetadata adds metadata to the transaction
func (t *Transaction) AddMetadata(key string, value interface{}) {
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

// IsCompleted returns true if the transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// IsFailed returns true if the transaction failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}

// IsPending returns true if the transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}
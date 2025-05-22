package payment

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// PaymentService defines the main payment service interface
type PaymentService interface {
	// Payment processing
	ProcessPayment(ctx context.Context, request *PaymentRequest) (*entities.Payment, error)
	GetPayment(ctx context.Context, id string) (*entities.Payment, error)
	GetPaymentsByInvoice(ctx context.Context, invoiceID string) ([]*entities.Payment, error)
	CancelPayment(ctx context.Context, id string) error
	RetryFailedPayment(ctx context.Context, id string) error

	// Invoice management
	CreateInvoice(ctx context.Context, request *InvoiceRequest) (*entities.Invoice, error)
	GetInvoice(ctx context.Context, id string) (*entities.Invoice, error)
	SendInvoice(ctx context.Context, id string) error
	MarkInvoicePaid(ctx context.Context, id string, payment *entities.Payment) error
	CancelInvoice(ctx context.Context, id string) error

	// Refund processing
	ProcessRefund(ctx context.Context, request *RefundRequest) (*entities.Refund, error)
	GetRefund(ctx context.Context, id string) (*entities.Refund, error)
	GetRefundsByPayment(ctx context.Context, paymentID string) ([]*entities.Refund, error)

	// Fraud detection
	AnalyzeFraud(ctx context.Context, payment *entities.Payment) (*entities.FraudDetectionResult, error)

	// Notifications
	SendNotification(ctx context.Context, notification *NotificationRequest) error

	// Analytics
	GetPaymentStats(ctx context.Context, startDate, endDate time.Time) (*PaymentStats, error)
}

// PaymentProcessor defines the interface for individual payment processors
type PaymentProcessor interface {
	GetName() string
	GetSupportedMethods() []entities.PaymentMethod
	ProcessPayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error)
	GetPaymentStatus(ctx context.Context, externalID string) (*PaymentStatusResponse, error)
	ProcessRefund(ctx context.Context, request *RefundRequest) (*RefundResponse, error)
	ValidateWebhook(ctx context.Context, payload []byte, signature string) bool
	ProcessWebhook(ctx context.Context, payload []byte) (*WebhookResponse, error)
	CalculateFees(amount int64, currency string) int64
}

// CryptoWallet defines the interface for cryptocurrency wallet operations
type CryptoWallet interface {
	GetAddress(currency string) (string, error)
	GetBalance(currency string) (int64, error)
	SendTransaction(ctx context.Context, to string, amount int64, currency string) (string, error)
	GetTransaction(txHash string) (*CryptoTransaction, error)
	GetTransactionStatus(txHash string) (CryptoTransactionStatus, error)
	MonitorAddress(address string, callback func(*CryptoTransaction)) error
	EstimateGasFee(currency string) (int64, error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendEmail(ctx context.Context, request *EmailRequest) error
	SendSMS(ctx context.Context, request *SMSRequest) error
	SendWebhook(ctx context.Context, request *WebhookRequest) error
	SendPushNotification(ctx context.Context, request *PushRequest) error
}

// FraudDetectionService defines the interface for fraud detection
type FraudDetectionService interface {
	AnalyzePayment(ctx context.Context, payment *entities.Payment) (*entities.FraudDetectionResult, error)
	UpdateWhitelist(ctx context.Context, clientID string) error
	UpdateBlacklist(ctx context.Context, ip string, reason string) error
	GetRiskProfile(ctx context.Context, clientID string) (*RiskProfile, error)
}

// InvoiceGenerator defines the interface for invoice generation
type InvoiceGenerator interface {
	GenerateInvoicePDF(ctx context.Context, invoice *entities.Invoice) ([]byte, error)
	GenerateReceiptPDF(ctx context.Context, receipt *entities.PaymentReceipt) ([]byte, error)
	GenerateInvoiceHTML(ctx context.Context, invoice *entities.Invoice) (string, error)
	GenerateReceiptHTML(ctx context.Context, receipt *entities.PaymentReceipt) (string, error)
}

// ReconciliationService defines the interface for payment reconciliation
type ReconciliationService interface {
	ReconcilePayments(ctx context.Context, startDate, endDate time.Time) (*ReconciliationReport, error)
	MatchTransactions(ctx context.Context, externalTransactions []ExternalTransaction) (*MatchingResult, error)
	ResolveDiscrepancies(ctx context.Context, discrepancies []Discrepancy) error
}

// Request/Response types

// PaymentRequest represents a payment processing request
type PaymentRequest struct {
	InvoiceID     *string                `json:"invoice_id"`
	ClientID      string                 `json:"client_id"`
	ProjectID     *string                `json:"project_id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentMethod entities.PaymentMethod `json:"payment_method"`
	PaymentType   entities.PaymentType   `json:"payment_type"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata"`
	ReturnURL     *string                `json:"return_url"`
	CancelURL     *string                `json:"cancel_url"`
	WebhookURL    *string                `json:"webhook_url"`
	IPAddress     *string                `json:"ip_address"`
	UserAgent     *string                `json:"user_agent"`

	// Credit card specific
	CardToken *string `json:"card_token"`

	// Crypto specific
	WalletAddress *string `json:"wallet_address"`

	// Additional verification
	CVV     *string `json:"cvv"`
	ZipCode *string `json:"zip_code"`
}

// PaymentResponse represents a payment processing response
type PaymentResponse struct {
	PaymentID             string                 `json:"payment_id"`
	ExternalID            *string                `json:"external_id"`
	Status                entities.PaymentStatus `json:"status"`
	Amount                int64                  `json:"amount"`
	Currency              string                 `json:"currency"`
	ProcessorFee          int64                  `json:"processor_fee"`
	NetAmount             int64                  `json:"net_amount"`
	RedirectURL           *string                `json:"redirect_url"`
	Message               string                 `json:"message"`
	TransactionHash       *string                `json:"transaction_hash"`
	EstimatedConfirmation *time.Time             `json:"estimated_confirmation"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// PaymentStatusResponse represents payment status from external processor
type PaymentStatusResponse struct {
	ExternalID    string                 `json:"external_id"`
	Status        entities.PaymentStatus `json:"status"`
	Amount        int64                  `json:"amount"`
	ProcessorFee  int64                  `json:"processor_fee"`
	ProcessedAt   *time.Time             `json:"processed_at"`
	FailureReason *string                `json:"failure_reason"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID   string                 `json:"payment_id"`
	Amount      int64                  `json:"amount"`
	Reason      entities.RefundReason  `json:"reason"`
	ReasonText  *string                `json:"reason_text"`
	RequestedBy string                 `json:"requested_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID            string                 `json:"refund_id"`
	ExternalID          *string                `json:"external_id"`
	Status              entities.RefundStatus  `json:"status"`
	Amount              int64                  `json:"amount"`
	ProcessorFee        int64                  `json:"processor_fee"`
	NetRefund           int64                  `json:"net_refund"`
	EstimatedCompletion *time.Time             `json:"estimated_completion"`
	Message             string                 `json:"message"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// InvoiceRequest represents an invoice creation request
type InvoiceRequest struct {
	ClientID       string                     `json:"client_id"`
	ProjectID      *string                    `json:"project_id"`
	Amount         int64                      `json:"amount"`
	Currency       string                     `json:"currency"`
	Description    string                     `json:"description"`
	LineItems      []entities.InvoiceLineItem `json:"line_items"`
	PaymentTerms   string                     `json:"payment_terms"`
	DueDate        time.Time                  `json:"due_date"`
	PaymentMethods []entities.PaymentMethod   `json:"payment_methods"`
	AutoReminders  bool                       `json:"auto_reminders"`
	Notes          *string                    `json:"notes"`
	Metadata       map[string]interface{}     `json:"metadata"`
}

// WebhookResponse represents webhook processing response
type WebhookResponse struct {
	EventType   string                  `json:"event_type"`
	PaymentID   *string                 `json:"payment_id"`
	ExternalID  *string                 `json:"external_id"`
	Status      *entities.PaymentStatus `json:"status"`
	Amount      *int64                  `json:"amount"`
	ProcessedAt *time.Time              `json:"processed_at"`
	Metadata    map[string]interface{}  `json:"metadata"`
}

// NotificationRequest represents a notification request
type NotificationRequest struct {
	PaymentID        *string                      `json:"payment_id"`
	InvoiceID        *string                      `json:"invoice_id"`
	ClientID         string                       `json:"client_id"`
	NotificationType entities.PaymentNotificationType    `json:"notification_type"`
	Channel          entities.NotificationChannel `json:"channel"`
	Recipient        string                       `json:"recipient"`
	Subject          string                       `json:"subject"`
	Content          string                       `json:"content"`
	Metadata         map[string]interface{}       `json:"metadata"`
}

// Email notification types
type EmailRequest struct {
	To          []string               `json:"to"`
	CC          []string               `json:"cc"`
	BCC         []string               `json:"bcc"`
	Subject     string                 `json:"subject"`
	TextContent string                 `json:"text_content"`
	HTMLContent string                 `json:"html_content"`
	Attachments []EmailAttachment      `json:"attachments"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

// SMS notification
type SMSRequest struct {
	To       string                 `json:"to"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Webhook notification
type WebhookRequest struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Payload map[string]interface{} `json:"payload"`
	Timeout time.Duration          `json:"timeout"`
}

// Push notification
type PushRequest struct {
	DeviceTokens []string               `json:"device_tokens"`
	Title        string                 `json:"title"`
	Body         string                 `json:"body"`
	Data         map[string]interface{} `json:"data"`
}

// Cryptocurrency types
type CryptoTransaction struct {
	Hash          string                  `json:"hash"`
	From          string                  `json:"from"`
	To            string                  `json:"to"`
	Amount        int64                   `json:"amount"`
	Currency      string                  `json:"currency"`
	GasFee        int64                   `json:"gas_fee"`
	BlockNumber   *int64                  `json:"block_number"`
	Confirmations int64                   `json:"confirmations"`
	Status        CryptoTransactionStatus `json:"status"`
	Timestamp     time.Time               `json:"timestamp"`
}

type CryptoTransactionStatus string

const (
	CryptoTxStatusPending   CryptoTransactionStatus = "pending"
	CryptoTxStatusConfirmed CryptoTransactionStatus = "confirmed"
	CryptoTxStatusFailed    CryptoTransactionStatus = "failed"
)

// Fraud detection types
type RiskProfile struct {
	ClientID            string                   `json:"client_id"`
	RiskScore           float64                  `json:"risk_score"`
	PaymentHistory      []string                 `json:"payment_history"`
	SuccessfulPayments  int64                    `json:"successful_payments"`
	FailedPayments      int64                    `json:"failed_payments"`
	LastPaymentAt       *time.Time               `json:"last_payment_at"`
	AverageAmount       float64                  `json:"average_amount"`
	PreferredMethods    []entities.PaymentMethod `json:"preferred_methods"`
	GeographicLocations []string                 `json:"geographic_locations"`
	DeviceFingerprints  []string                 `json:"device_fingerprints"`
}

// Reconciliation types
type ReconciliationReport struct {
	StartDate             time.Time             `json:"start_date"`
	EndDate               time.Time             `json:"end_date"`
	TotalTransactions     int64                 `json:"total_transactions"`
	MatchedTransactions   int64                 `json:"matched_transactions"`
	UnmatchedTransactions int64                 `json:"unmatched_transactions"`
	Discrepancies         []Discrepancy         `json:"discrepancies"`
	Summary               ReconciliationSummary `json:"summary"`
}

type ExternalTransaction struct {
	ExternalID      string    `json:"external_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	TransactionDate time.Time `json:"transaction_date"`
	Reference       string    `json:"reference"`
	Description     string    `json:"description"`
}

type MatchingResult struct {
	Matched    []TransactionMatch    `json:"matched"`
	Unmatched  []ExternalTransaction `json:"unmatched"`
	Duplicates []ExternalTransaction `json:"duplicates"`
}

type TransactionMatch struct {
	InternalPayment     *entities.Payment    `json:"internal_payment"`
	ExternalTransaction *ExternalTransaction `json:"external_transaction"`
	MatchConfidence     float64              `json:"match_confidence"`
}

type Discrepancy struct {
	Type        DiscrepancyType `json:"type"`
	PaymentID   *string         `json:"payment_id"`
	ExternalID  *string         `json:"external_id"`
	Description string          `json:"description"`
	Amount      *int64          `json:"amount"`
	Currency    *string         `json:"currency"`
}

type DiscrepancyType string

const (
	DiscrepancyTypeAmountMismatch  DiscrepancyType = "amount_mismatch"
	DiscrepancyTypeMissingExternal DiscrepancyType = "missing_external"
	DiscrepancyTypeMissingInternal DiscrepancyType = "missing_internal"
	DiscrepancyTypeDuplicate       DiscrepancyType = "duplicate"
	DiscrepancyTypeStatusMismatch  DiscrepancyType = "status_mismatch"
)

type ReconciliationSummary struct {
	TotalAmount        int64   `json:"total_amount"`
	MatchedAmount      int64   `json:"matched_amount"`
	UnmatchedAmount    int64   `json:"unmatched_amount"`
	DiscrepancyAmount  int64   `json:"discrepancy_amount"`
	ReconciliationRate float64 `json:"reconciliation_rate"`
}

type PaymentStats struct {
	TotalPayments      int64   `json:"total_payments"`
	TotalAmount        int64   `json:"total_amount"`
	SuccessfulPayments int64   `json:"successful_payments"`
	FailedPayments     int64   `json:"failed_payments"`
	RefundedPayments   int64   `json:"refunded_payments"`
	RefundedAmount     int64   `json:"refunded_amount"`
	AverageAmount      float64 `json:"average_amount"`
	SuccessRate        float64 `json:"success_rate"`
	TotalFees          int64   `json:"total_fees"`
}

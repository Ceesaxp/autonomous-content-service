package entities

import (
	"time"
)

// PaymentMethod represents different payment options
type PaymentMethod string

const (
	PaymentMethodStripe     PaymentMethod = "stripe"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodEthereum   PaymentMethod = "ethereum"
	PaymentMethodUSDC       PaymentMethod = "usdc"
	PaymentMethodDAI        PaymentMethod = "dai"
	PaymentMethodBitcoin    PaymentMethod = "bitcoin"
	PaymentMethodBankWire   PaymentMethod = "bank_wire"
	PaymentMethodACH        PaymentMethod = "ach"
)

// PaymentStatus represents the current state of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusConfirming PaymentStatus = "confirming"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusDisputed   PaymentStatus = "disputed"
)

// PaymentType represents the purpose of the payment
type PaymentType string

const (
	PaymentTypeInvoice      PaymentType = "invoice"
	PaymentTypeSubscription PaymentType = "subscription"
	PaymentTypeDeposit      PaymentType = "deposit"
	PaymentTypeRefund       PaymentType = "refund"
	PaymentTypePenalty      PaymentType = "penalty"
)

// Payment represents a payment transaction
type Payment struct {
	ID                    string                 `json:"id" db:"id"`
	InvoiceID             *string                `json:"invoice_id" db:"invoice_id"`
	ClientID              string                 `json:"client_id" db:"client_id"`
	ProjectID             *string                `json:"project_id" db:"project_id"`
	Amount                int64                  `json:"amount" db:"amount"`                           // Amount in smallest currency unit (cents, wei, etc.)
	Currency              string                 `json:"currency" db:"currency"`                       // USD, ETH, USDC, etc.
	PaymentMethod         PaymentMethod          `json:"payment_method" db:"payment_method"`
	PaymentType           PaymentType            `json:"payment_type" db:"payment_type"`
	Status                PaymentStatus          `json:"status" db:"status"`
	ExternalID            *string                `json:"external_id" db:"external_id"`                // External payment provider ID
	TransactionHash       *string                `json:"transaction_hash" db:"transaction_hash"`      // Blockchain transaction hash
	WalletAddress         *string                `json:"wallet_address" db:"wallet_address"`          // Crypto wallet address
	ProcessorFee          int64                  `json:"processor_fee" db:"processor_fee"`            // Fee charged by payment processor
	NetAmount             int64                  `json:"net_amount" db:"net_amount"`                  // Amount after fees
	Description           string                 `json:"description" db:"description"`
	Metadata              map[string]interface{} `json:"metadata" db:"metadata"`
	RetryAttempts         int                    `json:"retry_attempts" db:"retry_attempts"`
	MaxRetries            int                    `json:"max_retries" db:"max_retries"`
	LastRetryAt           *time.Time             `json:"last_retry_at" db:"last_retry_at"`
	NextRetryAt           *time.Time             `json:"next_retry_at" db:"next_retry_at"`
	FailureReason         *string                `json:"failure_reason" db:"failure_reason"`
	FraudScore            *float64               `json:"fraud_score" db:"fraud_score"`
	FraudFlags            []string               `json:"fraud_flags" db:"fraud_flags"`
	IPAddress             *string                `json:"ip_address" db:"ip_address"`
	UserAgent             *string                `json:"user_agent" db:"user_agent"`
	RefundedAmount        int64                  `json:"refunded_amount" db:"refunded_amount"`
	RefundReason          *string                `json:"refund_reason" db:"refund_reason"`
	ProcessedAt           *time.Time             `json:"processed_at" db:"processed_at"`
	ConfirmedAt           *time.Time             `json:"confirmed_at" db:"confirmed_at"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// Invoice represents an invoice for payment
type Invoice struct {
	ID               string                 `json:"id" db:"id"`
	InvoiceNumber    string                 `json:"invoice_number" db:"invoice_number"`
	ClientID         string                 `json:"client_id" db:"client_id"`
	ProjectID        *string                `json:"project_id" db:"project_id"`
	Amount           int64                  `json:"amount" db:"amount"`                    // Amount in cents
	Currency         string                 `json:"currency" db:"currency"`
	TaxAmount        int64                  `json:"tax_amount" db:"tax_amount"`
	TotalAmount      int64                  `json:"total_amount" db:"total_amount"`
	Status           InvoiceStatus          `json:"status" db:"status"`
	DueDate          time.Time              `json:"due_date" db:"due_date"`
	Description      string                 `json:"description" db:"description"`
	LineItems        []InvoiceLineItem      `json:"line_items" db:"line_items"`
	PaymentTerms     string                 `json:"payment_terms" db:"payment_terms"`
	Notes            *string                `json:"notes" db:"notes"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	PaidAmount       int64                  `json:"paid_amount" db:"paid_amount"`
	RemainingAmount  int64                  `json:"remaining_amount" db:"remaining_amount"`
	PaymentMethods   []PaymentMethod        `json:"payment_methods" db:"payment_methods"` // Accepted payment methods
	AutoReminders    bool                   `json:"auto_reminders" db:"auto_reminders"`
	RemindersSent    int                    `json:"reminders_sent" db:"reminders_sent"`
	LastReminderAt   *time.Time             `json:"last_reminder_at" db:"last_reminder_at"`
	PaidAt           *time.Time             `json:"paid_at" db:"paid_at"`
	CancelledAt      *time.Time             `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// InvoiceStatus represents the current state of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusPartial   InvoiceStatus = "partial"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   int64   `json:"unit_price"`   // Price in cents
	Amount      int64   `json:"amount"`       // Total amount in cents
	TaxRate     float64 `json:"tax_rate"`     // Tax rate as decimal (0.08 = 8%)
	TaxAmount   int64   `json:"tax_amount"`   // Tax amount in cents
}

// PaymentProcessor represents a payment service provider
type PaymentProcessor struct {
	ID               string                 `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Type             PaymentMethod          `json:"type" db:"type"`
	Enabled          bool                   `json:"enabled" db:"enabled"`
	Priority         int                    `json:"priority" db:"priority"`                // Higher priority = preferred processor
	Configuration    map[string]interface{} `json:"configuration" db:"configuration"`      // API keys, endpoints, etc.
	SupportedMethods []PaymentMethod        `json:"supported_methods" db:"supported_methods"`
	FeeStructure     FeeStructure           `json:"fee_structure" db:"fee_structure"`
	Capabilities     []string               `json:"capabilities" db:"capabilities"`       // refunds, subscriptions, etc.
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// FeeStructure represents payment processor fees
type FeeStructure struct {
	FixedFee      int64   `json:"fixed_fee"`       // Fixed fee in cents
	PercentageFee float64 `json:"percentage_fee"`  // Percentage fee as decimal
	MinFee        int64   `json:"min_fee"`         // Minimum fee in cents
	MaxFee        *int64  `json:"max_fee"`         // Maximum fee in cents (optional)
}

// PaymentNotification represents notifications sent for payment events
type PaymentNotification struct {
	ID              string                 `json:"id" db:"id"`
	PaymentID       string                 `json:"payment_id" db:"payment_id"`
	InvoiceID       *string                `json:"invoice_id" db:"invoice_id"`
	ClientID        string                 `json:"client_id" db:"client_id"`
	NotificationType NotificationType      `json:"notification_type" db:"notification_type"`
	Channel         NotificationChannel    `json:"channel" db:"channel"`
	Recipient       string                 `json:"recipient" db:"recipient"`
	Subject         string                 `json:"subject" db:"subject"`
	Content         string                 `json:"content" db:"content"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	Status          NotificationStatus     `json:"status" db:"status"`
	SentAt          *time.Time             `json:"sent_at" db:"sent_at"`
	DeliveredAt     *time.Time             `json:"delivered_at" db:"delivered_at"`
	OpenedAt        *time.Time             `json:"opened_at" db:"opened_at"`
	ClickedAt       *time.Time             `json:"clicked_at" db:"clicked_at"`
	FailureReason   *string                `json:"failure_reason" db:"failure_reason"`
	RetryAttempts   int                    `json:"retry_attempts" db:"retry_attempts"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationType represents different types of payment notifications
type NotificationType string

const (
	NotificationTypePaymentReceived  NotificationType = "payment_received"
	NotificationTypePaymentFailed    NotificationType = "payment_failed"
	NotificationTypeInvoiceSent      NotificationType = "invoice_sent"
	NotificationTypeInvoiceOverdue   NotificationType = "invoice_overdue"
	NotificationTypeRefundProcessed  NotificationType = "refund_processed"
	NotificationTypeFraudDetected    NotificationType = "fraud_detected"
	NotificationTypePaymentConfirmed NotificationType = "payment_confirmed"
)

// NotificationChannel represents delivery methods for notifications
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelWebhook NotificationChannel = "webhook"
	NotificationChannelPush    NotificationChannel = "push"
	NotificationChannelSlack   NotificationChannel = "slack"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusBounced   NotificationStatus = "bounced"
)

// PaymentReceipt represents a receipt for a completed payment
type PaymentReceipt struct {
	ID            string                 `json:"id" db:"id"`
	PaymentID     string                 `json:"payment_id" db:"payment_id"`
	InvoiceID     *string                `json:"invoice_id" db:"invoice_id"`
	ReceiptNumber string                 `json:"receipt_number" db:"receipt_number"`
	Amount        int64                  `json:"amount" db:"amount"`
	Currency      string                 `json:"currency" db:"currency"`
	PaymentMethod PaymentMethod          `json:"payment_method" db:"payment_method"`
	Description   string                 `json:"description" db:"description"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	GeneratedAt   time.Time              `json:"generated_at" db:"generated_at"`
	SentAt        *time.Time             `json:"sent_at" db:"sent_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// FraudDetectionResult represents the result of fraud detection analysis
type FraudDetectionResult struct {
	PaymentID    string             `json:"payment_id"`
	FraudScore   float64            `json:"fraud_score"`    // 0.0 = low risk, 1.0 = high risk
	RiskLevel    FraudRiskLevel     `json:"risk_level"`
	Flags        []FraudFlag        `json:"flags"`
	Explanation  string             `json:"explanation"`
	Action       FraudAction        `json:"action"`
	Confidence   float64            `json:"confidence"`
	ProcessedAt  time.Time          `json:"processed_at"`
}

// FraudRiskLevel represents the risk level of a payment
type FraudRiskLevel string

const (
	FraudRiskLevelLow      FraudRiskLevel = "low"
	FraudRiskLevelMedium   FraudRiskLevel = "medium"
	FraudRiskLevelHigh     FraudRiskLevel = "high"
	FraudRiskLevelCritical FraudRiskLevel = "critical"
)

// FraudFlag represents specific fraud indicators
type FraudFlag string

const (
	FraudFlagHighVelocity        FraudFlag = "high_velocity"         // Multiple payments in short time
	FraudFlagUnusualAmount       FraudFlag = "unusual_amount"        // Amount significantly different from normal
	FraudFlagSuspiciousLocation  FraudFlag = "suspicious_location"   // Payment from unusual geographic location
	FraudFlagNewPaymentMethod    FraudFlag = "new_payment_method"    // First time using this payment method
	FraudFlagFailedVerification  FraudFlag = "failed_verification"   // Payment verification failed
	FraudFlagBlacklistedIP       FraudFlag = "blacklisted_ip"        // IP address on blacklist
	FraudFlagSuspiciousUserAgent FraudFlag = "suspicious_user_agent" // User agent indicates automation/bot
	FraudFlagTimePatterns        FraudFlag = "time_patterns"         // Payments at unusual times
)

// FraudAction represents actions to take based on fraud detection
type FraudAction string

const (
	FraudActionAllow             FraudAction = "allow"
	FraudActionReview            FraudAction = "review"
	FraudActionRequireVerification FraudAction = "require_verification"
	FraudActionBlock             FraudAction = "block"
	FraudActionRefund            FraudAction = "refund"
)

// PaymentWebhook represents webhook events for payment processing
type PaymentWebhook struct {
	ID            string                 `json:"id" db:"id"`
	PaymentID     *string                `json:"payment_id" db:"payment_id"`
	InvoiceID     *string                `json:"invoice_id" db:"invoice_id"`
	Source        string                 `json:"source" db:"source"`           // stripe, paypal, blockchain, etc.
	EventType     string                 `json:"event_type" db:"event_type"`
	EventData     map[string]interface{} `json:"event_data" db:"event_data"`
	Signature     *string                `json:"signature" db:"signature"`     // Webhook signature for verification
	Processed     bool                   `json:"processed" db:"processed"`
	ProcessedAt   *time.Time             `json:"processed_at" db:"processed_at"`
	FailureReason *string                `json:"failure_reason" db:"failure_reason"`
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// Refund represents a refund transaction
type Refund struct {
	ID              string                 `json:"id" db:"id"`
	PaymentID       string                 `json:"payment_id" db:"payment_id"`
	Amount          int64                  `json:"amount" db:"amount"`
	Currency        string                 `json:"currency" db:"currency"`
	Reason          RefundReason           `json:"reason" db:"reason"`
	ReasonText      *string                `json:"reason_text" db:"reason_text"`
	Status          RefundStatus           `json:"status" db:"status"`
	ExternalID      *string                `json:"external_id" db:"external_id"`
	ProcessorFee    int64                  `json:"processor_fee" db:"processor_fee"`
	NetRefund       int64                  `json:"net_refund" db:"net_refund"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	RequestedBy     string                 `json:"requested_by" db:"requested_by"`
	ApprovedBy      *string                `json:"approved_by" db:"approved_by"`
	ProcessedAt     *time.Time             `json:"processed_at" db:"processed_at"`
	CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
	FailureReason   *string                `json:"failure_reason" db:"failure_reason"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// RefundReason represents the reason for a refund
type RefundReason string

const (
	RefundReasonDuplicate         RefundReason = "duplicate"
	RefundReasonFraud             RefundReason = "fraud"
	RefundReasonRequestedByCustomer RefundReason = "requested_by_customer"
	RefundReasonServiceNotDelivered RefundReason = "service_not_delivered"
	RefundReasonServiceDefective  RefundReason = "service_defective"
	RefundReasonOther             RefundReason = "other"
)

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusCompleted RefundStatus = "completed"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCancelled RefundStatus = "cancelled"
)

// Helper methods for Payment
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

func (p *Payment) CanRetry() bool {
	return p.Status == PaymentStatusFailed && p.RetryAttempts < p.MaxRetries
}

func (p *Payment) IsCryptocurrency() bool {
	return p.PaymentMethod == PaymentMethodEthereum ||
		p.PaymentMethod == PaymentMethodUSDC ||
		p.PaymentMethod == PaymentMethodDAI ||
		p.PaymentMethod == PaymentMethodBitcoin
}

// Helper methods for Invoice
func (i *Invoice) IsOverdue() bool {
	return i.Status != InvoiceStatusPaid && time.Now().After(i.DueDate)
}

func (i *Invoice) IsFullyPaid() bool {
	return i.Status == InvoiceStatusPaid && i.RemainingAmount == 0
}

func (i *Invoice) CalculateRemainingAmount() int64 {
	return i.TotalAmount - i.PaidAmount
}

// Helper methods for FraudDetectionResult
func (f *FraudDetectionResult) IsHighRisk() bool {
	return f.RiskLevel == FraudRiskLevelHigh || f.RiskLevel == FraudRiskLevelCritical
}

func (f *FraudDetectionResult) ShouldBlock() bool {
	return f.Action == FraudActionBlock || f.FraudScore >= 0.8
}

func (f *FraudDetectionResult) RequiresReview() bool {
	return f.Action == FraudActionReview || (f.FraudScore >= 0.5 && f.FraudScore < 0.8)
}
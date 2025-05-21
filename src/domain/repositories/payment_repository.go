package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// PaymentRepository defines the interface for payment data access
type PaymentRepository interface {
	// Payment operations
	CreatePayment(ctx context.Context, payment *entities.Payment) error
	GetPayment(ctx context.Context, id string) (*entities.Payment, error)
	GetPaymentByExternalID(ctx context.Context, externalID string) (*entities.Payment, error)
	GetPaymentsByInvoice(ctx context.Context, invoiceID string) ([]*entities.Payment, error)
	GetPaymentsByClient(ctx context.Context, clientID string, limit, offset int) ([]*entities.Payment, error)
	GetPaymentsByStatus(ctx context.Context, status entities.PaymentStatus, limit, offset int) ([]*entities.Payment, error)
	GetPendingRetries(ctx context.Context) ([]*entities.Payment, error)
	UpdatePayment(ctx context.Context, payment *entities.Payment) error
	UpdatePaymentStatus(ctx context.Context, id string, status entities.PaymentStatus) error
	DeletePayment(ctx context.Context, id string) error

	// Invoice operations
	CreateInvoice(ctx context.Context, invoice *entities.Invoice) error
	GetInvoice(ctx context.Context, id string) (*entities.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)
	GetInvoicesByClient(ctx context.Context, clientID string, limit, offset int) ([]*entities.Invoice, error)
	GetInvoicesByStatus(ctx context.Context, status entities.InvoiceStatus, limit, offset int) ([]*entities.Invoice, error)
	GetOverdueInvoices(ctx context.Context) ([]*entities.Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *entities.Invoice) error
	UpdateInvoiceStatus(ctx context.Context, id string, status entities.InvoiceStatus) error
	UpdateInvoicePaidAmount(ctx context.Context, id string, paidAmount int64) error
	DeleteInvoice(ctx context.Context, id string) error

	// Payment processor operations
	CreatePaymentProcessor(ctx context.Context, processor *entities.PaymentProcessor) error
	GetPaymentProcessor(ctx context.Context, id string) (*entities.PaymentProcessor, error)
	GetPaymentProcessorByType(ctx context.Context, processorType entities.PaymentMethod) (*entities.PaymentProcessor, error)
	GetEnabledPaymentProcessors(ctx context.Context) ([]*entities.PaymentProcessor, error)
	UpdatePaymentProcessor(ctx context.Context, processor *entities.PaymentProcessor) error
	DeletePaymentProcessor(ctx context.Context, id string) error

	// Notification operations
	CreateNotification(ctx context.Context, notification *entities.PaymentNotification) error
	GetNotification(ctx context.Context, id string) (*entities.PaymentNotification, error)
	GetNotificationsByPayment(ctx context.Context, paymentID string) ([]*entities.PaymentNotification, error)
	GetPendingNotifications(ctx context.Context) ([]*entities.PaymentNotification, error)
	UpdateNotification(ctx context.Context, notification *entities.PaymentNotification) error
	DeleteNotification(ctx context.Context, id string) error

	// Receipt operations
	CreateReceipt(ctx context.Context, receipt *entities.PaymentReceipt) error
	GetReceipt(ctx context.Context, id string) (*entities.PaymentReceipt, error)
	GetReceiptByPayment(ctx context.Context, paymentID string) (*entities.PaymentReceipt, error)
	GetReceiptsByClient(ctx context.Context, clientID string, limit, offset int) ([]*entities.PaymentReceipt, error)
	UpdateReceipt(ctx context.Context, receipt *entities.PaymentReceipt) error
	DeleteReceipt(ctx context.Context, id string) error

	// Webhook operations
	CreateWebhook(ctx context.Context, webhook *entities.PaymentWebhook) error
	GetWebhook(ctx context.Context, id string) (*entities.PaymentWebhook, error)
	GetUnprocessedWebhooks(ctx context.Context) ([]*entities.PaymentWebhook, error)
	GetWebhooksBySource(ctx context.Context, source string, limit, offset int) ([]*entities.PaymentWebhook, error)
	UpdateWebhook(ctx context.Context, webhook *entities.PaymentWebhook) error
	DeleteWebhook(ctx context.Context, id string) error

	// Refund operations
	CreateRefund(ctx context.Context, refund *entities.Refund) error
	GetRefund(ctx context.Context, id string) (*entities.Refund, error)
	GetRefundsByPayment(ctx context.Context, paymentID string) ([]*entities.Refund, error)
	GetRefundsByStatus(ctx context.Context, status entities.RefundStatus, limit, offset int) ([]*entities.Refund, error)
	UpdateRefund(ctx context.Context, refund *entities.Refund) error
	DeleteRefund(ctx context.Context, id string) error

	// Analytics and reporting
	GetPaymentStats(ctx context.Context, startDate, endDate time.Time) (*PaymentStats, error)
	GetRevenueByPeriod(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]*RevenueData, error)
	GetPaymentMethodStats(ctx context.Context, startDate, endDate time.Time) ([]*PaymentMethodStats, error)
	GetFraudStats(ctx context.Context, startDate, endDate time.Time) (*FraudStats, error)
}

// PaymentStats represents payment statistics
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

// RevenueData represents revenue data for a specific period
type RevenueData struct {
	Period string    `json:"period"`
	Date   time.Time `json:"date"`
	Amount int64     `json:"amount"`
	Count  int64     `json:"count"`
}

// PaymentMethodStats represents statistics by payment method
type PaymentMethodStats struct {
	PaymentMethod entities.PaymentMethod `json:"payment_method"`
	Count         int64                  `json:"count"`
	Amount        int64                  `json:"amount"`
	SuccessRate   float64                `json:"success_rate"`
	AverageAmount float64                `json:"average_amount"`
}

// FraudStats represents fraud detection statistics
type FraudStats struct {
	TotalPayments     int64    `json:"total_payments"`
	FlaggedPayments   int64    `json:"flagged_payments"`
	BlockedPayments   int64    `json:"blocked_payments"`
	FalsePositives    int64    `json:"false_positives"`
	FraudRate         float64  `json:"fraud_rate"`
	AverageFraudScore float64  `json:"average_fraud_score"`
	TopFraudFlags     []string `json:"top_fraud_flags"`
}

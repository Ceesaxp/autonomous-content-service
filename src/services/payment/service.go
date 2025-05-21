package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/autonomous-content-service/src/domain/entities"
	"github.com/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// ServiceImpl implements the PaymentService interface
type ServiceImpl struct {
	paymentRepo         repositories.PaymentRepository
	processors          map[entities.PaymentMethod]PaymentProcessor
	cryptoWallet        CryptoWallet
	notificationService NotificationService
	fraudDetection      FraudDetectionService
	invoiceGenerator    InvoiceGenerator
	reconciliation      ReconciliationService
	config              *Config
}

// Config holds payment service configuration
type Config struct {
	DefaultCurrency        string
	MaxRetryAttempts       int
	RetryBackoffMultiplier time.Duration
	InvoiceNumberPrefix    string
	ReceiptNumberPrefix    string
	FraudThreshold         float64
	NotificationRetries    int
	WebhookTimeout         time.Duration
	CryptoConfirmations    map[string]int64 // currency -> required confirmations
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	processors map[entities.PaymentMethod]PaymentProcessor,
	cryptoWallet CryptoWallet,
	notificationService NotificationService,
	fraudDetection FraudDetectionService,
	invoiceGenerator InvoiceGenerator,
	reconciliation ReconciliationService,
	config *Config,
) PaymentService {
	return &ServiceImpl{
		paymentRepo:         paymentRepo,
		processors:          processors,
		cryptoWallet:        cryptoWallet,
		notificationService: notificationService,
		fraudDetection:      fraudDetection,
		invoiceGenerator:    invoiceGenerator,
		reconciliation:      reconciliation,
		config:              config,
	}
}

// ProcessPayment processes a payment request
func (s *ServiceImpl) ProcessPayment(ctx context.Context, request *PaymentRequest) (*entities.Payment, error) {
	// Create payment record
	payment := &entities.Payment{
		ID:            uuid.New().String(),
		InvoiceID:     request.InvoiceID,
		ClientID:      request.ClientID,
		ProjectID:     request.ProjectID,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentType:   request.PaymentType,
		Status:        entities.PaymentStatusPending,
		Description:   request.Description,
		Metadata:      request.Metadata,
		MaxRetries:    s.config.MaxRetryAttempts,
		IPAddress:     request.IPAddress,
		UserAgent:     request.UserAgent,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Perform fraud detection
	fraudResult, err := s.fraudDetection.AnalyzePayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("fraud detection failed: %w", err)
	}

	payment.FraudScore = &fraudResult.FraudScore
	payment.FraudFlags = make([]string, len(fraudResult.Flags))
	for i, flag := range fraudResult.Flags {
		payment.FraudFlags[i] = string(flag)
	}

	// Block high-risk payments
	if fraudResult.ShouldBlock() {
		payment.Status = entities.PaymentStatusFailed
		payment.FailureReason = &fraudResult.Explanation
		
		if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
			return nil, fmt.Errorf("failed to create blocked payment: %w", err)
		}

		// Send fraud notification
		s.sendFraudNotification(ctx, payment, fraudResult)
		
		return payment, fmt.Errorf("payment blocked due to fraud risk: %s", fraudResult.Explanation)
	}

	// Save payment record
	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Get payment processor
	processor, exists := s.processors[request.PaymentMethod]
	if !exists {
		return nil, fmt.Errorf("payment method %s not supported", request.PaymentMethod)
	}

	// Update status to processing
	payment.Status = entities.PaymentStatusProcessing
	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Process payment through external processor
	response, err := processor.ProcessPayment(ctx, request)
	if err != nil {
		// Mark payment as failed
		payment.Status = entities.PaymentStatusFailed
		reason := err.Error()
		payment.FailureReason = &reason
		payment.UpdatedAt = time.Now()
		
		s.paymentRepo.UpdatePayment(ctx, payment)
		
		// Schedule retry if eligible
		if payment.CanRetry() {
			s.scheduleRetry(ctx, payment)
		}
		
		// Send failure notification
		s.sendPaymentNotification(ctx, payment, entities.NotificationTypePaymentFailed)
		
		return payment, fmt.Errorf("payment processing failed: %w", err)
	}

	// Update payment with processor response
	payment.ExternalID = response.ExternalID
	payment.Status = response.Status
	payment.ProcessorFee = response.ProcessorFee
	payment.NetAmount = response.NetAmount
	payment.TransactionHash = response.TransactionHash
	
	if response.Status == entities.PaymentStatusCompleted {
		now := time.Now()
		payment.ProcessedAt = &now
		payment.ConfirmedAt = &now
	}
	
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Handle successful payment
	if payment.IsCompleted() {
		// Update invoice if associated
		if payment.InvoiceID != nil {
			if err := s.handleInvoicePayment(ctx, *payment.InvoiceID, payment); err != nil {
				// Log error but don't fail the payment
				fmt.Printf("Failed to update invoice: %v\n", err)
			}
		}

		// Send confirmation notification
		if err := s.sendPaymentNotification(ctx, payment, entities.NotificationTypePaymentConfirmed); err != nil {
			// Log error but don't fail the payment
			fmt.Printf("Failed to send confirmation notification: %v\n", err)
		}

		// Generate and send receipt
		if err := s.generateAndSendReceipt(ctx, payment); err != nil {
			// Log error but don't fail the payment
			fmt.Printf("Failed to generate receipt: %v\n", err)
		}

		// Update treasury if completed
		if err := s.updateTreasury(ctx, payment); err != nil {
			// Log error but don't fail the payment
			fmt.Printf("Failed to update treasury: %v\n", err)
		}
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *ServiceImpl) GetPayment(ctx context.Context, id string) (*entities.Payment, error) {
	return s.paymentRepo.GetPayment(ctx, id)
}

// GetPaymentsByInvoice retrieves payments for an invoice
func (s *ServiceImpl) GetPaymentsByInvoice(ctx context.Context, invoiceID string) ([]*entities.Payment, error) {
	return s.paymentRepo.GetPaymentsByInvoice(ctx, invoiceID)
}

// CancelPayment cancels a pending payment
func (s *ServiceImpl) CancelPayment(ctx context.Context, id string) error {
	payment, err := s.paymentRepo.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != entities.PaymentStatusPending && payment.Status != entities.PaymentStatusProcessing {
		return fmt.Errorf("payment cannot be cancelled in status: %s", payment.Status)
	}

	// Cancel with external processor if has external ID
	if payment.ExternalID != nil {
		processor, exists := s.processors[payment.PaymentMethod]
		if exists {
			// Attempt to cancel with processor (implementation depends on processor)
			// For now, we'll just update our records
		}
	}

	payment.Status = entities.PaymentStatusCancelled
	payment.UpdatedAt = time.Now()

	return s.paymentRepo.UpdatePayment(ctx, payment)
}

// RetryFailedPayment retries a failed payment
func (s *ServiceImpl) RetryFailedPayment(ctx context.Context, id string) error {
	payment, err := s.paymentRepo.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if !payment.CanRetry() {
		return fmt.Errorf("payment cannot be retried")
	}

	// Reset payment status
	payment.Status = entities.PaymentStatusPending
	payment.RetryAttempts++
	now := time.Now()
	payment.LastRetryAt = &now
	payment.UpdatedAt = now

	// Calculate next retry time
	backoff := time.Duration(payment.RetryAttempts) * s.config.RetryBackoffMultiplier
	nextRetry := now.Add(backoff)
	payment.NextRetryAt = &nextRetry

	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment for retry: %w", err)
	}

	// Create new payment request
	request := &PaymentRequest{
		InvoiceID:     payment.InvoiceID,
		ClientID:      payment.ClientID,
		ProjectID:     payment.ProjectID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentMethod: payment.PaymentMethod,
		PaymentType:   payment.PaymentType,
		Description:   payment.Description,
		Metadata:      payment.Metadata,
		IPAddress:     payment.IPAddress,
		UserAgent:     payment.UserAgent,
	}

	// Process the retry
	_, err = s.ProcessPayment(ctx, request)
	return err
}

// CreateInvoice creates a new invoice
func (s *ServiceImpl) CreateInvoice(ctx context.Context, request *InvoiceRequest) (*entities.Invoice, error) {
	// Generate invoice number
	invoiceNumber := fmt.Sprintf("%s-%d", s.config.InvoiceNumberPrefix, time.Now().Unix())

	// Calculate tax amount
	taxAmount := int64(0)
	totalAmount := request.Amount
	
	for _, item := range request.LineItems {
		taxAmount += item.TaxAmount
	}
	
	if taxAmount > 0 {
		totalAmount += taxAmount
	}

	invoice := &entities.Invoice{
		ID:               uuid.New().String(),
		InvoiceNumber:    invoiceNumber,
		ClientID:         request.ClientID,
		ProjectID:        request.ProjectID,
		Amount:           request.Amount,
		Currency:         request.Currency,
		TaxAmount:        taxAmount,
		TotalAmount:      totalAmount,
		Status:           entities.InvoiceStatusDraft,
		DueDate:          request.DueDate,
		Description:      request.Description,
		LineItems:        request.LineItems,
		PaymentTerms:     request.PaymentTerms,
		PaymentMethods:   request.PaymentMethods,
		AutoReminders:    request.AutoReminders,
		Notes:            request.Notes,
		Metadata:         request.Metadata,
		RemainingAmount:  totalAmount,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.paymentRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *ServiceImpl) GetInvoice(ctx context.Context, id string) (*entities.Invoice, error) {
	return s.paymentRepo.GetInvoice(ctx, id)
}

// SendInvoice sends an invoice to the client
func (s *ServiceImpl) SendInvoice(ctx context.Context, id string) error {
	invoice, err := s.paymentRepo.GetInvoice(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Generate invoice PDF
	pdfContent, err := s.invoiceGenerator.GenerateInvoicePDF(ctx, invoice)
	if err != nil {
		return fmt.Errorf("failed to generate invoice PDF: %w", err)
	}

	// Get client information
	// client, err := s.clientRepo.GetClient(ctx, invoice.ClientID)
	// For now, we'll use a placeholder email

	// Send email with invoice
	emailRequest := &EmailRequest{
		To:      []string{"client@example.com"}, // TODO: Get from client
		Subject: fmt.Sprintf("Invoice %s", invoice.InvoiceNumber),
		HTMLContent: fmt.Sprintf(`
			<h2>Invoice %s</h2>
			<p>Dear Client,</p>
			<p>Please find your invoice attached.</p>
			<p>Amount: %s %.2f</p>
			<p>Due Date: %s</p>
			<p>Thank you for your business!</p>
		`, invoice.InvoiceNumber, invoice.Currency, float64(invoice.TotalAmount)/100, invoice.DueDate.Format("January 2, 2006")),
		Attachments: []EmailAttachment{
			{
				Filename:    fmt.Sprintf("invoice-%s.pdf", invoice.InvoiceNumber),
				ContentType: "application/pdf",
				Content:     pdfContent,
			},
		},
	}

	if err := s.notificationService.SendEmail(ctx, emailRequest); err != nil {
		return fmt.Errorf("failed to send invoice email: %w", err)
	}

	// Update invoice status
	invoice.Status = entities.InvoiceStatusSent
	invoice.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdateInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Create notification record
	notification := &entities.PaymentNotification{
		ID:               uuid.New().String(),
		InvoiceID:        &invoice.ID,
		ClientID:         invoice.ClientID,
		NotificationType: entities.NotificationTypeInvoiceSent,
		Channel:          entities.NotificationChannelEmail,
		Recipient:        "client@example.com", // TODO: Get from client
		Subject:          emailRequest.Subject,
		Content:          emailRequest.HTMLContent,
		Status:           entities.NotificationStatusSent,
		SentAt:           &invoice.UpdatedAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return s.paymentRepo.CreateNotification(ctx, notification)
}

// MarkInvoicePaid marks an invoice as paid
func (s *ServiceImpl) MarkInvoicePaid(ctx context.Context, id string, payment *entities.Payment) error {
	invoice, err := s.paymentRepo.GetInvoice(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	invoice.PaidAmount += payment.NetAmount
	invoice.RemainingAmount = invoice.TotalAmount - invoice.PaidAmount

	if invoice.RemainingAmount <= 0 {
		invoice.Status = entities.InvoiceStatusPaid
		now := time.Now()
		invoice.PaidAt = &now
	} else {
		invoice.Status = entities.InvoiceStatusPartial
	}

	invoice.UpdatedAt = time.Now()

	return s.paymentRepo.UpdateInvoice(ctx, invoice)
}

// CancelInvoice cancels an invoice
func (s *ServiceImpl) CancelInvoice(ctx context.Context, id string) error {
	invoice, err := s.paymentRepo.GetInvoice(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice.Status == entities.InvoiceStatusPaid {
		return fmt.Errorf("cannot cancel paid invoice")
	}

	invoice.Status = entities.InvoiceStatusCancelled
	now := time.Now()
	invoice.CancelledAt = &now
	invoice.UpdatedAt = now

	return s.paymentRepo.UpdateInvoice(ctx, invoice)
}

// ProcessRefund processes a refund request
func (s *ServiceImpl) ProcessRefund(ctx context.Context, request *RefundRequest) (*entities.Refund, error) {
	// Get original payment
	payment, err := s.paymentRepo.GetPayment(ctx, request.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if !payment.IsCompleted() {
		return nil, fmt.Errorf("cannot refund incomplete payment")
	}

	// Check if already fully refunded
	if payment.RefundedAmount >= payment.NetAmount {
		return nil, fmt.Errorf("payment already fully refunded")
	}

	// Check refund amount
	maxRefundable := payment.NetAmount - payment.RefundedAmount
	if request.Amount > maxRefundable {
		return nil, fmt.Errorf("refund amount exceeds refundable amount")
	}

	// Create refund record
	refund := &entities.Refund{
		ID:          uuid.New().String(),
		PaymentID:   request.PaymentID,
		Amount:      request.Amount,
		Currency:    payment.Currency,
		Reason:      request.Reason,
		ReasonText:  request.ReasonText,
		Status:      entities.RefundStatusPending,
		Metadata:    request.Metadata,
		RequestedBy: request.RequestedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save refund record
	if err := s.paymentRepo.CreateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Process refund with external processor
	processor, exists := s.processors[payment.PaymentMethod]
	if !exists {
		return nil, fmt.Errorf("payment processor not available for refunds")
	}

	refundRequest := &RefundRequest{
		PaymentID:   payment.ID,
		Amount:      request.Amount,
		Reason:      request.Reason,
		ReasonText:  request.ReasonText,
		RequestedBy: request.RequestedBy,
		Metadata:    request.Metadata,
	}

	response, err := processor.ProcessRefund(ctx, refundRequest)
	if err != nil {
		refund.Status = entities.RefundStatusFailed
		reason := err.Error()
		refund.FailureReason = &reason
		refund.UpdatedAt = time.Now()
		
		s.paymentRepo.UpdateRefund(ctx, refund)
		return refund, fmt.Errorf("refund processing failed: %w", err)
	}

	// Update refund with processor response
	refund.ExternalID = response.ExternalID
	refund.Status = response.Status
	refund.ProcessorFee = response.ProcessorFee
	refund.NetRefund = response.NetRefund
	refund.UpdatedAt = time.Now()

	if response.Status == entities.RefundStatusCompleted {
		now := time.Now()
		refund.ProcessedAt = &now
		refund.CompletedAt = &now

		// Update payment refunded amount
		payment.RefundedAmount += refund.NetRefund
		payment.UpdatedAt = time.Now()
		s.paymentRepo.UpdatePayment(ctx, payment)
	}

	if err := s.paymentRepo.UpdateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to update refund: %w", err)
	}

	// Send refund notification
	if refund.Status == entities.RefundStatusCompleted {
		s.sendRefundNotification(ctx, refund, payment)
	}

	return refund, nil
}

// GetRefund retrieves a refund by ID
func (s *ServiceImpl) GetRefund(ctx context.Context, id string) (*entities.Refund, error) {
	return s.paymentRepo.GetRefund(ctx, id)
}

// GetRefundsByPayment retrieves refunds for a payment
func (s *ServiceImpl) GetRefundsByPayment(ctx context.Context, paymentID string) ([]*entities.Refund, error) {
	return s.paymentRepo.GetRefundsByPayment(ctx, paymentID)
}

// AnalyzeFraud performs fraud analysis on a payment
func (s *ServiceImpl) AnalyzeFraud(ctx context.Context, payment *entities.Payment) (*entities.FraudDetectionResult, error) {
	return s.fraudDetection.AnalyzePayment(ctx, payment)
}

// SendNotification sends a payment notification
func (s *ServiceImpl) SendNotification(ctx context.Context, request *NotificationRequest) error {
	notification := &entities.PaymentNotification{
		ID:               uuid.New().String(),
		PaymentID:        request.PaymentID,
		InvoiceID:        request.InvoiceID,
		ClientID:         request.ClientID,
		NotificationType: request.NotificationType,
		Channel:          request.Channel,
		Recipient:        request.Recipient,
		Subject:          request.Subject,
		Content:          request.Content,
		Metadata:         request.Metadata,
		Status:           entities.NotificationStatusPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save notification record
	if err := s.paymentRepo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification based on channel
	var err error
	switch request.Channel {
	case entities.NotificationChannelEmail:
		emailReq := &EmailRequest{
			To:          []string{request.Recipient},
			Subject:     request.Subject,
			HTMLContent: request.Content,
		}
		err = s.notificationService.SendEmail(ctx, emailReq)
	case entities.NotificationChannelSMS:
		smsReq := &SMSRequest{
			To:      request.Recipient,
			Message: request.Content,
		}
		err = s.notificationService.SendSMS(ctx, smsReq)
	case entities.NotificationChannelWebhook:
		webhookReq := &WebhookRequest{
			URL:     request.Recipient,
			Method:  "POST",
			Payload: request.Metadata,
			Timeout: s.config.WebhookTimeout,
		}
		err = s.notificationService.SendWebhook(ctx, webhookReq)
	default:
		err = fmt.Errorf("unsupported notification channel: %s", request.Channel)
	}

	// Update notification status
	if err != nil {
		notification.Status = entities.NotificationStatusFailed
		reason := err.Error()
		notification.FailureReason = &reason
	} else {
		notification.Status = entities.NotificationStatusSent
		now := time.Now()
		notification.SentAt = &now
	}

	notification.UpdatedAt = time.Now()
	s.paymentRepo.UpdateNotification(ctx, notification)

	return err
}

// GetPaymentStats retrieves payment statistics
func (s *ServiceImpl) GetPaymentStats(ctx context.Context, startDate, endDate time.Time) (*PaymentStats, error) {
	return s.paymentRepo.GetPaymentStats(ctx, startDate, endDate)
}

// Helper methods

func (s *ServiceImpl) scheduleRetry(ctx context.Context, payment *entities.Payment) {
	backoff := time.Duration(payment.RetryAttempts+1) * s.config.RetryBackoffMultiplier
	nextRetry := time.Now().Add(backoff)
	payment.NextRetryAt = &nextRetry
	payment.UpdatedAt = time.Now()
	
	s.paymentRepo.UpdatePayment(ctx, payment)
}

func (s *ServiceImpl) handleInvoicePayment(ctx context.Context, invoiceID string, payment *entities.Payment) error {
	return s.MarkInvoicePaid(ctx, invoiceID, payment)
}

func (s *ServiceImpl) sendPaymentNotification(ctx context.Context, payment *entities.Payment, notificationType entities.NotificationType) error {
	var subject, content string
	
	switch notificationType {
	case entities.NotificationTypePaymentReceived:
		subject = "Payment Received"
		content = fmt.Sprintf("Your payment of %s %.2f has been received.", payment.Currency, float64(payment.Amount)/100)
	case entities.NotificationTypePaymentFailed:
		subject = "Payment Failed"
		content = fmt.Sprintf("Your payment of %s %.2f has failed.", payment.Currency, float64(payment.Amount)/100)
		if payment.FailureReason != nil {
			content += fmt.Sprintf(" Reason: %s", *payment.FailureReason)
		}
	case entities.NotificationTypePaymentConfirmed:
		subject = "Payment Confirmed"
		content = fmt.Sprintf("Your payment of %s %.2f has been confirmed.", payment.Currency, float64(payment.NetAmount)/100)
	}

	request := &NotificationRequest{
		PaymentID:        &payment.ID,
		ClientID:         payment.ClientID,
		NotificationType: notificationType,
		Channel:          entities.NotificationChannelEmail,
		Recipient:        "client@example.com", // TODO: Get from client
		Subject:          subject,
		Content:          content,
	}

	return s.SendNotification(ctx, request)
}

func (s *ServiceImpl) sendFraudNotification(ctx context.Context, payment *entities.Payment, fraudResult *entities.FraudDetectionResult) {
	request := &NotificationRequest{
		PaymentID:        &payment.ID,
		ClientID:         payment.ClientID,
		NotificationType: entities.NotificationTypeFraudDetected,
		Channel:          entities.NotificationChannelEmail,
		Recipient:        "security@company.com", // Internal notification
		Subject:          "Fraudulent Payment Detected",
		Content:          fmt.Sprintf("Payment %s flagged for fraud. Score: %.2f. Flags: %v", payment.ID, fraudResult.FraudScore, fraudResult.Flags),
	}

	s.SendNotification(ctx, request)
}

func (s *ServiceImpl) sendRefundNotification(ctx context.Context, refund *entities.Refund, payment *entities.Payment) {
	request := &NotificationRequest{
		PaymentID:        &payment.ID,
		ClientID:         payment.ClientID,
		NotificationType: entities.NotificationTypeRefundProcessed,
		Channel:          entities.NotificationChannelEmail,
		Recipient:        "client@example.com", // TODO: Get from client
		Subject:          "Refund Processed",
		Content:          fmt.Sprintf("Your refund of %s %.2f has been processed.", refund.Currency, float64(refund.NetRefund)/100),
	}

	s.SendNotification(ctx, request)
}

func (s *ServiceImpl) generateAndSendReceipt(ctx context.Context, payment *entities.Payment) error {
	receiptNumber := fmt.Sprintf("%s-%d", s.config.ReceiptNumberPrefix, time.Now().Unix())

	receipt := &entities.PaymentReceipt{
		ID:            uuid.New().String(),
		PaymentID:     payment.ID,
		InvoiceID:     payment.InvoiceID,
		ReceiptNumber: receiptNumber,
		Amount:        payment.NetAmount,
		Currency:      payment.Currency,
		PaymentMethod: payment.PaymentMethod,
		Description:   payment.Description,
		GeneratedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := s.paymentRepo.CreateReceipt(ctx, receipt); err != nil {
		return fmt.Errorf("failed to create receipt: %w", err)
	}

	// Generate receipt PDF
	pdfContent, err := s.invoiceGenerator.GenerateReceiptPDF(ctx, receipt)
	if err != nil {
		return fmt.Errorf("failed to generate receipt PDF: %w", err)
	}

	// Send receipt email
	emailRequest := &EmailRequest{
		To:      []string{"client@example.com"}, // TODO: Get from client
		Subject: fmt.Sprintf("Receipt %s", receipt.ReceiptNumber),
		HTMLContent: fmt.Sprintf(`
			<h2>Payment Receipt</h2>
			<p>Thank you for your payment!</p>
			<p>Receipt Number: %s</p>
			<p>Amount: %s %.2f</p>
			<p>Payment Method: %s</p>
		`, receipt.ReceiptNumber, receipt.Currency, float64(receipt.Amount)/100, receipt.PaymentMethod),
		Attachments: []EmailAttachment{
			{
				Filename:    fmt.Sprintf("receipt-%s.pdf", receipt.ReceiptNumber),
				ContentType: "application/pdf",
				Content:     pdfContent,
			},
		},
	}

	if err := s.notificationService.SendEmail(ctx, emailRequest); err != nil {
		return fmt.Errorf("failed to send receipt email: %w", err)
	}

	now := time.Now()
	receipt.SentAt = &now
	return s.paymentRepo.UpdateReceipt(ctx, receipt)
}

func (s *ServiceImpl) updateTreasury(ctx context.Context, payment *entities.Payment) error {
	// TODO: Integration with treasury system
	// This would call the treasury service to record the revenue
	fmt.Printf("TODO: Update treasury with payment %s, amount %d %s\n", payment.ID, payment.NetAmount, payment.Currency)
	return nil
}
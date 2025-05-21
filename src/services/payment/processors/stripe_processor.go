package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/autonomous-content-service/src/domain/entities"
	"github.com/autonomous-content-service/src/services/payment"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/client"
	"github.com/stripe/stripe-go/v74/webhook"
)

// StripeProcessor implements payment processing for Stripe
type StripeProcessor struct {
	client          *client.API
	webhookSecret   string
	config          *StripeConfig
}

// StripeConfig holds Stripe-specific configuration
type StripeConfig struct {
	SecretKey       string
	PublishableKey  string
	WebhookSecret   string
	EndpointSecret  string
	ConnectAccountID *string
	FeesMode        string // "separate" or "included"
	StatementDescriptor string
	CaptureMethod   string // "automatic" or "manual"
}

// NewStripeProcessor creates a new Stripe payment processor
func NewStripeProcessor(config *StripeConfig) payment.PaymentProcessor {
	// Initialize Stripe client
	sc := &client.API{}
	sc.Init(config.SecretKey, nil)

	return &StripeProcessor{
		client:        sc,
		webhookSecret: config.WebhookSecret,
		config:        config,
	}
}

// GetName returns the processor name
func (s *StripeProcessor) GetName() string {
	return "stripe"
}

// GetSupportedMethods returns supported payment methods
func (s *StripeProcessor) GetSupportedMethods() []entities.PaymentMethod {
	return []entities.PaymentMethod{
		entities.PaymentMethodStripe,
	}
}

// ProcessPayment processes a payment through Stripe
func (s *StripeProcessor) ProcessPayment(ctx context.Context, request *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Convert amount to cents (Stripe expects smallest currency unit)
	amount := request.Amount

	// Create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(request.Currency),
		Description: stripe.String(request.Description),
		CaptureMethod: stripe.String(s.config.CaptureMethod),
	}

	// Add metadata
	if request.Metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range request.Metadata {
			if str, ok := v.(string); ok {
				params.Metadata[k] = str
			}
		}
	}

	// Add client ID as metadata
	if params.Metadata == nil {
		params.Metadata = make(map[string]string)
	}
	params.Metadata["client_id"] = request.ClientID
	if request.InvoiceID != nil {
		params.Metadata["invoice_id"] = *request.InvoiceID
	}

	// Add statement descriptor
	if s.config.StatementDescriptor != "" {
		params.StatementDescriptor = stripe.String(s.config.StatementDescriptor)
	}

	// Add return URLs for redirect-based payments
	if request.ReturnURL != nil {
		params.ConfirmationMethod = stripe.String("manual")
		params.ReturnURL = stripe.String(*request.ReturnURL)
	}

	// Add payment method if provided (for card tokens)
	if request.CardToken != nil {
		params.PaymentMethod = stripe.String(*request.CardToken)
		params.ConfirmationMethod = stripe.String("automatic")
		params.Confirm = stripe.Bool(true)
	}

	// Create the payment intent
	intent, err := s.client.PaymentIntents.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Calculate processor fees (Stripe: 2.9% + 30¢ for cards)
	processorFee := s.CalculateFees(amount, request.Currency)
	netAmount := amount - processorFee

	// Determine status based on payment intent status
	var status entities.PaymentStatus
	var redirectURL *string

	switch intent.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = entities.PaymentStatusCompleted
	case stripe.PaymentIntentStatusProcessing:
		status = entities.PaymentStatusProcessing
	case stripe.PaymentIntentStatusRequiresConfirmation:
		status = entities.PaymentStatusConfirming
	case stripe.PaymentIntentStatusRequiresAction:
		status = entities.PaymentStatusPending
		if intent.NextAction != nil && intent.NextAction.RedirectToURL != nil {
			redirectURL = &intent.NextAction.RedirectToURL.URL
		}
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		status = entities.PaymentStatusPending
	case stripe.PaymentIntentStatusCanceled:
		status = entities.PaymentStatusCancelled
	default:
		status = entities.PaymentStatusPending
	}

	response := &payment.PaymentResponse{
		PaymentID:    request.ClientID, // Will be replaced with actual payment ID
		ExternalID:   &intent.ID,
		Status:       status,
		Amount:       amount,
		Currency:     request.Currency,
		ProcessorFee: processorFee,
		NetAmount:    netAmount,
		RedirectURL:  redirectURL,
		Message:      "Payment processed with Stripe",
		Metadata: map[string]interface{}{
			"stripe_payment_intent_id": intent.ID,
			"stripe_client_secret":     intent.ClientSecret,
		},
	}

	return response, nil
}

// GetPaymentStatus retrieves payment status from Stripe
func (s *StripeProcessor) GetPaymentStatus(ctx context.Context, externalID string) (*payment.PaymentStatusResponse, error) {
	intent, err := s.client.PaymentIntents.Get(externalID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	var status entities.PaymentStatus
	var failureReason *string

	switch intent.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = entities.PaymentStatusCompleted
	case stripe.PaymentIntentStatusProcessing:
		status = entities.PaymentStatusProcessing
	case stripe.PaymentIntentStatusRequiresConfirmation:
		status = entities.PaymentStatusConfirming
	case stripe.PaymentIntentStatusRequiresAction:
		status = entities.PaymentStatusPending
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		status = entities.PaymentStatusPending
	case stripe.PaymentIntentStatusCanceled:
		status = entities.PaymentStatusCancelled
	default:
		status = entities.PaymentStatusFailed
		if intent.LastPaymentError != nil {
			reason := intent.LastPaymentError.Message
			failureReason = &reason
		}
	}

	var processedAt *time.Time
	if intent.Created > 0 {
		timestamp := time.Unix(intent.Created, 0)
		processedAt = &timestamp
	}

	processorFee := s.CalculateFees(intent.Amount, string(intent.Currency))

	response := &payment.PaymentStatusResponse{
		ExternalID:    externalID,
		Status:        status,
		Amount:        intent.Amount,
		ProcessorFee:  processorFee,
		ProcessedAt:   processedAt,
		FailureReason: failureReason,
		Metadata: map[string]interface{}{
			"stripe_status": intent.Status,
			"charges":       len(intent.Charges.Data),
		},
	}

	return response, nil
}

// ProcessRefund processes a refund through Stripe
func (s *StripeProcessor) ProcessRefund(ctx context.Context, request *payment.RefundRequest) (*payment.RefundResponse, error) {
	// First, get the payment intent to find the charge
	intent, err := s.client.PaymentIntents.Get(request.PaymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	if len(intent.Charges.Data) == 0 {
		return nil, fmt.Errorf("no charges found for payment intent")
	}

	charge := intent.Charges.Data[0]

	// Create refund parameters
	params := &stripe.RefundParams{
		Charge: stripe.String(charge.ID),
		Amount: stripe.Int64(request.Amount),
		Reason: stripe.String(s.mapRefundReason(request.Reason)),
	}

	// Add metadata
	if request.Metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range request.Metadata {
			if str, ok := v.(string); ok {
				params.Metadata[k] = str
			}
		}
	}

	if request.ReasonText != nil {
		if params.Metadata == nil {
			params.Metadata = make(map[string]string)
		}
		params.Metadata["reason_text"] = *request.ReasonText
	}

	// Process the refund
	refund, err := s.client.Refunds.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Calculate refund fees (Stripe doesn't refund processing fees)
	processorFee := s.CalculateFees(request.Amount, string(intent.Currency))
	netRefund := request.Amount - processorFee

	// Map Stripe status to our status
	var status entities.RefundStatus
	switch refund.Status {
	case "succeeded":
		status = entities.RefundStatusCompleted
	case "pending":
		status = entities.RefundStatusProcessing
	case "failed":
		status = entities.RefundStatusFailed
	default:
		status = entities.RefundStatusPending
	}

	response := &payment.RefundResponse{
		RefundID:     refund.ID,
		ExternalID:   &refund.ID,
		Status:       status,
		Amount:       request.Amount,
		ProcessorFee: processorFee,
		NetRefund:    netRefund,
		Message:      "Refund processed with Stripe",
		Metadata: map[string]interface{}{
			"stripe_refund_id": refund.ID,
			"charge_id":        charge.ID,
		},
	}

	return response, nil
}

// ValidateWebhook validates Stripe webhook signatures
func (s *StripeProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) bool {
	_, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	return err == nil
}

// ProcessWebhook processes Stripe webhook events
func (s *StripeProcessor) ProcessWebhook(ctx context.Context, payload []byte) (*payment.WebhookResponse, error) {
	event, err := webhook.ConstructEvent(payload, "", s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	response := &payment.WebhookResponse{
		EventType: event.Type,
		Metadata: map[string]interface{}{
			"stripe_event_id": event.ID,
			"created":         event.Created,
		},
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var intent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return nil, fmt.Errorf("failed to parse payment intent: %w", err)
		}

		response.ExternalID = &intent.ID
		status := entities.PaymentStatusCompleted
		response.Status = &status
		response.Amount = &intent.Amount
		now := time.Now()
		response.ProcessedAt = &now

	case "payment_intent.payment_failed":
		var intent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return nil, fmt.Errorf("failed to parse payment intent: %w", err)
		}

		response.ExternalID = &intent.ID
		status := entities.PaymentStatusFailed
		response.Status = &status
		response.Amount = &intent.Amount

	case "charge.dispute.created":
		var dispute stripe.Dispute
		if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
			return nil, fmt.Errorf("failed to parse dispute: %w", err)
		}

		response.ExternalID = &dispute.Charge.ID
		status := entities.PaymentStatusDisputed
		response.Status = &status
		response.Amount = &dispute.Amount

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return nil, fmt.Errorf("failed to parse invoice: %w", err)
		}

		response.ExternalID = &invoice.ID
		status := entities.PaymentStatusCompleted
		response.Status = &status
		response.Amount = &invoice.AmountPaid

	default:
		return nil, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	return response, nil
}

// CalculateFees calculates Stripe processing fees
func (s *StripeProcessor) CalculateFees(amount int64, currency string) int64 {
	// Stripe standard rates: 2.9% + 30¢ for US cards
	// International cards: 3.4% + 30¢
	// For simplicity, using standard rate
	
	percentageFee := float64(amount) * 0.029 // 2.9%
	fixedFee := int64(30)                    // 30 cents
	
	if currency != "USD" {
		// International rate
		percentageFee = float64(amount) * 0.034 // 3.4%
	}
	
	totalFee := int64(percentageFee) + fixedFee
	return totalFee
}

// Helper methods

func (s *StripeProcessor) mapRefundReason(reason entities.RefundReason) string {
	switch reason {
	case entities.RefundReasonDuplicate:
		return "duplicate"
	case entities.RefundReasonFraud:
		return "fraudulent"
	case entities.RefundReasonRequestedByCustomer:
		return "requested_by_customer"
	default:
		return "requested_by_customer"
	}
}

// CreateCustomer creates a Stripe customer
func (s *StripeProcessor) CreateCustomer(ctx context.Context, email, name string, metadata map[string]string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	if metadata != nil {
		params.Metadata = metadata
	}

	return s.client.Customers.New(params)
}

// CreatePaymentMethod creates a payment method for reusable payments
func (s *StripeProcessor) CreatePaymentMethod(ctx context.Context, customerID, cardToken string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodParams{
		Type:     stripe.String("card"),
		Customer: stripe.String(customerID),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(cardToken),
		},
	}

	return s.client.PaymentMethods.New(params)
}

// CreateSubscription creates a recurring subscription
func (s *StripeProcessor) CreateSubscription(ctx context.Context, customerID, priceID string, metadata map[string]string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	if metadata != nil {
		params.Metadata = metadata
	}

	return s.client.Subscriptions.New(params)
}

// HandleResponse handles HTTP responses from Stripe API
func (s *StripeProcessor) HandleResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		return fmt.Errorf("stripe API error: %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}
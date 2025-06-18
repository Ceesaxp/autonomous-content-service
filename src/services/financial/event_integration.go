package financial

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/events"
	"github.com/Ceesaxp/autonomous-content-service/src/services/payment"
	"github.com/Ceesaxp/autonomous-content-service/src/services/pricing"
	"github.com/google/uuid"
)

// EventIntegratedFinancialService wraps financial services with event-driven capabilities
type EventIntegratedFinancialService struct {
	paymentProcessor payment.PaymentProcessor
	pricingEngine    *pricing.DynamicPricingEngine
	eventBus         *events.ServiceEventBus
	invoiceRepo      repositories.InvoiceRepository
	paymentRepo      repositories.PaymentRepository
	projectRepo      repositories.ProjectRepository
}

// NewEventIntegratedFinancialService creates a new event-integrated financial service
func NewEventIntegratedFinancialService(
	paymentProcessor payment.PaymentProcessor,
	pricingEngine *pricing.DynamicPricingEngine,
	eventBus *events.ServiceEventBus,
	invoiceRepo repositories.InvoiceRepository,
	paymentRepo repositories.PaymentRepository,
	projectRepo repositories.ProjectRepository,
) *EventIntegratedFinancialService {
	return &EventIntegratedFinancialService{
		paymentProcessor: paymentProcessor,
		pricingEngine:    pricingEngine,
		eventBus:         eventBus,
		invoiceRepo:      invoiceRepo,
		paymentRepo:      paymentRepo,
		projectRepo:      projectRepo,
	}
}

// HandleProjectCreated sets up financial tracking for new projects
func (s *EventIntegratedFinancialService) HandleProjectCreated(ctx context.Context, event events.Event) error {
	projectID, _ := event.Payload["project_id"].(string)
	clientID, _ := event.Payload["client_id"].(string)
	budget, _ := event.Payload["budget"].(float64)

	log.Printf("[FinancialService] Setting up financials for project %s (budget: %.2f)", projectID, budget)

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	// Load project details
	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Calculate pricing if not set
	if project.Budget == 0 {
		quote, err := s.pricingEngine.GenerateQuote(ctx, &pricing.QuoteRequest{
			ClientID:    clientID,
			ProjectType: project.Type,
			Complexity:  pricing.ComplexityMedium, // Should be calculated
			Urgency:     pricing.UrgencyStandard,
			Features:    []string{},
		})
		if err != nil {
			log.Printf("[FinancialService] Failed to generate quote: %v", err)
		} else {
			project.Budget = quote.TotalAmount
			s.projectRepo.Update(ctx, project)
		}
	}

	// Create initial invoice (deposit or full payment based on amount)
	invoice := &entities.Invoice{
		ID:          uuid.New(),
		ProjectID:   projectUUID,
		ClientID:    uuid.MustParse(clientID),
		Amount:      project.Budget * 0.5, // 50% deposit
		Currency:    "USD",
		Status:      entities.InvoiceStatusDraft,
		DueDate:     time.Now().AddDate(0, 0, 14), // 14 days payment terms
		Items: []entities.InvoiceItem{
			{
				Description: fmt.Sprintf("Initial deposit for project: %s", project.Title),
				Amount:      project.Budget * 0.5,
				Quantity:    1,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Publish invoice created event
	s.publishInvoiceCreatedEvent(ctx, invoice)

	// Set up payment tracking
	s.eventBus.PublishEvent(ctx, "financial.project_setup_complete", map[string]interface{}{
		"project_id":       projectID,
		"invoice_id":       invoice.ID.String(),
		"budget":           project.Budget,
		"deposit_amount":   invoice.Amount,
		"payment_terms":    "net-14",
	})

	return nil
}

// HandleProjectCompleted processes final payments for completed projects
func (s *EventIntegratedFinancialService) HandleProjectCompleted(ctx context.Context, event events.Event) error {
	projectID, _ := event.Payload["project_id"].(string)

	log.Printf("[FinancialService] Processing completion payment for project %s", projectID)

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	// Check if there's an outstanding balance
	invoices, err := s.invoiceRepo.ListByProject(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("failed to list invoices: %w", err)
	}

	var totalPaid float64
	var project *entities.Project

	for _, invoice := range invoices {
		if invoice.Status == entities.InvoiceStatusPaid {
			totalPaid += invoice.Amount
		}
		if project == nil {
			project, _ = s.projectRepo.GetByID(ctx, invoice.ProjectID)
		}
	}

	if project == nil {
		return fmt.Errorf("project not found")
	}

	// Create final invoice if there's a balance
	remainingBalance := project.Budget - totalPaid
	if remainingBalance > 0 {
		finalInvoice := &entities.Invoice{
			ID:        uuid.New(),
			ProjectID: projectUUID,
			ClientID:  project.ClientID,
			Amount:    remainingBalance,
			Currency:  "USD",
			Status:    entities.InvoiceStatusPending,
			DueDate:   time.Now().AddDate(0, 0, 14),
			Items: []entities.InvoiceItem{
				{
					Description: fmt.Sprintf("Final payment for project: %s", project.Title),
					Amount:      remainingBalance,
					Quantity:    1,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.invoiceRepo.Create(ctx, finalInvoice); err != nil {
			return fmt.Errorf("failed to create final invoice: %w", err)
		}

		// Send the invoice
		finalInvoice.Status = entities.InvoiceStatusSent
		finalInvoice.SentAt = timePtr(time.Now())
		s.invoiceRepo.Update(ctx, finalInvoice)

		s.publishInvoiceCreatedEvent(ctx, finalInvoice)
	}

	// Update project financial status
	s.eventBus.PublishEvent(ctx, "financial.project_finalized", map[string]interface{}{
		"project_id":    projectID,
		"total_budget":  project.Budget,
		"total_paid":    totalPaid,
		"balance_due":   remainingBalance,
	})

	return nil
}

// HandleContentApproved triggers milestone payments
func (s *EventIntegratedFinancialService) HandleContentApproved(ctx context.Context, event events.Event) error {
	contentID, _ := event.Payload["content_id"].(string)
	projectID, _ := event.Payload["project_id"].(string)

	log.Printf("[FinancialService] Processing milestone payment for approved content %s", contentID)

	// Check if this content approval triggers a milestone payment
	projectUUID, _ := uuid.Parse(projectID)
	project, _ := s.projectRepo.GetByID(ctx, projectUUID)
	
	if project == nil {
		return nil
	}

	// Simple milestone logic: create invoice every 5 pieces of approved content
	// In production, this would be based on project milestones
	contentCount := s.getApprovedContentCount(ctx, projectUUID)
	if contentCount%5 == 0 && contentCount > 0 {
		milestoneAmount := project.Budget * 0.2 // 20% per milestone

		milestoneInvoice := &entities.Invoice{
			ID:        uuid.New(),
			ProjectID: projectUUID,
			ClientID:  project.ClientID,
			Amount:    milestoneAmount,
			Currency:  "USD",
			Status:    entities.InvoiceStatusPending,
			DueDate:   time.Now().AddDate(0, 0, 7), // 7 days for milestones
			Items: []entities.InvoiceItem{
				{
					Description: fmt.Sprintf("Milestone payment - %d items delivered", contentCount),
					Amount:      milestoneAmount,
					Quantity:    1,
				},
			},
			Metadata: map[string]interface{}{
				"milestone_number": contentCount / 5,
				"content_count":    contentCount,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.invoiceRepo.Create(ctx, milestoneInvoice)
		s.publishInvoiceCreatedEvent(ctx, milestoneInvoice)

		log.Printf("[FinancialService] Created milestone invoice for %d delivered items", contentCount)
	}

	return nil
}

// HandleClientOnboarded sets up client payment profile
func (s *EventIntegratedFinancialService) HandleClientOnboarded(ctx context.Context, event events.Event) error {
	clientID, _ := event.Payload["client_id"].(string)
	tier, _ := event.Payload["tier"].(string)

	log.Printf("[FinancialService] Setting up payment profile for client %s (tier: %s)", clientID, tier)

	// Set up tier-based pricing preferences
	var discountRate float64
	var paymentTerms int

	switch tier {
	case "enterprise":
		discountRate = 0.20 // 20% discount
		paymentTerms = 30   // NET-30
	case "premium":
		discountRate = 0.10 // 10% discount
		paymentTerms = 21   // NET-21
	default:
		discountRate = 0.0 // No discount
		paymentTerms = 14  // NET-14
	}

	// Store client financial preferences
	s.eventBus.PublishEvent(ctx, "financial.client_profile_created", map[string]interface{}{
		"client_id":     clientID,
		"tier":          tier,
		"discount_rate": discountRate,
		"payment_terms": paymentTerms,
		"currency":      "USD",
	})

	return nil
}

// HandleDecisionExecuted implements financial decisions
func (s *EventIntegratedFinancialService) HandleDecisionExecuted(ctx context.Context, event events.Event) error {
	decisionID, _ := event.Payload["decision_id"].(string)
	decisionType, _ := event.Payload["type"].(string)
	selectedOption, _ := event.Payload["selected_option"].(string)

	log.Printf("[FinancialService] Implementing financial decision %s (type: %s, option: %s)", 
		decisionID, decisionType, selectedOption)

	switch decisionType {
	case "financial":
		return s.handleFinancialDecision(ctx, decisionID, selectedOption, event.Payload)
	case "pricing":
		return s.handlePricingDecision(ctx, decisionID, selectedOption, event.Payload)
	}

	return nil
}

// HandleRiskDetected adjusts financial operations based on risk
func (s *EventIntegratedFinancialService) HandleRiskDetected(ctx context.Context, event events.Event) error {
	category, _ := event.Payload["category"].(string)
	
	if category != "financial" {
		return nil
	}

	riskID, _ := event.Payload["risk_id"].(string)
	severity, _ := event.Payload["severity"].(string)
	score, _ := event.Payload["score"].(float64)

	log.Printf("[FinancialService] Responding to financial risk %s (severity: %s, score: %.2f)", 
		riskID, severity, score)

	// Adjust payment processing based on risk
	if severity == "high" || score > 0.8 {
		// Enable enhanced verification for all payments
		s.eventBus.PublishEvent(ctx, "financial.enhanced_verification_enabled", map[string]interface{}{
			"risk_id":  riskID,
			"duration": "24h",
			"reason":   "High financial risk detected",
		})

		// Lower automatic payment thresholds
		s.eventBus.PublishEvent(ctx, "financial.threshold_adjusted", map[string]interface{}{
			"threshold_type": "auto_approval",
			"old_value":      1000.0,
			"new_value":      100.0,
			"reason":         "Risk mitigation",
		})
	}

	return nil
}

// Payment processing methods

// ProcessInvoicePayment processes a payment for an invoice
func (s *EventIntegratedFinancialService) ProcessInvoicePayment(ctx context.Context, invoiceID uuid.UUID, paymentMethod entities.PaymentMethod) error {
	invoice, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	if invoice.Status == entities.InvoiceStatusPaid {
		return fmt.Errorf("invoice already paid")
	}

	// Create payment record
	payment := &entities.Payment{
		ID:            uuid.New(),
		InvoiceID:     &invoiceID,
		ClientID:      invoice.ClientID,
		Amount:        invoice.Amount,
		Currency:      invoice.Currency,
		Method:        paymentMethod,
		Status:        entities.PaymentStatusPending,
		TransactionID: uuid.New().String(),
		CreatedAt:     time.Now(),
	}

	// Process payment through payment processor
	result, err := s.paymentProcessor.ProcessPayment(ctx, &payment.PaymentRequest{
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		ClientID:      payment.ClientID.String(),
		Description:   fmt.Sprintf("Payment for invoice %s", invoiceID),
		PaymentMethod: string(paymentMethod),
		Metadata: map[string]string{
			"invoice_id": invoiceID.String(),
			"project_id": invoice.ProjectID.String(),
		},
	})

	if err != nil {
		payment.Status = entities.PaymentStatusFailed
		payment.FailureReason = err.Error()
		s.paymentRepo.Create(ctx, payment)
		
		s.publishPaymentFailedEvent(ctx, payment, invoice)
		return fmt.Errorf("payment processing failed: %w", err)
	}

	// Update payment record
	payment.Status = entities.PaymentStatusCompleted
	payment.TransactionID = result.TransactionID
	payment.ProcessedAt = timePtr(time.Now())
	s.paymentRepo.Create(ctx, payment)

	// Update invoice
	invoice.Status = entities.InvoiceStatusPaid
	invoice.PaidAt = timePtr(time.Now())
	s.invoiceRepo.Update(ctx, invoice)

	// Publish payment success event
	s.publishPaymentReceivedEvent(ctx, payment, invoice)

	return nil
}

// Helper methods

func (s *EventIntegratedFinancialService) handleFinancialDecision(ctx context.Context, decisionID, option string, payload map[string]interface{}) error {
	switch option {
	case "freeze_transactions":
		// Implement transaction freeze
		amount, _ := payload["freeze_threshold"].(float64)
		if amount == 0 {
			amount = 10000 // Default $10k
		}
		
		s.eventBus.PublishEvent(ctx, "financial.transactions_frozen", map[string]interface{}{
			"decision_id": decisionID,
			"threshold":   amount,
			"duration":    "24h",
		})

	case "increase_verification":
		// Implement enhanced verification
		s.eventBus.PublishEvent(ctx, "financial.verification_enhanced", map[string]interface{}{
			"decision_id": decisionID,
			"level":       "high",
		})
	}

	return nil
}

func (s *EventIntegratedFinancialService) handlePricingDecision(ctx context.Context, decisionID, option string, payload map[string]interface{}) error {
	// Update pricing engine parameters based on decision
	switch option {
	case "adjust_pricing":
		adjustment, _ := payload["price_adjustment"].(float64)
		s.pricingEngine.AdjustPricing(ctx, adjustment)
		
	case "enable_surge_pricing":
		s.pricingEngine.EnableSurgePricing(ctx, true)
	}

	return nil
}

func (s *EventIntegratedFinancialService) getApprovedContentCount(ctx context.Context, projectID uuid.UUID) int {
	// This would query the content repository
	// For now, return a mock value
	return 5
}

// Event publishing methods

func (s *EventIntegratedFinancialService) publishInvoiceCreatedEvent(ctx context.Context, invoice *entities.Invoice) {
	eventData := events.FinancialEventData{
		TransactionID: invoice.ID.String(),
		InvoiceID:     invoice.ID.String(),
		ClientID:      invoice.ClientID.String(),
		ProjectID:     invoice.ProjectID.String(),
		Amount:        invoice.Amount,
		Currency:      invoice.Currency,
		Status:        string(invoice.Status),
		Metadata:      invoice.Metadata,
	}

	event := events.CreateFinancialEvent(events.EventInvoiceCreated, "financial-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[FinancialService] Failed to publish invoice created event: %v", err)
	}
}

func (s *EventIntegratedFinancialService) publishPaymentReceivedEvent(ctx context.Context, payment *entities.Payment, invoice *entities.Invoice) {
	eventData := events.FinancialEventData{
		TransactionID: payment.TransactionID,
		InvoiceID:     invoice.ID.String(),
		ClientID:      payment.ClientID.String(),
		ProjectID:     invoice.ProjectID.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        string(payment.Status),
		PaymentMethod: string(payment.Method),
		Metadata: map[string]interface{}{
			"payment_id": payment.ID.String(),
		},
	}

	event := events.CreateFinancialEvent(events.EventPaymentReceived, "financial-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[FinancialService] Failed to publish payment received event: %v", err)
	}
}

func (s *EventIntegratedFinancialService) publishPaymentFailedEvent(ctx context.Context, payment *entities.Payment, invoice *entities.Invoice) {
	eventData := events.FinancialEventData{
		TransactionID: payment.TransactionID,
		InvoiceID:     invoice.ID.String(),
		ClientID:      payment.ClientID.String(),
		ProjectID:     invoice.ProjectID.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        string(payment.Status),
		PaymentMethod: string(payment.Method),
		Metadata: map[string]interface{}{
			"payment_id":     payment.ID.String(),
			"failure_reason": payment.FailureReason,
		},
	}

	event := events.CreateFinancialEvent(events.EventPaymentFailed, "financial-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[FinancialService] Failed to publish payment failed event: %v", err)
	}
}

// Utility function
func timePtr(t time.Time) *time.Time {
	return &t
}
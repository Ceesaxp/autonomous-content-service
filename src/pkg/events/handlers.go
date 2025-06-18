package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ServiceEventHandler defines the interface for handling events in services  
type ServiceEventHandler interface {
	Handle(ctx context.Context, event Event) error
	GetEventTypes() []string
	GetServiceName() string
}

// BaseEventHandler provides common functionality for event handlers
type BaseEventHandler struct {
	ServiceName string
	EventTypes  []string
}

func (h *BaseEventHandler) GetEventTypes() []string {
	return h.EventTypes
}

func (h *BaseEventHandler) GetServiceName() string {
	return h.ServiceName
}

// ContentServiceServiceEventHandler handles events for the content service
type ContentServiceServiceEventHandler struct {
	BaseEventHandler
}

func NewContentServiceServiceEventHandler() *ContentServiceServiceEventHandler {
	return &ContentServiceServiceEventHandler{
		BaseEventHandler: BaseEventHandler{
			ServiceName: "content-service",
			EventTypes: []string{
				EventProjectCreated,
				EventProjectUpdated,
				EventClientOnboarded,
				EventClientFeedback,
				EventDecisionExecuted,
				EventSystemAlert,
			},
		},
	}
}

func (h *ContentServiceServiceEventHandler) Handle(ctx context.Context, event Event) error {
	log.Printf("[ContentService] Processing event %s from %s", event.Type, event.Source)

	switch event.Type {
	case EventProjectCreated:
		return h.handleProjectCreated(ctx, event)
	case EventProjectUpdated:
		return h.handleProjectUpdated(ctx, event)
	case EventClientOnboarded:
		return h.handleClientOnboarded(ctx, event)
	case EventClientFeedback:
		return h.handleClientFeedback(ctx, event)
	case EventDecisionExecuted:
		return h.handleDecisionExecuted(ctx, event)
	case EventSystemAlert:
		return h.handleSystemAlert(ctx, event)
	default:
		log.Printf("[ContentService] Unhandled event type: %s", event.Type)
		return nil
	}
}

func (h *ContentServiceServiceEventHandler) handleProjectCreated(ctx context.Context, event Event) error {
	var data ProjectEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse project data: %w", err)
	}

	// Initialize content creation pipeline for the new project
	log.Printf("[ContentService] Initializing content pipeline for project %s", data.ProjectID)
	
	// TODO: Implement actual content pipeline initialization
	// - Create content templates based on project requirements
	// - Set up quality checkpoints
	// - Initialize LLM context for the project
	
	return nil
}

func (h *ContentServiceServiceEventHandler) handleProjectUpdated(ctx context.Context, event Event) error {
	var data ProjectEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse project data: %w", err)
	}

	log.Printf("[ContentService] Project %s updated with status %s", data.ProjectID, data.Status)
	
	// TODO: Implement project update handling
	// - Adjust content pipeline based on status changes
	// - Update quality thresholds if budget changed
	// - Reschedule content delivery if deadline changed
	
	return nil
}

func (h *ContentServiceServiceEventHandler) handleClientOnboarded(ctx context.Context, event Event) error {
	var data ClientEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	log.Printf("[ContentService] Setting up content preferences for client %s", data.ClientID)
	
	// TODO: Implement client onboarding for content service
	// - Initialize client-specific content templates
	// - Set up brand voice and style guidelines
	// - Configure quality preferences based on tier
	
	return nil
}

func (h *ContentServiceServiceEventHandler) handleClientFeedback(ctx context.Context, event Event) error {
	var data ClientEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	log.Printf("[ContentService] Processing feedback for client %s (satisfaction: %.2f)", data.ClientID, data.Satisfaction)
	
	// TODO: Implement feedback processing
	// - Adjust content quality parameters based on feedback
	// - Update client preferences and style guidelines
	// - Train content models with feedback data
	
	return nil
}

func (h *ContentServiceServiceEventHandler) handleDecisionExecuted(ctx context.Context, event Event) error {
	var data DecisionEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse decision data: %w", err)
	}

	log.Printf("[ContentService] Processing executed decision %s of type %s", data.DecisionID, data.Type)
	
	// TODO: Implement decision execution handling
	// - Apply content policy changes
	// - Adjust quality thresholds
	// - Implement new content creation strategies
	
	return nil
}

func (h *ContentServiceServiceEventHandler) handleSystemAlert(ctx context.Context, event Event) error {
	var data SystemEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse system data: %w", err)
	}

	log.Printf("[ContentService] System alert from %s: %s (level: %s)", data.Service, data.Message, data.AlertLevel)
	
	// TODO: Implement system alert handling
	// - Adjust content creation capacity based on system load
	// - Switch to backup LLM providers if needed
	// - Implement degraded service mode
	
	return nil
}

// DecisionServiceServiceEventHandler handles events for the decision service
type DecisionServiceServiceEventHandler struct {
	BaseEventHandler
}

func NewDecisionServiceServiceEventHandler() *DecisionServiceServiceEventHandler {
	return &DecisionServiceServiceEventHandler{
		BaseEventHandler: BaseEventHandler{
			ServiceName: "decision-service",
			EventTypes: []string{
				EventRiskDetected,
				EventIncidentCreated,
				EventComplianceIssue,
				EventSystemAlert,
				EventProposalCreated,
				EventClientFeedback,
			},
		},
	}
}

func (h *DecisionServiceServiceEventHandler) Handle(ctx context.Context, event Event) error {
	log.Printf("[DecisionService] Processing event %s from %s", event.Type, event.Source)

	switch event.Type {
	case EventRiskDetected:
		return h.handleRiskDetected(ctx, event)
	case EventIncidentCreated:
		return h.handleIncidentCreated(ctx, event)
	case EventComplianceIssue:
		return h.handleComplianceIssue(ctx, event)
	case EventSystemAlert:
		return h.handleSystemAlert(ctx, event)
	case EventProposalCreated:
		return h.handleProposalCreated(ctx, event)
	case EventClientFeedback:
		return h.handleClientFeedback(ctx, event)
	default:
		log.Printf("[DecisionService] Unhandled event type: %s", event.Type)
		return nil
	}
}

func (h *DecisionServiceServiceEventHandler) handleRiskDetected(ctx context.Context, event Event) error {
	var data RiskEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse risk data: %w", err)
	}

	log.Printf("[DecisionService] Risk detected: %s (severity: %s, score: %.2f)", data.RiskID, data.Severity, data.Score)
	
	// TODO: Implement risk-based decision making
	// - Create risk mitigation decisions
	// - Adjust operational parameters based on risk level
	// - Escalate to emergency protocols if needed
	
	return nil
}

func (h *DecisionServiceServiceEventHandler) handleIncidentCreated(ctx context.Context, event Event) error {
	var data RiskEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse incident data: %w", err)
	}

	log.Printf("[DecisionService] Incident created: %s (severity: %s)", data.IncidentID, data.Severity)
	
	// TODO: Implement incident response decisions
	// - Create emergency response decisions
	// - Allocate resources for incident resolution
	// - Implement communication decisions
	
	return nil
}

func (h *DecisionServiceServiceEventHandler) handleComplianceIssue(ctx context.Context, event Event) error {
	var data LegalEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse legal data: %w", err)
	}

	log.Printf("[DecisionService] Compliance issue detected: %s", data.ComplianceIssue)
	
	// TODO: Implement compliance-based decisions
	// - Create compliance remediation decisions
	// - Adjust policies and procedures
	// - Implement preventive measures
	
	return nil
}

func (h *DecisionServiceServiceEventHandler) handleSystemAlert(ctx context.Context, event Event) error {
	var data SystemEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse system data: %w", err)
	}

	log.Printf("[DecisionService] System alert: %s (level: %s)", data.Message, data.AlertLevel)
	
	// TODO: Implement system alert decisions
	// - Create operational adjustment decisions
	// - Implement resource reallocation decisions
	// - Create maintenance and optimization decisions
	
	return nil
}

func (h *DecisionServiceServiceEventHandler) handleProposalCreated(ctx context.Context, event Event) error {
	var data GovernanceEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse governance data: %w", err)
	}

	log.Printf("[DecisionService] Governance proposal created: %s (type: %s)", data.ProposalID, data.Type)
	
	// TODO: Implement governance proposal analysis
	// - Analyze proposal impact
	// - Create recommendation decisions
	// - Assess implementation feasibility
	
	return nil
}

func (h *DecisionServiceServiceEventHandler) handleClientFeedback(ctx context.Context, event Event) error {
	var data ClientEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	log.Printf("[DecisionService] Client feedback received for %s (satisfaction: %.2f)", data.ClientID, data.Satisfaction)
	
	// TODO: Implement feedback-based decisions
	// - Create service improvement decisions
	// - Adjust client relationship strategies
	// - Implement quality enhancement decisions
	
	return nil
}

// FinancialServiceServiceEventHandler handles events for the financial service
type FinancialServiceServiceEventHandler struct {
	BaseEventHandler
}

func NewFinancialServiceServiceEventHandler() *FinancialServiceServiceEventHandler {
	return &FinancialServiceServiceEventHandler{
		BaseEventHandler: BaseEventHandler{
			ServiceName: "financial-service",
			EventTypes: []string{
				EventProjectCreated,
				EventProjectCompleted,
				EventContentApproved,
				EventClientOnboarded,
				EventDecisionExecuted,
				EventRiskDetected,
			},
		},
	}
}

func (h *FinancialServiceServiceEventHandler) Handle(ctx context.Context, event Event) error {
	log.Printf("[FinancialService] Processing event %s from %s", event.Type, event.Source)

	switch event.Type {
	case EventProjectCreated:
		return h.handleProjectCreated(ctx, event)
	case EventProjectCompleted:
		return h.handleProjectCompleted(ctx, event)
	case EventContentApproved:
		return h.handleContentApproved(ctx, event)
	case EventClientOnboarded:
		return h.handleClientOnboarded(ctx, event)
	case EventDecisionExecuted:
		return h.handleDecisionExecuted(ctx, event)
	case EventRiskDetected:
		return h.handleRiskDetected(ctx, event)
	default:
		log.Printf("[FinancialService] Unhandled event type: %s", event.Type)
		return nil
	}
}

func (h *FinancialServiceServiceEventHandler) handleProjectCreated(ctx context.Context, event Event) error {
	var data ProjectEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse project data: %w", err)
	}

	log.Printf("[FinancialService] Creating budget allocation for project %s (budget: %.2f)", data.ProjectID, data.Budget)
	
	// TODO: Implement project financial setup
	// - Create budget allocation
	// - Set up milestone-based invoicing
	// - Initialize payment tracking
	
	return nil
}

func (h *FinancialServiceServiceEventHandler) handleProjectCompleted(ctx context.Context, event Event) error {
	var data ProjectEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse project data: %w", err)
	}

	log.Printf("[FinancialService] Processing completion for project %s", data.ProjectID)
	
	// TODO: Implement project completion financial processing
	// - Generate final invoice
	// - Process final payments
	// - Update revenue accounting
	
	return nil
}

func (h *FinancialServiceServiceEventHandler) handleContentApproved(ctx context.Context, event Event) error {
	var data ContentEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse content data: %w", err)
	}

	log.Printf("[FinancialService] Content approved for project %s, triggering milestone payment", data.ProjectID)
	
	// TODO: Implement milestone payment processing
	// - Trigger milestone invoice
	// - Update project financials
	// - Process automatic payments
	
	return nil
}

func (h *FinancialServiceServiceEventHandler) handleClientOnboarded(ctx context.Context, event Event) error {
	var data ClientEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	log.Printf("[FinancialService] Setting up financial profile for client %s (tier: %s)", data.ClientID, data.Tier)
	
	// TODO: Implement client financial setup
	// - Create payment profile
	// - Set up billing preferences
	// - Configure tier-based pricing
	
	return nil
}

func (h *FinancialServiceServiceEventHandler) handleDecisionExecuted(ctx context.Context, event Event) error {
	var data DecisionEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse decision data: %w", err)
	}

	log.Printf("[FinancialService] Processing financial impact of decision %s", data.DecisionID)
	
	// TODO: Implement decision-based financial adjustments
	// - Apply pricing changes
	// - Adjust budget allocations
	// - Implement cost optimization decisions
	
	return nil
}

func (h *FinancialServiceServiceEventHandler) handleRiskDetected(ctx context.Context, event Event) error {
	var data RiskEventData
	if err := mapToStruct(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to parse risk data: %w", err)
	}

	if data.Category == "financial" {
		log.Printf("[FinancialService] Financial risk detected: %s (score: %.2f)", data.RiskID, data.Score)
		
		// TODO: Implement financial risk response
		// - Adjust payment terms
		// - Implement additional verification
		// - Update risk-based pricing
	}
	
	return nil
}

// Helper function to convert map to struct
func mapToStruct(data map[string]interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, result)
}

// ServiceEventHandlerRegistry manages event handlers for different services
type ServiceEventHandlerRegistry struct {
	handlers map[string]ServiceEventHandler
}

func NewServiceEventHandlerRegistry() *ServiceEventHandlerRegistry {
	registry := &ServiceEventHandlerRegistry{
		handlers: make(map[string]ServiceEventHandler),
	}
	
	// Register default handlers
	registry.RegisterHandler(NewContentServiceServiceEventHandler())
	registry.RegisterHandler(NewDecisionServiceServiceEventHandler())
	registry.RegisterHandler(NewFinancialServiceServiceEventHandler())
	
	return registry
}

func (r *ServiceEventHandlerRegistry) RegisterHandler(handler ServiceEventHandler) {
	r.handlers[handler.GetServiceName()] = handler
}

func (r *ServiceEventHandlerRegistry) GetHandler(serviceName string) (ServiceEventHandler, bool) {
	handler, exists := r.handlers[serviceName]
	return handler, exists
}

func (r *ServiceEventHandlerRegistry) GetAllHandlers() map[string]ServiceEventHandler {
	return r.handlers
}

// EventRouter routes events to appropriate handlers based on service and event type
type EventRouter struct {
	registry *ServiceEventHandlerRegistry
}

func NewEventRouter() *EventRouter {
	return &EventRouter{
		registry: NewServiceEventHandlerRegistry(),
	}
}

func (r *EventRouter) RouteEvent(ctx context.Context, serviceName string, event Event) error {
	handler, exists := r.registry.GetHandler(serviceName)
	if !exists {
		log.Printf("[EventRouter] No handler found for service: %s", serviceName)
		return fmt.Errorf("no handler found for service: %s", serviceName)
	}

	// Check if handler supports this event type
	supportedTypes := handler.GetEventTypes()
	supported := false
	for _, eventType := range supportedTypes {
		if eventType == event.Type {
			supported = true
			break
		}
	}

	if !supported {
		log.Printf("[EventRouter] Handler %s does not support event type: %s", serviceName, event.Type)
		return nil // Not an error, just not supported
	}

	return handler.Handle(ctx, event)
}
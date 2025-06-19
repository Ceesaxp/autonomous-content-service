package decision_making

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/events"
	"github.com/google/uuid"
)

// EventIntegratedDecisionService wraps the decision engine with event-driven capabilities
type EventIntegratedDecisionService struct {
	engine            DecisionEngine
	eventBus          *events.ServiceEventBus
	decisionRepo      repositories.DecisionRepository
	policyEnforcer    PolicyEnforcer
	ethicalFramework  EthicalFramework
}

// NewEventIntegratedDecisionService creates a new event-integrated decision service
func NewEventIntegratedDecisionService(
	engine DecisionEngine,
	eventBus *events.ServiceEventBus,
	decisionRepo repositories.DecisionRepository,
	policyEnforcer PolicyEnforcer,
	ethicalFramework EthicalFramework,
) *EventIntegratedDecisionService {
	return &EventIntegratedDecisionService{
		engine:           engine,
		eventBus:         eventBus,
		decisionRepo:     decisionRepo,
		policyEnforcer:   policyEnforcer,
		ethicalFramework: ethicalFramework,
	}
}

// HandleRiskDetected creates and executes decisions based on detected risks
func (s *EventIntegratedDecisionService) HandleRiskDetected(ctx context.Context, event events.Event) error {
	riskID, _ := event.Payload["risk_id"].(string)
	category, _ := event.Payload["category"].(string)
	severity, _ := event.Payload["severity"].(string)
	score, _ := event.Payload["score"].(float64)

	log.Printf("[DecisionService] Creating decision for risk %s (category: %s, severity: %s, score: %.2f)", 
		riskID, category, severity, score)

	// Create decision based on risk
	var decisionType entities.DecisionType
	var priority entities.DecisionPriority

	// Map severity to priority
	switch severity {
	case "critical":
		decisionType = entities.DecisionTypeEmergency
		priority = entities.PriorityCritical
	case "high":
		decisionType = entities.DecisionTypeOperational
		priority = entities.PriorityHigh
	case "medium":
		decisionType = entities.DecisionTypeOperational
		priority = entities.PriorityMedium
	default:
		decisionType = entities.DecisionTypeOperational
		priority = entities.PriorityLow
	}

	// Create decision
	decision := &entities.Decision{
		ID:          uuid.New().String(),
		Type:        decisionType,
		Priority:    priority,
		Status:      entities.StatusPending,
		Title:       fmt.Sprintf("Response to %s risk in %s", severity, category),
		Description: fmt.Sprintf("Automated decision for risk %s with score %.2f", riskID, score),
		Context: map[string]interface{}{
			"risk_id":   riskID,
			"category":  category,
			"severity":  severity,
			"score":     score,
			"source":    "risk_management",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate options based on risk type
	options := s.generateRiskResponseOptions(category, severity, score)
	decision.Options = options

	// Analyze options using the engine
	err := s.engine.AnalyzeOptions(ctx, decision)
	if err != nil {
		return fmt.Errorf("failed to analyze options: %w", err)
	}
	// Simulate decision selection (in real implementation, this would be done by the engine)
	if len(decision.Options) > 0 {
		decision.SelectedOption = &decision.Options[0]
		decision.Confidence = 0.8
		decision.Reasoning = "Selected highest priority response option"
	}

	// Check policies
	policyResult, err := s.policyEnforcer.ValidateDecision(ctx, decision)
	if err != nil {
		log.Printf("[DecisionService] Failed to validate policies: %v", err)
	}
	violations := policyResult.Violations
	if len(violations) > 0 {
		decision.PolicyViolations = violations
		if s.hasBlockingViolations(violations) {
			decision.Status = entities.StatusRejected
			s.publishDecisionRejectedEvent(ctx, decision, "Policy violations detected")
			return nil
		}
	}

	// Check ethics
	ethicalResult, err := s.ethicalFramework.ValidateEthics(ctx, decision)
	if err != nil {
		log.Printf("[DecisionService] Failed to validate ethics: %v", err)
	}
	ethicalIssues := ethicalResult.Concerns
	if len(ethicalIssues) > 0 {
		log.Printf("[DecisionService] Ethical issues detected: %v", ethicalIssues)
		// Add to decision metadata
		decision.Metadata["ethical_issues"] = ethicalIssues
	}

	// Save decision
	if err := s.decisionRepo.CreateDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to save decision: %w", err)
	}

	// Execute decision
	decision.Status = entities.StatusExecuted
	if err := s.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		log.Printf("[DecisionService] Failed to update decision status: %v", err)
	}

	// Publish decision created event
	s.publishDecisionCreatedEvent(ctx, decision)

	// Execute the selected option
	if err := s.executeDecisionOption(ctx, decision, selectedOption); err != nil {
		decision.Status = entities.DecisionStatusFailed
		decision.ExecutionResult = &entities.ExecutionResult{
			Success: false,
			Error:   err.Error(),
			ExecutedAt: time.Now(),
		}
		s.decisionRepo.Update(ctx, decision)
		return fmt.Errorf("failed to execute decision: %w", err)
	}

	// Mark as executed
	decision.Status = entities.DecisionStatusExecuted
	executedAt := time.Now()
	decision.ExecutedAt = &executedAt
	decision.ExecutionResult = &entities.ExecutionResult{
		Success: true,
		ExecutedAt: executedAt,
	}
	s.decisionRepo.Update(ctx, decision)

	// Publish execution event
	s.publishDecisionExecutedEvent(ctx, decision)

	return nil
}

// HandleIncidentCreated creates incident response decisions
func (s *EventIntegratedDecisionService) HandleIncidentCreated(ctx context.Context, event events.Event) error {
	incidentID, _ := event.Payload["incident_id"].(string)
	severity, _ := event.Payload["severity"].(string)
	description, _ := event.Payload["description"].(string)

	log.Printf("[DecisionService] Creating incident response decision for %s (severity: %s)", incidentID, severity)

	// Start incident response workflow
	workflow, err := s.eventBus.StartWorkflow(ctx, "incident_response", map[string]interface{}{
		"incident_id": incidentID,
		"severity":    severity,
	})

	if err != nil {
		return fmt.Errorf("failed to start incident response workflow: %w", err)
	}

	log.Printf("[DecisionService] Started incident response workflow %s", workflow.ID)

	// Create immediate response decision
	decision := &entities.Decision{
		ID:          uuid.New().String(),
		Type:        entities.DecisionTypeEmergency,
		Priority:    entities.PriorityCritical,
		Status:      entities.StatusPending,
		Title:       fmt.Sprintf("Incident Response: %s", incidentID),
		Description: description,
		Context: map[string]interface{}{
			"incident_id": incidentID,
			"severity":    severity,
			"workflow_id": workflow.ID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate incident response options
	options := s.generateIncidentResponseOptions(severity)
	decision.Options = options

	// Fast-track execution for critical incidents
	if severity == "critical" {
		decision.SelectedOption = options[0] // Take most aggressive action
		decision.Confidence = 0.95
		decision.Reasoning = "Critical incident requires immediate automated response"
		
		// Execute immediately
		s.executeDecisionOption(ctx, decision, decision.SelectedOption)
	}

	// Save and publish
	s.decisionRepo.Create(ctx, decision)
	s.publishDecisionCreatedEvent(ctx, decision)

	return nil
}

// HandleComplianceIssue creates compliance remediation decisions
func (s *EventIntegratedDecisionService) HandleComplianceIssue(ctx context.Context, event events.Event) error {
	issue, _ := event.Payload["compliance_issue"].(string)
	contractID, _ := event.Payload["contract_id"].(string)
	clientID, _ := event.Payload["client_id"].(string)

	log.Printf("[DecisionService] Creating compliance decision for issue: %s", issue)

	decision := &entities.Decision{
		ID:          uuid.New().String(),
		Type:        entities.DecisionTypeCompliance,
		Priority:    entities.PriorityHigh,
		Status:      entities.StatusPending,
		Title:       fmt.Sprintf("Compliance Remediation: %s", issue),
		Description: fmt.Sprintf("Address compliance issue: %s", issue),
		Context: map[string]interface{}{
			"compliance_issue": issue,
			"contract_id":      contractID,
			"client_id":        clientID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate compliance options
	options := s.generateComplianceOptions(issue)
	decision.Options = options

	// Evaluate and save
	selectedOption, confidence, reasoning := s.engine.EvaluateOptions(ctx, options, decision.Context)
	decision.SelectedOption = selectedOption
	decision.Confidence = confidence
	decision.Reasoning = reasoning

	s.decisionRepo.Create(ctx, decision)
	s.publishDecisionCreatedEvent(ctx, decision)

	return nil
}

// HandleProposalCreated analyzes governance proposals
func (s *EventIntegratedDecisionService) HandleProposalCreated(ctx context.Context, event events.Event) error {
	proposalID, _ := event.Payload["proposal_id"].(string)
	proposalType, _ := event.Payload["type"].(string)

	log.Printf("[DecisionService] Analyzing governance proposal %s (type: %s)", proposalID, proposalType)

	// Create analysis decision
	decision := &entities.Decision{
		ID:          uuid.New().String(),
		Type:        entities.DecisionTypeStrategic,
		Priority:    entities.PriorityMedium,
		Status:      entities.StatusPending,
		Title:       fmt.Sprintf("Governance Proposal Analysis: %s", proposalID),
		Description: "Analyze governance proposal and provide recommendation",
		Context: map[string]interface{}{
			"proposal_id":   proposalID,
			"proposal_type": proposalType,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Analyze proposal impact
	impactAnalysis := s.analyzeProposalImpact(proposalType)
	decision.ImpactAnalysis = impactAnalysis

	// Generate recommendation options
	options := []entities.DecisionOption{
		{
			ID:          "support",
			Title:       "Support Proposal",
			Description: "Vote in favor of the proposal",
			Benefits:    []string{"Aligns with strategic goals", "Improves system efficiency"},
			Risks:       []string{"Implementation complexity", "Resource requirements"},
		},
		{
			ID:          "oppose",
			Title:       "Oppose Proposal",
			Description: "Vote against the proposal",
			Benefits:    []string{"Maintains stability", "Avoids disruption"},
			Risks:       []string{"Missed opportunity", "Stagnation"},
		},
		{
			ID:          "abstain",
			Title:       "Abstain from Voting",
			Description: "Neither support nor oppose",
			Benefits:    []string{"Neutral position", "More time to evaluate"},
			Risks:       []string{"No influence on outcome"},
		},
	}
	decision.Options = options

	// Make recommendation
	selectedOption, confidence, reasoning := s.engine.EvaluateOptions(ctx, options, decision.Context)
	decision.SelectedOption = selectedOption
	decision.Confidence = confidence
	decision.Reasoning = reasoning

	s.decisionRepo.Create(ctx, decision)
	s.publishDecisionCreatedEvent(ctx, decision)

	return nil
}

// Helper methods

func (s *EventIntegratedDecisionService) generateRiskResponseOptions(category, severity string, score float64) []entities.DecisionOption {
	var options []entities.DecisionOption

	switch category {
	case "financial":
		options = append(options, entities.DecisionOption{
			ID:          "freeze_transactions",
			Title:       "Freeze High-Risk Transactions",
			Description: "Temporarily freeze transactions above threshold",
			Benefits:    []string{"Prevents financial loss", "Allows investigation time"},
			Risks:       []string{"Customer dissatisfaction", "Business disruption"},
		})
		options = append(options, entities.DecisionOption{
			ID:          "increase_verification",
			Title:       "Increase Verification Requirements",
			Description: "Add additional verification steps",
			Benefits:    []string{"Reduces fraud risk", "Maintains operations"},
			Risks:       []string{"Increased friction", "Processing delays"},
		})

	case "operational":
		options = append(options, entities.DecisionOption{
			ID:          "scale_resources",
			Title:       "Scale Resources",
			Description: "Increase capacity to handle load",
			Benefits:    []string{"Maintains performance", "Prevents outages"},
			Risks:       []string{"Increased costs", "Over-provisioning"},
		})
		options = append(options, entities.DecisionOption{
			ID:          "enable_degraded_mode",
			Title:       "Enable Degraded Service Mode",
			Description: "Reduce functionality to maintain core services",
			Benefits:    []string{"Maintains availability", "Reduces load"},
			Risks:       []string{"Reduced functionality", "User experience impact"},
		})

	case "security":
		options = append(options, entities.DecisionOption{
			ID:          "emergency_lockdown",
			Title:       "Emergency Security Lockdown",
			Description: "Restrict access and enable enhanced monitoring",
			Benefits:    []string{"Prevents breach", "Enables investigation"},
			Risks:       []string{"Service disruption", "False positives"},
		})
		options = append(options, entities.DecisionOption{
			ID:          "rotate_credentials",
			Title:       "Rotate Security Credentials",
			Description: "Force rotation of all sensitive credentials",
			Benefits:    []string{"Eliminates compromised credentials", "Security refresh"},
			Risks:       []string{"Service interruption", "Configuration updates needed"},
		})
	}

	// Always include monitoring option
	options = append(options, entities.DecisionOption{
		ID:          "enhanced_monitoring",
		Title:       "Enhanced Monitoring Only",
		Description: "Increase monitoring without immediate action",
		Benefits:    []string{"Gathers more data", "Non-disruptive"},
		Risks:       []string{"Delayed response", "Risk escalation"},
	})

	return options
}

func (s *EventIntegratedDecisionService) generateIncidentResponseOptions(severity string) []entities.DecisionOption {
	options := []entities.DecisionOption{
		{
			ID:          "activate_response_team",
			Title:       "Activate Emergency Response",
			Description: "Trigger automated emergency response procedures",
			Benefits:    []string{"Rapid response", "Coordinated action"},
			Risks:       []string{"Resource intensive", "Potential overreaction"},
		},
		{
			ID:          "isolate_affected_systems",
			Title:       "Isolate Affected Systems",
			Description: "Quarantine affected components",
			Benefits:    []string{"Containment", "Prevents spread"},
			Risks:       []string{"Service degradation", "Recovery complexity"},
		},
		{
			ID:          "initiate_failover",
			Title:       "Initiate System Failover",
			Description: "Switch to backup systems",
			Benefits:    []string{"Maintains availability", "Clean environment"},
			Risks:       []string{"Data synchronization", "Performance impact"},
		},
	}

	if severity == "critical" {
		options = append([]entities.DecisionOption{{
			ID:          "emergency_shutdown",
			Title:       "Emergency System Shutdown",
			Description: "Immediate shutdown of affected systems",
			Benefits:    []string{"Prevents damage", "Full containment"},
			Risks:       []string{"Complete service outage", "Data loss potential"},
		}}, options...)
	}

	return options
}

func (s *EventIntegratedDecisionService) generateComplianceOptions(issue string) []entities.DecisionOption {
	return []entities.DecisionOption{
		{
			ID:          "immediate_remediation",
			Title:       "Immediate Remediation",
			Description: "Fix compliance issue immediately",
			Benefits:    []string{"Quick resolution", "Minimizes exposure"},
			Risks:       []string{"Rushed implementation", "Potential errors"},
		},
		{
			ID:          "phased_remediation",
			Title:       "Phased Remediation Plan",
			Description: "Implement fixes in controlled phases",
			Benefits:    []string{"Controlled rollout", "Testing opportunity"},
			Risks:       []string{"Extended exposure", "Regulatory scrutiny"},
		},
		{
			ID:          "compensating_controls",
			Title:       "Implement Compensating Controls",
			Description: "Add additional controls while planning full fix",
			Benefits:    []string{"Immediate risk reduction", "Time for proper fix"},
			Risks:       []string{"Not full compliance", "Additional complexity"},
		},
	}
}

func (s *EventIntegratedDecisionService) analyzeProposalImpact(proposalType string) *entities.ImpactAnalysis {
	// Simplified impact analysis
	return &entities.ImpactAnalysis{
		FinancialImpact: entities.ImpactLevel("medium"),
		OperationalImpact: entities.ImpactLevel("low"),
		ReputationalImpact: entities.ImpactLevel("low"),
		TimelineImpact: "1-3 months implementation",
		ResourceRequirements: []string{"Development time", "Testing resources"},
	}
}

func (s *EventIntegratedDecisionService) hasBlockingViolations(violations []entities.PolicyViolation) bool {
	for _, v := range violations {
		if v.Severity == "critical" {
			return true
		}
	}
	return false
}

func (s *EventIntegratedDecisionService) executeDecisionOption(ctx context.Context, decision *entities.Decision, option *entities.DecisionOption) error {
	log.Printf("[DecisionService] Executing decision option: %s", option.ID)

	// Publish specific action events based on option
	switch option.ID {
	case "freeze_transactions":
		return s.eventBus.PublishEvent(ctx, "financial.freeze_requested", map[string]interface{}{
			"decision_id": decision.ID,
			"reason":      decision.Title,
		})

	case "scale_resources":
		return s.eventBus.PublishEvent(ctx, "system.scale_requested", map[string]interface{}{
			"decision_id": decision.ID,
			"scale_up":    true,
		})

	case "emergency_lockdown":
		return s.eventBus.PublishEvent(ctx, "security.lockdown_activated", map[string]interface{}{
			"decision_id": decision.ID,
			"severity":    "critical",
		})

	case "activate_response_team":
		return s.eventBus.PublishEvent(ctx, "incident.response_activated", map[string]interface{}{
			"decision_id":  decision.ID,
			"incident_id":  decision.Context["incident_id"],
		})

	default:
		log.Printf("[DecisionService] No specific execution handler for option: %s", option.ID)
		return nil
	}
}

// Event publishing methods

func (s *EventIntegratedDecisionService) publishDecisionCreatedEvent(ctx context.Context, decision *entities.Decision) {
	eventData := events.DecisionEventData{
		DecisionID:      decision.ID,
		Type:            string(decision.Type),
		Priority:        string(decision.Priority),
		Status:          string(decision.Status),
		SelectedOption:  decision.SelectedOption.ID,
		ConfidenceScore: decision.Confidence,
		Justification:   decision.Reasoning,
		Metadata:        decision.Metadata,
	}

	event := events.CreateDecisionEvent(events.EventDecisionCreated, "decision-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[DecisionService] Failed to publish decision created event: %v", err)
	}
}

func (s *EventIntegratedDecisionService) publishDecisionExecutedEvent(ctx context.Context, decision *entities.Decision) {
	eventData := events.DecisionEventData{
		DecisionID:      decision.ID,
		Type:            string(decision.Type),
		Priority:        string(decision.Priority),
		Status:          string(decision.Status),
		SelectedOption:  decision.SelectedOption.ID,
		ConfidenceScore: decision.Confidence,
		ImpactScore:     0.0, // Calculate from impact analysis
		Justification:   decision.Reasoning,
		Metadata:        decision.Metadata,
	}

	event := events.CreateDecisionEvent(events.EventDecisionExecuted, "decision-service", eventData)
	if err := s.eventBus.PublishTypedEvent(ctx, event); err != nil {
		log.Printf("[DecisionService] Failed to publish decision executed event: %v", err)
	}
}

func (s *EventIntegratedDecisionService) publishDecisionRejectedEvent(ctx context.Context, decision *entities.Decision, reason string) {
	s.eventBus.PublishEvent(ctx, "decision.rejected", map[string]interface{}{
		"decision_id": decision.ID,
		"reason":      reason,
		"violations":  decision.PolicyViolations,
	})
}
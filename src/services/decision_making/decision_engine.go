package decision_making

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// DecisionEngineImpl implements the DecisionEngine interface
type DecisionEngineImpl struct {
	decisionRepo     repositories.DecisionRepository
	eventRepo        repositories.EventRepository
	policyEnforcer   PolicyEnforcer
	ethicalFramework EthicalFramework
	impactAnalyzer   ImpactAnalyzer
	conflictResolver ConflictResolver
	llmClient        LLMClient // Interface for LLM integration
}

// NewDecisionEngine creates a new decision engine instance
func NewDecisionEngine(
	decisionRepo repositories.DecisionRepository,
	eventRepo repositories.EventRepository,
	policyEnforcer PolicyEnforcer,
	ethicalFramework EthicalFramework,
	impactAnalyzer ImpactAnalyzer,
	conflictResolver ConflictResolver,
	llmClient LLMClient,
) *DecisionEngineImpl {
	return &DecisionEngineImpl{
		decisionRepo:     decisionRepo,
		eventRepo:        eventRepo,
		policyEnforcer:   policyEnforcer,
		ethicalFramework: ethicalFramework,
		impactAnalyzer:   impactAnalyzer,
		conflictResolver: conflictResolver,
		llmClient:        llmClient,
	}
}

// InitiateDecision starts a new decision-making process
func (de *DecisionEngineImpl) InitiateDecision(ctx context.Context, request DecisionRequest) (*entities.Decision, error) {
	// Create new decision
	decision := &entities.Decision{
		ID:          uuid.New().String(),
		Type:        request.Type,
		Priority:    request.Priority,
		Status:      entities.StatusPending,
		Title:       request.Title,
		Description: request.Description,
		Context:     request.Context,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Generate initial options based on decision type
	options, err := de.generateDecisionOptions(ctx, decision)
	if err != nil {
		return nil, fmt.Errorf("failed to generate options: %w", err)
	}
	decision.Options = options

	// Save decision
	if err := de.decisionRepo.CreateDecision(ctx, decision); err != nil {
		return nil, fmt.Errorf("failed to create decision: %w", err)
	}

	// Emit event
	event := &events.DecisionInitiated{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.initiated",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:   decision.ID,
		DecisionType: string(decision.Type),
		Priority:     string(decision.Priority),
		Context:      decision.Context,
		Requester:    "system",
	}
	if err := de.eventRepo.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return decision, nil
}

// AnalyzeOptions evaluates all available options for a decision
func (de *DecisionEngineImpl) AnalyzeOptions(ctx context.Context, decision *entities.Decision) error {
	startTime := time.Now()
	decision.Status = entities.StatusAnalyzing

	// Analyze each option
	for i := range decision.Options {
		option := &decision.Options[i]

		// Score based on multiple factors
		score, err := de.scoreOption(ctx, decision, option)
		if err != nil {
			return fmt.Errorf("failed to score option %s: %w", option.ID, err)
		}
		option.Score = score

		// Identify risks and benefits
		risks, benefits, err := de.analyzeRisksAndBenefits(ctx, decision, option)
		if err != nil {
			return fmt.Errorf("failed to analyze risks/benefits: %w", err)
		}
		option.Risks = risks
		option.Benefits = benefits
	}

	// Sort options by score
	sort.Slice(decision.Options, func(i, j int) bool {
		return decision.Options[i].Score > decision.Options[j].Score
	})

	// Calculate overall confidence
	decision.ConfidenceScore = de.calculateConfidence(decision.Options)

	// Perform impact analysis on top options
	if len(decision.Options) > 0 {
		impact, err := de.impactAnalyzer.AnalyzeImpact(ctx, decision)
		if err != nil {
			return fmt.Errorf("failed to analyze impact: %w", err)
		}
		decision.ImpactAnalysis = impact
	}

	// Update decision
	decision.UpdatedAt = time.Now()
	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	// Emit event
	analysisTime := time.Since(startTime).Milliseconds()
	topOptions := []string{}
	for i := 0; i < min(3, len(decision.Options)); i++ {
		topOptions = append(topOptions, decision.Options[i].ID)
	}

	event := &events.DecisionAnalyzed{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.analyzed",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:       decision.ID,
		OptionsEvaluated: len(decision.Options),
		TopOptions:       topOptions,
		ConfidenceScore:  decision.ConfidenceScore,
		AnalysisDuration: analysisTime,
	}
	return de.eventRepo.Save(ctx, event)
}

// MakeDecision selects the best option and finalizes the decision
func (de *DecisionEngineImpl) MakeDecision(ctx context.Context, decisionID string) (*entities.Decision, error) {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	if decision.Status != entities.StatusAnalyzing && decision.Status != entities.StatusPending {
		return nil, fmt.Errorf("decision not in correct state for making: %s", decision.Status)
	}

	// Validate against policies
	policyResult, err := de.policyEnforcer.ValidateDecision(ctx, decision)
	if err != nil {
		return nil, fmt.Errorf("policy validation failed: %w", err)
	}

	if !policyResult.Compliant {
		decision.PolicyViolations = policyResult.Violations
		decision.Status = entities.StatusRejected
		decision.Justification = "Decision rejected due to policy violations"

		// Save and emit violation events
		if err := de.handlePolicyViolations(ctx, decision, policyResult); err != nil {
			return nil, err
		}

		return decision, nil
	}

	// Validate against ethical guidelines
	ethicalResult, err := de.ethicalFramework.ValidateEthics(ctx, decision)
	if err != nil {
		return nil, fmt.Errorf("ethical validation failed: %w", err)
	}

	if !ethicalResult.Approved {
		decision.Status = entities.StatusRejected
		decision.Justification = fmt.Sprintf("Decision rejected due to ethical concerns: %s", ethicalResult.Justification)

		// Emit ethical concern events
		if err := de.handleEthicalConcerns(ctx, decision, ethicalResult); err != nil {
			return nil, err
		}

		return decision, nil
	}

	// Select best option
	if len(decision.Options) == 0 {
		return nil, fmt.Errorf("no options available for decision")
	}

	selectedOption := decision.Options[0] // Already sorted by score
	decision.SelectedOption = &selectedOption

	// Generate justification
	justification, err := de.generateJustification(ctx, decision, &selectedOption)
	if err != nil {
		return nil, fmt.Errorf("failed to generate justification: %w", err)
	}
	decision.Justification = justification

	// Check if auto-approval is allowed
	autoApproved := de.canAutoApprove(decision)
	if autoApproved {
		decision.Status = entities.StatusApproved
	} else {
		decision.Status = entities.StatusPending // Requires manual approval
	}

	// Update decision
	decision.UpdatedAt = time.Now()
	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return nil, fmt.Errorf("failed to update decision: %w", err)
	}

	// Emit event
	event := &events.DecisionMade{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.made",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:       decision.ID,
		SelectedOptionID: selectedOption.ID,
		Justification:    decision.Justification,
		ConfidenceScore:  decision.ConfidenceScore,
		AutoApproved:     autoApproved,
	}
	if err := de.eventRepo.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Create audit log
	log := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decision.ID,
		Timestamp:   time.Now(),
		EventType:   "decision_made",
		Description: fmt.Sprintf("Decision made: %s", selectedOption.Title),
		Actor:       "system",
		Changes: map[string]interface{}{
			"selected_option": selectedOption.ID,
			"confidence":      decision.ConfidenceScore,
			"auto_approved":   autoApproved,
		},
	}
	if err := de.decisionRepo.CreateDecisionLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return decision, nil
}

// ExecuteDecision puts the decision into action
func (de *DecisionEngineImpl) ExecuteDecision(ctx context.Context, decisionID string) (*entities.ExecutionResult, error) {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	if decision.Status != entities.StatusApproved {
		return nil, fmt.Errorf("decision not approved for execution")
	}

	if decision.SelectedOption == nil {
		return nil, fmt.Errorf("no option selected for execution")
	}

	startTime := time.Now()
	decision.Status = entities.StatusExecuted

	// Execute based on decision type
	result, err := de.executeDecisionType(ctx, decision)
	if err != nil {
		result = &entities.ExecutionResult{
			Success:      false,
			ErrorMessage: err.Error(),
			Metrics:      map[string]interface{}{},
			SideEffects:  []string{},
			Reversible:   false,
		}
	}

	// Update decision with result
	decision.ExecutionResult = result
	executedAt := time.Now()
	decision.ExecutedAt = &executedAt
	decision.UpdatedAt = time.Now()

	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return nil, fmt.Errorf("failed to update decision: %w", err)
	}

	// Emit event
	executionTime := time.Since(startTime).Milliseconds()
	event := &events.DecisionExecuted{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.executed",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:    decision.ID,
		Success:       result.Success,
		ExecutionTime: executionTime,
		Results:       result.Metrics,
		SideEffects:   result.SideEffects,
	}
	if err := de.eventRepo.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Create audit log
	log := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decision.ID,
		Timestamp:   time.Now(),
		EventType:   "decision_executed",
		Description: fmt.Sprintf("Decision executed: success=%v", result.Success),
		Actor:       "system",
		Changes: map[string]interface{}{
			"execution_result": result,
			"execution_time":   executionTime,
		},
	}
	if err := de.decisionRepo.CreateDecisionLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return result, nil
}

// RevertDecision rolls back a previously executed decision
func (de *DecisionEngineImpl) RevertDecision(ctx context.Context, decisionID string, reason string) error {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("failed to get decision: %w", err)
	}

	if decision.Status != entities.StatusExecuted {
		return fmt.Errorf("decision not in executed state")
	}

	if decision.ExecutionResult == nil || !decision.ExecutionResult.Reversible {
		return fmt.Errorf("decision is not reversible")
	}

	// Execute reversion based on decision type
	revertSuccess := true
	impact := ""

	// Attempt to revert
	if err := de.revertDecisionType(ctx, decision); err != nil {
		revertSuccess = false
		impact = err.Error()
	}

	// Update decision status
	decision.Status = entities.StatusReverted
	decision.UpdatedAt = time.Now()

	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	// Emit event
	event := &events.DecisionReverted{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.reverted",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:    decisionID,
		RevertReason:  reason,
		RevertSuccess: revertSuccess,
		Impact:        impact,
	}
	if err := de.eventRepo.Save(ctx, event); err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// OverrideDecision allows manual intervention in the decision process
func (de *DecisionEngineImpl) OverrideDecision(ctx context.Context, decisionID string, override *entities.DecisionOverride) error {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("failed to get decision: %w", err)
	}

	// Validate override authorization
	if !de.isAuthorizedToOverride(override.AuthorizedBy, decision) {
		return fmt.Errorf("user not authorized to override this decision")
	}

	// Apply override
	decision.Override = override
	decision.Status = entities.StatusOverridden
	decision.UpdatedAt = time.Now()

	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	// Emit event
	originalOptionID := ""
	if decision.SelectedOption != nil {
		originalOptionID = decision.SelectedOption.ID
	}

	event := &events.DecisionOverridden{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.overridden",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:       decisionID,
		OriginalOptionID: originalOptionID,
		OverrideOptionID: "", // Set based on override details
		OverrideReason:   override.Reason,
		AuthorizedBy:     override.AuthorizedBy,
	}
	return de.eventRepo.Save(ctx, event)
}

// EscalateDecision moves a decision to a higher priority or authority level
func (de *DecisionEngineImpl) EscalateDecision(ctx context.Context, decisionID string, reason string) error {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("failed to get decision: %w", err)
	}

	// Increase priority
	switch decision.Priority {
	case entities.PriorityLow:
		decision.Priority = entities.PriorityMedium
	case entities.PriorityMedium:
		decision.Priority = entities.PriorityHigh
	case entities.PriorityHigh:
		decision.Priority = entities.PriorityCritical
	}

	// Update decision
	decision.UpdatedAt = time.Now()
	if err := de.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	// Create audit log
	log := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decision.ID,
		Timestamp:   time.Now(),
		EventType:   "decision_escalated",
		Description: fmt.Sprintf("Decision escalated: %s", reason),
		Actor:       "system",
		Changes: map[string]interface{}{
			"new_priority": string(decision.Priority),
			"reason":       reason,
		},
	}
	return de.decisionRepo.CreateDecisionLog(ctx, log)
}

// AssessDecisionQuality evaluates the outcome of a decision
func (de *DecisionEngineImpl) AssessDecisionQuality(ctx context.Context, decisionID string) (*DecisionQualityReport, error) {
	// Get decision
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	if decision.ExecutionResult == nil {
		return nil, fmt.Errorf("decision not yet executed")
	}

	// Calculate quality metrics
	qualityScore := de.calculateQualityScore(decision)
	strengths := de.identifyStrengths(decision)
	weaknesses := de.identifyWeaknesses(decision)
	lessons := de.extractLessonsLearned(decision)
	improvements := de.suggestImprovements(decision)

	report := &DecisionQualityReport{
		DecisionID:     decisionID,
		QualityScore:   qualityScore,
		Strengths:      strengths,
		Weaknesses:     weaknesses,
		LessonsLearned: lessons,
		Improvements:   improvements,
	}

	// Emit event
	event := &events.DecisionQualityAssessed{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: "decision.quality_assessed",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
		DecisionID:       decisionID,
		QualityScore:     qualityScore,
		ExpectedOutcome:  "success",
		ActualOutcome:    fmt.Sprintf("success=%v", decision.ExecutionResult.Success),
		LessonsLearned:   lessons,
		ImprovementAreas: improvements,
	}
	if err := de.eventRepo.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return report, nil
}

// LearnFromDecision updates the system's decision-making based on outcomes
func (de *DecisionEngineImpl) LearnFromDecision(ctx context.Context, decisionID string) error {
	// Get decision and quality assessment
	decision, err := de.decisionRepo.GetDecision(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("failed to get decision: %w", err)
	}

	qualityReport, err := de.AssessDecisionQuality(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("failed to assess quality: %w", err)
	}

	// Update decision templates based on outcomes
	if qualityReport.QualityScore > 0.8 {
		// High-quality decision - consider creating a template
		template := de.createTemplateFromDecision(decision)
		if err := de.decisionRepo.CreateDecisionTemplate(ctx, template); err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}
	}

	// Update scoring weights based on success/failure
	// This would typically update ML model parameters or rule weights

	// Create learning record
	log := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decision.ID,
		Timestamp:   time.Now(),
		EventType:   "learning_applied",
		Description: "System learned from decision outcome",
		Actor:       "system",
		Changes: map[string]interface{}{
			"quality_score": qualityReport.QualityScore,
			"lessons":       qualityReport.LessonsLearned,
		},
	}
	return de.decisionRepo.CreateDecisionLog(ctx, log)
}

// Helper methods

func (de *DecisionEngineImpl) generateDecisionOptions(ctx context.Context, decision *entities.Decision) ([]entities.DecisionOption, error) {
	// Use templates if available
	templates, err := de.decisionRepo.ListDecisionTemplates(ctx, decision.Type)
	if err == nil && len(templates) > 0 {
		// Use template options as starting point
		return templates[0].DefaultOptions, nil
	}

	// Generate options using LLM
	prompt := fmt.Sprintf(`Generate decision options for:
Type: %s
Title: %s
Description: %s
Context: %v

Provide 3-5 viable options with pros, cons, and constraints.`,
		decision.Type, decision.Title, decision.Description, decision.Context)

	options, err := de.llmClient.GenerateOptions(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate options: %w", err)
	}

	return options, nil
}

func (de *DecisionEngineImpl) scoreOption(ctx context.Context, decision *entities.Decision, option *entities.DecisionOption) (float64, error) {
	// Multi-factor scoring
	scores := make([]float64, 0)

	// Feasibility score
	feasibility := de.assessFeasibility(decision, option)
	scores = append(scores, feasibility*0.3)

	// Impact score
	impact := de.assessImpact(decision, option)
	scores = append(scores, impact*0.25)

	// Risk score (inverted - lower risk is better)
	risk := de.assessRisk(decision, option)
	scores = append(scores, (1.0-risk)*0.2)

	// Alignment score
	alignment := de.assessAlignment(decision, option)
	scores = append(scores, alignment*0.15)

	// Cost efficiency
	efficiency := de.assessEfficiency(decision, option)
	scores = append(scores, efficiency*0.1)

	// Calculate weighted sum
	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	return totalScore, nil
}

func (de *DecisionEngineImpl) analyzeRisksAndBenefits(ctx context.Context, decision *entities.Decision, option *entities.DecisionOption) ([]string, []string, error) {
	// Use LLM to analyze risks and benefits
	prompt := fmt.Sprintf(`Analyze risks and benefits for option:
%s

In the context of:
%s

List specific risks and benefits.`, option.Description, decision.Description)

	analysis, err := de.llmClient.AnalyzeRisksBenefits(ctx, prompt)
	if err != nil {
		return nil, nil, err
	}

	return analysis.Risks, analysis.Benefits, nil
}

func (de *DecisionEngineImpl) calculateConfidence(options []entities.DecisionOption) float64 {
	if len(options) == 0 {
		return 0.0
	}

	// Base confidence on score distribution
	topScore := options[0].Score
	if len(options) > 1 {
		secondScore := options[1].Score
		// Higher confidence when there's clear separation
		separation := topScore - secondScore
		confidence := math.Min(0.5+separation, 1.0)

		// Adjust based on absolute score
		confidence *= topScore

		return confidence
	}

	return topScore * 0.8 // Single option, moderate confidence
}

func (de *DecisionEngineImpl) generateJustification(ctx context.Context, decision *entities.Decision, option *entities.DecisionOption) (string, error) {
	prompt := fmt.Sprintf(`Generate a clear justification for selecting this option:
Option: %s
Score: %.2f
Benefits: %v
Risks: %v

In the context of: %s

Provide a concise justification that explains why this is the best choice.`,
		option.Title, option.Score, option.Benefits, option.Risks, decision.Description)

	justification, err := de.llmClient.GenerateText(ctx, prompt)
	if err != nil {
		return "", err
	}

	return justification, nil
}

func (de *DecisionEngineImpl) canAutoApprove(decision *entities.Decision) bool {
	// Check confidence threshold
	if decision.ConfidenceScore < 0.8 {
		return false
	}

	// Check decision type
	switch decision.Type {
	case entities.DecisionTypeEmergency, entities.DecisionTypeFinancial:
		return false // Never auto-approve these
	case entities.DecisionTypeOperational:
		return decision.ConfidenceScore > 0.9
	default:
		return decision.ConfidenceScore > 0.85
	}
}

func (de *DecisionEngineImpl) handlePolicyViolations(ctx context.Context, decision *entities.Decision, result *PolicyValidationResult) error {
	for _, violation := range result.Violations {
		event := &events.PolicyViolationDetected{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New(),
				EventType: "policy.violation_detected",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{},
			},
			DecisionID:  decision.ID,
			PolicyID:    violation.PolicyID,
			PolicyName:  violation.PolicyName,
			Severity:    violation.Severity,
			Description: violation.Description,
			Violations:  []string{violation.Description},
		}
		if err := de.eventRepo.Save(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (de *DecisionEngineImpl) handleEthicalConcerns(ctx context.Context, decision *entities.Decision, result *EthicalValidationResult) error {
	for _, concern := range result.Concerns {
		event := &events.EthicalConcernRaised{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New(),
				EventType: "ethical.concern_raised",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{},
			},
			DecisionID:   decision.ID,
			GuidelineID:  concern.GuidelineID,
			ConcernLevel: concern.Severity,
			Description:  concern.Concern,
			RedLines:     result.RedLineViolations,
		}
		if err := de.eventRepo.Save(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (de *DecisionEngineImpl) executeDecisionType(ctx context.Context, decision *entities.Decision) (*entities.ExecutionResult, error) {
	// Route to appropriate execution handler based on type
	switch decision.Type {
	case entities.DecisionTypeContent:
		return de.executeContentDecision(ctx, decision)
	case entities.DecisionTypeFinancial:
		return de.executeFinancialDecision(ctx, decision)
	case entities.DecisionTypeClient:
		return de.executeClientDecision(ctx, decision)
	case entities.DecisionTypeOperational:
		return de.executeOperationalDecision(ctx, decision)
	default:
		return nil, fmt.Errorf("unsupported decision type: %s", decision.Type)
	}
}

func (de *DecisionEngineImpl) revertDecisionType(ctx context.Context, decision *entities.Decision) error {
	// Route to appropriate reversion handler based on type
	switch decision.Type {
	case entities.DecisionTypeContent:
		return de.revertContentDecision(ctx, decision)
	case entities.DecisionTypeFinancial:
		return de.revertFinancialDecision(ctx, decision)
	default:
		return fmt.Errorf("reversion not supported for decision type: %s", decision.Type)
	}
}

// Type-specific execution methods (to be implemented based on business logic)
func (de *DecisionEngineImpl) executeContentDecision(ctx context.Context, decision *entities.Decision) (*entities.ExecutionResult, error) {
	// Implement content-specific execution
	return &entities.ExecutionResult{
		Success:     true,
		Metrics:     map[string]interface{}{"content_created": true},
		SideEffects: []string{"content published"},
		Reversible:  true,
	}, nil
}

func (de *DecisionEngineImpl) executeFinancialDecision(ctx context.Context, decision *entities.Decision) (*entities.ExecutionResult, error) {
	// Implement financial-specific execution
	return &entities.ExecutionResult{
		Success:     true,
		Metrics:     map[string]interface{}{"transaction_completed": true},
		SideEffects: []string{"payment processed"},
		Reversible:  false,
	}, nil
}

func (de *DecisionEngineImpl) executeClientDecision(ctx context.Context, decision *entities.Decision) (*entities.ExecutionResult, error) {
	// Implement client-specific execution
	return &entities.ExecutionResult{
		Success:     true,
		Metrics:     map[string]interface{}{"client_updated": true},
		SideEffects: []string{"client notified"},
		Reversible:  true,
	}, nil
}

func (de *DecisionEngineImpl) executeOperationalDecision(ctx context.Context, decision *entities.Decision) (*entities.ExecutionResult, error) {
	// Implement operational-specific execution
	return &entities.ExecutionResult{
		Success:     true,
		Metrics:     map[string]interface{}{"operation_completed": true},
		SideEffects: []string{},
		Reversible:  true,
	}, nil
}

func (de *DecisionEngineImpl) revertContentDecision(ctx context.Context, decision *entities.Decision) error {
	// Implement content-specific reversion
	return nil
}

func (de *DecisionEngineImpl) revertFinancialDecision(ctx context.Context, decision *entities.Decision) error {
	// Financial decisions typically cannot be reverted
	return fmt.Errorf("financial decisions cannot be reverted")
}

// Scoring helper methods
func (de *DecisionEngineImpl) assessFeasibility(decision *entities.Decision, option *entities.DecisionOption) float64 {
	// Assess technical and resource feasibility
	return 0.8 // Placeholder
}

func (de *DecisionEngineImpl) assessImpact(decision *entities.Decision, option *entities.DecisionOption) float64 {
	// Assess positive impact
	return 0.7 // Placeholder
}

func (de *DecisionEngineImpl) assessRisk(decision *entities.Decision, option *entities.DecisionOption) float64 {
	// Assess risk level
	return 0.3 // Placeholder
}

func (de *DecisionEngineImpl) assessAlignment(decision *entities.Decision, option *entities.DecisionOption) float64 {
	// Assess alignment with goals
	return 0.9 // Placeholder
}

func (de *DecisionEngineImpl) assessEfficiency(decision *entities.Decision, option *entities.DecisionOption) float64 {
	// Assess cost/resource efficiency
	return 0.75 // Placeholder
}

// Quality assessment helpers
func (de *DecisionEngineImpl) calculateQualityScore(decision *entities.Decision) float64 {
	if decision.ExecutionResult == nil {
		return 0.0
	}

	score := 0.0
	if decision.ExecutionResult.Success {
		score += 0.5
	}

	// Add confidence factor
	score += decision.ConfidenceScore * 0.3

	// Add speed factor (faster decisions score higher)
	if decision.ExecutedAt != nil {
		duration := decision.ExecutedAt.Sub(decision.CreatedAt)
		if duration < time.Hour {
			score += 0.2
		} else if duration < time.Hour*24 {
			score += 0.1
		}
	}

	return score
}

func (de *DecisionEngineImpl) identifyStrengths(decision *entities.Decision) []string {
	strengths := []string{}

	if decision.ConfidenceScore > 0.8 {
		strengths = append(strengths, "High confidence in decision")
	}

	if decision.ExecutionResult != nil && decision.ExecutionResult.Success {
		strengths = append(strengths, "Successful execution")
	}

	if len(decision.PolicyViolations) == 0 {
		strengths = append(strengths, "Full policy compliance")
	}

	return strengths
}

func (de *DecisionEngineImpl) identifyWeaknesses(decision *entities.Decision) []string {
	weaknesses := []string{}

	if decision.ConfidenceScore < 0.6 {
		weaknesses = append(weaknesses, "Low confidence score")
	}

	if decision.Override != nil {
		weaknesses = append(weaknesses, "Required manual override")
	}

	if len(decision.Options) < 3 {
		weaknesses = append(weaknesses, "Limited options considered")
	}

	return weaknesses
}

func (de *DecisionEngineImpl) extractLessonsLearned(decision *entities.Decision) []string {
	lessons := []string{}

	if decision.ExecutionResult != nil {
		if decision.ExecutionResult.Success {
			lessons = append(lessons, fmt.Sprintf("Option '%s' was effective for %s decisions",
				decision.SelectedOption.Title, decision.Type))
		} else {
			lessons = append(lessons, fmt.Sprintf("Option '%s' failed due to: %s",
				decision.SelectedOption.Title, decision.ExecutionResult.ErrorMessage))
		}
	}

	return lessons
}

func (de *DecisionEngineImpl) suggestImprovements(decision *entities.Decision) []string {
	improvements := []string{}

	if len(decision.Options) < 3 {
		improvements = append(improvements, "Generate more diverse options")
	}

	if decision.ConfidenceScore < 0.7 {
		improvements = append(improvements, "Improve option scoring methodology")
	}

	if decision.Override != nil {
		improvements = append(improvements, "Review criteria to reduce manual overrides")
	}

	return improvements
}

func (de *DecisionEngineImpl) createTemplateFromDecision(decision *entities.Decision) *entities.DecisionTemplate {
	return &entities.DecisionTemplate{
		ID:               uuid.New().String(),
		Name:             fmt.Sprintf("Template from %s", decision.Title),
		Type:             decision.Type,
		Description:      decision.Description,
		RequiredContext:  de.extractRequiredContext(decision),
		DecisionCriteria: de.extractDecisionCriteria(decision),
		DefaultOptions:   decision.Options,
		PolicyChecks:     de.extractPolicyChecks(decision),
		Metadata: map[string]interface{}{
			"source_decision": decision.ID,
			"success_rate":    1.0,
		},
	}
}

func (de *DecisionEngineImpl) extractRequiredContext(decision *entities.Decision) []string {
	// Extract key context fields
	required := []string{}
	for key := range decision.Context {
		required = append(required, key)
	}
	return required
}

func (de *DecisionEngineImpl) extractDecisionCriteria(decision *entities.Decision) []string {
	return []string{
		"feasibility",
		"impact",
		"risk",
		"alignment",
		"efficiency",
	}
}

func (de *DecisionEngineImpl) extractPolicyChecks(decision *entities.Decision) []string {
	// Extract relevant policy IDs
	checks := []string{}
	for _, violation := range decision.PolicyViolations {
		checks = append(checks, violation.PolicyID)
	}
	return checks
}

func (de *DecisionEngineImpl) isAuthorizedToOverride(authorizer string, decision *entities.Decision) bool {
	// Check authorization based on decision type and priority
	// This would integrate with the access control system
	return true // Placeholder
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LLMClient interface for AI integration
type LLMClient interface {
	GenerateOptions(ctx context.Context, prompt string) ([]entities.DecisionOption, error)
	AnalyzeRisksBenefits(ctx context.Context, prompt string) (*RiskBenefitAnalysis, error)
	GenerateText(ctx context.Context, prompt string) (string, error)
}

type RiskBenefitAnalysis struct {
	Risks    []string
	Benefits []string
}

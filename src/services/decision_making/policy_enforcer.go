package decision_making

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// PolicyEnforcerImpl implements the PolicyEnforcer interface
type PolicyEnforcerImpl struct {
	decisionRepo repositories.DecisionRepository
	eventRepo    repositories.EventRepository
	ruleEngine   RuleEngine
}

// NewPolicyEnforcer creates a new policy enforcer instance
func NewPolicyEnforcer(
	decisionRepo repositories.DecisionRepository,
	eventRepo repositories.EventRepository,
) *PolicyEnforcerImpl {
	return &PolicyEnforcerImpl{
		decisionRepo: decisionRepo,
		eventRepo:    eventRepo,
		ruleEngine:   NewRuleEngine(),
	}
}

// ValidateDecision checks a decision against all applicable policies
func (pe *PolicyEnforcerImpl) ValidateDecision(ctx context.Context, decision *entities.Decision) (*PolicyValidationResult, error) {
	// Get applicable policies
	policies, err := pe.GetApplicablePolicies(ctx, decision.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get applicable policies: %w", err)
	}

	violations := []entities.PolicyViolation{}
	warnings := []string{}
	requiredActions := []string{}
	totalScore := 0.0
	policyCount := 0

	// Check each policy
	for _, policy := range policies {
		if !policy.Active {
			continue
		}

		// Check if policy is within effective date range
		if !pe.isPolicyEffective(policy) {
			continue
		}

		policyCount++
		policyPassed := true

		// Evaluate each rule in the policy
		for _, rule := range policy.Rules {
			passed, violation := pe.evaluateRule(decision, policy, rule)
			if !passed {
				policyPassed = false
				if violation != nil {
					violations = append(violations, *violation)
				}

				// Add required actions based on rule
				if action := pe.getRequiredAction(rule); action != "" {
					requiredActions = append(requiredActions, action)
				}
			} else if warning := pe.checkForWarning(decision, rule); warning != "" {
				warnings = append(warnings, warning)
			}
		}

		if policyPassed {
			totalScore += 1.0
		}
	}

	// Calculate compliance score
	complianceScore := 1.0
	if policyCount > 0 {
		complianceScore = totalScore / float64(policyCount)
	}

	result := &PolicyValidationResult{
		Compliant:       len(violations) == 0,
		Violations:      violations,
		Warnings:        warnings,
		RequiredActions: requiredActions,
		ComplianceScore: complianceScore,
	}

	// Log validation result
	pe.logValidation(ctx, decision.ID, result)

	return result, nil
}

// CheckPolicyCompliance performs a quick compliance check for an action
func (pe *PolicyEnforcerImpl) CheckPolicyCompliance(ctx context.Context, action string, context map[string]interface{}) (bool, error) {
	// Get all active policies
	policies, err := pe.decisionRepo.GetActivePolicies(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get active policies: %w", err)
	}

	// Quick check against all policies
	for _, policy := range policies {
		if !pe.isPolicyEffective(policy) {
			continue
		}

		for _, rule := range policy.Rules {
			if pe.doesRuleApplyToAction(rule, action) {
				if !pe.evaluateRuleCondition(rule, context) {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

// GetApplicablePolicies returns policies relevant to a decision type
func (pe *PolicyEnforcerImpl) GetApplicablePolicies(ctx context.Context, decisionType entities.DecisionType) ([]*entities.Policy, error) {
	// Get all active policies
	allPolicies, err := pe.decisionRepo.GetActivePolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active policies: %w", err)
	}

	// Filter by decision type
	applicable := []*entities.Policy{}
	for _, policy := range allPolicies {
		if pe.isPolicyApplicable(policy, decisionType) {
			applicable = append(applicable, policy)
		}
	}

	return applicable, nil
}

// RegisterPolicy adds a new policy to the system
func (pe *PolicyEnforcerImpl) RegisterPolicy(ctx context.Context, policy *entities.Policy) error {
	// Validate policy structure
	if err := pe.validatePolicy(policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	// Set defaults
	policy.ID = uuid.New().String()
	policy.EffectiveFrom = time.Now()
	policy.Active = true

	// Save policy
	if err := pe.decisionRepo.CreatePolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	// Log policy registration
	pe.logPolicyChange(ctx, policy.ID, "registered", map[string]interface{}{
		"name":     policy.Name,
		"category": policy.Category,
	})

	return nil
}

// UpdatePolicy modifies an existing policy
func (pe *PolicyEnforcerImpl) UpdatePolicy(ctx context.Context, policyID string, updates map[string]interface{}) error {
	// Get existing policy
	policy, err := pe.decisionRepo.GetPolicy(ctx, policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				policy.Name = name
			}
		case "description":
			if desc, ok := value.(string); ok {
				policy.Description = desc
			}
		case "priority":
			if priority, ok := value.(int); ok {
				policy.Priority = priority
			}
		case "rules":
			if rules, ok := value.([]entities.PolicyRule); ok {
				policy.Rules = rules
			}
		case "metadata":
			if metadata, ok := value.(map[string]interface{}); ok {
				policy.Metadata = metadata
			}
		}
	}

	// Validate updated policy
	if err := pe.validatePolicy(policy); err != nil {
		return fmt.Errorf("updated policy validation failed: %w", err)
	}

	// Save updated policy
	if err := pe.decisionRepo.UpdatePolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	// Log policy update
	pe.logPolicyChange(ctx, policyID, "updated", updates)

	return nil
}

// DeactivatePolicy marks a policy as inactive
func (pe *PolicyEnforcerImpl) DeactivatePolicy(ctx context.Context, policyID string, reason string) error {
	// Get policy
	policy, err := pe.decisionRepo.GetPolicy(ctx, policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Mark as inactive
	policy.Active = false
	now := time.Now()
	policy.EffectiveUntil = &now

	// Save updated policy
	if err := pe.decisionRepo.UpdatePolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	// Log deactivation
	pe.logPolicyChange(ctx, policyID, "deactivated", map[string]interface{}{
		"reason": reason,
	})

	return nil
}

// HandleViolation processes a policy violation
func (pe *PolicyEnforcerImpl) HandleViolation(ctx context.Context, violation *entities.PolicyViolation) error {
	// Create violation record
	record := &ViolationRecord{
		ViolationID: uuid.New().String(),
		PolicyID:    violation.PolicyID,
		DecisionID:  "", // Set by caller if applicable
		Timestamp:   time.Now().Format(time.RFC3339),
		Severity:    violation.Severity,
		Resolution:  "pending",
	}

	// Determine handling based on severity
	switch violation.Severity {
	case "critical":
		// Immediate escalation required
		if err := pe.escalateViolation(ctx, violation); err != nil {
			return fmt.Errorf("failed to escalate critical violation: %w", err)
		}
		record.Resolution = "escalated"

	case "high":
		// Notify relevant parties
		if err := pe.notifyViolation(ctx, violation); err != nil {
			return fmt.Errorf("failed to notify violation: %w", err)
		}
		record.Resolution = "notified"

	case "medium", "low":
		// Log for review
		record.Resolution = "logged"
	}

	// Store violation record
	// This would typically be stored in a violations table
	pe.logViolation(ctx, record)

	return nil
}

// GetViolationHistory retrieves past violations for a policy
func (pe *PolicyEnforcerImpl) GetViolationHistory(ctx context.Context, policyID string) ([]*ViolationRecord, error) {
	// This would query a violations table
	// For now, return empty list
	return []*ViolationRecord{}, nil
}

// Helper methods

func (pe *PolicyEnforcerImpl) isPolicyEffective(policy *entities.Policy) bool {
	now := time.Now()

	// Check if policy has started
	if policy.EffectiveFrom.After(now) {
		return false
	}

	// Check if policy has ended
	if policy.EffectiveUntil != nil && policy.EffectiveUntil.Before(now) {
		return false
	}

	return true
}

func (pe *PolicyEnforcerImpl) isPolicyApplicable(policy *entities.Policy, decisionType entities.DecisionType) bool {
	// Check if policy category matches decision type
	switch policy.Category {
	case "general":
		return true
	case "content":
		return decisionType == entities.DecisionTypeContent
	case "financial":
		return decisionType == entities.DecisionTypeFinancial
	case "client":
		return decisionType == entities.DecisionTypeClient
	case "operational":
		return decisionType == entities.DecisionTypeOperational
	case "compliance":
		return decisionType == entities.DecisionTypeCompliance
	case "ethical":
		return decisionType == entities.DecisionTypeEthical
	default:
		// Check metadata for specific decision types
		if types, ok := policy.Metadata["applicable_types"].([]string); ok {
			for _, t := range types {
				if t == string(decisionType) {
					return true
				}
			}
		}
	}

	return false
}

func (pe *PolicyEnforcerImpl) evaluateRule(decision *entities.Decision, policy *entities.Policy, rule entities.PolicyRule) (bool, *entities.PolicyViolation) {
	// Evaluate rule condition
	passed := pe.ruleEngine.Evaluate(rule, decision)

	if !passed {
		// Check if this is an exception case
		for _, exception := range rule.Exceptions {
			if pe.isException(decision, exception) {
				return true, nil
			}
		}

		// Create violation
		violation := &entities.PolicyViolation{
			PolicyID:    policy.ID,
			PolicyName:  policy.Name,
			Severity:    rule.Severity,
			Description: fmt.Sprintf("Rule '%s' violated: %s", rule.ID, rule.Condition),
			Confidence:  0.95, // High confidence in rule-based violations
		}

		return false, violation
	}

	return true, nil
}

func (pe *PolicyEnforcerImpl) evaluateRuleCondition(rule entities.PolicyRule, context map[string]interface{}) bool {
	// Simple rule evaluation logic
	// In production, this would use a proper rule engine

	// Parse condition
	parts := strings.Split(rule.Condition, " ")
	if len(parts) < 3 {
		return false
	}

	field := parts[0]
	operator := parts[1]
	value := strings.Join(parts[2:], " ")

	// Get field value from context
	fieldValue, exists := context[field]
	if !exists {
		return false
	}

	// Evaluate based on operator
	switch operator {
	case "==", "equals":
		return fmt.Sprintf("%v", fieldValue) == value
	case "!=", "not_equals":
		return fmt.Sprintf("%v", fieldValue) != value
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), value)
	case "matches":
		matched, _ := regexp.MatchString(value, fmt.Sprintf("%v", fieldValue))
		return matched
	default:
		return false
	}
}

func (pe *PolicyEnforcerImpl) doesRuleApplyToAction(rule entities.PolicyRule, action string) bool {
	// Check if rule action matches
	return rule.Action == action || rule.Action == "*"
}

func (pe *PolicyEnforcerImpl) isException(decision *entities.Decision, exception string) bool {
	// Check if decision matches exception criteria
	// This would be more sophisticated in production

	// Check for specific decision IDs
	if strings.HasPrefix(exception, "decision_id:") {
		return decision.ID == strings.TrimPrefix(exception, "decision_id:")
	}

	// Check for context values
	if strings.HasPrefix(exception, "context:") {
		contextKey := strings.TrimPrefix(exception, "context:")
		_, exists := decision.Context[contextKey]
		return exists
	}

	return false
}

func (pe *PolicyEnforcerImpl) checkForWarning(decision *entities.Decision, rule entities.PolicyRule) string {
	// Check if decision is close to violating the rule
	// This would implement warning thresholds
	return ""
}

func (pe *PolicyEnforcerImpl) getRequiredAction(rule entities.PolicyRule) string {
	// Determine required action based on rule
	if action, ok := rule.Parameters["required_action"].(string); ok {
		return action
	}
	return ""
}

func (pe *PolicyEnforcerImpl) validatePolicy(policy *entities.Policy) error {
	// Validate policy structure
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Category == "" {
		return fmt.Errorf("policy category is required")
	}

	if len(policy.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	// Validate each rule
	for i, rule := range policy.Rules {
		if rule.ID == "" {
			rule.ID = fmt.Sprintf("%s_rule_%d", policy.ID, i)
		}

		if rule.Condition == "" {
			return fmt.Errorf("rule condition is required")
		}

		if rule.Action == "" {
			return fmt.Errorf("rule action is required")
		}

		if rule.Severity == "" {
			rule.Severity = "medium"
		}
	}

	return nil
}

func (pe *PolicyEnforcerImpl) escalateViolation(ctx context.Context, violation *entities.PolicyViolation) error {
	// Implement escalation logic
	// This would notify administrators or trigger emergency protocols
	return nil
}

func (pe *PolicyEnforcerImpl) notifyViolation(ctx context.Context, violation *entities.PolicyViolation) error {
	// Implement notification logic
	// This would send alerts to relevant parties
	return nil
}

func (pe *PolicyEnforcerImpl) logValidation(ctx context.Context, decisionID string, result *PolicyValidationResult) {
	// Log validation result
	log := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decisionID,
		Timestamp:   time.Now(),
		EventType:   "policy_validation",
		Description: fmt.Sprintf("Policy validation: compliant=%v, score=%.2f", result.Compliant, result.ComplianceScore),
		Actor:       "policy_enforcer",
		Changes: map[string]interface{}{
			"violations": len(result.Violations),
			"warnings":   len(result.Warnings),
			"score":      result.ComplianceScore,
		},
	}
	if err := pe.decisionRepo.CreateDecisionLog(ctx, log); err != nil {
		// Log error but don't fail the function
		// Note: we may want to use a proper logger here
		// TODO: Add proper logging when logger is available
		_ = err // Explicitly ignore error to satisfy linter
	}
}

func (pe *PolicyEnforcerImpl) logPolicyChange(ctx context.Context, policyID string, action string, details map[string]interface{}) {
	// Log policy change
	// This would create an audit log entry
}

func (pe *PolicyEnforcerImpl) logViolation(ctx context.Context, record *ViolationRecord) {
	// Log violation record
	// This would store the violation in a database
}

// RuleEngine handles rule evaluation
type RuleEngine interface {
	Evaluate(rule entities.PolicyRule, decision *entities.Decision) bool
}

// SimpleRuleEngine provides basic rule evaluation
type SimpleRuleEngine struct{}

func NewRuleEngine() RuleEngine {
	return &SimpleRuleEngine{}
}

func (re *SimpleRuleEngine) Evaluate(rule entities.PolicyRule, decision *entities.Decision) bool {
	// Simple evaluation based on rule condition
	// In production, this would use a proper rules engine like Drools or custom DSL

	// Example conditions:
	// "amount < 10000" - Check if amount is less than 10000
	// "type == 'financial'" - Check if type equals financial
	// "priority != 'critical'" - Check if priority is not critical

	// For now, implement basic comparison logic
	switch rule.Condition {
	case "always":
		return true
	case "never":
		return false
	default:
		// Parse and evaluate condition
		// This is a simplified implementation
		return re.evaluateCondition(rule.Condition, decision)
	}
}

func (re *SimpleRuleEngine) evaluateCondition(condition string, decision *entities.Decision) bool {
	// Basic condition evaluation
	// This would be much more sophisticated in production

	// Check for specific decision types
	if strings.Contains(condition, "type") {
		if strings.Contains(condition, string(decision.Type)) {
			return true
		}
	}

	// Check for priority levels
	if strings.Contains(condition, "priority") {
		if strings.Contains(condition, string(decision.Priority)) {
			return true
		}
	}

	// Check for confidence thresholds
	if strings.Contains(condition, "confidence") {
		// Extract threshold value
		if strings.Contains(condition, ">") {
			parts := strings.Split(condition, ">")
			if len(parts) == 2 {
				threshold := 0.8 // Default threshold
				if val, err := parseFloat(strings.TrimSpace(parts[1])); err == nil {
					threshold = val
				}
				return decision.ConfidenceScore > threshold
			}
		}
	}

	// Default to true for unknown conditions
	return true
}

func parseFloat(s string) (float64, error) {
	// Simple float parsing helper
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

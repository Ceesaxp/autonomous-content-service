package entities

import (
	"time"
)

// DecisionType defines the type of decision being made
type DecisionType string

const (
	DecisionTypeOperational  DecisionType = "operational"
	DecisionTypeStrategic    DecisionType = "strategic"
	DecisionTypeEmergency    DecisionType = "emergency"
	DecisionTypeEthical      DecisionType = "ethical"
	DecisionTypeFinancial    DecisionType = "financial"
	DecisionTypeContent      DecisionType = "content"
	DecisionTypeClient       DecisionType = "client"
	DecisionTypeCompliance   DecisionType = "compliance"
)

// DecisionPriority defines the urgency of a decision
type DecisionPriority string

const (
	PriorityCritical  DecisionPriority = "critical"
	PriorityHigh      DecisionPriority = "high"
	PriorityMedium    DecisionPriority = "medium"
	PriorityLow       DecisionPriority = "low"
)

// DecisionStatus tracks the lifecycle of a decision
type DecisionStatus string

const (
	StatusPending    DecisionStatus = "pending"
	StatusAnalyzing  DecisionStatus = "analyzing"
	StatusApproved   DecisionStatus = "approved"
	StatusRejected   DecisionStatus = "rejected"
	StatusOverridden DecisionStatus = "overridden"
	StatusExecuted   DecisionStatus = "executed"
	StatusReverted   DecisionStatus = "reverted"
)

// Decision represents an autonomous decision made by the system
type Decision struct {
	ID                 string                 `json:"id"`
	Type               DecisionType           `json:"type"`
	Priority           DecisionPriority       `json:"priority"`
	Status             DecisionStatus         `json:"status"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	Context            map[string]interface{} `json:"context"`
	Options            []DecisionOption       `json:"options"`
	SelectedOption     *DecisionOption        `json:"selected_option,omitempty"`
	Justification      string                 `json:"justification"`
	ConfidenceScore    float64                `json:"confidence_score"`
	Confidence         float64                `json:"confidence"` // Alias for compatibility
	Reasoning          string                 `json:"reasoning"`  // Alias for justification
	ImpactAnalysis     *ImpactAnalysis        `json:"impact_analysis,omitempty"`
	StakeholderImpact  []StakeholderImpact    `json:"stakeholder_impact,omitempty"`
	Constraints        []string               `json:"constraints,omitempty"`
	PolicyViolations   []PolicyViolation      `json:"policy_violations,omitempty"`
	Override           *DecisionOverride      `json:"override,omitempty"`
	ExecutionResult    *ExecutionResult       `json:"execution_result,omitempty"`
	Deadline           *time.Time             `json:"deadline,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	ExecutedAt         *time.Time             `json:"executed_at,omitempty"`
}

// DecisionOption represents a possible choice in a decision
type DecisionOption struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`
	Risks       []string               `json:"risks"`
	Benefits    []string               `json:"benefits"`
	Constraints []string               `json:"constraints"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ImpactAnalysis evaluates the potential effects of a decision
type ImpactAnalysis struct {
	StakeholderImpacts []StakeholderImpact `json:"stakeholder_impacts"`
	FinancialImpact    *FinancialImpact    `json:"financial_impact,omitempty"`
	OperationalImpact  *OperationalImpact  `json:"operational_impact,omitempty"`
	ReputationalRisk   float64             `json:"reputational_risk"`
	ComplianceRisk     float64             `json:"compliance_risk"`
	ReversibilityScore float64             `json:"reversibility_score"`
}

// StakeholderImpact represents how a decision affects a stakeholder
type StakeholderImpact struct {
	StakeholderType string  `json:"stakeholder_type"`
	ImpactLevel     string  `json:"impact_level"`
	Description     string  `json:"description"`
	Sentiment       float64 `json:"sentiment"`
}

// FinancialImpact represents the financial consequences of a decision
type FinancialImpact struct {
	EstimatedCost    float64 `json:"estimated_cost"`
	EstimatedRevenue float64 `json:"estimated_revenue"`
	CashFlowImpact   float64 `json:"cash_flow_impact"`
	ROIEstimate      float64 `json:"roi_estimate"`
	PaybackPeriod    int     `json:"payback_period_days"`
}

// OperationalImpact represents how a decision affects operations
type OperationalImpact struct {
	ResourceRequirements map[string]float64 `json:"resource_requirements"`
	TimelineImpact       int                `json:"timeline_impact_hours"`
	ComplexityIncrease   float64            `json:"complexity_increase"`
	AutomationPotential  float64            `json:"automation_potential"`
}

// PolicyViolation represents a policy that would be violated by a decision
type PolicyViolation struct {
	PolicyID    string  `json:"policy_id"`
	PolicyName  string  `json:"policy_name"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// DecisionOverride represents manual intervention in a decision
type DecisionOverride struct {
	OverrideID     string    `json:"override_id"`
	AuthorizedBy   string    `json:"authorized_by"`
	Reason         string    `json:"reason"`
	RiskAcceptance string    `json:"risk_acceptance"`
	Timestamp      time.Time `json:"timestamp"`
}

// ExecutionResult captures the outcome of executing a decision
type ExecutionResult struct {
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metrics      map[string]interface{} `json:"metrics"`
	SideEffects  []string               `json:"side_effects"`
	Reversible   bool                   `json:"reversible"`
}

// Policy defines rules and guidelines for decision making
type Policy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	Rules           []PolicyRule           `json:"rules"`
	Priority        int                    `json:"priority"`
	EffectiveFrom   time.Time              `json:"effective_from"`
	EffectiveUntil  *time.Time             `json:"effective_until,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Active          bool                   `json:"active"`
}

// PolicyRule defines a specific rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Exceptions  []string               `json:"exceptions"`
	Severity    string                 `json:"severity"`
}

// EthicalGuideline represents ethical constraints for the system
type EthicalGuideline struct {
	ID          string   `json:"id"`
	Principle   string   `json:"principle"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	RedLines    []string `json:"red_lines"`
	Weight      float64  `json:"weight"`
}

// DecisionLog captures the full audit trail of a decision
type DecisionLog struct {
	ID           string                 `json:"id"`
	DecisionID   string                 `json:"decision_id"`
	Timestamp    time.Time              `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	Description  string                 `json:"description"`
	Actor        string                 `json:"actor"`
	Changes      map[string]interface{} `json:"changes"`
	SystemState  map[string]interface{} `json:"system_state"`
}

// DecisionTemplate provides reusable patterns for common decisions
type DecisionTemplate struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             DecisionType           `json:"type"`
	Description      string                 `json:"description"`
	RequiredContext  []string               `json:"required_context"`
	DecisionCriteria []string               `json:"decision_criteria"`
	DefaultOptions   []DecisionOption       `json:"default_options"`
	PolicyChecks     []string               `json:"policy_checks"`
	Metadata         map[string]interface{} `json:"metadata"`
}
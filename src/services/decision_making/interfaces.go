package decision_making

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// DecisionEngine orchestrates the decision-making process
type DecisionEngine interface {
	// Core decision operations
	InitiateDecision(ctx context.Context, request DecisionRequest) (*entities.Decision, error)
	AnalyzeOptions(ctx context.Context, decision *entities.Decision) error
	MakeDecision(ctx context.Context, decisionID string) (*entities.Decision, error)
	ExecuteDecision(ctx context.Context, decisionID string) (*entities.ExecutionResult, error)
	RevertDecision(ctx context.Context, decisionID string, reason string) error

	// Override and escalation
	OverrideDecision(ctx context.Context, decisionID string, override *entities.DecisionOverride) error
	EscalateDecision(ctx context.Context, decisionID string, reason string) error

	// Quality and learning
	AssessDecisionQuality(ctx context.Context, decisionID string) (*DecisionQualityReport, error)
	LearnFromDecision(ctx context.Context, decisionID string) error
}

// PolicyEnforcer validates decisions against policies
type PolicyEnforcer interface {
	// Policy validation
	ValidateDecision(ctx context.Context, decision *entities.Decision) (*PolicyValidationResult, error)
	CheckPolicyCompliance(ctx context.Context, action string, context map[string]interface{}) (bool, error)
	GetApplicablePolicies(ctx context.Context, decisionType entities.DecisionType) ([]*entities.Policy, error)

	// Policy management
	RegisterPolicy(ctx context.Context, policy *entities.Policy) error
	UpdatePolicy(ctx context.Context, policyID string, updates map[string]interface{}) error
	DeactivatePolicy(ctx context.Context, policyID string, reason string) error

	// Violation handling
	HandleViolation(ctx context.Context, violation *entities.PolicyViolation) error
	GetViolationHistory(ctx context.Context, policyID string) ([]*ViolationRecord, error)
}

// EthicalFramework ensures decisions align with ethical guidelines
type EthicalFramework interface {
	// Ethical validation
	ValidateEthics(ctx context.Context, decision *entities.Decision) (*EthicalValidationResult, error)
	CheckRedLines(ctx context.Context, action string, context map[string]interface{}) (bool, error)
	AssessBias(ctx context.Context, decision *entities.Decision) (*BiasAssessment, error)

	// Guideline management
	RegisterGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error
	GetActiveGuidelines(ctx context.Context) ([]*entities.EthicalGuideline, error)

	// Ethical analysis
	GenerateEthicalJustification(ctx context.Context, decision *entities.Decision) (string, error)
	IdentifyEthicalConcerns(ctx context.Context, scenario string) ([]EthicalConcern, error)
}

// ImpactAnalyzer evaluates the consequences of decisions
type ImpactAnalyzer interface {
	// Impact assessment
	AnalyzeImpact(ctx context.Context, decision *entities.Decision) (*entities.ImpactAnalysis, error)
	PredictOutcomes(ctx context.Context, decision *entities.Decision) (*OutcomePrediction, error)
	AssessRisk(ctx context.Context, decision *entities.Decision) (*RiskAssessment, error)

	// Stakeholder analysis
	IdentifyStakeholders(ctx context.Context, decision *entities.Decision) ([]Stakeholder, error)
	AnalyzeStakeholderImpact(ctx context.Context, decision *entities.Decision) ([]*entities.StakeholderImpact, error)

	// Financial analysis
	CalculateFinancialImpact(ctx context.Context, decision *entities.Decision) (*entities.FinancialImpact, error)
	EstimateROI(ctx context.Context, decision *entities.Decision) (float64, error)
}

// ConflictResolver handles competing priorities and decisions
type ConflictResolver interface {
	// Conflict detection
	DetectConflicts(ctx context.Context, decisions []*entities.Decision) ([]*DecisionConflict, error)
	AnalyzeConflict(ctx context.Context, conflictID string) (*ConflictAnalysis, error)

	// Resolution strategies
	ProposeResolution(ctx context.Context, conflict *DecisionConflict) (*ResolutionProposal, error)
	ResolveConflict(ctx context.Context, conflictID string, resolution *ConflictResolution) error

	// Priority management
	PrioritizeDecisions(ctx context.Context, decisions []*entities.Decision) ([]*entities.Decision, error)
	OptimizeResourceAllocation(ctx context.Context, decisions []*entities.Decision) (*ResourceAllocation, error)
}

// EmergencyProtocol handles critical situations
type EmergencyProtocol interface {
	// Emergency detection
	AssessSystemHealth(ctx context.Context) (*SystemHealthReport, error)
	DetectEmergency(ctx context.Context, indicators map[string]interface{}) (*EmergencyAssessment, error)

	// Emergency response
	ActivateEmergencyMode(ctx context.Context, reason string) error
	ExecuteEmergencyShutdown(ctx context.Context, scope string) error
	InitiateRecovery(ctx context.Context) error

	// Fallback mechanisms
	GetFallbackPlan(ctx context.Context, scenario string) (*FallbackPlan, error)
	ExecuteFallback(ctx context.Context, planID string) error
}

// Supporting structures

type DecisionRequest struct {
	Type        entities.DecisionType     `json:"type"`
	Priority    entities.DecisionPriority `json:"priority"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Context     map[string]interface{}    `json:"context"`
	Constraints []string                  `json:"constraints"`
	Deadline    *string                   `json:"deadline,omitempty"`
}

type PolicyValidationResult struct {
	Compliant       bool                       `json:"compliant"`
	Violations      []entities.PolicyViolation `json:"violations"`
	Warnings        []string                   `json:"warnings"`
	RequiredActions []string                   `json:"required_actions"`
	ComplianceScore float64                    `json:"compliance_score"`
}

type EthicalValidationResult struct {
	Approved          bool             `json:"approved"`
	Concerns          []EthicalConcern `json:"concerns"`
	RedLineViolations []string         `json:"red_line_violations"`
	EthicalScore      float64          `json:"ethical_score"`
	Justification     string           `json:"justification"`
}

type EthicalConcern struct {
	GuidelineID string `json:"guideline_id"`
	Principle   string `json:"principle"`
	Concern     string `json:"concern"`
	Severity    string `json:"severity"`
	Mitigation  string `json:"mitigation"`
}

type BiasAssessment struct {
	BiasDetected    bool     `json:"bias_detected"`
	BiasTypes       []string `json:"bias_types"`
	BiasScore       float64  `json:"bias_score"`
	AffectedGroups  []string `json:"affected_groups"`
	Recommendations []string `json:"recommendations"`
}

type OutcomePrediction struct {
	PrimaryOutcome      string    `json:"primary_outcome"`
	Probability         float64   `json:"probability"`
	AlternativeOutcomes []Outcome `json:"alternative_outcomes"`
	ConfidenceLevel     float64   `json:"confidence_level"`
	TimeHorizon         string    `json:"time_horizon"`
}

type Outcome struct {
	Description string  `json:"description"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
}

type RiskAssessment struct {
	OverallRisk     float64      `json:"overall_risk"`
	RiskFactors     []RiskFactor `json:"risk_factors"`
	MitigationPlans []string     `json:"mitigation_plans"`
	AcceptableRisk  bool         `json:"acceptable_risk"`
}

type RiskFactor struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Likelihood  float64 `json:"likelihood"`
	Impact      float64 `json:"impact"`
	RiskScore   float64 `json:"risk_score"`
}

type Stakeholder struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Influence float64 `json:"influence"`
	Interest  float64 `json:"interest"`
}

type DecisionConflict struct {
	ID                   string   `json:"id"`
	ConflictingDecisions []string `json:"conflicting_decisions"`
	ConflictType         string   `json:"conflict_type"`
	Severity             string   `json:"severity"`
	Description          string   `json:"description"`
	ResourcesInConflict  []string `json:"resources_in_conflict"`
}

type ConflictAnalysis struct {
	ConflictID        string             `json:"conflict_id"`
	RootCause         string             `json:"root_cause"`
	ImpactAssessment  map[string]float64 `json:"impact_assessment"`
	ResolutionOptions []ResolutionOption `json:"resolution_options"`
}

type ResolutionOption struct {
	ID          string   `json:"id"`
	Strategy    string   `json:"strategy"`
	Description string   `json:"description"`
	TradeOffs   []string `json:"trade_offs"`
	SuccessRate float64  `json:"success_rate"`
}

type ResolutionProposal struct {
	ConflictID       string   `json:"conflict_id"`
	ProposedStrategy string   `json:"proposed_strategy"`
	Actions          []string `json:"actions"`
	ExpectedOutcome  string   `json:"expected_outcome"`
	Timeline         string   `json:"timeline"`
}

type ConflictResolution struct {
	Strategy             string   `json:"strategy"`
	Actions              []string `json:"actions"`
	Justification        string   `json:"justification"`
	StakeholderAgreement bool     `json:"stakeholder_agreement"`
}

type ResourceAllocation struct {
	Allocations       map[string]float64 `json:"allocations"`
	Efficiency        float64            `json:"efficiency"`
	Conflicts         []string           `json:"conflicts"`
	OptimizationScore float64            `json:"optimization_score"`
}

type SystemHealthReport struct {
	OverallHealth      float64            `json:"overall_health"`
	ComponentStatus    map[string]string  `json:"component_status"`
	PerformanceMetrics map[string]float64 `json:"performance_metrics"`
	Anomalies          []string           `json:"anomalies"`
	Recommendations    []string           `json:"recommendations"`
}

type EmergencyAssessment struct {
	IsEmergency      bool     `json:"is_emergency"`
	Severity         string   `json:"severity"`
	AffectedSystems  []string `json:"affected_systems"`
	ImmediateActions []string `json:"immediate_actions"`
	EscalationNeeded bool     `json:"escalation_needed"`
}

type FallbackPlan struct {
	ID               string   `json:"id"`
	Scenario         string   `json:"scenario"`
	Steps            []string `json:"steps"`
	ResourcesNeeded  []string `json:"resources_needed"`
	ExpectedDuration string   `json:"expected_duration"`
	SuccessMetrics   []string `json:"success_metrics"`
}

type DecisionQualityReport struct {
	DecisionID     string   `json:"decision_id"`
	QualityScore   float64  `json:"quality_score"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
	LessonsLearned []string `json:"lessons_learned"`
	Improvements   []string `json:"improvements"`
}

type ViolationRecord struct {
	ViolationID string `json:"violation_id"`
	PolicyID    string `json:"policy_id"`
	DecisionID  string `json:"decision_id"`
	Timestamp   string `json:"timestamp"`
	Severity    string `json:"severity"`
	Resolution  string `json:"resolution"`
}

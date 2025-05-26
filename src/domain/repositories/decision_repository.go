package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// DecisionRepository defines the interface for decision data persistence
type DecisionRepository interface {
	// Decision CRUD operations
	CreateDecision(ctx context.Context, decision *entities.Decision) error
	GetDecision(ctx context.Context, id string) (*entities.Decision, error)
	UpdateDecision(ctx context.Context, decision *entities.Decision) error
	DeleteDecision(ctx context.Context, id string) error

	// Decision queries
	ListDecisions(ctx context.Context, filter DecisionFilter) ([]*entities.Decision, error)
	GetDecisionsByType(ctx context.Context, decisionType entities.DecisionType) ([]*entities.Decision, error)
	GetDecisionsByStatus(ctx context.Context, status entities.DecisionStatus) ([]*entities.Decision, error)
	GetPendingDecisions(ctx context.Context, priority entities.DecisionPriority) ([]*entities.Decision, error)
	GetDecisionHistory(ctx context.Context, days int) ([]*entities.Decision, error)

	// Policy operations
	CreatePolicy(ctx context.Context, policy *entities.Policy) error
	GetPolicy(ctx context.Context, id string) (*entities.Policy, error)
	UpdatePolicy(ctx context.Context, policy *entities.Policy) error
	DeletePolicy(ctx context.Context, id string) error
	ListPolicies(ctx context.Context, category string) ([]*entities.Policy, error)
	GetActivePolicies(ctx context.Context) ([]*entities.Policy, error)

	// Ethical guidelines operations
	CreateEthicalGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error
	GetEthicalGuideline(ctx context.Context, id string) (*entities.EthicalGuideline, error)
	UpdateEthicalGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error
	ListEthicalGuidelines(ctx context.Context) ([]*entities.EthicalGuideline, error)

	// Decision templates
	CreateDecisionTemplate(ctx context.Context, template *entities.DecisionTemplate) error
	GetDecisionTemplate(ctx context.Context, id string) (*entities.DecisionTemplate, error)
	ListDecisionTemplates(ctx context.Context, decisionType entities.DecisionType) ([]*entities.DecisionTemplate, error)

	// Decision logs and audit
	CreateDecisionLog(ctx context.Context, log *entities.DecisionLog) error
	GetDecisionLogs(ctx context.Context, decisionID string) ([]*entities.DecisionLog, error)
	GetAuditTrail(ctx context.Context, startTime, endTime time.Time) ([]*entities.DecisionLog, error)

	// Analytics and metrics
	GetDecisionMetrics(ctx context.Context, period string) (*DecisionMetrics, error)
	GetPolicyViolationStats(ctx context.Context, period string) (*PolicyViolationStats, error)
	GetDecisionQualityMetrics(ctx context.Context) (*QualityMetrics, error)
	GetStakeholderImpactSummary(ctx context.Context, period string) (*StakeholderImpactSummary, error)
}

// DecisionFilter defines criteria for filtering decisions
type DecisionFilter struct {
	Type          *entities.DecisionType
	Status        *entities.DecisionStatus
	Priority      *entities.DecisionPriority
	StartDate     *time.Time
	EndDate       *time.Time
	MinConfidence *float64
	HasOverride   *bool
	PolicyID      *string
}

// DecisionMetrics aggregates decision-related metrics
type DecisionMetrics struct {
	TotalDecisions       int            `json:"total_decisions"`
	DecisionsByType      map[string]int `json:"decisions_by_type"`
	DecisionsByStatus    map[string]int `json:"decisions_by_status"`
	AverageConfidence    float64        `json:"average_confidence"`
	OverrideRate         float64        `json:"override_rate"`
	ExecutionSuccessRate float64        `json:"execution_success_rate"`
	AverageExecutionTime float64        `json:"average_execution_time_ms"`
}

// PolicyViolationStats tracks policy compliance
type PolicyViolationStats struct {
	TotalViolations      int                      `json:"total_violations"`
	ViolationsBySeverity map[string]int           `json:"violations_by_severity"`
	ViolationsByPolicy   map[string]int           `json:"violations_by_policy"`
	ComplianceRate       float64                  `json:"compliance_rate"`
	TopViolatedPolicies  []PolicyViolationSummary `json:"top_violated_policies"`
}

// PolicyViolationSummary provides violation details for a policy
type PolicyViolationSummary struct {
	PolicyID       string `json:"policy_id"`
	PolicyName     string `json:"policy_name"`
	ViolationCount int    `json:"violation_count"`
	Severity       string `json:"severity"`
}

// QualityMetrics tracks decision quality over time
type QualityMetrics struct {
	AverageQualityScore float64            `json:"average_quality_score"`
	QualityTrend        float64            `json:"quality_trend"`
	QualityByType       map[string]float64 `json:"quality_by_type"`
	ImprovementAreas    []string           `json:"improvement_areas"`
	BestPerformingTypes []string           `json:"best_performing_types"`
}

// StakeholderImpactSummary aggregates stakeholder impact data
type StakeholderImpactSummary struct {
	AverageSentiment    float64               `json:"average_sentiment"`
	SentimentByGroup    map[string]float64    `json:"sentiment_by_group"`
	HighImpactDecisions int                   `json:"high_impact_decisions"`
	MitigationActions   int                   `json:"mitigation_actions"`
	StakeholderFeedback []StakeholderFeedback `json:"stakeholder_feedback"`
}

// StakeholderFeedback captures stakeholder responses
type StakeholderFeedback struct {
	StakeholderType string   `json:"stakeholder_type"`
	Sentiment       float64  `json:"sentiment"`
	FeedbackCount   int      `json:"feedback_count"`
	CommonConcerns  []string `json:"common_concerns"`
}

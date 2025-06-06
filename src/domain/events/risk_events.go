package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// Risk Events
const (
	RiskIdentified           = "risk.identified"
	RiskAssessed             = "risk.assessed"
	RiskMitigated            = "risk.mitigated"
	RiskResolved             = "risk.resolved"
	RiskEscalated            = "risk.escalated"
	RiskThresholdExceeded    = "risk.threshold_exceeded"
	ComplianceViolation      = "compliance.violation"
	ContentPolicyViolation   = "content.policy_violation"
	ServiceDependencyFailure = "service.dependency_failure"
	IncidentDetected         = "incident.detected"
	IncidentResolved         = "incident.resolved"
	VulnerabilityDetected    = "vulnerability.detected"
	BackupCompleted          = "backup.completed"
	BackupFailed             = "backup.failed"
)

// RiskIdentifiedEvent is emitted when a new risk is identified
type RiskIdentifiedEvent struct {
	BaseEvent
	RiskID           string                `json:"risk_id"`
	Type             entities.RiskType     `json:"type"`
	Severity         entities.RiskSeverity `json:"severity"`
	Title            string                `json:"title"`
	Source           string                `json:"source"`
	AffectedEntities []string              `json:"affected_entities"`
	Score            float64               `json:"score"`
}

// RiskAssessedEvent is emitted when a risk is assessed
type RiskAssessedEvent struct {
	BaseEvent
	RiskID           string                `json:"risk_id"`
	PreviousSeverity entities.RiskSeverity `json:"previous_severity"`
	NewSeverity      entities.RiskSeverity `json:"new_severity"`
	PreviousScore    float64               `json:"previous_score"`
	NewScore         float64               `json:"new_score"`
	Probability      float64               `json:"probability"`
	Impact           float64               `json:"impact"`
	Assessment       string                `json:"assessment"`
}

// RiskMitigatedEvent is emitted when risk mitigation is performed
type RiskMitigatedEvent struct {
	BaseEvent
	RiskID         string   `json:"risk_id"`
	MitigationPlan string   `json:"mitigation_plan"`
	Actions        []string `json:"actions"`
	EffectiveScore float64  `json:"effective_score"`
	Status         string   `json:"status"`
}

// RiskThresholdExceededEvent is emitted when a risk threshold is exceeded
type RiskThresholdExceededEvent struct {
	BaseEvent
	ThresholdID    string            `json:"threshold_id"`
	Type           entities.RiskType `json:"type"`
	Category       string            `json:"category"`
	CurrentValue   float64           `json:"current_value"`
	ThresholdValue float64           `json:"threshold_value"`
	Actions        []string          `json:"actions"`
}

// ComplianceViolationEvent is emitted when a compliance violation is detected
type ComplianceViolationEvent struct {
	BaseEvent
	RequirementID   string   `json:"requirement_id"`
	Regulation      string   `json:"regulation"`
	Article         string   `json:"article"`
	ViolationType   string   `json:"violation_type"`
	Description     string   `json:"description"`
	AffectedData    []string `json:"affected_data"`
	RequiredActions []string `json:"required_actions"`
}

// ContentPolicyViolationEvent is emitted when content violates policies
type ContentPolicyViolationEvent struct {
	BaseEvent
	PolicyID       string  `json:"policy_id"`
	ContentID      string  `json:"content_id"`
	PolicyType     string  `json:"policy_type"`
	Violation      string  `json:"violation"`
	Confidence     float64 `json:"confidence"`
	Action         string  `json:"action"`
	ContentSnippet string  `json:"content_snippet"`
}

// ServiceDependencyFailureEvent is emitted when a service dependency fails
type ServiceDependencyFailureEvent struct {
	BaseEvent
	DependencyID   string    `json:"dependency_id"`
	ServiceName    string    `json:"service_name"`
	Provider       string    `json:"provider"`
	Criticality    string    `json:"criticality"`
	FailureType    string    `json:"failure_type"`
	LastSuccessful time.Time `json:"last_successful"`
	FallbackUsed   bool      `json:"fallback_used"`
}

// IncidentDetectedEvent is emitted when an incident is detected
type IncidentDetectedEvent struct {
	BaseEvent
	IncidentID      string                `json:"incident_id"`
	Type            string                `json:"type"`
	Severity        entities.RiskSeverity `json:"severity"`
	Title           string                `json:"title"`
	AffectedService string                `json:"affected_service"`
	Impact          string                `json:"impact"`
	AutoResponse    bool                  `json:"auto_response"`
}

// IncidentResolvedEvent is emitted when an incident is resolved
type IncidentResolvedEvent struct {
	BaseEvent
	IncidentID      string        `json:"incident_id"`
	Resolution      string        `json:"resolution"`
	RootCause       string        `json:"root_cause"`
	TimeToDetect    time.Duration `json:"time_to_detect"`
	TimeToResolve   time.Duration `json:"time_to_resolve"`
	ActionsExecuted []string      `json:"actions_executed"`
	LessonsLearned  []string      `json:"lessons_learned"`
}

// VulnerabilityDetectedEvent is emitted when a security vulnerability is detected
type VulnerabilityDetectedEvent struct {
	BaseEvent
	VulnerabilityID string                `json:"vulnerability_id"`
	Type            string                `json:"type"`
	Severity        entities.RiskSeverity `json:"severity"`
	CVE             string                `json:"cve,omitempty"`
	Component       string                `json:"component"`
	Version         string                `json:"version"`
	Recommendation  string                `json:"recommendation"`
}

// BackupCompletedEvent is emitted when a backup is completed
type BackupCompletedEvent struct {
	BaseEvent
	BackupID   string        `json:"backup_id"`
	Type       string        `json:"type"`
	Size       int64         `json:"size"`
	Components []string      `json:"components"`
	Location   string        `json:"location"`
	Duration   time.Duration `json:"duration"`
	Verified   bool          `json:"verified"`
}

// BackupFailedEvent is emitted when a backup fails
type BackupFailedEvent struct {
	BaseEvent
	BackupID   string   `json:"backup_id"`
	Type       string   `json:"type"`
	Components []string `json:"components"`
	Error      string   `json:"error"`
	RetryCount int      `json:"retry_count"`
}

// RiskEscalatedEvent is emitted when a risk is escalated
type RiskEscalatedEvent struct {
	BaseEvent
	RiskID      string                `json:"risk_id"`
	NewSeverity entities.RiskSeverity `json:"new_severity"`
	Reason      string                `json:"reason"`
}

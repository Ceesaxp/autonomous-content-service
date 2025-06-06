package entities

import (
	"time"

	"github.com/google/uuid"
)

// RiskCategory represents the category of risk
type RiskCategory string

const (
	RiskCategoryContent     RiskCategory = "Content"
	RiskCategoryFinancial   RiskCategory = "Financial"
	RiskCategoryOperational RiskCategory = "Operational"
	RiskCategorySecurity    RiskCategory = "Security"
	RiskCategoryLegal       RiskCategory = "Legal"
	RiskCategoryReputation  RiskCategory = "Reputation"
)

// RiskType represents the category of risk (alias for compatibility)
type RiskType = RiskCategory

const (
	RiskTypeContent     = RiskCategoryContent
	RiskTypeFinancial   = RiskCategoryFinancial
	RiskTypeOperational = RiskCategoryOperational
	RiskTypeSecurity    = RiskCategorySecurity
	RiskTypeCompliance  = RiskCategoryLegal
	RiskTypeReputation  = RiskCategoryReputation
)

// RiskSeverity represents the severity level of a risk
type RiskSeverity string

const (
	RiskSeverityCritical RiskSeverity = "Critical"
	RiskSeverityHigh     RiskSeverity = "High"
	RiskSeverityMedium   RiskSeverity = "Medium"
	RiskSeverityLow      RiskSeverity = "Low"
)

// RiskStatus represents the current status of a risk
type RiskStatus string

const (
	RiskStatusIdentified RiskStatus = "Identified"
	RiskStatusAssessing  RiskStatus = "Assessing"
	RiskStatusMitigating RiskStatus = "Mitigating"
	RiskStatusMonitoring RiskStatus = "Monitoring"
	RiskStatusResolved   RiskStatus = "Resolved"
	RiskStatusAccepted   RiskStatus = "Accepted"
)

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "Open"
	IncidentStatusInProgress IncidentStatus = "InProgress"
	IncidentStatusResolved   IncidentStatus = "Resolved"
	IncidentStatusClosed     IncidentStatus = "Closed"
)

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

const (
	VulnerabilityStatusOpen          VulnerabilityStatus = "Open"
	VulnerabilityStatusInProgress    VulnerabilityStatus = "InProgress"
	VulnerabilityStatusFixed         VulnerabilityStatus = "Fixed"
	VulnerabilityStatusAccepted      VulnerabilityStatus = "Accepted"
	VulnerabilityStatusFalsePositive VulnerabilityStatus = "FalsePositive"
)

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusInProgress BackupStatus = "InProgress"
	BackupStatusCompleted  BackupStatus = "Completed"
	BackupStatusFailed     BackupStatus = "Failed"
	BackupStatusVerified   BackupStatus = "Verified"
)

// Risk represents an identified risk in the system
type Risk struct {
	ID                uuid.UUID              `json:"id"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Category          RiskCategory           `json:"category"`
	Severity          RiskSeverity           `json:"severity"`
	Likelihood        float64                `json:"likelihood"`        // 0.0 to 1.0
	Impact            float64                `json:"impact"`            // 0.0 to 1.0
	Status            RiskStatus             `json:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	MitigationActions []string               `json:"mitigation_actions,omitempty"`
	OwnerID           *uuid.UUID             `json:"owner_id,omitempty"`
	IdentifiedAt      time.Time              `json:"identified_at"`
	LastAssessment    time.Time              `json:"last_assessment"`
	ResolutionDate    *time.Time             `json:"resolution_date,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Incident represents a system incident
type Incident struct {
	ID             uuid.UUID              `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Severity       RiskSeverity           `json:"severity"`
	Status         IncidentStatus         `json:"status"`
	Category       RiskCategory           `json:"category"`
	Source         string                 `json:"source"`
	AssigneeID     *uuid.UUID             `json:"assignee_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ActionsTaken   []IncidentAction       `json:"actions_taken,omitempty"`
	RootCause      string                 `json:"root_cause,omitempty"`
	LessonsLearned string                 `json:"lessons_learned,omitempty"`
	Resolution     string                 `json:"resolution,omitempty"`
	OccurredAt     time.Time              `json:"occurred_at"`
	DetectedAt     time.Time              `json:"detected_at"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// IncidentAction represents an action taken during incident response
type IncidentAction struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Result      string    `json:"result"`
	ExecutedAt  time.Time `json:"executed_at"`
	ExecutedBy  string    `json:"executed_by"` // "system" or user ID
}

// Vulnerability represents a detected security vulnerability
type Vulnerability struct {
	ID               uuid.UUID           `json:"id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Severity         RiskSeverity        `json:"severity"`
	Status           VulnerabilityStatus `json:"status"`
	Component        string              `json:"component"`
	CVEID            string              `json:"cve_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	RemediationSteps []string            `json:"remediation_steps,omitempty"`
	DiscoveredAt     time.Time           `json:"discovered_at"`
	FixedAt          *time.Time          `json:"fixed_at,omitempty"`
	VerifiedAt       *time.Time          `json:"verified_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// Backup represents a system backup
type Backup struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	BackupType      string                 `json:"backup_type"`
	SizeBytes       int64                  `json:"size_bytes"`
	Status          BackupStatus           `json:"status"`
	StorageLocation string                 `json:"storage_location"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Checksum        string                 `json:"checksum,omitempty"`
	RetentionUntil  *time.Time             `json:"retention_until,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	VerifiedAt      *time.Time             `json:"verified_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// RiskThreshold represents configurable risk thresholds
type RiskThreshold struct {
	ID          string                 `json:"id"`
	Type        RiskType               `json:"type"`
	Category    string                 `json:"category"`
	Threshold   float64                `json:"threshold"`
	Unit        string                 `json:"unit"`
	Description string                 `json:"description"`
	Actions     []string               `json:"actions"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ComplianceRequirement represents a compliance requirement
type ComplianceRequirement struct {
	ID           string                 `json:"id"`
	Regulation   string                 `json:"regulation"` // e.g., "GDPR"
	Article      string                 `json:"article"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Requirements []string               `json:"requirements"`
	Controls     []string               `json:"controls"`
	IsActive     bool                   `json:"is_active"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ContentPolicy represents content moderation policies
type ContentPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // e.g., "violence", "hate_speech"
	Description string                 `json:"description"`
	Rules       []ContentRule          `json:"rules"`
	Actions     []string               `json:"actions"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ContentRule represents a specific content moderation rule
type ContentRule struct {
	Pattern     string  `json:"pattern"`
	Type        string  `json:"type"` // "keyword", "regex", "ml_classifier"
	Confidence  float64 `json:"confidence"`
	Action      string  `json:"action"`
	Description string  `json:"description"`
}

// ServiceDependency represents an external service dependency
type ServiceDependency struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"` // "api", "database", "payment", "llm"
	Provider         string                 `json:"provider"`
	Criticality      string                 `json:"criticality"` // "critical", "high", "medium", "low"
	HealthEndpoint   string                 `json:"health_endpoint"`
	FallbackService  string                 `json:"fallback_service"`
	MaxDowntime      int                    `json:"max_downtime"` // minutes
	LastHealthCheck  time.Time              `json:"last_health_check"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// SystemHealth represents overall system health status
type SystemHealth struct {
	Status         string               `json:"status"`
	Score          float64              `json:"score"`
	Components     []ComponentHealth    `json:"components"`
	ActiveRisks    int                  `json:"active_risks"`
	OpenIncidents  int                  `json:"open_incidents"`
	LastAssessment time.Time            `json:"last_assessment"`
	Recommendations []string            `json:"recommendations"`
}

// ComponentHealth represents the health status of a system component
type ComponentHealth struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ResponseTime int       `json:"response_time_ms"`
	ErrorRate    float64   `json:"error_rate"`
	LastCheck    time.Time `json:"last_check"`
}
package risk_management

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// RiskManagementService defines the main interface for risk management
type RiskManagementService interface {
	// Risk CRUD operations
	GetRiskByID(ctx context.Context, id uuid.UUID) (*entities.Risk, error)
	GetRisksByCategory(ctx context.Context, category entities.RiskCategory, offset, limit int) ([]*entities.Risk, int, error)
	GetRisksBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Risk, int, error)
	GetRisksByStatus(ctx context.Context, status entities.RiskStatus, offset, limit int) ([]*entities.Risk, int, error)
	CreateRisk(ctx context.Context, risk *entities.Risk) error
	UpdateRisk(ctx context.Context, risk *entities.Risk) error
	DeleteRisk(ctx context.Context, id uuid.UUID) error

	// Risk assessment and management
	AssessRisk(ctx context.Context, id uuid.UUID) (*entities.Risk, error)
	MitigateRisk(ctx context.Context, riskID uuid.UUID, actions []string) error
	MonitorRisks(ctx context.Context) error
	GetRiskDashboard(ctx context.Context) (*RiskDashboard, error)
	GetSystemRisks(ctx context.Context) ([]*entities.Risk, error)

	// Incident management
	CreateIncident(ctx context.Context, incident *entities.Incident) error
	RespondToIncident(ctx context.Context, incidentID uuid.UUID) error
	GetIncidentStatus(ctx context.Context, incidentID string) (*IncidentStatus, error)

	// Vulnerability management
	ScanVulnerabilities(ctx context.Context) ([]*entities.Vulnerability, error)
	FixVulnerability(ctx context.Context, id uuid.UUID) error

	// Backup management
	CreateBackup(ctx context.Context, name, backupType string) (*entities.Backup, error)
	RestoreBackup(ctx context.Context, id uuid.UUID) error
	VerifyBackup(ctx context.Context, id uuid.UUID) (bool, error)

	// System monitoring
	GetSystemHealth(ctx context.Context) (*entities.SystemHealth, error)
	GetDependencies(ctx context.Context) ([]*entities.ServiceDependency, error)
	RunSecurityScan(ctx context.Context) ([]SecurityScanResult, error)

	// Threshold management
	SetRiskThreshold(ctx context.Context, threshold *entities.RiskThreshold) error
	CheckThresholds(ctx context.Context) ([]*ThresholdViolation, error)

	// Compliance management
	CheckCompliance(ctx context.Context, data interface{}) (*ComplianceResult, error)
	GetComplianceReport(ctx context.Context) (*ComplianceReport, error)
}

// ContentRiskAnalyzer analyzes content for risks
type ContentRiskAnalyzer interface {
	AnalyzeContent(ctx context.Context, content string) (*ContentRiskResult, error)
	CheckPII(ctx context.Context, content string) (*PIICheckResult, error)
	CheckContentPolicy(ctx context.Context, content string) (*PolicyCheckResult, error)
	CheckCopyright(ctx context.Context, content string) (*CopyrightCheckResult, error)
}

// FinancialRiskAnalyzer analyzes financial risks
type FinancialRiskAnalyzer interface {
	AnalyzeTransaction(ctx context.Context, transaction *entities.Transaction) (*TransactionRiskResult, error)
	CheckFraudIndicators(ctx context.Context, data interface{}) (*FraudCheckResult, error)
	AssessPaymentRisk(ctx context.Context, payment *entities.Payment) (*PaymentRiskResult, error)
	CheckFinancialThresholds(ctx context.Context) ([]*ThresholdViolation, error)
}

// OperationalRiskMonitor monitors operational risks
type OperationalRiskMonitor interface {
	CheckServiceHealth(ctx context.Context) (*ServiceHealthResult, error)
	MonitorDependencies(ctx context.Context) ([]*DependencyStatus, error)
	AssessCapacityRisk(ctx context.Context) (*CapacityRiskResult, error)
	PredictFailures(ctx context.Context) ([]*FailurePrediction, error)
}

// SecurityRiskScanner scans for security risks
type SecurityRiskScanner interface {
	ScanVulnerabilities(ctx context.Context) ([]*entities.Vulnerability, error)
	CheckAccessPatterns(ctx context.Context) (*AccessPatternResult, error)
	AuditPermissions(ctx context.Context) (*PermissionAuditResult, error)
	DetectAnomalies(ctx context.Context) ([]*SecurityAnomaly, error)
}

// IncidentResponder handles incident response
type IncidentResponder interface {
	ExecutePlaybook(ctx context.Context, incident *entities.Incident) error
	NotifyStakeholders(ctx context.Context, incident *entities.Incident) error
	ContainIncident(ctx context.Context, incidentID string) error
	RecoverFromIncident(ctx context.Context, incidentID string) error
	GeneratePostMortem(ctx context.Context, incidentID string) (*PostMortem, error)
}

// BackupManager manages system backups
type BackupManager interface {
	CreateBackup(ctx context.Context, backupType string) (*entities.Backup, error)
	VerifyBackup(ctx context.Context, backupID string) error
	RestoreFromBackup(ctx context.Context, backupID string) error
	GetBackupStatus(ctx context.Context) (*BackupStatus, error)
	CleanupOldBackups(ctx context.Context) error
}

// RiskDashboard represents the risk management dashboard
type RiskDashboard struct {
	OverallRiskScore      float64                           `json:"overall_risk_score"`
	RisksByType           map[string]*TypeRiskSummary       `json:"risks_by_type"`
	ActiveIncidents       []*entities.Incident              `json:"active_incidents"`
	RecentVulnerabilities []*entities.Vulnerability `json:"recent_vulnerabilities"`
	ComplianceStatus      *ComplianceStatus                 `json:"compliance_status"`
	SystemHealth          *SystemHealthSummary              `json:"system_health"`
}

// TypeRiskSummary represents risk summary for a specific type
type TypeRiskSummary struct {
	Type         entities.RiskType `json:"type"`
	Count        int               `json:"count"`
	AverageScore float64           `json:"average_score"`
	Trend        string            `json:"trend"` // "increasing", "decreasing", "stable"
	TopRisks     []*entities.Risk  `json:"top_risks"`
}

// ThresholdViolation represents a threshold violation
type ThresholdViolation struct {
	ThresholdID  string  `json:"threshold_id"`
	Type         string  `json:"type"`
	Category     string  `json:"category"`
	CurrentValue float64 `json:"current_value"`
	Threshold    float64 `json:"threshold"`
	Severity     string  `json:"severity"`
	Message      string  `json:"message"`
}

// ContentRiskResult represents content analysis results
type ContentRiskResult struct {
	Safe            bool                   `json:"safe"`
	RiskScore       float64                `json:"risk_score"`
	Violations      []*ContentViolation    `json:"violations"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ContentViolation represents a content policy violation
type ContentViolation struct {
	PolicyType string  `json:"policy_type"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
	Location   string  `json:"location"`
	Text       string  `json:"text"`
	Action     string  `json:"action"`
}

// PIICheckResult represents PII detection results
type PIICheckResult struct {
	ContainsPII bool         `json:"contains_pii"`
	PIIEntities []*PIIEntity `json:"pii_entities"`
	Anonymized  string       `json:"anonymized,omitempty"`
}

// PIIEntity represents detected PII
type PIIEntity struct {
	Type       string  `json:"type"` // "email", "phone", "ssn", "credit_card", etc.
	Value      string  `json:"value"`
	Location   int     `json:"location"`
	Confidence float64 `json:"confidence"`
}

// PolicyCheckResult represents policy check results
type PolicyCheckResult struct {
	Compliant  bool               `json:"compliant"`
	Violations []*PolicyViolation `json:"violations"`
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	PolicyID    string `json:"policy_id"`
	PolicyName  string `json:"policy_name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Action      string `json:"action"`
}

// CopyrightCheckResult represents copyright check results
type CopyrightCheckResult struct {
	Clear     bool              `json:"clear"`
	Matches   []*CopyrightMatch `json:"matches"`
	RiskLevel string            `json:"risk_level"`
}

// CopyrightMatch represents a potential copyright match
type CopyrightMatch struct {
	Source     string  `json:"source"`
	Similarity float64 `json:"similarity"`
	URL        string  `json:"url,omitempty"`
	Owner      string  `json:"owner,omitempty"`
}

// TransactionRiskResult represents transaction risk analysis
type TransactionRiskResult struct {
	RiskScore      float64  `json:"risk_score"`
	RiskFactors    []string `json:"risk_factors"`
	RequiresReview bool     `json:"requires_review"`
	Action         string   `json:"action"` // "approve", "review", "reject"
	Reasons        []string `json:"reasons"`
}

// ServiceHealthResult represents service health check results
type ServiceHealthResult struct {
	Healthy             bool             `json:"healthy"`
	Services            []*ServiceStatus `json:"services"`
	FailingServices     []string         `json:"failing_services"`
	OverallAvailability float64          `json:"overall_availability"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	ResponseTime int     `json:"response_time"` // milliseconds
	Uptime       float64 `json:"uptime"`        // percentage
	LastCheck    string  `json:"last_check"`
}

// ComplianceResult represents compliance check results
type ComplianceResult struct {
	Compliant       bool                   `json:"compliant"`
	Violations      []*ComplianceViolation `json:"violations"`
	ComplianceScore float64                `json:"compliance_score"`
	Recommendations []string               `json:"recommendations"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	RequirementID string   `json:"requirement_id"`
	Regulation    string   `json:"regulation"`
	Article       string   `json:"article"`
	Description   string   `json:"description"`
	Severity      string   `json:"severity"`
	Actions       []string `json:"actions"`
}

// IncidentStatus represents current incident status
type IncidentStatus struct {
	IncidentID string                     `json:"incident_id"`
	Status     string                     `json:"status"`
	Actions    []*entities.IncidentAction `json:"actions"`
	Impact     string                     `json:"impact"`
	ETA        string                     `json:"eta,omitempty"`
}

// PostMortem represents incident post-mortem analysis
type PostMortem struct {
	IncidentID     string   `json:"incident_id"`
	Summary        string   `json:"summary"`
	Timeline       []string `json:"timeline"`
	RootCause      string   `json:"root_cause"`
	Impact         string   `json:"impact"`
	WhatWentWell   []string `json:"what_went_well"`
	WhatWentWrong  []string `json:"what_went_wrong"`
	ActionItems    []string `json:"action_items"`
	LessonsLearned []string `json:"lessons_learned"`
}

// SecurityScanResult represents security scan results
type SecurityScanResult struct {
	Type         string   `json:"type"`
	Severity     string   `json:"severity"`
	Component    string   `json:"component"`
	Description  string   `json:"description"`
	CVE          string   `json:"cve,omitempty"`
	Remediation  []string `json:"remediation"`
	Confidence   float64  `json:"confidence"`
}

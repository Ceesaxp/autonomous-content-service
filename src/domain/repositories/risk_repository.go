package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// RiskRepository defines the interface for risk data access
type RiskRepository interface {
	// Risk management
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Risk, error)
	FindByCategory(ctx context.Context, category entities.RiskCategory, offset, limit int) ([]*entities.Risk, int, error)
	FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Risk, int, error)
	FindByStatus(ctx context.Context, status entities.RiskStatus, offset, limit int) ([]*entities.Risk, int, error)
	FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Risk, int, error)
	Save(ctx context.Context, risk *entities.Risk) error
	Create(ctx context.Context, risk *entities.Risk) error
	Update(ctx context.Context, risk *entities.Risk) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Legacy methods for compatibility
	CreateRisk(ctx context.Context, risk *entities.Risk) error
	GetRiskByID(ctx context.Context, id string) (*entities.Risk, error)
	ListRisks(ctx context.Context, filters RiskFilters) ([]*entities.Risk, error)
	UpdateRisk(ctx context.Context, risk *entities.Risk) error
	DeleteRisk(ctx context.Context, id string) error
	GetRisksByType(ctx context.Context, riskType entities.RiskType) ([]*entities.Risk, error)
	GetActiveRisks(ctx context.Context) ([]*entities.Risk, error)
	GetRiskMetrics(ctx context.Context, timeRange TimeRange) (*RiskMetrics, error)

	// Risk thresholds
	CreateThreshold(ctx context.Context, threshold *entities.RiskThreshold) error
	GetThresholdByID(ctx context.Context, id string) (*entities.RiskThreshold, error)
	ListThresholds(ctx context.Context) ([]*entities.RiskThreshold, error)
	UpdateThreshold(ctx context.Context, threshold *entities.RiskThreshold) error
	DeleteThreshold(ctx context.Context, id string) error
	GetThresholdsByType(ctx context.Context, riskType entities.RiskType) ([]*entities.RiskThreshold, error)
	CheckThresholdViolations(ctx context.Context, riskType entities.RiskType, value float64) ([]*entities.RiskThreshold, error)

	// Compliance
	CreateComplianceRequirement(ctx context.Context, req *entities.ComplianceRequirement) error
	GetComplianceRequirement(ctx context.Context, id string) (*entities.ComplianceRequirement, error)
	ListComplianceRequirements(ctx context.Context, regulation string) ([]*entities.ComplianceRequirement, error)
	UpdateComplianceRequirement(ctx context.Context, req *entities.ComplianceRequirement) error
	GetActiveComplianceRequirements(ctx context.Context) ([]*entities.ComplianceRequirement, error)

	// Content policies
	CreateContentPolicy(ctx context.Context, policy *entities.ContentPolicy) error
	GetContentPolicy(ctx context.Context, id string) (*entities.ContentPolicy, error)
	ListContentPolicies(ctx context.Context) ([]*entities.ContentPolicy, error)
	UpdateContentPolicy(ctx context.Context, policy *entities.ContentPolicy) error
	GetActiveContentPolicies(ctx context.Context) ([]*entities.ContentPolicy, error)
	GetContentPoliciesByType(ctx context.Context, policyType string) ([]*entities.ContentPolicy, error)

	// Service dependencies
	CreateServiceDependency(ctx context.Context, dep *entities.ServiceDependency) error
	GetServiceDependency(ctx context.Context, id string) (*entities.ServiceDependency, error)
	ListServiceDependencies(ctx context.Context) ([]*entities.ServiceDependency, error)
	UpdateServiceDependency(ctx context.Context, dep *entities.ServiceDependency) error
	GetCriticalDependencies(ctx context.Context) ([]*entities.ServiceDependency, error)
	UpdateDependencyStatus(ctx context.Context, id string, status string) error

	// Incidents
	CreateIncident(ctx context.Context, incident *entities.Incident) error
	GetIncidentByID(ctx context.Context, id string) (*entities.Incident, error)
	ListIncidents(ctx context.Context, filters IncidentFilters) ([]*entities.Incident, error)
	UpdateIncident(ctx context.Context, incident *entities.Incident) error
	GetActiveIncidents(ctx context.Context) ([]*entities.Incident, error)
	GetIncidentsByService(ctx context.Context, service string) ([]*entities.Incident, error)
	AddIncidentAction(ctx context.Context, incidentID string, action *entities.IncidentAction) error

	// Security vulnerabilities
	CreateVulnerability(ctx context.Context, vuln *entities.Vulnerability) error
	GetVulnerability(ctx context.Context, id string) (*entities.Vulnerability, error)
	ListVulnerabilities(ctx context.Context, filters VulnerabilityFilters) ([]*entities.Vulnerability, error)
	UpdateVulnerability(ctx context.Context, vuln *entities.Vulnerability) error
	GetUnpatchedVulnerabilities(ctx context.Context) ([]*entities.Vulnerability, error)
	MarkVulnerabilityPatched(ctx context.Context, id string) error

	// Backups
	CreateBackupRecord(ctx context.Context, backup *entities.Backup) error
	GetBackupRecord(ctx context.Context, id string) (*entities.Backup, error)
	ListBackupRecords(ctx context.Context, filters BackupFilters) ([]*entities.Backup, error)
	UpdateBackupRecord(ctx context.Context, backup *entities.Backup) error
	GetLastSuccessfulBackup(ctx context.Context, backupType string) (*entities.Backup, error)
	CleanupOldBackups(ctx context.Context, retentionDays int) error
}

// RiskFilters represents filters for querying risks
type RiskFilters struct {
	Type      entities.RiskType
	Severity  entities.RiskSeverity
	Status    entities.RiskStatus
	StartDate time.Time
	EndDate   time.Time
	Source    string
}

// IncidentFilters represents filters for querying incidents
type IncidentFilters struct {
	Type      string
	Severity  entities.RiskSeverity
	Status    string
	Service   string
	StartDate time.Time
	EndDate   time.Time
}

// VulnerabilityFilters represents filters for querying vulnerabilities
type VulnerabilityFilters struct {
	Type      string
	Severity  entities.RiskSeverity
	Component string
	IsPatched bool
}

// BackupFilters represents filters for querying backups
type BackupFilters struct {
	Type      string
	Status    string
	StartDate time.Time
	EndDate   time.Time
}

// RiskMetrics represents aggregated risk metrics
type RiskMetrics struct {
	TotalRisks           int            `json:"total_risks"`
	RisksBySeverity      map[string]int `json:"risks_by_severity"`
	RisksByType          map[string]int `json:"risks_by_type"`
	RisksByStatus        map[string]int `json:"risks_by_status"`
	AverageScore         float64        `json:"average_score"`
	AverageTimeToResolve time.Duration  `json:"average_time_to_resolve"`
	OpenIncidents        int            `json:"open_incidents"`
	VulnerabilityCount   int            `json:"vulnerability_count"`
	ComplianceScore      float64        `json:"compliance_score"`
	LastBackup           time.Time      `json:"last_backup"`
}

// IncidentRepository defines the interface for incident data access
type IncidentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Incident, error)
	FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Incident, int, error)
	FindByStatus(ctx context.Context, status entities.IncidentStatus, offset, limit int) ([]*entities.Incident, int, error)
	FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Incident, int, error)
	Save(ctx context.Context, incident *entities.Incident) error
	Create(ctx context.Context, incident *entities.Incident) error
	Update(ctx context.Context, incident *entities.Incident) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// VulnerabilityRepository defines the interface for vulnerability data access
type VulnerabilityRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Vulnerability, error)
	FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Vulnerability, int, error)
	FindByStatus(ctx context.Context, status entities.VulnerabilityStatus, offset, limit int) ([]*entities.Vulnerability, int, error)
	FindByComponent(ctx context.Context, component string, offset, limit int) ([]*entities.Vulnerability, int, error)
	Save(ctx context.Context, vulnerability *entities.Vulnerability) error
	Create(ctx context.Context, vulnerability *entities.Vulnerability) error
	Update(ctx context.Context, vulnerability *entities.Vulnerability) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// BackupRepository defines the interface for backup data access
type BackupRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Backup, error)
	FindByStatus(ctx context.Context, status entities.BackupStatus, offset, limit int) ([]*entities.Backup, int, error)
	FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Backup, int, error)
	FindExpired(ctx context.Context) ([]*entities.Backup, error)
	Save(ctx context.Context, backup *entities.Backup) error
	Create(ctx context.Context, backup *entities.Backup) error
	Update(ctx context.Context, backup *entities.Backup) error
	Delete(ctx context.Context, id uuid.UUID) error
}

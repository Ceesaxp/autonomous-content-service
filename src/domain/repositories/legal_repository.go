package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// LegalRepository defines the interface for legal and compliance data operations
type LegalRepository interface {
	// Contract Management
	CreateContract(ctx context.Context, contract *entities.Contract) error
	GetContractByID(ctx context.Context, id uuid.UUID) (*entities.Contract, error)
	ListContracts(ctx context.Context, filters ContractFilters) ([]*entities.Contract, error)
	UpdateContract(ctx context.Context, contract *entities.Contract) error
	DeleteContract(ctx context.Context, id uuid.UUID) error
	GetContractsByClientID(ctx context.Context, clientID uuid.UUID) ([]*entities.Contract, error)
	GetContractsByStatus(ctx context.Context, status entities.ContractStatus) ([]*entities.Contract, error)
	GetActiveContracts(ctx context.Context) ([]*entities.Contract, error)
	GetExpiringContracts(ctx context.Context, days int) ([]*entities.Contract, error)
	ArchiveContract(ctx context.Context, id uuid.UUID) error

	// Contract Templates
	CreateContractTemplate(ctx context.Context, template *entities.ContractTemplate) error
	GetContractTemplateByID(ctx context.Context, id uuid.UUID) (*entities.ContractTemplate, error)
	ListContractTemplates(ctx context.Context, filters TemplateFilters) ([]*entities.ContractTemplate, error)
	UpdateContractTemplate(ctx context.Context, template *entities.ContractTemplate) error
	DeleteContractTemplate(ctx context.Context, id uuid.UUID) error
	GetActiveTemplates(ctx context.Context) ([]*entities.ContractTemplate, error)
	GetTemplatesByType(ctx context.Context, contractType entities.ContractType) ([]*entities.ContractTemplate, error)

	// Signatures
	CreateSignature(ctx context.Context, signature *entities.ContractSignature) error
	GetSignatureByID(ctx context.Context, id uuid.UUID) (*entities.ContractSignature, error)
	GetSignaturesByContract(ctx context.Context, contractID uuid.UUID) ([]*entities.ContractSignature, error)
	UpdateSignature(ctx context.Context, signature *entities.ContractSignature) error
	VerifySignature(ctx context.Context, id uuid.UUID) (bool, error)
	GetPendingSignatures(ctx context.Context) ([]*entities.ContractSignature, error)
	GetExpiredSignatures(ctx context.Context) ([]*entities.ContractSignature, error)

	// Compliance Checks
	CreateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error
	GetComplianceCheckByID(ctx context.Context, id uuid.UUID) (*entities.ComplianceCheck, error)
	ListComplianceChecks(ctx context.Context, filters ComplianceFilters) ([]*entities.ComplianceCheck, error)
	UpdateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error
	GetChecksByRegulation(ctx context.Context, regulation string) ([]*entities.ComplianceCheck, error)
	GetChecksByStatus(ctx context.Context, status entities.ComplianceStatus) ([]*entities.ComplianceCheck, error)
	GetOverdueChecks(ctx context.Context) ([]*entities.ComplianceCheck, error)

	// IP Licenses
	CreateIPLicense(ctx context.Context, license *entities.IPLicense) error
	GetIPLicenseByID(ctx context.Context, id uuid.UUID) (*entities.IPLicense, error)
	ListIPLicenses(ctx context.Context, filters IPLicenseFilters) ([]*entities.IPLicense, error)
	UpdateIPLicense(ctx context.Context, license *entities.IPLicense) error
	DeleteIPLicense(ctx context.Context, id uuid.UUID) error
	GetLicensesByType(ctx context.Context, ipType entities.IPType) ([]*entities.IPLicense, error)
	GetExpiringLicenses(ctx context.Context, days int) ([]*entities.IPLicense, error)
	GetActiveLicenses(ctx context.Context) ([]*entities.IPLicense, error)

	// Insurance Policies
	CreateInsurancePolicy(ctx context.Context, policy *entities.InsurancePolicy) error
	GetInsurancePolicyByID(ctx context.Context, id uuid.UUID) (*entities.InsurancePolicy, error)
	ListInsurancePolicies(ctx context.Context, filters InsuranceFilters) ([]*entities.InsurancePolicy, error)
	UpdateInsurancePolicy(ctx context.Context, policy *entities.InsurancePolicy) error
	DeleteInsurancePolicy(ctx context.Context, id uuid.UUID) error
	GetPoliciesByType(ctx context.Context, insuranceType entities.InsuranceType) ([]*entities.InsurancePolicy, error)
	GetExpiringPolicies(ctx context.Context, days int) ([]*entities.InsurancePolicy, error)
	GetActivePolicies(ctx context.Context) ([]*entities.InsurancePolicy, error)

	// Dispute Resolution
	CreateDispute(ctx context.Context, dispute *entities.DisputeResolution) error
	GetDisputeByID(ctx context.Context, id uuid.UUID) (*entities.DisputeResolution, error)
	ListDisputes(ctx context.Context, filters DisputeFilters) ([]*entities.DisputeResolution, error)
	UpdateDispute(ctx context.Context, dispute *entities.DisputeResolution) error
	GetDisputesByContract(ctx context.Context, contractID uuid.UUID) ([]*entities.DisputeResolution, error)
	GetActiveDisputes(ctx context.Context) ([]*entities.DisputeResolution, error)
	AddDisputeEvent(ctx context.Context, disputeID uuid.UUID, event *entities.DisputeEvent) error

	// Risk Assessments
	CreateRiskAssessment(ctx context.Context, assessment *entities.LegalRiskAssessment) error
	GetRiskAssessmentByID(ctx context.Context, id uuid.UUID) (*entities.LegalRiskAssessment, error)
	GetRiskAssessmentByContract(ctx context.Context, contractID uuid.UUID) (*entities.LegalRiskAssessment, error)
	ListRiskAssessments(ctx context.Context, filters RiskAssessmentFilters) ([]*entities.LegalRiskAssessment, error)
	UpdateRiskAssessment(ctx context.Context, assessment *entities.LegalRiskAssessment) error
	GetHighRiskAssessments(ctx context.Context) ([]*entities.LegalRiskAssessment, error)
	GetExpiredAssessments(ctx context.Context) ([]*entities.LegalRiskAssessment, error)

	// Regulatory Reporting
	CreateRegulatoryReport(ctx context.Context, report *entities.RegulatoryReport) error
	GetRegulatoryReportByID(ctx context.Context, id uuid.UUID) (*entities.RegulatoryReport, error)
	ListRegulatoryReports(ctx context.Context, filters ReportFilters) ([]*entities.RegulatoryReport, error)
	UpdateRegulatoryReport(ctx context.Context, report *entities.RegulatoryReport) error
	GetReportsByRegulation(ctx context.Context, regulation string) ([]*entities.RegulatoryReport, error)
	GetReportsByAuthority(ctx context.Context, authority string) ([]*entities.RegulatoryReport, error)
	GetOverdueReports(ctx context.Context) ([]*entities.RegulatoryReport, error)
	GetUpcomingReports(ctx context.Context, days int) ([]*entities.RegulatoryReport, error)

	// Audit Trails
	AddContractAuditEntry(ctx context.Context, contractID uuid.UUID, entry *entities.ContractAuditEntry) error
	GetContractAuditTrail(ctx context.Context, contractID uuid.UUID) ([]*entities.ContractAuditEntry, error)
	GetAuditTrailByTimeRange(ctx context.Context, start, end time.Time) ([]*entities.ContractAuditEntry, error)
	VerifyAuditTrail(ctx context.Context, contractID uuid.UUID) (bool, error)
}

// Filter structures for repository queries

type ContractFilters struct {
	Status      *entities.ContractStatus
	Type        *entities.ContractType
	ClientID    *uuid.UUID
	ProjectID   *uuid.UUID
	TemplateID  *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	IsExpiring  *bool
	ExpiringDays *int
	Offset      int
	Limit       int
}

type TemplateFilters struct {
	Type     *entities.ContractType
	IsActive *bool
	Version  *int
	Offset   int
	Limit    int
}

type ComplianceFilters struct {
	Type       *entities.ComplianceType
	Regulation *string
	Status     *entities.ComplianceStatus
	RiskLevel  *entities.RiskLevel
	StartDate  *time.Time
	EndDate    *time.Time
	Offset     int
	Limit      int
}

type IPLicenseFilters struct {
	Type        *entities.LicenseType
	IPType      *entities.IPType
	IsExclusive *bool
	IsActive    *bool
	StartDate   *time.Time
	EndDate     *time.Time
	Offset      int
	Limit       int
}

type InsuranceFilters struct {
	Type      *entities.InsuranceType
	Provider  *string
	Status    *entities.PolicyStatus
	IsActive  *bool
	StartDate *time.Time
	EndDate   *time.Time
	Offset    int
	Limit     int
}

type DisputeFilters struct {
	Type             *entities.DisputeType
	Status           *entities.DisputeStatus
	ContractID       *uuid.UUID
	ResolutionMethod *entities.ResolutionMethod
	StartDate        *time.Time
	EndDate          *time.Time
	Offset           int
	Limit            int
}

type RiskAssessmentFilters struct {
	RiskLevel        *entities.RiskLevel
	ContractID       *uuid.UUID
	RequiresReview   *bool
	RequiresInsurance *bool
	IsExpired        *bool
	StartDate        *time.Time
	EndDate          *time.Time
	Offset           int
	Limit            int
}

type ReportFilters struct {
	Type       *entities.ReportType
	Regulation *string
	Authority  *string
	Status     *entities.ReportStatus
	Period     *string
	StartDate  *time.Time
	EndDate    *time.Time
	Offset     int
	Limit      int
}
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Contract represents a legal contract in the system
type Contract struct {
	ID                uuid.UUID              `json:"id" db:"contract_id"`
	Title             string                 `json:"title" db:"title"`
	Type              ContractType           `json:"type" db:"type"`
	Status            ContractStatus         `json:"status" db:"status"`
	Version           int                    `json:"version" db:"version"`
	ParentContractID  *uuid.UUID             `json:"parent_contract_id,omitempty" db:"parent_contract_id"`
	ClientID          uuid.UUID              `json:"client_id" db:"client_id"`
	ProjectID         *uuid.UUID             `json:"project_id,omitempty" db:"project_id"`
	TemplateID        uuid.UUID              `json:"template_id" db:"template_id"`
	Content           string                 `json:"content" db:"content"`
	Parameters        map[string]interface{} `json:"parameters" db:"parameters"`
	Terms             []ContractTerm         `json:"terms" db:"terms"`
	Signatures        []ContractSignature    `json:"signatures" db:"signatures"`
	EffectiveDate     time.Time              `json:"effective_date" db:"effective_date"`
	ExpirationDate    *time.Time             `json:"expiration_date,omitempty" db:"expiration_date"`
	RenewalDate       *time.Time             `json:"renewal_date,omitempty" db:"renewal_date"`
	ComplianceChecks  []LegalComplianceCheck `json:"compliance_checks" db:"compliance_checks"`
	DisputeResolution *DisputeResolution     `json:"dispute_resolution,omitempty" db:"dispute_resolution"`
	IPLicenses        []IPLicense            `json:"ip_licenses" db:"ip_licenses"`
	InsuranceRequired bool                   `json:"insurance_required" db:"insurance_required"`
	InsurancePolicies []InsurancePolicy      `json:"insurance_policies" db:"insurance_policies"`
	RiskAssessment    *LegalRiskAssessment   `json:"risk_assessment,omitempty" db:"risk_assessment"`
	AuditTrail        []ContractAuditEntry   `json:"audit_trail" db:"audit_trail"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	ArchivedAt        *time.Time             `json:"archived_at,omitempty" db:"archived_at"`
}

// ContractTemplate represents a template for generating contracts
type ContractTemplate struct {
	ID           uuid.UUID              `json:"id" db:"template_id"`
	Name         string                 `json:"name" db:"name"`
	Type         ContractType           `json:"type" db:"type"`
	Version      int                    `json:"version" db:"version"`
	Content      string                 `json:"content" db:"content"`
	Parameters   []TemplateParameter    `json:"parameters" db:"parameters"`
	DefaultTerms []ContractTerm         `json:"default_terms" db:"default_terms"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ContractTerm represents a specific term or clause in a contract
type ContractTerm struct {
	ID          uuid.UUID   `json:"id" db:"term_id"`
	Type        TermType    `json:"type" db:"type"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Content     string      `json:"content" db:"content"`
	IsMandatory bool        `json:"is_mandatory" db:"is_mandatory"`
	Order       int         `json:"order" db:"order"`
	Metadata    interface{} `json:"metadata" db:"metadata"`
}

// ContractSignature represents a signature on a contract
type ContractSignature struct {
	ID               uuid.UUID       `json:"id" db:"signature_id"`
	SignerName       string          `json:"signer_name" db:"signer_name"`
	SignerEmail      string          `json:"signer_email" db:"signer_email"`
	SignerRole       string          `json:"signer_role" db:"signer_role"`
	SignatureType    SignatureType   `json:"signature_type" db:"signature_type"`
	SignatureData    string          `json:"signature_data" db:"signature_data"`
	SignatureHash    string          `json:"signature_hash" db:"signature_hash"`
	IPAddress        string          `json:"ip_address" db:"ip_address"`
	UserAgent        string          `json:"user_agent" db:"user_agent"`
	Timestamp        time.Time       `json:"timestamp" db:"timestamp"`
	Status           SignatureStatus `json:"status" db:"status"`
	VerificationHash string          `json:"verification_hash" db:"verification_hash"`
	CertificateID    *string         `json:"certificate_id,omitempty" db:"certificate_id"`
	IsValid          bool            `json:"is_valid" db:"is_valid"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
}

// ComplianceCheck represents a compliance verification
type LegalComplianceCheck struct { // prevent clash with HR
	ID          uuid.UUID        `json:"id" db:"check_id"`
	Type        ComplianceType   `json:"type" db:"type"`
	Regulation  string           `json:"regulation" db:"regulation"`
	Requirement string           `json:"requirement" db:"requirement"`
	Status      ComplianceStatus `json:"status" db:"status"`
	Result      string           `json:"result" db:"result"`
	Evidence    []string         `json:"evidence" db:"evidence"`
	CheckedAt   time.Time        `json:"checked_at" db:"checked_at"`
	CheckedBy   string           `json:"checked_by" db:"checked_by"`
	NextCheck   *time.Time       `json:"next_check,omitempty" db:"next_check"`
	Remediation *string          `json:"remediation,omitempty" db:"remediation"`
	RiskLevel   RiskLevel        `json:"risk_level" db:"risk_level"`
}

// IPLicense represents an intellectual property license
type IPLicense struct {
	ID             uuid.UUID    `json:"id" db:"license_id"`
	Type           LicenseType  `json:"type" db:"type"`
	Name           string       `json:"name" db:"name"`
	LicensorName   string       `json:"licensor_name" db:"licensor_name"`
	LicenseeName   string       `json:"licensee_name" db:"licensee_name"`
	IPType         IPType       `json:"ip_type" db:"ip_type"`
	IPDescription  string       `json:"ip_description" db:"ip_description"`
	UsageRights    []UsageRight `json:"usage_rights" db:"usage_rights"`
	Restrictions   []string     `json:"restrictions" db:"restrictions"`
	Territory      string       `json:"territory" db:"territory"`
	EffectiveDate  time.Time    `json:"effective_date" db:"effective_date"`
	ExpirationDate *time.Time   `json:"expiration_date,omitempty" db:"expiration_date"`
	RoyaltyRate    *float64     `json:"royalty_rate,omitempty" db:"royalty_rate"`
	Fee            *Money       `json:"fee,omitempty" db:"fee"`
	IsExclusive    bool         `json:"is_exclusive" db:"is_exclusive"`
	IsActive       bool         `json:"is_active" db:"is_active"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

// InsurancePolicy represents an insurance policy
type InsurancePolicy struct {
	ID              uuid.UUID     `json:"id" db:"policy_id"`
	Type            InsuranceType `json:"type" db:"type"`
	PolicyNumber    string        `json:"policy_number" db:"policy_number"`
	Provider        string        `json:"provider" db:"provider"`
	Coverage        Money         `json:"coverage" db:"coverage"`
	Deductible      Money         `json:"deductible" db:"deductible"`
	Premium         Money         `json:"premium" db:"premium"`
	EffectiveDate   time.Time     `json:"effective_date" db:"effective_date"`
	ExpirationDate  time.Time     `json:"expiration_date" db:"expiration_date"`
	RenewalDate     time.Time     `json:"renewal_date" db:"renewal_date"`
	CoverageDetails []string      `json:"coverage_details" db:"coverage_details"`
	Exclusions      []string      `json:"exclusions" db:"exclusions"`
	Status          PolicyStatus  `json:"status" db:"status"`
	IsActive        bool          `json:"is_active" db:"is_active"`
	DocumentURL     string        `json:"document_url" db:"document_url"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// DisputeResolution represents dispute resolution information
type DisputeResolution struct {
	ID               uuid.UUID        `json:"id" db:"dispute_id"`
	ContractID       uuid.UUID        `json:"contract_id" db:"contract_id"`
	Type             DisputeType      `json:"type" db:"type"`
	Status           DisputeStatus    `json:"status" db:"status"`
	Description      string           `json:"description" db:"description"`
	InitiatedBy      string           `json:"initiated_by" db:"initiated_by"`
	ResolutionMethod ResolutionMethod `json:"resolution_method" db:"resolution_method"`
	Mediator         *string          `json:"mediator,omitempty" db:"mediator"`
	Arbitrator       *string          `json:"arbitrator,omitempty" db:"arbitrator"`
	Venue            *string          `json:"venue,omitempty" db:"venue"`
	GoverningLaw     string           `json:"governing_law" db:"governing_law"`
	Timeline         []DisputeEvent   `json:"timeline" db:"timeline"`
	Resolution       *string          `json:"resolution,omitempty" db:"resolution"`
	Cost             *Money           `json:"cost,omitempty" db:"cost"`
	InitiatedAt      time.Time        `json:"initiated_at" db:"initiated_at"`
	ResolvedAt       *time.Time       `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// LegalRiskAssessment represents a legal risk assessment
type LegalRiskAssessment struct {
	ID                uuid.UUID    `json:"id" db:"assessment_id"`
	ContractID        uuid.UUID    `json:"contract_id" db:"contract_id"`
	RiskLevel         RiskLevel    `json:"risk_level" db:"risk_level"`
	RiskScore         float64      `json:"risk_score" db:"risk_score"`
	RiskFactors       []RiskFactor `json:"risk_factors" db:"risk_factors"`
	Recommendations   []string     `json:"recommendations" db:"recommendations"`
	RequiredClauses   []string     `json:"required_clauses" db:"required_clauses"`
	ComplianceIssues  []string     `json:"compliance_issues" db:"compliance_issues"`
	InsuranceRequired bool         `json:"insurance_required" db:"insurance_required"`
	LegalReview       bool         `json:"legal_review" db:"legal_review"`
	AssessedBy        string       `json:"assessed_by" db:"assessed_by"`
	AssessedAt        time.Time    `json:"assessed_at" db:"assessed_at"`
	ExpiresAt         time.Time    `json:"expires_at" db:"expires_at"`
}

// RegulatoryReport represents a regulatory filing or report
type RegulatoryReport struct {
	ID             uuid.UUID    `json:"id" db:"report_id"`
	Type           ReportType   `json:"type" db:"type"`
	Regulation     string       `json:"regulation" db:"regulation"`
	Authority      string       `json:"authority" db:"authority"`
	Period         string       `json:"period" db:"period"`
	Status         ReportStatus `json:"status" db:"status"`
	Content        string       `json:"content" db:"content"`
	Data           interface{}  `json:"data" db:"data"`
	FilingDeadline time.Time    `json:"filing_deadline" db:"filing_deadline"`
	FiledAt        *time.Time   `json:"filed_at,omitempty" db:"filed_at"`
	ConfirmationID *string      `json:"confirmation_id,omitempty" db:"confirmation_id"`
	DocumentURL    *string      `json:"document_url,omitempty" db:"document_url"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

// Supporting types and enums

type ContractType string

const (
	ContractTypeService     ContractType = "Service"
	ContractTypeEmployment  ContractType = "Employment"
	ContractTypeNDA         ContractType = "NDA"
	ContractTypeLicense     ContractType = "License"
	ContractTypePartnership ContractType = "Partnership"
	ContractTypeVendor      ContractType = "Vendor"
	ContractTypeClient      ContractType = "Client"
)

type ContractStatus string

const (
	ContractStatusDraft      ContractStatus = "Draft"
	ContractStatusReview     ContractStatus = "Review"
	ContractStatusPending    ContractStatus = "Pending"
	ContractStatusSigned     ContractStatus = "Signed"
	ContractStatusActive     ContractStatus = "Active"
	ContractStatusExpired    ContractStatus = "Expired"
	ContractStatusTerminated ContractStatus = "Terminated"
	ContractStatusDisputed   ContractStatus = "Disputed"
	ContractStatusArchived   ContractStatus = "Archived"
)

type TermType string

const (
	TermTypePayment         TermType = "Payment"
	TermTypeDelivery        TermType = "Delivery"
	TermTypeIP              TermType = "IP"
	TermTypeConfidentiality TermType = "Confidentiality"
	TermTypeTermination     TermType = "Termination"
	TermTypeLiability       TermType = "Liability"
	TermTypeDispute         TermType = "Dispute"
	TermTypeGoverning       TermType = "Governing"
)

type SignatureType string

const (
	SignatureTypeElectronic SignatureType = "Electronic"
	SignatureTypeDigital    SignatureType = "Digital"
	SignatureTypeWet        SignatureType = "Wet"
	SignatureTypeDocuSign   SignatureType = "DocuSign"
)

type SignatureStatus string

const (
	SignatureStatusPending  SignatureStatus = "Pending"
	SignatureStatusSigned   SignatureStatus = "Signed"
	SignatureStatusDeclined SignatureStatus = "Declined"
	SignatureStatusExpired  SignatureStatus = "Expired"
	SignatureStatusInvalid  SignatureStatus = "Invalid"
)

type ComplianceType string

const (
	ComplianceTypeGDPR     ComplianceType = "GDPR"
	ComplianceTypeCCPA     ComplianceType = "CCPA"
	ComplianceTypeSOX      ComplianceType = "SOX"
	ComplianceTypeHIPAA    ComplianceType = "HIPAA"
	ComplianceTypeSOC2     ComplianceType = "SOC2"
	ComplianceTypeISO27001 ComplianceType = "ISO27001"
	ComplianceTypeCOPPA    ComplianceType = "COPPA"
)

type ComplianceStatus string

const (
	ComplianceStatusCompliant    ComplianceStatus = "Compliant"
	ComplianceStatusNonCompliant ComplianceStatus = "NonCompliant"
	ComplianceStatusPending      ComplianceStatus = "Pending"
	ComplianceStatusExempt       ComplianceStatus = "Exempt"
	ComplianceStatusUnknown      ComplianceStatus = "Unknown"
)

type LicenseType string

const (
	LicenseTypeExclusive       LicenseType = "Exclusive"
	LicenseTypeNonExclusive    LicenseType = "NonExclusive"
	LicenseTypeSole            LicenseType = "Sole"
	LicenseTypeCreativeCommons LicenseType = "CreativeCommons"
	LicenseTypeMIT             LicenseType = "MIT"
	LicenseTypeGPL             LicenseType = "GPL"
	LicenseTypeProprietary     LicenseType = "Proprietary"
)

type IPType string

const (
	IPTypeCopyright   IPType = "Copyright"
	IPTypeTrademark   IPType = "Trademark"
	IPTypePatent      IPType = "Patent"
	IPTypeTradeSecret IPType = "TradeSecret"
	IPTypeSoftware    IPType = "Software"
	IPTypeContent     IPType = "Content"
)

type UsageRight string

const (
	UsageRightUse         UsageRight = "Use"
	UsageRightModify      UsageRight = "Modify"
	UsageRightDistribute  UsageRight = "Distribute"
	UsageRightSublicense  UsageRight = "Sublicense"
	UsageRightCommercial  UsageRight = "Commercial"
	UsageRightAttribution UsageRight = "Attribution"
)

type InsuranceType string

const (
	InsuranceTypeGeneral      InsuranceType = "General"
	InsuranceTypeProfessional InsuranceType = "Professional"
	InsuranceTypeCyber        InsuranceType = "Cyber"
	InsuranceTypeErrors       InsuranceType = "Errors"
	InsuranceTypeDirectors    InsuranceType = "Directors"
)

type PolicyStatus string

const (
	PolicyStatusActive    PolicyStatus = "Active"
	PolicyStatusInactive  PolicyStatus = "Inactive"
	PolicyStatusExpired   PolicyStatus = "Expired"
	PolicyStatusCancelled PolicyStatus = "Cancelled"
	PolicyStatusPending   PolicyStatus = "Pending"
)

type DisputeType string

const (
	DisputeTypeBreach      DisputeType = "Breach"
	DisputeTypePayment     DisputeType = "Payment"
	DisputeTypeDelivery    DisputeType = "Delivery"
	DisputeTypeQuality     DisputeType = "Quality"
	DisputeTypeIP          DisputeType = "IP"
	DisputeTypeTermination DisputeType = "Termination"
)

type DisputeStatus string

const (
	DisputeStatusOpen        DisputeStatus = "Open"
	DisputeStatusMediation   DisputeStatus = "Mediation"
	DisputeStatusArbitration DisputeStatus = "Arbitration"
	DisputeStatusLitigation  DisputeStatus = "Litigation"
	DisputeStatusResolved    DisputeStatus = "Resolved"
	DisputeStatusClosed      DisputeStatus = "Closed"
)

type ResolutionMethod string

const (
	ResolutionMethodNegotiation ResolutionMethod = "Negotiation"
	ResolutionMethodMediation   ResolutionMethod = "Mediation"
	ResolutionMethodArbitration ResolutionMethod = "Arbitration"
	ResolutionMethodLitigation  ResolutionMethod = "Litigation"
)

// RiskLevel is defined in pricing.go and imported here

type ReportType string

const (
	ReportTypeQuarterly ReportType = "Quarterly"
	ReportTypeAnnual    ReportType = "Annual"
	ReportTypeMonthly   ReportType = "Monthly"
	ReportTypeAdhoc     ReportType = "Adhoc"
	ReportTypeIncident  ReportType = "Incident"
)

type ReportStatus string

const (
	ReportStatusDraft    ReportStatus = "Draft"
	ReportStatusReview   ReportStatus = "Review"
	ReportStatusPending  ReportStatus = "Pending"
	ReportStatusFiled    ReportStatus = "Filed"
	ReportStatusRejected ReportStatus = "Rejected"
	ReportStatusAmended  ReportStatus = "Amended"
)

// Supporting structures

type TemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Validation   string      `json:"validation"`
}

type ContractAuditEntry struct {
	ID        uuid.UUID `json:"id"`
	Action    string    `json:"action"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	UserID    string    `json:"user_id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
}

type RiskFactor struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Likelihood  float64 `json:"likelihood"`
	Impact      float64 `json:"impact"`
	Mitigation  string  `json:"mitigation"`
}

type DisputeEvent struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Actor       string    `json:"actor"`
	Timestamp   time.Time `json:"timestamp"`
	Evidence    []string  `json:"evidence"`
}

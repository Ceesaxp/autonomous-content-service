package entities

import (
	"time"

	"github.com/google/uuid"
)

// TalentType represents the type of talent (human or AI)
type TalentType string

const (
	TalentTypeHuman TalentType = "Human"
	TalentTypeAI    TalentType = "AI"
)

// TalentStatus represents the status of a talent
type TalentStatus string

const (
	TalentStatusAvailable   TalentStatus = "Available"
	TalentStatusEngaged     TalentStatus = "Engaged"
	TalentStatusUnavailable TalentStatus = "Unavailable"
	TalentStatusOffboarded  TalentStatus = "Offboarded"
	TalentStatusDeprecated  TalentStatus = "Deprecated"
)

// EngagementStatus represents the status of an engagement
type EngagementStatus string

const (
	EngagementStatusDraft      EngagementStatus = "Draft"
	EngagementStatusPending    EngagementStatus = "Pending"
	EngagementStatusActive     EngagementStatus = "Active"
	EngagementStatusPaused     EngagementStatus = "Paused"
	EngagementStatusCompleted  EngagementStatus = "Completed"
	EngagementStatusTerminated EngagementStatus = "Terminated"
)

// EngagementType represents the type of engagement
type EngagementType string

const (
	EngagementTypeFullTime   EngagementType = "FullTime"
	EngagementTypePartTime   EngagementType = "PartTime"
	EngagementTypeContract   EngagementType = "Contract"
	EngagementTypeProject    EngagementType = "Project"
	EngagementTypeAPIAccess  EngagementType = "APIAccess"
)

// SkillLevel represents the proficiency level of a skill
type SkillLevel string

const (
	SkillLevelBeginner     SkillLevel = "Beginner"
	SkillLevelIntermediate SkillLevel = "Intermediate"
	SkillLevelAdvanced     SkillLevel = "Advanced"
	SkillLevelExpert       SkillLevel = "Expert"
)

// TrainingStatus represents the status of a training program
type TrainingStatus string

const (
	TrainingStatusNotStarted TrainingStatus = "NotStarted"
	TrainingStatusInProgress TrainingStatus = "InProgress"
	TrainingStatusCompleted  TrainingStatus = "Completed"
	TrainingStatusFailed     TrainingStatus = "Failed"
	TrainingStatusExpired    TrainingStatus = "Expired"
)

// PerformanceRating represents performance rating levels
type PerformanceRating string

const (
	PerformanceRatingExceptional PerformanceRating = "Exceptional"
	PerformanceRatingExceeds     PerformanceRating = "ExceedsExpectations"
	PerformanceRatingMeets       PerformanceRating = "MeetsExpectations"
	PerformanceRatingNeeds       PerformanceRating = "NeedsImprovement"
	PerformanceRatingUnsatisfactory PerformanceRating = "Unsatisfactory"
)

// ApplicationStatus represents the status of a talent application
type ApplicationStatus string

const (
	ApplicationStatusNew        ApplicationStatus = "New"
	ApplicationStatusScreening  ApplicationStatus = "Screening"
	ApplicationStatusInterview  ApplicationStatus = "Interview"
	ApplicationStatusAssessment ApplicationStatus = "Assessment"
	ApplicationStatusApproved   ApplicationStatus = "Approved"
	ApplicationStatusRejected   ApplicationStatus = "Rejected"
	ApplicationStatusWithdrawn  ApplicationStatus = "Withdrawn"
)

// Talent represents a base talent entity (human or AI)
type Talent struct {
	ID               uuid.UUID              `json:"id" db:"talent_id"`
	Type             TalentType             `json:"type" db:"type"`
	Name             string                 `json:"name" db:"name"`
	Email            string                 `json:"email,omitempty" db:"email"`
	Status           TalentStatus           `json:"status" db:"status"`
	ReputationScore  float64                `json:"reputation_score" db:"reputation_score"`
	Skills           []Skill                `json:"skills" db:"-"`
	Certifications   []Certification        `json:"certifications" db:"-"`
	Availability     map[string]interface{} `json:"availability" db:"availability"`
	HourlyRate       *Money                 `json:"hourly_rate,omitempty" db:"-"`
	Currency         string                 `json:"currency" db:"currency"`
	Location         string                 `json:"location,omitempty" db:"location"`
	Timezone         string                 `json:"timezone,omitempty" db:"timezone"`
	ProfileData      map[string]interface{} `json:"profile_data" db:"profile_data"`
	LastActiveAt     *time.Time             `json:"last_active_at" db:"last_active_at"`
	OnboardedAt      *time.Time             `json:"onboarded_at" db:"onboarded_at"`
	OffboardedAt     *time.Time             `json:"offboarded_at" db:"offboarded_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// HumanContributor represents a human talent with additional fields
type HumanContributor struct {
	Talent
	Phone            string                 `json:"phone,omitempty" db:"phone"`
	LinkedInURL      string                 `json:"linkedin_url,omitempty" db:"linkedin_url"`
	PortfolioURL     string                 `json:"portfolio_url,omitempty" db:"portfolio_url"`
	YearsExperience  int                    `json:"years_experience" db:"years_experience"`
	PreferredHours   int                    `json:"preferred_hours" db:"preferred_hours"`
	Languages        []string               `json:"languages" db:"-"`
	WorkAuthorization map[string]interface{} `json:"work_authorization" db:"work_authorization"`
}

// AIAgent represents an AI agent with specific capabilities
type AIAgent struct {
	Talent
	Provider         string                 `json:"provider" db:"provider"`
	Model            string                 `json:"model" db:"model"`
	APIEndpoint      string                 `json:"api_endpoint" db:"api_endpoint"`
	APIVersion       string                 `json:"api_version" db:"api_version"`
	Capabilities     []string               `json:"capabilities" db:"-"`
	RateLimits       map[string]interface{} `json:"rate_limits" db:"rate_limits"`
	CostPerRequest   *Money                 `json:"cost_per_request" db:"-"`
	CostPerToken     *Money                 `json:"cost_per_token" db:"-"`
	MaxTokens        int                    `json:"max_tokens" db:"max_tokens"`
	ResponseTimeMs   int                    `json:"response_time_ms" db:"response_time_ms"`
	Reliability      float64                `json:"reliability" db:"reliability"`
	LastHealthCheck  *time.Time             `json:"last_health_check" db:"last_health_check"`
}

// Skill represents a skill or capability
type Skill struct {
	ID          uuid.UUID  `json:"id" db:"skill_id"`
	TalentID    uuid.UUID  `json:"talent_id" db:"talent_id"`
	Name        string     `json:"name" db:"name"`
	Category    string     `json:"category" db:"category"`
	Level       SkillLevel `json:"level" db:"level"`
	YearsUsed   float64    `json:"years_used" db:"years_used"`
	LastUsed    *time.Time `json:"last_used" db:"last_used"`
	Verified    bool       `json:"verified" db:"verified"`
	VerifiedBy  string     `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt  *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Certification represents a professional certification or credential
type Certification struct {
	ID             uuid.UUID  `json:"id" db:"certification_id"`
	TalentID       uuid.UUID  `json:"talent_id" db:"talent_id"`
	Name           string     `json:"name" db:"name"`
	Issuer         string     `json:"issuer" db:"issuer"`
	CredentialID   string     `json:"credential_id" db:"credential_id"`
	IssueDate      time.Time  `json:"issue_date" db:"issue_date"`
	ExpiryDate     *time.Time `json:"expiry_date" db:"expiry_date"`
	VerificationURL string    `json:"verification_url,omitempty" db:"verification_url"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// Engagement represents an active work relationship
type Engagement struct {
	ID                uuid.UUID              `json:"id" db:"engagement_id"`
	TalentID          uuid.UUID              `json:"talent_id" db:"talent_id"`
	Type              EngagementType         `json:"type" db:"type"`
	Status            EngagementStatus       `json:"status" db:"status"`
	Title             string                 `json:"title" db:"title"`
	Description       string                 `json:"description" db:"description"`
	StartDate         time.Time              `json:"start_date" db:"start_date"`
	EndDate           *time.Time             `json:"end_date" db:"end_date"`
	HoursPerWeek      int                    `json:"hours_per_week" db:"hours_per_week"`
	RateType          string                 `json:"rate_type" db:"rate_type"` // Hourly, Fixed, Retainer
	Rate              *Money                 `json:"rate" db:"-"`
	Currency          string                 `json:"currency" db:"currency"`
	ContractID        *uuid.UUID             `json:"contract_id" db:"contract_id"`
	ManagerID         *uuid.UUID             `json:"manager_id" db:"manager_id"`
	TeamID            *uuid.UUID             `json:"team_id" db:"team_id"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics" db:"performance_metrics"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// WorkAssignment represents a specific task or project assigned
type WorkAssignment struct {
	ID              uuid.UUID              `json:"id" db:"assignment_id"`
	EngagementID    uuid.UUID              `json:"engagement_id" db:"engagement_id"`
	TalentID        uuid.UUID              `json:"talent_id" db:"talent_id"`
	ProjectID       *uuid.UUID             `json:"project_id" db:"project_id"`
	Title           string                 `json:"title" db:"title"`
	Description     string                 `json:"description" db:"description"`
	Priority        Priority               `json:"priority" db:"priority"`
	Status          string                 `json:"status" db:"status"`
	EstimatedHours  float64                `json:"estimated_hours" db:"estimated_hours"`
	ActualHours     float64                `json:"actual_hours" db:"actual_hours"`
	DueDate         *time.Time             `json:"due_date" db:"due_date"`
	CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
	Deliverables    []Deliverable          `json:"deliverables" db:"-"`
	QualityScore    *float64               `json:"quality_score" db:"quality_score"`
	FeedbackNotes   string                 `json:"feedback_notes" db:"feedback_notes"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// Deliverable represents work outputs and milestones
type Deliverable struct {
	ID             uuid.UUID              `json:"id" db:"deliverable_id"`
	AssignmentID   uuid.UUID              `json:"assignment_id" db:"assignment_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Type           string                 `json:"type" db:"type"`
	Status         string                 `json:"status" db:"status"`
	FileURL        string                 `json:"file_url,omitempty" db:"file_url"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	SubmittedAt    *time.Time             `json:"submitted_at" db:"submitted_at"`
	AcceptedAt     *time.Time             `json:"accepted_at" db:"accepted_at"`
	RejectedAt     *time.Time             `json:"rejected_at" db:"rejected_at"`
	RejectionReason string                `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// PerformanceMetric represents a performance metric data point
type PerformanceMetric struct {
	ID          uuid.UUID              `json:"id" db:"metric_id"`
	TalentID    uuid.UUID              `json:"talent_id" db:"talent_id"`
	Type        string                 `json:"type" db:"type"`
	Value       float64                `json:"value" db:"value"`
	Unit        string                 `json:"unit" db:"unit"`
	Description string                 `json:"description" db:"description"`
	Source      string                 `json:"source" db:"source"`
	Context     map[string]interface{} `json:"context" db:"context"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// PerformanceGoal represents a performance goal
type PerformanceGoal struct {
	ID           uuid.UUID              `json:"id" db:"goal_id"`
	TalentID     uuid.UUID              `json:"talent_id" db:"talent_id"`
	Title        string                 `json:"title" db:"title"`
	Description  string                 `json:"description" db:"description"`
	Type         string                 `json:"type" db:"type"`
	TargetValue  float64                `json:"target_value" db:"target_value"`
	CurrentValue float64                `json:"current_value" db:"current_value"`
	Unit         string                 `json:"unit" db:"unit"`
	Priority     Priority               `json:"priority" db:"priority"`
	DueDate      time.Time              `json:"due_date" db:"due_date"`
	Status       string                 `json:"status" db:"status"`
	Metrics      map[string]interface{} `json:"metrics" db:"metrics"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// PerformanceReview represents a performance evaluation
type PerformanceReview struct {
	ID               uuid.UUID              `json:"id" db:"review_id"`
	TalentID         uuid.UUID              `json:"talent_id" db:"talent_id"`
	EngagementID     *uuid.UUID             `json:"engagement_id" db:"engagement_id"`
	ReviewerID       *uuid.UUID             `json:"reviewer_id" db:"reviewer_id"`
	ReviewPeriodStart time.Time             `json:"review_period_start" db:"review_period_start"`
	ReviewPeriodEnd   time.Time             `json:"review_period_end" db:"review_period_end"`
	OverallRating    PerformanceRating      `json:"overall_rating" db:"overall_rating"`
	QualityScore     float64                `json:"quality_score" db:"quality_score"`
	ProductivityScore float64               `json:"productivity_score" db:"productivity_score"`
	ReliabilityScore float64                `json:"reliability_score" db:"reliability_score"`
	CommunicationScore float64              `json:"communication_score" db:"communication_score"`
	Strengths        []string               `json:"strengths" db:"-"`
	ImprovementAreas []string               `json:"improvement_areas" db:"-"`
	Goals            []string               `json:"goals" db:"-"`
	Comments         string                 `json:"comments" db:"comments"`
	Metrics          map[string]interface{} `json:"metrics" db:"metrics"`
	CompensationAdjustment *Money           `json:"compensation_adjustment" db:"-"`
	CompensationAdjustmentAmount *int64     `json:"compensation_adjustment_amount" db:"compensation_adjustment_amount"`
	CompensationAdjustmentCurrency *string  `json:"compensation_adjustment_currency" db:"compensation_adjustment_currency"`
	NextReviewDate   *time.Time             `json:"next_review_date" db:"next_review_date"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// CompensationPlan represents payment structures and terms
type CompensationPlan struct {
	ID                uuid.UUID              `json:"id" db:"compensation_plan_id"`
	TalentID          uuid.UUID              `json:"talent_id" db:"talent_id"`
	EngagementID      *uuid.UUID             `json:"engagement_id" db:"engagement_id"`
	Type              string                 `json:"type" db:"type"` // Salary, Hourly, Project, Retainer
	BaseAmount        *Money                 `json:"base_amount" db:"-"`
	Currency          string                 `json:"currency" db:"currency"`
	PaymentFrequency  string                 `json:"payment_frequency" db:"payment_frequency"`
	BonusStructure    map[string]interface{} `json:"bonus_structure" db:"bonus_structure"`
	Benefits          []string               `json:"benefits" db:"-"`
	EffectiveDate     time.Time              `json:"effective_date" db:"effective_date"`
	EndDate           *time.Time             `json:"end_date" db:"end_date"`
	TaxWithholding    float64                `json:"tax_withholding" db:"tax_withholding"`
	PaymentMethodID   *uuid.UUID             `json:"payment_method_id" db:"payment_method_id"`
	SmartContractAddr string                 `json:"smart_contract_addr,omitempty" db:"smart_contract_addr"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// PayrollRecord represents payment history
type PayrollRecord struct {
	ID               uuid.UUID              `json:"id" db:"payroll_id"`
	TalentID         uuid.UUID              `json:"talent_id" db:"talent_id"`
	EngagementID     *uuid.UUID             `json:"engagement_id" db:"engagement_id"`
	PayPeriodStart   time.Time              `json:"pay_period_start" db:"pay_period_start"`
	PayPeriodEnd     time.Time              `json:"pay_period_end" db:"pay_period_end"`
	GrossAmount      *Money                 `json:"gross_amount" db:"-"`
	NetAmount        *Money                 `json:"net_amount" db:"-"`
	Currency         string                 `json:"currency" db:"currency"`
	HoursWorked      float64                `json:"hours_worked" db:"hours_worked"`
	Deductions       map[string]interface{} `json:"deductions" db:"deductions"`
	Bonuses          map[string]interface{} `json:"bonuses" db:"bonuses"`
	PaymentDate      time.Time              `json:"payment_date" db:"payment_date"`
	PaymentMethod    string                 `json:"payment_method" db:"payment_method"`
	TransactionID    string                 `json:"transaction_id" db:"transaction_id"`
	Status           string                 `json:"status" db:"status"`
	TaxDocumentURLs  []string               `json:"tax_document_urls" db:"-"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// TrainingProgram represents onboarding and skill development
type TrainingProgram struct {
	ID               uuid.UUID              `json:"id" db:"training_id"`
	Name             string                 `json:"name" db:"name"`
	Description      string                 `json:"description" db:"description"`
	Type             string                 `json:"type" db:"type"` // Onboarding, Skill, Compliance
	TargetAudience   string                 `json:"target_audience" db:"target_audience"`
	Duration         int                    `json:"duration" db:"duration"` // in hours
	Format           string                 `json:"format" db:"format"` // Online, InPerson, Hybrid
	Materials        []TrainingMaterial     `json:"materials" db:"-"`
	Prerequisites    []string               `json:"prerequisites" db:"-"`
	LearningObjectives []string             `json:"learning_objectives" db:"-"`
	PassingScore     float64                `json:"passing_score" db:"passing_score"`
	CertificationID  *uuid.UUID             `json:"certification_id" db:"certification_id"`
	IsActive         bool                   `json:"is_active" db:"is_active"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// TrainingMaterial represents training content
type TrainingMaterial struct {
	ID          uuid.UUID              `json:"id" db:"material_id"`
	TrainingID  uuid.UUID              `json:"training_id" db:"training_id"`
	Title       string                 `json:"title" db:"title"`
	Type        string                 `json:"type" db:"type"` // Video, Document, Quiz, Exercise
	ContentURL  string                 `json:"content_url" db:"content_url"`
	Duration    int                    `json:"duration" db:"duration"` // in minutes
	Order       int                    `json:"order" db:"order"`
	IsRequired  bool                   `json:"is_required" db:"is_required"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// TrainingProgress represents talent's progress in training
type TrainingProgress struct {
	ID              uuid.UUID              `json:"id" db:"progress_id"`
	TalentID        uuid.UUID              `json:"talent_id" db:"talent_id"`
	TrainingID      uuid.UUID              `json:"training_id" db:"training_id"`
	Status          TrainingStatus         `json:"status" db:"status"`
	StartedAt       *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
	Progress        float64                `json:"progress" db:"progress"` // 0-100
	Score           *float64               `json:"score" db:"score"`
	Attempts        int                    `json:"attempts" db:"attempts"`
	MaterialProgress map[string]interface{} `json:"material_progress" db:"material_progress"`
	CertificateURL  string                 `json:"certificate_url,omitempty" db:"certificate_url"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// TalentApplication represents a job application
type TalentApplication struct {
	ID              uuid.UUID              `json:"id" db:"application_id"`
	TalentID        uuid.UUID              `json:"talent_id" db:"talent_id"`
	JobPostingID    uuid.UUID              `json:"job_posting_id" db:"job_posting_id"`
	Status          ApplicationStatus      `json:"status" db:"status"`
	CoverLetter     string                 `json:"cover_letter" db:"cover_letter"`
	ResumeURL       string                 `json:"resume_url,omitempty" db:"resume_url"`
	PortfolioURLs   []string               `json:"portfolio_urls" db:"-"`
	ScreeningScore  *float64               `json:"screening_score" db:"screening_score"`
	ScreeningNotes  string                 `json:"screening_notes" db:"screening_notes"`
	InterviewNotes  string                 `json:"interview_notes" db:"interview_notes"`
	AssessmentScore *float64               `json:"assessment_score" db:"assessment_score"`
	ReferenceChecks map[string]interface{} `json:"reference_checks" db:"reference_checks"`
	DecisionDate    *time.Time             `json:"decision_date" db:"decision_date"`
	DecisionReason  string                 `json:"decision_reason" db:"decision_reason"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// JobPosting represents an open position or requirement
type JobPosting struct {
	ID               uuid.UUID              `json:"id" db:"job_posting_id"`
	Title            string                 `json:"title" db:"title"`
	Description      string                 `json:"description" db:"description"`
	Type             EngagementType         `json:"type" db:"type"`
	Department       string                 `json:"department" db:"department"`
	RequiredSkills   []string               `json:"required_skills" db:"-"`
	PreferredSkills  []string               `json:"preferred_skills" db:"-"`
	ExperienceYears  int                    `json:"experience_years" db:"experience_years"`
	EducationLevel   string                 `json:"education_level" db:"education_level"`
	Location         string                 `json:"location" db:"location"`
	Remote           bool                   `json:"remote" db:"remote"`
	SalaryRange      map[string]interface{} `json:"salary_range" db:"salary_range"`
	Benefits         []string               `json:"benefits" db:"-"`
	PostedDate       time.Time              `json:"posted_date" db:"posted_date"`
	ClosingDate      *time.Time             `json:"closing_date" db:"closing_date"`
	IsActive         bool                   `json:"is_active" db:"is_active"`
	ApplicationCount int                    `json:"application_count" db:"application_count"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// ContractorAgreement represents legal agreements for contractors
type ContractorAgreement struct {
	ID               uuid.UUID              `json:"id" db:"agreement_id"`
	TalentID         uuid.UUID              `json:"talent_id" db:"talent_id"`
	EngagementID     *uuid.UUID             `json:"engagement_id" db:"engagement_id"`
	ContractType     string                 `json:"contract_type" db:"contract_type"`
	TemplateID       uuid.UUID              `json:"template_id" db:"template_id"`
	Terms            map[string]interface{} `json:"terms" db:"terms"`
	StartDate        time.Time              `json:"start_date" db:"start_date"`
	EndDate          *time.Time             `json:"end_date" db:"end_date"`
	RenewalDate      *time.Time             `json:"renewal_date" db:"renewal_date"`
	SignedAt         *time.Time             `json:"signed_at" db:"signed_at"`
	SignatureID      string                 `json:"signature_id" db:"signature_id"`
	DocumentURL      string                 `json:"document_url" db:"document_url"`
	Status           string                 `json:"status" db:"status"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// ComplianceCheck represents background checks and verifications
type ComplianceCheck struct {
	ID              uuid.UUID              `json:"id" db:"check_id"`
	TalentID        uuid.UUID              `json:"talent_id" db:"talent_id"`
	CheckType       string                 `json:"check_type" db:"check_type"`
	Provider        string                 `json:"provider" db:"provider"`
	Status          string                 `json:"status" db:"status"`
	Result          string                 `json:"result" db:"result"`
	Details         map[string]interface{} `json:"details" db:"details"`
	DocumentURLs    []string               `json:"document_urls" db:"-"`
	ValidUntil      *time.Time             `json:"valid_until" db:"valid_until"`
	CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// OffboardingChecklist represents tasks for offboarding
type OffboardingChecklist struct {
	ID               uuid.UUID              `json:"id" db:"checklist_id"`
	TalentID         uuid.UUID              `json:"talent_id" db:"talent_id"`
	EngagementID     *uuid.UUID             `json:"engagement_id" db:"engagement_id"`
	Reason           string                 `json:"reason" db:"reason"`
	LastWorkingDate  time.Time              `json:"last_working_date" db:"last_working_date"`
	Tasks            []OffboardingTask      `json:"tasks" db:"-"`
	KnowledgeTransfer map[string]interface{} `json:"knowledge_transfer" db:"knowledge_transfer"`
	ExitInterviewURL string                 `json:"exit_interview_url,omitempty" db:"exit_interview_url"`
	CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// OffboardingTask represents a single offboarding task
type OffboardingTask struct {
	ID           uuid.UUID  `json:"id" db:"task_id"`
	ChecklistID  uuid.UUID  `json:"checklist_id" db:"checklist_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	AssignedTo   *uuid.UUID `json:"assigned_to" db:"assigned_to"`
	DueDate      time.Time  `json:"due_date" db:"due_date"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	CompletedBy  *uuid.UUID `json:"completed_by" db:"completed_by"`
	Notes        string     `json:"notes" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}
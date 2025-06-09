package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// TalentRepository defines the interface for talent data access
type TalentRepository interface {
	// Basic CRUD operations
	CreateTalent(ctx context.Context, talent *entities.Talent) error
	GetTalentByID(ctx context.Context, id uuid.UUID) (*entities.Talent, error)
	GetTalentByEmail(ctx context.Context, email string) (*entities.Talent, error)
	UpdateTalent(ctx context.Context, talent *entities.Talent) error
	DeleteTalent(ctx context.Context, id uuid.UUID) error
	
	// Search and filtering
	SearchTalent(ctx context.Context, filter TalentFilter) ([]*entities.Talent, int, error)
	GetTalentBySkills(ctx context.Context, skills []string, minLevel entities.SkillLevel) ([]*entities.Talent, error)
	GetAvailableTalent(ctx context.Context, talentType entities.TalentType) ([]*entities.Talent, error)
	
	// Reputation and scoring
	UpdateReputationScore(ctx context.Context, talentID uuid.UUID, score float64) error
	GetTopTalentByScore(ctx context.Context, limit int) ([]*entities.Talent, error)
	
	// Skills and certifications
	AddTalentSkill(ctx context.Context, skill *entities.Skill) error
	UpdateTalentSkill(ctx context.Context, skill *entities.Skill) error
	RemoveTalentSkill(ctx context.Context, talentID, skillID uuid.UUID) error
	GetTalentSkills(ctx context.Context, talentID uuid.UUID) ([]*entities.Skill, error)
	
	AddTalentCertification(ctx context.Context, cert *entities.Certification) error
	UpdateTalentCertification(ctx context.Context, cert *entities.Certification) error
	GetTalentCertifications(ctx context.Context, talentID uuid.UUID) ([]*entities.Certification, error)
	GetExpiringCertifications(ctx context.Context, beforeDate time.Time) ([]*entities.Certification, error)
}

// EngagementRepository defines the interface for engagement data access
type EngagementRepository interface {
	// Basic CRUD operations
	CreateEngagement(ctx context.Context, engagement *entities.Engagement) error
	GetEngagementByID(ctx context.Context, id uuid.UUID) (*entities.Engagement, error)
	UpdateEngagement(ctx context.Context, engagement *entities.Engagement) error
	DeleteEngagement(ctx context.Context, id uuid.UUID) error
	
	// Search and filtering
	ListEngagements(ctx context.Context, filter EngagementFilter) ([]*entities.Engagement, int, error)
	GetEngagementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.Engagement, error)
	GetActiveEngagements(ctx context.Context) ([]*entities.Engagement, error)
	GetEngagementsEndingSoon(ctx context.Context, beforeDate time.Time) ([]*entities.Engagement, error)
	
	// Performance metrics
	UpdateEngagementMetrics(ctx context.Context, engagementID uuid.UUID, metrics map[string]interface{}) error
	GetEngagementMetrics(ctx context.Context, engagementID uuid.UUID) (map[string]interface{}, error)
}

// WorkAssignmentRepository defines the interface for work assignment data access
type WorkAssignmentRepository interface {
	// Basic CRUD operations
	CreateAssignment(ctx context.Context, assignment *entities.WorkAssignment) error
	GetAssignmentByID(ctx context.Context, id uuid.UUID) (*entities.WorkAssignment, error)
	UpdateAssignment(ctx context.Context, assignment *entities.WorkAssignment) error
	DeleteAssignment(ctx context.Context, id uuid.UUID) error
	
	// Search and filtering
	ListAssignments(ctx context.Context, filter AssignmentFilter) ([]*entities.WorkAssignment, int, error)
	GetAssignmentsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.WorkAssignment, error)
	GetAssignmentsByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*entities.WorkAssignment, error)
	GetOverdueAssignments(ctx context.Context) ([]*entities.WorkAssignment, error)
	
	// Deliverables
	CreateDeliverable(ctx context.Context, deliverable *entities.Deliverable) error
	GetDeliverableByID(ctx context.Context, id uuid.UUID) (*entities.Deliverable, error)
	UpdateDeliverable(ctx context.Context, deliverable *entities.Deliverable) error
	GetDeliverablesByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*entities.Deliverable, error)
	GetPendingDeliverables(ctx context.Context) ([]*entities.Deliverable, error)
}

// PerformanceRepository defines the interface for performance review data access
type PerformanceRepository interface {
	// Performance reviews
	CreatePerformanceReview(ctx context.Context, review *entities.PerformanceReview) error
	GetPerformanceReviewByID(ctx context.Context, id uuid.UUID) (*entities.PerformanceReview, error)
	UpdatePerformanceReview(ctx context.Context, review *entities.PerformanceReview) error
	
	// Search and filtering
	GetPerformanceReviewsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.PerformanceReview, error)
	GetPerformanceReviewsByPeriod(ctx context.Context, start, end time.Time) ([]*entities.PerformanceReview, error)
	GetReviewsDue(ctx context.Context, beforeDate time.Time) ([]*entities.PerformanceReview, error)
	
	// Performance analytics
	GetTalentPerformanceMetrics(ctx context.Context, talentID uuid.UUID, timeRange TimeRange) (*TalentPerformanceMetrics, error)
	GetPerformanceDistribution(ctx context.Context, timeRange TimeRange) (*PerformanceDistribution, error)
	GetTopPerformers(ctx context.Context, metric string, limit int) ([]*entities.Talent, error)
	GetUnderperformers(ctx context.Context, threshold float64) ([]*entities.Talent, error)
}

// CompensationRepository defines the interface for compensation data access
type CompensationRepository interface {
	// Compensation plans
	CreateCompensationPlan(ctx context.Context, plan *entities.CompensationPlan) error
	GetCompensationPlanByID(ctx context.Context, id uuid.UUID) (*entities.CompensationPlan, error)
	UpdateCompensationPlan(ctx context.Context, plan *entities.CompensationPlan) error
	GetCompensationPlansByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.CompensationPlan, error)
	GetActiveCompensationPlan(ctx context.Context, talentID uuid.UUID) (*entities.CompensationPlan, error)
	
	// Payroll records
	CreatePayrollRecord(ctx context.Context, record *entities.PayrollRecord) error
	GetPayrollRecordByID(ctx context.Context, id uuid.UUID) (*entities.PayrollRecord, error)
	UpdatePayrollRecord(ctx context.Context, record *entities.PayrollRecord) error
	GetPayrollRecordsByTalent(ctx context.Context, talentID uuid.UUID, timeRange TimeRange) ([]*entities.PayrollRecord, error)
	GetPendingPayroll(ctx context.Context) ([]*entities.PayrollRecord, error)
	
	// Compensation analytics
	GetCompensationSummary(ctx context.Context, timeRange TimeRange) (*CompensationSummary, error)
	GetTalentCompensationHistory(ctx context.Context, talentID uuid.UUID) ([]*entities.PayrollRecord, error)
	GetCompensationBenchmarks(ctx context.Context, skillSet []string) (*CompensationBenchmarks, error)
}

// TrainingRepository defines the interface for training data access
type TrainingRepository interface {
	// Training programs
	CreateTrainingProgram(ctx context.Context, program *entities.TrainingProgram) error
	GetTrainingProgramByID(ctx context.Context, id uuid.UUID) (*entities.TrainingProgram, error)
	UpdateTrainingProgram(ctx context.Context, program *entities.TrainingProgram) error
	ListTrainingPrograms(ctx context.Context, filter TrainingFilter) ([]*entities.TrainingProgram, int, error)
	GetRequiredTraining(ctx context.Context, talentType entities.TalentType) ([]*entities.TrainingProgram, error)
	
	// Training materials
	CreateTrainingMaterial(ctx context.Context, material *entities.TrainingMaterial) error
	GetTrainingMaterials(ctx context.Context, trainingID uuid.UUID) ([]*entities.TrainingMaterial, error)
	UpdateTrainingMaterial(ctx context.Context, material *entities.TrainingMaterial) error
	
	// Training progress
	CreateTrainingProgress(ctx context.Context, progress *entities.TrainingProgress) error
	GetTrainingProgress(ctx context.Context, talentID, trainingID uuid.UUID) (*entities.TrainingProgress, error)
	UpdateTrainingProgress(ctx context.Context, progress *entities.TrainingProgress) error
	GetTrainingProgressByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error)
	GetIncompleteTraining(ctx context.Context) ([]*entities.TrainingProgress, error)
	
	// Training analytics
	GetTrainingCompletionRates(ctx context.Context, timeRange TimeRange) (*TrainingCompletionRates, error)
	GetTrainingEffectiveness(ctx context.Context, trainingID uuid.UUID) (*TrainingEffectiveness, error)
}

// TalentApplicationRepository defines the interface for application data access
type TalentApplicationRepository interface {
	// Application management
	CreateApplication(ctx context.Context, application *entities.TalentApplication) error
	GetApplicationByID(ctx context.Context, id uuid.UUID) (*entities.TalentApplication, error)
	UpdateApplication(ctx context.Context, application *entities.TalentApplication) error
	ListApplications(ctx context.Context, filter ApplicationFilter) ([]*entities.TalentApplication, int, error)
	
	// Job postings
	CreateJobPosting(ctx context.Context, posting *entities.JobPosting) error
	GetJobPostingByID(ctx context.Context, id uuid.UUID) (*entities.JobPosting, error)
	UpdateJobPosting(ctx context.Context, posting *entities.JobPosting) error
	ListJobPostings(ctx context.Context, filter JobPostingFilter) ([]*entities.JobPosting, int, error)
	GetActiveJobPostings(ctx context.Context) ([]*entities.JobPosting, error)
	
	// Application analytics
	GetApplicationMetrics(ctx context.Context, timeRange TimeRange) (*ApplicationMetrics, error)
	GetJobPostingPerformance(ctx context.Context, jobPostingID uuid.UUID) (*JobPostingMetrics, error)
}

// ComplianceRepository defines the interface for compliance data access
type ComplianceRepository interface {
	// Compliance checks
	CreateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error
	GetComplianceCheckByID(ctx context.Context, id uuid.UUID) (*entities.ComplianceCheck, error)
	UpdateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error
	GetComplianceChecksByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.ComplianceCheck, error)
	GetPendingComplianceChecks(ctx context.Context) ([]*entities.ComplianceCheck, error)
	GetExpiringCompliance(ctx context.Context, beforeDate time.Time) ([]*entities.ComplianceCheck, error)
	
	// Contractor agreements
	CreateContractorAgreement(ctx context.Context, agreement *entities.ContractorAgreement) error
	GetContractorAgreementByID(ctx context.Context, id uuid.UUID) (*entities.ContractorAgreement, error)
	UpdateContractorAgreement(ctx context.Context, agreement *entities.ContractorAgreement) error
	GetContractorAgreementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.ContractorAgreement, error)
	GetExpiringAgreements(ctx context.Context, beforeDate time.Time) ([]*entities.ContractorAgreement, error)
	
	// Compliance reporting
	GetComplianceReport(ctx context.Context, timeRange TimeRange) (*ComplianceReport, error)
	GetTalentComplianceStatus(ctx context.Context, talentID uuid.UUID) (*TalentComplianceStatus, error)
}

// OffboardingRepository defines the interface for offboarding data access
type OffboardingRepository interface {
	// Offboarding checklists
	CreateOffboardingChecklist(ctx context.Context, checklist *entities.OffboardingChecklist) error
	GetOffboardingChecklistByID(ctx context.Context, id uuid.UUID) (*entities.OffboardingChecklist, error)
	UpdateOffboardingChecklist(ctx context.Context, checklist *entities.OffboardingChecklist) error
	GetOffboardingChecklistByTalent(ctx context.Context, talentID uuid.UUID) (*entities.OffboardingChecklist, error)
	GetPendingOffboarding(ctx context.Context) ([]*entities.OffboardingChecklist, error)
	
	// Offboarding tasks
	CreateOffboardingTask(ctx context.Context, task *entities.OffboardingTask) error
	GetOffboardingTaskByID(ctx context.Context, id uuid.UUID) (*entities.OffboardingTask, error)
	UpdateOffboardingTask(ctx context.Context, task *entities.OffboardingTask) error
	GetOffboardingTasksByChecklist(ctx context.Context, checklistID uuid.UUID) ([]*entities.OffboardingTask, error)
	GetOverdueOffboardingTasks(ctx context.Context) ([]*entities.OffboardingTask, error)
}

// Filter and query structures

// TalentFilter represents filtering options for talent queries
type TalentFilter struct {
	Type            *entities.TalentType    `json:"type,omitempty"`
	Status          *entities.TalentStatus  `json:"status,omitempty"`
	Skills          []string                `json:"skills,omitempty"`
	MinReputation   *float64                `json:"min_reputation,omitempty"`
	Location        *string                 `json:"location,omitempty"`
	Remote          *bool                   `json:"remote,omitempty"`
	MinHourlyRate   *float64                `json:"min_hourly_rate,omitempty"`
	MaxHourlyRate   *float64                `json:"max_hourly_rate,omitempty"`
	Search          string                  `json:"search,omitempty"`
	Offset          int                     `json:"offset"`
	Limit           int                     `json:"limit"`
	SortBy          string                  `json:"sort_by"`
	SortOrder       string                  `json:"sort_order"`
}

// EngagementFilter represents filtering options for engagement queries
type EngagementFilter struct {
	TalentID        *uuid.UUID                  `json:"talent_id,omitempty"`
	Type            *entities.EngagementType    `json:"type,omitempty"`
	Status          *entities.EngagementStatus  `json:"status,omitempty"`
	StartDateAfter  *time.Time                  `json:"start_date_after,omitempty"`
	StartDateBefore *time.Time                  `json:"start_date_before,omitempty"`
	EndDateAfter    *time.Time                  `json:"end_date_after,omitempty"`
	EndDateBefore   *time.Time                  `json:"end_date_before,omitempty"`
	Offset          int                         `json:"offset"`
	Limit           int                         `json:"limit"`
	SortBy          string                      `json:"sort_by"`
	SortOrder       string                      `json:"sort_order"`
}

// AssignmentFilter represents filtering options for assignment queries
type AssignmentFilter struct {
	TalentID        *uuid.UUID      `json:"talent_id,omitempty"`
	EngagementID    *uuid.UUID      `json:"engagement_id,omitempty"`
	ProjectID       *uuid.UUID      `json:"project_id,omitempty"`
	Status          *string         `json:"status,omitempty"`
	Priority        *entities.Priority `json:"priority,omitempty"`
	DueDateAfter    *time.Time      `json:"due_date_after,omitempty"`
	DueDateBefore   *time.Time      `json:"due_date_before,omitempty"`
	Offset          int             `json:"offset"`
	Limit           int             `json:"limit"`
	SortBy          string          `json:"sort_by"`
	SortOrder       string          `json:"sort_order"`
}

// TrainingFilter represents filtering options for training queries
type TrainingFilter struct {
	Type            *string                 `json:"type,omitempty"`
	TargetAudience  *string                 `json:"target_audience,omitempty"`
	IsActive        *bool                   `json:"is_active,omitempty"`
	Offset          int                     `json:"offset"`
	Limit           int                     `json:"limit"`
	SortBy          string                  `json:"sort_by"`
	SortOrder       string                  `json:"sort_order"`
}

// ApplicationFilter represents filtering options for application queries
type ApplicationFilter struct {
	TalentID        *uuid.UUID                      `json:"talent_id,omitempty"`
	JobPostingID    *uuid.UUID                      `json:"job_posting_id,omitempty"`
	Status          *entities.ApplicationStatus     `json:"status,omitempty"`
	MinScore        *float64                        `json:"min_score,omitempty"`
	CreatedAfter    *time.Time                      `json:"created_after,omitempty"`
	CreatedBefore   *time.Time                      `json:"created_before,omitempty"`
	Offset          int                             `json:"offset"`
	Limit           int                             `json:"limit"`
	SortBy          string                          `json:"sort_by"`
	SortOrder       string                          `json:"sort_order"`
}

// JobPostingFilter represents filtering options for job posting queries
type JobPostingFilter struct {
	Type            *entities.EngagementType `json:"type,omitempty"`
	Department      *string                  `json:"department,omitempty"`
	Location        *string                  `json:"location,omitempty"`
	Remote          *bool                    `json:"remote,omitempty"`
	IsActive        *bool                    `json:"is_active,omitempty"`
	Skills          []string                 `json:"skills,omitempty"`
	MinExperience   *int                     `json:"min_experience,omitempty"`
	Offset          int                      `json:"offset"`
	Limit           int                      `json:"limit"`
	SortBy          string                   `json:"sort_by"`
	SortOrder       string                   `json:"sort_order"`
}

// Analytics and reporting structures

// TalentPerformanceMetrics represents talent performance analytics
type TalentPerformanceMetrics struct {
	TalentID            uuid.UUID             `json:"talent_id"`
	OverallRating       entities.PerformanceRating `json:"overall_rating"`
	AverageQualityScore float64               `json:"average_quality_score"`
	AverageProductivityScore float64          `json:"average_productivity_score"`
	AverageReliabilityScore float64           `json:"average_reliability_score"`
	ProjectsCompleted   int                   `json:"projects_completed"`
	OnTimeDeliveryRate  float64               `json:"on_time_delivery_rate"`
	ClientSatisfactionScore float64           `json:"client_satisfaction_score"`
	SkillGrowthRate     float64               `json:"skill_growth_rate"`
	TrainingCompletion  float64               `json:"training_completion"`
}

// PerformanceDistribution represents performance distribution analytics
type PerformanceDistribution struct {
	Exceptional     int     `json:"exceptional"`
	ExceedsExpectations int `json:"exceeds_expectations"`
	MeetsExpectations   int `json:"meets_expectations"`
	NeedsImprovement    int `json:"needs_improvement"`
	Unsatisfactory      int `json:"unsatisfactory"`
	AverageScore        float64 `json:"average_score"`
	TotalReviews        int     `json:"total_reviews"`
}

// CompensationSummary represents compensation analytics
type CompensationSummary struct {
	TotalPayroll        *entities.Money       `json:"total_payroll"`
	AverageHourlyRate   *entities.Money       `json:"average_hourly_rate"`
	MedianHourlyRate    *entities.Money       `json:"median_hourly_rate"`
	TotalHoursWorked    float64               `json:"total_hours_worked"`
	PayrollByType       map[string]*entities.Money `json:"payroll_by_type"`
	TopEarners          []*entities.Talent    `json:"top_earners"`
	PayrollGrowthRate   float64               `json:"payroll_growth_rate"`
}

// CompensationBenchmarks represents market compensation benchmarks
type CompensationBenchmarks struct {
	SkillSet            []string              `json:"skill_set"`
	MinRate             *entities.Money       `json:"min_rate"`
	MaxRate             *entities.Money       `json:"max_rate"`
	AverageRate         *entities.Money       `json:"average_rate"`
	MedianRate          *entities.Money       `json:"median_rate"`
	MarketPercentile25  *entities.Money       `json:"market_percentile_25"`
	MarketPercentile75  *entities.Money       `json:"market_percentile_75"`
	SampleSize          int                   `json:"sample_size"`
	LastUpdated         time.Time             `json:"last_updated"`
}

// TrainingCompletionRates represents training completion analytics
type TrainingCompletionRates struct {
	TotalPrograms       int                   `json:"total_programs"`
	CompletedPrograms   int                   `json:"completed_programs"`
	InProgressPrograms  int                   `json:"in_progress_programs"`
	CompletionRate      float64               `json:"completion_rate"`
	AverageScore        float64               `json:"average_score"`
	AverageCompletionTime float64             `json:"average_completion_time"`
	CompletionByType    map[string]int        `json:"completion_by_type"`
}

// TrainingEffectiveness represents training effectiveness metrics
type TrainingEffectiveness struct {
	TrainingID          uuid.UUID             `json:"training_id"`
	EnrollmentCount     int                   `json:"enrollment_count"`
	CompletionCount     int                   `json:"completion_count"`
	CompletionRate      float64               `json:"completion_rate"`
	AverageScore        float64               `json:"average_score"`
	PassRate            float64               `json:"pass_rate"`
	AverageRating       float64               `json:"average_rating"`
	PerformanceImprovement float64            `json:"performance_improvement"`
	ROI                 float64               `json:"roi"`
}

// ApplicationMetrics represents application analytics
type ApplicationMetrics struct {
	TotalApplications   int                   `json:"total_applications"`
	ApprovedApplications int                  `json:"approved_applications"`
	RejectedApplications int                  `json:"rejected_applications"`
	PendingApplications int                   `json:"pending_applications"`
	ApprovalRate        float64               `json:"approval_rate"`
	AverageProcessingTime float64             `json:"average_processing_time"`
	ApplicationsBySource map[string]int       `json:"applications_by_source"`
	TopSkillsRequested  []string              `json:"top_skills_requested"`
}

// JobPostingMetrics represents job posting performance metrics
type JobPostingMetrics struct {
	JobPostingID        uuid.UUID             `json:"job_posting_id"`
	ViewCount           int                   `json:"view_count"`
	ApplicationCount    int                   `json:"application_count"`
	QualifiedCount      int                   `json:"qualified_count"`
	InterviewCount      int                   `json:"interview_count"`
	HireCount           int                   `json:"hire_count"`
	ConversionRate      float64               `json:"conversion_rate"`
	TimeToFill          float64               `json:"time_to_fill"`
	CostPerHire         *entities.Money       `json:"cost_per_hire"`
}

// ComplianceReport represents compliance status and analytics
type ComplianceReport struct {
	TotalTalent         int                   `json:"total_talent"`
	CompliantTalent     int                   `json:"compliant_talent"`
	NonCompliantTalent  int                   `json:"non_compliant_talent"`
	ComplianceRate      float64               `json:"compliance_rate"`
	ExpiringChecks      int                   `json:"expiring_checks"`
	OverdueChecks       int                   `json:"overdue_checks"`
	ComplianceByType    map[string]int        `json:"compliance_by_type"`
	RiskLevel           string                `json:"risk_level"`
}

// TalentComplianceStatus represents individual talent compliance status
type TalentComplianceStatus struct {
	TalentID            uuid.UUID             `json:"talent_id"`
	IsCompliant         bool                  `json:"is_compliant"`
	ComplianceScore     float64               `json:"compliance_score"`
	ActiveChecks        []*entities.ComplianceCheck `json:"active_checks"`
	ExpiringChecks      []*entities.ComplianceCheck `json:"expiring_checks"`
	MissingChecks       []string              `json:"missing_checks"`
	LastCheckDate       *time.Time            `json:"last_check_date"`
	NextCheckDue        *time.Time            `json:"next_check_due"`
	RiskLevel           string                `json:"risk_level"`
}
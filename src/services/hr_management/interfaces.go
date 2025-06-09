package hr_management

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// TalentAcquisitionService handles talent sourcing and recruitment
type TalentAcquisitionService interface {
	// Job posting management
	CreateJobPosting(ctx context.Context, request JobPostingRequest) (*entities.JobPosting, error)
	UpdateJobPosting(ctx context.Context, id uuid.UUID, request JobPostingRequest) (*entities.JobPosting, error)
	GetJobPosting(ctx context.Context, id uuid.UUID) (*entities.JobPosting, error)
	ListJobPostings(ctx context.Context, filter repositories.JobPostingFilter) ([]*entities.JobPosting, int, error)
	CloseJobPosting(ctx context.Context, id uuid.UUID) error

	// Application processing
	SubmitApplication(ctx context.Context, request ApplicationRequest) (*entities.TalentApplication, error)
	ScreenApplication(ctx context.Context, applicationID uuid.UUID) (*ApplicationScreeningResult, error)
	ProcessApplication(ctx context.Context, applicationID uuid.UUID, decision ApplicationDecision) error
	GetApplication(ctx context.Context, id uuid.UUID) (*entities.TalentApplication, error)
	ListApplications(ctx context.Context, filter repositories.ApplicationFilter) ([]*entities.TalentApplication, int, error)

	// Talent sourcing
	SourceTalent(ctx context.Context, requirements TalentRequirements) ([]*TalentMatch, error)
	SearchTalent(ctx context.Context, criteria TalentSearchCriteria) ([]*entities.Talent, error)
	AnalyzeTalentPool(ctx context.Context) (*TalentPoolAnalysis, error)

	// AI agent discovery
	DiscoverAIAgents(ctx context.Context, capabilities []string) ([]*entities.AIAgent, error)
	TestAIAgentCapabilities(ctx context.Context, agentID uuid.UUID, testSuite []CapabilityTest) (*CapabilityTestResult, error)
	RegisterAIAgent(ctx context.Context, request AIAgentRegistrationRequest) (*entities.AIAgent, error)
}

// OnboardingService handles new talent integration
type OnboardingService interface {
	// Onboarding workflow
	StartOnboarding(ctx context.Context, talentID uuid.UUID, onboardingType OnboardingType) (*OnboardingWorkflow, error)
	ProcessOnboardingStep(ctx context.Context, workflowID uuid.UUID, stepID string, data map[string]interface{}) error
	GetOnboardingStatus(ctx context.Context, talentID uuid.UUID) (*OnboardingStatus, error)
	CompleteOnboarding(ctx context.Context, workflowID uuid.UUID) error

	// Contract management
	GenerateContract(ctx context.Context, talentID uuid.UUID, engagementDetails EngagementDetails) (*ContractRequest, error)
	ProcessContractSigning(ctx context.Context, contractID uuid.UUID, signature SignatureData) error
	GetContractStatus(ctx context.Context, contractID uuid.UUID) (*ContractStatus, error)

	// Access and credentials
	ProvisionAccess(ctx context.Context, talentID uuid.UUID, accessRequirements []AccessRequirement) error
	SetupCredentials(ctx context.Context, talentID uuid.UUID, credentialType string) (*CredentialSetup, error)
	RevokeAccess(ctx context.Context, talentID uuid.UUID, accessType string) error

	// Training assignment
	AssignRequiredTraining(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error)
	TrackTrainingProgress(ctx context.Context, talentID uuid.UUID) (*TrainingProgressSummary, error)
}

// PerformanceManagementService handles performance tracking and reviews
type PerformanceManagementService interface {
	// Performance tracking
	RecordPerformanceMetric(ctx context.Context, talentID uuid.UUID, metric PerformanceMetric) error
	GetPerformanceMetrics(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) (*repositories.TalentPerformanceMetrics, error)
	AnalyzePerformanceTrends(ctx context.Context, talentID uuid.UUID) (*PerformanceTrendAnalysis, error)

	// Performance reviews
	SchedulePerformanceReview(ctx context.Context, talentID uuid.UUID, reviewType string) (*entities.PerformanceReview, error)
	ConductPerformanceReview(ctx context.Context, reviewID uuid.UUID, reviewData ReviewData) (*entities.PerformanceReview, error)
	GetPerformanceReviews(ctx context.Context, talentID uuid.UUID) ([]*entities.PerformanceReview, error)

	// Goal management
	SetPerformanceGoals(ctx context.Context, talentID uuid.UUID, goals []PerformanceGoal) error
	TrackGoalProgress(ctx context.Context, talentID uuid.UUID) (*GoalProgressSummary, error)
	UpdateGoalProgress(ctx context.Context, goalID uuid.UUID, progress float64, notes string) error

	// Performance alerts
	DetectPerformanceIssues(ctx context.Context) ([]*PerformanceAlert, error)
	GeneratePerformanceReport(ctx context.Context, timeRange repositories.TimeRange) (*PerformanceReport, error)
}

// CompensationService handles payment and compensation management
type CompensationService interface {
	// Compensation planning
	CreateCompensationPlan(ctx context.Context, request CompensationPlanRequest) (*entities.CompensationPlan, error)
	UpdateCompensationPlan(ctx context.Context, planID uuid.UUID, updates CompensationPlanUpdates) (*entities.CompensationPlan, error)
	GetCompensationPlan(ctx context.Context, talentID uuid.UUID) (*entities.CompensationPlan, error)
	CalculateCompensationAdjustment(ctx context.Context, talentID uuid.UUID, performanceMetrics map[string]float64) (*CompensationAdjustment, error)

	// Payroll processing
	ProcessPayroll(ctx context.Context, payPeriod PayPeriod) ([]*entities.PayrollRecord, error)
	GeneratePayrollRecord(ctx context.Context, talentID uuid.UUID, payPeriod PayPeriod) (*entities.PayrollRecord, error)
	CalculateHours(ctx context.Context, talentID uuid.UUID, startDate, endDate time.Time) (*HoursCalculation, error)
	ProcessPayment(ctx context.Context, payrollID uuid.UUID) (*PaymentResult, error)

	// Tax and compliance
	CalculateTaxWithholding(ctx context.Context, grossAmount *entities.Money, talentID uuid.UUID) (*TaxCalculation, error)
	GenerateTaxDocuments(ctx context.Context, talentID uuid.UUID, taxYear int) ([]*TaxDocument, error)
	ProcessContractorPayment(ctx context.Context, talentID uuid.UUID, amount *entities.Money, description string) (*PaymentResult, error)

	// Compensation analytics
	GetCompensationBenchmarks(ctx context.Context, skillSet []string, location string) (*repositories.CompensationBenchmarks, error)
	AnalyzeCompensationEquity(ctx context.Context) (*CompensationEquityAnalysis, error)
	GetCompensationReport(ctx context.Context, timeRange repositories.TimeRange) (*repositories.CompensationSummary, error)
}

// TrainingService handles training and development
type TrainingService interface {
	// Training program management
	CreateTrainingProgram(ctx context.Context, request TrainingProgramRequest) (*entities.TrainingProgram, error)
	UpdateTrainingProgram(ctx context.Context, programID uuid.UUID, updates TrainingProgramUpdates) (*entities.TrainingProgram, error)
	GetTrainingProgram(ctx context.Context, id uuid.UUID) (*entities.TrainingProgram, error)
	ListTrainingPrograms(ctx context.Context, filter repositories.TrainingFilter) ([]*entities.TrainingProgram, int, error)

	// Training enrollment and progress
	EnrollInTraining(ctx context.Context, talentID, trainingID uuid.UUID) (*entities.TrainingProgress, error)
	UpdateTrainingProgress(ctx context.Context, talentID, trainingID uuid.UUID, progress TrainingProgressUpdate) error
	CompleteTraining(ctx context.Context, talentID, trainingID uuid.UUID, finalScore float64) (*entities.Certification, error)
	GetTrainingProgress(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error)

	// Training content generation
	GenerateTrainingContent(ctx context.Context, topic string, targetAudience string, duration int) (*TrainingContentResult, error)
	CreateCustomTraining(ctx context.Context, talentID uuid.UUID, skillGaps []SkillGap) (*entities.TrainingProgram, error)
	AdaptTrainingContent(ctx context.Context, trainingID uuid.UUID, learnerProfile LearnerProfile) (*AdaptedTrainingContent, error)

	// Skill assessment
	AssessSkills(ctx context.Context, talentID uuid.UUID, skillSet []string) (*SkillAssessmentResult, error)
	IdentifySkillGaps(ctx context.Context, talentID uuid.UUID, targetRole string) ([]*SkillGap, error)
	RecommendTraining(ctx context.Context, talentID uuid.UUID) ([]*TrainingRecommendation, error)
}

// ComplianceManagementService handles HR compliance and legal requirements
type ComplianceManagementService interface {
	// Compliance checking
	CheckCompliance(ctx context.Context, talentID uuid.UUID) (*ComplianceStatus, error)
	UpdateComplianceDocument(ctx context.Context, talentID uuid.UUID, docType string, documentURL string) error
	GenerateComplianceReport(ctx context.Context, timeRange repositories.TimeRange) (*ComplianceReport, error)
}

// OffboardingService handles talent departure and knowledge transfer
type OffboardingService interface {
	// Offboarding workflow
	StartOffboarding(ctx context.Context, talentID uuid.UUID, reason string) (*OffboardingWorkflow, error)
	ProcessOffboardingStep(ctx context.Context, workflowID uuid.UUID, stepID string) error
	RevokeAllAccess(ctx context.Context, talentID uuid.UUID) error
	ConductExitInterview(ctx context.Context, talentID uuid.UUID, interviewData ExitInterviewData) (*ExitInterviewResult, error)
	GenerateOffboardingReport(ctx context.Context, talentID uuid.UUID) (*OffboardingReport, error)
}

// HRAnalyticsService provides HR metrics and reporting
type HRAnalyticsService interface {
	// Analytics
	GenerateTalentAnalytics(ctx context.Context) (*TalentAnalytics, error)
	GeneratePerformanceAnalytics(ctx context.Context, timeRange repositories.TimeRange) (*PerformanceAnalytics, error)
	GenerateCompensationAnalytics(ctx context.Context, timeRange repositories.TimeRange) (*CompensationAnalytics, error)
	PredictTalentNeeds(ctx context.Context) (*TalentPrediction, error)
}

// Request and Response Types

// JobPostingRequest represents a request to create a job posting
type JobPostingRequest struct {
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Type             entities.EngagementType `json:"type"`
	Department       string                 `json:"department"`
	RequiredSkills   []string               `json:"required_skills"`
	PreferredSkills  []string               `json:"preferred_skills"`
	ExperienceYears  int                    `json:"experience_years"`
	EducationLevel   string                 `json:"education_level"`
	Location         string                 `json:"location"`
	Remote           bool                   `json:"remote"`
	SalaryRange      SalaryRange            `json:"salary_range"`
	Benefits         []string               `json:"benefits"`
	ClosingDate      *time.Time             `json:"closing_date,omitempty"`
}

// SalaryRange represents a salary range for a position
type SalaryRange struct {
	Min      *entities.Money `json:"min"`
	Max      *entities.Money `json:"max"`
	Currency string          `json:"currency"`
}

// ApplicationRequest represents a talent application
type ApplicationRequest struct {
	TalentID        uuid.UUID `json:"talent_id"`
	JobPostingID    uuid.UUID `json:"job_posting_id"`
	CoverLetter     string    `json:"cover_letter"`
	ResumeURL       string    `json:"resume_url,omitempty"`
	PortfolioURLs   []string  `json:"portfolio_urls,omitempty"`
}

// ApplicationScreeningResult represents the result of application screening
type ApplicationScreeningResult struct {
	ApplicationID   uuid.UUID `json:"application_id"`
	Score           float64   `json:"score"`
	SkillMatch      float64   `json:"skill_match"`
	ExperienceMatch float64   `json:"experience_match"`
	RecommendedAction string  `json:"recommended_action"`
	Notes           string    `json:"notes"`
}

// ApplicationDecision represents a decision on an application
type ApplicationDecision struct {
	Decision string `json:"decision"` // Approve, Reject, Interview
	Reason   string `json:"reason"`
	Notes    string `json:"notes,omitempty"`
}

// TalentRequirements represents requirements for talent sourcing
type TalentRequirements struct {
	Skills          []string               `json:"skills"`
	MinExperience   int                    `json:"min_experience"`
	MaxBudget       *entities.Money        `json:"max_budget,omitempty"`
	Availability    map[string]interface{} `json:"availability"`
	Location        string                 `json:"location,omitempty"`
	Remote          bool                   `json:"remote"`
	TalentType      entities.TalentType    `json:"talent_type"`
}

// TalentMatch represents a potential talent match
type TalentMatch struct {
	Talent          *entities.Talent       `json:"talent"`
	MatchScore      float64                `json:"match_score"`
	SkillAlignment  map[string]float64     `json:"skill_alignment"`
	AvailabilityFit float64                `json:"availability_fit"`
	BudgetFit       float64                `json:"budget_fit"`
	Reasoning       string                 `json:"reasoning"`
}

// TalentSearchCriteria represents search criteria for talent
type TalentSearchCriteria struct {
	Keywords        []string               `json:"keywords"`
	Skills          []string               `json:"skills"`
	Location        string                 `json:"location,omitempty"`
	Remote          bool                   `json:"remote"`
	MinReputation   float64                `json:"min_reputation"`
	MaxRate         *entities.Money        `json:"max_rate,omitempty"`
	Availability    string                 `json:"availability,omitempty"`
}

// TalentPoolAnalysis represents analysis of the talent pool
type TalentPoolAnalysis struct {
	TotalTalent     int                    `json:"total_talent"`
	AvailableTalent int                    `json:"available_talent"`
	SkillDistribution map[string]int       `json:"skill_distribution"`
	LocationDistribution map[string]int    `json:"location_distribution"`
	AverageRates    map[string]*entities.Money `json:"average_rates"`
	GrowthTrends    map[string]float64     `json:"growth_trends"`
}

// OnboardingType represents the type of onboarding
type OnboardingType string

const (
	OnboardingTypeHuman OnboardingType = "Human"
	OnboardingTypeAI    OnboardingType = "AI"
)

// OnboardingWorkflow represents an onboarding workflow
type OnboardingWorkflow struct {
	ID            uuid.UUID              `json:"id"`
	TalentID      uuid.UUID              `json:"talent_id"`
	Type          OnboardingType         `json:"type"`
	Status        string                 `json:"status"`
	CurrentStep   string                 `json:"current_step"`
	Steps         []OnboardingStep       `json:"steps"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	EstimatedCompletion time.Time        `json:"estimated_completion"`
}

// OnboardingStep represents a step in the onboarding process
type OnboardingStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Required    bool                   `json:"required"`
	Order       int                    `json:"order"`
	Data        map[string]interface{} `json:"data,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// OnboardingStatus represents the status of onboarding
type OnboardingStatus struct {
	TalentID        uuid.UUID            `json:"talent_id"`
	WorkflowID      uuid.UUID            `json:"workflow_id"`
	Status          string               `json:"status"`
	Progress        float64              `json:"progress"`
	CurrentStep     string               `json:"current_step"`
	NextStep        string               `json:"next_step,omitempty"`
	BlockingIssues  []string             `json:"blocking_issues,omitempty"`
	EstimatedCompletion time.Time        `json:"estimated_completion"`
}

// EngagementDetails represents details for creating an engagement
type EngagementDetails struct {
	Type          entities.EngagementType `json:"type"`
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	StartDate     time.Time               `json:"start_date"`
	EndDate       *time.Time              `json:"end_date,omitempty"`
	HoursPerWeek  int                     `json:"hours_per_week"`
	Rate          *entities.Money         `json:"rate"`
	RateType      string                  `json:"rate_type"`
	Terms         map[string]interface{}  `json:"terms"`
}

// ContractRequest represents a contract generation request
type ContractRequest struct {
	ID            uuid.UUID              `json:"id"`
	TalentID      uuid.UUID              `json:"talent_id"`
	TemplateID    uuid.UUID              `json:"template_id"`
	Terms         map[string]interface{} `json:"terms"`
	DocumentURL   string                 `json:"document_url"`
	SigningURL    string                 `json:"signing_url"`
	ExpiresAt     time.Time              `json:"expires_at"`
}

// SignatureData represents signature information
type SignatureData struct {
	SignerID      uuid.UUID              `json:"signer_id"`
	SignatureID   string                 `json:"signature_id"`
	SignedAt      time.Time              `json:"signed_at"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ContractStatus represents the status of a contract
type ContractStatus struct {
	ContractID    uuid.UUID              `json:"contract_id"`
	Status        string                 `json:"status"`
	SignedBy      []uuid.UUID            `json:"signed_by"`
	PendingSigners []uuid.UUID           `json:"pending_signers"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	DocumentURL   string                 `json:"document_url"`
}

// Additional types for interface completeness

// CapabilityTest represents a test for AI agent capabilities
type CapabilityTest struct {
	TestID      string                 `json:"test_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Input       map[string]interface{} `json:"input"`
	Expected    map[string]interface{} `json:"expected"`
	MaxDuration time.Duration          `json:"max_duration"`
}

// CapabilityTestResult represents the result of capability testing
type CapabilityTestResult struct {
	AgentID     uuid.UUID              `json:"agent_id"`
	TestResults []SingleTestResult     `json:"test_results"`
	OverallScore float64               `json:"overall_score"`
	Passed      bool                   `json:"passed"`
	Duration    time.Duration          `json:"duration"`
	Notes       string                 `json:"notes"`
}

// SingleTestResult represents a single test result
type SingleTestResult struct {
	TestID      string                 `json:"test_id"`
	Passed      bool                   `json:"passed"`
	Score       float64                `json:"score"`
	Output      map[string]interface{} `json:"output"`
	Duration    time.Duration          `json:"duration"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// AIAgentRegistrationRequest represents a request to register an AI agent
type AIAgentRegistrationRequest struct {
	Name            string                 `json:"name"`
	Provider        string                 `json:"provider"`
	Model           string                 `json:"model"`
	APIEndpoint     string                 `json:"api_endpoint"`
	APIVersion      string                 `json:"api_version,omitempty"`
	Capabilities    []string               `json:"capabilities"`
	RateLimits      map[string]interface{} `json:"rate_limits"`
	CostPerRequest  *entities.Money        `json:"cost_per_request,omitempty"`
	CostPerToken    *entities.Money        `json:"cost_per_token,omitempty"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	ResponseTimeMs  int                    `json:"response_time_ms,omitempty"`
}

// AccessRequirement represents an access requirement for onboarding
type AccessRequirement struct {
	Type        string                 `json:"type"`
	System      string                 `json:"system"`
	Permissions []string               `json:"permissions"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialSetup represents credential setup information
type CredentialSetup struct {
	CredentialID   string                 `json:"credential_id"`
	Type           string                 `json:"type"`
	Username       string                 `json:"username"`
	TemporaryPassword string              `json:"temporary_password,omitempty"`
	SetupURL       string                 `json:"setup_url,omitempty"`
	ExpiresAt      time.Time              `json:"expires_at,omitempty"`
	Instructions   string                 `json:"instructions"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TrainingProgressSummary represents a summary of training progress
type TrainingProgressSummary struct {
	TalentID              uuid.UUID `json:"talent_id"`
	TotalTrainings        int       `json:"total_trainings"`
	CompletedTrainings    int       `json:"completed_trainings"`
	InProgressTrainings   int       `json:"in_progress_trainings"`
	RequiredTrainings     int       `json:"required_trainings"`
	CompletionPercentage  float64   `json:"completion_percentage"`
	AverageScore          float64   `json:"average_score"`
	CertificationsEarned  int       `json:"certifications_earned"`
	UpcomingDeadlines     []time.Time `json:"upcoming_deadlines"`
}

// PerformanceMetric represents a performance metric entry
type PerformanceMetric struct {
	MetricID    uuid.UUID              `json:"metric_id"`
	Type        string                 `json:"type"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit,omitempty"`
	Description string                 `json:"description,omitempty"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// PerformanceTrendAnalysis represents performance trend analysis
type PerformanceTrendAnalysis struct {
	TalentID        uuid.UUID              `json:"talent_id"`
	TimeRange       repositories.TimeRange `json:"time_range"`
	Trend           string                  `json:"trend"` // Improving, Declining, Stable
	TrendStrength   float64                 `json:"trend_strength"`
	KeyMetrics      []MetricTrend          `json:"key_metrics"`
	Predictions     []PerformancePrediction `json:"predictions"`
	Recommendations []string               `json:"recommendations"`
}

// MetricTrend represents a trend for a specific metric
type MetricTrend struct {
	MetricType    string    `json:"metric_type"`
	StartValue    float64   `json:"start_value"`
	EndValue      float64   `json:"end_value"`
	ChangePercent float64   `json:"change_percent"`
	Trend         string    `json:"trend"`
	Confidence    float64   `json:"confidence"`
}

// PerformancePrediction represents a performance prediction
type PerformancePrediction struct {
	MetricType     string    `json:"metric_type"`
	PredictedValue float64   `json:"predicted_value"`
	Confidence     float64   `json:"confidence"`
	TimeHorizon    time.Duration `json:"time_horizon"`
	Assumptions    []string  `json:"assumptions"`
}

// ReviewData represents data for conducting a performance review
type ReviewData struct {
	ReviewerID       uuid.UUID              `json:"reviewer_id"`
	QualityScore     float64                `json:"quality_score"`
	ProductivityScore float64               `json:"productivity_score"`
	ReliabilityScore float64                `json:"reliability_score"`
	CommunicationScore float64              `json:"communication_score"`
	OverallRating    entities.PerformanceRating `json:"overall_rating"`
	Strengths        []string               `json:"strengths"`
	ImprovementAreas []string               `json:"improvement_areas"`
	Goals            []string               `json:"goals"`
	Comments         string                 `json:"comments"`
	CompensationAdjustment *entities.Money  `json:"compensation_adjustment,omitempty"`
	NextReviewDate   time.Time              `json:"next_review_date"`
}

// PerformanceGoal represents a performance goal
type PerformanceGoal struct {
	GoalID      uuid.UUID              `json:"goal_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	TargetValue float64                `json:"target_value"`
	CurrentValue float64               `json:"current_value"`
	Unit        string                 `json:"unit,omitempty"`
	Priority    entities.Priority      `json:"priority"`
	DueDate     time.Time              `json:"due_date"`
	Status      string                 `json:"status"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
}

// GoalProgressSummary represents summary of goal progress
type GoalProgressSummary struct {
	TalentID        uuid.UUID         `json:"talent_id"`
	TotalGoals      int               `json:"total_goals"`
	CompletedGoals  int               `json:"completed_goals"`
	InProgressGoals int               `json:"in_progress_goals"`
	OverdueGoals    int               `json:"overdue_goals"`
	OverallProgress float64           `json:"overall_progress"`
	Goals           []PerformanceGoal `json:"goals"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	AlertID     uuid.UUID    `json:"alert_id"`
	TalentID    uuid.UUID    `json:"talent_id"`
	Type        string       `json:"type"`
	Severity    string       `json:"severity"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Metrics     []string     `json:"metrics"`
	CreatedAt   time.Time    `json:"created_at"`
	DueDate     time.Time    `json:"due_date,omitempty"`
	Resolved    bool         `json:"resolved"`
	ResolvedAt  *time.Time   `json:"resolved_at,omitempty"`
}

// PerformanceReport represents a performance report
type PerformanceReport struct {
	ReportID        uuid.UUID                     `json:"report_id"`
	TimeRange       repositories.TimeRange       `json:"time_range"`
	TotalTalent     int                           `json:"total_talent"`
	Distribution    repositories.PerformanceDistribution `json:"distribution"`
	TopPerformers   []*entities.Talent            `json:"top_performers"`
	Underperformers []*entities.Talent            `json:"underperformers"`
	Trends          []MetricTrend                 `json:"trends"`
	KeyInsights     []string                      `json:"key_insights"`
	Recommendations []string                      `json:"recommendations"`
	GeneratedAt     time.Time                     `json:"generated_at"`
}

// Compensation and payroll types

// CompensationPlanRequest represents a request to create a compensation plan
type CompensationPlanRequest struct {
	TalentID         uuid.UUID              `json:"talent_id"`
	EngagementID     *uuid.UUID             `json:"engagement_id,omitempty"`
	Type             string                 `json:"type"` // Salary, Hourly, Project, Retainer
	BaseAmount       *entities.Money        `json:"base_amount"`
	PaymentFrequency string                 `json:"payment_frequency"` // Weekly, BiWeekly, Monthly
	BonusStructure   map[string]interface{} `json:"bonus_structure,omitempty"`
	Benefits         []string               `json:"benefits,omitempty"`
	EffectiveDate    time.Time              `json:"effective_date"`
	EndDate          *time.Time             `json:"end_date,omitempty"`
	TaxWithholding   float64                `json:"tax_withholding"`
	PaymentMethodID  *uuid.UUID             `json:"payment_method_id,omitempty"`
}

// CompensationPlanUpdates represents updates to a compensation plan
type CompensationPlanUpdates struct {
	Type             *string                `json:"type,omitempty"`
	BaseAmount       *entities.Money        `json:"base_amount,omitempty"`
	PaymentFrequency *string                `json:"payment_frequency,omitempty"`
	BonusStructure   map[string]interface{} `json:"bonus_structure,omitempty"`
	Benefits         []string               `json:"benefits,omitempty"`
	EndDate          *time.Time             `json:"end_date,omitempty"`
	TaxWithholding   *float64               `json:"tax_withholding,omitempty"`
	PaymentMethodID  *uuid.UUID             `json:"payment_method_id,omitempty"`
}

// CompensationAdjustment represents a compensation adjustment
type CompensationAdjustment struct {
	TalentID         uuid.UUID       `json:"talent_id"`
	AdjustmentType   string          `json:"adjustment_type"` // Merit, Performance, Market, Promotion
	Amount           *entities.Money `json:"amount"`
	Percentage       float64         `json:"percentage"`
	Reason           string          `json:"reason"`
	EffectiveDate    time.Time       `json:"effective_date"`
	ApprovedBy       uuid.UUID       `json:"approved_by"`
	PerformanceScore float64         `json:"performance_score,omitempty"`
	MarketData       map[string]interface{} `json:"market_data,omitempty"`
}

// PayPeriod represents a payroll period
type PayPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	PayDate   time.Time `json:"pay_date"`
	Type      string    `json:"type"` // Weekly, BiWeekly, SemiMonthly, Monthly
}

// HoursCalculation represents calculated hours for a talent
type HoursCalculation struct {
	TalentID       uuid.UUID `json:"talent_id"`
	PayPeriod      PayPeriod `json:"pay_period"`
	RegularHours   float64   `json:"regular_hours"`
	OvertimeHours  float64   `json:"overtime_hours"`
	PaidTimeOff    float64   `json:"paid_time_off"`
	TotalHours     float64   `json:"total_hours"`
	BillableHours  float64   `json:"billable_hours"`
	NonBillableHours float64 `json:"non_billable_hours"`
	Projects       []ProjectHours `json:"projects"`
}

// ProjectHours represents hours worked on specific projects
type ProjectHours struct {
	ProjectID uuid.UUID `json:"project_id"`
	Hours     float64   `json:"hours"`
	Rate      *entities.Money `json:"rate,omitempty"`
}

// PaymentResult represents the result of a payment processing
type PaymentResult struct {
	PaymentID      uuid.UUID              `json:"payment_id"`
	Status         string                 `json:"status"`
	Amount         *entities.Money        `json:"amount"`
	ProcessedAt    time.Time              `json:"processed_at"`
	TransactionID  string                 `json:"transaction_id,omitempty"`
	PaymentMethod  string                 `json:"payment_method"`
	FailureReason  string                 `json:"failure_reason,omitempty"`
	Fees           *entities.Money        `json:"fees,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TaxCalculation represents tax calculation results
type TaxCalculation struct {
	TalentID         uuid.UUID       `json:"talent_id"`
	GrossAmount      *entities.Money `json:"gross_amount"`
	FederalTax       *entities.Money `json:"federal_tax"`
	StateTax         *entities.Money `json:"state_tax"`
	LocalTax         *entities.Money `json:"local_tax,omitempty"`
	SocialSecurity   *entities.Money `json:"social_security"`
	Medicare         *entities.Money `json:"medicare"`
	Unemployment     *entities.Money `json:"unemployment,omitempty"`
	TotalTax         *entities.Money `json:"total_tax"`
	NetAmount        *entities.Money `json:"net_amount"`
	TaxYear          int             `json:"tax_year"`
	Jurisdiction     string          `json:"jurisdiction"`
	CalculatedAt     time.Time       `json:"calculated_at"`
}

// TaxDocument represents a tax document
type TaxDocument struct {
	DocumentID   uuid.UUID `json:"document_id"`
	TalentID     uuid.UUID `json:"talent_id"`
	Type         string    `json:"type"` // W2, 1099, 1099-NEC, etc.
	TaxYear      int       `json:"tax_year"`
	DocumentURL  string    `json:"document_url"`
	GeneratedAt  time.Time `json:"generated_at"`
	Status       string    `json:"status"`
	EFileStatus  string    `json:"efile_status,omitempty"`
	Corrections  int       `json:"corrections"`
}

// Training and skill types

// TrainingProgramRequest represents a request to create a training program
type TrainingProgramRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Type               string   `json:"type"` // Onboarding, Skill, Compliance
	TargetAudience     string   `json:"target_audience"`
	Duration           int      `json:"duration"` // in hours
	Format             string   `json:"format"` // Online, InPerson, Hybrid
	Prerequisites      []string `json:"prerequisites,omitempty"`
	LearningObjectives []string `json:"learning_objectives"`
	PassingScore       float64  `json:"passing_score"`
	CertificationID    *uuid.UUID `json:"certification_id,omitempty"`
}

// TrainingProgramUpdates represents updates to a training program
type TrainingProgramUpdates struct {
	Name               *string   `json:"name,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Type               *string   `json:"type,omitempty"`
	TargetAudience     *string   `json:"target_audience,omitempty"`
	Duration           *int      `json:"duration,omitempty"`
	Format             *string   `json:"format,omitempty"`
	Prerequisites      []string  `json:"prerequisites,omitempty"`
	LearningObjectives []string  `json:"learning_objectives,omitempty"`
	PassingScore       *float64  `json:"passing_score,omitempty"`
	IsActive           *bool     `json:"is_active,omitempty"`
}

// TrainingProgressUpdate represents an update to training progress
type TrainingProgressUpdate struct {
	Progress        float64                `json:"progress"` // 0-100
	Status          entities.TrainingStatus `json:"status"`
	Score           *float64               `json:"score,omitempty"`
	MaterialProgress map[string]interface{} `json:"material_progress,omitempty"`
	Notes           string                 `json:"notes,omitempty"`
}

// TrainingContentResult represents generated training content
type TrainingContentResult struct {
	ContentID       uuid.UUID              `json:"content_id"`
	Topic           string                 `json:"topic"`
	TargetAudience  string                 `json:"target_audience"`
	Duration        int                    `json:"duration"`
	Format          string                 `json:"format"`
	Content         map[string]interface{} `json:"content"`
	Materials       []TrainingMaterialData `json:"materials"`
	Assessment      map[string]interface{} `json:"assessment,omitempty"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// TrainingMaterialData represents training material data
type TrainingMaterialData struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // Video, Document, Quiz, Exercise
	ContentURL  string                 `json:"content_url,omitempty"`
	Duration    int                    `json:"duration"` // in minutes
	OrderIndex  int                    `json:"order_index"`
	IsRequired  bool                   `json:"is_required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AdaptedTrainingContent represents adapted training content
type AdaptedTrainingContent struct {
	TrainingID      uuid.UUID              `json:"training_id"`
	TalentID        uuid.UUID              `json:"talent_id"`
	AdaptedContent  map[string]interface{} `json:"adapted_content"`
	PersonalizedPath []string              `json:"personalized_path"`
	EstimatedDuration int                  `json:"estimated_duration"`
	DifficultyLevel string                 `json:"difficulty_level"`
	AdaptedAt       time.Time              `json:"adapted_at"`
}

// LearnerProfile represents a learner's profile for content adaptation
type LearnerProfile struct {
	TalentID         uuid.UUID              `json:"talent_id"`
	LearningStyle    string                 `json:"learning_style"`
	PreferredFormat  string                 `json:"preferred_format"`
	Pace             string                 `json:"pace"` // Fast, Medium, Slow
	PriorKnowledge   map[string]float64     `json:"prior_knowledge"`
	CompletedTraining []uuid.UUID           `json:"completed_training"`
	SkillLevel       map[string]string      `json:"skill_level"`
	Preferences      map[string]interface{} `json:"preferences"`
	LastActive       time.Time              `json:"last_active"`
}

// SkillAssessmentResult represents skill assessment results
type SkillAssessmentResult struct {
	TalentID         uuid.UUID            `json:"talent_id"`
	AssessmentID     uuid.UUID            `json:"assessment_id"`
	SkillSet         []string             `json:"skill_set"`
	Results          map[string]SkillScore `json:"results"`
	OverallScore     float64              `json:"overall_score"`
	Recommendations  []string             `json:"recommendations"`
	AssessedAt       time.Time            `json:"assessed_at"`
	ValidUntil       time.Time            `json:"valid_until"`
}

// SkillScore represents a score for a specific skill
type SkillScore struct {
	Skill           string               `json:"skill"`
	Score           float64              `json:"score"`
	Level           entities.SkillLevel  `json:"level"`
	Confidence      float64              `json:"confidence"`
	Evidence        []string             `json:"evidence,omitempty"`
	Recommendations []string             `json:"recommendations,omitempty"`
}

// SkillGap represents a skill gap
type SkillGap struct {
	Skill           string               `json:"skill"`
	RequiredLevel   entities.SkillLevel  `json:"required_level"`
	CurrentLevel    entities.SkillLevel  `json:"current_level"`
	Gap             string               `json:"gap"` // Critical, High, Medium, Low
	Priority        entities.Priority    `json:"priority"`
	EstimatedEffort int                  `json:"estimated_effort"` // in hours
	Recommendations []TrainingRecommendation `json:"recommendations"`
}

// TrainingRecommendation represents a training recommendation
type TrainingRecommendation struct {
	TrainingID      uuid.UUID   `json:"training_id"`
	Title           string      `json:"title"`
	Type            string      `json:"type"`
	Priority        entities.Priority `json:"priority"`
	EstimatedDuration int       `json:"estimated_duration"`
	SkillsAddressed []string    `json:"skills_addressed"`
	Reason          string      `json:"reason"`
	PrerequisitesMet bool       `json:"prerequisites_met"`
	Cost            *entities.Money `json:"cost,omitempty"`
}

// Compliance and legal types

// ComplianceCheckResult represents the result of a compliance check
type ComplianceCheckResult struct {
	CheckID     uuid.UUID              `json:"check_id"`
	Status      string                 `json:"status"` // Passed, Failed, Pending, Error
	Result      string                 `json:"result"`
	Score       float64                `json:"score,omitempty"`
	Details     map[string]interface{} `json:"details"`
	Documents   []string               `json:"documents,omitempty"`
	ValidUntil  *time.Time             `json:"valid_until,omitempty"`
	CompletedAt time.Time              `json:"completed_at"`
	Provider    string                 `json:"provider"`
	Cost        *entities.Money        `json:"cost,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
}

// WorkAuthorizationStatus represents work authorization status
type WorkAuthorizationStatus struct {
	TalentID      uuid.UUID  `json:"talent_id"`
	IsAuthorized  bool       `json:"is_authorized"`
	AuthType      string     `json:"auth_type"` // Citizen, PermanentResident, H1B, etc.
	Jurisdiction  string     `json:"jurisdiction"`
	ValidFrom     time.Time  `json:"valid_from"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	DocumentType  string     `json:"document_type"`
	DocumentNumber string    `json:"document_number"`
	Restrictions  []string   `json:"restrictions,omitempty"`
	VerifiedAt    time.Time  `json:"verified_at"`
	VerifiedBy    string     `json:"verified_by"`
}

// ExpiringVisa represents a visa that's expiring soon
type ExpiringVisa struct {
	TalentID       uuid.UUID `json:"talent_id"`
	TalentName     string    `json:"talent_name"`
	VisaType       string    `json:"visa_type"`
	ExpirationDate time.Time `json:"expiration_date"`
	DaysUntilExpiry int      `json:"days_until_expiry"`
	RenewalRequired bool     `json:"renewal_required"`
	DocumentNumber  string   `json:"document_number"`
	Sponsor         string   `json:"sponsor,omitempty"`
	Priority        string   `json:"priority"` // Critical, High, Medium, Low
}

// WorkAuthorizationData represents work authorization data for updates
type WorkAuthorizationData struct {
	AuthType       string                 `json:"auth_type"`
	DocumentType   string                 `json:"document_type"`
	DocumentNumber string                 `json:"document_number"`
	ValidFrom      time.Time              `json:"valid_from"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	Restrictions   []string               `json:"restrictions,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ContractTerms represents contract terms
type ContractTerms struct {
	ContractType    string                 `json:"contract_type"` // Employment, Contractor, NDA, etc.
	Duration        *time.Duration         `json:"duration,omitempty"`
	RenewalTerms    string                 `json:"renewal_terms,omitempty"`
	TerminationTerms string                `json:"termination_terms"`
	CompensationTerms map[string]interface{} `json:"compensation_terms"`
	IPTerms         string                 `json:"ip_terms,omitempty"`
	NonCompeteTerms string                 `json:"non_compete_terms,omitempty"`
	ConfidentialityTerms string            `json:"confidentiality_terms,omitempty"`
	WorkLocation    string                 `json:"work_location,omitempty"`
	Benefits        []string               `json:"benefits,omitempty"`
	SpecialTerms    map[string]interface{} `json:"special_terms,omitempty"`
}

// ExpiringAgreement represents an agreement that's expiring
type ExpiringAgreement struct {
	AgreementID     uuid.UUID `json:"agreement_id"`
	TalentID        uuid.UUID `json:"talent_id"`
	TalentName      string    `json:"talent_name"`
	ContractType    string    `json:"contract_type"`
	ExpirationDate  time.Time `json:"expiration_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	AutoRenewal     bool      `json:"auto_renewal"`
	RenewalNoticed  bool      `json:"renewal_noticed"`
	Priority        string    `json:"priority"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ViolationID     uuid.UUID              `json:"violation_id"`
	TalentID        uuid.UUID              `json:"talent_id"`
	ViolationType   string                 `json:"violation_type"`
	Severity        string                 `json:"severity"` // Critical, High, Medium, Low
	Description     string                 `json:"description"`
	Regulation      string                 `json:"regulation"`
	DetectedAt      time.Time              `json:"detected_at"`
	Status          string                 `json:"status"` // Open, InProgress, Resolved, Dismissed
	AssignedTo      *uuid.UUID             `json:"assigned_to,omitempty"`
	DueDate         *time.Time             `json:"due_date,omitempty"`
	Evidence        []string               `json:"evidence,omitempty"`
	Impact          string                 `json:"impact"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceResolution represents a compliance resolution
type ComplianceResolution struct {
	ResolutionID    uuid.UUID              `json:"resolution_id"`
	ViolationID     uuid.UUID              `json:"violation_id"`
	ResolutionType  string                 `json:"resolution_type"` // Corrective, Preventive, Dismissed
	Description     string                 `json:"description"`
	Actions         []string               `json:"actions"`
	ResponsibleParty uuid.UUID             `json:"responsible_party"`
	CompletedAt     time.Time              `json:"completed_at"`
	VerifiedBy      uuid.UUID              `json:"verified_by"`
	Evidence        []string               `json:"evidence,omitempty"`
	Cost            *entities.Money        `json:"cost,omitempty"`
	Effectiveness   string                 `json:"effectiveness,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CompensationEquityAnalysis represents compensation equity analysis
type CompensationEquityAnalysis struct {
	AnalysisID      uuid.UUID              `json:"analysis_id"`
	AnalysisDate    time.Time              `json:"analysis_date"`
	TimeRange       repositories.TimeRange `json:"time_range"`
	OverallEquityScore float64             `json:"overall_equity_score"`
	GenderEquity    EquityMetrics          `json:"gender_equity"`
	RaceEquity      EquityMetrics          `json:"race_equity"`
	AgeEquity       EquityMetrics          `json:"age_equity"`
	LocationEquity  EquityMetrics          `json:"location_equity"`
	ExperienceEquity EquityMetrics         `json:"experience_equity"`
	Recommendations []EquityRecommendation `json:"recommendations"`
	RiskAreas       []string               `json:"risk_areas"`
	ComplianceStatus string                `json:"compliance_status"`
}

// EquityMetrics represents equity metrics for a specific dimension
type EquityMetrics struct {
	Dimension       string                 `json:"dimension"`
	PayGap          float64                `json:"pay_gap"` // percentage
	MedianDifference *entities.Money       `json:"median_difference"`
	Groups          map[string]GroupMetrics `json:"groups"`
	TrendDirection  string                 `json:"trend_direction"` // Improving, Worsening, Stable
	StatisticalSignificance bool           `json:"statistical_significance"`
}

// GroupMetrics represents metrics for a specific group
type GroupMetrics struct {
	GroupName       string          `json:"group_name"`
	Count           int             `json:"count"`
	MedianPay       *entities.Money `json:"median_pay"`
	AveragePay      *entities.Money `json:"average_pay"`
	PayRange        PayRange        `json:"pay_range"`
	Representation  float64         `json:"representation"` // percentage of total workforce
}

// PayRange represents a pay range
type PayRange struct {
	Min *entities.Money `json:"min"`
	Max *entities.Money `json:"max"`
}

// EquityRecommendation represents a recommendation for improving equity
type EquityRecommendation struct {
	Priority        entities.Priority `json:"priority"`
	Category        string           `json:"category"`
	Description     string           `json:"description"`
	ExpectedImpact  string           `json:"expected_impact"`
	TimeFrame       string           `json:"time_frame"`
	EstimatedCost   *entities.Money  `json:"estimated_cost,omitempty"`
	ResponsibleParty string          `json:"responsible_party"`
	Metrics         []string         `json:"metrics"` // metrics to track progress
}

// Offboarding types

// TaskCompletion represents completion of an offboarding task
type TaskCompletion struct {
	TaskID      uuid.UUID              `json:"task_id"`
	CompletedBy uuid.UUID              `json:"completed_by"`
	CompletedAt time.Time              `json:"completed_at"`
	Notes       string                 `json:"notes,omitempty"`
	Evidence    []string               `json:"evidence,omitempty"`
	Verification bool                  `json:"verification"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OffboardingStatus represents the status of offboarding
type OffboardingStatus struct {
	TalentID            uuid.UUID `json:"talent_id"`
	ChecklistID         uuid.UUID `json:"checklist_id"`
	Status              string    `json:"status"` // InProgress, Completed, Overdue
	Progress            float64   `json:"progress"` // 0-100
	TotalTasks          int       `json:"total_tasks"`
	CompletedTasks      int       `json:"completed_tasks"`
	OverdueTasks        int       `json:"overdue_tasks"`
	LastWorkingDate     time.Time `json:"last_working_date"`
	ExpectedCompletion  time.Time `json:"expected_completion"`
	ActualCompletion    *time.Time `json:"actual_completion,omitempty"`
	BlockingIssues      []string  `json:"blocking_issues,omitempty"`
	KnowledgeTransferStatus string `json:"knowledge_transfer_status"`
	AssetRecoveryStatus string    `json:"asset_recovery_status"`
	AccessRevocationStatus string `json:"access_revocation_status"`
}

// KnowledgeTransferPlan represents a knowledge transfer plan
type KnowledgeTransferPlan struct {
	PlanID            uuid.UUID        `json:"plan_id"`
	DepartingTalentID uuid.UUID        `json:"departing_talent_id"`
	SuccessorID       uuid.UUID        `json:"successor_id"`
	CreatedAt         time.Time        `json:"created_at"`
	TargetCompletion  time.Time        `json:"target_completion"`
	Status            string           `json:"status"`
	KnowledgeItems    []KnowledgeItem  `json:"knowledge_items"`
	Sessions          []TransferSession `json:"sessions"`
	Progress          float64          `json:"progress"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty"`
}

// KnowledgeItem represents an item of knowledge to be transferred
type KnowledgeItem struct {
	ItemID          uuid.UUID              `json:"item_id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"` // Process, System, Contact, Document, etc.
	Priority        entities.Priority      `json:"priority"`
	EstimatedTime   int                    `json:"estimated_time"` // in hours
	Status          string                 `json:"status"`
	TransferMethod  string                 `json:"transfer_method"` // Documentation, Meeting, Shadowing, etc.
	Documentation   []string               `json:"documentation,omitempty"`
	SystemAccess    []string               `json:"system_access,omitempty"`
	KeyContacts     []uuid.UUID            `json:"key_contacts,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	VerifiedBy      *uuid.UUID             `json:"verified_by,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TransferSession represents a knowledge transfer session
type TransferSession struct {
	SessionID       uuid.UUID              `json:"session_id"`
	ScheduledDate   time.Time              `json:"scheduled_date"`
	Duration        int                    `json:"duration"` // in minutes
	Topics          []string               `json:"topics"`
	Attendees       []uuid.UUID            `json:"attendees"`
	Status          string                 `json:"status"`
	Notes           string                 `json:"notes,omitempty"`
	Recordings      []string               `json:"recordings,omitempty"`
	ActionItems     []ActionItem           `json:"action_items,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ActionItem represents an action item from a transfer session
type ActionItem struct {
	ItemID      uuid.UUID  `json:"item_id"`
	Description string     `json:"description"`
	AssignedTo  uuid.UUID  `json:"assigned_to"`
	DueDate     time.Time  `json:"due_date"`
	Priority    entities.Priority `json:"priority"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// KnowledgeTransferResult represents the result of knowledge transfer
type KnowledgeTransferResult struct {
	PlanID              uuid.UUID `json:"plan_id"`
	CompletionRate      float64   `json:"completion_rate"`
	TotalSessions       int       `json:"total_sessions"`
	CompletedSessions   int       `json:"completed_sessions"`
	DocumentationGaps   []string  `json:"documentation_gaps,omitempty"`
	SuccessorReadiness  float64   `json:"successor_readiness"` // 0-100
	QualityScore        float64   `json:"quality_score"`
	RiskAssessment      string    `json:"risk_assessment"` // Low, Medium, High
	Recommendations     []string  `json:"recommendations"`
	CompletedAt         time.Time `json:"completed_at"`
}

// AccessRevocationResult represents the result of access revocation
type AccessRevocationResult struct {
	TalentID         uuid.UUID                    `json:"talent_id"`
	TotalSystems     int                          `json:"total_systems"`
	RevokedSystems   int                          `json:"revoked_systems"`
	FailedRevocations []FailedRevocation          `json:"failed_revocations,omitempty"`
	SystemStatus     map[string]AccessStatus     `json:"system_status"`
	CompletedAt      time.Time                   `json:"completed_at"`
	VerifiedBy       uuid.UUID                   `json:"verified_by"`
	RiskLevel        string                      `json:"risk_level"`
}

// FailedRevocation represents a failed access revocation
type FailedRevocation struct {
	System      string `json:"system"`
	AccountID   string `json:"account_id"`
	Reason      string `json:"reason"`
	RetryCount  int    `json:"retry_count"`
	LastAttempt time.Time `json:"last_attempt"`
	Contact     string `json:"contact,omitempty"`
}

// AccessStatus represents access status for a system
type AccessStatus struct {
	System       string     `json:"system"`
	AccountID    string     `json:"account_id"`
	Status       string     `json:"status"` // Active, Revoked, Suspended, Failed
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	LastVerified *time.Time `json:"last_verified,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// AssetItem represents an asset to be recovered
type AssetItem struct {
	AssetID       uuid.UUID              `json:"asset_id"`
	Type          string                 `json:"type"` // Laptop, Phone, Badge, Keys, etc.
	Brand         string                 `json:"brand,omitempty"`
	Model         string                 `json:"model,omitempty"`
	SerialNumber  string                 `json:"serial_number,omitempty"`
	AssetTag      string                 `json:"asset_tag,omitempty"`
	Value         *entities.Money        `json:"value,omitempty"`
	Condition     string                 `json:"condition,omitempty"`
	Location      string                 `json:"location,omitempty"`
	AssignedDate  time.Time              `json:"assigned_date"`
	DueDate       time.Time              `json:"due_date"`
	Priority      entities.Priority      `json:"priority"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AssetRecoveryResult represents the result of asset recovery
type AssetRecoveryResult struct {
	TalentID          uuid.UUID                `json:"talent_id"`
	TotalAssets       int                      `json:"total_assets"`
	RecoveredAssets   int                      `json:"recovered_assets"`
	MissingAssets     []AssetItem              `json:"missing_assets,omitempty"`
	DamagedAssets     []AssetItem              `json:"damaged_assets,omitempty"`
	AssetStatus       map[string]AssetRecovery `json:"asset_status"`
	TotalValue        *entities.Money          `json:"total_value"`
	RecoveredValue    *entities.Money          `json:"recovered_value"`
	CompletedAt       time.Time                `json:"completed_at"`
	VerifiedBy        uuid.UUID                `json:"verified_by"`
}

// AssetRecovery represents recovery status of an asset
type AssetRecovery struct {
	AssetID      uuid.UUID  `json:"asset_id"`
	Status       string     `json:"status"` // Recovered, Missing, Damaged, Disposed
	RecoveredAt  *time.Time `json:"recovered_at,omitempty"`
	Condition    string     `json:"condition,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	RecoveredBy  *uuid.UUID `json:"recovered_by,omitempty"`
	Replacement  bool       `json:"replacement"`
}

// ExitInterviewResultOld represents the result of an exit interview (detailed version)
type ExitInterviewResultOld struct {
	InterviewID      uuid.UUID              `json:"interview_id"`
	TalentID         uuid.UUID              `json:"talent_id"`
	InterviewerID    uuid.UUID              `json:"interviewer_id"`
	InterviewDate    time.Time              `json:"interview_date"`
	Duration         int                    `json:"duration"` // in minutes
	Responses        map[string]interface{} `json:"responses"`
	Ratings          map[string]float64     `json:"ratings"`
	Feedback         string                 `json:"feedback"`
	ImprovementSuggestions []string        `json:"improvement_suggestions"`
	WouldRecommend   bool                   `json:"would_recommend"`
	WouldReturnScale int                    `json:"would_return_scale"` // 1-10
	ReasonForLeaving string                 `json:"reason_for_leaving"`
	Manager          string                 `json:"manager"`
	Team             string                 `json:"team"`
	CompanyRating    float64                `json:"company_rating"`
	RoleRating       float64                `json:"role_rating"`
	Notes            string                 `json:"notes,omitempty"`
	RedFlags         []string               `json:"red_flags,omitempty"`
	RetentionOpportunity bool               `json:"retention_opportunity"`
}

// ReferenceProfile represents a reference profile for future hiring
type ReferenceProfile struct {
	TalentID          uuid.UUID   `json:"talent_id"`
	ProfileID         uuid.UUID   `json:"profile_id"`
	Name              string      `json:"name"`
	Role              string      `json:"role"`
	Department        string      `json:"department"`
	EmploymentPeriod  string      `json:"employment_period"`
	PerformanceRating string      `json:"performance_rating"`
	Strengths         []string    `json:"strengths"`
	DevelopmentAreas  []string    `json:"development_areas"`
	Achievements      []string    `json:"achievements"`
	TechnicalSkills   []string    `json:"technical_skills"`
	SoftSkills        []string    `json:"soft_skills"`
	WorkStyle         string      `json:"work_style"`
	TeamPlayer        bool        `json:"team_player"`
	Leadership        bool        `json:"leadership"`
	Innovation        bool        `json:"innovation"`
	Reliability       bool        `json:"reliability"`
	CommunicationSkills bool      `json:"communication_skills"`
	ProblemSolving    bool        `json:"problem_solving"`
	Adaptability      bool        `json:"adaptability"`
	WouldRehire       bool        `json:"would_rehire"`
	EligibleForRehire bool        `json:"eligible_for_rehire"`
	ReasonForLeaving  string      `json:"reason_for_leaving"`
	FinalAssessment   string      `json:"final_assessment"`
	ReferencingManager uuid.UUID  `json:"referencing_manager"`
	CreatedAt         time.Time   `json:"created_at"`
	ValidUntil        time.Time   `json:"valid_until"`
}

// Analytics types

// WorkforceMetrics represents overall workforce metrics
type WorkforceMetrics struct {
	TimeRange        repositories.TimeRange `json:"time_range"`
	TotalTalent      int                    `json:"total_talent"`
	ActiveTalent     int                    `json:"active_talent"`
	AvailableTalent  int                    `json:"available_talent"`
	HumanTalent      int                    `json:"human_talent"`
	AIAgents         int                    `json:"ai_agents"`
	UtilizationRate  float64                `json:"utilization_rate"`
	NewHires         int                    `json:"new_hires"`
	Departures       int                    `json:"departures"`
	TurnoverRate     float64                `json:"turnover_rate"`
	GrowthRate       float64                `json:"growth_rate"`
	AverageRating    float64                `json:"average_rating"`
	SkillDistribution map[string]int        `json:"skill_distribution"`
	LocationDistribution map[string]int     `json:"location_distribution"`
	TypeDistribution map[string]int         `json:"type_distribution"`
	TotalCost        *entities.Money        `json:"total_cost"`
	AverageCost      *entities.Money        `json:"average_cost"`
	ROI              float64                `json:"roi"`
}

// TalentUtilizationAnalysis represents talent utilization analysis
type TalentUtilizationAnalysis struct {
	TimeRange         repositories.TimeRange    `json:"time_range"`
	OverallUtilization float64                  `json:"overall_utilization"`
	ByTalentType      map[string]float64        `json:"by_talent_type"`
	BySkill           map[string]float64        `json:"by_skill"`
	ByLocation        map[string]float64        `json:"by_location"`
	Underutilized     []*entities.Talent        `json:"underutilized"`
	Overutilized      []*entities.Talent        `json:"overutilized"`
	OptimalAllocation []AllocationRecommendation `json:"optimal_allocation"`
	CapacityGaps      []CapacityGap              `json:"capacity_gaps"`
	Trends            []UtilizationTrend         `json:"trends"`
	Forecasts         []UtilizationForecast      `json:"forecasts"`
}

// AllocationRecommendation represents an allocation recommendation
type AllocationRecommendation struct {
	TalentID         uuid.UUID       `json:"talent_id"`
	CurrentProject   *uuid.UUID      `json:"current_project,omitempty"`
	RecommendedProject *uuid.UUID    `json:"recommended_project,omitempty"`
	Reason           string          `json:"reason"`
	ExpectedImpact   float64         `json:"expected_impact"`
	Priority         entities.Priority `json:"priority"`
	EstimatedEffort  int             `json:"estimated_effort"` // in hours
}

// CapacityGap represents a capacity gap
type CapacityGap struct {
	Skill            string          `json:"skill"`
	RequiredCapacity int             `json:"required_capacity"` // in hours
	AvailableCapacity int            `json:"available_capacity"` // in hours
	Gap              int             `json:"gap"` // in hours
	Impact           string          `json:"impact"` // Critical, High, Medium, Low
	Solutions        []GapSolution   `json:"solutions"`
}

// GapSolution represents a solution for capacity gaps
type GapSolution struct {
	Type             string          `json:"type"` // Hire, Train, Reallocate, Contract
	Description      string          `json:"description"`
	TimeToImplement  time.Duration   `json:"time_to_implement"`
	Cost             *entities.Money `json:"cost,omitempty"`
	Impact           float64         `json:"impact"` // 0-1
	Feasibility      float64         `json:"feasibility"` // 0-1
	Priority         entities.Priority `json:"priority"`
}

// UtilizationTrend represents a utilization trend
type UtilizationTrend struct {
	Period          string  `json:"period"` // Weekly, Monthly, Quarterly
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Utilization     float64 `json:"utilization"`
	ChangePercent   float64 `json:"change_percent"`
	Trend           string  `json:"trend"` // Increasing, Decreasing, Stable
}

// UtilizationForecast represents a utilization forecast
type UtilizationForecast struct {
	Period          string  `json:"period"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	PredictedUtilization float64 `json:"predicted_utilization"`
	Confidence      float64 `json:"confidence"`
	Assumptions     []string `json:"assumptions"`
}

// ProjectForecast represents a project forecast for demand planning
type ProjectForecast struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	ProjectName     string                 `json:"project_name"`
	StartDate       time.Time              `json:"start_date"`
	EndDate         time.Time              `json:"end_date"`
	RequiredSkills  []SkillRequirement     `json:"required_skills"`
	TotalEffort     int                    `json:"total_effort"` // in hours
	Priority        entities.Priority      `json:"priority"`
	Confidence      float64                `json:"confidence"`
	Dependencies    []uuid.UUID            `json:"dependencies,omitempty"`
}

// SkillRequirement represents a skill requirement for a project
type SkillRequirement struct {
	Skill           string                `json:"skill"`
	Level           entities.SkillLevel   `json:"level"`
	Hours           int                   `json:"hours"`
	Priority        entities.Priority     `json:"priority"`
	FlexibilityScore float64              `json:"flexibility_score"` // 0-1
}

// TalentDemandForecast represents a talent demand forecast
type TalentDemandForecast struct {
	ForecastID      uuid.UUID              `json:"forecast_id"`
	TimeHorizon     time.Duration          `json:"time_horizon"`
	GeneratedAt     time.Time              `json:"generated_at"`
	TotalDemand     int                    `json:"total_demand"` // in hours
	DemandBySkill   map[string]int         `json:"demand_by_skill"`
	DemandByPeriod  map[string]int         `json:"demand_by_period"`
	SupplyGaps      []SupplyGap            `json:"supply_gaps"`
	HiringPlan      []HiringRecommendation `json:"hiring_plan"`
	TrainingPlan    []TrainingPlan         `json:"training_plan"`
	Confidence      float64                `json:"confidence"`
	Assumptions     []string               `json:"assumptions"`
	RiskFactors     []string               `json:"risk_factors"`
}

// SupplyGap represents a supply gap
type SupplyGap struct {
	Skill           string          `json:"skill"`
	Level           entities.SkillLevel `json:"level"`
	RequiredHours   int             `json:"required_hours"`
	AvailableHours  int             `json:"available_hours"`
	GapHours        int             `json:"gap_hours"`
	GapTalent       int             `json:"gap_talent"` // number of people needed
	Urgency         string          `json:"urgency"` // Critical, High, Medium, Low
	Solutions       []GapSolution   `json:"solutions"`
}

// HiringRecommendation represents a hiring recommendation
type HiringRecommendation struct {
	Skill           string                 `json:"skill"`
	Level           entities.SkillLevel    `json:"level"`
	NumberToHire    int                    `json:"number_to_hire"`
	Priority        entities.Priority      `json:"priority"`
	Timeline        string                 `json:"timeline"`
	EstimatedCost   *entities.Money        `json:"estimated_cost"`
	ExpectedROI     float64                `json:"expected_roi"`
	Rationale       string                 `json:"rationale"`
	Prerequisites   []string               `json:"prerequisites,omitempty"`
}

// TrainingPlan represents a training plan recommendation
type TrainingPlan struct {
	TargetSkill     string                 `json:"target_skill"`
	CurrentTalent   []uuid.UUID            `json:"current_talent"`
	TrainingProgram *uuid.UUID             `json:"training_program,omitempty"`
	Duration        time.Duration          `json:"duration"`
	Cost            *entities.Money        `json:"cost"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	Priority        entities.Priority      `json:"priority"`
	Prerequisites   []string               `json:"prerequisites,omitempty"`
}

// PerformanceTrendReport represents a performance trend report
type PerformanceTrendReport struct {
	ReportID        uuid.UUID              `json:"report_id"`
	TimeRange       repositories.TimeRange `json:"time_range"`
	OverallTrend    string                 `json:"overall_trend"` // Improving, Declining, Stable
	TrendStrength   float64                `json:"trend_strength"`
	KeyFindings     []string               `json:"key_findings"`
	MetricTrends    []MetricTrend          `json:"metric_trends"`
	TalentTrends    []TalentTrend          `json:"talent_trends"`
	Alerts          []TrendAlert           `json:"alerts"`
	Recommendations []string               `json:"recommendations"`
	NextReviewDate  time.Time              `json:"next_review_date"`
}

// TalentTrend represents performance trend for specific talent
type TalentTrend struct {
	TalentID        uuid.UUID `json:"talent_id"`
	TalentName      string    `json:"talent_name"`
	Trend           string    `json:"trend"`
	TrendStrength   float64   `json:"trend_strength"`
	CurrentRating   float64   `json:"current_rating"`
	PreviousRating  float64   `json:"previous_rating"`
	ChangePercent   float64   `json:"change_percent"`
	RiskLevel       string    `json:"risk_level"`
}

// TrendAlert represents a performance trend alert
type TrendAlert struct {
	AlertID      uuid.UUID `json:"alert_id"`
	Type         string    `json:"type"` // Declining, Plateau, Volatility
	Severity     string    `json:"severity"` // Critical, High, Medium, Low
	Description  string    `json:"description"`
	AffectedTalent []uuid.UUID `json:"affected_talent"`
	Metric       string    `json:"metric"`
	Threshold    float64   `json:"threshold"`
	ActualValue  float64   `json:"actual_value"`
	Recommendations []string `json:"recommendations"`
}

// PerformanceCriteria represents criteria for performance analysis
type PerformanceCriteria struct {
	MetricTypes     []string  `json:"metric_types"`
	MinRating       float64   `json:"min_rating,omitempty"`
	TimeRange       repositories.TimeRange `json:"time_range"`
	TalentTypes     []entities.TalentType `json:"talent_types,omitempty"`
	Skills          []string  `json:"skills,omitempty"`
	Locations       []string  `json:"locations,omitempty"`
	Departments     []string  `json:"departments,omitempty"`
	IncludeInactive bool      `json:"include_inactive"`
	SortBy          string    `json:"sort_by"`
	Limit           int       `json:"limit"`
}

// TurnoverAnalysis represents turnover analysis
type TurnoverAnalysis struct {
	TimeRange        repositories.TimeRange `json:"time_range"`
	OverallTurnover  float64                `json:"overall_turnover"`
	VoluntaryTurnover float64               `json:"voluntary_turnover"`
	InvoluntaryTurnover float64             `json:"involuntary_turnover"`
	TurnoverByType   map[string]float64     `json:"turnover_by_type"`
	TurnoverBySkill  map[string]float64     `json:"turnover_by_skill"`
	TurnoverByLocation map[string]float64   `json:"turnover_by_location"`
	ExitReasons      map[string]int         `json:"exit_reasons"`
	CostOfTurnover   *entities.Money        `json:"cost_of_turnover"`
	ReplacementTime  float64                `json:"replacement_time"` // in days
	TopReasons       []ExitReason           `json:"top_reasons"`
	Trends           []TurnoverTrend        `json:"trends"`
	Benchmarks       TurnoverBenchmarks     `json:"benchmarks"`
}

// ExitReason represents an exit reason with analytics
type ExitReason struct {
	Reason          string  `json:"reason"`
	Count           int     `json:"count"`
	Percentage      float64 `json:"percentage"`
	Trend           string  `json:"trend"`
	AverageRating   float64 `json:"average_rating"`
	Preventability  string  `json:"preventability"` // High, Medium, Low
}

// TurnoverTrend represents turnover trend over time
type TurnoverTrend struct {
	Period         string  `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	TurnoverRate   float64 `json:"turnover_rate"`
	ExitCount      int     `json:"exit_count"`
	Trend          string  `json:"trend"`
}

// TurnoverBenchmarks represents industry turnover benchmarks
type TurnoverBenchmarks struct {
	Industry        string  `json:"industry"`
	IndustryAverage float64 `json:"industry_average"`
	CompanySize     string  `json:"company_size"`
	SizeAverage     float64 `json:"size_average"`
	GeographyAverage float64 `json:"geography_average"`
	PerformanceRating string `json:"performance_rating"` // Above, At, Below benchmark
}

// RetentionRisk represents retention risk for a talent
type RetentionRisk struct {
	TalentID         uuid.UUID `json:"talent_id"`
	TalentName       string    `json:"talent_name"`
	RiskScore        float64   `json:"risk_score"` // 0-100
	RiskLevel        string    `json:"risk_level"` // Critical, High, Medium, Low
	RiskFactors      []RiskFactor `json:"risk_factors"`
	Recommendations  []RetentionAction `json:"recommendations"`
	LastReviewDate   time.Time `json:"last_review_date"`
	NextReviewDate   time.Time `json:"next_review_date"`
	TimeInRole       time.Duration `json:"time_in_role"`
	PerformanceRating float64  `json:"performance_rating"`
	EngagementScore  float64   `json:"engagement_score"`
	FlightRisk       bool      `json:"flight_risk"`
}

// RiskFactor represents a risk factor for turnover
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Weight      float64 `json:"weight"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
	Trending    string  `json:"trending"` // Increasing, Decreasing, Stable
}

// RetentionAction represents an action to improve retention
type RetentionAction struct {
	Action          string                 `json:"action"`
	Priority        entities.Priority      `json:"priority"`
	ExpectedImpact  float64                `json:"expected_impact"`
	Cost            *entities.Money        `json:"cost,omitempty"`
	Timeline        string                 `json:"timeline"`
	ResponsibleParty string                `json:"responsible_party"`
	Prerequisites   []string               `json:"prerequisites,omitempty"`
}

// TurnoverPrediction represents a turnover prediction
type TurnoverPrediction struct {
	TalentID           uuid.UUID `json:"talent_id"`
	PredictionDate     time.Time `json:"prediction_date"`
	TurnoverProbability float64  `json:"turnover_probability"` // 0-1
	Confidence         float64   `json:"confidence"` // 0-1
	TimeFrame          string    `json:"time_frame"` // 30days, 60days, 90days, 6months, 1year
	KeyIndicators      []string  `json:"key_indicators"`
	ModelVersion       string    `json:"model_version"`
	LastUpdated        time.Time `json:"last_updated"`
}

// TalentROI represents talent return on investment
type TalentROI struct {
	TalentID           uuid.UUID              `json:"talent_id"`
	CalculationPeriod  repositories.TimeRange `json:"calculation_period"`
	TotalInvestment    *entities.Money        `json:"total_investment"`
	TotalValue         *entities.Money        `json:"total_value"`
	ROIPercentage      float64                `json:"roi_percentage"`
	ROIRatio           float64                `json:"roi_ratio"`
	PaybackPeriod      time.Duration          `json:"payback_period"`
	Investments        []Investment           `json:"investments"`
	ValueGenerators    []ValueGenerator       `json:"value_generators"`
	Benchmarks         ROIBenchmarks          `json:"benchmarks"`
	Projections        []ROIProjection        `json:"projections"`
}

// Investment represents an investment in talent
type Investment struct {
	Type        string          `json:"type"` // Salary, Training, Equipment, etc.
	Amount      *entities.Money `json:"amount"`
	Date        time.Time       `json:"date"`
	Recurring   bool            `json:"recurring"`
	Description string          `json:"description"`
}

// ValueGenerator represents value generated by talent
type ValueGenerator struct {
	Type        string          `json:"type"` // ProjectValue, CostSavings, Revenue, etc.
	Amount      *entities.Money `json:"amount"`
	Date        time.Time       `json:"date"`
	Attribution float64         `json:"attribution"` // percentage of value attributed to this talent
	Description string          `json:"description"`
	ProjectID   *uuid.UUID      `json:"project_id,omitempty"`
}

// ROIBenchmarks represents ROI benchmarks
type ROIBenchmarks struct {
	IndustryAverage    float64 `json:"industry_average"`
	RoleAverage        float64 `json:"role_average"`
	ExperienceAverage  float64 `json:"experience_average"`
	CompanyAverage     float64 `json:"company_average"`
	TopPerformerAverage float64 `json:"top_performer_average"`
}

// ROIProjection represents ROI projection
type ROIProjection struct {
	Period          string  `json:"period"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	ProjectedROI    float64 `json:"projected_roi"`
	Confidence      float64 `json:"confidence"`
	Assumptions     []string `json:"assumptions"`
}

// TalentCostAnalysis represents talent cost analysis
type TalentCostAnalysis struct {
	TimeRange           repositories.TimeRange    `json:"time_range"`
	TotalCost           *entities.Money           `json:"total_cost"`
	CostByType          map[string]*entities.Money `json:"cost_by_type"`
	CostBySkill         map[string]*entities.Money `json:"cost_by_skill"`
	CostByLocation      map[string]*entities.Money `json:"cost_by_location"`
	CostByTalent        map[uuid.UUID]*entities.Money `json:"cost_by_talent"`
	AverageCostPerHour  *entities.Money           `json:"average_cost_per_hour"`
	CostTrends          []CostTrend               `json:"cost_trends"`
	CostOptimization    []CostOptimization        `json:"cost_optimization"`
	Benchmarks          CostBenchmarks            `json:"benchmarks"`
}

// CostTrend represents cost trend over time
type CostTrend struct {
	Period       string          `json:"period"`
	StartDate    time.Time       `json:"start_date"`
	EndDate      time.Time       `json:"end_date"`
	TotalCost    *entities.Money `json:"total_cost"`
	ChangePercent float64        `json:"change_percent"`
	Trend        string          `json:"trend"`
}

// CostOptimization represents cost optimization opportunity
type CostOptimization struct {
	Opportunity      string                 `json:"opportunity"`
	CurrentCost      *entities.Money        `json:"current_cost"`
	OptimizedCost    *entities.Money        `json:"optimized_cost"`
	PotentialSavings *entities.Money        `json:"potential_savings"`
	Implementation   string                 `json:"implementation"`
	Risk             string                 `json:"risk"`
	Priority         entities.Priority      `json:"priority"`
}

// CostBenchmarks represents cost benchmarks
type CostBenchmarks struct {
	IndustryAverage   *entities.Money `json:"industry_average"`
	RegionAverage     *entities.Money `json:"region_average"`
	SizeAverage       *entities.Money `json:"size_average"`
	PerformanceRating string          `json:"performance_rating"`
}

// CompensationBenchmark represents compensation benchmark data
type CompensationBenchmark struct {
	Role               string                    `json:"role"`
	Location           string                    `json:"location"`
	ExperienceLevel    string                    `json:"experience_level"`
	MarketData         repositories.CompensationBenchmarks `json:"market_data"`
	CompanyPosition    string                    `json:"company_position"` // Above, At, Below market
	RecommendedRange   SalaryRange               `json:"recommended_range"`
	CompetitiveFactors []string                  `json:"competitive_factors"`
	LastUpdated        time.Time                 `json:"last_updated"`
	DataSources        []string                  `json:"data_sources"`
	Confidence         float64                   `json:"confidence"`
}

// ComplianceStatus represents compliance status for a talent
type ComplianceStatus struct {
	TalentID         uuid.UUID         `json:"talent_id"`
	OverallStatus    string            `json:"overall_status"`
	LastChecked      time.Time         `json:"last_checked"`
	ComplianceChecks []ComplianceCheck `json:"compliance_checks"`
	RequiredActions  []string          `json:"required_actions"`
	ExpiringItems    []ComplianceItem  `json:"expiring_items"`
}

// ComplianceCheck represents a specific compliance check
type ComplianceCheck struct {
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Notes       string    `json:"notes"`
}

// ComplianceItem represents a compliance item that may expire
type ComplianceItem struct {
	Type       string    `json:"type"`
	ExpiryDate time.Time `json:"expiry_date"`
	Status     string    `json:"status"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ReportID           uuid.UUID                     `json:"report_id"`
	TimeRange          repositories.TimeRange       `json:"time_range"`
	TotalTalent        int                          `json:"total_talent"`
	CompliantTalent    int                          `json:"compliant_talent"`
	NonCompliantTalent int                          `json:"non_compliant_talent"`
	ComplianceByType   map[string]ComplianceMetrics `json:"compliance_by_type"`
	Issues             []ComplianceIssue            `json:"issues"`
	GeneratedAt        time.Time                    `json:"generated_at"`
}

// ComplianceMetrics represents compliance metrics for a specific type
type ComplianceMetrics struct {
	Total     int     `json:"total"`
	Compliant int     `json:"compliant"`
	Rate      float64 `json:"rate"`
}

// ComplianceIssue represents a compliance issue
type ComplianceIssue struct {
	TalentID    uuid.UUID `json:"talent_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	DetectedAt  time.Time `json:"detected_at"`
}

// OffboardingWorkflow represents an offboarding workflow
type OffboardingWorkflow struct {
	ID                  uuid.UUID         `json:"id"`
	TalentID            uuid.UUID         `json:"talent_id"`
	Reason              string            `json:"reason"`
	Status              string            `json:"status"`
	Steps               []OffboardingStep `json:"steps"`
	StartedAt           time.Time         `json:"started_at"`
	EstimatedCompletion time.Time         `json:"estimated_completion"`
}

// OffboardingStep represents a step in the offboarding process
type OffboardingStep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Order       int    `json:"order"`
}

// ExitInterviewData represents data from an exit interview
type ExitInterviewData struct {
	InterviewerID uuid.UUID              `json:"interviewer_id"`
	Feedback      map[string]interface{} `json:"feedback"`
}

// ExitInterviewResult represents the result of an exit interview
type ExitInterviewResult struct {
	TalentID        uuid.UUID `json:"talent_id"`
	InterviewerID   uuid.UUID `json:"interviewer_id"`
	Feedback        map[string]interface{} `json:"feedback"`
	Recommendations []string  `json:"recommendations"`
	CompletedAt     time.Time `json:"completed_at"`
}

// OffboardingReport represents an offboarding report
type OffboardingReport struct {
	TalentID          uuid.UUID `json:"talent_id"`
	TalentName        string    `json:"talent_name"`
	Department        string    `json:"department"`
	OffboardingDate   time.Time `json:"offboarding_date"`
	Reason            string    `json:"reason"`
	AccessRevoked     bool      `json:"access_revoked"`
	AssetsReturned    bool      `json:"assets_returned"`
	FinalPayProcessed bool      `json:"final_pay_processed"`
	Documentation     []string  `json:"documentation"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// TalentAnalytics represents talent analytics data
type TalentAnalytics struct {
	TotalTalent          int            `json:"total_talent"`
	HumanTalent          int            `json:"human_talent"`
	AIAgents             int            `json:"ai_agents"`
	AvailableTalent      int            `json:"available_talent"`
	EngagedTalent        int            `json:"engaged_talent"`
	SkillDistribution    map[string]int `json:"skill_distribution"`
	LocationDistribution map[string]int `json:"location_distribution"`
	AverageReputation    float64        `json:"average_reputation"`
	GeneratedAt          time.Time      `json:"generated_at"`
}

// PerformanceAnalytics represents performance analytics data
type PerformanceAnalytics struct {
	TimeRange               repositories.TimeRange              `json:"time_range"`
	AveragePerformance      float64                             `json:"average_performance"`
	PerformanceDistribution repositories.PerformanceDistribution `json:"performance_distribution"`
	TopPerformers           int                                 `json:"top_performers"`
	UnderPerformers         int                                 `json:"under_performers"`
	PerformanceTrends       []MetricTrend                       `json:"performance_trends"`
	GeneratedAt             time.Time                           `json:"generated_at"`
}

// CompensationAnalytics represents compensation analytics data
type CompensationAnalytics struct {
	TimeRange                repositories.TimeRange        `json:"time_range"`
	TotalPayroll             *entities.Money               `json:"total_payroll"`
	AverageCompensation      *entities.Money               `json:"average_compensation"`
	CompensationDistribution CompensationDistribution      `json:"compensation_distribution"`
	PayrollTrends            []PayrollTrend                `json:"payroll_trends"`
	GeneratedAt              time.Time                     `json:"generated_at"`
}

// CompensationDistribution represents compensation distribution data
type CompensationDistribution struct {
	ByLevel    map[string]*entities.Money `json:"by_level"`
	BySkill    map[string]*entities.Money `json:"by_skill"`
	ByLocation map[string]*entities.Money `json:"by_location"`
	ByType     map[string]*entities.Money `json:"by_type"`
	Percentiles map[string]*entities.Money `json:"percentiles"`
}

// PayrollTrend represents payroll trend data (duplicate fix)
type PayrollTrend struct {
	Period        string          `json:"period"`
	Amount        *entities.Money `json:"amount"`
	ChangePercent float64         `json:"change_percent"`
	Trend         string          `json:"trend"`
}

// TalentPrediction represents talent needs prediction
type TalentPrediction struct {
	TimeHorizon     time.Duration `json:"time_horizon"`
	PredictedNeeds  []TalentNeed  `json:"predicted_needs"`
	ConfidenceScore float64       `json:"confidence_score"`
	Recommendations []string      `json:"recommendations"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

// TalentNeed represents a predicted talent need
type TalentNeed struct {
	SkillCategory     string        `json:"skill_category"`
	RequiredCount     int           `json:"required_count"`
	Confidence        float64       `json:"confidence"`
	TimeToFulfill     time.Duration `json:"time_to_fulfill"`
	RecommendedAction string        `json:"recommended_action"`
}
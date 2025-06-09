package hr_management

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// HRService is the main HR management service
type HRService struct {
	talentRepo       repositories.TalentRepository
	engagementRepo   repositories.EngagementRepository
	assignmentRepo   repositories.WorkAssignmentRepository
	performanceRepo  repositories.PerformanceRepository
	compensationRepo repositories.CompensationRepository
	trainingRepo     repositories.TrainingRepository
	applicationRepo  repositories.TalentApplicationRepository
	complianceRepo   repositories.ComplianceRepository
	offboardingRepo  repositories.OffboardingRepository
	eventRepo        repositories.EventRepository

	// Embedded services
	talentAcquisition     TalentAcquisitionService
	onboarding           OnboardingService
	performanceManagement PerformanceManagementService
	compensation         CompensationService
	training             TrainingService
	complianceManagement ComplianceManagementService
	offboarding          OffboardingService
	analytics            HRAnalyticsService
}

// NewHRService creates a new HR service instance
func NewHRService(
	talentRepo repositories.TalentRepository,
	engagementRepo repositories.EngagementRepository,
	assignmentRepo repositories.WorkAssignmentRepository,
	performanceRepo repositories.PerformanceRepository,
	compensationRepo repositories.CompensationRepository,
	trainingRepo repositories.TrainingRepository,
	applicationRepo repositories.TalentApplicationRepository,
	complianceRepo repositories.ComplianceRepository,
	offboardingRepo repositories.OffboardingRepository,
	eventRepo repositories.EventRepository,
) *HRService {
	service := &HRService{
		talentRepo:       talentRepo,
		engagementRepo:   engagementRepo,
		assignmentRepo:   assignmentRepo,
		performanceRepo:  performanceRepo,
		compensationRepo: compensationRepo,
		trainingRepo:     trainingRepo,
		applicationRepo:  applicationRepo,
		complianceRepo:   complianceRepo,
		offboardingRepo:  offboardingRepo,
		eventRepo:        eventRepo,
	}

	// Initialize embedded services
	service.talentAcquisition = NewTalentAcquisitionService(applicationRepo, talentRepo, eventRepo)
	service.onboarding = NewOnboardingService(talentRepo, complianceRepo, trainingRepo, eventRepo)
	service.performanceManagement = NewPerformanceManagementService(performanceRepo, talentRepo, eventRepo)
	service.compensation = NewCompensationService(compensationRepo, talentRepo, eventRepo)
	service.training = NewTrainingService(trainingRepo, talentRepo, eventRepo)
	service.complianceManagement = NewComplianceManagementService(complianceRepo, talentRepo, eventRepo)
	service.offboarding = NewOffboardingService(offboardingRepo, talentRepo, eventRepo)
	service.analytics = NewHRAnalyticsService(talentRepo, performanceRepo, compensationRepo)

	return service
}

// Talent Management

// CreateTalent creates a new talent profile
func (h *HRService) CreateTalent(ctx context.Context, request TalentCreationRequest) (*entities.Talent, error) {
	talent := &entities.Talent{
		ID:              uuid.New(),
		Type:            request.Type,
		Name:            request.Name,
		Email:           request.Email,
		Status:          entities.TalentStatusAvailable,
		ReputationScore: 0.0,
		Skills:          []entities.Skill{},
		Certifications:  []entities.Certification{},
		Availability:    request.Availability,
		HourlyRate:      request.HourlyRate,
		Currency:        request.Currency,
		Location:        request.Location,
		Timezone:        request.Timezone,
		ProfileData:     request.ProfileData,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.talentRepo.CreateTalent(ctx, talent); err != nil {
		return nil, fmt.Errorf("failed to create talent: %w", err)
	}

	// Add skills if provided
	for _, skillReq := range request.Skills {
		skill := &entities.Skill{
			ID:        uuid.New(),
			TalentID:  talent.ID,
			Name:      skillReq.Name,
			Category:  skillReq.Category,
			Level:     skillReq.Level,
			YearsUsed: skillReq.YearsUsed,
			Verified:  false,
			CreatedAt: time.Now(),
		}
		if err := h.talentRepo.AddTalentSkill(ctx, skill); err != nil {
			return nil, fmt.Errorf("failed to add skill: %w", err)
		}
	}

	return talent, nil
}

// GetTalent retrieves a talent by ID
func (h *HRService) GetTalent(ctx context.Context, id uuid.UUID) (*entities.Talent, error) {
	talent, err := h.talentRepo.GetTalentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get talent: %w", err)
	}

	// Load skills and certifications
	skills, err := h.talentRepo.GetTalentSkills(ctx, id)
	if err == nil && skills != nil {
		talent.Skills = make([]entities.Skill, len(skills))
		for i, skill := range skills {
			talent.Skills[i] = *skill
		}
	}

	certifications, err := h.talentRepo.GetTalentCertifications(ctx, id)
	if err == nil && certifications != nil {
		talent.Certifications = make([]entities.Certification, len(certifications))
		for i, cert := range certifications {
			talent.Certifications[i] = *cert
		}
	}

	return talent, nil
}

// UpdateTalent updates an existing talent profile
func (h *HRService) UpdateTalent(ctx context.Context, id uuid.UUID, updates TalentUpdateRequest) (*entities.Talent, error) {
	talent, err := h.talentRepo.GetTalentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get talent: %w", err)
	}

	// Apply updates
	if updates.Name != nil {
		talent.Name = *updates.Name
	}
	if updates.Email != nil {
		talent.Email = *updates.Email
	}
	if updates.Status != nil {
		oldStatus := talent.Status
		talent.Status = *updates.Status

		// Emit status change event
		if oldStatus != talent.Status {
			// TODO: Emit talent status changed event
			_ = oldStatus // prevent unused variable warning until event is implemented
		}
	}
	if updates.HourlyRate != nil {
		talent.HourlyRate = updates.HourlyRate
	}
	if updates.Location != nil {
		talent.Location = *updates.Location
	}
	if updates.Timezone != nil {
		talent.Timezone = *updates.Timezone
	}
	if updates.Availability != nil {
		talent.Availability = updates.Availability
	}
	if updates.ProfileData != nil {
		talent.ProfileData = updates.ProfileData
	}

	talent.UpdatedAt = time.Now()

	if err := h.talentRepo.UpdateTalent(ctx, talent); err != nil {
		return nil, fmt.Errorf("failed to update talent: %w", err)
	}

	return talent, nil
}

// SearchTalent searches for talent based on criteria
func (h *HRService) SearchTalent(ctx context.Context, filter repositories.TalentFilter) ([]*entities.Talent, int, error) {
	return h.talentRepo.SearchTalent(ctx, filter)
}

// Engagement Management

// CreateEngagement creates a new engagement
func (h *HRService) CreateEngagement(ctx context.Context, request EngagementCreationRequest) (*entities.Engagement, error) {
	// Verify talent exists and is available
	talent, err := h.talentRepo.GetTalentByID(ctx, request.TalentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	if talent.Status != entities.TalentStatusAvailable {
		return nil, fmt.Errorf("talent is not available for engagement")
	}

	engagement := &entities.Engagement{
		ID:                uuid.New(),
		TalentID:          request.TalentID,
		Type:              request.Type,
		Status:            entities.EngagementStatusDraft,
		Title:             request.Title,
		Description:       request.Description,
		StartDate:         request.StartDate,
		EndDate:           request.EndDate,
		HoursPerWeek:      request.HoursPerWeek,
		RateType:          request.RateType,
		Rate:              request.Rate,
		Currency:          request.Currency,
		ContractID:        request.ContractID,
		ManagerID:         request.ManagerID,
		TeamID:            request.TeamID,
		PerformanceMetrics: make(map[string]interface{}),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.engagementRepo.CreateEngagement(ctx, engagement); err != nil {
		return nil, fmt.Errorf("failed to create engagement: %w", err)
	}

	// Emit engagement created event
	event := events.NewEngagementCreatedEvent(engagement)
	if err := h.eventRepo.Save(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to emit engagement created event: %v\n", err)
	}

	return engagement, nil
}

// ActivateEngagement activates an engagement
func (h *HRService) ActivateEngagement(ctx context.Context, engagementID uuid.UUID) error {
	engagement, err := h.engagementRepo.GetEngagementByID(ctx, engagementID)
	if err != nil {
		return fmt.Errorf("engagement not found: %w", err)
	}

	if engagement.Status != entities.EngagementStatusPending {
		return fmt.Errorf("engagement must be in pending status to activate")
	}

	// Update engagement status
	engagement.Status = entities.EngagementStatusActive
	engagement.UpdatedAt = time.Now()

	if err := h.engagementRepo.UpdateEngagement(ctx, engagement); err != nil {
		return fmt.Errorf("failed to update engagement: %w", err)
	}

	// Update talent status
	talent, err := h.talentRepo.GetTalentByID(ctx, engagement.TalentID)
	if err != nil {
		return fmt.Errorf("failed to get talent: %w", err)
	}

	talent.Status = entities.TalentStatusEngaged
	talent.UpdatedAt = time.Now()

	if err := h.talentRepo.UpdateTalent(ctx, talent); err != nil {
		return fmt.Errorf("failed to update talent status: %w", err)
	}

	return nil
}

// CompleteEngagement completes an engagement
func (h *HRService) CompleteEngagement(ctx context.Context, engagementID uuid.UUID, completionNotes string) error {
	engagement, err := h.engagementRepo.GetEngagementByID(ctx, engagementID)
	if err != nil {
		return fmt.Errorf("engagement not found: %w", err)
	}

	if engagement.Status != entities.EngagementStatusActive {
		return fmt.Errorf("engagement must be active to complete")
	}

	// Update engagement status
	engagement.Status = entities.EngagementStatusCompleted
	now := time.Now()
	engagement.EndDate = &now
	engagement.UpdatedAt = now

	if err := h.engagementRepo.UpdateEngagement(ctx, engagement); err != nil {
		return fmt.Errorf("failed to update engagement: %w", err)
	}

	// Update talent status back to available
	talent, err := h.talentRepo.GetTalentByID(ctx, engagement.TalentID)
	if err != nil {
		return fmt.Errorf("failed to get talent: %w", err)
	}

	talent.Status = entities.TalentStatusAvailable
	talent.UpdatedAt = time.Now()

	if err := h.talentRepo.UpdateTalent(ctx, talent); err != nil {
		return fmt.Errorf("failed to update talent status: %w", err)
	}

	return nil
}

// Assignment Management

// CreateWorkAssignment creates a new work assignment
func (h *HRService) CreateWorkAssignment(ctx context.Context, request AssignmentCreationRequest) (*entities.WorkAssignment, error) {
	// Verify engagement exists and is active
	engagement, err := h.engagementRepo.GetEngagementByID(ctx, request.EngagementID)
	if err != nil {
		return nil, fmt.Errorf("engagement not found: %w", err)
	}

	if engagement.Status != entities.EngagementStatusActive {
		return nil, fmt.Errorf("engagement must be active to create assignments")
	}

	assignment := &entities.WorkAssignment{
		ID:             uuid.New(),
		EngagementID:   request.EngagementID,
		TalentID:       engagement.TalentID,
		ProjectID:      request.ProjectID,
		Title:          request.Title,
		Description:    request.Description,
		Priority:       request.Priority,
		Status:         "Created",
		EstimatedHours: request.EstimatedHours,
		ActualHours:    0,
		DueDate:        request.DueDate,
		Deliverables:   []entities.Deliverable{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.assignmentRepo.CreateAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	return assignment, nil
}

// CompleteAssignment marks an assignment as completed
func (h *HRService) CompleteAssignment(ctx context.Context, assignmentID uuid.UUID, actualHours float64, qualityScore float64) error {
	assignment, err := h.assignmentRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("assignment not found: %w", err)
	}

	// Update assignment
	assignment.Status = "Completed"
	assignment.ActualHours = actualHours
	assignment.QualityScore = &qualityScore
	now := time.Now()
	assignment.CompletedAt = &now
	assignment.UpdatedAt = now

	if err := h.assignmentRepo.UpdateAssignment(ctx, assignment); err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	// Create assignment completed event
	event := &events.AssignmentCompletedEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: events.HREventAssignmentCompleted,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "hr-service",
			},
		},
		AssignmentID:   assignment.ID,
		TalentID:       assignment.TalentID,
		EngagementID:   assignment.EngagementID,
		QualityScore:   qualityScore,
		ActualHours:    actualHours,
		EstimatedHours: assignment.EstimatedHours,
		CompletedAt:    now,
	}

	if err := h.eventRepo.Save(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to emit assignment completed event: %v\n", err)
	}

	return nil
}

// Service Delegation Methods

// TalentAcquisition returns the talent acquisition service
func (h *HRService) TalentAcquisition() TalentAcquisitionService {
	return h.talentAcquisition
}

// Onboarding returns the onboarding service
func (h *HRService) Onboarding() OnboardingService {
	return h.onboarding
}

// PerformanceManagement returns the performance management service
func (h *HRService) PerformanceManagement() PerformanceManagementService {
	return h.performanceManagement
}

// Compensation returns the compensation service
func (h *HRService) Compensation() CompensationService {
	return h.compensation
}

// Training returns the training service
func (h *HRService) Training() TrainingService {
	return h.training
}

// ComplianceManagement returns the compliance management service
func (h *HRService) ComplianceManagement() ComplianceManagementService {
	return h.complianceManagement
}

// Offboarding returns the offboarding service
func (h *HRService) Offboarding() OffboardingService {
	return h.offboarding
}

// Analytics returns the HR analytics service
func (h *HRService) Analytics() HRAnalyticsService {
	return h.analytics
}

// Utility Methods

// GetEngagementsByTalent retrieves all engagements for a talent
func (h *HRService) GetEngagementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.Engagement, error) {
	return h.engagementRepo.GetEngagementsByTalent(ctx, talentID)
}

// GetActiveEngagements retrieves all active engagements
func (h *HRService) GetActiveEngagements(ctx context.Context) ([]*entities.Engagement, error) {
	return h.engagementRepo.GetActiveEngagements(ctx)
}

// GetAssignmentsByTalent retrieves all assignments for a talent
func (h *HRService) GetAssignmentsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.WorkAssignment, error) {
	return h.assignmentRepo.GetAssignmentsByTalent(ctx, talentID)
}

// GetOverdueAssignments retrieves all overdue assignments
func (h *HRService) GetOverdueAssignments(ctx context.Context) ([]*entities.WorkAssignment, error) {
	return h.assignmentRepo.GetOverdueAssignments(ctx)
}

// Request Types

// TalentCreationRequest represents a request to create talent
type TalentCreationRequest struct {
	Type         entities.TalentType    `json:"type"`
	Name         string                 `json:"name"`
	Email        string                 `json:"email"`
	Skills       []SkillCreationRequest `json:"skills"`
	Availability map[string]interface{} `json:"availability"`
	HourlyRate   *entities.Money        `json:"hourly_rate,omitempty"`
	Currency     string                 `json:"currency"`
	Location     string                 `json:"location,omitempty"`
	Timezone     string                 `json:"timezone,omitempty"`
	ProfileData  map[string]interface{} `json:"profile_data"`
}

// SkillCreationRequest represents a request to create a skill
type SkillCreationRequest struct {
	Name      string              `json:"name"`
	Category  string              `json:"category"`
	Level     entities.SkillLevel `json:"level"`
	YearsUsed float64             `json:"years_used"`
}

// TalentUpdateRequest represents a request to update talent
type TalentUpdateRequest struct {
	Name         *string                `json:"name,omitempty"`
	Email        *string                `json:"email,omitempty"`
	Status       *entities.TalentStatus `json:"status,omitempty"`
	HourlyRate   *entities.Money        `json:"hourly_rate,omitempty"`
	Location     *string                `json:"location,omitempty"`
	Timezone     *string                `json:"timezone,omitempty"`
	Availability map[string]interface{} `json:"availability,omitempty"`
	ProfileData  map[string]interface{} `json:"profile_data,omitempty"`
}

// EngagementCreationRequest represents a request to create an engagement
type EngagementCreationRequest struct {
	TalentID     uuid.UUID              `json:"talent_id"`
	Type         entities.EngagementType `json:"type"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	StartDate    time.Time               `json:"start_date"`
	EndDate      *time.Time              `json:"end_date,omitempty"`
	HoursPerWeek int                     `json:"hours_per_week"`
	RateType     string                  `json:"rate_type"`
	Rate         *entities.Money         `json:"rate"`
	Currency     string                  `json:"currency"`
	ContractID   *uuid.UUID              `json:"contract_id,omitempty"`
	ManagerID    *uuid.UUID              `json:"manager_id,omitempty"`
	TeamID       *uuid.UUID              `json:"team_id,omitempty"`
}

// AssignmentCreationRequest represents a request to create an assignment
type AssignmentCreationRequest struct {
	EngagementID   uuid.UUID        `json:"engagement_id"`
	ProjectID      *uuid.UUID       `json:"project_id,omitempty"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Priority       entities.Priority `json:"priority"`
	EstimatedHours float64          `json:"estimated_hours"`
	DueDate        *time.Time       `json:"due_date,omitempty"`
}

// Placeholder service implementations (to be implemented in separate files)

func NewTalentAcquisitionService(applicationRepo repositories.TalentApplicationRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) TalentAcquisitionService {
	return NewTalentAcquisitionServiceImpl(applicationRepo, talentRepo, eventRepo)
}

func NewOnboardingService(talentRepo repositories.TalentRepository, complianceRepo repositories.ComplianceRepository, trainingRepo repositories.TrainingRepository, eventRepo repositories.EventRepository) OnboardingService {
	return NewOnboardingServiceImpl(talentRepo, complianceRepo, trainingRepo, eventRepo)
}

func NewPerformanceManagementService(performanceRepo repositories.PerformanceRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) PerformanceManagementService {
	return NewPerformanceManagementServiceImpl(performanceRepo, talentRepo, eventRepo)
}

func NewCompensationService(compensationRepo repositories.CompensationRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) CompensationService {
	return NewCompensationServiceImpl(compensationRepo, talentRepo, eventRepo)
}

func NewTrainingService(trainingRepo repositories.TrainingRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) TrainingService {
	return NewTrainingServiceImpl(trainingRepo, talentRepo, eventRepo)
}

func NewComplianceManagementService(complianceRepo repositories.ComplianceRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) ComplianceManagementService {
	return NewComplianceManagementServiceImpl(complianceRepo, talentRepo, eventRepo)
}

func NewOffboardingService(offboardingRepo repositories.OffboardingRepository, talentRepo repositories.TalentRepository, eventRepo repositories.EventRepository) OffboardingService {
	return NewOffboardingServiceImpl(offboardingRepo, talentRepo, eventRepo)
}

func NewHRAnalyticsService(talentRepo repositories.TalentRepository, performanceRepo repositories.PerformanceRepository, compensationRepo repositories.CompensationRepository) HRAnalyticsService {
	return NewHRAnalyticsServiceImpl(talentRepo, performanceRepo, compensationRepo)
}
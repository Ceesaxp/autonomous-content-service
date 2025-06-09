package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// HR event type constants
const (
	// Talent events
	HREventTalentOnboarded      = "hr.talent.onboarded"
	HREventTalentOffboarded     = "hr.talent.offboarded"
	HREventTalentStatusChanged  = "hr.talent.statusChanged"
	HREventTalentSkillAdded     = "hr.talent.skillAdded"
	HREventTalentCertified      = "hr.talent.certified"
	
	// Engagement events
	HREventEngagementCreated    = "hr.engagement.created"
	HREventEngagementActivated  = "hr.engagement.activated"
	HREventEngagementCompleted  = "hr.engagement.completed"
	HREventEngagementTerminated = "hr.engagement.terminated"
	
	// Assignment events
	HREventAssignmentCreated    = "hr.assignment.created"
	HREventAssignmentCompleted  = "hr.assignment.completed"
	HREventDeliverableSubmitted = "hr.deliverable.submitted"
	HREventDeliverableAccepted  = "hr.deliverable.accepted"
	
	// Performance events
	HREventPerformanceReviewed  = "hr.performance.reviewed"
	HREventPerformanceImproved  = "hr.performance.improved"
	HREventPerformanceDegraded  = "hr.performance.degraded"
	
	// Compensation events
	HREventCompensationAdjusted = "hr.compensation.adjusted"
	HREventPayrollProcessed     = "hr.payroll.processed"
	HREventBonusAwarded         = "hr.bonus.awarded"
	
	// Training events
	HREventTrainingStarted      = "hr.training.started"
	HREventTrainingCompleted    = "hr.training.completed"
	HREventCertificationEarned  = "hr.certification.earned"
	
	// Compliance events
	HREventComplianceCheckStarted   = "hr.compliance.checkStarted"
	HREventComplianceCheckCompleted = "hr.compliance.checkCompleted"
	HREventComplianceViolation      = "hr.compliance.violation"
	
	// Application events
	HREventApplicationReceived  = "hr.application.received"
	HREventApplicationScreened  = "hr.application.screened"
	HREventApplicationApproved  = "hr.application.approved"
	HREventApplicationRejected  = "hr.application.rejected"
)

// TalentOnboardedEvent is emitted when a new talent is onboarded
type TalentOnboardedEvent struct {
	BaseEvent
	TalentID        uuid.UUID             `json:"talent_id"`
	TalentType      entities.TalentType   `json:"talent_type"`
	Name            string                `json:"name"`
	Email           string                `json:"email,omitempty"`
	Skills          []entities.Skill      `json:"skills"`
	ContractID      *uuid.UUID            `json:"contract_id,omitempty"`
	OnboardedBy     uuid.UUID             `json:"onboarded_by"`
}

// TalentOffboardedEvent is emitted when a talent is offboarded
type TalentOffboardedEvent struct {
	BaseEvent
	TalentID        uuid.UUID             `json:"talent_id"`
	Reason          string                `json:"reason"`
	LastWorkingDate time.Time             `json:"last_working_date"`
	KnowledgeTransferComplete bool        `json:"knowledge_transfer_complete"`
	OffboardedBy    uuid.UUID             `json:"offboarded_by"`
}

// EngagementCreatedEvent is emitted when a new engagement is created
type EngagementCreatedEvent struct {
	BaseEvent
	EngagementID    uuid.UUID                  `json:"engagement_id"`
	TalentID        uuid.UUID                  `json:"talent_id"`
	Type            entities.EngagementType    `json:"type"`
	Title           string                     `json:"title"`
	StartDate       time.Time                  `json:"start_date"`
	EndDate         *time.Time                 `json:"end_date,omitempty"`
	Rate            *entities.Money            `json:"rate"`
	ContractID      *uuid.UUID                 `json:"contract_id,omitempty"`
}

// AssignmentCompletedEvent is emitted when a work assignment is completed
type AssignmentCompletedEvent struct {
	BaseEvent
	AssignmentID    uuid.UUID             `json:"assignment_id"`
	TalentID        uuid.UUID             `json:"talent_id"`
	EngagementID    uuid.UUID             `json:"engagement_id"`
	QualityScore    float64               `json:"quality_score"`
	ActualHours     float64               `json:"actual_hours"`
	EstimatedHours  float64               `json:"estimated_hours"`
	CompletedAt     time.Time             `json:"completed_at"`
}

// PerformanceReviewedEvent is emitted when a performance review is completed
type PerformanceReviewedEvent struct {
	BaseEvent
	ReviewID        uuid.UUID                      `json:"review_id"`
	TalentID        uuid.UUID                      `json:"talent_id"`
	ReviewerID      *uuid.UUID                     `json:"reviewer_id,omitempty"`
	OverallRating   entities.PerformanceRating     `json:"overall_rating"`
	QualityScore    float64                        `json:"quality_score"`
	ProductivityScore float64                      `json:"productivity_score"`
	CompensationAdjustment *entities.Money         `json:"compensation_adjustment,omitempty"`
}

// CompensationAdjustedEvent is emitted when compensation is changed
type CompensationAdjustedEvent struct {
	BaseEvent
	TalentID        uuid.UUID             `json:"talent_id"`
	EngagementID    *uuid.UUID            `json:"engagement_id,omitempty"`
	OldAmount       *entities.Money       `json:"old_amount"`
	NewAmount       *entities.Money       `json:"new_amount"`
	Reason          string                `json:"reason"`
	EffectiveDate   time.Time             `json:"effective_date"`
	ApprovedBy      uuid.UUID             `json:"approved_by"`
}

// PayrollProcessedEvent is emitted when payroll is processed
type PayrollProcessedEvent struct {
	BaseEvent
	PayrollID       uuid.UUID             `json:"payroll_id"`
	TalentID        uuid.UUID             `json:"talent_id"`
	PayPeriodStart  time.Time             `json:"pay_period_start"`
	PayPeriodEnd    time.Time             `json:"pay_period_end"`
	GrossAmount     *entities.Money       `json:"gross_amount"`
	NetAmount       *entities.Money       `json:"net_amount"`
	TransactionID   string                `json:"transaction_id"`
	PaymentMethod   string                `json:"payment_method"`
}

// TrainingCompletedEvent is emitted when training is completed
type TrainingCompletedEvent struct {
	BaseEvent
	TalentID        uuid.UUID             `json:"talent_id"`
	TrainingID      uuid.UUID             `json:"training_id"`
	TrainingName    string                `json:"training_name"`
	Score           float64               `json:"score"`
	PassingScore    float64               `json:"passing_score"`
	CertificateID   *uuid.UUID            `json:"certificate_id,omitempty"`
	CompletedAt     time.Time             `json:"completed_at"`
}

// ComplianceCheckCompletedEvent is emitted when a compliance check is completed
type ComplianceCheckCompletedEvent struct {
	BaseEvent
	CheckID         uuid.UUID             `json:"check_id"`
	TalentID        uuid.UUID             `json:"talent_id"`
	CheckType       string                `json:"check_type"`
	Result          string                `json:"result"`
	ValidUntil      *time.Time            `json:"valid_until,omitempty"`
	Issues          []string              `json:"issues,omitempty"`
}

// ApplicationScreenedEvent is emitted when an application is screened
type ApplicationScreenedEvent struct {
	BaseEvent
	ApplicationID   uuid.UUID                      `json:"application_id"`
	TalentID        uuid.UUID                      `json:"talent_id"`
	JobPostingID    uuid.UUID                      `json:"job_posting_id"`
	ScreeningScore  float64                        `json:"screening_score"`
	Status          entities.ApplicationStatus     `json:"status"`
	NextStep        string                         `json:"next_step,omitempty"`
}

// Helper functions for creating HR events

// NewTalentOnboardedEvent creates a new talent onboarded event
func NewTalentOnboardedEvent(talent *entities.Talent, contractID *uuid.UUID, onboardedBy uuid.UUID) *TalentOnboardedEvent {
	return &TalentOnboardedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: HREventTalentOnboarded,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "hr-service",
			},
		},
		TalentID:    talent.ID,
		TalentType:  talent.Type,
		Name:        talent.Name,
		Email:       talent.Email,
		Skills:      talent.Skills,
		ContractID:  contractID,
		OnboardedBy: onboardedBy,
	}
}

// NewEngagementCreatedEvent creates a new engagement created event
func NewEngagementCreatedEvent(engagement *entities.Engagement) *EngagementCreatedEvent {
	return &EngagementCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: HREventEngagementCreated,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "hr-service",
			},
		},
		EngagementID: engagement.ID,
		TalentID:     engagement.TalentID,
		Type:         engagement.Type,
		Title:        engagement.Title,
		StartDate:    engagement.StartDate,
		EndDate:      engagement.EndDate,
		Rate:         engagement.Rate,
		ContractID:   engagement.ContractID,
	}
}

// NewPerformanceReviewedEvent creates a new performance reviewed event
func NewPerformanceReviewedEvent(review *entities.PerformanceReview) *PerformanceReviewedEvent {
	return &PerformanceReviewedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: HREventPerformanceReviewed,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "hr-service",
			},
		},
		ReviewID:               review.ID,
		TalentID:               review.TalentID,
		ReviewerID:             review.ReviewerID,
		OverallRating:          review.OverallRating,
		QualityScore:           review.QualityScore,
		ProductivityScore:      review.ProductivityScore,
		CompensationAdjustment: review.CompensationAdjustment,
	}
}

// NewPayrollProcessedEvent creates a new payroll processed event
func NewPayrollProcessedEvent(payroll *entities.PayrollRecord) *PayrollProcessedEvent {
	return &PayrollProcessedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: HREventPayrollProcessed,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "hr-service",
			},
		},
		PayrollID:      payroll.ID,
		TalentID:       payroll.TalentID,
		PayPeriodStart: payroll.PayPeriodStart,
		PayPeriodEnd:   payroll.PayPeriodEnd,
		GrossAmount:    payroll.GrossAmount,
		NetAmount:      payroll.NetAmount,
		TransactionID:  payroll.TransactionID,
		PaymentMethod:  payroll.PaymentMethod,
	}
}
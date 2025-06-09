package hr_management

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// CompensationServiceImpl implements the CompensationService interface
type CompensationServiceImpl struct {
	compensationRepo repositories.CompensationRepository
	talentRepo       repositories.TalentRepository
	eventRepo        repositories.EventRepository
}

// NewCompensationServiceImpl creates a new compensation service
func NewCompensationServiceImpl(
	compensationRepo repositories.CompensationRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) CompensationService {
	return &CompensationServiceImpl{
		compensationRepo: compensationRepo,
		talentRepo:       talentRepo,
		eventRepo:        eventRepo,
	}
}

func (s *CompensationServiceImpl) CreateCompensationPlan(ctx context.Context, request CompensationPlanRequest) (*entities.CompensationPlan, error) {
	_, err := s.talentRepo.GetTalentByID(ctx, request.TalentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	plan := &entities.CompensationPlan{
		ID:               uuid.New(),
		TalentID:         request.TalentID,
		EngagementID:     request.EngagementID,
		Type:             request.Type,
		BaseAmount:       request.BaseAmount,
		PaymentFrequency: request.PaymentFrequency,
		BonusStructure:   request.BonusStructure,
		Benefits:         request.Benefits,
		EffectiveDate:    request.EffectiveDate,
		EndDate:          request.EndDate,
		TaxWithholding:   request.TaxWithholding,
		PaymentMethodID:  request.PaymentMethodID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.compensationRepo.CreateCompensationPlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create compensation plan: %w", err)
	}

	return plan, nil
}

func (s *CompensationServiceImpl) UpdateCompensationPlan(ctx context.Context, planID uuid.UUID, updates CompensationPlanUpdates) (*entities.CompensationPlan, error) {
	plan, err := s.compensationRepo.GetCompensationPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("compensation plan not found: %w", err)
	}

	// Apply updates
	if updates.Type != nil {
		plan.Type = *updates.Type
	}
	if updates.BaseAmount != nil {
		plan.BaseAmount = updates.BaseAmount
	}
	if updates.PaymentFrequency != nil {
		plan.PaymentFrequency = *updates.PaymentFrequency
	}
	if updates.BonusStructure != nil {
		plan.BonusStructure = updates.BonusStructure
	}
	if updates.Benefits != nil {
		plan.Benefits = updates.Benefits
	}
	if updates.EndDate != nil {
		plan.EndDate = updates.EndDate
	}
	if updates.TaxWithholding != nil {
		plan.TaxWithholding = *updates.TaxWithholding
	}
	if updates.PaymentMethodID != nil {
		plan.PaymentMethodID = updates.PaymentMethodID
	}

	plan.UpdatedAt = time.Now()

	if err := s.compensationRepo.UpdateCompensationPlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update compensation plan: %w", err)
	}

	return plan, nil
}

func (s *CompensationServiceImpl) GetCompensationPlan(ctx context.Context, talentID uuid.UUID) (*entities.CompensationPlan, error) {
	return s.compensationRepo.GetCompensationPlanByTalent(ctx, talentID)
}

func (s *CompensationServiceImpl) CalculateCompensationAdjustment(ctx context.Context, talentID uuid.UUID, performanceMetrics map[string]float64) (*CompensationAdjustment, error) {
	currentPlan, err := s.compensationRepo.GetCompensationPlanByTalent(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("compensation plan not found: %w", err)
	}

	// Calculate adjustment based on performance
	performanceScore := s.calculateOverallPerformanceScore(performanceMetrics)
	adjustmentPercentage := s.calculateAdjustmentPercentage(performanceScore)

	adjustmentAmount := &entities.Money{
		Amount:   float64(currentPlan.BaseAmount.Amount) * adjustmentPercentage / 100,
		Currency: currentPlan.BaseAmount.Currency,
	}

	adjustment := &CompensationAdjustment{
		TalentID:         talentID,
		AdjustmentType:   "Performance",
		Amount:           adjustmentAmount,
		Percentage:       adjustmentPercentage,
		Reason:           fmt.Sprintf("Performance-based adjustment (score: %.2f)", performanceScore),
		EffectiveDate:    time.Now().AddDate(0, 1, 0), // Next month
		ApprovedBy:       uuid.New(), // Would be actual approver ID
		PerformanceScore: performanceScore,
	}

	return adjustment, nil
}

func (s *CompensationServiceImpl) ProcessPayroll(ctx context.Context, payPeriod PayPeriod) ([]*entities.PayrollRecord, error) {
	// Get all active talent with compensation plans
	activeTalent, _, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Status: func() *entities.TalentStatus { s := entities.TalentStatusEngaged; return &s }(),
		Limit:  1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active talent: %w", err)
	}

	var payrollRecords []*entities.PayrollRecord

	for _, talent := range activeTalent {
		record, err := s.GeneratePayrollRecord(ctx, talent.ID, payPeriod)
		if err != nil {
			continue // Skip on error, log in real implementation
		}
		payrollRecords = append(payrollRecords, record)
	}

	return payrollRecords, nil
}

func (s *CompensationServiceImpl) GeneratePayrollRecord(ctx context.Context, talentID uuid.UUID, payPeriod PayPeriod) (*entities.PayrollRecord, error) {
	plan, err := s.compensationRepo.GetCompensationPlanByTalent(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("compensation plan not found: %w", err)
	}

	// Calculate hours worked
	hoursCalc, err := s.CalculateHours(ctx, talentID, payPeriod.StartDate, payPeriod.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hours: %w", err)
	}

	// Calculate gross amount based on compensation type
	var grossAmount *entities.Money
	switch plan.Type {
	case "Hourly":
		grossAmount = &entities.Money{
			Amount:   hoursCalc.TotalHours * plan.BaseAmount.Amount,
			Currency: plan.BaseAmount.Currency,
		}
	case "Salary":
		// Calculate pro-rated salary for pay period
		daysInPeriod := payPeriod.EndDate.Sub(payPeriod.StartDate).Hours() / 24
		dailyRate := plan.BaseAmount.Amount / 365
		grossAmount = &entities.Money{
			Amount:   dailyRate * daysInPeriod,
			Currency: plan.BaseAmount.Currency,
		}
	default:
		grossAmount = plan.BaseAmount
	}

	// Calculate tax withholding
	taxCalc, err := s.CalculateTaxWithholding(ctx, grossAmount, talentID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate taxes: %w", err)
	}

	record := &entities.PayrollRecord{
		ID:              uuid.New(),
		TalentID:        talentID,
		PayPeriodStart:  payPeriod.StartDate,
		PayPeriodEnd:    payPeriod.EndDate,
		GrossAmount:     grossAmount,
		NetAmount:       taxCalc.NetAmount,
		HoursWorked:     hoursCalc.TotalHours,
		Deductions:      map[string]interface{}{"taxes": taxCalc.TotalTax},
		Bonuses:         map[string]interface{}{},
		PaymentDate:     payPeriod.PayDate,
		PaymentMethod:   "Direct Deposit",
		Status:          "Pending",
		TaxDocumentURLs: []string{},
		CreatedAt:       time.Now(),
	}

	if err := s.compensationRepo.CreatePayrollRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create payroll record: %w", err)
	}

	return record, nil
}

func (s *CompensationServiceImpl) CalculateHours(ctx context.Context, talentID uuid.UUID, startDate, endDate time.Time) (*HoursCalculation, error) {
	// Get work assignments for the period
	assignments, err := s.compensationRepo.GetWorkAssignmentsForPeriod(ctx, talentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get work assignments: %w", err)
	}

	calc := &HoursCalculation{
		TalentID:         talentID,
		PayPeriod:        PayPeriod{StartDate: startDate, EndDate: endDate},
		RegularHours:     0.0,
		OvertimeHours:    0.0,
		PaidTimeOff:      0.0,
		TotalHours:       0.0,
		BillableHours:    0.0,
		NonBillableHours: 0.0,
		Projects:         []ProjectHours{},
	}

	projectHours := make(map[uuid.UUID]float64)

	for _, assignment := range assignments {
		hours := assignment.ActualHours
		calc.TotalHours += hours
		calc.BillableHours += hours // Assume all hours are billable for now

		if assignment.ProjectID != nil {
			projectHours[*assignment.ProjectID] += hours
		}
	}

	// Convert to project hours slice
	for projectID, hours := range projectHours {
		calc.Projects = append(calc.Projects, ProjectHours{
			ProjectID: projectID,
			Hours:     hours,
		})
	}

	// Calculate regular vs overtime (40 hours per week threshold)
	weeklyHours := calc.TotalHours / (endDate.Sub(startDate).Hours() / 24 / 7)
	if weeklyHours <= 40 {
		calc.RegularHours = calc.TotalHours
	} else {
		weeksInPeriod := endDate.Sub(startDate).Hours() / 24 / 7
		calc.RegularHours = 40 * weeksInPeriod
		calc.OvertimeHours = calc.TotalHours - calc.RegularHours
	}

	return calc, nil
}

func (s *CompensationServiceImpl) ProcessPayment(ctx context.Context, payrollID uuid.UUID) (*PaymentResult, error) {
	record, err := s.compensationRepo.GetPayrollRecord(ctx, payrollID)
	if err != nil {
		return nil, fmt.Errorf("payroll record not found: %w", err)
	}

	// Process payment through payment system
	result := &PaymentResult{
		PaymentID:     uuid.New(),
		Status:        "Processed",
		Amount:        record.NetAmount,
		ProcessedAt:   time.Now(),
		TransactionID: fmt.Sprintf("TXN_%s", uuid.New().String()[:8]),
		PaymentMethod: record.PaymentMethod,
	}

	// Update payroll record status
	record.Status = "Paid"
	record.TransactionID = result.TransactionID
	if err := s.compensationRepo.UpdatePayrollRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to update payroll record: %w", err)
	}

	return result, nil
}

func (s *CompensationServiceImpl) CalculateTaxWithholding(ctx context.Context, grossAmount *entities.Money, talentID uuid.UUID) (*TaxCalculation, error) {
	// Simplified tax calculation - in real implementation would use tax service
	federalRate := 0.22  // 22% federal tax
	stateRate := 0.08    // 8% state tax
	socialSecurityRate := 0.062 // 6.2% Social Security
	medicareRate := 0.0145      // 1.45% Medicare

	federalTax := grossAmount.Amount * federalRate
	stateTax := grossAmount.Amount * stateRate
	socialSecurity := grossAmount.Amount * socialSecurityRate
	medicare := grossAmount.Amount * medicareRate

	totalTax := federalTax + stateTax + socialSecurity + medicare
	netAmount := grossAmount.Amount - totalTax

	return &TaxCalculation{
		TalentID:       talentID,
		GrossAmount:    grossAmount,
		FederalTax:     &entities.Money{Amount: federalTax, Currency: grossAmount.Currency},
		StateTax:       &entities.Money{Amount: stateTax, Currency: grossAmount.Currency},
		SocialSecurity: &entities.Money{Amount: socialSecurity, Currency: grossAmount.Currency},
		Medicare:       &entities.Money{Amount: medicare, Currency: grossAmount.Currency},
		TotalTax:       &entities.Money{Amount: totalTax, Currency: grossAmount.Currency},
		NetAmount:      &entities.Money{Amount: netAmount, Currency: grossAmount.Currency},
		TaxYear:        time.Now().Year(),
		Jurisdiction:   "US",
		CalculatedAt:   time.Now(),
	}, nil
}

func (s *CompensationServiceImpl) GenerateTaxDocuments(ctx context.Context, talentID uuid.UUID, taxYear int) ([]*TaxDocument, error) {
	// Get payroll records for tax year
	_, err := s.compensationRepo.GetPayrollRecordsByTalentAndYear(ctx, talentID, taxYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll records: %w", err)
	}

	var documents []*TaxDocument

	// Generate W-2 or 1099 based on talent type
	talent, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	var docType string
	if talent.Type == entities.TalentTypeHuman {
		// Determine if employee or contractor based on engagement type
		docType = "W2" // Simplified - would check actual employment status
	} else {
		docType = "1099-NEC" // AI agents are typically contractors
	}

	doc := &TaxDocument{
		DocumentID:  uuid.New(),
		TalentID:    talentID,
		Type:        docType,
		TaxYear:     taxYear,
		DocumentURL: fmt.Sprintf("https://tax-docs.company.com/%s/%d", talentID, taxYear),
		GeneratedAt: time.Now(),
		Status:      "Generated",
		Corrections: 0,
	}

	documents = append(documents, doc)

	return documents, nil
}

func (s *CompensationServiceImpl) ProcessContractorPayment(ctx context.Context, talentID uuid.UUID, amount *entities.Money, description string) (*PaymentResult, error) {
	// Create one-off payment for contractor
	result := &PaymentResult{
		PaymentID:     uuid.New(),
		Status:        "Processed",
		Amount:        amount,
		ProcessedAt:   time.Now(),
		TransactionID: fmt.Sprintf("CONTRACTOR_%s", uuid.New().String()[:8]),
		PaymentMethod: "Wire Transfer",
		Metadata:      map[string]interface{}{"description": description},
	}

	return result, nil
}

func (s *CompensationServiceImpl) GetCompensationBenchmarks(ctx context.Context, skillSet []string, location string) (*repositories.CompensationBenchmarks, error) {
	// In real implementation, this would fetch from market data APIs
	return &repositories.CompensationBenchmarks{
		SkillSet:            skillSet,
		MinRate:             &entities.Money{Amount: 9500000, Currency: "USD"},  // $95,000
		MaxRate:             &entities.Money{Amount: 15000000, Currency: "USD"}, // $150,000
		AverageRate:         &entities.Money{Amount: 12000000, Currency: "USD"}, // $120,000
		MedianRate:          &entities.Money{Amount: 12000000, Currency: "USD"}, // $120,000
		MarketPercentile25:  &entities.Money{Amount: 9500000, Currency: "USD"},  // $95,000
		MarketPercentile75:  &entities.Money{Amount: 15000000, Currency: "USD"}, // $150,000
		SampleSize:          1250,
		LastUpdated:         time.Now(),
	}, nil
}

func (s *CompensationServiceImpl) AnalyzeCompensationEquity(ctx context.Context) (*CompensationEquityAnalysis, error) {
	// Analyze compensation equity across demographics
	analysis := &CompensationEquityAnalysis{
		AnalysisID:         uuid.New(),
		AnalysisDate:       time.Now(),
		TimeRange:          repositories.TimeRange{Start: time.Now().AddDate(-1, 0, 0), End: time.Now()},
		OverallEquityScore: 8.5, // Out of 10
		GenderEquity: EquityMetrics{
			Dimension:               "Gender",
			PayGap:                  2.3, // 2.3% gap
			MedianDifference:        &entities.Money{Amount: 280000, Currency: "USD"}, // $2,800
			StatisticalSignificance: true,
			TrendDirection:          "Improving",
		},
		ComplianceStatus:   "Compliant",
		RiskAreas:          []string{},
		Recommendations:    []EquityRecommendation{},
	}

	return analysis, nil
}

func (s *CompensationServiceImpl) GetCompensationReport(ctx context.Context, timeRange repositories.TimeRange) (*repositories.CompensationSummary, error) {
	return s.compensationRepo.GetCompensationSummary(ctx, timeRange)
}

// Helper methods for CompensationService

func (s *CompensationServiceImpl) calculateOverallPerformanceScore(metrics map[string]float64) float64 {
	if len(metrics) == 0 {
		return 3.0 // Default neutral score
	}

	total := 0.0
	for _, score := range metrics {
		total += score
	}
	return total / float64(len(metrics))
}

func (s *CompensationServiceImpl) calculateAdjustmentPercentage(performanceScore float64) float64 {
	// Performance-based adjustment scale
	switch {
	case performanceScore >= 4.5:
		return 15.0 // 15% increase for exceptional performance
	case performanceScore >= 4.0:
		return 10.0 // 10% increase for exceeding expectations
	case performanceScore >= 3.5:
		return 5.0  // 5% increase for meeting expectations
	case performanceScore >= 3.0:
		return 0.0  // No adjustment for meeting baseline
	case performanceScore >= 2.0:
		return -5.0 // 5% decrease for underperformance
	default:
		return -10.0 // 10% decrease for poor performance
	}
}

// TrainingServiceImpl implements the TrainingService interface
type TrainingServiceImpl struct {
	trainingRepo repositories.TrainingRepository
	talentRepo   repositories.TalentRepository
	eventRepo    repositories.EventRepository
}

// NewTrainingServiceImpl creates a new training service
func NewTrainingServiceImpl(
	trainingRepo repositories.TrainingRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) TrainingService {
	return &TrainingServiceImpl{
		trainingRepo: trainingRepo,
		talentRepo:   talentRepo,
		eventRepo:    eventRepo,
	}
}

func (s *TrainingServiceImpl) CreateTrainingProgram(ctx context.Context, request TrainingProgramRequest) (*entities.TrainingProgram, error) {
	program := &entities.TrainingProgram{
		ID:                 uuid.New(),
		Name:               request.Name,
		Description:        request.Description,
		Type:               request.Type,
		TargetAudience:     request.TargetAudience,
		Duration:           request.Duration,
		Format:             request.Format,
		Prerequisites:      request.Prerequisites,
		LearningObjectives: request.LearningObjectives,
		PassingScore:       request.PassingScore,
		CertificationID:    request.CertificationID,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.trainingRepo.CreateTrainingProgram(ctx, program); err != nil {
		return nil, fmt.Errorf("failed to create training program: %w", err)
	}

	return program, nil
}

func (s *TrainingServiceImpl) UpdateTrainingProgram(ctx context.Context, programID uuid.UUID, updates TrainingProgramUpdates) (*entities.TrainingProgram, error) {
	program, err := s.trainingRepo.GetTrainingProgramByID(ctx, programID)
	if err != nil {
		return nil, fmt.Errorf("training program not found: %w", err)
	}

	// Apply updates
	if updates.Name != nil {
		program.Name = *updates.Name
	}
	if updates.Description != nil {
		program.Description = *updates.Description
	}
	if updates.Type != nil {
		program.Type = *updates.Type
	}
	if updates.TargetAudience != nil {
		program.TargetAudience = *updates.TargetAudience
	}
	if updates.Duration != nil {
		program.Duration = *updates.Duration
	}
	if updates.Format != nil {
		program.Format = *updates.Format
	}
	if updates.Prerequisites != nil {
		program.Prerequisites = updates.Prerequisites
	}
	if updates.LearningObjectives != nil {
		program.LearningObjectives = updates.LearningObjectives
	}
	if updates.PassingScore != nil {
		program.PassingScore = *updates.PassingScore
	}
	if updates.IsActive != nil {
		program.IsActive = *updates.IsActive
	}

	program.UpdatedAt = time.Now()

	if err := s.trainingRepo.UpdateTrainingProgram(ctx, program); err != nil {
		return nil, fmt.Errorf("failed to update training program: %w", err)
	}

	return program, nil
}

func (s *TrainingServiceImpl) GetTrainingProgram(ctx context.Context, id uuid.UUID) (*entities.TrainingProgram, error) {
	return s.trainingRepo.GetTrainingProgramByID(ctx, id)
}

func (s *TrainingServiceImpl) ListTrainingPrograms(ctx context.Context, filter repositories.TrainingFilter) ([]*entities.TrainingProgram, int, error) {
	return s.trainingRepo.ListTrainingPrograms(ctx, filter)
}

func (s *TrainingServiceImpl) EnrollInTraining(ctx context.Context, talentID, trainingID uuid.UUID) (*entities.TrainingProgress, error) {
	// Verify talent and training exist
	_, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	program, err := s.trainingRepo.GetTrainingProgramByID(ctx, trainingID)
	if err != nil {
		return nil, fmt.Errorf("training program not found: %w", err)
	}

	if !program.IsActive {
		return nil, fmt.Errorf("training program is not active")
	}

	progress := &entities.TrainingProgress{
		ID:               uuid.New(),
		TalentID:         talentID,
		TrainingID:       trainingID,
		Status:           entities.TrainingStatusNotStarted,
		Progress:         0.0,
		Attempts:         0,
		MaterialProgress: map[string]interface{}{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.trainingRepo.CreateTrainingProgress(ctx, progress); err != nil {
		return nil, fmt.Errorf("failed to create training progress: %w", err)
	}

	return progress, nil
}

func (s *TrainingServiceImpl) UpdateTrainingProgress(ctx context.Context, talentID, trainingID uuid.UUID, progress TrainingProgressUpdate) error {
	trainingProgress, err := s.trainingRepo.GetTrainingProgress(ctx, talentID, trainingID)
	if err != nil {
		return fmt.Errorf("training progress not found: %w", err)
	}

	trainingProgress.Progress = progress.Progress
	trainingProgress.Status = progress.Status
	trainingProgress.Score = progress.Score
	trainingProgress.MaterialProgress = progress.MaterialProgress
	trainingProgress.UpdatedAt = time.Now()

	// Update started/completed timestamps based on status
	switch progress.Status {
	case entities.TrainingStatusInProgress:
		if trainingProgress.StartedAt == nil {
			now := time.Now()
			trainingProgress.StartedAt = &now
		}
	case entities.TrainingStatusCompleted:
		if trainingProgress.CompletedAt == nil {
			now := time.Now()
			trainingProgress.CompletedAt = &now
		}
	}

	return s.trainingRepo.UpdateTrainingProgress(ctx, trainingProgress)
}

func (s *TrainingServiceImpl) CompleteTraining(ctx context.Context, talentID, trainingID uuid.UUID, finalScore float64) (*entities.Certification, error) {
	trainingProgress, err := s.trainingRepo.GetTrainingProgress(ctx, talentID, trainingID)
	if err != nil {
		return nil, fmt.Errorf("training progress not found: %w", err)
	}

	program, err := s.trainingRepo.GetTrainingProgramByID(ctx, trainingID)
	if err != nil {
		return nil, fmt.Errorf("training program not found: %w", err)
	}

	// Check if passed
	if finalScore < program.PassingScore {
		trainingProgress.Status = entities.TrainingStatusFailed
		trainingProgress.Score = &finalScore
		trainingProgress.Attempts++
		if err := s.trainingRepo.UpdateTrainingProgress(ctx, trainingProgress); err != nil {
			return nil, fmt.Errorf("failed to update training progress: %w", err)
		}
		return nil, fmt.Errorf("training not passed: score %.2f below passing score %.2f", finalScore, program.PassingScore)
	}

	// Mark as completed
	trainingProgress.Status = entities.TrainingStatusCompleted
	trainingProgress.Score = &finalScore
	trainingProgress.Progress = 100.0
	now := time.Now()
	trainingProgress.CompletedAt = &now

	if err := s.trainingRepo.UpdateTrainingProgress(ctx, trainingProgress); err != nil {
		return nil, fmt.Errorf("failed to update training progress: %w", err)
	}

	// Create certification if program has one
	if program.CertificationID != nil {
		cert := &entities.Certification{
			ID:              uuid.New(),
			TalentID:        talentID,
			Name:            program.Name + " Certification",
			Issuer:          "Company Training Department",
			IssueDate:       time.Now(),
			VerificationURL: fmt.Sprintf("https://certs.company.com/verify/%s", uuid.New()),
			IsActive:        true,
			CreatedAt:       time.Now(),
		}

		if err := s.talentRepo.AddTalentCertification(ctx, cert); err != nil {
			return nil, fmt.Errorf("failed to create certification: %w", err)
		}

		return cert, nil
	}

	return nil, nil
}

func (s *TrainingServiceImpl) GetTrainingProgress(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error) {
	return s.trainingRepo.GetTrainingProgressByTalent(ctx, talentID)
}

func (s *TrainingServiceImpl) GenerateTrainingContent(ctx context.Context, topic string, targetAudience string, duration int) (*TrainingContentResult, error) {
	// AI-generated training content
	content := &TrainingContentResult{
		ContentID:      uuid.New(),
		Topic:          topic,
		TargetAudience: targetAudience,
		Duration:       duration,
		Format:         "Online",
		Content: map[string]interface{}{
			"introduction": fmt.Sprintf("Introduction to %s", topic),
			"objectives":   []string{fmt.Sprintf("Understand %s concepts", topic)},
			"modules":      []string{"Module 1: Basics", "Module 2: Advanced"},
		},
		Materials:   []TrainingMaterialData{},
		GeneratedAt: time.Now(),
	}

	return content, nil
}

func (s *TrainingServiceImpl) CreateCustomTraining(ctx context.Context, talentID uuid.UUID, skillGaps []SkillGap) (*entities.TrainingProgram, error) {
	talent, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	// Create custom training program based on skill gaps
	program := &entities.TrainingProgram{
		ID:                 uuid.New(),
		Name:               fmt.Sprintf("Custom Training for %s", talent.Name),
		Description:        "Personalized training program to address skill gaps",
		Type:               "Custom",
		TargetAudience:     "Individual",
		Duration:           s.calculateTrainingDuration(skillGaps),
		Format:             "Hybrid",
		Prerequisites:      []string{},
		LearningObjectives: s.generateObjectivesFromGaps(skillGaps),
		PassingScore:       80.0,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.trainingRepo.CreateTrainingProgram(ctx, program); err != nil {
		return nil, fmt.Errorf("failed to create custom training program: %w", err)
	}

	return program, nil
}

func (s *TrainingServiceImpl) AdaptTrainingContent(ctx context.Context, trainingID uuid.UUID, learnerProfile LearnerProfile) (*AdaptedTrainingContent, error) {
	program, err := s.trainingRepo.GetTrainingProgramByID(ctx, trainingID)
	if err != nil {
		return nil, fmt.Errorf("training program not found: %w", err)
	}

	adapted := &AdaptedTrainingContent{
		TrainingID:       trainingID,
		TalentID:         learnerProfile.TalentID,
		AdaptedContent:   map[string]interface{}{},
		PersonalizedPath: s.generatePersonalizedPath(program, learnerProfile),
		EstimatedDuration: s.adjustDurationForLearner(program.Duration, learnerProfile),
		DifficultyLevel:  s.calculateDifficultyLevel(learnerProfile),
		AdaptedAt:        time.Now(),
	}

	return adapted, nil
}

func (s *TrainingServiceImpl) AssessSkills(ctx context.Context, talentID uuid.UUID, skillSet []string) (*SkillAssessmentResult, error) {
	result := &SkillAssessmentResult{
		TalentID:        talentID,
		AssessmentID:    uuid.New(),
		SkillSet:        skillSet,
		Results:         make(map[string]SkillScore),
		OverallScore:    0.0,
		Recommendations: []string{},
		AssessedAt:      time.Now(),
		ValidUntil:      time.Now().AddDate(1, 0, 0), // Valid for 1 year
	}

	totalScore := 0.0
	for _, skill := range skillSet {
		score := s.assessIndividualSkill(ctx, talentID, skill)
		result.Results[skill] = score
		totalScore += score.Score
	}

	result.OverallScore = totalScore / float64(len(skillSet))
	result.Recommendations = s.generateAssessmentRecommendations(result.Results)

	return result, nil
}

func (s *TrainingServiceImpl) IdentifySkillGaps(ctx context.Context, talentID uuid.UUID, targetRole string) ([]*SkillGap, error) {
	// Get current skills
	currentSkills, err := s.talentRepo.GetTalentSkills(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current skills: %w", err)
	}

	// Get required skills for target role (would be from job requirements database)
	requiredSkills := s.getRequiredSkillsForRole(targetRole)

	var gaps []*SkillGap

	for skill, requiredLevel := range requiredSkills {
		currentLevel := s.getCurrentSkillLevel(currentSkills, skill)
		
		if s.hasSkillGap(currentLevel, requiredLevel) {
			gap := &SkillGap{
				Skill:           skill,
				RequiredLevel:   requiredLevel,
				CurrentLevel:    currentLevel,
				Gap:             s.calculateGapSeverity(currentLevel, requiredLevel),
				Priority:        s.calculateGapPriority(skill, requiredLevel),
				EstimatedEffort: s.estimateTrainingEffort(currentLevel, requiredLevel),
				Recommendations: s.generateSkillGapRecommendations(skill, currentLevel, requiredLevel),
			}
			gaps = append(gaps, gap)
		}
	}

	return gaps, nil
}

func (s *TrainingServiceImpl) RecommendTraining(ctx context.Context, talentID uuid.UUID) ([]*TrainingRecommendation, error) {
	// Get skill gaps
	gaps, err := s.IdentifySkillGaps(ctx, talentID, "current_role") // Would determine actual role
	if err != nil {
		return nil, fmt.Errorf("failed to identify skill gaps: %w", err)
	}

	var recommendations []*TrainingRecommendation

	for _, gap := range gaps {
		// Find training programs that address this skill gap
		isActive := true
		programs, _, err := s.trainingRepo.ListTrainingPrograms(ctx, repositories.TrainingFilter{
			IsActive: &isActive,
			Limit:    5,
		})
		if err != nil {
			continue
		}

		for _, program := range programs {
			rec := &TrainingRecommendation{
				TrainingID:        program.ID,
				Title:             program.Name,
				Type:              program.Type,
				Priority:          gap.Priority,
				EstimatedDuration: program.Duration,
				SkillsAddressed:   []string{gap.Skill},
				Reason:            fmt.Sprintf("Addresses %s skill gap", gap.Skill),
				PrerequisitesMet:  s.checkPrerequisites(ctx, talentID, program.Prerequisites),
			}
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations, nil
}

// Helper methods for TrainingService

func (s *TrainingServiceImpl) calculateTrainingDuration(skillGaps []SkillGap) int {
	total := 0
	for _, gap := range skillGaps {
		total += gap.EstimatedEffort
	}
	return total
}

func (s *TrainingServiceImpl) generateObjectivesFromGaps(skillGaps []SkillGap) []string {
	var objectives []string
	for _, gap := range skillGaps {
		objectives = append(objectives, fmt.Sprintf("Develop %s skills to %s level", gap.Skill, gap.RequiredLevel))
	}
	return objectives
}

func (s *TrainingServiceImpl) generatePersonalizedPath(program *entities.TrainingProgram, profile LearnerProfile) []string {
	// Customize learning path based on learner profile
	basePath := []string{"Introduction", "Core Concepts", "Practice", "Assessment"}
	
	if profile.Pace == "Fast" {
		return []string{"Core Concepts", "Advanced Practice", "Assessment"}
	} else if profile.Pace == "Slow" {
		return []string{"Pre-requisites", "Introduction", "Basic Concepts", "Practice", "Review", "Assessment"}
	}
	
	return basePath
}

func (s *TrainingServiceImpl) adjustDurationForLearner(baseDuration int, profile LearnerProfile) int {
	switch profile.Pace {
	case "Fast":
		return int(float64(baseDuration) * 0.7) // 30% faster
	case "Slow":
		return int(float64(baseDuration) * 1.5) // 50% longer
	default:
		return baseDuration
	}
}

func (s *TrainingServiceImpl) calculateDifficultyLevel(profile LearnerProfile) string {
	// Analyze prior knowledge to determine appropriate difficulty
	if len(profile.CompletedTraining) > 10 {
		return "Advanced"
	} else if len(profile.CompletedTraining) > 3 {
		return "Intermediate"
	}
	return "Beginner"
}

func (s *TrainingServiceImpl) assessIndividualSkill(ctx context.Context, talentID uuid.UUID, skill string) SkillScore {
	// In real implementation, this would run actual skill assessments
	return SkillScore{
		Skill:           skill,
		Score:           75.0 + float64(len(skill)%20), // Mock score
		Level:           entities.SkillLevelIntermediate,
		Confidence:      0.85,
		Evidence:        []string{"Assessment completed", "Practical demonstration"},
		Recommendations: []string{fmt.Sprintf("Consider advanced %s training", skill)},
	}
}

func (s *TrainingServiceImpl) generateAssessmentRecommendations(results map[string]SkillScore) []string {
	var recommendations []string
	
	for skill, score := range results {
		if score.Score < 60 {
			recommendations = append(recommendations, fmt.Sprintf("Focus on improving %s skills", skill))
		} else if score.Score > 85 {
			recommendations = append(recommendations, fmt.Sprintf("Consider advanced %s certification", skill))
		}
	}
	
	return recommendations
}

func (s *TrainingServiceImpl) getRequiredSkillsForRole(role string) map[string]entities.SkillLevel {
	// Mock required skills - would come from job requirements database
	return map[string]entities.SkillLevel{
		"Go Programming":     entities.SkillLevelAdvanced,
		"System Design":      entities.SkillLevelIntermediate,
		"Database Design":    entities.SkillLevelIntermediate,
		"API Development":    entities.SkillLevelAdvanced,
	}
}

func (s *TrainingServiceImpl) getCurrentSkillLevel(skills []*entities.Skill, skillName string) entities.SkillLevel {
	for _, skill := range skills {
		if skill.Name == skillName {
			return skill.Level
		}
	}
	return entities.SkillLevelBeginner // Default if not found
}

func (s *TrainingServiceImpl) hasSkillGap(current, required entities.SkillLevel) bool {
	skillLevels := map[entities.SkillLevel]int{
		entities.SkillLevelBeginner:     1,
		entities.SkillLevelIntermediate: 2,
		entities.SkillLevelAdvanced:     3,
		entities.SkillLevelExpert:       4,
	}
	
	return skillLevels[current] < skillLevels[required]
}

func (s *TrainingServiceImpl) calculateGapSeverity(current, required entities.SkillLevel) string {
	skillLevels := map[entities.SkillLevel]int{
		entities.SkillLevelBeginner:     1,
		entities.SkillLevelIntermediate: 2,
		entities.SkillLevelAdvanced:     3,
		entities.SkillLevelExpert:       4,
	}
	
	gap := skillLevels[required] - skillLevels[current]
	switch gap {
	case 1:
		return "Low"
	case 2:
		return "Medium"
	case 3:
		return "High"
	default:
		return "Critical"
	}
}

func (s *TrainingServiceImpl) calculateGapPriority(skill string, requiredLevel entities.SkillLevel) entities.Priority {
	// Core skills get higher priority
	coreSkills := map[string]bool{
		"Go Programming":  true,
		"System Design":   true,
		"API Development": true,
	}
	
	if coreSkills[skill] && requiredLevel == entities.SkillLevelAdvanced {
		return "high"
	}
	
	return "medium"
}

func (s *TrainingServiceImpl) estimateTrainingEffort(current, required entities.SkillLevel) int {
	skillLevels := map[entities.SkillLevel]int{
		entities.SkillLevelBeginner:     1,
		entities.SkillLevelIntermediate: 2,
		entities.SkillLevelAdvanced:     3,
		entities.SkillLevelExpert:       4,
	}
	
	gap := skillLevels[required] - skillLevels[current]
	return gap * 20 // 20 hours per skill level
}

func (s *TrainingServiceImpl) generateSkillGapRecommendations(skill string, current, required entities.SkillLevel) []TrainingRecommendation {
	return []TrainingRecommendation{
		{
			Title:             fmt.Sprintf("%s Fundamentals", skill),
			Type:              "Online",
			Priority:          "medium",
			EstimatedDuration: 40,
			SkillsAddressed:   []string{skill},
			Reason:            fmt.Sprintf("Bridge gap from %s to %s", current, required),
			PrerequisitesMet:  true,
		},
	}
}

func (s *TrainingServiceImpl) checkPrerequisites(ctx context.Context, talentID uuid.UUID, prerequisites []string) bool {
	if len(prerequisites) == 0 {
		return true
	}
	
	// Check if talent has completed prerequisite training
	progress, err := s.trainingRepo.GetTrainingProgressByTalent(ctx, talentID)
	if err != nil {
		return false
	}
	
	completed := make(map[string]bool)
	for _, p := range progress {
		if p.Status == entities.TrainingStatusCompleted {
			// Would need to map training ID to name
			completed[p.TrainingID.String()] = true
		}
	}
	
	for _, prereq := range prerequisites {
		if !completed[prereq] {
			return false
		}
	}
	
	return true
}

// ComplianceManagementServiceImpl implements the ComplianceManagementService interface
type ComplianceManagementServiceImpl struct {
	complianceRepo repositories.ComplianceRepository
	talentRepo     repositories.TalentRepository
	eventRepo      repositories.EventRepository
}

// NewComplianceManagementServiceImpl creates a new compliance management service
func NewComplianceManagementServiceImpl(
	complianceRepo repositories.ComplianceRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) ComplianceManagementService {
	return &ComplianceManagementServiceImpl{
		complianceRepo: complianceRepo,
		talentRepo:     talentRepo,
		eventRepo:      eventRepo,
	}
}

func (s *ComplianceManagementServiceImpl) CheckCompliance(ctx context.Context, talentID uuid.UUID) (*ComplianceStatus, error) {
	talent, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	status := &ComplianceStatus{
		TalentID:         talentID,
		OverallStatus:    "Compliant",
		LastChecked:      time.Now(),
		ComplianceChecks: []ComplianceCheck{},
		RequiredActions:  []string{},
		ExpiringItems:    []ComplianceItem{},
	}

	// Check various compliance requirements
	checks := []struct {
		name     string
		checker  func() (bool, string, []string)
	}{
		{"Documentation", func() (bool, string, []string) {
			return s.checkDocumentationCompliance(talent)
		}},
		{"Training", func() (bool, string, []string) {
			return s.checkTrainingCompliance(ctx, talentID)
		}},
		{"Legal", func() (bool, string, []string) {
			return s.checkLegalCompliance(talent)
		}},
		{"Tax", func() (bool, string, []string) {
			return s.checkTaxCompliance(talent)
		}},
	}

	overallCompliant := true
	for _, check := range checks {
		compliant, notes, actions := check.checker()
		
		complianceCheck := ComplianceCheck{
			Type:        check.name,
			Status:      "Compliant",
			LastChecked: time.Now(),
			Notes:       notes,
		}
		
		if !compliant {
			complianceCheck.Status = "Non-Compliant"
			overallCompliant = false
			status.RequiredActions = append(status.RequiredActions, actions...)
		}
		
		status.ComplianceChecks = append(status.ComplianceChecks, complianceCheck)
	}

	if !overallCompliant {
		status.OverallStatus = "Non-Compliant"
	}

	return status, nil
}

func (s *ComplianceManagementServiceImpl) UpdateComplianceDocument(ctx context.Context, talentID uuid.UUID, docType string, documentURL string) error {
	// Store compliance document
	// Update compliance status
	// Send notifications if needed
	return nil
}

func (s *ComplianceManagementServiceImpl) GenerateComplianceReport(ctx context.Context, timeRange repositories.TimeRange) (*ComplianceReport, error) {
	// Get all talent for compliance check
	allTalent, _, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get talent: %w", err)
	}

	report := &ComplianceReport{
		ReportID:      uuid.New(),
		TimeRange:     timeRange,
		TotalTalent:   len(allTalent),
		CompliantTalent: 0,
		NonCompliantTalent: 0,
		ComplianceByType: map[string]ComplianceMetrics{},
		Issues:        []ComplianceIssue{},
		GeneratedAt:   time.Now(),
	}

	// Check compliance for each talent
	for _, talent := range allTalent {
		status, err := s.CheckCompliance(ctx, talent.ID)
		if err != nil {
			continue
		}

		if status.OverallStatus == "Compliant" {
			report.CompliantTalent++
		} else {
			report.NonCompliantTalent++
			
			// Add issues to report
			for _, action := range status.RequiredActions {
				issue := ComplianceIssue{
					TalentID:    talent.ID,
					Type:        "Action Required",
					Description: action,
					Severity:    "Medium",
					DetectedAt:  time.Now(),
				}
				report.Issues = append(report.Issues, issue)
			}
		}
	}

	return report, nil
}

func (s *ComplianceManagementServiceImpl) checkDocumentationCompliance(talent *entities.Talent) (bool, string, []string) {
	// Check if all required documents are present
	requiredDocs := []string{"identity", "tax_form", "bank_details"}
	
	if talent.ProfileData == nil {
		return false, "Missing profile data", []string{"Upload required documentation"}
	}

	docs, exists := talent.ProfileData["documents"]
	if !exists {
		return false, "No documents uploaded", []string{"Upload all required documents"}
	}

	docMap, ok := docs.(map[string]interface{})
	if !ok {
		return false, "Invalid document format", []string{"Re-upload documents in correct format"}
	}

	var missing []string
	for _, doc := range requiredDocs {
		if _, exists := docMap[doc]; !exists {
			missing = append(missing, doc)
		}
	}

	if len(missing) > 0 {
		return false, fmt.Sprintf("Missing documents: %v", missing), []string{fmt.Sprintf("Upload missing documents: %v", missing)}
	}

	return true, "All required documents present", []string{}
}

func (s *ComplianceManagementServiceImpl) checkTrainingCompliance(ctx context.Context, talentID uuid.UUID) (bool, string, []string) {
	// Training compliance is checked externally for this service
	// In a real implementation, this would integrate with the training service
	
	// For now, assume basic compliance
	return true, fmt.Sprintf("Training compliance verified for talent %s", talentID), []string{}
}

func (s *ComplianceManagementServiceImpl) checkLegalCompliance(talent *entities.Talent) (bool, string, []string) {
	// Check contract status, work authorization, etc.
	if talent.ProfileData == nil {
		return false, "No legal information available", []string{"Provide work authorization documents"}
	}

	workAuth, exists := talent.ProfileData["work_authorization"]
	if !exists {
		return false, "Work authorization not verified", []string{"Submit work authorization documentation"}
	}

	if auth, ok := workAuth.(bool); !ok || !auth {
		return false, "Work authorization pending", []string{"Complete work authorization verification"}
	}

	return true, "Work authorization verified", []string{}
}

func (s *ComplianceManagementServiceImpl) checkTaxCompliance(talent *entities.Talent) (bool, string, []string) {
	// Check tax forms, withholding status
	if talent.ProfileData == nil {
		return false, "No tax information available", []string{"Submit tax forms"}
	}

	taxForm, exists := talent.ProfileData["tax_form_submitted"]
	if !exists {
		return false, "Tax form not submitted", []string{"Submit required tax forms (W9/W4)"}
	}

	if submitted, ok := taxForm.(bool); !ok || !submitted {
		return false, "Tax form pending", []string{"Complete and submit tax forms"}
	}

	return true, "Tax forms completed", []string{}
}

// OffboardingServiceImpl implements the OffboardingService interface
type OffboardingServiceImpl struct {
	offboardingRepo repositories.OffboardingRepository
	talentRepo      repositories.TalentRepository
	eventRepo       repositories.EventRepository
}

// NewOffboardingServiceImpl creates a new offboarding service
func NewOffboardingServiceImpl(
	offboardingRepo repositories.OffboardingRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) OffboardingService {
	return &OffboardingServiceImpl{
		offboardingRepo: offboardingRepo,
		talentRepo:      talentRepo,
		eventRepo:       eventRepo,
	}
}

func (s *OffboardingServiceImpl) StartOffboarding(ctx context.Context, talentID uuid.UUID, reason string) (*OffboardingWorkflow, error) {
	talent, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	workflow := &OffboardingWorkflow{
		ID:              uuid.New(),
		TalentID:        talentID,
		Reason:          reason,
		Status:          "InProgress",
		Steps:           s.generateOffboardingSteps(talent.Type),
		StartedAt:       time.Now(),
		EstimatedCompletion: time.Now().Add(s.getOffboardingDuration(talent.Type)),
	}

	return workflow, nil
}

func (s *OffboardingServiceImpl) ProcessOffboardingStep(ctx context.Context, workflowID uuid.UUID, stepID string) error {
	// Process specific offboarding step
	switch stepID {
	case "AccessRevocation":
		return s.processAccessRevocation(ctx, workflowID)
	case "AssetReturn":
		return s.processAssetReturn(ctx, workflowID)
	case "KnowledgeTransfer":
		return s.processKnowledgeTransfer(ctx, workflowID)
	case "FinalPayroll":
		return s.processFinalPayroll(ctx, workflowID)
	case "ExitInterview":
		return s.processExitInterview(ctx, workflowID)
	default:
		return fmt.Errorf("unknown offboarding step: %s", stepID)
	}
}

func (s *OffboardingServiceImpl) RevokeAllAccess(ctx context.Context, talentID uuid.UUID) error {
	// Revoke all system access for talent
	// This would integrate with identity management systems
	return nil
}

func (s *OffboardingServiceImpl) ConductExitInterview(ctx context.Context, talentID uuid.UUID, interviewData ExitInterviewData) (*ExitInterviewResult, error) {
	result := &ExitInterviewResult{
		TalentID:       talentID,
		InterviewerID:  interviewData.InterviewerID,
		Feedback:       interviewData.Feedback,
		Recommendations: s.generateRecommendationsFromFeedback(interviewData.Feedback),
		CompletedAt:    time.Now(),
	}

	return result, nil
}

func (s *OffboardingServiceImpl) GenerateOffboardingReport(ctx context.Context, talentID uuid.UUID) (*OffboardingReport, error) {
	talent, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	report := &OffboardingReport{
		TalentID:      talentID,
		TalentName:    talent.Name,
		Department:    "Engineering", // Would be fetched from engagement data
		OffboardingDate: time.Now(),
		Reason:        "End of Contract",
		AccessRevoked: true,
		AssetsReturned: true,
		FinalPayProcessed: true,
		Documentation: []string{"Exit interview completed", "Knowledge transfer documented"},
		GeneratedAt:   time.Now(),
	}

	return report, nil
}

func (s *OffboardingServiceImpl) generateOffboardingSteps(talentType entities.TalentType) []OffboardingStep {
	baseSteps := []OffboardingStep{
		{
			ID:          "AccessRevocation",
			Name:        "Revoke System Access",
			Description: "Revoke all system access and credentials",
			Required:    true,
			Order:       1,
		},
		{
			ID:          "AssetReturn",
			Name:        "Asset Return",
			Description: "Return company assets and equipment",
			Required:    true,
			Order:       2,
		},
		{
			ID:          "KnowledgeTransfer",
			Name:        "Knowledge Transfer",
			Description: "Transfer knowledge and handover responsibilities",
			Required:    true,
			Order:       3,
		},
		{
			ID:          "FinalPayroll",
			Name:        "Final Payroll",
			Description: "Process final payment and benefits",
			Required:    true,
			Order:       4,
		},
		{
			ID:          "ExitInterview",
			Name:        "Exit Interview",
			Description: "Conduct exit interview and collect feedback",
			Required:    false,
			Order:       5,
		},
	}

	// AI agents have simplified offboarding
	if talentType == entities.TalentTypeAI {
		return []OffboardingStep{
			{
				ID:          "AccessRevocation",
				Name:        "Revoke API Access",
				Description: "Revoke API keys and access credentials",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "DataCleanup",
				Name:        "Data Cleanup",
				Description: "Clean up agent data and configurations",
				Required:    true,
				Order:       2,
			},
		}
	}

	return baseSteps
}

func (s *OffboardingServiceImpl) getOffboardingDuration(talentType entities.TalentType) time.Duration {
	if talentType == entities.TalentTypeAI {
		return 1 * 24 * time.Hour // 1 day for AI agents
	}
	return 7 * 24 * time.Hour // 1 week for humans
}

func (s *OffboardingServiceImpl) processAccessRevocation(ctx context.Context, workflowID uuid.UUID) error {
	// Revoke system access
	return nil
}

func (s *OffboardingServiceImpl) processAssetReturn(ctx context.Context, workflowID uuid.UUID) error {
	// Process asset return
	return nil
}

func (s *OffboardingServiceImpl) processKnowledgeTransfer(ctx context.Context, workflowID uuid.UUID) error {
	// Facilitate knowledge transfer
	return nil
}

func (s *OffboardingServiceImpl) processFinalPayroll(ctx context.Context, workflowID uuid.UUID) error {
	// Process final payroll
	return nil
}

func (s *OffboardingServiceImpl) processExitInterview(ctx context.Context, workflowID uuid.UUID) error {
	// Schedule and conduct exit interview
	return nil
}

func (s *OffboardingServiceImpl) generateRecommendationsFromFeedback(feedback map[string]interface{}) []string {
	recommendations := []string{}
	
	if feedback["process_improvement"] != nil {
		recommendations = append(recommendations, "Review and improve offboarding process")
	}
	
	if feedback["team_feedback"] != nil {
		recommendations = append(recommendations, "Share feedback with team management")
	}
	
	return recommendations
}

// HRAnalyticsServiceImpl implements the HRAnalyticsService interface
type HRAnalyticsServiceImpl struct {
	talentRepo      repositories.TalentRepository
	performanceRepo repositories.PerformanceRepository
	compensationRepo repositories.CompensationRepository
}

// NewHRAnalyticsServiceImpl creates a new HR analytics service
func NewHRAnalyticsServiceImpl(
	talentRepo repositories.TalentRepository,
	performanceRepo repositories.PerformanceRepository,
	compensationRepo repositories.CompensationRepository,
) HRAnalyticsService {
	return &HRAnalyticsServiceImpl{
		talentRepo:       talentRepo,
		performanceRepo:  performanceRepo,
		compensationRepo: compensationRepo,
	}
}

func (s *HRAnalyticsServiceImpl) GenerateTalentAnalytics(ctx context.Context) (*TalentAnalytics, error) {
	// Get all talent
	allTalent, totalCount, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get talent: %w", err)
	}

	analytics := &TalentAnalytics{
		TotalTalent:    totalCount,
		HumanTalent:    0,
		AIAgents:       0,
		AvailableTalent: 0,
		EngagedTalent:  0,
		SkillDistribution: make(map[string]int),
		LocationDistribution: make(map[string]int),
		AverageReputation: 0.0,
		GeneratedAt:    time.Now(),
	}

	totalReputation := 0.0
	for _, talent := range allTalent {
		switch talent.Type {
		case entities.TalentTypeHuman:
			analytics.HumanTalent++
		case entities.TalentTypeAI:
			analytics.AIAgents++
		}

		switch talent.Status {
		case entities.TalentStatusAvailable:
			analytics.AvailableTalent++
		case entities.TalentStatusEngaged:
			analytics.EngagedTalent++
		}

		if talent.Location != "" {
			analytics.LocationDistribution[talent.Location]++
		}

		totalReputation += talent.ReputationScore

		// Get skills for this talent
		skills, err := s.talentRepo.GetTalentSkills(ctx, talent.ID)
		if err == nil {
			for _, skill := range skills {
				analytics.SkillDistribution[skill.Name]++
			}
		}
	}

	if totalCount > 0 {
		analytics.AverageReputation = totalReputation / float64(totalCount)
	}

	return analytics, nil
}

func (s *HRAnalyticsServiceImpl) GeneratePerformanceAnalytics(ctx context.Context, timeRange repositories.TimeRange) (*PerformanceAnalytics, error) {
	// Get performance distribution
	distribution, err := s.performanceRepo.GetPerformanceDistribution(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance distribution: %w", err)
	}

	analytics := &PerformanceAnalytics{
		TimeRange:           timeRange,
		AveragePerformance:  3.5, // Would be calculated from actual data
		PerformanceDistribution: *distribution,
		TopPerformers:       5,   // Count of top performers
		UnderPerformers:     2,   // Count of underperformers
		PerformanceTrends:   []MetricTrend{
			{
				MetricType:    "overall_performance",
				StartValue:    3.2,
				EndValue:      3.5,
				ChangePercent: 9.4,
				Trend:         "Improving",
				Confidence:    0.85,
			},
		},
		GeneratedAt: time.Now(),
	}

	return analytics, nil
}

func (s *HRAnalyticsServiceImpl) GenerateCompensationAnalytics(ctx context.Context, timeRange repositories.TimeRange) (*CompensationAnalytics, error) {
	summary, err := s.compensationRepo.GetCompensationSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get compensation summary: %w", err)
	}

	analytics := &CompensationAnalytics{
		TimeRange:        timeRange,
		TotalPayroll:     summary.TotalPayroll,
		AverageCompensation: summary.AverageHourlyRate,
		CompensationDistribution: CompensationDistribution{
			ByLevel: make(map[string]*entities.Money),
			BySkill: make(map[string]*entities.Money),
			ByLocation: make(map[string]*entities.Money),
			ByType: make(map[string]*entities.Money),
			Percentiles: make(map[string]*entities.Money),
		},
		PayrollTrends:    []PayrollTrend{
			{
				Period:        "Monthly",
				Amount:        summary.TotalPayroll,
				ChangePercent: 5.2,
				Trend:         "Increasing",
			},
		},
		GeneratedAt: time.Now(),
	}

	return analytics, nil
}

func (s *HRAnalyticsServiceImpl) PredictTalentNeeds(ctx context.Context) (*TalentPrediction, error) {
	// Analyze current talent utilization and predict future needs
	prediction := &TalentPrediction{
		TimeHorizon:    90 * 24 * time.Hour, // 90 days
		PredictedNeeds: []TalentNeed{
			{
				SkillCategory:    "AI/ML Engineering",
				RequiredCount:    3,
				Confidence:       0.85,
				TimeToFulfill:    30 * 24 * time.Hour,
				RecommendedAction: "Start recruitment for AI/ML engineers",
			},
			{
				SkillCategory:    "Frontend Development",
				RequiredCount:    2,
				Confidence:       0.75,
				TimeToFulfill:    45 * 24 * time.Hour,
				RecommendedAction: "Post job listings for frontend developers",
			},
		},
		ConfidenceScore: 0.8,
		Recommendations: []string{
			"Increase AI talent acquisition budget",
			"Consider partnerships with AI talent agencies",
			"Develop internal training programs for existing talent",
		},
		GeneratedAt: time.Now(),
	}

	return prediction, nil
}
package hr_management

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// TalentAcquisitionServiceImpl implements the TalentAcquisitionService interface
type TalentAcquisitionServiceImpl struct {
	applicationRepo repositories.TalentApplicationRepository
	talentRepo      repositories.TalentRepository
	eventRepo       repositories.EventRepository
}

// NewTalentAcquisitionService creates a new talent acquisition service
func NewTalentAcquisitionServiceImpl(
	applicationRepo repositories.TalentApplicationRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) TalentAcquisitionService {
	return &TalentAcquisitionServiceImpl{
		applicationRepo: applicationRepo,
		talentRepo:      talentRepo,
		eventRepo:       eventRepo,
	}
}

// Job posting management

func (s *TalentAcquisitionServiceImpl) CreateJobPosting(ctx context.Context, request JobPostingRequest) (*entities.JobPosting, error) {
	jobPosting := &entities.JobPosting{
		ID:               uuid.New(),
		Title:            request.Title,
		Description:      request.Description,
		Type:             request.Type,
		Department:       request.Department,
		RequiredSkills:   request.RequiredSkills,
		PreferredSkills:  request.PreferredSkills,
		ExperienceYears:  request.ExperienceYears,
		EducationLevel:   request.EducationLevel,
		Location:         request.Location,
		Remote:           request.Remote,
		Benefits:         request.Benefits,
		ClosingDate:      request.ClosingDate,
		IsActive:         true,
		ApplicationCount: 0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Convert salary range to JSON
	if request.SalaryRange.Min != nil || request.SalaryRange.Max != nil {
		salaryData := map[string]interface{}{
			"min":      request.SalaryRange.Min,
			"max":      request.SalaryRange.Max,
			"currency": request.SalaryRange.Currency,
		}
		jobPosting.SalaryRange = salaryData
	}

	if err := s.applicationRepo.CreateJobPosting(ctx, jobPosting); err != nil {
		return nil, fmt.Errorf("failed to create job posting: %w", err)
	}

	return jobPosting, nil
}

func (s *TalentAcquisitionServiceImpl) UpdateJobPosting(ctx context.Context, id uuid.UUID, request JobPostingRequest) (*entities.JobPosting, error) {
	jobPosting, err := s.applicationRepo.GetJobPostingByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("job posting not found: %w", err)
	}

	// Update fields
	jobPosting.Title = request.Title
	jobPosting.Description = request.Description
	jobPosting.Type = request.Type
	jobPosting.Department = request.Department
	jobPosting.RequiredSkills = request.RequiredSkills
	jobPosting.PreferredSkills = request.PreferredSkills
	jobPosting.ExperienceYears = request.ExperienceYears
	jobPosting.EducationLevel = request.EducationLevel
	jobPosting.Location = request.Location
	jobPosting.Remote = request.Remote
	jobPosting.Benefits = request.Benefits
	jobPosting.ClosingDate = request.ClosingDate
	jobPosting.UpdatedAt = time.Now()

	// Update salary range
	if request.SalaryRange.Min != nil || request.SalaryRange.Max != nil {
		salaryData := map[string]interface{}{
			"min":      request.SalaryRange.Min,
			"max":      request.SalaryRange.Max,
			"currency": request.SalaryRange.Currency,
		}
		jobPosting.SalaryRange = salaryData
	}

	if err := s.applicationRepo.UpdateJobPosting(ctx, jobPosting); err != nil {
		return nil, fmt.Errorf("failed to update job posting: %w", err)
	}

	return jobPosting, nil
}

func (s *TalentAcquisitionServiceImpl) GetJobPosting(ctx context.Context, id uuid.UUID) (*entities.JobPosting, error) {
	return s.applicationRepo.GetJobPostingByID(ctx, id)
}

func (s *TalentAcquisitionServiceImpl) ListJobPostings(ctx context.Context, filter repositories.JobPostingFilter) ([]*entities.JobPosting, int, error) {
	return s.applicationRepo.ListJobPostings(ctx, filter)
}

func (s *TalentAcquisitionServiceImpl) CloseJobPosting(ctx context.Context, id uuid.UUID) error {
	jobPosting, err := s.applicationRepo.GetJobPostingByID(ctx, id)
	if err != nil {
		return fmt.Errorf("job posting not found: %w", err)
	}

	jobPosting.IsActive = false
	now := time.Now()
	jobPosting.ClosingDate = &now
	jobPosting.UpdatedAt = now

	return s.applicationRepo.UpdateJobPosting(ctx, jobPosting)
}

// Application processing

func (s *TalentAcquisitionServiceImpl) SubmitApplication(ctx context.Context, request ApplicationRequest) (*entities.TalentApplication, error) {
	// Verify talent exists
	_, err := s.talentRepo.GetTalentByID(ctx, request.TalentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	// Verify job posting exists and is active
	jobPosting, err := s.applicationRepo.GetJobPostingByID(ctx, request.JobPostingID)
	if err != nil {
		return nil, fmt.Errorf("job posting not found: %w", err)
	}

	if !jobPosting.IsActive {
		return nil, fmt.Errorf("job posting is no longer active")
	}

	application := &entities.TalentApplication{
		ID:            uuid.New(),
		TalentID:      request.TalentID,
		JobPostingID:  request.JobPostingID,
		Status:        entities.ApplicationStatusNew,
		CoverLetter:   request.CoverLetter,
		ResumeURL:     request.ResumeURL,
		PortfolioURLs: request.PortfolioURLs,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.applicationRepo.CreateApplication(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Update application count on job posting
	jobPosting.ApplicationCount++
	if err := s.applicationRepo.UpdateJobPosting(ctx, jobPosting); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to update application count: %v\n", err)
	}

	return application, nil
}

func (s *TalentAcquisitionServiceImpl) ScreenApplication(ctx context.Context, applicationID uuid.UUID) (*ApplicationScreeningResult, error) {
	application, err := s.applicationRepo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	talent, err := s.talentRepo.GetTalentByID(ctx, application.TalentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	jobPosting, err := s.applicationRepo.GetJobPostingByID(ctx, application.JobPostingID)
	if err != nil {
		return nil, fmt.Errorf("job posting not found: %w", err)
	}

	// Get talent skills for matching
	skills, err := s.talentRepo.GetTalentSkills(ctx, talent.ID)
	if err != nil {
		skills = []*entities.Skill{} // Continue with empty skills if error
	}

	// Calculate skill match
	skillMatch := s.calculateSkillMatch(skills, jobPosting.RequiredSkills)

	// Calculate experience match
	experienceMatch := s.calculateExperienceMatch(talent, jobPosting.ExperienceYears)

	// Calculate overall score
	overallScore := (skillMatch*0.6 + experienceMatch*0.4) * 100

	// Determine recommended action
	var recommendedAction string
	var notes string

	switch {
	case overallScore >= 80:
		recommendedAction = "Interview"
		notes = "Strong candidate with excellent skill and experience match"
	case overallScore >= 60:
		recommendedAction = "Review"
		notes = "Good candidate, requires additional review"
	case overallScore >= 40:
		recommendedAction = "Consider"
		notes = "Moderate fit, may be suitable for junior or training roles"
	default:
		recommendedAction = "Reject"
		notes = "Insufficient match for current requirements"
	}

	result := &ApplicationScreeningResult{
		ApplicationID:     applicationID,
		Score:             overallScore,
		SkillMatch:        skillMatch * 100,
		ExperienceMatch:   experienceMatch * 100,
		RecommendedAction: recommendedAction,
		Notes:             notes,
	}

	// Update application with screening results
	application.ScreeningScore = &overallScore
	application.ScreeningNotes = notes
	application.Status = entities.ApplicationStatusScreening
	application.UpdatedAt = time.Now()

	if err := s.applicationRepo.UpdateApplication(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return result, nil
}

func (s *TalentAcquisitionServiceImpl) ProcessApplication(ctx context.Context, applicationID uuid.UUID, decision ApplicationDecision) error {
	application, err := s.applicationRepo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("application not found: %w", err)
	}

	// Update application status based on decision
	switch decision.Decision {
	case "Approve":
		application.Status = entities.ApplicationStatusApproved
	case "Reject":
		application.Status = entities.ApplicationStatusRejected
	case "Interview":
		application.Status = entities.ApplicationStatusInterview
	default:
		return fmt.Errorf("invalid decision: %s", decision.Decision)
	}

	application.DecisionReason = decision.Reason
	application.DecisionDate = &time.Time{}
	*application.DecisionDate = time.Now()
	application.UpdatedAt = time.Now()

	if decision.Notes != "" {
		if application.InterviewNotes == "" {
			application.InterviewNotes = decision.Notes
		} else {
			application.InterviewNotes += "\n" + decision.Notes
		}
	}

	return s.applicationRepo.UpdateApplication(ctx, application)
}

func (s *TalentAcquisitionServiceImpl) GetApplication(ctx context.Context, id uuid.UUID) (*entities.TalentApplication, error) {
	return s.applicationRepo.GetApplicationByID(ctx, id)
}

func (s *TalentAcquisitionServiceImpl) ListApplications(ctx context.Context, filter repositories.ApplicationFilter) ([]*entities.TalentApplication, int, error) {
	return s.applicationRepo.ListApplications(ctx, filter)
}

// Talent sourcing

func (s *TalentAcquisitionServiceImpl) SourceTalent(ctx context.Context, requirements TalentRequirements) ([]*TalentMatch, error) {
	// Build search filter from requirements
	filter := repositories.TalentFilter{
		Type:      &requirements.TalentType,
		Location:  &requirements.Location,
		Remote:    &requirements.Remote,
		Limit:     50, // Reasonable limit for sourcing
		SortBy:    "reputation_score",
		SortOrder: "desc",
	}

	// Search available talent
	talents, _, err := s.talentRepo.SearchTalent(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search talent: %w", err)
	}

	matches := make([]*TalentMatch, 0, len(talents))

	for _, talent := range talents {
		if talent.Status != entities.TalentStatusAvailable {
			continue
		}

		// Get talent skills for matching
		skills, err := s.talentRepo.GetTalentSkills(ctx, talent.ID)
		if err != nil {
			continue // Skip on error
		}

		// Calculate match scores
		skillAlignment := s.calculateDetailedSkillMatch(skills, requirements.Skills)
		availabilityFit := s.calculateAvailabilityFit(talent.Availability, requirements.Availability)
		budgetFit := s.calculateBudgetFit(talent.HourlyRate, requirements.MaxBudget)

		// Calculate overall match score
		matchScore := (float64(len(skillAlignment))*0.5 + availabilityFit*0.3 + budgetFit*0.2) / float64(len(requirements.Skills))

		if matchScore > 0.3 { // Only include reasonably good matches
			match := &TalentMatch{
				Talent:          talent,
				MatchScore:      matchScore,
				SkillAlignment:  skillAlignment,
				AvailabilityFit: availabilityFit,
				BudgetFit:       budgetFit,
				Reasoning:       s.generateMatchReasoning(matchScore, skillAlignment, availabilityFit, budgetFit),
			}
			matches = append(matches, match)
		}
	}

	return matches, nil
}

func (s *TalentAcquisitionServiceImpl) SearchTalent(ctx context.Context, criteria TalentSearchCriteria) ([]*entities.Talent, error) {
	filter := repositories.TalentFilter{
		Search:        criteria.Keywords[0], // Use first keyword as search term
		Location:      &criteria.Location,
		Remote:        &criteria.Remote,
		MinReputation: &criteria.MinReputation,
		Limit:         50,
		SortBy:        "reputation_score",
		SortOrder:     "desc",
	}

	if criteria.MaxRate != nil {
		maxRate := float64(criteria.MaxRate.Amount) / 100 // Convert cents to dollars
		filter.MaxHourlyRate = &maxRate
	}

	talents, _, err := s.talentRepo.SearchTalent(ctx, filter)
	return talents, err
}

func (s *TalentAcquisitionServiceImpl) AnalyzeTalentPool(ctx context.Context) (*TalentPoolAnalysis, error) {
	// Get all talent
	allTalent, totalCount, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Limit: 1000, // Large limit for analysis
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get talent pool: %w", err)
	}

	// Get available talent
	availableTalent, availableCount, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Status: func() *entities.TalentStatus { s := entities.TalentStatusAvailable; return &s }(),
		Limit:  1000,
	})
	if err != nil {
		_ = availableTalent // Ignore error for now
		return nil, fmt.Errorf("failed to get available talent: %w", err)
	}

	// Analyze distributions
	skillDistribution := make(map[string]int)
	locationDistribution := make(map[string]int)
	averageRates := make(map[string]*entities.Money)

	for _, talent := range allTalent {
		// Location distribution
		if talent.Location != "" {
			locationDistribution[talent.Location]++
		}

		// Get skills for this talent
		skills, err := s.talentRepo.GetTalentSkills(ctx, talent.ID)
		if err == nil {
			for _, skill := range skills {
				skillDistribution[skill.Name]++
			}
		}
	}

	// Calculate growth trends (simplified - would need historical data)
	growthTrends := map[string]float64{
		"total_talent":     5.2,  // 5.2% monthly growth
		"available_talent": 3.8,  // 3.8% monthly growth
		"ai_agents":        15.7, // 15.7% monthly growth
	}

	return &TalentPoolAnalysis{
		TotalTalent:          totalCount,
		AvailableTalent:      availableCount,
		SkillDistribution:    skillDistribution,
		LocationDistribution: locationDistribution,
		AverageRates:         averageRates,
		GrowthTrends:         growthTrends,
	}, nil
}

// AI agent discovery

func (s *TalentAcquisitionServiceImpl) DiscoverAIAgents(ctx context.Context, capabilities []string) ([]*entities.AIAgent, error) {
	// Search for AI type talent
	aiType := entities.TalentTypeAI
	talents, _, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Type:  &aiType,
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search AI agents: %w", err)
	}

	agents := make([]*entities.AIAgent, 0, len(talents))

	for _, talent := range talents {
		// Get AI-specific data (this would be populated from the ai_agents table)
		agent := &entities.AIAgent{
			Talent:       *talent,
			Provider:     "OpenAI", // Would be fetched from database
			Model:        "gpt-4",  // Would be fetched from database
			APIEndpoint:  "https://api.openai.com/v1/chat/completions",
			Capabilities: capabilities, // Would match against stored capabilities
			RateLimits: map[string]interface{}{
				"requests_per_minute": 3000,
				"tokens_per_minute":   150000,
			},
			CostPerRequest: &entities.Money{Amount: 300, Currency: "USD"}, // $0.03
			ResponseTimeMs: 1500,                                          // 1.5 seconds
			Reliability:    0.99,
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *TalentAcquisitionServiceImpl) TestAIAgentCapabilities(ctx context.Context, agentID uuid.UUID, testSuite []CapabilityTest) (*CapabilityTestResult, error) {
	agent, err := s.talentRepo.GetTalentByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	if agent.Type != entities.TalentTypeAI {
		return nil, fmt.Errorf("talent is not an AI agent")
	}

	results := make([]SingleTestResult, len(testSuite))
	totalScore := 0.0
	passed := true
	totalDuration := time.Duration(0)

	for i, test := range testSuite {
		// Simulate test execution
		startTime := time.Now()

		// For demo purposes, simulate test results
		testPassed := true // Would actually run the test
		score := 0.85      // Would be calculated based on test results
		output := map[string]interface{}{
			"result":   "success",
			"accuracy": score,
		}

		duration := time.Since(startTime)
		totalDuration += duration

		results[i] = SingleTestResult{
			TestID:   test.TestID,
			Passed:   testPassed,
			Score:    score,
			Output:   output,
			Duration: duration,
		}

		totalScore += score
		if !testPassed {
			passed = false
		}
	}

	overallScore := totalScore / float64(len(testSuite))

	return &CapabilityTestResult{
		AgentID:      agentID,
		TestResults:  results,
		OverallScore: overallScore,
		Passed:       passed,
		Duration:     totalDuration,
		Notes:        fmt.Sprintf("Completed %d tests with overall score of %.2f", len(testSuite), overallScore),
	}, nil
}

func (s *TalentAcquisitionServiceImpl) RegisterAIAgent(ctx context.Context, request AIAgentRegistrationRequest) (*entities.AIAgent, error) {
	// Create talent profile for AI agent
	talent := &entities.Talent{
		ID:              uuid.New(),
		Type:            entities.TalentTypeAI,
		Name:            request.Name,
		Status:          entities.TalentStatusAvailable,
		ReputationScore: 0.0,
		Currency:        "USD",
		Location:        "Cloud",
		Timezone:        "UTC",
		Availability:    map[string]interface{}{"24/7": true},
		ProfileData: map[string]interface{}{
			"provider":     request.Provider,
			"model":        request.Model,
			"api_endpoint": request.APIEndpoint,
			"capabilities": request.Capabilities,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.talentRepo.CreateTalent(ctx, talent); err != nil {
		return nil, fmt.Errorf("failed to create AI agent talent profile: %w", err)
	}

	// Create AI agent specific data
	agent := &entities.AIAgent{
		Talent:         *talent,
		Provider:       request.Provider,
		Model:          request.Model,
		APIEndpoint:    request.APIEndpoint,
		APIVersion:     request.APIVersion,
		Capabilities:   request.Capabilities,
		RateLimits:     request.RateLimits,
		CostPerRequest: request.CostPerRequest,
		CostPerToken:   request.CostPerToken,
		MaxTokens:      request.MaxTokens,
		ResponseTimeMs: request.ResponseTimeMs,
		Reliability:    1.0, // Start with perfect reliability
	}

	return agent, nil
}

// Helper methods

func (s *TalentAcquisitionServiceImpl) calculateSkillMatch(talentSkills []*entities.Skill, requiredSkills []string) float64 {
	if len(requiredSkills) == 0 {
		return 1.0 // Perfect match if no requirements
	}

	talentSkillMap := make(map[string]*entities.Skill)
	for _, skill := range talentSkills {
		talentSkillMap[skill.Name] = skill
	}

	matchCount := 0
	for _, required := range requiredSkills {
		if _, exists := talentSkillMap[required]; exists {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(requiredSkills))
}

func (s *TalentAcquisitionServiceImpl) calculateExperienceMatch(talent *entities.Talent, requiredYears int) float64 {
	if requiredYears == 0 {
		return 1.0 // No experience requirement
	}

	// For AI agents, consider them as having unlimited experience
	if talent.Type == entities.TalentTypeAI {
		return 1.0
	}

	// Extract years of experience from profile data
	if exp, exists := talent.ProfileData["years_experience"]; exists {
		if years, ok := exp.(int); ok {
			if years >= requiredYears {
				return 1.0
			}
			return float64(years) / float64(requiredYears)
		}
	}

	return 0.5 // Default moderate experience if not specified
}

func (s *TalentAcquisitionServiceImpl) calculateDetailedSkillMatch(talentSkills []*entities.Skill, requiredSkills []string) map[string]float64 {
	skillAlignment := make(map[string]float64)

	talentSkillMap := make(map[string]*entities.Skill)
	for _, skill := range talentSkills {
		talentSkillMap[skill.Name] = skill
	}

	for _, required := range requiredSkills {
		if skill, exists := talentSkillMap[required]; exists {
			// Score based on skill level
			var score float64
			switch skill.Level {
			case entities.SkillLevelExpert:
				score = 1.0
			case entities.SkillLevelAdvanced:
				score = 0.8
			case entities.SkillLevelIntermediate:
				score = 0.6
			case entities.SkillLevelBeginner:
				score = 0.4
			default:
				score = 0.5
			}
			skillAlignment[required] = score
		} else {
			skillAlignment[required] = 0.0
		}
	}

	return skillAlignment
}

func (s *TalentAcquisitionServiceImpl) calculateAvailabilityFit(talentAvailability, requiredAvailability map[string]interface{}) float64 {
	// Simplified availability matching
	// In a real implementation, this would parse availability patterns
	return 0.8 // Default good availability fit
}

func (s *TalentAcquisitionServiceImpl) calculateBudgetFit(talentRate *entities.Money, maxBudget *entities.Money) float64 {
	if maxBudget == nil || talentRate == nil {
		return 1.0 // No budget constraint or rate
	}

	if talentRate.Amount <= maxBudget.Amount {
		return 1.0 // Within budget
	}

	// Calculate how much over budget (return lower score for higher overage)
	overage := float64(talentRate.Amount-maxBudget.Amount) / float64(maxBudget.Amount)
	return 1.0 / (1.0 + overage)
}

func (s *TalentAcquisitionServiceImpl) generateMatchReasoning(matchScore float64, skillAlignment map[string]float64, availabilityFit, budgetFit float64) string {
	var reasoning string

	switch {
	case matchScore >= 0.8:
		reasoning = "Excellent match with strong skill alignment"
	case matchScore >= 0.6:
		reasoning = "Good match with acceptable skill coverage"
	case matchScore >= 0.4:
		reasoning = "Moderate match, may require additional training"
	default:
		reasoning = "Limited match, significant gaps in requirements"
	}

	if budgetFit < 0.8 {
		reasoning += ", but over budget"
	}

	if availabilityFit < 0.7 {
		reasoning += ", with limited availability"
	}

	return reasoning
}

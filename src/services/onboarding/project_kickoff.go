package onboarding

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ProjectKickoffService handles automatic project creation after onboarding completion
type ProjectKickoffService struct {
	// These would be injected in a real implementation
	projectRepository ProjectRepository
	contentService    ContentService
	pricingService    PricingService
	notificationService NotificationService
}

// ProjectRepository interface for project operations
type ProjectRepository interface {
	CreateProject(ctx context.Context, project *entities.Project) error
	GetProject(ctx context.Context, projectID uuid.UUID) (*entities.Project, error)
	UpdateProject(ctx context.Context, project *entities.Project) error
}

// ContentService interface for content operations
type ContentService interface {
	CreateInitialContent(ctx context.Context, projectID uuid.UUID, profile *entities.ClientProfile) error
	GenerateContentPlan(ctx context.Context, profile *entities.ClientProfile) (*ContentPlan, error)
}

// PricingService interface for pricing operations
type PricingService interface {
	CalculateProjectPricing(ctx context.Context, profile *entities.ClientProfile, plan *ContentPlan) (*ProjectPricing, error)
}

// NotificationService interface for client notifications
type NotificationService interface {
	SendWelcomeEmail(ctx context.Context, clientID uuid.UUID, project *entities.Project) error
	SendProjectStartNotification(ctx context.Context, clientID uuid.UUID, project *entities.Project) error
}

// Data structures for project kickoff

// ContentPlan represents a structured content plan for a client
type ContentPlan struct {
	PlanID          uuid.UUID         `json:"planId"`
	ClientID        uuid.UUID         `json:"clientId"`
	ContentPillars  []ContentPillar   `json:"contentPillars"`
	ContentCalendar []ContentItem     `json:"contentCalendar"`
	Frequency       string            `json:"frequency"`
	Channels        []string          `json:"channels"`
	CreatedAt       time.Time         `json:"createdAt"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ContentPillar represents a key content theme or topic area
type ContentPillar struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Priority    int      `json:"priority"`
	Frequency   string   `json:"frequency"`
}

// ContentItem represents a specific piece of content to be created
type ContentItem struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Pillar      string    `json:"pillar"`
	Channel     string    `json:"channel"`
	ScheduledAt time.Time `json:"scheduledAt"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	WordCount   int       `json:"wordCount"`
	Description string    `json:"description"`
}

// ProjectPricing represents pricing information for a project
type ProjectPricing struct {
	ProjectID       uuid.UUID                  `json:"projectId"`
	BasePricing     map[string]float64         `json:"basePricing"`
	VolumeDiscount  float64                    `json:"volumeDiscount"`
	LoyaltyDiscount float64                    `json:"loyaltyDiscount"`
	TotalMonthly    float64                    `json:"totalMonthly"`
	TotalAnnual     float64                    `json:"totalAnnual"`
	Breakdown       map[string]interface{}     `json:"breakdown"`
	CreatedAt       time.Time                  `json:"createdAt"`
}

// NewProjectKickoffService creates a new project kickoff service
func NewProjectKickoffService(
	projectRepo ProjectRepository,
	contentService ContentService,
	pricingService PricingService,
	notificationService NotificationService,
) *ProjectKickoffService {
	return &ProjectKickoffService{
		projectRepository:   projectRepo,
		contentService:      contentService,
		pricingService:      pricingService,
		notificationService: notificationService,
	}
}

// InitiateProjectFromProfile creates a new project based on completed client profile
func (p *ProjectKickoffService) InitiateProjectFromProfile(ctx context.Context, profile *entities.ClientProfile) (*entities.Project, error) {
	if !profile.OnboardingComplete {
		return nil, fmt.Errorf("client profile onboarding is not complete")
	}
	
	// Generate content plan
	contentPlan, err := p.contentService.GenerateContentPlan(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content plan: %w", err)
	}
	
	// Calculate pricing
	pricing, err := p.pricingService.CalculateProjectPricing(ctx, profile, contentPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pricing: %w", err)
	}
	
	// Create project
	project := &entities.Project{
		ProjectID:    uuid.New(),
		ClientID:     profile.ClientID,
		Title:        p.generateProjectTitle(profile),
		Description:  p.generateProjectDescription(profile, contentPlan),
		Status:       "planning",
		Priority:     "medium",
		Budget:       entities.Money{Amount: pricing.TotalMonthly, Currency: "USD"},
		Requirements: p.extractRequirements(profile, contentPlan),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"contentPlan":  contentPlan,
			"pricing":      pricing,
			"profile":      profile,
			"source":       "onboarding",
			"automated":    true,
		},
	}
	
	// Save project
	if err := p.projectRepository.CreateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	// Create initial content items
	if err := p.contentService.CreateInitialContent(ctx, project.ProjectID, profile); err != nil {
		// Log error but don't fail the project creation
		fmt.Printf("Failed to create initial content: %v\n", err)
	}
	
	// Send notifications
	if err := p.notificationService.SendWelcomeEmail(ctx, profile.ClientID, project); err != nil {
		// Log error but don't fail the project creation
		fmt.Printf("Failed to send welcome email: %v\n", err)
	}
	
	if err := p.notificationService.SendProjectStartNotification(ctx, profile.ClientID, project); err != nil {
		// Log error but don't fail the project creation
		fmt.Printf("Failed to send project start notification: %v\n", err)
	}
	
	// Update project status to active
	project.Status = "active"
	project.UpdatedAt = time.Now()
	
	if err := p.projectRepository.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project status: %w", err)
	}
	
	return project, nil
}

// generateProjectTitle creates a descriptive title for the project
func (p *ProjectKickoffService) generateProjectTitle(profile *entities.ClientProfile) string {
	if len(profile.BusinessGoals) > 0 {
		primaryGoal := profile.BusinessGoals[0].Name
		return fmt.Sprintf("%s Content Strategy - %s", profile.Industry, primaryGoal)
	}
	
	return fmt.Sprintf("%s Content Marketing Strategy", profile.Industry)
}

// generateProjectDescription creates a detailed description for the project
func (p *ProjectKickoffService) generateProjectDescription(profile *entities.ClientProfile, plan *ContentPlan) string {
	description := fmt.Sprintf(
		"Autonomous content creation project for %s industry, targeting %s market. ",
		profile.Industry,
		profile.CompanySize,
	)
	
	if len(profile.BusinessGoals) > 0 {
		description += "Primary objectives: "
		for i, goal := range profile.BusinessGoals {
			if i > 0 {
				description += ", "
			}
			description += goal.Name
		}
		description += ". "
	}
	
	if len(plan.ContentPillars) > 0 {
		description += "Content strategy focuses on: "
		for i, pillar := range plan.ContentPillars {
			if i > 0 {
				description += ", "
			}
			description += pillar.Name
		}
		description += ". "
	}
	
	description += fmt.Sprintf(
		"Content frequency: %s. Brand voice: %s.",
		plan.Frequency,
		profile.BrandGuidelines.Voice,
	)
	
	return description
}

// extractRequirements generates project requirements from profile and plan
func (p *ProjectKickoffService) extractRequirements(profile *entities.ClientProfile, plan *ContentPlan) []string {
	requirements := []string{}
	
	// Add content goals as requirements
	for _, goal := range profile.ContentGoals {
		requirements = append(requirements, fmt.Sprintf("Create content for: %s", goal))
	}
	
	// Add brand voice requirements
	if profile.BrandGuidelines.Voice != "" {
		requirements = append(requirements, fmt.Sprintf("Maintain brand voice: %s", profile.BrandGuidelines.Voice))
	}
	
	// Add tone requirements
	if len(profile.BrandGuidelines.Tone) > 0 {
		toneList := ""
		for i, tone := range profile.BrandGuidelines.Tone {
			if i > 0 {
				toneList += ", "
			}
			toneList += tone
		}
		requirements = append(requirements, fmt.Sprintf("Use tone: %s", toneList))
	}
	
	// Add target audience requirements
	requirements = append(requirements, "Target audience: "+profile.TargetAudience.Demographics["description"].(string))
	
	// Add frequency requirements
	requirements = append(requirements, fmt.Sprintf("Publishing frequency: %s", plan.Frequency))
	
	// Add channel requirements
	if len(plan.Channels) > 0 {
		channelList := ""
		for i, channel := range plan.Channels {
			if i > 0 {
				channelList += ", "
			}
			channelList += channel
		}
		requirements = append(requirements, fmt.Sprintf("Distribution channels: %s", channelList))
	}
	
	// Add SEO requirements
	requirements = append(requirements, "Optimize for search engines and organic traffic")
	
	// Add quality requirements
	requirements = append(requirements, "Maintain high quality standards with multi-pass review")
	
	return requirements
}

// GenerateOnboardingProjectPlan creates a recommended project plan from client profile
func (p *ProjectKickoffService) GenerateOnboardingProjectPlan(ctx context.Context, profile *entities.ClientProfile) (*OnboardingProjectPlan, error) {
	// Generate content plan
	contentPlan, err := p.contentService.GenerateContentPlan(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content plan: %w", err)
	}
	
	// Calculate pricing
	pricing, err := p.pricingService.CalculateProjectPricing(ctx, profile, contentPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pricing: %w", err)
	}
	
	// Generate timeline
	timeline := p.generateProjectTimeline(contentPlan)
	
	// Generate deliverables
	deliverables := p.generateProjectDeliverables(contentPlan)
	
	// Create project plan
	plan := &OnboardingProjectPlan{
		PlanID:        uuid.New(),
		ClientID:      profile.ClientID,
		Title:         p.generateProjectTitle(profile),
		Description:   p.generateProjectDescription(profile, contentPlan),
		ContentPlan:   contentPlan,
		Pricing:       pricing,
		Timeline:      timeline,
		Deliverables:  deliverables,
		Recommendations: p.generateRecommendations(profile, contentPlan),
		NextSteps:     p.generateNextSteps(profile),
		CreatedAt:     time.Now(),
	}
	
	return plan, nil
}

// OnboardingProjectPlan represents a complete project plan generated from onboarding
type OnboardingProjectPlan struct {
	PlanID          uuid.UUID                  `json:"planId"`
	ClientID        uuid.UUID                  `json:"clientId"`
	Title           string                     `json:"title"`
	Description     string                     `json:"description"`
	ContentPlan     *ContentPlan               `json:"contentPlan"`
	Pricing         *ProjectPricing            `json:"pricing"`
	Timeline        []TimelineItem             `json:"timeline"`
	Deliverables    []Deliverable              `json:"deliverables"`
	Recommendations []string                   `json:"recommendations"`
	NextSteps       []string                   `json:"nextSteps"`
	CreatedAt       time.Time                  `json:"createdAt"`
}

// TimelineItem represents a milestone in the project timeline
type TimelineItem struct {
	Phase       string    `json:"phase"`
	Description string    `json:"description"`
	Duration    string    `json:"duration"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Deliverables []string `json:"deliverables"`
}

// Deliverable represents a specific project deliverable
type Deliverable struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Quantity    int       `json:"quantity"`
	Timeline    string    `json:"timeline"`
	Status      string    `json:"status"`
}

// generateProjectTimeline creates a timeline based on content plan
func (p *ProjectKickoffService) generateProjectTimeline(plan *ContentPlan) []TimelineItem {
	timeline := []TimelineItem{}
	
	now := time.Now()
	
	// Setup phase
	timeline = append(timeline, TimelineItem{
		Phase:       "Setup & Strategy",
		Description: "Content strategy refinement and initial setup",
		Duration:    "1-2 weeks",
		StartDate:   now,
		EndDate:     now.Add(14 * 24 * time.Hour),
		Deliverables: []string{
			"Content strategy document",
			"Content calendar setup",
			"Brand voice guidelines",
		},
	})
	
	// Content creation phase
	timeline = append(timeline, TimelineItem{
		Phase:       "Content Creation",
		Description: "Ongoing content creation and optimization",
		Duration:    "Ongoing",
		StartDate:   now.Add(14 * 24 * time.Hour),
		EndDate:     now.Add(90 * 24 * time.Hour),
		Deliverables: []string{
			"Regular content pieces",
			"Performance monitoring",
			"Content optimization",
		},
	})
	
	// Review and optimization phase
	timeline = append(timeline, TimelineItem{
		Phase:       "Review & Optimization",
		Description: "Monthly strategy review and content optimization",
		Duration:    "Monthly",
		StartDate:   now.Add(30 * 24 * time.Hour),
		EndDate:     now.Add(365 * 24 * time.Hour),
		Deliverables: []string{
			"Performance reports",
			"Strategy adjustments",
			"Content recommendations",
		},
	})
	
	return timeline
}

// generateProjectDeliverables creates deliverables list from content plan
func (p *ProjectKickoffService) generateProjectDeliverables(plan *ContentPlan) []Deliverable {
	deliverables := []Deliverable{}
	
	// Count content types
	contentTypes := make(map[string]int)
	for _, item := range plan.ContentCalendar {
		contentTypes[item.Type]++
	}
	
	// Create deliverables based on content types
	for contentType, quantity := range contentTypes {
		deliverable := Deliverable{
			Name:        contentType,
			Description: fmt.Sprintf("High-quality %s optimized for your audience and goals", contentType),
			Type:        contentType,
			Quantity:    quantity,
			Timeline:    p.getContentTimeline(contentType),
			Status:      "planned",
		}
		deliverables = append(deliverables, deliverable)
	}
	
	// Add strategy deliverables
	deliverables = append(deliverables, Deliverable{
		Name:        "Content Strategy Document",
		Description: "Comprehensive strategy document outlining approach, goals, and tactics",
		Type:        "document",
		Quantity:    1,
		Timeline:    "Week 1",
		Status:      "planned",
	})
	
	deliverables = append(deliverables, Deliverable{
		Name:        "Performance Reports",
		Description: "Monthly performance reports with insights and recommendations",
		Type:        "report",
		Quantity:    12,
		Timeline:    "Monthly",
		Status:      "planned",
	})
	
	return deliverables
}

// getContentTimeline returns appropriate timeline for content type
func (p *ProjectKickoffService) getContentTimeline(contentType string) string {
	timelines := map[string]string{
		"Blog Post":        "Weekly",
		"Social Media":     "Daily",
		"Email Newsletter": "Weekly",
		"Case Study":       "Monthly",
		"Whitepaper":       "Quarterly",
		"Video Script":     "Bi-weekly",
		"Product Description": "As needed",
		"Press Release":    "As needed",
	}
	
	if timeline, exists := timelines[contentType]; exists {
		return timeline
	}
	
	return "Regular"
}

// generateRecommendations creates strategic recommendations
func (p *ProjectKickoffService) generateRecommendations(profile *entities.ClientProfile, plan *ContentPlan) []string {
	recommendations := []string{}
	
	// Industry-specific recommendations
	switch profile.Industry {
	case "Technology":
		recommendations = append(recommendations, 
			"Focus on thought leadership content to establish industry expertise",
			"Create technical tutorials and how-to guides for your audience",
			"Leverage case studies to demonstrate real-world value",
		)
	case "Healthcare":
		recommendations = append(recommendations,
			"Prioritize educational content that builds trust and credibility",
			"Ensure all content complies with healthcare regulations",
			"Focus on patient success stories and testimonials",
		)
	case "Finance":
		recommendations = append(recommendations,
			"Create educational content that simplifies complex financial concepts",
			"Focus on building trust through transparent, accurate information",
			"Leverage market insights and trend analysis",
		)
	default:
		recommendations = append(recommendations,
			"Focus on building brand awareness through consistent messaging",
			"Create valuable, educational content for your target audience",
			"Optimize content for search engines to increase organic visibility",
		)
	}
	
	// Goal-specific recommendations
	for _, goal := range profile.BusinessGoals {
		switch goal.Name {
		case "Lead Generation":
			recommendations = append(recommendations, "Create gated content like whitepapers and guides to capture leads")
		case "Brand Awareness":
			recommendations = append(recommendations, "Invest in storytelling content that showcases your brand personality")
		case "Customer Retention":
			recommendations = append(recommendations, "Develop educational content that helps existing customers succeed")
		}
	}
	
	// Content frequency recommendations
	if plan.Frequency == "Daily" {
		recommendations = append(recommendations, "Consider batch content creation to maintain consistent quality at high volume")
	}
	
	return recommendations
}

// generateNextSteps creates actionable next steps for the client
func (p *ProjectKickoffService) generateNextSteps(profile *entities.ClientProfile) []string {
	nextSteps := []string{
		"Review and approve your personalized content strategy",
		"Set up your client dashboard access",
		"Schedule a strategy review call with your content manager",
		"Approve the content calendar for the first month",
		"Set up analytics tracking for performance monitoring",
	}
	
	// Add industry-specific next steps
	if profile.Industry == "E-commerce" {
		nextSteps = append(nextSteps, "Provide product catalog access for product description creation")
	}
	
	if len(profile.CompetitorURLs) > 0 {
		nextSteps = append(nextSteps, "Review competitive analysis insights and differentiation strategy")
	}
	
	return nextSteps
}
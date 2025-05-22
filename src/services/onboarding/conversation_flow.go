package onboarding

import (
	"errors"
	"fmt"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ConversationFlowImpl implements the ConversationFlow interface
type ConversationFlowImpl struct {
	questions map[entities.OnboardingStage][]Question
}

// NewConversationFlow creates a new conversation flow manager
func NewConversationFlow() *ConversationFlowImpl {
	flow := &ConversationFlowImpl{
		questions: make(map[entities.OnboardingStage][]Question),
	}
	flow.initializeQuestions()
	return flow
}

// GetCurrentStageQuestions returns questions for the current stage
func (c *ConversationFlowImpl) GetCurrentStageQuestions(stage entities.OnboardingStage) []Question {
	return c.questions[stage]
}

// ProcessResponse processes a client response and validates it
func (c *ConversationFlowImpl) ProcessResponse(stage entities.OnboardingStage, responses map[string]interface{}, key string, value interface{}) error {
	questions := c.questions[stage]
	
	// Find the question being answered
	var question *Question
	for i := range questions {
		if questions[i].ID == key {
			question = &questions[i]
			break
		}
	}
	
	if question == nil {
		return fmt.Errorf("question with ID %s not found for stage %s", key, stage)
	}
	
	// Validate the response
	if err := c.validateResponse(question, value); err != nil {
		return err
	}
	
	// Store the response
	responses[key] = value
	
	return nil
}

// GetNextStage determines the next stage based on current stage and responses
func (c *ConversationFlowImpl) GetNextStage(stage entities.OnboardingStage, responses map[string]interface{}) entities.OnboardingStage {
	switch stage {
	case entities.StageInitial:
		return entities.StageIndustry
	case entities.StageIndustry:
		return entities.StageGoals
	case entities.StageGoals:
		return entities.StageAudience
	case entities.StageAudience:
		return entities.StageStyle
	case entities.StageStyle:
		return entities.StageBrand
	case entities.StageBrand:
		return entities.StageCompetitors
	case entities.StageCompetitors:
		return entities.StageWelcome
	case entities.StageWelcome:
		return entities.StageComplete
	default:
		return entities.StageComplete
	}
}

// IsStageComplete checks if all required questions for a stage have been answered
func (c *ConversationFlowImpl) IsStageComplete(stage entities.OnboardingStage, responses map[string]interface{}) bool {
	questions := c.questions[stage]
	
	for _, question := range questions {
		if question.Required {
			if _, exists := responses[question.ID]; !exists {
				return false
			}
		}
	}
	
	return true
}

// validateResponse validates a response against question rules
func (c *ConversationFlowImpl) validateResponse(question *Question, value interface{}) error {
	if question.Required && (value == nil || value == "") {
		return errors.New("this field is required")
	}
	
	if question.Validation == nil {
		return nil
	}
	
	validation := question.Validation
	
	// Validate string responses
	if str, ok := value.(string); ok {
		if validation.MinLength != nil && len(str) < *validation.MinLength {
			return fmt.Errorf("response must be at least %d characters", *validation.MinLength)
		}
		if validation.MaxLength != nil && len(str) > *validation.MaxLength {
			return fmt.Errorf("response must be no more than %d characters", *validation.MaxLength)
		}
	}
	
	// Validate numeric responses
	if num, ok := value.(float64); ok {
		if validation.MinValue != nil && num < *validation.MinValue {
			return fmt.Errorf("value must be at least %f", *validation.MinValue)
		}
		if validation.MaxValue != nil && num > *validation.MaxValue {
			return fmt.Errorf("value must be no more than %f", *validation.MaxValue)
		}
	}
	
	// Validate choice responses
	if question.Type == "choice" || question.Type == "multiple" {
		if err := c.validateChoiceResponse(question, value); err != nil {
			return err
		}
	}
	
	return nil
}

// validateChoiceResponse validates choice-type responses
func (c *ConversationFlowImpl) validateChoiceResponse(question *Question, value interface{}) error {
	validOptions := make(map[string]bool)
	for _, option := range question.Options {
		validOptions[option.Value] = true
	}
	
	switch v := value.(type) {
	case string:
		if !validOptions[v] {
			return errors.New("invalid option selected")
		}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				if !validOptions[str] {
					return errors.New("invalid option selected")
				}
			}
		}
	case []string:
		for _, str := range v {
			if !validOptions[str] {
				return errors.New("invalid option selected")
			}
		}
	default:
		return errors.New("invalid response format for choice question")
	}
	
	return nil
}

// initializeQuestions sets up all the questions for each onboarding stage
func (c *ConversationFlowImpl) initializeQuestions() {
	// Initial stage questions
	c.questions[entities.StageInitial] = []Question{
		{
			ID:       "welcome_confirmation",
			Type:     "choice",
			Question: "Welcome! I'm here to help you get the most out of our content creation service. Are you ready to get started?",
			Required: true,
			Options: []Option{
				{Value: "yes", Label: "Yes, let's get started!"},
				{Value: "learn_more", Label: "I'd like to learn more first"},
			},
		},
	}
	
	// Industry identification questions
	c.questions[entities.StageIndustry] = []Question{
		{
			ID:          "company_name",
			Type:        "text",
			Question:    "What's your company name?",
			Description: "This helps us personalize your experience",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(2), MaxLength: intPtr(100)},
		},
		{
			ID:       "industry",
			Type:     "choice",
			Question: "Which industry best describes your business?",
			Required: true,
			Options: []Option{
				{Value: "technology", Label: "Technology & Software"},
				{Value: "healthcare", Label: "Healthcare & Medical"},
				{Value: "finance", Label: "Finance & Banking"},
				{Value: "ecommerce", Label: "E-commerce & Retail"},
				{Value: "education", Label: "Education & Training"},
				{Value: "manufacturing", Label: "Manufacturing & Industrial"},
				{Value: "professional_services", Label: "Professional Services"},
				{Value: "real_estate", Label: "Real Estate"},
				{Value: "food_beverage", Label: "Food & Beverage"},
				{Value: "travel_hospitality", Label: "Travel & Hospitality"},
				{Value: "nonprofit", Label: "Non-profit & Social Impact"},
				{Value: "other", Label: "Other"},
			},
		},
		{
			ID:          "industry_other",
			Type:        "text",
			Question:    "Please describe your industry:",
			Description: "Help us understand your specific field",
			Required:    false,
			Validation:  &Validation{MinLength: intPtr(5), MaxLength: intPtr(200)},
		},
		{
			ID:       "company_size",
			Type:     "choice",
			Question: "What's the size of your company?",
			Required: true,
			Options: []Option{
				{Value: "solo", Label: "Just me (Solo entrepreneur)"},
				{Value: "startup", Label: "Startup (2-10 employees)"},
				{Value: "small", Label: "Small business (11-50 employees)"},
				{Value: "medium", Label: "Medium business (51-200 employees)"},
				{Value: "large", Label: "Large company (200+ employees)"},
			},
		},
	}
	
	// Business goals questions
	c.questions[entities.StageGoals] = []Question{
		{
			ID:       "primary_goals",
			Type:     "multiple",
			Question: "What are your primary business goals? (Select all that apply)",
			Required: true,
			Options: []Option{
				{Value: "brand_awareness", Label: "Increase brand awareness"},
				{Value: "lead_generation", Label: "Generate more leads"},
				{Value: "customer_retention", Label: "Improve customer retention"},
				{Value: "thought_leadership", Label: "Establish thought leadership"},
				{Value: "product_education", Label: "Educate customers about products/services"},
				{Value: "seo_traffic", Label: "Improve SEO and organic traffic"},
				{Value: "social_engagement", Label: "Increase social media engagement"},
				{Value: "sales_support", Label: "Support sales team with content"},
				{Value: "customer_support", Label: "Reduce customer support burden"},
				{Value: "competitive_advantage", Label: "Gain competitive advantage"},
			},
		},
		{
			ID:          "content_goals",
			Type:        "multiple",
			Question:    "What types of content are you most interested in?",
			Description: "This helps us prioritize your content strategy",
			Required:    true,
			Options: []Option{
				{Value: "blog_posts", Label: "Blog posts & articles"},
				{Value: "social_media", Label: "Social media content"},
				{Value: "email_campaigns", Label: "Email newsletters & campaigns"},
				{Value: "whitepapers", Label: "Whitepapers & case studies"},
				{Value: "product_descriptions", Label: "Product descriptions"},
				{Value: "website_copy", Label: "Website copy & landing pages"},
				{Value: "press_releases", Label: "Press releases"},
				{Value: "video_scripts", Label: "Video scripts"},
				{Value: "presentations", Label: "Presentations & slide decks"},
				{Value: "documentation", Label: "Technical documentation"},
			},
		},
		{
			ID:          "content_frequency",
			Type:        "choice",
			Question:    "How often do you plan to publish content?",
			Description: "This helps us plan your content calendar",
			Required:    true,
			Options: []Option{
				{Value: "daily", Label: "Daily"},
				{Value: "weekly", Label: "Weekly"},
				{Value: "bi_weekly", Label: "Bi-weekly"},
				{Value: "monthly", Label: "Monthly"},
				{Value: "quarterly", Label: "Quarterly"},
				{Value: "as_needed", Label: "As needed"},
			},
		},
	}
	
	// Target audience questions
	c.questions[entities.StageAudience] = []Question{
		{
			ID:          "target_audience_description",
			Type:        "text",
			Question:    "Describe your ideal customer or target audience:",
			Description: "Include demographics, job roles, interests, etc.",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(20), MaxLength: intPtr(500)},
		},
		{
			ID:       "audience_size",
			Type:     "choice",
			Question: "What's the approximate size of your target audience?",
			Required: false,
			Options: []Option{
				{Value: "niche", Label: "Niche market (under 10K people)"},
				{Value: "small", Label: "Small market (10K-100K people)"},
				{Value: "medium", Label: "Medium market (100K-1M people)"},
				{Value: "large", Label: "Large market (1M+ people)"},
				{Value: "unknown", Label: "I'm not sure"},
			},
		},
		{
			ID:          "pain_points",
			Type:        "text",
			Question:    "What are the main challenges or pain points your audience faces?",
			Description: "Understanding their problems helps us create relevant content",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(20), MaxLength: intPtr(500)},
		},
		{
			ID:       "content_channels",
			Type:     "multiple",
			Question: "Where does your audience typically consume content?",
			Required: true,
			Options: []Option{
				{Value: "company_blog", Label: "Company blog/website"},
				{Value: "linkedin", Label: "LinkedIn"},
				{Value: "twitter", Label: "Twitter/X"},
				{Value: "facebook", Label: "Facebook"},
				{Value: "instagram", Label: "Instagram"},
				{Value: "youtube", Label: "YouTube"},
				{Value: "email", Label: "Email newsletters"},
				{Value: "industry_publications", Label: "Industry publications"},
				{Value: "podcasts", Label: "Podcasts"},
				{Value: "webinars", Label: "Webinars & events"},
			},
		},
	}
	
	// Style preferences questions
	c.questions[entities.StageStyle] = []Question{
		{
			ID:       "writing_tone",
			Type:     "multiple",
			Question: "What tone should your content have? (Select all that apply)",
			Required: true,
			Options: []Option{
				{Value: "professional", Label: "Professional"},
				{Value: "conversational", Label: "Conversational"},
				{Value: "authoritative", Label: "Authoritative"},
				{Value: "friendly", Label: "Friendly"},
				{Value: "educational", Label: "Educational"},
				{Value: "inspiring", Label: "Inspiring"},
				{Value: "humorous", Label: "Humorous"},
				{Value: "technical", Label: "Technical"},
				{Value: "casual", Label: "Casual"},
				{Value: "formal", Label: "Formal"},
			},
		},
		{
			ID:       "content_complexity",
			Type:     "choice",
			Question: "What level of complexity should your content have?",
			Required: true,
			Options: []Option{
				{Value: "beginner", Label: "Beginner-friendly (accessible to everyone)"},
				{Value: "intermediate", Label: "Intermediate (some industry knowledge assumed)"},
				{Value: "advanced", Label: "Advanced (expert-level content)"},
				{Value: "mixed", Label: "Mixed (varies by content piece)"},
			},
		},
		{
			ID:          "avoid_topics",
			Type:        "text",
			Question:    "Are there any topics, words, or approaches you want us to avoid?",
			Description: "This helps us maintain your brand standards",
			Required:    false,
			Validation:  &Validation{MaxLength: intPtr(300)},
		},
	}
	
	// Brand voice questions
	c.questions[entities.StageBrand] = []Question{
		{
			ID:          "brand_personality",
			Type:        "text",
			Question:    "How would you describe your brand's personality in 3-5 words?",
			Description: "For example: innovative, trustworthy, approachable",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(5), MaxLength: intPtr(100)},
		},
		{
			ID:          "brand_values",
			Type:        "text",
			Question:    "What are your core brand values?",
			Description: "What principles guide your business decisions?",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(10), MaxLength: intPtr(300)},
		},
		{
			ID:          "unique_value_prop",
			Type:        "text",
			Question:    "What makes your business unique from competitors?",
			Description: "Your unique value proposition or differentiator",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(20), MaxLength: intPtr(300)},
		},
		{
			ID:          "example_content",
			Type:        "text",
			Question:    "Do you have any existing content (URLs) that represents your ideal brand voice?",
			Description: "Share links to blog posts, pages, or content you love",
			Required:    false,
			Validation:  &Validation{MaxLength: intPtr(500)},
		},
	}
	
	// Competitor analysis questions
	c.questions[entities.StageCompetitors] = []Question{
		{
			ID:          "main_competitors",
			Type:        "text",
			Question:    "Who are your main competitors?",
			Description: "List 3-5 companies you compete with directly",
			Required:    true,
			Validation:  &Validation{MinLength: intPtr(10), MaxLength: intPtr(300)},
		},
		{
			ID:          "competitor_websites",
			Type:        "text",
			Question:    "Please share competitor website URLs (one per line):",
			Description: "We'll analyze their content strategy to help differentiate yours",
			Required:    false,
			Validation:  &Validation{MaxLength: intPtr(500)},
		},
		{
			ID:          "competitive_advantage",
			Type:        "text",
			Question:    "What content or messaging gaps do you see in your industry?",
			Description: "Opportunities where you could stand out from competitors",
			Required:    false,
			Validation:  &Validation{MaxLength: intPtr(400)},
		},
	}
	
	// Welcome/completion questions
	c.questions[entities.StageWelcome] = []Question{
		{
			ID:       "onboarding_feedback",
			Type:     "choice",
			Question: "How was this onboarding experience?",
			Required: false,
			Options: []Option{
				{Value: "excellent", Label: "Excellent"},
				{Value: "good", Label: "Good"},
				{Value: "okay", Label: "Okay"},
				{Value: "needs_improvement", Label: "Needs improvement"},
			},
		},
		{
			ID:          "additional_notes",
			Type:        "text",
			Question:    "Any additional information or special requests?",
			Description: "Anything else you'd like us to know?",
			Required:    false,
			Validation:  &Validation{MaxLength: intPtr(500)},
		},
	}
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

// GetStageProgress calculates the completion percentage for the onboarding
func (c *ConversationFlowImpl) GetStageProgress(stage entities.OnboardingStage) float64 {
	stageOrder := []entities.OnboardingStage{
		entities.StageInitial,
		entities.StageIndustry,
		entities.StageGoals,
		entities.StageAudience,
		entities.StageStyle,
		entities.StageBrand,
		entities.StageCompetitors,
		entities.StageWelcome,
		entities.StageComplete,
	}
	
	for i, s := range stageOrder {
		if s == stage {
			return float64(i) / float64(len(stageOrder)-1)
		}
	}
	
	return 0.0
}

// GetStageTitle returns a human-readable title for the stage
func (c *ConversationFlowImpl) GetStageTitle(stage entities.OnboardingStage) string {
	titles := map[entities.OnboardingStage]string{
		entities.StageInitial:     "Welcome",
		entities.StageIndustry:    "About Your Business",
		entities.StageGoals:       "Your Goals",
		entities.StageAudience:    "Target Audience",
		entities.StageStyle:       "Content Style",
		entities.StageBrand:       "Brand Voice",
		entities.StageCompetitors: "Competitive Landscape",
		entities.StageWelcome:     "Almost Done!",
		entities.StageComplete:    "Complete",
	}
	
	if title, exists := titles[stage]; exists {
		return title
	}
	
	return string(stage)
}
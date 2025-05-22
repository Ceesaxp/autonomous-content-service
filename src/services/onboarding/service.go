package onboarding

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// OnboardingServiceImpl implements the OnboardingService interface
type OnboardingServiceImpl struct {
	repository       OnboardingRepository
	conversationFlow ConversationFlow
	industryAnalyzer IndustryAnalyzer
	competitorAnalyzer CompetitorAnalyzer
	brandExtractor   BrandVoiceExtractor
}

// NewOnboardingService creates a new onboarding service
func NewOnboardingService(
	repo OnboardingRepository,
	flow ConversationFlow,
	industryAnalyzer IndustryAnalyzer,
	competitorAnalyzer CompetitorAnalyzer,
	brandExtractor BrandVoiceExtractor,
) *OnboardingServiceImpl {
	return &OnboardingServiceImpl{
		repository:       repo,
		conversationFlow: flow,
		industryAnalyzer: industryAnalyzer,
		competitorAnalyzer: competitorAnalyzer,
		brandExtractor:   brandExtractor,
	}
}

// StartOnboarding creates a new onboarding session for a client
func (s *OnboardingServiceImpl) StartOnboarding(ctx context.Context, clientID uuid.UUID) (*entities.OnboardingSession, error) {
	session := entities.NewOnboardingSession(clientID)
	
	// Add initial welcome message
	session.AddMessage("system", "Welcome! I'm here to help you set up your content strategy. Let's start by learning about your business.", nil)
	
	if err := s.repository.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save onboarding session: %w", err)
	}
	
	return session, nil
}

// GetSession retrieves an onboarding session by ID
func (s *OnboardingServiceImpl) GetSession(ctx context.Context, sessionID uuid.UUID) (*entities.OnboardingSession, error) {
	return s.repository.GetSession(ctx, sessionID)
}

// UpdateSession updates an existing onboarding session
func (s *OnboardingServiceImpl) UpdateSession(ctx context.Context, session *entities.OnboardingSession) error {
	return s.repository.SaveSession(ctx, session)
}

// ProcessMessage processes a client message and generates an appropriate response
func (s *OnboardingServiceImpl) ProcessMessage(ctx context.Context, sessionID uuid.UUID, message string) (*ConversationResponse, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Add client message to conversation log
	session.AddMessage("client", message, nil)
	
	// Process the message based on current stage
	response, err := s.processStageMessage(ctx, session, message)
	if err != nil {
		return nil, fmt.Errorf("failed to process message: %w", err)
	}
	
	// Add system response to conversation log
	session.AddMessage("system", response.Message, map[string]interface{}{
		"stage":    response.Stage,
		"progress": response.Progress,
	})
	
	// Save updated session
	if err := s.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	
	return response, nil
}

// processStageMessage processes a message based on the current onboarding stage
func (s *OnboardingServiceImpl) processStageMessage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	switch session.Stage {
	case entities.StageInitial:
		return s.processInitialStage(ctx, session, message)
	case entities.StageIndustry:
		return s.processIndustryStage(ctx, session, message)
	case entities.StageGoals:
		return s.processGoalsStage(ctx, session, message)
	case entities.StageAudience:
		return s.processAudienceStage(ctx, session, message)
	case entities.StageStyle:
		return s.processStyleStage(ctx, session, message)
	case entities.StageBrand:
		return s.processBrandStage(ctx, session, message)
	case entities.StageCompetitors:
		return s.processCompetitorsStage(ctx, session, message)
	case entities.StageWelcome:
		return s.processWelcomeStage(ctx, session, message)
	default:
		return s.generateGenericResponse(session), nil
	}
}

// processInitialStage handles the initial welcome stage
func (s *OnboardingServiceImpl) processInitialStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	message = strings.ToLower(strings.TrimSpace(message))
	
	if strings.Contains(message, "yes") || strings.Contains(message, "start") || strings.Contains(message, "ready") {
		session.AddResponse("welcome_confirmation", "yes")
		session.UpdateStage(entities.StageIndustry)
		
		return &ConversationResponse{
			Message:    "Great! Let's start by learning about your business. This helps me create content that resonates with your industry and audience.",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageIndustry),
			Stage:      entities.StageIndustry,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageIndustry),
			NextAction: "answer_questions",
		}, nil
	}
	
	if strings.Contains(message, "learn") || strings.Contains(message, "more") {
		return &ConversationResponse{
			Message: "I'd be happy to explain! I'm an AI assistant that helps create personalized content strategies. I'll ask you questions about your business, goals, and preferences, then create a content plan tailored specifically for you. The whole process takes about 10-15 minutes. Ready to get started?",
			Questions: s.conversationFlow.GetCurrentStageQuestions(entities.StageInitial),
			Stage:     entities.StageInitial,
			Progress:  0.0,
			NextAction: "confirm_start",
		}, nil
	}
	
	// Default response for unclear input
	return &ConversationResponse{
		Message:    "I'm here to help you create a personalized content strategy! Are you ready to get started, or would you like to learn more about the process first?",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageInitial),
		Stage:      entities.StageInitial,
		Progress:   0.0,
		NextAction: "clarify_intent",
	}, nil
}

// processIndustryStage handles industry identification questions
func (s *OnboardingServiceImpl) processIndustryStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	// Extract structured data from natural language response
	response := s.extractIndustryInfo(message)
	
	// Store responses
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	// Check if stage is complete
	if s.conversationFlow.IsStageComplete(entities.StageIndustry, session.Responses) {
		// Analyze industry
		if industry, ok := session.Responses["industry"].(string); ok {
			analysis, err := s.industryAnalyzer.GenerateIndustryReport(ctx, industry, []string{})
			if err == nil {
				session.AddResponse("industry_analysis", analysis)
			}
		}
		
		session.UpdateStage(entities.StageGoals)
		
		return &ConversationResponse{
			Message:    "Perfect! Now I understand your business context. Let's talk about your goals - what do you want to achieve with your content?",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageGoals),
			Stage:      entities.StageGoals,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageGoals),
			NextAction: "define_goals",
		}, nil
	}
	
	// Request missing information
	return s.requestMissingIndustryInfo(session)
}

// processGoalsStage handles business goals identification
func (s *OnboardingServiceImpl) processGoalsStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractGoalsInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	if s.conversationFlow.IsStageComplete(entities.StageGoals, session.Responses) {
		session.UpdateStage(entities.StageAudience)
		
		return &ConversationResponse{
			Message:    "Excellent! Understanding your goals helps me recommend the right content types. Now, let's talk about your target audience - who are you trying to reach?",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageAudience),
			Stage:      entities.StageAudience,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageAudience),
			NextAction: "define_audience",
		}, nil
	}
	
	return s.requestMissingGoalsInfo(session)
}

// processAudienceStage handles target audience definition
func (s *OnboardingServiceImpl) processAudienceStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractAudienceInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	if s.conversationFlow.IsStageComplete(entities.StageAudience, session.Responses) {
		session.UpdateStage(entities.StageStyle)
		
		return &ConversationResponse{
			Message:    "Great! Knowing your audience helps me tailor the messaging perfectly. Now let's define your content style and tone preferences.",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageStyle),
			Stage:      entities.StageStyle,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageStyle),
			NextAction: "define_style",
		}, nil
	}
	
	return s.requestMissingAudienceInfo(session)
}

// processStyleStage handles content style preferences
func (s *OnboardingServiceImpl) processStyleStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractStyleInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	if s.conversationFlow.IsStageComplete(entities.StageStyle, session.Responses) {
		session.UpdateStage(entities.StageBrand)
		
		return &ConversationResponse{
			Message:    "Perfect! Your style preferences will ensure consistent, on-brand content. Now let's capture your unique brand voice and personality.",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageBrand),
			Stage:      entities.StageBrand,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageBrand),
			NextAction: "define_brand",
		}, nil
	}
	
	return s.requestMissingStyleInfo(session)
}

// processBrandStage handles brand voice definition
func (s *OnboardingServiceImpl) processBrandStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractBrandInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	if s.conversationFlow.IsStageComplete(entities.StageBrand, session.Responses) {
		session.UpdateStage(entities.StageCompetitors)
		
		return &ConversationResponse{
			Message:    "Excellent! Your brand voice will make your content distinctive and memorable. Finally, let's understand your competitive landscape to identify opportunities.",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageCompetitors),
			Stage:      entities.StageCompetitors,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageCompetitors),
			NextAction: "analyze_competitors",
		}, nil
	}
	
	return s.requestMissingBrandInfo(session)
}

// processCompetitorsStage handles competitive analysis
func (s *OnboardingServiceImpl) processCompetitorsStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractCompetitorInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	if s.conversationFlow.IsStageComplete(entities.StageCompetitors, session.Responses) {
		// Perform competitive analysis if URLs provided
		if urls, ok := session.Responses["competitor_websites"].([]string); ok && len(urls) > 0 {
			analysis, err := s.competitorAnalyzer.ExtractContentStrategy(urls)
			if err == nil {
				session.AddResponse("competitor_analysis", analysis)
			}
		}
		
		session.UpdateStage(entities.StageWelcome)
		
		return &ConversationResponse{
			Message:    "Fantastic! I now have everything I need to create your personalized content strategy. Let me prepare your recommendations...",
			Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageWelcome),
			Stage:      entities.StageWelcome,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageWelcome),
			NextAction: "finalize_onboarding",
		}, nil
	}
	
	return s.requestMissingCompetitorInfo(session)
}

// processWelcomeStage handles the final welcome and completion
func (s *OnboardingServiceImpl) processWelcomeStage(ctx context.Context, session *entities.OnboardingSession, message string) (*ConversationResponse, error) {
	response := s.extractWelcomeInfo(message)
	
	for key, value := range response {
		session.AddResponse(key, value)
	}
	
	session.Complete()
	
	return &ConversationResponse{
		Message:    "Thank you for completing the onboarding! Your personalized content strategy is ready. I'll now create your client profile and you can start your first project whenever you're ready.",
		Stage:      entities.StageComplete,
		Progress:   1.0,
		NextAction: "create_profile",
		Metadata: map[string]interface{}{
			"onboarding_complete": true,
			"next_steps": []string{
				"Review your content strategy",
				"Start your first project",
				"Access your client dashboard",
			},
		},
	}, nil
}

// CompleteOnboarding finalizes the onboarding process and creates a client profile
func (s *OnboardingServiceImpl) CompleteOnboarding(ctx context.Context, sessionID uuid.UUID) (*entities.ClientProfile, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	if session.Stage != entities.StageComplete {
		return nil, fmt.Errorf("onboarding session is not complete")
	}
	
	// Create client profile from session data
	profile, err := entities.NewClientProfileFromOnboarding(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create client profile: %w", err)
	}
	
	// Clean up session (optional - you might want to keep for analytics)
	// s.repository.DeleteSession(ctx, sessionID)
	
	return profile, nil
}

// GetNextQuestions returns the next set of questions for the current stage
func (s *OnboardingServiceImpl) GetNextQuestions(ctx context.Context, sessionID uuid.UUID) ([]Question, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return s.conversationFlow.GetCurrentStageQuestions(session.Stage), nil
}

// ValidateResponse validates a response for the current stage
func (s *OnboardingServiceImpl) ValidateResponse(ctx context.Context, sessionID uuid.UUID, key string, value interface{}) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	return s.conversationFlow.ProcessResponse(session.Stage, session.Responses, key, value)
}

// Helper methods for extracting information from natural language

func (s *OnboardingServiceImpl) extractIndustryInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Simple keyword extraction (in production, use NLP)
	message = strings.ToLower(message)
	
	industries := map[string]string{
		"tech":       "technology",
		"software":   "technology",
		"health":     "healthcare",
		"medical":    "healthcare",
		"finance":    "finance",
		"banking":    "finance",
		"retail":     "ecommerce",
		"ecommerce":  "ecommerce",
		"education":  "education",
		"learning":   "education",
	}
	
	for keyword, industry := range industries {
		if strings.Contains(message, keyword) {
			response["industry"] = industry
			break
		}
	}
	
	sizes := map[string]string{
		"solo":      "solo",
		"startup":   "startup",
		"small":     "small",
		"medium":    "medium",
		"large":     "large",
		"just me":   "solo",
		"employees": "small",
	}
	
	for keyword, size := range sizes {
		if strings.Contains(message, keyword) {
			response["company_size"] = size
			break
		}
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractGoalsInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Extract goals from message
	goals := []string{}
	goalKeywords := map[string]string{
		"brand":     "brand_awareness",
		"leads":     "lead_generation",
		"traffic":   "seo_traffic",
		"social":    "social_engagement",
		"retention": "customer_retention",
		"thought":   "thought_leadership",
	}
	
	message = strings.ToLower(message)
	for keyword, goal := range goalKeywords {
		if strings.Contains(message, keyword) {
			goals = append(goals, goal)
		}
	}
	
	if len(goals) > 0 {
		response["primary_goals"] = goals
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractAudienceInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Simple extraction - in production use NLP
	if len(message) > 20 {
		response["target_audience_description"] = message
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractStyleInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	tones := []string{}
	toneKeywords := map[string]string{
		"professional": "professional",
		"casual":       "casual",
		"friendly":     "friendly",
		"formal":       "formal",
		"conversational": "conversational",
	}
	
	message = strings.ToLower(message)
	for keyword, tone := range toneKeywords {
		if strings.Contains(message, keyword) {
			tones = append(tones, tone)
		}
	}
	
	if len(tones) > 0 {
		response["writing_tone"] = tones
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractBrandInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Extract brand personality and values
	if len(message) > 10 {
		response["brand_personality"] = message
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractCompetitorInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Extract competitor information
	if len(message) > 5 {
		response["main_competitors"] = message
	}
	
	return response
}

func (s *OnboardingServiceImpl) extractWelcomeInfo(message string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Extract feedback and additional notes
	if len(message) > 0 {
		response["additional_notes"] = message
	}
	
	return response
}

// Helper methods for requesting missing information

func (s *OnboardingServiceImpl) requestMissingIndustryInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	questions := s.conversationFlow.GetCurrentStageQuestions(entities.StageIndustry)
	missingQuestions := []Question{}
	
	for _, q := range questions {
		if q.Required {
			if _, exists := session.Responses[q.ID]; !exists {
				missingQuestions = append(missingQuestions, q)
			}
		}
	}
	
	if len(missingQuestions) > 0 {
		return &ConversationResponse{
			Message:    "I'd like to learn more about your business. Could you help me with a few more details?",
			Questions:  missingQuestions,
			Stage:      entities.StageIndustry,
			Progress:   s.conversationFlow.GetStageProgress(entities.StageIndustry),
			NextAction: "provide_business_info",
		}, nil
	}
	
	return s.generateGenericResponse(session), nil
}

func (s *OnboardingServiceImpl) requestMissingGoalsInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	return &ConversationResponse{
		Message:    "What are your main goals for content marketing? Are you looking to increase brand awareness, generate leads, or something else?",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageGoals),
		Stage:      entities.StageGoals,
		Progress:   s.conversationFlow.GetStageProgress(entities.StageGoals),
		NextAction: "clarify_goals",
	}, nil
}

func (s *OnboardingServiceImpl) requestMissingAudienceInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	return &ConversationResponse{
		Message:    "Tell me more about your target audience. Who are you trying to reach with your content?",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageAudience),
		Stage:      entities.StageAudience,
		Progress:   s.conversationFlow.GetStageProgress(entities.StageAudience),
		NextAction: "describe_audience",
	}, nil
}

func (s *OnboardingServiceImpl) requestMissingStyleInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	return &ConversationResponse{
		Message:    "What style and tone should your content have? Should it be professional, casual, technical, or something else?",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageStyle),
		Stage:      entities.StageStyle,
		Progress:   s.conversationFlow.GetStageProgress(entities.StageStyle),
		NextAction: "define_style",
	}, nil
}

func (s *OnboardingServiceImpl) requestMissingBrandInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	return &ConversationResponse{
		Message:    "Help me understand your brand personality. How would you describe your brand in a few words?",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageBrand),
		Stage:      entities.StageBrand,
		Progress:   s.conversationFlow.GetStageProgress(entities.StageBrand),
		NextAction: "describe_brand",
	}, nil
}

func (s *OnboardingServiceImpl) requestMissingCompetitorInfo(session *entities.OnboardingSession) (*ConversationResponse, error) {
	return &ConversationResponse{
		Message:    "Who are your main competitors? Understanding your competitive landscape helps me create content that differentiates you.",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(entities.StageCompetitors),
		Stage:      entities.StageCompetitors,
		Progress:   s.conversationFlow.GetStageProgress(entities.StageCompetitors),
		NextAction: "identify_competitors",
	}, nil
}

func (s *OnboardingServiceImpl) generateGenericResponse(session *entities.OnboardingSession) *ConversationResponse {
	return &ConversationResponse{
		Message:    "I understand. Let's continue with the next step.",
		Questions:  s.conversationFlow.GetCurrentStageQuestions(session.Stage),
		Stage:      session.Stage,
		Progress:   s.conversationFlow.GetStageProgress(session.Stage),
		NextAction: "continue",
	}
}
package onboarding

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// OnboardingService defines the interface for client onboarding operations
type OnboardingService interface {
	// Session Management
	StartOnboarding(ctx context.Context, clientID uuid.UUID) (*entities.OnboardingSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*entities.OnboardingSession, error)
	UpdateSession(ctx context.Context, session *entities.OnboardingSession) error
	CompleteOnboarding(ctx context.Context, sessionID uuid.UUID) (*entities.ClientProfile, error)

	// Conversation Flow
	ProcessMessage(ctx context.Context, sessionID uuid.UUID, message string) (*ConversationResponse, error)
	GetNextQuestions(ctx context.Context, sessionID uuid.UUID) ([]Question, error)
	ValidateResponse(ctx context.Context, sessionID uuid.UUID, key string, value interface{}) error

	// Data Analysis
	AnalyzeIndustry(ctx context.Context, industry string) (*IndustryAnalysis, error)
	AnalyzeCompetitors(ctx context.Context, competitorURLs []string) (*CompetitorAnalysis, error)
	ExtractBrandVoice(ctx context.Context, exampleContent []string) (*entities.BrandGuidelines, error)
}

// OnboardingRepository defines the interface for onboarding data persistence
type OnboardingRepository interface {
	// Session persistence
	SaveSession(ctx context.Context, session *entities.OnboardingSession) error
	GetSession(ctx context.Context, sessionID uuid.UUID) (*entities.OnboardingSession, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error

	// Analytics and tracking
	GetSessionsByClient(ctx context.Context, clientID uuid.UUID) ([]*entities.OnboardingSession, error)
	GetIncompleteSessionsOlderThan(ctx context.Context, hours int) ([]*entities.OnboardingSession, error)
}

// ConversationFlow defines the interface for managing conversation logic
type ConversationFlow interface {
	GetCurrentStageQuestions(stage entities.OnboardingStage) []Question
	ProcessResponse(stage entities.OnboardingStage, responses map[string]interface{}, key string, value interface{}) error
	GetNextStage(stage entities.OnboardingStage, responses map[string]interface{}) entities.OnboardingStage
	IsStageComplete(stage entities.OnboardingStage, responses map[string]interface{}) bool
	GetStageProgress(stage entities.OnboardingStage) float64
}

// IndustryAnalyzer defines the interface for industry classification and analysis
type IndustryAnalyzer interface {
	ClassifyIndustry(description string) (*IndustryClassification, error)
	GetIndustryInsights(industry string) (*IndustryInsights, error)
	SuggestContentTypes(industry string) ([]string, error)
	GenerateIndustryReport(ctx context.Context, industry string, clientGoals []string) (*IndustryAnalysis, error)
}

// CompetitorAnalyzer defines the interface for competitive analysis
type CompetitorAnalyzer interface {
	AnalyzeWebsite(url string) (*WebsiteAnalysis, error)
	ExtractContentStrategy(urls []string) (*ContentStrategy, error)
	IdentifyGaps(clientGoals []string, competitorStrategies []*ContentStrategy) (*GapAnalysis, error)
}

// BrandVoiceExtractor defines the interface for brand voice analysis
type BrandVoiceExtractor interface {
	AnalyzeContent(content []string) (*entities.BrandGuidelines, error)
	ExtractTone(content string) ([]string, error)
	IdentifyValues(content []string) ([]string, error)
	GenerateGuidelines(analysis *BrandAnalysis) (*entities.BrandGuidelines, error)
}

// Data Transfer Objects

// ConversationResponse represents a response from the onboarding system
type ConversationResponse struct {
	Message     string                 `json:"message"`
	Questions   []Question             `json:"questions,omitempty"`
	Stage       entities.OnboardingStage `json:"stage"`
	Progress    float64                `json:"progress"`
	NextAction  string                 `json:"nextAction"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Question represents a question in the onboarding flow
type Question struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "text", "choice", "multiple", "scale", "upload"
	Question    string      `json:"question"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	Options     []Option    `json:"options,omitempty"`
	Validation  *Validation `json:"validation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Option represents a choice option for questions
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Validation represents validation rules for questions
type Validation struct {
	MinLength   *int     `json:"minLength,omitempty"`
	MaxLength   *int     `json:"maxLength,omitempty"`
	Pattern     *string  `json:"pattern,omitempty"`
	MinValue    *float64 `json:"minValue,omitempty"`
	MaxValue    *float64 `json:"maxValue,omitempty"`
	AllowedTypes []string `json:"allowedTypes,omitempty"`
}

// Industry Analysis Objects

// IndustryClassification represents the result of industry classification
type IndustryClassification struct {
	Industry     string  `json:"industry"`
	Category     string  `json:"category"`
	Subcategory  string  `json:"subcategory"`
	Confidence   float64 `json:"confidence"`
	Alternatives []string `json:"alternatives"`
}

// IndustryInsights provides insights about a specific industry
type IndustryInsights struct {
	Industry        string                 `json:"industry"`
	MarketSize      string                 `json:"marketSize"`
	GrowthRate      string                 `json:"growthRate"`
	KeyTrends       []string               `json:"keyTrends"`
	ContentTypes    []string               `json:"contentTypes"`
	Channels        []string               `json:"channels"`
	Challenges      []string               `json:"challenges"`
	Opportunities   []string               `json:"opportunities"`
	BestPractices   []string               `json:"bestPractices"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Competitive Analysis Objects

// WebsiteAnalysis represents analysis of a competitor website
type WebsiteAnalysis struct {
	URL             string                 `json:"url"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	ContentTypes    []string               `json:"contentTypes"`
	PublishingFreq  string                 `json:"publishingFrequency"`
	SocialChannels  []string               `json:"socialChannels"`
	SEOScore        float64                `json:"seoScore"`
	TechnicalStack  []string               `json:"technicalStack"`
	BrandVoice      []string               `json:"brandVoice"`
	ContentTopics   []string               `json:"contentTopics"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ContentStrategy represents a competitor's content strategy
type ContentStrategy struct {
	CompetitorName  string   `json:"competitorName"`
	ContentPillars  []string `json:"contentPillars"`
	PublishingPlan  string   `json:"publishingPlan"`
	TargetAudience  string   `json:"targetAudience"`
	ContentFormats  []string `json:"contentFormats"`
	Distribution    []string `json:"distribution"`
	Engagement      string   `json:"engagement"`
	Strengths       []string `json:"strengths"`
	Weaknesses      []string `json:"weaknesses"`
}

// GapAnalysis identifies opportunities in the competitive landscape
type GapAnalysis struct {
	IdentifiedGaps   []string `json:"identifiedGaps"`
	Opportunities    []string `json:"opportunities"`
	ContentSuggestions []string `json:"contentSuggestions"`
	ChannelGaps      []string `json:"channelGaps"`
	AudienceGaps     []string `json:"audienceGaps"`
	RecommendedFocus []string `json:"recommendedFocus"`
}

// Brand Analysis Objects

// BrandAnalysis represents the result of brand voice analysis
type BrandAnalysis struct {
	Voice           string   `json:"voice"`
	ToneAttributes  []string `json:"toneAttributes"`
	Values          []string `json:"values"`
	Personality     []string `json:"personality"`
	WritingStyle    string   `json:"writingStyle"`
	Vocabulary      []string `json:"vocabulary"`
	MessageThemes   []string `json:"messageThemes"`
	EmotionalTone   string   `json:"emotionalTone"`
	Examples        []string `json:"examples"`
	Recommendations []string `json:"recommendations"`
}

// IndustryAnalysis represents comprehensive industry analysis
type IndustryAnalysis struct {
	Classification *IndustryClassification `json:"classification"`
	Insights       *IndustryInsights       `json:"insights"`
	Competitors    []string                `json:"competitors"`
	ContentGaps    []string                `json:"contentGaps"`
}

// CompetitorAnalysis represents comprehensive competitive analysis
type CompetitorAnalysis struct {
	Websites    []*WebsiteAnalysis `json:"websites"`
	Strategies  []*ContentStrategy `json:"strategies"`
	GapAnalysis *GapAnalysis       `json:"gapAnalysis"`
	Summary     string             `json:"summary"`
}
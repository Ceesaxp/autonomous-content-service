package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OnboardingStage represents the current stage of client onboarding
type OnboardingStage string

const (
	StageInitial     OnboardingStage = "initial"
	StageIndustry    OnboardingStage = "industry"
	StageGoals       OnboardingStage = "goals"
	StageAudience    OnboardingStage = "audience"
	StageStyle       OnboardingStage = "style"
	StageBrand       OnboardingStage = "brand"
	StageCompetitors OnboardingStage = "competitors"
	StageWelcome     OnboardingStage = "welcome"
	StageComplete    OnboardingStage = "complete"
)

// BusinessGoal represents a specific business objective
type BusinessGoal struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// TargetAudience represents demographic and psychographic information
type TargetAudience struct {
	Demographics map[string]interface{} `json:"demographics"`
	Psychographics map[string]interface{} `json:"psychographics"`
	PainPoints     []string              `json:"painPoints"`
	Channels       []string              `json:"channels"`
	Behavior       map[string]interface{} `json:"behavior"`
}

// BrandGuidelines represents brand voice and style guidelines
type BrandGuidelines struct {
	Voice        string   `json:"voice"`
	Tone         []string `json:"tone"`
	Values       []string `json:"values"`
	Personality  []string `json:"personality"`
	DoNotUse     []string `json:"doNotUse"`
	Examples     []string `json:"examples"`
}

// OnboardingSession tracks the client's onboarding progress
type OnboardingSession struct {
	SessionID      uuid.UUID                 `json:"sessionId"`
	ClientID       uuid.UUID                 `json:"clientId"`
	Stage          OnboardingStage           `json:"stage"`
	Responses      map[string]interface{}    `json:"responses"`
	ConversationLog []ConversationMessage    `json:"conversationLog"`
	StartedAt      time.Time                 `json:"startedAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
	CompletedAt    *time.Time                `json:"completedAt,omitempty"`
}

// ConversationMessage represents a single message in the onboarding conversation
type ConversationMessage struct {
	ID        uuid.UUID `json:"id"`
	Speaker   string    `json:"speaker"` // "client" or "system"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ClientProfile contains detailed information about client preferences
type ClientProfile struct {
	ProfileID        uuid.UUID        `json:"profileId"`
	ClientID         uuid.UUID        `json:"clientId"`
	Industry         string           `json:"industry"`
	IndustryCategory string           `json:"industryCategory"`
	CompanySize      string           `json:"companySize"`
	BusinessGoals    []BusinessGoal   `json:"businessGoals"`
	TargetAudience   TargetAudience   `json:"targetAudience"`
	BrandGuidelines  BrandGuidelines  `json:"brandGuidelines"`
	ContentGoals     []string         `json:"contentGoals"`
	StylePreferences map[string]interface{} `json:"stylePreferences"`
	ExampleContent   []string         `json:"exampleContent"`
	CompetitorURLs   []string         `json:"competitorUrls"`
	CompetitorAnalysis map[string]interface{} `json:"competitorAnalysis"`
	OnboardingComplete bool           `json:"onboardingComplete"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// NewOnboardingSession creates a new onboarding session for a client
func NewOnboardingSession(clientID uuid.UUID) *OnboardingSession {
	return &OnboardingSession{
		SessionID:       uuid.New(),
		ClientID:        clientID,
		Stage:           StageInitial,
		Responses:       make(map[string]interface{}),
		ConversationLog: []ConversationMessage{},
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// AddMessage adds a message to the conversation log
func (s *OnboardingSession) AddMessage(speaker, message string, metadata map[string]interface{}) {
	msg := ConversationMessage{
		ID:        uuid.New(),
		Speaker:   speaker,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	s.ConversationLog = append(s.ConversationLog, msg)
	s.UpdatedAt = time.Now()
}

// UpdateStage updates the current onboarding stage
func (s *OnboardingSession) UpdateStage(stage OnboardingStage) {
	s.Stage = stage
	s.UpdatedAt = time.Now()
}

// AddResponse stores a response for the current stage
func (s *OnboardingSession) AddResponse(key string, value interface{}) {
	s.Responses[key] = value
	s.UpdatedAt = time.Now()
}

// Complete marks the onboarding session as complete
func (s *OnboardingSession) Complete() {
	s.Stage = StageComplete
	now := time.Now()
	s.CompletedAt = &now
	s.UpdatedAt = now
}

// NewClientProfile creates a new client profile from onboarding session
func NewClientProfileFromOnboarding(session *OnboardingSession) (*ClientProfile, error) {
	if session.Stage != StageComplete {
		return nil, errors.New("onboarding session must be complete")
	}

	profile := &ClientProfile{
		ProfileID:          uuid.New(),
		ClientID:           session.ClientID,
		OnboardingComplete: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		StylePreferences:   make(map[string]interface{}),
		ExampleContent:     []string{},
		CompetitorURLs:     []string{},
		CompetitorAnalysis: make(map[string]interface{}),
	}

	// Extract data from session responses
	if industry, ok := session.Responses["industry"].(string); ok {
		profile.Industry = industry
	}
	if industryCategory, ok := session.Responses["industryCategory"].(string); ok {
		profile.IndustryCategory = industryCategory
	}
	if companySize, ok := session.Responses["companySize"].(string); ok {
		profile.CompanySize = companySize
	}
	if goals, ok := session.Responses["businessGoals"].([]BusinessGoal); ok {
		profile.BusinessGoals = goals
	}
	if contentGoals, ok := session.Responses["contentGoals"].([]string); ok {
		profile.ContentGoals = contentGoals
	}
	if audience, ok := session.Responses["targetAudience"].(TargetAudience); ok {
		profile.TargetAudience = audience
	}
	if brand, ok := session.Responses["brandGuidelines"].(BrandGuidelines); ok {
		profile.BrandGuidelines = brand
	}
	if competitors, ok := session.Responses["competitorURLs"].([]string); ok {
		profile.CompetitorURLs = competitors
	}

	if err := profile.Validate(); err != nil {
		return nil, err
	}

	return profile, nil
}

// Validate ensures the client profile has all required fields
func (p *ClientProfile) Validate() error {
	if p.Industry == "" {
		return errors.New("industry is required")
	}

	if len(p.ContentGoals) == 0 {
		return errors.New("at least one content goal is required")
	}

	if len(p.BusinessGoals) == 0 {
		return errors.New("at least one business goal is required")
	}

	return nil
}

// UpdateStylePreference updates or adds a style preference
func (p *ClientProfile) UpdateStylePreference(key string, value interface{}) {
	p.StylePreferences[key] = value
	p.UpdateTimestamp()
}

// AddExampleContent adds a URL to example content
func (p *ClientProfile) AddExampleContent(url string) {
	p.ExampleContent = append(p.ExampleContent, url)
	p.UpdateTimestamp()
}

// AddCompetitorURL adds a competitor URL
func (p *ClientProfile) AddCompetitorURL(url string) {
	p.CompetitorURLs = append(p.CompetitorURLs, url)
	p.UpdateTimestamp()
}

// UpdateTimestamp updates the UpdatedAt timestamp to the current time
func (p *ClientProfile) UpdateTimestamp() {
	p.UpdatedAt = time.Now()
}

// UpdateBrandGuidelines updates the brand guidelines
func (p *ClientProfile) UpdateBrandGuidelines(guidelines BrandGuidelines) {
	p.BrandGuidelines = guidelines
	p.UpdateTimestamp()
}

// UpdateTargetAudience updates the target audience information
func (p *ClientProfile) UpdateTargetAudience(audience TargetAudience) {
	p.TargetAudience = audience
	p.UpdateTimestamp()
}

// AddBusinessGoal adds a new business goal
func (p *ClientProfile) AddBusinessGoal(goal BusinessGoal) {
	p.BusinessGoals = append(p.BusinessGoals, goal)
	p.UpdateTimestamp()
}

// UpdateCompetitorAnalysis updates competitive analysis data
func (p *ClientProfile) UpdateCompetitorAnalysis(analysis map[string]interface{}) {
	p.CompetitorAnalysis = analysis
	p.UpdateTimestamp()
}

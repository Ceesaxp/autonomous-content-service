package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ClientProfile contains detailed information about client preferences
type ClientProfile struct {
	ProfileID      uuid.UUID   `json:"profileId"`
	ClientID       uuid.UUID   `json:"clientId"`
	Industry       string      `json:"industry"`
	BrandVoice     string      `json:"brandVoice"`
	TargetAudience string      `json:"targetAudience"`
	ContentGoals   []string    `json:"contentGoals"`
	StylePreferences map[string]interface{} `json:"stylePreferences"`
	ExampleContent []string    `json:"exampleContent"`
	CompetitorURLs []string    `json:"competitorUrls"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

// NewClientProfile creates a new client profile with the given properties
func NewClientProfile(clientID uuid.UUID, industry, brandVoice, targetAudience string, contentGoals []string) (*ClientProfile, error) {
	if len(contentGoals) == 0 {
		return nil, errors.New("at least one content goal is required")
	}

	profile := &ClientProfile{
		ProfileID:      uuid.New(),
		ClientID:       clientID,
		Industry:       industry,
		BrandVoice:     brandVoice,
		TargetAudience: targetAudience,
		ContentGoals:   contentGoals,
		StylePreferences: make(map[string]interface{}),
		ExampleContent: []string{},
		CompetitorURLs: []string{},
		UpdatedAt:      time.Now(),
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

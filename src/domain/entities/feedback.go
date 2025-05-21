package entities

import (
	"time"

	"github.com/google/uuid"
)

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackTypePositive    FeedbackType = "Positive"
	FeedbackTypeNegative    FeedbackType = "Negative"
	FeedbackTypeNeutral     FeedbackType = "Neutral"
	FeedbackTypeSuggestion  FeedbackType = "Suggestion"
	FeedbackTypeComplaint   FeedbackType = "Complaint"
	FeedbackTypeTestimonial FeedbackType = "Testimonial"
)

// FeedbackSource represents where the feedback came from
type FeedbackSource string

const (
	FeedbackSourceClient     FeedbackSource = "Client"
	FeedbackSourceSystem     FeedbackSource = "System"
	FeedbackSourceAutomatic  FeedbackSource = "Automatic"
	FeedbackSourceThirdParty FeedbackSource = "ThirdParty"
)

// FeedbackStatus represents the status of feedback
type FeedbackStatus string

const (
	FeedbackStatusOpen       FeedbackStatus = "Open"
	FeedbackStatusInProgress FeedbackStatus = "InProgress"
	FeedbackStatusResolved   FeedbackStatus = "Resolved"
	FeedbackStatusClosed     FeedbackStatus = "Closed"
	FeedbackStatusDismissed  FeedbackStatus = "Dismissed"
)

// FeedbackRating represents a numeric rating
type FeedbackRating struct {
	Score   float64 `json:"score"`   // 0.0 to 5.0 or 0.0 to 10.0
	MaxScore float64 `json:"maxScore"` // Maximum possible score
}

// Feedback represents feedback on content or projects
type Feedback struct {
	FeedbackID  uuid.UUID             `json:"feedbackId"`
	ContentID   *uuid.UUID            `json:"contentId,omitempty"`
	ProjectID   *uuid.UUID            `json:"projectId,omitempty"`
	ClientID    uuid.UUID             `json:"clientId"`
	Type        FeedbackType          `json:"type"`
	Source      FeedbackSource        `json:"source"`
	Status      FeedbackStatus        `json:"status"`
	Title       string                `json:"title"`
	Message     string                `json:"message"`
	Rating      *FeedbackRating       `json:"rating,omitempty"`
	Tags        []string              `json:"tags"`
	IsResolved  bool                  `json:"isResolved"`
	ResolvedAt  *time.Time            `json:"resolvedAt,omitempty"`
	ResolvedBy  *uuid.UUID            `json:"resolvedBy,omitempty"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewFeedback creates a new feedback instance
func NewFeedback(clientID uuid.UUID, feedbackType FeedbackType, source FeedbackSource, title, message string) *Feedback {
	return &Feedback{
		FeedbackID: uuid.New(),
		ClientID:   clientID,
		Type:       feedbackType,
		Source:     source,
		Status:     FeedbackStatusOpen,
		Title:      title,
		Message:    message,
		Tags:       []string{},
		IsResolved: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}

// SetRating sets the feedback rating
func (f *Feedback) SetRating(score, maxScore float64) {
	f.Rating = &FeedbackRating{
		Score:    score,
		MaxScore: maxScore,
	}
	f.UpdatedAt = time.Now()
}

// AddTag adds a tag to the feedback
func (f *Feedback) AddTag(tag string) {
	f.Tags = append(f.Tags, tag)
	f.UpdatedAt = time.Now()
}

// Resolve marks the feedback as resolved
func (f *Feedback) Resolve(resolvedBy *uuid.UUID) {
	f.IsResolved = true
	now := time.Now()
	f.ResolvedAt = &now
	f.ResolvedBy = resolvedBy
	f.UpdatedAt = time.Now()
}

// AttachToContent associates the feedback with content
func (f *Feedback) AttachToContent(contentID uuid.UUID) {
	f.ContentID = &contentID
	f.UpdatedAt = time.Now()
}

// AttachToProject associates the feedback with a project
func (f *Feedback) AttachToProject(projectID uuid.UUID) {
	f.ProjectID = &projectID
	f.UpdatedAt = time.Now()
}
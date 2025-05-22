package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusDraft      ProjectStatus = "Draft"
	ProjectStatusPlanning   ProjectStatus = "Planning"
	ProjectStatusInProgress ProjectStatus = "InProgress"
	ProjectStatusReview     ProjectStatus = "Review"
	ProjectStatusCompleted  ProjectStatus = "Completed"
	ProjectStatusCancelled  ProjectStatus = "Cancelled"
)

// Priority represents the priority level of a project
type Priority string

const (
	PriorityHigh   Priority = "High"
	PriorityMedium Priority = "Medium"
	PriorityLow    Priority = "Low"
)

// Money represents a monetary amount with currency
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Project represents a content creation project
type Project struct {
	ProjectID    uuid.UUID              `json:"projectId"`
	ClientID     uuid.UUID              `json:"clientId"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	ContentType  ContentType            `json:"contentType"`
	Deadline     time.Time              `json:"deadline"`
	Budget       Money                  `json:"budget"`
	Priority     Priority               `json:"priority"`
	Status       ProjectStatus          `json:"status"`
	Requirements []string               `json:"requirements,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	Contents     []*Content             `json:"contents,omitempty"`
}

// NewProject creates a new project with the given properties
func NewProject(clientID uuid.UUID, title, description string, contentType ContentType, deadline time.Time, budget Money) (*Project, error) {
	project := &Project{
		ProjectID:    uuid.New(),
		ClientID:     clientID,
		Title:        title,
		Description:  description,
		ContentType:  contentType,
		Deadline:     deadline,
		Budget:       budget,
		Priority:     PriorityMedium, // Default priority
		Status:       ProjectStatusDraft,
		Requirements: []string{},
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Contents:     []*Content{},
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	return project, nil
}

// Validate ensures the project has all required fields and valid data
func (p *Project) Validate() error {
	if len(p.Title) < 5 || len(p.Title) > 100 {
		return errors.New("title must be between 5 and 100 characters")
	}

	if p.Description == "" {
		return errors.New("description is required")
	}

	if p.Deadline.Before(time.Now()) {
		return errors.New("deadline must be in the future")
	}

	if p.Budget.Amount <= 0 {
		return errors.New("budget must be greater than zero")
	}

	return nil
}

// AddContent adds a new content item to the project
func (p *Project) AddContent(content *Content) {
	p.Contents = append(p.Contents, content)
	p.UpdateTimestamp()
}

// UpdateStatus changes the project status
func (p *Project) UpdateStatus(status ProjectStatus) {
	p.Status = status
	p.UpdateTimestamp()
}

// UpdatePriority changes the project priority
func (p *Project) UpdatePriority(priority Priority) {
	p.Priority = priority
	p.UpdateTimestamp()
}

// UpdateDeadline changes the project deadline
func (p *Project) UpdateDeadline(deadline time.Time) error {
	if deadline.Before(time.Now()) {
		return errors.New("deadline must be in the future")
	}

	p.Deadline = deadline
	p.UpdateTimestamp()
	return nil
}

// UpdateTimestamp updates the UpdatedAt timestamp to the current time
func (p *Project) UpdateTimestamp() {
	p.UpdatedAt = time.Now()
}

// IsActive returns true if the project is not Completed or Cancelled
func (p *Project) IsActive() bool {
	return p.Status != ProjectStatusCompleted && p.Status != ProjectStatusCancelled
}

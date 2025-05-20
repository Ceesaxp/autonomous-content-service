package repositories

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// FeedbackRepository defines the interface for feedback persistence operations
type FeedbackRepository interface {
	// FindByID retrieves feedback by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Feedback, error)

	// FindByContentID retrieves feedback for a specific content
	FindByContentID(ctx context.Context, contentID uuid.UUID) ([]*entities.Feedback, error)

	// FindByProjectID retrieves feedback for a specific project
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.Feedback, error)

	// FindBySource retrieves feedback by source
	FindBySource(ctx context.Context, source entities.FeedbackSource, offset, limit int) ([]*entities.Feedback, int, error)

	// FindByStatus retrieves feedback by status
	FindByStatus(ctx context.Context, status entities.FeedbackStatus, offset, limit int) ([]*entities.Feedback, int, error)

	// FindByType retrieves feedback by type
	FindByType(ctx context.Context, feedbackType entities.FeedbackType, offset, limit int) ([]*entities.Feedback, int, error)

	// Save persists feedback to the repository
	Save(ctx context.Context, feedback *entities.Feedback) error

	// Create adds new feedback to the repository
	Create(ctx context.Context, feedback *entities.Feedback) error

	// Update updates existing feedback in the repository
	Update(ctx context.Context, feedback *entities.Feedback) error

	// Delete removes feedback from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

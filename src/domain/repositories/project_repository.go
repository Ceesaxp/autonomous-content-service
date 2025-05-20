package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project persistence operations
type ProjectRepository interface {
	// FindByID retrieves a project by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Project, error)

	// FindByClientID retrieves projects for a specific client
	FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]*entities.Project, int, error)

	// FindByStatus retrieves projects by status
	FindByStatus(ctx context.Context, status entities.ProjectStatus, offset, limit int) ([]*entities.Project, int, error)

	// FindActive retrieves all active projects
	FindActive(ctx context.Context, offset, limit int) ([]*entities.Project, int, error)

	// FindByDeadlineRange retrieves projects with deadlines in a specific range
	FindByDeadlineRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Project, int, error)

	// FindAll retrieves all projects with optional pagination
	FindAll(ctx context.Context, offset, limit int) ([]*entities.Project, int, error)

	// Save persists a project to the repository
	Save(ctx context.Context, project *entities.Project) error

	// Create adds a new project to the repository
	Create(ctx context.Context, project *entities.Project) error

	// Update updates an existing project in the repository
	Update(ctx context.Context, project *entities.Project) error

	// Delete removes a project from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

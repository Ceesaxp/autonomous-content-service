package repositories

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// ContentRepository defines the interface for content persistence operations
type ContentRepository interface {
	// FindByID retrieves content by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)

	// FindByProjectID retrieves content for a specific project
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.Content, error)

	// FindByStatus retrieves content by status
	FindByStatus(ctx context.Context, status entities.ContentStatus, offset, limit int) ([]*entities.Content, int, error)

	// FindByType retrieves content by type
	FindByType(ctx context.Context, contentType entities.ContentType, offset, limit int) ([]*entities.Content, int, error)

	// Save persists content to the repository
	Save(ctx context.Context, content *entities.Content) error

	// Create adds new content to the repository
	Create(ctx context.Context, content *entities.Content) error

	// Update updates existing content in the repository
	Update(ctx context.Context, content *entities.Content) error

	// Delete removes content from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

// ContentVersionRepository defines the interface for content version persistence operations
type ContentVersionRepository interface {
	// FindByID retrieves a content version by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ContentVersion, error)

	// FindByContentID retrieves all versions for a specific content
	FindByContentID(ctx context.Context, contentID uuid.UUID) ([]*entities.ContentVersion, error)

	// FindByContentIDAndVersion retrieves a specific version of content
	FindByContentIDAndVersion(ctx context.Context, contentID uuid.UUID, version int) (*entities.ContentVersion, error)

	// Save persists a content version to the repository
	Save(ctx context.Context, version *entities.ContentVersion) error

	// Create adds a new content version to the repository
	Create(ctx context.Context, version *entities.ContentVersion) error
}

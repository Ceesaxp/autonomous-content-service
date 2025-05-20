package repositories

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// SystemCapabilityRepository defines the interface for system capability persistence operations
type SystemCapabilityRepository interface {
	// FindByID retrieves a system capability by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.SystemCapability, error)

	// FindByName retrieves a system capability by name
	FindByName(ctx context.Context, name string) (*entities.SystemCapability, error)

	// FindByType retrieves system capabilities by type
	FindByType(ctx context.Context, capabilityType entities.CapabilityType) ([]*entities.SystemCapability, error)

	// FindByStatus retrieves system capabilities by status
	FindByStatus(ctx context.Context, status entities.CapabilityStatus) ([]*entities.SystemCapability, error)

	// FindAll retrieves all system capabilities
	FindAll(ctx context.Context) ([]*entities.SystemCapability, error)

	// Save persists a system capability to the repository
	Save(ctx context.Context, capability *entities.SystemCapability) error

	// Create adds a new system capability to the repository
	Create(ctx context.Context, capability *entities.SystemCapability) error

	// Update updates an existing system capability in the repository
	Update(ctx context.Context, capability *entities.SystemCapability) error

	// Delete removes a system capability from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

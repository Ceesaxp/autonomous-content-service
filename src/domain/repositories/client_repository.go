package repositories

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// ClientRepository defines the interface for client persistence operations
type ClientRepository interface {
	// FindByID retrieves a client by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Client, error)

	// FindByEmail retrieves a client by email address
	FindByEmail(ctx context.Context, email string) (*entities.Client, error)

	// FindAll retrieves all clients with optional pagination
	FindAll(ctx context.Context, offset, limit int) ([]*entities.Client, int, error)

	// FindByStatus retrieves clients by status
	FindByStatus(ctx context.Context, status entities.ClientStatus, offset, limit int) ([]*entities.Client, int, error)

	// Save persists a client to the repository
	Save(ctx context.Context, client *entities.Client) error

	// Create adds a new client to the repository
	Create(ctx context.Context, client *entities.Client) error

	// Update updates an existing client in the repository
	Update(ctx context.Context, client *entities.Client) error

	// Delete removes a client from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

// ClientProfileRepository defines the interface for client profile persistence operations
type ClientProfileRepository interface {
	// FindByID retrieves a client profile by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ClientProfile, error)

	// FindByClientID retrieves a client profile by client ID
	FindByClientID(ctx context.Context, clientID uuid.UUID) (*entities.ClientProfile, error)

	// Save persists a client profile to the repository
	Save(ctx context.Context, profile *entities.ClientProfile) error

	// Create adds a new client profile to the repository
	Create(ctx context.Context, profile *entities.ClientProfile) error

	// Update updates an existing client profile in the repository
	Update(ctx context.Context, profile *entities.ClientProfile) error

	// Delete removes a client profile from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}

package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// TransactionRepository defines the interface for transaction persistence operations
type TransactionRepository interface {
	// FindByID retrieves a transaction by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Transaction, error)

	// FindByClientID retrieves transactions for a specific client
	FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]*entities.Transaction, int, error)

	// FindByProjectID retrieves transactions for a specific project
	FindByProjectID(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*entities.Transaction, int, error)

	// FindByStatus retrieves transactions by status
	FindByStatus(ctx context.Context, status entities.TransactionStatus, offset, limit int) ([]*entities.Transaction, int, error)

	// FindByType retrieves transactions by type
	FindByType(ctx context.Context, transactionType entities.TransactionType, offset, limit int) ([]*entities.Transaction, int, error)

	// FindByDateRange retrieves transactions within a specific date range
	FindByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Transaction, int, error)

	// Save persists a transaction to the repository
	Save(ctx context.Context, transaction *entities.Transaction) error

	// Create adds a new transaction to the repository
	Create(ctx context.Context, transaction *entities.Transaction) error

	// Update updates an existing transaction in the repository
	Update(ctx context.Context, transaction *entities.Transaction) error
}

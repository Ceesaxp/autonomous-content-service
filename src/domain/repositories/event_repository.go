package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventRepository defines the interface for event persistence operations
type EventRepository interface {
	// Save persists an event to the repository
	Save(ctx context.Context, event interface{}) error

	// FindByID retrieves an event by ID
	FindByID(ctx context.Context, id uuid.UUID) (interface{}, error)

	// FindByType retrieves events by type
	FindByType(ctx context.Context, eventType string, offset, limit int) ([]interface{}, int, error)

	// FindByAggregateID retrieves events for a specific aggregate
	FindByAggregateID(ctx context.Context, aggregateID uuid.UUID, offset, limit int) ([]interface{}, int, error)

	// FindByTimeRange retrieves events within a specific time range
	FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]interface{}, int, error)

	// FindLatest retrieves the latest events
	FindLatest(ctx context.Context, limit int) ([]interface{}, error)
}

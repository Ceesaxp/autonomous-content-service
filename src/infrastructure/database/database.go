package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *sql.DB) error {
	// For now, this is a placeholder
	// In a real implementation, this would run SQL migration scripts
	fmt.Println("Running database migrations...")
	
	// Check if tables exist and create them if they don't
	// This is a simplified approach - in production you'd use a proper migration tool
	
	return nil
}
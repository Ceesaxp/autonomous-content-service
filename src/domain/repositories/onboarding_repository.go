package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// OnboardingRepositoryImpl implements the OnboardingRepository interface
type OnboardingRepositoryImpl struct {
	db *sql.DB
}

// NewOnboardingRepository creates a new onboarding repository
func NewOnboardingRepository(db *sql.DB) *OnboardingRepositoryImpl {
	return &OnboardingRepositoryImpl{db: db}
}

// SaveSession saves an onboarding session to the database
func (r *OnboardingRepositoryImpl) SaveSession(ctx context.Context, session *entities.OnboardingSession) error {
	// Serialize responses and conversation log to JSON
	responsesJSON, err := json.Marshal(session.Responses)
	if err != nil {
		return fmt.Errorf("failed to marshal responses: %w", err)
	}
	
	conversationJSON, err := json.Marshal(session.ConversationLog)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation log: %w", err)
	}
	
	query := `
		INSERT INTO onboarding_sessions (
			session_id, client_id, stage, responses, conversation_log, 
			started_at, updated_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (session_id) DO UPDATE SET
			stage = EXCLUDED.stage,
			responses = EXCLUDED.responses,
			conversation_log = EXCLUDED.conversation_log,
			updated_at = EXCLUDED.updated_at,
			completed_at = EXCLUDED.completed_at`
	
	_, err = r.db.ExecContext(ctx, query,
		session.SessionID,
		session.ClientID,
		string(session.Stage),
		responsesJSON,
		conversationJSON,
		session.StartedAt,
		session.UpdatedAt,
		session.CompletedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save onboarding session: %w", err)
	}
	
	return nil
}

// GetSession retrieves an onboarding session by ID
func (r *OnboardingRepositoryImpl) GetSession(ctx context.Context, sessionID uuid.UUID) (*entities.OnboardingSession, error) {
	query := `
		SELECT session_id, client_id, stage, responses, conversation_log, 
			   started_at, updated_at, completed_at
		FROM onboarding_sessions 
		WHERE session_id = $1`
	
	row := r.db.QueryRowContext(ctx, query, sessionID)
	
	var session entities.OnboardingSession
	var stageStr string
	var responsesJSON, conversationJSON []byte
	
	err := row.Scan(
		&session.SessionID,
		&session.ClientID,
		&stageStr,
		&responsesJSON,
		&conversationJSON,
		&session.StartedAt,
		&session.UpdatedAt,
		&session.CompletedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("onboarding session not found")
		}
		return nil, fmt.Errorf("failed to get onboarding session: %w", err)
	}
	
	// Convert stage string to enum
	session.Stage = entities.OnboardingStage(stageStr)
	
	// Deserialize JSON fields
	if err := json.Unmarshal(responsesJSON, &session.Responses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal responses: %w", err)
	}
	
	if err := json.Unmarshal(conversationJSON, &session.ConversationLog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation log: %w", err)
	}
	
	return &session, nil
}

// DeleteSession removes an onboarding session from the database
func (r *OnboardingRepositoryImpl) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM onboarding_sessions WHERE session_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete onboarding session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("onboarding session not found")
	}
	
	return nil
}

// GetSessionsByClient retrieves all onboarding sessions for a client
func (r *OnboardingRepositoryImpl) GetSessionsByClient(ctx context.Context, clientID uuid.UUID) ([]*entities.OnboardingSession, error) {
	query := `
		SELECT session_id, client_id, stage, responses, conversation_log, 
			   started_at, updated_at, completed_at
		FROM onboarding_sessions 
		WHERE client_id = $1
		ORDER BY started_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to query onboarding sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*entities.OnboardingSession
	
	for rows.Next() {
		var session entities.OnboardingSession
		var stageStr string
		var responsesJSON, conversationJSON []byte
		
		err := rows.Scan(
			&session.SessionID,
			&session.ClientID,
			&stageStr,
			&responsesJSON,
			&conversationJSON,
			&session.StartedAt,
			&session.UpdatedAt,
			&session.CompletedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan onboarding session: %w", err)
		}
		
		// Convert stage string to enum
		session.Stage = entities.OnboardingStage(stageStr)
		
		// Deserialize JSON fields
		if err := json.Unmarshal(responsesJSON, &session.Responses); err != nil {
			return nil, fmt.Errorf("failed to unmarshal responses: %w", err)
		}
		
		if err := json.Unmarshal(conversationJSON, &session.ConversationLog); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conversation log: %w", err)
		}
		
		sessions = append(sessions, &session)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating onboarding sessions: %w", err)
	}
	
	return sessions, nil
}

// GetIncompleteSessionsOlderThan retrieves incomplete sessions older than specified hours
func (r *OnboardingRepositoryImpl) GetIncompleteSessionsOlderThan(ctx context.Context, hours int) ([]*entities.OnboardingSession, error) {
	query := `
		SELECT session_id, client_id, stage, responses, conversation_log, 
			   started_at, updated_at, completed_at
		FROM onboarding_sessions 
		WHERE completed_at IS NULL 
		  AND updated_at < $1
		ORDER BY updated_at ASC`
	
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	rows, err := r.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query incomplete onboarding sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*entities.OnboardingSession
	
	for rows.Next() {
		var session entities.OnboardingSession
		var stageStr string
		var responsesJSON, conversationJSON []byte
		
		err := rows.Scan(
			&session.SessionID,
			&session.ClientID,
			&stageStr,
			&responsesJSON,
			&conversationJSON,
			&session.StartedAt,
			&session.UpdatedAt,
			&session.CompletedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan onboarding session: %w", err)
		}
		
		// Convert stage string to enum
		session.Stage = entities.OnboardingStage(stageStr)
		
		// Deserialize JSON fields
		if err := json.Unmarshal(responsesJSON, &session.Responses); err != nil {
			return nil, fmt.Errorf("failed to unmarshal responses: %w", err)
		}
		
		if err := json.Unmarshal(conversationJSON, &session.ConversationLog); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conversation log: %w", err)
		}
		
		sessions = append(sessions, &session)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incomplete onboarding sessions: %w", err)
	}
	
	return sessions, nil
}

// GetOnboardingStats returns statistics about onboarding sessions
func (r *OnboardingRepositoryImpl) GetOnboardingStats(ctx context.Context, days int) (*OnboardingStats, error) {
	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	
	// Get total sessions
	var totalSessions int
	err := r.db.QueryRowContext(ctx, 
		`SELECT COUNT(*) FROM onboarding_sessions WHERE started_at >= $1`, 
		cutoffTime).Scan(&totalSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sessions: %w", err)
	}
	
	// Get completed sessions
	var completedSessions int
	err = r.db.QueryRowContext(ctx, 
		`SELECT COUNT(*) FROM onboarding_sessions WHERE started_at >= $1 AND completed_at IS NOT NULL`, 
		cutoffTime).Scan(&completedSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed sessions: %w", err)
	}
	
	// Get completion rate
	completionRate := 0.0
	if totalSessions > 0 {
		completionRate = float64(completedSessions) / float64(totalSessions) * 100
	}
	
	// Get average completion time for completed sessions
	var avgCompletionMinutes sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at))/60) 
		FROM onboarding_sessions 
		WHERE started_at >= $1 AND completed_at IS NOT NULL`, 
		cutoffTime).Scan(&avgCompletionMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to get average completion time: %w", err)
	}
	
	avgCompletionTime := "N/A"
	if avgCompletionMinutes.Valid {
		avgCompletionTime = fmt.Sprintf("%.1f minutes", avgCompletionMinutes.Float64)
	}
	
	// Get stage dropoff data
	stageDropoff := make(map[string]int)
	stages := []string{"initial", "industry", "goals", "audience", "style", "brand", "competitors", "welcome"}
	
	for _, stage := range stages {
		var count int
		err = r.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM onboarding_sessions 
			WHERE started_at >= $1 AND stage = $2 AND completed_at IS NULL`, 
			cutoffTime, stage).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get dropoff for stage %s: %w", stage, err)
		}
		stageDropoff[stage] = count
	}
	
	return &OnboardingStats{
		Period:               fmt.Sprintf("Last %d days", days),
		TotalSessions:        totalSessions,
		CompletedSessions:    completedSessions,
		CompletionRate:       completionRate,
		AverageCompletionTime: avgCompletionTime,
		StageDropoff:         stageDropoff,
	}, nil
}

// OnboardingStats represents onboarding analytics data
type OnboardingStats struct {
	Period               string         `json:"period"`
	TotalSessions        int            `json:"totalSessions"`
	CompletedSessions    int            `json:"completedSessions"`
	CompletionRate       float64        `json:"completionRate"`
	AverageCompletionTime string        `json:"averageCompletionTime"`
	StageDropoff         map[string]int `json:"stageDropoff"`
}

// CreateOnboardingTables creates the database tables for onboarding
func CreateOnboardingTables(db *sql.DB) error {
	createSessionsTable := `
		CREATE TABLE IF NOT EXISTS onboarding_sessions (
			session_id UUID PRIMARY KEY,
			client_id UUID NOT NULL,
			stage VARCHAR(50) NOT NULL,
			responses JSONB DEFAULT '{}',
			conversation_log JSONB DEFAULT '[]',
			started_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP NULL,
			CONSTRAINT fk_onboarding_client FOREIGN KEY (client_id) REFERENCES clients(client_id) ON DELETE CASCADE
		);`
	
	createIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_client_id ON onboarding_sessions(client_id);`,
		`CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_stage ON onboarding_sessions(stage);`,
		`CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_started_at ON onboarding_sessions(started_at);`,
		`CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_completed_at ON onboarding_sessions(completed_at);`,
		`CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_updated_at ON onboarding_sessions(updated_at);`,
	}
	
	// Create tables
	if _, err := db.Exec(createSessionsTable); err != nil {
		return fmt.Errorf("failed to create onboarding_sessions table: %w", err)
	}
	
	// Create indexes
	for _, indexSQL := range createIndexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	
	return nil
}
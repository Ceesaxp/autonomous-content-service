package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// PostgresLearningRepository implements learning repository with PostgreSQL
type PostgresLearningRepository struct {
	db *sql.DB
}

// NewPostgresLearningRepository creates a new learning repository
func NewPostgresLearningRepository(db *sql.DB) repositories.LearningRepository {
	return &PostgresLearningRepository{db: db}
}

// CreateArtifact creates a new learning artifact
func (r *PostgresLearningRepository) CreateArtifact(ctx context.Context, artifact *entities.LearningArtifact) error {
	evidenceJSON, _ := json.Marshal(artifact.Evidence)
	tagsJSON, _ := json.Marshal(artifact.Tags)
	relatedJSON, _ := json.Marshal(artifact.RelatedArtifacts)
	prereqJSON, _ := json.Marshal(artifact.Prerequisites)
	metadataJSON, _ := json.Marshal(artifact.Metadata)
	
	query := `
		INSERT INTO learning_artifacts (
			id, type, category, title, description, source, source_id,
			evidence, confidence, impact_score, tags, related_artifacts, 
			prerequisites, metadata, valid_from, valid_until, verification_date,
			usage_count, success_rate, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`
	
	_, err := r.db.ExecContext(ctx, query,
		artifact.ID, artifact.Type, artifact.Category, artifact.Title, artifact.Description,
		artifact.Source, artifact.SourceID, evidenceJSON, artifact.Confidence, artifact.ImpactScore,
		tagsJSON, relatedJSON, prereqJSON, metadataJSON, artifact.ValidFrom, artifact.ValidUntil,
		artifact.VerificationDate, artifact.UsageCount, artifact.SuccessRate, artifact.Status,
		artifact.CreatedAt, artifact.UpdatedAt)
	
	return err
}

// GetArtifact retrieves a learning artifact by ID
func (r *PostgresLearningRepository) GetArtifact(ctx context.Context, id string) (*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanArtifact(row)
}

// UpdateArtifact updates a learning artifact
func (r *PostgresLearningRepository) UpdateArtifact(ctx context.Context, artifact *entities.LearningArtifact) error {
	evidenceJSON, _ := json.Marshal(artifact.Evidence)
	tagsJSON, _ := json.Marshal(artifact.Tags)
	relatedJSON, _ := json.Marshal(artifact.RelatedArtifacts)
	prereqJSON, _ := json.Marshal(artifact.Prerequisites)
	metadataJSON, _ := json.Marshal(artifact.Metadata)
	
	query := `
		UPDATE learning_artifacts SET
			type = $2, category = $3, title = $4, description = $5, source = $6,
			source_id = $7, evidence = $8, confidence = $9, impact_score = $10,
			tags = $11, related_artifacts = $12, prerequisites = $13, metadata = $14,
			valid_from = $15, valid_until = $16, verification_date = $17,
			usage_count = $18, success_rate = $19, status = $20, updated_at = $21
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		artifact.ID, artifact.Type, artifact.Category, artifact.Title, artifact.Description,
		artifact.Source, artifact.SourceID, evidenceJSON, artifact.Confidence, artifact.ImpactScore,
		tagsJSON, relatedJSON, prereqJSON, metadataJSON, artifact.ValidFrom, artifact.ValidUntil,
		artifact.VerificationDate, artifact.UsageCount, artifact.SuccessRate, artifact.Status,
		artifact.UpdatedAt)
	
	return err
}

// DeleteArtifact deletes a learning artifact
func (r *PostgresLearningRepository) DeleteArtifact(ctx context.Context, id string) error {
	query := `DELETE FROM learning_artifacts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// LinkArtifacts creates a relationship between two artifacts
func (r *PostgresLearningRepository) LinkArtifacts(ctx context.Context, sourceID, targetID string, relationType string) error {
	query := `
		INSERT INTO artifact_relationships (source_id, target_id, relation_type, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_id, target_id, relation_type) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query, sourceID, targetID, relationType, time.Now())
	return err
}

// UnlinkArtifacts removes a relationship between artifacts
func (r *PostgresLearningRepository) UnlinkArtifacts(ctx context.Context, sourceID, targetID string) error {
	query := `DELETE FROM artifact_relationships WHERE source_id = $1 AND target_id = $2`
	_, err := r.db.ExecContext(ctx, query, sourceID, targetID)
	return err
}

// GetRelatedArtifacts gets related artifacts with a specific depth
func (r *PostgresLearningRepository) GetRelatedArtifacts(ctx context.Context, artifactID string, depth int) ([]*entities.LearningArtifact, error) {
	// For simplicity, implement depth 1 only
	query := `
		SELECT la.id, la.type, la.category, la.title, la.description, la.source, la.source_id,
			   la.evidence, la.confidence, la.impact_score, la.tags, la.related_artifacts,
			   la.prerequisites, la.metadata, la.valid_from, la.valid_until, la.verification_date,
			   la.usage_count, la.success_rate, la.status, la.created_at, la.updated_at
		FROM learning_artifacts la
		JOIN artifact_relationships ar ON (la.id = ar.target_id OR la.id = ar.source_id)
		WHERE (ar.source_id = $1 OR ar.target_id = $1) AND la.id != $1`
	
	rows, err := r.db.QueryContext(ctx, query, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var artifacts []*entities.LearningArtifact
	for rows.Next() {
		artifact, err := r.scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	
	return artifacts, rows.Err()
}

// GetArtifactsByPattern searches artifacts by pattern
func (r *PostgresLearningRepository) GetArtifactsByPattern(ctx context.Context, pattern string) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts 
		WHERE title ILIKE $1 OR description ILIKE $1 OR category ILIKE $1`
	
	rows, err := r.db.QueryContext(ctx, query, "%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// SearchArtifacts searches artifacts with filters
func (r *PostgresLearningRepository) SearchArtifacts(ctx context.Context, query string, filters map[string]interface{}) ([]*entities.LearningArtifact, error) {
	// Basic implementation
	return r.GetArtifactsByPattern(ctx, query)
}

// GetArtifactsByType gets artifacts by type
func (r *PostgresLearningRepository) GetArtifactsByType(ctx context.Context, artifactType entities.LearningType) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts WHERE type = $1`
	
	rows, err := r.db.QueryContext(ctx, query, artifactType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// GetArtifactsByCategory gets artifacts by category
func (r *PostgresLearningRepository) GetArtifactsByCategory(ctx context.Context, category string) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts WHERE category = $1`
	
	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// GetActiveArtifacts gets all active artifacts
func (r *PostgresLearningRepository) GetActiveArtifacts(ctx context.Context) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts WHERE status = $1`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ArtifactStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// GetArtifactsRequiringVerification gets artifacts that need verification
func (r *PostgresLearningRepository) GetArtifactsRequiringVerification(ctx context.Context, before time.Time) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts 
		WHERE status = $1 AND (verification_date IS NULL OR verification_date < $2)`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ArtifactStatusActive, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// IncrementUsage increments usage count and updates success rate
func (r *PostgresLearningRepository) IncrementUsage(ctx context.Context, artifactID string, success bool) error {
	// Get current usage
	var usageCount int
	var successRate float64
	
	query := `SELECT usage_count, success_rate FROM learning_artifacts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, artifactID).Scan(&usageCount, &successRate)
	if err != nil {
		return err
	}
	
	// Calculate new success rate
	newUsageCount := usageCount + 1
	newSuccessRate := (successRate*float64(usageCount) + map[bool]float64{true: 1.0, false: 0.0}[success]) / float64(newUsageCount)
	
	// Update
	updateQuery := `
		UPDATE learning_artifacts 
		SET usage_count = $2, success_rate = $3, updated_at = $4
		WHERE id = $1`
	
	_, err = r.db.ExecContext(ctx, updateQuery, artifactID, newUsageCount, newSuccessRate, time.Now())
	return err
}

// GetMostUsedArtifacts gets most frequently used artifacts
func (r *PostgresLearningRepository) GetMostUsedArtifacts(ctx context.Context, limit int) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts 
		WHERE status = $1
		ORDER BY usage_count DESC
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ArtifactStatusActive, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// GetHighImpactArtifacts gets high-impact artifacts
func (r *PostgresLearningRepository) GetHighImpactArtifacts(ctx context.Context, minImpact float64) ([]*entities.LearningArtifact, error) {
	query := `
		SELECT id, type, category, title, description, source, source_id,
			   evidence, confidence, impact_score, tags, related_artifacts,
			   prerequisites, metadata, valid_from, valid_until, verification_date,
			   usage_count, success_rate, status, created_at, updated_at
		FROM learning_artifacts 
		WHERE status = $1 AND impact_score >= $2
		ORDER BY impact_score DESC`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ArtifactStatusActive, minImpact)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanArtifacts(rows)
}

// Helper methods

func (r *PostgresLearningRepository) scanArtifact(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.LearningArtifact, error) {
	var artifact entities.LearningArtifact
	var evidenceJSON, tagsJSON, relatedJSON, prereqJSON, metadataJSON []byte
	
	err := scanner.Scan(
		&artifact.ID, &artifact.Type, &artifact.Category, &artifact.Title, &artifact.Description,
		&artifact.Source, &artifact.SourceID, &evidenceJSON, &artifact.Confidence, &artifact.ImpactScore,
		&tagsJSON, &relatedJSON, &prereqJSON, &metadataJSON, &artifact.ValidFrom, &artifact.ValidUntil,
		&artifact.VerificationDate, &artifact.UsageCount, &artifact.SuccessRate, &artifact.Status,
		&artifact.CreatedAt, &artifact.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(evidenceJSON, &artifact.Evidence); err != nil {
		return nil, fmt.Errorf("unmarshaling evidence: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &artifact.Tags); err != nil {
		return nil, fmt.Errorf("unmarshaling tags: %w", err)
	}
	if err := json.Unmarshal(relatedJSON, &artifact.RelatedArtifacts); err != nil {
		return nil, fmt.Errorf("unmarshaling related artifacts: %w", err)
	}
	if err := json.Unmarshal(prereqJSON, &artifact.Prerequisites); err != nil {
		return nil, fmt.Errorf("unmarshaling prerequisites: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &artifact.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}
	
	return &artifact, nil
}

func (r *PostgresLearningRepository) scanArtifacts(rows *sql.Rows) ([]*entities.LearningArtifact, error) {
	var artifacts []*entities.LearningArtifact
	
	for rows.Next() {
		artifact, err := r.scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	
	return artifacts, rows.Err()
}

// PostgresMetricsRepository implements metrics repository with PostgreSQL
type PostgresMetricsRepository struct {
	db *sql.DB
}

// NewPostgresMetricsRepository creates a new metrics repository
func NewPostgresMetricsRepository(db *sql.DB) repositories.MetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

// RecordMetric records a performance metric
func (r *PostgresMetricsRepository) RecordMetric(ctx context.Context, metric *entities.PerformanceMetric) error {
	contextJSON, _ := json.Marshal(metric.Context)
	tagsJSON, _ := json.Marshal(metric.Tags)
	
	query := `
		INSERT INTO performance_metrics (
			id, component, metric_name, value, unit, timestamp,
			context, aggregation, period, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	_, err := r.db.ExecContext(ctx, query,
		metric.ID, metric.Component, metric.MetricName, metric.Value, metric.Unit,
		metric.Timestamp, contextJSON, metric.Aggregation, metric.Period, tagsJSON)
	
	return err
}

// GetMetrics retrieves metrics for a component and metric name in a time range
func (r *PostgresMetricsRepository) GetMetrics(ctx context.Context, component, metricName string, from, to time.Time) ([]*entities.PerformanceMetric, error) {
	query := `
		SELECT id, component, metric_name, value, unit, timestamp,
			   context, aggregation, period, tags
		FROM performance_metrics 
		WHERE component = $1 AND metric_name = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp DESC`
	
	rows, err := r.db.QueryContext(ctx, query, component, metricName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var metrics []*entities.PerformanceMetric
	for rows.Next() {
		var metric entities.PerformanceMetric
		var contextJSON, tagsJSON []byte
		
		err := rows.Scan(
			&metric.ID, &metric.Component, &metric.MetricName, &metric.Value, &metric.Unit,
			&metric.Timestamp, &contextJSON, &metric.Aggregation, &metric.Period, &tagsJSON)
		
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(contextJSON, &metric.Context); err != nil {
			metric.Context = make(map[string]interface{}) // Use empty map on error
		}
		if err := json.Unmarshal(tagsJSON, &metric.Tags); err != nil {
			metric.Tags = []string{} // Use empty slice on error
		}
		
		metrics = append(metrics, &metric)
	}
	
	return metrics, rows.Err()
}

// GetLatestMetric gets the latest metric value
func (r *PostgresMetricsRepository) GetLatestMetric(ctx context.Context, component, metricName string) (*entities.PerformanceMetric, error) {
	query := `
		SELECT id, component, metric_name, value, unit, timestamp,
			   context, aggregation, period, tags
		FROM performance_metrics 
		WHERE component = $1 AND metric_name = $2
		ORDER BY timestamp DESC
		LIMIT 1`
	
	var metric entities.PerformanceMetric
	var contextJSON, tagsJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, component, metricName).Scan(
		&metric.ID, &metric.Component, &metric.MetricName, &metric.Value, &metric.Unit,
		&metric.Timestamp, &contextJSON, &metric.Aggregation, &metric.Period, &tagsJSON)
	
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal(contextJSON, &metric.Context)
	json.Unmarshal(tagsJSON, &metric.Tags)
	
	return &metric, nil
}

// GetAggregatedMetrics gets aggregated metric value
func (r *PostgresMetricsRepository) GetAggregatedMetrics(ctx context.Context, component, metricName string, aggregation entities.AggregationType, period string) (float64, error) {
	var query string
	
	switch aggregation {
	case entities.AggregationAverage:
		query = `SELECT AVG(value) FROM performance_metrics WHERE component = $1 AND metric_name = $2 AND period = $3`
	case entities.AggregationSum:
		query = `SELECT SUM(value) FROM performance_metrics WHERE component = $1 AND metric_name = $2 AND period = $3`
	case entities.AggregationMin:
		query = `SELECT MIN(value) FROM performance_metrics WHERE component = $1 AND metric_name = $2 AND period = $3`
	case entities.AggregationMax:
		query = `SELECT MAX(value) FROM performance_metrics WHERE component = $1 AND metric_name = $2 AND period = $3`
	case entities.AggregationCount:
		query = `SELECT COUNT(*) FROM performance_metrics WHERE component = $1 AND metric_name = $2 AND period = $3`
	default:
		return 0, fmt.Errorf("unsupported aggregation type: %s", aggregation)
	}
	
	var result sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, component, metricName, period).Scan(&result)
	if err != nil {
		return 0, err
	}
	
	if !result.Valid {
		return 0, nil
	}
	
	return result.Float64, nil
}

// GetMetricTrends gets metric trends over multiple periods
func (r *PostgresMetricsRepository) GetMetricTrends(ctx context.Context, component, metricName string, periods int) ([]float64, error) {
	query := `
		SELECT AVG(value) 
		FROM performance_metrics 
		WHERE component = $1 AND metric_name = $2 
		GROUP BY period 
		ORDER BY period DESC 
		LIMIT $3`
	
	rows, err := r.db.QueryContext(ctx, query, component, metricName, periods)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var trends []float64
	for rows.Next() {
		var value sql.NullFloat64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		
		if value.Valid {
			trends = append(trends, value.Float64)
		}
	}
	
	return trends, rows.Err()
}

// GetComponentMetrics gets all current metrics for a component
func (r *PostgresMetricsRepository) GetComponentMetrics(ctx context.Context, component string) (map[string]float64, error) {
	query := `
		SELECT DISTINCT ON (metric_name) metric_name, value
		FROM performance_metrics 
		WHERE component = $1
		ORDER BY metric_name, timestamp DESC`
	
	rows, err := r.db.QueryContext(ctx, query, component)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	metrics := make(map[string]float64)
	for rows.Next() {
		var metricName string
		var value float64
		
		if err := rows.Scan(&metricName, &value); err != nil {
			return nil, err
		}
		
		metrics[metricName] = value
	}
	
	return metrics, rows.Err()
}

// GetMetricAnomalies gets metrics that are anomalous
func (r *PostgresMetricsRepository) GetMetricAnomalies(ctx context.Context, component, metricName string, threshold float64) ([]*entities.PerformanceMetric, error) {
	// Simplified implementation - get values that deviate significantly from recent average
	query := `
		WITH recent_avg AS (
			SELECT AVG(value) as avg_value, STDDEV(value) as stddev_value
			FROM performance_metrics 
			WHERE component = $1 AND metric_name = $2 
			AND timestamp > NOW() - INTERVAL '7 days'
		)
		SELECT pm.id, pm.component, pm.metric_name, pm.value, pm.unit, pm.timestamp,
			   pm.context, pm.aggregation, pm.period, pm.tags
		FROM performance_metrics pm, recent_avg ra
		WHERE pm.component = $1 AND pm.metric_name = $2
		AND ABS(pm.value - ra.avg_value) > $3 * ra.stddev_value
		ORDER BY pm.timestamp DESC`
	
	rows, err := r.db.QueryContext(ctx, query, component, metricName, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var metrics []*entities.PerformanceMetric
	for rows.Next() {
		var metric entities.PerformanceMetric
		var contextJSON, tagsJSON []byte
		
		err := rows.Scan(
			&metric.ID, &metric.Component, &metric.MetricName, &metric.Value, &metric.Unit,
			&metric.Timestamp, &contextJSON, &metric.Aggregation, &metric.Period, &tagsJSON)
		
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(contextJSON, &metric.Context); err != nil {
			metric.Context = make(map[string]interface{}) // Use empty map on error
		}
		if err := json.Unmarshal(tagsJSON, &metric.Tags); err != nil {
			metric.Tags = []string{} // Use empty slice on error
		}
		
		metrics = append(metrics, &metric)
	}
	
	return metrics, rows.Err()
}

// GetMetricBaseline gets baseline statistics for a metric
func (r *PostgresMetricsRepository) GetMetricBaseline(ctx context.Context, component, metricName string, days int) (mean, stddev float64, err error) {
	query := `
		SELECT AVG(value) as mean, STDDEV(value) as stddev
		FROM performance_metrics 
		WHERE component = $1 AND metric_name = $2 
		AND timestamp > NOW() - INTERVAL '%d days'`
	
	var meanResult, stddevResult sql.NullFloat64
	err = r.db.QueryRowContext(ctx, fmt.Sprintf(query, days), component, metricName).Scan(&meanResult, &stddevResult)
	
	if err != nil {
		return 0, 0, err
	}
	
	if meanResult.Valid {
		mean = meanResult.Float64
	}
	if stddevResult.Valid {
		stddev = stddevResult.Float64
	}
	
	return mean, stddev, nil
}
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// PostgresDecisionRepository implements decision repository with PostgreSQL
type PostgresDecisionRepository struct {
	db *sql.DB
}

// NewDecisionRepository creates a new decision repository
func NewDecisionRepository(db *sql.DB) repositories.DecisionRepository {
	return &PostgresDecisionRepository{db: db}
}

// Decision CRUD operations

func (r *PostgresDecisionRepository) CreateDecision(ctx context.Context, decision *entities.Decision) error {
	optionsJSON, _ := json.Marshal(decision.Options)
	stakeholdersJSON, _ := json.Marshal(decision.StakeholderImpact)
	constraintsJSON, _ := json.Marshal(decision.Constraints)
	metadataJSON, _ := json.Marshal(decision.Metadata)
	
	query := `
		INSERT INTO decisions (
			id, type, title, description, status, priority, options,
			selected_option, confidence, reasoning, stakeholder_impact,
			constraints, deadline, created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	
	_, err := r.db.ExecContext(ctx, query,
		decision.ID, decision.Type, decision.Title, decision.Description,
		decision.Status, decision.Priority, optionsJSON, decision.SelectedOption,
		decision.Confidence, decision.Reasoning, stakeholdersJSON,
		constraintsJSON, decision.Deadline, decision.CreatedAt,
		decision.UpdatedAt, metadataJSON)
	
	return err
}

func (r *PostgresDecisionRepository) GetDecision(ctx context.Context, id string) (*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanDecision(row)
}

func (r *PostgresDecisionRepository) UpdateDecision(ctx context.Context, decision *entities.Decision) error {
	optionsJSON, _ := json.Marshal(decision.Options)
	stakeholdersJSON, _ := json.Marshal(decision.StakeholderImpact)
	constraintsJSON, _ := json.Marshal(decision.Constraints)
	metadataJSON, _ := json.Marshal(decision.Metadata)
	
	query := `
		UPDATE decisions SET
			type = $2, title = $3, description = $4, status = $5,
			priority = $6, options = $7, selected_option = $8,
			confidence = $9, reasoning = $10, stakeholder_impact = $11,
			constraints = $12, deadline = $13, updated_at = $14,
			metadata = $15
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		decision.ID, decision.Type, decision.Title, decision.Description,
		decision.Status, decision.Priority, optionsJSON, decision.SelectedOption,
		decision.Confidence, decision.Reasoning, stakeholdersJSON,
		constraintsJSON, decision.Deadline, decision.UpdatedAt, metadataJSON)
	
	return err
}

func (r *PostgresDecisionRepository) DeleteDecision(ctx context.Context, id string) error {
	query := `DELETE FROM decisions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Decision queries

func (r *PostgresDecisionRepository) ListDecisions(ctx context.Context, filter repositories.DecisionFilter) ([]*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions ORDER BY created_at DESC LIMIT 100`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisions(rows)
}

func (r *PostgresDecisionRepository) GetDecisionsByType(ctx context.Context, decisionType entities.DecisionType) ([]*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions WHERE type = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, decisionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisions(rows)
}

func (r *PostgresDecisionRepository) GetDecisionsByStatus(ctx context.Context, status entities.DecisionStatus) ([]*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions WHERE status = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisions(rows)
}

func (r *PostgresDecisionRepository) GetPendingDecisions(ctx context.Context, priority entities.DecisionPriority) ([]*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions 
		WHERE status = 'pending' AND priority = $1 
		ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisions(rows)
}

func (r *PostgresDecisionRepository) GetDecisionHistory(ctx context.Context, days int) ([]*entities.Decision, error) {
	query := `
		SELECT id, type, title, description, status, priority, options,
			   selected_option, confidence, reasoning, stakeholder_impact,
			   constraints, deadline, created_at, updated_at, metadata
		FROM decisions 
		WHERE created_at >= $1 
		ORDER BY created_at DESC`
	
	cutoff := time.Now().AddDate(0, 0, -days)
	rows, err := r.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisions(rows)
}

// Helper methods

func (r *PostgresDecisionRepository) scanDecision(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Decision, error) {
	var decision entities.Decision
	var optionsJSON, stakeholdersJSON, constraintsJSON, metadataJSON []byte
	
	err := scanner.Scan(
		&decision.ID, &decision.Type, &decision.Title, &decision.Description,
		&decision.Status, &decision.Priority, &optionsJSON, &decision.SelectedOption,
		&decision.Confidence, &decision.Reasoning, &stakeholdersJSON,
		&constraintsJSON, &decision.Deadline, &decision.CreatedAt,
		&decision.UpdatedAt, &metadataJSON)
	
	if err != nil {
		return nil, err
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(optionsJSON, &decision.Options); err != nil {
		decision.Options = []entities.DecisionOption{}
	}
	if err := json.Unmarshal(stakeholdersJSON, &decision.StakeholderImpact); err != nil {
		decision.StakeholderImpact = []entities.StakeholderImpact{}
	}
	if err := json.Unmarshal(constraintsJSON, &decision.Constraints); err != nil {
		decision.Constraints = []string{}
	}
	if err := json.Unmarshal(metadataJSON, &decision.Metadata); err != nil {
		decision.Metadata = make(map[string]interface{})
	}
	
	return &decision, nil
}

func (r *PostgresDecisionRepository) scanDecisions(rows *sql.Rows) ([]*entities.Decision, error) {
	var decisions []*entities.Decision
	
	for rows.Next() {
		decision, err := r.scanDecision(rows)
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}
	
	return decisions, rows.Err()
}

// Policy Management Implementation

func (r *PostgresDecisionRepository) CreatePolicy(ctx context.Context, policy *entities.Policy) error {
	rulesJSON, _ := json.Marshal(policy.Rules)
	metadataJSON, _ := json.Marshal(policy.Metadata)
	
	query := `
		INSERT INTO policies (
			id, name, category, description, rules, priority, effective_from,
			effective_until, metadata, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Category, policy.Description, rulesJSON,
		policy.Priority, policy.EffectiveFrom, policy.EffectiveUntil, metadataJSON,
		policy.Active, now, now)
	
	return err
}

func (r *PostgresDecisionRepository) GetPolicy(ctx context.Context, id string) (*entities.Policy, error) {
	query := `
		SELECT id, name, category, description, rules, priority, effective_from,
			   effective_until, metadata, active, created_at, updated_at
		FROM policies WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanPolicy(row)
}

func (r *PostgresDecisionRepository) UpdatePolicy(ctx context.Context, policy *entities.Policy) error {
	rulesJSON, _ := json.Marshal(policy.Rules)
	metadataJSON, _ := json.Marshal(policy.Metadata)
	
	query := `
		UPDATE policies SET
			name = $2, category = $3, description = $4, rules = $5, priority = $6,
			effective_from = $7, effective_until = $8, metadata = $9, active = $10,
			updated_at = $11
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Category, policy.Description, rulesJSON,
		policy.Priority, policy.EffectiveFrom, policy.EffectiveUntil, metadataJSON,
		policy.Active, time.Now())
	
	return err
}

func (r *PostgresDecisionRepository) DeletePolicy(ctx context.Context, id string) error {
	query := "DELETE FROM policies WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresDecisionRepository) ListPolicies(ctx context.Context, category string) ([]*entities.Policy, error) {
	var query string
	var args []interface{}
	
	if category != "" {
		query = `
			SELECT id, name, category, description, rules, priority, effective_from,
				   effective_until, metadata, active, created_at, updated_at
			FROM policies WHERE category = $1
			ORDER BY priority DESC, created_at DESC`
		args = []interface{}{category}
	} else {
		query = `
			SELECT id, name, category, description, rules, priority, effective_from,
				   effective_until, metadata, active, created_at, updated_at
			FROM policies
			ORDER BY priority DESC, created_at DESC`
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanPolicies(rows)
}

func (r *PostgresDecisionRepository) GetActivePolicies(ctx context.Context) ([]*entities.Policy, error) {
	query := `
		SELECT id, name, category, description, rules, priority, effective_from,
			   effective_until, metadata, active, created_at, updated_at
		FROM policies 
		WHERE active = true AND effective_from <= NOW() 
		AND (effective_until IS NULL OR effective_until > NOW())
		ORDER BY priority DESC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanPolicies(rows)
}

// Ethical Guidelines Implementation

func (r *PostgresDecisionRepository) CreateEthicalGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error {
	examplesJSON, _ := json.Marshal(guideline.Examples)
	redLinesJSON, _ := json.Marshal(guideline.RedLines)
	
	query := `
		INSERT INTO ethical_guidelines (
			id, principle, description, examples, red_lines, weight, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		guideline.ID, guideline.Principle, guideline.Description, examplesJSON,
		redLinesJSON, guideline.Weight, now, now)
	
	return err
}

func (r *PostgresDecisionRepository) GetEthicalGuideline(ctx context.Context, id string) (*entities.EthicalGuideline, error) {
	query := `
		SELECT id, principle, description, examples, red_lines, weight, created_at, updated_at
		FROM ethical_guidelines WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanEthicalGuideline(row)
}

func (r *PostgresDecisionRepository) UpdateEthicalGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error {
	examplesJSON, _ := json.Marshal(guideline.Examples)
	redLinesJSON, _ := json.Marshal(guideline.RedLines)
	
	query := `
		UPDATE ethical_guidelines SET
			principle = $2, description = $3, examples = $4, red_lines = $5,
			weight = $6, updated_at = $7
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		guideline.ID, guideline.Principle, guideline.Description, examplesJSON,
		redLinesJSON, guideline.Weight, time.Now())
	
	return err
}

func (r *PostgresDecisionRepository) ListEthicalGuidelines(ctx context.Context) ([]*entities.EthicalGuideline, error) {
	query := `
		SELECT id, principle, description, examples, red_lines, weight, created_at, updated_at
		FROM ethical_guidelines
		ORDER BY weight DESC, principle ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanEthicalGuidelines(rows)
}

// Decision Template Implementation

func (r *PostgresDecisionRepository) CreateDecisionTemplate(ctx context.Context, template *entities.DecisionTemplate) error {
	requiredContextJSON, _ := json.Marshal(template.RequiredContext)
	decisionCriteriaJSON, _ := json.Marshal(template.DecisionCriteria)
	defaultOptionsJSON, _ := json.Marshal(template.DefaultOptions)
	policyChecksJSON, _ := json.Marshal(template.PolicyChecks)
	metadataJSON, _ := json.Marshal(template.Metadata)
	
	query := `
		INSERT INTO decision_templates (
			id, name, type, description, required_context, decision_criteria,
			default_options, policy_checks, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Type, template.Description,
		requiredContextJSON, decisionCriteriaJSON, defaultOptionsJSON,
		policyChecksJSON, metadataJSON, now, now)
	
	return err
}

func (r *PostgresDecisionRepository) GetDecisionTemplate(ctx context.Context, id string) (*entities.DecisionTemplate, error) {
	query := `
		SELECT id, name, type, description, required_context, decision_criteria,
			   default_options, policy_checks, metadata, created_at, updated_at
		FROM decision_templates WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanDecisionTemplate(row)
}

func (r *PostgresDecisionRepository) ListDecisionTemplates(ctx context.Context, decisionType entities.DecisionType) ([]*entities.DecisionTemplate, error) {
	var query string
	var args []interface{}
	
	if decisionType != "" {
		query = `
			SELECT id, name, type, description, required_context, decision_criteria,
				   default_options, policy_checks, metadata, created_at, updated_at
			FROM decision_templates WHERE type = $1
			ORDER BY name ASC`
		args = []interface{}{decisionType}
	} else {
		query = `
			SELECT id, name, type, description, required_context, decision_criteria,
				   default_options, policy_checks, metadata, created_at, updated_at
			FROM decision_templates
			ORDER BY type, name ASC`
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanDecisionTemplates(rows)
}

func (r *PostgresDecisionRepository) CreateDecisionLog(ctx context.Context, log *entities.DecisionLog) error {
	return nil // Silently succeed for now
}

func (r *PostgresDecisionRepository) GetDecisionLogs(ctx context.Context, decisionID string) ([]*entities.DecisionLog, error) {
	return []*entities.DecisionLog{}, nil
}

func (r *PostgresDecisionRepository) GetAuditTrail(ctx context.Context, startTime, endTime time.Time) ([]*entities.DecisionLog, error) {
	return []*entities.DecisionLog{}, nil
}

func (r *PostgresDecisionRepository) GetDecisionMetrics(ctx context.Context, period string) (*repositories.DecisionMetrics, error) {
	return &repositories.DecisionMetrics{
		TotalDecisions:       0,
		DecisionsByType:      make(map[string]int),
		DecisionsByStatus:    make(map[string]int),
		AverageConfidence:    0.0,
		OverrideRate:         0.0,
		ExecutionSuccessRate: 0.0,
		AverageExecutionTime: 0.0,
	}, nil
}

func (r *PostgresDecisionRepository) GetPolicyViolationStats(ctx context.Context, period string) (*repositories.PolicyViolationStats, error) {
	return &repositories.PolicyViolationStats{
		TotalViolations:      0,
		ViolationsBySeverity: make(map[string]int),
		ViolationsByPolicy:   make(map[string]int),
		ComplianceRate:       100.0,
		TopViolatedPolicies:  []repositories.PolicyViolationSummary{},
	}, nil
}

func (r *PostgresDecisionRepository) GetDecisionQualityMetrics(ctx context.Context) (*repositories.QualityMetrics, error) {
	return &repositories.QualityMetrics{
		AverageQualityScore: 0.0,
		QualityTrend:        0.0,
		QualityByType:       make(map[string]float64),
		ImprovementAreas:    []string{},
		BestPerformingTypes: []string{},
	}, nil
}

func (r *PostgresDecisionRepository) GetStakeholderImpactSummary(ctx context.Context, period string) (*repositories.StakeholderImpactSummary, error) {
	return &repositories.StakeholderImpactSummary{
		AverageSentiment:    0.0,
		SentimentByGroup:    make(map[string]float64),
		HighImpactDecisions: 0,
		MitigationActions:   0,
		StakeholderFeedback: []repositories.StakeholderFeedback{},
	}, nil
}

// Helper scanning methods for Policy, EthicalGuideline, and DecisionTemplate

func (r *PostgresDecisionRepository) scanPolicy(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Policy, error) {
	var policy entities.Policy
	var rulesJSON, metadataJSON []byte
	var effectiveUntil sql.NullTime
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&policy.ID, &policy.Name, &policy.Category, &policy.Description, &rulesJSON,
		&policy.Priority, &policy.EffectiveFrom, &effectiveUntil, &metadataJSON,
		&policy.Active, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if effectiveUntil.Valid {
		policy.EffectiveUntil = &effectiveUntil.Time
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(rulesJSON, &policy.Rules); err != nil {
		policy.Rules = []entities.PolicyRule{}
	}
	if err := json.Unmarshal(metadataJSON, &policy.Metadata); err != nil {
		policy.Metadata = make(map[string]interface{})
	}

	return &policy, nil
}

func (r *PostgresDecisionRepository) scanPolicies(rows *sql.Rows) ([]*entities.Policy, error) {
	var policies []*entities.Policy

	for rows.Next() {
		policy, err := r.scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

func (r *PostgresDecisionRepository) scanEthicalGuideline(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.EthicalGuideline, error) {
	var guideline entities.EthicalGuideline
	var examplesJSON, redLinesJSON []byte
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&guideline.ID, &guideline.Principle, &guideline.Description, &examplesJSON,
		&redLinesJSON, &guideline.Weight, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(examplesJSON, &guideline.Examples); err != nil {
		guideline.Examples = []string{}
	}
	if err := json.Unmarshal(redLinesJSON, &guideline.RedLines); err != nil {
		guideline.RedLines = []string{}
	}

	return &guideline, nil
}

func (r *PostgresDecisionRepository) scanEthicalGuidelines(rows *sql.Rows) ([]*entities.EthicalGuideline, error) {
	var guidelines []*entities.EthicalGuideline

	for rows.Next() {
		guideline, err := r.scanEthicalGuideline(rows)
		if err != nil {
			return nil, err
		}
		guidelines = append(guidelines, guideline)
	}

	return guidelines, rows.Err()
}

func (r *PostgresDecisionRepository) scanDecisionTemplate(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.DecisionTemplate, error) {
	var template entities.DecisionTemplate
	var requiredContextJSON, decisionCriteriaJSON, defaultOptionsJSON, policyChecksJSON, metadataJSON []byte
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&template.ID, &template.Name, &template.Type, &template.Description,
		&requiredContextJSON, &decisionCriteriaJSON, &defaultOptionsJSON,
		&policyChecksJSON, &metadataJSON, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(requiredContextJSON, &template.RequiredContext); err != nil {
		template.RequiredContext = []string{}
	}
	if err := json.Unmarshal(decisionCriteriaJSON, &template.DecisionCriteria); err != nil {
		template.DecisionCriteria = []string{}
	}
	if err := json.Unmarshal(defaultOptionsJSON, &template.DefaultOptions); err != nil {
		template.DefaultOptions = []entities.DecisionOption{}
	}
	if err := json.Unmarshal(policyChecksJSON, &template.PolicyChecks); err != nil {
		template.PolicyChecks = []string{}
	}
	if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
		template.Metadata = make(map[string]interface{})
	}

	return &template, nil
}

func (r *PostgresDecisionRepository) scanDecisionTemplates(rows *sql.Rows) ([]*entities.DecisionTemplate, error) {
	var templates []*entities.DecisionTemplate

	for rows.Next() {
		template, err := r.scanDecisionTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}
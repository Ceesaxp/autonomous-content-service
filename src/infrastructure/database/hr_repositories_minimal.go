package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Minimal HR repository implementations that compile

// PostgresMinimalTalentRepository - minimal implementation
type PostgresMinimalTalentRepository struct {
	db *sql.DB
}

func NewTalentRepository(db *sql.DB) repositories.TalentRepository {
	return &PostgresMinimalTalentRepository{db: db}
}

func (r *PostgresMinimalTalentRepository) CreateTalent(ctx context.Context, talent *entities.Talent) error {
	query := `
		INSERT INTO talents (talent_id, type, name, email, status, reputation_score, availability, currency, location, timezone, profile_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	if talent.ID == uuid.Nil {
		talent.ID = uuid.New()
	}
	
	if talent.CreatedAt.IsZero() {
		talent.CreatedAt = time.Now()
	}
	talent.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		talent.ID, talent.Type, talent.Name, talent.Email, talent.Status,
		talent.ReputationScore, talent.Availability, talent.Currency,
		talent.Location, talent.Timezone, talent.ProfileData,
		talent.CreatedAt, talent.UpdatedAt)
	
	return err
}

func (r *PostgresMinimalTalentRepository) GetTalentByID(ctx context.Context, id uuid.UUID) (*entities.Talent, error) {
	query := `
		SELECT talent_id, type, name, email, status, reputation_score, 
		       availability, currency, location, timezone, profile_data,
		       last_active_at, onboarded_at, offboarded_at, created_at, updated_at
		FROM talents WHERE talent_id = $1`
	
	talent := &entities.Talent{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
		&talent.ReputationScore, &talent.Availability, &talent.Currency,
		&talent.Location, &talent.Timezone, &talent.ProfileData,
		&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
		&talent.CreatedAt, &talent.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return talent, nil
}

func (r *PostgresMinimalTalentRepository) GetTalentByEmail(ctx context.Context, email string) (*entities.Talent, error) {
	query := `
		SELECT talent_id, type, name, email, status, reputation_score, 
		       availability, currency, location, timezone, profile_data,
		       last_active_at, onboarded_at, offboarded_at, created_at, updated_at
		FROM talents WHERE email = $1`
	
	talent := &entities.Talent{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
		&talent.ReputationScore, &talent.Availability, &talent.Currency,
		&talent.Location, &talent.Timezone, &talent.ProfileData,
		&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
		&talent.CreatedAt, &talent.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return talent, nil
}

func (r *PostgresMinimalTalentRepository) UpdateTalent(ctx context.Context, talent *entities.Talent) error {
	query := `
		UPDATE talents SET 
			type = $2, name = $3, email = $4, status = $5, reputation_score = $6,
			availability = $7, currency = $8, location = $9, timezone = $10,
			profile_data = $11, last_active_at = $12, onboarded_at = $13,
			offboarded_at = $14, updated_at = $15
		WHERE talent_id = $1`
	
	talent.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		talent.ID, talent.Type, talent.Name, talent.Email, talent.Status,
		talent.ReputationScore, talent.Availability, talent.Currency,
		talent.Location, talent.Timezone, talent.ProfileData,
		talent.LastActiveAt, talent.OnboardedAt, talent.OffboardedAt,
		talent.UpdatedAt)
	
	return err
}

func (r *PostgresMinimalTalentRepository) DeleteTalent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM talents WHERE talent_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresMinimalTalentRepository) SearchTalent(ctx context.Context, filter repositories.TalentFilter) ([]*entities.Talent, int, error) {
	query := `
		SELECT talent_id, type, name, email, status, reputation_score, 
		       availability, currency, location, timezone, profile_data,
		       last_active_at, onboarded_at, offboarded_at, created_at, updated_at
		FROM talents WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.Type != nil {
		query += ` AND type = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}
	
	if filter.Status != nil {
		query += ` AND status = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	
	if filter.Location != nil {
		query += ` AND location ILIKE $` + fmt.Sprintf("%d", argIndex)
		args = append(args, "%"+*filter.Location+"%")
		argIndex++
	}
	
	if filter.MinReputation != nil {
		query += ` AND reputation_score >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.MinReputation)
		argIndex++
	}
	
	if filter.Search != "" {
		query += ` AND (name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR email ILIKE $` + fmt.Sprintf("%d", argIndex) + `)`
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}
	
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	if filter.SortBy != "" {
		query += ` ORDER BY ` + filter.SortBy
		if filter.SortOrder == "desc" {
			query += ` DESC`
		} else {
			query += ` ASC`
		}
	} else {
		query += ` ORDER BY reputation_score DESC`
	}
	
	if filter.Limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	
	if filter.Offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		talents = append(talents, talent)
	}
	
	return talents, total, rows.Err()
}

func (r *PostgresMinimalTalentRepository) GetTalentBySkills(ctx context.Context, skills []string, minLevel entities.SkillLevel) ([]*entities.Talent, error) {
	if len(skills) == 0 {
		return []*entities.Talent{}, nil
	}
	
	query := `
		SELECT DISTINCT t.talent_id, t.type, t.name, t.email, t.status, t.reputation_score, 
		       t.availability, t.currency, t.location, t.timezone, t.profile_data,
		       t.last_active_at, t.onboarded_at, t.offboarded_at, t.created_at, t.updated_at
		FROM talents t
		JOIN skills s ON t.talent_id = s.talent_id
		WHERE s.name = ANY($1) AND s.level >= $2
		GROUP BY t.talent_id, t.type, t.name, t.email, t.status, t.reputation_score, 
		         t.availability, t.currency, t.location, t.timezone, t.profile_data,
		         t.last_active_at, t.onboarded_at, t.offboarded_at, t.created_at, t.updated_at
		HAVING COUNT(DISTINCT s.name) = $3
		ORDER BY t.reputation_score DESC`
	
	rows, err := r.db.QueryContext(ctx, query, pq.Array(skills), minLevel, len(skills))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		talents = append(talents, talent)
	}
	
	return talents, rows.Err()
}

func (r *PostgresMinimalTalentRepository) GetAvailableTalent(ctx context.Context, talentType entities.TalentType) ([]*entities.Talent, error) {
	query := `
		SELECT talent_id, type, name, email, status, reputation_score, 
		       availability, currency, location, timezone, profile_data,
		       last_active_at, onboarded_at, offboarded_at, created_at, updated_at
		FROM talents 
		WHERE type = $1 AND status = 'Available'
		ORDER BY reputation_score DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		talents = append(talents, talent)
	}
	
	return talents, rows.Err()
}

func (r *PostgresMinimalTalentRepository) UpdateReputationScore(ctx context.Context, talentID uuid.UUID, score float64) error {
	query := `UPDATE talents SET reputation_score = $2, updated_at = $3 WHERE talent_id = $1`
	_, err := r.db.ExecContext(ctx, query, talentID, score, time.Now())
	return err
}

func (r *PostgresMinimalTalentRepository) GetTopTalentByScore(ctx context.Context, limit int) ([]*entities.Talent, error) {
	query := `
		SELECT talent_id, type, name, email, status, reputation_score, 
		       availability, currency, location, timezone, profile_data,
		       last_active_at, onboarded_at, offboarded_at, created_at, updated_at
		FROM talents 
		WHERE status != 'Offboarded'
		ORDER BY reputation_score DESC 
		LIMIT $1`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		talents = append(talents, talent)
	}
	
	return talents, rows.Err()
}

func (r *PostgresMinimalTalentRepository) AddTalentSkill(ctx context.Context, skill *entities.Skill) error {
	query := `
		INSERT INTO skills (skill_id, talent_id, name, category, level, years_used, last_used, 
		                    verified, verified_by, verified_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	if skill.ID == uuid.Nil {
		skill.ID = uuid.New()
	}
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		skill.ID, skill.TalentID, skill.Name, skill.Category, skill.Level,
		skill.YearsUsed, skill.LastUsed, skill.Verified, skill.VerifiedBy,
		skill.VerifiedAt, skill.CreatedAt)
	
	return err
}

func (r *PostgresMinimalTalentRepository) UpdateTalentSkill(ctx context.Context, skill *entities.Skill) error {
	query := `
		UPDATE skills SET 
			name = $2, category = $3, level = $4, years_used = $5, last_used = $6,
			verified = $7, verified_by = $8, verified_at = $9
		WHERE skill_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		skill.ID, skill.Name, skill.Category, skill.Level, skill.YearsUsed,
		skill.LastUsed, skill.Verified, skill.VerifiedBy, skill.VerifiedAt)
	
	return err
}

func (r *PostgresMinimalTalentRepository) RemoveTalentSkill(ctx context.Context, talentID, skillID uuid.UUID) error {
	query := `DELETE FROM skills WHERE talent_id = $1 AND skill_id = $2`
	_, err := r.db.ExecContext(ctx, query, talentID, skillID)
	return err
}

func (r *PostgresMinimalTalentRepository) GetTalentSkills(ctx context.Context, talentID uuid.UUID) ([]*entities.Skill, error) {
	query := `
		SELECT skill_id, talent_id, name, category, level, years_used, last_used,
		       verified, verified_by, verified_at, created_at
		FROM skills WHERE talent_id = $1 ORDER BY level DESC, name ASC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var skills []*entities.Skill
	for rows.Next() {
		skill := &entities.Skill{}
		err := rows.Scan(
			&skill.ID, &skill.TalentID, &skill.Name, &skill.Category, &skill.Level,
			&skill.YearsUsed, &skill.LastUsed, &skill.Verified, &skill.VerifiedBy,
			&skill.VerifiedAt, &skill.CreatedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	
	return skills, rows.Err()
}

func (r *PostgresMinimalTalentRepository) AddTalentCertification(ctx context.Context, cert *entities.Certification) error {
	query := `
		INSERT INTO certifications (certification_id, talent_id, name, issuer, credential_id,
		                            issue_date, expiry_date, verification_url, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}
	if cert.CreatedAt.IsZero() {
		cert.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		cert.ID, cert.TalentID, cert.Name, cert.Issuer, cert.CredentialID,
		cert.IssueDate, cert.ExpiryDate, cert.VerificationURL, cert.IsActive, cert.CreatedAt)
	
	return err
}

func (r *PostgresMinimalTalentRepository) UpdateTalentCertification(ctx context.Context, cert *entities.Certification) error {
	query := `
		UPDATE certifications SET 
			name = $2, issuer = $3, credential_id = $4, issue_date = $5,
			expiry_date = $6, verification_url = $7, is_active = $8
		WHERE certification_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		cert.ID, cert.Name, cert.Issuer, cert.CredentialID, cert.IssueDate,
		cert.ExpiryDate, cert.VerificationURL, cert.IsActive)
	
	return err
}

func (r *PostgresMinimalTalentRepository) GetTalentCertifications(ctx context.Context, talentID uuid.UUID) ([]*entities.Certification, error) {
	query := `
		SELECT certification_id, talent_id, name, issuer, credential_id, issue_date,
		       expiry_date, verification_url, is_active, created_at
		FROM certifications WHERE talent_id = $1 ORDER BY issue_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var certs []*entities.Certification
	for rows.Next() {
		cert := &entities.Certification{}
		err := rows.Scan(
			&cert.ID, &cert.TalentID, &cert.Name, &cert.Issuer, &cert.CredentialID,
			&cert.IssueDate, &cert.ExpiryDate, &cert.VerificationURL, &cert.IsActive, &cert.CreatedAt)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	
	return certs, rows.Err()
}

func (r *PostgresMinimalTalentRepository) GetExpiringCertifications(ctx context.Context, beforeDate time.Time) ([]*entities.Certification, error) {
	query := `
		SELECT certification_id, talent_id, name, issuer, credential_id, issue_date,
		       expiry_date, verification_url, is_active, created_at
		FROM certifications 
		WHERE expiry_date IS NOT NULL AND expiry_date <= $1 AND is_active = true
		ORDER BY expiry_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var certs []*entities.Certification
	for rows.Next() {
		cert := &entities.Certification{}
		err := rows.Scan(
			&cert.ID, &cert.TalentID, &cert.Name, &cert.Issuer, &cert.CredentialID,
			&cert.IssueDate, &cert.ExpiryDate, &cert.VerificationURL, &cert.IsActive, &cert.CreatedAt)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	
	return certs, rows.Err()
}

// Minimal implementations for other repositories - only basic CRUD operations

type PostgresMinimalEngagementRepository struct{ db *sql.DB }

// GetEngagementMetrics implements repositories.EngagementRepository.
func (r *PostgresMinimalEngagementRepository) GetEngagementMetrics(ctx context.Context, engagementID uuid.UUID) (map[string]interface{}, error) {
	query := `SELECT performance_metrics FROM engagements WHERE engagement_id = $1`
	var metrics map[string]interface{}
	err := r.db.QueryRowContext(ctx, query, engagementID).Scan(&metrics)
	if err != nil {
		if err == sql.ErrNoRows {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	return metrics, nil
}

// GetEngagementsEndingSoon implements repositories.EngagementRepository.
func (r *PostgresMinimalEngagementRepository) GetEngagementsEndingSoon(ctx context.Context, beforeDate time.Time) ([]*entities.Engagement, error) {
	query := `
		SELECT engagement_id, talent_id, type, status, title, description, start_date, end_date,
		       hours_per_week, rate_type, currency, contract_id, manager_id, team_id,
		       performance_metrics, created_at, updated_at
		FROM engagements 
		WHERE end_date IS NOT NULL AND end_date <= $1 AND status = 'Active'
		ORDER BY end_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var engagements []*entities.Engagement
	for rows.Next() {
		engagement := &entities.Engagement{}
		err := rows.Scan(
			&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
			&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
			&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
			&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
			&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		engagements = append(engagements, engagement)
	}
	return engagements, rows.Err()
}

// UpdateEngagementMetrics implements repositories.EngagementRepository.
func (r *PostgresMinimalEngagementRepository) UpdateEngagementMetrics(ctx context.Context, engagementID uuid.UUID, metrics map[string]interface{}) error {
	query := `UPDATE engagements SET performance_metrics = $2, updated_at = $3 WHERE engagement_id = $1`
	_, err := r.db.ExecContext(ctx, query, engagementID, metrics, time.Now())
	return err
}

func NewEngagementRepository(db *sql.DB) repositories.EngagementRepository {
	return &PostgresMinimalEngagementRepository{db: db}
}
func (r *PostgresMinimalEngagementRepository) CreateEngagement(ctx context.Context, engagement *entities.Engagement) error {
	query := `
		INSERT INTO engagements (engagement_id, talent_id, type, status, title, description, 
		                         start_date, end_date, hours_per_week, rate_type, currency,
		                         contract_id, manager_id, team_id, performance_metrics, 
		                         created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	
	if engagement.ID == uuid.Nil {
		engagement.ID = uuid.New()
	}
	if engagement.CreatedAt.IsZero() {
		engagement.CreatedAt = time.Now()
	}
	engagement.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		engagement.ID, engagement.TalentID, engagement.Type, engagement.Status,
		engagement.Title, engagement.Description, engagement.StartDate, engagement.EndDate,
		engagement.HoursPerWeek, engagement.RateType, engagement.Currency,
		engagement.ContractID, engagement.ManagerID, engagement.TeamID,
		engagement.PerformanceMetrics, engagement.CreatedAt, engagement.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalEngagementRepository) GetEngagementByID(ctx context.Context, id uuid.UUID) (*entities.Engagement, error) {
	query := `
		SELECT engagement_id, talent_id, type, status, title, description, start_date, end_date,
		       hours_per_week, rate_type, currency, contract_id, manager_id, team_id,
		       performance_metrics, created_at, updated_at
		FROM engagements WHERE engagement_id = $1`
	
	engagement := &entities.Engagement{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
		&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
		&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
		&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
		&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return engagement, nil
}
func (r *PostgresMinimalEngagementRepository) UpdateEngagement(ctx context.Context, engagement *entities.Engagement) error {
	query := `
		UPDATE engagements SET 
			title = $2, description = $3, status = $4, end_date = $5, hours_per_week = $6,
			rate_type = $7, currency = $8, manager_id = $9, team_id = $10,
			performance_metrics = $11, updated_at = $12
		WHERE engagement_id = $1`
	
	engagement.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		engagement.ID, engagement.Title, engagement.Description, engagement.Status,
		engagement.EndDate, engagement.HoursPerWeek, engagement.RateType, engagement.Currency,
		engagement.ManagerID, engagement.TeamID, engagement.PerformanceMetrics, engagement.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalEngagementRepository) DeleteEngagement(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM engagements WHERE engagement_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
func (r *PostgresMinimalEngagementRepository) ListEngagements(ctx context.Context, filter repositories.EngagementFilter) ([]*entities.Engagement, int, error) {
	query := `
		SELECT engagement_id, talent_id, type, status, title, description, start_date, end_date,
		       hours_per_week, rate_type, currency, contract_id, manager_id, team_id,
		       performance_metrics, created_at, updated_at
		FROM engagements WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.TalentID != nil {
		query += ` AND talent_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.TalentID)
		argIndex++
	}
	
	if filter.Type != nil {
		query += ` AND type = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}
	
	if filter.Status != nil {
		query += ` AND status = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var engagements []*entities.Engagement
	for rows.Next() {
		engagement := &entities.Engagement{}
		err := rows.Scan(
			&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
			&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
			&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
			&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
			&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		engagements = append(engagements, engagement)
	}
	
	return engagements, total, rows.Err()
}
func (r *PostgresMinimalEngagementRepository) GetEngagementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.Engagement, error) {
	query := `
		SELECT engagement_id, talent_id, type, status, title, description, start_date, end_date,
		       hours_per_week, rate_type, currency, contract_id, manager_id, team_id,
		       performance_metrics, created_at, updated_at
		FROM engagements WHERE talent_id = $1 ORDER BY start_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var engagements []*entities.Engagement
	for rows.Next() {
		engagement := &entities.Engagement{}
		err := rows.Scan(
			&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
			&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
			&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
			&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
			&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		engagements = append(engagements, engagement)
	}
	
	return engagements, rows.Err()
}
func (r *PostgresMinimalEngagementRepository) GetEngagementsByProject(ctx context.Context, projectID uuid.UUID) ([]*entities.Engagement, error) {
	query := `
		SELECT DISTINCT e.engagement_id, e.talent_id, e.type, e.status, e.title, e.description, 
		       e.start_date, e.end_date, e.hours_per_week, e.rate_type, e.currency, e.contract_id, 
		       e.manager_id, e.team_id, e.performance_metrics, e.created_at, e.updated_at
		FROM engagements e
		JOIN work_assignments wa ON e.engagement_id = wa.engagement_id
		WHERE wa.project_id = $1
		ORDER BY e.start_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var engagements []*entities.Engagement
	for rows.Next() {
		engagement := &entities.Engagement{}
		err := rows.Scan(
			&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
			&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
			&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
			&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
			&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		engagements = append(engagements, engagement)
	}
	
	return engagements, rows.Err()
}
func (r *PostgresMinimalEngagementRepository) GetActiveEngagements(ctx context.Context) ([]*entities.Engagement, error) {
	query := `
		SELECT engagement_id, talent_id, type, status, title, description, start_date, end_date,
		       hours_per_week, rate_type, currency, contract_id, manager_id, team_id,
		       performance_metrics, created_at, updated_at
		FROM engagements 
		WHERE status = 'Active' AND (end_date IS NULL OR end_date > NOW())
		ORDER BY start_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var engagements []*entities.Engagement
	for rows.Next() {
		engagement := &entities.Engagement{}
		err := rows.Scan(
			&engagement.ID, &engagement.TalentID, &engagement.Type, &engagement.Status,
			&engagement.Title, &engagement.Description, &engagement.StartDate, &engagement.EndDate,
			&engagement.HoursPerWeek, &engagement.RateType, &engagement.Currency,
			&engagement.ContractID, &engagement.ManagerID, &engagement.TeamID,
			&engagement.PerformanceMetrics, &engagement.CreatedAt, &engagement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		engagements = append(engagements, engagement)
	}
	
	return engagements, rows.Err()
}
func (r *PostgresMinimalEngagementRepository) UpdateEngagementStatus(ctx context.Context, engagementID uuid.UUID, status entities.EngagementStatus) error {
	query := `UPDATE engagements SET status = $2, updated_at = $3 WHERE engagement_id = $1`
	_, err := r.db.ExecContext(ctx, query, engagementID, status, time.Now())
	return err
}

type PostgresMinimalWorkAssignmentRepository struct{ db *sql.DB }

// CreateDeliverable implements repositories.WorkAssignmentRepository.
func (r *PostgresMinimalWorkAssignmentRepository) CreateDeliverable(ctx context.Context, deliverable *entities.Deliverable) error {
	query := `
		INSERT INTO deliverables (deliverable_id, assignment_id, name, description, type, status, 
		                          file_url, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if deliverable.ID == uuid.Nil {
		deliverable.ID = uuid.New()
	}
	if deliverable.CreatedAt.IsZero() {
		deliverable.CreatedAt = time.Now()
	}
	deliverable.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		deliverable.ID, deliverable.AssignmentID, deliverable.Name, deliverable.Description,
		deliverable.Type, deliverable.Status, deliverable.FileURL, deliverable.Metadata,
		deliverable.CreatedAt, deliverable.UpdatedAt)
	
	return err
}

// GetDeliverableByID implements repositories.WorkAssignmentRepository.
func (r *PostgresMinimalWorkAssignmentRepository) GetDeliverableByID(ctx context.Context, id uuid.UUID) (*entities.Deliverable, error) {
	query := `
		SELECT deliverable_id, assignment_id, name, description, type, status, file_url, metadata,
		       submitted_at, accepted_at, rejected_at, rejection_reason, created_at, updated_at
		FROM deliverables WHERE deliverable_id = $1`
	
	deliverable := &entities.Deliverable{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deliverable.ID, &deliverable.AssignmentID, &deliverable.Name, &deliverable.Description,
		&deliverable.Type, &deliverable.Status, &deliverable.FileURL, &deliverable.Metadata,
		&deliverable.SubmittedAt, &deliverable.AcceptedAt, &deliverable.RejectedAt,
		&deliverable.RejectionReason, &deliverable.CreatedAt, &deliverable.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return deliverable, nil
}

// GetDeliverablesByAssignment implements repositories.WorkAssignmentRepository.
func (r *PostgresMinimalWorkAssignmentRepository) GetDeliverablesByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*entities.Deliverable, error) {
	query := `
		SELECT deliverable_id, assignment_id, name, description, type, status, file_url, metadata,
		       submitted_at, accepted_at, rejected_at, rejection_reason, created_at, updated_at
		FROM deliverables WHERE assignment_id = $1 ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var deliverables []*entities.Deliverable
	for rows.Next() {
		deliverable := &entities.Deliverable{}
		err := rows.Scan(
			&deliverable.ID, &deliverable.AssignmentID, &deliverable.Name, &deliverable.Description,
			&deliverable.Type, &deliverable.Status, &deliverable.FileURL, &deliverable.Metadata,
			&deliverable.SubmittedAt, &deliverable.AcceptedAt, &deliverable.RejectedAt,
			&deliverable.RejectionReason, &deliverable.CreatedAt, &deliverable.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deliverables = append(deliverables, deliverable)
	}
	return deliverables, rows.Err()
}

// GetPendingDeliverables implements repositories.WorkAssignmentRepository.
func (r *PostgresMinimalWorkAssignmentRepository) GetPendingDeliverables(ctx context.Context) ([]*entities.Deliverable, error) {
	query := `
		SELECT deliverable_id, assignment_id, name, description, type, status, file_url, metadata,
		       submitted_at, accepted_at, rejected_at, rejection_reason, created_at, updated_at
		FROM deliverables WHERE status = 'Pending' OR status = 'Submitted' ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var deliverables []*entities.Deliverable
	for rows.Next() {
		deliverable := &entities.Deliverable{}
		err := rows.Scan(
			&deliverable.ID, &deliverable.AssignmentID, &deliverable.Name, &deliverable.Description,
			&deliverable.Type, &deliverable.Status, &deliverable.FileURL, &deliverable.Metadata,
			&deliverable.SubmittedAt, &deliverable.AcceptedAt, &deliverable.RejectedAt,
			&deliverable.RejectionReason, &deliverable.CreatedAt, &deliverable.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deliverables = append(deliverables, deliverable)
	}
	return deliverables, rows.Err()
}

// UpdateDeliverable implements repositories.WorkAssignmentRepository.
func (r *PostgresMinimalWorkAssignmentRepository) UpdateDeliverable(ctx context.Context, deliverable *entities.Deliverable) error {
	query := `
		UPDATE deliverables SET 
			name = $2, description = $3, type = $4, status = $5, file_url = $6, metadata = $7,
			submitted_at = $8, accepted_at = $9, rejected_at = $10, rejection_reason = $11,
			updated_at = $12
		WHERE deliverable_id = $1`
	
	deliverable.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		deliverable.ID, deliverable.Name, deliverable.Description, deliverable.Type,
		deliverable.Status, deliverable.FileURL, deliverable.Metadata,
		deliverable.SubmittedAt, deliverable.AcceptedAt, deliverable.RejectedAt,
		deliverable.RejectionReason, deliverable.UpdatedAt)
	
	return err
}

func NewWorkAssignmentRepository(db *sql.DB) repositories.WorkAssignmentRepository {
	return &PostgresMinimalWorkAssignmentRepository{db: db}
}
func (r *PostgresMinimalWorkAssignmentRepository) CreateAssignment(ctx context.Context, assignment *entities.WorkAssignment) error {
	query := `
		INSERT INTO work_assignments (assignment_id, engagement_id, talent_id, project_id, title, 
		                              description, priority, status, estimated_hours, actual_hours,
		                              due_date, quality_score, feedback_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	if assignment.CreatedAt.IsZero() {
		assignment.CreatedAt = time.Now()
	}
	assignment.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.EngagementID, assignment.TalentID, assignment.ProjectID,
		assignment.Title, assignment.Description, assignment.Priority, assignment.Status,
		assignment.EstimatedHours, assignment.ActualHours, assignment.DueDate,
		assignment.QualityScore, assignment.FeedbackNotes, assignment.CreatedAt, assignment.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalWorkAssignmentRepository) GetAssignmentByID(ctx context.Context, id uuid.UUID) (*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments WHERE assignment_id = $1`
	
	assignment := &entities.WorkAssignment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
		&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
		&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
		&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
		&assignment.CreatedAt, &assignment.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return assignment, nil
}
func (r *PostgresMinimalWorkAssignmentRepository) UpdateAssignment(ctx context.Context, assignment *entities.WorkAssignment) error {
	query := `
		UPDATE work_assignments SET 
			title = $2, description = $3, priority = $4, status = $5, estimated_hours = $6,
			actual_hours = $7, due_date = $8, completed_at = $9, quality_score = $10,
			feedback_notes = $11, updated_at = $12
		WHERE assignment_id = $1`
	
	assignment.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.Title, assignment.Description, assignment.Priority,
		assignment.Status, assignment.EstimatedHours, assignment.ActualHours, assignment.DueDate,
		assignment.CompletedAt, assignment.QualityScore, assignment.FeedbackNotes, assignment.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalWorkAssignmentRepository) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM work_assignments WHERE assignment_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
func (r *PostgresMinimalWorkAssignmentRepository) ListAssignments(ctx context.Context, filter repositories.AssignmentFilter) ([]*entities.WorkAssignment, int, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.TalentID != nil {
		query += ` AND talent_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.TalentID)
		argIndex++
	}
	
	if filter.EngagementID != nil {
		query += ` AND engagement_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.EngagementID)
		argIndex++
	}
	
	if filter.Status != nil {
		query += ` AND status = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		assignments = append(assignments, assignment)
	}
	
	return assignments, total, rows.Err()
}
func (r *PostgresMinimalWorkAssignmentRepository) GetAssignmentsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments WHERE talent_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	
	return assignments, rows.Err()
}
func (r *PostgresMinimalWorkAssignmentRepository) GetAssignmentsByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments WHERE engagement_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, engagementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	
	return assignments, rows.Err()
}
func (r *PostgresMinimalWorkAssignmentRepository) GetActiveAssignments(ctx context.Context, talentID uuid.UUID) ([]*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments 
		WHERE talent_id = $1 AND status IN ('Active', 'InProgress', 'Assigned')
		ORDER BY priority DESC, due_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	
	return assignments, rows.Err()
}
func (r *PostgresMinimalWorkAssignmentRepository) UpdateAssignmentStatus(ctx context.Context, assignmentID uuid.UUID, status string) error {
	query := `UPDATE work_assignments SET status = $2, updated_at = $3 WHERE assignment_id = $1`
	_, err := r.db.ExecContext(ctx, query, assignmentID, status, time.Now())
	return err
}
func (r *PostgresMinimalWorkAssignmentRepository) GetOverdueAssignments(ctx context.Context) ([]*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments 
		WHERE due_date < $1 AND status NOT IN ('Completed', 'Cancelled')
		ORDER BY due_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

type PostgresMinimalPerformanceRepository struct{ db *sql.DB }

// GetPerformanceReview implements repositories.PerformanceRepository.
func (r *PostgresMinimalPerformanceRepository) GetPerformanceReview(ctx context.Context, id uuid.UUID) (*entities.PerformanceReview, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews WHERE review_id = $1`
	
	review := &entities.PerformanceReview{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
		&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
		&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
		&review.CommunicationScore, &review.Comments, &review.Metrics,
		&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
		&review.NextReviewDate, &review.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return review, nil
}

// GetPerformanceReviewsByPeriod implements repositories.PerformanceRepository.
func (r *PostgresMinimalPerformanceRepository) GetPerformanceReviewsByPeriod(ctx context.Context, start time.Time, end time.Time) ([]*entities.PerformanceReview, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews 
		WHERE review_period_start >= $1 AND review_period_end <= $2
		ORDER BY review_period_start ASC`
	
	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reviews []*entities.PerformanceReview
	for rows.Next() {
		review := &entities.PerformanceReview{}
		err := rows.Scan(
			&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
			&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
			&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
			&review.CommunicationScore, &review.Comments, &review.Metrics,
			&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
			&review.NextReviewDate, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

// GetPerformanceReviewsByTalent implements repositories.PerformanceRepository.
func (r *PostgresMinimalPerformanceRepository) GetPerformanceReviewsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.PerformanceReview, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews 
		WHERE talent_id = $1
		ORDER BY review_period_start DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reviews []*entities.PerformanceReview
	for rows.Next() {
		review := &entities.PerformanceReview{}
		err := rows.Scan(
			&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
			&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
			&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
			&review.CommunicationScore, &review.Comments, &review.Metrics,
			&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
			&review.NextReviewDate, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

// GetReviewsDue implements repositories.PerformanceRepository.
func (r *PostgresMinimalPerformanceRepository) GetReviewsDue(ctx context.Context, beforeDate time.Time) ([]*entities.PerformanceReview, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews 
		WHERE next_review_date IS NOT NULL AND next_review_date <= $1
		ORDER BY next_review_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reviews []*entities.PerformanceReview
	for rows.Next() {
		review := &entities.PerformanceReview{}
		err := rows.Scan(
			&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
			&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
			&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
			&review.CommunicationScore, &review.Comments, &review.Metrics,
			&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
			&review.NextReviewDate, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func NewPerformanceRepository(db *sql.DB) repositories.PerformanceRepository {
	return &PostgresMinimalPerformanceRepository{db: db}
}
func (r *PostgresMinimalPerformanceRepository) GetAllPerformanceMetrics(ctx context.Context, timeRange repositories.TimeRange) (map[uuid.UUID]*repositories.TalentPerformanceMetrics, error) {
	return nil, nil
}
func (r *PostgresMinimalPerformanceRepository) CreatePerformanceReview(ctx context.Context, review *entities.PerformanceReview) error {
	query := `
		INSERT INTO performance_reviews (review_id, talent_id, engagement_id, reviewer_id,
		                                  review_period_start, review_period_end, overall_rating,
		                                  quality_score, productivity_score, reliability_score,
		                                  communication_score, comments, metrics,
		                                  compensation_adjustment_amount, compensation_adjustment_currency,
		                                  next_review_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	if review.CreatedAt.IsZero() {
		review.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.TalentID, review.EngagementID, review.ReviewerID,
		review.ReviewPeriodStart, review.ReviewPeriodEnd, review.OverallRating,
		review.QualityScore, review.ProductivityScore, review.ReliabilityScore,
		review.CommunicationScore, review.Comments, review.Metrics,
		review.CompensationAdjustmentAmount, review.CompensationAdjustmentCurrency,
		review.NextReviewDate, review.CreatedAt)
	
	return err
}
func (r *PostgresMinimalPerformanceRepository) GetPerformanceReviewByID(ctx context.Context, id uuid.UUID) (*entities.PerformanceReview, error) {
	return r.GetPerformanceReview(ctx, id)
}
func (r *PostgresMinimalPerformanceRepository) UpdatePerformanceReview(ctx context.Context, review *entities.PerformanceReview) error {
	query := `
		UPDATE performance_reviews SET 
			overall_rating = $2, quality_score = $3, productivity_score = $4,
			reliability_score = $5, communication_score = $6, comments = $7,
			metrics = $8, compensation_adjustment_amount = $9,
			compensation_adjustment_currency = $10, next_review_date = $11
		WHERE review_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.OverallRating, review.QualityScore, review.ProductivityScore,
		review.ReliabilityScore, review.CommunicationScore, review.Comments,
		review.Metrics, review.CompensationAdjustmentAmount,
		review.CompensationAdjustmentCurrency, review.NextReviewDate)
	
	return err
}
func (r *PostgresMinimalPerformanceRepository) DeletePerformanceReview(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM performance_reviews WHERE review_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
func (r *PostgresMinimalPerformanceRepository) ListPerformanceReviews(ctx context.Context, filter interface{}) ([]*entities.PerformanceReview, int, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews ORDER BY created_at DESC`
	
	countQuery := `SELECT COUNT(*) FROM performance_reviews`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var reviews []*entities.PerformanceReview
	for rows.Next() {
		review := &entities.PerformanceReview{}
		err := rows.Scan(
			&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
			&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
			&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
			&review.CommunicationScore, &review.Comments, &review.Metrics,
			&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
			&review.NextReviewDate, &review.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, review)
	}
	
	return reviews, total, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) GetReviewsByTalent(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) ([]*entities.PerformanceReview, error) {
	query := `
		SELECT review_id, talent_id, engagement_id, reviewer_id, review_period_start, review_period_end,
		       overall_rating, quality_score, productivity_score, reliability_score, communication_score,
		       comments, metrics, compensation_adjustment_amount, compensation_adjustment_currency,
		       next_review_date, created_at
		FROM performance_reviews 
		WHERE talent_id = $1 AND review_period_start >= $2 AND review_period_end <= $3
		ORDER BY review_period_start DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reviews []*entities.PerformanceReview
	for rows.Next() {
		review := &entities.PerformanceReview{}
		err := rows.Scan(
			&review.ID, &review.TalentID, &review.EngagementID, &review.ReviewerID,
			&review.ReviewPeriodStart, &review.ReviewPeriodEnd, &review.OverallRating,
			&review.QualityScore, &review.ProductivityScore, &review.ReliabilityScore,
			&review.CommunicationScore, &review.Comments, &review.Metrics,
			&review.CompensationAdjustmentAmount, &review.CompensationAdjustmentCurrency,
			&review.NextReviewDate, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	
	return reviews, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) CreatePerformanceMetric(ctx context.Context, metric *entities.PerformanceMetric) error {
	query := `
		INSERT INTO performance_metrics (metric_id, talent_id, type, value, unit, description, 
		                                 source, context, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		metric.ID, metric.TalentID, metric.Type, metric.Value, metric.Unit,
		metric.Description, metric.Source, metric.Context, metric.CreatedAt)
	
	return err
}
func (r *PostgresMinimalPerformanceRepository) GetPerformanceMetrics(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) (*repositories.TalentPerformanceMetrics, error) {
	query := `
		SELECT metric_id, talent_id, type, value, unit, description, source, context, created_at
		FROM performance_metrics 
		WHERE talent_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	metrics := &repositories.TalentPerformanceMetrics{
		TalentID: talentID,
	}
	
	// Aggregate metrics from raw data
	var qualitySum, productivitySum, reliabilitySum float64
	var qualityCount, productivityCount, reliabilityCount int
	
	for rows.Next() {
		metric := &entities.PerformanceMetric{}
		err := rows.Scan(
			&metric.ID, &metric.TalentID, &metric.Type, &metric.Value, &metric.Unit,
			&metric.Description, &metric.Source, &metric.Context, &metric.CreatedAt)
		if err != nil {
			return nil, err
		}
		
		// Aggregate by metric type
		switch metric.Type {
		case "quality":
			qualitySum += metric.Value
			qualityCount++
		case "productivity":
			productivitySum += metric.Value
			productivityCount++
		case "reliability":
			reliabilitySum += metric.Value
			reliabilityCount++
		}
	}
	
	// Calculate averages
	if qualityCount > 0 {
		metrics.AverageQualityScore = qualitySum / float64(qualityCount)
	}
	if productivityCount > 0 {
		metrics.AverageProductivityScore = productivitySum / float64(productivityCount)
	}
	if reliabilityCount > 0 {
		metrics.AverageReliabilityScore = reliabilitySum / float64(reliabilityCount)
	}
	
	return metrics, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) CreatePerformanceGoal(ctx context.Context, goal *entities.PerformanceGoal) error {
	query := `
		INSERT INTO performance_goals (goal_id, talent_id, title, description, type, target_value,
		                               current_value, unit, priority, due_date, status, metrics,
		                               created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	if goal.ID == uuid.Nil {
		goal.ID = uuid.New()
	}
	if goal.CreatedAt.IsZero() {
		goal.CreatedAt = time.Now()
	}
	goal.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		goal.ID, goal.TalentID, goal.Title, goal.Description, goal.Type,
		goal.TargetValue, goal.CurrentValue, goal.Unit, goal.Priority,
		goal.DueDate, goal.Status, goal.Metrics, goal.CreatedAt, goal.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalPerformanceRepository) GetPerformanceGoal(ctx context.Context, goalID uuid.UUID) (*entities.PerformanceGoal, error) {
	query := `
		SELECT goal_id, talent_id, title, description, type, target_value, current_value,
		       unit, priority, due_date, status, metrics, created_at, updated_at
		FROM performance_goals WHERE goal_id = $1`
	
	goal := &entities.PerformanceGoal{}
	err := r.db.QueryRowContext(ctx, query, goalID).Scan(
		&goal.ID, &goal.TalentID, &goal.Title, &goal.Description, &goal.Type,
		&goal.TargetValue, &goal.CurrentValue, &goal.Unit, &goal.Priority,
		&goal.DueDate, &goal.Status, &goal.Metrics, &goal.CreatedAt, &goal.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return goal, nil
}
func (r *PostgresMinimalPerformanceRepository) UpdatePerformanceGoal(ctx context.Context, goal *entities.PerformanceGoal) error {
	query := `
		UPDATE performance_goals SET 
			title = $2, description = $3, type = $4, target_value = $5, current_value = $6,
			unit = $7, priority = $8, due_date = $9, status = $10, metrics = $11, updated_at = $12
		WHERE goal_id = $1`
	
	goal.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		goal.ID, goal.Title, goal.Description, goal.Type, goal.TargetValue,
		goal.CurrentValue, goal.Unit, goal.Priority, goal.DueDate, goal.Status,
		goal.Metrics, goal.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalPerformanceRepository) GetPerformanceGoalsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.PerformanceGoal, error) {
	query := `
		SELECT goal_id, talent_id, title, description, type, target_value, current_value,
		       unit, priority, due_date, status, metrics, created_at, updated_at
		FROM performance_goals WHERE talent_id = $1 ORDER BY due_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var goals []*entities.PerformanceGoal
	for rows.Next() {
		goal := &entities.PerformanceGoal{}
		err := rows.Scan(
			&goal.ID, &goal.TalentID, &goal.Title, &goal.Description, &goal.Type,
			&goal.TargetValue, &goal.CurrentValue, &goal.Unit, &goal.Priority,
			&goal.DueDate, &goal.Status, &goal.Metrics, &goal.CreatedAt, &goal.UpdatedAt)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	
	return goals, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) GetTalentPerformanceMetrics(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) (*repositories.TalentPerformanceMetrics, error) {
	return r.GetPerformanceMetrics(ctx, talentID, timeRange)
}
func (r *PostgresMinimalPerformanceRepository) GetPerformanceDistribution(ctx context.Context, timeRange repositories.TimeRange) (*repositories.PerformanceDistribution, error) {
	query := `
		SELECT overall_rating, COUNT(*) as count
		FROM performance_reviews 
		WHERE review_period_start >= $1 AND review_period_end <= $2
		GROUP BY overall_rating
		ORDER BY overall_rating`
	
	rows, err := r.db.QueryContext(ctx, query, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	distribution := &repositories.PerformanceDistribution{}
	
	for rows.Next() {
		var rating string
		var count int
		err := rows.Scan(&rating, &count)
		if err != nil {
			return nil, err
		}
		
		// Map ratings to struct fields
		switch rating {
		case "Exceptional":
			distribution.Exceptional = count
		case "ExceedsExpectations":
			distribution.ExceedsExpectations = count
		case "MeetsExpectations":
			distribution.MeetsExpectations = count
		case "NeedsImprovement":
			distribution.NeedsImprovement = count
		case "Unsatisfactory":
			distribution.Unsatisfactory = count
		}
		distribution.TotalReviews += count
	}
	
	return distribution, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) GetTopPerformers(ctx context.Context, criteria repositories.PerformanceCriteria) ([]*entities.Talent, error) {
	query := `
		SELECT DISTINCT t.talent_id, t.type, t.name, t.email, t.status, t.reputation_score, 
		       t.availability, t.currency, t.location, t.timezone, t.profile_data,
		       t.last_active_at, t.onboarded_at, t.offboarded_at, t.created_at, t.updated_at
		FROM talents t
		JOIN performance_reviews pr ON t.talent_id = pr.talent_id
		WHERE pr.review_period_start >= $1 AND pr.review_period_end <= $2
		      AND pr.overall_rating IN ('Exceptional', 'ExceedsExpectations')
		ORDER BY t.reputation_score DESC
		LIMIT $3`
	
	limit := 10
	if criteria.Limit > 0 {
		limit = criteria.Limit
	}
	
	rows, err := r.db.QueryContext(ctx, query, criteria.TimeRange.Start, criteria.TimeRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		talents = append(talents, talent)
	}
	
	return talents, rows.Err()
}
func (r *PostgresMinimalPerformanceRepository) GetUnderperformers(ctx context.Context, criteria repositories.PerformanceCriteria) ([]*entities.Talent, error) {
	query := `
		SELECT DISTINCT t.talent_id, t.type, t.name, t.email, t.status, t.reputation_score, 
		       t.availability, t.currency, t.location, t.timezone, t.profile_data,
		       t.last_active_at, t.onboarded_at, t.offboarded_at, t.created_at, t.updated_at
		FROM talents t
		JOIN performance_reviews pr ON t.talent_id = pr.talent_id
		WHERE pr.review_period_start >= $1 AND pr.review_period_end <= $2
		      AND pr.overall_rating IN ('NeedsImprovement', 'Unsatisfactory')
		ORDER BY t.reputation_score ASC
		LIMIT $3`
	
	limit := 10
	if criteria.Limit > 0 {
		limit = criteria.Limit
	}
	
	rows, err := r.db.QueryContext(ctx, query, criteria.TimeRange.Start, criteria.TimeRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var talents []*entities.Talent
	for rows.Next() {
		talent := &entities.Talent{}
		err := rows.Scan(
			&talent.ID, &talent.Type, &talent.Name, &talent.Email, &talent.Status,
			&talent.ReputationScore, &talent.Availability, &talent.Currency,
			&talent.Location, &talent.Timezone, &talent.ProfileData,
			&talent.LastActiveAt, &talent.OnboardedAt, &talent.OffboardedAt,
			&talent.CreatedAt, &talent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		talents = append(talents, talent)
	}
	
	return talents, rows.Err()
}

// Minimal stub implementations for remaining repositories

type PostgresMinimalCompensationRepository struct{ db *sql.DB }

// GetPayrollRecordsByTalentAndYear implements repositories.CompensationRepository.
func (r *PostgresMinimalCompensationRepository) GetPayrollRecordsByTalentAndYear(ctx context.Context, talentID uuid.UUID, taxYear int) ([]*entities.PayrollRecord, error) {
	query := `
		SELECT payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		       currency, hours_worked, deductions, bonuses, payment_date, payment_method,
		       transaction_id, status, created_at
		FROM payroll_records 
		WHERE talent_id = $1 AND EXTRACT(YEAR FROM payment_date) = $2
		ORDER BY payment_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID, taxYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*entities.PayrollRecord
	for rows.Next() {
		record := &entities.PayrollRecord{}
		err := rows.Scan(
			&record.ID, &record.TalentID, &record.EngagementID, &record.PayPeriodStart,
			&record.PayPeriodEnd, &record.Currency, &record.HoursWorked, &record.Deductions,
			&record.Bonuses, &record.PaymentDate, &record.PaymentMethod, &record.TransactionID,
			&record.Status, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// GetPendingPayroll implements repositories.CompensationRepository.
func (r *PostgresMinimalCompensationRepository) GetPendingPayroll(ctx context.Context) ([]*entities.PayrollRecord, error) {
	query := `
		SELECT payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		       currency, hours_worked, deductions, bonuses, payment_date, payment_method,
		       transaction_id, status, created_at
		FROM payroll_records 
		WHERE status = 'Pending' OR status = 'Processing'
		ORDER BY payment_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*entities.PayrollRecord
	for rows.Next() {
		record := &entities.PayrollRecord{}
		err := rows.Scan(
			&record.ID, &record.TalentID, &record.EngagementID, &record.PayPeriodStart,
			&record.PayPeriodEnd, &record.Currency, &record.HoursWorked, &record.Deductions,
			&record.Bonuses, &record.PaymentDate, &record.PaymentMethod, &record.TransactionID,
			&record.Status, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// GetTalentCompensationHistory implements repositories.CompensationRepository.
func (r *PostgresMinimalCompensationRepository) GetTalentCompensationHistory(ctx context.Context, talentID uuid.UUID) ([]*entities.PayrollRecord, error) {
	query := `
		SELECT payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		       currency, hours_worked, deductions, bonuses, payment_date, payment_method,
		       transaction_id, status, created_at
		FROM payroll_records 
		WHERE talent_id = $1
		ORDER BY payment_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*entities.PayrollRecord
	for rows.Next() {
		record := &entities.PayrollRecord{}
		err := rows.Scan(
			&record.ID, &record.TalentID, &record.EngagementID, &record.PayPeriodStart,
			&record.PayPeriodEnd, &record.Currency, &record.HoursWorked, &record.Deductions,
			&record.Bonuses, &record.PaymentDate, &record.PaymentMethod, &record.TransactionID,
			&record.Status, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// GetWorkAssignmentsForPeriod implements repositories.CompensationRepository.
func (r *PostgresMinimalCompensationRepository) GetWorkAssignmentsForPeriod(ctx context.Context, talentID uuid.UUID, startDate time.Time, endDate time.Time) ([]*entities.WorkAssignment, error) {
	query := `
		SELECT assignment_id, engagement_id, talent_id, project_id, title, description, 
		       priority, status, estimated_hours, actual_hours, due_date, completed_at,
		       quality_score, feedback_notes, created_at, updated_at
		FROM work_assignments 
		WHERE talent_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assignments []*entities.WorkAssignment
	for rows.Next() {
		assignment := &entities.WorkAssignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.EngagementID, &assignment.TalentID, &assignment.ProjectID,
			&assignment.Title, &assignment.Description, &assignment.Priority, &assignment.Status,
			&assignment.EstimatedHours, &assignment.ActualHours, &assignment.DueDate,
			&assignment.CompletedAt, &assignment.QualityScore, &assignment.FeedbackNotes,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func NewCompensationRepository(db *sql.DB) repositories.CompensationRepository {
	return &PostgresMinimalCompensationRepository{db: db}
}
func (r *PostgresMinimalCompensationRepository) CreateCompensationPlan(ctx context.Context, plan *entities.CompensationPlan) error {
	query := `
		INSERT INTO compensation_plans (compensation_plan_id, talent_id, engagement_id, type, currency,
		                                payment_frequency, bonus_structure, effective_date, end_date,
		                                tax_withholding, payment_method_id, smart_contract_addr,
		                                created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now()
	}
	plan.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.TalentID, plan.EngagementID, plan.Type, plan.Currency,
		plan.PaymentFrequency, plan.BonusStructure, plan.EffectiveDate, plan.EndDate,
		plan.TaxWithholding, plan.PaymentMethodID, plan.SmartContractAddr,
		plan.CreatedAt, plan.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalCompensationRepository) GetCompensationPlan(ctx context.Context, id uuid.UUID) (*entities.CompensationPlan, error) {
	query := `
		SELECT compensation_plan_id, talent_id, engagement_id, type, currency, payment_frequency,
		       bonus_structure, effective_date, end_date, tax_withholding, payment_method_id,
		       smart_contract_addr, created_at, updated_at
		FROM compensation_plans WHERE compensation_plan_id = $1`
	
	plan := &entities.CompensationPlan{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&plan.ID, &plan.TalentID, &plan.EngagementID, &plan.Type, &plan.Currency,
		&plan.PaymentFrequency, &plan.BonusStructure, &plan.EffectiveDate, &plan.EndDate,
		&plan.TaxWithholding, &plan.PaymentMethodID, &plan.SmartContractAddr,
		&plan.CreatedAt, &plan.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return plan, nil
}
func (r *PostgresMinimalCompensationRepository) GetCompensationPlanByID(ctx context.Context, id uuid.UUID) (*entities.CompensationPlan, error) {
	return r.GetCompensationPlan(ctx, id)
}
func (r *PostgresMinimalCompensationRepository) GetCompensationPlanByTalent(ctx context.Context, talentID uuid.UUID) (*entities.CompensationPlan, error) {
	return r.GetActiveCompensationPlan(ctx, talentID)
}
func (r *PostgresMinimalCompensationRepository) UpdateCompensationPlan(ctx context.Context, plan *entities.CompensationPlan) error {
	query := `
		UPDATE compensation_plans SET 
			type = $2, currency = $3, payment_frequency = $4, bonus_structure = $5,
			end_date = $6, tax_withholding = $7, payment_method_id = $8,
			smart_contract_addr = $9, updated_at = $10
		WHERE compensation_plan_id = $1`
	
	plan.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.Type, plan.Currency, plan.PaymentFrequency, plan.BonusStructure,
		plan.EndDate, plan.TaxWithholding, plan.PaymentMethodID, plan.SmartContractAddr,
		plan.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalCompensationRepository) GetCompensationPlansByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.CompensationPlan, error) {
	query := `
		SELECT compensation_plan_id, talent_id, engagement_id, type, currency, payment_frequency,
		       bonus_structure, effective_date, end_date, tax_withholding, payment_method_id,
		       smart_contract_addr, created_at, updated_at
		FROM compensation_plans WHERE talent_id = $1 ORDER BY effective_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var plans []*entities.CompensationPlan
	for rows.Next() {
		plan := &entities.CompensationPlan{}
		err := rows.Scan(
			&plan.ID, &plan.TalentID, &plan.EngagementID, &plan.Type, &plan.Currency,
			&plan.PaymentFrequency, &plan.BonusStructure, &plan.EffectiveDate, &plan.EndDate,
			&plan.TaxWithholding, &plan.PaymentMethodID, &plan.SmartContractAddr,
			&plan.CreatedAt, &plan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	
	return plans, rows.Err()
}
func (r *PostgresMinimalCompensationRepository) GetActiveCompensationPlan(ctx context.Context, talentID uuid.UUID) (*entities.CompensationPlan, error) {
	query := `
		SELECT compensation_plan_id, talent_id, engagement_id, type, currency, payment_frequency,
		       bonus_structure, effective_date, end_date, tax_withholding, payment_method_id,
		       smart_contract_addr, created_at, updated_at
		FROM compensation_plans 
		WHERE talent_id = $1 AND effective_date <= NOW() 
		      AND (end_date IS NULL OR end_date > NOW())
		ORDER BY effective_date DESC LIMIT 1`
	
	plan := &entities.CompensationPlan{}
	err := r.db.QueryRowContext(ctx, query, talentID).Scan(
		&plan.ID, &plan.TalentID, &plan.EngagementID, &plan.Type, &plan.Currency,
		&plan.PaymentFrequency, &plan.BonusStructure, &plan.EffectiveDate, &plan.EndDate,
		&plan.TaxWithholding, &plan.PaymentMethodID, &plan.SmartContractAddr,
		&plan.CreatedAt, &plan.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return plan, nil
}
func (r *PostgresMinimalCompensationRepository) CreatePayrollRecord(ctx context.Context, record *entities.PayrollRecord) error {
	query := `
		INSERT INTO payroll_records (payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		                             currency, hours_worked, deductions, bonuses, payment_date,
		                             payment_method, transaction_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.TalentID, record.EngagementID, record.PayPeriodStart, record.PayPeriodEnd,
		record.Currency, record.HoursWorked, record.Deductions, record.Bonuses, record.PaymentDate,
		record.PaymentMethod, record.TransactionID, record.Status, record.CreatedAt)
	
	return err
}
func (r *PostgresMinimalCompensationRepository) GetPayrollRecord(ctx context.Context, id uuid.UUID) (*entities.PayrollRecord, error) {
	query := `
		SELECT payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		       currency, hours_worked, deductions, bonuses, payment_date, payment_method,
		       transaction_id, status, created_at
		FROM payroll_records WHERE payroll_id = $1`
	
	record := &entities.PayrollRecord{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID, &record.TalentID, &record.EngagementID, &record.PayPeriodStart,
		&record.PayPeriodEnd, &record.Currency, &record.HoursWorked, &record.Deductions,
		&record.Bonuses, &record.PaymentDate, &record.PaymentMethod, &record.TransactionID,
		&record.Status, &record.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return record, nil
}
func (r *PostgresMinimalCompensationRepository) GetPayrollRecordByID(ctx context.Context, id uuid.UUID) (*entities.PayrollRecord, error) {
	return r.GetPayrollRecord(ctx, id)
}
func (r *PostgresMinimalCompensationRepository) UpdatePayrollRecord(ctx context.Context, record *entities.PayrollRecord) error {
	query := `
		UPDATE payroll_records SET 
			hours_worked = $2, deductions = $3, bonuses = $4, payment_date = $5,
			payment_method = $6, transaction_id = $7, status = $8
		WHERE payroll_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.HoursWorked, record.Deductions, record.Bonuses,
		record.PaymentDate, record.PaymentMethod, record.TransactionID, record.Status)
	
	return err
}
func (r *PostgresMinimalCompensationRepository) GetPayrollRecordsByTalent(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) ([]*entities.PayrollRecord, error) {
	query := `
		SELECT payroll_id, talent_id, engagement_id, pay_period_start, pay_period_end,
		       currency, hours_worked, deductions, bonuses, payment_date, payment_method,
		       transaction_id, status, created_at
		FROM payroll_records 
		WHERE talent_id = $1 AND payment_date >= $2 AND payment_date <= $3
		ORDER BY payment_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*entities.PayrollRecord
	for rows.Next() {
		record := &entities.PayrollRecord{}
		err := rows.Scan(
			&record.ID, &record.TalentID, &record.EngagementID, &record.PayPeriodStart,
			&record.PayPeriodEnd, &record.Currency, &record.HoursWorked, &record.Deductions,
			&record.Bonuses, &record.PaymentDate, &record.PaymentMethod, &record.TransactionID,
			&record.Status, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	
	return records, rows.Err()
}
func (r *PostgresMinimalCompensationRepository) CalculatePayroll(ctx context.Context, talentID uuid.UUID, period interface{}) (*entities.PayrollRecord, error) {
	// This is a complex calculation that would typically involve:
	// - Getting active compensation plan
	// - Calculating hours worked from work assignments
	// - Applying taxes and deductions
	// - For now, return a basic structure
	plan, err := r.GetActiveCompensationPlan(ctx, talentID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("no active compensation plan for talent %s", talentID)
	}
	
	// Create a basic payroll record structure
	record := &entities.PayrollRecord{
		ID:               uuid.New(),
		TalentID:         talentID,
		EngagementID:     plan.EngagementID,
		PayPeriodStart:   time.Now().AddDate(0, 0, -14), // 2 weeks ago
		PayPeriodEnd:     time.Now(),
		Currency:         plan.Currency,
		HoursWorked:      0, // Would be calculated from work assignments
		Deductions:       make(map[string]interface{}),
		Bonuses:          make(map[string]interface{}),
		PaymentDate:      time.Now(),
		PaymentMethod:    "Direct Deposit",
		Status:           "Calculated",
		CreatedAt:        time.Now(),
	}
	
	return record, nil
}
func (r *PostgresMinimalCompensationRepository) GetCompensationSummary(ctx context.Context, timeRange repositories.TimeRange) (*repositories.CompensationSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_payroll_records,
			SUM(hours_worked) as total_hours,
			COUNT(DISTINCT talent_id) as unique_talents
		FROM payroll_records 
		WHERE payment_date >= $1 AND payment_date <= $2`
	
	var totalRecords, uniqueTalents int
	var totalHours float64
	err := r.db.QueryRowContext(ctx, query, timeRange.Start, timeRange.End).Scan(
		&totalRecords, &totalHours, &uniqueTalents)
	if err != nil {
		return nil, err
	}
	
	summary := &repositories.CompensationSummary{
		TotalHoursWorked: totalHours,
		PayrollByType:    make(map[string]*entities.Money),
		TopEarners:       []*entities.Talent{},
		PayrollGrowthRate: 0, // Would need historical data calculation
	}
	
	return summary, nil
}
func (r *PostgresMinimalCompensationRepository) GetCompensationBenchmarks(ctx context.Context, roles []string) (*repositories.CompensationBenchmarks, error) {
	// This would typically query external salary data sources
	// For now, return a basic structure
	benchmarks := &repositories.CompensationBenchmarks{
		SkillSet:           roles,
		MinRate:            &entities.Money{Amount: 50000, Currency: "USD"},
		MaxRate:            &entities.Money{Amount: 100000, Currency: "USD"},
		AverageRate:        &entities.Money{Amount: 75000, Currency: "USD"},
		MedianRate:         &entities.Money{Amount: 70000, Currency: "USD"},
		MarketPercentile25: &entities.Money{Amount: 60000, Currency: "USD"},
		MarketPercentile75: &entities.Money{Amount: 85000, Currency: "USD"},
		SampleSize:         100,
		LastUpdated:        time.Now(),
	}
	
	return benchmarks, nil
}

type PostgresMinimalTrainingRepository struct{ db *sql.DB }

func NewTrainingRepository(db *sql.DB) repositories.TrainingRepository {
	return &PostgresMinimalTrainingRepository{db: db}
}
func (r *PostgresMinimalTrainingRepository) CreateTrainingProgram(ctx context.Context, program *entities.TrainingProgram) error {
	query := `
		INSERT INTO training_programs (training_id, name, description, type, target_audience, duration,
		                               format, passing_score, certification_id, is_active, 
		                               created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	
	if program.ID == uuid.Nil {
		program.ID = uuid.New()
	}
	if program.CreatedAt.IsZero() {
		program.CreatedAt = time.Now()
	}
	program.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		program.ID, program.Name, program.Description, program.Type, program.TargetAudience,
		program.Duration, program.Format, program.PassingScore, program.CertificationID,
		program.IsActive, program.CreatedAt, program.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) GetTrainingProgramByID(ctx context.Context, id uuid.UUID) (*entities.TrainingProgram, error) {
	query := `
		SELECT training_id, name, description, type, target_audience, duration, format,
		       passing_score, certification_id, is_active, created_at, updated_at
		FROM training_programs WHERE training_id = $1`
	
	program := &entities.TrainingProgram{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&program.ID, &program.Name, &program.Description, &program.Type, &program.TargetAudience,
		&program.Duration, &program.Format, &program.PassingScore, &program.CertificationID,
		&program.IsActive, &program.CreatedAt, &program.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return program, nil
}
func (r *PostgresMinimalTrainingRepository) UpdateTrainingProgram(ctx context.Context, program *entities.TrainingProgram) error {
	query := `
		UPDATE training_programs SET 
			name = $2, description = $3, type = $4, target_audience = $5, duration = $6,
			format = $7, passing_score = $8, certification_id = $9, is_active = $10,
			updated_at = $11
		WHERE training_id = $1`
	
	program.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		program.ID, program.Name, program.Description, program.Type, program.TargetAudience,
		program.Duration, program.Format, program.PassingScore, program.CertificationID,
		program.IsActive, program.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) DeleteTrainingProgram(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM training_programs WHERE training_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
func (r *PostgresMinimalTrainingRepository) ListTrainingPrograms(ctx context.Context, filter repositories.TrainingFilter) ([]*entities.TrainingProgram, int, error) {
	query := `
		SELECT training_id, name, description, type, target_audience, duration, format,
		       passing_score, certification_id, is_active, created_at, updated_at
		FROM training_programs WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.Type != nil {
		query += ` AND type = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}
	
	if filter.IsActive != nil {
		query += ` AND is_active = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}
	
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var programs []*entities.TrainingProgram
	for rows.Next() {
		program := &entities.TrainingProgram{}
		err := rows.Scan(
			&program.ID, &program.Name, &program.Description, &program.Type, &program.TargetAudience,
			&program.Duration, &program.Format, &program.PassingScore, &program.CertificationID,
			&program.IsActive, &program.CreatedAt, &program.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		programs = append(programs, program)
	}
	
	return programs, total, rows.Err()
}
func (r *PostgresMinimalTrainingRepository) CreateTrainingMaterial(ctx context.Context, material *entities.TrainingMaterial) error {
	query := `
		INSERT INTO training_materials (material_id, training_id, title, type, content_url,
		                                duration, order_num, is_required, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if material.ID == uuid.Nil {
		material.ID = uuid.New()
	}
	if material.CreatedAt.IsZero() {
		material.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		material.ID, material.TrainingID, material.Title, material.Type, material.ContentURL,
		material.Duration, material.Order, material.IsRequired, material.Metadata, material.CreatedAt)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) GetTrainingMaterial(ctx context.Context, id uuid.UUID) (*entities.TrainingMaterial, error) {
	query := `
		SELECT material_id, training_id, title, type, content_url, duration, order_num,
		       is_required, metadata, created_at
		FROM training_materials WHERE material_id = $1`
	
	material := &entities.TrainingMaterial{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&material.ID, &material.TrainingID, &material.Title, &material.Type, &material.ContentURL,
		&material.Duration, &material.Order, &material.IsRequired, &material.Metadata, &material.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return material, nil
}
func (r *PostgresMinimalTrainingRepository) UpdateTrainingMaterial(ctx context.Context, material *entities.TrainingMaterial) error {
	query := `
		UPDATE training_materials SET 
			title = $2, type = $3, content_url = $4, duration = $5, order_num = $6,
			is_required = $7, metadata = $8
		WHERE material_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		material.ID, material.Title, material.Type, material.ContentURL, material.Duration,
		material.Order, material.IsRequired, material.Metadata)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) GetMaterialsByProgram(ctx context.Context, programID uuid.UUID) ([]*entities.TrainingMaterial, error) {
	return nil, nil
}
func (r *PostgresMinimalTrainingRepository) EnrollTalentInProgram(ctx context.Context, talentID, programID uuid.UUID) (*entities.TrainingProgress, error) {
	return nil, nil
}
func (r *PostgresMinimalTrainingRepository) CreateTrainingProgress(ctx context.Context, progress *entities.TrainingProgress) error {
	query := `
		INSERT INTO training_progress (progress_id, talent_id, training_id, status, started_at, progress, attempts, material_progress, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if progress.ID == uuid.Nil {
		progress.ID = uuid.New()
	}
	
	if progress.CreatedAt.IsZero() {
		progress.CreatedAt = time.Now()
	}
	progress.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		progress.ID, progress.TalentID, progress.TrainingID, progress.Status,
		progress.StartedAt, progress.Progress, progress.Attempts, progress.MaterialProgress,
		progress.CreatedAt, progress.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) UpdateTrainingProgress(ctx context.Context, progress *entities.TrainingProgress) error {
	query := `
		UPDATE training_progress 
		SET status = $3, completed_at = $4, progress = $5, score = $6, attempts = $7,
		    material_progress = $8, certificate_url = $9, updated_at = $10
		WHERE talent_id = $1 AND training_id = $2`
	
	progress.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		progress.TalentID, progress.TrainingID, progress.Status, progress.CompletedAt,
		progress.Progress, progress.Score, progress.Attempts, progress.MaterialProgress,
		progress.CertificateURL, progress.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTrainingRepository) GetTrainingProgress(ctx context.Context, talentID, programID uuid.UUID) (*entities.TrainingProgress, error) {
	query := `
		SELECT progress_id, talent_id, training_id, status, started_at, completed_at, progress,
		       score, attempts, material_progress, certificate_url, created_at, updated_at
		FROM training_progress WHERE talent_id = $1 AND training_id = $2`
	
	progress := &entities.TrainingProgress{}
	err := r.db.QueryRowContext(ctx, query, talentID, programID).Scan(
		&progress.ID, &progress.TalentID, &progress.TrainingID, &progress.Status,
		&progress.StartedAt, &progress.CompletedAt, &progress.Progress,
		&progress.Score, &progress.Attempts, &progress.MaterialProgress,
		&progress.CertificateURL, &progress.CreatedAt, &progress.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return progress, nil
}
func (r *PostgresMinimalTrainingRepository) GetIncompleteTraining(ctx context.Context) ([]*entities.TrainingProgress, error) {
	query := `
		SELECT progress_id, talent_id, training_id, status, started_at, completed_at, progress,
		       score, attempts, material_progress, certificate_url, created_at, updated_at
		FROM training_progress WHERE status IN ('NotStarted', 'InProgress') ORDER BY started_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var progressList []*entities.TrainingProgress
	for rows.Next() {
		progress := &entities.TrainingProgress{}
		err := rows.Scan(
			&progress.ID, &progress.TalentID, &progress.TrainingID, &progress.Status,
			&progress.StartedAt, &progress.CompletedAt, &progress.Progress,
			&progress.Score, &progress.Attempts, &progress.MaterialProgress,
			&progress.CertificateURL, &progress.CreatedAt, &progress.UpdatedAt)
		if err != nil {
			return nil, err
		}
		progressList = append(progressList, progress)
	}
	
	return progressList, rows.Err()
}
func (r *PostgresMinimalTrainingRepository) GetCompletedTrainings(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error) {
	return nil, nil
}
func (r *PostgresMinimalTrainingRepository) GetRequiredTraining(ctx context.Context, talentType entities.TalentType) ([]*entities.TrainingProgram, error) {
	query := `
		SELECT training_id, name, description, type, target_audience, duration, format,
		       passing_score, certification_id, is_active, created_at, updated_at
		FROM training_programs 
		WHERE is_active = true AND (target_audience = $1 OR target_audience = 'All')
		      AND type IN ('Onboarding', 'Compliance')
		ORDER BY type, name`
	
	rows, err := r.db.QueryContext(ctx, query, talentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var programs []*entities.TrainingProgram
	for rows.Next() {
		program := &entities.TrainingProgram{}
		err := rows.Scan(
			&program.ID, &program.Name, &program.Description, &program.Type, &program.TargetAudience,
			&program.Duration, &program.Format, &program.PassingScore, &program.CertificationID,
			&program.IsActive, &program.CreatedAt, &program.UpdatedAt)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	
	return programs, rows.Err()
}
func (r *PostgresMinimalTrainingRepository) GetTrainingMaterials(ctx context.Context, trainingID uuid.UUID) ([]*entities.TrainingMaterial, error) {
	query := `
		SELECT material_id, training_id, title, type, content_url, duration, order_num,
		       is_required, metadata, created_at
		FROM training_materials WHERE training_id = $1 ORDER BY order_num ASC`
	
	rows, err := r.db.QueryContext(ctx, query, trainingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var materials []*entities.TrainingMaterial
	for rows.Next() {
		material := &entities.TrainingMaterial{}
		err := rows.Scan(
			&material.ID, &material.TrainingID, &material.Title, &material.Type, &material.ContentURL,
			&material.Duration, &material.Order, &material.IsRequired, &material.Metadata, &material.CreatedAt)
		if err != nil {
			return nil, err
		}
		materials = append(materials, material)
	}
	
	return materials, rows.Err()
}
func (r *PostgresMinimalTrainingRepository) GetTrainingProgressByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.TrainingProgress, error) {
	query := `
		SELECT progress_id, talent_id, training_id, status, started_at, completed_at, progress,
		       score, attempts, material_progress, certificate_url, created_at, updated_at
		FROM training_progress WHERE talent_id = $1 ORDER BY started_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var progressList []*entities.TrainingProgress
	for rows.Next() {
		progress := &entities.TrainingProgress{}
		err := rows.Scan(
			&progress.ID, &progress.TalentID, &progress.TrainingID, &progress.Status,
			&progress.StartedAt, &progress.CompletedAt, &progress.Progress,
			&progress.Score, &progress.Attempts, &progress.MaterialProgress,
			&progress.CertificateURL, &progress.CreatedAt, &progress.UpdatedAt)
		if err != nil {
			return nil, err
		}
		progressList = append(progressList, progress)
	}
	
	return progressList, rows.Err()
}
func (r *PostgresMinimalTrainingRepository) GetTrainingCompletionRates(ctx context.Context, timeRange repositories.TimeRange) (*repositories.TrainingCompletionRates, error) {
	query := `
		SELECT 
			COUNT(*) as total_programs,
			COUNT(CASE WHEN status = 'Completed' THEN 1 END) as completed_programs,
			COUNT(CASE WHEN status = 'InProgress' THEN 1 END) as in_progress_programs,
			AVG(CASE WHEN score IS NOT NULL THEN score END) as average_score,
			AVG(CASE WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (completed_at - started_at))/3600 END) as avg_completion_hours
		FROM training_progress 
		WHERE created_at BETWEEN $1 AND $2`
	
	var totalPrograms, completedPrograms, inProgressPrograms int
	var averageScore, avgCompletionHours sql.NullFloat64
	
	err := r.db.QueryRowContext(ctx, query, timeRange.Start, timeRange.End).Scan(
		&totalPrograms, &completedPrograms, &inProgressPrograms, &averageScore, &avgCompletionHours)
	
	if err != nil {
		return nil, err
	}
	
	completionRate := 0.0
	if totalPrograms > 0 {
		completionRate = float64(completedPrograms) / float64(totalPrograms) * 100
	}
	
	return &repositories.TrainingCompletionRates{
		TotalPrograms:         totalPrograms,
		CompletedPrograms:     completedPrograms,
		InProgressPrograms:    inProgressPrograms,
		CompletionRate:        completionRate,
		AverageScore:          averageScore.Float64,
		AverageCompletionTime: avgCompletionHours.Float64,
		CompletionByType:      make(map[string]int),
	}, nil
}
func (r *PostgresMinimalTrainingRepository) GetTrainingEffectiveness(ctx context.Context, trainingID uuid.UUID) (*repositories.TrainingEffectiveness, error) {
	query := `
		SELECT 
			COUNT(*) as enrollment_count,
			COUNT(CASE WHEN status = 'Completed' THEN 1 END) as completion_count,
			AVG(CASE WHEN score IS NOT NULL THEN score END) as average_score,
			COUNT(CASE WHEN score >= tp.passing_score THEN 1 END) as pass_count
		FROM training_progress tp_prog
		JOIN training_programs tp ON tp_prog.training_id = tp.training_id
		WHERE tp_prog.training_id = $1
		GROUP BY tp.passing_score`
	
	var enrollmentCount, completionCount, passCount int
	var averageScore sql.NullFloat64
	
	err := r.db.QueryRowContext(ctx, query, trainingID).Scan(
		&enrollmentCount, &completionCount, &averageScore, &passCount)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &repositories.TrainingEffectiveness{
				TrainingID:      trainingID,
				EnrollmentCount: 0,
				CompletionCount: 0,
				CompletionRate:  0,
				AverageScore:    0,
				PassRate:        0,
				AverageRating:   0,
				PerformanceImprovement: 0,
				ROI:             0,
			}, nil
		}
		return nil, err
	}
	
	completionRate := 0.0
	if enrollmentCount > 0 {
		completionRate = float64(completionCount) / float64(enrollmentCount) * 100
	}
	
	passRate := 0.0
	if completionCount > 0 {
		passRate = float64(passCount) / float64(completionCount) * 100
	}
	
	return &repositories.TrainingEffectiveness{
		TrainingID:      trainingID,
		EnrollmentCount: enrollmentCount,
		CompletionCount: completionCount,
		CompletionRate:  completionRate,
		AverageScore:    averageScore.Float64,
		PassRate:        passRate,
		AverageRating:   0.0, // Would require additional rating data
		PerformanceImprovement: 0.0, // Would require pre/post assessment data
		ROI:             0.0, // Would require cost and benefit calculations
	}, nil
}

type PostgresMinimalTalentApplicationRepository struct{ db *sql.DB }

func NewTalentApplicationRepository(db *sql.DB) repositories.TalentApplicationRepository {
	return &PostgresMinimalTalentApplicationRepository{db: db}
}
func (r *PostgresMinimalTalentApplicationRepository) CreateApplication(ctx context.Context, application *entities.TalentApplication) error {
	query := `
		INSERT INTO talent_applications (application_id, talent_id, job_posting_id, status, cover_letter, resume_url,
			screening_score, screening_notes, interview_notes, assessment_score, reference_checks,
			decision_date, decision_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	
	if application.ID == uuid.Nil {
		application.ID = uuid.New()
	}
	
	if application.CreatedAt.IsZero() {
		application.CreatedAt = time.Now()
	}
	application.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		application.ID, application.TalentID, application.JobPostingID, application.Status,
		application.CoverLetter, application.ResumeURL, application.ScreeningScore,
		application.ScreeningNotes, application.InterviewNotes, application.AssessmentScore,
		application.ReferenceChecks, application.DecisionDate, application.DecisionReason,
		application.CreatedAt, application.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTalentApplicationRepository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*entities.TalentApplication, error) {
	query := `
		SELECT application_id, talent_id, job_posting_id, status, cover_letter, resume_url,
			screening_score, screening_notes, interview_notes, assessment_score, reference_checks,
			decision_date, decision_reason, created_at, updated_at
		FROM talent_applications WHERE application_id = $1`
	
	application := &entities.TalentApplication{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&application.ID, &application.TalentID, &application.JobPostingID, &application.Status,
		&application.CoverLetter, &application.ResumeURL, &application.ScreeningScore,
		&application.ScreeningNotes, &application.InterviewNotes, &application.AssessmentScore,
		&application.ReferenceChecks, &application.DecisionDate, &application.DecisionReason,
		&application.CreatedAt, &application.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return application, nil
}
func (r *PostgresMinimalTalentApplicationRepository) UpdateApplication(ctx context.Context, application *entities.TalentApplication) error {
	query := `
		UPDATE talent_applications 
		SET status = $2, cover_letter = $3, resume_url = $4, screening_score = $5, screening_notes = $6,
			interview_notes = $7, assessment_score = $8, reference_checks = $9, decision_date = $10,
			decision_reason = $11, updated_at = $12
		WHERE application_id = $1`
	
	application.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		application.ID, application.Status, application.CoverLetter, application.ResumeURL,
		application.ScreeningScore, application.ScreeningNotes, application.InterviewNotes,
		application.AssessmentScore, application.ReferenceChecks, application.DecisionDate,
		application.DecisionReason, application.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTalentApplicationRepository) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (r *PostgresMinimalTalentApplicationRepository) ListApplications(ctx context.Context, filter repositories.ApplicationFilter) ([]*entities.TalentApplication, int, error) {
	query := `
		SELECT application_id, talent_id, job_posting_id, status, cover_letter, resume_url,
			screening_score, screening_notes, interview_notes, assessment_score, reference_checks,
			decision_date, decision_reason, created_at, updated_at
		FROM talent_applications WHERE 1=1`
	
	var args []interface{}
	argIndex := 1
	
	if filter.TalentID != nil {
		query += fmt.Sprintf(" AND talent_id = $%d", argIndex)
		args = append(args, *filter.TalentID)
		argIndex++
	}
	
	if filter.JobPostingID != nil {
		query += fmt.Sprintf(" AND job_posting_id = $%d", argIndex)
		args = append(args, *filter.JobPostingID)
		argIndex++
	}
	
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	
	if filter.MinScore != nil {
		query += fmt.Sprintf(" AND (screening_score >= $%d OR assessment_score >= $%d)", argIndex, argIndex)
		args = append(args, *filter.MinScore)
		argIndex++
	}
	
	if filter.CreatedAfter != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}
	
	if filter.CreatedBefore != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}
	
	// Count total before pagination
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Add sorting and pagination
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var applications []*entities.TalentApplication
	for rows.Next() {
		application := &entities.TalentApplication{}
		err := rows.Scan(
			&application.ID, &application.TalentID, &application.JobPostingID, &application.Status,
			&application.CoverLetter, &application.ResumeURL, &application.ScreeningScore,
			&application.ScreeningNotes, &application.InterviewNotes, &application.AssessmentScore,
			&application.ReferenceChecks, &application.DecisionDate, &application.DecisionReason,
			&application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		applications = append(applications, application)
	}
	
	return applications, total, rows.Err()
}
func (r *PostgresMinimalTalentApplicationRepository) GetApplicationsByPosting(ctx context.Context, postingID uuid.UUID) ([]*entities.TalentApplication, error) {
	query := `
		SELECT application_id, talent_id, job_posting_id, status, cover_letter, resume_url,
			screening_score, screening_notes, interview_notes, assessment_score, reference_checks,
			decision_date, decision_reason, created_at, updated_at
		FROM talent_applications WHERE job_posting_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, postingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var applications []*entities.TalentApplication
	for rows.Next() {
		application := &entities.TalentApplication{}
		err := rows.Scan(
			&application.ID, &application.TalentID, &application.JobPostingID, &application.Status,
			&application.CoverLetter, &application.ResumeURL, &application.ScreeningScore,
			&application.ScreeningNotes, &application.InterviewNotes, &application.AssessmentScore,
			&application.ReferenceChecks, &application.DecisionDate, &application.DecisionReason,
			&application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	
	return applications, rows.Err()
}
func (r *PostgresMinimalTalentApplicationRepository) GetApplicationsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.TalentApplication, error) {
	query := `
		SELECT application_id, talent_id, job_posting_id, status, cover_letter, resume_url,
			screening_score, screening_notes, interview_notes, assessment_score, reference_checks,
			decision_date, decision_reason, created_at, updated_at
		FROM talent_applications WHERE talent_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var applications []*entities.TalentApplication
	for rows.Next() {
		application := &entities.TalentApplication{}
		err := rows.Scan(
			&application.ID, &application.TalentID, &application.JobPostingID, &application.Status,
			&application.CoverLetter, &application.ResumeURL, &application.ScreeningScore,
			&application.ScreeningNotes, &application.InterviewNotes, &application.AssessmentScore,
			&application.ReferenceChecks, &application.DecisionDate, &application.DecisionReason,
			&application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	
	return applications, rows.Err()
}
func (r *PostgresMinimalTalentApplicationRepository) UpdateApplicationStatus(ctx context.Context, applicationID uuid.UUID, status entities.ApplicationStatus) error {
	query := `UPDATE talent_applications SET status = $2, updated_at = $3 WHERE application_id = $1`
	_, err := r.db.ExecContext(ctx, query, applicationID, status, time.Now())
	return err
}
func (r *PostgresMinimalTalentApplicationRepository) CreateJobPosting(ctx context.Context, posting *entities.JobPosting) error {
	query := `
		INSERT INTO job_postings (job_posting_id, title, description, type, department, experience_years,
			education_level, location, remote, salary_range, posted_date, closing_date, is_active,
			application_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	
	if posting.ID == uuid.Nil {
		posting.ID = uuid.New()
	}
	
	if posting.CreatedAt.IsZero() {
		posting.CreatedAt = time.Now()
	}
	posting.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		posting.ID, posting.Title, posting.Description, posting.Type, posting.Department,
		posting.ExperienceYears, posting.EducationLevel, posting.Location, posting.Remote,
		posting.SalaryRange, posting.PostedDate, posting.ClosingDate, posting.IsActive,
		posting.ApplicationCount, posting.CreatedAt, posting.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTalentApplicationRepository) GetJobPostingByID(ctx context.Context, id uuid.UUID) (*entities.JobPosting, error) {
	query := `
		SELECT job_posting_id, title, description, type, department, experience_years,
			education_level, location, remote, salary_range, posted_date, closing_date,
			is_active, application_count, created_at, updated_at
		FROM job_postings WHERE job_posting_id = $1`
	
	posting := &entities.JobPosting{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&posting.ID, &posting.Title, &posting.Description, &posting.Type, &posting.Department,
		&posting.ExperienceYears, &posting.EducationLevel, &posting.Location, &posting.Remote,
		&posting.SalaryRange, &posting.PostedDate, &posting.ClosingDate, &posting.IsActive,
		&posting.ApplicationCount, &posting.CreatedAt, &posting.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return posting, nil
}
func (r *PostgresMinimalTalentApplicationRepository) UpdateJobPosting(ctx context.Context, posting *entities.JobPosting) error {
	query := `
		UPDATE job_postings 
		SET title = $2, description = $3, type = $4, department = $5, experience_years = $6,
			education_level = $7, location = $8, remote = $9, salary_range = $10, closing_date = $11,
			is_active = $12, application_count = $13, updated_at = $14
		WHERE job_posting_id = $1`
	
	posting.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		posting.ID, posting.Title, posting.Description, posting.Type, posting.Department,
		posting.ExperienceYears, posting.EducationLevel, posting.Location, posting.Remote,
		posting.SalaryRange, posting.ClosingDate, posting.IsActive, posting.ApplicationCount,
		posting.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalTalentApplicationRepository) ListJobPostings(ctx context.Context, filter repositories.JobPostingFilter) ([]*entities.JobPosting, int, error) {
	query := `
		SELECT job_posting_id, title, description, type, department, experience_years,
			education_level, location, remote, salary_range, posted_date, closing_date,
			is_active, application_count, created_at, updated_at
		FROM job_postings WHERE 1=1`
	
	var args []interface{}
	argIndex := 1
	
	if filter.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}
	
	if filter.Department != nil {
		query += fmt.Sprintf(" AND department = $%d", argIndex)
		args = append(args, *filter.Department)
		argIndex++
	}
	
	if filter.Location != nil {
		query += fmt.Sprintf(" AND location ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Location+"%")
		argIndex++
	}
	
	if filter.Remote != nil {
		query += fmt.Sprintf(" AND remote = $%d", argIndex)
		args = append(args, *filter.Remote)
		argIndex++
	}
	
	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}
	
	if filter.MinExperience != nil {
		query += fmt.Sprintf(" AND experience_years >= $%d", argIndex)
		args = append(args, *filter.MinExperience)
		argIndex++
	}
	
	// Count total before pagination
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as filtered"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Add sorting and pagination
	sortBy := "posted_date"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var postings []*entities.JobPosting
	for rows.Next() {
		posting := &entities.JobPosting{}
		err := rows.Scan(
			&posting.ID, &posting.Title, &posting.Description, &posting.Type, &posting.Department,
			&posting.ExperienceYears, &posting.EducationLevel, &posting.Location, &posting.Remote,
			&posting.SalaryRange, &posting.PostedDate, &posting.ClosingDate, &posting.IsActive,
			&posting.ApplicationCount, &posting.CreatedAt, &posting.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		postings = append(postings, posting)
	}
	
	return postings, total, rows.Err()
}
func (r *PostgresMinimalTalentApplicationRepository) GetActiveJobPostings(ctx context.Context) ([]*entities.JobPosting, error) {
	query := `
		SELECT job_posting_id, title, description, type, department, experience_years,
			education_level, location, remote, salary_range, posted_date, closing_date,
			is_active, application_count, created_at, updated_at
		FROM job_postings 
		WHERE is_active = true AND (closing_date IS NULL OR closing_date > $1)
		ORDER BY posted_date DESC`
	
	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var postings []*entities.JobPosting
	for rows.Next() {
		posting := &entities.JobPosting{}
		err := rows.Scan(
			&posting.ID, &posting.Title, &posting.Description, &posting.Type, &posting.Department,
			&posting.ExperienceYears, &posting.EducationLevel, &posting.Location, &posting.Remote,
			&posting.SalaryRange, &posting.PostedDate, &posting.ClosingDate, &posting.IsActive,
			&posting.ApplicationCount, &posting.CreatedAt, &posting.UpdatedAt)
		if err != nil {
			return nil, err
		}
		postings = append(postings, posting)
	}
	
	return postings, rows.Err()
}
func (r *PostgresMinimalTalentApplicationRepository) GetApplicationMetrics(ctx context.Context, timeRange repositories.TimeRange) (*repositories.ApplicationMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_applications,
			COUNT(CASE WHEN status = 'Approved' THEN 1 END) as approved_applications,
			COUNT(CASE WHEN status = 'Rejected' THEN 1 END) as rejected_applications,
			COUNT(CASE WHEN status IN ('New', 'Screening', 'Interview', 'Assessment') THEN 1 END) as pending_applications,
			AVG(CASE WHEN decision_date IS NOT NULL AND created_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (decision_date - created_at))/86400 END) as avg_processing_days
		FROM talent_applications 
		WHERE created_at BETWEEN $1 AND $2`
	
	var totalApplications, approvedApplications, rejectedApplications, pendingApplications int
	var avgProcessingDays sql.NullFloat64
	
	err := r.db.QueryRowContext(ctx, query, timeRange.Start, timeRange.End).Scan(
		&totalApplications, &approvedApplications, &rejectedApplications, &pendingApplications, &avgProcessingDays)
	
	if err != nil {
		return nil, err
	}
	
	approvalRate := 0.0
	if totalApplications > 0 {
		approvalRate = float64(approvedApplications) / float64(totalApplications) * 100
	}
	
	return &repositories.ApplicationMetrics{
		TotalApplications:     totalApplications,
		ApprovedApplications:  approvedApplications,
		RejectedApplications:  rejectedApplications,
		PendingApplications:   pendingApplications,
		ApprovalRate:          approvalRate,
		AverageProcessingTime: avgProcessingDays.Float64,
		ApplicationsBySource:  make(map[string]int),
		TopSkillsRequested:    []string{},
	}, nil
}
func (r *PostgresMinimalTalentApplicationRepository) GetJobPostingPerformance(ctx context.Context, jobPostingID uuid.UUID) (*repositories.JobPostingMetrics, error) {
	query := `
		SELECT 
			jp.application_count as application_count,
			COUNT(CASE WHEN ta.status = 'Approved' THEN 1 END) as hire_count,
			COUNT(CASE WHEN ta.status IN ('Interview', 'Assessment') THEN 1 END) as interview_count,
			COUNT(CASE WHEN ta.screening_score >= 70 THEN 1 END) as qualified_count,
			AVG(CASE WHEN ta.decision_date IS NOT NULL AND ta.created_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (ta.decision_date - ta.created_at))/86400 END) as avg_time_to_fill
		FROM job_postings jp
		LEFT JOIN talent_applications ta ON jp.job_posting_id = ta.job_posting_id
		WHERE jp.job_posting_id = $1
		GROUP BY jp.job_posting_id, jp.application_count`
	
	var applicationCount, hireCount, interviewCount, qualifiedCount int
	var avgTimeToFill sql.NullFloat64
	
	err := r.db.QueryRowContext(ctx, query, jobPostingID).Scan(
		&applicationCount, &hireCount, &interviewCount, &qualifiedCount, &avgTimeToFill)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &repositories.JobPostingMetrics{
				JobPostingID:    jobPostingID,
				ViewCount:       0,
				ApplicationCount: 0,
				QualifiedCount:  0,
				InterviewCount:  0,
				HireCount:       0,
				ConversionRate:  0,
				TimeToFill:      0,
				CostPerHire:     nil,
			}, nil
		}
		return nil, err
	}
	
	conversionRate := 0.0
	if applicationCount > 0 {
		conversionRate = float64(hireCount) / float64(applicationCount) * 100
	}
	
	return &repositories.JobPostingMetrics{
		JobPostingID:     jobPostingID,
		ViewCount:        0, // Would require separate tracking
		ApplicationCount: applicationCount,
		QualifiedCount:   qualifiedCount,
		InterviewCount:   interviewCount,
		HireCount:        hireCount,
		ConversionRate:   conversionRate,
		TimeToFill:       avgTimeToFill.Float64,
		CostPerHire:      nil, // Would require cost tracking
	}, nil
}

type PostgresMinimalComplianceRepository struct{ db *sql.DB }

func NewComplianceRepository(db *sql.DB) repositories.ComplianceRepository {
	return &PostgresMinimalComplianceRepository{db: db}
}
func (r *PostgresMinimalComplianceRepository) CreateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error {
	query := `
		INSERT INTO compliance_checks (check_id, talent_id, check_type, provider, status, result,
			details, valid_until, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if check.ID == uuid.Nil {
		check.ID = uuid.New()
	}
	
	if check.CreatedAt.IsZero() {
		check.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		check.ID, check.TalentID, check.CheckType, check.Provider, check.Status,
		check.Result, check.Details, check.ValidUntil, check.CompletedAt, check.CreatedAt)
	
	return err
}
func (r *PostgresMinimalComplianceRepository) GetComplianceCheckByID(ctx context.Context, id uuid.UUID) (*entities.ComplianceCheck, error) {
	query := `
		SELECT check_id, talent_id, check_type, provider, status, result, details,
			valid_until, completed_at, created_at
		FROM compliance_checks WHERE check_id = $1`
	
	check := &entities.ComplianceCheck{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&check.ID, &check.TalentID, &check.CheckType, &check.Provider, &check.Status,
		&check.Result, &check.Details, &check.ValidUntil, &check.CompletedAt, &check.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return check, nil
}
func (r *PostgresMinimalComplianceRepository) UpdateComplianceCheck(ctx context.Context, check *entities.ComplianceCheck) error {
	query := `
		UPDATE compliance_checks 
		SET status = $2, result = $3, details = $4, valid_until = $5, completed_at = $6
		WHERE check_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		check.ID, check.Status, check.Result, check.Details, check.ValidUntil, check.CompletedAt)
	
	return err
}
func (r *PostgresMinimalComplianceRepository) ListComplianceChecks(ctx context.Context, filter interface{}) ([]*entities.ComplianceCheck, int, error) {
	return nil, 0, nil
}
func (r *PostgresMinimalComplianceRepository) GetComplianceChecksByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.ComplianceCheck, error) {
	query := `
		SELECT check_id, talent_id, check_type, provider, status, result, details,
			valid_until, completed_at, created_at
		FROM compliance_checks WHERE talent_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var checks []*entities.ComplianceCheck
	for rows.Next() {
		check := &entities.ComplianceCheck{}
		err := rows.Scan(
			&check.ID, &check.TalentID, &check.CheckType, &check.Provider, &check.Status,
			&check.Result, &check.Details, &check.ValidUntil, &check.CompletedAt, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	
	return checks, rows.Err()
}
func (r *PostgresMinimalComplianceRepository) GetPendingComplianceChecks(ctx context.Context) ([]*entities.ComplianceCheck, error) {
	query := `
		SELECT check_id, talent_id, check_type, provider, status, result, details,
			valid_until, completed_at, created_at
		FROM compliance_checks WHERE status IN ('Pending', 'InProgress') ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var checks []*entities.ComplianceCheck
	for rows.Next() {
		check := &entities.ComplianceCheck{}
		err := rows.Scan(
			&check.ID, &check.TalentID, &check.CheckType, &check.Provider, &check.Status,
			&check.Result, &check.Details, &check.ValidUntil, &check.CompletedAt, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	
	return checks, rows.Err()
}
func (r *PostgresMinimalComplianceRepository) GetExpiringCompliance(ctx context.Context, beforeDate time.Time) ([]*entities.ComplianceCheck, error) {
	query := `
		SELECT check_id, talent_id, check_type, provider, status, result, details,
			valid_until, completed_at, created_at
		FROM compliance_checks 
		WHERE valid_until IS NOT NULL AND valid_until <= $1 AND status = 'Completed'
		ORDER BY valid_until ASC`
	
	rows, err := r.db.QueryContext(ctx, query, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var checks []*entities.ComplianceCheck
	for rows.Next() {
		check := &entities.ComplianceCheck{}
		err := rows.Scan(
			&check.ID, &check.TalentID, &check.CheckType, &check.Provider, &check.Status,
			&check.Result, &check.Details, &check.ValidUntil, &check.CompletedAt, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	
	return checks, rows.Err()
}
func (r *PostgresMinimalComplianceRepository) GetContractorAgreementByID(ctx context.Context, id uuid.UUID) (*entities.ContractorAgreement, error) {
	query := `
		SELECT agreement_id, talent_id, engagement_id, contract_type, template_id, terms,
			start_date, end_date, renewal_date, signed_at, signature_id, document_url,
			status, created_at, updated_at
		FROM contractor_agreements WHERE agreement_id = $1`
	
	agreement := &entities.ContractorAgreement{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&agreement.ID, &agreement.TalentID, &agreement.EngagementID, &agreement.ContractType,
		&agreement.TemplateID, &agreement.Terms, &agreement.StartDate, &agreement.EndDate,
		&agreement.RenewalDate, &agreement.SignedAt, &agreement.SignatureID, &agreement.DocumentURL,
		&agreement.Status, &agreement.CreatedAt, &agreement.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return agreement, nil
}
func (r *PostgresMinimalComplianceRepository) GetContractorAgreementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.ContractorAgreement, error) {
	query := `
		SELECT agreement_id, talent_id, engagement_id, contract_type, template_id, terms,
			start_date, end_date, renewal_date, signed_at, signature_id, document_url,
			status, created_at, updated_at
		FROM contractor_agreements WHERE talent_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var agreements []*entities.ContractorAgreement
	for rows.Next() {
		agreement := &entities.ContractorAgreement{}
		err := rows.Scan(
			&agreement.ID, &agreement.TalentID, &agreement.EngagementID, &agreement.ContractType,
			&agreement.TemplateID, &agreement.Terms, &agreement.StartDate, &agreement.EndDate,
			&agreement.RenewalDate, &agreement.SignedAt, &agreement.SignatureID, &agreement.DocumentURL,
			&agreement.Status, &agreement.CreatedAt, &agreement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		agreements = append(agreements, agreement)
	}
	
	return agreements, rows.Err()
}
func (r *PostgresMinimalComplianceRepository) GetTalentComplianceStatus(ctx context.Context, talentID uuid.UUID) (*repositories.TalentComplianceStatus, error) {
	// Get all active compliance checks for the talent
	query := `
		SELECT check_id, talent_id, check_type, provider, status, result, details,
			valid_until, completed_at, created_at
		FROM compliance_checks WHERE talent_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var activeChecks []*entities.ComplianceCheck
	var expiringChecks []*entities.ComplianceCheck
	var missingChecks []string
	var lastCheckDate *time.Time
	complianceScore := 100.0
	
	for rows.Next() {
		check := &entities.ComplianceCheck{}
		err := rows.Scan(
			&check.ID, &check.TalentID, &check.CheckType, &check.Provider, &check.Status,
			&check.Result, &check.Details, &check.ValidUntil, &check.CompletedAt, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		
		if check.Status == "Completed" && check.Result == "Pass" {
			activeChecks = append(activeChecks, check)
			if check.ValidUntil != nil && check.ValidUntil.Before(time.Now().AddDate(0, 0, 30)) {
				expiringChecks = append(expiringChecks, check)
			}
		} else if check.Status == "Failed" || check.Result == "Fail" {
			complianceScore -= 20.0
			missingChecks = append(missingChecks, check.CheckType)
		}
		
		if lastCheckDate == nil || check.CreatedAt.After(*lastCheckDate) {
			lastCheckDate = &check.CreatedAt
		}
	}
	
	isCompliant := complianceScore >= 80.0 && len(missingChecks) == 0
	
	// Calculate next check due (example: 90 days from last check)
	var nextCheckDue *time.Time
	if lastCheckDate != nil {
		nextCheck := lastCheckDate.AddDate(0, 0, 90)
		nextCheckDue = &nextCheck
	}
	
	riskLevel := "Low"
	if !isCompliant {
		riskLevel = "High"
	} else if len(expiringChecks) > 0 {
		riskLevel = "Medium"
	}
	
	return &repositories.TalentComplianceStatus{
		TalentID:        talentID,
		IsCompliant:     isCompliant,
		ComplianceScore: complianceScore,
		ActiveChecks:    activeChecks,
		ExpiringChecks:  expiringChecks,
		MissingChecks:   missingChecks,
		LastCheckDate:   lastCheckDate,
		NextCheckDue:    nextCheckDue,
		RiskLevel:       riskLevel,
	}, nil
}
func (r *PostgresMinimalComplianceRepository) CreateContractorAgreement(ctx context.Context, agreement *entities.ContractorAgreement) error {
	query := `
		INSERT INTO contractor_agreements (agreement_id, talent_id, engagement_id, contract_type, template_id,
			terms, start_date, end_date, renewal_date, signed_at, signature_id, document_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	
	if agreement.ID == uuid.Nil {
		agreement.ID = uuid.New()
	}
	
	if agreement.CreatedAt.IsZero() {
		agreement.CreatedAt = time.Now()
	}
	agreement.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		agreement.ID, agreement.TalentID, agreement.EngagementID, agreement.ContractType,
		agreement.TemplateID, agreement.Terms, agreement.StartDate, agreement.EndDate,
		agreement.RenewalDate, agreement.SignedAt, agreement.SignatureID, agreement.DocumentURL,
		agreement.Status, agreement.CreatedAt, agreement.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalComplianceRepository) GetContractorAgreement(ctx context.Context, id uuid.UUID) (*entities.ContractorAgreement, error) {
	return nil, nil
}
func (r *PostgresMinimalComplianceRepository) UpdateContractorAgreement(ctx context.Context, agreement *entities.ContractorAgreement) error {
	query := `
		UPDATE contractor_agreements 
		SET terms = $2, end_date = $3, renewal_date = $4, signed_at = $5, signature_id = $6,
			document_url = $7, status = $8, updated_at = $9
		WHERE agreement_id = $1`
	
	agreement.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		agreement.ID, agreement.Terms, agreement.EndDate, agreement.RenewalDate,
		agreement.SignedAt, agreement.SignatureID, agreement.DocumentURL,
		agreement.Status, agreement.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalComplianceRepository) GetAgreementsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.ContractorAgreement, error) {
	return nil, nil
}
func (r *PostgresMinimalComplianceRepository) GetComplianceReport(ctx context.Context, timeRange repositories.TimeRange) (*repositories.ComplianceReport, error) {
	query := `
		SELECT 
			COUNT(DISTINCT t.talent_id) as total_talent,
			COUNT(DISTINCT CASE WHEN cc.status = 'Completed' AND cc.result = 'Pass' THEN t.talent_id END) as compliant_talent,
			COUNT(CASE WHEN cc.valid_until <= $3 AND cc.status = 'Completed' THEN 1 END) as expiring_checks,
			COUNT(CASE WHEN cc.status IN ('Pending', 'Failed') THEN 1 END) as overdue_checks
		FROM talents t
		LEFT JOIN compliance_checks cc ON t.talent_id = cc.talent_id
			AND cc.created_at BETWEEN $1 AND $2`
	
	var totalTalent, compliantTalent, expiringChecks, overdueChecks int
	
	err := r.db.QueryRowContext(ctx, query, timeRange.Start, timeRange.End, time.Now().AddDate(0, 0, 30)).Scan(
		&totalTalent, &compliantTalent, &expiringChecks, &overdueChecks)
	
	if err != nil {
		return nil, err
	}
	
	nonCompliantTalent := totalTalent - compliantTalent
	complianceRate := 0.0
	if totalTalent > 0 {
		complianceRate = float64(compliantTalent) / float64(totalTalent) * 100
	}
	
	riskLevel := "Low"
	if complianceRate < 80 {
		riskLevel = "High"
	} else if complianceRate < 90 {
		riskLevel = "Medium"
	}
	
	return &repositories.ComplianceReport{
		TotalTalent:        totalTalent,
		CompliantTalent:    compliantTalent,
		NonCompliantTalent: nonCompliantTalent,
		ComplianceRate:     complianceRate,
		ExpiringChecks:     expiringChecks,
		OverdueChecks:      overdueChecks,
		ComplianceByType:   make(map[string]int),
		RiskLevel:          riskLevel,
	}, nil
}
func (r *PostgresMinimalComplianceRepository) GetExpiringAgreements(ctx context.Context, beforeDate time.Time) ([]*entities.ContractorAgreement, error) {
	query := `
		SELECT agreement_id, talent_id, engagement_id, contract_type, template_id, terms,
			start_date, end_date, renewal_date, signed_at, signature_id, document_url,
			status, created_at, updated_at
		FROM contractor_agreements 
		WHERE (end_date IS NOT NULL AND end_date <= $1) OR (renewal_date IS NOT NULL AND renewal_date <= $1)
			AND status = 'Active'
		ORDER BY COALESCE(end_date, renewal_date) ASC`
	
	rows, err := r.db.QueryContext(ctx, query, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var agreements []*entities.ContractorAgreement
	for rows.Next() {
		agreement := &entities.ContractorAgreement{}
		err := rows.Scan(
			&agreement.ID, &agreement.TalentID, &agreement.EngagementID, &agreement.ContractType,
			&agreement.TemplateID, &agreement.Terms, &agreement.StartDate, &agreement.EndDate,
			&agreement.RenewalDate, &agreement.SignedAt, &agreement.SignatureID, &agreement.DocumentURL,
			&agreement.Status, &agreement.CreatedAt, &agreement.UpdatedAt)
		if err != nil {
			return nil, err
		}
		agreements = append(agreements, agreement)
	}
	
	return agreements, rows.Err()
}

type PostgresMinimalOffboardingRepository struct{ db *sql.DB }

func NewOffboardingRepository(db *sql.DB) repositories.OffboardingRepository {
	return &PostgresMinimalOffboardingRepository{db: db}
}
func (r *PostgresMinimalOffboardingRepository) CreateOffboardingChecklist(ctx context.Context, checklist *entities.OffboardingChecklist) error {
	query := `
		INSERT INTO offboarding_checklists (checklist_id, talent_id, engagement_id, reason, last_working_date,
			knowledge_transfer, exit_interview_url, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if checklist.ID == uuid.Nil {
		checklist.ID = uuid.New()
	}
	
	if checklist.CreatedAt.IsZero() {
		checklist.CreatedAt = time.Now()
	}
	checklist.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		checklist.ID, checklist.TalentID, checklist.EngagementID, checklist.Reason,
		checklist.LastWorkingDate, checklist.KnowledgeTransfer, checklist.ExitInterviewURL,
		checklist.CompletedAt, checklist.CreatedAt, checklist.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalOffboardingRepository) GetOffboardingChecklistByID(ctx context.Context, id uuid.UUID) (*entities.OffboardingChecklist, error) {
	query := `
		SELECT checklist_id, talent_id, engagement_id, reason, last_working_date,
			knowledge_transfer, exit_interview_url, completed_at, created_at, updated_at
		FROM offboarding_checklists WHERE checklist_id = $1`
	
	checklist := &entities.OffboardingChecklist{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&checklist.ID, &checklist.TalentID, &checklist.EngagementID, &checklist.Reason,
		&checklist.LastWorkingDate, &checklist.KnowledgeTransfer, &checklist.ExitInterviewURL,
		&checklist.CompletedAt, &checklist.CreatedAt, &checklist.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return checklist, nil
}
func (r *PostgresMinimalOffboardingRepository) UpdateOffboardingChecklist(ctx context.Context, checklist *entities.OffboardingChecklist) error {
	query := `
		UPDATE offboarding_checklists 
		SET reason = $2, last_working_date = $3, knowledge_transfer = $4, exit_interview_url = $5,
			completed_at = $6, updated_at = $7
		WHERE checklist_id = $1`
	
	checklist.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		checklist.ID, checklist.Reason, checklist.LastWorkingDate, checklist.KnowledgeTransfer,
		checklist.ExitInterviewURL, checklist.CompletedAt, checklist.UpdatedAt)
	
	return err
}
func (r *PostgresMinimalOffboardingRepository) ListOffboardingChecklists(ctx context.Context, filter interface{}) ([]*entities.OffboardingChecklist, int, error) {
	return nil, 0, nil
}
func (r *PostgresMinimalOffboardingRepository) GetChecklistsByTalent(ctx context.Context, talentID uuid.UUID) ([]*entities.OffboardingChecklist, error) {
	return nil, nil
}
func (r *PostgresMinimalOffboardingRepository) CreateOffboardingTask(ctx context.Context, task *entities.OffboardingTask) error {
	query := `
		INSERT INTO offboarding_tasks (task_id, checklist_id, name, description, assigned_to,
			due_date, completed_at, completed_by, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.ChecklistID, task.Name, task.Description, task.AssignedTo,
		task.DueDate, task.CompletedAt, task.CompletedBy, task.Notes, task.CreatedAt)
	
	return err
}
func (r *PostgresMinimalOffboardingRepository) UpdateOffboardingTask(ctx context.Context, task *entities.OffboardingTask) error {
	query := `
		UPDATE offboarding_tasks 
		SET name = $2, description = $3, assigned_to = $4, due_date = $5,
			completed_at = $6, completed_by = $7, notes = $8
		WHERE task_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Name, task.Description, task.AssignedTo, task.DueDate,
		task.CompletedAt, task.CompletedBy, task.Notes)
	
	return err
}
func (r *PostgresMinimalOffboardingRepository) GetTasksByChecklist(ctx context.Context, checklistID uuid.UUID) ([]*entities.OffboardingTask, error) {
	return nil, nil
}
func (r *PostgresMinimalOffboardingRepository) GetPendingTasks(ctx context.Context) ([]*entities.OffboardingTask, error) {
	return nil, nil
}
func (r *PostgresMinimalOffboardingRepository) GetOffboardingChecklistByTalent(ctx context.Context, talentID uuid.UUID) (*entities.OffboardingChecklist, error) {
	query := `
		SELECT checklist_id, talent_id, engagement_id, reason, last_working_date,
			knowledge_transfer, exit_interview_url, completed_at, created_at, updated_at
		FROM offboarding_checklists WHERE talent_id = $1 ORDER BY created_at DESC LIMIT 1`
	
	checklist := &entities.OffboardingChecklist{}
	err := r.db.QueryRowContext(ctx, query, talentID).Scan(
		&checklist.ID, &checklist.TalentID, &checklist.EngagementID, &checklist.Reason,
		&checklist.LastWorkingDate, &checklist.KnowledgeTransfer, &checklist.ExitInterviewURL,
		&checklist.CompletedAt, &checklist.CreatedAt, &checklist.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return checklist, nil
}
func (r *PostgresMinimalOffboardingRepository) GetPendingOffboarding(ctx context.Context) ([]*entities.OffboardingChecklist, error) {
	query := `
		SELECT checklist_id, talent_id, engagement_id, reason, last_working_date,
			knowledge_transfer, exit_interview_url, completed_at, created_at, updated_at
		FROM offboarding_checklists WHERE completed_at IS NULL ORDER BY last_working_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var checklists []*entities.OffboardingChecklist
	for rows.Next() {
		checklist := &entities.OffboardingChecklist{}
		err := rows.Scan(
			&checklist.ID, &checklist.TalentID, &checklist.EngagementID, &checklist.Reason,
			&checklist.LastWorkingDate, &checklist.KnowledgeTransfer, &checklist.ExitInterviewURL,
			&checklist.CompletedAt, &checklist.CreatedAt, &checklist.UpdatedAt)
		if err != nil {
			return nil, err
		}
		checklists = append(checklists, checklist)
	}
	
	return checklists, rows.Err()
}
func (r *PostgresMinimalOffboardingRepository) GetOffboardingTaskByID(ctx context.Context, id uuid.UUID) (*entities.OffboardingTask, error) {
	query := `
		SELECT task_id, checklist_id, name, description, assigned_to, due_date,
			completed_at, completed_by, notes, created_at
		FROM offboarding_tasks WHERE task_id = $1`
	
	task := &entities.OffboardingTask{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.ChecklistID, &task.Name, &task.Description, &task.AssignedTo,
		&task.DueDate, &task.CompletedAt, &task.CompletedBy, &task.Notes, &task.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return task, nil
}
func (r *PostgresMinimalOffboardingRepository) GetOffboardingTasksByChecklist(ctx context.Context, checklistID uuid.UUID) ([]*entities.OffboardingTask, error) {
	query := `
		SELECT task_id, checklist_id, name, description, assigned_to, due_date,
			completed_at, completed_by, notes, created_at
		FROM offboarding_tasks WHERE checklist_id = $1 ORDER BY due_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, checklistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tasks []*entities.OffboardingTask
	for rows.Next() {
		task := &entities.OffboardingTask{}
		err := rows.Scan(
			&task.ID, &task.ChecklistID, &task.Name, &task.Description, &task.AssignedTo,
			&task.DueDate, &task.CompletedAt, &task.CompletedBy, &task.Notes, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}
func (r *PostgresMinimalOffboardingRepository) GetOverdueOffboardingTasks(ctx context.Context) ([]*entities.OffboardingTask, error) {
	query := `
		SELECT task_id, checklist_id, name, description, assigned_to, due_date,
			completed_at, completed_by, notes, created_at
		FROM offboarding_tasks 
		WHERE due_date < $1 AND completed_at IS NULL
		ORDER BY due_date ASC`
	
	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tasks []*entities.OffboardingTask
	for rows.Next() {
		task := &entities.OffboardingTask{}
		err := rows.Scan(
			&task.ID, &task.ChecklistID, &task.Name, &task.Description, &task.AssignedTo,
			&task.DueDate, &task.CompletedAt, &task.CompletedBy, &task.Notes, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

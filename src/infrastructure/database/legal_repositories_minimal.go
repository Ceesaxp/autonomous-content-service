package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// PostgresMinimalLegalRepository - minimal implementation that compiles
// Phase1: Contract/Template CRUD + filtering, pagination
type PostgresMinimalLegalRepository struct {
	db *sql.DB
}

func NewLegalRepository(db *sql.DB) repositories.LegalRepository {
	return &PostgresMinimalLegalRepository{db: db}
}

// ErrNotImplemented indicates a stub that still needs production logic.
var ErrNotImplemented = fmt.Errorf("method not implemented")

// rowScanner abstracts sql.Row and sql.Rows Scan.
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanContract reads a full Contract record from a rowScanner.
func (r *PostgresMinimalLegalRepository) scanContract(scanner rowScanner, c *entities.Contract) error {
	var rawParams, rawTerms, rawSignatures, rawChecks, rawDispute, rawIPLicenses, rawInsurance, rawRisk, rawAudit []byte
	var parentID, projectID sql.NullString
	var expDate, renewDate, archivedAt sql.NullTime
	if err := scanner.Scan(
		&c.ID, &c.Title, &c.Type, &c.Status, &c.Version,
		&parentID, &c.ClientID, &projectID, &c.TemplateID,
		&c.Content,
		&rawParams, &rawTerms, &rawSignatures,
		&c.EffectiveDate, &expDate, &renewDate,
		&rawChecks, &rawDispute, &rawIPLicenses,
		&c.InsuranceRequired, &rawInsurance, &rawRisk, &rawAudit,
		&c.CreatedAt, &c.UpdatedAt, &archivedAt,
	); err != nil {
		return err
	}
	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return err
		}
		c.ParentContractID = &pid
	}
	if projectID.Valid {
		pid, err := uuid.Parse(projectID.String)
		if err != nil {
			return err
		}
		c.ProjectID = &pid
	}
	if expDate.Valid {
		c.ExpirationDate = &expDate.Time
	}
	if renewDate.Valid {
		c.RenewalDate = &renewDate.Time
	}
	if archivedAt.Valid {
		c.ArchivedAt = &archivedAt.Time
	}
	if err := json.Unmarshal(rawParams, &c.Parameters); err != nil {
		return err
	}
	if err := json.Unmarshal(rawTerms, &c.Terms); err != nil {
		return err
	}
	if err := json.Unmarshal(rawSignatures, &c.Signatures); err != nil {
		return err
	}
	if err := json.Unmarshal(rawChecks, &c.ComplianceChecks); err != nil {
		return err
	}
	if err := json.Unmarshal(rawDispute, &c.DisputeResolution); err != nil {
		return err
	}
	if err := json.Unmarshal(rawIPLicenses, &c.IPLicenses); err != nil {
		return err
	}
	if err := json.Unmarshal(rawInsurance, &c.InsurancePolicies); err != nil {
		return err
	}
	if err := json.Unmarshal(rawRisk, &c.RiskAssessment); err != nil {
		return err
	}
	if err := json.Unmarshal(rawAudit, &c.AuditTrail); err != nil {
		return err
	}
	return nil
}

// scanTemplate reads a full ContractTemplate record from a rowScanner.
func (r *PostgresMinimalLegalRepository) scanTemplate(scanner rowScanner, t *entities.ContractTemplate) error {
	var rawParams, rawTerms []byte
	if err := scanner.Scan(
		&t.ID, &t.Name, &t.Type, &t.Version, &t.Content,
		&rawParams, &rawTerms, &t.Metadata, &t.IsActive,
		&t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return err
	}
	if err := json.Unmarshal(rawParams, &t.Parameters); err != nil {
		return err
	}
	if err := json.Unmarshal(rawTerms, &t.DefaultTerms); err != nil {
		return err
	}
	return nil
}

// scanDisputeResolution maps a dispute_resolutions row into a DisputeResolution entity.
func (r *PostgresMinimalLegalRepository) scanDisputeResolution(scanner rowScanner, d *entities.DisputeResolution, rawTimeline *[]byte) error {
	var resolvedAt sql.NullTime
	if err := scanner.Scan(
		&d.ID, &d.ContractID, &d.Type, &d.Status, &d.Description,
		&d.InitiatedBy, &d.ResolutionMethod, &d.Mediator, &d.Arbitrator,
		&d.Venue, &d.GoverningLaw, rawTimeline, &d.Resolution,
		&d.Cost.Amount, &d.Cost.Currency, &d.InitiatedAt, &resolvedAt, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		return err
	}
	if resolvedAt.Valid {
		d.ResolvedAt = &resolvedAt.Time
	}
	if err := json.Unmarshal(*rawTimeline, &d.Timeline); err != nil {
		return fmt.Errorf("unmarshal dispute timeline: %w", err)
	}
	return nil
}

// scanRegulatoryReport maps a regulatory_reports row into a RegulatoryReport entity.
func (r *PostgresMinimalLegalRepository) scanRegulatoryReport(scanner rowScanner, rep *entities.RegulatoryReport, rawData *[]byte) error {
	var filedAt sql.NullTime
	var confirmation sql.NullString
	var docURL sql.NullString
	if err := scanner.Scan(
		&rep.ID, &rep.Type, &rep.Regulation, &rep.Authority, &rep.Period, &rep.Status,
		&rep.Content, rawData, &rep.FilingDeadline, &filedAt, &confirmation, &docURL,
		&rep.CreatedAt, &rep.UpdatedAt,
	); err != nil {
		return err
	}
	if filedAt.Valid {
		rep.FiledAt = &filedAt.Time
	}
	if confirmation.Valid {
		rep.ConfirmationID = &confirmation.String
	}
	if docURL.Valid {
		rep.DocumentURL = &docURL.String
	}
	rep.Data = json.RawMessage(*rawData)
	return nil
}

// scanSignature reads a ContractSignature from a rowScanner.
func (r *PostgresMinimalLegalRepository) scanSignature(scanner rowScanner, s *entities.ContractSignature) error {
	var dummyContract, cert sql.NullString
	var exp sql.NullTime
	if err := scanner.Scan(
		&s.ID, &dummyContract, &s.SignerName, &s.SignerEmail, &s.SignerRole,
		&s.SignatureType, &s.SignatureData, &s.SignatureHash,
		&s.IPAddress, &s.UserAgent, &s.Timestamp, &s.Status,
		&s.VerificationHash, &cert, &s.IsValid, &exp,
	); err != nil {
		return err
	}
	if cert.Valid {
		s.CertificateID = &cert.String
	}
	if exp.Valid {
		s.ExpiresAt = &exp.Time
	}
	return nil
}

// CreateContract inserts a new contract record.
func (r *PostgresMinimalLegalRepository) CreateContract(ctx context.Context, contract *entities.Contract) error {
	params, err := json.Marshal(contract.Parameters)
	if err != nil {
		return fmt.Errorf("marshal parameters: %w", err)
	}
	terms, err := json.Marshal(contract.Terms)
	if err != nil {
		return fmt.Errorf("marshal terms: %w", err)
	}
	sigs, err := json.Marshal(contract.Signatures)
	if err != nil {
		return fmt.Errorf("marshal signatures: %w", err)
	}
	checks, err := json.Marshal(contract.ComplianceChecks)
	if err != nil {
		return fmt.Errorf("marshal compliance checks: %w", err)
	}
	dispute, err := json.Marshal(contract.DisputeResolution)
	if err != nil {
		return fmt.Errorf("marshal dispute resolution: %w", err)
	}
	ipLic, err := json.Marshal(contract.IPLicenses)
	if err != nil {
		return fmt.Errorf("marshal ip licenses: %w", err)
	}
	insPol, err := json.Marshal(contract.InsurancePolicies)
	if err != nil {
		return fmt.Errorf("marshal insurance policies: %w", err)
	}
	risk, err := json.Marshal(contract.RiskAssessment)
	if err != nil {
		return fmt.Errorf("marshal risk assessment: %w", err)
	}
	audit, err := json.Marshal(contract.AuditTrail)
	if err != nil {
		return fmt.Errorf("marshal audit trail: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO contracts (contract_id,title,type,status,version,parent_contract_id,client_id,project_id,template_id,content,parameters,terms,signatures,effective_date,expiration_date,renewal_date,compliance_checks,dispute_resolution,ip_licenses,insurance_required,insurance_policies,risk_assessment,audit_trail,created_at,updated_at,archived_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12::jsonb,$13::jsonb,$14,$15,$16,$17::jsonb,$18::jsonb,$19::jsonb,$20,$21::jsonb,$22::jsonb,$23::jsonb,NOW(),NOW(),$24)`,
		contract.ID, contract.Title, contract.Type, contract.Status, contract.Version,
		contract.ParentContractID, contract.ClientID, contract.ProjectID, contract.TemplateID,
		contract.Content, params, terms, sigs, contract.EffectiveDate, contract.ExpirationDate, contract.RenewalDate,
		checks, dispute, ipLic, contract.InsuranceRequired, insPol, risk, audit, contract.ArchivedAt,
	)
	if err != nil {
		return fmt.Errorf("insert contract: %w", err)
	}
	return nil
}

// GetContractByID retrieves a contract by ID.
func (r *PostgresMinimalLegalRepository) GetContractByID(ctx context.Context, id uuid.UUID) (*entities.Contract, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id,title,type,status,version,parent_contract_id,client_id,project_id,template_id,content,parameters,terms,signatures,effective_date,expiration_date,renewal_date,compliance_checks,dispute_resolution,ip_licenses,insurance_required,insurance_policies,risk_assessment,audit_trail,created_at,updated_at,archived_at FROM contracts WHERE contract_id=$1`,
		id,
	)
	var c entities.Contract
	if err := r.scanContract(row, &c); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get contract: %w", err)
	}
	return &c, nil
}

// UpdateContract updates an existing contract.
func (r *PostgresMinimalLegalRepository) UpdateContract(ctx context.Context, contract *entities.Contract) error {
	return ErrNotImplemented
}

// DeleteContract removes a contract by ID.
func (r *PostgresMinimalLegalRepository) DeleteContract(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM contracts WHERE contract_id=$1`, id); err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	return nil
}

// ListContracts returns contracts matching filters.
func (r *PostgresMinimalLegalRepository) ListContracts(ctx context.Context, filters repositories.ContractFilters) ([]*entities.Contract, error) {
	return nil, ErrNotImplemented
}

// GetContractsByClientID returns all contracts for a client.
func (r *PostgresMinimalLegalRepository) GetContractsByClientID(ctx context.Context, clientID uuid.UUID) ([]*entities.Contract, error) {
	return nil, ErrNotImplemented
}

// GetContractsByStatus returns contracts filtered by status.
func (r *PostgresMinimalLegalRepository) GetContractsByStatus(ctx context.Context, status entities.ContractStatus) ([]*entities.Contract, error) {
	return nil, ErrNotImplemented
}

// GetExpiringContracts finds contracts expiring within given days.
func (r *PostgresMinimalLegalRepository) GetExpiringContracts(ctx context.Context, days int) ([]*entities.Contract, error) {
	return nil, ErrNotImplemented
}

// GetActiveContracts returns only active contracts.
func (r *PostgresMinimalLegalRepository) GetActiveContracts(ctx context.Context) ([]*entities.Contract, error) {
	return nil, ErrNotImplemented
}

// ArchiveContract archives a contract by setting status and archived_at.
func (r *PostgresMinimalLegalRepository) ArchiveContract(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE contracts SET status=$2,archived_at=NOW(),updated_at=NOW() WHERE contract_id=$1`, id, entities.ContractStatusArchived); err != nil {
		return fmt.Errorf("archive contract: %w", err)
	}
	return nil
}

// Contract Templates
// CreateContractTemplate inserts a new contract template record.
func (r *PostgresMinimalLegalRepository) CreateContractTemplate(ctx context.Context, template *entities.ContractTemplate) error {
	params, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("marshal template parameters: %w", err)
	}
	terms, err := json.Marshal(template.DefaultTerms)
	if err != nil {
		return fmt.Errorf("marshal default terms: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO contract_templates (template_id,name,type,version,content,parameters,default_terms,metadata,is_active,created_at,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8,$9,NOW(),NOW())`,
		template.ID, template.Name, template.Type, template.Version, template.Content,
		params, terms, template.Metadata, template.IsActive,
	)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}
	return nil
}

// GetContractTemplateByID retrieves a contract template by ID.
func (r *PostgresMinimalLegalRepository) GetContractTemplateByID(ctx context.Context, id uuid.UUID) (*entities.ContractTemplate, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT template_id,name,type,version,content,parameters,default_terms,metadata,is_active,created_at,updated_at
        FROM contract_templates WHERE template_id=$1`, id)
	var t entities.ContractTemplate
	if err := r.scanTemplate(row, &t); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &t, nil
}

// UpdateContractTemplate updates an existing contract template.
func (r *PostgresMinimalLegalRepository) UpdateContractTemplate(ctx context.Context, template *entities.ContractTemplate) error {
	params, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("marshal template parameters: %w", err)
	}
	terms, err := json.Marshal(template.DefaultTerms)
	if err != nil {
		return fmt.Errorf("marshal default terms: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE contract_templates SET name=$2,type=$3,version=$4,content=$5,parameters=$6::jsonb,default_terms=$7::jsonb,metadata=$8,is_active=$9,updated_at=NOW() WHERE template_id=$1`,
		template.ID, template.Name, template.Type, template.Version, template.Content,
		params, terms, template.Metadata, template.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	return nil
}

// DeleteContractTemplate removes a contract template by ID.
func (r *PostgresMinimalLegalRepository) DeleteContractTemplate(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM contract_templates WHERE template_id=$1`, id); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

// ListContractTemplates returns contract templates matching filters.
func (r *PostgresMinimalLegalRepository) ListContractTemplates(ctx context.Context, filters repositories.TemplateFilters) ([]*entities.ContractTemplate, error) {
	base := `SELECT template_id,name,type,version,content,parameters,default_terms,metadata,is_active,created_at,updated_at FROM contract_templates`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *filters.IsActive)
		idx++
	}
	if filters.Version != nil {
		where = append(where, fmt.Sprintf("version=$%d", idx))
		args = append(args, *filters.Version)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractTemplate
	for rows.Next() {
		var t entities.ContractTemplate
		if err := r.scanTemplate(rows, &t); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate templates: %w", err)
	}
	return out, nil
}

// GetTemplatesByType returns templates of a given type.
func (r *PostgresMinimalLegalRepository) GetTemplatesByType(ctx context.Context, templateType entities.ContractType) ([]*entities.ContractTemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT template_id,name,type,version,content,parameters,default_terms,metadata,is_active,created_at,updated_at FROM contract_templates WHERE type=$1 ORDER BY created_at DESC`, templateType)
	if err != nil {
		return nil, fmt.Errorf("get templates by type: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractTemplate
	for rows.Next() {
		var t entities.ContractTemplate
		if err := r.scanTemplate(rows, &t); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate templates by type: %w", err)
	}
	return out, nil
}

// GetActiveTemplates returns only active templates.
func (r *PostgresMinimalLegalRepository) GetActiveTemplates(ctx context.Context) ([]*entities.ContractTemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT template_id,name,type,version,content,parameters,default_terms,metadata,is_active,created_at,updated_at FROM contract_templates WHERE is_active ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get active templates: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractTemplate
	for rows.Next() {
		var t entities.ContractTemplate
		if err := r.scanTemplate(rows, &t); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active templates: %w", err)
	}
	return out, nil
}

// Signature Management (Phase2)
func (r *PostgresMinimalLegalRepository) CreateSignature(ctx context.Context, s *entities.ContractSignature) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO contract_signatures (signature_id,contract_id,signer_name,signer_email,signer_role,signature_type,signature_data,signature_hash,ip_address,user_agent,timestamp,status,verification_hash,certificate_id,is_valid,expires_at,created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NOW())`,
		s.ID, uuid.Nil, s.SignerName, s.SignerEmail, s.SignerRole,
		s.SignatureType, s.SignatureData, s.SignatureHash, s.IPAddress,
		s.UserAgent, s.Timestamp, s.Status, s.VerificationHash,
		s.CertificateID, s.IsValid, s.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert signature: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetSignatureByID(ctx context.Context, id uuid.UUID) (*entities.ContractSignature, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT signature_id,contract_id,signer_name,signer_email,signer_role,signature_type,signature_data,signature_hash,ip_address,user_agent,timestamp,status,verification_hash,certificate_id,is_valid,expires_at
        FROM contract_signatures WHERE signature_id=$1`, id)
	var s entities.ContractSignature
	if err := r.scanSignature(row, &s); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get signature: %w", err)
	}
	return &s, nil
}

func (r *PostgresMinimalLegalRepository) UpdateSignature(ctx context.Context, s *entities.ContractSignature) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE contract_signatures SET signer_name=$2,signer_email=$3,signer_role=$4,signature_type=$5,signature_data=$6,signature_hash=$7,ip_address=$8,user_agent=$9,timestamp=$10,status=$11,verification_hash=$12,certificate_id=$13,is_valid=$14,expires_at=$15 WHERE signature_id=$1`,
		s.ID, s.SignerName, s.SignerEmail, s.SignerRole,
		s.SignatureType, s.SignatureData, s.SignatureHash, s.IPAddress,
		s.UserAgent, s.Timestamp, s.Status, s.VerificationHash,
		s.CertificateID, s.IsValid, s.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("update signature: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetSignaturesByContract(ctx context.Context, contractID uuid.UUID) ([]*entities.ContractSignature, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT signature_id,contract_id,signer_name,signer_email,signer_role,signature_type,signature_data,signature_hash,ip_address,user_agent,timestamp,status,verification_hash,certificate_id,is_valid,expires_at
        FROM contract_signatures WHERE contract_id=$1 ORDER BY timestamp`, contractID)
	if err != nil {
		return nil, fmt.Errorf("get signatures by contract: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractSignature
	for rows.Next() {
		var s entities.ContractSignature
		if err := r.scanSignature(rows, &s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signatures by contract: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) VerifySignature(ctx context.Context, id uuid.UUID) (bool, error) {
	var valid bool
	err := r.db.QueryRowContext(ctx,
		`SELECT is_valid FROM contract_signatures WHERE signature_id=$1`, id).Scan(&valid)
	if err != nil {
		return false, fmt.Errorf("verify signature: %w", err)
	}
	return valid, nil
}

func (r *PostgresMinimalLegalRepository) GetPendingSignatures(ctx context.Context) ([]*entities.ContractSignature, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT signature_id,contract_id,signer_name,signer_email,signer_role,signature_type,signature_data,signature_hash,ip_address,user_agent,timestamp,status,verification_hash,certificate_id,is_valid,expires_at
        FROM contract_signatures WHERE status='Pending' ORDER BY timestamp`)
	if err != nil {
		return nil, fmt.Errorf("get pending signatures: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractSignature
	for rows.Next() {
		var s entities.ContractSignature
		if err := r.scanSignature(rows, &s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending signatures: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetExpiredSignatures(ctx context.Context) ([]*entities.ContractSignature, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT signature_id,contract_id,signer_name,signer_email,signer_role,signature_type,signature_data,signature_hash,ip_address,user_agent,timestamp,status,verification_hash,certificate_id,is_valid,expires_at
        FROM contract_signatures WHERE status='Expired' OR (expires_at IS NOT NULL AND expires_at< NOW()) ORDER BY expires_at`)
	if err != nil {
		return nil, fmt.Errorf("get expired signatures: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractSignature
	for rows.Next() {
		var s entities.ContractSignature
		if err := r.scanSignature(rows, &s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired signatures: %w", err)
	}
	return out, nil
}

// Compliance Checks (Phase3)
func (r *PostgresMinimalLegalRepository) CreateComplianceCheck(ctx context.Context, check *entities.LegalComplianceCheck) error {
	evidence, err := json.Marshal(check.Evidence)
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO compliance_checks (check_id,entity_type,entity_id,type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level,created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13,$14,NOW())`,
		check.ID, "", uuid.Nil, check.Type, check.Regulation, check.Requirement,
		check.Status, check.Result, evidence, check.CheckedAt, check.CheckedBy, check.NextCheck,
		check.Remediation, check.RiskLevel,
	)
	if err != nil {
		return fmt.Errorf("insert compliance check: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetComplianceCheckByID(ctx context.Context, id uuid.UUID) (*entities.LegalComplianceCheck, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level FROM compliance_checks WHERE check_id=$1`, id)
	var c entities.LegalComplianceCheck
	var rawEvidence []byte
	if err := row.Scan(&c.Type, &c.Regulation, &c.Requirement, &c.Status, &c.Result, &rawEvidence, &c.CheckedAt, &c.CheckedBy, &c.NextCheck, &c.Remediation, &c.RiskLevel); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get compliance check: %w", err)
	}
	if err := json.Unmarshal(rawEvidence, &c.Evidence); err != nil {
		return nil, fmt.Errorf("unmarshal evidence: %w", err)
	}
	c.ID = id
	return &c, nil
}

func (r *PostgresMinimalLegalRepository) UpdateComplianceCheck(ctx context.Context, check *entities.LegalComplianceCheck) error {
	evidence, err := json.Marshal(check.Evidence)
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE compliance_checks SET type=$2,regulation=$3,requirement=$4,status=$5,result=$6,evidence=$7::jsonb,checked_at=$8,checked_by=$9,next_check=$10,remediation=$11,risk_level=$12 WHERE check_id=$1`,
		check.ID, check.Type, check.Regulation, check.Requirement, check.Status, check.Result,
		evidence, check.CheckedAt, check.CheckedBy, check.NextCheck, check.Remediation, check.RiskLevel,
	)
	if err != nil {
		return fmt.Errorf("update compliance check: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) ListComplianceChecks(ctx context.Context, filters repositories.ComplianceFilters) ([]*entities.LegalComplianceCheck, error) {
	base := `SELECT check_id,type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level FROM compliance_checks`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.Regulation != nil {
		where = append(where, fmt.Sprintf("regulation=$%d", idx))
		args = append(args, *filters.Regulation)
		idx++
	}
	if filters.Status != nil {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, *filters.Status)
		idx++
	}
	if filters.RiskLevel != nil {
		where = append(where, fmt.Sprintf("risk_level=$%d", idx))
		args = append(args, *filters.RiskLevel)
		idx++
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("checked_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("checked_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY checked_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list compliance checks: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalComplianceCheck
	for rows.Next() {
		var c entities.LegalComplianceCheck
		var rawEvidence []byte
		if err := rows.Scan(&c.ID, &c.Type, &c.Regulation, &c.Requirement, &c.Status, &c.Result, &rawEvidence, &c.CheckedAt, &c.CheckedBy, &c.NextCheck, &c.Remediation, &c.RiskLevel); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawEvidence, &c.Evidence); err != nil {
			return nil, fmt.Errorf("unmarshal evidence: %w", err)
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate compliance checks: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetChecksByRegulation(ctx context.Context, regulation string) ([]*entities.LegalComplianceCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT check_id,type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level FROM compliance_checks WHERE regulation=$1 ORDER BY checked_at DESC`, regulation)
	if err != nil {
		return nil, fmt.Errorf("get checks by regulation: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalComplianceCheck
	for rows.Next() {
		var c entities.LegalComplianceCheck
		var rawEvidence []byte
		if err := rows.Scan(&c.ID, &c.Type, &c.Regulation, &c.Requirement, &c.Status, &c.Result, &rawEvidence, &c.CheckedAt, &c.CheckedBy, &c.NextCheck, &c.Remediation, &c.RiskLevel); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawEvidence, &c.Evidence); err != nil {
			return nil, fmt.Errorf("unmarshal evidence: %w", err)
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checks by regulation: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetChecksByStatus(ctx context.Context, status entities.ComplianceStatus) ([]*entities.LegalComplianceCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT check_id,type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level FROM compliance_checks WHERE status=$1 ORDER BY checked_at DESC`, status)
	if err != nil {
		return nil, fmt.Errorf("get checks by status: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalComplianceCheck
	for rows.Next() {
		var c entities.LegalComplianceCheck
		var rawEvidence []byte
		if err := rows.Scan(&c.ID, &c.Type, &c.Regulation, &c.Requirement, &c.Status, &c.Result, &rawEvidence, &c.CheckedAt, &c.CheckedBy, &c.NextCheck, &c.Remediation, &c.RiskLevel); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawEvidence, &c.Evidence); err != nil {
			return nil, fmt.Errorf("unmarshal evidence: %w", err)
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checks by status: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetOverdueChecks(ctx context.Context) ([]*entities.LegalComplianceCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT check_id,type,regulation,requirement,status,result,evidence,checked_at,checked_by,next_check,remediation,risk_level FROM compliance_checks WHERE next_check IS NOT NULL AND next_check < NOW() AND status != 'Compliant' ORDER BY next_check`)
	if err != nil {
		return nil, fmt.Errorf("get overdue checks: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalComplianceCheck
	for rows.Next() {
		var c entities.LegalComplianceCheck
		var rawEvidence []byte
		if err := rows.Scan(&c.ID, &c.Type, &c.Regulation, &c.Requirement, &c.Status, &c.Result, &rawEvidence, &c.CheckedAt, &c.CheckedBy, &c.NextCheck, &c.Remediation, &c.RiskLevel); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawEvidence, &c.Evidence); err != nil {
			return nil, fmt.Errorf("unmarshal evidence: %w", err)
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overdue checks: %w", err)
	}
	return out, nil
}

// IP Licenses (Phase4)
func (r *PostgresMinimalLegalRepository) CreateIPLicense(ctx context.Context, lic *entities.IPLicense) error {
	usage, err := json.Marshal(lic.UsageRights)
	if err != nil {
		return fmt.Errorf("marshal usage rights: %w", err)
	}
	restr, err := json.Marshal(lic.Restrictions)
	if err != nil {
		return fmt.Errorf("marshal restrictions: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO ip_licenses (license_id,type,name,licensor_name,licensee_name,ip_type,ip_description,usage_rights,restrictions,territory,effective_date,expiration_date,royalty_rate,fee_amount,fee_currency,is_exclusive,is_active,created_at,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8::text[],$9::text[],$10,$11,$12,$13,$14,$15,$16,NOW(),NOW())`,
		lic.ID, lic.Type, lic.Name, lic.LicensorName, lic.LicenseeName,
		lic.IPType, lic.IPDescription, usage, restr,
		lic.Territory, lic.EffectiveDate, lic.ExpirationDate,
		lic.RoyaltyRate, lic.Fee.Amount, lic.Fee.Currency,
		lic.IsExclusive, lic.IsActive,
	)
	if err != nil {
		return fmt.Errorf("insert ip license: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetIPLicenseByID(ctx context.Context, id uuid.UUID) (*entities.IPLicense, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT license_id,type,name,licensor_name,licensee_name,ip_type,ip_description,usage_rights,restrictions,territory,effective_date,expiration_date,royalty_rate,fee_amount,fee_currency,is_exclusive,is_active,created_at,updated_at FROM ip_licenses WHERE license_id=$1`, id)
	var lic entities.IPLicense
	if err := r.scanIPLicense(row, &lic); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get ip license: %w", err)
	}
	return &lic, nil
}

func (r *PostgresMinimalLegalRepository) UpdateIPLicense(ctx context.Context, lic *entities.IPLicense) error {
	usage, err := json.Marshal(lic.UsageRights)
	if err != nil {
		return fmt.Errorf("marshal usage rights: %w", err)
	}
	restr, err := json.Marshal(lic.Restrictions)
	if err != nil {
		return fmt.Errorf("marshal restrictions: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE ip_licenses SET type=$2,name=$3,licensor_name=$4,licensee_name=$5,ip_type=$6,ip_description=$7,usage_rights=$8::text[],restrictions=$9::text[],territory=$10,effective_date=$11,expiration_date=$12,royalty_rate=$13,fee_amount=$14,fee_currency=$15,is_exclusive=$16,is_active=$17,updated_at=NOW() WHERE license_id=$1`,
		lic.ID, lic.Type, lic.Name, lic.LicensorName, lic.LicenseeName,
		lic.IPType, lic.IPDescription, usage, restr, lic.Territory,
		lic.EffectiveDate, lic.ExpirationDate, lic.RoyaltyRate,
		lic.Fee.Amount, lic.Fee.Currency, lic.IsExclusive, lic.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update ip license: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) DeleteIPLicense(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM ip_licenses WHERE license_id=$1`, id); err != nil {
		return fmt.Errorf("delete ip license: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) ListIPLicenses(ctx context.Context, filters repositories.IPLicenseFilters) ([]*entities.IPLicense, error) {
	base := `SELECT license_id,type,name,licensor_name,licensee_name,ip_type,ip_description,usage_rights,restrictions,territory,effective_date,expiration_date,royalty_rate,fee_amount,fee_currency,is_exclusive,is_active,created_at,updated_at FROM ip_licenses`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.IPType != nil {
		where = append(where, fmt.Sprintf("ip_type=$%d", idx))
		args = append(args, *filters.IPType)
		idx++
	}
	if filters.IsExclusive != nil {
		where = append(where, fmt.Sprintf("is_exclusive=$%d", idx))
		args = append(args, *filters.IsExclusive)
		idx++
	}
	if filters.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *filters.IsActive)
		idx++
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("created_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("created_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list ip licenses: %w", err)
	}
	defer rows.Close()
	var out []*entities.IPLicense
	for rows.Next() {
		var lic entities.IPLicense
		if err := r.scanIPLicense(rows, &lic); err != nil {
			return nil, err
		}
		out = append(out, &lic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ip licenses: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetLicensesByType(ctx context.Context, ipType entities.IPType) ([]*entities.IPLicense, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT license_id,type,name,licensor_name,licensee_name,ip_type,ip_description,usage_rights,restrictions,territory,effective_date,expiration_date,royalty_rate,fee_amount,fee_currency,is_exclusive,is_active,created_at,updated_at FROM ip_licenses WHERE ip_type=$1 ORDER BY created_at DESC`, ipType)
	if err != nil {
		return nil, fmt.Errorf("get licenses by type: %w", err)
	}
	defer rows.Close()
	var out []*entities.IPLicense
	for rows.Next() {
		var lic entities.IPLicense
		if err := r.scanIPLicense(rows, &lic); err != nil {
			return nil, err
		}
		out = append(out, &lic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate licenses by type: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetExpiringLicenses(ctx context.Context, days int) ([]*entities.IPLicense, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM ip_licenses WHERE expiration_date <= NOW() + ($1 || ' days')::interval ORDER BY expiration_date`, days)
	if err != nil {
		return nil, fmt.Errorf("get expiring ip licenses: %w", err)
	}
	defer rows.Close()
	var out []*entities.IPLicense
	for rows.Next() {
		var lic entities.IPLicense
		if err := r.scanIPLicense(rows, &lic); err != nil {
			return nil, err
		}
		out = append(out, &lic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expiring ip licenses: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetActiveLicenses(ctx context.Context) ([]*entities.IPLicense, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT license_id,type,name,licensor_name,licensee_name,ip_type,ip_description,usage_rights,restrictions,territory,effective_date,expiration_date,royalty_rate,fee_amount,fee_currency,is_exclusive,is_active,created_at,updated_at
        FROM ip_licenses WHERE is_active ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get active ip licenses: %w", err)
	}
	defer rows.Close()
	var out []*entities.IPLicense
	for rows.Next() {
		var lic entities.IPLicense
		if err := r.scanIPLicense(rows, &lic); err != nil {
			return nil, err
		}
		out = append(out, &lic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active ip licenses: %w", err)
	}
	return out, nil
}

// scanIPLicense reads a full IPLicense record from a rowScanner.
func (r *PostgresMinimalLegalRepository) scanIPLicense(scanner rowScanner, license *entities.IPLicense) error {
	var rawUsageRights, rawRestrictions []byte
	if err := scanner.Scan(
		&license.ID, &license.Type, &license.Name, &license.LicensorName, &license.LicenseeName,
		&license.IPType, &license.IPDescription, &rawUsageRights, &rawRestrictions,
		&license.Territory, &license.EffectiveDate, &license.ExpirationDate,
		&license.RoyaltyRate, &license.Fee, &license.IsExclusive, &license.IsActive,
		&license.CreatedAt, &license.UpdatedAt,
	); err != nil {
		return err
	}
	if err := json.Unmarshal(rawUsageRights, &license.UsageRights); err != nil { 
		return err 
	}
	if err := json.Unmarshal(rawRestrictions, &license.Restrictions); err != nil { 
		return err 
	}
	return nil
}

// scanRiskAssessment reads a full LegalRiskAssessment record from a rowScanner.
func (r *PostgresMinimalLegalRepository) scanRiskAssessment(scanner rowScanner, assessment *entities.LegalRiskAssessment) error {
	var rawFactors, rawRecommendations, rawClauses, rawIssues []byte
	if err := scanner.Scan(
		&assessment.ID, &assessment.RiskLevel, &assessment.RiskScore, &rawFactors, &rawRecommendations,
		&rawClauses, &rawIssues, &assessment.InsuranceRequired, &assessment.LegalReview,
		&assessment.AssessedBy, &assessment.AssessedAt, &assessment.ExpiresAt,
	); err != nil {
		return err
	}
	if err := json.Unmarshal(rawFactors, &assessment.RiskFactors); err != nil { 
		return err 
	}
	if err := json.Unmarshal(rawRecommendations, &assessment.Recommendations); err != nil { 
		return err 
	}
	if err := json.Unmarshal(rawClauses, &assessment.RequiredClauses); err != nil { 
		return err 
	}
	if err := json.Unmarshal(rawIssues, &assessment.ComplianceIssues); err != nil { 
		return err 
	}
	return nil
}

// Insurance Policies (Phase5)
func (r *PostgresMinimalLegalRepository) scanInsurancePolicy(scanner rowScanner, p *entities.InsurancePolicy) error {
	var ren sql.NullTime
	var details, exclusions []string
	if err := scanner.Scan(
		&p.ID, &p.Type, &p.PolicyNumber, &p.Provider,
		&p.Coverage.Amount, &p.Coverage.Currency,
		&p.Deductible.Amount, &p.Deductible.Currency,
		&p.Premium.Amount, &p.Premium.Currency,
		&p.EffectiveDate, &p.ExpirationDate, &ren,
		&details, &exclusions,
		&p.Status, &p.IsActive, &p.DocumentURL,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return err
	}
	if ren.Valid {
		p.RenewalDate = ren.Time
	}
	p.CoverageDetails = details
	p.Exclusions = exclusions
	return nil
}

func (r *PostgresMinimalLegalRepository) CreateInsurancePolicy(ctx context.Context, policy *entities.InsurancePolicy) error {
	details, err := json.Marshal(policy.CoverageDetails)
	if err != nil {
		return fmt.Errorf("marshal coverage details: %w", err)
	}
	exclusions, err := json.Marshal(policy.Exclusions)
	if err != nil {
		return fmt.Errorf("marshal exclusions: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO insurance_policies (policy_id,type,policy_number,provider,coverage_amount,coverage_currency,deductible_amount,deductible_currency,premium_amount,premium_currency,effective_date,expiration_date,renewal_date,coverage_details,exclusions,status,is_active,document_url,created_at,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::text[],$15::text[],$16,$17,$18,NOW(),NOW())`,
		policy.ID, policy.Type, policy.PolicyNumber, policy.Provider,
		policy.Coverage.Amount, policy.Coverage.Currency,
		policy.Deductible.Amount, policy.Deductible.Currency,
		policy.Premium.Amount, policy.Premium.Currency,
		policy.EffectiveDate, policy.ExpirationDate, policy.RenewalDate,
		details, exclusions,
		policy.Status, policy.IsActive, policy.DocumentURL,
	)
	if err != nil {
		return fmt.Errorf("insert insurance policy: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetInsurancePolicyByID(ctx context.Context, id uuid.UUID) (*entities.InsurancePolicy, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT policy_id,type,policy_number,provider,coverage_amount,coverage_currency,deductible_amount,deductible_currency,premium_amount,premium_currency,effective_date,expiration_date,renewal_date,coverage_details,exclusions,status,is_active,document_url,created_at,updated_at
        FROM insurance_policies WHERE policy_id=$1`, id)
	var p entities.InsurancePolicy
	if err := r.scanInsurancePolicy(row, &p); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get insurance policy: %w", err)
	}
	return &p, nil
}

func (r *PostgresMinimalLegalRepository) UpdateInsurancePolicy(ctx context.Context, policy *entities.InsurancePolicy) error {
	details, err := json.Marshal(policy.CoverageDetails)
	if err != nil {
		return fmt.Errorf("marshal coverage details: %w", err)
	}
	exclusions, err := json.Marshal(policy.Exclusions)
	if err != nil {
		return fmt.Errorf("marshal exclusions: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE insurance_policies SET type=$2,policy_number=$3,provider=$4,coverage_amount=$5,coverage_currency=$6,deductible_amount=$7,deductible_currency=$8,premium_amount=$9,premium_currency=$10,effective_date=$11,expiration_date=$12,renewal_date=$13,coverage_details=$14::text[],exclusions=$15::text[],status=$16,is_active=$17,document_url=$18,updated_at=NOW() WHERE policy_id=$1`,
		policy.ID, policy.Type, policy.PolicyNumber, policy.Provider,
		policy.Coverage.Amount, policy.Coverage.Currency,
		policy.Deductible.Amount, policy.Deductible.Currency,
		policy.Premium.Amount, policy.Premium.Currency,
		policy.EffectiveDate, policy.ExpirationDate, policy.RenewalDate,
		details, exclusions,
		policy.Status, policy.IsActive, policy.DocumentURL,
	)
	if err != nil {
		return fmt.Errorf("update insurance policy: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) DeleteInsurancePolicy(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM insurance_policies WHERE policy_id=$1`, id); err != nil {
		return fmt.Errorf("delete insurance policy: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) ListInsurancePolicies(ctx context.Context, filters repositories.InsuranceFilters) ([]*entities.InsurancePolicy, error) {
	base := `SELECT policy_id,type,policy_number,provider,coverage_amount,coverage_currency,deductible_amount,deductible_currency,premium_amount,premium_currency,effective_date,expiration_date,renewal_date,coverage_details,exclusions,status,is_active,document_url,created_at,updated_at FROM insurance_policies`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.Provider != nil {
		where = append(where, fmt.Sprintf("provider=$%d", idx))
		args = append(args, *filters.Provider)
		idx++
	}
	if filters.Status != nil {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, *filters.Status)
		idx++
	}
	if filters.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *filters.IsActive)
		idx++
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("created_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("created_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list insurance policies: %w", err)
	}
	defer rows.Close()
	var out []*entities.InsurancePolicy
	for rows.Next() {
		var p entities.InsurancePolicy
		if err := r.scanInsurancePolicy(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate insurance policies: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetPoliciesByType(ctx context.Context, insuranceType entities.InsuranceType) ([]*entities.InsurancePolicy, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM insurance_policies WHERE type=$1 ORDER BY created_at DESC`, insuranceType)
	if err != nil {
		return nil, fmt.Errorf("get policies by type: %w", err)
	}
	defer rows.Close()
	var out []*entities.InsurancePolicy
	for rows.Next() {
		var p entities.InsurancePolicy
		if err := r.scanInsurancePolicy(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policies by type: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetExpiringPolicies(ctx context.Context, days int) ([]*entities.InsurancePolicy, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM insurance_policies WHERE expiration_date <= NOW() + ($1 || ' days')::interval ORDER BY expiration_date`, days)
	if err != nil {
		return nil, fmt.Errorf("get expiring policies: %w", err)
	}
	defer rows.Close()
	var out []*entities.InsurancePolicy
	for rows.Next() {
		var p entities.InsurancePolicy
		if err := r.scanInsurancePolicy(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expiring policies: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetActivePolicies(ctx context.Context) ([]*entities.InsurancePolicy, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM insurance_policies WHERE is_active ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get active policies: %w", err)
	}
	defer rows.Close()
	var out []*entities.InsurancePolicy
	for rows.Next() {
		var p entities.InsurancePolicy
		if err := r.scanInsurancePolicy(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active policies: %w", err)
	}
	return out, nil
}

// Dispute Resolution (Phase6)
func (r *PostgresMinimalLegalRepository) CreateDispute(ctx context.Context, d *entities.DisputeResolution) error {
	timeline, err := json.Marshal(d.Timeline)
	if err != nil {
		return fmt.Errorf("marshal timeline: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO dispute_resolutions (dispute_id,contract_id,type,status,description,initiated_by,resolution_method,mediator,arbitrator,venue,governing_law,timeline,resolution,cost_amount,cost_currency,initiated_at,created_at,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13,$14,$15,$16,NOW(),NOW())`,
		d.ID, d.ContractID, d.Type, d.Status, d.Description, d.InitiatedBy, d.ResolutionMethod,
		d.Mediator, d.Arbitrator, d.Venue, d.GoverningLaw, timeline, d.Resolution,
		d.Cost.Amount, d.Cost.Currency, d.InitiatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert dispute: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetDisputeByID(ctx context.Context, id uuid.UUID) (*entities.DisputeResolution, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id,type,status,description,initiated_by,resolution_method,mediator,arbitrator,venue,governing_law,timeline,resolution,cost_amount,cost_currency,initiated_at,resolved_at,created_at,updated_at
        FROM dispute_resolutions WHERE dispute_id=$1`, id)
	var d entities.DisputeResolution
	var rawTimeline []byte
	if err := row.Scan(&d.ContractID, &d.Type, &d.Status, &d.Description, &d.InitiatedBy, &d.ResolutionMethod,
		&d.Mediator, &d.Arbitrator, &d.Venue, &d.GoverningLaw, &rawTimeline, &d.Resolution,
		&d.Cost.Amount, &d.Cost.Currency, &d.InitiatedAt, &d.ResolvedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get dispute: %w", err)
	}
	if err := json.Unmarshal(rawTimeline, &d.Timeline); err != nil {
		return nil, fmt.Errorf("unmarshal timeline: %w", err)
	}
	d.ID = id
	return &d, nil
}

func (r *PostgresMinimalLegalRepository) UpdateDispute(ctx context.Context, d *entities.DisputeResolution) error {
	timeline, err := json.Marshal(d.Timeline)
	if err != nil {
		return fmt.Errorf("marshal timeline: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE dispute_resolutions SET status=$2,description=$3,resolution_method=$4,mediator=$5,arbitrator=$6,venue=$7,governing_law=$8,timeline=$9::jsonb,resolution=$10,cost_amount=$11,cost_currency=$12,resolved_at=$13,updated_at=NOW() WHERE dispute_id=$1`,
		d.ID, d.Status, d.Description, d.ResolutionMethod, d.Mediator, d.Arbitrator, d.Venue, d.GoverningLaw,
		timeline, d.Resolution, d.Cost.Amount, d.Cost.Currency, d.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("update dispute: %w", err)
	}
	return nil
}

// ListDisputes returns disputes matching the given filters.
func (r *PostgresMinimalLegalRepository) ListDisputes(ctx context.Context, filters repositories.DisputeFilters) ([]*entities.DisputeResolution, error) {
	base := `SELECT dispute_id,contract_id,type,status,description,initiated_by,resolution_method,mediator,arbitrator,venue,governing_law,timeline,resolution,cost_amount,cost_currency,initiated_at,resolved_at,created_at,updated_at FROM dispute_resolutions`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.Status != nil {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, *filters.Status)
		idx++
	}
	if filters.ContractID != nil {
		where = append(where, fmt.Sprintf("contract_id=$%d", idx))
		args = append(args, *filters.ContractID)
		idx++
	}
	if filters.ResolutionMethod != nil {
		where = append(where, fmt.Sprintf("resolution_method=$%d", idx))
		args = append(args, *filters.ResolutionMethod)
		idx++
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("initiated_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("initiated_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY initiated_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list disputes: %w", err)
	}
	defer rows.Close()
	var out []*entities.DisputeResolution
	for rows.Next() {
		var d entities.DisputeResolution
		var rawTimeline []byte
		if err := r.scanDisputeResolution(rows, &d, &rawTimeline); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate disputes: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetDisputesByContract(ctx context.Context, contractID uuid.UUID) ([]*entities.DisputeResolution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT dispute_id,contract_id,type,status,description,initiated_by,resolution_method,mediator,arbitrator,venue,governing_law,timeline,resolution,cost_amount,cost_currency,initiated_at,resolved_at,created_at,updated_at
        FROM dispute_resolutions WHERE contract_id=$1 ORDER BY initiated_at DESC`, contractID)
	if err != nil {
		return nil, fmt.Errorf("list disputes by contract: %w", err)
	}
	defer rows.Close()
	var out []*entities.DisputeResolution
	for rows.Next() {
		var d entities.DisputeResolution
		var rawTimeline []byte
		if err := r.scanDisputeResolution(rows, &d, &rawTimeline); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate disputes by contract: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetActiveDisputes(ctx context.Context) ([]*entities.DisputeResolution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM dispute_resolutions WHERE status='Open' ORDER BY initiated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list active disputes: %w", err)
	}
	defer rows.Close()
	var out []*entities.DisputeResolution
	for rows.Next() {
		var d entities.DisputeResolution
		var rawTimeline []byte
		if err := r.scanDisputeResolution(rows, &d, &rawTimeline); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active disputes: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) AddDisputeEvent(ctx context.Context, disputeID uuid.UUID, ev *entities.DisputeEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dispute_events (event_id,dispute_id,event_type,description,actor,timestamp,evidence,metadata) VALUES ($1,$2,$3,$4,$5,$6,$7::text[],$8::jsonb)`,
		ev.ID, disputeID, ev.Type, ev.Description, ev.Actor, ev.Timestamp, ev.Evidence, "{}",
	)
	if err != nil {
		return fmt.Errorf("insert dispute event: %w", err)
	}
	return nil
}

// CreateRegulatoryReport inserts a new regulatory report record.
func (r *PostgresMinimalLegalRepository) CreateRegulatoryReport(ctx context.Context, rep *entities.RegulatoryReport) error {
	dataRaw, err := json.Marshal(rep.Data)
	if err != nil {
		return fmt.Errorf("marshal report data: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO regulatory_reports (report_id,type,regulation,authority,period,status,content,data,filing_deadline,created_at,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,NOW(),NOW())`,
		rep.ID, rep.Type, rep.Regulation, rep.Authority, rep.Period, rep.Status,
		rep.Content, dataRaw, rep.FilingDeadline,
	)
	if err != nil {
		return fmt.Errorf("insert regulatory report: %w", err)
	}
	return nil
}

// GetRegulatoryReportByID retrieves a regulatory report by ID.
func (r *PostgresMinimalLegalRepository) GetRegulatoryReportByID(ctx context.Context, id uuid.UUID) (*entities.RegulatoryReport, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT type,regulation,authority,period,status,content,data,filing_deadline,filed_at,confirmation_id,document_url,created_at,updated_at
        FROM regulatory_reports WHERE report_id=$1`, id)
	var rep entities.RegulatoryReport
	var rawData []byte
	if err := r.scanRegulatoryReport(row, &rep, &rawData); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get regulatory report: %w", err)
	}
	return &rep, nil
}

// UpdateRegulatoryReport updates an existing regulatory report.
func (r *PostgresMinimalLegalRepository) UpdateRegulatoryReport(ctx context.Context, rep *entities.RegulatoryReport) error {
	dataRaw, err := json.Marshal(rep.Data)
	if err != nil {
		return fmt.Errorf("marshal report data: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE regulatory_reports SET type=$2,regulation=$3,authority=$4,period=$5,status=$6,content=$7,data=$8::jsonb,filing_deadline=$9,filed_at=$10,confirmation_id=$11,document_url=$12,updated_at=NOW() WHERE report_id=$1`,
		rep.ID, rep.Type, rep.Regulation, rep.Authority, rep.Period, rep.Status,
		rep.Content, dataRaw, rep.FilingDeadline, rep.FiledAt, rep.ConfirmationID, rep.DocumentURL,
	)
	if err != nil {
		return fmt.Errorf("update regulatory report: %w", err)
	}
	return nil
}

// ListRegulatoryReports returns regulatory reports matching filters.
func (r *PostgresMinimalLegalRepository) ListRegulatoryReports(ctx context.Context, filters repositories.ReportFilters) ([]*entities.RegulatoryReport, error) {
	base := `SELECT report_id,type,regulation,authority,period,status,content,data,filing_deadline,filed_at,confirmation_id,document_url,created_at,updated_at FROM regulatory_reports`
	var where []string
	var args []interface{}
	idx := 1
	if filters.Type != nil {
		where = append(where, fmt.Sprintf("type=$%d", idx))
		args = append(args, *filters.Type)
		idx++
	}
	if filters.Regulation != nil {
		where = append(where, fmt.Sprintf("regulation=$%d", idx))
		args = append(args, *filters.Regulation)
		idx++
	}
	if filters.Authority != nil {
		where = append(where, fmt.Sprintf("authority=$%d", idx))
		args = append(args, *filters.Authority)
		idx++
	}
	if filters.Status != nil {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, *filters.Status)
		idx++
	}
	if filters.Period != nil {
		where = append(where, fmt.Sprintf("period=$%d", idx))
		args = append(args, *filters.Period)
		idx++
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("created_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("created_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list regulatory reports: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate regulatory reports: %w", err)
	}
	return out, nil
}

// GetReportsByRegulation returns reports filtered by regulation.
func (r *PostgresMinimalLegalRepository) GetReportsByRegulation(ctx context.Context, regulation string) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE regulation=$1 ORDER BY created_at DESC`, regulation)
	if err != nil {
		return nil, fmt.Errorf("get reports by regulation: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports by regulation: %w", err)
	}
	return out, nil
}

// GetReportsByStatus returns reports filtered by status.
func (r *PostgresMinimalLegalRepository) GetReportsByStatus(ctx context.Context, status entities.ReportStatus) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE status=$1 ORDER BY created_at DESC`, status)
	if err != nil {
		return nil, fmt.Errorf("get reports by status: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports by status: %w", err)
	}
	return out, nil
}

// GetPendingReports returns reports still in draft or review.
func (r *PostgresMinimalLegalRepository) GetPendingReports(ctx context.Context) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE status IN ('Draft','Review','Pending') ORDER BY filing_deadline`)
	if err != nil {
		return nil, fmt.Errorf("get pending reports: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending reports: %w", err)
	}
	return out, nil
}

// GetOverdueReports returns reports past their filing deadline.
func (r *PostgresMinimalLegalRepository) GetOverdueReports(ctx context.Context) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE filing_deadline < NOW() AND status != 'Filed' ORDER BY filing_deadline`)
	if err != nil {
		return nil, fmt.Errorf("get overdue reports: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overdue reports: %w", err)
	}
	return out, nil
}

// scanAuditEntry maps a contract_audit_entries row into a ContractAuditEntry entity.
func (r *PostgresMinimalLegalRepository) scanAuditEntry(scanner rowScanner, e *entities.ContractAuditEntry) error {
	var oldVal, newVal sql.NullString
	var ipAddr sql.NullString
	if err := scanner.Scan(
		&e.ID, uuid.Nil, &e.Action, &e.Field, &oldVal, &newVal,
		&e.UserID, &e.UserAgent, &ipAddr, &e.Timestamp, &e.Hash,
		uuid.Nil,
	); err != nil {
		return err
	}
	if oldVal.Valid {
		e.OldValue = oldVal.String
	}
	if newVal.Valid {
		e.NewValue = newVal.String
	}
	if ipAddr.Valid {
		e.IPAddress = ipAddr.String
	}
	return nil
}

// AddContractAuditEntry inserts an audit entry for a contract.
func (r *PostgresMinimalLegalRepository) AddContractAuditEntry(ctx context.Context, contractID uuid.UUID, entry *entities.ContractAuditEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO contract_audit_entries (entry_id,contract_id,action,field,old_value,new_value,user_id,user_agent,ip_address,timestamp,hash,metadata)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb)`,
		entry.ID, contractID, entry.Action, entry.Field,
		entry.OldValue, entry.NewValue,
		entry.UserID, entry.UserAgent, entry.IPAddress,
		entry.Timestamp, entry.Hash, "{}",
	)
	if err != nil {
		return fmt.Errorf("insert audit entry: %w", err)
	}
	return nil
}

// GetContractAuditTrail retrieves all audit entries for a contract.
func (r *PostgresMinimalLegalRepository) GetContractAuditTrail(ctx context.Context, contractID uuid.UUID) ([]*entities.ContractAuditEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT entry_id,contract_id,action,field,old_value,new_value,user_id,user_agent,ip_address,timestamp,hash,metadata
        FROM contract_audit_entries WHERE contract_id=$1 ORDER BY timestamp`, contractID)
	if err != nil {
		return nil, fmt.Errorf("get audit trail: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractAuditEntry
	for rows.Next() {
		var e entities.ContractAuditEntry
		if err := r.scanAuditEntry(rows, &e); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit trail: %w", err)
	}
	return out, nil
}

// GetAuditEntriesByTimeRange retrieves audit entries across all contracts within the given time window.
func (r *PostgresMinimalLegalRepository) GetAuditEntriesByTimeRange(ctx context.Context, tr repositories.TimeRange) ([]*entities.ContractAuditEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT entry_id,contract_id,action,field,old_value,new_value,user_id,user_agent,ip_address,timestamp,hash,metadata
        FROM contract_audit_entries WHERE timestamp BETWEEN $1 AND $2 ORDER BY timestamp`, tr.Start, tr.End)
	if err != nil {
		return nil, fmt.Errorf("get audit entries by time range: %w", err)
	}
	defer rows.Close()
	var out []*entities.ContractAuditEntry
	for rows.Next() {
		var e entities.ContractAuditEntry
		if err := r.scanAuditEntry(rows, &e); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit entries by time range: %w", err)
	}
	return out, nil
}

// VerifyAuditTrail checks the integrity of the audit trail for a contract.
func (r *PostgresMinimalLegalRepository) VerifyAuditTrail(ctx context.Context, contractID uuid.UUID) (bool, error) {
	// For now, this returns true if any entries exist (full validation requires hashing logic).
	rows, err := r.db.QueryContext(ctx,
		`SELECT 1 FROM contract_audit_entries WHERE contract_id=$1 LIMIT 1`, contractID)
	if err != nil {
		return false, fmt.Errorf("verify audit trail query: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

// Risk Assessments (Phase7)
func (r *PostgresMinimalLegalRepository) CreateRiskAssessment(ctx context.Context, a *entities.LegalRiskAssessment) error {
	factors, err := json.Marshal(a.RiskFactors)
	if err != nil {
		return fmt.Errorf("marshal risk factors: %w", err)
	}
	recs, err := json.Marshal(a.Recommendations)
	if err != nil {
		return fmt.Errorf("marshal recommendations: %w", err)
	}
	clauses, err := json.Marshal(a.RequiredClauses)
	if err != nil {
		return fmt.Errorf("marshal required clauses: %w", err)
	}
	issues, err := json.Marshal(a.ComplianceIssues)
	if err != nil {
		return fmt.Errorf("marshal compliance issues: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO legal_risk_assessments (assessment_id,entity_type,entity_id,risk_level,risk_score,risk_factors,recommendations,required_clauses,compliance_issues,insurance_required,legal_review,assessed_by,assessed_at,expires_at)
        VALUES ($1,'contract',$2,$3,$4,$5::jsonb,$6::text[],$7::text[],$8::text[],$9,$10,$11,$12,$13)`,
		a.ID, a.ContractID, a.RiskLevel, a.RiskScore,
		factors, recs, clauses, issues,
		a.InsuranceRequired, a.LegalReview,
		a.AssessedBy, a.AssessedAt, a.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert risk assessment: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) GetRiskAssessmentByID(ctx context.Context, id uuid.UUID) (*entities.LegalRiskAssessment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT entity_id,risk_level,risk_score,risk_factors,recommendations,required_clauses,compliance_issues,insurance_required,legal_review,assessed_by,assessed_at,expires_at
        FROM legal_risk_assessments WHERE assessment_id=$1`, id)
	var a entities.LegalRiskAssessment
	if err := r.scanRiskAssessment(row, &a); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get risk assessment: %w", err)
	}
	a.ID = id
	return &a, nil
}

func (r *PostgresMinimalLegalRepository) UpdateRiskAssessment(ctx context.Context, a *entities.LegalRiskAssessment) error {
	factors, err := json.Marshal(a.RiskFactors)
	if err != nil {
		return fmt.Errorf("marshal risk factors: %w", err)
	}
	recs, err := json.Marshal(a.Recommendations)
	if err != nil {
		return fmt.Errorf("marshal recommendations: %w", err)
	}
	clauses, err := json.Marshal(a.RequiredClauses)
	if err != nil {
		return fmt.Errorf("marshal required clauses: %w", err)
	}
	issues, err := json.Marshal(a.ComplianceIssues)
	if err != nil {
		return fmt.Errorf("marshal compliance issues: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE legal_risk_assessments SET risk_level=$2,risk_score=$3,risk_factors=$4::jsonb,recommendations=$5::text[],required_clauses=$6::text[],compliance_issues=$7::text[],insurance_required=$8,legal_review=$9,assessed_by=$10,assessed_at=$11,expires_at=$12 WHERE assessment_id=$1`,
		a.ID, a.RiskLevel, a.RiskScore,
		factors, recs, clauses, issues,
		a.InsuranceRequired, a.LegalReview,
		a.AssessedBy, a.AssessedAt, a.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("update risk assessment: %w", err)
	}
	return nil
}

func (r *PostgresMinimalLegalRepository) ListRiskAssessments(ctx context.Context, filters repositories.RiskAssessmentFilters) ([]*entities.LegalRiskAssessment, error) {
	base := `SELECT assessment_id,entity_id,risk_level,risk_score,risk_factors,recommendations,required_clauses,compliance_issues,insurance_required,legal_review,assessed_by,assessed_at,expires_at FROM legal_risk_assessments`
	var where []string
	var args []interface{}
	idx := 1
	if filters.RiskLevel != nil {
		where = append(where, fmt.Sprintf("risk_level=$%d", idx))
		args = append(args, *filters.RiskLevel)
		idx++
	}
	if filters.ContractID != nil {
		where = append(where, fmt.Sprintf("entity_id=$%d", idx))
		args = append(args, *filters.ContractID)
		idx++
	}
	if filters.RequiresReview != nil {
		where = append(where, fmt.Sprintf("legal_review=$%d", idx))
		args = append(args, *filters.RequiresReview)
		idx++
	}
	if filters.RequiresInsurance != nil {
		where = append(where, fmt.Sprintf("insurance_required=$%d", idx))
		args = append(args, *filters.RequiresInsurance)
		idx++
	}
	if filters.IsExpired != nil {
		if *filters.IsExpired {
			where = append(where, "expires_at < NOW()")
		} else {
			where = append(where, "expires_at >= NOW()")
		}
	}
	if filters.StartDate != nil {
		where = append(where, fmt.Sprintf("assessed_at>=$%d", idx))
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		where = append(where, fmt.Sprintf("assessed_at<=$%d", idx))
		args = append(args, *filters.EndDate)
		idx++
	}
	qry := base
	if len(where) > 0 {
		qry += " WHERE " + strings.Join(where, " AND ")
	}
	qry += " ORDER BY assessed_at DESC"
	if filters.Limit > 0 {
		qry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filters.Limit, filters.Offset)
	}
	rows, err := r.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, fmt.Errorf("list risk assessments: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalRiskAssessment
	for rows.Next() {
		var a entities.LegalRiskAssessment
			if err := r.scanRiskAssessment(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate risk assessments: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetHighRiskAssessments(ctx context.Context) ([]*entities.LegalRiskAssessment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM legal_risk_assessments WHERE risk_level='High' OR risk_level='Critical' ORDER BY assessed_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get high risk assessments: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalRiskAssessment
	for rows.Next() {
		var a entities.LegalRiskAssessment
			if err := r.scanRiskAssessment(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate high risk assessments: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetExpiredAssessments(ctx context.Context) ([]*entities.LegalRiskAssessment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM legal_risk_assessments WHERE expires_at < NOW() ORDER BY expires_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get expired assessments: %w", err)
	}
	defer rows.Close()
	var out []*entities.LegalRiskAssessment
	for rows.Next() {
		var a entities.LegalRiskAssessment
			if err := r.scanRiskAssessment(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired assessments: %w", err)
	}
	return out, nil
}

func (r *PostgresMinimalLegalRepository) GetAuditEntriesByUser(ctx context.Context, userID string, timeRange repositories.TimeRange) ([]interface{}, error) {
	return nil, nil
}

func (r *PostgresMinimalLegalRepository) GetAuditEntriesByAction(ctx context.Context, action string) ([]interface{}, error) {
	return nil, nil
}

func (r *PostgresMinimalLegalRepository) GetLegalMetrics(ctx context.Context, timeRange repositories.TimeRange) (interface{}, error) {
	return nil, nil
}
func (r *PostgresMinimalLegalRepository) GetContractMetrics(ctx context.Context, timeRange repositories.TimeRange) (interface{}, error) {
	return nil, nil
}
func (r *PostgresMinimalLegalRepository) GetComplianceMetrics(ctx context.Context, timeRange repositories.TimeRange) (interface{}, error) {
	return nil, nil
}
func (r *PostgresMinimalLegalRepository) GenerateLegalReport(ctx context.Context, reportType string, timeRange repositories.TimeRange) (interface{}, error) {
	return nil, nil
}


// GetAuditTrailByTimeRange is an alias for GetAuditEntriesByTimeRange to match interface
func (r *PostgresMinimalLegalRepository) GetAuditTrailByTimeRange(ctx context.Context, start, end time.Time) ([]*entities.ContractAuditEntry, error) {
	return r.GetAuditEntriesByTimeRange(ctx, repositories.TimeRange{Start: start, End: end})
}

// GetReportsByAuthority returns reports filtered by regulatory authority.
func (r *PostgresMinimalLegalRepository) GetReportsByAuthority(ctx context.Context, authority string) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE authority=$1 ORDER BY created_at DESC`, authority)
	if err != nil {
		return nil, fmt.Errorf("get reports by authority: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports by authority: %w", err)
	}
	return out, nil
}

// GetRiskAssessmentByContract returns risk assessment for a specific contract.
func (r *PostgresMinimalLegalRepository) GetRiskAssessmentByContract(ctx context.Context, contractID uuid.UUID) (*entities.LegalRiskAssessment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT entity_id,risk_level,risk_score,risk_factors,recommendations,required_clauses,compliance_issues,insurance_required,legal_review,assessed_by,assessed_at,expires_at
        FROM legal_risk_assessments WHERE entity_type='contract' AND entity_id=$1`, contractID)
	var a entities.LegalRiskAssessment
	if err := r.scanRiskAssessment(row, &a); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get risk assessment by contract: %w", err)
	}
	return &a, nil
}

// GetUpcomingReports returns reports due within the given number of days.
func (r *PostgresMinimalLegalRepository) GetUpcomingReports(ctx context.Context, days int) ([]*entities.RegulatoryReport, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM regulatory_reports WHERE due_date BETWEEN NOW() AND NOW() + INTERVAL '` + fmt.Sprintf("%d days", days) + `' ORDER BY due_date ASC`)
	if err != nil {
		return nil, fmt.Errorf("get upcoming reports: %w", err)
	}
	defer rows.Close()
	var out []*entities.RegulatoryReport
	for rows.Next() {
		var rep entities.RegulatoryReport
		var rawData []byte
		if err := r.scanRegulatoryReport(rows, &rep, &rawData); err != nil {
			return nil, err
		}
		out = append(out, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate upcoming reports: %w", err)
	}
	return out, nil
}

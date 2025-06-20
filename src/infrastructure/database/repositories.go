package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// PostgresClientRepository implements the ClientRepository interface
type PostgresClientRepository struct {
	db *sql.DB
}

// NewClientRepository creates a new PostgreSQL client repository
func NewClientRepository(db *sql.DB) repositories.ClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) FindByStatus(ctx context.Context, status entities.ClientStatus, offset, limit int) ([]*entities.Client, int, error) {
	countQuery := "SELECT COUNT(*) FROM clients WHERE status = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT client_id, name, contact_email, contact_phone, billing_address, 
			   timezone, created_at, updated_at, status
		FROM clients 
		WHERE status = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clients []*entities.Client
	for rows.Next() {
		client, err := r.scanClient(rows)
		if err != nil {
			return nil, 0, err
		}
		clients = append(clients, client)
	}

	return clients, total, rows.Err()
}

func (r *PostgresClientRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Client, error) {
	query := `
		SELECT client_id, name, contact_email, contact_phone, billing_address, 
			   timezone, created_at, updated_at, status
		FROM clients WHERE client_id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanClient(row)
}

func (r *PostgresClientRepository) FindByEmail(ctx context.Context, email string) (*entities.Client, error) {
	query := `
		SELECT client_id, name, contact_email, contact_phone, billing_address, 
			   timezone, created_at, updated_at, status
		FROM clients WHERE contact_email = $1`

	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanClient(row)
}

func (r *PostgresClientRepository) FindAll(ctx context.Context, offset, limit int) ([]*entities.Client, int, error) {
	countQuery := "SELECT COUNT(*) FROM clients"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT client_id, name, contact_email, contact_phone, billing_address, 
			   timezone, created_at, updated_at, status
		FROM clients 
		ORDER BY created_at DESC
		OFFSET $1 LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clients []*entities.Client
	for rows.Next() {
		client, err := r.scanClient(rows)
		if err != nil {
			return nil, 0, err
		}
		clients = append(clients, client)
	}

	return clients, total, rows.Err()
}

func (r *PostgresClientRepository) Save(ctx context.Context, client *entities.Client) error {
	if client.ClientID == uuid.Nil {
		return r.Create(ctx, client)
	}
	return r.Update(ctx, client)
}

func (r *PostgresClientRepository) Create(ctx context.Context, client *entities.Client) error {
	if client.ClientID == uuid.Nil {
		client.ClientID = uuid.New()
	}

	now := time.Now()
	client.CreatedAt = now
	client.UpdatedAt = now

	addressJSON, _ := json.Marshal(client.BillingAddress)

	query := `
		INSERT INTO clients (client_id, name, contact_email, contact_phone, billing_address, 
							timezone, created_at, updated_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		client.ClientID, client.Name, client.ContactEmail, client.ContactPhone,
		addressJSON, client.Timezone, client.CreatedAt, client.UpdatedAt, client.Status)

	return err
}

func (r *PostgresClientRepository) Update(ctx context.Context, client *entities.Client) error {
	client.UpdatedAt = time.Now()

	addressJSON, _ := json.Marshal(client.BillingAddress)

	query := `
		UPDATE clients SET 
			name = $2, contact_email = $3, contact_phone = $4, billing_address = $5,
			timezone = $6, updated_at = $7, status = $8
		WHERE client_id = $1`

	_, err := r.db.ExecContext(ctx, query,
		client.ClientID, client.Name, client.ContactEmail, client.ContactPhone,
		addressJSON, client.Timezone, client.UpdatedAt, client.Status)

	return err
}

func (r *PostgresClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM clients WHERE client_id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// PostgresClientProfileRepository implements the ClientProfileRepository interface
type PostgresClientProfileRepository struct {
	db *sql.DB
}

// NewClientProfileRepository creates a new PostgreSQL client profile repository
func NewClientProfileRepository(db *sql.DB) repositories.ClientProfileRepository {
	return &PostgresClientProfileRepository{db: db}
}

func (r *PostgresClientProfileRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ClientProfile, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresClientProfileRepository) FindByClientID(ctx context.Context, clientID uuid.UUID) (*entities.ClientProfile, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresClientProfileRepository) Save(ctx context.Context, profile *entities.ClientProfile) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresClientProfileRepository) Create(ctx context.Context, profile *entities.ClientProfile) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresClientProfileRepository) Update(ctx context.Context, profile *entities.ClientProfile) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresClientProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Placeholder implementation
	return nil
}

// PostgresProjectRepository implements the ProjectRepository interface
type PostgresProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new PostgreSQL project repository
func NewProjectRepository(db *sql.DB) repositories.ProjectRepository {
	return &PostgresProjectRepository{db: db}
}

func (r *PostgresProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Project, error) {
	query := `
		SELECT project_id, client_id, title, description, content_type, deadline, budget_amount,
			   budget_currency, priority, status, requirements, metadata, created_at, updated_at
		FROM projects WHERE project_id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanProject(row)
}

func (r *PostgresProjectRepository) FindByStatus(ctx context.Context, status entities.ProjectStatus, offset, limit int) ([]*entities.Project, int, error) {
	countQuery := "SELECT COUNT(*) FROM projects WHERE status = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT project_id, client_id, title, description, content_type, deadline, budget_amount,
			   budget_currency, priority, status, requirements, metadata, created_at, updated_at
		FROM projects WHERE status = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`
	
	rows, err := r.db.QueryContext(ctx, query, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	projects, err := r.scanProjects(rows)
	return projects, total, err
}

func (r *PostgresProjectRepository) FindActive(ctx context.Context, offset, limit int) ([]*entities.Project, int, error) {
	countQuery := "SELECT COUNT(*) FROM projects WHERE status NOT IN ('Completed', 'Cancelled')"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT project_id, client_id, title, description, content_type, deadline, budget_amount,
			   budget_currency, priority, status, requirements, metadata, created_at, updated_at
		FROM projects WHERE status NOT IN ('Completed', 'Cancelled')
		ORDER BY deadline ASC
		OFFSET $1 LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	projects, err := r.scanProjects(rows)
	return projects, total, err
}

func (r *PostgresProjectRepository) FindAll(ctx context.Context, offset, limit int) ([]*entities.Project, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresProjectRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]*entities.Project, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresProjectRepository) FindByDeadlineRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Project, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresProjectRepository) Save(ctx context.Context, project *entities.Project) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresProjectRepository) Create(ctx context.Context, project *entities.Project) error {
	if project.ProjectID == uuid.Nil {
		project.ProjectID = uuid.New()
	}
	
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	
	requirementsJSON, _ := json.Marshal(project.Requirements)
	metadataJSON, _ := json.Marshal(project.Metadata)
	
	query := `
		INSERT INTO projects (
			project_id, client_id, title, description, content_type, deadline,
			budget_amount, budget_currency, priority, status, requirements,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	_, err := r.db.ExecContext(ctx, query,
		project.ProjectID, project.ClientID, project.Title, project.Description,
		project.ContentType, project.Deadline, project.Budget.Amount, project.Budget.Currency,
		project.Priority, project.Status, requirementsJSON, metadataJSON,
		project.CreatedAt, project.UpdatedAt)
	
	return err
}

func (r *PostgresProjectRepository) Update(ctx context.Context, project *entities.Project) error {
	project.UpdatedAt = time.Now()
	
	requirementsJSON, _ := json.Marshal(project.Requirements)
	metadataJSON, _ := json.Marshal(project.Metadata)
	
	query := `
		UPDATE projects SET
			title = $2, description = $3, content_type = $4, deadline = $5,
			budget_amount = $6, budget_currency = $7, priority = $8, status = $9,
			requirements = $10, metadata = $11, updated_at = $12
		WHERE project_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		project.ProjectID, project.Title, project.Description, project.ContentType,
		project.Deadline, project.Budget.Amount, project.Budget.Currency,
		project.Priority, project.Status, requirementsJSON, metadataJSON,
		project.UpdatedAt)
	
	return err
}

func (r *PostgresProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Project, error) {
	// Delegate to FindByID for now
	return r.FindByID(ctx, id)
}

func (r *PostgresProjectRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*entities.Project, error) {
	// Delegate to FindByClientID for now
	projects, _, err := r.FindByClientID(ctx, clientID, 0, 100) // Get first 100 projects
	return projects, err
}

// PostgresContentRepository implements the ContentRepository interface
type PostgresContentRepository struct {
	db *sql.DB
}

// NewContentRepository creates a new PostgreSQL content repository
func NewContentRepository(db *sql.DB) repositories.ContentRepository {
	return &PostgresContentRepository{db: db}
}

func (r *PostgresContentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	query := `
		SELECT content_id, project_id, title, type, status, data, metadata, version,
			   word_count, created_at, updated_at
		FROM content WHERE content_id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanContent(row)
}

func (r *PostgresContentRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.Content, error) {
	query := `
		SELECT content_id, project_id, title, type, status, data, metadata, version,
			   word_count, created_at, updated_at
		FROM content WHERE project_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanContents(rows)
}

func (r *PostgresContentRepository) FindByStatus(ctx context.Context, status entities.ContentStatus, offset, limit int) ([]*entities.Content, int, error) {
	countQuery := "SELECT COUNT(*) FROM content WHERE status = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT content_id, project_id, title, type, status, data, metadata, version,
			   word_count, created_at, updated_at
		FROM content WHERE status = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`
	
	rows, err := r.db.QueryContext(ctx, query, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	contents, err := r.scanContents(rows)
	return contents, total, err
}

func (r *PostgresContentRepository) FindByType(ctx context.Context, contentType entities.ContentType, offset, limit int) ([]*entities.Content, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresContentRepository) Save(ctx context.Context, content *entities.Content) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresContentRepository) Create(ctx context.Context, content *entities.Content) error {
	if content.ContentID == uuid.Nil {
		content.ContentID = uuid.New()
	}
	
	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now
	
	metadataJSON, _ := json.Marshal(content.Metadata)
	
	query := `
		INSERT INTO content (
			content_id, project_id, title, type, status, data, metadata,
			version, word_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	_, err := r.db.ExecContext(ctx, query,
		content.ContentID, content.ProjectID, content.Title, content.Type,
		content.Status, content.Data, metadataJSON, content.Version,
		content.WordCount, content.CreatedAt, content.UpdatedAt)
	
	return err
}

func (r *PostgresContentRepository) Update(ctx context.Context, content *entities.Content) error {
	content.UpdatedAt = time.Now()
	
	metadataJSON, _ := json.Marshal(content.Metadata)
	
	query := `
		UPDATE content SET
			title = $2, type = $3, status = $4, data = $5, metadata = $6,
			version = $7, word_count = $8, updated_at = $9
		WHERE content_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		content.ContentID, content.Title, content.Type, content.Status,
		content.Data, metadataJSON, content.Version, content.WordCount,
		content.UpdatedAt)
	
	return err
}

func (r *PostgresContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Placeholder implementation
	return nil
}

// PostgresContentVersionRepository implements the ContentVersionRepository interface
type PostgresContentVersionRepository struct {
	db *sql.DB
}

// NewContentVersionRepository creates a new PostgreSQL content version repository
func NewContentVersionRepository(db *sql.DB) repositories.ContentVersionRepository {
	return &PostgresContentVersionRepository{db: db}
}

func (r *PostgresContentVersionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ContentVersion, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresContentVersionRepository) FindByContentID(ctx context.Context, contentID uuid.UUID) ([]*entities.ContentVersion, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresContentVersionRepository) FindByContentIDAndVersion(ctx context.Context, contentID uuid.UUID, version int) (*entities.ContentVersion, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresContentVersionRepository) Save(ctx context.Context, version *entities.ContentVersion) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresContentVersionRepository) Create(ctx context.Context, version *entities.ContentVersion) error {
	// Placeholder implementation
	return nil
}

// PostgresFeedbackRepository implements the FeedbackRepository interface
type PostgresFeedbackRepository struct {
	db *sql.DB
}

// NewFeedbackRepository creates a new PostgreSQL feedback repository
func NewFeedbackRepository(db *sql.DB) repositories.FeedbackRepository {
	return &PostgresFeedbackRepository{db: db}
}

func (r *PostgresFeedbackRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Feedback, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresFeedbackRepository) FindByContentID(ctx context.Context, contentID uuid.UUID) ([]*entities.Feedback, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresFeedbackRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.Feedback, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresFeedbackRepository) FindByClientID(ctx context.Context, clientID uuid.UUID) ([]*entities.Feedback, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresFeedbackRepository) FindByType(ctx context.Context, feedbackType entities.FeedbackType, offset, limit int) ([]*entities.Feedback, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresFeedbackRepository) FindBySource(ctx context.Context, source entities.FeedbackSource, offset, limit int) ([]*entities.Feedback, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresFeedbackRepository) FindByStatus(ctx context.Context, status entities.FeedbackStatus, offset, limit int) ([]*entities.Feedback, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresFeedbackRepository) Save(ctx context.Context, feedback *entities.Feedback) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresFeedbackRepository) Create(ctx context.Context, feedback *entities.Feedback) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresFeedbackRepository) Update(ctx context.Context, feedback *entities.Feedback) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresFeedbackRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Placeholder implementation
	return nil
}

// PostgresSystemCapabilityRepository implements the SystemCapabilityRepository interface
type PostgresSystemCapabilityRepository struct {
	db *sql.DB
}

// NewSystemCapabilityRepository creates a new PostgreSQL system capability repository
func NewSystemCapabilityRepository(db *sql.DB) repositories.SystemCapabilityRepository {
	return &PostgresSystemCapabilityRepository{db: db}
}

func (r *PostgresSystemCapabilityRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SystemCapability, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresSystemCapabilityRepository) FindByName(ctx context.Context, name string) (*entities.SystemCapability, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresSystemCapabilityRepository) FindByType(ctx context.Context, capabilityType entities.CapabilityType) ([]*entities.SystemCapability, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresSystemCapabilityRepository) FindByStatus(ctx context.Context, status entities.CapabilityStatus) ([]*entities.SystemCapability, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresSystemCapabilityRepository) FindAll(ctx context.Context) ([]*entities.SystemCapability, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresSystemCapabilityRepository) Save(ctx context.Context, capability *entities.SystemCapability) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresSystemCapabilityRepository) Create(ctx context.Context, capability *entities.SystemCapability) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresSystemCapabilityRepository) Update(ctx context.Context, capability *entities.SystemCapability) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresSystemCapabilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Placeholder implementation
	return nil
}

// PostgresEventRepository implements the EventRepository interface
type PostgresEventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *sql.DB) repositories.EventRepository {
	return &PostgresEventRepository{db: db}
}

func (r *PostgresEventRepository) FindByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	// Placeholder implementation
	return nil, nil
}

func (r *PostgresEventRepository) FindByAggregateID(ctx context.Context, aggregateID uuid.UUID, offset, limit int) ([]interface{}, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresEventRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID, eventType string, offset, limit int) ([]*events.Event, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresEventRepository) FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]interface{}, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresEventRepository) FindByType(ctx context.Context, eventType string, offset, limit int) ([]interface{}, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresEventRepository) FindLatest(ctx context.Context, limit int) ([]interface{}, error) {
	return nil, nil
}

func (r *PostgresEventRepository) Save(ctx context.Context, event interface{}) error {
	// Placeholder implementation
	return nil
}

// PostgresTransactionRepository implements the TransactionRepository interface
type PostgresTransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new PostgreSQL transaction repository
func NewTransactionRepository(db *sql.DB) repositories.TransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	query := `
		SELECT transaction_id, client_id, project_id, type, status, amount, currency,
			   payment_method, payment_reference, description, processed_at,
			   created_at, updated_at, metadata
		FROM transactions WHERE transaction_id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanTransaction(row)
}

func (r *PostgresTransactionRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]*entities.Transaction, int, error) {
	countQuery := "SELECT COUNT(*) FROM transactions WHERE client_id = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, clientID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT transaction_id, client_id, project_id, type, status, amount, currency,
			   payment_method, payment_reference, description, processed_at,
			   created_at, updated_at, metadata
		FROM transactions WHERE client_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`
	
	rows, err := r.db.QueryContext(ctx, query, clientID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	transactions, err := r.scanTransactions(rows)
	return transactions, total, err
}

func (r *PostgresTransactionRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*entities.Transaction, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresTransactionRepository) FindByStatus(ctx context.Context, status entities.TransactionStatus, offset, limit int) ([]*entities.Transaction, int, error) {
	countQuery := "SELECT COUNT(*) FROM transactions WHERE status = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT transaction_id, client_id, project_id, type, status, amount, currency,
			   payment_method, payment_reference, description, processed_at,
			   created_at, updated_at, metadata
		FROM transactions WHERE status = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`
	
	rows, err := r.db.QueryContext(ctx, query, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	transactions, err := r.scanTransactions(rows)
	return transactions, total, err
}

func (r *PostgresTransactionRepository) FindByType(ctx context.Context, transactionType entities.TransactionType, offset, limit int) ([]*entities.Transaction, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresTransactionRepository) FindByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Transaction, int, error) {
	// Placeholder implementation
	return nil, 0, nil
}

func (r *PostgresTransactionRepository) Save(ctx context.Context, transaction *entities.Transaction) error {
	// Placeholder implementation
	return nil
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, transaction *entities.Transaction) error {
	if transaction.TransactionID == uuid.Nil {
		transaction.TransactionID = uuid.New()
	}
	
	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now
	
	metadataJSON, _ := json.Marshal(transaction.Metadata)
	
	query := `
		INSERT INTO transactions (
			transaction_id, client_id, project_id, type, status, amount, currency,
			payment_method, payment_reference, description, processed_at,
			created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	_, err := r.db.ExecContext(ctx, query,
		transaction.TransactionID, transaction.ClientID, transaction.ProjectID,
		transaction.Type, transaction.Status, transaction.Amount.Amount, transaction.Amount.Currency,
		transaction.PaymentMethod, transaction.PaymentReference, transaction.Description,
		transaction.ProcessedAt, transaction.CreatedAt, transaction.UpdatedAt, metadataJSON)
	
	return err
}

func (r *PostgresTransactionRepository) Update(ctx context.Context, transaction *entities.Transaction) error {
	transaction.UpdatedAt = time.Now()
	
	metadataJSON, _ := json.Marshal(transaction.Metadata)
	
	query := `
		UPDATE transactions SET
			type = $2, status = $3, amount = $4, currency = $5, payment_method = $6,
			payment_reference = $7, description = $8, processed_at = $9,
			updated_at = $10, metadata = $11
		WHERE transaction_id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		transaction.TransactionID, transaction.Type, transaction.Status,
		transaction.Amount.Amount, transaction.Amount.Currency, transaction.PaymentMethod,
		transaction.PaymentReference, transaction.Description, transaction.ProcessedAt,
		transaction.UpdatedAt, metadataJSON)
	
	return err
}

// PostgresRiskRepository implements the RiskRepository interface
type PostgresRiskRepository struct {
	db *sql.DB
}

// GetIncidentByID implements repositories.RiskRepository.
func (r *PostgresRiskRepository) GetIncidentByID(ctx context.Context, id string) (*entities.Incident, error) {
	panic("unimplemented")
}

// UpdateDependencyStatus implements repositories.RiskRepository.
func (r *PostgresRiskRepository) UpdateDependencyStatus(ctx context.Context, id string, status string) error {
	panic("unimplemented")
}

// NewRiskRepository creates a new PostgreSQL risk repository
func NewRiskRepository(db *sql.DB) repositories.RiskRepository {
	return &PostgresRiskRepository{db: db}
}

func (r *PostgresRiskRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Risk, error) {
	query := `
		SELECT risk_id, title, description, category, severity, likelihood, impact,
			   status, metadata, mitigation_actions, owner_id, identified_at,
			   last_assessment, resolution_date, created_at, updated_at
		FROM risks
		WHERE risk_id = $1`

	var risk entities.Risk
	var metadata, mitigationActions sql.NullString
	var ownerID, resolutionDate sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&risk.ID, &risk.Title, &risk.Description, &risk.Category, &risk.Severity,
		&risk.Likelihood, &risk.Impact, &risk.Status, &metadata, &mitigationActions,
		&ownerID, &risk.IdentifiedAt, &risk.LastAssessment, &resolutionDate,
		&risk.CreatedAt, &risk.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse JSON metadata if present
	if metadata.Valid {
		if err := json.Unmarshal([]byte(metadata.String), &risk.Metadata); err != nil {
			// Log error but continue
			fmt.Printf("Failed to unmarshal risk metadata: %v\n", err)
		}
	}

	// Parse mitigation actions if present
	if mitigationActions.Valid {
		// Parse comma-separated string array
		actions := strings.Split(mitigationActions.String, ",")
		for i, action := range actions {
			actions[i] = strings.TrimSpace(action)
		}
		risk.MitigationActions = actions
	}

	if ownerID.Valid {
		ownerUUID, _ := uuid.Parse(ownerID.String)
		risk.OwnerID = &ownerUUID
	}

	if resolutionDate.Valid {
		if parsed, err := time.Parse(time.RFC3339, resolutionDate.String); err == nil {
			risk.ResolutionDate = &parsed
		} else {
			fmt.Printf("Failed to parse resolution date: %v\n", err)
		}
	}

	return &risk, nil
}

func (r *PostgresRiskRepository) FindByCategory(ctx context.Context, category entities.RiskCategory, offset, limit int) ([]*entities.Risk, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM risks WHERE category = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, category).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	query := `
		SELECT risk_id, title, description, category, severity, likelihood, impact,
			   status, metadata, mitigation_actions, owner_id, identified_at,
			   last_assessment, resolution_date, created_at, updated_at
		FROM risks
		WHERE category = $1
		ORDER BY severity DESC, identified_at DESC
		OFFSET $2 LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, category, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var risks []*entities.Risk
	for rows.Next() {
		var risk entities.Risk
		var metadata, mitigationActions sql.NullString
		var ownerID, resolutionDate sql.NullString

		err := rows.Scan(
			&risk.ID, &risk.Title, &risk.Description, &risk.Category, &risk.Severity,
			&risk.Likelihood, &risk.Impact, &risk.Status, &metadata, &mitigationActions,
			&ownerID, &risk.IdentifiedAt, &risk.LastAssessment, &resolutionDate,
			&risk.CreatedAt, &risk.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// TODO: Parse metadata and mitigation actions as above

		risks = append(risks, &risk)
	}

	return risks, total, nil
}

func (r *PostgresRiskRepository) FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Risk, int, error) {
	// Similar implementation to FindByCategory
	return nil, 0, nil
}

func (r *PostgresRiskRepository) FindByStatus(ctx context.Context, status entities.RiskStatus, offset, limit int) ([]*entities.Risk, int, error) {
	// Similar implementation to FindByCategory
	return nil, 0, nil
}

func (r *PostgresRiskRepository) FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Risk, int, error) {
	// Implementation for time range queries
	return nil, 0, nil
}

func (r *PostgresRiskRepository) Save(ctx context.Context, risk *entities.Risk) error {
	if risk.ID == uuid.Nil {
		return r.Create(ctx, risk)
	}
	return r.Update(ctx, risk)
}

func (r *PostgresRiskRepository) Create(ctx context.Context, risk *entities.Risk) error {
	if risk.ID == uuid.Nil {
		risk.ID = uuid.New()
	}

	query := `
		INSERT INTO risks (risk_id, title, description, category, severity, likelihood,
						  impact, status, metadata, mitigation_actions, owner_id,
						  identified_at, last_assessment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	now := time.Now()
	risk.CreatedAt = now
	risk.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		risk.ID, risk.Title, risk.Description, risk.Category, risk.Severity,
		risk.Likelihood, risk.Impact, risk.Status, nil, // TODO: Marshal metadata
		nil, // TODO: Marshal mitigation actions
		risk.OwnerID, risk.IdentifiedAt, risk.LastAssessment,
		risk.CreatedAt, risk.UpdatedAt,
	)

	return err
}

func (r *PostgresRiskRepository) Update(ctx context.Context, risk *entities.Risk) error {
	query := `
		UPDATE risks
		SET title = $2, description = $3, category = $4, severity = $5,
			likelihood = $6, impact = $7, status = $8, metadata = $9,
			mitigation_actions = $10, owner_id = $11, last_assessment = $12,
			resolution_date = $13, updated_at = $14
		WHERE risk_id = $1`

	risk.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		risk.ID, risk.Title, risk.Description, risk.Category, risk.Severity,
		risk.Likelihood, risk.Impact, risk.Status, nil, // TODO: Marshal metadata
		nil, // TODO: Marshal mitigation actions
		risk.OwnerID, risk.LastAssessment, risk.ResolutionDate,
		risk.UpdatedAt,
	)

	return err
}

func (r *PostgresRiskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM risks WHERE risk_id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresBackupRepository) AddIncidentAction(ctx context.Context, incidentID string, action *entities.IncidentAction) error {
	return nil
}

// Similar implementations for Incident, Vulnerability, and Backup repositories...

// PostgresIncidentRepository implements the IncidentRepository interface
type PostgresIncidentRepository struct {
	db *sql.DB
}

func NewIncidentRepository(db *sql.DB) repositories.IncidentRepository {
	return &PostgresIncidentRepository{db: db}
}

func (r *PostgresIncidentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Incident, error) {
	// Implementation similar to RiskRepository
	return nil, nil
}

func (r *PostgresIncidentRepository) FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Incident, int, error) {
	return nil, 0, nil
}

func (r *PostgresIncidentRepository) FindByStatus(ctx context.Context, status entities.IncidentStatus, offset, limit int) ([]*entities.Incident, int, error) {
	return nil, 0, nil
}

func (r *PostgresIncidentRepository) FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Incident, int, error) {
	return nil, 0, nil
}

func (r *PostgresIncidentRepository) Save(ctx context.Context, incident *entities.Incident) error {
	return nil
}

func (r *PostgresIncidentRepository) Create(ctx context.Context, incident *entities.Incident) error {
	return nil
}

func (r *PostgresIncidentRepository) Update(ctx context.Context, incident *entities.Incident) error {
	return nil
}

func (r *PostgresIncidentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// PostgresVulnerabilityRepository implements the VulnerabilityRepository interface
type PostgresVulnerabilityRepository struct {
	db *sql.DB
}

func NewVulnerabilityRepository(db *sql.DB) repositories.VulnerabilityRepository {
	return &PostgresVulnerabilityRepository{db: db}
}

func (r *PostgresVulnerabilityRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Vulnerability, error) {
	return nil, nil
}

func (r *PostgresVulnerabilityRepository) FindBySeverity(ctx context.Context, severity entities.RiskSeverity, offset, limit int) ([]*entities.Vulnerability, int, error) {
	return nil, 0, nil
}

func (r *PostgresVulnerabilityRepository) FindByStatus(ctx context.Context, status entities.VulnerabilityStatus, offset, limit int) ([]*entities.Vulnerability, int, error) {
	return nil, 0, nil
}

func (r *PostgresVulnerabilityRepository) FindByComponent(ctx context.Context, component string, offset, limit int) ([]*entities.Vulnerability, int, error) {
	return nil, 0, nil
}

func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vulnerability *entities.Vulnerability) error {
	return nil
}

func (r *PostgresVulnerabilityRepository) Create(ctx context.Context, vulnerability *entities.Vulnerability) error {
	return nil
}

func (r *PostgresVulnerabilityRepository) Update(ctx context.Context, vulnerability *entities.Vulnerability) error {
	return nil
}

func (r *PostgresVulnerabilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// PostgresBackupRepository implements the BackupRepository interface
type PostgresBackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) repositories.BackupRepository {
	return &PostgresBackupRepository{db: db}
}

func (r *PostgresBackupRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Backup, error) {
	return nil, nil
}

func (r *PostgresBackupRepository) FindByStatus(ctx context.Context, status entities.BackupStatus, offset, limit int) ([]*entities.Backup, int, error) {
	return nil, 0, nil
}

func (r *PostgresBackupRepository) FindByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.Backup, int, error) {
	return nil, 0, nil
}

func (r *PostgresBackupRepository) FindExpired(ctx context.Context) ([]*entities.Backup, error) {
	return nil, nil
}

func (r *PostgresBackupRepository) Save(ctx context.Context, backup *entities.Backup) error {
	return nil
}

func (r *PostgresBackupRepository) Create(ctx context.Context, backup *entities.Backup) error {
	return nil
}

func (r *PostgresBackupRepository) Update(ctx context.Context, backup *entities.Backup) error {
	return nil
}

func (r *PostgresBackupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// Add missing methods to PostgresRiskRepository

func (r *PostgresRiskRepository) CreateRisk(ctx context.Context, risk *entities.Risk) error {
	return r.Create(ctx, risk)
}

func (r *PostgresRiskRepository) GetRiskByID(ctx context.Context, id string) (*entities.Risk, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, uid)
}

func (r *PostgresRiskRepository) ListRisks(ctx context.Context, filters repositories.RiskFilters) ([]*entities.Risk, error) {
	// TODO: Implement filtering
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateRisk(ctx context.Context, risk *entities.Risk) error {
	return r.Update(ctx, risk)
}

func (r *PostgresRiskRepository) DeleteRisk(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.Delete(ctx, uid)
}

func (r *PostgresRiskRepository) GetRisksByType(ctx context.Context, riskType entities.RiskType) ([]*entities.Risk, error) {
	risks, _, err := r.FindByCategory(ctx, riskType, 0, 100)
	if err != nil {
		return nil, err
	}
	return risks, err
}

func (r *PostgresRiskRepository) GetActiveRisks(ctx context.Context) ([]*entities.Risk, error) {
	// Return all non-resolved risks
	return nil, nil
}

func (r *PostgresRiskRepository) GetRiskMetrics(ctx context.Context, timeRange repositories.TimeRange) (*repositories.RiskMetrics, error) {
	// TODO: Implement metrics calculation
	return nil, nil
}

func (r *PostgresRiskRepository) CreateThreshold(ctx context.Context, threshold *entities.RiskThreshold) error {
	return nil
}

func (r *PostgresRiskRepository) GetThresholdByID(ctx context.Context, id string) (*entities.RiskThreshold, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListThresholds(ctx context.Context) ([]*entities.RiskThreshold, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateThreshold(ctx context.Context, threshold *entities.RiskThreshold) error {
	return nil
}

func (r *PostgresRiskRepository) DeleteThreshold(ctx context.Context, id string) error {
	return nil
}

func (r *PostgresRiskRepository) GetThresholdsByType(ctx context.Context, riskType entities.RiskType) ([]*entities.RiskThreshold, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CheckThresholdViolations(ctx context.Context, riskType entities.RiskType, value float64) ([]*entities.RiskThreshold, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CreateComplianceRequirement(ctx context.Context, req *entities.ComplianceRequirement) error {
	return nil
}

func (r *PostgresRiskRepository) GetComplianceRequirement(ctx context.Context, id string) (*entities.ComplianceRequirement, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListComplianceRequirements(ctx context.Context, regulation string) ([]*entities.ComplianceRequirement, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateComplianceRequirement(ctx context.Context, req *entities.ComplianceRequirement) error {
	return nil
}

func (r *PostgresRiskRepository) GetActiveComplianceRequirements(ctx context.Context) ([]*entities.ComplianceRequirement, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CreateContentPolicy(ctx context.Context, policy *entities.ContentPolicy) error {
	return nil
}

func (r *PostgresRiskRepository) GetContentPolicy(ctx context.Context, id string) (*entities.ContentPolicy, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListContentPolicies(ctx context.Context) ([]*entities.ContentPolicy, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateContentPolicy(ctx context.Context, policy *entities.ContentPolicy) error {
	return nil
}

func (r *PostgresRiskRepository) GetActiveContentPolicies(ctx context.Context) ([]*entities.ContentPolicy, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) GetContentPoliciesByType(ctx context.Context, policyType string) ([]*entities.ContentPolicy, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CreateServiceDependency(ctx context.Context, dep *entities.ServiceDependency) error {
	return nil
}

func (r *PostgresRiskRepository) GetServiceDependency(ctx context.Context, id string) (*entities.ServiceDependency, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListServiceDependencies(ctx context.Context) ([]*entities.ServiceDependency, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateServiceDependency(ctx context.Context, dep *entities.ServiceDependency) error {
	return nil
}

func (r *PostgresRiskRepository) DeleteServiceDependency(ctx context.Context, id string) error {
	return nil
}

func (r *PostgresRiskRepository) GetCriticalDependencies(ctx context.Context) ([]*entities.ServiceDependency, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) GetUnhealthyDependencies(ctx context.Context) ([]*entities.ServiceDependency, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CreateIncident(ctx context.Context, incident *entities.Incident) error {
	return nil
}

func (r *PostgresRiskRepository) GetIncident(ctx context.Context, id string) (*entities.Incident, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListIncidents(ctx context.Context, filters repositories.IncidentFilters) ([]*entities.Incident, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateIncident(ctx context.Context, incident *entities.Incident) error {
	return nil
}

func (r *PostgresRiskRepository) GetActiveIncidents(ctx context.Context) ([]*entities.Incident, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) GetIncidentsByService(ctx context.Context, service string) ([]*entities.Incident, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) AddIncidentAction(ctx context.Context, incidentID string, action *entities.IncidentAction) error {
	return nil
}

func (r *PostgresRiskRepository) CreateVulnerability(ctx context.Context, vuln *entities.Vulnerability) error {
	return nil
}

func (r *PostgresRiskRepository) GetVulnerability(ctx context.Context, id string) (*entities.Vulnerability, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListVulnerabilities(ctx context.Context, filters repositories.VulnerabilityFilters) ([]*entities.Vulnerability, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateVulnerability(ctx context.Context, vuln *entities.Vulnerability) error {
	return nil
}

func (r *PostgresRiskRepository) GetUnpatchedVulnerabilities(ctx context.Context) ([]*entities.Vulnerability, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) MarkVulnerabilityPatched(ctx context.Context, id string) error {
	return nil
}

func (r *PostgresRiskRepository) CreateBackupRecord(ctx context.Context, backup *entities.Backup) error {
	return nil
}

func (r *PostgresRiskRepository) GetBackupRecord(ctx context.Context, id string) (*entities.Backup, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) ListBackupRecords(ctx context.Context, filters repositories.BackupFilters) ([]*entities.Backup, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) UpdateBackupRecord(ctx context.Context, backup *entities.Backup) error {
	return nil
}

func (r *PostgresRiskRepository) GetLastSuccessfulBackup(ctx context.Context, backupType string) (*entities.Backup, error) {
	return nil, nil
}

func (r *PostgresRiskRepository) CleanupOldBackups(ctx context.Context, retentionDays int) error {
	return nil
}

// Helper methods for scanning database rows

// scanClient scans a database row into a Client entity
func (r *PostgresClientRepository) scanClient(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Client, error) {
	var client entities.Client
	var addressJSON []byte

	err := scanner.Scan(
		&client.ClientID, &client.Name, &client.ContactEmail, &client.ContactPhone,
		&addressJSON, &client.Timezone, &client.CreatedAt, &client.UpdatedAt, &client.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal address JSON
	if err := json.Unmarshal(addressJSON, &client.BillingAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing address: %w", err)
	}

	return &client, nil
}

// Scanning helper methods for Project, Content, and Transaction entities

func (r *PostgresProjectRepository) scanProject(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Project, error) {
	var project entities.Project
	var requirementsJSON, metadataJSON []byte
	
	err := scanner.Scan(
		&project.ProjectID, &project.ClientID, &project.Title, &project.Description,
		&project.ContentType, &project.Deadline, &project.Budget.Amount, &project.Budget.Currency,
		&project.Priority, &project.Status, &requirementsJSON, &metadataJSON,
		&project.CreatedAt, &project.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(requirementsJSON, &project.Requirements); err != nil {
		project.Requirements = []string{}
	}
	if err := json.Unmarshal(metadataJSON, &project.Metadata); err != nil {
		project.Metadata = make(map[string]interface{})
	}
	
	return &project, nil
}

func (r *PostgresProjectRepository) scanProjects(rows *sql.Rows) ([]*entities.Project, error) {
	var projects []*entities.Project
	
	for rows.Next() {
		project, err := r.scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	
	return projects, rows.Err()
}

func (r *PostgresContentRepository) scanContent(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Content, error) {
	var content entities.Content
	var metadataJSON []byte
	
	err := scanner.Scan(
		&content.ContentID, &content.ProjectID, &content.Title, &content.Type,
		&content.Status, &content.Data, &metadataJSON, &content.Version,
		&content.WordCount, &content.CreatedAt, &content.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(metadataJSON, &content.Metadata); err != nil {
		content.Metadata = make(map[string]interface{})
	}
	
	return &content, nil
}

func (r *PostgresContentRepository) scanContents(rows *sql.Rows) ([]*entities.Content, error) {
	var contents []*entities.Content
	
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	
	return contents, rows.Err()
}

func (r *PostgresTransactionRepository) scanTransaction(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Transaction, error) {
	var transaction entities.Transaction
	var metadataJSON []byte
	var projectID sql.NullString
	var processedAt sql.NullTime
	
	err := scanner.Scan(
		&transaction.TransactionID, &transaction.ClientID, &projectID, &transaction.Type,
		&transaction.Status, &transaction.Amount.Amount, &transaction.Amount.Currency,
		&transaction.PaymentMethod, &transaction.PaymentReference, &transaction.Description,
		&processedAt, &transaction.CreatedAt, &transaction.UpdatedAt, &metadataJSON)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Handle nullable fields
	if projectID.Valid {
		if id, err := uuid.Parse(projectID.String); err == nil {
			transaction.ProjectID = &id
		}
	}
	if processedAt.Valid {
		transaction.ProcessedAt = &processedAt.Time
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(metadataJSON, &transaction.Metadata); err != nil {
		transaction.Metadata = make(map[string]interface{})
	}
	
	return &transaction, nil
}

func (r *PostgresTransactionRepository) scanTransactions(rows *sql.Rows) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	
	for rows.Next() {
		transaction, err := r.scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	
	return transactions, rows.Err()
}


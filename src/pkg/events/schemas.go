package events

import (
	"time"
)

// Event is imported from redis_streams.go - no need to redeclare

// Event type constants
const (
	// Content Events
	EventContentCreated   = "content.created"
	EventContentUpdated   = "content.updated"
	EventContentApproved  = "content.approved"
	EventContentRejected  = "content.rejected"
	EventContentPublished = "content.published"
	EventContentArchived  = "content.archived"

	// Project Events
	EventProjectCreated    = "project.created"
	EventProjectUpdated    = "project.updated"
	EventProjectCompleted  = "project.completed"
	EventProjectCancelled  = "project.cancelled"
	EventProjectStatusChanged = "project.status_changed"

	// Client Events
	EventClientOnboarded     = "client.onboarded"
	EventClientUpdated       = "client.updated"
	EventClientStatusChanged = "client.status_changed"
	EventClientFeedback      = "client.feedback"

	// Financial Events
	EventInvoiceCreated  = "invoice.created"
	EventInvoiceUpdated  = "invoice.updated"
	EventInvoicePaid     = "invoice.paid"
	EventInvoiceFailed   = "invoice.failed"
	EventPaymentReceived = "payment.received"
	EventPaymentFailed   = "payment.failed"

	// Decision Events
	EventDecisionCreated  = "decision.created"
	EventDecisionExecuted = "decision.executed"
	EventDecisionOverride = "decision.override"

	// Risk Events
	EventRiskDetected    = "risk.detected"
	EventRiskMitigated   = "risk.mitigated"
	EventIncidentCreated = "incident.created"
	EventIncidentResolved = "incident.resolved"

	// HR Events
	EventTalentOnboarded   = "talent.onboarded"
	EventTalentOffboarded  = "talent.offboarded"
	EventEngagementCreated = "engagement.created"
	EventEngagementCompleted = "engagement.completed"

	// Governance Events
	EventProposalCreated   = "proposal.created"
	EventProposalVoted     = "proposal.voted"
	EventProposalExecuted  = "proposal.executed"
	EventMemberJoined      = "member.joined"

	// Legal Events
	EventContractCreated  = "contract.created"
	EventContractSigned   = "contract.signed"
	EventContractExpired  = "contract.expired"
	EventComplianceIssue  = "compliance.issue"

	// System Events
	EventSystemHealthCheck = "system.health_check"
	EventSystemAlert       = "system.alert"
	EventSystemMaintenance = "system.maintenance"
)

// Stream names for different event categories
const (
	StreamContent    = "content-events"
	StreamProjects   = "project-events"
	StreamClients    = "client-events"
	StreamFinancial  = "financial-events"
	StreamDecisions  = "decision-events"
	StreamRisk       = "risk-events"
	StreamHR         = "hr-events"
	StreamGovernance = "governance-events"
	StreamLegal      = "legal-events"
	StreamSystem     = "system-events"
)

// Event data schemas for different event types

// ContentEventData represents content-related event data
type ContentEventData struct {
	ContentID    string                 `json:"content_id"`
	ProjectID    string                 `json:"project_id"`
	ClientID     string                 `json:"client_id"`
	ContentType  string                 `json:"content_type"`
	Title        string                 `json:"title"`
	Status       string                 `json:"status"`
	WordCount    int                    `json:"word_count,omitempty"`
	QualityScore float64                `json:"quality_score,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectEventData represents project-related event data
type ProjectEventData struct {
	ProjectID     string                 `json:"project_id"`
	ClientID      string                 `json:"client_id"`
	Title         string                 `json:"title"`
	Status        string                 `json:"status"`
	Priority      string                 `json:"priority"`
	Budget        float64                `json:"budget,omitempty"`
	Deadline      *time.Time             `json:"deadline,omitempty"`
	Progress      float64                `json:"progress,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ClientEventData represents client-related event data
type ClientEventData struct {
	ClientID      string                 `json:"client_id"`
	Name          string                 `json:"name"`
	Email         string                 `json:"email"`
	Status        string                 `json:"status"`
	Tier          string                 `json:"tier"`
	Industry      string                 `json:"industry,omitempty"`
	Satisfaction  float64                `json:"satisfaction,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// FinancialEventData represents financial-related event data
type FinancialEventData struct {
	TransactionID string                 `json:"transaction_id"`
	InvoiceID     string                 `json:"invoice_id,omitempty"`
	ClientID      string                 `json:"client_id,omitempty"`
	ProjectID     string                 `json:"project_id,omitempty"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	PaymentMethod string                 `json:"payment_method,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DecisionEventData represents decision-related event data
type DecisionEventData struct {
	DecisionID       string                 `json:"decision_id"`
	Type             string                 `json:"type"`
	Priority         string                 `json:"priority"`
	Status           string                 `json:"status"`
	SelectedOption   string                 `json:"selected_option,omitempty"`
	ConfidenceScore  float64                `json:"confidence_score,omitempty"`
	ImpactScore      float64                `json:"impact_score,omitempty"`
	Justification    string                 `json:"justification,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RiskEventData represents risk-related event data
type RiskEventData struct {
	RiskID       string                 `json:"risk_id"`
	IncidentID   string                 `json:"incident_id,omitempty"`
	Category     string                 `json:"category"`
	Severity     string                 `json:"severity"`
	Score        float64                `json:"score"`
	Status       string                 `json:"status"`
	Description  string                 `json:"description"`
	Mitigation   string                 `json:"mitigation,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// HREventData represents HR-related event data
type HREventData struct {
	TalentID      string                 `json:"talent_id"`
	EngagementID  string                 `json:"engagement_id,omitempty"`
	TalentType    string                 `json:"talent_type"`
	Status        string                 `json:"status"`
	Role          string                 `json:"role,omitempty"`
	Skills        []string               `json:"skills,omitempty"`
	Performance   float64                `json:"performance,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GovernanceEventData represents governance-related event data
type GovernanceEventData struct {
	ProposalID    string                 `json:"proposal_id,omitempty"`
	MemberID      string                 `json:"member_id,omitempty"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	VotingPower   float64                `json:"voting_power,omitempty"`
	VoteChoice    string                 `json:"vote_choice,omitempty"`
	QuorumMet     bool                   `json:"quorum_met,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// LegalEventData represents legal-related event data
type LegalEventData struct {
	ContractID    string                 `json:"contract_id,omitempty"`
	ClientID      string                 `json:"client_id,omitempty"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	SignatureType string                 `json:"signature_type,omitempty"`
	ExpiryDate    *time.Time             `json:"expiry_date,omitempty"`
	ComplianceIssue string               `json:"compliance_issue,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SystemEventData represents system-related event data
type SystemEventData struct {
	Service       string                 `json:"service"`
	Component     string                 `json:"component,omitempty"`
	HealthStatus  string                 `json:"health_status,omitempty"`
	AlertLevel    string                 `json:"alert_level,omitempty"`
	Message       string                 `json:"message"`
	Metrics       map[string]float64     `json:"metrics,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Helper functions to create typed events

// CreateContentEvent creates a content-related event
func CreateContentEvent(eventType, source string, data ContentEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"content_id":    data.ContentID,
		"project_id":    data.ProjectID,
		"client_id":     data.ClientID,
		"content_type":  data.ContentType,
		"title":         data.Title,
		"status":        data.Status,
		"word_count":    data.WordCount,
		"quality_score": data.QualityScore,
		"metadata":      data.Metadata,
	})
}

// CreateProjectEvent creates a project-related event
func CreateProjectEvent(eventType, source string, data ProjectEventData) Event {
	eventData := map[string]interface{}{
		"project_id": data.ProjectID,
		"client_id":  data.ClientID,
		"title":      data.Title,
		"status":     data.Status,
		"priority":   data.Priority,
		"budget":     data.Budget,
		"progress":   data.Progress,
		"metadata":   data.Metadata,
	}
	if data.Deadline != nil {
		eventData["deadline"] = data.Deadline.Format(time.RFC3339)
	}
	return CreateEvent(eventType, source, eventData)
}

// CreateClientEvent creates a client-related event
func CreateClientEvent(eventType, source string, data ClientEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"client_id":    data.ClientID,
		"name":         data.Name,
		"email":        data.Email,
		"status":       data.Status,
		"tier":         data.Tier,
		"industry":     data.Industry,
		"satisfaction": data.Satisfaction,
		"metadata":     data.Metadata,
	})
}

// CreateFinancialEvent creates a financial-related event
func CreateFinancialEvent(eventType, source string, data FinancialEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"transaction_id": data.TransactionID,
		"invoice_id":     data.InvoiceID,
		"client_id":      data.ClientID,
		"project_id":     data.ProjectID,
		"amount":         data.Amount,
		"currency":       data.Currency,
		"status":         data.Status,
		"payment_method": data.PaymentMethod,
		"metadata":       data.Metadata,
	})
}

// CreateDecisionEvent creates a decision-related event
func CreateDecisionEvent(eventType, source string, data DecisionEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"decision_id":      data.DecisionID,
		"type":             data.Type,
		"priority":         data.Priority,
		"status":           data.Status,
		"selected_option":  data.SelectedOption,
		"confidence_score": data.ConfidenceScore,
		"impact_score":     data.ImpactScore,
		"justification":    data.Justification,
		"metadata":         data.Metadata,
	})
}

// CreateRiskEvent creates a risk-related event
func CreateRiskEvent(eventType, source string, data RiskEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"risk_id":     data.RiskID,
		"incident_id": data.IncidentID,
		"category":    data.Category,
		"severity":    data.Severity,
		"score":       data.Score,
		"status":      data.Status,
		"description": data.Description,
		"mitigation":  data.Mitigation,
		"metadata":    data.Metadata,
	})
}

// CreateHREvent creates an HR-related event
func CreateHREvent(eventType, source string, data HREventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"talent_id":     data.TalentID,
		"engagement_id": data.EngagementID,
		"talent_type":   data.TalentType,
		"status":        data.Status,
		"role":          data.Role,
		"skills":        data.Skills,
		"performance":   data.Performance,
		"metadata":      data.Metadata,
	})
}

// CreateGovernanceEvent creates a governance-related event
func CreateGovernanceEvent(eventType, source string, data GovernanceEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"proposal_id":  data.ProposalID,
		"member_id":    data.MemberID,
		"type":         data.Type,
		"status":       data.Status,
		"voting_power": data.VotingPower,
		"vote_choice":  data.VoteChoice,
		"quorum_met":   data.QuorumMet,
		"metadata":     data.Metadata,
	})
}

// CreateLegalEvent creates a legal-related event
func CreateLegalEvent(eventType, source string, data LegalEventData) Event {
	eventData := map[string]interface{}{
		"contract_id":      data.ContractID,
		"client_id":        data.ClientID,
		"type":             data.Type,
		"status":           data.Status,
		"signature_type":   data.SignatureType,
		"compliance_issue": data.ComplianceIssue,
		"metadata":         data.Metadata,
	}
	if data.ExpiryDate != nil {
		eventData["expiry_date"] = data.ExpiryDate.Format(time.RFC3339)
	}
	return CreateEvent(eventType, source, eventData)
}

// CreateSystemEvent creates a system-related event
func CreateSystemEvent(eventType, source string, data SystemEventData) Event {
	return CreateEvent(eventType, source, map[string]interface{}{
		"service":       data.Service,
		"component":     data.Component,
		"health_status": data.HealthStatus,
		"alert_level":   data.AlertLevel,
		"message":       data.Message,
		"metrics":       data.Metrics,
		"metadata":      data.Metadata,
	})
}
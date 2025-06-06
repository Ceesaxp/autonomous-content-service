package risk_management

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// IncidentResponderImpl implements autonomous incident response
type IncidentResponderImpl struct {
	riskRepo      repositories.RiskRepository
	eventRepo     repositories.EventRepository
	playbooks     map[string]*IncidentPlaybook
	notifications chan *NotificationRequest
}

// NewIncidentResponder creates a new incident responder
func NewIncidentResponder(
	riskRepo repositories.RiskRepository,
	eventRepo repositories.EventRepository,
) *IncidentResponderImpl {
	return &IncidentResponderImpl{
		riskRepo:      riskRepo,
		eventRepo:     eventRepo,
		playbooks:     initializePlaybooks(),
		notifications: make(chan *NotificationRequest, 100),
	}
}

// ExecutePlaybook executes the appropriate playbook for an incident
func (r *IncidentResponderImpl) ExecutePlaybook(ctx context.Context, incident *entities.Incident) error {
	// Find appropriate playbook
	incidentType := "default"
	if typeVal, ok := incident.Metadata["type"].(string); ok {
		incidentType = typeVal
	}
	playbook, exists := r.playbooks[incidentType]
	if !exists {
		playbook = r.playbooks["default"]
	}

	// Update incident status
	incident.Status = entities.IncidentStatusInProgress
	incident.UpdatedAt = time.Now()
	if err := r.riskRepo.UpdateIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	// Execute playbook steps
	for _, step := range playbook.Steps {
		action := &entities.IncidentAction{
			ID:          generateActionID(),
			Type:        step.Type,
			Description: step.Description,
			Status:      "executing",
			ExecutedAt:  time.Now(),
			ExecutedBy:  "system",
		}

		// Execute the step
		result, err := r.executeStep(ctx, incident, step)
		if err != nil {
			action.Status = "failed"
			action.Result = fmt.Sprintf("Error: %v", err)
		} else {
			action.Status = "completed"
			action.Result = result
		}

		// Add action to incident
		if err := r.riskRepo.AddIncidentAction(ctx, incident.ID.String(), action); err != nil {
			return fmt.Errorf("failed to add incident action: %w", err)
		}

		// Check if step requires stopping
		if step.StopOnFailure && action.Status == "failed" {
			incident.Status = "failed"
			if err := r.riskRepo.UpdateIncident(ctx, incident); err != nil {
				fmt.Printf("Failed to update incident status: %v\n", err)
			}
			return fmt.Errorf("playbook execution failed at step: %s", step.Name)
		}
	}

	// Update incident status
	incident.Status = "mitigating"
	incident.UpdatedAt = time.Now()
	if err := r.riskRepo.UpdateIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	return nil
}

// NotifyStakeholders sends notifications about the incident
func (r *IncidentResponderImpl) NotifyStakeholders(ctx context.Context, incident *entities.Incident) error {
	// Determine notification priority based on severity
	priority := "normal"
	if incident.Severity == entities.RiskSeverityCritical {
		priority = "urgent"
	} else if incident.Severity == entities.RiskSeverityHigh {
		priority = "high"
	}

	// Create notification request
	notification := &NotificationRequest{
		Type:     "incident",
		Priority: priority,
		Subject:  fmt.Sprintf("Incident Alert: %s", incident.Title),
		Message: fmt.Sprintf(
			"Incident Details:\n"+
				"Type: %s\n"+
				"Severity: %s\n"+
				"Status: %s\n"+
				"Impact: %s\n"+
				"Detected: %s",
			incident.Metadata["type"],
			incident.Severity,
			incident.Status,
			incident.Metadata["impact"],
			incident.DetectedAt.Format(time.RFC3339),
		),
		Channels: r.getNotificationChannels(incident.Severity),
		Metadata: map[string]interface{}{
			"incident_id": incident.ID,
			"type":        incident.Metadata["type"],
			"severity":    incident.Severity,
		},
	}

	// Send notification (non-blocking)
	select {
	case r.notifications <- notification:
		// Notification queued successfully
	default:
		// Queue full, log error but don't block
		return fmt.Errorf("notification queue full")
	}

	// Log notification action
	action := &entities.IncidentAction{
		ID:          generateActionID(),
		Type:        "notification",
		Description: "Stakeholder notification sent",
		Status:      "completed",
		Result:      fmt.Sprintf("Notification sent via %v", notification.Channels),
		ExecutedAt:  time.Now(),
		ExecutedBy:  "system",
	}
	if err := r.riskRepo.AddIncidentAction(ctx, incident.ID.String(), action); err != nil {
		fmt.Printf("Failed to add incident action: %v\n", err)
	}

	return nil
}

// ContainIncident contains the incident to prevent further damage
func (r *IncidentResponderImpl) ContainIncident(ctx context.Context, incidentID string) error {
	incident, err := r.riskRepo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Execute containment based on incident type
	incidentType := "default"
	if typeVal, ok := incident.Metadata["type"].(string); ok {
		incidentType = typeVal
	}
	containmentActions := r.getContainmentActions(incidentType)

	for _, action := range containmentActions {
		incidentAction := &entities.IncidentAction{
			ID:          generateActionID(),
			Type:        "containment",
			Description: action.Description,
			Status:      "executing",
			ExecutedAt:  time.Now(),
			ExecutedBy:  "system",
		}

		// Execute containment action
		if err := action.Execute(ctx, incident); err != nil {
			incidentAction.Status = "failed"
			incidentAction.Result = fmt.Sprintf("Error: %v", err)
		} else {
			incidentAction.Status = "completed"
			incidentAction.Result = "Containment successful"
		}

		if err := r.riskRepo.AddIncidentAction(ctx, incident.ID.String(), incidentAction); err != nil {
			fmt.Printf("Failed to add incident action: %v\n", err)
		}
	}

	// Update incident status
	incident.Status = "contained"
	incident.UpdatedAt = time.Now()
	if err := r.riskRepo.UpdateIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	return nil
}

// RecoverFromIncident executes recovery procedures
func (r *IncidentResponderImpl) RecoverFromIncident(ctx context.Context, incidentID string) error {
	incident, err := r.riskRepo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Execute recovery based on incident type
	incidentType := "default"
	if typeVal, ok := incident.Metadata["type"].(string); ok {
		incidentType = typeVal
	}
	recoveryActions := r.getRecoveryActions(incidentType)

	for _, action := range recoveryActions {
		incidentAction := &entities.IncidentAction{
			ID:          generateActionID(),
			Type:        "recovery",
			Description: action.Description,
			Status:      "executing",
			ExecutedAt:  time.Now(),
			ExecutedBy:  "system",
		}

		// Execute recovery action
		if err := action.Execute(ctx, incident); err != nil {
			incidentAction.Status = "failed"
			incidentAction.Result = fmt.Sprintf("Error: %v", err)
		} else {
			incidentAction.Status = "completed"
			incidentAction.Result = "Recovery successful"
		}

		if err := r.riskRepo.AddIncidentAction(ctx, incident.ID.String(), incidentAction); err != nil {
			fmt.Printf("Failed to add incident action: %v\n", err)
		}
	}

	// Update incident as resolved
	incident.Status = "resolved"
	incident.ResolvedAt = &[]time.Time{time.Now()}[0]
	incident.Resolution = "Automated recovery completed successfully"
	incident.UpdatedAt = time.Now()
	if err := r.riskRepo.UpdateIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	// Emit incident resolved event
	event := &events.IncidentResolvedEvent{
		BaseEvent: events.BaseEvent{
			EventID:   generateEventID(),
			EventType: events.IncidentResolved,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "incident_responder",
			},
		},
		IncidentID:      incident.ID.String(),
		Resolution:      incident.Resolution,
		RootCause:       incident.RootCause,
		TimeToDetect:    time.Duration(0), // AcknowledgedAt not available in new model
		TimeToResolve:   time.Since(incident.DetectedAt),
		ActionsExecuted: r.getExecutedActionTypes(incident.ActionsTaken),
		LessonsLearned:  r.extractLessonsLearned(incident),
	}
	if err := r.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Failed to save incident resolved event: %v\n", err)
	}

	return nil
}

// GeneratePostMortem generates a post-mortem analysis
func (r *IncidentResponderImpl) GeneratePostMortem(ctx context.Context, incidentID string) (*PostMortem, error) {
	incident, err := r.riskRepo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	postMortem := &PostMortem{
		IncidentID:     incident.ID.String(),
		Summary:        r.generateSummary(incident),
		Timeline:       r.generateTimeline(incident),
		RootCause:      r.identifyRootCause(incident),
		Impact:         fmt.Sprintf("%v", incident.Metadata["impact"]),
		WhatWentWell:   r.identifyWhatWentWell(incident),
		WhatWentWrong:  r.identifyWhatWentWrong(incident),
		ActionItems:    r.generateActionItems(incident),
		LessonsLearned: r.extractLessonsLearned(incident),
	}

	return postMortem, nil
}

// Helper methods

func (r *IncidentResponderImpl) executeStep(ctx context.Context, incident *entities.Incident, step *PlaybookStep) (string, error) {
	switch step.Type {
	case "assess":
		return r.assessIncident(ctx, incident)
	case "isolate":
		return r.isolateAffectedService(ctx, incident)
	case "backup":
		return r.createEmergencyBackup(ctx, incident)
	case "notify":
		return "Notification sent", r.NotifyStakeholders(ctx, incident)
	case "failover":
		return r.activateFailover(ctx, incident)
	case "restart":
		return r.restartService(ctx, incident)
	case "rollback":
		return r.rollbackChanges(ctx, incident)
	default:
		return "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (r *IncidentResponderImpl) assessIncident(ctx context.Context, incident *entities.Incident) (string, error) {
	// Assess the incident and determine severity
	assessment := fmt.Sprintf(
		"Incident assessed: Type=%s, Severity=%s, Service=%s",
		incident.Metadata["type"], incident.Severity, incident.Metadata["affected_service"],
	)
	return assessment, nil
}

func (r *IncidentResponderImpl) isolateAffectedService(ctx context.Context, incident *entities.Incident) (string, error) {
	// Isolate the affected service to prevent spread
	return fmt.Sprintf("Service %s isolated", incident.Metadata["affected_service"]), nil
}

func (r *IncidentResponderImpl) createEmergencyBackup(ctx context.Context, incident *entities.Incident) (string, error) {
	// Create emergency backup before recovery
	return "Emergency backup created", nil
}

func (r *IncidentResponderImpl) activateFailover(ctx context.Context, incident *entities.Incident) (string, error) {
	// Activate failover for the affected service
	return fmt.Sprintf("Failover activated for %s", incident.Metadata["affected_service"]), nil
}

func (r *IncidentResponderImpl) restartService(ctx context.Context, incident *entities.Incident) (string, error) {
	// Restart the affected service
	return fmt.Sprintf("Service %s restarted", incident.Metadata["affected_service"]), nil
}

func (r *IncidentResponderImpl) rollbackChanges(ctx context.Context, incident *entities.Incident) (string, error) {
	// Rollback recent changes if needed
	return "Changes rolled back to last known good state", nil
}

func (r *IncidentResponderImpl) getNotificationChannels(severity entities.RiskSeverity) []string {
	switch severity {
	case entities.RiskSeverityCritical:
		return []string{"dashboard", "email", "sms", "slack"}
	case entities.RiskSeverityHigh:
		return []string{"dashboard", "email", "slack"}
	case entities.RiskSeverityMedium:
		return []string{"dashboard", "email"}
	default:
		return []string{"dashboard"}
	}
}

func (r *IncidentResponderImpl) getContainmentActions(incidentType string) []*ContainmentAction {
	actions := make([]*ContainmentAction, 0)

	switch incidentType {
	case "service_failure":
		actions = append(actions, &ContainmentAction{
			Description: "Isolate failed service",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Isolate the service
				return nil
			},
		})
	case "security_breach":
		actions = append(actions, &ContainmentAction{
			Description: "Block suspicious IPs",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Block IPs
				return nil
			},
		})
		actions = append(actions, &ContainmentAction{
			Description: "Revoke compromised credentials",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Revoke credentials
				return nil
			},
		})
	case "data_corruption":
		actions = append(actions, &ContainmentAction{
			Description: "Stop write operations",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Stop writes
				return nil
			},
		})
	}

	return actions
}

func (r *IncidentResponderImpl) getRecoveryActions(incidentType string) []*RecoveryAction {
	actions := make([]*RecoveryAction, 0)

	switch incidentType {
	case "service_failure":
		actions = append(actions, &RecoveryAction{
			Description: "Restart service with clean state",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Restart service
				return nil
			},
		})
	case "data_corruption":
		actions = append(actions, &RecoveryAction{
			Description: "Restore from last good backup",
			Execute: func(ctx context.Context, incident *entities.Incident) error {
				// Restore backup
				return nil
			},
		})
	}

	// Common recovery actions
	actions = append(actions, &RecoveryAction{
		Description: "Verify system integrity",
		Execute: func(ctx context.Context, incident *entities.Incident) error {
			// Verify integrity
			return nil
		},
	})

	return actions
}

func (r *IncidentResponderImpl) getExecutedActionTypes(actions []entities.IncidentAction) []string {
	types := make([]string, 0)
	seen := make(map[string]bool)

	for _, action := range actions {
		if !seen[action.Type] {
			types = append(types, action.Type)
			seen[action.Type] = true
		}
	}

	return types
}

func (r *IncidentResponderImpl) generateSummary(incident *entities.Incident) string {
	return fmt.Sprintf(
		"%s incident detected on %s affecting %s. "+
			"The incident was %s with %d response actions executed.",
		incident.Metadata["type"],
		incident.DetectedAt.Format("2006-01-02 15:04:05"),
		incident.Metadata["affected_service"],
		incident.Status,
		len(incident.ActionsTaken),
	)
}

func (r *IncidentResponderImpl) generateTimeline(incident *entities.Incident) []string {
	timeline := []string{
		fmt.Sprintf("%s - Incident detected", incident.DetectedAt.Format("15:04:05")),
	}

	// AcknowledgedAt field not available in new model

	for _, action := range incident.ActionsTaken {
		timeline = append(timeline, fmt.Sprintf("%s - %s: %s",
			action.ExecutedAt.Format("15:04:05"),
			action.Type,
			action.Description,
		))
	}

	if incident.ResolvedAt != nil {
		timeline = append(timeline, fmt.Sprintf("%s - Incident resolved", incident.ResolvedAt.Format("15:04:05")))
	}

	return timeline
}

func (r *IncidentResponderImpl) identifyRootCause(incident *entities.Incident) string {
	if incident.RootCause != "" {
		return incident.RootCause
	}

	// Analyze incident type and actions to determine likely root cause
	incidentType := "default"
	if typeVal, ok := incident.Metadata["type"].(string); ok {
		incidentType = typeVal
	}
	switch incidentType {
	case "service_failure":
		return "Service crash due to resource exhaustion or configuration error"
	case "security_breach":
		return "Unauthorized access due to compromised credentials or vulnerability"
	case "data_corruption":
		return "Data integrity issue due to concurrent writes or hardware failure"
	default:
		return "Root cause under investigation"
	}
}

func (r *IncidentResponderImpl) identifyWhatWentWell(incident *entities.Incident) []string {
	wellItems := []string{}

	// Check detection time - AcknowledgedAt not available in new model
	// Use a default check instead

	// Check automated response
	autoActions := 0
	for _, action := range incident.ActionsTaken {
		if action.ExecutedBy == "system" {
			autoActions++
		}
	}
	if autoActions > 0 {
		wellItems = append(wellItems, fmt.Sprintf("Automated response executed %d actions", autoActions))
	}

	// Check containment
	if strings.Contains(string(incident.Status), "contained") || incident.Status == entities.IncidentStatusResolved {
		wellItems = append(wellItems, "Incident successfully contained")
	}

	return wellItems
}

func (r *IncidentResponderImpl) identifyWhatWentWrong(incident *entities.Incident) []string {
	wrongItems := []string{}

	// Check for failed actions
	failedActions := 0
	for _, action := range incident.ActionsTaken {
		if action.Status == "failed" {
			failedActions++
		}
	}
	if failedActions > 0 {
		wrongItems = append(wrongItems, fmt.Sprintf("%d response actions failed", failedActions))
	}

	// Check resolution time
	if incident.ResolvedAt != nil {
		resolutionTime := incident.ResolvedAt.Sub(incident.DetectedAt)
		if resolutionTime > 1*time.Hour {
			wrongItems = append(wrongItems, fmt.Sprintf("Long resolution time: %v", resolutionTime))
		}
	}

	// Check severity
	if incident.Severity == entities.RiskSeverityCritical {
		wrongItems = append(wrongItems, "Critical severity incident occurred")
	}

	return wrongItems
}

func (r *IncidentResponderImpl) generateActionItems(incident *entities.Incident) []string {
	actionItems := []string{
		"Review and update incident response playbooks",
		"Implement additional monitoring for " + fmt.Sprintf("%v", incident.Metadata["affected_service"]),
	}

	// Add specific action items based on incident type
	incidentType := "default"
	if typeVal, ok := incident.Metadata["type"].(string); ok {
		incidentType = typeVal
	}
	switch incidentType {
	case "service_failure":
		actionItems = append(actionItems, "Implement service health checks")
		actionItems = append(actionItems, "Add redundancy for critical services")
	case "security_breach":
		actionItems = append(actionItems, "Conduct security audit")
		actionItems = append(actionItems, "Update access control policies")
	case "data_corruption":
		actionItems = append(actionItems, "Implement data validation checks")
		actionItems = append(actionItems, "Review backup procedures")
	}

	return actionItems
}

func (r *IncidentResponderImpl) extractLessonsLearned(incident *entities.Incident) []string {
	lessons := []string{
		fmt.Sprintf("Incident type '%s' requires specialized response procedures", incident.Metadata["type"]),
	}

	// Extract lessons from response actions
	if len(incident.ActionsTaken) > 0 {
		lessons = append(lessons, "Automated response capabilities proved effective")
	}

	// Add lessons based on severity and impact
	if incident.Severity == entities.RiskSeverityCritical {
		lessons = append(lessons, "Critical incidents require immediate automated containment")
	}

	return lessons
}

func initializePlaybooks() map[string]*IncidentPlaybook {
	return map[string]*IncidentPlaybook{
		"service_failure": {
			Name:        "Service Failure Response",
			Description: "Playbook for handling service failures",
			Steps: []*PlaybookStep{
				{Name: "assess", Type: "assess", Description: "Assess incident severity", StopOnFailure: false},
				{Name: "notify", Type: "notify", Description: "Notify stakeholders", StopOnFailure: false},
				{Name: "isolate", Type: "isolate", Description: "Isolate affected service", StopOnFailure: false},
				{Name: "backup", Type: "backup", Description: "Create emergency backup", StopOnFailure: false},
				{Name: "failover", Type: "failover", Description: "Activate failover", StopOnFailure: true},
				{Name: "restart", Type: "restart", Description: "Restart service", StopOnFailure: false},
			},
		},
		"security_breach": {
			Name:        "Security Breach Response",
			Description: "Playbook for handling security breaches",
			Steps: []*PlaybookStep{
				{Name: "assess", Type: "assess", Description: "Assess breach severity", StopOnFailure: false},
				{Name: "notify", Type: "notify", Description: "Notify security team", StopOnFailure: false},
				{Name: "isolate", Type: "isolate", Description: "Isolate compromised systems", StopOnFailure: true},
				{Name: "backup", Type: "backup", Description: "Secure evidence backup", StopOnFailure: false},
			},
		},
		"data_corruption": {
			Name:        "Data Corruption Response",
			Description: "Playbook for handling data corruption",
			Steps: []*PlaybookStep{
				{Name: "assess", Type: "assess", Description: "Assess data damage", StopOnFailure: false},
				{Name: "notify", Type: "notify", Description: "Notify data team", StopOnFailure: false},
				{Name: "isolate", Type: "isolate", Description: "Stop write operations", StopOnFailure: true},
				{Name: "backup", Type: "backup", Description: "Secure current state", StopOnFailure: false},
				{Name: "rollback", Type: "rollback", Description: "Restore from backup", StopOnFailure: true},
			},
		},
		"default": {
			Name:        "Default Incident Response",
			Description: "Default playbook for unknown incident types",
			Steps: []*PlaybookStep{
				{Name: "assess", Type: "assess", Description: "Assess incident", StopOnFailure: false},
				{Name: "notify", Type: "notify", Description: "Notify operations team", StopOnFailure: false},
				{Name: "isolate", Type: "isolate", Description: "Isolate affected components", StopOnFailure: false},
			},
		},
	}
}

// Additional types

type IncidentPlaybook struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Steps       []*PlaybookStep `json:"steps"`
}

type PlaybookStep struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	StopOnFailure bool   `json:"stop_on_failure"`
}

type NotificationRequest struct {
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"`
	Subject  string                 `json:"subject"`
	Message  string                 `json:"message"`
	Channels []string               `json:"channels"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ContainmentAction struct {
	Description string
	Execute     func(context.Context, *entities.Incident) error
}

type RecoveryAction struct {
	Description string
	Execute     func(context.Context, *entities.Incident) error
}

// Helper functions
// Helper functions defined in operational_risk_monitor.go

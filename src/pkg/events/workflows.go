package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Workflow represents a business process workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      WorkflowStatus         `json:"status"`
	Steps       []WorkflowStep         `json:"steps"`
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        StepType               `json:"type"`
	Status      StepStatus             `json:"status"`
	TriggerEvent string                `json:"trigger_event,omitempty"`
	PublishEvent string                `json:"publish_event,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

type StepType string

const (
	StepTypeEvent      StepType = "event"
	StepTypeWait       StepType = "wait"
	StepTypeCondition  StepType = "condition"
	StepTypeParallel   StepType = "parallel"
	StepTypeSequential StepType = "sequential"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// WorkflowEngine manages workflow execution
type WorkflowEngine struct {
	eventClient     EventPublisher
	eventConsumer   EventConsumer
	activeWorkflows map[string]*Workflow
	stepHandlers    map[StepType]StepHandler
}

// StepHandler defines the interface for handling workflow steps
type StepHandler interface {
	Execute(ctx context.Context, workflow *Workflow, step *WorkflowStep) error
}

func NewWorkflowEngine(eventClient EventPublisher, eventConsumer EventConsumer) *WorkflowEngine {
	engine := &WorkflowEngine{
		eventClient:     eventClient,
		eventConsumer:   eventConsumer,
		activeWorkflows: make(map[string]*Workflow),
		stepHandlers:    make(map[StepType]StepHandler),
	}

	// Register default step handlers
	engine.RegisterStepHandler(StepTypeEvent, &EventStepHandler{eventClient: eventClient})
	engine.RegisterStepHandler(StepTypeWait, &WaitStepHandler{})
	engine.RegisterStepHandler(StepTypeCondition, &ConditionStepHandler{})

	return engine
}

func (e *WorkflowEngine) RegisterStepHandler(stepType StepType, handler StepHandler) {
	e.stepHandlers[stepType] = handler
}

func (e *WorkflowEngine) StartWorkflow(ctx context.Context, workflow *Workflow) error {
	log.Printf("[WorkflowEngine] Starting workflow %s (%s)", workflow.ID, workflow.Name)

	workflow.Status = WorkflowStatusRunning
	workflow.UpdatedAt = time.Now()
	e.activeWorkflows[workflow.ID] = workflow

	// Start executing the first step
	return e.executeNextStep(ctx, workflow)
}

func (e *WorkflowEngine) executeNextStep(ctx context.Context, workflow *Workflow) error {
	// Find the next pending step
	var nextStep *WorkflowStep
	for i := range workflow.Steps {
		if workflow.Steps[i].Status == StepStatusPending {
			nextStep = &workflow.Steps[i]
			break
		}
	}

	if nextStep == nil {
		// No more steps, workflow is complete
		return e.completeWorkflow(workflow)
	}

	return e.executeStep(ctx, workflow, nextStep)
}

func (e *WorkflowEngine) executeStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	log.Printf("[WorkflowEngine] Executing step %s (%s) in workflow %s", step.ID, step.Name, workflow.ID)

	step.Status = StepStatusRunning
	now := time.Now()
	step.StartedAt = &now
	workflow.UpdatedAt = time.Now()

	handler, exists := e.stepHandlers[step.Type]
	if !exists {
		return fmt.Errorf("no handler found for step type: %s", step.Type)
	}

	// Execute the step
	err := handler.Execute(ctx, workflow, step)
	if err != nil {
		step.Status = StepStatusFailed
		step.Error = err.Error()
		
		// Check if we should retry
		if step.RetryCount < step.MaxRetries {
			step.RetryCount++
			log.Printf("[WorkflowEngine] Retrying step %s (attempt %d/%d)", step.ID, step.RetryCount, step.MaxRetries)
			step.Status = StepStatusPending
			time.AfterFunc(time.Second*time.Duration(step.RetryCount), func() {
				_ = e.executeStep(ctx, workflow, step) // Execute retry step
			})
			return nil
		}

		workflow.Status = WorkflowStatusFailed
		return fmt.Errorf("step %s failed: %w", step.ID, err)
	}

	// Step completed successfully
	step.Status = StepStatusCompleted
	completedAt := time.Now()
	step.CompletedAt = &completedAt

	// Continue to next step
	return e.executeNextStep(ctx, workflow)
}

func (e *WorkflowEngine) completeWorkflow(workflow *Workflow) error {
	log.Printf("[WorkflowEngine] Completing workflow %s", workflow.ID)

	workflow.Status = WorkflowStatusCompleted
	completedAt := time.Now()
	workflow.CompletedAt = &completedAt
	workflow.UpdatedAt = time.Now()

	// Remove from active workflows
	delete(e.activeWorkflows, workflow.ID)

	// Publish workflow completion event
	if e.eventClient != nil {
		event := CreateSystemEvent(
			"workflow.completed",
			"workflow-engine",
			SystemEventData{
				Service:   "workflow-engine",
				Component: "workflow",
				Message:   fmt.Sprintf("Workflow %s completed successfully", workflow.Name),
				Metadata: map[string]interface{}{
					"workflow_id": workflow.ID,
					"duration":    time.Since(workflow.CreatedAt).String(),
				},
			},
		)
		_ = e.eventClient.Publish(context.Background(), StreamSystem, event) // Best effort workflow completion event
	}

	return nil
}

// Step Handlers

// EventStepHandler handles event-based steps
type EventStepHandler struct {
	eventClient EventPublisher
}

func (h *EventStepHandler) Execute(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	if step.PublishEvent == "" {
		return fmt.Errorf("event step must specify publish_event")
	}

	// Create event from step data
	event := Event{
		Type:    step.PublishEvent,
		Source:  "workflow-engine",
		Payload: step.Payload,
	}

	// Merge workflow context into event data
	if event.Payload == nil {
		event.Payload = make(map[string]interface{})
	}
	for k, v := range workflow.Context {
		if _, exists := event.Payload[k]; !exists {
			event.Payload[k] = v
		}
	}

	// Determine target stream based on event type
	stream := getStreamForEventType(step.PublishEvent)
	
	log.Printf("[EventStepHandler] Publishing event %s to stream %s", step.PublishEvent, stream)
	return h.eventClient.Publish(ctx, stream, event)
}

// WaitStepHandler handles wait/delay steps
type WaitStepHandler struct{}

func (h *WaitStepHandler) Execute(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	if step.Timeout <= 0 {
		step.Timeout = 5 * time.Second // Default wait time
	}

	log.Printf("[WaitStepHandler] Waiting for %v", step.Timeout)
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(step.Timeout):
		return nil
	}
}

// ConditionStepHandler handles conditional logic steps
type ConditionStepHandler struct{}

func (h *ConditionStepHandler) Execute(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	// Simple condition evaluation based on workflow context
	condition, ok := step.Payload["condition"].(string)
	if !ok {
		return fmt.Errorf("condition step must specify condition")
	}

	// Basic condition evaluation (in production, use a proper expression evaluator)
	result := h.evaluateCondition(condition, workflow.Context)
	
	log.Printf("[ConditionStepHandler] Condition '%s' evaluated to %t", condition, result)
	
	if !result {
		step.Status = StepStatusSkipped
		return nil
	}

	return nil
}

func (h *ConditionStepHandler) evaluateCondition(condition string, context map[string]interface{}) bool {
	// Simplified condition evaluation
	// In production, use a proper expression language like CEL or govaluate
	
	switch condition {
	case "always_true":
		return true
	case "always_false":
		return false
	default:
		// Try to find boolean value in context
		if value, exists := context[condition]; exists {
			if boolValue, ok := value.(bool); ok {
				return boolValue
			}
		}
		return false
	}
}

// Predefined Workflows

// CreateClientOnboardingWorkflow creates a workflow for client onboarding
func CreateClientOnboardingWorkflow(clientID string) *Workflow {
	workflowID := fmt.Sprintf("client-onboarding-%s-%d", clientID, time.Now().Unix())
	
	return &Workflow{
		ID:          workflowID,
		Name:        "Client Onboarding",
		Description: "Complete client onboarding process",
		Status:      WorkflowStatusPending,
		Context: map[string]interface{}{
			"client_id": clientID,
		},
		Steps: []WorkflowStep{
			{
				ID:           "validate-client",
				Name:         "Validate Client Information",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "client.validation_requested",
				Payload: map[string]interface{}{
					"validation_type": "onboarding",
				},
				MaxRetries: 3,
			},
			{
				ID:      "wait-validation",
				Name:    "Wait for Validation",
				Type:    StepTypeWait,
				Status:  StepStatusPending,
				Timeout: 30 * time.Second,
			},
			{
				ID:           "setup-profile",
				Name:         "Setup Client Profile",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventClientOnboarded,
				MaxRetries:   2,
			},
			{
				ID:           "create-welcome-project",
				Name:         "Create Welcome Project",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventProjectCreated,
				Payload: map[string]interface{}{
					"project_type": "welcome",
					"priority":     "high",
				},
				MaxRetries: 2,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateContentCreationWorkflow creates a workflow for content creation
func CreateContentCreationWorkflow(projectID, contentType string) *Workflow {
	workflowID := fmt.Sprintf("content-creation-%s-%d", projectID, time.Now().Unix())
	
	return &Workflow{
		ID:          workflowID,
		Name:        "Content Creation",
		Description: "Complete content creation and approval process",
		Status:      WorkflowStatusPending,
		Context: map[string]interface{}{
			"project_id":   projectID,
			"content_type": contentType,
		},
		Steps: []WorkflowStep{
			{
				ID:           "create-content",
				Name:         "Create Content",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "content.creation_requested",
				MaxRetries:   3,
			},
			{
				ID:      "wait-creation",
				Name:    "Wait for Content Creation",
				Type:    StepTypeWait,
				Status:  StepStatusPending,
				Timeout: 2 * time.Minute,
			},
			{
				ID:           "quality-review",
				Name:         "Quality Review",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "content.review_requested",
				MaxRetries:   2,
			},
			{
				ID:      "wait-review",
				Name:    "Wait for Review",
				Type:    StepTypeWait,
				Status:  StepStatusPending,
				Timeout: 1 * time.Minute,
			},
			{
				ID:           "client-approval",
				Name:         "Request Client Approval",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "content.approval_requested",
				MaxRetries:   1,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateIncidentResponseWorkflow creates a workflow for incident response
func CreateIncidentResponseWorkflow(incidentID, severity string) *Workflow {
	workflowID := fmt.Sprintf("incident-response-%s-%d", incidentID, time.Now().Unix())
	
	return &Workflow{
		ID:          workflowID,
		Name:        "Incident Response",
		Description: "Automated incident response and resolution",
		Status:      WorkflowStatusPending,
		Context: map[string]interface{}{
			"incident_id": incidentID,
			"severity":    severity,
		},
		Steps: []WorkflowStep{
			{
				ID:           "assess-incident",
				Name:         "Assess Incident Impact",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "incident.assessment_requested",
				MaxRetries:   2,
			},
			{
				ID:           "check-severity",
				Name:         "Check if High Severity",
				Type:         StepTypeCondition,
				Status:       StepStatusPending,
				Payload: map[string]interface{}{
					"condition": "high_severity",
				},
			},
			{
				ID:           "emergency-response",
				Name:         "Emergency Response",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "system.emergency_activated",
				MaxRetries:   1,
			},
			{
				ID:           "notify-stakeholders",
				Name:         "Notify Stakeholders",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "incident.notification_sent",
				MaxRetries:   2,
			},
			{
				ID:           "auto-remediation",
				Name:         "Attempt Auto-Remediation",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: "incident.remediation_requested",
				MaxRetries:   3,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper function to determine target stream for event type
func getStreamForEventType(eventType string) string {
	switch {
	case containsPrefix(eventType, "content."):
		return StreamContent
	case containsPrefix(eventType, "project."):
		return StreamProjects
	case containsPrefix(eventType, "client."):
		return StreamClients
	case containsPrefix(eventType, "invoice."), containsPrefix(eventType, "payment."):
		return StreamFinancial
	case containsPrefix(eventType, "decision."):
		return StreamDecisions
	case containsPrefix(eventType, "risk."), containsPrefix(eventType, "incident."):
		return StreamRisk
	case containsPrefix(eventType, "talent."), containsPrefix(eventType, "engagement."):
		return StreamHR
	case containsPrefix(eventType, "proposal."), containsPrefix(eventType, "member."):
		return StreamGovernance
	case containsPrefix(eventType, "contract."), containsPrefix(eventType, "compliance."):
		return StreamLegal
	case containsPrefix(eventType, "system."), containsPrefix(eventType, "workflow."):
		return StreamSystem
	default:
		return StreamSystem
	}
}

func containsPrefix(str, prefix string) bool {
	return len(str) >= len(prefix) && str[:len(prefix)] == prefix
}

// WorkflowManager provides high-level workflow management
type WorkflowManager struct {
	engine    *WorkflowEngine
	workflows map[string]*Workflow
}

func NewWorkflowManager(eventClient EventPublisher, eventConsumer EventConsumer) *WorkflowManager {
	return &WorkflowManager{
		engine:    NewWorkflowEngine(eventClient, eventConsumer),
		workflows: make(map[string]*Workflow),
	}
}

func (m *WorkflowManager) StartClientOnboarding(ctx context.Context, clientID string) (*Workflow, error) {
	workflow := CreateClientOnboardingWorkflow(clientID)
	m.workflows[workflow.ID] = workflow
	
	err := m.engine.StartWorkflow(ctx, workflow)
	if err != nil {
		delete(m.workflows, workflow.ID)
		return nil, err
	}
	
	return workflow, nil
}

func (m *WorkflowManager) StartContentCreation(ctx context.Context, projectID, contentType string) (*Workflow, error) {
	workflow := CreateContentCreationWorkflow(projectID, contentType)
	m.workflows[workflow.ID] = workflow
	
	err := m.engine.StartWorkflow(ctx, workflow)
	if err != nil {
		delete(m.workflows, workflow.ID)
		return nil, err
	}
	
	return workflow, nil
}

func (m *WorkflowManager) StartIncidentResponse(ctx context.Context, incidentID, severity string) (*Workflow, error) {
	workflow := CreateIncidentResponseWorkflow(incidentID, severity)
	m.workflows[workflow.ID] = workflow
	
	err := m.engine.StartWorkflow(ctx, workflow)
	if err != nil {
		delete(m.workflows, workflow.ID)
		return nil, err
	}
	
	return workflow, nil
}

func (m *WorkflowManager) GetWorkflow(workflowID string) (*Workflow, bool) {
	workflow, exists := m.workflows[workflowID]
	return workflow, exists
}

func (m *WorkflowManager) ListActiveWorkflows() []*Workflow {
	var active []*Workflow
	for _, workflow := range m.workflows {
		if workflow.Status == WorkflowStatusRunning || workflow.Status == WorkflowStatusPending {
			active = append(active, workflow)
		}
	}
	return active
}
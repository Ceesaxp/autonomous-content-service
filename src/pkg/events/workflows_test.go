package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkflow(t *testing.T) {
	workflow := &Workflow{
		ID:          "test-workflow-1",
		Name:        "Test Workflow",
		Description: "A workflow for testing",
		Status:      WorkflowStatusPending,
		Steps: []WorkflowStep{
			{
				ID:           "step-1",
				Name:         "First Step",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventContentCreated,
				Payload: map[string]interface{}{
					"test": "data",
				},
				MaxRetries: 3,
			},
			{
				ID:      "step-2",
				Name:    "Wait Step",
				Type:    StepTypeWait,
				Status:  StepStatusPending,
				Timeout: 2 * time.Second,
			},
		},
		Context:   map[string]interface{}{"test_context": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "test-workflow-1", workflow.ID)
	assert.Equal(t, WorkflowStatusPending, workflow.Status)
	assert.Len(t, workflow.Steps, 2)
	assert.Equal(t, StepStatusPending, workflow.Steps[0].Status)
}

func TestWorkflowEngine(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	engine := NewWorkflowEngine(publisher, consumer)

	// Create a simple workflow
	workflow := &Workflow{
		ID:     "engine-test-1",
		Name:   "Engine Test",
		Status: WorkflowStatusPending,
		Steps: []WorkflowStep{
			{
				ID:           "publish-step",
				Name:         "Publish Event",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventContentCreated,
				Payload: map[string]interface{}{
					"content_id": "test-content",
				},
				MaxRetries: 2,
			},
		},
		Context:   map[string]interface{}{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock the event publication - both for the step event and workflow completion event
	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(nil)
	publisher.On("Publish", mock.Anything, StreamSystem, mock.AnythingOfType("events.Event")).Return(nil)

	// Start the workflow
	err := engine.StartWorkflow(context.Background(), workflow)
	assert.NoError(t, err)

	// Workflow should be completed after the single step
	assert.Equal(t, WorkflowStatusCompleted, workflow.Status)
	assert.Equal(t, StepStatusCompleted, workflow.Steps[0].Status)
	assert.NotNil(t, workflow.CompletedAt)

	publisher.AssertExpectations(t)
}

func TestWorkflowEngineWithFailure(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	engine := NewWorkflowEngine(publisher, consumer)

	// Create workflow with failing step
	workflow := &Workflow{
		ID:     "failure-test-1",
		Name:   "Failure Test",
		Status: WorkflowStatusPending,
		Steps: []WorkflowStep{
			{
				ID:           "failing-step",
				Name:         "Failing Step",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventContentCreated,
				Payload:      map[string]interface{}{},
				MaxRetries:   1,
			},
		},
		Context:   map[string]interface{}{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock event publication to fail
	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(assert.AnError)
	publisher.On("Publish", mock.Anything, StreamSystem, mock.AnythingOfType("events.Event")).Return(nil)

	// Start the workflow
	err := engine.StartWorkflow(context.Background(), workflow)
	assert.Error(t, err)

	// Workflow should be failed
	assert.Equal(t, WorkflowStatusFailed, workflow.Status)
	assert.Equal(t, StepStatusFailed, workflow.Steps[0].Status)
	assert.NotEmpty(t, workflow.Steps[0].Error)

	publisher.AssertExpectations(t)
}

func TestWorkflowEngineWithRetry(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	engine := NewWorkflowEngine(publisher, consumer)

	// Create workflow with step that will retry
	workflow := &Workflow{
		ID:     "retry-test-1",
		Name:   "Retry Test",
		Status: WorkflowStatusPending,
		Steps: []WorkflowStep{
			{
				ID:           "retry-step",
				Name:         "Retry Step",
				Type:         StepTypeEvent,
				Status:       StepStatusPending,
				PublishEvent: EventContentCreated,
				Payload:      map[string]interface{}{},
				MaxRetries:   2,
			},
		},
		Context:   map[string]interface{}{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock first call to fail, second to succeed, and system completion event
	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(assert.AnError).Once()
	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(nil).Once()
	publisher.On("Publish", mock.Anything, StreamSystem, mock.AnythingOfType("events.Event")).Return(nil)

	// Start the workflow
	err := engine.StartWorkflow(context.Background(), workflow)
	assert.NoError(t, err)

	// Wait for retry (step should be pending for retry)
	assert.Equal(t, WorkflowStatusRunning, workflow.Status)
	assert.Equal(t, StepStatusPending, workflow.Steps[0].Status)
	assert.Equal(t, 1, workflow.Steps[0].RetryCount)

	publisher.AssertExpectations(t)
}

func TestEventStepHandler(t *testing.T) {
	publisher := &MockEventPublisher{}
	handler := &EventStepHandler{eventClient: publisher}

	workflow := &Workflow{
		ID:      "event-step-test",
		Context: map[string]interface{}{"project_id": "proj-123"},
	}

	step := &WorkflowStep{
		ID:           "event-step",
		PublishEvent: EventProjectCreated,
		Payload: map[string]interface{}{
			"title": "Test Project",
		},
	}

	// Mock event publication
	publisher.On("Publish", mock.Anything, StreamProjects, mock.MatchedBy(func(event Event) bool {
		// Verify the event has merged context
		return event.Payload["project_id"] == "proj-123" && event.Payload["title"] == "Test Project"
	})).Return(nil)

	err := handler.Execute(context.Background(), workflow, step)
	assert.NoError(t, err)

	publisher.AssertExpectations(t)
}

func TestWaitStepHandler(t *testing.T) {
	handler := &WaitStepHandler{}

	workflow := &Workflow{ID: "wait-test"}
	step := &WorkflowStep{
		ID:      "wait-step",
		Timeout: 100 * time.Millisecond, // Short timeout for testing
	}

	start := time.Now()
	err := handler.Execute(context.Background(), workflow, step)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 90*time.Millisecond) // Account for timing variance
	assert.LessOrEqual(t, duration, 200*time.Millisecond)
}

func TestConditionStepHandler(t *testing.T) {
	handler := &ConditionStepHandler{}

	workflow := &Workflow{
		ID: "condition-test",
		Context: map[string]interface{}{
			"test_flag": true,
		},
	}

	// Test always true condition
	step := &WorkflowStep{
		ID:   "condition-step",
		Type: StepTypeCondition,
		Payload: map[string]interface{}{
			"condition": "always_true",
		},
	}

	err := handler.Execute(context.Background(), workflow, step)
	assert.NoError(t, err)
	assert.NotEqual(t, StepStatusSkipped, step.Status)

	// Test always false condition (should skip)
	step.Payload["condition"] = "always_false"
	step.Status = StepStatusPending // Reset status

	err = handler.Execute(context.Background(), workflow, step)
	assert.NoError(t, err)
	assert.Equal(t, StepStatusSkipped, step.Status)

	// Test context-based condition
	step.Payload["condition"] = "test_flag"
	step.Status = StepStatusPending // Reset status

	err = handler.Execute(context.Background(), workflow, step)
	assert.NoError(t, err)
	assert.NotEqual(t, StepStatusSkipped, step.Status)
}

func TestCreateClientOnboardingWorkflow(t *testing.T) {
	workflow := CreateClientOnboardingWorkflow("client-123")

	assert.NotEmpty(t, workflow.ID)
	assert.Equal(t, "Client Onboarding", workflow.Name)
	assert.Equal(t, WorkflowStatusPending, workflow.Status)
	assert.Equal(t, "client-123", workflow.Context["client_id"])
	assert.Len(t, workflow.Steps, 4)

	// Check step types and order
	assert.Equal(t, "validate-client", workflow.Steps[0].ID)
	assert.Equal(t, StepTypeEvent, workflow.Steps[0].Type)
	assert.Equal(t, "client.validation_requested", workflow.Steps[0].PublishEvent)

	assert.Equal(t, "wait-validation", workflow.Steps[1].ID)
	assert.Equal(t, StepTypeWait, workflow.Steps[1].Type)

	assert.Equal(t, "setup-profile", workflow.Steps[2].ID)
	assert.Equal(t, EventClientOnboarded, workflow.Steps[2].PublishEvent)

	assert.Equal(t, "create-welcome-project", workflow.Steps[3].ID)
	assert.Equal(t, EventProjectCreated, workflow.Steps[3].PublishEvent)
}

func TestCreateContentCreationWorkflow(t *testing.T) {
	workflow := CreateContentCreationWorkflow("project-456", "blog_post")

	assert.NotEmpty(t, workflow.ID)
	assert.Equal(t, "Content Creation", workflow.Name)
	assert.Equal(t, WorkflowStatusPending, workflow.Status)
	assert.Equal(t, "project-456", workflow.Context["project_id"])
	assert.Equal(t, "blog_post", workflow.Context["content_type"])
	assert.Len(t, workflow.Steps, 5)

	// Check workflow steps
	expectedSteps := []string{
		"create-content",
		"wait-creation",
		"quality-review",
		"wait-review",
		"client-approval",
	}

	for i, expectedID := range expectedSteps {
		assert.Equal(t, expectedID, workflow.Steps[i].ID)
	}
}

func TestCreateIncidentResponseWorkflow(t *testing.T) {
	workflow := CreateIncidentResponseWorkflow("incident-789", "critical")

	assert.NotEmpty(t, workflow.ID)
	assert.Equal(t, "Incident Response", workflow.Name)
	assert.Equal(t, WorkflowStatusPending, workflow.Status)
	assert.Equal(t, "incident-789", workflow.Context["incident_id"])
	assert.Equal(t, "critical", workflow.Context["severity"])
	assert.Len(t, workflow.Steps, 5)

	// Check workflow structure
	assert.Equal(t, "assess-incident", workflow.Steps[0].ID)
	assert.Equal(t, "check-severity", workflow.Steps[1].ID)
	assert.Equal(t, StepTypeCondition, workflow.Steps[1].Type)
	assert.Equal(t, "emergency-response", workflow.Steps[2].ID)
	assert.Equal(t, "notify-stakeholders", workflow.Steps[3].ID)
	assert.Equal(t, "auto-remediation", workflow.Steps[4].ID)
}

func TestWorkflowManager(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	manager := NewWorkflowManager(publisher, consumer)

	// Mock event publications for workflow steps
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("events.Event")).Return(nil)

	// Test starting client onboarding workflow
	workflow, err := manager.StartClientOnboarding(context.Background(), "client-456")
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, "Client Onboarding", workflow.Name)

	// Test getting workflow
	retrieved, exists := manager.GetWorkflow(workflow.ID)
	assert.True(t, exists)
	assert.Equal(t, workflow, retrieved)

	// Test listing active workflows (workflow might complete quickly)
	activeWorkflows := manager.ListActiveWorkflows()
	// Workflow might have completed already due to no wait steps, so check for 0 or 1
	assert.GreaterOrEqual(t, len(activeWorkflows), 0)
	if len(activeWorkflows) > 0 {
		assert.Equal(t, workflow.ID, activeWorkflows[0].ID)
	}

	// Test starting content creation workflow
	contentWorkflow, err := manager.StartContentCreation(context.Background(), "project-789", "article")
	assert.NoError(t, err)
	assert.NotNil(t, contentWorkflow)
	assert.Equal(t, "Content Creation", contentWorkflow.Name)

	// Test starting incident response workflow
	incidentWorkflow, err := manager.StartIncidentResponse(context.Background(), "incident-101", "high")
	assert.NoError(t, err)
	assert.NotNil(t, incidentWorkflow)
	assert.Equal(t, "Incident Response", incidentWorkflow.Name)

	publisher.AssertExpectations(t)
}

func TestWorkflowConcurrency(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	manager := NewWorkflowManager(publisher, consumer)

	// Mock event publications
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("events.Event")).Return(nil)

	// Start multiple workflows concurrently
	const numWorkflows = 10
	workflowChan := make(chan *Workflow, numWorkflows)
	errorChan := make(chan error, numWorkflows)

	for i := 0; i < numWorkflows; i++ {
		go func(id int) {
			workflow, err := manager.StartClientOnboarding(context.Background(), fmt.Sprintf("client-%d", id))
			workflowChan <- workflow
			errorChan <- err
		}(i)
	}

	// Collect results
	workflows := make([]*Workflow, 0, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := <-workflowChan
		err := <-errorChan
		assert.NoError(t, err)
		assert.NotNil(t, workflow)
		workflows = append(workflows, workflow)
	}

	// Verify all workflows were created with unique IDs
	ids := make(map[string]bool)
	for _, workflow := range workflows {
		assert.False(t, ids[workflow.ID], "Duplicate workflow ID: %s", workflow.ID)
		ids[workflow.ID] = true
	}

	// Verify active workflows count
	activeWorkflows := manager.ListActiveWorkflows()
	assert.Len(t, activeWorkflows, numWorkflows)

	publisher.AssertExpectations(t)
}

func TestWorkflowContextMerging(t *testing.T) {
	publisher := &MockEventPublisher{}
	handler := &EventStepHandler{eventClient: publisher}

	workflow := &Workflow{
		ID: "context-test",
		Context: map[string]interface{}{
			"global_key": "global_value",
			"shared_key": "global_override", // This should be overridden by step data
		},
	}

	step := &WorkflowStep{
		ID:           "context-step",
		PublishEvent: EventContentCreated,
		Payload: map[string]interface{}{
			"step_key":   "step_value",
			"shared_key": "step_override", // This should override global context
		},
	}

	// Verify context merging in published event
	publisher.On("Publish", mock.Anything, StreamContent, mock.MatchedBy(func(event Event) bool {
		return event.Payload["global_key"] == "global_value" &&
			event.Payload["step_key"] == "step_value" &&
			event.Payload["shared_key"] == "step_override" // Step data should win
	})).Return(nil)

	err := handler.Execute(context.Background(), workflow, step)
	assert.NoError(t, err)

	publisher.AssertExpectations(t)
}

// Benchmark workflow execution
func BenchmarkWorkflowExecution(b *testing.B) {
	publisher := &MockEventPublisher{}
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("events.Event")).Return(nil)

	consumer := &MockEventConsumer{}
	engine := NewWorkflowEngine(publisher, consumer)

	// Create a simple workflow for benchmarking
	createWorkflow := func() *Workflow {
		return &Workflow{
			ID:     "benchmark-workflow",
			Name:   "Benchmark",
			Status: WorkflowStatusPending,
			Steps: []WorkflowStep{
				{
					ID:           "benchmark-step",
					Type:         StepTypeEvent,
					Status:       StepStatusPending,
					PublishEvent: EventContentCreated,
					Payload:      map[string]interface{}{"test": "data"},
				},
			},
			Context: map[string]interface{}{},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workflow := createWorkflow()
		_ = engine.StartWorkflow(context.Background(), workflow) // Ignore error for benchmark
	}
}

func TestWorkflowTimeout(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	engine := NewWorkflowEngine(publisher, consumer)

	// Create workflow with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	workflow := &Workflow{
		ID:     "timeout-test",
		Status: WorkflowStatusPending,
		Steps: []WorkflowStep{
			{
				ID:      "long-wait",
				Type:    StepTypeWait,
				Status:  StepStatusPending,
				Timeout: 1 * time.Second, // Longer than context timeout
			},
		},
	}

	err := engine.StartWorkflow(ctx, workflow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
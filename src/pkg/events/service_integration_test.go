package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Ceesaxp/autonomous-content-service/src/pkg/discovery"
)

// MockServiceDiscovery for testing
type MockServiceDiscovery struct {
	mock.Mock
}

func (m *MockServiceDiscovery) RegisterService(serviceName, address, healthCheck string) error {
	args := m.Called(serviceName, address, healthCheck)
	return args.Error(0)
}

func (m *MockServiceDiscovery) DeregisterService(serviceID string) error {
	args := m.Called(serviceID)
	return args.Error(0)
}

func (m *MockServiceDiscovery) DiscoverService(serviceName string) ([]*discovery.ServiceEndpoint, error) {
	args := m.Called(serviceName)
	return args.Get(0).([]*discovery.ServiceEndpoint), args.Error(1)
}

func (m *MockServiceDiscovery) DiscoverHealthyService(serviceName string) (*discovery.ServiceEndpoint, error) {
	args := m.Called(serviceName)
	return args.Get(0).(*discovery.ServiceEndpoint), args.Error(1)
}

func (m *MockServiceDiscovery) DeregisterAll() error {
	args := m.Called()
	return args.Error(0)
}

func TestServiceEventBus(t *testing.T) {
	// Create mocks
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	// Configure service event bus
	config := ServiceEventConfig{
		ServiceName:      "test-service",
		EventClient:      publisher,
		EventConsumer:    consumer,
		ServiceDiscovery: discovery,
		ConsumerGroup:    "test-consumers",
	}

	bus := NewServiceEventBus(config)

	// Test service registration
	discovery.On("RegisterService", "test-service", "localhost:8080", "/health").Return(nil)
	// No Subscribe call expected since no handler is registered for test-service

	err := bus.Start(context.Background())
	assert.NoError(t, err)

	// Test event publishing
	eventData := map[string]interface{}{
		"test_key": "test_value",
	}

	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(nil)

	err = bus.PublishEvent(context.Background(), EventContentCreated, eventData)
	assert.NoError(t, err)

	// Test stopping the bus
	// DeregisterAll is only called for ConsulClient instances, not for mock
	// No Close call expected in current implementation

	err = bus.Stop()
	assert.NoError(t, err)

	// Verify all expectations
	publisher.AssertExpectations(t)
	consumer.AssertExpectations(t)
	discovery.AssertExpectations(t)
}

func TestServiceEventBusFactory(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	factory := NewServiceEventBusFactory(publisher, consumer, discovery)

	// Test creating content service bus
	contentBus := factory.CreateEventBus("content-service")
	assert.NotNil(t, contentBus)
	assert.Equal(t, "content-service", contentBus.serviceName)

	// Test creating decision service bus
	decisionBus := factory.CreateEventBus("decision-service")
	assert.NotNil(t, decisionBus)
	assert.Equal(t, "decision-service", decisionBus.serviceName)

	// Test creating unknown service bus (should still work but with no default handler)
	unknownBus := factory.CreateEventBus("unknown-service")
	assert.NotNil(t, unknownBus)
	assert.Equal(t, "unknown-service", unknownBus.serviceName)
}

func TestWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow integration test in short mode")
	}

	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	config := ServiceEventConfig{
		ServiceName:      "test-service",
		EventClient:      publisher,
		EventConsumer:    consumer,
		ServiceDiscovery: discovery,
	}

	bus := NewServiceEventBus(config)

	// Test invalid workflow type first (no timeout)
	_, err := bus.StartWorkflow(context.Background(), "invalid_workflow", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown workflow type")

	// Test starting incident response workflow (fastest one)
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("events.Event")).Return(nil)

	workflow, err := bus.StartWorkflow(context.Background(), "incident_response", map[string]interface{}{
		"incident_id": "incident-789",
		"severity":    "high",
	})

	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, "Incident Response", workflow.Name)

	publisher.AssertExpectations(t)
}

func TestEventMetrics(t *testing.T) {
	metrics := NewEventMetrics()

	// Test recording events
	metrics.RecordEventPublished("test.event")
	metrics.RecordEventPublished("test.event")
	metrics.RecordEventPublished("other.event")

	metrics.RecordEventConsumed("test.event")
	metrics.RecordEventProcessed("test.event", 50*time.Millisecond)
	metrics.RecordEventFailed("test.event")

	metrics.RecordWorkflowStarted()
	metrics.RecordWorkflowCompleted()
	metrics.RecordWorkflowFailed()

	// Test getting metrics
	metricsData := metrics.GetMetrics()

	eventsPublished := metricsData["events_published"].(map[string]int64)
	assert.Equal(t, int64(2), eventsPublished["test.event"])
	assert.Equal(t, int64(1), eventsPublished["other.event"])

	eventsConsumed := metricsData["events_consumed"].(map[string]int64)
	assert.Equal(t, int64(1), eventsConsumed["test.event"])

	eventsProcessed := metricsData["events_processed"].(map[string]int64)
	assert.Equal(t, int64(1), eventsProcessed["test.event"])

	eventsFailed := metricsData["events_failed"].(map[string]int64)
	assert.Equal(t, int64(1), eventsFailed["test.event"])

	assert.Equal(t, int64(1), metricsData["workflows_started"])
	assert.Equal(t, int64(1), metricsData["workflows_completed"])
	assert.Equal(t, int64(1), metricsData["workflows_failed"])
}

func TestHealthChecker(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	config := ServiceEventConfig{
		ServiceName:      "test-service",
		EventClient:      publisher,
		EventConsumer:    consumer,
		ServiceDiscovery: discovery,
	}

	bus := NewServiceEventBus(config)
	metricsCollector := NewMetricsCollector(bus)
	healthChecker := NewHealthChecker(bus, metricsCollector)

	// Test health check
	health := healthChecker.CheckHealth()

	assert.Equal(t, "healthy", health["status"])

	checks := health["checks"].(map[string]interface{})
	assert.Contains(t, checks, "event_client")
	assert.Contains(t, checks, "subscriptions")
	assert.Contains(t, checks, "metrics")

	eventClientCheck := checks["event_client"].(map[string]interface{})
	assert.Equal(t, "healthy", eventClientCheck["status"])

	subscriptionsCheck := checks["subscriptions"].(map[string]interface{})
	assert.Equal(t, "healthy", subscriptionsCheck["status"])
	assert.Equal(t, 0, subscriptionsCheck["count"]) // No active subscriptions in test
}

func TestEventErrorHandling(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	config := ServiceEventConfig{
		ServiceName:      "test-service",
		EventClient:      publisher,
		EventConsumer:    consumer,
		ServiceDiscovery: discovery,
	}

	bus := NewServiceEventBus(config)

	// Create a failing handler
	handler := &MockServiceEventHandler{}
	handler.On("GetServiceName").Return("test-service")
	handler.On("GetEventTypes").Return([]string{EventContentCreated})
	handler.On("Handle", mock.Anything, mock.AnythingOfType("events.Event")).Return(assert.AnError)

	// Note: RegisterEventHandler method not exposed in current implementation
	// This test would require access to internal handler registry
	
	// Test error handling - should publish error event
	event := CreateEvent(EventContentCreated, "external-service", map[string]interface{}{})
	
	// Expect error event to be published
	publisher.On("Publish", mock.Anything, StreamSystem, mock.AnythingOfType("events.Event")).Return(nil)

	err := bus.handleEvent(context.Background(), event, handler)
	assert.Error(t, err)

	handler.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestStreamSelection(t *testing.T) {
	testCases := []struct {
		eventType      string
		expectedStream string
	}{
		{"content.created", StreamContent},
		{"content.updated", StreamContent},
		{"project.created", StreamProjects},
		{"project.status_changed", StreamProjects},
		{"client.onboarded", StreamClients},
		{"client.feedback", StreamClients},
		{"invoice.created", StreamFinancial},
		{"payment.received", StreamFinancial},
		{"decision.created", StreamDecisions},
		{"decision.executed", StreamDecisions},
		{"risk.detected", StreamRisk},
		{"incident.created", StreamRisk},
		{"talent.onboarded", StreamHR},
		{"engagement.created", StreamHR},
		{"proposal.created", StreamGovernance},
		{"member.joined", StreamGovernance},
		{"contract.created", StreamLegal},
		{"compliance.issue", StreamLegal},
		{"system.alert", StreamSystem},
		{"workflow.completed", StreamSystem},
		{"unknown.event", StreamSystem}, // Fallback
	}

	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			stream := getStreamForEventType(tc.eventType)
			assert.Equal(t, tc.expectedStream, stream)
		})
	}
}

func TestEventBusLifecycle(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}
	discovery := &MockServiceDiscovery{}

	config := ServiceEventConfig{
		ServiceName:      "lifecycle-test",
		EventClient:      publisher,
		EventConsumer:    consumer,
		ServiceDiscovery: discovery,
	}

	bus := NewServiceEventBus(config)

	// Test full lifecycle: Start -> Publish -> Stop
	discovery.On("RegisterService", "lifecycle-test", "localhost:8080", "/health").Return(nil)
	consumer.On("Subscribe", mock.Anything, mock.AnythingOfType("string"), "lifecycle-test-consumers", mock.AnythingOfType("events.EventHandler")).Return(nil)

	// Start
	err := bus.Start(context.Background())
	assert.NoError(t, err)

	// Publish events
	publisher.On("Publish", mock.Anything, StreamContent, mock.AnythingOfType("events.Event")).Return(nil)
	publisher.On("Publish", mock.Anything, StreamProjects, mock.AnythingOfType("events.Event")).Return(nil)

	err = bus.PublishEvent(context.Background(), EventContentCreated, map[string]interface{}{
		"content_id": "test-content",
	})
	assert.NoError(t, err)

	err = bus.PublishEvent(context.Background(), EventProjectCreated, map[string]interface{}{
		"project_id": "test-project",
	})
	assert.NoError(t, err)

	// Stop
	discovery.On("DeregisterAll").Return()
	publisher.On("Close").Return(nil)

	err = bus.Stop()
	assert.NoError(t, err)

	// Verify all interactions
	discovery.AssertExpectations(t)
	publisher.AssertExpectations(t)
	consumer.AssertExpectations(t)
}

// Benchmark tests for service integration
func BenchmarkEventBusPublish(b *testing.B) {
	publisher := &MockEventPublisher{}
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("events.Event")).Return(nil)

	config := ServiceEventConfig{
		ServiceName: "benchmark-service",
		EventClient: publisher,
	}

	bus := NewServiceEventBus(config)

	eventData := map[string]interface{}{
		"benchmark": true,
		"iteration": 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventData["iteration"] = i
		_ = bus.PublishEvent(context.Background(), EventContentCreated, eventData) // Ignore error for benchmark
	}
}

func BenchmarkEventHandling(b *testing.B) {
	handler := NewContentServiceServiceEventHandler()

	event := CreateEvent(EventProjectCreated, "test-service", map[string]interface{}{
		"project_id": "benchmark-project",
		"title":      "Benchmark Project",
		"status":     "active",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(context.Background(), event) // Ignore error for benchmark
	}
}

func TestConcurrentEventHandling(t *testing.T) {
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}

	config := ServiceEventConfig{
		ServiceName:   "concurrent-test",
		EventClient:   publisher,
		EventConsumer: consumer,
	}

	bus := NewServiceEventBus(config)

	// Register handler
	handler := &MockServiceEventHandler{}
	handler.On("GetServiceName").Return("concurrent-test")
	handler.On("GetEventTypes").Return([]string{EventContentCreated})
	handler.On("Handle", mock.Anything, mock.AnythingOfType("events.Event")).Return(nil)

	// Note: RegisterEventHandler method not exposed in current implementation
	// This test would require access to internal handler registry

	// Simulate concurrent event handling
	event := CreateEvent(EventContentCreated, "external-service", map[string]interface{}{
		"test": "concurrent",
	})

	// Handle the same event concurrently
	const numGoroutines = 10
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := bus.handleEvent(context.Background(), event, handler)
			errChan <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify handler was called the expected number of times
	handler.AssertNumberOfCalls(t, "Handle", numGoroutines)
}
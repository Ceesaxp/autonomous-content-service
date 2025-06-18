package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventPublisher for testing
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, stream string, event Event) error {
	args := m.Called(ctx, stream, event)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockEventConsumer for testing
type MockEventConsumer struct {
	mock.Mock
}

func (m *MockEventConsumer) Subscribe(ctx context.Context, stream, consumerGroup string, handler EventHandler) error {
	args := m.Called(ctx, stream, consumerGroup, handler)
	return args.Error(0)
}

func (m *MockEventConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockServiceEventHandler for testing
type MockServiceEventHandler struct {
	mock.Mock
}

func (m *MockServiceEventHandler) Handle(ctx context.Context, event Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockServiceEventHandler) GetEventTypes() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockServiceEventHandler) GetServiceName() string {
	args := m.Called()
	return args.String(0)
}

func TestCreateEvent(t *testing.T) {
	eventType := "test.event"
	source := "test-service"
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	event := CreateEvent(eventType, source, payload)

	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, source, event.Source)
	assert.Equal(t, payload, event.Payload)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestCreateContentEvent(t *testing.T) {
	data := ContentEventData{
		ContentID:    "content-123",
		ProjectID:    "project-456",
		ClientID:     "client-789",
		ContentType:  "article",
		Title:        "Test Article",
		Status:       "draft",
		WordCount:    500,
		QualityScore: 0.85,
	}

	event := CreateContentEvent(EventContentCreated, "content-service", data)

	assert.Equal(t, EventContentCreated, event.Type)
	assert.Equal(t, "content-service", event.Source)
	assert.Equal(t, data.ContentID, event.Payload["content_id"])
	assert.Equal(t, data.ProjectID, event.Payload["project_id"])
	assert.Equal(t, data.QualityScore, event.Payload["quality_score"])
}

func TestCreateProjectEvent(t *testing.T) {
	deadline := time.Now().Add(7 * 24 * time.Hour)
	data := ProjectEventData{
		ProjectID: "project-123",
		ClientID:  "client-456",
		Title:     "Test Project",
		Status:    "active",
		Priority:  "high",
		Budget:    5000.0,
		Deadline:  &deadline,
		Progress:  0.3,
	}

	event := CreateProjectEvent(EventProjectCreated, "project-service", data)

	assert.Equal(t, EventProjectCreated, event.Type)
	assert.Equal(t, "project-service", event.Source)
	assert.Equal(t, data.ProjectID, event.Payload["project_id"])
	assert.Equal(t, data.Budget, event.Payload["budget"])
	assert.Equal(t, deadline.Format(time.RFC3339), event.Payload["deadline"])
}

func TestCreateFinancialEvent(t *testing.T) {
	data := FinancialEventData{
		TransactionID: "txn-123",
		InvoiceID:     "inv-456",
		ClientID:      "client-789",
		ProjectID:     "project-101",
		Amount:        1500.0,
		Currency:      "USD",
		Status:        "completed",
		PaymentMethod: "stripe",
	}

	event := CreateFinancialEvent(EventPaymentReceived, "financial-service", data)

	assert.Equal(t, EventPaymentReceived, event.Type)
	assert.Equal(t, "financial-service", event.Source)
	assert.Equal(t, data.TransactionID, event.Payload["transaction_id"])
	assert.Equal(t, data.Amount, event.Payload["amount"])
	assert.Equal(t, data.Currency, event.Payload["currency"])
}

func TestGetStreamForEventType(t *testing.T) {
	testCases := []struct {
		eventType      string
		expectedStream string
	}{
		{EventContentCreated, StreamContent},
		{EventProjectCreated, StreamProjects},
		{EventClientOnboarded, StreamClients},
		{EventInvoiceCreated, StreamFinancial},
		{EventPaymentReceived, StreamFinancial},
		{EventDecisionCreated, StreamDecisions},
		{EventRiskDetected, StreamRisk},
		{EventIncidentCreated, StreamRisk},
		{EventTalentOnboarded, StreamHR},
		{EventProposalCreated, StreamGovernance},
		{EventContractCreated, StreamLegal},
		{EventSystemAlert, StreamSystem},
		{"unknown.event", StreamSystem}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			stream := getStreamForEventType(tc.eventType)
			assert.Equal(t, tc.expectedStream, stream)
		})
	}
}

func TestServiceEventHandlerRegistry(t *testing.T) {
	registry := NewServiceEventHandlerRegistry()

	// Create a mock handler
	handler := &MockServiceEventHandler{}
	handler.On("GetServiceName").Return("test-service")

	// Register the handler
	registry.RegisterHandler(handler)

	// Retrieve the handler
	retrieved, exists := registry.GetHandler("test-service")
	assert.True(t, exists)
	assert.Equal(t, handler, retrieved)

	// Try to get non-existent handler
	_, exists = registry.GetHandler("non-existent")
	assert.False(t, exists)

	// Get all handlers - should include both our test handler and any default handlers
	allHandlers := registry.GetAllHandlers()
	assert.Contains(t, allHandlers, "test-service")
	assert.Equal(t, handler, allHandlers["test-service"])
}

func TestEventRouter(t *testing.T) {
	router := NewEventRouter()

	// Create mock handler
	handler := &MockServiceEventHandler{}
	handler.On("GetServiceName").Return("test-service")
	handler.On("GetEventTypes").Return([]string{EventContentCreated, EventProjectCreated})

	// Register handler
	router.registry.RegisterHandler(handler)

	// Test routing supported event
	event := CreateEvent(EventContentCreated, "external-service", map[string]interface{}{
		"test": "data",
	})

	handler.On("Handle", mock.Anything, event).Return(nil)

	err := router.RouteEvent(context.Background(), "test-service", event)
	assert.NoError(t, err)

	// Test routing unsupported event (should not error, just ignore)
	unsupportedEvent := CreateEvent("unsupported.event", "external-service", map[string]interface{}{})
	err = router.RouteEvent(context.Background(), "test-service", unsupportedEvent)
	assert.NoError(t, err)

	// Test routing to non-existent service
	err = router.RouteEvent(context.Background(), "non-existent", event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler found for service")

	handler.AssertExpectations(t)
}

func TestContentServiceEventHandler(t *testing.T) {
	handler := NewContentServiceServiceEventHandler()

	// Test service name and event types
	assert.Equal(t, "content-service", handler.GetServiceName())
	eventTypes := handler.GetEventTypes()
	assert.Contains(t, eventTypes, EventProjectCreated)
	assert.Contains(t, eventTypes, EventClientOnboarded)
	assert.Contains(t, eventTypes, EventDecisionExecuted)

	// Test handling project created event
	event := CreateEvent(EventProjectCreated, "project-service", map[string]interface{}{
		"project_id": "project-123",
		"title":      "Test Project",
		"status":     "active",
		"budget":     5000.0,
	})

	err := handler.Handle(context.Background(), event)
	assert.NoError(t, err)
}

func TestDecisionServiceEventHandler(t *testing.T) {
	handler := NewDecisionServiceServiceEventHandler()

	// Test service name and event types
	assert.Equal(t, "decision-service", handler.GetServiceName())
	eventTypes := handler.GetEventTypes()
	assert.Contains(t, eventTypes, EventRiskDetected)
	assert.Contains(t, eventTypes, EventIncidentCreated)
	assert.Contains(t, eventTypes, EventComplianceIssue)

	// Test handling risk detected event
	event := CreateEvent(EventRiskDetected, "risk-service", map[string]interface{}{
		"risk_id":  "risk-123",
		"category": "financial",
		"severity": "high",
		"score":    0.85,
	})

	err := handler.Handle(context.Background(), event)
	assert.NoError(t, err)
}

func TestFinancialServiceEventHandler(t *testing.T) {
	handler := NewFinancialServiceServiceEventHandler()

	// Test service name and event types
	assert.Equal(t, "financial-service", handler.GetServiceName())
	eventTypes := handler.GetEventTypes()
	assert.Contains(t, eventTypes, EventProjectCreated)
	assert.Contains(t, eventTypes, EventContentApproved)
	assert.Contains(t, eventTypes, EventClientOnboarded)

	// Test handling project created event
	event := CreateEvent(EventProjectCreated, "project-service", map[string]interface{}{
		"project_id": "project-123",
		"client_id":  "client-456",
		"budget":     2500.0,
	})

	err := handler.Handle(context.Background(), event)
	assert.NoError(t, err)
}

func TestMapToStruct(t *testing.T) {
	// Test successful mapping
	data := map[string]interface{}{
		"content_id":    "content-123",
		"project_id":    "project-456",
		"title":         "Test Content",
		"quality_score": 0.9,
	}

	var result ContentEventData
	err := mapToStruct(data, &result)

	assert.NoError(t, err)
	assert.Equal(t, "content-123", result.ContentID)
	assert.Equal(t, "project-456", result.ProjectID)
	assert.Equal(t, "Test Content", result.Title)
	assert.Equal(t, 0.9, result.QualityScore)

	// Test mapping with invalid data
	invalidData := map[string]interface{}{
		"invalid_field": make(chan int), // Channels can't be marshaled to JSON
	}

	err = mapToStruct(invalidData, &result)
	assert.Error(t, err)
}

func TestEventWithMetadata(t *testing.T) {
	event := CreateEvent("test.event", "test-service", map[string]interface{}{
		"data": "value",
	})

	// Add metadata
	eventWithMeta := event.WithMetadata("correlation_id", "corr-123")
	eventWithMeta = eventWithMeta.WithMetadata("trace_id", "trace-456")

	assert.Equal(t, "corr-123", eventWithMeta.Metadata["correlation_id"])
	assert.Equal(t, "trace-456", eventWithMeta.Metadata["trace_id"])
}

// Benchmark tests
func BenchmarkCreateEvent(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": 3.14,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateEvent("benchmark.event", "test-service", payload)
	}
}

func BenchmarkCreateContentEvent(b *testing.B) {
	data := ContentEventData{
		ContentID:    "content-123",
		ProjectID:    "project-456",
		ContentType:  "article",
		Title:        "Benchmark Article",
		QualityScore: 0.85,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateContentEvent(EventContentCreated, "content-service", data)
	}
}

func BenchmarkMapToStruct(b *testing.B) {
	data := map[string]interface{}{
		"content_id":    "content-123",
		"project_id":    "project-456",
		"title":         "Test Content",
		"quality_score": 0.9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result ContentEventData
		mapToStruct(data, &result)
	}
}

// Integration test with mock Redis client
func TestEventSystemIntegration(t *testing.T) {
	// This would test with a real Redis instance in integration tests
	// For unit tests, we use mocks
	
	publisher := &MockEventPublisher{}
	consumer := &MockEventConsumer{}

	// Test event publishing
	event := CreateEvent(EventContentCreated, "content-service", map[string]interface{}{
		"content_id": "content-123",
	})

	publisher.On("Publish", mock.Anything, StreamContent, event).Return(nil)

	err := publisher.Publish(context.Background(), StreamContent, event)
	assert.NoError(t, err)

	// Test event consumption
	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	consumer.On("Subscribe", mock.Anything, StreamContent, "test-group", mock.AnythingOfType("events.EventHandler")).Return(nil)

	err = consumer.Subscribe(context.Background(), StreamContent, "test-group", handler)
	assert.NoError(t, err)

	publisher.AssertExpectations(t)
	consumer.AssertExpectations(t)
}
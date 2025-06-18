package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/pkg/discovery"
)

// ServiceEventBus provides event-driven communication between microservices
type ServiceEventBus struct {
	serviceName     string
	eventClient     EventPublisher
	eventConsumer   EventConsumer
	serviceDiscovery discovery.ServiceDiscovery
	eventHandlers   map[string]ServiceEventHandler
	subscriptions   map[string]context.CancelFunc
	workflowManager *WorkflowManager
	mu              sync.RWMutex
}

// ServiceEventConfig configures the service event bus
type ServiceEventConfig struct {
	ServiceName      string
	EventClient      EventPublisher
	EventConsumer    EventConsumer
	ServiceDiscovery discovery.ServiceDiscovery
	ConsumerGroup    string
}

func NewServiceEventBus(config ServiceEventConfig) *ServiceEventBus {
	bus := &ServiceEventBus{
		serviceName:      config.ServiceName,
		eventClient:      config.EventClient,
		eventConsumer:    config.EventConsumer,
		serviceDiscovery: config.ServiceDiscovery,
		eventHandlers:    make(map[string]ServiceEventHandler),
		subscriptions:    make(map[string]context.CancelFunc),
	}

	// Initialize workflow manager
	bus.workflowManager = NewWorkflowManager(config.EventClient, config.EventConsumer)

	return bus
}

// RegisterServiceEventHandler registers an event handler for the service
func (bus *ServiceEventBus) RegisterServiceEventHandler(handler ServiceEventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.eventHandlers[handler.GetServiceName()] = handler
	log.Printf("[ServiceEventBus] Registered event handler for %s", handler.GetServiceName())
	return nil
}

// Start begins event consumption and service registration
func (bus *ServiceEventBus) Start(ctx context.Context) error {
	log.Printf("[ServiceEventBus] Starting event bus for service %s", bus.serviceName)

	// Register service with discovery
	if bus.serviceDiscovery != nil {
		address := "localhost:8080" // This should be configurable
		err := bus.serviceDiscovery.RegisterService(bus.serviceName, address, "/health")
		if err != nil {
			log.Printf("[ServiceEventBus] Warning: Failed to register with service discovery: %v", err)
		}
	}

	// Start consuming events for this service
	return bus.startEventConsumption(ctx)
}

// Stop stops event consumption and deregisters the service
func (bus *ServiceEventBus) Stop() error {
	log.Printf("[ServiceEventBus] Stopping event bus for service %s", bus.serviceName)

	bus.mu.Lock()
	defer bus.mu.Unlock()

	// Cancel all subscriptions
	for stream, cancel := range bus.subscriptions {
		log.Printf("[ServiceEventBus] Cancelling subscription to %s", stream)
		cancel()
	}
	bus.subscriptions = make(map[string]context.CancelFunc)

	// Deregister from service discovery
	if bus.serviceDiscovery != nil {
		if consulClient, ok := bus.serviceDiscovery.(*discovery.ConsulClient); ok {
			consulClient.DeregisterAll()
		}
	}

	return nil
}

// startEventConsumption starts consuming events from relevant streams
func (bus *ServiceEventBus) startEventConsumption(ctx context.Context) error {
	handler, exists := bus.eventHandlers[bus.serviceName]
	if !exists {
		log.Printf("[ServiceEventBus] No event handler registered for service %s", bus.serviceName)
		return nil
	}

	// Determine which streams to subscribe to based on event types
	streams := bus.getStreamsForEventTypes(handler.GetEventTypes())
	
	for _, stream := range streams {
		err := bus.subscribeToStream(ctx, stream, handler)
		if err != nil {
			log.Printf("[ServiceEventBus] Failed to subscribe to stream %s: %v", stream, err)
			continue
		}
	}

	return nil
}

// subscribeToStream subscribes to a specific event stream
func (bus *ServiceEventBus) subscribeToStream(ctx context.Context, stream string, handler ServiceEventHandler) error {
	consumerGroup := fmt.Sprintf("%s-consumers", bus.serviceName)
	
	streamCtx, cancel := context.WithCancel(ctx)
	bus.mu.Lock()
	bus.subscriptions[stream] = cancel
	bus.mu.Unlock()

	log.Printf("[ServiceEventBus] Subscribing to stream %s with consumer group %s", stream, consumerGroup)

	eventHandler := func(ctx context.Context, event Event) error {
		return bus.handleEvent(ctx, event, handler)
	}

	go func() {
		err := bus.eventConsumer.Subscribe(streamCtx, stream, consumerGroup, eventHandler)
		if err != nil {
			log.Printf("[ServiceEventBus] Error subscribing to stream %s: %v", stream, err)
		}
	}()

	return nil
}

// handleEvent processes an incoming event
func (bus *ServiceEventBus) handleEvent(ctx context.Context, event Event, handler ServiceEventHandler) error {
	log.Printf("[ServiceEventBus] Received event %s from %s for service %s", event.Type, event.Source, bus.serviceName)

	// Skip events from the same service to avoid loops
	if event.Source == bus.serviceName {
		return nil
	}

	// Handle the event
	err := handler.Handle(ctx, event)
	if err != nil {
		log.Printf("[ServiceEventBus] Error handling event %s: %v", event.Type, err)
		
		// Publish error event for monitoring
		bus.publishErrorEvent(ctx, event, err)
		return err
	}

	return nil
}

// publishErrorEvent publishes an error event for monitoring
func (bus *ServiceEventBus) publishErrorEvent(ctx context.Context, originalEvent Event, err error) {
	errorEvent := CreateSystemEvent(
		"event.processing_error",
		bus.serviceName,
		SystemEventData{
			Service:   bus.serviceName,
			Component: "event_handler",
			Message:   fmt.Sprintf("Failed to process event %s: %v", originalEvent.Type, err),
			AlertLevel: "error",
			Metadata: map[string]interface{}{
				"original_event_id":   originalEvent.ID,
				"original_event_type": originalEvent.Type,
				"original_source":     originalEvent.Source,
				"error_message":       err.Error(),
			},
		},
	)

	bus.eventClient.Publish(ctx, StreamSystem, errorEvent)
}

// PublishEvent publishes an event to the appropriate stream
func (bus *ServiceEventBus) PublishEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := CreateEvent(eventType, bus.serviceName, data)
	stream := getStreamForEventType(eventType)
	
	log.Printf("[ServiceEventBus] Publishing event %s to stream %s", eventType, stream)
	return bus.eventClient.Publish(ctx, stream, event)
}

// PublishTypedEvent publishes a typed event using helper functions
func (bus *ServiceEventBus) PublishTypedEvent(ctx context.Context, event Event) error {
	stream := getStreamForEventType(event.Type)
	
	// Set source if not already set
	if event.Source == "" {
		event.Source = bus.serviceName
	}
	
	log.Printf("[ServiceEventBus] Publishing typed event %s to stream %s", event.Type, stream)
	return bus.eventClient.Publish(ctx, stream, event)
}

// StartWorkflow starts a predefined workflow
func (bus *ServiceEventBus) StartWorkflow(ctx context.Context, workflowType string, params map[string]interface{}) (*Workflow, error) {
	switch workflowType {
	case "client_onboarding":
		clientID, ok := params["client_id"].(string)
		if !ok {
			return nil, fmt.Errorf("client_onboarding workflow requires client_id parameter")
		}
		return bus.workflowManager.StartClientOnboarding(ctx, clientID)
		
	case "content_creation":
		projectID, ok := params["project_id"].(string)
		if !ok {
			return nil, fmt.Errorf("content_creation workflow requires project_id parameter")
		}
		contentType, ok := params["content_type"].(string)
		if !ok {
			contentType = "article" // Default content type
		}
		return bus.workflowManager.StartContentCreation(ctx, projectID, contentType)
		
	case "incident_response":
		incidentID, ok := params["incident_id"].(string)
		if !ok {
			return nil, fmt.Errorf("incident_response workflow requires incident_id parameter")
		}
		severity, ok := params["severity"].(string)
		if !ok {
			severity = "medium" // Default severity
		}
		return bus.workflowManager.StartIncidentResponse(ctx, incidentID, severity)
		
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", workflowType)
	}
}

// GetWorkflow retrieves a workflow by ID
func (bus *ServiceEventBus) GetWorkflow(workflowID string) (*Workflow, bool) {
	return bus.workflowManager.GetWorkflow(workflowID)
}

// ListActiveWorkflows returns all active workflows
func (bus *ServiceEventBus) ListActiveWorkflows() []*Workflow {
	return bus.workflowManager.ListActiveWorkflows()
}

// getStreamsForEventTypes returns the streams to subscribe to for given event types
func (bus *ServiceEventBus) getStreamsForEventTypes(eventTypes []string) []string {
	streamSet := make(map[string]bool)
	
	for _, eventType := range eventTypes {
		stream := getStreamForEventType(eventType)
		streamSet[stream] = true
	}
	
	var streams []string
	for stream := range streamSet {
		streams = append(streams, stream)
	}
	
	return streams
}

// ServiceEventBusFactory creates event buses for different services
type ServiceEventBusFactory struct {
	eventClient      EventPublisher
	eventConsumer    EventConsumer
	serviceDiscovery discovery.ServiceDiscovery
}

func NewServiceEventBusFactory(eventClient EventPublisher, eventConsumer EventConsumer, serviceDiscovery discovery.ServiceDiscovery) *ServiceEventBusFactory {
	return &ServiceEventBusFactory{
		eventClient:      eventClient,
		eventConsumer:    eventConsumer,
		serviceDiscovery: serviceDiscovery,
	}
}

func (factory *ServiceEventBusFactory) CreateEventBus(serviceName string) *ServiceEventBus {
	config := ServiceEventConfig{
		ServiceName:      serviceName,
		EventClient:      factory.eventClient,
		EventConsumer:    factory.eventConsumer,
		ServiceDiscovery: factory.serviceDiscovery,
		ConsumerGroup:    fmt.Sprintf("%s-consumers", serviceName),
	}
	
	bus := NewServiceEventBus(config)
	
	// Register appropriate event handler based on service name
	switch serviceName {
	case "content-service":
		bus.RegisterServiceEventHandler(NewContentServiceServiceEventHandler())
	case "decision-service":
		bus.RegisterServiceEventHandler(NewDecisionServiceServiceEventHandler())
	case "financial-service":
		bus.RegisterServiceEventHandler(NewFinancialServiceServiceEventHandler())
	// Add more services as needed
	default:
		log.Printf("[ServiceEventBusFactory] No default handler for service: %s", serviceName)
	}
	
	return bus
}

// EventMetrics tracks event processing metrics
type EventMetrics struct {
	EventsPublished   map[string]int64 `json:"events_published"`
	EventsConsumed    map[string]int64 `json:"events_consumed"`
	EventsProcessed   map[string]int64 `json:"events_processed"`
	EventsFailed      map[string]int64 `json:"events_failed"`
	ProcessingTimes   map[string]time.Duration `json:"processing_times"`
	LastEventTime     map[string]time.Time `json:"last_event_time"`
	WorkflowsStarted  int64 `json:"workflows_started"`
	WorkflowsCompleted int64 `json:"workflows_completed"`
	WorkflowsFailed   int64 `json:"workflows_failed"`
	mu                sync.RWMutex
}

func NewEventMetrics() *EventMetrics {
	return &EventMetrics{
		EventsPublished: make(map[string]int64),
		EventsConsumed:  make(map[string]int64),
		EventsProcessed: make(map[string]int64),
		EventsFailed:    make(map[string]int64),
		ProcessingTimes: make(map[string]time.Duration),
		LastEventTime:   make(map[string]time.Time),
	}
}

func (m *EventMetrics) RecordEventPublished(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventsPublished[eventType]++
	m.LastEventTime[eventType] = time.Now()
}

func (m *EventMetrics) RecordEventConsumed(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventsConsumed[eventType]++
}

func (m *EventMetrics) RecordEventProcessed(eventType string, processingTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventsProcessed[eventType]++
	m.ProcessingTimes[eventType] = processingTime
}

func (m *EventMetrics) RecordEventFailed(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventsFailed[eventType]++
}

func (m *EventMetrics) RecordWorkflowStarted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WorkflowsStarted++
}

func (m *EventMetrics) RecordWorkflowCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WorkflowsCompleted++
}

func (m *EventMetrics) RecordWorkflowFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WorkflowsFailed++
}

func (m *EventMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Create a copy to avoid concurrent map access
	metrics := map[string]interface{}{
		"events_published":    m.copyInt64Map(m.EventsPublished),
		"events_consumed":     m.copyInt64Map(m.EventsConsumed),
		"events_processed":    m.copyInt64Map(m.EventsProcessed),
		"events_failed":       m.copyInt64Map(m.EventsFailed),
		"workflows_started":   m.WorkflowsStarted,
		"workflows_completed": m.WorkflowsCompleted,
		"workflows_failed":    m.WorkflowsFailed,
	}
	
	return metrics
}

func (m *EventMetrics) copyInt64Map(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// MetricsCollector collects metrics from the event bus
type MetricsCollector struct {
	metrics *EventMetrics
	bus     *ServiceEventBus
}

func NewMetricsCollector(bus *ServiceEventBus) *MetricsCollector {
	return &MetricsCollector{
		metrics: NewEventMetrics(),
		bus:     bus,
	}
}

func (c *MetricsCollector) GetMetrics() map[string]interface{} {
	metrics := c.metrics.GetMetrics()
	
	// Add workflow metrics
	activeWorkflows := c.bus.ListActiveWorkflows()
	metrics["active_workflows"] = len(activeWorkflows)
	
	return metrics
}

func (c *MetricsCollector) ExportMetrics() ([]byte, error) {
	metrics := c.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// HealthChecker provides health check functionality for the event system
type HealthChecker struct {
	bus     *ServiceEventBus
	metrics *MetricsCollector
}

func NewHealthChecker(bus *ServiceEventBus, metrics *MetricsCollector) *HealthChecker {
	return &HealthChecker{
		bus:     bus,
		metrics: metrics,
	}
}

func (h *HealthChecker) CheckHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{},
	}
	
	checks := health["checks"].(map[string]interface{})
	
	// Check event client health
	if h.bus.eventClient != nil {
		checks["event_client"] = map[string]interface{}{
			"status": "healthy",
			"message": "Event client is operational",
		}
	} else {
		checks["event_client"] = map[string]interface{}{
			"status": "unhealthy",
			"message": "Event client is not available",
		}
		health["status"] = "unhealthy"
	}
	
	// Check active subscriptions
	h.bus.mu.RLock()
	subscriptionCount := len(h.bus.subscriptions)
	h.bus.mu.RUnlock()
	
	checks["subscriptions"] = map[string]interface{}{
		"status": "healthy",
		"count":  subscriptionCount,
	}
	
	// Add metrics summary
	if h.metrics != nil {
		metricsData := h.metrics.GetMetrics()
		checks["metrics"] = map[string]interface{}{
			"status": "healthy",
			"data":   metricsData,
		}
	}
	
	return health
}
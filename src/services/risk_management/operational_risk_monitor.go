package risk_management

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/google/uuid"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// OperationalRiskMonitorImpl implements operational risk monitoring
type OperationalRiskMonitorImpl struct {
	riskRepo     repositories.RiskRepository
	eventRepo    repositories.EventRepository
	httpClient   *http.Client
	dependencies map[string]*entities.ServiceDependency
}

// NewOperationalRiskMonitor creates a new operational risk monitor
func NewOperationalRiskMonitor(
	riskRepo repositories.RiskRepository,
	eventRepo repositories.EventRepository,
) *OperationalRiskMonitorImpl {
	return &OperationalRiskMonitorImpl{
		riskRepo:  riskRepo,
		eventRepo: eventRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		dependencies: initializeDefaultDependencies(),
	}
}

// CheckServiceHealth checks the health of all services
func (m *OperationalRiskMonitorImpl) CheckServiceHealth(ctx context.Context) (*ServiceHealthResult, error) {
	result := &ServiceHealthResult{
		Healthy:             true,
		Services:            make([]*ServiceStatus, 0),
		FailingServices:     make([]string, 0),
		OverallAvailability: 100.0,
	}

	// Get all service dependencies
	dependencies, err := m.riskRepo.ListServiceDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list dependencies: %w", err)
	}

	healthyCount := 0
	totalCount := len(dependencies)

	for _, dep := range dependencies {
		status := m.checkServiceStatus(ctx, dep)
		result.Services = append(result.Services, status)

		if status.Status == "healthy" {
			healthyCount++
		} else {
			result.FailingServices = append(result.FailingServices, dep.Name)
			result.Healthy = false

			// Create incident for critical service failure
			if dep.Criticality == "critical" {
				m.createServiceIncident(ctx, dep, status)
			}
		}

		// Update dependency status
		if err := m.riskRepo.UpdateDependencyStatus(ctx, dep.ID, status.Status); err != nil {
			fmt.Printf("Failed to update dependency status: %v\n", err)
		}
	}

	if totalCount > 0 {
		result.OverallAvailability = float64(healthyCount) / float64(totalCount) * 100
	}

	// Create risk if availability is low
	if result.OverallAvailability < 90 {
		risk := &entities.Risk{
			ID:                uuid.New(),
			Category:          entities.RiskTypeOperational,
			Severity:          entities.RiskSeverityHigh,
			Status:            entities.RiskStatusIdentified,
			Title:             "Low Service Availability",
			Description:       fmt.Sprintf("Overall service availability: %.2f%%", result.OverallAvailability),
			Likelihood:        0.8,
			Impact:            0.7,
			MitigationActions: []string{"Investigate failing services and implement redundancy"},
			Metadata: map[string]interface{}{
				"source":            "operational_monitor",
				"affected_entities": result.FailingServices,
				"availability":      result.OverallAvailability,
			},
			IdentifiedAt:   time.Now(),
			LastAssessment: time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := m.riskRepo.CreateRisk(ctx, risk); err != nil {
			return nil, fmt.Errorf("failed to create risk: %w", err)
		}
	}

	return result, nil
}

// MonitorDependencies monitors all service dependencies
func (m *OperationalRiskMonitorImpl) MonitorDependencies(ctx context.Context) ([]*DependencyStatus, error) {
	dependencies, err := m.riskRepo.ListServiceDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list dependencies: %w", err)
	}

	statuses := make([]*DependencyStatus, 0, len(dependencies))

	for _, dep := range dependencies {
		status := &DependencyStatus{
			DependencyID:    dep.ID,
			Name:            dep.Name,
			Type:            dep.Type,
			Provider:        dep.Provider,
			Status:          "unknown",
			ResponseTime:    0,
			LastSuccessful:  dep.LastHealthCheck,
			Criticality:     dep.Criticality,
			HasFallback:     dep.FallbackService != "",
			FallbackActive:  false,
		}

		// Check health endpoint if available
		if dep.HealthEndpoint != "" {
			healthy, responseTime := m.checkHealthEndpoint(dep.HealthEndpoint)
			if healthy {
				status.Status = "healthy"
				status.ResponseTime = responseTime
			} else {
				status.Status = "unhealthy"
				
				// Check if fallback should be activated
				if dep.FallbackService != "" && time.Since(dep.LastHealthCheck) > time.Duration(dep.MaxDowntime)*time.Minute {
					status.FallbackActive = true
					m.activateFallback(ctx, dep)
				}
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// AssessCapacityRisk assesses system capacity risks
func (m *OperationalRiskMonitorImpl) AssessCapacityRisk(ctx context.Context) (*CapacityRiskResult, error) {
	result := &CapacityRiskResult{
		RiskLevel:    "low",
		Utilization:  make(map[string]float64),
		Projections:  make(map[string]*CapacityProjection),
		Warnings:     make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Check memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	memoryUsage := float64(memStats.Alloc) / float64(memStats.Sys) * 100
	result.Utilization["memory"] = memoryUsage
	
	if memoryUsage > 80 {
		result.RiskLevel = "high"
		result.Warnings = append(result.Warnings, fmt.Sprintf("High memory usage: %.2f%%", memoryUsage))
		result.Recommendations = append(result.Recommendations, "Consider scaling up memory resources")
	}

	// Check goroutine count
	goroutineCount := runtime.NumGoroutine()
	result.Utilization["goroutines"] = float64(goroutineCount)
	
	if goroutineCount > 10000 {
		result.RiskLevel = "medium"
		result.Warnings = append(result.Warnings, fmt.Sprintf("High goroutine count: %d", goroutineCount))
		result.Recommendations = append(result.Recommendations, "Review for potential goroutine leaks")
	}

	// Check CPU usage (simplified)
	cpuCount := runtime.NumCPU()
	result.Utilization["cpu_cores"] = float64(cpuCount)

	// Project future capacity needs (simplified)
	result.Projections["memory"] = &CapacityProjection{
		TimeHorizon:      "24h",
		ProjectedUsage:   memoryUsage * 1.1, // 10% growth projection
		RequiredCapacity: 100.0,
		RecommendedAction: "Monitor",
	}

	return result, nil
}

// PredictFailures predicts potential service failures
func (m *OperationalRiskMonitorImpl) PredictFailures(ctx context.Context) ([]*FailurePrediction, error) {
	predictions := make([]*FailurePrediction, 0)

	// Get recent incidents
	incidents, err := m.riskRepo.ListIncidents(ctx, repositories.IncidentFilters{
		StartDate: time.Now().Add(-7 * 24 * time.Hour),
		EndDate:   time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	// Analyze incident patterns
	serviceIncidentCount := make(map[string]int)
	for _, incident := range incidents {
		if service, ok := incident.Metadata["affected_service"].(string); ok {
			serviceIncidentCount[service]++
		}
	}

	// Predict failures based on patterns
	for service, count := range serviceIncidentCount {
		if count >= 3 { // 3 or more incidents in a week
			prediction := &FailurePrediction{
				Service:        service,
				Probability:    float64(count) / 7.0, // Simplified probability
				TimeFrame:      "next 24 hours",
				RiskFactors:    []string{fmt.Sprintf("%d incidents in last 7 days", count)},
				PreventiveActions: []string{
					"Increase monitoring frequency",
					"Review service logs",
					"Prepare fallback activation",
				},
			}
			predictions = append(predictions, prediction)
		}
	}

	// Check dependency health trends
	dependencies, err := m.riskRepo.ListServiceDependencies(ctx)
	if err == nil {
		for _, dep := range dependencies {
			if time.Since(dep.LastHealthCheck) > 24*time.Hour {
				prediction := &FailurePrediction{
					Service:     dep.Name,
					Probability: 0.3,
					TimeFrame:   "next 48 hours",
					RiskFactors: []string{"No health check in 24+ hours"},
					PreventiveActions: []string{
						"Verify service connectivity",
						"Update health check configuration",
					},
				}
				predictions = append(predictions, prediction)
			}
		}
	}

	return predictions, nil
}

// Helper methods

func (m *OperationalRiskMonitorImpl) checkServiceStatus(ctx context.Context, dep *entities.ServiceDependency) *ServiceStatus {
	status := &ServiceStatus{
		Name:         dep.Name,
		Status:       "unknown",
		ResponseTime: 0,
		Uptime:       99.9, // Default assumption
		LastCheck:    time.Now().Format(time.RFC3339),
	}

	if dep.HealthEndpoint != "" {
		healthy, responseTime := m.checkHealthEndpoint(dep.HealthEndpoint)
		if healthy {
			status.Status = "healthy"
			status.ResponseTime = responseTime
		} else {
			status.Status = "unhealthy"
			status.Uptime = 0.0
		}
	}

	return status
}

func (m *OperationalRiskMonitorImpl) checkHealthEndpoint(endpoint string) (bool, int) {
	start := time.Now()
	
	resp, err := m.httpClient.Get(endpoint)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	responseTime := int(time.Since(start).Milliseconds())

	// Consider 2xx and 3xx as healthy
	return resp.StatusCode >= 200 && resp.StatusCode < 400, responseTime
}

func (m *OperationalRiskMonitorImpl) createServiceIncident(ctx context.Context, dep *entities.ServiceDependency, status *ServiceStatus) {
	incident := &entities.Incident{
		ID:          uuid.New(),
		Severity:    entities.RiskSeverityHigh,
		Status:      entities.IncidentStatusOpen,
		Title:       fmt.Sprintf("%s Service Failure", dep.Name),
		Description: fmt.Sprintf("Critical service %s is not responding", dep.Name),
		Category:    entities.RiskCategoryOperational,
		Source:      "operational_monitor",
		Metadata: map[string]interface{}{
			"type":             "service_failure",
			"affected_service": dep.Name,
			"impact":           "Service functionality degraded or unavailable",
		},
		ActionsTaken: []entities.IncidentAction{
			{
				ID:          uuid.New().String(),
				Type:        "notification",
				Description: "Alert sent to monitoring dashboard",
				Status:      "completed",
				Result:      "success",
				ExecutedAt:  time.Now(),
				ExecutedBy:  "system",
			},
		},
		OccurredAt: time.Now(),
		DetectedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := m.riskRepo.CreateIncident(ctx, incident); err != nil {
		fmt.Printf("Failed to create service incident: %v\n", err)
	}

	// Emit event
	event := &events.IncidentDetectedEvent{
		BaseEvent: events.BaseEvent{
			EventID:   generateEventID(),
			EventType: events.IncidentDetected,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "operational_monitor",
			},
		},
		IncidentID:      incident.ID.String(),
		Type:            fmt.Sprintf("%v", incident.Metadata["type"]),
		Severity:        incident.Severity,
		Title:           incident.Title,
		AffectedService: fmt.Sprintf("%v", incident.Metadata["affected_service"]),
		Impact:          fmt.Sprintf("%v", incident.Metadata["impact"]),
		AutoResponse:    true,
	}

	if err := m.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Failed to save incident detected event: %v\n", err)
	}
}

func (m *OperationalRiskMonitorImpl) activateFallback(ctx context.Context, dep *entities.ServiceDependency) {
	// Log fallback activation
	event := &events.ServiceDependencyFailureEvent{
		BaseEvent: events.BaseEvent{
			EventID:   generateEventID(),
			EventType: events.ServiceDependencyFailure,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "operational_monitor",
			},
		},
		DependencyID:   dep.ID,
		ServiceName:    dep.Name,
		Provider:       dep.Provider,
		Criticality:    dep.Criticality,
		FailureType:    "health_check_failure",
		LastSuccessful: dep.LastHealthCheck,
		FallbackUsed:   true,
	}

	if err := m.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Failed to save service dependency failure event: %v\n", err)
	}
}

func initializeDefaultDependencies() map[string]*entities.ServiceDependency {
	return map[string]*entities.ServiceDependency{
		"openai": {
			ID:              "dep_openai",
			Name:            "OpenAI API",
			Type:            "llm",
			Provider:        "OpenAI",
			Criticality:     "critical",
			HealthEndpoint:  "https://api.openai.com/v1/models",
			FallbackService: "anthropic",
			MaxDowntime:     5,
			Status:          "unknown",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		"anthropic": {
			ID:              "dep_anthropic",
			Name:            "Anthropic Claude API",
			Type:            "llm",
			Provider:        "Anthropic",
			Criticality:     "critical",
			HealthEndpoint:  "https://api.anthropic.com/v1/models",
			FallbackService: "openai",
			MaxDowntime:     5,
			Status:          "unknown",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		"stripe": {
			ID:              "dep_stripe",
			Name:            "Stripe Payment API",
			Type:            "payment",
			Provider:        "Stripe",
			Criticality:     "critical",
			HealthEndpoint:  "https://api.stripe.com/v1/",
			FallbackService: "",
			MaxDowntime:     10,
			Status:          "unknown",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		"postgresql": {
			ID:              "dep_postgresql",
			Name:            "PostgreSQL Database",
			Type:            "database",
			Provider:        "Internal",
			Criticality:     "critical",
			HealthEndpoint:  "", // Would use internal health check
			FallbackService: "",
			MaxDowntime:     1,
			Status:          "unknown",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		"redis": {
			ID:              "dep_redis",
			Name:            "Redis Cache",
			Type:            "cache",
			Provider:        "Internal",
			Criticality:     "high",
			HealthEndpoint:  "", // Would use internal health check
			FallbackService: "",
			MaxDowntime:     30,
			Status:          "unknown",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}
}


func generateActionID() string {
	return fmt.Sprintf("act_%d", time.Now().UnixNano())
}

func generateEventID() uuid.UUID {
	return uuid.New()
}

// Additional types

type DependencyStatus struct {
	DependencyID   string    `json:"dependency_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Provider       string    `json:"provider"`
	Status         string    `json:"status"`
	ResponseTime   int       `json:"response_time"`
	LastSuccessful time.Time `json:"last_successful"`
	Criticality    string    `json:"criticality"`
	HasFallback    bool      `json:"has_fallback"`
	FallbackActive bool      `json:"fallback_active"`
}

type CapacityRiskResult struct {
	RiskLevel       string                      `json:"risk_level"`
	Utilization     map[string]float64          `json:"utilization"`
	Projections     map[string]*CapacityProjection `json:"projections"`
	Warnings        []string                    `json:"warnings"`
	Recommendations []string                    `json:"recommendations"`
}

type CapacityProjection struct {
	TimeHorizon       string  `json:"time_horizon"`
	ProjectedUsage    float64 `json:"projected_usage"`
	RequiredCapacity  float64 `json:"required_capacity"`
	RecommendedAction string  `json:"recommended_action"`
}

type FailurePrediction struct {
	Service           string   `json:"service"`
	Probability       float64  `json:"probability"`
	TimeFrame         string   `json:"time_frame"`
	RiskFactors       []string `json:"risk_factors"`
	PreventiveActions []string `json:"preventive_actions"`
}
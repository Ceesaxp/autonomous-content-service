package entities

import (
	"time"

	"github.com/google/uuid"
)

// CapabilityType represents the type of system capability
type CapabilityType string

const (
	CapabilityTypeContentGeneration CapabilityType = "ContentGeneration"
	CapabilityTypeResearch          CapabilityType = "Research"
	CapabilityTypeQualityAssurance  CapabilityType = "QualityAssurance"
	CapabilityTypePaymentProcessing CapabilityType = "PaymentProcessing"
	CapabilityTypeClientManagement  CapabilityType = "ClientManagement"
	CapabilityTypeProjectManagement CapabilityType = "ProjectManagement"
	CapabilityTypeAnalytics         CapabilityType = "Analytics"
	CapabilityTypeIntegration       CapabilityType = "Integration"
)

// CapabilityStatus represents the status of a capability
type CapabilityStatus string

const (
	CapabilityStatusActive      CapabilityStatus = "Active"
	CapabilityStatusInactive    CapabilityStatus = "Inactive"
	CapabilityStatusMaintenance CapabilityStatus = "Maintenance"
	CapabilityStatusUpgrading   CapabilityStatus = "Upgrading"
	CapabilityStatusDeprecated  CapabilityStatus = "Deprecated"
)

// SystemCapability represents a capability of the autonomous system
type SystemCapability struct {
	CapabilityID  uuid.UUID        `json:"capabilityId"`
	Type          CapabilityType   `json:"type"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Version       string           `json:"version"`
	Status        CapabilityStatus `json:"status"`
	Configuration map[string]interface{} `json:"configuration"`
	Metrics       CapabilityMetrics      `json:"metrics"`
	Dependencies  []uuid.UUID            `json:"dependencies"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	LastUsed      *time.Time             `json:"lastUsed,omitempty"`
}

// CapabilityMetrics represents performance metrics for a capability
type CapabilityMetrics struct {
	UsageCount       int64         `json:"usageCount"`
	AverageLatency   time.Duration `json:"averageLatency"`
	SuccessRate      float64       `json:"successRate"`
	ErrorRate        float64       `json:"errorRate"`
	LastError        *string       `json:"lastError,omitempty"`
	LastErrorTime    *time.Time    `json:"lastErrorTime,omitempty"`
	PerformanceScore float64       `json:"performanceScore"`
}

// NewSystemCapability creates a new system capability
func NewSystemCapability(
	capabilityType CapabilityType,
	name, description, version string,
) *SystemCapability {
	return &SystemCapability{
		CapabilityID:  uuid.New(),
		Type:          capabilityType,
		Name:          name,
		Description:   description,
		Version:       version,
		Status:        CapabilityStatusActive,
		Configuration: make(map[string]interface{}),
		Metrics: CapabilityMetrics{
			UsageCount:       0,
			AverageLatency:   0,
			SuccessRate:      100.0,
			ErrorRate:        0.0,
			PerformanceScore: 100.0,
		},
		Dependencies: []uuid.UUID{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// UpdateStatus updates the capability status
func (c *SystemCapability) UpdateStatus(status CapabilityStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}

// SetConfiguration sets configuration parameters
func (c *SystemCapability) SetConfiguration(key string, value interface{}) {
	c.Configuration[key] = value
	c.UpdatedAt = time.Now()
}

// RecordUsage records usage of the capability
func (c *SystemCapability) RecordUsage(latency time.Duration, success bool) {
	c.Metrics.UsageCount++
	
	// Update average latency
	if c.Metrics.UsageCount == 1 {
		c.Metrics.AverageLatency = latency
	} else {
		// Simple running average
		totalLatency := c.Metrics.AverageLatency * time.Duration(c.Metrics.UsageCount-1)
		c.Metrics.AverageLatency = (totalLatency + latency) / time.Duration(c.Metrics.UsageCount)
	}
	
	// Update success/error rates
	if success {
		c.Metrics.SuccessRate = (c.Metrics.SuccessRate*float64(c.Metrics.UsageCount-1) + 100.0) / float64(c.Metrics.UsageCount)
	} else {
		c.Metrics.SuccessRate = (c.Metrics.SuccessRate * float64(c.Metrics.UsageCount-1)) / float64(c.Metrics.UsageCount)
	}
	c.Metrics.ErrorRate = 100.0 - c.Metrics.SuccessRate
	
	// Update performance score (combination of success rate and latency)
	latencyScore := 100.0
	if latency > 5*time.Second {
		latencyScore = 50.0
	} else if latency > 2*time.Second {
		latencyScore = 75.0
	} else if latency > 1*time.Second {
		latencyScore = 90.0
	}
	
	c.Metrics.PerformanceScore = (c.Metrics.SuccessRate + latencyScore) / 2.0
	
	now := time.Now()
	c.LastUsed = &now
	c.UpdatedAt = time.Now()
}

// RecordError records an error for the capability
func (c *SystemCapability) RecordError(errorMessage string) {
	c.Metrics.LastError = &errorMessage
	now := time.Now()
	c.Metrics.LastErrorTime = &now
	c.UpdatedAt = time.Now()
}

// AddDependency adds a dependency to another capability
func (c *SystemCapability) AddDependency(dependencyID uuid.UUID) {
	c.Dependencies = append(c.Dependencies, dependencyID)
	c.UpdatedAt = time.Now()
}

// IsActive returns true if the capability is active
func (c *SystemCapability) IsActive() bool {
	return c.Status == CapabilityStatusActive
}

// IsHealthy returns true if the capability is performing well
func (c *SystemCapability) IsHealthy() bool {
	return c.Metrics.PerformanceScore >= 80.0 && c.Metrics.ErrorRate <= 5.0
}
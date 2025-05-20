package events

import (
	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// System event types
const (
	EventTypeCapabilityPerformanceDeclined EventType = "CapabilityPerformanceDeclined"
	EventTypeCapabilityUpgraded           EventType = "CapabilityUpgraded"
	EventTypeAnomalyDetected              EventType = "AnomalyDetected"
)

// CapabilityPerformanceDeclinedEvent is triggered when performance metrics for a capability fall below threshold
type CapabilityPerformanceDeclinedEvent struct {
	BaseEvent
	CapabilityID     uuid.UUID              `json:"capabilityId"`
	CapabilityType   entities.CapabilityType `json:"capabilityType"`
	AffectedMetrics  []string               `json:"affectedMetrics"`
	CurrentValues    map[string]float64     `json:"currentValues"`
	ThresholdValues  map[string]float64     `json:"thresholdValues"`
}

// NewCapabilityPerformanceDeclinedEvent creates a new CapabilityPerformanceDeclinedEvent
func NewCapabilityPerformanceDeclinedEvent(
	capability *entities.SystemCapability,
	affectedMetrics []string,
	currentValues, thresholdValues map[string]float64,
) CapabilityPerformanceDeclinedEvent {
	return CapabilityPerformanceDeclinedEvent{
		BaseEvent:       NewBaseEvent(EventTypeCapabilityPerformanceDeclined, capability.CapabilityID),
		CapabilityID:    capability.CapabilityID,
		CapabilityType:  capability.Type,
		AffectedMetrics: affectedMetrics,
		CurrentValues:   currentValues,
		ThresholdValues: thresholdValues,
	}
}

// CapabilityUpgradedEvent is triggered when a system capability is enhanced
type CapabilityUpgradedEvent struct {
	BaseEvent
	CapabilityID      uuid.UUID              `json:"capabilityId"`
	CapabilityType    entities.CapabilityType `json:"capabilityType"`
	UpgradeDetails    string                 `json:"upgradeDetails"`
	PerformanceGains  map[string]interface{} `json:"performanceGains"`
}

// NewCapabilityUpgradedEvent creates a new CapabilityUpgradedEvent
func NewCapabilityUpgradedEvent(
	capability *entities.SystemCapability,
	upgradeDetails string,
	performanceGains map[string]interface{},
) CapabilityUpgradedEvent {
	return CapabilityUpgradedEvent{
		BaseEvent:       NewBaseEvent(EventTypeCapabilityUpgraded, capability.CapabilityID),
		CapabilityID:    capability.CapabilityID,
		CapabilityType:  capability.Type,
		UpgradeDetails:  upgradeDetails,
		PerformanceGains: performanceGains,
	}
}

// AnomalyDetectedEvent is triggered when unusual patterns are detected in system behavior
type AnomalyDetectedEvent struct {
	BaseEvent
	AnomalyType       string                 `json:"anomalyType"`
	AffectedComponents []string               `json:"affectedComponents"`
	AnomalyData       map[string]interface{} `json:"anomalyData"`
	Severity          string                 `json:"severity"` // "Low", "Medium", "High", "Critical"
}

// NewAnomalyDetectedEvent creates a new AnomalyDetectedEvent
func NewAnomalyDetectedEvent(
	anomalyType string,
	affectedComponents []string,
	anomalyData map[string]interface{},
	severity string,
) AnomalyDetectedEvent {
	// Generate a UUID for the anomaly - this is different since it's not tied to a specific entity
	anomalyID := uuid.New()

	return AnomalyDetectedEvent{
		BaseEvent:          NewBaseEvent(EventTypeAnomalyDetected, anomalyID),
		AnomalyType:        anomalyType,
		AffectedComponents: affectedComponents,
		AnomalyData:        anomalyData,
		Severity:           severity,
	}
}

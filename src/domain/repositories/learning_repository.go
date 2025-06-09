package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// LearningRepository manages learning artifacts and knowledge graph
type LearningRepository interface {
	// Learning Artifacts - Graph-based storage
	CreateArtifact(ctx context.Context, artifact *entities.LearningArtifact) error
	GetArtifact(ctx context.Context, id string) (*entities.LearningArtifact, error)
	UpdateArtifact(ctx context.Context, artifact *entities.LearningArtifact) error
	DeleteArtifact(ctx context.Context, id string) error
	
	// Graph operations for knowledge relationships
	LinkArtifacts(ctx context.Context, sourceID, targetID string, relationType string) error
	UnlinkArtifacts(ctx context.Context, sourceID, targetID string) error
	GetRelatedArtifacts(ctx context.Context, artifactID string, depth int) ([]*entities.LearningArtifact, error)
	GetArtifactsByPattern(ctx context.Context, pattern string) ([]*entities.LearningArtifact, error)
	
	// Knowledge retrieval
	SearchArtifacts(ctx context.Context, query string, filters map[string]interface{}) ([]*entities.LearningArtifact, error)
	GetArtifactsByType(ctx context.Context, artifactType entities.LearningType) ([]*entities.LearningArtifact, error)
	GetArtifactsByCategory(ctx context.Context, category string) ([]*entities.LearningArtifact, error)
	GetActiveArtifacts(ctx context.Context) ([]*entities.LearningArtifact, error)
	GetArtifactsRequiringVerification(ctx context.Context, before time.Time) ([]*entities.LearningArtifact, error)
	
	// Usage tracking
	IncrementUsage(ctx context.Context, artifactID string, success bool) error
	GetMostUsedArtifacts(ctx context.Context, limit int) ([]*entities.LearningArtifact, error)
	GetHighImpactArtifacts(ctx context.Context, minImpact float64) ([]*entities.LearningArtifact, error)
}

// MetricsRepository manages performance metrics
type MetricsRepository interface {
	// Metric storage
	RecordMetric(ctx context.Context, metric *entities.SystemPerformanceMetric) error
	GetMetrics(ctx context.Context, component, metricName string, from, to time.Time) ([]*entities.SystemPerformanceMetric, error)
	GetLatestMetric(ctx context.Context, component, metricName string) (*entities.SystemPerformanceMetric, error)
	
	// Aggregations
	GetAggregatedMetrics(ctx context.Context, component, metricName string, aggregation entities.AggregationType, period string) (float64, error)
	GetMetricTrends(ctx context.Context, component, metricName string, periods int) ([]float64, error)
	GetComponentMetrics(ctx context.Context, component string) (map[string]float64, error)
	
	// Anomaly detection
	GetMetricAnomalies(ctx context.Context, component, metricName string, threshold float64) ([]*entities.SystemPerformanceMetric, error)
	GetMetricBaseline(ctx context.Context, component, metricName string, days int) (mean, stddev float64, err error)
}

// ExperimentRepository manages A/B tests and experiments
type ExperimentRepository interface {
	// Experiment lifecycle
	CreateExperiment(ctx context.Context, experiment *entities.Experiment) error
	GetExperiment(ctx context.Context, id string) (*entities.Experiment, error)
	UpdateExperiment(ctx context.Context, experiment *entities.Experiment) error
	ListActiveExperiments(ctx context.Context) ([]*entities.Experiment, error)
	
	// Experiment assignment
	AssignToVariant(ctx context.Context, experimentID, entityID, variantID string) error
	GetEntityVariant(ctx context.Context, experimentID, entityID string) (variantID string, err error)
	RecordConversion(ctx context.Context, experimentID, variantID, entityID string, metrics map[string]float64) error
	
	// Results and analysis
	GetExperimentResults(ctx context.Context, experimentID string) (*entities.ExperimentResults, error)
	CalculateStatisticalSignificance(ctx context.Context, experimentID string) (pValue float64, significant bool, err error)
	GetWinningVariant(ctx context.Context, experimentID string) (variantID string, confidence float64, err error)
	
	// Historical experiments
	GetCompletedExperiments(ctx context.Context, component string, limit int) ([]*entities.Experiment, error)
	GetSuccessfulOptimizations(ctx context.Context, minLift float64) ([]*entities.Experiment, error)
}

// CapabilityRepository manages capability gaps and acquisitions
type CapabilityRepository interface {
	// Gap management
	CreateCapabilityGap(ctx context.Context, gap *entities.CapabilityGap) error
	GetCapabilityGap(ctx context.Context, id string) (*entities.CapabilityGap, error)
	UpdateCapabilityGap(ctx context.Context, gap *entities.CapabilityGap) error
	ListCapabilityGaps(ctx context.Context, status entities.GapStatus) ([]*entities.CapabilityGap, error)
	
	// Gap analysis
	FindSimilarGaps(ctx context.Context, description string) ([]*entities.CapabilityGap, error)
	GetGapsByType(ctx context.Context, capType entities.CapabilityType) ([]*entities.CapabilityGap, error)
	GetHighPriorityGaps(ctx context.Context, minPriority float64) ([]*entities.CapabilityGap, error)
	IncrementGapFrequency(ctx context.Context, gapID string) error
	
	// Resolution tracking
	RecordGapResolution(ctx context.Context, gapID string, resolution *entities.GapResolution) error
	GetResolvedGaps(ctx context.Context, method string) ([]*entities.CapabilityGap, error)
	GetAverageResolutionTime(ctx context.Context, capType entities.CapabilityType) (time.Duration, error)
}

// PromptRepository manages prompt optimizations
type PromptRepository interface {
	// Prompt management
	CreatePromptOptimization(ctx context.Context, optimization *entities.PromptOptimization) error
	GetPromptOptimization(ctx context.Context, id string) (*entities.PromptOptimization, error)
	UpdatePromptOptimization(ctx context.Context, optimization *entities.PromptOptimization) error
	GetActivePrompts(ctx context.Context, component string) ([]*entities.PromptOptimization, error)
	
	// Version history
	GetPromptHistory(ctx context.Context, component string) ([]*entities.PromptOptimization, error)
	RollbackPrompt(ctx context.Context, component string, version string) error
	GetBestPerformingPrompt(ctx context.Context, component string, metric string) (*entities.PromptOptimization, error)
	
	// LLM provider optimization
	GetProviderPerformance(ctx context.Context, component string) (map[string]map[string]float64, error)
	GetOptimalProvider(ctx context.Context, component string, requirements map[string]float64) (provider string, err error)
}
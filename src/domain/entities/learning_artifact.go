package entities

import (
	"time"
)

// LearningArtifact represents a piece of knowledge learned by the system
type LearningArtifact struct {
	ID               string                 `json:"id"`
	Type             LearningType           `json:"type"`
	Category         string                 `json:"category"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Source           LearningSource         `json:"source"`
	SourceID         string                 `json:"source_id"`
	Evidence         []Evidence             `json:"evidence"`
	Confidence       float64                `json:"confidence"`
	ImpactScore      float64                `json:"impact_score"`
	Tags             []string               `json:"tags"`
	RelatedArtifacts []string               `json:"related_artifacts"`
	Prerequisites    []string               `json:"prerequisites"`
	Metadata         map[string]interface{} `json:"metadata"`
	ValidFrom        time.Time              `json:"valid_from"`
	ValidUntil       *time.Time             `json:"valid_until"`
	VerificationDate *time.Time             `json:"verification_date"`
	UsageCount       int                    `json:"usage_count"`
	SuccessRate      float64                `json:"success_rate"`
	Status           ArtifactStatus         `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// LearningType defines the type of learning
type LearningType string

const (
	LearningTypePattern       LearningType = "pattern"
	LearningTypeRule          LearningType = "rule"
	LearningTypeOptimization  LearningType = "optimization"
	LearningTypeCapability    LearningType = "capability"
	LearningTypeConstraint    LearningType = "constraint"
	LearningTypeRelationship  LearningType = "relationship"
	LearningTypeHeuristic     LearningType = "heuristic"
	LearningTypeException     LearningType = "exception"
)

// LearningSource defines where the learning came from
type LearningSource string

const (
	SourceProjectAnalysis   LearningSource = "project_analysis"
	SourceClientFeedback    LearningSource = "client_feedback"
	SourceSystemMonitoring  LearningSource = "system_monitoring"
	SourceExperiment        LearningSource = "experiment"
	SourceCompetitorAnalysis LearningSource = "competitor_analysis"
	SourceManualEntry       LearningSource = "manual_entry"
	SourceAPIDiscovery      LearningSource = "api_discovery"
	SourceErrorAnalysis     LearningSource = "error_analysis"
)

// Evidence supports a learning artifact
type Evidence struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Strength    float64                `json:"strength"`
}

// ArtifactStatus represents the lifecycle state
type ArtifactStatus string

const (
	ArtifactStatusDraft      ArtifactStatus = "draft"
	ArtifactStatusActive     ArtifactStatus = "active"
	ArtifactStatusTesting    ArtifactStatus = "testing"
	ArtifactStatusDeprecated ArtifactStatus = "deprecated"
	ArtifactStatusArchived   ArtifactStatus = "archived"
)

// SystemPerformanceMetric represents a system performance measurement
type SystemPerformanceMetric struct {
	ID          string                 `json:"id"`
	Component   string                 `json:"component"`
	MetricName  string                 `json:"metric_name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	Aggregation AggregationType        `json:"aggregation"`
	Period      string                 `json:"period"`
	Tags        []string               `json:"tags"`
}

// AggregationType defines how metrics are aggregated
type AggregationType string

const (
	AggregationAverage AggregationType = "average"
	AggregationSum     AggregationType = "sum"
	AggregationMin     AggregationType = "min"
	AggregationMax     AggregationType = "max"
	AggregationCount   AggregationType = "count"
	AggregationP95     AggregationType = "p95"
	AggregationP99     AggregationType = "p99"
)

// Experiment represents an A/B test or experiment
type Experiment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Hypothesis      string                 `json:"hypothesis"`
	Type            ExperimentType         `json:"type"`
	Status          ExperimentStatus       `json:"status"`
	Variants        []ExperimentVariant    `json:"variants"`
	MetricsTracked  []string               `json:"metrics_tracked"`
	SuccessCriteria SuccessCriteria        `json:"success_criteria"`
	SampleSize      int                    `json:"sample_size"`
	CurrentSample   int                    `json:"current_sample"`
	Results         *ExperimentResults     `json:"results,omitempty"`
	StartDate       time.Time              `json:"start_date"`
	EndDate         *time.Time             `json:"end_date"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ExperimentType defines the type of experiment
type ExperimentType string

const (
	ExperimentTypeAB          ExperimentType = "ab_test"
	ExperimentTypeMultivariate ExperimentType = "multivariate"
	ExperimentTypeBandit      ExperimentType = "bandit"
	ExperimentTypeFeatureFlag ExperimentType = "feature_flag"
)


// ExperimentVariant represents a variant in an experiment
type ExperimentVariant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Config      map[string]interface{} `json:"config"`
	IsControl   bool                   `json:"is_control"`
}

// SuccessCriteria defines what makes an experiment successful
type SuccessCriteria struct {
	PrimaryMetric     string  `json:"primary_metric"`
	MinimumEffect     float64 `json:"minimum_effect"`
	ConfidenceLevel   float64 `json:"confidence_level"`
	MinimumSampleSize int     `json:"minimum_sample_size"`
}


// VariantResult contains results for a specific variant
type VariantResult struct {
	SampleSize  int                    `json:"sample_size"`
	Conversions int                    `json:"conversions"`
	Metrics     map[string]float64     `json:"metrics"`
	Confidence  ConfidenceInterval     `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MetricResult contains results for a specific metric
type MetricResult struct {
	Control   float64            `json:"control"`
	Variants  map[string]float64 `json:"variants"`
	BestValue float64            `json:"best_value"`
	Lift      float64            `json:"lift"`
}

// ConfidenceInterval represents a statistical confidence interval
type ConfidenceInterval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Level float64 `json:"level"`
}

// CapabilityGap represents a missing capability identified by the system
type CapabilityGap struct {
	ID               string              `json:"id"`
	Type             string              `json:"type"`
	Description      string              `json:"description"`
	RequestedBy      []string            `json:"requested_by"`
	Frequency        int                 `json:"frequency"`
	Priority         float64             `json:"priority"`
	EstimatedImpact  float64             `json:"estimated_impact"`
	EstimatedEffort  float64             `json:"estimated_effort"`
	PotentialSources []CapabilitySource  `json:"potential_sources"`
	Status           GapStatus           `json:"status"`
	Resolution       *GapResolution      `json:"resolution,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// Capability gap types
const (
	CapabilityGapTypeAPI        = "api_integration"
	CapabilityGapTypeContent    = "content_type"
	CapabilityGapTypeLanguage   = "language"
	CapabilityGapTypeIndustry   = "industry_knowledge"
	CapabilityGapTypeTool       = "tool"
	CapabilityGapTypeSkill      = "skill"
	CapabilityGapTypeData       = "data_source"
)

// CapabilitySource represents a potential source for acquiring a capability
type CapabilitySource struct {
	Type        string                 `json:"type"`
	Provider    string                 `json:"provider"`
	Cost        float64                `json:"cost"`
	TimeToAcquire string               `json:"time_to_acquire"`
	Confidence  float64                `json:"confidence"`
	Requirements []string              `json:"requirements"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GapStatus represents the status of a capability gap
type GapStatus string

const (
	GapStatusIdentified GapStatus = "identified"
	GapStatusAnalyzing  GapStatus = "analyzing"
	GapStatusApproved   GapStatus = "approved"
	GapStatusAcquiring  GapStatus = "acquiring"
	GapStatusResolved   GapStatus = "resolved"
	GapStatusDismissed  GapStatus = "dismissed"
)

// GapResolution describes how a gap was resolved
type GapResolution struct {
	Method       string                 `json:"method"`
	Source       string                 `json:"source"`
	Cost         float64                `json:"cost"`
	TimeToResolve string                `json:"time_to_resolve"`
	Effectiveness float64               `json:"effectiveness"`
	Details      map[string]interface{} `json:"details"`
	ResolvedAt   time.Time              `json:"resolved_at"`
}

// PromptOptimization represents an optimized prompt configuration
type PromptOptimization struct {
	ID              string                 `json:"id"`
	Component       string                 `json:"component"`
	OriginalPrompt  string                 `json:"original_prompt"`
	OptimizedPrompt string                 `json:"optimized_prompt"`
	LLMProvider     string                 `json:"llm_provider"`
	ModelVersion    string                 `json:"model_version"`
	Improvements    map[string]float64     `json:"improvements"`
	TestResults     []PromptTestResult     `json:"test_results"`
	Status          OptimizationStatus     `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	ActivatedAt     *time.Time             `json:"activated_at"`
}

// PromptTestResult contains results from testing a prompt
type PromptTestResult struct {
	TestCase    string             `json:"test_case"`
	Score       float64            `json:"score"`
	Metrics     map[string]float64 `json:"metrics"`
	SampleOutput string            `json:"sample_output"`
	Timestamp   time.Time          `json:"timestamp"`
}

// OptimizationStatus represents the status of an optimization
type OptimizationStatus string

const (
	OptimizationStatusTesting   OptimizationStatus = "testing"
	OptimizationStatusActive    OptimizationStatus = "active"
	OptimizationStatusRollback  OptimizationStatus = "rollback"
	OptimizationStatusArchived  OptimizationStatus = "archived"
)
package self_improvement

import (
	"context"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// SelfImprovementService orchestrates the continuous learning and improvement
type SelfImprovementService interface {
	// Performance monitoring
	CollectMetrics(ctx context.Context) error
	AnalyzePerformance(ctx context.Context, component string, period string) (*PerformanceAnalysis, error)
	DetectAnomalies(ctx context.Context) ([]*Anomaly, error)
	
	// Learning and knowledge management
	LearnFromProject(ctx context.Context, projectID string) ([]*entities.LearningArtifact, error)
	LearnFromFeedback(ctx context.Context, feedbackID string) (*entities.LearningArtifact, error)
	LearnFromError(ctx context.Context, errorData ErrorData) (*entities.LearningArtifact, error)
	SynthesizeLearnings(ctx context.Context, period string) ([]*entities.LearningArtifact, error)
	
	// Capability management
	IdentifyCapabilityGaps(ctx context.Context) ([]*entities.CapabilityGap, error)
	PrioritizeGaps(ctx context.Context, gaps []*entities.CapabilityGap) ([]*entities.CapabilityGap, error)
	AcquireCapability(ctx context.Context, gapID string) error
	
	// Experimentation
	ProposeExperiment(ctx context.Context, hypothesis string, component string) (*entities.Experiment, error)
	RunExperiment(ctx context.Context, experimentID string) error
	EvaluateExperiment(ctx context.Context, experimentID string) (*ExperimentEvaluation, error)
	ApplyExperimentResults(ctx context.Context, experimentID string) error
	
	// Optimization
	OptimizePrompts(ctx context.Context, component string) ([]*entities.PromptOptimization, error)
	SelectOptimalLLM(ctx context.Context, task string, requirements map[string]float64) (string, error)
	OptimizeWorkflow(ctx context.Context, workflow string) (*WorkflowOptimization, error)
	
	// ROI and prioritization
	CalculateImprovementROI(ctx context.Context, improvement Improvement) (float64, error)
	PrioritizeImprovements(ctx context.Context) ([]*PrioritizedImprovement, error)
	
	// Competitive intelligence
	AnalyzeCompetitors(ctx context.Context) ([]*CompetitorInsight, error)
	IdentifyMarketGaps(ctx context.Context) ([]*MarketOpportunity, error)
}

// MetricsCollector gathers performance metrics from all system components
type MetricsCollector interface {
	// Component-specific collectors
	CollectContentMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error)
	CollectDecisionMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error)
	CollectFinancialMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error)
	CollectClientMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error)
	CollectSystemMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error)
	
	// Aggregation
	AggregateMetrics(ctx context.Context, metrics []*entities.SystemPerformanceMetric) (*MetricsSummary, error)
}

// LearningEngine processes experiences into actionable knowledge
type LearningEngine interface {
	// Pattern recognition
	IdentifyPatterns(ctx context.Context, data []DataPoint) ([]*Pattern, error)
	ValidatePattern(ctx context.Context, pattern *Pattern) (bool, float64, error)
	
	// Rule extraction
	ExtractRules(ctx context.Context, outcomes []Outcome) ([]*Rule, error)
	TestRule(ctx context.Context, rule *Rule, testData []DataPoint) (*RuleValidation, error)
	
	// Knowledge synthesis
	ConnectConcepts(ctx context.Context, artifacts []*entities.LearningArtifact) (*KnowledgeGraph, error)
	IdentifyContradictions(ctx context.Context) ([]*Contradiction, error)
	ResolveContradiction(ctx context.Context, contradiction *Contradiction) (*entities.LearningArtifact, error)
}

// ExperimentEngine manages A/B tests and experiments
type ExperimentEngine interface {
	// Experiment design
	DesignExperiment(ctx context.Context, hypothesis string, metrics []string) (*entities.Experiment, error)
	ValidateExperimentDesign(ctx context.Context, experiment *entities.Experiment) error
	
	// Execution
	StartExperiment(ctx context.Context, experimentID string) error
	AssignVariant(ctx context.Context, experimentID, entityID string) (string, error)
	TrackMetric(ctx context.Context, experimentID, variantID string, metric string, value float64) error
	
	// Analysis
	CalculateSignificance(ctx context.Context, experimentID string) (*StatisticalAnalysis, error)
	DetermineWinner(ctx context.Context, experimentID string) (*entities.ExperimentResults, error)
	GenerateRecommendations(ctx context.Context, results *entities.ExperimentResults) ([]string, error)
}

// CapabilityAcquisition handles acquiring new capabilities
type CapabilityAcquisition interface {
	// Discovery
	DiscoverAPIs(ctx context.Context, capability string) ([]*APIDiscovery, error)
	EvaluateAPI(ctx context.Context, api *APIDiscovery) (*APIEvaluation, error)
	
	// Integration
	IntegrateAPI(ctx context.Context, api *APIDiscovery) error
	TestIntegration(ctx context.Context, integrationID string) (*IntegrationTest, error)
	
	// Internal development
	GenerateCapabilityScript(ctx context.Context, capability string, language string) (*Script, error)
	ValidateScript(ctx context.Context, script *Script) (*ScriptValidation, error)
	DeployScript(ctx context.Context, script *Script) error
}

// OptimizationEngine handles various system optimizations
type OptimizationEngine interface {
	// Prompt optimization
	AnalyzePromptPerformance(ctx context.Context, component string) (*PromptAnalysis, error)
	GeneratePromptVariants(ctx context.Context, prompt string, count int) ([]string, error)
	TestPromptVariants(ctx context.Context, variants []string, testCases []TestCase) (*PromptTestResults, error)
	
	// LLM selection
	BenchmarkLLMs(ctx context.Context, task string) (map[string]*LLMBenchmark, error)
	SelectLLMForTask(ctx context.Context, task string, constraints Constraints) (string, error)
	
	// Workflow optimization
	AnalyzeWorkflow(ctx context.Context, workflow string) (*WorkflowAnalysis, error)
	OptimizeWorkflowSteps(ctx context.Context, workflow string) ([]*WorkflowStep, error)
	SimulateWorkflow(ctx context.Context, steps []*WorkflowStep) (*WorkflowSimulation, error)
}
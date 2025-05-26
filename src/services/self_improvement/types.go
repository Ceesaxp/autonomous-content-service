package self_improvement

import (
	"time"
)

// PerformanceAnalysis contains analysis results for a component
type PerformanceAnalysis struct {
	Component   string                 `json:"component"`
	Period      string                 `json:"period"`
	Metrics     map[string]MetricStats `json:"metrics"`
	Trends      map[string]Trend       `json:"trends"`
	Insights    []Insight              `json:"insights"`
	Anomalies   []*Anomaly             `json:"anomalies"`
	Comparisons map[string]Comparison  `json:"comparisons"`
}

// MetricStats contains statistical information about a metric
type MetricStats struct {
	Mean        float64   `json:"mean"`
	Median      float64   `json:"median"`
	StdDev      float64   `json:"std_dev"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	P95         float64   `json:"p95"`
	P99         float64   `json:"p99"`
	Count       int       `json:"count"`
	LastValue   float64   `json:"last_value"`
	LastUpdated time.Time `json:"last_updated"`
}

// Trend represents a metric trend
type Trend struct {
	Direction   string  `json:"direction"` // increasing, decreasing, stable
	Magnitude   float64 `json:"magnitude"`
	Confidence  float64 `json:"confidence"`
	Prediction  float64 `json:"prediction"`
	Description string  `json:"description"`
}

// Insight represents an actionable insight
type Insight struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Impact      float64   `json:"impact"`
	Actions     []string  `json:"actions"`
	Evidence    []string  `json:"evidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Component   string                 `json:"component"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Expected    float64                `json:"expected"`
	Deviation   float64                `json:"deviation"`
	Severity    string                 `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	PossibleCauses []string            `json:"possible_causes"`
}

// Comparison represents a comparison with baseline or previous period
type Comparison struct {
	BaselineValue  float64 `json:"baseline_value"`
	CurrentValue   float64 `json:"current_value"`
	Change         float64 `json:"change"`
	ChangePercent  float64 `json:"change_percent"`
	IsImprovement  bool    `json:"is_improvement"`
}

// ErrorData contains information about an error
type ErrorData struct {
	Component   string                 `json:"component"`
	ErrorType   string                 `json:"error_type"`
	Message     string                 `json:"message"`
	StackTrace  string                 `json:"stack_trace"`
	Context     map[string]interface{} `json:"context"`
	Frequency   int                    `json:"frequency"`
	FirstSeen   time.Time              `json:"first_seen"`
	LastSeen    time.Time              `json:"last_seen"`
}

// ExperimentEvaluation contains experiment evaluation results
type ExperimentEvaluation struct {
	ExperimentID     string                 `json:"experiment_id"`
	Status           string                 `json:"status"`
	Winner           string                 `json:"winner"`
	Confidence       float64                `json:"confidence"`
	ImpactEstimate   float64                `json:"impact_estimate"`
	Recommendations  []string               `json:"recommendations"`
	RiskAssessment   RiskAssessment         `json:"risk_assessment"`
	Implementation   ImplementationPlan     `json:"implementation"`
}

// RiskAssessment evaluates risks
type RiskAssessment struct {
	OverallRisk string       `json:"overall_risk"`
	Risks       []Risk       `json:"risks"`
	Mitigations []Mitigation `json:"mitigations"`
}

// Risk represents a potential risk
type Risk struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"`
	Impact      float64 `json:"impact"`
	Score       float64 `json:"score"`
}

// Mitigation represents a risk mitigation strategy
type Mitigation struct {
	RiskType    string `json:"risk_type"`
	Strategy    string `json:"strategy"`
	Description string `json:"description"`
	Cost        float64 `json:"cost"`
}

// ImplementationPlan describes how to implement a change
type ImplementationPlan struct {
	Steps       []ImplementationStep `json:"steps"`
	Timeline    string               `json:"timeline"`
	Resources   []string             `json:"resources"`
	Rollback    RollbackPlan         `json:"rollback"`
}

// ImplementationStep is a single implementation step
type ImplementationStep struct {
	Order       int      `json:"order"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
	Duration    string   `json:"duration"`
	Validation  string   `json:"validation"`
}

// RollbackPlan describes how to rollback changes
type RollbackPlan struct {
	Trigger     string   `json:"trigger"`
	Steps       []string `json:"steps"`
	TimeLimit   string   `json:"time_limit"`
}

// WorkflowOptimization contains workflow optimization results
type WorkflowOptimization struct {
	WorkflowID      string         `json:"workflow_id"`
	OriginalSteps   []*WorkflowStep `json:"original_steps"`
	OptimizedSteps  []*WorkflowStep `json:"optimized_steps"`
	Improvements    map[string]float64 `json:"improvements"`
	SimulationResults *WorkflowSimulation `json:"simulation_results"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Duration    time.Duration          `json:"duration"`
	Cost        float64                `json:"cost"`
	Resources   []string               `json:"resources"`
	Dependencies []string              `json:"dependencies"`
	Config      map[string]interface{} `json:"config"`
}

// WorkflowSimulation contains simulation results
type WorkflowSimulation struct {
	TotalDuration   time.Duration      `json:"total_duration"`
	TotalCost       float64            `json:"total_cost"`
	Bottlenecks     []string           `json:"bottlenecks"`
	ResourceUsage   map[string]float64 `json:"resource_usage"`
	SuccessRate     float64            `json:"success_rate"`
	Parallelization float64            `json:"parallelization"`
}

// Improvement represents a potential improvement
type Improvement struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Component   string  `json:"component"`
	Impact      float64 `json:"impact"`
	Effort      float64 `json:"effort"`
	Cost        float64 `json:"cost"`
	TimeToValue string  `json:"time_to_value"`
}

// PrioritizedImprovement is an improvement with priority
type PrioritizedImprovement struct {
	Improvement Improvement `json:"improvement"`
	Priority    float64     `json:"priority"`
	ROI         float64     `json:"roi"`
	Score       float64     `json:"score"`
	Rationale   string      `json:"rationale"`
}

// CompetitorInsight contains competitor analysis
type CompetitorInsight struct {
	Competitor   string                 `json:"competitor"`
	Strengths    []string               `json:"strengths"`
	Weaknesses   []string               `json:"weaknesses"`
	Features     []string               `json:"features"`
	Pricing      map[string]float64     `json:"pricing"`
	Performance  map[string]float64     `json:"performance"`
	MarketShare  float64                `json:"market_share"`
	Differentiators []string            `json:"differentiators"`
	Opportunities []string              `json:"opportunities"`
}

// MarketOpportunity represents a market opportunity
type MarketOpportunity struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	MarketSize   float64  `json:"market_size"`
	GrowthRate   float64  `json:"growth_rate"`
	Competition  string   `json:"competition"`
	Requirements []string `json:"requirements"`
	Investment   float64  `json:"investment"`
	TimeToMarket string   `json:"time_to_market"`
	Confidence   float64  `json:"confidence"`
}

// MetricsSummary summarizes metrics
type MetricsSummary struct {
	Period          string                    `json:"period"`
	ComponentStats  map[string]ComponentStats `json:"component_stats"`
	OverallHealth   float64                   `json:"overall_health"`
	TopPerformers   []string                  `json:"top_performers"`
	NeedsAttention  []string                  `json:"needs_attention"`
	Recommendations []string                  `json:"recommendations"`
}

// ComponentStats contains component statistics
type ComponentStats struct {
	Health          float64            `json:"health"`
	Availability    float64            `json:"availability"`
	Performance     float64            `json:"performance"`
	ErrorRate       float64            `json:"error_rate"`
	Metrics         map[string]float64 `json:"metrics"`
}

// Additional types for learning engine
type DataPoint struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

type Pattern struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Conditions  []Condition `json:"conditions"`
	Frequency   int         `json:"frequency"`
	Confidence  float64     `json:"confidence"`
	Examples    []string    `json:"examples"`
}

type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type Outcome struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Result    string                 `json:"result"`
	Success   bool                   `json:"success"`
	Metrics   map[string]float64     `json:"metrics"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

type Rule struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Conditions  []Condition `json:"conditions"`
	Actions     []string    `json:"actions"`
	Priority    int         `json:"priority"`
	Confidence  float64     `json:"confidence"`
}

type RuleValidation struct {
	RuleID        string  `json:"rule_id"`
	TotalTests    int     `json:"total_tests"`
	Successful    int     `json:"successful"`
	Failed        int     `json:"failed"`
	Accuracy      float64 `json:"accuracy"`
	FalsePositives int    `json:"false_positives"`
	FalseNegatives int    `json:"false_negatives"`
}

type KnowledgeGraph struct {
	Nodes []KnowledgeNode `json:"nodes"`
	Edges []KnowledgeEdge `json:"edges"`
	Clusters []Cluster    `json:"clusters"`
}

type KnowledgeNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties"`
	Importance float64                `json:"importance"`
}

type KnowledgeEdge struct {
	ID       string                 `json:"id"`
	Source   string                 `json:"source"`
	Target   string                 `json:"target"`
	Type     string                 `json:"type"`
	Weight   float64                `json:"weight"`
	Properties map[string]interface{} `json:"properties"`
}

type Cluster struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NodeIDs     []string `json:"node_ids"`
	Coherence   float64  `json:"coherence"`
	Description string   `json:"description"`
}

type Contradiction struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Artifact1   string   `json:"artifact1"`
	Artifact2   string   `json:"artifact2"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Evidence    []string `json:"evidence"`
}

// API discovery types
type APIDiscovery struct {
	Provider     string                 `json:"provider"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Pricing      map[string]float64     `json:"pricing"`
	RateLimit    int                    `json:"rate_limit"`
	Documentation string               `json:"documentation"`
	SDKLanguages []string              `json:"sdk_languages"`
	Requirements []string              `json:"requirements"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type APIEvaluation struct {
	APIID           string             `json:"api_id"`
	Score           float64            `json:"score"`
	CostBenefit     float64            `json:"cost_benefit"`
	IntegrationTime string             `json:"integration_time"`
	Compatibility   float64            `json:"compatibility"`
	Reliability     float64            `json:"reliability"`
	Performance     map[string]float64 `json:"performance"`
	Risks           []Risk             `json:"risks"`
	Recommendation  string             `json:"recommendation"`
}

type IntegrationTest struct {
	IntegrationID string             `json:"integration_id"`
	TestCases     []TestCase         `json:"test_cases"`
	Results       []TestResult       `json:"results"`
	OverallStatus string             `json:"overall_status"`
	Performance   map[string]float64 `json:"performance"`
	Issues        []string           `json:"issues"`
}

type Script struct {
	ID          string `json:"id"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Inputs      []ScriptInput `json:"inputs"`
	Outputs     []ScriptOutput `json:"outputs"`
	Dependencies []string `json:"dependencies"`
	Version     string `json:"version"`
}

type ScriptInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     interface{} `json:"default"`
}

type ScriptOutput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ScriptValidation struct {
	ScriptID      string   `json:"script_id"`
	Valid         bool     `json:"valid"`
	Errors        []string `json:"errors"`
	Warnings      []string `json:"warnings"`
	SecurityScore float64  `json:"security_score"`
	Performance   float64  `json:"performance"`
}

// Optimization types
type PromptAnalysis struct {
	Component        string             `json:"component"`
	CurrentPrompt    string             `json:"current_prompt"`
	Performance      map[string]float64 `json:"performance"`
	Weaknesses       []string           `json:"weaknesses"`
	Opportunities    []string           `json:"opportunities"`
	SuggestedChanges []string           `json:"suggested_changes"`
}

type TestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Input       map[string]interface{} `json:"input"`
	Expected    map[string]interface{} `json:"expected"`
	Weight      float64                `json:"weight"`
}

type TestResult struct {
	TestCaseID string                 `json:"test_case_id"`
	Passed     bool                   `json:"passed"`
	Actual     map[string]interface{} `json:"actual"`
	Score      float64                `json:"score"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error,omitempty"`
}

type PromptTestResults struct {
	Variants []PromptVariantResult `json:"variants"`
	Winner   string                `json:"winner"`
	Analysis map[string]interface{} `json:"analysis"`
}

type PromptVariantResult struct {
	Variant      string             `json:"variant"`
	TestResults  []TestResult       `json:"test_results"`
	OverallScore float64            `json:"overall_score"`
	Metrics      map[string]float64 `json:"metrics"`
}

type LLMBenchmark struct {
	Provider     string             `json:"provider"`
	Model        string             `json:"model"`
	Performance  map[string]float64 `json:"performance"`
	Cost         float64            `json:"cost"`
	Latency      time.Duration      `json:"latency"`
	Reliability  float64            `json:"reliability"`
	Features     []string           `json:"features"`
}

type Constraints struct {
	MaxCost     float64            `json:"max_cost"`
	MaxLatency  time.Duration      `json:"max_latency"`
	MinQuality  float64            `json:"min_quality"`
	Required    []string           `json:"required"`
	Preferred   []string           `json:"preferred"`
	Weights     map[string]float64 `json:"weights"`
}

type WorkflowAnalysis struct {
	WorkflowID   string              `json:"workflow_id"`
	Steps        []*WorkflowStep     `json:"steps"`
	Bottlenecks  []Bottleneck        `json:"bottlenecks"`
	Inefficiencies []Inefficiency    `json:"inefficiencies"`
	Opportunities []Opportunity      `json:"opportunities"`
	Metrics      map[string]float64  `json:"metrics"`
}

type Bottleneck struct {
	StepID      string  `json:"step_id"`
	Type        string  `json:"type"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Solutions   []string `json:"solutions"`
}

type Inefficiency struct {
	Type        string   `json:"type"`
	StepIDs     []string `json:"step_ids"`
	Description string   `json:"description"`
	WastedTime  time.Duration `json:"wasted_time"`
	WastedCost  float64  `json:"wasted_cost"`
}

type Opportunity struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Benefit     float64 `json:"benefit"`
	Effort      float64 `json:"effort"`
	Priority    float64 `json:"priority"`
}

type StatisticalAnalysis struct {
	SampleSize       int                    `json:"sample_size"`
	PowerAnalysis    float64                `json:"power_analysis"`
	EffectSize       float64                `json:"effect_size"`
	PValue           float64                `json:"p_value"`
	ConfidenceLevel  float64                `json:"confidence_level"`
	SignificantDiff  bool                   `json:"significant_diff"`
	VariantStats     map[string]VariantStats `json:"variant_stats"`
}

type VariantStats struct {
	Mean       float64 `json:"mean"`
	StdDev     float64 `json:"std_dev"`
	SampleSize int     `json:"sample_size"`
	Conversions int    `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
}
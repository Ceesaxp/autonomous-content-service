package pricing

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// PricingService provides unified interface for pricing operations
type PricingService interface {
	// Quote Management
	GeneratePriceQuote(ctx context.Context, req *GeneratePriceQuoteRequest) (*entities.PriceQuote, error)
	UpdatePriceQuote(ctx context.Context, quoteID string, updates *UpdatePriceQuoteRequest) (*entities.PriceQuote, error)
	GetPriceQuote(ctx context.Context, quoteID string) (*entities.PriceQuote, error)
	AcceptPriceQuote(ctx context.Context, quoteID string) error
	RejectPriceQuote(ctx context.Context, quoteID string, reason string) error
	ExpirePriceQuote(ctx context.Context, quoteID string) error

	// Pricing Model Management
	CreatePricingModel(ctx context.Context, req *CreatePricingModelRequest) (*entities.PricingModel, error)
	UpdatePricingModel(ctx context.Context, modelID string, req *UpdatePricingModelRequest) (*entities.PricingModel, error)
	GetPricingModel(ctx context.Context, modelID string) (*entities.PricingModel, error)
	ListPricingModels(ctx context.Context, filter *PricingModelFilter) ([]*entities.PricingModel, error)

	// Client Pricing
	CreateClientPricingProfile(ctx context.Context, req *CreateClientPricingProfileRequest) (*entities.ClientPricingProfile, error)
	UpdateClientPricingProfile(ctx context.Context, profileID string, req *UpdateClientPricingProfileRequest) (*entities.ClientPricingProfile, error)
	GetClientPricingProfile(ctx context.Context, clientID string) (*entities.ClientPricingProfile, error)
	CalculateClientDiscount(ctx context.Context, clientID string, orderValue float64, volume int) (float64, error)

	// Market Intelligence
	UpdateMarketData(ctx context.Context, req *UpdateMarketDataRequest) (*entities.MarketData, error)
	GetLatestMarketData(ctx context.Context, contentType entities.ContentType, segment string) (*entities.MarketData, error)
	AnalyzeCompetitorPricing(ctx context.Context, req *CompetitorAnalysisRequest) (*repositories.CompetitorAnalysisResult, error)
	CalculatePriceElasticity(ctx context.Context, contentType entities.ContentType, timeRange repositories.TimeRange) (*repositories.PriceElasticityResult, error)

	// Cost Management
	CalculateResourceCost(ctx context.Context, req *CalculateResourceCostRequest) (*CalculateResourceCostResponse, error)
	RecordResourceUsage(ctx context.Context, req *RecordResourceUsageRequest) error
	GetCostAnalysis(ctx context.Context, req *CostAnalysisRequest) (*repositories.CostAnalysisResult, error)
	GetProfitabilityReport(ctx context.Context, req *ProfitabilityRequest) (*repositories.ProfitabilityResult, error)

	// A/B Testing
	CreatePricingExperiment(ctx context.Context, req *CreatePricingExperimentRequest) (*entities.PricingExperiment, error)
	UpdatePricingExperiment(ctx context.Context, experimentID string, req *UpdatePricingExperimentRequest) (*entities.PricingExperiment, error)
	GetPricingExperiment(ctx context.Context, experimentID string) (*entities.PricingExperiment, error)
	StartPricingExperiment(ctx context.Context, experimentID string) error
	StopPricingExperiment(ctx context.Context, experimentID string) error
	AnalyzePricingExperiment(ctx context.Context, experimentID string) (*entities.ExperimentResults, error)
	GetExperimentVariantForClient(ctx context.Context, clientID string, contentType entities.ContentType) (*entities.PricingVariant, error)

	// Analytics
	GetQuoteAcceptanceRate(ctx context.Context, filter *QuoteAnalyticsFilter) (float64, error)
	GetPriceDistribution(ctx context.Context, filter *PriceDistributionFilter) (map[string]int, error)
	GetPricingTrends(ctx context.Context, filter *PricingTrendsFilter) (*PricingTrendsResponse, error)
}

// MarketIntelligenceService handles market data collection and analysis
type MarketIntelligenceService interface {
	// Data Collection
	CollectCompetitorPricing(ctx context.Context, contentType entities.ContentType) error
	CollectMarketTrends(ctx context.Context, contentType entities.ContentType) error
	UpdateDemandLevel(ctx context.Context, contentType entities.ContentType, demandLevel entities.DemandLevel) error

	// Analysis
	AnalyzeMarketPosition(ctx context.Context, contentType entities.ContentType, ourPrice float64) (*MarketPositionAnalysis, error)
	DetectPricingOpportunities(ctx context.Context, contentType entities.ContentType) (*PricingOpportunityAnalysis, error)
	RecommendPriceAdjustment(ctx context.Context, req *PriceAdjustmentRecommendationRequest) (*PriceAdjustmentRecommendation, error)

	// Monitoring
	MonitorPriceChanges(ctx context.Context, contentType entities.ContentType) error
	DetectMarketAnomalies(ctx context.Context, contentType entities.ContentType) (*MarketAnomalyReport, error)
	GenerateMarketReport(ctx context.Context, req *MarketReportRequest) (*MarketReport, error)
}

// DynamicPricingEngine handles real-time price calculations
type DynamicPricingEngine interface {
	// Price Calculation
	CalculatePrice(ctx context.Context, req *CalculatePriceRequest) (*CalculatePriceResponse, error)
	ApplyComplexityAdjustments(ctx context.Context, basePrice float64, contentSpec *ContentSpecification) ([]entities.PriceAdjustment, error)
	ApplyMarketAdjustments(ctx context.Context, basePrice float64, contentType entities.ContentType) ([]entities.PriceAdjustment, error)
	ApplyClientAdjustments(ctx context.Context, basePrice float64, clientID string) ([]entities.PriceAdjustment, error)
	CalculateSurgeMultiplier(ctx context.Context, contentType entities.ContentType, deliveryTime time.Duration) (float64, error)

	// Real-time Adjustments
	AdjustPriceForDemand(ctx context.Context, basePrice float64, contentType entities.ContentType) (float64, error)
	AdjustPriceForCapacity(ctx context.Context, basePrice float64, currentLoad float64) (float64, error)
	AdjustPriceForTiming(ctx context.Context, basePrice float64, requestTime time.Time) (float64, error)

	// Optimization
	OptimizePriceForRevenue(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error)
	OptimizePriceForConversion(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error)
	OptimizePriceForMarketShare(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error)
}

// CostCalculationEngine handles resource cost tracking and calculation
type CostCalculationEngine interface {
	// Cost Calculation
	CalculateContentCreationCost(ctx context.Context, req *ContentCreationCostRequest) (*ContentCreationCostResponse, error)
	CalculateLLMCost(ctx context.Context, req *LLMCostRequest) (*LLMCostResponse, error)
	CalculateProcessingCost(ctx context.Context, req *ProcessingCostRequest) (*ProcessingCostResponse, error)
	CalculateInfrastructureCost(ctx context.Context, req *InfrastructureCostRequest) (*InfrastructureCostResponse, error)

	// Cost Tracking
	TrackResourceUsage(ctx context.Context, req *TrackResourceUsageRequest) error
	GetResourceUtilization(ctx context.Context, req *ResourceUtilizationRequest) (*ResourceUtilizationResponse, error)
	EstimateResourceCost(ctx context.Context, req *EstimateResourceCostRequest) (*EstimateResourceCostResponse, error)

	// Cost Optimization
	OptimizeResourceAllocation(ctx context.Context, req *OptimizeResourceAllocationRequest) (*OptimizeResourceAllocationResponse, error)
	IdentifyCostSavingOpportunities(ctx context.Context, req *CostSavingOpportunitiesRequest) (*CostSavingOpportunitiesResponse, error)
	GenerateCostReport(ctx context.Context, req *CostReportRequest) (*CostReportResponse, error)
}

// PricingExperimentService handles A/B testing for pricing strategies
type PricingExperimentService interface {
	// Experiment Management
	DesignExperiment(ctx context.Context, req *DesignExperimentRequest) (*entities.PricingExperiment, error)
	ValidateExperiment(ctx context.Context, experimentID string) (*ExperimentValidationResult, error)
	ExecuteExperiment(ctx context.Context, experimentID string) error
	MonitorExperiment(ctx context.Context, experimentID string) (*ExperimentMonitoringResult, error)

	// Variant Management
	CreateExperimentVariant(ctx context.Context, req *CreateExperimentVariantRequest) (*entities.PricingVariant, error)
	AssignClientToVariant(ctx context.Context, experimentID, clientID string) (*entities.PricingVariant, error)
	RecordExperimentEvent(ctx context.Context, req *RecordExperimentEventRequest) error

	// Analysis
	CalculateStatisticalSignificance(ctx context.Context, experimentID string) (*StatisticalSignificanceResult, error)
	GenerateExperimentReport(ctx context.Context, experimentID string) (*ExperimentReportResult, error)
	RecommendWinningVariant(ctx context.Context, experimentID string) (*VariantRecommendationResult, error)

	// Automation
	AutoCreateExperiments(ctx context.Context, req *AutoCreateExperimentsRequest) ([]*entities.PricingExperiment, error)
	AutoAnalyzeExperiments(ctx context.Context) ([]*entities.ExperimentResults, error)
	AutoApplyWinningVariants(ctx context.Context) ([]string, error)
}

// Request/Response types
type GeneratePriceQuoteRequest struct {
	ProjectID     string                 `json:"project_id"`
	ClientID      string                 `json:"client_id"`
	ContentType   entities.ContentType   `json:"content_type"`
	ContentSpec   *ContentSpecification  `json:"content_spec"`
	DeliveryTime  time.Duration          `json:"delivery_time"`
	Priority      string                 `json:"priority"` // standard, expedited, urgent
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

type UpdatePriceQuoteRequest struct {
	Status         *entities.QuoteStatus `json:"status,omitempty"`
	ValidUntil     *time.Time            `json:"valid_until,omitempty"`
	DiscountAmount *float64              `json:"discount_amount,omitempty"`
	Notes          *string               `json:"notes,omitempty"`
}

type CreatePricingModelRequest struct {
	Name            string                    `json:"name"`
	ContentType     entities.ContentType      `json:"content_type"`
	BasePrice       float64                   `json:"base_price"`
	Currency        string                    `json:"currency"`
	ComplexityRules []entities.ComplexityRule `json:"complexity_rules"`
	PricingFactors  map[string]float64        `json:"pricing_factors"`
}

type UpdatePricingModelRequest struct {
	Name           *string            `json:"name,omitempty"`
	BasePrice      *float64           `json:"base_price,omitempty"`
	PricingFactors map[string]float64 `json:"pricing_factors,omitempty"`
	IsActive       *bool              `json:"is_active,omitempty"`
}

type CreateClientPricingProfileRequest struct {
	ClientID        string                    `json:"client_id"`
	Tier            entities.ClientTier       `json:"tier"`
	VolumeDiscounts []entities.VolumeDiscount `json:"volume_discounts"`
	LoyaltyDiscount float64                   `json:"loyalty_discount"`
	CustomRates     map[string]float64        `json:"custom_rates"`
	PaymentTerms    entities.PaymentTerms     `json:"payment_terms"`
	CreditLimit     float64                   `json:"credit_limit"`
	RiskLevel       entities.RiskLevel        `json:"risk_level"`
}

type UpdateClientPricingProfileRequest struct {
	Tier            *entities.ClientTier   `json:"tier,omitempty"`
	LoyaltyDiscount *float64               `json:"loyalty_discount,omitempty"`
	CustomRates     map[string]float64     `json:"custom_rates,omitempty"`
	PaymentTerms    *entities.PaymentTerms `json:"payment_terms,omitempty"`
	CreditLimit     *float64               `json:"credit_limit,omitempty"`
	RiskLevel       *entities.RiskLevel    `json:"risk_level,omitempty"`
	IsActive        *bool                  `json:"is_active,omitempty"`
}

type UpdateMarketDataRequest struct {
	ContentType     entities.ContentType         `json:"content_type"`
	MarketSegment   string                       `json:"market_segment"`
	AveragePrice    float64                      `json:"average_price"`
	MedianPrice     float64                      `json:"median_price"`
	MinPrice        float64                      `json:"min_price"`
	MaxPrice        float64                      `json:"max_price"`
	SampleSize      int                          `json:"sample_size"`
	CompetitorData  []entities.CompetitorPricing `json:"competitor_data"`
	DemandLevel     entities.DemandLevel         `json:"demand_level"`
	TrendDirection  entities.TrendDirection      `json:"trend_direction"`
	ConfidenceScore float64                      `json:"confidence_score"`
	DataSource      string                       `json:"data_source"`
}

type CompetitorAnalysisRequest struct {
	ContentType   entities.ContentType   `json:"content_type"`
	CompetitorIDs []string               `json:"competitor_ids,omitempty"`
	TimeRange     repositories.TimeRange `json:"time_range"`
}

type CalculateResourceCostRequest struct {
	ProjectID     string                 `json:"project_id"`
	ContentType   entities.ContentType   `json:"content_type"`
	ContentSpec   *ContentSpecification  `json:"content_spec"`
	ResourceUsage *ResourceUsageEstimate `json:"resource_usage"`
}

type CalculateResourceCostResponse struct {
	TotalCost       float64            `json:"total_cost"`
	CostBreakdown   map[string]float64 `json:"cost_breakdown"`
	EstimatedTime   time.Duration      `json:"estimated_time"`
	ConfidenceLevel float64            `json:"confidence_level"`
}

type RecordResourceUsageRequest struct {
	ProjectID    string                 `json:"project_id"`
	ContentType  entities.ContentType   `json:"content_type"`
	ResourceType string                 `json:"resource_type"`
	ResourceName string                 `json:"resource_name"`
	Quantity     float64                `json:"quantity"`
	Unit         string                 `json:"unit"`
	Cost         float64                `json:"cost"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type CostAnalysisRequest struct {
	ContentType *entities.ContentType  `json:"content_type,omitempty"`
	TimeRange   repositories.TimeRange `json:"time_range"`
	GroupBy     string                 `json:"group_by"`
}

type ProfitabilityRequest struct {
	ContentType *entities.ContentType  `json:"content_type,omitempty"`
	ClientID    *string                `json:"client_id,omitempty"`
	TimeRange   repositories.TimeRange `json:"time_range"`
	GroupBy     string                 `json:"group_by"`
}

type CreatePricingExperimentRequest struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Hypothesis        string                    `json:"hypothesis"`
	Variants          []entities.PricingVariant `json:"variants"`
	TargetMetric      string                    `json:"target_metric"`
	TargetSegment     map[string]interface{}    `json:"target_segment"`
	TrafficSplit      map[string]float64        `json:"traffic_split"`
	Duration          time.Duration             `json:"duration"`
	SampleSize        int                       `json:"sample_size"`
	SignificanceLevel float64                   `json:"significance_level"`
}

type UpdatePricingExperimentRequest struct {
	Name         *string                    `json:"name,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Status       *entities.ExperimentStatus `json:"status,omitempty"`
	EndDate      *time.Time                 `json:"end_date,omitempty"`
	TrafficSplit map[string]float64         `json:"traffic_split,omitempty"`
}

// Cost calculation types
type LLMCostRequest struct {
	ModelName    string  `json:"model_name"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	RequestCount int     `json:"request_count"`
}

type LLMCostResponse struct {
	TotalCost       float64 `json:"total_cost"`
	InputCost       float64 `json:"input_cost"`
	OutputCost      float64 `json:"output_cost"`
	CostPerToken    float64 `json:"cost_per_token"`
	CostPerRequest  float64 `json:"cost_per_request"`
}

type ProcessingCostRequest struct {
	ProcessingTime time.Duration          `json:"processing_time"`
	CPUUsage       float64                `json:"cpu_usage"`
	MemoryUsage    float64                `json:"memory_usage"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type ProcessingCostResponse struct {
	TotalCost   float64            `json:"total_cost"`
	CPUCost     float64            `json:"cpu_cost"`
	MemoryCost  float64            `json:"memory_cost"`
	OtherCosts  map[string]float64 `json:"other_costs,omitempty"`
}

type InfrastructureCostRequest struct {
	ServiceType  string             `json:"service_type"`
	Usage        float64            `json:"usage"`
	Unit         string             `json:"unit"`
	Duration     time.Duration      `json:"duration"`
	Region       string             `json:"region"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type InfrastructureCostResponse struct {
	TotalCost     float64            `json:"total_cost"`
	CostBreakdown map[string]float64 `json:"cost_breakdown"`
	Unit          string             `json:"unit"`
	Rate          float64            `json:"rate"`
}

type ResourceUtilizationRequest struct {
	TimeRange   repositories.TimeRange `json:"time_range"`
	ResourceType string                `json:"resource_type"`
	Granularity string                 `json:"granularity"`
}

type ResourceUtilizationResponse struct {
	Utilization []ResourceUtilizationPoint `json:"utilization"`
	Average     float64                    `json:"average"`
	Peak        float64                    `json:"peak"`
	TotalCost   float64                    `json:"total_cost"`
}

type ResourceUtilizationPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
	Cost      float64   `json:"cost"`
}

type EstimateResourceCostRequest struct {
	ProjectID       string                 `json:"project_id"`
	ContentType     entities.ContentType   `json:"content_type"`
	EstimatedUsage  ResourceUsageEstimate  `json:"estimated_usage"`
	Duration        time.Duration          `json:"duration"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type EstimateResourceCostResponse struct {
	EstimatedCost   float64            `json:"estimated_cost"`
	CostBreakdown   map[string]float64 `json:"cost_breakdown"`
	ConfidenceLevel float64            `json:"confidence_level"`
	ValidUntil      time.Time          `json:"valid_until"`
}

// Additional request/response types
type TrackResourceUsageRequest struct {
	ProjectID      string                 `json:"project_id"`
	ResourceType   string                 `json:"resource_type"`
	Usage          float64                `json:"usage"`
	Unit           string                 `json:"unit"`
	Cost           float64                `json:"cost"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type OptimizeResourceAllocationRequest struct {
	ProjectID       string                 `json:"project_id"`
	CurrentUsage    ResourceUsageEstimate  `json:"current_usage"`
	OptimizationGoal string                `json:"optimization_goal"` // cost, performance, efficiency
	Constraints     map[string]interface{} `json:"constraints,omitempty"`
}

type OptimizeResourceAllocationResponse struct {
	RecommendedAllocation ResourceUsageEstimate  `json:"recommended_allocation"`
	EstimatedSavings      float64                `json:"estimated_savings"`
	PerformanceImpact     float64                `json:"performance_impact"`
	ImplementationSteps   []string               `json:"implementation_steps"`
}

type CostSavingOpportunitiesRequest struct {
	TimeRange           repositories.TimeRange `json:"time_range"`
	CurrentSpending     float64                `json:"current_spending"`
	OptimizationTargets []string               `json:"optimization_targets"`
}

type CostSavingOpportunitiesResponse struct {
	Opportunities      []CostSavingOpportunity `json:"opportunities"`
	TotalPotentialSavings float64              `json:"total_potential_savings"`
	ImplementationPlan []string                `json:"implementation_plan"`
}

type CostSavingOpportunity struct {
	Type                 string  `json:"type"`
	Description          string  `json:"description"`
	PotentialSavings     float64 `json:"potential_savings"`
	ImplementationEffort string  `json:"implementation_effort"`
	Priority             string  `json:"priority"`
}

type CostReportRequest struct {
	TimeRange      repositories.TimeRange `json:"time_range"`
	ReportType     string                 `json:"report_type"`
	Granularity    string                 `json:"granularity"`
	IncludeForecast bool                   `json:"include_forecast"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
}

type CostReportResponse struct {
	ReportData      map[string]interface{} `json:"report_data"`
	Summary         CostReportSummary      `json:"summary"`
	Trends          []CostTrendPoint       `json:"trends"`
	Forecast        []CostForecastPoint    `json:"forecast,omitempty"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

type CostReportSummary struct {
	TotalCost       float64 `json:"total_cost"`
	AverageDailyCost float64 `json:"average_daily_cost"`
	CostChange      float64 `json:"cost_change"`
	MajorDrivers    []string `json:"major_drivers"`
}

type CostTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Cost      float64   `json:"cost"`
	Category  string    `json:"category"`
}

type CostForecastPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	PredictedCost  float64   `json:"predicted_cost"`
	ConfidenceRange [2]float64 `json:"confidence_range"`
}

type DesignExperimentRequest struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Hypothesis       string                 `json:"hypothesis"`
	ExperimentType   string                 `json:"experiment_type"`
	TargetMetric     string                 `json:"target_metric"`
	Duration         time.Duration          `json:"duration"`
	TrafficPercent   float64                `json:"traffic_percent"`
	Parameters       map[string]interface{} `json:"parameters"`
}

// Experiment management types
type ExperimentValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Issues      []string `json:"issues"`
	Warnings    []string `json:"warnings"`
	Confidence  float64  `json:"confidence"`
}

type ExperimentMonitoringResult struct {
	Status       string             `json:"status"`
	Progress     float64            `json:"progress"`
	Metrics      map[string]float64 `json:"metrics"`
	Participants int                `json:"participants"`
	Issues       []string           `json:"issues"`
}

type CreateExperimentVariantRequest struct {
	ExperimentID string                 `json:"experiment_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Parameters   map[string]interface{} `json:"parameters"`
	Traffic      float64                `json:"traffic"`
}

type RecordExperimentEventRequest struct {
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	EventType    string                 `json:"event_type"`
	Properties   map[string]interface{} `json:"properties"`
	Timestamp    time.Time              `json:"timestamp"`
}

type StatisticalSignificanceResult struct {
	IsSignificant   bool    `json:"is_significant"`
	PValue          float64 `json:"p_value"`
	Confidence      float64 `json:"confidence"`
	SampleSize      int     `json:"sample_size"`
	EffectSize      float64 `json:"effect_size"`
	PowerAnalysis   float64 `json:"power_analysis"`
}

type ExperimentReportResult struct {
	Summary     string                 `json:"summary"`
	Results     map[string]interface{} `json:"results"`
	Charts      []string               `json:"charts"`
	Insights    []string               `json:"insights"`
	GeneratedAt time.Time              `json:"generated_at"`
}

type VariantRecommendationResult struct {
	WinningVariant string             `json:"winning_variant"`
	Confidence     float64            `json:"confidence"`
	ExpectedLift   float64            `json:"expected_lift"`
	Reasoning      []string           `json:"reasoning"`
	NextSteps      []string           `json:"next_steps"`
}

type AutoCreateExperimentsRequest struct {
	ContentTypes     []entities.ContentType `json:"content_types"`
	MaxExperiments   int                    `json:"max_experiments"`
	Duration         time.Duration          `json:"duration"`
	TrafficThreshold float64                `json:"traffic_threshold"`
}

type ContentSpecification struct {
	Title           string                 `json:"title"`
	WordCount       int                    `json:"word_count"`
	ComplexityLevel string                 `json:"complexity_level"`
	ResearchDepth   string                 `json:"research_depth"`
	TechnicalLevel  string                 `json:"technical_level"`
	TargetAudience  string                 `json:"target_audience"`
	Requirements    []string               `json:"requirements"`
	AdditionalSpecs map[string]interface{} `json:"additional_specs,omitempty"`
}

type ResourceUsageEstimate struct {
	LLMTokens           int                `json:"llm_tokens"`
	LLMCalls            int                `json:"llm_calls"`
	ProcessingTime      time.Duration      `json:"processing_time"`
	StorageRequired     float64            `json:"storage_required"`
	Bandwidth           float64            `json:"bandwidth"`
	AdditionalResources map[string]float64 `json:"additional_resources,omitempty"`
}

// Filter and analytics types
type PricingModelFilter struct {
	ContentType *entities.ContentType   `json:"content_type,omitempty"`
	IsActive    *bool                   `json:"is_active,omitempty"`
	CreatedBy   *string                 `json:"created_by,omitempty"`
	TimeRange   *repositories.TimeRange `json:"time_range,omitempty"`
	Limit       int                     `json:"limit"`
	Offset      int                     `json:"offset"`
}

type QuoteAnalyticsFilter struct {
	ContentType *entities.ContentType   `json:"content_type,omitempty"`
	ClientTier  *entities.ClientTier    `json:"client_tier,omitempty"`
	TimeRange   *repositories.TimeRange `json:"time_range,omitempty"`
}

type PriceDistributionFilter struct {
	ContentType *entities.ContentType   `json:"content_type,omitempty"`
	TimeRange   *repositories.TimeRange `json:"time_range,omitempty"`
	Buckets     int                     `json:"buckets"`
}

type PricingTrendsFilter struct {
	ContentType *entities.ContentType  `json:"content_type,omitempty"`
	TimeRange   repositories.TimeRange `json:"time_range"`
	Granularity string                 `json:"granularity"` // daily, weekly, monthly
}

type PricingTrendsResponse struct {
	Trends    []PricingTrendPoint `json:"trends"`
	Summary   TrendSummary        `json:"summary"`
	Forecasts []PricingForecast   `json:"forecasts"`
}

type PricingTrendPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	AveragePrice   float64   `json:"average_price"`
	Volume         int       `json:"volume"`
	AcceptanceRate float64   `json:"acceptance_rate"`
}

type TrendSummary struct {
	OverallTrend    string             `json:"overall_trend"`
	GrowthRate      float64            `json:"growth_rate"`
	Volatility      float64            `json:"volatility"`
	SeasonalFactors map[string]float64 `json:"seasonal_factors"`
}

type PricingForecast struct {
	Timestamp          time.Time          `json:"timestamp"`
	PredictedPrice     float64            `json:"predicted_price"`
	ConfidenceInterval [2]float64         `json:"confidence_interval"`
	Factors            map[string]float64 `json:"factors"`
}

// Market intelligence types
type MarketPositionAnalysis struct {
	Position        string   `json:"position"`
	MarketShare     float64  `json:"market_share"`
	PriceAdvantage  float64  `json:"price_advantage"`
	CompetitiveGap  float64  `json:"competitive_gap"`
	Recommendations []string `json:"recommendations"`
}

type PricingOpportunityAnalysis struct {
	Opportunities    []PricingOpportunity `json:"opportunities"`
	PotentialRevenue float64              `json:"potential_revenue"`
	RiskLevel        entities.RiskLevel   `json:"risk_level"`
	Timeframe        string               `json:"timeframe"`
}

type PricingOpportunity struct {
	Type                 string             `json:"type"`
	Description          string             `json:"description"`
	RevenueImpact        float64            `json:"revenue_impact"`
	ImplementationEffort string             `json:"implementation_effort"`
	RiskLevel            entities.RiskLevel `json:"risk_level"`
}

type PriceAdjustmentRecommendationRequest struct {
	ContentType       entities.ContentType   `json:"content_type"`
	CurrentPrice      float64                `json:"current_price"`
	MarketConditions  map[string]interface{} `json:"market_conditions"`
	BusinessObjective string                 `json:"business_objective"` // revenue, market_share, profit
}

type PriceAdjustmentRecommendation struct {
	RecommendedPrice   float64             `json:"recommended_price"`
	AdjustmentPercent  float64             `json:"adjustment_percent"`
	ExpectedImpact     PriceImpactAnalysis `json:"expected_impact"`
	ImplementationPlan []string            `json:"implementation_plan"`
	MonitoringMetrics  []string            `json:"monitoring_metrics"`
}

type PriceImpactAnalysis struct {
	RevenueImpact     float64            `json:"revenue_impact"`
	VolumeImpact      float64            `json:"volume_impact"`
	MarketShareImpact float64            `json:"market_share_impact"`
	CompetitiveRisk   entities.RiskLevel `json:"competitive_risk"`
}

type MarketAnomalyReport struct {
	Anomalies          []MarketAnomaly `json:"anomalies"`
	Severity           string          `json:"severity"`
	RecommendedActions []string        `json:"recommended_actions"`
	DetectedAt         time.Time       `json:"detected_at"`
}

type MarketAnomaly struct {
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Severity    string             `json:"severity"`
	DetectedAt  time.Time          `json:"detected_at"`
	Metrics     map[string]float64 `json:"metrics"`
}

type MarketReportRequest struct {
	ContentTypes              []entities.ContentType `json:"content_types"`
	TimeRange                 repositories.TimeRange `json:"time_range"`
	IncludeForecast           bool                   `json:"include_forecast"`
	IncludeCompetitorAnalysis bool                   `json:"include_competitor_analysis"`
}

type MarketReport struct {
	Summary            MarketSummary                          `json:"summary"`
	Trends             []MarketTrend                          `json:"trends"`
	CompetitorAnalysis *repositories.CompetitorAnalysisResult `json:"competitor_analysis,omitempty"`
	Forecast           []PricingForecast                      `json:"forecast,omitempty"`
	Recommendations    []string                               `json:"recommendations"`
	GeneratedAt        time.Time                              `json:"generated_at"`
}

type MarketSummary struct {
	AveragePrice   float64                 `json:"average_price"`
	PriceRange     [2]float64              `json:"price_range"`
	DemandLevel    entities.DemandLevel    `json:"demand_level"`
	TrendDirection entities.TrendDirection `json:"trend_direction"`
	MarketSize     float64                 `json:"market_size"`
	GrowthRate     float64                 `json:"growth_rate"`
}

type MarketTrend struct {
	Metric    string    `json:"metric"`
	Trend     string    `json:"trend"`
	Magnitude float64   `json:"magnitude"`
	Period    string    `json:"period"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

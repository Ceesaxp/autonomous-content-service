package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// PricingModelRepository manages pricing model data access
type PricingModelRepository interface {
	// Pricing Models
	CreatePricingModel(ctx context.Context, model *entities.PricingModel) error
	GetPricingModel(ctx context.Context, id string) (*entities.PricingModel, error)
	GetPricingModelByContentType(ctx context.Context, contentType entities.ContentType) (*entities.PricingModel, error)
	UpdatePricingModel(ctx context.Context, model *entities.PricingModel) error
	DeletePricingModel(ctx context.Context, id string) error
	ListPricingModels(ctx context.Context, filter PricingModelFilter) ([]*entities.PricingModel, error)

	// Complexity Rules
	CreateComplexityRule(ctx context.Context, rule *entities.ComplexityRule) error
	GetComplexityRule(ctx context.Context, id string) (*entities.ComplexityRule, error)
	UpdateComplexityRule(ctx context.Context, rule *entities.ComplexityRule) error
	DeleteComplexityRule(ctx context.Context, id string) error
	ListComplexityRules(ctx context.Context, filter ComplexityRuleFilter) ([]*entities.ComplexityRule, error)
}

// PriceQuoteRepository manages price quote data access
type PriceQuoteRepository interface {
	// Price Quotes
	CreatePriceQuote(ctx context.Context, quote *entities.PriceQuote) error
	GetPriceQuote(ctx context.Context, id string) (*entities.PriceQuote, error)
	GetPriceQuotesByProject(ctx context.Context, projectID string) ([]*entities.PriceQuote, error)
	GetPriceQuotesByClient(ctx context.Context, clientID string, filter PriceQuoteFilter) ([]*entities.PriceQuote, error)
	UpdatePriceQuote(ctx context.Context, quote *entities.PriceQuote) error
	DeletePriceQuote(ctx context.Context, id string) error
	ListPriceQuotes(ctx context.Context, filter PriceQuoteFilter) ([]*entities.PriceQuote, error)

	// Quote Analytics
	GetQuoteAcceptanceRate(ctx context.Context, filter QuoteAnalyticsFilter) (float64, error)
	GetAveragePriceByContentType(ctx context.Context, contentType entities.ContentType, timeRange TimeRange) (float64, error)
	GetPriceDistribution(ctx context.Context, filter PriceDistributionFilter) (map[string]int, error)
}

// MarketDataRepository manages market intelligence data
type MarketDataRepository interface {
	// Market Data
	CreateMarketData(ctx context.Context, data *entities.MarketData) error
	GetMarketData(ctx context.Context, id string) (*entities.MarketData, error)
	GetLatestMarketData(ctx context.Context, contentType entities.ContentType, segment string) (*entities.MarketData, error)
	UpdateMarketData(ctx context.Context, data *entities.MarketData) error
	DeleteMarketData(ctx context.Context, id string) error
	ListMarketData(ctx context.Context, filter MarketDataFilter) ([]*entities.MarketData, error)

	// Competitor Data
	CreateCompetitorPricing(ctx context.Context, pricing *entities.CompetitorPricing) error
	GetCompetitorPricing(ctx context.Context, competitorID string, contentType entities.ContentType) ([]*entities.CompetitorPricing, error)
	UpdateCompetitorPricing(ctx context.Context, pricing *entities.CompetitorPricing) error
	DeleteCompetitorPricing(ctx context.Context, competitorID string, id string) error

	// Market Analytics
	GetMarketTrends(ctx context.Context, contentType entities.ContentType, timeRange TimeRange) ([]*entities.MarketData, error)
	GetCompetitorAnalysis(ctx context.Context, filter CompetitorAnalysisFilter) (*CompetitorAnalysisResult, error)
	GetPriceElasticity(ctx context.Context, contentType entities.ContentType, timeRange TimeRange) (*PriceElasticityResult, error)
}

// ClientPricingRepository manages client-specific pricing data
type ClientPricingRepository interface {
	// Client Pricing Profiles
	CreateClientPricingProfile(ctx context.Context, profile *entities.ClientPricingProfile) error
	GetClientPricingProfile(ctx context.Context, id string) (*entities.ClientPricingProfile, error)
	GetClientPricingProfileByClient(ctx context.Context, clientID string) (*entities.ClientPricingProfile, error)
	UpdateClientPricingProfile(ctx context.Context, profile *entities.ClientPricingProfile) error
	DeleteClientPricingProfile(ctx context.Context, id string) error
	ListClientPricingProfiles(ctx context.Context, filter ClientPricingFilter) ([]*entities.ClientPricingProfile, error)

	// Volume Discounts
	CreateVolumeDiscount(ctx context.Context, discount *entities.VolumeDiscount, clientID string) error
	UpdateVolumeDiscount(ctx context.Context, discount *entities.VolumeDiscount, clientID string) error
	DeleteVolumeDiscount(ctx context.Context, clientID string, minVolume int) error

	// Special Offers
	CreateSpecialOffer(ctx context.Context, offer *entities.SpecialOffer, clientID string) error
	GetSpecialOffer(ctx context.Context, id string) (*entities.SpecialOffer, error)
	UpdateSpecialOffer(ctx context.Context, offer *entities.SpecialOffer) error
	DeleteSpecialOffer(ctx context.Context, id string) error
	GetActiveSpecialOffers(ctx context.Context, clientID string) ([]*entities.SpecialOffer, error)

	// Client Analytics
	GetClientPricingMetrics(ctx context.Context, clientID string, timeRange TimeRange) (*entities.ClientPricingMetrics, error)
	UpdateClientPricingMetrics(ctx context.Context, clientID string, metrics *entities.ClientPricingMetrics) error
	GetClientChurnRisk(ctx context.Context, clientID string) (float64, error)
	GetClientLifetimeValue(ctx context.Context, clientID string) (float64, error)
}

// CostModelRepository manages cost calculation data
type CostModelRepository interface {
	// Cost Models
	CreateCostModel(ctx context.Context, model *entities.CostModel) error
	GetCostModel(ctx context.Context, id string) (*entities.CostModel, error)
	GetCostModelByContentType(ctx context.Context, contentType entities.ContentType) (*entities.CostModel, error)
	UpdateCostModel(ctx context.Context, model *entities.CostModel) error
	DeleteCostModel(ctx context.Context, id string) error
	ListCostModels(ctx context.Context, filter CostModelFilter) ([]*entities.CostModel, error)

	// Cost Tracking
	RecordResourceUsage(ctx context.Context, usage *ResourceUsageRecord) error
	GetResourceUsage(ctx context.Context, filter ResourceUsageFilter) ([]*ResourceUsageRecord, error)
	GetCostAnalysis(ctx context.Context, filter CostAnalysisFilter) (*CostAnalysisResult, error)
	GetProfitabilityReport(ctx context.Context, filter ProfitabilityFilter) (*ProfitabilityResult, error)
}

// PricingExperimentRepository manages A/B testing for pricing
type PricingExperimentRepository interface {
	// Pricing Experiments
	CreatePricingExperiment(ctx context.Context, experiment *entities.PricingExperiment) error
	GetPricingExperiment(ctx context.Context, id string) (*entities.PricingExperiment, error)
	UpdatePricingExperiment(ctx context.Context, experiment *entities.PricingExperiment) error
	DeletePricingExperiment(ctx context.Context, id string) error
	ListPricingExperiments(ctx context.Context, filter ExperimentFilter) ([]*entities.PricingExperiment, error)

	// Experiment Variants
	CreatePricingVariant(ctx context.Context, variant *entities.PricingVariant, experimentID string) error
	UpdatePricingVariant(ctx context.Context, variant *entities.PricingVariant) error
	DeletePricingVariant(ctx context.Context, id string) error

	// Experiment Results
	RecordExperimentEvent(ctx context.Context, event *ExperimentEvent) error
	GetExperimentMetrics(ctx context.Context, experimentID string, variantID string) (*entities.VariantMetrics, error)
	UpdateExperimentResults(ctx context.Context, experimentID string, results *entities.ExperimentResults) error
	GetActiveExperiments(ctx context.Context, contentType entities.ContentType) ([]*entities.PricingExperiment, error)
}

// Data transfer objects and filters
type PricingModelFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	CreatedBy   *string               `json:"created_by,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	Limit       int                   `json:"limit"`
	Offset      int                   `json:"offset"`
}

type ComplexityRuleFilter struct {
	Metric    *string    `json:"metric,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	TimeRange *TimeRange `json:"time_range,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type PriceQuoteFilter struct {
	ProjectID   *string               `json:"project_id,omitempty"`
	ClientID    *string               `json:"client_id,omitempty"`
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	Status      *entities.QuoteStatus `json:"status,omitempty"`
	PriceRange  *PriceRange           `json:"price_range,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	Limit       int                   `json:"limit"`
	Offset      int                   `json:"offset"`
}

type QuoteAnalyticsFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	ClientTier  *entities.ClientTier  `json:"client_tier,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
}

type PriceDistributionFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	Buckets     int                   `json:"buckets"`
}

type MarketDataFilter struct {
	ContentType     *entities.ContentType    `json:"content_type,omitempty"`
	MarketSegment   *string                  `json:"market_segment,omitempty"`
	DemandLevel     *entities.DemandLevel    `json:"demand_level,omitempty"`
	TrendDirection  *entities.TrendDirection `json:"trend_direction,omitempty"`
	ConfidenceRange *ConfidenceRange         `json:"confidence_range,omitempty"`
	TimeRange       *TimeRange               `json:"time_range,omitempty"`
	Limit           int                      `json:"limit"`
	Offset          int                      `json:"offset"`
}

type CompetitorAnalysisFilter struct {
	ContentType   *entities.ContentType `json:"content_type,omitempty"`
	CompetitorIDs []string              `json:"competitor_ids,omitempty"`
	TimeRange     *TimeRange            `json:"time_range,omitempty"`
}

type ClientPricingFilter struct {
	ClientID  *string              `json:"client_id,omitempty"`
	Tier      *entities.ClientTier `json:"tier,omitempty"`
	RiskLevel *entities.RiskLevel  `json:"risk_level,omitempty"`
	IsActive  *bool                `json:"is_active,omitempty"`
	TimeRange *TimeRange           `json:"time_range,omitempty"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
}

type CostModelFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	Limit       int                   `json:"limit"`
	Offset      int                   `json:"offset"`
}

type ResourceUsageFilter struct {
	ContentType  *entities.ContentType `json:"content_type,omitempty"`
	ProjectID    *string               `json:"project_id,omitempty"`
	ResourceType *string               `json:"resource_type,omitempty"`
	TimeRange    *TimeRange            `json:"time_range,omitempty"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type CostAnalysisFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	GroupBy     string                `json:"group_by"` // content_type, project, client, resource_type
}

type ProfitabilityFilter struct {
	ContentType *entities.ContentType `json:"content_type,omitempty"`
	ClientID    *string               `json:"client_id,omitempty"`
	TimeRange   *TimeRange            `json:"time_range,omitempty"`
	GroupBy     string                `json:"group_by"` // content_type, client, project
}

type ExperimentFilter struct {
	ContentType *entities.ContentType      `json:"content_type,omitempty"`
	Status      *entities.ExperimentStatus `json:"status,omitempty"`
	CreatedBy   *string                    `json:"created_by,omitempty"`
	TimeRange   *TimeRange                 `json:"time_range,omitempty"`
	Limit       int                        `json:"limit"`
	Offset      int                        `json:"offset"`
}

// Supporting types
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type ConfidenceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type ResourceUsageRecord struct {
	ID           string                 `json:"id" db:"id"`
	ProjectID    string                 `json:"project_id" db:"project_id"`
	ContentType  entities.ContentType   `json:"content_type" db:"content_type"`
	ResourceType string                 `json:"resource_type" db:"resource_type"` // llm_api, processing, storage
	ResourceName string                 `json:"resource_name" db:"resource_name"` // gpt-4, claude-3, cpu_time
	Quantity     float64                `json:"quantity" db:"quantity"`
	Unit         string                 `json:"unit" db:"unit"` // tokens, minutes, mb
	Cost         float64                `json:"cost" db:"cost"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	RecordedAt   time.Time              `json:"recorded_at" db:"recorded_at"`
}

type CompetitorAnalysisResult struct {
	ContentType        entities.ContentType `json:"content_type"`
	OurAveragePrice    float64              `json:"our_average_price"`
	MarketAveragePrice float64              `json:"market_average_price"`
	PricePosition      string               `json:"price_position"` // lowest, below_market, market_rate, above_market, highest
	CompetitorCount    int                  `json:"competitor_count"`
	MarketShare        float64              `json:"market_share"`
	PriceGap           float64              `json:"price_gap"`
	Recommendation     string               `json:"recommendation"`
	AnalyzedAt         time.Time            `json:"analyzed_at"`
}

type PriceElasticityResult struct {
	ContentType     entities.ContentType `json:"content_type"`
	ElasticityScore float64              `json:"elasticity_score"` // Negative values indicate elastic demand
	OptimalPrice    float64              `json:"optimal_price"`
	RevenueImpact   float64              `json:"revenue_impact"`
	ConfidenceLevel float64              `json:"confidence_level"`
	DataPoints      int                  `json:"data_points"`
	AnalyzedAt      time.Time            `json:"analyzed_at"`
}

type CostAnalysisResult struct {
	TotalCost       float64            `json:"total_cost"`
	CostByType      map[string]float64 `json:"cost_by_type"`
	AverageCost     float64            `json:"average_cost"`
	CostTrend       string             `json:"cost_trend"` // increasing, decreasing, stable
	EfficiencyScore float64            `json:"efficiency_score"`
	Recommendations []string           `json:"recommendations"`
	AnalyzedAt      time.Time          `json:"analyzed_at"`
}

type ProfitabilityResult struct {
	TotalRevenue    float64            `json:"total_revenue"`
	TotalCost       float64            `json:"total_cost"`
	GrossProfit     float64            `json:"gross_profit"`
	ProfitMargin    float64            `json:"profit_margin"`
	ProfitBySegment map[string]float64 `json:"profit_by_segment"`
	ROI             float64            `json:"roi"`
	BreakevenPoint  float64            `json:"breakeven_point"`
	AnalyzedAt      time.Time          `json:"analyzed_at"`
}

type ExperimentEvent struct {
	ID           string                 `json:"id" db:"id"`
	ExperimentID string                 `json:"experiment_id" db:"experiment_id"`
	VariantID    string                 `json:"variant_id" db:"variant_id"`
	ClientID     string                 `json:"client_id" db:"client_id"`
	EventType    string                 `json:"event_type" db:"event_type"` // quote_generated, quote_sent, quote_accepted, quote_rejected
	Price        float64                `json:"price" db:"price"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	RecordedAt   time.Time              `json:"recorded_at" db:"recorded_at"`
}

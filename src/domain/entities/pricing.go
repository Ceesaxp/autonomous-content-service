package entities

import (
	"time"
)

// PricingModel defines base pricing structures for different content types
type PricingModel struct {
	ID              string             `json:"id" db:"id"`
	Name            string             `json:"name" db:"name"`
	ContentType     ContentType        `json:"content_type" db:"content_type"`
	BasePrice       float64            `json:"base_price" db:"base_price"`
	Currency        string             `json:"currency" db:"currency"`
	ComplexityRules []ComplexityRule   `json:"complexity_rules" db:"complexity_rules"`
	PricingFactors  map[string]float64 `json:"pricing_factors" db:"pricing_factors"`
	IsActive        bool               `json:"is_active" db:"is_active"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
	CreatedBy       string             `json:"created_by" db:"created_by"`
}

// ComplexityRule defines pricing adjustments based on content complexity
type ComplexityRule struct {
	ID             string                `json:"id" db:"id"`
	Name           string                `json:"name" db:"name"`
	Metric         string                `json:"metric" db:"metric"` // word_count, research_depth, technical_level, etc.
	Thresholds     []ComplexityThreshold `json:"thresholds" db:"thresholds"`
	Multiplier     float64               `json:"multiplier" db:"multiplier"`
	AdditionalCost float64               `json:"additional_cost" db:"additional_cost"`
	IsActive       bool                  `json:"is_active" db:"is_active"`
}

// ComplexityThreshold defines condition-based pricing adjustments
type ComplexityThreshold struct {
	Condition  string   `json:"condition" db:"condition"` // gt, gte, lt, lte, eq, range
	Value      float64  `json:"value" db:"value"`
	UpperValue *float64 `json:"upper_value,omitempty" db:"upper_value"`
	Multiplier float64  `json:"multiplier" db:"multiplier"`
	FixedCost  float64  `json:"fixed_cost" db:"fixed_cost"`
}

// PriceQuote represents a calculated price for specific content request
type PriceQuote struct {
	ID                    string                 `json:"id" db:"id"`
	ProjectID             string                 `json:"project_id" db:"project_id"`
	ClientID              string                 `json:"client_id" db:"client_id"`
	ContentType           ContentType            `json:"content_type" db:"content_type"`
	BasePrice             float64                `json:"base_price" db:"base_price"`
	ComplexityAdjustments []PriceAdjustment      `json:"complexity_adjustments" db:"complexity_adjustments"`
	MarketAdjustments     []PriceAdjustment      `json:"market_adjustments" db:"market_adjustments"`
	ClientAdjustments     []PriceAdjustment      `json:"client_adjustments" db:"client_adjustments"`
	SurgeMultiplier       float64                `json:"surge_multiplier" db:"surge_multiplier"`
	DiscountAmount        float64                `json:"discount_amount" db:"discount_amount"`
	FinalPrice            float64                `json:"final_price" db:"final_price"`
	Currency              string                 `json:"currency" db:"currency"`
	ValidUntil            time.Time              `json:"valid_until" db:"valid_until"`
	Status                QuoteStatus            `json:"status" db:"status"`
	Metadata              map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// PriceAdjustment represents individual pricing adjustments
type PriceAdjustment struct {
	Type        string  `json:"type" db:"type"`             // complexity, market, client, surge, discount
	Reason      string  `json:"reason" db:"reason"`         // word_count, demand, loyalty, expedited, etc.
	Amount      float64 `json:"amount" db:"amount"`         // Fixed amount adjustment
	Multiplier  float64 `json:"multiplier" db:"multiplier"` // Percentage adjustment
	Description string  `json:"description" db:"description"`
}

// MarketData represents competitor and market pricing information
type MarketData struct {
	ID              string              `json:"id" db:"id"`
	ContentType     ContentType         `json:"content_type" db:"content_type"`
	MarketSegment   string              `json:"market_segment" db:"market_segment"`
	AveragePrice    float64             `json:"average_price" db:"average_price"`
	MedianPrice     float64             `json:"median_price" db:"median_price"`
	MinPrice        float64             `json:"min_price" db:"min_price"`
	MaxPrice        float64             `json:"max_price" db:"max_price"`
	SampleSize      int                 `json:"sample_size" db:"sample_size"`
	CompetitorData  []CompetitorPricing `json:"competitor_data" db:"competitor_data"`
	DemandLevel     DemandLevel         `json:"demand_level" db:"demand_level"`
	TrendDirection  TrendDirection      `json:"trend_direction" db:"trend_direction"`
	ConfidenceScore float64             `json:"confidence_score" db:"confidence_score"`
	DataSource      string              `json:"data_source" db:"data_source"`
	CollectedAt     time.Time           `json:"collected_at" db:"collected_at"`
	ValidUntil      time.Time           `json:"valid_until" db:"valid_until"`
}

// CompetitorPricing represents individual competitor pricing data
type CompetitorPricing struct {
	CompetitorID   string                 `json:"competitor_id" db:"competitor_id"`
	CompetitorName string                 `json:"competitor_name" db:"competitor_name"`
	Price          float64                `json:"price" db:"price"`
	Currency       string                 `json:"currency" db:"currency"`
	ServiceLevel   string                 `json:"service_level" db:"service_level"`
	DeliveryTime   string                 `json:"delivery_time" db:"delivery_time"`
	QualityRating  float64                `json:"quality_rating" db:"quality_rating"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CollectedAt    time.Time              `json:"collected_at" db:"collected_at"`
}

// ClientPricingProfile represents client-specific pricing configurations
type ClientPricingProfile struct {
	ID              string               `json:"id" db:"id"`
	ClientID        string               `json:"client_id" db:"client_id"`
	Tier            ClientTier           `json:"tier" db:"tier"`
	VolumeDiscounts []VolumeDiscount     `json:"volume_discounts" db:"volume_discounts"`
	LoyaltyDiscount float64              `json:"loyalty_discount" db:"loyalty_discount"`
	CustomRates     map[string]float64   `json:"custom_rates" db:"custom_rates"`
	PaymentTerms    PaymentTerms         `json:"payment_terms" db:"payment_terms"`
	CreditLimit     float64              `json:"credit_limit" db:"credit_limit"`
	RiskLevel       RiskLevel            `json:"risk_level" db:"risk_level"`
	SpecialOffers   []SpecialOffer       `json:"special_offers" db:"special_offers"`
	Metrics         ClientPricingMetrics `json:"metrics" db:"metrics"`
	IsActive        bool                 `json:"is_active" db:"is_active"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" db:"updated_at"`
}

// VolumeDiscount defines discount based on purchase volume
type VolumeDiscount struct {
	MinVolume    int     `json:"min_volume" db:"min_volume"`
	MaxVolume    *int    `json:"max_volume,omitempty" db:"max_volume"`
	DiscountRate float64 `json:"discount_rate" db:"discount_rate"`
	DiscountType string  `json:"discount_type" db:"discount_type"` // percentage, fixed_amount
	Period       string  `json:"period" db:"period"`               // monthly, quarterly, yearly
	IsActive     bool    `json:"is_active" db:"is_active"`
}

// SpecialOffer represents time-limited pricing offers
type SpecialOffer struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Description  string                 `json:"description" db:"description"`
	DiscountRate float64                `json:"discount_rate" db:"discount_rate"`
	DiscountType string                 `json:"discount_type" db:"discount_type"`
	Conditions   map[string]interface{} `json:"conditions" db:"conditions"`
	ValidFrom    time.Time              `json:"valid_from" db:"valid_from"`
	ValidUntil   time.Time              `json:"valid_until" db:"valid_until"`
	UsageLimit   *int                   `json:"usage_limit,omitempty" db:"usage_limit"`
	UsageCount   int                    `json:"usage_count" db:"usage_count"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
}

// ClientPricingMetrics tracks client pricing performance
type ClientPricingMetrics struct {
	TotalRevenue        float64   `json:"total_revenue" db:"total_revenue"`
	AverageOrderValue   float64   `json:"average_order_value" db:"average_order_value"`
	ProjectCount        int       `json:"project_count" db:"project_count"`
	PriceAcceptanceRate float64   `json:"price_acceptance_rate" db:"price_acceptance_rate"`
	ChurnRisk           float64   `json:"churn_risk" db:"churn_risk"`
	LifetimeValue       float64   `json:"lifetime_value" db:"lifetime_value"`
	LastOrderDate       time.Time `json:"last_order_date" db:"last_order_date"`
	PaymentReliability  float64   `json:"payment_reliability" db:"payment_reliability"`
}

// CostModel represents resource cost calculations
type CostModel struct {
	ID                  string                 `json:"id" db:"id"`
	Name                string                 `json:"name" db:"name"`
	ContentType         ContentType            `json:"content_type" db:"content_type"`
	BaseCost            float64                `json:"base_cost" db:"base_cost"`
	LLMCosts            map[string]float64     `json:"llm_costs" db:"llm_costs"`               // Cost per API call by model
	ProcessingCosts     map[string]float64     `json:"processing_costs" db:"processing_costs"` // Cost per minute/operation
	InfrastructureCosts map[string]float64     `json:"infrastructure_costs" db:"infrastructure_costs"`
	OverheadRate        float64                `json:"overhead_rate" db:"overhead_rate"`
	ProfitMargin        float64                `json:"profit_margin" db:"profit_margin"`
	Factors             map[string]interface{} `json:"factors" db:"factors"`
	IsActive            bool                   `json:"is_active" db:"is_active"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// PricingExperiment represents A/B testing for pricing strategies
type PricingExperiment struct {
	ID                string                 `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`
	Hypothesis        string                 `json:"hypothesis" db:"hypothesis"`
	Variants          []PricingVariant       `json:"variants" db:"variants"`
	TargetMetric      string                 `json:"target_metric" db:"target_metric"`
	TargetSegment     map[string]interface{} `json:"target_segment" db:"target_segment"`
	TrafficSplit      map[string]float64     `json:"traffic_split" db:"traffic_split"`
	StartDate         time.Time              `json:"start_date" db:"start_date"`
	EndDate           time.Time              `json:"end_date" db:"end_date"`
	Status            ExperimentStatus       `json:"status" db:"status"`
	Results           *ExperimentResults     `json:"results,omitempty" db:"results"`
	SampleSize        int                    `json:"sample_size" db:"sample_size"`
	SignificanceLevel float64                `json:"significance_level" db:"significance_level"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy         string                 `json:"created_by" db:"created_by"`
}

// PricingVariant represents different pricing approaches in A/B tests
type PricingVariant struct {
	ID             string                 `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	PricingModel   string                 `json:"pricing_model" db:"pricing_model"`
	Adjustments    []PriceAdjustment      `json:"adjustments" db:"adjustments"`
	Configuration  map[string]interface{} `json:"configuration" db:"configuration"`
	TrafficPercent float64                `json:"traffic_percent" db:"traffic_percent"`
	Metrics        VariantMetrics         `json:"metrics" db:"metrics"`
}

// VariantMetrics tracks performance of pricing variants
type VariantMetrics struct {
	QuotesGenerated      int     `json:"quotes_generated" db:"quotes_generated"`
	QuotesAccepted       int     `json:"quotes_accepted" db:"quotes_accepted"`
	AcceptanceRate       float64 `json:"acceptance_rate" db:"acceptance_rate"`
	AveragePrice         float64 `json:"average_price" db:"average_price"`
	TotalRevenue         float64 `json:"total_revenue" db:"total_revenue"`
	ConversionRate       float64 `json:"conversion_rate" db:"conversion_rate"`
	CustomerSatisfaction float64 `json:"customer_satisfaction" db:"customer_satisfaction"`
}

// ExperimentResults contains statistical analysis of pricing experiments
type ExperimentResults struct {
	WinningVariant          string                 `json:"winning_variant" db:"winning_variant"`
	ConfidenceLevel         float64                `json:"confidence_level" db:"confidence_level"`
	StatisticalSignificance bool                   `json:"statistical_significance" db:"statistical_significance"`
	EffectSize              float64                `json:"effect_size" db:"effect_size"`
	Recommendation          string                 `json:"recommendation" db:"recommendation"`
	MetricImprovements      map[string]float64     `json:"metric_improvements" db:"metric_improvements"`
	AnalysisData            map[string]interface{} `json:"analysis_data" db:"analysis_data"`
	AnalyzedAt              time.Time              `json:"analyzed_at" db:"analyzed_at"`
}

// Enums and constants
type QuoteStatus string

const (
	QuoteStatusDraft     QuoteStatus = "draft"
	QuoteStatusPending   QuoteStatus = "pending"
	QuoteStatusSent      QuoteStatus = "sent"
	QuoteStatusAccepted  QuoteStatus = "accepted"
	QuoteStatusRejected  QuoteStatus = "rejected"
	QuoteStatusExpired   QuoteStatus = "expired"
	QuoteStatusCancelled QuoteStatus = "cancelled"
)

type ClientTier string

const (
	ClientTierBasic      ClientTier = "basic"
	ClientTierPremium    ClientTier = "premium"
	ClientTierEnterprise ClientTier = "enterprise"
	ClientTierVIP        ClientTier = "vip"
)

type PaymentTerms string

const (
	PaymentTermsImmediate PaymentTerms = "immediate"
	PaymentTermsNet15     PaymentTerms = "net_15"
	PaymentTermsNet30     PaymentTerms = "net_30"
	PaymentTermsNet60     PaymentTerms = "net_60"
	PaymentTermsCustom    PaymentTerms = "custom"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type DemandLevel string

const (
	DemandLevelVeryLow  DemandLevel = "very_low"
	DemandLevelLow      DemandLevel = "low"
	DemandLevelMedium   DemandLevel = "medium"
	DemandLevelHigh     DemandLevel = "high"
	DemandLevelVeryHigh DemandLevel = "very_high"
)

type TrendDirection string

const (
	TrendDirectionUp       TrendDirection = "up"
	TrendDirectionDown     TrendDirection = "down"
	TrendDirectionStable   TrendDirection = "stable"
	TrendDirectionVolatile TrendDirection = "volatile"
)

type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusActive    ExperimentStatus = "active"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "cancelled"
)

// Helper methods
func (q *PriceQuote) CalculateFinalPrice() {
	q.FinalPrice = q.BasePrice

	// Apply complexity adjustments
	for _, adj := range q.ComplexityAdjustments {
		if adj.Multiplier > 0 {
			q.FinalPrice *= adj.Multiplier
		}
		q.FinalPrice += adj.Amount
	}

	// Apply market adjustments
	for _, adj := range q.MarketAdjustments {
		if adj.Multiplier > 0 {
			q.FinalPrice *= adj.Multiplier
		}
		q.FinalPrice += adj.Amount
	}

	// Apply client adjustments
	for _, adj := range q.ClientAdjustments {
		if adj.Multiplier > 0 {
			q.FinalPrice *= adj.Multiplier
		}
		q.FinalPrice += adj.Amount
	}

	// Apply surge multiplier
	if q.SurgeMultiplier > 0 {
		q.FinalPrice *= q.SurgeMultiplier
	}

	// Apply discount
	q.FinalPrice -= q.DiscountAmount

	// Ensure non-negative price
	if q.FinalPrice < 0 {
		q.FinalPrice = 0
	}
}

func (q *PriceQuote) IsExpired() bool {
	return time.Now().After(q.ValidUntil)
}

func (q *PriceQuote) GetPriceBreakdown() map[string]float64 {
	breakdown := map[string]float64{
		"base_price": q.BasePrice,
	}

	complexityTotal := 0.0
	for _, adj := range q.ComplexityAdjustments {
		adj_amount := adj.Amount
		if adj.Multiplier > 0 {
			adj_amount += q.BasePrice * (adj.Multiplier - 1)
		}
		complexityTotal += adj_amount
	}
	breakdown["complexity_adjustments"] = complexityTotal

	marketTotal := 0.0
	for _, adj := range q.MarketAdjustments {
		adj_amount := adj.Amount
		if adj.Multiplier > 0 {
			adj_amount += q.BasePrice * (adj.Multiplier - 1)
		}
		marketTotal += adj_amount
	}
	breakdown["market_adjustments"] = marketTotal

	clientTotal := 0.0
	for _, adj := range q.ClientAdjustments {
		adj_amount := adj.Amount
		if adj.Multiplier > 0 {
			adj_amount += q.BasePrice * (adj.Multiplier - 1)
		}
		clientTotal += adj_amount
	}
	breakdown["client_adjustments"] = clientTotal

	if q.SurgeMultiplier > 1 {
		breakdown["surge_pricing"] = q.BasePrice * (q.SurgeMultiplier - 1)
	}

	if q.DiscountAmount > 0 {
		breakdown["discount"] = -q.DiscountAmount
	}

	breakdown["final_price"] = q.FinalPrice

	return breakdown
}

func (p *ClientPricingProfile) GetApplicableDiscount(volume int, orderValue float64) float64 {
	discount := p.LoyaltyDiscount

	// Check volume discounts
	for _, vd := range p.VolumeDiscounts {
		if !vd.IsActive {
			continue
		}
		if volume >= vd.MinVolume && (vd.MaxVolume == nil || volume <= *vd.MaxVolume) {
			if vd.DiscountType == "percentage" {
				volDiscount := orderValue * (vd.DiscountRate / 100)
				if volDiscount > discount {
					discount = volDiscount
				}
			} else {
				if vd.DiscountRate > discount {
					discount = vd.DiscountRate
				}
			}
		}
	}

	// Check special offers
	now := time.Now()
	for _, offer := range p.SpecialOffers {
		if !offer.IsActive || now.Before(offer.ValidFrom) || now.After(offer.ValidUntil) {
			continue
		}
		if offer.UsageLimit != nil && offer.UsageCount >= *offer.UsageLimit {
			continue
		}

		offerDiscount := 0.0
		if offer.DiscountType == "percentage" {
			offerDiscount = orderValue * (offer.DiscountRate / 100)
		} else {
			offerDiscount = offer.DiscountRate
		}

		if offerDiscount > discount {
			discount = offerDiscount
		}
	}

	return discount
}

func (m *MarketData) IsStale(maxAge time.Duration) bool {
	return time.Since(m.CollectedAt) > maxAge
}

func (m *MarketData) GetMarketPosition(ourPrice float64) string {
	if ourPrice <= m.MinPrice {
		return "lowest"
	}
	if ourPrice <= m.MedianPrice*0.8 {
		return "below_market"
	}
	if ourPrice <= m.MedianPrice*1.2 {
		return "market_rate"
	}
	if ourPrice <= m.MaxPrice {
		return "above_market"
	}
	return "highest"
}

func (e *PricingExperiment) IsActive() bool {
	now := time.Now()
	return e.Status == ExperimentStatusActive && now.After(e.StartDate) && now.Before(e.EndDate)
}

func (e *PricingExperiment) GetVariantForClient(clientID string) *PricingVariant {
	if !e.IsActive() {
		return nil
	}

	// Use client ID to deterministically assign variant
	hash := 0
	for _, b := range clientID {
		hash = hash*31 + int(b)
	}
	hash = hash % 100
	if hash < 0 {
		hash = -hash
	}

	cumulative := 0.0
	for _, variant := range e.Variants {
		cumulative += variant.TrafficPercent
		if float64(hash) < cumulative {
			return &variant
		}
	}

	return nil
}

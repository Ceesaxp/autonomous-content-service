package pricing

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// DynamicPricingEngineImpl implements DynamicPricingEngine
type DynamicPricingEngineImpl struct {
	pricingModelRepo  repositories.PricingModelRepository
	marketDataRepo    repositories.MarketDataRepository
	clientPricingRepo repositories.ClientPricingRepository
	costModelRepo     repositories.CostModelRepository
}

// NewDynamicPricingEngine creates a new dynamic pricing engine
func NewDynamicPricingEngine(
	pricingModelRepo repositories.PricingModelRepository,
	marketDataRepo repositories.MarketDataRepository,
	clientPricingRepo repositories.ClientPricingRepository,
	costModelRepo repositories.CostModelRepository,
) *DynamicPricingEngineImpl {
	return &DynamicPricingEngineImpl{
		pricingModelRepo:  pricingModelRepo,
		marketDataRepo:    marketDataRepo,
		clientPricingRepo: clientPricingRepo,
		costModelRepo:     costModelRepo,
	}
}

// CalculatePrice performs comprehensive price calculation
func (e *DynamicPricingEngineImpl) CalculatePrice(ctx context.Context, req *CalculatePriceRequest) (*CalculatePriceResponse, error) {
	// Get base pricing model
	pricingModel, err := e.pricingModelRepo.GetPricingModelByContentType(ctx, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing model: %w", err)
	}

	basePrice := pricingModel.BasePrice

	// Apply complexity adjustments
	complexityAdjustments, err := e.ApplyComplexityAdjustments(ctx, basePrice, req.ContentSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to apply complexity adjustments: %w", err)
	}

	// Apply market adjustments
	marketAdjustments, err := e.ApplyMarketAdjustments(ctx, basePrice, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to apply market adjustments: %w", err)
	}

	// Apply client adjustments
	clientAdjustments, err := e.ApplyClientAdjustments(ctx, basePrice, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to apply client adjustments: %w", err)
	}

	// Calculate surge multiplier
	surgeMultiplier, err := e.CalculateSurgeMultiplier(ctx, req.ContentType, req.ExpectedDeliveryTime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate surge multiplier: %w", err)
	}

	// Apply demand-based adjustments
	demandAdjustedPrice, err := e.AdjustPriceForDemand(ctx, basePrice, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to apply demand adjustments: %w", err)
	}

	// Apply capacity-based adjustments
	capacityAdjustedPrice, err := e.AdjustPriceForCapacity(ctx, demandAdjustedPrice, req.CurrentSystemLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to apply capacity adjustments: %w", err)
	}

	// Apply timing-based adjustments
	finalPrice, err := e.AdjustPriceForTiming(ctx, capacityAdjustedPrice, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to apply timing adjustments: %w", err)
	}

	// Apply all adjustments to get final price
	for _, adj := range complexityAdjustments {
		if adj.Multiplier > 0 {
			finalPrice *= adj.Multiplier
		}
		finalPrice += adj.Amount
	}

	for _, adj := range marketAdjustments {
		if adj.Multiplier > 0 {
			finalPrice *= adj.Multiplier
		}
		finalPrice += adj.Amount
	}

	for _, adj := range clientAdjustments {
		if adj.Multiplier > 0 {
			finalPrice *= adj.Multiplier
		}
		finalPrice += adj.Amount
	}

	if surgeMultiplier > 1 {
		finalPrice *= surgeMultiplier
	}

	return &CalculatePriceResponse{
		BasePrice:             basePrice,
		ComplexityAdjustments: complexityAdjustments,
		MarketAdjustments:     marketAdjustments,
		ClientAdjustments:     clientAdjustments,
		SurgeMultiplier:       surgeMultiplier,
		FinalPrice:            finalPrice,
		Currency:              pricingModel.Currency,
		ConfidenceLevel:       0.85, // Default confidence
		PriceBreakdown: map[string]float64{
			"base_price":   basePrice,
			"final_price":  finalPrice,
			"surge_factor": surgeMultiplier,
			"adjustments":  finalPrice - basePrice,
		},
	}, nil
}

// ApplyComplexityAdjustments calculates price adjustments based on content complexity
func (e *DynamicPricingEngineImpl) ApplyComplexityAdjustments(ctx context.Context, basePrice float64, contentSpec *ContentSpecification) ([]entities.PriceAdjustment, error) {
	if contentSpec == nil {
		return []entities.PriceAdjustment{}, nil
	}

	var adjustments []entities.PriceAdjustment

	// Word count adjustment
	if contentSpec.WordCount > 1000 {
		multiplier := 1.0 + float64(contentSpec.WordCount-1000)/1000*0.1 // 10% per 1000 words
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "complexity",
			Reason:      "word_count",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Word count adjustment for %d words", contentSpec.WordCount),
		})
	}

	// Complexity level adjustment
	complexityMultipliers := map[string]float64{
		"basic":        1.0,
		"intermediate": 1.2,
		"advanced":     1.5,
		"expert":       2.0,
	}

	if multiplier, exists := complexityMultipliers[contentSpec.ComplexityLevel]; exists && multiplier > 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "complexity",
			Reason:      "complexity_level",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Complexity level adjustment for %s", contentSpec.ComplexityLevel),
		})
	}

	// Research depth adjustment
	researchMultipliers := map[string]float64{
		"minimal":       1.0,
		"basic":         1.1,
		"thorough":      1.3,
		"extensive":     1.6,
		"comprehensive": 2.0,
	}

	if multiplier, exists := researchMultipliers[contentSpec.ResearchDepth]; exists && multiplier > 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "complexity",
			Reason:      "research_depth",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Research depth adjustment for %s", contentSpec.ResearchDepth),
		})
	}

	// Technical level adjustment
	technicalMultipliers := map[string]float64{
		"general":     1.0,
		"technical":   1.3,
		"specialized": 1.7,
		"expert":      2.2,
	}

	if multiplier, exists := technicalMultipliers[contentSpec.TechnicalLevel]; exists && multiplier > 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "complexity",
			Reason:      "technical_level",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Technical level adjustment for %s", contentSpec.TechnicalLevel),
		})
	}

	// Special requirements adjustment
	if len(contentSpec.Requirements) > 3 {
		additionalCost := basePrice * 0.1 * float64(len(contentSpec.Requirements)-3) // 10% per extra requirement
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "complexity",
			Reason:      "special_requirements",
			Amount:      additionalCost,
			Multiplier:  1.0,
			Description: fmt.Sprintf("Special requirements adjustment for %d requirements", len(contentSpec.Requirements)),
		})
	}

	return adjustments, nil
}

// ApplyMarketAdjustments calculates price adjustments based on market conditions
func (e *DynamicPricingEngineImpl) ApplyMarketAdjustments(ctx context.Context, basePrice float64, contentType entities.ContentType) ([]entities.PriceAdjustment, error) {
	marketData, err := e.marketDataRepo.GetLatestMarketData(ctx, contentType, "general")
	if err != nil {
		return []entities.PriceAdjustment{}, nil // No market adjustment if no data
	}

	if marketData.IsStale(24 * time.Hour) {
		return []entities.PriceAdjustment{}, nil // Don't use stale data
	}

	var adjustments []entities.PriceAdjustment

	// Demand level adjustment
	demandMultipliers := map[entities.DemandLevel]float64{
		entities.DemandLevelVeryLow:  0.9,
		entities.DemandLevelLow:      0.95,
		entities.DemandLevelMedium:   1.0,
		entities.DemandLevelHigh:     1.1,
		entities.DemandLevelVeryHigh: 1.25,
	}

	if multiplier, exists := demandMultipliers[marketData.DemandLevel]; exists && multiplier != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "market",
			Reason:      "demand_level",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Demand level adjustment for %s demand", marketData.DemandLevel),
		})
	}

	// Market position adjustment
	position := marketData.GetMarketPosition(basePrice)
	positionAdjustments := map[string]float64{
		"lowest":       1.0,  // No adjustment for being lowest
		"below_market": 1.05, // Small premium if below market
		"market_rate":  1.0,  // No adjustment for market rate
		"above_market": 0.95, // Small discount if above market
		"highest":      0.9,  // Larger discount if highest
	}

	if adjustment, exists := positionAdjustments[position]; exists && adjustment != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "market",
			Reason:      "market_position",
			Amount:      0,
			Multiplier:  adjustment,
			Description: fmt.Sprintf("Market position adjustment for %s position", position),
		})
	}

	// Trend direction adjustment
	trendAdjustments := map[entities.TrendDirection]float64{
		entities.TrendDirectionDown:     0.95,
		entities.TrendDirectionStable:   1.0,
		entities.TrendDirectionUp:       1.05,
		entities.TrendDirectionVolatile: 1.0, // No adjustment for volatile
	}

	if multiplier, exists := trendAdjustments[marketData.TrendDirection]; exists && multiplier != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "market",
			Reason:      "trend_direction",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Trend direction adjustment for %s trend", marketData.TrendDirection),
		})
	}

	return adjustments, nil
}

// ApplyClientAdjustments calculates price adjustments based on client profile
func (e *DynamicPricingEngineImpl) ApplyClientAdjustments(ctx context.Context, basePrice float64, clientID string) ([]entities.PriceAdjustment, error) {
	profile, err := e.clientPricingRepo.GetClientPricingProfileByClient(ctx, clientID)
	if err != nil {
		return []entities.PriceAdjustment{}, nil // No client adjustment if no profile
	}

	var adjustments []entities.PriceAdjustment

	// Client tier adjustment
	tierMultipliers := map[entities.ClientTier]float64{
		entities.ClientTierBasic:      1.0,
		entities.ClientTierPremium:    0.95,
		entities.ClientTierEnterprise: 0.9,
		entities.ClientTierVIP:        0.85,
	}

	if multiplier, exists := tierMultipliers[profile.Tier]; exists && multiplier != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "client",
			Reason:      "tier_discount",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Client tier discount for %s tier", profile.Tier),
		})
	}

	// Risk level adjustment
	riskMultipliers := map[entities.RiskLevel]float64{
		entities.RiskLevelLow:      0.98,
		entities.RiskLevelMedium:   1.0,
		entities.RiskLevelHigh:     1.05,
		entities.RiskLevelCritical: 1.15,
	}

	if multiplier, exists := riskMultipliers[profile.RiskLevel]; exists && multiplier != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "client",
			Reason:      "risk_adjustment",
			Amount:      0,
			Multiplier:  multiplier,
			Description: fmt.Sprintf("Risk level adjustment for %s risk", profile.RiskLevel),
		})
	}

	// Payment terms adjustment
	paymentAdjustments := map[entities.PaymentTerms]float64{
		entities.PaymentTermsImmediate: 0.98, // Discount for immediate payment
		entities.PaymentTermsNet15:     1.0,
		entities.PaymentTermsNet30:     1.02,
		entities.PaymentTermsNet60:     1.05,
		entities.PaymentTermsCustom:    1.0,
	}

	if adjustment, exists := paymentAdjustments[profile.PaymentTerms]; exists && adjustment != 1.0 {
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "client",
			Reason:      "payment_terms",
			Amount:      0,
			Multiplier:  adjustment,
			Description: fmt.Sprintf("Payment terms adjustment for %s", profile.PaymentTerms),
		})
	}

	// Loyalty discount
	if profile.LoyaltyDiscount > 0 {
		discountAmount := basePrice * (profile.LoyaltyDiscount / 100)
		adjustments = append(adjustments, entities.PriceAdjustment{
			Type:        "client",
			Reason:      "loyalty_discount",
			Amount:      -discountAmount,
			Multiplier:  1.0,
			Description: fmt.Sprintf("Loyalty discount of %.1f%%", profile.LoyaltyDiscount),
		})
	}

	return adjustments, nil
}

// CalculateSurgeMultiplier calculates surge pricing based on delivery urgency
func (e *DynamicPricingEngineImpl) CalculateSurgeMultiplier(ctx context.Context, contentType entities.ContentType, deliveryTime time.Duration) (float64, error) {
	// Standard delivery times by content type (in hours)
	standardDeliveryTimes := map[entities.ContentType]time.Duration{
		entities.ContentTypeBlogPost: 24 * time.Hour,
		//sentities.ContentTypeArticle:         48 * time.Hour,
		// entities.ContentTypeWhitepaper:      72 * time.Hour,
		// entities.ContentTypeWebsiteContent:  12 * time.Hour,
		// entities.ContentTypeSocialMediaPost: 4 * time.Hour,
		// entities.ContentTypeNewsletter:      24 * time.Hour,
		// entities.ContentTypeEbook:           168 * time.Hour, // 1 week
	}

	standardTime, exists := standardDeliveryTimes[contentType]
	if !exists {
		standardTime = 24 * time.Hour // Default
	}

	// No surge if delivery time is standard or longer
	if deliveryTime >= standardTime {
		return 1.0, nil
	}

	// Calculate surge multiplier based on urgency
	urgencyRatio := float64(deliveryTime) / float64(standardTime)

	if urgencyRatio > 0.75 {
		return 1.0, nil // No surge for > 75% of standard time
	} else if urgencyRatio > 0.5 {
		return 1.2, nil // 20% surge for 50-75% of standard time
	} else if urgencyRatio > 0.25 {
		return 1.5, nil // 50% surge for 25-50% of standard time
	} else {
		return 2.0, nil // 100% surge for < 25% of standard time
	}
}

// AdjustPriceForDemand adjusts price based on current demand levels
func (e *DynamicPricingEngineImpl) AdjustPriceForDemand(ctx context.Context, basePrice float64, contentType entities.ContentType) (float64, error) {
	marketData, err := e.marketDataRepo.GetLatestMarketData(ctx, contentType, "general")
	if err != nil {
		return basePrice, nil // No adjustment if no market data
	}

	demandAdjustments := map[entities.DemandLevel]float64{
		entities.DemandLevelVeryLow:  0.9,
		entities.DemandLevelLow:      0.95,
		entities.DemandLevelMedium:   1.0,
		entities.DemandLevelHigh:     1.1,
		entities.DemandLevelVeryHigh: 1.25,
	}

	if adjustment, exists := demandAdjustments[marketData.DemandLevel]; exists {
		return basePrice * adjustment, nil
	}

	return basePrice, nil
}

// AdjustPriceForCapacity adjusts price based on current system capacity
func (e *DynamicPricingEngineImpl) AdjustPriceForCapacity(ctx context.Context, basePrice float64, currentLoad float64) (float64, error) {
	// Adjust pricing based on system capacity utilization
	if currentLoad < 0.5 {
		return basePrice * 0.95, nil // 5% discount for low load
	} else if currentLoad < 0.7 {
		return basePrice, nil // No adjustment for normal load
	} else if currentLoad < 0.85 {
		return basePrice * 1.1, nil // 10% premium for high load
	} else {
		return basePrice * 1.25, nil // 25% premium for very high load
	}
}

// AdjustPriceForTiming adjusts price based on time of request
func (e *DynamicPricingEngineImpl) AdjustPriceForTiming(ctx context.Context, basePrice float64, requestTime time.Time) (float64, error) {
	// Time-based pricing adjustments
	hour := requestTime.Hour()
	weekday := requestTime.Weekday()

	// Weekend premium
	if weekday == time.Saturday || weekday == time.Sunday {
		basePrice *= 1.15 // 15% weekend premium
	}

	// Off-hours premium (outside 9 AM - 6 PM)
	if hour < 9 || hour > 18 {
		basePrice *= 1.1 // 10% off-hours premium
	}

	// Holiday check (simplified - would integrate with holiday calendar)
	// This is a placeholder for more sophisticated holiday detection
	if isHoliday(requestTime) {
		basePrice *= 1.2 // 20% holiday premium
	}

	return basePrice, nil
}

// OptimizePriceForRevenue optimizes price to maximize revenue
func (e *DynamicPricingEngineImpl) OptimizePriceForRevenue(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error) {
	// Get price elasticity data
	elasticity, err := e.marketDataRepo.GetPriceElasticity(ctx, req.ContentType, repositories.TimeRange{
		Start: time.Now().AddDate(0, -3, 0), // Last 3 months
		End:   time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get price elasticity: %w", err)
	}

	// Calculate optimal price using elasticity
	elasticityScore := math.Abs(elasticity.ElasticityScore)

	var optimalPrice float64
	if elasticityScore > 1.0 {
		// Elastic demand - lower price to increase volume
		optimalPrice = req.CurrentPrice * 0.9
	} else if elasticityScore < 0.5 {
		// Inelastic demand - can increase price
		optimalPrice = req.CurrentPrice * 1.1
	} else {
		// Unit elastic - optimal price is current
		optimalPrice = req.CurrentPrice
	}

	// Calculate expected impact
	priceChange := (optimalPrice - req.CurrentPrice) / req.CurrentPrice
	volumeChange := -elasticityScore * priceChange
	revenueChange := (1+priceChange)*(1+volumeChange) - 1

	return &PriceOptimizationResponse{
		OptimalPrice:          optimalPrice,
		PriceChange:           priceChange,
		ExpectedVolumeChange:  volumeChange,
		ExpectedRevenueChange: revenueChange,
		Confidence:            elasticity.ConfidenceLevel,
		Recommendations: []string{
			fmt.Sprintf("Adjust price from $%.2f to $%.2f", req.CurrentPrice, optimalPrice),
			fmt.Sprintf("Expected revenue change: %.1f%%", revenueChange*100),
		},
	}, nil
}

// OptimizePriceForConversion optimizes price to maximize conversion
func (e *DynamicPricingEngineImpl) OptimizePriceForConversion(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error) {
	// For conversion optimization, we generally want to find the sweet spot
	// where price is low enough to encourage conversion but high enough to maintain value

	// Get market data for benchmarking
	marketData, err := e.marketDataRepo.GetLatestMarketData(ctx, req.ContentType, "general")
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Position slightly below median to encourage conversion
	optimalPrice := marketData.MedianPrice * 0.95

	priceChange := (optimalPrice - req.CurrentPrice) / req.CurrentPrice

	return &PriceOptimizationResponse{
		OptimalPrice:          optimalPrice,
		PriceChange:           priceChange,
		ExpectedVolumeChange:  math.Abs(priceChange) * 1.5, // Assume volume increases 1.5x price decrease
		ExpectedRevenueChange: (optimalPrice/req.CurrentPrice - 1) * 1.5,
		Confidence:            marketData.ConfidenceScore,
		Recommendations: []string{
			"Price positioned for conversion optimization",
			fmt.Sprintf("Set price 5%% below market median (%.2f)", marketData.MedianPrice),
		},
	}, nil
}

// OptimizePriceForMarketShare optimizes price to maximize market share
func (e *DynamicPricingEngineImpl) OptimizePriceForMarketShare(ctx context.Context, req *PriceOptimizationRequest) (*PriceOptimizationResponse, error) {
	// Get competitor analysis
	competitorAnalysis, err := e.marketDataRepo.GetCompetitorAnalysis(ctx, repositories.CompetitorAnalysisFilter{
		ContentType: &req.ContentType,
		TimeRange: &repositories.TimeRange{
			Start: time.Now().AddDate(0, -1, 0), // Last month
			End:   time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get competitor analysis: %w", err)
	}

	// Price below market average to gain market share
	optimalPrice := competitorAnalysis.MarketAveragePrice * 0.85

	priceChange := (optimalPrice - req.CurrentPrice) / req.CurrentPrice

	return &PriceOptimizationResponse{
		OptimalPrice:          optimalPrice,
		PriceChange:           priceChange,
		ExpectedVolumeChange:  math.Abs(priceChange) * 2.0, // Aggressive volume gain
		ExpectedRevenueChange: (optimalPrice/req.CurrentPrice - 1) * 2.0,
		Confidence:            0.7, // Lower confidence for aggressive strategy
		Recommendations: []string{
			"Aggressive pricing for market share capture",
			fmt.Sprintf("Price 15%% below market average (%.2f)", competitorAnalysis.MarketAveragePrice),
			"Monitor competitor responses closely",
		},
	}, nil
}

// Helper functions
func isHoliday(t time.Time) bool {
	// Simplified holiday detection - would typically integrate with holiday calendar service
	// Check for major US holidays (simplified)
	month := t.Month()
	day := t.Day()

	// New Year's Day
	if month == time.January && day == 1 {
		return true
	}

	// Christmas
	if month == time.December && day == 25 {
		return true
	}

	// Independence Day
	if month == time.July && day == 4 {
		return true
	}

	return false
}

// Additional request/response types for the engine
type CalculatePriceRequest struct {
	ClientID             string                `json:"client_id"`
	ContentType          entities.ContentType  `json:"content_type"`
	ContentSpec          *ContentSpecification `json:"content_spec"`
	ExpectedDeliveryTime time.Duration         `json:"expected_delivery_time"`
	CurrentSystemLoad    float64               `json:"current_system_load"`
	RequestTime          time.Time             `json:"request_time"`
}

type CalculatePriceResponse struct {
	BasePrice             float64                    `json:"base_price"`
	ComplexityAdjustments []entities.PriceAdjustment `json:"complexity_adjustments"`
	MarketAdjustments     []entities.PriceAdjustment `json:"market_adjustments"`
	ClientAdjustments     []entities.PriceAdjustment `json:"client_adjustments"`
	SurgeMultiplier       float64                    `json:"surge_multiplier"`
	FinalPrice            float64                    `json:"final_price"`
	Currency              string                     `json:"currency"`
	ConfidenceLevel       float64                    `json:"confidence_level"`
	PriceBreakdown        map[string]float64         `json:"price_breakdown"`
}

type PriceOptimizationRequest struct {
	ContentType  entities.ContentType `json:"content_type"`
	CurrentPrice float64              `json:"current_price"`
	TargetMetric string               `json:"target_metric"` // revenue, conversion, market_share
	Constraints  map[string]float64   `json:"constraints"`   // min_price, max_price, etc.
}

type PriceOptimizationResponse struct {
	OptimalPrice          float64  `json:"optimal_price"`
	PriceChange           float64  `json:"price_change"`
	ExpectedVolumeChange  float64  `json:"expected_volume_change"`
	ExpectedRevenueChange float64  `json:"expected_revenue_change"`
	Confidence            float64  `json:"confidence"`
	Recommendations       []string `json:"recommendations"`
	RiskFactors           []string `json:"risk_factors"`
}

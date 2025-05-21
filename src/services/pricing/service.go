package pricing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// PricingServiceImpl implements the PricingService interface
type PricingServiceImpl struct {
	pricingModelRepo   repositories.PricingModelRepository
	priceQuoteRepo     repositories.PriceQuoteRepository
	marketDataRepo     repositories.MarketDataRepository
	clientPricingRepo  repositories.ClientPricingRepository
	costModelRepo      repositories.CostModelRepository
	experimentRepo     repositories.PricingExperimentRepository
	dynamicEngine      DynamicPricingEngine
	marketIntelligence MarketIntelligenceService
	costCalculator     CostCalculationEngine
	experimentService  PricingExperimentService
}

// NewPricingService creates a new pricing service
func NewPricingService(
	pricingModelRepo repositories.PricingModelRepository,
	priceQuoteRepo repositories.PriceQuoteRepository,
	marketDataRepo repositories.MarketDataRepository,
	clientPricingRepo repositories.ClientPricingRepository,
	costModelRepo repositories.CostModelRepository,
	experimentRepo repositories.PricingExperimentRepository,
	dynamicEngine DynamicPricingEngine,
	marketIntelligence MarketIntelligenceService,
	costCalculator CostCalculationEngine,
	experimentService PricingExperimentService,
) *PricingServiceImpl {
	return &PricingServiceImpl{
		pricingModelRepo:   pricingModelRepo,
		priceQuoteRepo:     priceQuoteRepo,
		marketDataRepo:     marketDataRepo,
		clientPricingRepo:  clientPricingRepo,
		costModelRepo:      costModelRepo,
		experimentRepo:     experimentRepo,
		dynamicEngine:      dynamicEngine,
		marketIntelligence: marketIntelligence,
		costCalculator:     costCalculator,
		experimentService:  experimentService,
	}
}

// GeneratePriceQuote creates a comprehensive price quote
func (s *PricingServiceImpl) GeneratePriceQuote(ctx context.Context, req *GeneratePriceQuoteRequest) (*entities.PriceQuote, error) {
	// Get pricing model for content type
	pricingModel, err := s.pricingModelRepo.GetPricingModelByContentType(ctx, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing model: %w", err)
	}

	if !pricingModel.IsActive {
		return nil, fmt.Errorf("pricing model not active for content type: %s", req.ContentType)
	}

	// Start with base price
	basePrice := pricingModel.BasePrice

	// Initialize quote
	quote := &entities.PriceQuote{
		ID:          generateID(),
		ProjectID:   req.ProjectID,
		ClientID:    req.ClientID,
		ContentType: req.ContentType,
		BasePrice:   basePrice,
		Currency:    pricingModel.Currency,
		Status:      entities.QuoteStatusDraft,
		ValidUntil:  time.Now().Add(7 * 24 * time.Hour), // 7 days validity
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Apply complexity adjustments
	if req.ContentSpec != nil {
		complexityAdjustments, err := s.dynamicEngine.ApplyComplexityAdjustments(ctx, basePrice, req.ContentSpec)
		if err != nil {
			log.Printf("Error calculating complexity adjustments: %v", err)
		} else {
			quote.ComplexityAdjustments = complexityAdjustments
		}
	}

	// Apply market adjustments
	marketAdjustments, err := s.dynamicEngine.ApplyMarketAdjustments(ctx, basePrice, req.ContentType)
	if err != nil {
		log.Printf("Error calculating market adjustments: %v", err)
	} else {
		quote.MarketAdjustments = marketAdjustments
	}

	// Apply client-specific adjustments
	clientAdjustments, err := s.dynamicEngine.ApplyClientAdjustments(ctx, basePrice, req.ClientID)
	if err != nil {
		log.Printf("Error calculating client adjustments: %v", err)
	} else {
		quote.ClientAdjustments = clientAdjustments
	}

	// Calculate surge pricing
	if req.DeliveryTime > 0 {
		surgeMultiplier, err := s.dynamicEngine.CalculateSurgeMultiplier(ctx, req.ContentType, req.DeliveryTime)
		if err != nil {
			log.Printf("Error calculating surge multiplier: %v", err)
		} else {
			quote.SurgeMultiplier = surgeMultiplier
		}
	}

	// Calculate client discount
	discount, err := s.CalculateClientDiscount(ctx, req.ClientID, basePrice, 1)
	if err != nil {
		log.Printf("Error calculating client discount: %v", err)
	} else {
		quote.DiscountAmount = discount
	}

	// Check for active experiments
	variant, err := s.GetExperimentVariantForClient(ctx, req.ClientID, req.ContentType)
	if err == nil && variant != nil {
		// Apply experimental adjustments
		quote.ComplexityAdjustments = append(quote.ComplexityAdjustments, variant.Adjustments...)
		quote.Metadata["experiment_variant"] = variant.ID
	}

	// Calculate final price
	quote.CalculateFinalPrice()

	// Store the quote
	if err := s.priceQuoteRepo.CreatePriceQuote(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to create price quote: %w", err)
	}

	return quote, nil
}

// UpdatePriceQuote updates an existing price quote
func (s *PricingServiceImpl) UpdatePriceQuote(ctx context.Context, quoteID string, updates *UpdatePriceQuoteRequest) (*entities.PriceQuote, error) {
	quote, err := s.priceQuoteRepo.GetPriceQuote(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price quote: %w", err)
	}

	// Apply updates
	if updates.Status != nil {
		quote.Status = *updates.Status
	}
	if updates.ValidUntil != nil {
		quote.ValidUntil = *updates.ValidUntil
	}
	if updates.DiscountAmount != nil {
		quote.DiscountAmount = *updates.DiscountAmount
		quote.CalculateFinalPrice()
	}
	if updates.Notes != nil {
		quote.Metadata["notes"] = *updates.Notes
	}

	quote.UpdatedAt = time.Now()

	if err := s.priceQuoteRepo.UpdatePriceQuote(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to update price quote: %w", err)
	}

	return quote, nil
}

// GetPriceQuote retrieves a price quote by ID
func (s *PricingServiceImpl) GetPriceQuote(ctx context.Context, quoteID string) (*entities.PriceQuote, error) {
	return s.priceQuoteRepo.GetPriceQuote(ctx, quoteID)
}

// AcceptPriceQuote marks a quote as accepted
func (s *PricingServiceImpl) AcceptPriceQuote(ctx context.Context, quoteID string) error {
	_, err := s.UpdatePriceQuote(ctx, quoteID, &UpdatePriceQuoteRequest{
		Status: &[]entities.QuoteStatus{entities.QuoteStatusAccepted}[0],
	})
	return err
}

// RejectPriceQuote marks a quote as rejected
func (s *PricingServiceImpl) RejectPriceQuote(ctx context.Context, quoteID string, reason string) error {
	_, err := s.UpdatePriceQuote(ctx, quoteID, &UpdatePriceQuoteRequest{
		Status: &[]entities.QuoteStatus{entities.QuoteStatusRejected}[0],
		Notes:  &reason,
	})
	return err
}

// ExpirePriceQuote marks a quote as expired
func (s *PricingServiceImpl) ExpirePriceQuote(ctx context.Context, quoteID string) error {
	_, err := s.UpdatePriceQuote(ctx, quoteID, &UpdatePriceQuoteRequest{
		Status: &[]entities.QuoteStatus{entities.QuoteStatusExpired}[0],
	})
	return err
}

// CreatePricingModel creates a new pricing model
func (s *PricingServiceImpl) CreatePricingModel(ctx context.Context, req *CreatePricingModelRequest) (*entities.PricingModel, error) {
	model := &entities.PricingModel{
		ID:              generateID(),
		Name:            req.Name,
		ContentType:     req.ContentType,
		BasePrice:       req.BasePrice,
		Currency:        req.Currency,
		ComplexityRules: req.ComplexityRules,
		PricingFactors:  req.PricingFactors,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       "system",
	}

	if err := s.pricingModelRepo.CreatePricingModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to create pricing model: %w", err)
	}

	return model, nil
}

// UpdatePricingModel updates an existing pricing model
func (s *PricingServiceImpl) UpdatePricingModel(ctx context.Context, modelID string, req *UpdatePricingModelRequest) (*entities.PricingModel, error) {
	model, err := s.pricingModelRepo.GetPricingModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing model: %w", err)
	}

	// Apply updates
	if req.Name != nil {
		model.Name = *req.Name
	}
	if req.BasePrice != nil {
		model.BasePrice = *req.BasePrice
	}
	if req.PricingFactors != nil {
		model.PricingFactors = req.PricingFactors
	}
	if req.IsActive != nil {
		model.IsActive = *req.IsActive
	}

	model.UpdatedAt = time.Now()

	if err := s.pricingModelRepo.UpdatePricingModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to update pricing model: %w", err)
	}

	return model, nil
}

// GetPricingModel retrieves a pricing model by ID
func (s *PricingServiceImpl) GetPricingModel(ctx context.Context, modelID string) (*entities.PricingModel, error) {
	return s.pricingModelRepo.GetPricingModel(ctx, modelID)
}

// ListPricingModels lists pricing models with filtering
func (s *PricingServiceImpl) ListPricingModels(ctx context.Context, filter *PricingModelFilter) ([]*entities.PricingModel, error) {
	repoFilter := repositories.PricingModelFilter{
		ContentType: filter.ContentType,
		IsActive:    filter.IsActive,
		CreatedBy:   filter.CreatedBy,
		TimeRange:   filter.TimeRange,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
	}
	return s.pricingModelRepo.ListPricingModels(ctx, repoFilter)
}

// CalculateClientDiscount calculates applicable discount for a client
func (s *PricingServiceImpl) CalculateClientDiscount(ctx context.Context, clientID string, orderValue float64, volume int) (float64, error) {
	profile, err := s.clientPricingRepo.GetClientPricingProfileByClient(ctx, clientID)
	if err != nil {
		return 0, nil // No discount if no profile found
	}

	return profile.GetApplicableDiscount(volume, orderValue), nil
}

// CreateClientPricingProfile creates a client pricing profile
func (s *PricingServiceImpl) CreateClientPricingProfile(ctx context.Context, req *CreateClientPricingProfileRequest) (*entities.ClientPricingProfile, error) {
	profile := &entities.ClientPricingProfile{
		ID:              generateID(),
		ClientID:        req.ClientID,
		Tier:            req.Tier,
		VolumeDiscounts: req.VolumeDiscounts,
		LoyaltyDiscount: req.LoyaltyDiscount,
		CustomRates:     req.CustomRates,
		PaymentTerms:    req.PaymentTerms,
		CreditLimit:     req.CreditLimit,
		RiskLevel:       req.RiskLevel,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.clientPricingRepo.CreateClientPricingProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create client pricing profile: %w", err)
	}

	return profile, nil
}

// UpdateClientPricingProfile updates a client pricing profile
func (s *PricingServiceImpl) UpdateClientPricingProfile(ctx context.Context, profileID string, req *UpdateClientPricingProfileRequest) (*entities.ClientPricingProfile, error) {
	profile, err := s.clientPricingRepo.GetClientPricingProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client pricing profile: %w", err)
	}

	// Apply updates
	if req.Tier != nil {
		profile.Tier = *req.Tier
	}
	if req.LoyaltyDiscount != nil {
		profile.LoyaltyDiscount = *req.LoyaltyDiscount
	}
	if req.CustomRates != nil {
		profile.CustomRates = req.CustomRates
	}
	if req.PaymentTerms != nil {
		profile.PaymentTerms = *req.PaymentTerms
	}
	if req.CreditLimit != nil {
		profile.CreditLimit = *req.CreditLimit
	}
	if req.RiskLevel != nil {
		profile.RiskLevel = *req.RiskLevel
	}
	if req.IsActive != nil {
		profile.IsActive = *req.IsActive
	}

	profile.UpdatedAt = time.Now()

	if err := s.clientPricingRepo.UpdateClientPricingProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update client pricing profile: %w", err)
	}

	return profile, nil
}

// GetClientPricingProfile retrieves a client pricing profile
func (s *PricingServiceImpl) GetClientPricingProfile(ctx context.Context, clientID string) (*entities.ClientPricingProfile, error) {
	return s.clientPricingRepo.GetClientPricingProfileByClient(ctx, clientID)
}

// UpdateMarketData updates market intelligence data
func (s *PricingServiceImpl) UpdateMarketData(ctx context.Context, req *UpdateMarketDataRequest) (*entities.MarketData, error) {
	marketData := &entities.MarketData{
		ID:              generateID(),
		ContentType:     req.ContentType,
		MarketSegment:   req.MarketSegment,
		AveragePrice:    req.AveragePrice,
		MedianPrice:     req.MedianPrice,
		MinPrice:        req.MinPrice,
		MaxPrice:        req.MaxPrice,
		SampleSize:      req.SampleSize,
		CompetitorData:  req.CompetitorData,
		DemandLevel:     req.DemandLevel,
		TrendDirection:  req.TrendDirection,
		ConfidenceScore: req.ConfidenceScore,
		DataSource:      req.DataSource,
		CollectedAt:     time.Now(),
		ValidUntil:      time.Now().Add(24 * time.Hour),
	}

	if err := s.marketDataRepo.CreateMarketData(ctx, marketData); err != nil {
		return nil, fmt.Errorf("failed to create market data: %w", err)
	}

	return marketData, nil
}

// GetLatestMarketData retrieves the latest market data
func (s *PricingServiceImpl) GetLatestMarketData(ctx context.Context, contentType entities.ContentType, segment string) (*entities.MarketData, error) {
	return s.marketDataRepo.GetLatestMarketData(ctx, contentType, segment)
}

// AnalyzeCompetitorPricing performs competitor analysis
func (s *PricingServiceImpl) AnalyzeCompetitorPricing(ctx context.Context, req *CompetitorAnalysisRequest) (*repositories.CompetitorAnalysisResult, error) {
	filter := repositories.CompetitorAnalysisFilter{
		ContentType:   &req.ContentType,
		CompetitorIDs: req.CompetitorIDs,
		TimeRange:     &req.TimeRange,
	}
	return s.marketDataRepo.GetCompetitorAnalysis(ctx, filter)
}

// CalculatePriceElasticity calculates price elasticity for a content type
func (s *PricingServiceImpl) CalculatePriceElasticity(ctx context.Context, contentType entities.ContentType, timeRange repositories.TimeRange) (*repositories.PriceElasticityResult, error) {
	return s.marketDataRepo.GetPriceElasticity(ctx, contentType, timeRange)
}

// Delegate to cost calculation engine
func (s *PricingServiceImpl) CalculateResourceCost(ctx context.Context, req *CalculateResourceCostRequest) (*CalculateResourceCostResponse, error) {
	return s.costCalculator.CalculateContentCreationCost(ctx, &ContentCreationCostRequest{
		ProjectID:     req.ProjectID,
		ContentType:   req.ContentType,
		ContentSpec:   req.ContentSpec,
		ResourceUsage: req.ResourceUsage,
	})
}

func (s *PricingServiceImpl) RecordResourceUsage(ctx context.Context, req *RecordResourceUsageRequest) error {
	return s.costCalculator.TrackResourceUsage(ctx, &TrackResourceUsageRequest{
		ProjectID:    req.ProjectID,
		ResourceType: req.ResourceType,
		Usage:        req.Quantity,
		Unit:         req.Unit,
		Cost:         req.Cost,
		Timestamp:    time.Now(),
		Metadata:     req.Metadata,
	})
}

func (s *PricingServiceImpl) GetCostAnalysis(ctx context.Context, req *CostAnalysisRequest) (*repositories.CostAnalysisResult, error) {
	filter := repositories.CostAnalysisFilter{
		ContentType: req.ContentType,
		TimeRange:   &req.TimeRange,
		GroupBy:     req.GroupBy,
	}
	return s.costModelRepo.GetCostAnalysis(ctx, filter)
}

func (s *PricingServiceImpl) GetProfitabilityReport(ctx context.Context, req *ProfitabilityRequest) (*repositories.ProfitabilityResult, error) {
	filter := repositories.ProfitabilityFilter{
		ContentType: req.ContentType,
		ClientID:    req.ClientID,
		TimeRange:   &req.TimeRange,
		GroupBy:     req.GroupBy,
	}
	return s.costModelRepo.GetProfitabilityReport(ctx, filter)
}

// Delegate to experiment service
func (s *PricingServiceImpl) CreatePricingExperiment(ctx context.Context, req *CreatePricingExperimentRequest) (*entities.PricingExperiment, error) {
	experiment := &entities.PricingExperiment{
		ID:                generateID(),
		Name:              req.Name,
		Description:       req.Description,
		Hypothesis:        req.Hypothesis,
		Variants:          req.Variants,
		TargetMetric:      req.TargetMetric,
		TargetSegment:     req.TargetSegment,
		TrafficSplit:      req.TrafficSplit,
		StartDate:         time.Now(),
		EndDate:           time.Now().Add(req.Duration),
		Status:            entities.ExperimentStatusDraft,
		SampleSize:        req.SampleSize,
		SignificanceLevel: req.SignificanceLevel,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         "system",
	}

	if err := s.experimentRepo.CreatePricingExperiment(ctx, experiment); err != nil {
		return nil, fmt.Errorf("failed to create pricing experiment: %w", err)
	}

	return experiment, nil
}

func (s *PricingServiceImpl) UpdatePricingExperiment(ctx context.Context, experimentID string, req *UpdatePricingExperimentRequest) (*entities.PricingExperiment, error) {
	experiment, err := s.experimentRepo.GetPricingExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing experiment: %w", err)
	}

	// Apply updates
	if req.Name != nil {
		experiment.Name = *req.Name
	}
	if req.Description != nil {
		experiment.Description = *req.Description
	}
	if req.Status != nil {
		experiment.Status = *req.Status
	}
	if req.EndDate != nil {
		experiment.EndDate = *req.EndDate
	}
	if req.TrafficSplit != nil {
		experiment.TrafficSplit = req.TrafficSplit
	}

	experiment.UpdatedAt = time.Now()

	if err := s.experimentRepo.UpdatePricingExperiment(ctx, experiment); err != nil {
		return nil, fmt.Errorf("failed to update pricing experiment: %w", err)
	}

	return experiment, nil
}

func (s *PricingServiceImpl) GetPricingExperiment(ctx context.Context, experimentID string) (*entities.PricingExperiment, error) {
	return s.experimentRepo.GetPricingExperiment(ctx, experimentID)
}

func (s *PricingServiceImpl) StartPricingExperiment(ctx context.Context, experimentID string) error {
	_, err := s.UpdatePricingExperiment(ctx, experimentID, &UpdatePricingExperimentRequest{
		Status: &[]entities.ExperimentStatus{entities.ExperimentStatusActive}[0],
	})
	return err
}

func (s *PricingServiceImpl) StopPricingExperiment(ctx context.Context, experimentID string) error {
	_, err := s.UpdatePricingExperiment(ctx, experimentID, &UpdatePricingExperimentRequest{
		Status: &[]entities.ExperimentStatus{entities.ExperimentStatusCompleted}[0],
	})
	return err
}

func (s *PricingServiceImpl) AnalyzePricingExperiment(ctx context.Context, experimentID string) (*entities.ExperimentResults, error) {
	report, err := s.experimentService.GenerateExperimentReport(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate experiment report: %w", err)
	}

	// Convert ExperimentReportResult to entities.ExperimentResults
	return &entities.ExperimentResults{
		WinningVariant:          "", // Will be determined from the report
		ConfidenceLevel:         0.0, // Will be calculated from the report
		StatisticalSignificance: false, // Will be determined from the report
		EffectSize:              0.0, // Will be calculated from the report
		Recommendation:          report.Summary,
		MetricImprovements:      make(map[string]float64),
		AnalysisData:            report.Results,
		AnalyzedAt:              report.GeneratedAt,
	}, nil
}

func (s *PricingServiceImpl) GetExperimentVariantForClient(ctx context.Context, clientID string, contentType entities.ContentType) (*entities.PricingVariant, error) {
	experiments, err := s.experimentRepo.GetActiveExperiments(ctx, contentType)
	if err != nil {
		return nil, err
	}

	for _, experiment := range experiments {
		if variant := experiment.GetVariantForClient(clientID); variant != nil {
			return variant, nil
		}
	}

	return nil, nil
}

// Analytics methods
func (s *PricingServiceImpl) GetQuoteAcceptanceRate(ctx context.Context, filter *QuoteAnalyticsFilter) (float64, error) {
	repoFilter := repositories.QuoteAnalyticsFilter{
		ContentType: filter.ContentType,
		ClientTier:  filter.ClientTier,
		TimeRange:   filter.TimeRange,
	}
	return s.priceQuoteRepo.GetQuoteAcceptanceRate(ctx, repoFilter)
}

func (s *PricingServiceImpl) GetPriceDistribution(ctx context.Context, filter *PriceDistributionFilter) (map[string]int, error) {
	repoFilter := repositories.PriceDistributionFilter{
		ContentType: filter.ContentType,
		TimeRange:   filter.TimeRange,
		Buckets:     filter.Buckets,
	}
	return s.priceQuoteRepo.GetPriceDistribution(ctx, repoFilter)
}

func (s *PricingServiceImpl) GetPricingTrends(ctx context.Context, filter *PricingTrendsFilter) (*PricingTrendsResponse, error) {
	// Implementation would aggregate pricing data over time
	// This is a simplified implementation
	return &PricingTrendsResponse{
		Trends:    []PricingTrendPoint{},
		Summary:   TrendSummary{},
		Forecasts: []PricingForecast{},
	}, nil
}

// Helper function to generate IDs
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Type aliases for missing request types
type ContentCreationCostRequest = CalculateResourceCostRequest
type ContentCreationCostResponse = CalculateResourceCostResponse

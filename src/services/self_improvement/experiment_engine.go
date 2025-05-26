package self_improvement

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// experimentEngine implements ExperimentEngine interface
type experimentEngine struct {
	experimentRepo repositories.ExperimentRepository
	rand           *rand.Rand
}

// NewExperimentEngine creates a new experiment engine
func NewExperimentEngine(experimentRepo repositories.ExperimentRepository) ExperimentEngine {
	return &experimentEngine{
		experimentRepo: experimentRepo,
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DesignExperiment creates a new experiment design
func (ee *experimentEngine) DesignExperiment(ctx context.Context, hypothesis string, metrics []string) (*entities.Experiment, error) {
	// Create experiment with default configuration
	experiment := &entities.Experiment{
		ID:          fmt.Sprintf("exp_%d", time.Now().Unix()),
		Name:        ee.generateExperimentName(hypothesis),
		Description: fmt.Sprintf("Testing hypothesis: %s", hypothesis),
		Hypothesis:  hypothesis,
		Type:        entities.ExperimentTypeAB,
		Status:      entities.ExperimentStatusDraft,
		MetricsTracked: metrics,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config:      make(map[string]interface{}),
	}
	
	// Design variants based on hypothesis
	variants := ee.designVariants(hypothesis)
	experiment.Variants = variants
	
	// Set success criteria
	experiment.SuccessCriteria = entities.SuccessCriteria{
		PrimaryMetric:     metrics[0], // First metric is primary
		MinimumEffect:     0.05,       // 5% minimum effect size
		ConfidenceLevel:   0.95,       // 95% confidence
		MinimumSampleSize: ee.calculateSampleSize(0.05, 0.95, 0.8),
	}
	
	// Set sample size
	experiment.SampleSize = experiment.SuccessCriteria.MinimumSampleSize
	
	return experiment, nil
}

// ValidateExperimentDesign validates an experiment design
func (ee *experimentEngine) ValidateExperimentDesign(ctx context.Context, experiment *entities.Experiment) error {
	// Check variants
	if len(experiment.Variants) < 2 {
		return fmt.Errorf("experiment must have at least 2 variants")
	}
	
	// Check for control variant
	hasControl := false
	totalWeight := 0.0
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			hasControl = true
		}
		totalWeight += variant.Weight
	}
	
	if !hasControl {
		return fmt.Errorf("experiment must have a control variant")
	}
	
	// Check weights sum to 1
	if math.Abs(totalWeight-1.0) > 0.001 {
		return fmt.Errorf("variant weights must sum to 1.0, got %.3f", totalWeight)
	}
	
	// Check metrics
	if len(experiment.MetricsTracked) == 0 {
		return fmt.Errorf("experiment must track at least one metric")
	}
	
	// Check success criteria
	if experiment.SuccessCriteria.MinimumSampleSize < 10 {
		return fmt.Errorf("minimum sample size too small: %d", experiment.SuccessCriteria.MinimumSampleSize)
	}
	
	return nil
}

// StartExperiment starts running an experiment
func (ee *experimentEngine) StartExperiment(ctx context.Context, experimentID string) error {
	experiment, err := ee.experimentRepo.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("getting experiment: %w", err)
	}
	
	// Validate before starting
	if err := ee.ValidateExperimentDesign(ctx, experiment); err != nil {
		return fmt.Errorf("invalid experiment design: %w", err)
	}
	
	// Update status
	experiment.Status = entities.ExperimentStatusActive
	experiment.StartDate = time.Now()
	experiment.UpdatedAt = time.Now()
	
	return ee.experimentRepo.UpdateExperiment(ctx, experiment)
}

// AssignVariant assigns an entity to a variant
func (ee *experimentEngine) AssignVariant(ctx context.Context, experimentID, entityID string) (string, error) {
	// Check if already assigned
	existingVariant, err := ee.experimentRepo.GetEntityVariant(ctx, experimentID, entityID)
	if err == nil && existingVariant != "" {
		return existingVariant, nil
	}
	
	// Get experiment
	experiment, err := ee.experimentRepo.GetExperiment(ctx, experimentID)
	if err != nil {
		return "", fmt.Errorf("getting experiment: %w", err)
	}
	
	if experiment.Status != entities.ExperimentStatusActive {
		return "", fmt.Errorf("experiment is not running")
	}
	
	// Assign based on weights
	variantID := ee.selectVariantByWeight(experiment.Variants)
	
	// Store assignment
	if err := ee.experimentRepo.AssignToVariant(ctx, experimentID, entityID, variantID); err != nil {
		return "", fmt.Errorf("assigning variant: %w", err)
	}
	
	return variantID, nil
}

// TrackMetric records a metric value for a variant
func (ee *experimentEngine) TrackMetric(ctx context.Context, experimentID, variantID string, metric string, value float64) error {
	// Record conversion with metric data
	metrics := map[string]float64{metric: value}
	
	// Use a generated entity ID for the metric tracking
	entityID := fmt.Sprintf("metric_%s_%d", metric, time.Now().UnixNano())
	
	return ee.experimentRepo.RecordConversion(ctx, experimentID, variantID, entityID, metrics)
}

// CalculateSignificance performs statistical analysis
func (ee *experimentEngine) CalculateSignificance(ctx context.Context, experimentID string) (*StatisticalAnalysis, error) {
	// Get experiment
	experiment, err := ee.experimentRepo.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("getting experiment: %w", err)
	}
	
	// Get results
	_, err = ee.experimentRepo.GetExperimentResults(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("getting results: %w", err)
	}
	
	// Perform statistical analysis
	analysis := &StatisticalAnalysis{
		SampleSize:      experiment.CurrentSample,
		ConfidenceLevel: experiment.SuccessCriteria.ConfidenceLevel,
		VariantStats:    make(map[string]VariantStats),
	}
	
	// Get control variant
	var controlID string
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			controlID = variant.ID
			break
		}
	}
	
	if controlID == "" {
		return nil, fmt.Errorf("no control variant found")
	}
	
	// Calculate stats for each variant
	// Note: This is simplified since VariantResults is not available in the current struct
	// This would need to be implemented based on the actual experiment data storage
	for _, variant := range experiment.Variants {
		stats := VariantStats{
			SampleSize:     100, // Placeholder
			Conversions:    50,  // Placeholder
			ConversionRate: 0.5, // Placeholder
			Mean:           0.5,
			StdDev:         0.1,
		}
		
		analysis.VariantStats[variant.ID] = stats
	}
	
	// Calculate p-value and effect size
	if len(analysis.VariantStats) >= 2 {
		controlStats := analysis.VariantStats[controlID]
		
		for variantID, variantStats := range analysis.VariantStats {
			if variantID != controlID {
				// Calculate z-score
				pooledP := (float64(controlStats.Conversions+variantStats.Conversions)) / 
				          float64(controlStats.SampleSize+variantStats.SampleSize)
				pooledSE := math.Sqrt(pooledP * (1 - pooledP) * 
				           (1.0/float64(controlStats.SampleSize) + 1.0/float64(variantStats.SampleSize)))
				
				if pooledSE > 0 {
					zScore := (variantStats.ConversionRate - controlStats.ConversionRate) / pooledSE
					pValue := 2 * (1 - ee.normalCDF(math.Abs(zScore)))
					
					analysis.PValue = pValue
					analysis.EffectSize = (variantStats.ConversionRate - controlStats.ConversionRate) / controlStats.ConversionRate
					analysis.SignificantDiff = pValue < (1 - experiment.SuccessCriteria.ConfidenceLevel)
				}
			}
		}
	}
	
	// Calculate statistical power
	analysis.PowerAnalysis = ee.calculatePower(
		analysis.EffectSize,
		experiment.SuccessCriteria.ConfidenceLevel,
		experiment.CurrentSample,
	)
	
	return analysis, nil
}

// DetermineWinner determines the winning variant
func (ee *experimentEngine) DetermineWinner(ctx context.Context, experimentID string) (*entities.ExperimentResults, error) {
	// Get experiment
	experiment, err := ee.experimentRepo.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("getting experiment: %w", err)
	}
	
	// Calculate significance
	analysis, err := ee.CalculateSignificance(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("calculating significance: %w", err)
	}
	
	// Get current results
	results, err := ee.experimentRepo.GetExperimentResults(ctx, experimentID)
	if err != nil {
		// Initialize results if not exists
		results = &entities.ExperimentResults{
			MetricImprovements: make(map[string]float64),
			AnalysisData:       make(map[string]interface{}),
		}
	}
	
	// Update with statistical analysis
	results.ConfidenceLevel = analysis.ConfidenceLevel
	results.EffectSize = analysis.EffectSize
	results.StatisticalSignificance = analysis.SignificantDiff
	
	// Determine winner
	if analysis.SignificantDiff && experiment.CurrentSample >= experiment.SuccessCriteria.MinimumSampleSize {
		// Find variant with best performance
		var bestVariant string
		bestPerformance := -1.0
		
		for variantID, stats := range analysis.VariantStats {
			if stats.ConversionRate > bestPerformance {
				bestPerformance = stats.ConversionRate
				bestVariant = variantID
			}
		}
		
		results.WinningVariant = bestVariant
		
		// Update experiment status
		experiment.Status = entities.ExperimentStatusCompleted
		experiment.EndDate = &[]time.Time{time.Now()}[0]
		experiment.Results = results
		
		if err := ee.experimentRepo.UpdateExperiment(ctx, experiment); err != nil {
			return nil, fmt.Errorf("updating experiment: %w", err)
		}
	}
	
	// Store analysis in results
	results.AnalysisData["statistical_analysis"] = analysis
	
	return results, nil
}

// GenerateRecommendations generates recommendations based on results
func (ee *experimentEngine) GenerateRecommendations(ctx context.Context, results *entities.ExperimentResults) ([]string, error) {
	var recommendations []string
	
	// Check if there's a clear winner
	if results.WinningVariant != "" {
		recommendations = append(recommendations, 
			fmt.Sprintf("Implement variant %s as it showed %.1f%% improvement", 
				results.WinningVariant, results.EffectSize*100))
		
		// Check confidence level
		if results.ConfidenceLevel >= 0.95 {
			recommendations = append(recommendations, 
				"High confidence in results - safe to implement immediately")
		} else if results.ConfidenceLevel >= 0.90 {
			recommendations = append(recommendations,
				"Good confidence in results - consider implementing with monitoring")
		}
	} else {
		// No clear winner
		if !results.StatisticalSignificance {
			recommendations = append(recommendations,
				"No significant difference detected - consider testing more radical changes")
		}
	}
	
	// Analyze metric-specific results
	for metric, improvement := range results.MetricImprovements {
		if improvement > 0.1 {
			recommendations = append(recommendations,
				fmt.Sprintf("%s showed %.1f%% improvement - strong positive signal", 
					metric, improvement*100))
		} else if improvement < -0.1 {
			recommendations = append(recommendations,
				fmt.Sprintf("WARNING: %s showed %.1f%% decrease - investigate impact", 
					metric, improvement*100))
		}
	}
	
	// General recommendations
	// Note: Can't check VariantResults as it's not in the struct
	recommendations = append(recommendations,
		"Consider running follow-up A/B test with top 2 variants for clearer results")
	
	return recommendations, nil
}

// Helper methods

func (ee *experimentEngine) generateExperimentName(hypothesis string) string {
	// Extract key words from hypothesis
	words := []string{}
	for _, word := range []string{"improve", "increase", "decrease", "optimize", "test"} {
		if strings.Contains(strings.ToLower(hypothesis), word) {
			words = append(words, word)
			break
		}
	}
	
	// Add timestamp
	return fmt.Sprintf("Experiment_%s_%s", strings.Join(words, "_"), time.Now().Format("20060102"))
}

func (ee *experimentEngine) designVariants(hypothesis string) []entities.ExperimentVariant {
	// Create control variant
	control := entities.ExperimentVariant{
		ID:          "control",
		Name:        "Control",
		Description: "Current implementation",
		Weight:      0.5,
		IsControl:   true,
		Config:      make(map[string]interface{}),
	}
	
	// Create treatment variant based on hypothesis
	treatment := entities.ExperimentVariant{
		ID:          "treatment",
		Name:        "Treatment",
		Description: "Modified implementation testing: " + hypothesis,
		Weight:      0.5,
		IsControl:   false,
		Config:      ee.generateVariantConfig(hypothesis),
	}
	
	return []entities.ExperimentVariant{control, treatment}
}

func (ee *experimentEngine) generateVariantConfig(hypothesis string) map[string]interface{} {
	config := make(map[string]interface{})
	
	// Parse hypothesis for configuration hints
	hypothesis = strings.ToLower(hypothesis)
	
	if strings.Contains(hypothesis, "prompt") {
		config["prompt_variation"] = true
	}
	if strings.Contains(hypothesis, "price") || strings.Contains(hypothesis, "pricing") {
		config["pricing_variation"] = true
	}
	if strings.Contains(hypothesis, "quality") {
		config["quality_threshold_variation"] = true
	}
	if strings.Contains(hypothesis, "speed") || strings.Contains(hypothesis, "fast") {
		config["performance_variation"] = true
	}
	
	return config
}

func (ee *experimentEngine) calculateSampleSize(minEffect, confidence, power float64) int {
	// Simplified sample size calculation for proportions
	// n = (Z_alpha + Z_beta)^2 * 2 * p * (1-p) / delta^2
	
	// Z-scores for common values
	zAlpha := 1.96 // 95% confidence
	zBeta := 0.84  // 80% power
	
	if confidence >= 0.99 {
		zAlpha = 2.58
	} else if confidence >= 0.95 {
		zAlpha = 1.96
	} else if confidence >= 0.90 {
		zAlpha = 1.645
	}
	
	if power >= 0.90 {
		zBeta = 1.28
	} else if power >= 0.80 {
		zBeta = 0.84
	}
	
	// Assume baseline conversion rate of 0.1 (10%)
	p := 0.1
	
	// Calculate sample size per variant
	n := math.Pow(zAlpha+zBeta, 2) * 2 * p * (1 - p) / math.Pow(minEffect, 2)
	
	// Total sample size (for 2 variants)
	return int(math.Ceil(n * 2))
}

func (ee *experimentEngine) selectVariantByWeight(variants []entities.ExperimentVariant) string {
	// Create cumulative weights
	cumulative := make([]float64, len(variants))
	total := 0.0
	
	for i, variant := range variants {
		total += variant.Weight
		cumulative[i] = total
	}
	
	// Random selection based on weights
	r := ee.rand.Float64() * total
	
	for i, threshold := range cumulative {
		if r <= threshold {
			return variants[i].ID
		}
	}
	
	// Fallback to last variant
	return variants[len(variants)-1].ID
}

func (ee *experimentEngine) normalCDF(x float64) float64 {
	// Approximation of the cumulative distribution function of standard normal
	// Using error function approximation
	return 0.5 * (1 + ee.erf(x/math.Sqrt(2)))
}

func (ee *experimentEngine) erf(x float64) float64 {
	// Approximation of error function
	// Using Abramowitz and Stegun approximation
	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911
	
	sign := 1.0
	if x < 0 {
		sign = -1.0
	}
	x = math.Abs(x)
	
	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)
	
	return sign * y
}

func (ee *experimentEngine) calculatePower(effectSize, confidence float64, sampleSize int) float64 {
	// Simplified power calculation
	// Power = Φ(|δ|√(n/2) - Z_α/2)
	
	zAlpha := 1.96 // 95% confidence
	if confidence >= 0.99 {
		zAlpha = 2.58
	} else if confidence >= 0.90 {
		zAlpha = 1.645
	}
	
	// Calculate power
	if effectSize == 0 || sampleSize == 0 {
		return 0
	}
	
	z := math.Abs(effectSize) * math.Sqrt(float64(sampleSize)/2) - zAlpha
	power := ee.normalCDF(z)
	
	return power
}
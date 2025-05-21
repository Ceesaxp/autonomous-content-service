package content_creation

import (
	"math"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ScoringEngine quantifies content quality using various metrics
type ScoringEngine struct {
	weightingSchemes map[entities.ContentType]ContentTypeWeights
}

// NewScoringEngine creates a new scoring engine
func NewScoringEngine() *ScoringEngine {
	return &ScoringEngine{
		weightingSchemes: initializeWeightingSchemes(),
	}
}

// ContentTypeWeights defines scoring weights for different content types
type ContentTypeWeights struct {
	Readability     float64 `json:"readability"`
	Accuracy        float64 `json:"accuracy"`
	Engagement      float64 `json:"engagement"`
	Clarity         float64 `json:"clarity"`
	Coherence       float64 `json:"coherence"`
	Completeness    float64 `json:"completeness"`
	Relevance       float64 `json:"relevance"`
	Originality     float64 `json:"originality"`
	Tone            float64 `json:"tone"`
	Structure       float64 `json:"structure"`
	Grammar         float64 `json:"grammar"`
	SEO             float64 `json:"seo"`
	CallToAction    float64 `json:"callToAction"`
	EmotionalImpact float64 `json:"emotionalImpact"`
	Credibility     float64 `json:"credibility"`
}

// OverallScoreRequest contains parameters for overall score calculation
type OverallScoreRequest struct {
	CriteriaScores  map[string]float64
	FactCheckScore  float64
	PlagiarismScore float64
	StyleScore      float64
	MultiPassScore  float64
}

// ScoreBreakdown provides detailed score breakdown
type ScoreBreakdown struct {
	OverallScore      float64            `json:"overallScore"`
	WeightedScores    map[string]float64 `json:"weightedScores"`
	CategoryScores    map[string]float64 `json:"categoryScores"`
	QualityGrade      QualityGrade       `json:"qualityGrade"`
	ConfidenceLevel   float64            `json:"confidenceLevel"`
	ScoreDistribution ScoreDistribution  `json:"scoreDistribution"`
}

// QualityGrade represents content quality levels
type QualityGrade string

const (
	GradeExcellent QualityGrade = "excellent"  // 90-100
	GradeGood      QualityGrade = "good"       // 80-89
	GradeSatisfactory QualityGrade = "satisfactory" // 70-79
	GradeNeedsImprovement QualityGrade = "needs_improvement" // 60-69
	GradePoor      QualityGrade = "poor"       // 0-59
)

// ScoreDistribution shows score distribution across categories
type ScoreDistribution struct {
	ContentQuality   float64 `json:"contentQuality"`
	TechnicalQuality float64 `json:"technicalQuality"`
	AudienceAlignment float64 `json:"audienceAlignment"`
	Optimization     float64 `json:"optimization"`
}

// CalculateCriterionScore calculates score for a specific criterion
func (s *ScoringEngine) CalculateCriterionScore(evaluation *DetailedEvaluation, criterion EvaluationCriterion) float64 {
	if result, exists := evaluation.CriteriaResults[string(criterion)]; exists {
		return result.Score
	}
	return 0.0
}

// CalculateOverallScore calculates the overall content quality score
func (s *ScoringEngine) CalculateOverallScore(request OverallScoreRequest) float64 {
	// Base weights for overall calculation
	baseWeights := map[string]float64{
		"criteria":    0.60, // 60% from criteria evaluation
		"factcheck":   0.15, // 15% from fact checking
		"plagiarism":  0.10, // 10% from plagiarism detection
		"style":       0.10, // 10% from style analysis
		"multipass":   0.05, // 5% from multi-pass improvement
	}

	// Calculate weighted criteria score
	criteriaScore := s.calculateWeightedCriteriaScore(request.CriteriaScores)
	
	// Calculate overall score
	overallScore := (criteriaScore * baseWeights["criteria"]) +
		(request.FactCheckScore * baseWeights["factcheck"]) +
		(request.PlagiarismScore * baseWeights["plagiarism"]) +
		(request.StyleScore * baseWeights["style"]) +
		(request.MultiPassScore * baseWeights["multipass"])

	// Ensure score is within bounds
	return math.Max(0, math.Min(100, overallScore))
}

// CalculateDetailedScore provides comprehensive score breakdown
func (s *ScoringEngine) CalculateDetailedScore(request OverallScoreRequest, contentType entities.ContentType) *ScoreBreakdown {
	weights := s.getWeightsForContentType(contentType)
	
	breakdown := &ScoreBreakdown{
		WeightedScores: make(map[string]float64),
		CategoryScores: make(map[string]float64),
	}

	// Calculate weighted scores for each criterion
	totalWeight := 0.0
	weightedSum := 0.0

	for criterion, score := range request.CriteriaScores {
		weight := s.getCriterionWeight(criterion, weights)
		weightedScore := score * weight
		
		breakdown.WeightedScores[criterion] = weightedScore
		weightedSum += weightedScore
		totalWeight += weight
	}

	// Calculate criteria score
	criteriaScore := 0.0
	if totalWeight > 0 {
		criteriaScore = weightedSum / totalWeight
	}

	// Calculate category scores
	breakdown.CategoryScores = s.calculateCategoryScores(request.CriteriaScores, weights)

	// Calculate overall score
	breakdown.OverallScore = s.CalculateOverallScore(request)

	// Determine quality grade
	breakdown.QualityGrade = s.determineQualityGrade(breakdown.OverallScore)

	// Calculate confidence level
	breakdown.ConfidenceLevel = s.calculateConfidenceLevel(request.CriteriaScores)

	// Calculate score distribution
	breakdown.ScoreDistribution = s.calculateScoreDistribution(request.CriteriaScores)

	return breakdown
}

// calculateWeightedCriteriaScore calculates weighted average of criteria scores
func (s *ScoringEngine) calculateWeightedCriteriaScore(criteriaScores map[string]float64) float64 {
	if len(criteriaScores) == 0 {
		return 0.0
	}

	totalWeight := 0.0
	weightedSum := 0.0

	for criterion, score := range criteriaScores {
		weight := s.getBaseCriterionWeight(criterion)
		weightedSum += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return weightedSum / totalWeight
}

// calculateCategoryScores groups criteria into categories and calculates scores
func (s *ScoringEngine) calculateCategoryScores(criteriaScores map[string]float64, weights ContentTypeWeights) map[string]float64 {
	categories := map[string][]string{
		"contentQuality": {
			string(CriterionAccuracy),
			string(CriterionCompleteness),
			string(CriterionOriginality),
			string(CriterionCredibility),
		},
		"technicalQuality": {
			string(CriterionGrammar),
			string(CriterionStructure),
			string(CriterionClarity),
			string(CriterionCoherence),
		},
		"audienceAlignment": {
			string(CriterionTone),
			string(CriterionRelevance),
			string(CriterionEmotionalImpact),
		},
		"optimization": {
			string(CriterionSEO),
			string(CriterionEngagement),
			string(CriterionCallToAction),
			string(CriterionReadability),
		},
	}

	categoryScores := make(map[string]float64)

	for category, criteria := range categories {
		totalScore := 0.0
		totalWeight := 0.0
		count := 0

		for _, criterion := range criteria {
			if score, exists := criteriaScores[criterion]; exists {
				weight := s.getCriterionWeight(criterion, weights)
				totalScore += score * weight
				totalWeight += weight
				count++
			}
		}

		if count > 0 && totalWeight > 0 {
			categoryScores[category] = totalScore / totalWeight
		} else {
			categoryScores[category] = 0.0
		}
	}

	return categoryScores
}

// calculateScoreDistribution calculates score distribution
func (s *ScoringEngine) calculateScoreDistribution(criteriaScores map[string]float64) ScoreDistribution {
	categories := s.calculateCategoryScores(criteriaScores, ContentTypeWeights{
		// Use default weights for distribution calculation
		Readability:     1.0,
		Accuracy:        1.0,
		Engagement:      1.0,
		Clarity:         1.0,
		Coherence:       1.0,
		Completeness:    1.0,
		Relevance:       1.0,
		Originality:     1.0,
		Tone:            1.0,
		Structure:       1.0,
		Grammar:         1.0,
		SEO:             1.0,
		CallToAction:    1.0,
		EmotionalImpact: 1.0,
		Credibility:     1.0,
	})

	return ScoreDistribution{
		ContentQuality:    categories["contentQuality"],
		TechnicalQuality:  categories["technicalQuality"],
		AudienceAlignment: categories["audienceAlignment"],
		Optimization:      categories["optimization"],
	}
}

// calculateConfidenceLevel calculates confidence in the scoring
func (s *ScoringEngine) calculateConfidenceLevel(criteriaScores map[string]float64) float64 {
	if len(criteriaScores) == 0 {
		return 0.0
	}

	// Calculate variance in scores
	mean := 0.0
	for _, score := range criteriaScores {
		mean += score
	}
	mean /= float64(len(criteriaScores))

	variance := 0.0
	for _, score := range criteriaScores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(criteriaScores))

	standardDeviation := math.Sqrt(variance)

	// Lower variance = higher confidence
	// Map standard deviation (0-50) to confidence (1.0-0.5)
	confidence := 1.0 - (standardDeviation / 100.0)
	return math.Max(0.5, math.Min(1.0, confidence))
}

// determineQualityGrade determines quality grade based on score
func (s *ScoringEngine) determineQualityGrade(score float64) QualityGrade {
	switch {
	case score >= 90:
		return GradeExcellent
	case score >= 80:
		return GradeGood
	case score >= 70:
		return GradeSatisfactory
	case score >= 60:
		return GradeNeedsImprovement
	default:
		return GradePoor
	}
}

// getWeightsForContentType returns weights for specific content type
func (s *ScoringEngine) getWeightsForContentType(contentType entities.ContentType) ContentTypeWeights {
	if weights, exists := s.weightingSchemes[contentType]; exists {
		return weights
	}
	return s.weightingSchemes[entities.ContentTypeBlogPost] // Default
}

// getCriterionWeight gets weight for a criterion using content type weights
func (s *ScoringEngine) getCriterionWeight(criterion string, weights ContentTypeWeights) float64 {
	switch EvaluationCriterion(criterion) {
	case CriterionReadability:
		return weights.Readability
	case CriterionAccuracy:
		return weights.Accuracy
	case CriterionEngagement:
		return weights.Engagement
	case CriterionClarity:
		return weights.Clarity
	case CriterionCoherence:
		return weights.Coherence
	case CriterionCompleteness:
		return weights.Completeness
	case CriterionRelevance:
		return weights.Relevance
	case CriterionOriginality:
		return weights.Originality
	case CriterionTone:
		return weights.Tone
	case CriterionStructure:
		return weights.Structure
	case CriterionGrammar:
		return weights.Grammar
	case CriterionSEO:
		return weights.SEO
	case CriterionCallToAction:
		return weights.CallToAction
	case CriterionEmotionalImpact:
		return weights.EmotionalImpact
	case CriterionCredibility:
		return weights.Credibility
	default:
		return 1.0
	}
}

// getBaseCriterionWeight gets base weight for a criterion
func (s *ScoringEngine) getBaseCriterionWeight(criterion string) float64 {
	baseWeights := map[string]float64{
		string(CriterionReadability):    1.2,
		string(CriterionAccuracy):       1.5,
		string(CriterionEngagement):     1.3,
		string(CriterionClarity):        1.2,
		string(CriterionCoherence):      1.1,
		string(CriterionCompleteness):   1.0,
		string(CriterionRelevance):      1.2,
		string(CriterionOriginality):    1.0,
		string(CriterionTone):           1.1,
		string(CriterionStructure):      1.1,
		string(CriterionGrammar):        1.3,
		string(CriterionSEO):            1.0,
		string(CriterionCallToAction):   1.0,
		string(CriterionEmotionalImpact): 1.0,
		string(CriterionCredibility):    1.2,
	}

	if weight, exists := baseWeights[criterion]; exists {
		return weight
	}
	return 1.0
}

// initializeWeightingSchemes sets up content-type-specific weights
func initializeWeightingSchemes() map[entities.ContentType]ContentTypeWeights {
	schemes := make(map[entities.ContentType]ContentTypeWeights)

	// Blog Post weights
	schemes[entities.ContentTypeBlogPost] = ContentTypeWeights{
		Readability:     1.2,
		Accuracy:        1.3,
		Engagement:      1.4,
		Clarity:         1.2,
		Coherence:       1.2,
		Completeness:    1.1,
		Relevance:       1.3,
		Originality:     1.2,
		Tone:            1.1,
		Structure:       1.2,
		Grammar:         1.3,
		SEO:             1.4,
		CallToAction:    1.0,
		EmotionalImpact: 1.1,
		Credibility:     1.2,
	}

	// Social Post weights
	schemes[entities.ContentTypeSocialPost] = ContentTypeWeights{
		Readability:     1.1,
		Accuracy:        1.0,
		Engagement:      1.5,
		Clarity:         1.3,
		Coherence:       1.0,
		Completeness:    0.8,
		Relevance:       1.4,
		Originality:     1.2,
		Tone:            1.4,
		Structure:       0.9,
		Grammar:         1.2,
		SEO:             0.7,
		CallToAction:    1.5,
		EmotionalImpact: 1.5,
		Credibility:     1.0,
	}

	// Technical Article weights
	schemes[entities.ContentTypeTechnicalArticle] = ContentTypeWeights{
		Readability:     1.1,
		Accuracy:        1.8,
		Engagement:      1.0,
		Clarity:         1.5,
		Coherence:       1.4,
		Completeness:    1.6,
		Relevance:       1.3,
		Originality:     1.3,
		Tone:            1.0,
		Structure:       1.4,
		Grammar:         1.4,
		SEO:             0.8,
		CallToAction:    0.7,
		EmotionalImpact: 0.7,
		Credibility:     1.8,
	}

	// Email Newsletter weights
	schemes[entities.ContentTypeEmailNewsletter] = ContentTypeWeights{
		Readability:     1.3,
		Accuracy:        1.2,
		Engagement:      1.4,
		Clarity:         1.3,
		Coherence:       1.2,
		Completeness:    1.0,
		Relevance:       1.4,
		Originality:     1.0,
		Tone:            1.3,
		Structure:       1.2,
		Grammar:         1.3,
		SEO:             0.8,
		CallToAction:    1.5,
		EmotionalImpact: 1.2,
		Credibility:     1.1,
	}

	// Product Description weights
	schemes[entities.ContentTypeProductDescription] = ContentTypeWeights{
		Readability:     1.3,
		Accuracy:        1.4,
		Engagement:      1.3,
		Clarity:         1.4,
		Coherence:       1.1,
		Completeness:    1.3,
		Relevance:       1.4,
		Originality:     1.0,
		Tone:            1.2,
		Structure:       1.1,
		Grammar:         1.3,
		SEO:             1.3,
		CallToAction:    1.5,
		EmotionalImpact: 1.1,
		Credibility:     1.4,
	}

	// Press Release weights
	schemes[entities.ContentTypePressRelease] = ContentTypeWeights{
		Readability:     1.2,
		Accuracy:        1.6,
		Engagement:      1.1,
		Clarity:         1.4,
		Coherence:       1.3,
		Completeness:    1.4,
		Relevance:       1.3,
		Originality:     1.1,
		Tone:            1.3,
		Structure:       1.4,
		Grammar:         1.4,
		SEO:             1.0,
		CallToAction:    1.0,
		EmotionalImpact: 1.0,
		Credibility:     1.6,
	}

	// Website Copy weights
	schemes[entities.ContentTypeWebsiteCopy] = ContentTypeWeights{
		Readability:     1.3,
		Accuracy:        1.2,
		Engagement:      1.3,
		Clarity:         1.4,
		Coherence:       1.2,
		Completeness:    1.1,
		Relevance:       1.3,
		Originality:     1.1,
		Tone:            1.3,
		Structure:       1.2,
		Grammar:         1.3,
		SEO:             1.4,
		CallToAction:    1.4,
		EmotionalImpact: 1.2,
		Credibility:     1.3,
	}

	return schemes
}
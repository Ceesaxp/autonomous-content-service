package content_creation

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// QualityAssuranceSystem manages the self-review quality assurance process
type QualityAssuranceSystem struct {
	evaluationEngine   *EvaluationEngine
	multiPassReviewer  *MultiPassReviewer
	scoringEngine      *ScoringEngine
	factChecker        *FactChecker
	plagiarismDetector *PlagiarismDetector
	styleChecker       *StyleChecker
	improvementEngine  *ImprovementEngine
	revisionTracker    *RevisionTracker
	benchmarkEngine    *BenchmarkEngine
}

// NewQualityAssuranceSystem creates a new quality assurance system
func NewQualityAssuranceSystem(
	llmClient LLMClient,
	searchService SearchService,
	plagiarismAPI PlagiarismAPI,
) *QualityAssuranceSystem {
	evaluationEngine := NewEvaluationEngine(llmClient)
	multiPassReviewer := NewMultiPassReviewer(llmClient, evaluationEngine)
	scoringEngine := NewScoringEngine()
	factChecker := NewFactChecker(llmClient, searchService)
	plagiarismDetector := NewPlagiarismDetector(plagiarismAPI, llmClient)
	styleChecker := NewStyleChecker(llmClient)
	improvementEngine := NewImprovementEngine(llmClient, evaluationEngine)
	revisionTracker := NewRevisionTracker()
	benchmarkEngine := NewBenchmarkEngine()

	return &QualityAssuranceSystem{
		evaluationEngine:   evaluationEngine,
		multiPassReviewer:  multiPassReviewer,
		scoringEngine:      scoringEngine,
		factChecker:        factChecker,
		plagiarismDetector: plagiarismDetector,
		styleChecker:       styleChecker,
		improvementEngine:  improvementEngine,
		revisionTracker:    revisionTracker,
		benchmarkEngine:    benchmarkEngine,
	}
}

// QualityAssessmentRequest contains the parameters for quality assessment
type QualityAssessmentRequest struct {
	Content         *entities.Content
	ContentText     string
	EvaluationCriteria []EvaluationCriterion
	RequiredThreshold float64
	MaxRevisions     int
	TargetAudience   string
	IndustryBenchmark string
}

// QualityAssessmentResult contains the complete quality assessment results
type QualityAssessmentResult struct {
	OverallScore        float64                 `json:"overallScore"`
	PassedThreshold     bool                    `json:"passedThreshold"`
	CriteriaScores      map[string]float64      `json:"criteriaScores"`
	DetailedEvaluation  DetailedEvaluation      `json:"detailedEvaluation"`
	MultiPassResults    []PassResult            `json:"multiPassResults"`
	FactCheckResults    FactCheckResult         `json:"factCheckResults"`
	PlagiarismResults   PlagiarismResult        `json:"plagiarismResults"`
	StyleAnalysis       StyleAnalysisResult     `json:"styleAnalysis"`
	ImprovementSuggestions []ImprovementSuggestion `json:"improvementSuggestions"`
	RevisionHistory     []RevisionRecord        `json:"revisionHistory"`
	BenchmarkComparison BenchmarkComparison     `json:"benchmarkComparison"`
	RecommendedActions  []string                `json:"recommendedActions"`
}

// PerformAssessment conducts a comprehensive quality assessment
func (qa *QualityAssuranceSystem) PerformAssessment(ctx context.Context, request QualityAssessmentRequest) (*QualityAssessmentResult, error) {
	result := &QualityAssessmentResult{
		CriteriaScores:      make(map[string]float64),
		MultiPassResults:    []PassResult{},
		ImprovementSuggestions: []ImprovementSuggestion{},
		RevisionHistory:     []RevisionRecord{},
		RecommendedActions:  []string{},
	}

	// Track this assessment
	revisionID := qa.revisionTracker.StartRevision(request.Content.ContentID, request.ContentText)

	// 1. Perform multi-pass review
	multiPassResults, err := qa.multiPassReviewer.PerformMultiPassReview(ctx, MultiPassRequest{
		Content:       request.ContentText,
		ContentType:   request.Content.Type,
		TargetAudience: request.TargetAudience,
		Criteria:      request.EvaluationCriteria,
	})
	if err != nil {
		return nil, fmt.Errorf("multi-pass review failed: %w", err)
	}
	result.MultiPassResults = multiPassResults.Passes

	// 2. Evaluate content against criteria
	evaluation, err := qa.evaluationEngine.EvaluateContent(ctx, EvaluationRequest{
		Content:    request.ContentText,
		ContentType: request.Content.Type,
		Criteria:   request.EvaluationCriteria,
		TargetAudience: request.TargetAudience,
	})
	if err != nil {
		return nil, fmt.Errorf("content evaluation failed: %w", err)
	}
	result.DetailedEvaluation = *evaluation

	// 3. Calculate scores for each criterion
	for _, criterion := range request.EvaluationCriteria {
		score := qa.scoringEngine.CalculateCriterionScore(evaluation, criterion)
		result.CriteriaScores[string(criterion)] = score
	}

	// 4. Fact-check content
	factCheckResult, err := qa.factChecker.CheckFacts(ctx, FactCheckRequest{
		Content:     request.ContentText,
		ContentType: request.Content.Type,
		Sources:     []string{}, // Will be automatically determined
	})
	if err != nil {
		return nil, fmt.Errorf("fact checking failed: %w", err)
	}
	result.FactCheckResults = *factCheckResult

	// 5. Check for plagiarism
	plagiarismResult, err := qa.plagiarismDetector.CheckPlagiarism(ctx, PlagiarismCheckRequest{
		Content:     request.ContentText,
		ContentType: request.Content.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("plagiarism detection failed: %w", err)
	}
	result.PlagiarismResults = *plagiarismResult

	// 6. Analyze style consistency
	styleResult, err := qa.styleChecker.AnalyzeStyle(ctx, StyleCheckRequest{
		Content:        request.ContentText,
		ContentType:    request.Content.Type,
		TargetAudience: request.TargetAudience,
	})
	if err != nil {
		return nil, fmt.Errorf("style analysis failed: %w", err)
	}
	result.StyleAnalysis = *styleResult

	// 7. Calculate overall score
	result.OverallScore = qa.scoringEngine.CalculateOverallScore(OverallScoreRequest{
		CriteriaScores:    result.CriteriaScores,
		FactCheckScore:    factCheckResult.OverallScore,
		PlagiarismScore:   plagiarismResult.OriginalityScore,
		StyleScore:        styleResult.OverallScore,
		MultiPassScore:    multiPassResults.OverallScore,
	})

	// 8. Check if threshold is met
	result.PassedThreshold = result.OverallScore >= request.RequiredThreshold

	// 9. Generate improvement suggestions if needed
	if !result.PassedThreshold || result.OverallScore < 85.0 {
		improvements, err := qa.improvementEngine.GenerateImprovements(ctx, ImprovementRequest{
			Content:         request.ContentText,
			ContentType:     request.Content.Type,
			EvaluationResults: evaluation,
			FactCheckResults:  factCheckResult,
			StyleResults:     styleResult,
			TargetScore:      request.RequiredThreshold,
			CurrentScore:     result.OverallScore,
		})
		if err != nil {
			return nil, fmt.Errorf("improvement generation failed: %w", err)
		}
		result.ImprovementSuggestions = improvements.Suggestions
	}

	// 10. Compare against industry benchmarks
	if request.IndustryBenchmark != "" {
		benchmarkResult, err := qa.benchmarkEngine.CompareToBenchmark(ctx, BenchmarkRequest{
			Content:         request.ContentText,
			ContentType:     request.Content.Type,
			IndustryType:    request.IndustryBenchmark,
			QualityMetrics:  result.CriteriaScores,
		})
		if err != nil {
			return nil, fmt.Errorf("benchmark comparison failed: %w", err)
		}
		result.BenchmarkComparison = *benchmarkResult
	}

	// 11. Generate recommended actions
	result.RecommendedActions = qa.generateRecommendedActions(result)

	// 12. Record revision
	qa.revisionTracker.CompleteRevision(revisionID, RevisionResult{
		Score:        result.OverallScore,
		PassedThreshold: result.PassedThreshold,
		Improvements: len(result.ImprovementSuggestions),
	})
	result.RevisionHistory = qa.revisionTracker.GetRevisionHistory(request.Content.ContentID)

	return result, nil
}

// generateRecommendedActions creates actionable recommendations based on the assessment
func (qa *QualityAssuranceSystem) generateRecommendedActions(result *QualityAssessmentResult) []string {
	actions := []string{}

	// Priority-based recommendations
	if !result.PassedThreshold {
		actions = append(actions, "CRITICAL: Content does not meet minimum quality threshold - requires revision")
	}

	// Fact-checking issues
	if result.FactCheckResults.ErrorCount > 0 {
		actions = append(actions, fmt.Sprintf("Address %d factual errors identified", result.FactCheckResults.ErrorCount))
	}

	// Plagiarism issues
	if result.PlagiarismResults.OriginalityScore < 0.8 {
		actions = append(actions, "Rewrite content to improve originality - potential plagiarism detected")
	}

	// Style issues
	if result.StyleAnalysis.ConsistencyScore < 0.7 {
		actions = append(actions, "Improve style consistency throughout the content")
	}

	// Specific criteria improvements
	for criterion, score := range result.CriteriaScores {
		if score < 70.0 {
			actions = append(actions, fmt.Sprintf("Focus on improving %s (current score: %.1f)", criterion, score))
		}
	}

	// Benchmark comparison
	if result.BenchmarkComparison.PerformanceGap > 20.0 {
		actions = append(actions, "Content significantly below industry standards - major improvements needed")
	}

	return actions
}

// EvaluationCriterion defines different content evaluation criteria
type EvaluationCriterion string

const (
	CriterionReadability    EvaluationCriterion = "readability"
	CriterionAccuracy       EvaluationCriterion = "accuracy"
	CriterionEngagement     EvaluationCriterion = "engagement"
	CriterionClarity        EvaluationCriterion = "clarity"
	CriterionCoherence      EvaluationCriterion = "coherence"
	CriterionCompleteness   EvaluationCriterion = "completeness"
	CriterionRelevance      EvaluationCriterion = "relevance"
	CriterionOriginality    EvaluationCriterion = "originality"
	CriterionTone           EvaluationCriterion = "tone"
	CriterionStructure      EvaluationCriterion = "structure"
	CriterionGrammar        EvaluationCriterion = "grammar"
	CriterionSEO            EvaluationCriterion = "seo"
	CriterionCallToAction   EvaluationCriterion = "call_to_action"
	CriterionEmotionalImpact EvaluationCriterion = "emotional_impact"
	CriterionCredibility    EvaluationCriterion = "credibility"
)

// GetCriteriaForContentType returns the evaluation criteria for a specific content type
func GetCriteriaForContentType(contentType entities.ContentType) []EvaluationCriterion {
	baseCriteria := []EvaluationCriterion{
		CriterionReadability,
		CriterionAccuracy,
		CriterionClarity,
		CriterionGrammar,
		CriterionStructure,
	}

	switch contentType {
	case entities.ContentTypeBlogPost:
		return append(baseCriteria, 
			CriterionEngagement, 
			CriterionSEO, 
			CriterionCoherence, 
			CriterionOriginality,
			CriterionRelevance,
		)
	case entities.ContentTypeSocialPost:
		return append(baseCriteria,
			CriterionEngagement,
			CriterionEmotionalImpact,
			CriterionCallToAction,
			CriterionTone,
		)
	case entities.ContentTypeEmailNewsletter:
		return append(baseCriteria,
			CriterionEngagement,
			CriterionCallToAction,
			CriterionTone,
			CriterionRelevance,
		)
	case entities.ContentTypeTechnicalArticle:
		return append(baseCriteria,
			CriterionAccuracy,
			CriterionCompleteness,
			CriterionCredibility,
			CriterionCoherence,
		)
	case entities.ContentTypeProductDescription:
		return append(baseCriteria,
			CriterionEngagement,
			CriterionCallToAction,
			CriterionSEO,
			CriterionCredibility,
		)
	case entities.ContentTypePressRelease:
		return append(baseCriteria,
			CriterionCredibility,
			CriterionCompleteness,
			CriterionTone,
			CriterionStructure,
		)
	default:
		return baseCriteria
	}
}
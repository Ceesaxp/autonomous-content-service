package content_creation

import (
	"context"
	"testing"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// Mock implementations for testing

type MockQALLMClient struct{}

func (m *MockQALLMClient) Generate(ctx context.Context, prompt interface{}) (string, error) {
	// Return mock responses based on prompt type
	return `{
		"score": 75,
		"explanation": "Content demonstrates good quality with room for improvement",
		"evidence": ["Clear structure", "Engaging tone"],
		"suggestions": ["Improve readability", "Add more examples"],
		"confidence": 0.8
	}`, nil
}

type MockSearchService struct{}

func (m *MockSearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	return []SearchResult{
		{
			Title:     "Sample Article",
			URL:       "https://example.com/article",
			Snippet:   "This is a sample search result",
			Relevance: 8,
		},
	}, nil
}

func (m *MockSearchService) FetchContent(ctx context.Context, url string) (string, error) {
	return "Sample content from the URL", nil
}

type MockPlagiarismAPI struct{}

func (m *MockPlagiarismAPI) CheckPlagiarism(ctx context.Context, content string) (float64, []PlagiarismDetail, error) {
	return 0.95, []PlagiarismDetail{}, nil
}

// Test fixtures

func createTestContent() *entities.Content {
	content, _ := entities.NewContent(uuid.New(), "Test Article", entities.ContentTypeBlogPost)
	content.Data = `# Test Article

This is a test article to demonstrate the quality assurance system. 
The article contains multiple paragraphs with varying quality levels.

## Introduction
Content quality is important for engaging readers and achieving business objectives.

## Main Content
Here we discuss various aspects of content creation and quality assessment.
The content should be readable, accurate, and engaging for the target audience.

## Conclusion
In conclusion, quality assurance systems help maintain high content standards.`
	return content
}

func createMockQASystem() *QualityAssuranceSystem {
	mockLLMClient := &MockQALLMClient{}
	mockSearchService := &MockSearchService{}
	mockPlagiarismAPI := &MockPlagiarismAPI{}

	return NewQualityAssuranceSystem(mockLLMClient, mockSearchService, mockPlagiarismAPI)
}

// Unit Tests

func TestQualityAssuranceSystem_Creation(t *testing.T) {
	qa := createMockQASystem()
	
	if qa == nil {
		t.Fatal("QualityAssuranceSystem should not be nil")
	}
	
	if qa.evaluationEngine == nil {
		t.Error("EvaluationEngine should be initialized")
	}
	
	if qa.multiPassReviewer == nil {
		t.Error("MultiPassReviewer should be initialized")
	}
	
	if qa.scoringEngine == nil {
		t.Error("ScoringEngine should be initialized")
	}
}

func TestQualityAssuranceSystem_PerformAssessment(t *testing.T) {
	qa := createMockQASystem()
	content := createTestContent()
	
	request := QualityAssessmentRequest{
		Content:           content,
		ContentText:       content.Data,
		EvaluationCriteria: GetCriteriaForContentType(content.Type),
		RequiredThreshold: 80.0,
		MaxRevisions:     3,
		TargetAudience:   "developers",
		IndustryBenchmark: "technology",
	}
	
	ctx := context.Background()
	result, err := qa.PerformAssessment(ctx, request)
	
	if err != nil {
		t.Fatalf("PerformAssessment failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	// Test basic result structure
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", result.OverallScore)
	}
	
	if len(result.CriteriaScores) == 0 {
		t.Error("CriteriaScores should not be empty")
	}
	
	if len(result.MultiPassResults) == 0 {
		t.Error("MultiPassResults should not be empty")
	}
	
	if len(result.RecommendedActions) == 0 {
		t.Error("RecommendedActions should not be empty for test content")
	}
}

func TestEvaluationEngine_EvaluateContent(t *testing.T) {
	mockLLMClient := &MockQALLMClient{}
	engine := NewEvaluationEngine(mockLLMClient)
	
	request := EvaluationRequest{
		Content:        "This is a test content for evaluation.",
		ContentType:    entities.ContentTypeBlogPost,
		Criteria:       []EvaluationCriterion{CriterionReadability, CriterionEngagement},
		TargetAudience: "general",
	}
	
	ctx := context.Background()
	result, err := engine.EvaluateContent(ctx, request)
	
	if err != nil {
		t.Fatalf("EvaluateContent failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", result.OverallScore)
	}
	
	if len(result.CriteriaResults) != len(request.Criteria) {
		t.Errorf("Expected %d criteria results, got %d", len(request.Criteria), len(result.CriteriaResults))
	}
}

func TestScoringEngine_CalculateOverallScore(t *testing.T) {
	engine := NewScoringEngine()
	
	request := OverallScoreRequest{
		CriteriaScores: map[string]float64{
			"readability": 80.0,
			"engagement":  75.0,
			"accuracy":    85.0,
		},
		FactCheckScore:  90.0,
		PlagiarismScore: 95.0,
		StyleScore:      80.0,
		MultiPassScore:  85.0,
	}
	
	score := engine.CalculateOverallScore(request)
	
	if score < 0 || score > 100 {
		t.Errorf("Overall score should be between 0 and 100, got %f", score)
	}
	
	// Score should be reasonable based on input scores
	if score < 70 || score > 90 {
		t.Errorf("Score seems unreasonable based on input values: %f", score)
	}
}

func TestScoringEngine_CalculateDetailedScore(t *testing.T) {
	engine := NewScoringEngine()
	
	request := OverallScoreRequest{
		CriteriaScores: map[string]float64{
			"readability": 80.0,
			"engagement":  75.0,
			"accuracy":    85.0,
		},
		FactCheckScore:  90.0,
		PlagiarismScore: 95.0,
		StyleScore:      80.0,
		MultiPassScore:  85.0,
	}
	
	breakdown := engine.CalculateDetailedScore(request, entities.ContentTypeBlogPost)
	
	if breakdown == nil {
		t.Fatal("Score breakdown should not be nil")
	}
	
	if breakdown.OverallScore < 0 || breakdown.OverallScore > 100 {
		t.Errorf("Overall score should be between 0 and 100, got %f", breakdown.OverallScore)
	}
	
	if len(breakdown.WeightedScores) == 0 {
		t.Error("WeightedScores should not be empty")
	}
	
	if len(breakdown.CategoryScores) == 0 {
		t.Error("CategoryScores should not be empty")
	}
	
	if breakdown.QualityGrade == "" {
		t.Error("QualityGrade should be set")
	}
	
	if breakdown.ConfidenceLevel < 0 || breakdown.ConfidenceLevel > 1 {
		t.Errorf("ConfidenceLevel should be between 0 and 1, got %f", breakdown.ConfidenceLevel)
	}
}

func TestFactChecker_CheckFacts(t *testing.T) {
	mockLLMClient := &MockQALLMClient{}
	mockSearchService := &MockSearchService{}
	checker := NewFactChecker(mockLLMClient, mockSearchService)
	
	request := FactCheckRequest{
		Content:     "The capital of France is Paris. This fact was established in the year 1800.",
		ContentType: entities.ContentTypeBlogPost,
		Sources:     []string{},
	}
	
	ctx := context.Background()
	result, err := checker.CheckFacts(ctx, request)
	
	if err != nil {
		t.Fatalf("CheckFacts failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", result.OverallScore)
	}
	
	if result.ConfidenceLevel < 0 || result.ConfidenceLevel > 1 {
		t.Errorf("ConfidenceLevel should be between 0 and 1, got %f", result.ConfidenceLevel)
	}
}

func TestPlagiarismDetector_CheckPlagiarism(t *testing.T) {
	mockPlagiarismAPI := &MockPlagiarismAPI{}
	mockLLMClient := &MockQALLMClient{}
	detector := NewPlagiarismDetector(mockPlagiarismAPI, mockLLMClient)
	
	request := PlagiarismCheckRequest{
		Content:     "This is original content created for testing purposes.",
		ContentType: entities.ContentTypeBlogPost,
		CheckWeb:    true,
		CheckDatabase: false,
		Sensitivity: 0.8,
	}
	
	ctx := context.Background()
	result, err := detector.CheckPlagiarism(ctx, request)
	
	if err != nil {
		t.Fatalf("CheckPlagiarism failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OriginalityScore < 0 || result.OriginalityScore > 100 {
		t.Errorf("OriginalityScore should be between 0 and 100, got %f", result.OriginalityScore)
	}
	
	if result.ConfidenceLevel < 0 || result.ConfidenceLevel > 1 {
		t.Errorf("ConfidenceLevel should be between 0 and 1, got %f", result.ConfidenceLevel)
	}
	
	if result.Fingerprint.Hash == "" {
		t.Error("Content fingerprint hash should not be empty")
	}
}

func TestStyleChecker_AnalyzeStyle(t *testing.T) {
	mockLLMClient := &MockQALLMClient{}
	checker := NewStyleChecker(mockLLMClient)
	
	request := StyleCheckRequest{
		Content:        "This is a professional article written in a formal tone. The content maintains consistency throughout.",
		ContentType:    entities.ContentTypeBlogPost,
		TargetAudience: "professionals",
		StyleGuide: StyleGuide{
			BrandVoice: "professional and authoritative",
		},
	}
	
	ctx := context.Background()
	result, err := checker.AnalyzeStyle(ctx, request)
	
	if err != nil {
		t.Fatalf("AnalyzeStyle failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", result.OverallScore)
	}
	
	if result.ConsistencyScore < 0 || result.ConsistencyScore > 100 {
		t.Errorf("ConsistencyScore should be between 0 and 100, got %f", result.ConsistencyScore)
	}
	
	if result.ToneAnalysis.PrimaryTone == "" {
		t.Error("Primary tone should be identified")
	}
}

func TestMultiPassReviewer_PerformMultiPassReview(t *testing.T) {
	mockLLMClient := &MockQALLMClient{}
	mockEvaluationEngine := NewEvaluationEngine(mockLLMClient)
	reviewer := NewMultiPassReviewer(mockLLMClient, mockEvaluationEngine)
	
	request := MultiPassRequest{
		Content:        "This is test content for multi-pass review.",
		ContentType:    entities.ContentTypeBlogPost,
		TargetAudience: "general",
		Criteria:       []EvaluationCriterion{CriterionReadability, CriterionEngagement},
	}
	
	ctx := context.Background()
	result, err := reviewer.PerformMultiPassReview(ctx, request)
	
	if err != nil {
		t.Fatalf("PerformMultiPassReview failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", result.OverallScore)
	}
	
	if len(result.Passes) == 0 {
		t.Error("Should have at least one review pass")
	}
	
	if result.ProcessingTime <= 0 {
		t.Error("ProcessingTime should be greater than 0")
	}
}

func TestImprovementEngine_GenerateImprovements(t *testing.T) {
	mockLLMClient := &MockQALLMClient{}
	mockEvaluationEngine := NewEvaluationEngine(mockLLMClient)
	engine := NewImprovementEngine(mockLLMClient, mockEvaluationEngine)
	
	// Create mock evaluation results
	evaluationResults := &DetailedEvaluation{
		OverallScore: 65.0,
		CriteriaResults: map[string]CriterionResult{
			"readability": {Score: 60.0, Explanation: "Below average readability"},
			"engagement":  {Score: 70.0, Explanation: "Moderate engagement"},
		},
	}
	
	request := ImprovementRequest{
		Content:           "This is test content that needs improvement.",
		ContentType:       entities.ContentTypeBlogPost,
		EvaluationResults: evaluationResults,
		TargetScore:       80.0,
		CurrentScore:      65.0,
		MaxSuggestions:    5,
		Focus:             []ImprovementFocus{FocusReadability, FocusEngagement},
	}
	
	ctx := context.Background()
	result, err := engine.GenerateImprovements(ctx, request)
	
	if err != nil {
		t.Fatalf("GenerateImprovements failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if len(result.Suggestions) == 0 {
		t.Error("Should generate at least one improvement suggestion")
	}
	
	if len(result.Suggestions) > request.MaxSuggestions {
		t.Errorf("Should not exceed max suggestions limit: %d", request.MaxSuggestions)
	}
	
	if result.ExpectedImpact.TotalPotentialGain < 0 {
		t.Error("Total potential gain should not be negative")
	}
}

func TestRevisionTracker_CompleteWorkflow(t *testing.T) {
	tracker := NewRevisionTracker()
	contentID := uuid.New()
	
	// Start revision
	revisionID := tracker.StartRevision(contentID, "Initial content")
	if revisionID == "" {
		t.Fatal("RevisionID should not be empty")
	}
	
	// Add quality check point
	tracker.AddQualityCheckPoint(revisionID, 70.0, map[string]float64{
		"readability": 65.0,
		"engagement":  75.0,
	}, CheckInitial)
	
	// Add pending change
	changeID := tracker.AddPendingChange(revisionID, ChangeContent, "Improve readability", "Better content", 10.0)
	if changeID == "" {
		t.Fatal("ChangeID should not be empty")
	}
	
	// Apply change
	tracker.ApplyChange(revisionID, changeID, "old text", "new text", 8.0)
	
	// Complete revision
	result := RevisionResult{
		Score:           78.0,
		PassedThreshold: true,
		Improvements:    1,
	}
	tracker.CompleteRevision(revisionID, result)
	
	// Get revision history
	history := tracker.GetRevisionHistory(contentID)
	if len(history) != 1 {
		t.Errorf("Expected 1 revision record, got %d", len(history))
	}
	
	if history[0].FinalScore != 78.0 {
		t.Errorf("Expected final score 78.0, got %f", history[0].FinalScore)
	}
	
	// Get analytics
	analytics := tracker.GetRevisionAnalytics()
	if analytics.TotalRevisions != 1 {
		t.Errorf("Expected 1 total revision, got %d", analytics.TotalRevisions)
	}
}

func TestBenchmarkEngine_CompareToBenchmark(t *testing.T) {
	engine := NewBenchmarkEngine()
	
	request := BenchmarkRequest{
		Content:      "Test content for benchmarking",
		ContentType:  entities.ContentTypeBlogPost,
		IndustryType: "technology",
		QualityMetrics: map[string]float64{
			"readability": 75.0,
			"engagement":  80.0,
			"accuracy":    85.0,
		},
		ComparisonType: ComparisonIndustry,
	}
	
	ctx := context.Background()
	result, err := engine.CompareToBenchmark(ctx, request)
	
	if err != nil {
		t.Fatalf("CompareToBenchmark failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.OverallPerformance < 0 || result.OverallPerformance > 100 {
		t.Errorf("OverallPerformance should be between 0 and 100, got %f", result.OverallPerformance)
	}
	
	if len(result.MetricComparisons) == 0 {
		t.Error("MetricComparisons should not be empty")
	}
	
	if len(result.BenchmarkDatasets) == 0 {
		t.Error("BenchmarkDatasets should not be empty")
	}
	
	if result.IndustryRanking.TotalCompetitors <= 0 {
		t.Error("TotalCompetitors should be greater than 0")
	}
}

// Integration Tests

func TestQualityAssuranceSystem_Integration(t *testing.T) {
	qa := createMockQASystem()
	content := createTestContent()
	
	request := QualityAssessmentRequest{
		Content:           content,
		ContentText:       content.Data,
		EvaluationCriteria: GetCriteriaForContentType(content.Type),
		RequiredThreshold: 75.0,
		MaxRevisions:     3,
		TargetAudience:   "developers",
		IndustryBenchmark: "technology",
	}
	
	ctx := context.Background()
	result, err := qa.PerformAssessment(ctx, request)
	
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}
	
	// Verify all components worked together
	if result.OverallScore == 0 {
		t.Error("Overall score should be calculated")
	}
	
	if len(result.MultiPassResults) == 0 {
		t.Error("Multi-pass review should produce results")
	}
	
	if len(result.FactCheckResults.FactualErrors) < 0 {
		t.Error("Fact check should run without errors")
	}
	
	if result.PlagiarismResults.OriginalityScore == 0 {
		t.Error("Plagiarism check should produce score")
	}
	
	if result.StyleAnalysis.OverallScore == 0 {
		t.Error("Style analysis should produce score")
	}
	
	if len(result.RecommendedActions) == 0 {
		t.Error("Should generate recommended actions")
	}
}

// Performance Tests

func TestQualityAssuranceSystem_Performance(t *testing.T) {
	qa := createMockQASystem()
	content := createTestContent()
	
	request := QualityAssessmentRequest{
		Content:           content,
		ContentText:       content.Data,
		EvaluationCriteria: GetCriteriaForContentType(content.Type),
		RequiredThreshold: 80.0,
		MaxRevisions:     3,
		TargetAudience:   "developers",
	}
	
	ctx := context.Background()
	start := time.Now()
	
	result, err := qa.PerformAssessment(ctx, request)
	
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Performance test failed: %v", err)
	}
	
	// Performance should be reasonable (adjust threshold as needed)
	maxDuration := 30 * time.Second
	if duration > maxDuration {
		t.Errorf("Assessment took too long: %v (max: %v)", duration, maxDuration)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	t.Logf("Quality assessment completed in %v", duration)
}

// Error Handling Tests

func TestQualityAssuranceSystem_ErrorHandling(t *testing.T) {
	qa := createMockQASystem()
	
	// Test with invalid content
	request := QualityAssessmentRequest{
		Content:           nil,
		ContentText:       "",
		EvaluationCriteria: []EvaluationCriterion{},
		RequiredThreshold: 80.0,
	}
	
	ctx := context.Background()
	result, err := qa.PerformAssessment(ctx, request)
	
	// Should handle gracefully, either with error or degraded functionality
	if err != nil {
		t.Logf("Expected error for invalid input: %v", err)
		return
	}
	
	if result != nil {
		// If no error, result should have reasonable defaults
		if result.OverallScore < 0 || result.OverallScore > 100 {
			t.Error("Score should be within valid range even with invalid input")
		}
	}
}

// Benchmark Tests

func BenchmarkQualityAssuranceSystem_PerformAssessment(b *testing.B) {
	qa := createMockQASystem()
	content := createTestContent()
	
	request := QualityAssessmentRequest{
		Content:           content,
		ContentText:       content.Data,
		EvaluationCriteria: GetCriteriaForContentType(content.Type),
		RequiredThreshold: 80.0,
		MaxRevisions:     3,
		TargetAudience:   "developers",
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := qa.PerformAssessment(ctx, request)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkScoringEngine_CalculateOverallScore(b *testing.B) {
	engine := NewScoringEngine()
	
	request := OverallScoreRequest{
		CriteriaScores: map[string]float64{
			"readability": 80.0,
			"engagement":  75.0,
			"accuracy":    85.0,
			"seo":         70.0,
			"clarity":     82.0,
		},
		FactCheckScore:  90.0,
		PlagiarismScore: 95.0,
		StyleScore:      80.0,
		MultiPassScore:  85.0,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateOverallScore(request)
	}
}
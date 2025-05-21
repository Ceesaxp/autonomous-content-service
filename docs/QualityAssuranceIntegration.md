# Quality Assurance System Integration Guide

## Overview

The Self-Review Quality Assurance System is a comprehensive content evaluation and improvement framework that provides automated quality assessment, multi-pass review, fact-checking, plagiarism detection, style analysis, and benchmark comparison capabilities.

## System Architecture

### Core Components

1. **Quality Assurance System** (`quality_assurance.go`)
   - Main orchestrator component
   - Coordinates all quality assessment processes
   - Provides unified assessment results

2. **Evaluation Engine** (`evaluation_engine.go`)
   - Evaluates content against multiple criteria
   - Provides detailed scoring and explanations
   - Supports 15 different evaluation criteria

3. **Multi-Pass Reviewer** (`multi_pass_reviewer.go`)
   - Conducts specialized review passes
   - Focuses on different aspects per pass
   - Tracks improvement across iterations

4. **Scoring Engine** (`scoring_engine.go`)
   - Quantifies content quality using weighted metrics
   - Provides detailed score breakdowns
   - Supports content-type-specific weighting

5. **Fact Checker** (`fact_checker.go`)
   - Verifies factual accuracy against reliable sources
   - Detects contradictions and inconsistencies
   - Provides credibility assessment

6. **Plagiarism Detector** (`plagiarism_detector.go`)
   - Detects potential plagiarism using multiple methods
   - Creates content fingerprints
   - Provides originality scoring

7. **Style Checker** (`style_checker.go`)
   - Analyzes style consistency and brand alignment
   - Detects tone and voice inconsistencies
   - Provides formatting recommendations

8. **Improvement Engine** (`improvement_engine.go`)
   - Generates targeted improvement suggestions
   - Prioritizes recommendations by impact and effort
   - Creates implementation plans

9. **Revision Tracker** (`revision_tracker.go`)
   - Tracks revision history and quality improvements
   - Provides analytics on system performance
   - Identifies improvement patterns

10. **Benchmark Engine** (`benchmark_engine.go`)
    - Compares content against industry standards
    - Provides competitive analysis
    - Tracks performance trends

## Integration Steps

### 1. Initialize the Quality Assurance System

```go
package main

import (
    "github.com/Ceesaxp/autonomous-content-service/src/services/content_creation"
)

func main() {
    // Initialize dependencies
    llmClient := content_creation.NewOpenAIClient(apiKey, "gpt-4", 2000, 0.7)
    searchService := content_creation.NewWebSearchService(searchAPIKey, searchURL)
    plagiarismAPI := content_creation.NewSimplePlagiarismAPI()

    // Create quality assurance system
    qaSystem := content_creation.NewQualityAssuranceSystem(
        llmClient,
        searchService,
        plagiarismAPI,
    )
}
```

### 2. Configure Evaluation Criteria

```go
// Get default criteria for content type
criteria := content_creation.GetCriteriaForContentType(entities.ContentTypeBlogPost)

// Or create custom criteria
customCriteria := []content_creation.EvaluationCriterion{
    content_creation.CriterionReadability,
    content_creation.CriterionEngagement,
    content_creation.CriterionAccuracy,
    content_creation.CriterionSEO,
}
```

### 3. Perform Quality Assessment

```go
// Create assessment request
request := content_creation.QualityAssessmentRequest{
    Content:           contentEntity,
    ContentText:       contentText,
    EvaluationCriteria: criteria,
    RequiredThreshold: 80.0,
    MaxRevisions:     3,
    TargetAudience:   "developers",
    IndustryBenchmark: "technology",
}

// Perform assessment
ctx := context.Background()
result, err := qaSystem.PerformAssessment(ctx, request)
if err != nil {
    log.Fatalf("Quality assessment failed: %v", err)
}

// Process results
fmt.Printf("Overall Score: %.1f\n", result.OverallScore)
fmt.Printf("Passed Threshold: %v\n", result.PassedThreshold)
fmt.Printf("Recommendations: %d\n", len(result.ImprovementSuggestions))
```

### 4. Handle Assessment Results

```go
// Check if content meets quality threshold
if !result.PassedThreshold {
    // Apply improvement suggestions
    for _, suggestion := range result.ImprovementSuggestions {
        if suggestion.Priority == content_creation.PriorityQuickWin {
            // Implement quick wins first
            fmt.Printf("Quick Win: %s\n", suggestion.Title)
            fmt.Printf("Action: %s\n", suggestion.Implementation)
        }
    }
}

// Review fact-checking results
if result.FactCheckResults.ErrorCount > 0 {
    for _, error := range result.FactCheckResults.FactualErrors {
        fmt.Printf("Factual Error: %s\n", error.Issue)
        fmt.Printf("Correction: %s\n", error.Correction)
    }
}

// Check plagiarism results
if result.PlagiarismResults.OriginalityScore < 0.8 {
    fmt.Printf("Originality concerns detected\n")
    for _, match := range result.PlagiarismResults.Matches {
        fmt.Printf("Match: %s (%.1f%% similar)\n", 
            match.Source.URL, match.SimilarityScore*100)
    }
}
```

### 5. Integrate with Content Pipeline

```go
// Example integration with content creation pipeline
func (p *ContentPipeline) finalizeContent(ctx context.Context, content *entities.Content) error {
    // Perform quality assessment
    qaRequest := content_creation.QualityAssessmentRequest{
        Content:           content,
        ContentText:       content.Data,
        EvaluationCriteria: content_creation.GetCriteriaForContentType(content.Type),
        RequiredThreshold: 80.0,
        TargetAudience:   p.config.TargetAudience,
    }
    
    qaResult, err := p.qaSystem.PerformAssessment(ctx, qaRequest)
    if err != nil {
        return fmt.Errorf("quality assessment failed: %w", err)
    }
    
    // Update content with quality metrics
    content.UpdateStatistics(entities.ContentStatistics{
        ReadabilityScore: qaResult.CriteriaScores["readability"],
        SEOScore:        qaResult.CriteriaScores["seo"],
        EngagementScore: qaResult.CriteriaScores["engagement"],
        PlagiarismScore: qaResult.PlagiarismResults.OriginalityScore,
    })
    
    // Check if revision is needed
    if !qaResult.PassedThreshold {
        return p.reviseContent(ctx, content, qaResult.ImprovementSuggestions)
    }
    
    return nil
}
```

## Configuration Options

### Quality Thresholds

```go
type QualityThresholds struct {
    OverallMinimum    float64 // Minimum overall score (e.g., 80.0)
    ReadabilityMin    float64 // Minimum readability score (e.g., 70.0)
    AccuracyMin       float64 // Minimum accuracy score (e.g., 90.0)
    OriginalityMin    float64 // Minimum originality score (e.g., 0.8)
    FactErrorMax      int     // Maximum factual errors (e.g., 0)
}
```

### Content Type Weights

```go
// Customize scoring weights for different content types
blogPostWeights := content_creation.ContentTypeWeights{
    Readability:     1.2,
    Accuracy:        1.3,
    Engagement:      1.4,
    SEO:             1.4,
    Clarity:         1.2,
    // ... other criteria
}
```

### Style Guidelines

```go
// Configure brand-specific style guidelines
styleGuide := content_creation.StyleGuide{
    BrandVoice:       "professional and authoritative",
    ToneGuidelines:   map[string]string{
        "blog_post": "conversational yet expert",
        "technical": "formal and precise",
    },
    ForbiddenPhrases: []string{"very", "really", "totally"},
    PreferredPhrases: []string{"significantly", "substantially"},
}
```

## API Integration

### REST API Endpoints

```go
// POST /api/v1/content/assess
type AssessmentRequest struct {
    ContentID    string   `json:"contentId"`
    ContentText  string   `json:"contentText"`
    ContentType  string   `json:"contentType"`
    Criteria     []string `json:"criteria"`
    Threshold    float64  `json:"threshold"`
    Audience     string   `json:"audience"`
}

type AssessmentResponse struct {
    OverallScore    float64            `json:"overallScore"`
    PassedThreshold bool               `json:"passedThreshold"`
    CriteriaScores  map[string]float64 `json:"criteriaScores"`
    Recommendations []Recommendation   `json:"recommendations"`
    ProcessingTime  string             `json:"processingTime"`
}
```

### WebSocket for Real-Time Assessment

```go
// Real-time quality assessment during content editing
type QualityUpdateMessage struct {
    Type         string             `json:"type"`
    ContentID    string             `json:"contentId"`
    OverallScore float64            `json:"overallScore"`
    Issues       []QualityIssue     `json:"issues"`
    Suggestions  []Suggestion       `json:"suggestions"`
}
```

## Monitoring and Analytics

### Quality Metrics Dashboard

```go
// Track quality metrics over time
type QualityMetrics struct {
    AverageScore       float64   `json:"averageScore"`
    PassRate          float64   `json:"passRate"`
    CommonIssues      []string  `json:"commonIssues"`
    ImprovementTrends []Trend   `json:"improvementTrends"`
    ProcessingTime    Duration  `json:"processingTime"`
}
```

### Performance Monitoring

```go
// Monitor system performance
func (qa *QualityAssuranceSystem) GetPerformanceMetrics() PerformanceMetrics {
    analytics := qa.revisionTracker.GetRevisionAnalytics()
    return PerformanceMetrics{
        AverageProcessingTime: analytics.AverageRevisionTime,
        ThroughputPerHour:     analytics.PerformanceMetrics.ThroughputPerHour,
        ErrorRate:             analytics.PerformanceMetrics.ErrorRate,
        SuccessRate:           analytics.SuccessRate,
    }
}
```

## Error Handling

### Graceful Degradation

```go
// Handle partial failures gracefully
func (qa *QualityAssuranceSystem) PerformAssessment(ctx context.Context, request QualityAssessmentRequest) (*QualityAssessmentResult, error) {
    result := &QualityAssessmentResult{}
    
    // Continue with other assessments even if one fails
    if evaluation, err := qa.evaluationEngine.EvaluateContent(ctx, evalRequest); err == nil {
        result.DetailedEvaluation = *evaluation
    } else {
        log.Printf("Evaluation failed: %v", err)
        // Use fallback scoring
    }
    
    // Similar pattern for other components...
    
    return result, nil
}
```

### Timeout Handling

```go
// Use context timeouts for LLM calls
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := llmClient.Generate(ctx, prompt)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout gracefully
        return fallbackResponse, nil
    }
    return nil, err
}
```

## Best Practices

### 1. Chunked Processing

For large content, process in chunks to avoid timeouts:

```go
func (qa *QualityAssuranceSystem) processLargeContent(content string) {
    chunks := qa.chunkContent(content, 1000) // 1000 words per chunk
    
    for i, chunk := range chunks {
        chunkResult := qa.processChunk(chunk)
        // Aggregate results
    }
}
```

### 2. Caching

Implement caching for expensive operations:

```go
type CachedQualityAssurance struct {
    qa    *QualityAssuranceSystem
    cache map[string]*QualityAssessmentResult
}

func (c *CachedQualityAssurance) PerformAssessment(ctx context.Context, request QualityAssessmentRequest) (*QualityAssessmentResult, error) {
    // Check cache first
    key := c.generateCacheKey(request)
    if cached, exists := c.cache[key]; exists {
        return cached, nil
    }
    
    // Perform assessment and cache result
    result, err := c.qa.PerformAssessment(ctx, request)
    if err == nil {
        c.cache[key] = result
    }
    
    return result, err
}
```

### 3. Incremental Assessment

For real-time editing, use incremental assessment:

```go
type IncrementalAssessment struct {
    lastContent string
    lastResult  *QualityAssessmentResult
}

func (ia *IncrementalAssessment) UpdateAssessment(newContent string) *QualityAssessmentResult {
    changes := ia.detectChanges(ia.lastContent, newContent)
    
    if len(changes) < 100 { // Small changes
        return ia.updatePartialAssessment(changes)
    }
    
    // Full reassessment for large changes
    return ia.performFullAssessment(newContent)
}
```

## Testing

### Unit Testing

```go
func TestQualityAssurance(t *testing.T) {
    // Use mock implementations for testing
    mockLLM := &MockLLMClient{}
    mockSearch := &MockSearchService{}
    mockPlagiarism := &MockPlagiarismAPI{}
    
    qa := NewQualityAssuranceSystem(mockLLM, mockSearch, mockPlagiarism)
    
    // Test assessment
    result, err := qa.PerformAssessment(ctx, testRequest)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Testing

```go
func TestQualityAssuranceIntegration(t *testing.T) {
    // Test with real LLM but controlled input
    qa := createTestQualityAssuranceSystem()
    
    testCases := []struct {
        content  string
        expected float64
    }{
        {"High quality content with proper structure", 85.0},
        {"Poor quality content", 45.0},
    }
    
    for _, tc := range testCases {
        result, err := qa.PerformAssessment(ctx, createRequest(tc.content))
        assert.NoError(t, err)
        assert.InDelta(t, tc.expected, result.OverallScore, 10.0)
    }
}
```

### Performance Testing

```go
func BenchmarkQualityAssessment(b *testing.B) {
    qa := createQualityAssuranceSystem()
    request := createBenchmarkRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := qa.PerformAssessment(context.Background(), request)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Deployment Considerations

### 1. Resource Requirements

- **Memory**: 2-4 GB for typical workloads
- **CPU**: Multi-core recommended for parallel processing
- **Network**: Stable connection for LLM API calls
- **Storage**: Minimal for core system, scales with revision history

### 2. Scalability

- Horizontal scaling through multiple instances
- Load balancing for high-throughput scenarios
- Database sharding for revision history
- CDN for benchmark datasets

### 3. Security

- API key management for LLM services
- Content encryption in transit and at rest
- Access control for quality assessment results
- Audit logging for compliance

## Troubleshooting

### Common Issues

1. **High Processing Times**
   - Check LLM API response times
   - Implement request batching
   - Use content chunking for large documents

2. **Inconsistent Scores**
   - Verify LLM model consistency
   - Check evaluation criteria weights
   - Review benchmark datasets

3. **Memory Usage**
   - Implement result caching with TTL
   - Clear revision history periodically
   - Use streaming for large content processing

### Debugging Tools

```go
// Enable detailed logging
qa := NewQualityAssuranceSystem(llmClient, searchService, plagiarismAPI)
qa.SetLogLevel(LogLevelDebug)

// Performance profiling
result, err := qa.PerformAssessmentWithProfiling(ctx, request)
fmt.Printf("Processing breakdown: %+v\n", result.ProcessingBreakdown)
```

This integration guide provides comprehensive instructions for implementing and using the Self-Review Quality Assurance System in your autonomous content creation service.
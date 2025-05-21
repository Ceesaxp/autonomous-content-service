package content_creation

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// MultiPassReviewer conducts specialized multi-pass reviews with different focuses
type MultiPassReviewer struct {
	llmClient        LLMClient
	evaluationEngine *EvaluationEngine
}

// NewMultiPassReviewer creates a new multi-pass reviewer
func NewMultiPassReviewer(llmClient LLMClient, evaluationEngine *EvaluationEngine) *MultiPassReviewer {
	return &MultiPassReviewer{
		llmClient:        llmClient,
		evaluationEngine: evaluationEngine,
	}
}

// ReviewPass represents different types of review passes
type ReviewPass string

const (
	PassContentStructure  ReviewPass = "content_structure"
	PassLanguageQuality   ReviewPass = "language_quality"
	PassFactualAccuracy   ReviewPass = "factual_accuracy"
	PassAudienceAlignment ReviewPass = "audience_alignment"
	PassEngagementOptimization ReviewPass = "engagement_optimization"
	PassFinalPolish       ReviewPass = "final_polish"
)

// MultiPassRequest contains parameters for multi-pass review
type MultiPassRequest struct {
	Content        string
	ContentType    entities.ContentType
	TargetAudience string
	Criteria       []EvaluationCriterion
	CustomPasses   []ReviewPass
}

// MultiPassResult contains results from multi-pass review
type MultiPassResult struct {
	OverallScore    float64      `json:"overallScore"`
	Passes          []PassResult `json:"passes"`
	FinalRecommendations []string `json:"finalRecommendations"`
	QualityImprovement   float64  `json:"qualityImprovement"`
	ProcessingTime       time.Duration `json:"processingTime"`
}

// PassResult contains results from a single review pass
type PassResult struct {
	PassType        ReviewPass     `json:"passType"`
	Score           float64        `json:"score"`
	PreviousScore   float64        `json:"previousScore"`
	Improvement     float64        `json:"improvement"`
	Focus           string         `json:"focus"`
	Findings        []ReviewFinding `json:"findings"`
	Recommendations []string        `json:"recommendations"`
	ProcessingTime  time.Duration   `json:"processingTime"`
	Confidence      float64         `json:"confidence"`
}

// ReviewFinding represents a specific finding from a review pass
type ReviewFinding struct {
	Type        FindingType `json:"type"`
	Severity    Severity    `json:"severity"`
	Description string      `json:"description"`
	Location    string      `json:"location"`
	Suggestion  string      `json:"suggestion"`
	Example     string      `json:"example,omitempty"`
}

// FindingType categorizes different types of review findings
type FindingType string

const (
	FindingStructuralIssue    FindingType = "structural_issue"
	FindingLanguageError      FindingType = "language_error"
	FindingFactualError       FindingType = "factual_error"
	FindingAudienceMismatch   FindingType = "audience_mismatch"
	FindingEngagementGap      FindingType = "engagement_gap"
	FindingStyleInconsistency FindingType = "style_inconsistency"
	FindingOptimizationOpportunity FindingType = "optimization_opportunity"
)

// Severity levels for review findings
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityMajor    Severity = "major"
	SeverityMinor    Severity = "minor"
	SeverityInfo     Severity = "info"
)

// PerformMultiPassReview conducts a comprehensive multi-pass review
func (r *MultiPassReviewer) PerformMultiPassReview(ctx context.Context, request MultiPassRequest) (*MultiPassResult, error) {
	startTime := time.Now()
	
	result := &MultiPassResult{
		Passes:              []PassResult{},
		FinalRecommendations: []string{},
	}

	// Determine review passes
	passes := r.determinePasses(request)
	
	// Conduct each review pass
	currentContent := request.Content
	previousScore := 0.0
	
	for i, pass := range passes {
		passStartTime := time.Now()
		
		passResult, err := r.conductPass(ctx, pass, currentContent, request, previousScore)
		if err != nil {
			return nil, fmt.Errorf("pass %s failed: %w", pass, err)
		}
		
		passResult.ProcessingTime = time.Since(passStartTime)
		result.Passes = append(result.Passes, *passResult)
		
		// Update previous score for next pass
		previousScore = passResult.Score
		
		// Apply improvements if significant
		if i < len(passes)-1 && passResult.Improvement > 5.0 {
			// In a real implementation, this would apply the suggested improvements
			// For now, we'll simulate content improvement
			currentContent = r.simulateContentImprovement(currentContent, passResult)
		}
	}

	// Calculate overall metrics
	result.OverallScore = r.calculateOverallScore(result.Passes)
	result.QualityImprovement = r.calculateQualityImprovement(result.Passes)
	result.FinalRecommendations = r.generateFinalRecommendations(result.Passes)
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// determinePasses selects appropriate review passes based on content type and criteria
func (r *MultiPassReviewer) determinePasses(request MultiPassRequest) []ReviewPass {
	// Use custom passes if provided
	if len(request.CustomPasses) > 0 {
		return request.CustomPasses
	}

	// Default pass sequence based on content type
	basePasses := []ReviewPass{
		PassContentStructure,
		PassLanguageQuality,
		PassFactualAccuracy,
		PassAudienceAlignment,
		PassEngagementOptimization,
		PassFinalPolish,
	}

	// Customize based on content type
	switch request.ContentType {
	case entities.ContentTypeTechnicalArticle:
		return []ReviewPass{
			PassContentStructure,
			PassFactualAccuracy,
			PassLanguageQuality,
			PassAudienceAlignment,
			PassFinalPolish,
		}
	case entities.ContentTypeSocialPost:
		return []ReviewPass{
			PassEngagementOptimization,
			PassAudienceAlignment,
			PassLanguageQuality,
			PassFinalPolish,
		}
	case entities.ContentTypeBlogPost:
		return basePasses
	case entities.ContentTypeProductDescription:
		return []ReviewPass{
			PassAudienceAlignment,
			PassEngagementOptimization,
			PassLanguageQuality,
			PassFinalPolish,
		}
	default:
		return basePasses
	}
}

// conductPass performs a specific review pass
func (r *MultiPassReviewer) conductPass(ctx context.Context, pass ReviewPass, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	switch pass {
	case PassContentStructure:
		return r.reviewContentStructure(ctx, content, request, previousScore)
	case PassLanguageQuality:
		return r.reviewLanguageQuality(ctx, content, request, previousScore)
	case PassFactualAccuracy:
		return r.reviewFactualAccuracy(ctx, content, request, previousScore)
	case PassAudienceAlignment:
		return r.reviewAudienceAlignment(ctx, content, request, previousScore)
	case PassEngagementOptimization:
		return r.reviewEngagementOptimization(ctx, content, request, previousScore)
	case PassFinalPolish:
		return r.reviewFinalPolish(ctx, content, request, previousScore)
	default:
		return r.reviewGeneric(ctx, pass, content, request, previousScore)
	}
}

// reviewContentStructure focuses on content organization and structure
func (r *MultiPassReviewer) reviewContentStructure(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing ONLY on the content structure and organization of this %s for %s audience.

FOCUS AREAS:
1. Logical flow and organization
2. Use of headings and subheadings
3. Paragraph structure and breaks
4. Introduction, body, and conclusion effectiveness
5. Transition quality between sections

Analyze the content and identify specific structural issues, rate the structural quality (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Content Structure and Organization",
  "findings": [
    {
      "type": "structural_issue",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<specific suggestion>",
      "example": "<example if applicable>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, request.TargetAudience, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassContentStructure, response, previousScore)
}

// reviewLanguageQuality focuses on grammar, style, and language usage
func (r *MultiPassReviewer) reviewLanguageQuality(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing ONLY on language quality of this %s for %s audience.

FOCUS AREAS:
1. Grammar and syntax accuracy
2. Spelling and punctuation
3. Sentence variety and flow
4. Word choice and vocabulary
5. Consistency in language style

Analyze the content and identify specific language issues, rate the language quality (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Language Quality and Grammar",
  "findings": [
    {
      "type": "language_error",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<specific suggestion>",
      "example": "<corrected example>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, request.TargetAudience, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassLanguageQuality, response, previousScore)
}

// reviewFactualAccuracy focuses on accuracy and credibility
func (r *MultiPassReviewer) reviewFactualAccuracy(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing ONLY on factual accuracy and credibility of this %s.

FOCUS AREAS:
1. Verifiable claims and statements
2. Data accuracy and currentness
3. Source credibility (where applicable)
4. Consistency of facts throughout
5. Absence of contradictions

Analyze the content and identify potential factual issues, rate the factual accuracy (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Factual Accuracy and Credibility",
  "findings": [
    {
      "type": "factual_error",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<correction or verification needed>",
      "example": "<corrected version if applicable>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassFactualAccuracy, response, previousScore)
}

// reviewAudienceAlignment focuses on target audience appropriateness
func (r *MultiPassReviewer) reviewAudienceAlignment(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing ONLY on audience alignment of this %s for %s audience.

FOCUS AREAS:
1. Appropriate complexity level for audience
2. Relevant topics and examples
3. Tone and formality level
4. Cultural and demographic considerations
5. Audience-specific needs and interests

Analyze the content and identify audience misalignment issues, rate the audience alignment (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Audience Alignment and Relevance",
  "findings": [
    {
      "type": "audience_mismatch",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<how to better align with audience>",
      "example": "<better approach example>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, request.TargetAudience, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassAudienceAlignment, response, previousScore)
}

// reviewEngagementOptimization focuses on engagement and persuasiveness
func (r *MultiPassReviewer) reviewEngagementOptimization(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing ONLY on engagement optimization of this %s for %s audience.

FOCUS AREAS:
1. Hook effectiveness and opening impact
2. Emotional connection and appeal
3. Call-to-action clarity and persuasiveness
4. Interactive elements and questions
5. Overall engagement potential

Analyze the content and identify engagement gaps, rate the engagement potential (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Engagement and Persuasiveness",
  "findings": [
    {
      "type": "engagement_gap",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<how to improve engagement>",
      "example": "<more engaging alternative>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, request.TargetAudience, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassEngagementOptimization, response, previousScore)
}

// reviewFinalPolish focuses on final refinements and optimization
func (r *MultiPassReviewer) reviewFinalPolish(ctx context.Context, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a final polish review of this %s for %s audience, focusing on overall refinement.

FOCUS AREAS:
1. Final readability and flow
2. Professional presentation
3. Consistency throughout
4. Optimization opportunities
5. Publication readiness

Analyze the content for final polish needs, rate the publication readiness (0-100), and provide final recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "Final Polish and Optimization",
  "findings": [
    {
      "type": "optimization_opportunity",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<final improvement>",
      "example": "<polished version>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, request.ContentType, request.TargetAudience, content)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(PassFinalPolish, response, previousScore)
}

// reviewGeneric handles custom review passes
func (r *MultiPassReviewer) reviewGeneric(ctx context.Context, pass ReviewPass, content string, request MultiPassRequest, previousScore float64) (*PassResult, error) {
	prompt := fmt.Sprintf(`Conduct a specialized review focusing on "%s" for this %s content targeting %s audience.

Provide specific analysis for this focus area, identify relevant issues, rate the quality (0-100), and provide targeted recommendations.

Content to review:
%s

Respond in JSON format:
{
  "score": <number 0-100>,
  "focus": "%s",
  "findings": [
    {
      "type": "style_inconsistency",
      "severity": "major|minor|critical|info",
      "description": "<description>",
      "location": "<where in content>",
      "suggestion": "<specific suggestion>",
      "example": "<example if applicable>"
    }
  ],
  "recommendations": ["<recommendation 1>", "<recommendation 2>"],
  "confidence": <0-1>
}`, pass, request.ContentType, request.TargetAudience, content, pass)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return r.parsePassResult(pass, response, previousScore)
}

// parsePassResult parses LLM response into PassResult
func (r *MultiPassReviewer) parsePassResult(passType ReviewPass, response string, previousScore float64) (*PassResult, error) {
	// Try to parse JSON response first
	// This is a simplified implementation - in a real system, you'd use proper JSON parsing
	result := &PassResult{
		PassType:        passType,
		PreviousScore:   previousScore,
		Findings:        []ReviewFinding{},
		Recommendations: []string{},
		Confidence:      0.8,
	}

	// Extract score (simplified regex approach)
	scoreRegex := `"score"\s*:\s*(\d+(?:\.\d+)?)`
	if matches := regexp.MustCompile(scoreRegex).FindStringSubmatch(response); len(matches) > 1 {
		if score, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.Score = score
		}
	}

	// Extract focus
	focusRegex := `"focus"\s*:\s*"([^"]+)"`
	if matches := regexp.MustCompile(focusRegex).FindStringSubmatch(response); len(matches) > 1 {
		result.Focus = matches[1]
	}

	// Calculate improvement
	result.Improvement = result.Score - previousScore

	// In a real implementation, you would properly parse the JSON
	// and extract findings and recommendations

	return result, nil
}

// simulateContentImprovement simulates applying improvements to content
func (r *MultiPassReviewer) simulateContentImprovement(content string, passResult *PassResult) string {
	// In a real implementation, this would apply specific improvements
	// For now, we just return the original content
	return content
}

// calculateOverallScore calculates the overall score from all passes
func (r *MultiPassReviewer) calculateOverallScore(passes []PassResult) float64 {
	if len(passes) == 0 {
		return 0.0
	}

	totalScore := 0.0
	totalWeight := 0.0

	for _, pass := range passes {
		weight := r.getPassWeight(pass.PassType)
		totalScore += pass.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateQualityImprovement calculates the total quality improvement
func (r *MultiPassReviewer) calculateQualityImprovement(passes []PassResult) float64 {
	if len(passes) == 0 {
		return 0.0
	}

	firstScore := passes[0].PreviousScore
	lastScore := passes[len(passes)-1].Score

	return lastScore - firstScore
}

// generateFinalRecommendations generates final consolidated recommendations
func (r *MultiPassReviewer) generateFinalRecommendations(passes []PassResult) []string {
	recommendations := []string{}
	
	// Collect high-priority recommendations
	for _, pass := range passes {
		for _, finding := range pass.Findings {
			if finding.Severity == SeverityCritical || finding.Severity == SeverityMajor {
				recommendations = append(recommendations, 
					fmt.Sprintf("[%s] %s", pass.PassType, finding.Suggestion))
			}
		}
	}

	// Add pass-specific recommendations
	for _, pass := range passes {
		if pass.Score < 70.0 { // Low score threshold
			for _, rec := range pass.Recommendations {
				recommendations = append(recommendations, 
					fmt.Sprintf("[%s] %s", pass.PassType, rec))
			}
		}
	}

	return recommendations
}

// getPassWeight returns the importance weight for different pass types
func (r *MultiPassReviewer) getPassWeight(pass ReviewPass) float64 {
	weights := map[ReviewPass]float64{
		PassContentStructure:        1.2,
		PassLanguageQuality:         1.3,
		PassFactualAccuracy:         1.5,
		PassAudienceAlignment:       1.2,
		PassEngagementOptimization:  1.1,
		PassFinalPolish:             1.0,
	}

	if weight, exists := weights[pass]; exists {
		return weight
	}
	return 1.0
}
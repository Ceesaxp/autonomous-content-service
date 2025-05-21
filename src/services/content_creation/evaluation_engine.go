package content_creation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// EvaluationEngine handles detailed content evaluation against multiple criteria
type EvaluationEngine struct {
	llmClient LLMClient
}

// NewEvaluationEngine creates a new evaluation engine
func NewEvaluationEngine(llmClient LLMClient) *EvaluationEngine {
	return &EvaluationEngine{
		llmClient: llmClient,
	}
}

// EvaluationRequest contains parameters for content evaluation
type EvaluationRequest struct {
	Content        string
	ContentType    entities.ContentType
	Criteria       []EvaluationCriterion
	TargetAudience string
	Context        map[string]interface{}
}

// DetailedEvaluation contains comprehensive evaluation results
type DetailedEvaluation struct {
	OverallScore     float64                    `json:"overallScore"`
	CriteriaResults  map[string]CriterionResult `json:"criteriaResults"`
	Strengths        []string                   `json:"strengths"`
	Weaknesses       []string                   `json:"weaknesses"`
	Recommendations  []string                   `json:"recommendations"`
	DetailedAnalysis string                     `json:"detailedAnalysis"`
}

// CriterionResult contains results for a specific evaluation criterion
type CriterionResult struct {
	Score        float64  `json:"score"`
	MaxScore     float64  `json:"maxScore"`
	Explanation  string   `json:"explanation"`
	Evidence     []string `json:"evidence"`
	Suggestions  []string `json:"suggestions"`
	Confidence   float64  `json:"confidence"`
}

// EvaluateContent performs comprehensive evaluation of content against specified criteria
func (e *EvaluationEngine) EvaluateContent(ctx context.Context, request EvaluationRequest) (*DetailedEvaluation, error) {
	evaluation := &DetailedEvaluation{
		CriteriaResults: make(map[string]CriterionResult),
		Strengths:       []string{},
		Weaknesses:      []string{},
		Recommendations: []string{},
	}

	// Evaluate each criterion
	totalScore := 0.0
	totalWeight := 0.0

	for _, criterion := range request.Criteria {
		result, err := e.evaluateCriterion(ctx, request.Content, request.ContentType, criterion, request.TargetAudience)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate criterion %s: %w", criterion, err)
		}

		evaluation.CriteriaResults[string(criterion)] = *result
		weight := getCriterionWeight(criterion, request.ContentType)
		totalScore += result.Score * weight
		totalWeight += weight

		// Collect strengths and weaknesses
		if result.Score >= 80.0 {
			evaluation.Strengths = append(evaluation.Strengths, 
				fmt.Sprintf("%s: %s", criterion, result.Explanation))
		} else if result.Score < 60.0 {
			evaluation.Weaknesses = append(evaluation.Weaknesses, 
				fmt.Sprintf("%s: %s", criterion, result.Explanation))
		}

		// Collect recommendations
		evaluation.Recommendations = append(evaluation.Recommendations, result.Suggestions...)
	}

	// Calculate overall score
	if totalWeight > 0 {
		evaluation.OverallScore = totalScore / totalWeight
	}

	// Generate detailed analysis
	analysisPrompt := e.createAnalysisPrompt(request, evaluation)
	analysis, err := e.llmClient.Generate(ctx, analysisPrompt)
	if err != nil {
		// Continue without detailed analysis if LLM fails
		evaluation.DetailedAnalysis = "Detailed analysis unavailable"
	} else {
		evaluation.DetailedAnalysis = analysis
	}

	return evaluation, nil
}

// evaluateCriterion evaluates content against a specific criterion
func (e *EvaluationEngine) evaluateCriterion(ctx context.Context, content string, contentType entities.ContentType, criterion EvaluationCriterion, targetAudience string) (*CriterionResult, error) {
	switch criterion {
	case CriterionReadability:
		return e.evaluateReadability(ctx, content, contentType)
	case CriterionAccuracy:
		return e.evaluateAccuracy(ctx, content, contentType)
	case CriterionEngagement:
		return e.evaluateEngagement(ctx, content, contentType, targetAudience)
	case CriterionClarity:
		return e.evaluateClarity(ctx, content, contentType)
	case CriterionCoherence:
		return e.evaluateCoherence(ctx, content, contentType)
	case CriterionCompleteness:
		return e.evaluateCompleteness(ctx, content, contentType)
	case CriterionRelevance:
		return e.evaluateRelevance(ctx, content, contentType, targetAudience)
	case CriterionOriginality:
		return e.evaluateOriginality(ctx, content, contentType)
	case CriterionTone:
		return e.evaluateTone(ctx, content, contentType, targetAudience)
	case CriterionStructure:
		return e.evaluateStructure(ctx, content, contentType)
	case CriterionGrammar:
		return e.evaluateGrammar(ctx, content, contentType)
	case CriterionSEO:
		return e.evaluateSEO(ctx, content, contentType)
	case CriterionCallToAction:
		return e.evaluateCallToAction(ctx, content, contentType)
	case CriterionEmotionalImpact:
		return e.evaluateEmotionalImpact(ctx, content, contentType, targetAudience)
	case CriterionCredibility:
		return e.evaluateCredibility(ctx, content, contentType)
	default:
		return e.evaluateGeneric(ctx, content, contentType, criterion)
	}
}

// evaluateReadability assesses content readability
func (e *EvaluationEngine) evaluateReadability(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the readability of the following %s content. Consider:
1. Sentence length and complexity
2. Vocabulary difficulty
3. Paragraph structure
4. Use of active vs passive voice
5. Clarity of expression

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateAccuracy assesses factual accuracy
func (e *EvaluationEngine) evaluateAccuracy(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the factual accuracy of the following %s content. Consider:
1. Verifiable facts and claims
2. Consistency of information
3. Use of credible sources
4. Absence of contradictions
5. Currency of information

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateEngagement assesses engagement potential
func (e *EvaluationEngine) evaluateEngagement(ctx context.Context, content string, contentType entities.ContentType, targetAudience string) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the engagement potential of the following %s content for %s audience. Consider:
1. Hook and opening effectiveness
2. Use of storytelling elements
3. Interactive elements or questions
4. Emotional connection
5. Call to action effectiveness
6. Visual appeal (if applicable)

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, targetAudience, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateClarity assesses content clarity
func (e *EvaluationEngine) evaluateClarity(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the clarity of the following %s content. Consider:
1. Clear and specific language
2. Logical flow of ideas
3. Absence of ambiguity
4. Effective use of examples
5. Consistent terminology

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateCoherence assesses content coherence
func (e *EvaluationEngine) evaluateCoherence(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the coherence of the following %s content. Consider:
1. Logical organization and structure
2. Smooth transitions between ideas
3. Consistent theme throughout
4. Unified message and purpose
5. Supporting details align with main points

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateCompleteness assesses content completeness
func (e *EvaluationEngine) evaluateCompleteness(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the completeness of the following %s content. Consider:
1. All important topics covered
2. Sufficient depth and detail
3. Addresses potential questions
4. Provides necessary context
5. Meets content type expectations

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateRelevance assesses content relevance
func (e *EvaluationEngine) evaluateRelevance(ctx context.Context, content string, contentType entities.ContentType, targetAudience string) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the relevance of the following %s content for %s audience. Consider:
1. Addresses audience needs and interests
2. Timely and current information
3. Practical applicability
4. Appropriate level of detail
5. Aligns with audience expectations

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, targetAudience, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateOriginality assesses content originality
func (e *EvaluationEngine) evaluateOriginality(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the originality of the following %s content. Consider:
1. Unique perspective or approach
2. Fresh insights or ideas
3. Creative presentation
4. Avoids clichés and overused phrases
5. Brings new value to the topic

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateTone assesses content tone appropriateness
func (e *EvaluationEngine) evaluateTone(ctx context.Context, content string, contentType entities.ContentType, targetAudience string) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the tone appropriateness of the following %s content for %s audience. Consider:
1. Matches intended audience expectations
2. Consistent throughout the content
3. Appropriate formality level
4. Emotional resonance
5. Brand voice alignment (if applicable)

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, targetAudience, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateStructure assesses content structure
func (e *EvaluationEngine) evaluateStructure(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the structure of the following %s content. Consider:
1. Clear introduction, body, and conclusion
2. Logical organization of sections
3. Appropriate use of headings and subheadings
4. Effective paragraph breaks
5. Content type best practices

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateGrammar assesses grammar and language usage
func (e *EvaluationEngine) evaluateGrammar(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the grammar and language usage of the following %s content. Consider:
1. Correct grammar and syntax
2. Proper spelling and punctuation
3. Appropriate word choice
4. Sentence variety
5. Professional language standards

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateSEO assesses SEO optimization
func (e *EvaluationEngine) evaluateSEO(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the SEO optimization of the following %s content. Consider:
1. Keyword usage and density
2. Meta elements potential
3. Header tag structure
4. Content length and depth
5. Internal/external linking opportunities

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateCallToAction assesses call-to-action effectiveness
func (e *EvaluationEngine) evaluateCallToAction(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the call-to-action effectiveness of the following %s content. Consider:
1. Clear and specific action requested
2. Compelling and motivating language
3. Appropriate placement and timing
4. Easy to follow instructions
5. Aligned with content goals

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateEmotionalImpact assesses emotional impact
func (e *EvaluationEngine) evaluateEmotionalImpact(ctx context.Context, content string, contentType entities.ContentType, targetAudience string) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the emotional impact of the following %s content for %s audience. Consider:
1. Evokes appropriate emotions
2. Creates emotional connection
3. Uses compelling language
4. Includes relatable scenarios
5. Balances emotional and rational appeals

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, targetAudience, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateCredibility assesses content credibility
func (e *EvaluationEngine) evaluateCredibility(ctx context.Context, content string, contentType entities.ContentType) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the credibility of the following %s content. Consider:
1. Use of authoritative sources
2. Professional presentation
3. Balanced perspective
4. Transparent about limitations
5. Evidence-based claims

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// evaluateGeneric handles custom or unknown criteria
func (e *EvaluationEngine) evaluateGeneric(ctx context.Context, content string, contentType entities.ContentType, criterion EvaluationCriterion) (*CriterionResult, error) {
	prompt := fmt.Sprintf(`Evaluate the following %s content against the criterion of "%s". 
Provide a detailed assessment considering what this criterion means for this type of content.

Rate on a scale of 0-100 and provide specific evidence and suggestions.

Content to evaluate:
%s

Respond in JSON format:
{
  "score": <number>,
  "explanation": "<brief explanation>",
  "evidence": ["<evidence point 1>", "<evidence point 2>"],
  "suggestions": ["<suggestion 1>", "<suggestion 2>"],
  "confidence": <0-1>
}`, contentType, criterion, content)

	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return e.parseEvaluationResponse(response, 100.0)
}

// parseEvaluationResponse parses LLM response into CriterionResult
func (e *EvaluationEngine) parseEvaluationResponse(response string, maxScore float64) (*CriterionResult, error) {
	// Try to parse as JSON first
	var result CriterionResult
	if err := json.Unmarshal([]byte(response), &result); err == nil {
		result.MaxScore = maxScore
		return &result, nil
	}

	// Fallback to regex parsing
	result = CriterionResult{
		MaxScore:    maxScore,
		Evidence:    []string{},
		Suggestions: []string{},
		Confidence:  0.8, // Default confidence
	}

	// Extract score
	scoreRegex := regexp.MustCompile(`(?i)score["\s]*:?\s*(\d+(?:\.\d+)?)`)
	scoreMatch := scoreRegex.FindStringSubmatch(response)
	if len(scoreMatch) > 1 {
		if score, err := strconv.ParseFloat(scoreMatch[1], 64); err == nil {
			result.Score = score
		}
	}

	// Extract explanation
	explanationRegex := regexp.MustCompile(`(?i)explanation["\s]*:?\s*["']([^"']+)["']`)
	explanationMatch := explanationRegex.FindStringSubmatch(response)
	if len(explanationMatch) > 1 {
		result.Explanation = explanationMatch[1]
	} else {
		// Fallback: use first sentence of response
		sentences := strings.Split(response, ".")
		if len(sentences) > 0 {
			result.Explanation = strings.TrimSpace(sentences[0])
		}
	}

	// Extract evidence and suggestions from response (simplified)
	lines := strings.Split(response, "\n")
	inEvidence := false
	inSuggestions := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lowerLine := strings.ToLower(line)
		
		if strings.Contains(lowerLine, "evidence") {
			inEvidence = true
			inSuggestions = false
			continue
		} else if strings.Contains(lowerLine, "suggestion") {
			inSuggestions = true
			inEvidence = false
			continue
		}
		
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•") {
			cleanLine := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*"), "•")
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				if inEvidence {
					result.Evidence = append(result.Evidence, cleanLine)
				} else if inSuggestions {
					result.Suggestions = append(result.Suggestions, cleanLine)
				}
			}
		}
	}

	return &result, nil
}

// createAnalysisPrompt creates a prompt for detailed analysis
func (e *EvaluationEngine) createAnalysisPrompt(request EvaluationRequest, evaluation *DetailedEvaluation) string {
	return fmt.Sprintf(`Provide a comprehensive analysis of the following %s content evaluation results:

Overall Score: %.1f/100

Criteria Results:
%s

Strengths:
%s

Weaknesses:
%s

Please provide a detailed analysis that:
1. Explains the overall quality assessment
2. Identifies the most critical areas for improvement
3. Highlights what works well in the content
4. Provides strategic recommendations for enhancement
5. Considers the target audience: %s

Keep the analysis professional, actionable, and specific to the content type.`,
		request.ContentType,
		evaluation.OverallScore,
		formatCriteriaResults(evaluation.CriteriaResults),
		strings.Join(evaluation.Strengths, "; "),
		strings.Join(evaluation.Weaknesses, "; "),
		request.TargetAudience,
	)
}

// formatCriteriaResults formats criteria results for display
func formatCriteriaResults(results map[string]CriterionResult) string {
	var formatted []string
	for criterion, result := range results {
		formatted = append(formatted, fmt.Sprintf("- %s: %.1f/%.1f", criterion, result.Score, result.MaxScore))
	}
	return strings.Join(formatted, "\n")
}

// getCriterionWeight returns the importance weight for a criterion based on content type
func getCriterionWeight(criterion EvaluationCriterion, contentType entities.ContentType) float64 {
	// Base weights
	weights := map[EvaluationCriterion]float64{
		CriterionReadability:    1.2,
		CriterionAccuracy:       1.5,
		CriterionEngagement:     1.3,
		CriterionClarity:        1.2,
		CriterionCoherence:      1.1,
		CriterionCompleteness:   1.0,
		CriterionRelevance:      1.2,
		CriterionOriginality:    1.0,
		CriterionTone:           1.1,
		CriterionStructure:      1.1,
		CriterionGrammar:        1.3,
		CriterionSEO:            1.0,
		CriterionCallToAction:   1.0,
		CriterionEmotionalImpact: 1.0,
		CriterionCredibility:    1.2,
	}

	baseWeight := weights[criterion]
	if baseWeight == 0 {
		baseWeight = 1.0
	}

	// Adjust weights based on content type
	switch contentType {
	case entities.ContentTypeTechnicalArticle:
		if criterion == CriterionAccuracy || criterion == CriterionCredibility {
			return baseWeight * 1.5
		}
	case entities.ContentTypeSocialPost:
		if criterion == CriterionEngagement || criterion == CriterionEmotionalImpact {
			return baseWeight * 1.5
		}
	case entities.ContentTypeBlogPost:
		if criterion == CriterionSEO || criterion == CriterionEngagement {
			return baseWeight * 1.3
		}
	case entities.ContentTypeProductDescription:
		if criterion == CriterionCallToAction || criterion == CriterionCredibility {
			return baseWeight * 1.4
		}
	}

	return baseWeight
}
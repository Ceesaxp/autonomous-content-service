package content_creation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// QualityCheckInput contains inputs for the quality checking process
type QualityCheckInput struct {
	Content           string
	CheckPlagiarism   bool
	CheckFactAccuracy bool
	EvaluateSEO       bool
}

// QualityCheckOutput contains the results of quality assessment
type QualityCheckOutput struct {
	ReadabilityScore    float64               `json:"readabilityScore"`
	SEOScore           float64               `json:"seoScore"`
	EngagementScore    float64               `json:"engagementScore"`
	PlagiarismScore    float64               `json:"plagiarismScore"`
	FactualErrors      []FactualError        `json:"factualErrors,omitempty"`
	SuggestionsByCategory map[string][]string `json:"suggestionsByCategory,omitempty"`
	Keywords           []string              `json:"keywords,omitempty"`
}

// FactualError represents a factual error in content
type FactualError struct {
	ErrorText  string `json:"errorText"`
	Correction string `json:"correction"`
	Source     string `json:"source,omitempty"`
}

// QualityChecker defines the interface for content quality assessment
type QualityChecker interface {
	// CheckContent assesses the quality of content
	CheckContent(ctx context.Context, content *entities.Content, input QualityCheckInput) (QualityCheckOutput, error)
}

// LLMQualityChecker uses LLM for content quality assessment
type LLMQualityChecker struct {
	LLMClient        LLMClient
	PlagiarismAPI    PlagiarismAPI
	ReadabilityScorer ReadabilityScorer
	SEOAnalyzer      SEOAnalyzer
}

// NewLLMQualityChecker creates a new quality checker with LLM capabilities
func NewLLMQualityChecker(llmClient LLMClient, plagiarismAPI PlagiarismAPI, readabilityScorer ReadabilityScorer, seoAnalyzer SEOAnalyzer) *LLMQualityChecker {
	return &LLMQualityChecker{
		LLMClient:        llmClient,
		PlagiarismAPI:    plagiarismAPI,
		ReadabilityScorer: readabilityScorer,
		SEOAnalyzer:      seoAnalyzer,
	}
}

// CheckContent assesses the quality of content
func (q *LLMQualityChecker) CheckContent(ctx context.Context, content *entities.Content, input QualityCheckInput) (QualityCheckOutput, error) {
	// Initialize output
	output := QualityCheckOutput{
		FactualErrors:      []FactualError{},
		SuggestionsByCategory: make(map[string][]string),
		Keywords:           []string{},
	}

	// Run readability analysis
	readabilityScore, readabilitySuggestions, err := q.ReadabilityScorer.AnalyzeReadability(ctx, input.Content)
	if err != nil {
		return output, fmt.Errorf("readability analysis failed: %w", err)
	}
	output.ReadabilityScore = readabilityScore
	output.SuggestionsByCategory["Readability"] = readabilitySuggestions

	// Run engagement analysis using LLM
	engagementScore, engagementSuggestions, err := q.analyzeEngagement(ctx, input.Content, content.Type)
	if err != nil {
		return output, fmt.Errorf("engagement analysis failed: %w", err)
	}
	output.EngagementScore = engagementScore
	output.SuggestionsByCategory["Engagement"] = engagementSuggestions

	// Check for plagiarism if requested
	if input.CheckPlagiarism {
		plagiarismScore, plagiarismDetails, err := q.PlagiarismAPI.CheckPlagiarism(ctx, input.Content)
		if err != nil {
			// Log the error but continue with other checks
			fmt.Printf("Plagiarism check failed: %v\n", err)
			// Use a default high score (assuming 1.0 is completely original)
			output.PlagiarismScore = 0.95
		} else {
			output.PlagiarismScore = plagiarismScore
			if plagiarismScore < 0.8 {
				plagiarismSuggestions := []string{}
				for _, detail := range plagiarismDetails {
					plagiarismSuggestions = append(plagiarismSuggestions, 
						fmt.Sprintf("Potential plagiarism: '%s' matches source '%s'", 
							detail.Fragment, detail.Source))
				}
				output.SuggestionsByCategory["Plagiarism"] = plagiarismSuggestions
			}
		}
	} else {
		// Skip plagiarism check
		output.PlagiarismScore = 1.0
	}

	// Check factual accuracy if requested
	if input.CheckFactAccuracy {
		factErrors, err := q.checkFactualAccuracy(ctx, input.Content)
		if err != nil {
			// Log the error but continue with other checks
			fmt.Printf("Factual accuracy check failed: %v\n", err)
		} else {
			output.FactualErrors = factErrors
		}
	}

	// Evaluate SEO if requested
	if input.EvaluateSEO {
		seoScore, keywords, seoSuggestions, err := q.SEOAnalyzer.AnalyzeSEO(ctx, content.Title, input.Content)
		if err != nil {
			// Log the error but continue with other checks
			fmt.Printf("SEO analysis failed: %v\n", err)
			output.SEOScore = 70.0 // Default moderate score
		} else {
			output.SEOScore = seoScore
			output.Keywords = keywords
			output.SuggestionsByCategory["SEO"] = seoSuggestions
		}
	} else {
		// Skip SEO evaluation
		output.SEOScore = 75.0 // Default moderate score
	}

	return output, nil
}

// analyzeEngagement evaluates content engagement using LLM
func (q *LLMQualityChecker) analyzeEngagement(ctx context.Context, content string, contentType entities.ContentType) (float64, []string, error) {
	// Create a prompt for engagement analysis
	prompt := fmt.Sprintf(
		"Analyze the engagement quality of the following %s. " +
		"Rate it on a scale of 0-100 and provide specific suggestions for improvement. " +
		"Focus on narrative flow, use of examples, compelling arguments, emotional appeal, and audience connection. " +
		"Format your response as a JSON object with 'score' (number) and 'suggestions' (array of strings).\n\n" +
		"Content to analyze:\n%s",
		contentType,
		content,
	)

	// Use LLM to analyze engagement
	response, err := q.LLMClient.Generate(ctx, prompt)
	if err != nil {
		return 0, nil, fmt.Errorf("engagement analysis generation failed: %w", err)
	}

	// Parse the response to extract score and suggestions
	score, suggestions := parseEngagementAnalysis(response)

	return score, suggestions, nil
}

// parseEngagementAnalysis extracts engagement score and suggestions from LLM response
func parseEngagementAnalysis(response string) (float64, []string) {
	// Look for JSON structure in response
	jsonRegex := regexp.MustCompile(`\{[\s\S]*\}`)
	jsonMatch := jsonRegex.FindString(response)

	// Default values
	score := 70.0 // Moderate default score
	suggestions := []string{}

	// Parse JSON if found
	if jsonMatch != "" {
		// Extract score
		scoreRegex := regexp.MustCompile(`"score"\s*:\s*(\d+(\.\d+)?)`)
		scoreMatch := scoreRegex.FindStringSubmatch(jsonMatch)
		if len(scoreMatch) > 1 {
			fmt.Sscanf(scoreMatch[1], "%f", &score)
		}

		// Extract suggestions
		suggRegex := regexp.MustCompile(`"suggestions"\s*:\s*\[(.*?)\]`)
		suggMatch := suggRegex.FindStringSubmatch(jsonMatch)
		if len(suggMatch) > 1 {
			// Split suggestions by comma, considering quoted strings
			suggItems := strings.Split(suggMatch[1], ",")
			for _, item := range suggItems {
				// Clean up and extract the suggestion text
				item = strings.TrimSpace(item)
				item = strings.Trim(item, `"'`)
				if item != "" {
					suggestions = append(suggestions, item)
				}
			}
		}
	} else {
		// Fallback parsing for non-JSON response
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			// Try to extract score
			if strings.Contains(strings.ToLower(line), "score") {
				scoreRegex := regexp.MustCompile(`(\d+(\.\d+)?)`)
				scoreMatch := scoreRegex.FindString(line)
				if scoreMatch != "" {
					fmt.Sscanf(scoreMatch, "%f", &score)
				}
			}

			// Extract suggestion-like lines
			if strings.Contains(line, ":") || 
			   strings.HasPrefix(strings.TrimSpace(line), "-") || 
			   strings.HasPrefix(strings.TrimSpace(line), "*") {
				// Clean up the line
				line = strings.TrimSpace(line)
				line = regexp.MustCompile(`^\d+\.\s*|^-\s*|^\*\s*`).ReplaceAllString(line, "")
				if line != "" && !strings.Contains(strings.ToLower(line), "score") {
					suggestions = append(suggestions, line)
				}
			}
		}
	}

	// Ensure score is within 0-100 range
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score, suggestions
}

// checkFactualAccuracy verifies facts in content using LLM and external data
func (q *LLMQualityChecker) checkFactualAccuracy(ctx context.Context, content string) ([]FactualError, error) {
	// Create a prompt for fact-checking
	prompt := fmt.Sprintf(
		"Fact-check the following content. Identify any factual errors or inaccuracies. " +
		"For each error, provide the incorrect statement, the correct information, and a reliable source if possible. " +
		"Format your response as a JSON array of objects, each with 'errorText', 'correction', and 'source' fields.\n\n" +
		"Content to fact-check:\n%s",
		content,
	)

	// Use LLM to fact-check content
	response, err := q.LLMClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("fact-checking failed: %w", err)
	}

	// Parse the response to extract factual errors
	return parseFactualErrors(response), nil
}

// parseFactualErrors extracts factual errors from LLM response
func parseFactualErrors(response string) []FactualError {
	// Look for JSON array in response
	jsonRegex := regexp.MustCompile(`\[[\s\S]*\]`)
	jsonMatch := jsonRegex.FindString(response)

	errors := []FactualError{}

	// Parse JSON if found
	if jsonMatch != "" {
		// Extract individual error objects
		objRegex := regexp.MustCompile(`\{[^\{\}]*\}`)
		objMatches := objRegex.FindAllString(jsonMatch, -1)

		for _, objMatch := range objMatches {
			// Extract error text
			errorRegex := regexp.MustCompile(`"errorText"\s*:\s*"([^"]*)"`)
			errorMatch := errorRegex.FindStringSubmatch(objMatch)
			if len(errorMatch) < 2 {
				continue
			}
			errorText := errorMatch[1]

			// Extract correction
			correctionRegex := regexp.MustCompile(`"correction"\s*:\s*"([^"]*)"`)
			correctionMatch := correctionRegex.FindStringSubmatch(objMatch)
			correction := ""
			if len(correctionMatch) > 1 {
				correction = correctionMatch[1]
			}

			// Extract source
			sourceRegex := regexp.MustCompile(`"source"\s*:\s*"([^"]*)"`)
			sourceMatch := sourceRegex.FindStringSubmatch(objMatch)
			source := ""
			if len(sourceMatch) > 1 {
				source = sourceMatch[1]
			}

			// Add to errors
			errors = append(errors, FactualError{
				ErrorText:  errorText,
				Correction: correction,
				Source:     source,
			})
		}
	} else {
		// Fallback parsing for non-JSON response
		// This is a simplified approach; a real implementation would be more robust
		lines := strings.Split(response, "\n")
		currentError := FactualError{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				// Empty line might separate errors
				if currentError.ErrorText != "" {
					errors = append(errors, currentError)
					currentError = FactualError{}
				}
				continue
			}

			// Try to identify parts of the error
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "incorrect") {
				currentError.ErrorText = extractContent(line)
			} else if strings.Contains(lowerLine, "correct") {
				currentError.Correction = extractContent(line)
			} else if strings.Contains(lowerLine, "source") {
				currentError.Source = extractContent(line)
			}
		}

		// Add the last error if not empty
		if currentError.ErrorText != "" {
			errors = append(errors, currentError)
		}
	}

	return errors
}

// extractContent extracts content after a colon or similar separator
func extractContent(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(line)
}

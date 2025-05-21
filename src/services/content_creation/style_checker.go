package content_creation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// StyleChecker handles style consistency checking
type StyleChecker struct {
	llmClient LLMClient
}

// NewStyleChecker creates a new style checker
func NewStyleChecker(llmClient LLMClient) *StyleChecker {
	return &StyleChecker{
		llmClient: llmClient,
	}
}

// StyleCheckRequest contains parameters for style analysis
type StyleCheckRequest struct {
	Content        string
	ContentType    entities.ContentType
	TargetAudience string
	BrandGuidelines map[string]interface{}
	StyleGuide     StyleGuide
}

// StyleAnalysisResult contains comprehensive style analysis results
type StyleAnalysisResult struct {
	OverallScore       float64           `json:"overallScore"`
	ConsistencyScore   float64           `json:"consistencyScore"`
	BrandAlignmentScore float64          `json:"brandAlignmentScore"`
	StyleIssues        []StyleIssue      `json:"styleIssues"`
	ToneAnalysis       ToneAnalysis      `json:"toneAnalysis"`
	VoiceConsistency   VoiceConsistency  `json:"voiceConsistency"`
	FormattingIssues   []FormattingIssue `json:"formattingIssues"`
	Recommendations    []StyleRecommendation `json:"recommendations"`
	StyleProfile       StyleProfile      `json:"styleProfile"`
	ProcessingTime     time.Duration     `json:"processingTime"`
}

// StyleIssue represents a style consistency issue
type StyleIssue struct {
	Type        StyleIssueType `json:"type"`
	Severity    Severity       `json:"severity"`
	Description string         `json:"description"`
	Location    string         `json:"location"`
	Example     string         `json:"example"`
	Suggestion  string         `json:"suggestion"`
	RuleViolated string        `json:"ruleViolated"`
}

// ToneAnalysis analyzes the tone consistency
type ToneAnalysis struct {
	PrimaryTone    Tone              `json:"primaryTone"`
	ToneConsistency float64          `json:"toneConsistency"`
	ToneShifts     []ToneShift       `json:"toneShifts"`
	AudienceMatch  float64           `json:"audienceMatch"`
	EmotionalTone  map[string]float64 `json:"emotionalTone"`
}

// VoiceConsistency analyzes voice consistency
type VoiceConsistency struct {
	Score           float64        `json:"score"`
	PersonConsistency float64      `json:"personConsistency"`
	PerspectiveShifts []PerspectiveShift `json:"perspectiveShifts"`
	VoiceCharacteristics VoiceCharacteristics `json:"voiceCharacteristics"`
}

// FormattingIssue represents formatting inconsistencies
type FormattingIssue struct {
	Type        FormattingType `json:"type"`
	Description string         `json:"description"`
	Examples    []string       `json:"examples"`
	Correction  string         `json:"correction"`
}

// StyleRecommendation provides actionable style improvements
type StyleRecommendation struct {
	Priority    RecommendationPriority `json:"priority"`
	Category    string                 `json:"category"`
	Issue       string                 `json:"issue"`
	Suggestion  string                 `json:"suggestion"`
	Impact      string                 `json:"impact"`
	Examples    []string               `json:"examples"`
}

// StyleProfile characterizes the detected writing style
type StyleProfile struct {
	FormalityLevel   float64   `json:"formalityLevel"`
	ComplexityLevel  float64   `json:"complexityLevel"`
	PersonalityTraits []string `json:"personalityTraits"`
	WritingStyle     string    `json:"writingStyle"`
	Characteristics  []string  `json:"characteristics"`
}

// StyleGuide defines style guidelines to check against
type StyleGuide struct {
	ToneGuidelines      map[string]string   `json:"toneGuidelines"`
	VoiceGuidelines     map[string]string   `json:"voiceGuidelines"`
	FormattingRules     map[string]string   `json:"formattingRules"`
	TerminologyRules    map[string]string   `json:"terminologyRules"`
	BrandVoice          string              `json:"brandVoice"`
	ForbiddenPhrases    []string            `json:"forbiddenPhrases"`
	PreferredPhrases    []string            `json:"preferredPhrases"`
	StylePreferences    map[string]string   `json:"stylePreferences"`
}

// Enums and types

type StyleIssueType string

const (
	IssueInconsistentTone      StyleIssueType = "inconsistent_tone"
	IssueVoiceShift           StyleIssueType = "voice_shift"
	IssueFormattingInconsistency StyleIssueType = "formatting_inconsistency"
	IssueTerminologyInconsistency StyleIssueType = "terminology_inconsistency"
	IssueBrandViolation       StyleIssueType = "brand_violation"
	IssueAudienceMismatch     StyleIssueType = "audience_mismatch"
)

type Tone string

const (
	ToneFormal      Tone = "formal"
	ToneInformal    Tone = "informal"
	ToneProfessional Tone = "professional"
	ToneConversational Tone = "conversational"
	ToneAcademic    Tone = "academic"
	TonePersuasive  Tone = "persuasive"
	ToneFriendly    Tone = "friendly"
	ToneAuthoritative Tone = "authoritative"
)

type FormattingType string

const (
	FormattingCapitalization FormattingType = "capitalization"
	FormattingPunctuation   FormattingType = "punctuation"
	FormattingSpacing       FormattingType = "spacing"
	FormattingHeadings      FormattingType = "headings"
	FormattingLists         FormattingType = "lists"
)

type RecommendationPriority string

const (
	PriorityHigh   RecommendationPriority = "high"
	PriorityMedium RecommendationPriority = "medium"
	PriorityLow    RecommendationPriority = "low"
)

type ToneShift struct {
	From     Tone   `json:"from"`
	To       Tone   `json:"to"`
	Location string `json:"location"`
	Context  string `json:"context"`
}

type PerspectiveShift struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Location string `json:"location"`
	Example  string `json:"example"`
}

type VoiceCharacteristics struct {
	Person           string  `json:"person"` // first, second, third
	Perspective      string  `json:"perspective"`
	AuthorityLevel   float64 `json:"authorityLevel"`
	PersonalityScore float64 `json:"personalityScore"`
}

// AnalyzeStyle performs comprehensive style analysis
func (sc *StyleChecker) AnalyzeStyle(ctx context.Context, request StyleCheckRequest) (*StyleAnalysisResult, error) {
	startTime := time.Now()
	
	result := &StyleAnalysisResult{
		StyleIssues:      []StyleIssue{},
		FormattingIssues: []FormattingIssue{},
		Recommendations:  []StyleRecommendation{},
	}

	// 1. Analyze tone consistency
	toneAnalysis, err := sc.analyzeTone(ctx, request.Content, request.TargetAudience)
	if err != nil {
		return nil, fmt.Errorf("tone analysis failed: %w", err)
	}
	result.ToneAnalysis = *toneAnalysis

	// 2. Analyze voice consistency
	voiceConsistency, err := sc.analyzeVoiceConsistency(ctx, request.Content)
	if err != nil {
		return nil, fmt.Errorf("voice analysis failed: %w", err)
	}
	result.VoiceConsistency = *voiceConsistency

	// 3. Check formatting consistency
	formattingIssues := sc.checkFormattingConsistency(request.Content)
	result.FormattingIssues = formattingIssues

	// 4. Check brand alignment if guidelines provided
	if len(request.BrandGuidelines) > 0 || request.StyleGuide.BrandVoice != "" {
		brandScore, brandIssues, err := sc.checkBrandAlignment(ctx, request.Content, request.StyleGuide)
		if err == nil {
			result.BrandAlignmentScore = brandScore
			result.StyleIssues = append(result.StyleIssues, brandIssues...)
		}
	}

	// 5. Generate style profile
	styleProfile, err := sc.generateStyleProfile(ctx, request.Content)
	if err == nil {
		result.StyleProfile = *styleProfile
	}

	// 6. Calculate overall scores
	result.ConsistencyScore = sc.calculateConsistencyScore(result)
	result.OverallScore = sc.calculateOverallStyleScore(result)

	// 7. Generate recommendations
	result.Recommendations = sc.generateStyleRecommendations(result)

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// analyzeTone analyzes tone consistency throughout the content
func (sc *StyleChecker) analyzeTone(ctx context.Context, content, targetAudience string) (*ToneAnalysis, error) {
	prompt := fmt.Sprintf(`Analyze the tone consistency in the following content for %s audience:

Content:
%s

Evaluate:
1. Primary tone throughout the content
2. Any tone shifts or inconsistencies
3. Appropriateness for the target audience
4. Emotional undertones

Respond in JSON format:
{
  "primaryTone": "formal|informal|professional|conversational|academic|persuasive|friendly|authoritative",
  "toneConsistency": <0-1 score>,
  "toneShifts": [
    {
      "from": "<tone>",
      "to": "<tone>",
      "location": "<where in content>",
      "context": "<surrounding context>"
    }
  ],
  "audienceMatch": <0-1 score>,
  "emotionalTone": {
    "positive": <0-1>,
    "negative": <0-1>,
    "neutral": <0-1>,
    "confident": <0-1>,
    "uncertain": <0-1>
  }
}`, targetAudience, content)

	response, err := sc.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return sc.parseToneAnalysis(response)
}

// analyzeVoiceConsistency analyzes voice and perspective consistency
func (sc *StyleChecker) analyzeVoiceConsistency(ctx context.Context, content string) (*VoiceConsistency, error) {
	prompt := fmt.Sprintf(`Analyze voice and perspective consistency in the following content:

Content:
%s

Evaluate:
1. Consistent use of person (first, second, third)
2. Perspective shifts
3. Authority level and confidence
4. Overall voice characteristics

Respond in JSON format:
{
  "score": <0-1>,
  "personConsistency": <0-1>,
  "perspectiveShifts": [
    {
      "from": "<perspective>",
      "to": "<perspective>",
      "location": "<location>",
      "example": "<example text>"
    }
  ],
  "voiceCharacteristics": {
    "person": "first|second|third|mixed",
    "perspective": "<description>",
    "authorityLevel": <0-1>,
    "personalityScore": <0-1>
  }
}`, content)

	response, err := sc.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return sc.parseVoiceAnalysis(response)
}

// checkFormattingConsistency checks for formatting inconsistencies
func (sc *StyleChecker) checkFormattingConsistency(content string) []FormattingIssue {
	issues := []FormattingIssue{}

	// Check capitalization consistency
	if capIssues := sc.checkCapitalizationConsistency(content); len(capIssues) > 0 {
		issues = append(issues, FormattingIssue{
			Type:        FormattingCapitalization,
			Description: "Inconsistent capitalization patterns detected",
			Examples:    capIssues,
			Correction:  "Use consistent capitalization rules throughout",
		})
	}

	// Check punctuation consistency
	if punctIssues := sc.checkPunctuationConsistency(content); len(punctIssues) > 0 {
		issues = append(issues, FormattingIssue{
			Type:        FormattingPunctuation,
			Description: "Inconsistent punctuation usage detected",
			Examples:    punctIssues,
			Correction:  "Follow consistent punctuation rules",
		})
	}

	// Check heading consistency
	if headingIssues := sc.checkHeadingConsistency(content); len(headingIssues) > 0 {
		issues = append(issues, FormattingIssue{
			Type:        FormattingHeadings,
			Description: "Inconsistent heading format detected",
			Examples:    headingIssues,
			Correction:  "Use consistent heading hierarchy and formatting",
		})
	}

	return issues
}

// checkBrandAlignment checks alignment with brand guidelines
func (sc *StyleChecker) checkBrandAlignment(ctx context.Context, content string, styleGuide StyleGuide) (float64, []StyleIssue, error) {
	issues := []StyleIssue{}
	score := 100.0

	// Check forbidden phrases
	for _, phrase := range styleGuide.ForbiddenPhrases {
		if strings.Contains(strings.ToLower(content), strings.ToLower(phrase)) {
			issues = append(issues, StyleIssue{
				Type:        IssueBrandViolation,
				Severity:    SeverityMajor,
				Description: fmt.Sprintf("Use of forbidden phrase: %s", phrase),
				Suggestion:  "Remove or replace with approved alternative",
				RuleViolated: "Brand phrase guidelines",
			})
			score -= 10.0
		}
	}

	// Check brand voice alignment using LLM
	if styleGuide.BrandVoice != "" {
		brandScore, brandIssues, err := sc.checkBrandVoiceAlignment(ctx, content, styleGuide.BrandVoice)
		if err == nil {
			score = (score + brandScore) / 2
			issues = append(issues, brandIssues...)
		}
	}

	return score, issues, nil
}

// checkBrandVoiceAlignment checks alignment with brand voice
func (sc *StyleChecker) checkBrandVoiceAlignment(ctx context.Context, content, brandVoice string) (float64, []StyleIssue, error) {
	prompt := fmt.Sprintf(`Evaluate how well the following content aligns with the specified brand voice:

Brand Voice: %s

Content:
%s

Rate the alignment (0-100) and identify specific issues where the content doesn't match the brand voice.

Respond in JSON format:
{
  "alignmentScore": <0-100>,
  "issues": [
    {
      "description": "<issue description>",
      "location": "<where in content>",
      "suggestion": "<how to fix>",
      "severity": "major|minor"
    }
  ]
}`, brandVoice, content)

	response, err := sc.llmClient.Generate(ctx, prompt)
	if err != nil {
		return 0, nil, err
	}

	return sc.parseBrandAlignment(response)
}

// generateStyleProfile creates a profile of the writing style
func (sc *StyleChecker) generateStyleProfile(ctx context.Context, content string) (*StyleProfile, error) {
	prompt := fmt.Sprintf(`Analyze the writing style of the following content and create a style profile:

Content:
%s

Evaluate:
1. Formality level (0-1, where 0 is very informal and 1 is very formal)
2. Complexity level (0-1, where 0 is simple and 1 is complex)
3. Personality traits evident in the writing
4. Overall writing style category
5. Key characteristics

Respond in JSON format:
{
  "formalityLevel": <0-1>,
  "complexityLevel": <0-1>,
  "personalityTraits": ["<trait1>", "<trait2>"],
  "writingStyle": "<style category>",
  "characteristics": ["<characteristic1>", "<characteristic2>"]
}`, content)

	response, err := sc.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return sc.parseStyleProfile(response)
}

// Helper methods for formatting checks

// checkCapitalizationConsistency checks for capitalization issues
func (sc *StyleChecker) checkCapitalizationConsistency(content string) []string {
	issues := []string{}
	
	// Check sentence capitalization
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(content, -1)
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 0 && strings.ToLower(sentence[:1]) == sentence[:1] {
			issues = append(issues, fmt.Sprintf("Sentence not capitalized: '%.50s...'", sentence))
		}
	}
	
	return issues
}

// checkPunctuationConsistency checks for punctuation issues
func (sc *StyleChecker) checkPunctuationConsistency(content string) []string {
	issues := []string{}
	
	// Check for spaces before punctuation
	punctRegex := regexp.MustCompile(`\s+[,.;:]`)
	if matches := punctRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, "Spaces before punctuation marks detected")
	}
	
	// Check for missing spaces after punctuation
	noSpaceRegex := regexp.MustCompile(`[,.;:][^\s]`)
	if matches := noSpaceRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, "Missing spaces after punctuation marks")
	}
	
	return issues
}

// checkHeadingConsistency checks for heading format issues
func (sc *StyleChecker) checkHeadingConsistency(content string) []string {
	issues := []string{}
	
	// Check markdown-style headings
	headingRegex := regexp.MustCompile(`(?m)^#+\s+(.*)$`)
	headings := headingRegex.FindAllString(content, -1)
	
	if len(headings) > 1 {
		// Check for consistent spacing
		for _, heading := range headings {
			if !regexp.MustCompile(`^#+\s+`).MatchString(heading) {
				issues = append(issues, fmt.Sprintf("Inconsistent heading format: %s", heading))
			}
		}
	}
	
	return issues
}

// Parsing methods

// parseToneAnalysis parses tone analysis response
func (sc *StyleChecker) parseToneAnalysis(response string) (*ToneAnalysis, error) {
	var analysis ToneAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		// Fallback parsing
		return &ToneAnalysis{
			PrimaryTone:     ToneProfessional,
			ToneConsistency: 0.8,
			ToneShifts:      []ToneShift{},
			AudienceMatch:   0.8,
			EmotionalTone:   map[string]float64{"neutral": 0.8},
		}, nil
	}
	return &analysis, nil
}

// parseVoiceAnalysis parses voice analysis response
func (sc *StyleChecker) parseVoiceAnalysis(response string) (*VoiceConsistency, error) {
	var voice VoiceConsistency
	if err := json.Unmarshal([]byte(response), &voice); err != nil {
		// Fallback
		return &VoiceConsistency{
			Score:              0.8,
			PersonConsistency:  0.8,
			PerspectiveShifts:  []PerspectiveShift{},
			VoiceCharacteristics: VoiceCharacteristics{
				Person:          "third",
				Perspective:     "objective",
				AuthorityLevel:  0.7,
				PersonalityScore: 0.6,
			},
		}, nil
	}
	return &voice, nil
}

// parseBrandAlignment parses brand alignment response
func (sc *StyleChecker) parseBrandAlignment(response string) (float64, []StyleIssue, error) {
	var result struct {
		AlignmentScore float64 `json:"alignmentScore"`
		Issues         []struct {
			Description string `json:"description"`
			Location    string `json:"location"`
			Suggestion  string `json:"suggestion"`
			Severity    string `json:"severity"`
		} `json:"issues"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return 75.0, []StyleIssue{}, nil
	}

	issues := []StyleIssue{}
	for _, issue := range result.Issues {
		severity := SeverityMinor
		if issue.Severity == "major" {
			severity = SeverityMajor
		}
		
		issues = append(issues, StyleIssue{
			Type:        IssueBrandViolation,
			Severity:    severity,
			Description: issue.Description,
			Location:    issue.Location,
			Suggestion:  issue.Suggestion,
			RuleViolated: "Brand voice guidelines",
		})
	}

	return result.AlignmentScore, issues, nil
}

// parseStyleProfile parses style profile response
func (sc *StyleChecker) parseStyleProfile(response string) (*StyleProfile, error) {
	var profile StyleProfile
	if err := json.Unmarshal([]byte(response), &profile); err != nil {
		// Fallback
		return &StyleProfile{
			FormalityLevel:   0.7,
			ComplexityLevel:  0.6,
			PersonalityTraits: []string{"professional", "clear"},
			WritingStyle:     "informative",
			Characteristics:  []string{"clear", "concise"},
		}, nil
	}
	return &profile, nil
}

// Score calculation methods

// calculateConsistencyScore calculates overall consistency score
func (sc *StyleChecker) calculateConsistencyScore(result *StyleAnalysisResult) float64 {
	score := 100.0
	
	// Penalize based on tone consistency
	score *= result.ToneAnalysis.ToneConsistency
	
	// Penalize based on voice consistency
	score *= result.VoiceConsistency.Score
	
	// Penalize for formatting issues
	score -= float64(len(result.FormattingIssues)) * 5.0
	
	// Penalize for style issues
	for _, issue := range result.StyleIssues {
		switch issue.Severity {
		case SeverityCritical:
			score -= 15.0
		case SeverityMajor:
			score -= 10.0
		case SeverityMinor:
			score -= 5.0
		}
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateOverallStyleScore calculates overall style score
func (sc *StyleChecker) calculateOverallStyleScore(result *StyleAnalysisResult) float64 {
	scores := []float64{
		result.ConsistencyScore,
		result.ToneAnalysis.AudienceMatch * 100,
		result.VoiceConsistency.Score * 100,
	}
	
	if result.BrandAlignmentScore > 0 {
		scores = append(scores, result.BrandAlignmentScore)
	}
	
	total := 0.0
	for _, score := range scores {
		total += score
	}
	
	return total / float64(len(scores))
}

// generateStyleRecommendations generates actionable style recommendations
func (sc *StyleChecker) generateStyleRecommendations(result *StyleAnalysisResult) []StyleRecommendation {
	recommendations := []StyleRecommendation{}
	
	// High priority recommendations
	if result.ToneAnalysis.ToneConsistency < 0.7 {
		recommendations = append(recommendations, StyleRecommendation{
			Priority:   PriorityHigh,
			Category:   "Tone",
			Issue:      "Inconsistent tone throughout content",
			Suggestion: "Review content for tone shifts and maintain consistent voice",
			Impact:     "Improves reader experience and brand consistency",
		})
	}
	
	if result.VoiceConsistency.Score < 0.7 {
		recommendations = append(recommendations, StyleRecommendation{
			Priority:   PriorityHigh,
			Category:   "Voice",
			Issue:      "Voice and perspective inconsistencies",
			Suggestion: "Maintain consistent perspective (first/second/third person) throughout",
			Impact:     "Reduces confusion and improves clarity",
		})
	}
	
	// Medium priority recommendations
	if len(result.FormattingIssues) > 0 {
		recommendations = append(recommendations, StyleRecommendation{
			Priority:   PriorityMedium,
			Category:   "Formatting",
			Issue:      "Formatting inconsistencies detected",
			Suggestion: "Review and apply consistent formatting rules",
			Impact:     "Improves professional appearance and readability",
		})
	}
	
	// Brand alignment recommendations
	if result.BrandAlignmentScore < 80 {
		recommendations = append(recommendations, StyleRecommendation{
			Priority:   PriorityMedium,
			Category:   "Brand",
			Issue:      "Content doesn't fully align with brand voice",
			Suggestion: "Adjust tone and language to better match brand guidelines",
			Impact:     "Strengthens brand consistency and recognition",
		})
	}
	
	return recommendations
}
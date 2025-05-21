package content_creation

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// PlagiarismDetector handles comprehensive plagiarism detection
type PlagiarismDetector struct {
	plagiarismAPI PlagiarismAPI
	llmClient     LLMClient
	fingerprinter *ContentFingerprinter
}

// NewPlagiarismDetector creates a new plagiarism detector
func NewPlagiarismDetector(plagiarismAPI PlagiarismAPI, llmClient LLMClient) *PlagiarismDetector {
	return &PlagiarismDetector{
		plagiarismAPI: plagiarismAPI,
		llmClient:     llmClient,
		fingerprinter: NewContentFingerprinter(),
	}
}

// PlagiarismCheckRequest contains parameters for plagiarism detection
type PlagiarismCheckRequest struct {
	Content     string
	ContentType entities.ContentType
	CheckWeb    bool
	CheckDatabase bool
	Sensitivity float64 // 0.0 to 1.0, higher = more sensitive
}

// PlagiarismResult contains comprehensive plagiarism detection results
type PlagiarismResult struct {
	OriginalityScore    float64             `json:"originalityScore"`
	OverallRisk         RiskLevel           `json:"overallRisk"`
	TotalMatches        int                 `json:"totalMatches"`
	HighRiskMatches     int                 `json:"highRiskMatches"`
	MediumRiskMatches   int                 `json:"mediumRiskMatches"`
	LowRiskMatches      int                 `json:"lowRiskMatches"`
	Matches             []PlagiarismMatch   `json:"matches"`
	Patterns            []SuspiciousPattern `json:"patterns"`
	Fingerprint         ContentFingerprint  `json:"fingerprint"`
	Recommendations     []string            `json:"recommendations"`
	ProcessingTime      time.Duration       `json:"processingTime"`
	ConfidenceLevel     float64             `json:"confidenceLevel"`
}

// PlagiarismMatch represents a detected match with external content
type PlagiarismMatch struct {
	ID              string    `json:"id"`
	MatchedText     string    `json:"matchedText"`
	SourceText      string    `json:"sourceText"`
	Source          Source    `json:"source"`
	SimilarityScore float64   `json:"similarityScore"`
	RiskLevel       RiskLevel `json:"riskLevel"`
	Location        Location  `json:"location"`
	Type            MatchType `json:"type"`
	Context         string    `json:"context"`
}

// SuspiciousPattern represents patterns that might indicate plagiarism
type SuspiciousPattern struct {
	Type        PatternType `json:"type"`
	Description string      `json:"description"`
	Instances   []string    `json:"instances"`
	RiskLevel   RiskLevel   `json:"riskLevel"`
}

// Source represents a source where plagiarized content was found
type Source struct {
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Author      string     `json:"author,omitempty"`
	Domain      string     `json:"domain"`
	PublishDate string     `json:"publishDate,omitempty"`
	Type        SourceType `json:"type"`
}

// Location represents where in the content the match was found
type Location struct {
	StartChar int `json:"startChar"`
	EndChar   int `json:"endChar"`
	Paragraph int `json:"paragraph"`
	Sentence  int `json:"sentence"`
}

// ContentFingerprint represents a unique fingerprint of content
type ContentFingerprint struct {
	Hash            string            `json:"hash"`
	WordFrequency   map[string]int    `json:"wordFrequency"`
	PhraseHashes    []string          `json:"phraseHashes"`
	StyleMetrics    StyleMetrics      `json:"styleMetrics"`
	SemanticVectors []float64         `json:"semanticVectors"`
}

// StyleMetrics captures writing style characteristics
type StyleMetrics struct {
	AverageSentenceLength float64 `json:"averageSentenceLength"`
	VocabularyRichness    float64 `json:"vocabularyRichness"`
	ReadabilityScore      float64 `json:"readabilityScore"`
	PunctuationDensity    float64 `json:"punctuationDensity"`
	PassiveVoiceRatio     float64 `json:"passiveVoiceRatio"`
}

// RiskLevel indicates the risk level of plagiarism
type RiskLevel string

const (
	RiskHigh   RiskLevel = "high"
	RiskMedium RiskLevel = "medium"
	RiskLow    RiskLevel = "low"
	RiskNone   RiskLevel = "none"
)

// MatchType categorizes different types of matches
type MatchType string

const (
	MatchExact      MatchType = "exact"
	MatchNearExact  MatchType = "near_exact"
	MatchParaphrase MatchType = "paraphrase"
	MatchStructural MatchType = "structural"
	MatchIdea       MatchType = "idea"
)

// PatternType categorizes suspicious patterns
type PatternType string

const (
	PatternRepetitiveStructure PatternType = "repetitive_structure"
	PatternUnusualPhrasing     PatternType = "unusual_phrasing"
	PatternStyleInconsistency  PatternType = "style_inconsistency"
	PatternSuspiciousQuoting   PatternType = "suspicious_quoting"
)

// CheckPlagiarism performs comprehensive plagiarism detection
func (p *PlagiarismDetector) CheckPlagiarism(ctx context.Context, request PlagiarismCheckRequest) (*PlagiarismResult, error) {
	startTime := time.Now()
	
	result := &PlagiarismResult{
		Matches:         []PlagiarismMatch{},
		Patterns:        []SuspiciousPattern{},
		Recommendations: []string{},
	}

	// 1. Generate content fingerprint
	fingerprint, err := p.fingerprinter.GenerateFingerprint(request.Content)
	if err != nil {
		return nil, fmt.Errorf("fingerprint generation failed: %w", err)
	}
	result.Fingerprint = *fingerprint

	// 2. Check against external sources if enabled
	if request.CheckWeb {
		webMatches, err := p.checkAgainstWeb(ctx, request.Content)
		if err == nil {
			result.Matches = append(result.Matches, webMatches...)
		}
	}

	// 3. Check against database if enabled
	if request.CheckDatabase {
		dbMatches, err := p.checkAgainstDatabase(ctx, fingerprint)
		if err == nil {
			result.Matches = append(result.Matches, dbMatches...)
		}
	}

	// 4. Detect suspicious patterns
	patterns, err := p.detectSuspiciousPatterns(ctx, request.Content)
	if err == nil {
		result.Patterns = patterns
	}

	// 5. Perform semantic similarity analysis
	semanticMatches, err := p.checkSemanticSimilarity(ctx, request.Content)
	if err == nil {
		result.Matches = append(result.Matches, semanticMatches...)
	}

	// 6. Calculate metrics
	result.TotalMatches = len(result.Matches)
	p.categorizeMatches(result)
	result.OriginalityScore = p.calculateOriginalityScore(result)
	result.OverallRisk = p.determineOverallRisk(result)
	result.ConfidenceLevel = p.calculateConfidence(result)
	result.Recommendations = p.generateRecommendations(result)
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// checkAgainstWeb checks content against web sources
func (p *PlagiarismDetector) checkAgainstWeb(ctx context.Context, content string) ([]PlagiarismMatch, error) {
	// Use external plagiarism API
	if p.plagiarismAPI != nil {
		_, details, err := p.plagiarismAPI.CheckPlagiarism(ctx, content)
		if err != nil {
			return nil, err
		}

		matches := []PlagiarismMatch{}
		for _, detail := range details {
			match := PlagiarismMatch{
				ID:              generateMatchID(detail.Fragment, detail.Source),
				MatchedText:     detail.Fragment,
				SimilarityScore: detail.Percentage / 100.0,
				Source: Source{
					URL:    detail.Source,
					Domain: extractDomain(detail.Source),
					Type:   SourceNews, // Default
				},
				RiskLevel: p.determineRiskLevel(detail.Percentage / 100.0),
				Type:      MatchExact,
			}
			matches = append(matches, match)
		}

		return matches, nil
	}

	// Fallback: simulate basic web checking
	return p.simulateWebCheck(content), nil
}

// checkAgainstDatabase checks content against internal database
func (p *PlagiarismDetector) checkAgainstDatabase(ctx context.Context, fingerprint *ContentFingerprint) ([]PlagiarismMatch, error) {
	// In a real implementation, this would check against a database of previously processed content
	// For now, we'll return empty results
	return []PlagiarismMatch{}, nil
}

// detectSuspiciousPatterns identifies patterns that might indicate plagiarism
func (p *PlagiarismDetector) detectSuspiciousPatterns(ctx context.Context, content string) ([]SuspiciousPattern, error) {
	prompt := fmt.Sprintf(`Analyze the following content for patterns that might indicate plagiarism or non-original writing:

1. Repetitive sentence structures
2. Unusual phrasing or word choices
3. Inconsistent writing style
4. Suspicious use of quotes or citations
5. Abrupt topic changes
6. Inconsistent voice or tone

Content to analyze:
%s

Identify suspicious patterns and explain why they might indicate non-original content.

Respond in JSON format:
{
  "patterns": [
    {
      "type": "repetitive_structure|unusual_phrasing|style_inconsistency|suspicious_quoting",
      "description": "<description of the pattern>",
      "instances": ["<example 1>", "<example 2>"],
      "riskLevel": "high|medium|low"
    }
  ]
}`, content)

	response, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return p.parsePatternsResponse(response)
}

// checkSemanticSimilarity performs semantic similarity analysis
func (p *PlagiarismDetector) checkSemanticSimilarity(ctx context.Context, content string) ([]PlagiarismMatch, error) {
	// Break content into segments
	segments := p.segmentContent(content)
	matches := []PlagiarismMatch{}

	for _, segment := range segments {
		// In a real implementation, this would use semantic similarity models
		// For now, we'll simulate with basic pattern matching
		if p.isSemanticallySuspicious(segment) {
			match := PlagiarismMatch{
				ID:              generateMatchID(segment, "semantic_analysis"),
				MatchedText:     segment,
				SimilarityScore: 0.7,
				Source: Source{
					URL:   "semantic_analysis",
					Title: "Semantic Analysis",
					Type:  SourceExpert,
				},
				RiskLevel: RiskMedium,
				Type:      MatchParaphrase,
			}
			matches = append(matches, match)
		}
	}

	return matches, nil
}

// Helper methods

// simulateWebCheck simulates basic web plagiarism checking
func (p *PlagiarismDetector) simulateWebCheck(content string) []PlagiarismMatch {
	// This is a placeholder implementation
	// In a real system, this would perform actual web searches
	matches := []PlagiarismMatch{}

	// Look for common phrases that might be plagiarized
	commonPhrases := []string{
		"in conclusion",
		"it is important to note",
		"according to recent studies",
		"research has shown",
	}

	for _, phrase := range commonPhrases {
		if strings.Contains(strings.ToLower(content), phrase) {
			match := PlagiarismMatch{
				ID:              generateMatchID(phrase, "common_phrase"),
				MatchedText:     phrase,
				SimilarityScore: 0.3,
				Source: Source{
					URL:   "common_phrases",
					Title: "Common Academic Phrases",
					Type:  SourceExpert,
				},
				RiskLevel: RiskLow,
				Type:      MatchNearExact,
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// parsePatternsResponse parses the patterns detection response
func (p *PlagiarismDetector) parsePatternsResponse(response string) ([]SuspiciousPattern, error) {
	var result struct {
		Patterns []SuspiciousPattern `json:"patterns"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback parsing
		return []SuspiciousPattern{}, nil
	}

	return result.Patterns, nil
}

// segmentContent breaks content into analyzable segments
func (p *PlagiarismDetector) segmentContent(content string) []string {
	// Split by sentences for basic segmentation
	sentences := strings.Split(content, ".")
	segments := []string{}

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 { // Only analyze substantial sentences
			segments = append(segments, sentence)
		}
	}

	return segments
}

// isSemanticallySuspicious checks if a segment seems semantically suspicious
func (p *PlagiarismDetector) isSemanticallySuspicious(segment string) bool {
	// Simplified heuristics for semantic suspicion
	suspiciousPatterns := []string{
		"it is widely known",
		"experts agree",
		"studies have shown",
		"it is a fact that",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(segment), pattern) {
			return true
		}
	}

	return false
}

// categorizeMatches categorizes matches by risk level
func (p *PlagiarismDetector) categorizeMatches(result *PlagiarismResult) {
	for _, match := range result.Matches {
		switch match.RiskLevel {
		case RiskHigh:
			result.HighRiskMatches++
		case RiskMedium:
			result.MediumRiskMatches++
		case RiskLow:
			result.LowRiskMatches++
		}
	}
}

// calculateOriginalityScore calculates the overall originality score
func (p *PlagiarismDetector) calculateOriginalityScore(result *PlagiarismResult) float64 {
	if result.TotalMatches == 0 {
		return 100.0
	}

	// Calculate penalty based on match risk levels
	penalty := float64(result.HighRiskMatches)*20.0 +
		float64(result.MediumRiskMatches)*10.0 +
		float64(result.LowRiskMatches)*5.0

	// Calculate average similarity penalty
	totalSimilarity := 0.0
	for _, match := range result.Matches {
		totalSimilarity += match.SimilarityScore
	}
	avgSimilarity := totalSimilarity / float64(result.TotalMatches)
	similarityPenalty := avgSimilarity * 30.0

	// Calculate final score
	score := 100.0 - penalty - similarityPenalty

	if score < 0 {
		score = 0
	}

	return score
}

// determineOverallRisk determines the overall plagiarism risk
func (p *PlagiarismDetector) determineOverallRisk(result *PlagiarismResult) RiskLevel {
	if result.HighRiskMatches > 0 || result.OriginalityScore < 60 {
		return RiskHigh
	} else if result.MediumRiskMatches > 2 || result.OriginalityScore < 80 {
		return RiskMedium
	} else if result.LowRiskMatches > 5 || result.OriginalityScore < 95 {
		return RiskLow
	}
	return RiskNone
}

// determineRiskLevel determines risk level based on similarity score
func (p *PlagiarismDetector) determineRiskLevel(similarity float64) RiskLevel {
	if similarity >= 0.8 {
		return RiskHigh
	} else if similarity >= 0.6 {
		return RiskMedium
	} else if similarity >= 0.3 {
		return RiskLow
	}
	return RiskNone
}

// calculateConfidence calculates confidence in plagiarism detection
func (p *PlagiarismDetector) calculateConfidence(result *PlagiarismResult) float64 {
	// Base confidence
	confidence := 0.8

	// Increase confidence with more matches
	if result.TotalMatches > 5 {
		confidence += 0.1
	}

	// Decrease confidence if only low-risk matches
	if result.HighRiskMatches == 0 && result.MediumRiskMatches == 0 {
		confidence -= 0.2
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.5 {
		confidence = 0.5
	}

	return confidence
}

// generateRecommendations generates recommendations based on results
func (p *PlagiarismDetector) generateRecommendations(result *PlagiarismResult) []string {
	recommendations := []string{}

	if result.OriginalityScore < 80 {
		recommendations = append(recommendations, "Review and rewrite sections with high similarity scores")
	}

	if result.HighRiskMatches > 0 {
		recommendations = append(recommendations, "Address high-risk plagiarism matches immediately")
	}

	if len(result.Patterns) > 0 {
		recommendations = append(recommendations, "Review content for suspicious writing patterns")
	}

	if result.OriginalityScore < 95 {
		recommendations = append(recommendations, "Add proper citations and attributions")
		recommendations = append(recommendations, "Paraphrase similar content more effectively")
	}

	return recommendations
}

// generateMatchID generates a unique ID for a match
func generateMatchID(text, source string) string {
	hasher := md5.New()
	hasher.Write([]byte(text + source))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:8]
}

// ContentFingerprinter generates unique fingerprints for content
type ContentFingerprinter struct{}

// NewContentFingerprinter creates a new content fingerprinter
func NewContentFingerprinter() *ContentFingerprinter {
	return &ContentFingerprinter{}
}

// GenerateFingerprint creates a unique fingerprint for content
func (cf *ContentFingerprinter) GenerateFingerprint(content string) (*ContentFingerprint, error) {
	// Generate hash
	hasher := md5.New()
	hasher.Write([]byte(content))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Calculate word frequency
	wordFreq := cf.calculateWordFrequency(content)

	// Generate phrase hashes
	phraseHashes := cf.generatePhraseHashes(content)

	// Calculate style metrics
	styleMetrics := cf.calculateStyleMetrics(content)

	return &ContentFingerprint{
		Hash:            hash,
		WordFrequency:   wordFreq,
		PhraseHashes:    phraseHashes,
		StyleMetrics:    styleMetrics,
		SemanticVectors: []float64{}, // Placeholder for semantic vectors
	}, nil
}

// calculateWordFrequency calculates word frequency distribution
func (cf *ContentFingerprinter) calculateWordFrequency(content string) map[string]int {
	words := strings.Fields(strings.ToLower(content))
	freq := make(map[string]int)

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]")
		if len(word) > 2 && !isStopWord(word) {
			freq[word]++
		}
	}

	return freq
}

// generatePhraseHashes generates hashes for common phrases
func (cf *ContentFingerprinter) generatePhraseHashes(content string) []string {
	sentences := strings.Split(content, ".")
	hashes := []string{}

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 {
			hasher := md5.New()
			hasher.Write([]byte(strings.ToLower(sentence)))
			hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

// calculateStyleMetrics calculates writing style metrics
func (cf *ContentFingerprinter) calculateStyleMetrics(content string) StyleMetrics {
	sentences := strings.Split(content, ".")
	words := strings.Fields(content)
	
	// Average sentence length
	avgSentenceLength := 0.0
	if len(sentences) > 0 {
		avgSentenceLength = float64(len(words)) / float64(len(sentences))
	}

	// Vocabulary richness (unique words / total words)
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[strings.ToLower(word)] = true
	}
	vocabRichness := 0.0
	if len(words) > 0 {
		vocabRichness = float64(len(uniqueWords)) / float64(len(words))
	}

	// Punctuation density
	punctuation := 0
	for _, char := range content {
		if strings.ContainsRune(".,!?;:", char) {
			punctuation++
		}
	}
	punctuationDensity := float64(punctuation) / float64(len(content))

	return StyleMetrics{
		AverageSentenceLength: avgSentenceLength,
		VocabularyRichness:    vocabRichness,
		ReadabilityScore:      70.0, // Placeholder
		PunctuationDensity:    punctuationDensity,
		PassiveVoiceRatio:     0.2,  // Placeholder
	}
}
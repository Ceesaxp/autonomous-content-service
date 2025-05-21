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

// FactChecker handles fact-checking capabilities against reliable sources
type FactChecker struct {
	llmClient     LLMClient
	searchService SearchService
	sourcesDB     FactSourceDatabase
}

// NewFactChecker creates a new fact checker
func NewFactChecker(llmClient LLMClient, searchService SearchService) *FactChecker {
	return &FactChecker{
		llmClient:     llmClient,
		searchService: searchService,
		sourcesDB:     NewFactSourceDatabase(),
	}
}

// FactCheckRequest contains parameters for fact checking
type FactCheckRequest struct {
	Content     string
	ContentType entities.ContentType
	Sources     []string
	Domain      string
}

// FactCheckResult contains comprehensive fact-checking results
type FactCheckResult struct {
	OverallScore    float64       `json:"overallScore"`
	ErrorCount      int           `json:"errorCount"`
	WarningCount    int           `json:"warningCount"`
	VerifiedClaims  int           `json:"verifiedClaims"`
	TotalClaims     int           `json:"totalClaims"`
	FactualErrors   []FactualIssue `json:"factualErrors"`
	Warnings        []FactualIssue `json:"warnings"`
	VerifiedFacts   []VerifiedFact `json:"verifiedFacts"`
	Sources         []FactSource   `json:"sources"`
	ConfidenceLevel float64        `json:"confidenceLevel"`
	ProcessingTime  time.Duration  `json:"processingTime"`
}

// FactualIssue represents a factual error or warning
type FactualIssue struct {
	ID          string      `json:"id"`
	Type        IssueType   `json:"type"`
	Severity    IssueSeverity `json:"severity"`
	Claim       string      `json:"claim"`
	Location    string      `json:"location"`
	Issue       string      `json:"issue"`
	Correction  string      `json:"correction"`
	Evidence    []Evidence  `json:"evidence"`
	Confidence  float64     `json:"confidence"`
}

// VerifiedFact represents a successfully verified fact
type VerifiedFact struct {
	Claim       string     `json:"claim"`
	Location    string     `json:"location"`
	Sources     []FactSource `json:"sources"`
	Confidence  float64    `json:"confidence"`
}

// Evidence represents supporting evidence for a fact check
type Evidence struct {
	Source      FactSource `json:"source"`
	Quote       string     `json:"quote"`
	URL         string     `json:"url"`
	Relevance   float64    `json:"relevance"`
}

// FactSource represents a reliable source for fact checking
type FactSource struct {
	Name           string      `json:"name"`
	URL            string      `json:"url"`
	Type           SourceType  `json:"type"`
	Credibility    float64     `json:"credibility"`
	Domain         string      `json:"domain"`
	LastVerified   time.Time   `json:"lastVerified"`
}

// IssueType categorizes factual issues
type IssueType string

const (
	IssueFactualError     IssueType = "factual_error"
	IssueOutdatedInfo     IssueType = "outdated_info"
	IssueUnverifiableClaim IssueType = "unverifiable_claim"
	IssueMissingSource    IssueType = "missing_source"
	IssueContradiction    IssueType = "contradiction"
)

// IssueSeverity indicates the severity of factual issues
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "error"
	SeverityWarning IssueSeverity = "warning"
	SeverityNotice  IssueSeverity = "notice"
)

// SourceType categorizes different types of sources
type SourceType string

const (
	SourceAcademic     SourceType = "academic"
	SourceGovernment   SourceType = "government"
	SourceNews         SourceType = "news"
	SourceEncyclopedia SourceType = "encyclopedia"
	SourceOfficial     SourceType = "official"
	SourceExpert       SourceType = "expert"
)

// CheckFacts performs comprehensive fact checking
func (f *FactChecker) CheckFacts(ctx context.Context, request FactCheckRequest) (*FactCheckResult, error) {
	startTime := time.Now()
	
	result := &FactCheckResult{
		FactualErrors:  []FactualIssue{},
		Warnings:       []FactualIssue{},
		VerifiedFacts:  []VerifiedFact{},
		Sources:        []FactSource{},
	}

	// 1. Extract claims from content
	claims, err := f.extractClaims(ctx, request.Content, request.ContentType)
	if err != nil {
		return nil, fmt.Errorf("claim extraction failed: %w", err)
	}
	result.TotalClaims = len(claims)

	// 2. Check each claim
	for _, claim := range claims {
		claimResult, err := f.checkClaim(ctx, claim, request.Domain)
		if err != nil {
			// Log error but continue with other claims
			continue
		}

		// Process claim result
		f.processClaimResult(claimResult, result)
	}

	// 3. Cross-reference against reliable sources
	if len(request.Sources) > 0 {
		sourceResults, err := f.checkAgainstSources(ctx, claims, request.Sources)
		if err == nil {
			f.incorporateSourceResults(sourceResults, result)
		}
	}

	// 4. Detect contradictions within content
	contradictions, err := f.detectContradictions(ctx, request.Content)
	if err == nil {
		for _, contradiction := range contradictions {
			result.FactualErrors = append(result.FactualErrors, contradiction)
		}
	}

	// 5. Calculate metrics
	result.ErrorCount = len(result.FactualErrors)
	result.WarningCount = len(result.Warnings)
	result.VerifiedClaims = len(result.VerifiedFacts)
	result.OverallScore = f.calculateFactCheckScore(result)
	result.ConfidenceLevel = f.calculateConfidence(result)
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// extractClaims extracts factual claims from content
func (f *FactChecker) extractClaims(ctx context.Context, content string, contentType entities.ContentType) ([]FactualClaim, error) {
	prompt := fmt.Sprintf(`Extract all factual claims from the following %s content. 
Focus on statements that can be verified with external sources, including:
- Statistical data and numbers
- Historical facts and dates
- Scientific claims
- Quotes and attributions
- Product specifications
- Company information

For each claim, identify its location in the text and classify its type.

Content:
%s

Respond in JSON format:
{
  "claims": [
    {
      "text": "<exact claim text>",
      "location": "<position in content>",
      "type": "statistic|historical|scientific|quote|specification|other",
      "verifiable": true|false
    }
  ]
}`, contentType, content)

	response, err := f.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return f.parseClaimsResponse(response)
}

// checkClaim verifies a specific claim
func (f *FactChecker) checkClaim(ctx context.Context, claim FactualClaim, domain string) (*ClaimCheckResult, error) {
	// Search for supporting/contradicting evidence
	searchQuery := f.buildSearchQuery(claim)
	searchResults, err := f.searchService.Search(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	// Analyze search results
	evidence := []Evidence{}
	for _, result := range searchResults {
		// Get credibility score for source
		credibility := f.sourcesDB.GetSourceCredibility(result.URL)
		
		if credibility >= 0.7 { // Only use highly credible sources
			evidence = append(evidence, Evidence{
				Source: FactSource{
					Name:        result.Title,
					URL:         result.URL,
					Credibility: credibility,
					Domain:      extractDomain(result.URL),
				},
				Quote:     result.Snippet,
				URL:       result.URL,
				Relevance: float64(result.Relevance) / 10.0,
			})
		}
	}

	// Use LLM to assess claim against evidence
	return f.assessClaimWithEvidence(ctx, claim, evidence)
}

// assessClaimWithEvidence uses LLM to verify claim against evidence
func (f *FactChecker) assessClaimWithEvidence(ctx context.Context, claim FactualClaim, evidence []Evidence) (*ClaimCheckResult, error) {
	evidenceText := ""
	for i, ev := range evidence {
		evidenceText += fmt.Sprintf("\n%d. Source: %s (Credibility: %.2f)\n   Quote: %s\n   URL: %s\n",
			i+1, ev.Source.Name, ev.Source.Credibility, ev.Quote, ev.URL)
	}

	prompt := fmt.Sprintf(`Verify the following claim against the provided evidence:

CLAIM: %s

EVIDENCE:
%s

Analyze whether the claim is:
1. VERIFIED: Supported by credible evidence
2. CONTRADICTED: Contradicted by credible evidence  
3. UNVERIFIABLE: Cannot be verified with available evidence
4. OUTDATED: Once true but now outdated

Provide your assessment with confidence level and explanation.

Respond in JSON format:
{
  "status": "verified|contradicted|unverifiable|outdated",
  "confidence": <0-1>,
  "explanation": "<detailed explanation>",
  "supporting_sources": ["<source1>", "<source2>"],
  "contradicting_sources": ["<source1>", "<source2>"],
  "correction": "<correct information if claim is wrong>",
  "severity": "error|warning|notice"
}`, claim.Text, evidenceText)

	response, err := f.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return f.parseClaimCheckResponse(response, claim, evidence)
}

// checkAgainstSources checks claims against specific provided sources
func (f *FactChecker) checkAgainstSources(ctx context.Context, claims []FactualClaim, sources []string) ([]*ClaimCheckResult, error) {
	results := []*ClaimCheckResult{}

	for _, source := range sources {
		// Fetch source content
		content, err := f.searchService.FetchContent(ctx, source)
		if err != nil {
			continue
		}

		// Check each claim against this source
		for _, claim := range claims {
			result, err := f.checkClaimAgainstSource(ctx, claim, content, source)
			if err == nil {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// checkClaimAgainstSource checks a claim against a specific source
func (f *FactChecker) checkClaimAgainstSource(ctx context.Context, claim FactualClaim, sourceContent, sourceURL string) (*ClaimCheckResult, error) {
	prompt := fmt.Sprintf(`Check if the following claim is supported, contradicted, or not mentioned in the source content:

CLAIM: %s

SOURCE CONTENT:
%s

Analyze the relationship and provide assessment.

Respond in JSON format:
{
  "status": "verified|contradicted|not_mentioned",
  "confidence": <0-1>,
  "explanation": "<explanation>",
  "relevant_quote": "<relevant quote from source if any>",
  "correction": "<correction if needed>"
}`, claim.Text, sourceContent)

	response, err := f.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return f.parseSourceCheckResponse(response, claim, sourceURL)
}

// detectContradictions finds contradictions within the content itself
func (f *FactChecker) detectContradictions(ctx context.Context, content string) ([]FactualIssue, error) {
	prompt := fmt.Sprintf(`Analyze the following content for internal contradictions - statements that contradict each other within the same text.

Content:
%s

Identify any contradictory statements and explain the contradictions.

Respond in JSON format:
{
  "contradictions": [
    {
      "statement1": "<first contradictory statement>",
      "statement2": "<second contradictory statement>",
      "location1": "<location of first statement>",
      "location2": "<location of second statement>",
      "explanation": "<explanation of contradiction>",
      "severity": "error|warning|notice"
    }
  ]
}`, content)

	response, err := f.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return f.parseContradictionsResponse(response)
}

// Helper functions and parsing methods

// FactualClaim represents a claim that can be fact-checked
type FactualClaim struct {
	Text       string `json:"text"`
	Location   string `json:"location"`
	Type       string `json:"type"`
	Verifiable bool   `json:"verifiable"`
}

// ClaimCheckResult contains the result of checking a single claim
type ClaimCheckResult struct {
	Claim      FactualClaim
	Status     string
	Confidence float64
	Evidence   []Evidence
	Issue      *FactualIssue
	Verified   *VerifiedFact
}

// parseClaimsResponse parses the claims extraction response
func (f *FactChecker) parseClaimsResponse(response string) ([]FactualClaim, error) {
	var result struct {
		Claims []FactualClaim `json:"claims"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback parsing
		return f.parseClaimsFallback(response), nil
	}

	return result.Claims, nil
}

// parseClaimsFallback provides fallback parsing for claims
func (f *FactChecker) parseClaimsFallback(response string) []FactualClaim {
	claims := []FactualClaim{}
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") && len(line) > 10 {
			claims = append(claims, FactualClaim{
				Text:       line,
				Location:   "Unknown",
				Type:       "other",
				Verifiable: true,
			})
		}
	}
	
	return claims
}

// parseClaimCheckResponse parses claim check results
func (f *FactChecker) parseClaimCheckResponse(response string, claim FactualClaim, evidence []Evidence) (*ClaimCheckResult, error) {
	// Simplified parsing - in a real implementation, would use proper JSON parsing
	result := &ClaimCheckResult{
		Claim:    claim,
		Evidence: evidence,
	}

	// Extract status
	if strings.Contains(strings.ToLower(response), "verified") {
		result.Status = "verified"
	} else if strings.Contains(strings.ToLower(response), "contradicted") {
		result.Status = "contradicted"
	} else {
		result.Status = "unverifiable"
	}

	// Set confidence (simplified)
	result.Confidence = 0.8

	return result, nil
}

// parseSourceCheckResponse parses source-specific check results
func (f *FactChecker) parseSourceCheckResponse(response string, claim FactualClaim, sourceURL string) (*ClaimCheckResult, error) {
	// Simplified implementation
	return &ClaimCheckResult{
		Claim:      claim,
		Status:     "not_mentioned",
		Confidence: 0.7,
	}, nil
}

// parseContradictionsResponse parses contradiction detection results
func (f *FactChecker) parseContradictionsResponse(response string) ([]FactualIssue, error) {
	// Simplified implementation
	return []FactualIssue{}, nil
}

// buildSearchQuery creates a search query for a claim
func (f *FactChecker) buildSearchQuery(claim FactualClaim) string {
	// Extract key terms from claim
	words := strings.Fields(claim.Text)
	keyWords := []string{}
	
	for _, word := range words {
		if len(word) > 3 && !isStopWord(word) {
			keyWords = append(keyWords, word)
		}
	}
	
	return strings.Join(keyWords, " ")
}

// extractDomain extracts domain from URL
func extractDomain(url string) string {
	regex := regexp.MustCompile(`https?://([^/]+)`)
	matches := regex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// processClaimResult processes the result of checking a claim
func (f *FactChecker) processClaimResult(result *ClaimCheckResult, factCheckResult *FactCheckResult) {
	switch result.Status {
	case "verified":
		if result.Verified != nil {
			factCheckResult.VerifiedFacts = append(factCheckResult.VerifiedFacts, *result.Verified)
		}
	case "contradicted":
		if result.Issue != nil {
			factCheckResult.FactualErrors = append(factCheckResult.FactualErrors, *result.Issue)
		}
	case "unverifiable":
		if result.Issue != nil {
			factCheckResult.Warnings = append(factCheckResult.Warnings, *result.Issue)
		}
	}
}

// incorporateSourceResults incorporates results from source checking
func (f *FactChecker) incorporateSourceResults(results []*ClaimCheckResult, factCheckResult *FactCheckResult) {
	for _, result := range results {
		f.processClaimResult(result, factCheckResult)
	}
}

// calculateFactCheckScore calculates overall fact check score
func (f *FactChecker) calculateFactCheckScore(result *FactCheckResult) float64 {
	if result.TotalClaims == 0 {
		return 100.0 // No claims to verify
	}

	// Base score calculation
	verifiedRatio := float64(result.VerifiedClaims) / float64(result.TotalClaims)
	errorPenalty := float64(result.ErrorCount) * 10.0
	warningPenalty := float64(result.WarningCount) * 5.0

	score := (verifiedRatio * 100.0) - errorPenalty - warningPenalty

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// calculateConfidence calculates confidence level
func (f *FactChecker) calculateConfidence(result *FactCheckResult) float64 {
	if result.TotalClaims == 0 {
		return 1.0
	}

	// Confidence based on verification ratio and source quality
	verificationRatio := float64(result.VerifiedClaims+len(result.FactualErrors)) / float64(result.TotalClaims)
	return verificationRatio * 0.9 // Max 90% confidence
}

// FactSourceDatabase manages reliable sources for fact checking
type FactSourceDatabase struct {
	sources map[string]float64 // URL pattern -> credibility score
}

// NewFactSourceDatabase creates a new fact source database
func NewFactSourceDatabase() FactSourceDatabase {
	sources := map[string]float64{
		// Academic and educational
		"edu":           0.9,
		"scholar.google": 0.95,
		"pubmed":        0.95,
		"arxiv.org":     0.85,
		
		// Government sources
		"gov":           0.9,
		"who.int":       0.9,
		"cdc.gov":       0.9,
		
		// News organizations
		"reuters.com":   0.85,
		"ap.org":        0.85,
		"bbc.com":       0.8,
		
		// Encyclopedias
		"britannica.com": 0.8,
		"wikipedia.org":  0.7,
		
		// Default for unknown sources
		"default":       0.5,
	}

	return FactSourceDatabase{sources: sources}
}

// GetSourceCredibility returns credibility score for a source
func (db FactSourceDatabase) GetSourceCredibility(url string) float64 {
	for pattern, score := range db.sources {
		if strings.Contains(url, pattern) {
			return score
		}
	}
	return db.sources["default"]
}
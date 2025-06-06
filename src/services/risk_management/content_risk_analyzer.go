package risk_management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// ContentRiskAnalyzerImpl implements content risk analysis
type ContentRiskAnalyzerImpl struct {
	riskRepo         repositories.RiskRepository
	eventRepo        repositories.EventRepository
	violencePatterns []*regexp.Regexp
	hatePatterns     []*regexp.Regexp
	piiPatterns      map[string]*regexp.Regexp
}

// NewContentRiskAnalyzer creates a new content risk analyzer
func NewContentRiskAnalyzer(
	riskRepo repositories.RiskRepository,
	eventRepo repositories.EventRepository,
) *ContentRiskAnalyzerImpl {
	return &ContentRiskAnalyzerImpl{
		riskRepo:         riskRepo,
		eventRepo:        eventRepo,
		violencePatterns: initializeViolencePatterns(),
		hatePatterns:     initializeHatePatterns(),
		piiPatterns:      initializePIIPatterns(),
	}
}

// AnalyzeContent analyzes content for various risks
func (a *ContentRiskAnalyzerImpl) AnalyzeContent(ctx context.Context, content string) (*ContentRiskResult, error) {
	result := &ContentRiskResult{
		Safe:       true,
		RiskScore:  0.0,
		Violations: make([]*ContentViolation, 0),
		Metadata:   make(map[string]interface{}),
	}

	// Check content policies
	policyResult, err := a.CheckContentPolicy(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("content policy check failed: %w", err)
	}

	if !policyResult.Compliant {
		result.Safe = false
		for _, violation := range policyResult.Violations {
			result.Violations = append(result.Violations, &ContentViolation{
				PolicyType: violation.PolicyName,
				Severity:   violation.Severity,
				Confidence: 0.9, // High confidence for pattern matching
				Location:   "",
				Text:       "",
				Action:     violation.Action,
			})
		}
	}

	// Check for PII
	piiResult, err := a.CheckPII(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("PII check failed: %w", err)
	}

	if piiResult.ContainsPII {
		result.Safe = false
		result.Violations = append(result.Violations, &ContentViolation{
			PolicyType: "pii_exposure",
			Severity:   "high",
			Confidence: 0.95,
			Location:   "content",
			Text:       "PII detected",
			Action:     "redact",
		})
		result.Metadata["pii_entities"] = piiResult.PIIEntities
		result.Recommendations = append(result.Recommendations, "Remove or anonymize PII before publishing")
	}

	// Check copyright (simplified for now)
	copyrightResult, err := a.CheckCopyright(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("copyright check failed: %w", err)
	}

	if !copyrightResult.Clear {
		result.RiskScore += 0.3
		result.Recommendations = append(result.Recommendations, "Review content for potential copyright issues")
	}

	// Calculate overall risk score
	result.RiskScore = a.calculateRiskScore(result)

	// Create risk entity if score is significant
	if result.RiskScore > 0.5 {
		risk := &entities.Risk{
			ID:                uuid.New(),
			Category:          entities.RiskTypeContent,
			Severity:          a.scoresToSeverity(result.RiskScore),
			Status:            entities.RiskStatusIdentified,
			Title:             "Content Risk Detected",
			Description:       fmt.Sprintf("Content analysis identified %d violations", len(result.Violations)),
			Likelihood:        0.8,
			Impact:            0.6,
			MitigationActions: result.Recommendations,
			Metadata:          map[string]interface{}{"source": "content_analyzer", "analysis": result.Metadata},
			IdentifiedAt:      time.Now(),
			LastAssessment:    time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := a.riskRepo.CreateRisk(ctx, risk); err != nil {
			return nil, fmt.Errorf("failed to create risk: %w", err)
		}
	}

	return result, nil
}

// CheckPII checks content for personally identifiable information
func (a *ContentRiskAnalyzerImpl) CheckPII(ctx context.Context, content string) (*PIICheckResult, error) {
	result := &PIICheckResult{
		ContainsPII: false,
		PIIEntities: make([]*PIIEntity, 0),
	}

	// Check for various PII patterns
	for piiType, pattern := range a.piiPatterns {
		matches := pattern.FindAllStringIndex(content, -1)
		for _, match := range matches {
			result.ContainsPII = true
			entity := &PIIEntity{
				Type:       piiType,
				Value:      a.maskPII(content[match[0]:match[1]], piiType),
				Location:   match[0],
				Confidence: 0.9,
			}
			result.PIIEntities = append(result.PIIEntities, entity)
		}
	}

	// Generate anonymized version if PII found
	if result.ContainsPII {
		result.Anonymized = a.anonymizeContent(content, result.PIIEntities)
	}

	return result, nil
}

// CheckContentPolicy checks content against active policies
func (a *ContentRiskAnalyzerImpl) CheckContentPolicy(ctx context.Context, content string) (*PolicyCheckResult, error) {
	result := &PolicyCheckResult{
		Compliant:  true,
		Violations: make([]*PolicyViolation, 0),
	}

	// Get active content policies
	policies, err := a.riskRepo.GetActiveContentPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get content policies: %w", err)
	}

	contentLower := strings.ToLower(content)

	for _, policy := range policies {
		if policy.Type == "violence" {
			if a.checkViolence(contentLower) {
				result.Compliant = false
				result.Violations = append(result.Violations, &PolicyViolation{
					PolicyID:    policy.ID,
					PolicyName:  policy.Name,
					Description: "Content contains violence-related terms",
					Severity:    "high",
					Action:      "reject",
				})
			}
		} else if policy.Type == "hate_speech" {
			if a.checkHateSpeech(contentLower) {
				result.Compliant = false
				result.Violations = append(result.Violations, &PolicyViolation{
					PolicyID:    policy.ID,
					PolicyName:  policy.Name,
					Description: "Content contains hate speech",
					Severity:    "critical",
					Action:      "reject",
				})
			}
		}
	}

	return result, nil
}

// CheckCopyright performs basic copyright checking
func (a *ContentRiskAnalyzerImpl) CheckCopyright(ctx context.Context, content string) (*CopyrightCheckResult, error) {
	result := &CopyrightCheckResult{
		Clear:     true,
		Matches:   make([]*CopyrightMatch, 0),
		RiskLevel: "low",
	}

	// Generate content hash for comparison
	hash := a.generateContentHash(content)

	// Check for common copyright indicators
	copyrightIndicators := []string{
		"© ",
		"copyright",
		"all rights reserved",
		"reprinted with permission",
		"proprietary",
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range copyrightIndicators {
		if strings.Contains(contentLower, indicator) {
			result.Clear = false
			result.RiskLevel = "medium"
			result.Matches = append(result.Matches, &CopyrightMatch{
				Source:     "internal_check",
				Similarity: 0.0,
				Owner:      "Unknown",
			})
			break
		}
	}

	// Store hash for future comparisons
	result.Matches = append(result.Matches, &CopyrightMatch{
		Source:     "content_hash",
		Similarity: 0.0,
		URL:        hash,
	})

	return result, nil
}

// Helper methods

func (a *ContentRiskAnalyzerImpl) checkViolence(content string) bool {
	for _, pattern := range a.violencePatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

func (a *ContentRiskAnalyzerImpl) checkHateSpeech(content string) bool {
	for _, pattern := range a.hatePatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

func (a *ContentRiskAnalyzerImpl) calculateRiskScore(result *ContentRiskResult) float64 {
	score := 0.0

	// Weight violations by severity
	for _, violation := range result.Violations {
		switch violation.Severity {
		case "critical":
			score += 0.4
		case "high":
			score += 0.3
		case "medium":
			score += 0.2
		case "low":
			score += 0.1
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (a *ContentRiskAnalyzerImpl) scoresToSeverity(score float64) entities.RiskSeverity {
	switch {
	case score >= 0.8:
		return entities.RiskSeverityCritical
	case score >= 0.6:
		return entities.RiskSeverityHigh
	case score >= 0.4:
		return entities.RiskSeverityMedium
	default:
		return entities.RiskSeverityLow
	}
}

func (a *ContentRiskAnalyzerImpl) maskPII(value, piiType string) string {
	switch piiType {
	case "email":
		parts := strings.Split(value, "@")
		if len(parts) == 2 {
			return maskString(parts[0], 2) + "@" + parts[1]
		}
	case "phone":
		if len(value) > 6 {
			return value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
		}
	case "ssn":
		if len(value) > 4 {
			return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
		}
	case "credit_card":
		if len(value) > 4 {
			return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
		}
	}
	return strings.Repeat("*", len(value))
}

func (a *ContentRiskAnalyzerImpl) anonymizeContent(content string, entities []*PIIEntity) string {
	anonymized := content

	// Sort entities by location in reverse order to avoid offset issues
	for i := len(entities) - 1; i >= 0; i-- {
		entity := entities[i]
		// Find the actual value at the location
		start := entity.Location
		pattern := a.piiPatterns[entity.Type]
		if matches := pattern.FindStringIndex(content[start:]); matches != nil {
			end := start + matches[1]
			replacement := fmt.Sprintf("[%s_REDACTED]", strings.ToUpper(entity.Type))
			anonymized = anonymized[:start] + replacement + anonymized[end:]
		}
	}

	return anonymized
}

func (a *ContentRiskAnalyzerImpl) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func maskString(s string, visibleChars int) string {
	if len(s) <= visibleChars {
		return s
	}
	return s[:visibleChars] + strings.Repeat("*", len(s)-visibleChars)
}


// Pattern initialization functions

func initializeViolencePatterns() []*regexp.Regexp {
	patterns := []string{
		`\b(kill|murder|assault|attack|violence|violent|weapon|gun|knife|bomb|explosive)\b`,
		`\b(harm|hurt|injure|wound|damage|destroy|death|die|dead)\b`,
		`\b(fight|combat|war|battle|conflict)\b`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

func initializeHatePatterns() []*regexp.Regexp {
	// Simplified patterns for hate speech detection
	// In production, this would use more sophisticated ML models
	patterns := []string{
		`\b(hate|hatred|discriminat|racist|sexist|bigot)\b`,
		`\b(inferior|superior) (race|gender|religion)\b`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

func initializePIIPatterns() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"email":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		"phone":       regexp.MustCompile(`\b(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})\b`),
		"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		"credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
		"ip_address":  regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	}
}

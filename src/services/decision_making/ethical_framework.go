package decision_making

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// EthicalFrameworkImpl implements the EthicalFramework interface
type EthicalFrameworkImpl struct {
	decisionRepo repositories.DecisionRepository
	eventRepo    repositories.EventRepository
	llmClient    LLMClient
	guidelines   map[string]*entities.EthicalGuideline // Cache of guidelines
}

// NewEthicalFramework creates a new ethical framework instance
func NewEthicalFramework(
	decisionRepo repositories.DecisionRepository,
	eventRepo repositories.EventRepository,
	llmClient LLMClient,
) *EthicalFrameworkImpl {
	return &EthicalFrameworkImpl{
		decisionRepo: decisionRepo,
		eventRepo:    eventRepo,
		llmClient:    llmClient,
		guidelines:   make(map[string]*entities.EthicalGuideline),
	}
}

// Core ethical principles
var coreEthicalPrinciples = []entities.EthicalGuideline{
	{
		ID:          "eth-001",
		Principle:   "Do No Harm",
		Description: "Avoid actions that could cause physical, emotional, financial, or reputational harm to individuals or society",
		Examples: []string{
			"Refusing to create misleading content",
			"Protecting user privacy and data",
			"Avoiding discriminatory practices",
		},
		RedLines: []string{
			"Creating content that promotes violence or hate",
			"Violating user privacy without consent",
			"Engaging in deceptive practices",
		},
		Weight: 1.0,
	},
	{
		ID:          "eth-002",
		Principle:   "Fairness and Non-Discrimination",
		Description: "Treat all stakeholders equitably and avoid bias based on protected characteristics",
		Examples: []string{
			"Equal pricing for similar services",
			"Unbiased content creation",
			"Fair resource allocation",
		},
		RedLines: []string{
			"Discriminatory pricing based on demographics",
			"Biased content that stereotypes groups",
			"Unfair treatment of clients or partners",
		},
		Weight: 0.9,
	},
	{
		ID:          "eth-003",
		Principle:   "Transparency and Honesty",
		Description: "Be truthful and open about capabilities, limitations, and decision-making processes",
		Examples: []string{
			"Clear disclosure of AI-generated content",
			"Honest representation of services",
			"Transparent pricing and terms",
		},
		RedLines: []string{
			"Misrepresenting capabilities",
			"Hidden fees or terms",
			"Plagiarism or uncredited work",
		},
		Weight: 0.9,
	},
	{
		ID:          "eth-004",
		Principle:   "Respect for Autonomy",
		Description: "Respect individual choice and consent in all interactions",
		Examples: []string{
			"Obtaining consent for data usage",
			"Allowing opt-out mechanisms",
			"Respecting client preferences",
		},
		RedLines: []string{
			"Using data without consent",
			"Forcing unwanted services",
			"Ignoring explicit user preferences",
		},
		Weight: 0.85,
	},
	{
		ID:          "eth-005",
		Principle:   "Beneficence",
		Description: "Actively work to benefit stakeholders and society",
		Examples: []string{
			"Creating valuable, helpful content",
			"Contributing to positive outcomes",
			"Supporting client success",
		},
		RedLines: []string{
			"Creating harmful or useless content",
			"Wasting resources on low-value activities",
			"Prioritizing profit over stakeholder benefit",
		},
		Weight: 0.8,
	},
	{
		ID:          "eth-006",
		Principle:   "Environmental Responsibility",
		Description: "Minimize environmental impact and promote sustainability",
		Examples: []string{
			"Efficient resource usage",
			"Promoting sustainable practices",
			"Reducing computational waste",
		},
		RedLines: []string{
			"Excessive resource consumption",
			"Promoting environmentally harmful practices",
			"Ignoring environmental impact",
		},
		Weight: 0.7,
	},
}

// ValidateEthics checks if a decision aligns with ethical guidelines
func (ef *EthicalFrameworkImpl) ValidateEthics(ctx context.Context, decision *entities.Decision) (*EthicalValidationResult, error) {
	// Load guidelines if not cached
	if err := ef.loadGuidelines(ctx); err != nil {
		return nil, fmt.Errorf("failed to load guidelines: %w", err)
	}

	concerns := []EthicalConcern{}
	redLineViolations := []string{}
	totalScore := 0.0
	totalWeight := 0.0

	// Check each guideline
	for _, guideline := range ef.guidelines {
		score, concern, redLine := ef.evaluateGuideline(ctx, decision, guideline)

		if redLine != "" {
			redLineViolations = append(redLineViolations, redLine)
		}

		if concern != nil {
			concerns = append(concerns, *concern)
		}

		totalScore += score * guideline.Weight
		totalWeight += guideline.Weight
	}

	// Calculate ethical score
	ethicalScore := 0.0
	if totalWeight > 0 {
		ethicalScore = totalScore / totalWeight
	}

	// Generate justification
	justification, err := ef.GenerateEthicalJustification(ctx, decision)
	if err != nil {
		justification = "Unable to generate ethical justification"
	}

	// Decision is approved if no red lines are violated and score is acceptable
	approved := len(redLineViolations) == 0 && ethicalScore >= 0.7

	result := &EthicalValidationResult{
		Approved:          approved,
		Concerns:          concerns,
		RedLineViolations: redLineViolations,
		EthicalScore:      ethicalScore,
		Justification:     justification,
	}

	// Log ethical validation
	ef.logEthicalValidation(ctx, decision.ID, result)

	return result, nil
}

// CheckRedLines performs a quick check for ethical red line violations
func (ef *EthicalFrameworkImpl) CheckRedLines(ctx context.Context, action string, context map[string]interface{}) (bool, error) {
	// Load guidelines if not cached
	if err := ef.loadGuidelines(ctx); err != nil {
		return false, fmt.Errorf("failed to load guidelines: %w", err)
	}

	// Check each guideline's red lines
	for _, guideline := range ef.guidelines {
		for _, redLine := range guideline.RedLines {
			if ef.violatesRedLine(action, context, redLine) {
				return false, nil
			}
		}
	}

	return true, nil
}

// AssessBias evaluates potential bias in a decision
func (ef *EthicalFrameworkImpl) AssessBias(ctx context.Context, decision *entities.Decision) (*BiasAssessment, error) {
	biasTypes := []string{}
	affectedGroups := []string{}
	recommendations := []string{}

	// Check for various types of bias
	demographicBias := ef.checkDemographicBias(decision)
	if demographicBias.detected {
		biasTypes = append(biasTypes, "demographic")
		affectedGroups = append(affectedGroups, demographicBias.affectedGroups...)
		recommendations = append(recommendations, demographicBias.recommendations...)
	}

	confirmationBias := ef.checkConfirmationBias(decision)
	if confirmationBias.detected {
		biasTypes = append(biasTypes, "confirmation")
		recommendations = append(recommendations, confirmationBias.recommendations...)
	}

	availabilityBias := ef.checkAvailabilityBias(decision)
	if availabilityBias.detected {
		biasTypes = append(biasTypes, "availability")
		recommendations = append(recommendations, availabilityBias.recommendations...)
	}

	// Calculate overall bias score
	biasScore := float64(len(biasTypes)) / 10.0 // Normalized score

	assessment := &BiasAssessment{
		BiasDetected:    len(biasTypes) > 0,
		BiasTypes:       biasTypes,
		BiasScore:       biasScore,
		AffectedGroups:  affectedGroups,
		Recommendations: recommendations,
	}

	return assessment, nil
}

// RegisterGuideline adds a new ethical guideline
func (ef *EthicalFrameworkImpl) RegisterGuideline(ctx context.Context, guideline *entities.EthicalGuideline) error {
	// Validate guideline
	if err := ef.validateGuideline(guideline); err != nil {
		return fmt.Errorf("guideline validation failed: %w", err)
	}

	// Set ID if not provided
	if guideline.ID == "" {
		guideline.ID = uuid.New().String()
	}

	// Save guideline
	if err := ef.decisionRepo.CreateEthicalGuideline(ctx, guideline); err != nil {
		return fmt.Errorf("failed to create guideline: %w", err)
	}

	// Update cache
	ef.guidelines[guideline.ID] = guideline

	return nil
}

// GetActiveGuidelines returns all active ethical guidelines
func (ef *EthicalFrameworkImpl) GetActiveGuidelines(ctx context.Context) ([]*entities.EthicalGuideline, error) {
	guidelines, err := ef.decisionRepo.ListEthicalGuidelines(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list guidelines: %w", err)
	}

	// Sort by weight (most important first)
	sort.Slice(guidelines, func(i, j int) bool {
		return guidelines[i].Weight > guidelines[j].Weight
	})

	return guidelines, nil
}

// GenerateEthicalJustification creates an explanation for ethical decisions
func (ef *EthicalFrameworkImpl) GenerateEthicalJustification(ctx context.Context, decision *entities.Decision) (string, error) {
	// Build context for LLM
	prompt := fmt.Sprintf(`Generate an ethical justification for this decision:

Decision: %s
Type: %s
Description: %s

Selected Option: %s

Ethical Principles Considered:
%s

Provide a clear, concise justification that explains how this decision aligns with ethical principles.`,
		decision.Title,
		decision.Type,
		decision.Description,
		decision.SelectedOption.Title,
		ef.formatPrinciples(),
	)

	justification, err := ef.llmClient.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate justification: %w", err)
	}

	return justification, nil
}

// IdentifyEthicalConcerns analyzes a scenario for potential ethical issues
func (ef *EthicalFrameworkImpl) IdentifyEthicalConcerns(ctx context.Context, scenario string) ([]EthicalConcern, error) {
	// Use LLM to identify concerns
	_ = fmt.Sprintf(`Analyze this scenario for ethical concerns:

%s

Consider these ethical principles:
%s

Identify any ethical concerns, their severity, and potential mitigations.`,
		scenario,
		ef.formatPrinciples(),
	)

	// This would call LLM to analyze concerns
	// For now, return empty list
	return []EthicalConcern{}, nil
}

// Helper methods

func (ef *EthicalFrameworkImpl) loadGuidelines(ctx context.Context) error {
	if len(ef.guidelines) > 0 {
		return nil // Already loaded
	}

	// Load core principles
	for i := range coreEthicalPrinciples {
		guideline := &coreEthicalPrinciples[i]
		ef.guidelines[guideline.ID] = guideline
	}

	// Load custom guidelines from database
	customGuidelines, err := ef.decisionRepo.ListEthicalGuidelines(ctx)
	if err != nil {
		return err
	}

	for _, guideline := range customGuidelines {
		ef.guidelines[guideline.ID] = guideline
	}

	return nil
}

func (ef *EthicalFrameworkImpl) evaluateGuideline(ctx context.Context, decision *entities.Decision, guideline *entities.EthicalGuideline) (float64, *EthicalConcern, string) {
	score := 1.0 // Start with perfect score
	var concern *EthicalConcern
	redLine := ""

	// Check for red line violations
	for _, rl := range guideline.RedLines {
		if ef.decisionViolatesRedLine(decision, rl) {
			redLine = fmt.Sprintf("%s: %s", guideline.Principle, rl)
			score = 0.0
			break
		}
	}

	// Evaluate alignment with principle
	if redLine == "" {
		alignment := ef.evaluateAlignment(decision, guideline)
		score = alignment

		// Create concern if score is low
		if score < 0.7 {
			severity := "low"
			if score < 0.5 {
				severity = "medium"
			}
			if score < 0.3 {
				severity = "high"
			}

			concern = &EthicalConcern{
				GuidelineID: guideline.ID,
				Principle:   guideline.Principle,
				Concern:     fmt.Sprintf("Decision may not fully align with %s", guideline.Principle),
				Severity:    severity,
				Mitigation:  ef.suggestMitigation(decision, guideline),
			}
		}
	}

	return score, concern, redLine
}

func (ef *EthicalFrameworkImpl) violatesRedLine(action string, context map[string]interface{}, redLine string) bool {
	// Check if action violates red line
	actionLower := strings.ToLower(action)
	redLineLower := strings.ToLower(redLine)

	// Simple keyword matching
	keywords := []string{"violence", "hate", "deceptive", "discriminatory", "privacy", "harmful"}
	for _, keyword := range keywords {
		if strings.Contains(redLineLower, keyword) && strings.Contains(actionLower, keyword) {
			return true
		}
	}

	return false
}

func (ef *EthicalFrameworkImpl) decisionViolatesRedLine(decision *entities.Decision, redLine string) bool {
	// Check decision content against red line
	_ = strings.ToLower(redLine)

	// Check title and description
	titleLower := strings.ToLower(decision.Title)
	descLower := strings.ToLower(decision.Description)

	// Look for violation indicators
	if strings.Contains(redLine, "violence") || strings.Contains(redLine, "hate") {
		violenceKeywords := []string{"attack", "harm", "hurt", "destroy", "kill", "hate"}
		for _, keyword := range violenceKeywords {
			if strings.Contains(titleLower, keyword) || strings.Contains(descLower, keyword) {
				return true
			}
		}
	}

	if strings.Contains(redLine, "privacy") {
		privacyKeywords := []string{"personal data", "private information", "without consent"}
		for _, keyword := range privacyKeywords {
			if strings.Contains(titleLower, keyword) || strings.Contains(descLower, keyword) {
				return true
			}
		}
	}

	if strings.Contains(redLine, "deceptive") || strings.Contains(redLine, "misleading") {
		deceptionKeywords := []string{"fake", "false", "mislead", "deceive", "trick"}
		for _, keyword := range deceptionKeywords {
			if strings.Contains(titleLower, keyword) || strings.Contains(descLower, keyword) {
				return true
			}
		}
	}

	return false
}

func (ef *EthicalFrameworkImpl) evaluateAlignment(decision *entities.Decision, guideline *entities.EthicalGuideline) float64 {
	// Evaluate how well the decision aligns with the guideline
	// This is a simplified scoring mechanism

	score := 0.5 // Neutral starting point

	// Check for positive alignment
	for _, example := range guideline.Examples {
		if ef.decisionAlignsWithExample(decision, example) {
			score += 0.1
		}
	}

	// Cap at 1.0
	return math.Min(score, 1.0)
}

func (ef *EthicalFrameworkImpl) decisionAlignsWithExample(decision *entities.Decision, example string) bool {
	// Simple keyword matching for alignment
	exampleLower := strings.ToLower(example)
	titleLower := strings.ToLower(decision.Title)
	descLower := strings.ToLower(decision.Description)

	// Extract key concepts from example
	positiveKeywords := []string{"protect", "fair", "equal", "transparent", "honest", "consent", "benefit", "help", "sustainable"}

	for _, keyword := range positiveKeywords {
		if strings.Contains(exampleLower, keyword) &&
			(strings.Contains(titleLower, keyword) || strings.Contains(descLower, keyword)) {
			return true
		}
	}

	return false
}

func (ef *EthicalFrameworkImpl) suggestMitigation(decision *entities.Decision, guideline *entities.EthicalGuideline) string {
	// Suggest how to better align with the guideline
	suggestions := map[string]string{
		"Do No Harm":                      "Consider potential negative impacts and add safeguards",
		"Fairness and Non-Discrimination": "Review for bias and ensure equitable treatment",
		"Transparency and Honesty":        "Add clear disclosures and be explicit about limitations",
		"Respect for Autonomy":            "Ensure proper consent and provide opt-out options",
		"Beneficence":                     "Focus on maximizing positive outcomes for stakeholders",
		"Environmental Responsibility":    "Optimize for resource efficiency and sustainability",
	}

	if suggestion, exists := suggestions[guideline.Principle]; exists {
		return suggestion
	}

	return "Review decision to better align with ethical principles"
}

// Bias checking methods

type biasCheckResult struct {
	detected        bool
	affectedGroups  []string
	recommendations []string
}

func (ef *EthicalFrameworkImpl) checkDemographicBias(decision *entities.Decision) biasCheckResult {
	result := biasCheckResult{detected: false}

	// Check for demographic-related terms
	demographics := []string{"age", "gender", "race", "ethnicity", "religion", "nationality", "disability"}
	content := strings.ToLower(decision.Title + " " + decision.Description)

	for _, demo := range demographics {
		if strings.Contains(content, demo) {
			result.detected = true
			result.affectedGroups = append(result.affectedGroups, demo)
		}
	}

	if result.detected {
		result.recommendations = append(result.recommendations,
			"Review decision for demographic bias",
			"Ensure equal treatment across all demographic groups",
			"Consider using demographic-neutral criteria")
	}

	return result
}

func (ef *EthicalFrameworkImpl) checkConfirmationBias(decision *entities.Decision) biasCheckResult {
	result := biasCheckResult{detected: false}

	// Check if all options are too similar
	if len(decision.Options) > 1 {
		// Simple check: if all options have very similar scores
		scoreRange := decision.Options[0].Score - decision.Options[len(decision.Options)-1].Score
		if scoreRange < 0.1 {
			result.detected = true
			result.recommendations = append(result.recommendations,
				"Consider more diverse options",
				"Challenge existing assumptions",
				"Seek contrarian viewpoints")
		}
	}

	return result
}

func (ef *EthicalFrameworkImpl) checkAvailabilityBias(decision *entities.Decision) biasCheckResult {
	result := biasCheckResult{detected: false}

	// Check if decision relies too heavily on recent events
	if ctx, exists := decision.Context["based_on_recent"].(bool); exists && ctx {
		result.detected = true
		result.recommendations = append(result.recommendations,
			"Consider historical data beyond recent events",
			"Analyze long-term trends",
			"Avoid overweighting recent experiences")
	}

	return result
}

func (ef *EthicalFrameworkImpl) validateGuideline(guideline *entities.EthicalGuideline) error {
	if guideline.Principle == "" {
		return fmt.Errorf("principle is required")
	}

	if guideline.Description == "" {
		return fmt.Errorf("description is required")
	}

	if guideline.Weight < 0 || guideline.Weight > 1 {
		return fmt.Errorf("weight must be between 0 and 1")
	}

	return nil
}

func (ef *EthicalFrameworkImpl) formatPrinciples() string {
	principles := []string{}
	for _, guideline := range ef.guidelines {
		principles = append(principles, fmt.Sprintf("- %s: %s", guideline.Principle, guideline.Description))
	}
	return strings.Join(principles, "\n")
}

func (ef *EthicalFrameworkImpl) logEthicalValidation(ctx context.Context, decisionID string, result *EthicalValidationResult) {
	// Log ethical validation result
	logValidation := &entities.DecisionLog{
		ID:          uuid.New().String(),
		DecisionID:  decisionID,
		Timestamp:   time.Now(),
		EventType:   "ethical_validation",
		Description: fmt.Sprintf("Ethical validation: approved=%v, score=%.2f", result.Approved, result.EthicalScore),
		Actor:       "ethical_framework",
		Changes: map[string]interface{}{
			"concerns":  len(result.Concerns),
			"red_lines": len(result.RedLineViolations),
			"score":     result.EthicalScore,
		},
	}
	if err := ef.decisionRepo.CreateDecisionLog(ctx, logValidation); err != nil {
		// Log error but don't fail the function
		log.Printf("Failed to create decision log: %v", err)
	}
}

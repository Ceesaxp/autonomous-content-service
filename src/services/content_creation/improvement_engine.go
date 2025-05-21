package content_creation

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ImprovementEngine generates targeted improvement suggestions for low-scoring content
type ImprovementEngine struct {
	llmClient        LLMClient
	evaluationEngine *EvaluationEngine
}

// NewImprovementEngine creates a new improvement engine
func NewImprovementEngine(llmClient LLMClient, evaluationEngine *EvaluationEngine) *ImprovementEngine {
	return &ImprovementEngine{
		llmClient:        llmClient,
		evaluationEngine: evaluationEngine,
	}
}

// ImprovementRequest contains parameters for generating improvements
type ImprovementRequest struct {
	Content           string
	ContentType       entities.ContentType
	EvaluationResults *DetailedEvaluation
	FactCheckResults  *FactCheckResult
	StyleResults      *StyleAnalysisResult
	TargetScore       float64
	CurrentScore      float64
	MaxSuggestions    int
	Focus             []ImprovementFocus
}

// ImprovementResult contains generated improvement suggestions
type ImprovementResult struct {
	Suggestions       []ImprovementSuggestion `json:"suggestions"`
	PriorityMatrix    PriorityMatrix          `json:"priorityMatrix"`
	ExpectedImpact    ExpectedImpact          `json:"expectedImpact"`
	ImplementationPlan []ImplementationStep   `json:"implementationPlan"`
	ProcessingTime    time.Duration           `json:"processingTime"`
}

// ImprovementSuggestion represents a specific improvement recommendation
type ImprovementSuggestion struct {
	ID              string                 `json:"id"`
	Priority        ImprovementPriority    `json:"priority"`
	Category        ImprovementCategory    `json:"category"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Rationale       string                 `json:"rationale"`
	Implementation  string                 `json:"implementation"`
	Examples        []Example              `json:"examples"`
	ExpectedGain    float64                `json:"expectedGain"`
	Effort          EffortLevel            `json:"effort"`
	Prerequisites   []string               `json:"prerequisites"`
	Tags            []string               `json:"tags"`
}

// Example shows before/after for improvement suggestions
type Example struct {
	Before      string `json:"before"`
	After       string `json:"after"`
	Explanation string `json:"explanation"`
}

// PriorityMatrix helps prioritize improvements based on impact and effort
type PriorityMatrix struct {
	HighImpactLowEffort   []string `json:"highImpactLowEffort"`
	HighImpactHighEffort  []string `json:"highImpactHighEffort"`
	LowImpactLowEffort    []string `json:"lowImpactLowEffort"`
	LowImpactHighEffort   []string `json:"lowImpactHighEffort"`
}

// ExpectedImpact forecasts the improvement potential
type ExpectedImpact struct {
	TotalPotentialGain    float64            `json:"totalPotentialGain"`
	CategoryImpacts       map[string]float64 `json:"categoryImpacts"`
	RiskAssessment        RiskAssessment     `json:"riskAssessment"`
	TimeToImplement       time.Duration      `json:"timeToImplement"`
	SuccessProbability   float64            `json:"successProbability"`
}

// ImplementationStep provides a step-by-step improvement plan
type ImplementationStep struct {
	Step        int                 `json:"step"`
	Phase       ImplementationPhase `json:"phase"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Suggestions []string            `json:"suggestions"`
	EstimatedTime time.Duration     `json:"estimatedTime"`
	Dependencies  []int             `json:"dependencies"`
}

// RiskAssessment evaluates risks of implementing improvements
type RiskAssessment struct {
	OverallRisk     RiskLevel        `json:"overallRisk"`
	SpecificRisks   []Risk           `json:"specificRisks"`
	MitigationSteps []string         `json:"mitigationSteps"`
}

// Risk represents a specific implementation risk
type Risk struct {
	Type        RiskType  `json:"type"`
	Description string    `json:"description"`
	Impact      RiskLevel `json:"impact"`
	Likelihood  RiskLevel `json:"likelihood"`
	Mitigation  string    `json:"mitigation"`
}

// Enums and types

type ImprovementPriority string

const (
	PriorityQuickWin   ImprovementPriority = "quick_win"
	PriorityEssential  ImprovementPriority = "essential"
	PriorityImportant  ImprovementPriority = "important"
	PriorityNiceToHave ImprovementPriority = "nice_to_have"
)

type ImprovementCategory string

const (
	CategoryStructure     ImprovementCategory = "structure"
	CategoryContent       ImprovementCategory = "content"
	CategoryLanguage      ImprovementCategory = "language"
	CategoryEngagement    ImprovementCategory = "engagement"
	CategoryCredibility   ImprovementCategory = "credibility"
	CategoryTechnical     ImprovementCategory = "technical"
	CategorySEO           ImprovementCategory = "seo"
	CategoryAccessibility ImprovementCategory = "accessibility"
	CategoryOptimization  ImprovementCategory = "optimization"
)

type ImprovementFocus string

const (
	FocusReadability    ImprovementFocus = "readability"
	FocusEngagement     ImprovementFocus = "engagement"
	FocusAccuracy       ImprovementFocus = "accuracy"
	FocusStyle          ImprovementFocus = "style"
	FocusStructure      ImprovementFocus = "structure"
	FocusOptimization   ImprovementFocus = "optimization"
)

type EffortLevel string

const (
	EffortLow    EffortLevel = "low"
	EffortMedium EffortLevel = "medium"
	EffortHigh   EffortLevel = "high"
)

type ImplementationPhase string

const (
	PhaseQuickFixes    ImplementationPhase = "quick_fixes"
	PhaseContentRevision ImplementationPhase = "content_revision"
	PhaseStructuralChanges ImplementationPhase = "structural_changes"
	PhaseOptimization  ImplementationPhase = "optimization"
)

type RiskType string

const (
	RiskContentIntegrity RiskType = "content_integrity"
	RiskToneShift       RiskType = "tone_shift"
	RiskOverOptimization RiskType = "over_optimization"
	RiskTimeConstraint   RiskType = "time_constraint"
)

// GenerateImprovements creates targeted improvement suggestions
func (ie *ImprovementEngine) GenerateImprovements(ctx context.Context, request ImprovementRequest) (*ImprovementResult, error) {
	startTime := time.Now()
	
	result := &ImprovementResult{
		Suggestions: []ImprovementSuggestion{},
	}

	// 1. Analyze current weaknesses
	weaknesses := ie.identifyWeaknesses(request)

	// 2. Generate targeted suggestions for each weakness
	for _, weakness := range weaknesses {
		suggestions, err := ie.generateSuggestionsForWeakness(ctx, weakness, request)
		if err != nil {
			continue // Log error but continue with other weaknesses
		}
		result.Suggestions = append(result.Suggestions, suggestions...)
	}

	// 3. Generate additional suggestions based on focus areas
	if len(request.Focus) > 0 {
		focusedSuggestions, err := ie.generateFocusedSuggestions(ctx, request)
		if err == nil {
			result.Suggestions = append(result.Suggestions, focusedSuggestions...)
		}
	}

	// 4. Prioritize suggestions
	ie.prioritizeSuggestions(result.Suggestions)

	// 5. Limit to max suggestions if specified
	if request.MaxSuggestions > 0 && len(result.Suggestions) > request.MaxSuggestions {
		result.Suggestions = result.Suggestions[:request.MaxSuggestions]
	}

	// 6. Generate priority matrix
	result.PriorityMatrix = ie.createPriorityMatrix(result.Suggestions)

	// 7. Calculate expected impact
	result.ExpectedImpact = ie.calculateExpectedImpact(result.Suggestions, request)

	// 8. Create implementation plan
	result.ImplementationPlan = ie.createImplementationPlan(result.Suggestions)

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// identifyWeaknesses analyzes evaluation results to identify key weaknesses
func (ie *ImprovementEngine) identifyWeaknesses(request ImprovementRequest) []Weakness {
	weaknesses := []Weakness{}

	// Analyze evaluation results
	if request.EvaluationResults != nil {
		for criterion, result := range request.EvaluationResults.CriteriaResults {
			if result.Score < 70.0 { // Threshold for weakness
				weaknesses = append(weaknesses, Weakness{
					Type:        WeaknessType(criterion),
					Severity:    ie.determineSeverity(result.Score),
					Score:       result.Score,
					Description: result.Explanation,
					Evidence:    result.Evidence,
				})
			}
		}
	}

	// Analyze fact-check results
	if request.FactCheckResults != nil && request.FactCheckResults.ErrorCount > 0 {
		weaknesses = append(weaknesses, Weakness{
			Type:        WeaknessFactualAccuracy,
			Severity:    SeverityHigh,
			Score:       request.FactCheckResults.OverallScore,
			Description: fmt.Sprintf("%d factual errors detected", request.FactCheckResults.ErrorCount),
		})
	}

	// Analyze style results
	if request.StyleResults != nil && request.StyleResults.ConsistencyScore < 80.0 {
		weaknesses = append(weaknesses, Weakness{
			Type:        WeaknessStyleConsistency,
			Severity:    ie.determineSeverity(request.StyleResults.ConsistencyScore),
			Score:       request.StyleResults.ConsistencyScore,
			Description: "Style consistency issues detected",
		})
	}

	return weaknesses
}

// generateSuggestionsForWeakness creates specific suggestions for a weakness
func (ie *ImprovementEngine) generateSuggestionsForWeakness(ctx context.Context, weakness Weakness, request ImprovementRequest) ([]ImprovementSuggestion, error) {
	prompt := fmt.Sprintf(`Generate specific, actionable improvement suggestions for the following weakness in %s content:

Weakness: %s
Current Score: %.1f/100
Description: %s
Evidence: %s

Current content:
%s

Generate 2-3 highly specific suggestions with:
1. Clear implementation steps
2. Before/after examples
3. Expected impact on quality
4. Implementation effort level

Respond in JSON format:
{
  "suggestions": [
    {
      "title": "<concise title>",
      "description": "<detailed description>",
      "rationale": "<why this improvement helps>",
      "implementation": "<step-by-step implementation>",
      "examples": [
        {
          "before": "<current problematic text>",
          "after": "<improved version>",
          "explanation": "<why this is better>"
        }
      ],
      "expectedGain": <0-20 point improvement>,
      "effort": "low|medium|high"
    }
  ]
}`, request.ContentType, weakness.Type, weakness.Score, weakness.Description, weakness.Evidence, request.Content)

	response, err := ie.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ie.parseSuggestions(response, weakness)
}

// generateFocusedSuggestions creates suggestions based on focus areas
func (ie *ImprovementEngine) generateFocusedSuggestions(ctx context.Context, request ImprovementRequest) ([]ImprovementSuggestion, error) {
	focusAreas := ""
	for _, focus := range request.Focus {
		focusAreas += string(focus) + ", "
	}

	prompt := fmt.Sprintf(`Generate improvement suggestions focused specifically on: %s

Content Type: %s
Current Score: %.1f/100
Target Score: %.1f/100

Content:
%s

Generate 3-5 targeted suggestions that will have the highest impact on the specified focus areas.

Respond in JSON format:
{
  "suggestions": [
    {
      "title": "<concise title>",
      "description": "<detailed description>",
      "rationale": "<why this improvement helps the focus area>",
      "implementation": "<step-by-step implementation>",
      "examples": [
        {
          "before": "<current text>",
          "after": "<improved version>",
          "explanation": "<improvement explanation>"
        }
      ],
      "expectedGain": <0-20 point improvement>,
      "effort": "low|medium|high",
      "tags": ["<focus area>", "<additional tags>"]
    }
  ]
}`, focusAreas, request.ContentType, request.CurrentScore, request.TargetScore, request.Content)

	response, err := ie.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ie.parseFocusedSuggestions(response)
}

// Helper types for weakness analysis

type Weakness struct {
	Type        WeaknessType `json:"type"`
	Severity    SeverityLevel `json:"severity"`
	Score       float64      `json:"score"`
	Description string       `json:"description"`
	Evidence    []string     `json:"evidence"`
}

type WeaknessType string

const (
	WeaknessReadability       WeaknessType = "readability"
	WeaknessEngagement        WeaknessType = "engagement"
	WeaknessStructure         WeaknessType = "structure"
	WeaknessClarity           WeaknessType = "clarity"
	WeaknessFactualAccuracy   WeaknessType = "factual_accuracy"
	WeaknessStyleConsistency  WeaknessType = "style_consistency"
	WeaknessSEO               WeaknessType = "seo"
	WeaknessCredibility       WeaknessType = "credibility"
)

type SeverityLevel string

const (
	SeverityLow    SeverityLevel = "low"
	SeverityMedium SeverityLevel = "medium"
	SeverityHigh   SeverityLevel = "high"
)

// determineSeverity determines severity based on score
func (ie *ImprovementEngine) determineSeverity(score float64) SeverityLevel {
	if score < 50 {
		return SeverityHigh
	} else if score < 70 {
		return SeverityMedium
	}
	return SeverityLow
}

// prioritizeSuggestions sorts suggestions by priority and expected impact
func (ie *ImprovementEngine) prioritizeSuggestions(suggestions []ImprovementSuggestion) {
	sort.Slice(suggestions, func(i, j int) bool {
		// Primary sort by priority
		priorityOrder := map[ImprovementPriority]int{
			PriorityQuickWin:   1,
			PriorityEssential:  2,
			PriorityImportant:  3,
			PriorityNiceToHave: 4,
		}
		
		if priorityOrder[suggestions[i].Priority] != priorityOrder[suggestions[j].Priority] {
			return priorityOrder[suggestions[i].Priority] < priorityOrder[suggestions[j].Priority]
		}
		
		// Secondary sort by expected gain
		return suggestions[i].ExpectedGain > suggestions[j].ExpectedGain
	})
}

// createPriorityMatrix categorizes suggestions into a priority matrix
func (ie *ImprovementEngine) createPriorityMatrix(suggestions []ImprovementSuggestion) PriorityMatrix {
	matrix := PriorityMatrix{
		HighImpactLowEffort:  []string{},
		HighImpactHighEffort: []string{},
		LowImpactLowEffort:   []string{},
		LowImpactHighEffort:  []string{},
	}

	for _, suggestion := range suggestions {
		highImpact := suggestion.ExpectedGain >= 10.0
		lowEffort := suggestion.Effort == EffortLow

		switch {
		case highImpact && lowEffort:
			matrix.HighImpactLowEffort = append(matrix.HighImpactLowEffort, suggestion.ID)
		case highImpact && !lowEffort:
			matrix.HighImpactHighEffort = append(matrix.HighImpactHighEffort, suggestion.ID)
		case !highImpact && lowEffort:
			matrix.LowImpactLowEffort = append(matrix.LowImpactLowEffort, suggestion.ID)
		default:
			matrix.LowImpactHighEffort = append(matrix.LowImpactHighEffort, suggestion.ID)
		}
	}

	return matrix
}

// calculateExpectedImpact calculates the expected overall impact
func (ie *ImprovementEngine) calculateExpectedImpact(suggestions []ImprovementSuggestion, request ImprovementRequest) ExpectedImpact {
	totalGain := 0.0
	categoryImpacts := make(map[string]float64)

	for _, suggestion := range suggestions {
		totalGain += suggestion.ExpectedGain
		categoryImpacts[string(suggestion.Category)] += suggestion.ExpectedGain
	}

	// Estimate implementation time
	totalTime := time.Duration(0)
	for _, suggestion := range suggestions {
		switch suggestion.Effort {
		case EffortLow:
			totalTime += 30 * time.Minute
		case EffortMedium:
			totalTime += 2 * time.Hour
		case EffortHigh:
			totalTime += 8 * time.Hour
		}
	}

	// Calculate success probability based on effort distribution
	successProb := ie.calculateSuccessProbability(suggestions)

	return ExpectedImpact{
		TotalPotentialGain:  totalGain,
		CategoryImpacts:     categoryImpacts,
		TimeToImplement:     totalTime,
		SuccessProbability:  successProb,
		RiskAssessment:      ie.assessImplementationRisks(suggestions),
	}
}

// calculateSuccessProbability estimates likelihood of successful implementation
func (ie *ImprovementEngine) calculateSuccessProbability(suggestions []ImprovementSuggestion) float64 {
	if len(suggestions) == 0 {
		return 1.0
	}

	lowEffortCount := 0
	for _, suggestion := range suggestions {
		if suggestion.Effort == EffortLow {
			lowEffortCount++
		}
	}

	// Higher probability with more low-effort suggestions
	baseProb := 0.7
	effortBonus := float64(lowEffortCount) / float64(len(suggestions)) * 0.3
	
	return baseProb + effortBonus
}

// assessImplementationRisks assesses risks of implementing suggestions
func (ie *ImprovementEngine) assessImplementationRisks(suggestions []ImprovementSuggestion) RiskAssessment {
	risks := []Risk{}
	
	// Assess content integrity risk
	highEffortCount := 0
	for _, suggestion := range suggestions {
		if suggestion.Effort == EffortHigh {
			highEffortCount++
		}
	}
	
	if highEffortCount > 3 {
		risks = append(risks, Risk{
			Type:        RiskContentIntegrity,
			Description: "Multiple high-effort changes may affect content integrity",
			Impact:      RiskMedium,
			Likelihood:  RiskMedium,
			Mitigation:  "Implement changes incrementally and review after each change",
		})
	}

	// Determine overall risk
	overallRisk := RiskLow
	if len(risks) > 2 {
		overallRisk = RiskMedium
	}

	return RiskAssessment{
		OverallRisk:   overallRisk,
		SpecificRisks: risks,
		MitigationSteps: []string{
			"Implement suggestions in priority order",
			"Test changes incrementally",
			"Maintain backup of original content",
		},
	}
}

// createImplementationPlan creates a step-by-step implementation plan
func (ie *ImprovementEngine) createImplementationPlan(suggestions []ImprovementSuggestion) []ImplementationStep {
	steps := []ImplementationStep{}
	
	// Phase 1: Quick fixes (low effort, high impact)
	quickFixes := []string{}
	for _, suggestion := range suggestions {
		if suggestion.Effort == EffortLow && suggestion.ExpectedGain >= 5.0 {
			quickFixes = append(quickFixes, suggestion.ID)
		}
	}
	
	if len(quickFixes) > 0 {
		steps = append(steps, ImplementationStep{
			Step:          1,
			Phase:         PhaseQuickFixes,
			Title:         "Quick Fixes",
			Description:   "Implement low-effort, high-impact improvements",
			Suggestions:   quickFixes,
			EstimatedTime: time.Duration(len(quickFixes)) * 30 * time.Minute,
		})
	}

	// Phase 2: Content revision (medium effort improvements)
	contentRevisions := []string{}
	for _, suggestion := range suggestions {
		if suggestion.Effort == EffortMedium {
			contentRevisions = append(contentRevisions, suggestion.ID)
		}
	}
	
	if len(contentRevisions) > 0 {
		steps = append(steps, ImplementationStep{
			Step:          2,
			Phase:         PhaseContentRevision,
			Title:         "Content Revision",
			Description:   "Implement medium-effort content improvements",
			Suggestions:   contentRevisions,
			EstimatedTime: time.Duration(len(contentRevisions)) * 2 * time.Hour,
			Dependencies:  []int{1},
		})
	}

	// Phase 3: Structural changes (high effort improvements)
	structuralChanges := []string{}
	for _, suggestion := range suggestions {
		if suggestion.Effort == EffortHigh {
			structuralChanges = append(structuralChanges, suggestion.ID)
		}
	}
	
	if len(structuralChanges) > 0 {
		steps = append(steps, ImplementationStep{
			Step:          3,
			Phase:         PhaseStructuralChanges,
			Title:         "Structural Changes",
			Description:   "Implement high-effort structural improvements",
			Suggestions:   structuralChanges,
			EstimatedTime: time.Duration(len(structuralChanges)) * 4 * time.Hour,
			Dependencies:  []int{1, 2},
		})
	}

	return steps
}

// Parsing methods

// parseSuggestions parses suggestions from LLM response
func (ie *ImprovementEngine) parseSuggestions(response string, weakness Weakness) ([]ImprovementSuggestion, error) {
	var result struct {
		Suggestions []struct {
			Title          string    `json:"title"`
			Description    string    `json:"description"`
			Rationale      string    `json:"rationale"`
			Implementation string    `json:"implementation"`
			Examples       []Example `json:"examples"`
			ExpectedGain   float64   `json:"expectedGain"`
			Effort         string    `json:"effort"`
		} `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return []ImprovementSuggestion{}, nil
	}

	suggestions := []ImprovementSuggestion{}
	for i, s := range result.Suggestions {
		suggestion := ImprovementSuggestion{
			ID:              fmt.Sprintf("%s_%d", weakness.Type, i+1),
			Priority:        ie.determinePriority(s.ExpectedGain, s.Effort),
			Category:        ie.mapToCategory(weakness.Type),
			Title:           s.Title,
			Description:     s.Description,
			Rationale:       s.Rationale,
			Implementation:  s.Implementation,
			Examples:        s.Examples,
			ExpectedGain:    s.ExpectedGain,
			Effort:          ie.mapEffort(s.Effort),
			Tags:            []string{string(weakness.Type)},
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// parseFocusedSuggestions parses focused suggestions from LLM response
func (ie *ImprovementEngine) parseFocusedSuggestions(response string) ([]ImprovementSuggestion, error) {
	var result struct {
		Suggestions []struct {
			Title          string    `json:"title"`
			Description    string    `json:"description"`
			Rationale      string    `json:"rationale"`
			Implementation string    `json:"implementation"`
			Examples       []Example `json:"examples"`
			ExpectedGain   float64   `json:"expectedGain"`
			Effort         string    `json:"effort"`
			Tags           []string  `json:"tags"`
		} `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return []ImprovementSuggestion{}, nil
	}

	suggestions := []ImprovementSuggestion{}
	for i, s := range result.Suggestions {
		suggestion := ImprovementSuggestion{
			ID:              fmt.Sprintf("focused_%d", i+1),
			Priority:        ie.determinePriority(s.ExpectedGain, s.Effort),
			Category:        CategoryOptimization, // Default for focused suggestions
			Title:           s.Title,
			Description:     s.Description,
			Rationale:       s.Rationale,
			Implementation:  s.Implementation,
			Examples:        s.Examples,
			ExpectedGain:    s.ExpectedGain,
			Effort:          ie.mapEffort(s.Effort),
			Tags:            s.Tags,
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// Helper mapping functions

// determinePriority determines priority based on expected gain and effort
func (ie *ImprovementEngine) determinePriority(gain float64, effort string) ImprovementPriority {
	if gain >= 15.0 {
		return PriorityEssential
	} else if gain >= 10.0 && effort == "low" {
		return PriorityQuickWin
	} else if gain >= 8.0 {
		return PriorityImportant
	}
	return PriorityNiceToHave
}

// mapToCategory maps weakness type to improvement category
func (ie *ImprovementEngine) mapToCategory(weaknessType WeaknessType) ImprovementCategory {
	mapping := map[WeaknessType]ImprovementCategory{
		WeaknessReadability:      CategoryLanguage,
		WeaknessEngagement:       CategoryEngagement,
		WeaknessStructure:        CategoryStructure,
		WeaknessClarity:          CategoryLanguage,
		WeaknessFactualAccuracy:  CategoryContent,
		WeaknessStyleConsistency: CategoryLanguage,
		WeaknessSEO:              CategorySEO,
		WeaknessCredibility:      CategoryCredibility,
	}
	
	if category, exists := mapping[weaknessType]; exists {
		return category
	}
	return CategoryContent
}

// mapEffort maps string effort to EffortLevel
func (ie *ImprovementEngine) mapEffort(effort string) EffortLevel {
	switch effort {
	case "low":
		return EffortLow
	case "medium":
		return EffortMedium
	case "high":
		return EffortHigh
	default:
		return EffortMedium
	}
}
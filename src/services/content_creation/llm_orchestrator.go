package content_creation

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// LLMOrchestrator manages the LLM interactions and context
type LLMOrchestrator struct {
	LLMClient         LLMClient
	ContextManager    ContextManager
	TemplateManager   *PromptTemplateManager
	MetricsCollector  *MetricsCollector
}

// NewLLMOrchestrator creates a new LLM orchestrator
func NewLLMOrchestrator(
	client LLMClient,
	contextManager ContextManager,
	templateManager *PromptTemplateManager,
) *LLMOrchestrator {
	return &LLMOrchestrator{
		LLMClient:        client,
		ContextManager:   contextManager,
		TemplateManager:  templateManager,
		MetricsCollector: NewMetricsCollector(),
	}
}

// GenerateContent generates content using the LLM with managed context
func (o *LLMOrchestrator) GenerateContent(
	ctx context.Context,
	projectID uuid.UUID,
	contentType entities.ContentType,
	stage string,
	data PromptData,
) (string, error) {
	startTime := time.Now()
	
	// Get or initialize context
	err := o.ContextManager.SwitchContext(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to switch context: %w", err)
	}
	
	// Generate the prompt using the template
	prompt, err := o.TemplateManager.GeneratePrompt(contentType, stage, data)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}
	
	// Add the prompt to the context
	promptEntry := ContextEntry{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
		Priority:  5, // Medium priority
		Metadata: map[string]interface{}{
			"stage":       stage,
			"contentType": contentType,
		},
	}
	err = o.ContextManager.AddEntry(ctx, projectID, promptEntry)
	if err != nil {
		return "", fmt.Errorf("failed to add prompt to context: %w", err)
	}
	
	// Get the full context window
	contextWindow, err := o.ContextManager.GetContext(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}
	
	// Prepare the context for the LLM
	var messages []string
	for _, entry := range contextWindow.Entries {
		messages = append(messages, entry.Content)
	}
	
	// Generate the content
	response, err := o.LLMClient.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}
	
	// Add the response to the context
	responseEntry := ContextEntry{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
		Priority:  8, // High priority for model responses
		Metadata: map[string]interface{}{
			"stage":       stage,
			"contentType": contentType,
			"tokenCount":  estimateTokens(response),
		},
	}
	err = o.ContextManager.AddEntry(ctx, projectID, responseEntry)
	if err != nil {
		return "", fmt.Errorf("failed to add response to context: %w", err)
	}
	
	// Collect metrics
	o.MetricsCollector.RecordGeneration(
		contentType,
		stage,
		time.Since(startTime),
		estimateTokens(prompt),
		estimateTokens(response),
	)
	
	return response, nil
}

// InjectClientContext adds client-specific context to the context window
func (o *LLMOrchestrator) InjectClientContext(
	ctx context.Context,
	projectID uuid.UUID,
	clientProfile *entities.ClientProfile,
) error {
	// Create a domain knowledge map from the client profile
	knowledge := map[string]interface{}{
		"industry":       clientProfile.Industry,
		"brandVoice":     clientProfile.BrandVoice,
		"targetAudience": clientProfile.TargetAudience,
		"contentGoals":   clientProfile.ContentGoals,
	}
	
	// Add style preferences
	for k, v := range clientProfile.StylePreferences {
		knowledge["style_"+k] = v
	}
	
	// Inject the knowledge
	return o.ContextManager.InjectDomainKnowledge(ctx, projectID, knowledge)
}

// MetricsCollector collects metrics about LLM usage
type MetricsCollector struct {
	Generations               int
	TotalLatency              time.Duration
	TotalPromptTokens         int
	TotalResponseTokens       int
	GenerationsByContentType  map[entities.ContentType]int
	GenerationsByStage        map[string]int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		GenerationsByContentType: make(map[entities.ContentType]int),
		GenerationsByStage:       make(map[string]int),
	}
}

// RecordGeneration records metrics for a content generation
func (m *MetricsCollector) RecordGeneration(
	contentType entities.ContentType,
	stage string,
	latency time.Duration,
	promptTokens int,
	responseTokens int,
) {
	m.Generations++
	m.TotalLatency += latency
	m.TotalPromptTokens += promptTokens
	m.TotalResponseTokens += responseTokens
	m.GenerationsByContentType[contentType]++
	m.GenerationsByStage[stage]++
}

// GetMetrics returns the collected metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	avgLatency := float64(0)
	if m.Generations > 0 {
		avgLatency = float64(m.TotalLatency) / float64(m.Generations) / float64(time.Millisecond)
	}
	
	return map[string]interface{}{
		"totalGenerations":     m.Generations,
		"averageLatencyMs":     avgLatency,
		"totalPromptTokens":    m.TotalPromptTokens,
		"totalResponseTokens":  m.TotalResponseTokens,
		"totalTokens":          m.TotalPromptTokens + m.TotalResponseTokens,
		"contentTypeBreakdown": m.GenerationsByContentType,
		"stageBreakdown":       m.GenerationsByStage,
	}
}
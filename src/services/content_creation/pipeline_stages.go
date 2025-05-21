package content_creation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// researchStage conducts research for the content
func (p *ContentPipeline) researchStage(ctx context.Context, content *entities.Content) (*StageResult, error) {
	startTime := time.Now()

	// Get project to understand context
	project, err := p.projectRepo.FindByID(ctx, content.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Prepare research requirements
	requirements := ResearchRequirements{
		Topics:          []string{content.Title},
		TargetAudience:  getTargetAudienceFromProject(project),
		Industry:        getIndustryFromProject(project),
		ContentPurpose:  getContentPurposeFromType(content.Type),
		MaxSources:      10,
		RequireRecent:   true,
		RequireCredible: true,
	}

	// Conduct research
	researchOutput, err := p.researcher.Research(ctx, content, requirements)
	if err != nil {
		return nil, fmt.Errorf("research failed: %w", err)
	}

	// Prepare metadata for the next stage
	metadata := map[string]interface{}{
		"topics":     researchOutput.Topics,
		"sources":    researchOutput.Sources,
		"keyFacts":   researchOutput.KeyFacts,
		"references": researchOutput.References,
		"summary":    researchOutput.Summary,
	}

	return &StageResult{
		Content:     researchOutput.Summary,
		Status:      "completed",
		Metadata:    metadata,
		ElapsedTime: time.Since(startTime),
	}, nil
}

// outliningStage creates a structured outline for the content
func (p *ContentPipeline) outliningStage(ctx context.Context, content *entities.Content) (*StageResult, error) {
	startTime := time.Now()

	// Get research data from content metadata
	researchData, exists := content.Metadata["research"]
	if !exists {
		return nil, fmt.Errorf("research data not found in content metadata")
	}

	// Prepare prompt data for outlining
	promptData := PromptData{
		ContentTitle:   content.Title,
		ContentType:    content.Type,
		AdditionalContext: map[string]interface{}{
			"research": researchData,
		},
	}

	// Get project context for more personalized outlining
	project, err := p.projectRepo.FindByID(ctx, content.ProjectID)
	if err == nil {
		promptData.ProjectTitle = project.Title
		promptData.TargetAudience = getTargetAudienceFromProject(project)
		promptData.BrandVoice = getBrandVoiceFromProject(project)
		promptData.ContentGoals = getContentGoalsFromProject(project)
	}

	// Switch to project context
	err = p.contextManager.SwitchContext(ctx, content.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to switch context: %w", err)
	}

	// Generate outline using template
	templateManager := NewPromptTemplateManager()
	prompt, err := templateManager.GeneratePrompt(content.Type, "outline", promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate outline prompt: %w", err)
	}

	// Generate outline using LLM
	outline, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("outline generation failed: %w", err)
	}

	// Add outline to context
	contextEntry := ContextEntry{
		Role:      "assistant",
		Content:   outline,
		Timestamp: time.Now(),
		Priority:  7,
		Metadata: map[string]interface{}{
			"stage":       "outline",
			"contentType": content.Type,
		},
	}
	err = p.contextManager.AddEntry(ctx, content.ProjectID, contextEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to add outline to context: %w", err)
	}

	return &StageResult{
		Content:     outline,
		Status:      "completed",
		Metadata:    map[string]interface{}{"stage": "outline"},
		ElapsedTime: time.Since(startTime),
	}, nil
}

// draftingStage creates the initial draft of the content
func (p *ContentPipeline) draftingStage(ctx context.Context, content *entities.Content) (*StageResult, error) {
	startTime := time.Now()

	// Get outline from content metadata
	outline, exists := content.Metadata["outline"]
	if !exists {
		return nil, fmt.Errorf("outline not found in content metadata")
	}

	// Get research data
	researchData, _ := content.Metadata["research"]

	// Prepare prompt data for drafting
	promptData := PromptData{
		ContentTitle: content.Title,
		ContentType:  content.Type,
		AdditionalContext: map[string]interface{}{
			"Outline":  outline,
			"research": researchData,
		},
	}

	// Get project context
	project, err := p.projectRepo.FindByID(ctx, content.ProjectID)
	if err == nil {
		promptData.ProjectTitle = project.Title
		promptData.TargetAudience = getTargetAudienceFromProject(project)
		promptData.BrandVoice = getBrandVoiceFromProject(project)
		promptData.ContentGoals = getContentGoalsFromProject(project)
		promptData.Keywords = getKeywordsFromProject(project)
	}

	// Generate draft using template
	templateManager := NewPromptTemplateManager()
	prompt, err := templateManager.GeneratePrompt(content.Type, "draft", promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate draft prompt: %w", err)
	}

	// Generate draft using LLM
	draft, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("draft generation failed: %w", err)
	}

	// Add draft to context
	contextEntry := ContextEntry{
		Role:      "assistant",
		Content:   draft,
		Timestamp: time.Now(),
		Priority:  8,
		Metadata: map[string]interface{}{
			"stage":       "draft",
			"contentType": content.Type,
			"wordCount":   estimateWords(draft),
		},
	}
	err = p.contextManager.AddEntry(ctx, content.ProjectID, contextEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to add draft to context: %w", err)
	}

	return &StageResult{
		Content:     draft,
		Status:      "completed",
		Metadata:    map[string]interface{}{"stage": "draft", "wordCount": estimateWords(draft)},
		ElapsedTime: time.Since(startTime),
	}, nil
}

// editingStage improves and refines the content
func (p *ContentPipeline) editingStage(ctx context.Context, content *entities.Content) (*StageResult, error) {
	startTime := time.Now()

	// Get the current content (draft)
	if content.Data == "" {
		return nil, fmt.Errorf("no draft content available for editing")
	}

	// Prepare prompt data for editing
	promptData := PromptData{
		ContentTitle: content.Title,
		ContentType:  content.Type,
		AdditionalContext: map[string]interface{}{
			"Draft": content.Data,
		},
	}

	// Get project context
	project, err := p.projectRepo.FindByID(ctx, content.ProjectID)
	if err == nil {
		promptData.ClientName = getClientNameFromProject(project)
		promptData.TargetAudience = getTargetAudienceFromProject(project)
		promptData.BrandVoice = getBrandVoiceFromProject(project)
		promptData.Keywords = getKeywordsFromProject(project)
	}

	// Generate editing prompt using template
	templateManager := NewPromptTemplateManager()
	prompt, err := templateManager.GeneratePrompt(content.Type, "edit", promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate edit prompt: %w", err)
	}

	// Generate edited content using LLM
	editedContent, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("editing failed: %w", err)
	}

	// Add edited content to context
	contextEntry := ContextEntry{
		Role:      "assistant",
		Content:   editedContent,
		Timestamp: time.Now(),
		Priority:  9,
		Metadata: map[string]interface{}{
			"stage":       "edit",
			"contentType": content.Type,
			"wordCount":   estimateWords(editedContent),
		},
	}
	err = p.contextManager.AddEntry(ctx, content.ProjectID, contextEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to add edited content to context: %w", err)
	}

	return &StageResult{
		Content:     editedContent,
		Status:      "completed",
		Metadata:    map[string]interface{}{"stage": "edit", "wordCount": estimateWords(editedContent)},
		ElapsedTime: time.Since(startTime),
	}, nil
}

// finalizationStage prepares the content for delivery
func (p *ContentPipeline) finalizationStage(ctx context.Context, content *entities.Content) (*StageResult, error) {
	startTime := time.Now()

	// Get the current content (edited version)
	if content.Data == "" {
		return nil, fmt.Errorf("no edited content available for finalization")
	}

	// Prepare prompt data for finalization
	promptData := PromptData{
		ContentTitle: content.Title,
		ContentType:  content.Type,
		AdditionalContext: map[string]interface{}{
			"EditedDraft": content.Data,
		},
	}

	// Get project context
	project, err := p.projectRepo.FindByID(ctx, content.ProjectID)
	if err == nil {
		promptData.ClientName = getClientNameFromProject(project)
		promptData.Keywords = getKeywordsFromProject(project)
	}

	// Get keywords from metadata if available
	if keywords, exists := content.Metadata["keywords"]; exists {
		if keywordSlice, ok := keywords.([]string); ok {
			promptData.Keywords = keywordSlice
		}
	}

	// Generate finalization prompt using template
	templateManager := NewPromptTemplateManager()
	prompt, err := templateManager.GeneratePrompt(content.Type, "finalize", promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate finalization prompt: %w", err)
	}

	// Generate final content using LLM
	finalContent, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("finalization failed: %w", err)
	}

	// Add delivery metadata
	deliveryMetadata := map[string]interface{}{
		"stage":          "finalized",
		"wordCount":      estimateWords(finalContent),
		"deliveryReady":  true,
		"finalizedAt":    time.Now(),
		"contentFormat":  getContentFormat(content.Type),
	}

	// Add final content to context
	contextEntry := ContextEntry{
		Role:      "assistant",
		Content:   finalContent,
		Timestamp: time.Now(),
		Priority:  10, // Highest priority for final content
		Metadata: deliveryMetadata,
	}
	err = p.contextManager.AddEntry(ctx, content.ProjectID, contextEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to add final content to context: %w", err)
	}

	return &StageResult{
		Content:     finalContent,
		Status:      "completed",
		Metadata:    deliveryMetadata,
		ElapsedTime: time.Since(startTime),
	}, nil
}

// Helper functions to extract project information

func getTargetAudienceFromProject(project *entities.Project) string {
	if project.Description != "" {
		// In a real implementation, this would parse or extract audience info
		return "target audience" // Placeholder
	}
	return "general audience"
}

func getIndustryFromProject(project *entities.Project) string {
	// In a real implementation, this would extract industry info from project
	return "general"
}

func getContentPurposeFromType(contentType entities.ContentType) string {
	switch contentType {
	case entities.ContentTypeBlogPost:
		return "informational and engaging blog content"
	case entities.ContentTypeSocialPost:
		return "engaging social media content"
	case entities.ContentTypeTechnicalArticle:
		return "detailed technical information"
	case entities.ContentTypeEmailNewsletter:
		return "newsletter communication"
	case entities.ContentTypeWebsiteCopy:
		return "website marketing copy"
	case entities.ContentTypeProductDescription:
		return "product marketing description"
	case entities.ContentTypePressRelease:
		return "press and media communication"
	default:
		return "general content creation"
	}
}

func getBrandVoiceFromProject(project *entities.Project) string {
	// In a real implementation, this would extract brand voice from project or client profile
	return "professional"
}

func getContentGoalsFromProject(project *entities.Project) []string {
	// In a real implementation, this would extract goals from project
	return []string{"inform", "engage", "convert"}
}

func getKeywordsFromProject(project *entities.Project) []string {
	// In a real implementation, this would extract keywords from project
	return []string{}
}

func getClientNameFromProject(project *entities.Project) string {
	// In a real implementation, this would get client name from the client entity
	return "Client"
}

func getContentFormat(contentType entities.ContentType) string {
	switch contentType {
	case entities.ContentTypeBlogPost:
		return "HTML/Markdown"
	case entities.ContentTypeSocialPost:
		return "Plain Text"
	case entities.ContentTypeTechnicalArticle:
		return "Markdown"
	case entities.ContentTypeEmailNewsletter:
		return "HTML"
	case entities.ContentTypeWebsiteCopy:
		return "HTML"
	case entities.ContentTypeProductDescription:
		return "HTML/Plain Text"
	case entities.ContentTypePressRelease:
		return "Plain Text"
	default:
		return "Plain Text"
	}
}

func estimateWords(text string) int {
	if text == "" {
		return 0
	}
	// Simple word count estimation
	return len(strings.Fields(text))
}
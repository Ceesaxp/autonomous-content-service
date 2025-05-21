package content_creation

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// PipelineStage represents a stage in the content creation pipeline
type PipelineStage string

const (
	StageResearch     PipelineStage = "research"
	StageOutlining    PipelineStage = "outlining"
	StageDrafting     PipelineStage = "drafting"
	StageEditing      PipelineStage = "editing"
	StageFinalization PipelineStage = "finalization"
)

// PipelineProgressEvent represents a progress event in the content creation pipeline
type PipelineProgressEvent struct {
	ContentID   uuid.UUID     `json:"contentId"`
	ProjectID   uuid.UUID     `json:"projectId"`
	Stage       PipelineStage `json:"stage"`
	Status      string        `json:"status"` // started, completed, failed
	TimeElapsed time.Duration `json:"timeElapsed"`
	Details     string        `json:"details,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// PipelineConfig contains configuration options for the content pipeline
type PipelineConfig struct {
	MaxRetries            int  `json:"maxRetries"`
	ContextWindowSize     int  `json:"contextWindowSize"`
	EnableFactChecking    bool `json:"enableFactChecking"`
	EnablePlagiarismCheck bool `json:"enablePlagiarismCheck"`
	SEOOptimization       bool `json:"seoOptimization"`
	StageTimeoutSeconds   int  `json:"stageTimeoutSeconds"`
}

// StageResult contains the result of executing a pipeline stage
type StageResult struct {
	Content     string                 `json:"content"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ElapsedTime time.Duration          `json:"elapsedTime"`
	Error       error                  `json:"error,omitempty"`
}

// ContentPipeline orchestrates the content creation process
type ContentPipeline struct {
	contentRepo        repositories.ContentRepository
	ContentVersionRepo repositories.ContentVersionRepository
	projectRepo        repositories.ProjectRepository
	eventRepo          repositories.EventRepository
	llmClient          LLMClient
	contextManager     ContextManager
	researcher         Researcher
	qualityChecker     QualityChecker
	config             PipelineConfig
}

// NewContentPipeline creates a new content creation pipeline
func NewContentPipeline(
	contentRepo repositories.ContentRepository,
	contentVersionRepo repositories.ContentVersionRepository,
	projectRepo repositories.ProjectRepository,
	eventRepo repositories.EventRepository,
	llmClient LLMClient,
	contextManager ContextManager,
	researcher Researcher,
	qualityChecker QualityChecker,
	config PipelineConfig,
) *ContentPipeline {
	return &ContentPipeline{
		contentRepo:        contentRepo,
		ContentVersionRepo: contentVersionRepo,
		projectRepo:        projectRepo,
		eventRepo:          eventRepo,
		llmClient:          llmClient,
		contextManager:     contextManager,
		researcher:         researcher,
		qualityChecker:     qualityChecker,
		config:             config,
	}
}

// CreateContent orchestrates the full content creation process
func (p *ContentPipeline) CreateContent(ctx context.Context, projectID uuid.UUID, title string, contentType entities.ContentType) (*entities.Content, error) {
	startTime := time.Now()

	// Create the content entity
	content, err := entities.NewContent(projectID, title, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create content entity: %w", err)
	}

	// Persist the initial content
	err = p.contentRepo.Create(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to persist initial content: %w", err)
	}

	// Record event
	p.recordEvent(ctx, content.ContentID, content.ProjectID, StageResearch, "started", time.Since(startTime), "Starting content creation process")

	// Run the pipeline
	err = p.executePipeline(ctx, content)
	if err != nil {
		// Update content status to indicate error
		content.UpdateStatus(entities.ContentStatusPlanning)
		content.UpdateMetadata("error", err.Error())
		p.contentRepo.Update(ctx, content)

		// Record event
		p.recordEvent(ctx, content.ContentID, content.ProjectID, StageResearch, "failed", time.Since(startTime), fmt.Sprintf("Pipeline failed: %v", err))

		return content, fmt.Errorf("pipeline execution failed: %w", err)
	}

	// Record final event
	p.recordEvent(ctx, content.ContentID, content.ProjectID, StageFinalization, "completed", time.Since(startTime), "Content creation completed successfully")

	return content, nil
}

// executePipeline runs the content through all pipeline stages
func (p *ContentPipeline) executePipeline(ctx context.Context, content *entities.Content) error {
	// Research stage
	content.UpdateStatus(entities.ContentStatusResearching)
	p.contentRepo.Update(ctx, content)

	researchResult, err := p.executeStage(ctx, content, StageResearch)
	if err != nil {
		return fmt.Errorf("research stage failed: %w", err)
	}

	// Store research data in content metadata
	content.UpdateMetadata("research", researchResult.Metadata)
	p.contentRepo.Update(ctx, content)

	// Outlining stage
	outlineResult, err := p.executeStage(ctx, content, StageOutlining)
	if err != nil {
		return fmt.Errorf("outlining stage failed: %w", err)
	}

	// Store outline in content metadata and update content
	content.UpdateMetadata("outline", outlineResult.Content)
	p.contentRepo.Update(ctx, content)

	// Drafting stage
	content.UpdateStatus(entities.ContentStatusDrafting)
	p.contentRepo.Update(ctx, content)

	draftResult, err := p.executeStage(ctx, content, StageDrafting)
	if err != nil {
		return fmt.Errorf("drafting stage failed: %w", err)
	}

	// Update content with draft
	err = content.UpdateContent(draftResult.Content, string(StageDrafting))
	if err != nil {
		return fmt.Errorf("failed to update content with draft: %w", err)
	}
	p.contentRepo.Update(ctx, content)

	// Editing stage
	content.UpdateStatus(entities.ContentStatusEditing)
	p.contentRepo.Update(ctx, content)

	editResult, err := p.executeStage(ctx, content, StageEditing)
	if err != nil {
		return fmt.Errorf("editing stage failed: %w", err)
	}

	// Update content with edited version
	err = content.UpdateContent(editResult.Content, string(StageEditing))
	if err != nil {
		return fmt.Errorf("failed to update content with edited version: %w", err)
	}
	p.contentRepo.Update(ctx, content)

	// Quality check
	qualityInput := QualityCheckInput{
		Content:           editResult.Content,
		CheckPlagiarism:   p.config.EnablePlagiarismCheck,
		CheckFactAccuracy: p.config.EnableFactChecking,
		EvaluateSEO:       p.config.SEOOptimization,
	}

	qualityOutput, err := p.qualityChecker.CheckContent(ctx, content, qualityInput)
	if err != nil {
		// Log but don't fail the pipeline
		fmt.Printf("Warning: Quality check encountered errors: %v\n", err)
	} else {
		// Update content statistics
		content.UpdateStatistics(entities.ContentStatistics{
			ReadabilityScore: qualityOutput.ReadabilityScore,
			SEOScore:         qualityOutput.SEOScore,
			EngagementScore:  qualityOutput.EngagementScore,
			PlagiarismScore:  qualityOutput.PlagiarismScore,
		})

		// Add suggestions to metadata
		content.UpdateMetadata("qualitySuggestions", qualityOutput.SuggestionsByCategory)
		content.UpdateMetadata("keywords", qualityOutput.Keywords)
		p.contentRepo.Update(ctx, content)
	}

	// Finalization stage
	finalResult, err := p.executeStage(ctx, content, StageFinalization)
	if err != nil {
		return fmt.Errorf("finalization stage failed: %w", err)
	}

	// Update content with final version
	err = content.UpdateContent(finalResult.Content, string(StageFinalization))
	if err != nil {
		return fmt.Errorf("failed to update content with final version: %w", err)
	}

	// Update content status to review
	content.UpdateStatus(entities.ContentStatusReview)
	p.contentRepo.Update(ctx, content)

	return nil
}

// executeStage executes a specific pipeline stage with retry logic
func (p *ContentPipeline) executeStage(ctx context.Context, content *entities.Content, stage PipelineStage) (*StageResult, error) {
	var result *StageResult
	var err error
	var attemptCount int

	stageTimeout := time.Duration(p.config.StageTimeoutSeconds) * time.Second
	if stageTimeout == 0 {
		stageTimeout = 60 * time.Second // default 60 second timeout
	}

	// Record start of stage
	startTime := time.Now()
	p.recordEvent(ctx, content.ContentID, content.ProjectID, stage, "started", 0, fmt.Sprintf("Starting %s stage", stage))

	// Execute stage with retries
	for attemptCount = 1; attemptCount <= p.config.MaxRetries; attemptCount++ {
		// Create a timeout context for this stage
		stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
		defer cancel()

		switch stage {
		case StageResearch:
			result, err = p.researchStage(stageCtx, content)
		case StageOutlining:
			result, err = p.outliningStage(stageCtx, content)
		case StageDrafting:
			result, err = p.draftingStage(stageCtx, content)
		case StageEditing:
			result, err = p.editingStage(stageCtx, content)
		case StageFinalization:
			result, err = p.finalizationStage(stageCtx, content)
		default:
			return nil, fmt.Errorf("unknown pipeline stage: %s", stage)
		}

		// Check if context timed out
		if stageCtx.Err() == context.DeadlineExceeded {
			p.recordEvent(ctx, content.ContentID, content.ProjectID, stage, "timeout", time.Since(startTime),
				fmt.Sprintf("Stage timed out after %v. Attempt %d of %d", stageTimeout, attemptCount, p.config.MaxRetries))

			// If we've exhausted retries, fail
			if attemptCount == p.config.MaxRetries {
				return nil, fmt.Errorf("stage %s timed out after %d attempts", stage, attemptCount)
			}

			// Otherwise retry
			continue
		}

		if err != nil {
			p.recordEvent(ctx, content.ContentID, content.ProjectID, stage, "error", time.Since(startTime),
				fmt.Sprintf("Stage error: %v. Attempt %d of %d", err, attemptCount, p.config.MaxRetries))

			// If we've exhausted retries, fail
			if attemptCount == p.config.MaxRetries {
				return nil, fmt.Errorf("stage %s failed after %d attempts: %w", stage, attemptCount, err)
			}

			// Otherwise retry
			time.Sleep(time.Duration(attemptCount) * time.Second) // Exponential backoff
			continue
		}

		// Success - break out of retry loop
		break
	}

	// Record completion of stage
	p.recordEvent(ctx, content.ContentID, content.ProjectID, stage, "completed", time.Since(startTime),
		fmt.Sprintf("Completed %s stage after %d attempts", stage, attemptCount))

	return result, nil
}

// recordEvent creates and stores an event for pipeline progress
func (p *ContentPipeline) recordEvent(ctx context.Context, contentID, projectID uuid.UUID, stage PipelineStage, status string, elapsed time.Duration, details string) {
	event := &events.ContentRequestedEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New(),
			EventType: fmt.Sprintf("pipeline.%s.%s", stage, status),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"stage":       stage,
				"status":      status,
				"timeElapsed": elapsed.String(),
				"details":     details,
			},
		},
		ContentID: contentID,
		ProjectID: projectID,
	}

	// Non-blocking event recording
	go func() {
		contextTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := p.eventRepo.Save(contextTimeout, event)
		if err != nil {
			// Just log the error and continue
			fmt.Printf("Failed to record pipeline event: %v\n", err)
		}
	}()
}

// The following methods will be implemented in separate files:
// - researchStage
// - outliningStage
// - draftingStage
// - editingStage
// - finalizationStage

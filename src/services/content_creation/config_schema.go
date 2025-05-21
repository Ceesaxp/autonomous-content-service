package content_creation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ContentTypeConfig defines configuration for a specific content type
type ContentTypeConfig struct {
	Type                entities.ContentType `json:"type"`
	MinWordCount        int                  `json:"minWordCount"`
	MaxWordCount        int                  `json:"maxWordCount"`
	RequiredStages      []PipelineStage      `json:"requiredStages"`
	OptionalStages      []PipelineStage      `json:"optionalStages"`
	QualityThresholds   QualityThresholds    `json:"qualityThresholds"`
	TimeoutPerStage     time.Duration        `json:"timeoutPerStage"`
	ResearchRequirements ResearchConfig      `json:"researchRequirements"`
}

// QualityThresholds defines minimum quality scores required
type QualityThresholds struct {
	MinReadabilityScore float64 `json:"minReadabilityScore"`
	MinSEOScore         float64 `json:"minSEOScore"`
	MinEngagementScore  float64 `json:"minEngagementScore"`
	MinPlagiarismScore  float64 `json:"minPlagiarismScore"`
}

// ResearchConfig defines research requirements for content types
type ResearchConfig struct {
	RequiredTopics    int  `json:"requiredTopics"`
	MinSources        int  `json:"minSources"`
	MaxSources        int  `json:"maxSources"`
	RequireCredible   bool `json:"requireCredible"`
	RequireRecent     bool `json:"requireRecent"`
	FactCheckRequired bool `json:"factCheckRequired"`
}

// StageConfig defines configuration for individual pipeline stages
type StageConfig struct {
	Stage           PipelineStage `json:"stage"`
	MaxRetries      int           `json:"maxRetries"`
	Timeout         time.Duration `json:"timeout"`
	RequiredInputs  []string      `json:"requiredInputs"`
	OptionalInputs  []string      `json:"optionalInputs"`
	OutputFormat    string        `json:"outputFormat"`
	QualityChecks   []string      `json:"qualityChecks"`
}

// PipelineConfigSchema defines the complete configuration schema
type PipelineConfigSchema struct {
	Version           string                           `json:"version"`
	DefaultConfig     PipelineConfig                   `json:"defaultConfig"`
	ContentTypeConfigs map[entities.ContentType]ContentTypeConfig `json:"contentTypeConfigs"`
	StageConfigs      map[PipelineStage]StageConfig    `json:"stageConfigs"`
	LLMConfigs        map[string]LLMConfig            `json:"llmConfigs"`
	QualityConfigs    QualityConfig                   `json:"qualityConfigs"`
}

// LLMConfig defines configuration for LLM clients
type LLMConfig struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"maxTokens"`
	Timeout      time.Duration `json:"timeout"`
	RetryPolicy  RetryConfig   `json:"retryPolicy"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries      int           `json:"maxRetries"`
	InitialDelay    time.Duration `json:"initialDelay"`
	BackoffMultiplier float64     `json:"backoffMultiplier"`
	MaxDelay        time.Duration `json:"maxDelay"`
}

// QualityConfig defines quality checking configuration
type QualityConfig struct {
	EnableReadabilityCheck bool                      `json:"enableReadabilityCheck"`
	EnableSEOCheck         bool                      `json:"enableSEOCheck"`
	EnableEngagementCheck  bool                      `json:"enableEngagementCheck"`
	EnablePlagiarismCheck  bool                      `json:"enablePlagiarismCheck"`
	EnableFactCheck        bool                      `json:"enableFactCheck"`
	ReadabilityTools      []string                  `json:"readabilityTools"`
	SEOTools              []string                  `json:"seoTools"`
	PlagiarismTools       []string                  `json:"plagiarismTools"`
	FactCheckTools        []string                  `json:"factCheckTools"`
	QualityThresholds     map[entities.ContentType]QualityThresholds `json:"qualityThresholds"`
}

// DefaultPipelineConfigSchema returns a default configuration schema
func DefaultPipelineConfigSchema() *PipelineConfigSchema {
	return &PipelineConfigSchema{
		Version: "1.0.0",
		DefaultConfig: PipelineConfig{
			MaxRetries:           3,
			ContextWindowSize:    8192,
			EnableFactChecking:   true,
			EnablePlagiarismCheck: true,
			SEOOptimization:      true,
			StageTimeoutSeconds:  300, // 5 minutes per stage
		},
		ContentTypeConfigs: map[entities.ContentType]ContentTypeConfig{
			entities.ContentTypeBlogPost: {
				Type:         entities.ContentTypeBlogPost,
				MinWordCount: 800,
				MaxWordCount: 3000,
				RequiredStages: []PipelineStage{
					StageResearch, StageOutlining, StageDrafting, StageEditing, StageFinalization,
				},
				OptionalStages: []PipelineStage{},
				QualityThresholds: QualityThresholds{
					MinReadabilityScore: 70.0,
					MinSEOScore:         60.0,
					MinEngagementScore:  70.0,
					MinPlagiarismScore:  0.85,
				},
				TimeoutPerStage: 300 * time.Second,
				ResearchRequirements: ResearchConfig{
					RequiredTopics:    3,
					MinSources:        5,
					MaxSources:        15,
					RequireCredible:   true,
					RequireRecent:     true,
					FactCheckRequired: true,
				},
			},
			entities.ContentTypeSocialPost: {
				Type:         entities.ContentTypeSocialPost,
				MinWordCount: 10,
				MaxWordCount: 280,
				RequiredStages: []PipelineStage{
					StageResearch, StageDrafting, StageEditing,
				},
				OptionalStages: []PipelineStage{StageOutlining, StageFinalization},
				QualityThresholds: QualityThresholds{
					MinReadabilityScore: 80.0,
					MinSEOScore:         50.0,
					MinEngagementScore:  85.0,
					MinPlagiarismScore:  0.90,
				},
				TimeoutPerStage: 120 * time.Second,
				ResearchRequirements: ResearchConfig{
					RequiredTopics:    1,
					MinSources:        2,
					MaxSources:        5,
					RequireCredible:   false,
					RequireRecent:     true,
					FactCheckRequired: false,
				},
			},
			entities.ContentTypeTechnicalArticle: {
				Type:         entities.ContentTypeTechnicalArticle,
				MinWordCount: 1200,
				MaxWordCount: 5000,
				RequiredStages: []PipelineStage{
					StageResearch, StageOutlining, StageDrafting, StageEditing, StageFinalization,
				},
				OptionalStages: []PipelineStage{},
				QualityThresholds: QualityThresholds{
					MinReadabilityScore: 60.0,
					MinSEOScore:         65.0,
					MinEngagementScore:  65.0,
					MinPlagiarismScore:  0.90,
				},
				TimeoutPerStage: 450 * time.Second,
				ResearchRequirements: ResearchConfig{
					RequiredTopics:    5,
					MinSources:        10,
					MaxSources:        25,
					RequireCredible:   true,
					RequireRecent:     true,
					FactCheckRequired: true,
				},
			},
			entities.ContentTypeEmailNewsletter: {
				Type:         entities.ContentTypeEmailNewsletter,
				MinWordCount: 400,
				MaxWordCount: 1200,
				RequiredStages: []PipelineStage{
					StageResearch, StageOutlining, StageDrafting, StageEditing, StageFinalization,
				},
				OptionalStages: []PipelineStage{},
				QualityThresholds: QualityThresholds{
					MinReadabilityScore: 75.0,
					MinSEOScore:         40.0,
					MinEngagementScore:  80.0,
					MinPlagiarismScore:  0.85,
				},
				TimeoutPerStage: 240 * time.Second,
				ResearchRequirements: ResearchConfig{
					RequiredTopics:    2,
					MinSources:        3,
					MaxSources:        8,
					RequireCredible:   true,
					RequireRecent:     true,
					FactCheckRequired: true,
				},
			},
		},
		StageConfigs: map[PipelineStage]StageConfig{
			StageResearch: {
				Stage:      StageResearch,
				MaxRetries: 3,
				Timeout:    300 * time.Second,
				RequiredInputs: []string{"title", "contentType", "targetAudience"},
				OptionalInputs: []string{"industry", "keywords", "brandVoice"},
				OutputFormat:   "ResearchOutput",
				QualityChecks:  []string{"sourceCredibility", "topicRelevance"},
			},
			StageOutlining: {
				Stage:      StageOutlining,
				MaxRetries: 3,
				Timeout:    180 * time.Second,
				RequiredInputs: []string{"research", "title", "contentType"},
				OptionalInputs: []string{"brandVoice", "targetAudience"},
				OutputFormat:   "StructuredOutline",
				QualityChecks:  []string{"logicalFlow", "completeness"},
			},
			StageDrafting: {
				Stage:      StageDrafting,
				MaxRetries: 3,
				Timeout:    360 * time.Second,
				RequiredInputs: []string{"outline", "research", "title"},
				OptionalInputs: []string{"keywords", "brandVoice"},
				OutputFormat:   "FormattedContent",
				QualityChecks:  []string{"wordCount", "coherence", "styleConsistency"},
			},
			StageEditing: {
				Stage:      StageEditing,
				MaxRetries: 3,
				Timeout:    240 * time.Second,
				RequiredInputs: []string{"draft", "title", "contentType"},
				OptionalInputs: []string{"keywords", "styleGuide"},
				OutputFormat:   "EditedContent",
				QualityChecks:  []string{"grammar", "readability", "factAccuracy"},
			},
			StageFinalization: {
				Stage:      StageFinalization,
				MaxRetries: 2,
				Timeout:    120 * time.Second,
				RequiredInputs: []string{"editedContent", "contentType"},
				OptionalInputs: []string{"keywords", "deliveryFormat"},
				OutputFormat:   "PublicationReady",
				QualityChecks:  []string{"formatting", "metadata", "deliveryStandards"},
			},
		},
		LLMConfigs: map[string]LLMConfig{
			"default": {
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2048,
				Timeout:     60 * time.Second,
				RetryPolicy: RetryConfig{
					MaxRetries:        3,
					InitialDelay:      2 * time.Second,
					BackoffMultiplier: 2.0,
					MaxDelay:          30 * time.Second,
				},
			},
			"creative": {
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.9,
				MaxTokens:   2048,
				Timeout:     60 * time.Second,
				RetryPolicy: RetryConfig{
					MaxRetries:        3,
					InitialDelay:      2 * time.Second,
					BackoffMultiplier: 2.0,
					MaxDelay:          30 * time.Second,
				},
			},
			"technical": {
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.3,
				MaxTokens:   3072,
				Timeout:     90 * time.Second,
				RetryPolicy: RetryConfig{
					MaxRetries:        3,
					InitialDelay:      2 * time.Second,
					BackoffMultiplier: 2.0,
					MaxDelay:          30 * time.Second,
				},
			},
		},
		QualityConfigs: QualityConfig{
			EnableReadabilityCheck: true,
			EnableSEOCheck:         true,
			EnableEngagementCheck:  true,
			EnablePlagiarismCheck:  true,
			EnableFactCheck:        true,
			ReadabilityTools:       []string{"flesch-kincaid", "coleman-liau"},
			SEOTools:               []string{"keyword-density", "meta-analysis"},
			PlagiarismTools:        []string{"similarity-check", "source-verification"},
			FactCheckTools:         []string{"llm-verification", "source-validation"},
			QualityThresholds: map[entities.ContentType]QualityThresholds{
				entities.ContentTypeBlogPost: {
					MinReadabilityScore: 70.0,
					MinSEOScore:         60.0,
					MinEngagementScore:  70.0,
					MinPlagiarismScore:  0.85,
				},
				entities.ContentTypeSocialPost: {
					MinReadabilityScore: 80.0,
					MinSEOScore:         50.0,
					MinEngagementScore:  85.0,
					MinPlagiarismScore:  0.90,
				},
			},
		},
	}
}

// LoadConfigFromJSON loads configuration from JSON
func LoadConfigFromJSON(jsonData []byte) (*PipelineConfigSchema, error) {
	var config PipelineConfigSchema
	err := json.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config JSON: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// ToJSON converts the configuration to JSON
func (c *PipelineConfigSchema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Validate validates the configuration schema
func (c *PipelineConfigSchema) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	// Validate content type configs
	for contentType, config := range c.ContentTypeConfigs {
		if config.Type != contentType {
			return fmt.Errorf("content type mismatch for %s", contentType)
		}
		
		if config.MinWordCount < 0 {
			return fmt.Errorf("minWordCount cannot be negative for %s", contentType)
		}
		
		if config.MaxWordCount > 0 && config.MinWordCount > config.MaxWordCount {
			return fmt.Errorf("minWordCount cannot be greater than maxWordCount for %s", contentType)
		}
		
		if len(config.RequiredStages) == 0 {
			return fmt.Errorf("at least one required stage must be specified for %s", contentType)
		}
	}
	
	// Validate stage configs
	for stage, config := range c.StageConfigs {
		if config.Stage != stage {
			return fmt.Errorf("stage mismatch for %s", stage)
		}
		
		if config.MaxRetries < 0 {
			return fmt.Errorf("maxRetries cannot be negative for stage %s", stage)
		}
		
		if config.Timeout <= 0 {
			return fmt.Errorf("timeout must be positive for stage %s", stage)
		}
	}
	
	// Validate LLM configs
	for name, config := range c.LLMConfigs {
		if config.Provider == "" {
			return fmt.Errorf("provider is required for LLM config %s", name)
		}
		
		if config.Model == "" {
			return fmt.Errorf("model is required for LLM config %s", name)
		}
		
		if config.Temperature < 0 || config.Temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2 for LLM config %s", name)
		}
		
		if config.MaxTokens <= 0 {
			return fmt.Errorf("maxTokens must be positive for LLM config %s", name)
		}
	}
	
	return nil
}

// GetContentTypeConfig returns configuration for a specific content type
func (c *PipelineConfigSchema) GetContentTypeConfig(contentType entities.ContentType) (ContentTypeConfig, bool) {
	config, exists := c.ContentTypeConfigs[contentType]
	return config, exists
}

// GetStageConfig returns configuration for a specific stage
func (c *PipelineConfigSchema) GetStageConfig(stage PipelineStage) (StageConfig, bool) {
	config, exists := c.StageConfigs[stage]
	return config, exists
}

// GetLLMConfig returns configuration for a specific LLM
func (c *PipelineConfigSchema) GetLLMConfig(name string) (LLMConfig, bool) {
	config, exists := c.LLMConfigs[name]
	return config, exists
}

// CreatePipelineConfig creates a PipelineConfig from the schema for a specific content type
func (c *PipelineConfigSchema) CreatePipelineConfig(contentType entities.ContentType) PipelineConfig {
	contentConfig, exists := c.GetContentTypeConfig(contentType)
	if !exists {
		// Return default config if content type not found
		return c.DefaultConfig
	}
	
	// Merge default config with content-specific config
	config := c.DefaultConfig
	config.StageTimeoutSeconds = int(contentConfig.TimeoutPerStage.Seconds())
	
	return config
}
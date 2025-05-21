package content_creation

import (
	"context"
	"testing"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockContentRepository struct {
	mock.Mock
}

func (m *MockContentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Content), args.Error(1)
}

func (m *MockContentRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.Content, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entities.Content), args.Error(1)
}

func (m *MockContentRepository) FindByStatus(ctx context.Context, status entities.ContentStatus, offset, limit int) ([]*entities.Content, int, error) {
	args := m.Called(ctx, status, offset, limit)
	return args.Get(0).([]*entities.Content), args.Int(1), args.Error(2)
}

func (m *MockContentRepository) FindByType(ctx context.Context, contentType entities.ContentType, offset, limit int) ([]*entities.Content, int, error) {
	args := m.Called(ctx, contentType, offset, limit)
	return args.Get(0).([]*entities.Content), args.Int(1), args.Error(2)
}

func (m *MockContentRepository) Save(ctx context.Context, content *entities.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockContentRepository) Create(ctx context.Context, content *entities.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockContentRepository) Update(ctx context.Context, content *entities.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockContentVersionRepository struct {
	mock.Mock
}

func (m *MockContentVersionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ContentVersion, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.ContentVersion), args.Error(1)
}

func (m *MockContentVersionRepository) FindByContentID(ctx context.Context, contentID uuid.UUID) ([]*entities.ContentVersion, error) {
	args := m.Called(ctx, contentID)
	return args.Get(0).([]*entities.ContentVersion), args.Error(1)
}

func (m *MockContentVersionRepository) FindByContentIDAndVersion(ctx context.Context, contentID uuid.UUID, version int) (*entities.ContentVersion, error) {
	args := m.Called(ctx, contentID, version)
	return args.Get(0).(*entities.ContentVersion), args.Error(1)
}

func (m *MockContentVersionRepository) Save(ctx context.Context, version *entities.ContentVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockContentVersionRepository) Create(ctx context.Context, version *entities.ContentVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Project), args.Error(1)
}

func (m *MockProjectRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, page int, pageSize int) ([]*entities.Project, int, error) {
	args := m.Called(ctx, clientID, page, pageSize)
	return args.Get(0).([]*entities.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) FindByDeadlineRange(ctx context.Context, startDate time.Time, endDate time.Time, page int, pageSize int) ([]*entities.Project, int, error) {
	args := m.Called(ctx, startDate, endDate, page, pageSize)
	return args.Get(0).([]*entities.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) FindByStatus(ctx context.Context, status entities.ProjectStatus, offset, limit int) ([]*entities.Project, int, error) {
	args := m.Called(ctx, status, offset, limit)
	return args.Get(0).([]*entities.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) Save(ctx context.Context, project *entities.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Create(ctx context.Context, project *entities.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *entities.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) FindActive(ctx context.Context, page int, pageSize int) ([]*entities.Project, int, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*entities.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) FindAll(ctx context.Context, page int, pageSize int) ([]*entities.Project, int, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*entities.Project), args.Int(1), args.Error(2)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) FindByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockEventRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID, eventType string, offset, limit int) ([]*events.Event, int, error) {
	args := m.Called(ctx, entityID, eventType, offset, limit)
	return args.Get(0).([]*events.Event), args.Int(1), args.Error(2)
}

func (m *MockEventRepository) Create(ctx context.Context, event *events.BaseEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Save(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) FindByAggregateID(ctx context.Context, aggregateID uuid.UUID, page int, pageSize int) ([]interface{}, int, error) {
	args := m.Called(ctx, aggregateID, page, pageSize)
	return args.Get(0).([]interface{}), args.Int(1), args.Error(2)
}

func (m *MockEventRepository) FindByTimeRange(ctx context.Context, startTime time.Time, endTime time.Time, page int, pageSize int) ([]interface{}, int, error) {
	args := m.Called(ctx, startTime, endTime, page, pageSize)
	return args.Get(0).([]interface{}), args.Int(1), args.Error(2)
}

func (m *MockEventRepository) FindByType(ctx context.Context, eventType string, page int, pageSize int) ([]interface{}, int, error) {
	args := m.Called(ctx, eventType, page, pageSize)
	return args.Get(0).([]interface{}), args.Int(1), args.Error(2)
}

func (m *MockEventRepository) FindLatest(ctx context.Context, limit int) ([]interface{}, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]interface{}), args.Error(1)
}

type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt interface{}) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

type MockContextManager struct {
	mock.Mock
}

func (m *MockContextManager) GetContext(ctx context.Context, projectID uuid.UUID) (*ContextWindow, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*ContextWindow), args.Error(1)
}

func (m *MockContextManager) AddEntry(ctx context.Context, projectID uuid.UUID, entry ContextEntry) error {
	args := m.Called(ctx, projectID, entry)
	return args.Error(0)
}

func (m *MockContextManager) SwitchContext(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockContextManager) InjectDomainKnowledge(ctx context.Context, projectID uuid.UUID, knowledge map[string]interface{}) error {
	args := m.Called(ctx, projectID, knowledge)
	return args.Error(0)
}

func (m *MockContextManager) SerializeContext(ctx context.Context, projectID uuid.UUID) (string, error) {
	args := m.Called(ctx, projectID)
	return args.String(0), args.Error(1)
}

func (m *MockContextManager) DeserializeContext(ctx context.Context, serialized string) (uuid.UUID, error) {
	args := m.Called(ctx, serialized)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockContextManager) GetContextMetrics(ctx context.Context, projectID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

type MockResearcher struct {
	mock.Mock
}

func (m *MockResearcher) Research(ctx context.Context, content *entities.Content, requirements ResearchRequirements) (*ResearchOutput, error) {
	args := m.Called(ctx, content, requirements)
	return args.Get(0).(*ResearchOutput), args.Error(1)
}

func (m *MockResearcher) EvaluateSourceCredibility(ctx context.Context, source ResearchSource) (float64, error) {
	args := m.Called(ctx, source)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockResearcher) ExtractKeyFacts(ctx context.Context, sources []ResearchSource) ([]string, error) {
	args := m.Called(ctx, sources)
	return args.Get(0).([]string), args.Error(1)
}

type MockQualityChecker struct {
	mock.Mock
}

func (m *MockQualityChecker) CheckContent(ctx context.Context, content *entities.Content, input QualityCheckInput) (QualityCheckOutput, error) {
	args := m.Called(ctx, content, input)
	return args.Get(0).(QualityCheckOutput), args.Error(1)
}

// Types are declared elsewhere in the codebase and imported here for testing

// MockPromptTemplateManager mocks the prompt template manager
type MockPromptTemplateManager struct {
	mock.Mock
}

func (m *MockPromptTemplateManager) GeneratePrompt(contentType entities.ContentType, stage string, data interface{}) (string, error) {
	args := m.Called(contentType, stage, data)
	return args.String(0), args.Error(1)
}

// Test functions

func TestContentPipeline_CreateContent(t *testing.T) {
	// Setup mocks
	mockContentRepo := new(MockContentRepository)
	mockContentVersionRepo := new(MockContentVersionRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockEventRepo := new(MockEventRepository)
	mockLLMClient := new(MockLLMClient)
	mockContextManager := new(MockContextManager)
	mockResearcher := new(MockResearcher)
	mockQualityChecker := new(MockQualityChecker)

	// Setup pipeline
	config := PipelineConfig{
		MaxRetries:            2,
		ContextWindowSize:     4096,
		EnableFactChecking:    true,
		EnablePlagiarismCheck: true,
		SEOOptimization:       true,
		StageTimeoutSeconds:   30,
	}

	pipeline := NewContentPipeline(
		mockContentRepo,
		mockContentVersionRepo,
		mockProjectRepo,
		mockEventRepo,
		mockLLMClient,
		mockContextManager,
		mockResearcher,
		mockQualityChecker,
		config,
	)

	// Test data
	projectID := uuid.New()
	title := "Test Blog Post"
	contentType := entities.ContentTypeBlogPost

	// Create test project
	testProject, _ := entities.NewProject(
		uuid.New(),
		"Test Project",
		"Test project description",
		contentType,
		time.Now().Add(24*time.Hour),
		entities.Money{Amount: 100.0, Currency: "USD"},
	)
	testProject.ProjectID = projectID

	// Setup expectations
	mockContentRepo.On("Create", mock.Anything, mock.MatchedBy(func(content *entities.Content) bool {
		return content.Title == title && content.Type == contentType && content.ProjectID == projectID
	})).Return(nil)

	mockContentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	mockProjectRepo.On("FindByID", mock.Anything, projectID).Return(testProject, nil)

	mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	mockContextManager.On("SwitchContext", mock.Anything, projectID).Return(nil)
	mockContextManager.On("AddEntry", mock.Anything, projectID, mock.Anything).Return(nil)

	// Research stage expectations
	mockResearcher.On("Research", mock.Anything, mock.Anything, mock.Anything).Return(&ResearchOutput{
		Topics: []ResearchTopic{
			{Topic: "Test Topic", Keywords: []string{"test", "blog"}, Priority: 8, MaxSources: 5},
		},
		Sources: []ResearchSource{
			{Type: "web", URL: "https://example.com", Title: "Test Source", Content: "Test content", Credibility: 0.8},
		},
		KeyFacts:   []string{"Test fact 1", "Test fact 2"},
		References: []string{"Test Reference"},
		Summary:    "Test research summary",
	}, nil)

	// LLM generation expectations
	mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("Generated content", nil)

	// Quality check expectations
	mockQualityChecker.On("CheckContent", mock.Anything, mock.Anything, mock.Anything).Return(QualityCheckOutput{
		ReadabilityScore: 85.0,
		SEOScore:         78.0,
		EngagementScore:  82.0,
		PlagiarismScore:  0.95,
		SuggestionsByCategory: map[string][]string{
			"SEO": {"Add more keywords"},
		},
		Keywords: []string{"test", "blog", "content"},
	}, nil)

	// Execute test
	ctx := context.Background()
	result, err := pipeline.CreateContent(ctx, projectID, title, contentType)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, title, result.Title)
	assert.Equal(t, contentType, result.Type)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, entities.ContentStatusReview, result.Status)
	assert.NotEmpty(t, result.Data)

	// Verify all mocks were called as expected
	mockContentRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockContextManager.AssertExpectations(t)
	mockResearcher.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
	mockQualityChecker.AssertExpectations(t)
}

func TestContentPipeline_StageExecution(t *testing.T) {
	// Setup mocks
	mockContentRepo := new(MockContentRepository)
	mockContentVersionRepo := new(MockContentVersionRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockEventRepo := new(MockEventRepository)
	mockLLMClient := new(MockLLMClient)
	mockContextManager := new(MockContextManager)
	mockResearcher := new(MockResearcher)
	mockQualityChecker := new(MockQualityChecker)

	config := PipelineConfig{
		MaxRetries:          2,
		StageTimeoutSeconds: 5,
	}

	pipeline := NewContentPipeline(
		mockContentRepo,
		mockContentVersionRepo,
		mockProjectRepo,
		mockEventRepo,
		mockLLMClient,
		mockContextManager,
		mockResearcher,
		mockQualityChecker,
		config,
	)

	// Test content
	content, _ := entities.NewContent(uuid.New(), "Test Content", entities.ContentTypeBlogPost)
	content.UpdateMetadata("research", map[string]interface{}{"summary": "test research"})
	content.UpdateMetadata("outline", "Test outline")

	// Test project
	testProject, _ := entities.NewProject(
		uuid.New(),
		"Test Project",
		"Test description",
		entities.ContentTypeBlogPost,
		time.Now().Add(24*time.Hour),
		entities.Money{Amount: 100.0, Currency: "USD"},
	)

	mockProjectRepo.On("FindByID", mock.Anything, content.ProjectID).Return(testProject, nil)
	mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	t.Run("Research Stage", func(t *testing.T) {
		mockResearcher.On("Research", mock.Anything, content, mock.Anything).Return(&ResearchOutput{
			Summary: "Research completed successfully",
		}, nil).Once()

		result, err := pipeline.researchStage(context.Background(), content)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Contains(t, result.Content, "Research completed")
	})

	t.Run("Outlining Stage", func(t *testing.T) {
		mockContextManager.On("SwitchContext", mock.Anything, content.ProjectID).Return(nil).Once()
		mockContextManager.On("AddEntry", mock.Anything, content.ProjectID, mock.Anything).Return(nil).Once()
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("Generated outline", nil).Once()

		result, err := pipeline.outliningStage(context.Background(), content)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "Generated outline", result.Content)
	})

	t.Run("Drafting Stage", func(t *testing.T) {
		mockContextManager.On("AddEntry", mock.Anything, content.ProjectID, mock.Anything).Return(nil).Once()
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("Generated draft", nil).Once()

		result, err := pipeline.draftingStage(context.Background(), content)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "Generated draft", result.Content)
	})

	t.Run("Editing Stage", func(t *testing.T) {
		content.UpdateContent("Draft content to edit", "test")

		mockContextManager.On("AddEntry", mock.Anything, content.ProjectID, mock.Anything).Return(nil).Once()
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("Edited content", nil).Once()

		result, err := pipeline.editingStage(context.Background(), content)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "Edited content", result.Content)
	})

	t.Run("Finalization Stage", func(t *testing.T) {
		content.UpdateContent("Content to finalize", "test")

		mockContextManager.On("AddEntry", mock.Anything, content.ProjectID, mock.Anything).Return(nil).Once()
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("Final content", nil).Once()

		result, err := pipeline.finalizationStage(context.Background(), content)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "Final content", result.Content)
		assert.True(t, result.Metadata["deliveryReady"].(bool))
	})

	// Verify all expectations
	mockResearcher.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
	mockContextManager.AssertExpectations(t)
}

func TestContentPipeline_ErrorHandling(t *testing.T) {
	// Setup mocks for error scenarios
	mockContentRepo := new(MockContentRepository)
	mockContentVersionRepo := new(MockContentVersionRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockEventRepo := new(MockEventRepository)
	mockLLMClient := new(MockLLMClient)
	mockContextManager := new(MockContextManager)
	mockResearcher := new(MockResearcher)
	mockQualityChecker := new(MockQualityChecker)

	config := PipelineConfig{
		MaxRetries:          1,
		StageTimeoutSeconds: 1,
	}

	pipeline := NewContentPipeline(
		mockContentRepo,
		mockContentVersionRepo,
		mockProjectRepo,
		mockEventRepo,
		mockLLMClient,
		mockContextManager,
		mockResearcher,
		mockQualityChecker,
		config,
	)

	t.Run("Research Stage Error", func(t *testing.T) {
		content, _ := entities.NewContent(uuid.New(), "Test Content", entities.ContentTypeBlogPost)

		mockProjectRepo.On("FindByID", mock.Anything, content.ProjectID).Return(nil, assert.AnError)
		mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		result, err := pipeline.researchStage(context.Background(), content)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get project")
	})

	t.Run("Outlining Stage Missing Research", func(t *testing.T) {
		content, _ := entities.NewContent(uuid.New(), "Test Content", entities.ContentTypeBlogPost)
		// Don't add research metadata

		result, err := pipeline.outliningStage(context.Background(), content)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "research data not found")
	})

	t.Run("Context Timeout", func(t *testing.T) {
		content, _ := entities.NewContent(uuid.New(), "Test Content", entities.ContentTypeBlogPost)
		content.UpdateMetadata("research", map[string]interface{}{"summary": "test"})

		testProject, _ := entities.NewProject(
			uuid.New(),
			"Test Project",
			"Test description",
			entities.ContentTypeBlogPost,
			time.Now().Add(24*time.Hour),
			entities.Money{Amount: 100.0, Currency: "USD"},
		)

		mockProjectRepo.On("FindByID", mock.Anything, content.ProjectID).Return(testProject, nil)
		mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockContextManager.On("SwitchContext", mock.Anything, content.ProjectID).Return(nil)

		// Simulate a slow LLM response that times out
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			time.Sleep(2 * time.Second) // Longer than timeout
		}).Return("", context.DeadlineExceeded)

		result, err := pipeline.executeStage(context.Background(), content, StageOutlining)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "timed out")
	})
}

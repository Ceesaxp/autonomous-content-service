package content_creation

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// ContextEntry represents a single entry in the conversation context
type ContextEntry struct {
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  int                    `json:"priority"` // Higher number = higher priority
	Metadata  map[string]interface{} `json:"metadata"` // Additional info about the entry
}

// ContextWindow represents the active context for a conversation
type ContextWindow struct {
	Entries            []ContextEntry          `json:"entries"`
	ProjectID          uuid.UUID               `json:"projectId"`
	ClientID           uuid.UUID               `json:"clientId"`
	ContentType        entities.ContentType    `json:"contentType"`
	MaxTokens          int                     `json:"maxTokens"`
	CurrentTokenCount  int                     `json:"currentTokenCount"`
	LastAccessed       time.Time               `json:"lastAccessed"`
	DomainKnowledge    map[string]interface{}  `json:"domainKnowledge"`
}

// ContextManager defines the interface for managing LLM conversation context
type ContextManager interface {
	// GetContext retrieves the context for a specific client or project
	GetContext(ctx context.Context, projectID uuid.UUID) (*ContextWindow, error)
	
	// AddEntry adds a new entry to the context window
	AddEntry(ctx context.Context, projectID uuid.UUID, entry ContextEntry) error
	
	// SwitchContext switches to a different project context
	SwitchContext(ctx context.Context, projectID uuid.UUID) error
	
	// InjectDomainKnowledge adds domain-specific knowledge to the context
	InjectDomainKnowledge(ctx context.Context, projectID uuid.UUID, knowledge map[string]interface{}) error
	
	// SerializeContext converts context to a storable format
	SerializeContext(ctx context.Context, projectID uuid.UUID) (string, error)
	
	// DeserializeContext loads context from a serialized format
	DeserializeContext(ctx context.Context, serialized string) (uuid.UUID, error)
	
	// GetContextMetrics retrieves metrics about context usage
	GetContextMetrics(ctx context.Context, projectID uuid.UUID) (map[string]interface{}, error)
}

// InMemoryContextManager implements ContextManager with in-memory storage
type InMemoryContextManager struct {
	contextWindows     map[uuid.UUID]*ContextWindow
	clientRepository   repositories.ClientRepository
	maxWindowSize      int
	mutex              sync.RWMutex
}

// NewInMemoryContextManager creates a new context manager
func NewInMemoryContextManager(clientRepo repositories.ClientRepository, maxTokens int) *InMemoryContextManager {
	return &InMemoryContextManager{
		contextWindows:   make(map[uuid.UUID]*ContextWindow),
		clientRepository: clientRepo,
		maxWindowSize:    maxTokens,
		mutex:            sync.RWMutex{},
	}
}

// GetContext retrieves the context for a specific project
func (cm *InMemoryContextManager) GetContext(ctx context.Context, projectID uuid.UUID) (*ContextWindow, error) {
	cm.mutex.RLock()
	window, exists := cm.contextWindows[projectID]
	cm.mutex.RUnlock()
	
	if !exists {
		return nil, errors.New("context window not found for project")
	}
	
	// Update last accessed time
	cm.mutex.Lock()
	window.LastAccessed = time.Now()
	cm.mutex.Unlock()
	
	return window, nil
}

// AddEntry adds a new entry to the context window with priority-based retention
func (cm *InMemoryContextManager) AddEntry(ctx context.Context, projectID uuid.UUID, entry ContextEntry) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	window, exists := cm.contextWindows[projectID]
	if !exists {
		return errors.New("context window not found for project")
	}
	
	// Estimate tokens in the new entry
	entryTokens := estimateTokens(entry.Content)
	
	// Check if adding this would exceed the window size
	if window.CurrentTokenCount + entryTokens > window.MaxTokens {
		// Need to remove entries to make space
		cm.pruneContextWindow(window, entryTokens)
	}
	
	// Add the new entry
	window.Entries = append(window.Entries, entry)
	window.CurrentTokenCount += entryTokens
	window.LastAccessed = time.Now()
	
	return nil
}

// pruneContextWindow removes lower priority entries to make room for new content
func (cm *InMemoryContextManager) pruneContextWindow(window *ContextWindow, requiredTokens int) {
	// Sort entries by priority (ascending) so we remove lowest priority first
	sort.Slice(window.Entries, func(i, j int) bool {
		return window.Entries[i].Priority < window.Entries[j].Priority
	})
	
	tokensToFree := window.CurrentTokenCount + requiredTokens - window.MaxTokens
	freedTokens := 0
	entriesRemoved := 0
	
	// Remove entries until we've freed enough tokens
	for i, entry := range window.Entries {
		entryTokens := estimateTokens(entry.Content)
		freedTokens += entryTokens
		entriesRemoved = i + 1
		
		if freedTokens >= tokensToFree {
			break
		}
	}
	
	// Remove the entries
	if entriesRemoved > 0 {
		window.Entries = window.Entries[entriesRemoved:]
		window.CurrentTokenCount -= freedTokens
	}
}

// SwitchContext switches to a different project context
func (cm *InMemoryContextManager) SwitchContext(ctx context.Context, projectID uuid.UUID) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	_, exists := cm.contextWindows[projectID]
	if !exists {
		// Initialize a new context window
		window := &ContextWindow{
			Entries:          []ContextEntry{},
			ProjectID:        projectID,
			MaxTokens:        cm.maxWindowSize,
			CurrentTokenCount: 0,
			LastAccessed:     time.Now(),
			DomainKnowledge:  make(map[string]interface{}),
		}
		cm.contextWindows[projectID] = window
	}
	
	return nil
}

// InjectDomainKnowledge adds domain-specific knowledge to the context
func (cm *InMemoryContextManager) InjectDomainKnowledge(ctx context.Context, projectID uuid.UUID, knowledge map[string]interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	window, exists := cm.contextWindows[projectID]
	if !exists {
		return errors.New("context window not found for project")
	}
	
	// Merge the knowledge maps
	for k, v := range knowledge {
		window.DomainKnowledge[k] = v
	}
	
	return nil
}

// SerializeContext converts context to a storable format
func (cm *InMemoryContextManager) SerializeContext(ctx context.Context, projectID uuid.UUID) (string, error) {
	cm.mutex.RLock()
	window, exists := cm.contextWindows[projectID]
	cm.mutex.RUnlock()
	
	if !exists {
		return "", errors.New("context window not found for project")
	}
	
	// Convert to JSON
	serialized, err := json.Marshal(window)
	if err != nil {
		return "", err
	}
	
	return string(serialized), nil
}

// DeserializeContext loads context from a serialized format
func (cm *InMemoryContextManager) DeserializeContext(ctx context.Context, serialized string) (uuid.UUID, error) {
	var window ContextWindow
	
	err := json.Unmarshal([]byte(serialized), &window)
	if err != nil {
		return uuid.Nil, err
	}
	
	// Store the loaded context
	cm.mutex.Lock()
	cm.contextWindows[window.ProjectID] = &window
	cm.mutex.Unlock()
	
	return window.ProjectID, nil
}

// GetContextMetrics retrieves metrics about context usage
func (cm *InMemoryContextManager) GetContextMetrics(ctx context.Context, projectID uuid.UUID) (map[string]interface{}, error) {
	cm.mutex.RLock()
	window, exists := cm.contextWindows[projectID]
	cm.mutex.RUnlock()
	
	if !exists {
		return nil, errors.New("context window not found for project")
	}
	
	metrics := map[string]interface{}{
		"entryCount":       len(window.Entries),
		"tokenUsage":       window.CurrentTokenCount,
		"tokenCapacity":    window.MaxTokens,
		"utilizationPct":   float64(window.CurrentTokenCount) / float64(window.MaxTokens) * 100,
		"lastAccessed":     window.LastAccessed,
		"domainKnowledgeEntries": len(window.DomainKnowledge),
	}
	
	return metrics, nil
}

// estimateTokens provides a rough estimate of tokens in a string
// A more accurate implementation would use the tokenizer from the LLM provider
func estimateTokens(text string) int {
	// Simple approximation: 1 token ≈ 4 characters
	return len(text) / 4
}
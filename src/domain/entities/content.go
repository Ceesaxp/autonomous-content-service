package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ContentType represents the type of content
type ContentType string

const (
	ContentTypeBlogPost          ContentType = "BlogPost"
	ContentTypeSocialPost        ContentType = "SocialPost"
	ContentTypeEmailNewsletter   ContentType = "EmailNewsletter"
	ContentTypeWebsiteCopy       ContentType = "WebsiteCopy"
	ContentTypeTechnicalArticle  ContentType = "TechnicalArticle"
	ContentTypeProductDescription ContentType = "ProductDescription"
	ContentTypePressRelease      ContentType = "PressRelease"
)

// ContentStatus represents the status of content
type ContentStatus string

const (
	ContentStatusPlanning    ContentStatus = "Planning"
	ContentStatusResearching ContentStatus = "Researching"
	ContentStatusDrafting    ContentStatus = "Drafting"
	ContentStatusEditing     ContentStatus = "Editing"
	ContentStatusReview      ContentStatus = "Review"
	ContentStatusApproved    ContentStatus = "Approved"
	ContentStatusPublished   ContentStatus = "Published"
	ContentStatusArchived    ContentStatus = "Archived"
)

// ContentStatistics contains metrics about the content
type ContentStatistics struct {
	ReadabilityScore float64 `json:"readabilityScore"`
	SEOScore        float64 `json:"seoScore"`
	EngagementScore float64 `json:"engagementScore"`
	PlagiarismScore float64 `json:"plagiarismScore"`
}

// Content represents a piece of content created within a project
type Content struct {
	ContentID   uuid.UUID      `json:"contentId"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Title       string         `json:"title"`
	Type        ContentType    `json:"type"`
	Status      ContentStatus  `json:"status"`
	Data        string         `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     int            `json:"version"`
	WordCount   int            `json:"wordCount"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Versions    []*ContentVersion `json:"versions,omitempty"`
	Statistics  *ContentStatistics `json:"statistics,omitempty"`
}

// NewContent creates a new content item with the given properties
func NewContent(projectID uuid.UUID, title string, contentType ContentType) (*Content, error) {
	content := &Content{
		ContentID:   uuid.New(),
		ProjectID:   projectID,
		Title:       title,
		Type:        contentType,
		Status:      ContentStatusPlanning,
		Data:        "",
		Metadata:    make(map[string]interface{}),
		Version:     1,
		WordCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Versions:    []*ContentVersion{},
		Statistics:  &ContentStatistics{},
	}

	if err := content.Validate(); err != nil {
		return nil, err
	}

	return content, nil
}

// Validate ensures the content has all required fields
func (c *Content) Validate() error {
	if c.Title == "" {
		return errors.New("title is required")
	}

	// Validate minimum word count based on content type if content has data
	if c.Data != "" && c.WordCount > 0 {
		minWordCount := getMinWordCount(c.Type)
		if c.WordCount < minWordCount {
			return errors.New("content does not meet minimum word count requirement")
		}
	}

	return nil
}

// getMinWordCount returns the minimum word count required for each content type
func getMinWordCount(contentType ContentType) int {
	switch contentType {
	case ContentTypeBlogPost:
		return 500
	case ContentTypeSocialPost:
		return 50
	case ContentTypeEmailNewsletter:
		return 300
	case ContentTypeWebsiteCopy:
		return 200
	case ContentTypeTechnicalArticle:
		return 800
	case ContentTypeProductDescription:
		return 150
	case ContentTypePressRelease:
		return 400
	default:
		return 100
	}
}

// UpdateContent updates the content data and creates a new version
func (c *Content) UpdateContent(data string, source string) error {
	// Create a version of the current content
	version := NewContentVersion(c.ContentID, c.Version, c.Data, c.Metadata, source)
	c.Versions = append(c.Versions, version)

	// Update content with new data
	c.Data = data
	c.Version++
	c.WordCount = countWords(data)
	c.UpdateTimestamp()

	return c.Validate()
}

// countWords is a simple function to count words in a string
func countWords(s string) int {
	// This is a simplified implementation - a real one would be more sophisticated
	if s == "" {
		return 0
	}

	// For simplicity, just counting spaces + 1
	count := 1
	for _, char := range s {
		if char == ' ' || char == '\n' || char == '\t' || char == '\r' {
			count++
		}
	}
	return count
}

// UpdateStatus changes the content status
func (c *Content) UpdateStatus(status ContentStatus) {
	c.Status = status
	c.UpdateTimestamp()
}

// UpdateMetadata adds or updates metadata
func (c *Content) UpdateMetadata(key string, value interface{}) {
	c.Metadata[key] = value
	c.UpdateTimestamp()
}

// UpdateStatistics updates the content statistics
func (c *Content) UpdateStatistics(stats ContentStatistics) {
	c.Statistics = &stats
	c.UpdateTimestamp()
}

// UpdateTimestamp updates the UpdatedAt timestamp to the current time
func (c *Content) UpdateTimestamp() {
	c.UpdatedAt = time.Now()
}

// IsComplete returns true if the content is Approved or Published
func (c *Content) IsComplete() bool {
	return c.Status == ContentStatusApproved || c.Status == ContentStatusPublished
}

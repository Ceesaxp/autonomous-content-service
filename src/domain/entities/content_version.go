package entities

import (
	"time"

	"github.com/google/uuid"
)

// ContentVersion stores a previous version of content
type ContentVersion struct {
	VersionID     uuid.UUID              `json:"versionId"`
	ContentID     uuid.UUID              `json:"contentId"`
	VersionNumber int                    `json:"versionNumber"`
	Data          string                 `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"createdAt"`
	CreatedBy     string                 `json:"createdBy"`
}

// NewContentVersion creates a new content version
func NewContentVersion(contentID uuid.UUID, versionNumber int, data string, metadata map[string]interface{}, createdBy string) *ContentVersion {
	// Create a deep copy of metadata
	metadataCopy := make(map[string]interface{})
	for k, v := range metadata {
		metadataCopy[k] = v
	}

	return &ContentVersion{
		VersionID:     uuid.New(),
		ContentID:     contentID,
		VersionNumber: versionNumber,
		Data:          data,
		Metadata:      metadataCopy,
		CreatedAt:     time.Now(),
		CreatedBy:     createdBy,
	}
}

package risk_management

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// BackupManagerImpl implements backup management
type BackupManagerImpl struct {
	riskRepo  repositories.RiskRepository
	eventRepo repositories.EventRepository
	config    *BackupConfig
}

// BackupConfig holds backup configuration
type BackupConfig struct {
	RetentionDays       int
	MaxBackupSize       int64
	BackupLocation      string
	FullBackupSchedule  string
	IncrementalSchedule string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(
	riskRepo repositories.RiskRepository,
	eventRepo repositories.EventRepository,
) *BackupManagerImpl {
	return &BackupManagerImpl{
		riskRepo:  riskRepo,
		eventRepo: eventRepo,
		config: &BackupConfig{
			RetentionDays:       30,
			MaxBackupSize:       10 * 1024 * 1024 * 1024, // 10GB
			BackupLocation:      "/backups",
			FullBackupSchedule:  "0 2 * * 0",   // Weekly at 2 AM on Sunday
			IncrementalSchedule: "0 2 * * 1-6", // Daily at 2 AM Mon-Sat
		},
	}
}

// CreateBackup creates a new backup
func (b *BackupManagerImpl) CreateBackup(ctx context.Context, backupType string) (*entities.Backup, error) {
	startTime := time.Now()

	// Determine components to backup
	components := b.getBackupComponents(backupType)

	// Create backup record
	backup := &entities.Backup{
		ID:              uuid.New(),
		Name:            fmt.Sprintf("%s_%s", backupType, time.Now().Format("20060102_150405")),
		BackupType:      backupType,
		SizeBytes:       0,
		Status:          entities.BackupStatusInProgress,
		StorageLocation: fmt.Sprintf("%s/%s_%s", b.config.BackupLocation, backupType, time.Now().Format("20060102_150405")),
		Metadata:        map[string]interface{}{"components": components},
		RetentionUntil:  &[]time.Time{time.Now().Add(time.Duration(b.config.RetentionDays) * 24 * time.Hour)}[0],
		StartedAt:       time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save initial backup record
	if err := b.riskRepo.CreateBackupRecord(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Perform backup for each component
	totalSize := int64(0)
	failedComponents := []string{}

	for _, component := range components {
		size, err := b.backupComponent(ctx, component, backup.StorageLocation)
		if err != nil {
			failedComponents = append(failedComponents, component)
			continue
		}
		totalSize += size
	}

	// Update backup record
	backup.SizeBytes = totalSize
	backup.UpdatedAt = time.Now()

	if len(failedComponents) > 0 {
		backup.Status = entities.BackupStatusFailed
		backup.Metadata = map[string]interface{}{
			"failed_components": failedComponents,
		}

		// Emit backup failed event
		event := &events.BackupFailedEvent{
			BaseEvent: events.BaseEvent{
				EventID:   generateEventID(),
				EventType: events.BackupFailed,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"source": "backup_manager",
				},
			},
			BackupID:   backup.ID.String(),
			Type:       backup.BackupType,
			Components: failedComponents,
			Error:      fmt.Sprintf("Failed to backup %d components", len(failedComponents)),
			RetryCount: 0,
		}
		if err := b.eventRepo.Save(ctx, event); err != nil {
			// Log error but continue with backup process
			fmt.Printf("Failed to save backup failed event: %v\n", err)
		}
	} else {
		backup.Status = entities.BackupStatusCompleted

		// Emit backup completed event
		event := &events.BackupCompletedEvent{
			BaseEvent: events.BaseEvent{
				EventID:   generateEventID(),
				EventType: events.BackupCompleted,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"source": "backup_manager",
				},
			},
			BackupID:   backup.ID.String(),
			Type:       backup.BackupType,
			Size:       backup.SizeBytes,
			Components: components,
			Location:   backup.StorageLocation,
			Duration:   time.Since(startTime),
			Verified:   false,
		}
		if err := b.eventRepo.Save(ctx, event); err != nil {
			// Log error but continue with backup process
			fmt.Printf("Failed to save backup completed event: %v\n", err)
		}
	}

	// Update backup record
	if err := b.riskRepo.UpdateBackupRecord(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to update backup record: %w", err)
	}

	// Verify backup if successful
	if backup.Status == entities.BackupStatusCompleted {
		go func() {
			if err := b.VerifyBackup(context.Background(), backup.ID.String()); err != nil {
				fmt.Printf("Failed to verify backup %s: %v\n", backup.ID.String(), err)
			}
		}()
	}

	return backup, nil
}

// VerifyBackup verifies backup integrity
func (b *BackupManagerImpl) VerifyBackup(ctx context.Context, backupID string) error {
	backup, err := b.riskRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup record: %w", err)
	}

	// Perform verification for each component
	allValid := true
	components, _ := backup.Metadata["components"].([]string)
	for _, component := range components {
		if !b.verifyComponent(ctx, component, backup.StorageLocation) {
			allValid = false
			break
		}
	}

	// Update verification status
	if allValid {
		now := time.Now()
		backup.VerifiedAt = &now
		backup.Status = entities.BackupStatusVerified
	} else {
		backup.Status = entities.BackupStatusFailed
	}

	backup.UpdatedAt = time.Now()
	return b.riskRepo.UpdateBackupRecord(ctx, backup)
}

// RestoreFromBackup restores system from backup
func (b *BackupManagerImpl) RestoreFromBackup(ctx context.Context, backupID string) error {
	backup, err := b.riskRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup record: %w", err)
	}

	// Verify backup is valid
	if backup.Status != entities.BackupStatusVerified && backup.Status != entities.BackupStatusCompleted {
		return fmt.Errorf("backup %s is not valid for restore (status: %s)", backupID, backup.Status)
	}

	// Create restore point before restoration
	restorePoint, err := b.CreateBackup(ctx, "restore_point")
	if err != nil {
		return fmt.Errorf("failed to create restore point: %w", err)
	}

	// Restore each component
	failedComponents := []string{}
	components, _ := backup.Metadata["components"].([]string)
	for _, component := range components {
		if err := b.restoreComponent(ctx, component, backup.StorageLocation); err != nil {
			failedComponents = append(failedComponents, component)
		}
	}

	if len(failedComponents) > 0 {
		// Attempt to restore from restore point
		if err := b.rollbackToRestorePoint(ctx, restorePoint); err != nil {
			fmt.Printf("Failed to rollback to restore point: %v\n", err)
		}
		return fmt.Errorf("restore failed for components: %v", failedComponents)
	}

	// Update restore timestamp
	now := time.Now()
	backup.Metadata["restored_at"] = now
	backup.UpdatedAt = now
	if err := b.riskRepo.UpdateBackupRecord(ctx, backup); err != nil {
		fmt.Printf("Failed to update backup record: %v\n", err)
	}

	return nil
}

// GetBackupStatus returns current backup status
func (b *BackupManagerImpl) GetBackupStatus(ctx context.Context) (*BackupStatus, error) {
	// Get last successful backup
	lastBackup, err := b.riskRepo.GetLastSuccessfulBackup(ctx, "full")
	if err != nil {
		return nil, fmt.Errorf("failed to get last backup: %w", err)
	}

	// Get all backups
	backups, err := b.riskRepo.ListBackupRecords(ctx, repositories.BackupFilters{
		Status:    "completed",
		StartDate: time.Now().Add(-30 * 24 * time.Hour),
		EndDate:   time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// Calculate total size
	totalSize := int64(0)
	oldestBackup := time.Now()
	for _, backup := range backups {
		totalSize += backup.SizeBytes
		if backup.CreatedAt.Before(oldestBackup) {
			oldestBackup = backup.CreatedAt
		}
	}

	status := &BackupStatus{
		LastSuccessfulBackup: lastBackup.CreatedAt,
		NextScheduledBackup:  b.getNextScheduledBackup(),
		BackupHealth:         b.assessBackupHealth(backups),
		AvailableBackups:     len(backups),
		OldestBackup:         oldestBackup,
		TotalBackupSize:      totalSize,
	}

	return status, nil
}

// CleanupOldBackups removes backups past retention period
func (b *BackupManagerImpl) CleanupOldBackups(ctx context.Context) error {
	return b.riskRepo.CleanupOldBackups(ctx, b.config.RetentionDays)
}

// Helper methods

func (b *BackupManagerImpl) getBackupComponents(backupType string) []string {
	switch backupType {
	case "full":
		return []string{"database", "files", "configuration", "logs"}
	case "incremental":
		return []string{"database", "files"}
	case "emergency":
		return []string{"database", "configuration"}
	case "restore_point":
		return []string{"database", "configuration"}
	default:
		return []string{"database"}
	}
}

func (b *BackupManagerImpl) backupComponent(ctx context.Context, component string, location string) (int64, error) {
	// Simulate backup operation
	// In production, this would perform actual backup
	switch component {
	case "database":
		// Backup database
		return 1024 * 1024 * 100, nil // 100MB
	case "files":
		// Backup files
		return 1024 * 1024 * 500, nil // 500MB
	case "configuration":
		// Backup configuration
		return 1024 * 100, nil // 100KB
	case "logs":
		// Backup logs
		return 1024 * 1024 * 50, nil // 50MB
	default:
		return 0, fmt.Errorf("unknown component: %s", component)
	}
}

func (b *BackupManagerImpl) verifyComponent(ctx context.Context, component string, location string) bool {
	// Simulate verification
	// In production, this would perform actual verification (checksums, integrity checks)
	return true
}

func (b *BackupManagerImpl) restoreComponent(ctx context.Context, component string, location string) error {
	// Simulate restore operation
	// In production, this would perform actual restoration
	return nil
}

func (b *BackupManagerImpl) rollbackToRestorePoint(ctx context.Context, restorePoint *entities.Backup) error {
	// Restore from the restore point created before failed restoration
	components, _ := restorePoint.Metadata["components"].([]string)
	for _, component := range components {
		if err := b.restoreComponent(ctx, component, restorePoint.StorageLocation); err != nil {
			fmt.Printf("Failed to restore component %s: %v\n", component, err)
		}
	}
	return nil
}

func (b *BackupManagerImpl) getNextScheduledBackup() time.Time {
	// Calculate next scheduled backup based on cron schedule
	// Simplified implementation - in production would use cron parser
	now := time.Now()

	// If it's Sunday, next full backup is next Sunday at 2 AM
	// Otherwise, next backup is tomorrow at 2 AM
	nextBackup := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())

	return nextBackup
}

func (b *BackupManagerImpl) assessBackupHealth(backups []*entities.Backup) string {
	if len(backups) == 0 {
		return "critical"
	}

	// Check if we have recent backups
	lastBackup := backups[len(backups)-1]
	timeSinceLastBackup := time.Since(lastBackup.CreatedAt)

	if timeSinceLastBackup > 48*time.Hour {
		return "warning"
	} else if timeSinceLastBackup > 72*time.Hour {
		return "critical"
	}

	// Check for failed backups
	failedCount := 0
	for _, backup := range backups {
		if backup.Status != entities.BackupStatusCompleted && backup.Status != entities.BackupStatusVerified {
			failedCount++
		}
	}

	if float64(failedCount)/float64(len(backups)) > 0.2 {
		return "warning"
	}

	return "healthy"
}


// generateEventID defined in operational_risk_monitor.go

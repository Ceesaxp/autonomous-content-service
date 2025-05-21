package content_creation

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// RevisionTracker manages revision tracking to measure quality improvement
type RevisionTracker struct {
	revisions map[uuid.UUID][]RevisionRecord
	active    map[string]*ActiveRevision
	mutex     sync.RWMutex
}

// NewRevisionTracker creates a new revision tracker
func NewRevisionTracker() *RevisionTracker {
	return &RevisionTracker{
		revisions: make(map[uuid.UUID][]RevisionRecord),
		active:    make(map[string]*ActiveRevision),
	}
}

// RevisionRecord represents a single content revision
type RevisionRecord struct {
	ID                string                 `json:"id"`
	ContentID         uuid.UUID              `json:"contentId"`
	RevisionNumber    int                    `json:"revisionNumber"`
	StartTime         time.Time              `json:"startTime"`
	EndTime           time.Time              `json:"endTime"`
	Duration          time.Duration          `json:"duration"`
	InitialScore      float64                `json:"initialScore"`
	FinalScore        float64                `json:"finalScore"`
	ScoreImprovement  float64                `json:"scoreImprovement"`
	QualityMetrics    map[string]float64     `json:"qualityMetrics"`
	ChangesApplied    []AppliedChange        `json:"changesApplied"`
	ImprovementAreas  []string               `json:"improvementAreas"`
	Reviewer          string                 `json:"reviewer"`
	Status            RevisionStatus         `json:"status"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ActiveRevision tracks ongoing revisions
type ActiveRevision struct {
	ID               string                 `json:"id"`
	ContentID        uuid.UUID              `json:"contentId"`
	StartTime        time.Time              `json:"startTime"`
	InitialScore     float64                `json:"initialScore"`
	InitialContent   string                 `json:"initialContent"`
	CurrentContent   string                 `json:"currentContent"`
	Changes          []PendingChange        `json:"changes"`
	QualityChecks    []QualityCheckPoint    `json:"qualityChecks"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// AppliedChange represents a change that was applied during revision
type AppliedChange struct {
	ID              string                 `json:"id"`
	Type            ChangeType             `json:"type"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	BeforeText      string                 `json:"beforeText"`
	AfterText       string                 `json:"afterText"`
	ExpectedImpact  float64                `json:"expectedImpact"`
	ActualImpact    float64                `json:"actualImpact"`
	AppliedAt       time.Time              `json:"appliedAt"`
	SuccessRating   float64                `json:"successRating"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PendingChange represents a change that hasn't been applied yet
type PendingChange struct {
	ID              string                 `json:"id"`
	Type            ChangeType             `json:"type"`
	Description     string                 `json:"description"`
	ProposedText    string                 `json:"proposedText"`
	ExpectedImpact  float64                `json:"expectedImpact"`
	Priority        int                    `json:"priority"`
	Status          ChangeStatus           `json:"status"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// QualityCheckPoint represents a quality measurement point during revision
type QualityCheckPoint struct {
	Timestamp      time.Time          `json:"timestamp"`
	OverallScore   float64            `json:"overallScore"`
	CriteriaScores map[string]float64 `json:"criteriaScores"`
	CheckType      CheckType          `json:"checkType"`
	Notes          string             `json:"notes"`
}

// RevisionAnalytics provides analytics on revision performance
type RevisionAnalytics struct {
	TotalRevisions        int                    `json:"totalRevisions"`
	AverageImprovement    float64                `json:"averageImprovement"`
	AverageRevisionTime   time.Duration          `json:"averageRevisionTime"`
	SuccessRate           float64                `json:"successRate"`
	CommonImprovements    []ImprovementPattern   `json:"commonImprovements"`
	PerformanceMetrics    PerformanceMetrics     `json:"performanceMetrics"`
	TrendAnalysis         TrendAnalysis          `json:"trendAnalysis"`
	Recommendations       []SystemRecommendation `json:"recommendations"`
}

// ImprovementPattern identifies common improvement patterns
type ImprovementPattern struct {
	Pattern         string  `json:"pattern"`
	Frequency       int     `json:"frequency"`
	AverageImpact   float64 `json:"averageImpact"`
	SuccessRate     float64 `json:"successRate"`
	Category        string  `json:"category"`
}

// PerformanceMetrics tracks system performance
type PerformanceMetrics struct {
	AverageProcessingTime time.Duration          `json:"averageProcessingTime"`
	ThroughputPerHour     float64                `json:"throughputPerHour"`
	ErrorRate             float64                `json:"errorRate"`
	QualityDistribution   map[string]int         `json:"qualityDistribution"`
	ImprovementVelocity   float64                `json:"improvementVelocity"`
}

// TrendAnalysis analyzes trends in revision performance
type TrendAnalysis struct {
	QualityTrend          TrendDirection         `json:"qualityTrend"`
	EfficiencyTrend       TrendDirection         `json:"efficiencyTrend"`
	TimeSeriesData        []TimeSeriesPoint      `json:"timeSeriesData"`
	SeasonalPatterns      []SeasonalPattern      `json:"seasonalPatterns"`
	Predictions           []QualityPrediction    `json:"predictions"`
}

// SystemRecommendation provides system improvement recommendations
type SystemRecommendation struct {
	Type           RecommendationType `json:"type"`
	Priority       int                `json:"priority"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	Impact         string             `json:"impact"`
	Implementation string             `json:"implementation"`
}

// Enums and types

type RevisionStatus string

const (
	StatusInProgress RevisionStatus = "in_progress"
	StatusCompleted  RevisionStatus = "completed"
	StatusFailed     RevisionStatus = "failed"
	StatusCancelled  RevisionStatus = "cancelled"
)

type ChangeType string

const (
	ChangeStructural  ChangeType = "structural"
	ChangeContent     ChangeType = "content"
	ChangeLanguage    ChangeType = "language"
	ChangeFormatting  ChangeType = "formatting"
	ChangeOptimization ChangeType = "optimization"
)

type ChangeStatus string

const (
	ChangePending  ChangeStatus = "pending"
	ChangeApplied  ChangeStatus = "applied"
	ChangeRejected ChangeStatus = "rejected"
	ChangeReverted ChangeStatus = "reverted"
)

type CheckType string

const (
	CheckInitial     CheckType = "initial"
	CheckIntermediate CheckType = "intermediate"
	CheckFinal       CheckType = "final"
	CheckValidation  CheckType = "validation"
)

type TrendDirection string

const (
	TrendImproving TrendDirection = "improving"
	TrendStable    TrendDirection = "stable"
	TrendDeclining TrendDirection = "declining"
)

type RecommendationType string

const (
	RecommendationProcess    RecommendationType = "process"
	RecommendationTechnical  RecommendationType = "technical"
	RecommendationWorkflow   RecommendationType = "workflow"
	RecommendationTraining   RecommendationType = "training"
)

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Metric    string    `json:"metric"`
}

type SeasonalPattern struct {
	Period      string  `json:"period"`
	Pattern     string  `json:"pattern"`
	Correlation float64 `json:"correlation"`
}

type QualityPrediction struct {
	Timestamp       time.Time `json:"timestamp"`
	PredictedScore  float64   `json:"predictedScore"`
	Confidence      float64   `json:"confidence"`
	Factors         []string  `json:"factors"`
}

// RevisionResult contains the result of completing a revision
type RevisionResult struct {
	Score           float64 `json:"score"`
	PassedThreshold bool    `json:"passedThreshold"`
	Improvements    int     `json:"improvements"`
}

// Core revision tracking methods

// StartRevision begins tracking a new revision
func (rt *RevisionTracker) StartRevision(contentID uuid.UUID, initialContent string) string {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	revisionID := uuid.New().String()
	
	activeRevision := &ActiveRevision{
		ID:             revisionID,
		ContentID:      contentID,
		StartTime:      time.Now(),
		InitialContent: initialContent,
		CurrentContent: initialContent,
		Changes:        []PendingChange{},
		QualityChecks:  []QualityCheckPoint{},
		Metadata:       make(map[string]interface{}),
	}

	rt.active[revisionID] = activeRevision
	return revisionID
}

// AddQualityCheckPoint adds a quality measurement point
func (rt *RevisionTracker) AddQualityCheckPoint(revisionID string, score float64, criteriaScores map[string]float64, checkType CheckType) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if active, exists := rt.active[revisionID]; exists {
		checkPoint := QualityCheckPoint{
			Timestamp:      time.Now(),
			OverallScore:   score,
			CriteriaScores: criteriaScores,
			CheckType:      checkType,
		}
		active.QualityChecks = append(active.QualityChecks, checkPoint)

		// Update initial score if this is the first check
		if checkType == CheckInitial {
			active.InitialScore = score
		}
	}
}

// AddPendingChange adds a proposed change to the revision
func (rt *RevisionTracker) AddPendingChange(revisionID string, changeType ChangeType, description, proposedText string, expectedImpact float64) string {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	changeID := uuid.New().String()
	
	if active, exists := rt.active[revisionID]; exists {
		change := PendingChange{
			ID:             changeID,
			Type:           changeType,
			Description:    description,
			ProposedText:   proposedText,
			ExpectedImpact: expectedImpact,
			Priority:       len(active.Changes) + 1,
			Status:         ChangePending,
			CreatedAt:      time.Now(),
		}
		active.Changes = append(active.Changes, change)
	}

	return changeID
}

// ApplyChange applies a pending change and tracks the result
func (rt *RevisionTracker) ApplyChange(revisionID, changeID string, beforeText, afterText string, actualImpact float64) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if active, exists := rt.active[revisionID]; exists {
		// Find and update the pending change
		for i, change := range active.Changes {
			if change.ID == changeID {
				active.Changes[i].Status = ChangeApplied
				break
			}
		}
		
		// Update current content
		active.CurrentContent = afterText
	}
}

// CompleteRevision finalizes a revision and creates a revision record
func (rt *RevisionTracker) CompleteRevision(revisionID string, result RevisionResult) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if active, exists := rt.active[revisionID]; exists {
		endTime := time.Now()
		
		// Create applied changes from pending changes
		appliedChanges := []AppliedChange{}
		for _, change := range active.Changes {
			if change.Status == ChangeApplied {
				appliedChanges = append(appliedChanges, AppliedChange{
					ID:             change.ID,
					Type:           change.Type,
					Description:    change.Description,
					ExpectedImpact: change.ExpectedImpact,
					ActualImpact:   0, // Would be calculated from before/after scores
					AppliedAt:      change.CreatedAt,
					SuccessRating:  0.8, // Default success rating
				})
			}
		}

		// Create revision record
		revisionRecord := RevisionRecord{
			ID:               revisionID,
			ContentID:        active.ContentID,
			RevisionNumber:   rt.getNextRevisionNumber(active.ContentID),
			StartTime:        active.StartTime,
			EndTime:          endTime,
			Duration:         endTime.Sub(active.StartTime),
			InitialScore:     active.InitialScore,
			FinalScore:       result.Score,
			ScoreImprovement: result.Score - active.InitialScore,
			QualityMetrics:   rt.extractQualityMetrics(active.QualityChecks),
			ChangesApplied:   appliedChanges,
			ImprovementAreas: rt.extractImprovementAreas(appliedChanges),
			Reviewer:         "quality_assurance_system",
			Status:           StatusCompleted,
			Metadata:         active.Metadata,
		}

		// Store revision record
		rt.revisions[active.ContentID] = append(rt.revisions[active.ContentID], revisionRecord)
		
		// Remove from active revisions
		delete(rt.active, revisionID)
	}
}

// GetRevisionHistory returns the revision history for a content item
func (rt *RevisionTracker) GetRevisionHistory(contentID uuid.UUID) []RevisionRecord {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	if records, exists := rt.revisions[contentID]; exists {
		return records
	}
	return []RevisionRecord{}
}

// GetRevisionAnalytics provides comprehensive analytics
func (rt *RevisionTracker) GetRevisionAnalytics() *RevisionAnalytics {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	analytics := &RevisionAnalytics{}
	
	allRevisions := []RevisionRecord{}
	for _, revisions := range rt.revisions {
		allRevisions = append(allRevisions, revisions...)
	}

	if len(allRevisions) == 0 {
		return analytics
	}

	// Calculate basic metrics
	totalImprovement := 0.0
	totalDuration := time.Duration(0)
	successCount := 0

	for _, revision := range allRevisions {
		totalImprovement += revision.ScoreImprovement
		totalDuration += revision.Duration
		if revision.ScoreImprovement > 0 {
			successCount++
		}
	}

	analytics.TotalRevisions = len(allRevisions)
	analytics.AverageImprovement = totalImprovement / float64(len(allRevisions))
	analytics.AverageRevisionTime = totalDuration / time.Duration(len(allRevisions))
	analytics.SuccessRate = float64(successCount) / float64(len(allRevisions))

	// Analyze improvement patterns
	analytics.CommonImprovements = rt.analyzeImprovementPatterns(allRevisions)

	// Calculate performance metrics
	analytics.PerformanceMetrics = rt.calculatePerformanceMetrics(allRevisions)

	// Analyze trends
	analytics.TrendAnalysis = rt.analyzeTrends(allRevisions)

	// Generate recommendations
	analytics.Recommendations = rt.generateSystemRecommendations(analytics)

	return analytics
}

// Helper methods

// getNextRevisionNumber gets the next revision number for a content item
func (rt *RevisionTracker) getNextRevisionNumber(contentID uuid.UUID) int {
	if revisions, exists := rt.revisions[contentID]; exists {
		return len(revisions) + 1
	}
	return 1
}

// extractQualityMetrics extracts quality metrics from check points
func (rt *RevisionTracker) extractQualityMetrics(checkPoints []QualityCheckPoint) map[string]float64 {
	metrics := make(map[string]float64)
	
	if len(checkPoints) > 0 {
		// Use the last check point as final metrics
		lastCheck := checkPoints[len(checkPoints)-1]
		for criterion, score := range lastCheck.CriteriaScores {
			metrics[criterion] = score
		}
		metrics["overall"] = lastCheck.OverallScore
	}
	
	return metrics
}

// extractImprovementAreas extracts improvement areas from applied changes
func (rt *RevisionTracker) extractImprovementAreas(changes []AppliedChange) []string {
	areas := make(map[string]bool)
	
	for _, change := range changes {
		if change.Category != "" {
			areas[change.Category] = true
		}
		areas[string(change.Type)] = true
	}
	
	result := []string{}
	for area := range areas {
		result = append(result, area)
	}
	
	return result
}

// analyzeImprovementPatterns identifies common improvement patterns
func (rt *RevisionTracker) analyzeImprovementPatterns(revisions []RevisionRecord) []ImprovementPattern {
	patterns := make(map[string]*ImprovementPattern)
	
	for _, revision := range revisions {
		for _, change := range revision.ChangesApplied {
			pattern := string(change.Type)
			
			if existing, exists := patterns[pattern]; exists {
				existing.Frequency++
				existing.AverageImpact = (existing.AverageImpact + change.ActualImpact) / 2
				if change.SuccessRating > 0.7 {
					existing.SuccessRate += 1
				}
			} else {
				patterns[pattern] = &ImprovementPattern{
					Pattern:       pattern,
					Frequency:     1,
					AverageImpact: change.ActualImpact,
					SuccessRate:   0,
					Category:      change.Category,
				}
				if change.SuccessRating > 0.7 {
					patterns[pattern].SuccessRate = 1
				}
			}
		}
	}
	
	// Convert to slice and calculate final success rates
	result := []ImprovementPattern{}
	for _, pattern := range patterns {
		pattern.SuccessRate = pattern.SuccessRate / float64(pattern.Frequency)
		result = append(result, *pattern)
	}
	
	return result
}

// calculatePerformanceMetrics calculates system performance metrics
func (rt *RevisionTracker) calculatePerformanceMetrics(revisions []RevisionRecord) PerformanceMetrics {
	if len(revisions) == 0 {
		return PerformanceMetrics{}
	}

	totalDuration := time.Duration(0)
	errorCount := 0
	qualityDistribution := make(map[string]int)

	for _, revision := range revisions {
		totalDuration += revision.Duration
		
		if revision.Status == StatusFailed {
			errorCount++
		}
		
		// Categorize final scores
		score := revision.FinalScore
		switch {
		case score >= 90:
			qualityDistribution["excellent"]++
		case score >= 80:
			qualityDistribution["good"]++
		case score >= 70:
			qualityDistribution["satisfactory"]++
		default:
			qualityDistribution["poor"]++
		}
	}

	avgProcessingTime := totalDuration / time.Duration(len(revisions))
	errorRate := float64(errorCount) / float64(len(revisions))
	throughput := float64(len(revisions)) / totalDuration.Hours()

	return PerformanceMetrics{
		AverageProcessingTime: avgProcessingTime,
		ThroughputPerHour:     throughput,
		ErrorRate:             errorRate,
		QualityDistribution:   qualityDistribution,
		ImprovementVelocity:   rt.calculateImprovementVelocity(revisions),
	}
}

// calculateImprovementVelocity calculates the rate of quality improvement
func (rt *RevisionTracker) calculateImprovementVelocity(revisions []RevisionRecord) float64 {
	if len(revisions) < 2 {
		return 0
	}

	totalImprovement := 0.0
	for _, revision := range revisions {
		totalImprovement += revision.ScoreImprovement
	}

	// Calculate improvement per day
	firstRevision := revisions[0]
	lastRevision := revisions[len(revisions)-1]
	
	duration := lastRevision.EndTime.Sub(firstRevision.StartTime)
	days := duration.Hours() / 24
	
	if days > 0 {
		return totalImprovement / days
	}
	
	return 0
}

// analyzeTrends analyzes trends in revision performance
func (rt *RevisionTracker) analyzeTrends(revisions []RevisionRecord) TrendAnalysis {
	// Simplified trend analysis
	trend := TrendAnalysis{
		QualityTrend:    TrendStable,
		EfficiencyTrend: TrendStable,
		TimeSeriesData:  []TimeSeriesPoint{},
	}

	// Create time series data
	for _, revision := range revisions {
		trend.TimeSeriesData = append(trend.TimeSeriesData, TimeSeriesPoint{
			Timestamp: revision.EndTime,
			Value:     revision.FinalScore,
			Metric:    "quality_score",
		})
	}

	// Simple trend detection (compare first and last quarters)
	if len(revisions) >= 4 {
		firstQuarter := revisions[:len(revisions)/4]
		lastQuarter := revisions[len(revisions)*3/4:]
		
		firstAvg := 0.0
		lastAvg := 0.0
		
		for _, rev := range firstQuarter {
			firstAvg += rev.FinalScore
		}
		firstAvg /= float64(len(firstQuarter))
		
		for _, rev := range lastQuarter {
			lastAvg += rev.FinalScore
		}
		lastAvg /= float64(len(lastQuarter))
		
		if lastAvg > firstAvg+5 {
			trend.QualityTrend = TrendImproving
		} else if lastAvg < firstAvg-5 {
			trend.QualityTrend = TrendDeclining
		}
	}

	return trend
}

// generateSystemRecommendations generates system improvement recommendations
func (rt *RevisionTracker) generateSystemRecommendations(analytics *RevisionAnalytics) []SystemRecommendation {
	recommendations := []SystemRecommendation{}

	// Low success rate recommendation
	if analytics.SuccessRate < 0.7 {
		recommendations = append(recommendations, SystemRecommendation{
			Type:           RecommendationProcess,
			Priority:       1,
			Title:          "Improve Success Rate",
			Description:    "Current revision success rate is below 70%",
			Impact:         "Higher success rate will improve overall system effectiveness",
			Implementation: "Review and optimize improvement suggestion algorithms",
		})
	}

	// Slow processing recommendation
	if analytics.PerformanceMetrics.AverageProcessingTime > 10*time.Minute {
		recommendations = append(recommendations, SystemRecommendation{
			Type:           RecommendationTechnical,
			Priority:       2,
			Title:          "Optimize Processing Time",
			Description:    "Average processing time exceeds 10 minutes",
			Impact:         "Faster processing will improve user experience",
			Implementation: "Optimize algorithms and consider parallel processing",
		})
	}

	return recommendations
}
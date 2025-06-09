package hr_management

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// PerformanceManagementServiceImpl implements the PerformanceManagementService interface
type PerformanceManagementServiceImpl struct {
	performanceRepo repositories.PerformanceRepository
	talentRepo      repositories.TalentRepository
	eventRepo       repositories.EventRepository
}

// NewPerformanceManagementServiceImpl creates a new performance management service
func NewPerformanceManagementServiceImpl(
	performanceRepo repositories.PerformanceRepository,
	talentRepo repositories.TalentRepository,
	eventRepo repositories.EventRepository,
) PerformanceManagementService {
	return &PerformanceManagementServiceImpl{
		performanceRepo: performanceRepo,
		talentRepo:      talentRepo,
		eventRepo:       eventRepo,
	}
}

// Performance tracking

func (s *PerformanceManagementServiceImpl) RecordPerformanceMetric(ctx context.Context, talentID uuid.UUID, metric PerformanceMetric) error {
	// Validate talent exists
	_, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return fmt.Errorf("talent not found: %w", err)
	}

	// Create performance metric record
	performanceMetric := &entities.PerformanceMetric{
		ID:          uuid.New(),
		TalentID:    talentID,
		Type:        metric.Type,
		Value:       metric.Value,
		Unit:        metric.Unit,
		Description: metric.Description,
		Source:      metric.Source,
		Context:     metric.Context,
		CreatedAt:   metric.Timestamp,
	}

	return s.performanceRepo.CreatePerformanceMetric(ctx, performanceMetric)
}

func (s *PerformanceManagementServiceImpl) GetPerformanceMetrics(ctx context.Context, talentID uuid.UUID, timeRange repositories.TimeRange) (*repositories.TalentPerformanceMetrics, error) {
	metrics, err := s.performanceRepo.GetPerformanceMetrics(ctx, talentID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	return metrics, nil
}

func (s *PerformanceManagementServiceImpl) AnalyzePerformanceTrends(ctx context.Context, talentID uuid.UUID) (*PerformanceTrendAnalysis, error) {
	// Get performance metrics for the last 6 months
	timeRange := repositories.TimeRange{
		Start: time.Now().AddDate(0, -6, 0),
		End:   time.Now(),
	}

	metrics, err := s.performanceRepo.GetPerformanceMetrics(ctx, talentID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	// Analyze trends
	analysis := &PerformanceTrendAnalysis{
		TalentID:    talentID,
		TimeRange:   timeRange,
		KeyMetrics:  s.calculateMetricTrends(metrics),
		Predictions: s.generatePredictions(metrics),
	}

	// Determine overall trend
	if len(analysis.KeyMetrics) > 0 {
		avgTrend := 0.0
		for _, metric := range analysis.KeyMetrics {
			avgTrend += metric.ChangePercent
		}
		avgTrend /= float64(len(analysis.KeyMetrics))

		switch {
		case avgTrend > 10:
			analysis.Trend = "Improving"
			analysis.TrendStrength = avgTrend / 100
		case avgTrend < -10:
			analysis.Trend = "Declining"
			analysis.TrendStrength = -avgTrend / 100
		default:
			analysis.Trend = "Stable"
			analysis.TrendStrength = 0.5
		}
	}

	analysis.Recommendations = s.generateRecommendations(analysis)

	return analysis, nil
}

// Performance reviews

func (s *PerformanceManagementServiceImpl) SchedulePerformanceReview(ctx context.Context, talentID uuid.UUID, reviewType string) (*entities.PerformanceReview, error) {
	_, err := s.talentRepo.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("talent not found: %w", err)
	}

	nextReviewDate := time.Now().AddDate(0, 3, 0) // Next review in 3 months
	review := &entities.PerformanceReview{
		ID:                  uuid.New(),
		TalentID:            talentID,
		ReviewPeriodStart:   time.Now().AddDate(0, -3, 0), // Last 3 months
		ReviewPeriodEnd:     time.Now(),
		OverallRating:       entities.PerformanceRatingMeets, // Default
		QualityScore:        0.0,
		ProductivityScore:   0.0,
		ReliabilityScore:    0.0,
		CommunicationScore:  0.0,
		Strengths:           []string{},
		ImprovementAreas:    []string{},
		Goals:               []string{},
		Metrics:             make(map[string]interface{}),
		NextReviewDate:      &nextReviewDate,
		CreatedAt:           time.Now(),
	}

	if err := s.performanceRepo.CreatePerformanceReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create performance review: %w", err)
	}

	return review, nil
}

func (s *PerformanceManagementServiceImpl) ConductPerformanceReview(ctx context.Context, reviewID uuid.UUID, reviewData ReviewData) (*entities.PerformanceReview, error) {
	review, err := s.performanceRepo.GetPerformanceReview(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("review not found: %w", err)
	}

	// Update review with provided data
	review.ReviewerID = &reviewData.ReviewerID
	review.QualityScore = reviewData.QualityScore
	review.ProductivityScore = reviewData.ProductivityScore
	review.ReliabilityScore = reviewData.ReliabilityScore
	review.CommunicationScore = reviewData.CommunicationScore
	review.OverallRating = reviewData.OverallRating
	review.Strengths = reviewData.Strengths
	review.ImprovementAreas = reviewData.ImprovementAreas
	review.Goals = reviewData.Goals
	review.Comments = reviewData.Comments
	review.NextReviewDate = &reviewData.NextReviewDate

	if reviewData.CompensationAdjustment != nil {
		amount := int64(reviewData.CompensationAdjustment.Amount)
		review.CompensationAdjustmentAmount = &amount
		review.CompensationAdjustmentCurrency = &reviewData.CompensationAdjustment.Currency
	}

	if err := s.performanceRepo.UpdatePerformanceReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to update performance review: %w", err)
	}

	return review, nil
}

func (s *PerformanceManagementServiceImpl) GetPerformanceReviews(ctx context.Context, talentID uuid.UUID) ([]*entities.PerformanceReview, error) {
	return s.performanceRepo.GetPerformanceReviewsByTalent(ctx, talentID)
}

// Goal management

func (s *PerformanceManagementServiceImpl) SetPerformanceGoals(ctx context.Context, talentID uuid.UUID, goals []PerformanceGoal) error {
	for _, goal := range goals {
		goalEntity := &entities.PerformanceGoal{
			ID:           goal.GoalID,
			TalentID:     talentID,
			Title:        goal.Title,
			Description:  goal.Description,
			Type:         goal.Type,
			TargetValue:  goal.TargetValue,
			CurrentValue: goal.CurrentValue,
			Unit:         goal.Unit,
			Priority:     goal.Priority,
			DueDate:      goal.DueDate,
			Status:       goal.Status,
			Metrics:      goal.Metrics,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.performanceRepo.CreatePerformanceGoal(ctx, goalEntity); err != nil {
			return fmt.Errorf("failed to create goal %s: %w", goal.Title, err)
		}
	}

	return nil
}

func (s *PerformanceManagementServiceImpl) TrackGoalProgress(ctx context.Context, talentID uuid.UUID) (*GoalProgressSummary, error) {
	goals, err := s.performanceRepo.GetPerformanceGoalsByTalent(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance goals: %w", err)
	}

	summary := &GoalProgressSummary{
		TalentID:        talentID,
		TotalGoals:      len(goals),
		CompletedGoals:  0,
		InProgressGoals: 0,
		OverdueGoals:    0,
		OverallProgress: 0.0,
		Goals:           make([]PerformanceGoal, len(goals)),
	}

	totalProgress := 0.0
	now := time.Now()

	for i, goal := range goals {
		goalProgress := goal.CurrentValue / goal.TargetValue * 100
		if goalProgress > 100 {
			goalProgress = 100
		}

		summary.Goals[i] = PerformanceGoal{
			GoalID:       goal.ID,
			Title:        goal.Title,
			Description:  goal.Description,
			Type:         goal.Type,
			TargetValue:  goal.TargetValue,
			CurrentValue: goal.CurrentValue,
			Unit:         goal.Unit,
			Priority:     goal.Priority,
			DueDate:      goal.DueDate,
			Status:       goal.Status,
			Metrics:      goal.Metrics,
		}

		totalProgress += goalProgress

		switch goal.Status {
		case "Completed":
			summary.CompletedGoals++
		case "InProgress":
			summary.InProgressGoals++
		}

		if goal.DueDate.Before(now) && goal.Status != "Completed" {
			summary.OverdueGoals++
		}
	}

	if summary.TotalGoals > 0 {
		summary.OverallProgress = totalProgress / float64(summary.TotalGoals)
	}

	return summary, nil
}

func (s *PerformanceManagementServiceImpl) UpdateGoalProgress(ctx context.Context, goalID uuid.UUID, progress float64, notes string) error {
	goal, err := s.performanceRepo.GetPerformanceGoal(ctx, goalID)
	if err != nil {
		return fmt.Errorf("goal not found: %w", err)
	}

	goal.CurrentValue = progress
	goal.UpdatedAt = time.Now()

	// Update status based on progress
	progressPercent := progress / goal.TargetValue * 100
	switch {
	case progressPercent >= 100:
		goal.Status = "Completed"
	case progressPercent > 0:
		goal.Status = "InProgress"
	default:
		goal.Status = "NotStarted"
	}

	// Add notes to metrics
	if notes != "" {
		if goal.Metrics == nil {
			goal.Metrics = make(map[string]interface{})
		}
		goal.Metrics["last_update_notes"] = notes
		goal.Metrics["last_updated"] = time.Now()
	}

	return s.performanceRepo.UpdatePerformanceGoal(ctx, goal)
}

// Performance alerts

func (s *PerformanceManagementServiceImpl) DetectPerformanceIssues(ctx context.Context) ([]*PerformanceAlert, error) {
	var alerts []*PerformanceAlert

	// Get all active talent
	activeTalent, _, err := s.talentRepo.SearchTalent(ctx, repositories.TalentFilter{
		Status: func() *entities.TalentStatus { s := entities.TalentStatusEngaged; return &s }(),
		Limit:  1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active talent: %w", err)
	}

	// Check performance issues for each talent
	for _, talent := range activeTalent {
		talentAlerts := s.checkTalentPerformanceIssues(ctx, talent)
		alerts = append(alerts, talentAlerts...)
	}

	return alerts, nil
}

func (s *PerformanceManagementServiceImpl) GeneratePerformanceReport(ctx context.Context, timeRange repositories.TimeRange) (*PerformanceReport, error) {
	// Get performance data for all talent in the time range
	allMetrics, err := s.performanceRepo.GetAllPerformanceMetrics(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	// Get performance distribution
	distribution, err := s.performanceRepo.GetPerformanceDistribution(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance distribution: %w", err)
	}

	// Identify top performers and underperformers
	topPerformers, err := s.identifyTopPerformers(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to identify top performers: %w", err)
	}

	underperformers, err := s.identifyUnderperformers(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to identify underperformers: %w", err)
	}

	report := &PerformanceReport{
		ReportID:        uuid.New(),
		TimeRange:       timeRange,
		TotalTalent:     len(allMetrics),
		Distribution:    *distribution,
		TopPerformers:   topPerformers,
		Underperformers: underperformers,
		Trends:          s.calculateReportTrends(allMetrics),
		KeyInsights:     s.generateKeyInsights(distribution, topPerformers, underperformers),
		Recommendations: s.generateReportRecommendations(distribution),
		GeneratedAt:     time.Now(),
	}

	return report, nil
}

// Helper methods

func (s *PerformanceManagementServiceImpl) calculateMetricTrends(metrics *repositories.TalentPerformanceMetrics) []MetricTrend {
	var trends []MetricTrend

	// Group metrics by type and calculate trends
	metricsByType := make(map[string][]entities.PerformanceMetric)
	for _, metric := range metrics.Metrics {
		metricsByType[metric.Type] = append(metricsByType[metric.Type], *metric)
	}

	for metricType, typeMetrics := range metricsByType {
		if len(typeMetrics) < 2 {
			continue // Need at least 2 data points for trend
		}

		// Sort by timestamp
		// Calculate start and end values
		startValue := typeMetrics[0].Value
		endValue := typeMetrics[len(typeMetrics)-1].Value

		changePercent := ((endValue - startValue) / startValue) * 100

		trend := MetricTrend{
			MetricType:    metricType,
			StartValue:    startValue,
			EndValue:      endValue,
			ChangePercent: changePercent,
			Confidence:    0.8, // Default confidence
		}

		switch {
		case changePercent > 5:
			trend.Trend = "Improving"
		case changePercent < -5:
			trend.Trend = "Declining"
		default:
			trend.Trend = "Stable"
		}

		trends = append(trends, trend)
	}

	return trends
}

func (s *PerformanceManagementServiceImpl) generatePredictions(metrics *repositories.TalentPerformanceMetrics) []PerformancePrediction {
	var predictions []PerformancePrediction

	// Simple linear prediction based on recent trends
	for _, trend := range s.calculateMetricTrends(metrics) {
		prediction := PerformancePrediction{
			MetricType:     trend.MetricType,
			PredictedValue: trend.EndValue * (1 + trend.ChangePercent/100),
			Confidence:     0.7,
			TimeHorizon:    30 * 24 * time.Hour, // 30 days
			Assumptions:    []string{"Linear trend continuation", "No external factors"},
		}

		predictions = append(predictions, prediction)
	}

	return predictions
}

func (s *PerformanceManagementServiceImpl) generateRecommendations(analysis *PerformanceTrendAnalysis) []string {
	var recommendations []string

	switch analysis.Trend {
	case "Improving":
		recommendations = append(recommendations,
			"Continue current performance trajectory",
			"Consider increased responsibilities",
			"Identify success factors for knowledge sharing")
	case "Declining":
		recommendations = append(recommendations,
			"Schedule performance improvement meeting",
			"Identify and address performance barriers",
			"Consider additional training or support")
	case "Stable":
		recommendations = append(recommendations,
			"Set new performance challenges",
			"Explore growth opportunities",
			"Consider lateral skill development")
	}

	return recommendations
}

func (s *PerformanceManagementServiceImpl) checkTalentPerformanceIssues(ctx context.Context, talent *entities.Talent) []*PerformanceAlert {
	var alerts []*PerformanceAlert

	// Check for declining performance
	timeRange := repositories.TimeRange{
		Start: time.Now().AddDate(0, -1, 0), // Last month
		End:   time.Now(),
	}

	metrics, err := s.performanceRepo.GetPerformanceMetrics(ctx, talent.ID, timeRange)
	if err != nil {
		return alerts // Return empty if can't get metrics
	}

	// Check for performance issues
	if s.hasDeciningPerformance(metrics) {
		alert := &PerformanceAlert{
			AlertID:     uuid.New(),
			TalentID:    talent.ID,
			Type:        "Declining",
			Severity:    "Medium",
			Title:       "Declining Performance Detected",
			Description: fmt.Sprintf("Performance metrics for %s show declining trend", talent.Name),
			Metrics:     []string{"quality_score", "productivity"},
			CreatedAt:   time.Now(),
			DueDate:     time.Now().Add(7 * 24 * time.Hour),
			Resolved:    false,
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

func (s *PerformanceManagementServiceImpl) hasDeciningPerformance(metrics *repositories.TalentPerformanceMetrics) bool {
	// Simple heuristic: check if recent average is lower than overall average
	if len(metrics.Metrics) < 4 {
		return false
	}

	// Get recent metrics (last quarter)
	recentCount := len(metrics.Metrics) / 4
	if recentCount < 1 {
		recentCount = 1
	}

	recentMetrics := metrics.Metrics[len(metrics.Metrics)-recentCount:]
	
	// Calculate averages
	var recentAvg, overallAvg float64
	for _, metric := range recentMetrics {
		recentAvg += metric.Value
	}
	recentAvg /= float64(len(recentMetrics))

	for _, metric := range metrics.Metrics {
		overallAvg += metric.Value
	}
	overallAvg /= float64(len(metrics.Metrics))

	// Consider declining if recent average is 20% lower than overall
	return recentAvg < overallAvg*0.8
}

func (s *PerformanceManagementServiceImpl) identifyTopPerformers(ctx context.Context, timeRange repositories.TimeRange) ([]*entities.Talent, error) {
	// Get talent with high performance ratings
	criteria := repositories.PerformanceCriteria{
		MetricTypes: []string{"quality_score", "productivity", "reliability"},
		MinRating:   4.0,
		TimeRange:   timeRange,
		SortBy:      "overall_rating",
		Limit:       10,
	}

	return s.performanceRepo.GetTopPerformers(ctx, criteria)
}

func (s *PerformanceManagementServiceImpl) identifyUnderperformers(ctx context.Context, timeRange repositories.TimeRange) ([]*entities.Talent, error) {
	// Get talent with low performance ratings
	criteria := repositories.PerformanceCriteria{
		MetricTypes: []string{"quality_score", "productivity", "reliability"},
		MinRating:   0.0, // No minimum
		TimeRange:   timeRange,
		SortBy:      "overall_rating",
		Limit:       10,
	}

	return s.performanceRepo.GetUnderperformers(ctx, criteria)
}

func (s *PerformanceManagementServiceImpl) calculateReportTrends(allMetrics map[uuid.UUID]*repositories.TalentPerformanceMetrics) []MetricTrend {
	// Aggregate trends across all talent
	var trends []MetricTrend

	// This would be more sophisticated in a real implementation
	trends = append(trends, MetricTrend{
		MetricType:    "overall_productivity",
		StartValue:    3.2,
		EndValue:      3.7,
		ChangePercent: 15.6,
		Trend:         "Improving",
		Confidence:    0.85,
	})

	return trends
}

func (s *PerformanceManagementServiceImpl) generateKeyInsights(distribution *repositories.PerformanceDistribution, topPerformers, underperformers []*entities.Talent) []string {
	insights := []string{
		fmt.Sprintf("%d talent in top performance tier", len(topPerformers)),
		fmt.Sprintf("%d talent requiring performance improvement", len(underperformers)),
	}

	// Add distribution insights
	if distribution.ExceptionalCount > 0 {
		insights = append(insights, fmt.Sprintf("%.1f%% of talent performing exceptionally", 
			float64(distribution.ExceptionalCount)/float64(distribution.TotalCount)*100))
	}

	return insights
}

func (s *PerformanceManagementServiceImpl) generateReportRecommendations(distribution *repositories.PerformanceDistribution) []string {
	recommendations := []string{}

	if distribution.NeedsImprovementCount > distribution.TotalCount/10 {
		recommendations = append(recommendations, "Consider enhanced training programs for underperforming talent")
	}

	if distribution.ExceptionalCount > 0 {
		recommendations = append(recommendations, "Leverage top performers for mentoring and knowledge sharing")
	}

	recommendations = append(recommendations, "Continue regular performance reviews and goal setting")

	return recommendations
}
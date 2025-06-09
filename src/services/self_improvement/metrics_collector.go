package self_improvement

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// metricsCollector implements MetricsCollector interface
type metricsCollector struct {
	projectRepo  repositories.ProjectRepository
	contentRepo  repositories.ContentRepository
	decisionRepo repositories.DecisionRepository
	feedbackRepo repositories.FeedbackRepository
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(
	projectRepo repositories.ProjectRepository,
	contentRepo repositories.ContentRepository,
	decisionRepo repositories.DecisionRepository,
	feedbackRepo repositories.FeedbackRepository,
) MetricsCollector {
	return &metricsCollector{
		projectRepo:  projectRepo,
		contentRepo:  contentRepo,
		decisionRepo: decisionRepo,
		feedbackRepo: feedbackRepo,
	}
}

// CollectContentMetrics collects metrics from content creation
func (mc *metricsCollector) CollectContentMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error) {
	var metrics []*entities.SystemPerformanceMetric
	
	// Get content by status to approximate recent content
	recentContent, _, err := mc.contentRepo.FindByStatus(ctx, entities.ContentStatusPublished, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("getting recent content: %w", err)
	}
	
	// Calculate average quality score
	totalQuality := 0.0
	qualityCount := 0
	revisionTotal := 0
	
	for _, content := range recentContent {
		if content.Statistics != nil {
			qualityScore := (content.Statistics.ReadabilityScore + content.Statistics.SEOScore + content.Statistics.EngagementScore) / 3.0
			if qualityScore > 0 {
				totalQuality += qualityScore
				qualityCount++
			}
		}
		revisionTotal += content.Version
	}
	
	if qualityCount > 0 {
		avgQuality := totalQuality / float64(qualityCount)
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("content_quality_%d", time.Now().Unix()),
			Component:  "content_creation",
			MetricName: "quality_score",
			Value:      avgQuality,
			Unit:       "score",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": qualityCount,
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1h",
		})
	}
	
	// Average revision count
	if len(recentContent) > 0 {
		avgRevisions := float64(revisionTotal) / float64(len(recentContent))
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("content_revisions_%d", time.Now().Unix()),
			Component:  "content_creation",
			MetricName: "revision_rate",
			Value:      avgRevisions,
			Unit:       "count",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": len(recentContent),
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1h",
		})
	}
	
	// Content creation time (if we track it)
	// This would analyze ContentVersion timestamps to calculate average creation time
	
	// Content approval rate
	approved := 0
	for _, content := range recentContent {
		if content.Status == entities.ContentStatusApproved || content.Status == entities.ContentStatusPublished {
			approved++
		}
	}
	
	if len(recentContent) > 0 {
		approvalRate := float64(approved) / float64(len(recentContent)) * 100
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("content_approval_%d", time.Now().Unix()),
			Component:  "content_creation",
			MetricName: "approval_rate",
			Value:      approvalRate,
			Unit:       "percentage",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"approved":    approved,
				"total":       len(recentContent),
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1h",
		})
	}
	
	return metrics, nil
}

// CollectDecisionMetrics collects metrics from decision making
func (mc *metricsCollector) CollectDecisionMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error) {
	var metrics []*entities.SystemPerformanceMetric
	
	// Get recent decisions (using GetDecisionsByStatus)
	recentDecisions, err := mc.decisionRepo.GetDecisionsByStatus(ctx, entities.StatusExecuted)
	if err != nil {
		return nil, fmt.Errorf("getting recent decisions: %w", err)
	}
	
	// Calculate decision metrics
	totalConfidence := 0.0
	executedCount := 0
	successCount := 0
	totalExecutionTime := 0.0
	
	for _, decision := range recentDecisions {
		totalConfidence += decision.ConfidenceScore
		
		if decision.Status == entities.StatusExecuted {
			executedCount++
			if decision.ExecutionResult != nil && decision.ExecutionResult.Success {
				successCount++
			}
			if decision.ExecutedAt != nil {
				executionTime := decision.ExecutedAt.Sub(decision.CreatedAt).Seconds()
				totalExecutionTime += executionTime
			}
		}
	}
	
	if len(recentDecisions) > 0 {
		// Average confidence
		avgConfidence := totalConfidence / float64(len(recentDecisions))
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("decision_confidence_%d", time.Now().Unix()),
			Component:  "decision_making",
			MetricName: "confidence_score",
			Value:      avgConfidence,
			Unit:       "score",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": len(recentDecisions),
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1h",
		})
		
		// Success rate
		if executedCount > 0 {
			successRate := float64(successCount) / float64(executedCount) * 100
			metrics = append(metrics, &entities.SystemPerformanceMetric{
				ID:         fmt.Sprintf("decision_success_%d", time.Now().Unix()),
				Component:  "decision_making",
				MetricName: "success_rate",
				Value:      successRate,
				Unit:       "percentage",
				Timestamp:  time.Now(),
				Context: map[string]interface{}{
					"successful": successCount,
					"executed":   executedCount,
				},
				Aggregation: entities.AggregationAverage,
				Period:      "1h",
			})
			
			// Average execution time
			avgExecutionTime := totalExecutionTime / float64(executedCount)
			metrics = append(metrics, &entities.SystemPerformanceMetric{
				ID:         fmt.Sprintf("decision_exec_time_%d", time.Now().Unix()),
				Component:  "decision_making",
				MetricName: "execution_time",
				Value:      avgExecutionTime,
				Unit:       "seconds",
				Timestamp:  time.Now(),
				Context: map[string]interface{}{
					"sample_size": executedCount,
				},
				Aggregation: entities.AggregationAverage,
				Period:      "1h",
			})
		}
	}
	
	return metrics, nil
}

// CollectFinancialMetrics collects financial performance metrics
func (mc *metricsCollector) CollectFinancialMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error) {
	var metrics []*entities.SystemPerformanceMetric
	
	// Get recent projects for financial analysis (using FindByStatus)
	recentProjects, _, err := mc.projectRepo.FindByStatus(ctx, entities.ProjectStatusCompleted, 0, 30)
	if err != nil {
		return nil, fmt.Errorf("getting recent projects: %w", err)
	}
	
	// Calculate financial metrics
	totalRevenue := 0.0
	completedProjects := 0
	totalMargin := 0.0
	
	for _, project := range recentProjects {
		if project.Status == entities.ProjectStatusCompleted && project.Budget.Amount > 0 {
			totalRevenue += project.Budget.Amount
			completedProjects++
			
			// Simplified margin calculation (assume 30% profit margin)
			margin := 30.0 // Default profit margin percentage
			totalMargin += margin
		}
	}
	
	// Average revenue per project
	if completedProjects > 0 {
		avgRevenue := totalRevenue / float64(completedProjects)
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("financial_avg_revenue_%d", time.Now().Unix()),
			Component:  "pricing",
			MetricName: "average_revenue",
			Value:      avgRevenue,
			Unit:       "currency",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"completed_projects": completedProjects,
				"total_revenue":      totalRevenue,
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1d",
		})
		
		// Average margin
		avgMargin := totalMargin / float64(completedProjects)
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("financial_margin_%d", time.Now().Unix()),
			Component:  "pricing",
			MetricName: "profit_margin",
			Value:      avgMargin,
			Unit:       "percentage",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": completedProjects,
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1d",
		})
	}
	
	// Pricing accuracy (estimated vs actual)
	pricingVariance := 0.0
	pricingCount := 0
	
	for _, project := range recentProjects {
		if project.Status == entities.ProjectStatusCompleted && 
		   project.Budget.Amount > 0 {
			// Simplified variance calculation (assume 10% variance)
			variance := 10.0 // Default variance percentage
			pricingVariance += variance
			pricingCount++
		}
	}
	
	if pricingCount > 0 {
		avgVariance := pricingVariance / float64(pricingCount)
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("pricing_accuracy_%d", time.Now().Unix()),
			Component:  "pricing",
			MetricName: "pricing_variance",
			Value:      avgVariance,
			Unit:       "percentage",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": pricingCount,
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1d",
		})
	}
	
	return metrics, nil
}

// CollectClientMetrics collects client-related metrics
func (mc *metricsCollector) CollectClientMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error) {
	var metrics []*entities.SystemPerformanceMetric
	
	// Get recent feedback (using FindByStatus)
	recentFeedback, _, err := mc.feedbackRepo.FindByStatus(ctx, entities.FeedbackStatusResolved, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("getting recent feedback: %w", err)
	}
	
	// Calculate satisfaction metrics
	totalRating := 0.0
	ratingCount := 0
	positiveCount := 0
	negativeCount := 0
	
	for _, feedback := range recentFeedback {
		if feedback.Rating != nil {
			totalRating += feedback.Rating.Score
			ratingCount++
		}
		
		switch feedback.Type {
		case entities.FeedbackTypePositive, entities.FeedbackTypeTestimonial:
			positiveCount++
		case entities.FeedbackTypeNegative, entities.FeedbackTypeComplaint:
			negativeCount++
		}
	}
	
	// Average satisfaction rating
	if ratingCount > 0 {
		avgRating := totalRating / float64(ratingCount)
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("client_satisfaction_%d", time.Now().Unix()),
			Component:  "client",
			MetricName: "satisfaction_rating",
			Value:      avgRating,
			Unit:       "rating",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"sample_size": ratingCount,
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1d",
		})
	}
	
	// Sentiment ratio
	if len(recentFeedback) > 0 {
		sentimentRatio := float64(positiveCount-negativeCount) / float64(len(recentFeedback)) * 100
		metrics = append(metrics, &entities.SystemPerformanceMetric{
			ID:         fmt.Sprintf("client_sentiment_%d", time.Now().Unix()),
			Component:  "client",
			MetricName: "sentiment_ratio",
			Value:      sentimentRatio,
			Unit:       "percentage",
			Timestamp:  time.Now(),
			Context: map[string]interface{}{
				"positive":    positiveCount,
				"negative":    negativeCount,
				"total":       len(recentFeedback),
			},
			Aggregation: entities.AggregationAverage,
			Period:      "1d",
		})
	}
	
	// Response time (if tracked)
	// This would analyze response times to client requests
	
	// Retention rate (returning clients)
	// This would analyze client project history
	
	return metrics, nil
}

// CollectSystemMetrics collects overall system health metrics
func (mc *metricsCollector) CollectSystemMetrics(ctx context.Context) ([]*entities.SystemPerformanceMetric, error) {
	var metrics []*entities.SystemPerformanceMetric
	
	// System availability (assumed to be tracked elsewhere)
	// For now, we'll simulate with a high availability
	metrics = append(metrics, &entities.SystemPerformanceMetric{
		ID:         fmt.Sprintf("system_availability_%d", time.Now().Unix()),
		Component:  "system",
		MetricName: "availability",
		Value:      99.5,
		Unit:       "percentage",
		Timestamp:  time.Now(),
		Context:    map[string]interface{}{},
		Aggregation: entities.AggregationAverage,
		Period:      "1h",
	})
	
	// Error rate (would be collected from logs/monitoring)
	metrics = append(metrics, &entities.SystemPerformanceMetric{
		ID:         fmt.Sprintf("system_error_rate_%d", time.Now().Unix()),
		Component:  "system",
		MetricName: "error_rate",
		Value:      0.5,
		Unit:       "percentage",
		Timestamp:  time.Now(),
		Context:    map[string]interface{}{},
		Aggregation: entities.AggregationAverage,
		Period:      "1h",
	})
	
	// API response time (would be collected from API monitoring)
	metrics = append(metrics, &entities.SystemPerformanceMetric{
		ID:         fmt.Sprintf("api_response_time_%d", time.Now().Unix()),
		Component:  "api",
		MetricName: "response_time",
		Value:      150,
		Unit:       "milliseconds",
		Timestamp:  time.Now(),
		Context:    map[string]interface{}{},
		Aggregation: entities.AggregationP95,
		Period:      "1h",
	})
	
	// Resource utilization (CPU, memory, etc.)
	// This would be collected from system monitoring
	
	return metrics, nil
}

// AggregateMetrics aggregates metrics into a summary
func (mc *metricsCollector) AggregateMetrics(ctx context.Context, metrics []*entities.SystemPerformanceMetric) (*MetricsSummary, error) {
	summary := &MetricsSummary{
		Period:         "1h",
		ComponentStats: make(map[string]ComponentStats),
		OverallHealth:  0,
		TopPerformers:  []string{},
		NeedsAttention: []string{},
		Recommendations: []string{},
	}
	
	// Group metrics by component
	componentMetrics := make(map[string][]*entities.SystemPerformanceMetric)
	for _, metric := range metrics {
		componentMetrics[metric.Component] = append(componentMetrics[metric.Component], metric)
	}
	
	// Calculate stats for each component
	totalHealth := 0.0
	componentCount := 0
	
	for component, compMetrics := range componentMetrics {
		stats := mc.calculateComponentStats(compMetrics)
		summary.ComponentStats[component] = stats
		
		totalHealth += stats.Health
		componentCount++
		
		// Identify top performers and those needing attention
		if stats.Health > 90 {
			summary.TopPerformers = append(summary.TopPerformers, component)
		} else if stats.Health < 70 {
			summary.NeedsAttention = append(summary.NeedsAttention, component)
		}
	}
	
	// Calculate overall health
	if componentCount > 0 {
		summary.OverallHealth = totalHealth / float64(componentCount)
	}
	
	// Generate recommendations
	if summary.OverallHealth < 80 {
		summary.Recommendations = append(summary.Recommendations, 
			"Overall system health is below target. Focus on improving underperforming components.")
	}
	
	if len(summary.NeedsAttention) > 0 {
		summary.Recommendations = append(summary.Recommendations,
			fmt.Sprintf("Components needing attention: %v", summary.NeedsAttention))
	}
	
	return summary, nil
}

func (mc *metricsCollector) calculateComponentStats(metrics []*entities.SystemPerformanceMetric) ComponentStats {
	stats := ComponentStats{
		Health:       85.0, // Default health
		Availability: 99.0, // Default availability
		Performance:  0,
		ErrorRate:    0,
		Metrics:      make(map[string]float64),
	}
	
	// Process each metric
	for _, metric := range metrics {
		stats.Metrics[metric.MetricName] = metric.Value
		
		// Update specific stats based on metric type
		switch metric.MetricName {
		case "availability":
			stats.Availability = metric.Value
		case "error_rate":
			stats.ErrorRate = metric.Value
		case "quality_score", "success_rate", "satisfaction_rating":
			// These contribute to performance
			stats.Performance += metric.Value / 3 // Simple average
		}
	}
	
	// Calculate overall health
	// Health = (Availability * 0.3) + (100 - ErrorRate) * 0.3 + Performance * 0.4
	stats.Health = stats.Availability*0.3 + (100-stats.ErrorRate)*0.3 + stats.Performance*0.4
	
	return stats
}
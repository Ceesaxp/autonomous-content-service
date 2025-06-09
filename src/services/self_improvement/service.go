package self_improvement

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// Service implements the self-improvement service
type Service struct {
	learningRepo     repositories.LearningRepository
	metricsRepo      repositories.MetricsRepository
	experimentRepo   repositories.ExperimentRepository
	capabilityRepo   repositories.CapabilityRepository
	promptRepo       repositories.PromptRepository
	projectRepo      repositories.ProjectRepository
	feedbackRepo     repositories.FeedbackRepository
	contentRepo      repositories.ContentRepository
	decisionRepo     repositories.DecisionRepository
	metricsCollector MetricsCollector
	learningEngine   LearningEngine
	experimentEngine ExperimentEngine
	capabilityAcq    CapabilityAcquisition
	optimizationEng  OptimizationEngine
}

// NewService creates a new self-improvement service
func NewService(
	learningRepo repositories.LearningRepository,
	metricsRepo repositories.MetricsRepository,
	experimentRepo repositories.ExperimentRepository,
	capabilityRepo repositories.CapabilityRepository,
	promptRepo repositories.PromptRepository,
	projectRepo repositories.ProjectRepository,
	feedbackRepo repositories.FeedbackRepository,
	contentRepo repositories.ContentRepository,
	decisionRepo repositories.DecisionRepository,
) *Service {
	s := &Service{
		learningRepo:   learningRepo,
		metricsRepo:    metricsRepo,
		experimentRepo: experimentRepo,
		capabilityRepo: capabilityRepo,
		promptRepo:     promptRepo,
		projectRepo:    projectRepo,
		feedbackRepo:   feedbackRepo,
		contentRepo:    contentRepo,
		decisionRepo:   decisionRepo,
	}
	
	// Initialize sub-components
	s.metricsCollector = NewMetricsCollector(projectRepo, contentRepo, decisionRepo, feedbackRepo)
	s.learningEngine = NewLearningEngine(learningRepo)
	s.experimentEngine = NewExperimentEngine(experimentRepo)
	s.capabilityAcq = NewCapabilityAcquisition(capabilityRepo)
	s.optimizationEng = NewOptimizationEngine(promptRepo, metricsRepo)
	
	return s
}

// CollectMetrics collects performance metrics from all system components
func (s *Service) CollectMetrics(ctx context.Context) error {
	// Collect from all sources
	contentMetrics, err := s.metricsCollector.CollectContentMetrics(ctx)
	if err != nil {
		return fmt.Errorf("collecting content metrics: %w", err)
	}
	
	decisionMetrics, err := s.metricsCollector.CollectDecisionMetrics(ctx)
	if err != nil {
		return fmt.Errorf("collecting decision metrics: %w", err)
	}
	
	financialMetrics, err := s.metricsCollector.CollectFinancialMetrics(ctx)
	if err != nil {
		return fmt.Errorf("collecting financial metrics: %w", err)
	}
	
	clientMetrics, err := s.metricsCollector.CollectClientMetrics(ctx)
	if err != nil {
		return fmt.Errorf("collecting client metrics: %w", err)
	}
	
	systemMetrics, err := s.metricsCollector.CollectSystemMetrics(ctx)
	if err != nil {
		return fmt.Errorf("collecting system metrics: %w", err)
	}
	
	// Store all metrics
	allMetrics := append(contentMetrics, decisionMetrics...)
	allMetrics = append(allMetrics, financialMetrics...)
	allMetrics = append(allMetrics, clientMetrics...)
	allMetrics = append(allMetrics, systemMetrics...)
	
	for _, metric := range allMetrics {
		if err := s.metricsRepo.RecordMetric(ctx, metric); err != nil {
			return fmt.Errorf("recording metric %s: %w", metric.MetricName, err)
		}
	}
	
	// Detect anomalies
	anomalies, err := s.DetectAnomalies(ctx)
	if err != nil {
		return fmt.Errorf("detecting anomalies: %w", err)
	}
	
	// Create learning artifacts from anomalies
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" || anomaly.Severity == "high" {
			artifact := &entities.LearningArtifact{
				Type:        entities.LearningTypePattern,
				Category:    "anomaly",
				Title:       fmt.Sprintf("Anomaly in %s.%s", anomaly.Component, anomaly.Metric),
				Description: fmt.Sprintf("Detected %s anomaly: value %f (expected %f)", anomaly.Severity, anomaly.Value, anomaly.Expected),
				Source:      entities.SourceSystemMonitoring,
				Evidence: []entities.Evidence{{
					Type:        "anomaly_detection",
					Description: "Statistical anomaly detected",
					Data: map[string]interface{}{
						"value":     anomaly.Value,
						"expected":  anomaly.Expected,
						"deviation": anomaly.Deviation,
					},
					Timestamp: anomaly.Timestamp,
					Strength:  0.8,
				}},
				Confidence:  0.9,
				ImpactScore: s.calculateAnomalyImpact(anomaly),
				Status:      entities.ArtifactStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			
			if err := s.learningRepo.CreateArtifact(ctx, artifact); err != nil {
				return fmt.Errorf("creating anomaly artifact: %w", err)
			}
		}
	}
	
	return nil
}

// AnalyzePerformance analyzes performance for a specific component
func (s *Service) AnalyzePerformance(ctx context.Context, component string, period string) (*PerformanceAnalysis, error) {
	// Parse period
	duration, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("parsing period: %w", err)
	}
	
	from := time.Now().Add(-duration)
	to := time.Now()
	
	// Get component metrics
	componentMetrics, err := s.metricsRepo.GetComponentMetrics(ctx, component)
	if err != nil {
		return nil, fmt.Errorf("getting component metrics: %w", err)
	}
	
	analysis := &PerformanceAnalysis{
		Component:   component,
		Period:      period,
		Metrics:     make(map[string]MetricStats),
		Trends:      make(map[string]Trend),
		Comparisons: make(map[string]Comparison),
		Insights:    []Insight{},
		Anomalies:   []*Anomaly{},
	}
	
	// Analyze each metric
	for metricName, currentValue := range componentMetrics {
		// Get historical data
		metrics, err := s.metricsRepo.GetMetrics(ctx, component, metricName, from, to)
		if err != nil {
			continue
		}
		
		// Calculate statistics
		stats := s.calculateMetricStats(metrics)
		analysis.Metrics[metricName] = stats
		
		// Analyze trends
		trend := s.analyzeTrend(metrics)
		analysis.Trends[metricName] = trend
		
		// Compare with baseline
		mean, stddev, err := s.metricsRepo.GetMetricBaseline(ctx, component, metricName, 30)
		if err == nil {
			comparison := Comparison{
				BaselineValue: mean,
				CurrentValue:  currentValue,
				Change:        currentValue - mean,
				ChangePercent: ((currentValue - mean) / mean) * 100,
				IsImprovement: s.isImprovement(metricName, currentValue-mean),
			}
			analysis.Comparisons[metricName] = comparison
			
			// Check for anomalies
			if math.Abs(currentValue-mean) > 2*stddev {
				anomaly := &Anomaly{
					Component: component,
					Metric:    metricName,
					Value:     currentValue,
					Expected:  mean,
					Deviation: (currentValue - mean) / stddev,
					Severity:  s.calculateAnomalySeverity((currentValue - mean) / stddev),
					Timestamp: time.Now(),
				}
				analysis.Anomalies = append(analysis.Anomalies, anomaly)
			}
		}
		
		// Generate insights
		if trend.Direction == "decreasing" && s.isImportantMetric(metricName) {
			analysis.Insights = append(analysis.Insights, Insight{
				Type:        "performance_degradation",
				Severity:    "high",
				Description: fmt.Sprintf("%s is showing a declining trend", metricName),
				Impact:      trend.Magnitude,
				Actions:     []string{"investigate_root_cause", "optimize_component"},
			})
		}
	}
	
	return analysis, nil
}

// DetectAnomalies detects anomalies across all components
func (s *Service) DetectAnomalies(ctx context.Context) ([]*Anomaly, error) {
	var anomalies []*Anomaly
	
	components := []string{"content_creation", "decision_making", "pricing", "payment", "onboarding"}
	
	for _, component := range components {
		metrics, err := s.metricsRepo.GetComponentMetrics(ctx, component)
		if err != nil {
			continue
		}
		
		for metricName, value := range metrics {
			// Get baseline
			mean, stddev, err := s.metricsRepo.GetMetricBaseline(ctx, component, metricName, 30)
			if err != nil {
				continue
			}
			
			// Check for anomaly (2 standard deviations)
			deviation := (value - mean) / stddev
			if math.Abs(deviation) > 2 {
				anomaly := &Anomaly{
					Component:   component,
					Metric:      metricName,
					Value:       value,
					Expected:    mean,
					Deviation:   deviation,
					Severity:    s.calculateAnomalySeverity(deviation),
					Timestamp:   time.Now(),
					PossibleCauses: s.identifyPossibleCauses(component, metricName, deviation),
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}
	
	return anomalies, nil
}

// LearnFromProject analyzes a completed project and extracts learnings
func (s *Service) LearnFromProject(ctx context.Context, projectIDStr string) ([]*entities.LearningArtifact, error) {
	var artifacts []*entities.LearningArtifact
	
	// Convert string ID to UUID
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}
	
	// Get project details
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	
	// Get project content
	content, err := s.contentRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("getting project content: %w", err)
	}
	
	// Get project feedback
	feedback, err := s.feedbackRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("getting project feedback: %w", err)
	}
	
	// Analyze project success
	success := s.evaluateProjectSuccess(project, content, feedback)
	
	// Extract patterns from successful projects
	if success.Score > 0.8 {
		// Content quality patterns
		if len(content) > 0 {
			avgQuality := s.calculateAverageQuality(content)
			if avgQuality > 85 {
				artifact := &entities.LearningArtifact{
					Type:        entities.LearningTypePattern,
					Category:    "content_quality",
					Title:       "High-quality content pattern",
					Description: fmt.Sprintf("Project %s achieved average quality score of %.1f", project.Title, avgQuality),
					Source:      entities.SourceProjectAnalysis,
					SourceID:    projectID.String(),
					Evidence: []entities.Evidence{{
						Type:        "quality_score",
						Description: "Average content quality score",
						Data:        map[string]interface{}{"score": avgQuality, "content_count": len(content)},
						Timestamp:   time.Now(),
						Strength:    0.9,
					}},
					Confidence:  0.85,
					ImpactScore: 0.8,
					Status:      entities.ArtifactStatusActive,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				artifacts = append(artifacts, artifact)
			}
		}
		
		// Client satisfaction patterns
		if success.ClientSatisfaction > 4.5 {
			artifact := &entities.LearningArtifact{
				Type:        entities.LearningTypePattern,
				Category:    "client_satisfaction",
				Title:       "High client satisfaction pattern",
				Description: fmt.Sprintf("Project %s achieved client satisfaction of %.1f/5", project.Title, success.ClientSatisfaction),
				Source:      entities.SourceProjectAnalysis,
				SourceID:    projectID.String(),
				Evidence:    s.extractSatisfactionEvidence(project, feedback),
				Confidence:  0.9,
				ImpactScore: 0.9,
				Status:      entities.ArtifactStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			artifacts = append(artifacts, artifact)
		}
	}
	
	// Extract learnings from failures or challenges
	if success.Score < 0.6 || len(success.Challenges) > 0 {
		for _, challenge := range success.Challenges {
			artifact := &entities.LearningArtifact{
				Type:        entities.LearningTypeException,
				Category:    "project_challenge",
				Title:       fmt.Sprintf("Challenge: %s", challenge.Type),
				Description: challenge.Description,
				Source:      entities.SourceProjectAnalysis,
				SourceID:    projectID.String(),
				Evidence: []entities.Evidence{{
					Type:        "challenge",
					Description: challenge.Description,
					Data:        map[string]interface{}{"impact": challenge.Impact, "resolution": challenge.Resolution},
					Timestamp:   time.Now(),
					Strength:    0.8,
				}},
				Confidence:  0.8,
				ImpactScore: challenge.Impact,
				Status:      entities.ArtifactStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			artifacts = append(artifacts, artifact)
		}
	}
	
	// Store artifacts
	for _, artifact := range artifacts {
		if err := s.learningRepo.CreateArtifact(ctx, artifact); err != nil {
			return nil, fmt.Errorf("creating learning artifact: %w", err)
		}
	}
	
	return artifacts, nil
}

// LearnFromFeedback extracts learnings from client feedback
func (s *Service) LearnFromFeedback(ctx context.Context, feedbackIDStr string) (*entities.LearningArtifact, error) {
	// Convert string ID to UUID
	feedbackID, err := uuid.Parse(feedbackIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid feedback ID: %w", err)
	}
	
	// Get feedback
	feedback, err := s.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("getting feedback: %w", err)
	}
	
	// Determine learning type based on feedback
	var learningType entities.LearningType
	var category string
	
	switch feedback.Type {
	case entities.FeedbackTypeComplaint:
		learningType = entities.LearningTypeConstraint
		category = "client_complaint"
	case entities.FeedbackTypeSuggestion:
		learningType = entities.LearningTypeOptimization
		category = "client_suggestion"
	case entities.FeedbackTypeTestimonial:
		learningType = entities.LearningTypePattern
		category = "success_pattern"
	default:
		learningType = entities.LearningTypeRule
		category = "client_feedback"
	}
	
	// Create learning artifact
	artifact := &entities.LearningArtifact{
		Type:        learningType,
		Category:    category,
		Title:       fmt.Sprintf("Learning from %s feedback", feedback.Type),
		Description: feedback.Message,
		Source:      entities.SourceClientFeedback,
		SourceID:    feedbackID.String(),
		Evidence: []entities.Evidence{{
			Type:        "client_feedback",
			Description: string(feedback.Type),
			Data: map[string]interface{}{
				"rating":    feedback.Rating,
				"source":    feedback.Source,
				"content_id": feedback.ContentID,
			},
			Timestamp: feedback.CreatedAt,
			Strength:  s.calculateFeedbackStrength(feedback),
		}},
		Confidence:  s.calculateFeedbackConfidence(feedback),
		ImpactScore: s.calculateFeedbackImpact(feedback),
		Tags:        feedback.Tags,
		Status:      entities.ArtifactStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Store artifact
	if err := s.learningRepo.CreateArtifact(ctx, artifact); err != nil {
		return nil, fmt.Errorf("creating learning artifact: %w", err)
	}
	
	// Update feedback status
	feedback.Status = entities.FeedbackStatusResolved
	if err := s.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, fmt.Errorf("updating feedback status: %w", err)
	}
	
	return artifact, nil
}

// LearnFromError extracts learnings from system errors
func (s *Service) LearnFromError(ctx context.Context, errorData ErrorData) (*entities.LearningArtifact, error) {
	// Create learning artifact from error
	artifact := &entities.LearningArtifact{
		Type:        entities.LearningTypeException,
		Category:    "system_error",
		Title:       fmt.Sprintf("Error pattern: %s", errorData.ErrorType),
		Description: errorData.Message,
		Source:      entities.SourceErrorAnalysis,
		Evidence: []entities.Evidence{{
			Type:        "error_trace",
			Description: "System error occurrence",
			Data: map[string]interface{}{
				"component":   errorData.Component,
				"error_type":  errorData.ErrorType,
				"frequency":   errorData.Frequency,
				"stack_trace": errorData.StackTrace,
			},
			Timestamp: errorData.LastSeen,
			Strength:  0.9,
		}},
		Confidence:  0.95,
		ImpactScore: s.calculateErrorImpact(errorData),
		Status:      entities.ArtifactStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Store artifact
	if err := s.learningRepo.CreateArtifact(ctx, artifact); err != nil {
		return nil, fmt.Errorf("creating error learning artifact: %w", err)
	}
	
	return artifact, nil
}

// SynthesizeLearnings combines related learnings into higher-level insights
func (s *Service) SynthesizeLearnings(ctx context.Context, period string) ([]*entities.LearningArtifact, error) {
	// Parse period
	duration, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("parsing period: %w", err)
	}
	
	// Get recent artifacts
	artifacts, err := s.learningRepo.GetActiveArtifacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting active artifacts: %w", err)
	}
	
	// Filter by period
	var recentArtifacts []*entities.LearningArtifact
	cutoff := time.Now().Add(-duration)
	for _, artifact := range artifacts {
		if artifact.CreatedAt.After(cutoff) {
			recentArtifacts = append(recentArtifacts, artifact)
		}
	}
	
	// Build knowledge graph
	graph, err := s.learningEngine.ConnectConcepts(ctx, recentArtifacts)
	if err != nil {
		return nil, fmt.Errorf("building knowledge graph: %w", err)
	}
	
	// Identify patterns and relationships
	var synthesized []*entities.LearningArtifact
	
	// Find clusters of related concepts
	for _, cluster := range graph.Clusters {
		if cluster.Coherence > 0.7 {
			// Create synthesis artifact
			artifact := &entities.LearningArtifact{
				Type:        entities.LearningTypeRelationship,
				Category:    "synthesis",
				Title:       cluster.Name,
				Description: cluster.Description,
				Source:      entities.SourceSystemMonitoring,
				Evidence:    s.extractClusterEvidence(cluster, recentArtifacts),
				Confidence:  cluster.Coherence,
				ImpactScore: s.calculateClusterImpact(cluster, recentArtifacts),
				RelatedArtifacts: cluster.NodeIDs,
				Status:      entities.ArtifactStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			synthesized = append(synthesized, artifact)
		}
	}
	
	// Identify contradictions
	contradictions, err := s.learningEngine.IdentifyContradictions(ctx)
	if err == nil {
		for _, contradiction := range contradictions {
			// Resolve contradiction
			resolution, err := s.learningEngine.ResolveContradiction(ctx, contradiction)
			if err == nil && resolution != nil {
				synthesized = append(synthesized, resolution)
			}
		}
	}
	
	// Store synthesized learnings
	for _, artifact := range synthesized {
		if err := s.learningRepo.CreateArtifact(ctx, artifact); err != nil {
			return nil, fmt.Errorf("creating synthesized artifact: %w", err)
		}
	}
	
	return synthesized, nil
}

// IdentifyCapabilityGaps identifies missing capabilities based on various signals
func (s *Service) IdentifyCapabilityGaps(ctx context.Context) ([]*entities.CapabilityGap, error) {
	// Get existing gaps
	existingGaps, err := s.capabilityRepo.ListCapabilityGaps(ctx, entities.GapStatusIdentified)
	if err != nil {
		return nil, fmt.Errorf("getting existing gaps: %w", err)
	}
	
	// Analyze recent failures and feedback
	gaps := make(map[string]*entities.CapabilityGap)
	
	// Check for content type requests we couldn't fulfill
	contentGaps := s.identifyContentCapabilityGaps(ctx)
	for _, gap := range contentGaps {
		gaps[gap.Description] = gap
	}
	
	// Check for integration requests
	integrationGaps := s.identifyIntegrationGaps(ctx)
	for _, gap := range integrationGaps {
		gaps[gap.Description] = gap
	}
	
	// Check for language/industry gaps
	domainGaps := s.identifyDomainGaps(ctx)
	for _, gap := range domainGaps {
		gaps[gap.Description] = gap
	}
	
	// Convert map to slice and calculate priorities
	var gapList []*entities.CapabilityGap
	for _, gap := range gaps {
		// Calculate priority based on frequency and impact
		gap.Priority = s.calculateGapPriority(gap)
		
		// Check if gap already exists
		exists := false
		for _, existing := range existingGaps {
			if existing.Description == gap.Description {
				exists = true
				// Update frequency
				existing.Frequency += gap.Frequency
				if err := s.capabilityRepo.UpdateCapabilityGap(ctx, existing); err != nil {
					return nil, fmt.Errorf("updating gap frequency: %w", err)
				}
				break
			}
		}
		
		if !exists {
			// Create new gap
			if err := s.capabilityRepo.CreateCapabilityGap(ctx, gap); err != nil {
				return nil, fmt.Errorf("creating capability gap: %w", err)
			}
			gapList = append(gapList, gap)
		}
	}
	
	return gapList, nil
}

// PrioritizeGaps prioritizes capability gaps based on ROI and impact
func (s *Service) PrioritizeGaps(ctx context.Context, gaps []*entities.CapabilityGap) ([]*entities.CapabilityGap, error) {
	// Calculate priority score for each gap
	for _, gap := range gaps {
		// ROI = (Impact * Frequency) / (Effort * Cost)
		roi := (gap.EstimatedImpact * float64(gap.Frequency)) / (gap.EstimatedEffort + 1)
		
		// Adjust for strategic importance
		if gap.Type == entities.CapabilityGapTypeAPI || gap.Type == entities.CapabilityGapTypeContent {
			roi *= 1.2
		}
		
		gap.Priority = roi
	}
	
	// Sort by priority
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].Priority > gaps[j].Priority
	})
	
	return gaps, nil
}

// Helper methods

func (s *Service) calculateMetricStats(metrics []*entities.SystemPerformanceMetric) MetricStats {
	if len(metrics) == 0 {
		return MetricStats{}
	}
	
	var values []float64
	for _, m := range metrics {
		values = append(values, m.Value)
	}
	
	sort.Float64s(values)
	
	// Calculate statistics
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate standard deviation
	var variance float64
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(values)))
	
	// Calculate percentiles
	p95Index := int(float64(len(values)) * 0.95)
	p99Index := int(float64(len(values)) * 0.99)
	
	return MetricStats{
		Mean:        mean,
		Median:      values[len(values)/2],
		StdDev:      stdDev,
		Min:         values[0],
		Max:         values[len(values)-1],
		P95:         values[p95Index],
		P99:         values[p99Index],
		Count:       len(values),
		LastValue:   metrics[len(metrics)-1].Value,
		LastUpdated: metrics[len(metrics)-1].Timestamp,
	}
}

func (s *Service) analyzeTrend(metrics []*entities.SystemPerformanceMetric) Trend {
	if len(metrics) < 2 {
		return Trend{Direction: "stable", Confidence: 0.5}
	}
	
	// Simple linear regression
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(metrics))
	
	for i, m := range metrics {
		x := float64(i)
		y := m.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// Determine direction
	direction := "stable"
	if slope > 0.1 {
		direction = "increasing"
	} else if slope < -0.1 {
		direction = "decreasing"
	}
	
	// Calculate R-squared for confidence
	meanY := sumY / n
	var ssTotal, ssResidual float64
	
	for i, m := range metrics {
		y := m.Value
		yPred := slope*float64(i) + (sumY-slope*sumX)/n
		ssTotal += math.Pow(y-meanY, 2)
		ssResidual += math.Pow(y-yPred, 2)
	}
	
	rSquared := 1 - (ssResidual / ssTotal)
	
	// Predict next value
	prediction := slope*n + (sumY-slope*sumX)/n
	
	return Trend{
		Direction:  direction,
		Magnitude:  math.Abs(slope),
		Confidence: rSquared,
		Prediction: prediction,
	}
}

func (s *Service) isImprovement(metricName string, change float64) bool {
	// Define which metrics improve when they increase
	increasingGood := map[string]bool{
		"quality_score":        true,
		"client_satisfaction":  true,
		"conversion_rate":      true,
		"success_rate":         true,
		"availability":         true,
		"performance":          true,
	}
	
	// Define which metrics improve when they decrease
	decreasingGood := map[string]bool{
		"error_rate":      true,
		"response_time":   true,
		"revision_rate":   true,
		"complaint_rate":  true,
		"cost_per_unit":   true,
	}
	
	if increasingGood[metricName] {
		return change > 0
	}
	if decreasingGood[metricName] {
		return change < 0
	}
	
	// Default: increasing is good
	return change > 0
}

func (s *Service) calculateAnomalySeverity(deviation float64) string {
	absDeviation := math.Abs(deviation)
	if absDeviation > 4 {
		return "critical"
	} else if absDeviation > 3 {
		return "high"
	} else if absDeviation > 2 {
		return "medium"
	}
	return "low"
}

func (s *Service) calculateAnomalyImpact(anomaly *Anomaly) float64 {
	// Base impact on severity
	impactMap := map[string]float64{
		"critical": 0.9,
		"high":     0.7,
		"medium":   0.5,
		"low":      0.3,
	}
	
	impact := impactMap[anomaly.Severity]
	
	// Adjust based on component importance
	componentWeights := map[string]float64{
		"content_creation": 1.2,
		"payment":          1.3,
		"decision_making":  1.1,
		"pricing":          1.0,
		"onboarding":       0.9,
	}
	
	if weight, ok := componentWeights[anomaly.Component]; ok {
		impact *= weight
	}
	
	return math.Min(impact, 1.0)
}

func (s *Service) isImportantMetric(metricName string) bool {
	importantMetrics := []string{
		"quality_score",
		"client_satisfaction",
		"conversion_rate",
		"error_rate",
		"availability",
		"revenue",
	}
	
	for _, important := range importantMetrics {
		if metricName == important {
			return true
		}
	}
	return false
}

func (s *Service) identifyPossibleCauses(component, metric string, deviation float64) []string {
	causes := []string{}
	
	// Generic causes based on deviation direction
	if deviation > 0 {
		causes = append(causes, "increased_load", "configuration_change", "external_dependency_issue")
	} else {
		causes = append(causes, "reduced_usage", "optimization_applied", "data_quality_issue")
	}
	
	// Component-specific causes
	switch component {
	case "content_creation":
		if metric == "quality_score" && deviation < 0 {
			causes = append(causes, "prompt_degradation", "llm_model_change", "context_overflow")
		}
	case "payment":
		if metric == "failure_rate" && deviation > 0 {
			causes = append(causes, "payment_provider_issue", "fraud_detection_trigger", "network_connectivity")
		}
	}
	
	return causes
}
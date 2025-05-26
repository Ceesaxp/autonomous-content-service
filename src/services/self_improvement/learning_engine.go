package self_improvement

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// learningEngine implements LearningEngine interface
type learningEngine struct {
	learningRepo repositories.LearningRepository
}

// NewLearningEngine creates a new learning engine
func NewLearningEngine(learningRepo repositories.LearningRepository) LearningEngine {
	return &learningEngine{
		learningRepo: learningRepo,
	}
}

// IdentifyPatterns identifies patterns in data points
func (le *learningEngine) IdentifyPatterns(ctx context.Context, data []DataPoint) ([]*Pattern, error) {
	var patterns []*Pattern
	
	// Group data points by type
	typeGroups := make(map[string][]DataPoint)
	for _, dp := range data {
		typeGroups[dp.Type] = append(typeGroups[dp.Type], dp)
	}
	
	// Analyze each group for patterns
	for dataType, points := range typeGroups {
		if len(points) < 3 { // Need at least 3 points for a pattern
			continue
		}
		
		// Look for common conditions
		conditionFreq := make(map[string]int)
		for _, point := range points {
			conditions := le.extractConditions(point)
			for _, cond := range conditions {
				key := fmt.Sprintf("%s_%s_%v", cond.Field, cond.Operator, cond.Value)
				conditionFreq[key]++
			}
		}
		
		// Create patterns from frequent conditions
		for condKey, freq := range conditionFreq {
			if float64(freq)/float64(len(points)) > 0.5 { // Appears in >50% of points
				pattern := &Pattern{
					ID:          fmt.Sprintf("pattern_%s_%d", dataType, time.Now().Unix()),
					Type:        dataType,
					Description: fmt.Sprintf("Common pattern in %s data", dataType),
					Conditions:  le.parseConditionKey(condKey),
					Frequency:   freq,
					Confidence:  float64(freq) / float64(len(points)),
					Examples:    le.getExampleIDs(points, condKey, 3),
				}
				patterns = append(patterns, pattern)
			}
		}
		
		// Look for sequential patterns
		sequentialPatterns := le.findSequentialPatterns(points)
		patterns = append(patterns, sequentialPatterns...)
		
		// Look for temporal patterns
		temporalPatterns := le.findTemporalPatterns(points)
		patterns = append(patterns, temporalPatterns...)
	}
	
	return patterns, nil
}

// ValidatePattern validates a pattern against new data
func (le *learningEngine) ValidatePattern(ctx context.Context, pattern *Pattern) (bool, float64, error) {
	// Get recent data points of the same type
	recentData := le.getRecentDataPoints(pattern.Type, 100)
	
	// Count how many match the pattern
	matches := 0
	for _, dp := range recentData {
		if le.matchesPattern(dp, pattern) {
			matches++
		}
	}
	
	// Calculate validation confidence
	confidence := float64(matches) / float64(len(recentData))
	
	// Pattern is valid if confidence > 0.6
	isValid := confidence > 0.6
	
	// Update pattern confidence
	pattern.Confidence = (pattern.Confidence + confidence) / 2
	
	return isValid, confidence, nil
}

// ExtractRules extracts rules from outcomes
func (le *learningEngine) ExtractRules(ctx context.Context, outcomes []Outcome) ([]*Rule, error) {
	var rules []*Rule
	
	// Group outcomes by action
	actionGroups := make(map[string][]Outcome)
	for _, outcome := range outcomes {
		actionGroups[outcome.Action] = append(actionGroups[outcome.Action], outcome)
	}
	
	// Analyze each action group
	for action, actionOutcomes := range actionGroups {
		// Separate successful and failed outcomes
		var successful, failed []Outcome
		for _, outcome := range actionOutcomes {
			if outcome.Success {
				successful = append(successful, outcome)
			} else {
				failed = append(failed, outcome)
			}
		}
		
		// Extract conditions that lead to success
		if len(successful) > 0 {
			successConditions := le.findCommonConditions(successful)
			if len(successConditions) > 0 {
				rule := &Rule{
					ID:          fmt.Sprintf("rule_%s_success_%d", action, time.Now().Unix()),
					Name:        fmt.Sprintf("Success conditions for %s", action),
					Description: fmt.Sprintf("Conditions that typically lead to successful %s", action),
					Conditions:  successConditions,
					Actions:     []string{action},
					Priority:    1,
					Confidence:  float64(len(successful)) / float64(len(actionOutcomes)),
				}
				rules = append(rules, rule)
			}
		}
		
		// Extract conditions that lead to failure
		if len(failed) > 0 {
			failureConditions := le.findCommonConditions(failed)
			if len(failureConditions) > 0 {
				rule := &Rule{
					ID:          fmt.Sprintf("rule_%s_failure_%d", action, time.Now().Unix()),
					Name:        fmt.Sprintf("Failure conditions for %s", action),
					Description: fmt.Sprintf("Conditions to avoid for %s", action),
					Conditions:  failureConditions,
					Actions:     []string{fmt.Sprintf("avoid_%s", action)},
					Priority:    2,
					Confidence:  float64(len(failed)) / float64(len(actionOutcomes)),
				}
				rules = append(rules, rule)
			}
		}
		
		// Extract metric-based rules
		metricRules := le.extractMetricRules(actionOutcomes)
		rules = append(rules, metricRules...)
	}
	
	return rules, nil
}

// TestRule tests a rule against test data
func (le *learningEngine) TestRule(ctx context.Context, rule *Rule, testData []DataPoint) (*RuleValidation, error) {
	validation := &RuleValidation{
		RuleID:     rule.ID,
		TotalTests: len(testData),
	}
	
	for _, dp := range testData {
		// Check if data point matches rule conditions
		if le.matchesConditions(dp, rule.Conditions) {
			// This would trigger the rule
			outcome := le.simulateRuleApplication(rule, dp)
			
			if outcome.Success {
				validation.Successful++
			} else {
				validation.Failed++
				if outcome.ExpectedSuccess {
					validation.FalseNegatives++
				} else {
					validation.FalsePositives++
				}
			}
		}
	}
	
	// Calculate accuracy
	if validation.TotalTests > 0 {
		validation.Accuracy = float64(validation.Successful) / float64(validation.TotalTests)
	}
	
	return validation, nil
}

// ConnectConcepts builds a knowledge graph from learning artifacts
func (le *learningEngine) ConnectConcepts(ctx context.Context, artifacts []*entities.LearningArtifact) (*KnowledgeGraph, error) {
	graph := &KnowledgeGraph{
		Nodes:    []KnowledgeNode{},
		Edges:    []KnowledgeEdge{},
		Clusters: []Cluster{},
	}
	
	// Create nodes from artifacts
	for _, artifact := range artifacts {
		node := KnowledgeNode{
			ID:    artifact.ID,
			Type:  string(artifact.Type),
			Label: artifact.Title,
			Properties: map[string]interface{}{
				"category":    artifact.Category,
				"confidence":  artifact.Confidence,
				"impact":      artifact.ImpactScore,
				"usage_count": artifact.UsageCount,
			},
			Importance: artifact.ImpactScore * artifact.Confidence,
		}
		graph.Nodes = append(graph.Nodes, node)
	}
	
	// Create edges based on relationships
	for i, artifact1 := range artifacts {
		for j, artifact2 := range artifacts {
			if i >= j {
				continue
			}
			
			// Calculate similarity/relationship strength
			similarity := le.calculateArtifactSimilarity(artifact1, artifact2)
			
			if similarity > 0.3 { // Threshold for creating an edge
				edge := KnowledgeEdge{
					ID:     fmt.Sprintf("edge_%s_%s", artifact1.ID, artifact2.ID),
					Source: artifact1.ID,
					Target: artifact2.ID,
					Type:   le.determineRelationType(artifact1, artifact2),
					Weight: similarity,
					Properties: map[string]interface{}{
						"created": time.Now(),
					},
				}
				graph.Edges = append(graph.Edges, edge)
				
				// Update artifact relationships
				artifact1.RelatedArtifacts = append(artifact1.RelatedArtifacts, artifact2.ID)
				artifact2.RelatedArtifacts = append(artifact2.RelatedArtifacts, artifact1.ID)
			}
		}
	}
	
	// Identify clusters using community detection
	clusters := le.detectCommunities(graph)
	graph.Clusters = clusters
	
	// Store relationships in repository
	for _, edge := range graph.Edges {
		if err := le.learningRepo.LinkArtifacts(ctx, edge.Source, edge.Target, edge.Type); err != nil {
			return nil, fmt.Errorf("linking artifacts: %w", err)
		}
	}
	
	return graph, nil
}

// IdentifyContradictions finds contradicting learning artifacts
func (le *learningEngine) IdentifyContradictions(ctx context.Context) ([]*Contradiction, error) {
	var contradictions []*Contradiction
	
	// Get all active artifacts
	artifacts, err := le.learningRepo.GetActiveArtifacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting active artifacts: %w", err)
	}
	
	// Group by category for efficient comparison
	categoryGroups := make(map[string][]*entities.LearningArtifact)
	for _, artifact := range artifacts {
		categoryGroups[artifact.Category] = append(categoryGroups[artifact.Category], artifact)
	}
	
	// Check for contradictions within each category
	for _, catArtifacts := range categoryGroups {
		for i, artifact1 := range catArtifacts {
			for j, artifact2 := range catArtifacts {
				if i >= j {
					continue
				}
				
				// Check if artifacts contradict each other
				if le.areContradictory(artifact1, artifact2) {
					contradiction := &Contradiction{
						ID:          fmt.Sprintf("contradiction_%d", time.Now().UnixNano()),
						Type:        le.classifyContradiction(artifact1, artifact2),
						Artifact1:   artifact1.ID,
						Artifact2:   artifact2.ID,
						Description: le.describeContradiction(artifact1, artifact2),
						Severity:    le.assessContradictionSeverity(artifact1, artifact2),
						Evidence:    le.gatherContradictionEvidence(artifact1, artifact2),
					}
					contradictions = append(contradictions, contradiction)
				}
			}
		}
	}
	
	return contradictions, nil
}

// ResolveContradiction attempts to resolve a contradiction
func (le *learningEngine) ResolveContradiction(ctx context.Context, contradiction *Contradiction) (*entities.LearningArtifact, error) {
	// Get the contradicting artifacts
	artifact1, err := le.learningRepo.GetArtifact(ctx, contradiction.Artifact1)
	if err != nil {
		return nil, fmt.Errorf("getting artifact1: %w", err)
	}
	
	artifact2, err := le.learningRepo.GetArtifact(ctx, contradiction.Artifact2)
	if err != nil {
		return nil, fmt.Errorf("getting artifact2: %w", err)
	}
	
	// Determine resolution strategy
	var resolution *entities.LearningArtifact
	
	switch contradiction.Type {
	case "direct_conflict":
		// Choose the artifact with higher confidence and more recent evidence
		resolution = le.resolveByConfidenceAndRecency(artifact1, artifact2)
		
	case "scope_overlap":
		// Create a new artifact that defines the scope boundaries
		resolution = le.createScopeBoundaryArtifact(artifact1, artifact2)
		
	case "temporal_change":
		// Create a new artifact that captures the evolution
		resolution = le.createEvolutionArtifact(artifact1, artifact2)
		
	default:
		// Create a synthesis artifact
		resolution = le.synthesizeArtifacts(artifact1, artifact2, contradiction)
	}
	
	// Store the resolution
	if resolution != nil {
		resolution.ID = fmt.Sprintf("resolution_%d", time.Now().Unix())
		resolution.Source = entities.SourceSystemMonitoring
		resolution.RelatedArtifacts = []string{artifact1.ID, artifact2.ID}
		resolution.Status = entities.ArtifactStatusActive
		resolution.CreatedAt = time.Now()
		resolution.UpdatedAt = time.Now()
		
		if err := le.learningRepo.CreateArtifact(ctx, resolution); err != nil {
			return nil, fmt.Errorf("creating resolution artifact: %w", err)
		}
		
		// Update original artifacts
		artifact1.Status = entities.ArtifactStatusDeprecated
		artifact2.Status = entities.ArtifactStatusDeprecated
		
		if err := le.learningRepo.UpdateArtifact(ctx, artifact1); err != nil {
			return nil, fmt.Errorf("updating artifact1: %w", err)
		}
		
		if err := le.learningRepo.UpdateArtifact(ctx, artifact2); err != nil {
			return nil, fmt.Errorf("updating artifact2: %w", err)
		}
	}
	
	return resolution, nil
}

// Helper methods

func (le *learningEngine) extractConditions(dp DataPoint) []Condition {
	var conditions []Condition
	
	// Extract conditions from context
	for field, value := range dp.Context {
		condition := Condition{
			Field:    field,
			Operator: "equals",
			Value:    value,
		}
		conditions = append(conditions, condition)
	}
	
	// Extract conditions from value if it's a map
	if valueMap, ok := dp.Value.(map[string]interface{}); ok {
		for field, value := range valueMap {
			condition := Condition{
				Field:    fmt.Sprintf("value.%s", field),
				Operator: "equals",
				Value:    value,
			}
			conditions = append(conditions, condition)
		}
	}
	
	return conditions
}

func (le *learningEngine) parseConditionKey(key string) []Condition {
	parts := strings.Split(key, "_")
	if len(parts) >= 3 {
		return []Condition{{
			Field:    parts[0],
			Operator: parts[1],
			Value:    strings.Join(parts[2:], "_"),
		}}
	}
	return []Condition{}
}

func (le *learningEngine) getExampleIDs(points []DataPoint, condKey string, limit int) []string {
	var examples []string
	count := 0
	
	for _, point := range points {
		conditions := le.extractConditions(point)
		for _, cond := range conditions {
			key := fmt.Sprintf("%s_%s_%v", cond.Field, cond.Operator, cond.Value)
			if key == condKey {
				examples = append(examples, point.ID)
				count++
				if count >= limit {
					return examples
				}
				break
			}
		}
	}
	
	return examples
}

func (le *learningEngine) findSequentialPatterns(points []DataPoint) []*Pattern {
	var patterns []*Pattern
	
	// Sort by timestamp
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})
	
	// Look for sequences of events
	for i := 0; i < len(points)-2; i++ {
		// Check if there's a pattern in the next 3 points
		if le.isSequentialPattern(points[i:i+3]) {
			pattern := &Pattern{
				ID:          fmt.Sprintf("seq_pattern_%d", time.Now().Unix()),
				Type:        "sequential",
				Description: "Sequential pattern detected",
				Frequency:   1,
				Confidence:  0.7,
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

func (le *learningEngine) findTemporalPatterns(points []DataPoint) []*Pattern {
	var patterns []*Pattern
	
	// Group by hour of day
	hourGroups := make(map[int][]DataPoint)
	for _, point := range points {
		hour := point.Timestamp.Hour()
		hourGroups[hour] = append(hourGroups[hour], point)
	}
	
	// Find hours with significantly more activity
	avgPerHour := float64(len(points)) / 24.0
	for hour, hourPoints := range hourGroups {
		if float64(len(hourPoints)) > avgPerHour*1.5 {
			pattern := &Pattern{
				ID:          fmt.Sprintf("temporal_pattern_%d", time.Now().Unix()),
				Type:        "temporal",
				Description: fmt.Sprintf("High activity during hour %d", hour),
				Conditions: []Condition{{
					Field:    "timestamp.hour",
					Operator: "equals",
					Value:    hour,
				}},
				Frequency:  len(hourPoints),
				Confidence: float64(len(hourPoints)) / float64(len(points)),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

func (le *learningEngine) getRecentDataPoints(dataType string, limit int) []DataPoint {
	// This would fetch from a data store
	// For now, returning empty slice
	return []DataPoint{}
}

func (le *learningEngine) matchesPattern(dp DataPoint, pattern *Pattern) bool {
	// Check if all pattern conditions are met
	for _, condition := range pattern.Conditions {
		if !le.checkCondition(dp, condition) {
			return false
		}
	}
	return true
}

func (le *learningEngine) checkCondition(dp DataPoint, condition Condition) bool {
	// Get field value from data point
	value := le.getFieldValue(dp, condition.Field)
	if value == nil {
		return false
	}
	
	// Check condition based on operator
	switch condition.Operator {
	case "equals":
		return value == condition.Value
	case "greater":
		if vFloat, ok := value.(float64); ok {
			if cFloat, ok := condition.Value.(float64); ok {
				return vFloat > cFloat
			}
		}
	case "less":
		if vFloat, ok := value.(float64); ok {
			if cFloat, ok := condition.Value.(float64); ok {
				return vFloat < cFloat
			}
		}
	case "contains":
		if vStr, ok := value.(string); ok {
			if cStr, ok := condition.Value.(string); ok {
				return strings.Contains(vStr, cStr)
			}
		}
	}
	
	return false
}

func (le *learningEngine) getFieldValue(dp DataPoint, field string) interface{} {
	// Handle nested fields
	parts := strings.Split(field, ".")
	
	if parts[0] == "value" && len(parts) > 1 {
		if valueMap, ok := dp.Value.(map[string]interface{}); ok {
			return valueMap[parts[1]]
		}
	} else if parts[0] == "timestamp" && len(parts) > 1 {
		switch parts[1] {
		case "hour":
			return dp.Timestamp.Hour()
		case "day":
			return dp.Timestamp.Day()
		case "weekday":
			return int(dp.Timestamp.Weekday())
		}
	} else if val, ok := dp.Context[field]; ok {
		return val
	}
	
	return nil
}

func (le *learningEngine) findCommonConditions(outcomes []Outcome) []Condition {
	var conditions []Condition
	
	// Count frequency of each context key-value pair
	contextFreq := make(map[string]int)
	for _, outcome := range outcomes {
		for key, value := range outcome.Context {
			contextKey := fmt.Sprintf("%s=%v", key, value)
			contextFreq[contextKey]++
		}
	}
	
	// Extract conditions that appear in >60% of outcomes
	threshold := int(float64(len(outcomes)) * 0.6)
	for contextKey, freq := range contextFreq {
		if freq >= threshold {
			parts := strings.SplitN(contextKey, "=", 2)
			if len(parts) == 2 {
				condition := Condition{
					Field:    parts[0],
					Operator: "equals",
					Value:    parts[1],
				}
				conditions = append(conditions, condition)
			}
		}
	}
	
	return conditions
}

func (le *learningEngine) extractMetricRules(outcomes []Outcome) []*Rule {
	var rules []*Rule
	
	// Group by success/failure
	var successful, failed []Outcome
	for _, outcome := range outcomes {
		if outcome.Success {
			successful = append(successful, outcome)
		} else {
			failed = append(failed, outcome)
		}
	}
	
	// Analyze metric thresholds
	for metricName := range outcomes[0].Metrics {
		// Find threshold that best separates success/failure
		threshold, accuracy := le.findOptimalThreshold(metricName, successful, failed)
		
		if accuracy > 0.7 { // Good separation
			rule := &Rule{
				ID:          fmt.Sprintf("metric_rule_%s_%d", metricName, time.Now().Unix()),
				Name:        fmt.Sprintf("Threshold rule for %s", metricName),
				Description: fmt.Sprintf("%s should be greater than %.2f", metricName, threshold),
				Conditions: []Condition{{
					Field:    metricName,
					Operator: "greater",
					Value:    threshold,
				}},
				Priority:   3,
				Confidence: accuracy,
			}
			rules = append(rules, rule)
		}
	}
	
	return rules
}

func (le *learningEngine) findOptimalThreshold(metric string, successful, failed []Outcome) (float64, float64) {
	// Collect all metric values
	var values []float64
	for _, outcome := range successful {
		if val, ok := outcome.Metrics[metric]; ok {
			values = append(values, val)
		}
	}
	for _, outcome := range failed {
		if val, ok := outcome.Metrics[metric]; ok {
			values = append(values, val)
		}
	}
	
	if len(values) == 0 {
		return 0, 0
	}
	
	// Sort values
	sort.Float64s(values)
	
	// Try each value as a threshold
	bestThreshold := values[0]
	bestAccuracy := 0.0
	
	for _, threshold := range values {
		correctCount := 0
		
		// Count correct classifications
		for _, outcome := range successful {
			if val, ok := outcome.Metrics[metric]; ok && val > threshold {
				correctCount++
			}
		}
		for _, outcome := range failed {
			if val, ok := outcome.Metrics[metric]; ok && val <= threshold {
				correctCount++
			}
		}
		
		accuracy := float64(correctCount) / float64(len(successful)+len(failed))
		if accuracy > bestAccuracy {
			bestAccuracy = accuracy
			bestThreshold = threshold
		}
	}
	
	return bestThreshold, bestAccuracy
}

func (le *learningEngine) matchesConditions(dp DataPoint, conditions []Condition) bool {
	for _, condition := range conditions {
		if !le.checkCondition(dp, condition) {
			return false
		}
	}
	return true
}

func (le *learningEngine) simulateRuleApplication(rule *Rule, dp DataPoint) struct {
	Success         bool
	ExpectedSuccess bool
} {
	// Simulate applying the rule
	// In real implementation, this would check against actual outcomes
	return struct {
		Success         bool
		ExpectedSuccess bool
	}{
		Success:         true,
		ExpectedSuccess: true,
	}
}

func (le *learningEngine) calculateArtifactSimilarity(a1, a2 *entities.LearningArtifact) float64 {
	similarity := 0.0
	
	// Category similarity
	if a1.Category == a2.Category {
		similarity += 0.3
	}
	
	// Type similarity
	if a1.Type == a2.Type {
		similarity += 0.2
	}
	
	// Tag overlap
	tagOverlap := le.calculateTagOverlap(a1.Tags, a2.Tags)
	similarity += tagOverlap * 0.2
	
	// Description similarity (simplified)
	descSimilarity := le.calculateTextSimilarity(a1.Description, a2.Description)
	similarity += descSimilarity * 0.3
	
	return similarity
}

func (le *learningEngine) calculateTagOverlap(tags1, tags2 []string) float64 {
	if len(tags1) == 0 || len(tags2) == 0 {
		return 0
	}
	
	tagSet := make(map[string]bool)
	for _, tag := range tags1 {
		tagSet[tag] = true
	}
	
	overlap := 0
	for _, tag := range tags2 {
		if tagSet[tag] {
			overlap++
		}
	}
	
	return float64(overlap) / float64(math.Max(float64(len(tags1)), float64(len(tags2))))
}

func (le *learningEngine) calculateTextSimilarity(text1, text2 string) float64 {
	// Simple word overlap calculation
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}
	
	wordSet := make(map[string]bool)
	for _, word := range words1 {
		wordSet[word] = true
	}
	
	overlap := 0
	for _, word := range words2 {
		if wordSet[word] {
			overlap++
		}
	}
	
	return float64(overlap) / float64(math.Max(float64(len(words1)), float64(len(words2))))
}

func (le *learningEngine) determineRelationType(a1, a2 *entities.LearningArtifact) string {
	if a1.Type == a2.Type && a1.Category == a2.Category {
		return "similar"
	}
	
	// Check if one is prerequisite of another
	for _, prereq := range a2.Prerequisites {
		if prereq == a1.ID {
			return "prerequisite"
		}
	}
	
	// Check if they're from the same source
	if a1.Source == a2.Source && a1.SourceID == a2.SourceID {
		return "same_source"
	}
	
	return "related"
}

func (le *learningEngine) detectCommunities(graph *KnowledgeGraph) []Cluster {
	// Simple community detection using connected components
	visited := make(map[string]bool)
	var clusters []Cluster
	clusterID := 0
	
	// Build adjacency list
	adjacency := make(map[string][]string)
	for _, edge := range graph.Edges {
		adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
		adjacency[edge.Target] = append(adjacency[edge.Target], edge.Source)
	}
	
	// Find connected components
	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			cluster := Cluster{
				ID:      fmt.Sprintf("cluster_%d", clusterID),
				NodeIDs: []string{},
			}
			
			// DFS to find all connected nodes
			stack := []string{node.ID}
			for len(stack) > 0 {
				current := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				
				if visited[current] {
					continue
				}
				
				visited[current] = true
				cluster.NodeIDs = append(cluster.NodeIDs, current)
				
				// Add neighbors
				for _, neighbor := range adjacency[current] {
					if !visited[neighbor] {
						stack = append(stack, neighbor)
					}
				}
			}
			
			// Calculate cluster coherence
			cluster.Coherence = le.calculateClusterCoherence(cluster, graph)
			cluster.Name = le.generateClusterName(cluster, graph)
			cluster.Description = le.generateClusterDescription(cluster, graph)
			
			clusters = append(clusters, cluster)
			clusterID++
		}
	}
	
	return clusters
}

func (le *learningEngine) calculateClusterCoherence(cluster Cluster, graph *KnowledgeGraph) float64 {
	if len(cluster.NodeIDs) <= 1 {
		return 1.0
	}
	
	// Calculate internal edge density
	internalEdges := 0
	possibleEdges := len(cluster.NodeIDs) * (len(cluster.NodeIDs) - 1) / 2
	
	nodeSet := make(map[string]bool)
	for _, nodeID := range cluster.NodeIDs {
		nodeSet[nodeID] = true
	}
	
	for _, edge := range graph.Edges {
		if nodeSet[edge.Source] && nodeSet[edge.Target] {
			internalEdges++
		}
	}
	
	if possibleEdges == 0 {
		return 0
	}
	
	return float64(internalEdges) / float64(possibleEdges)
}

func (le *learningEngine) generateClusterName(cluster Cluster, graph *KnowledgeGraph) string {
	// Find most common category in cluster
	categoryCount := make(map[string]int)
	
	for _, nodeID := range cluster.NodeIDs {
		for _, node := range graph.Nodes {
			if node.ID == nodeID {
				if category, ok := node.Properties["category"].(string); ok {
					categoryCount[category]++
				}
				break
			}
		}
	}
	
	// Find most frequent category
	maxCount := 0
	dominantCategory := "General"
	for category, count := range categoryCount {
		if count > maxCount {
			maxCount = count
			dominantCategory = category
		}
	}
	
	// Simple capitalization without using deprecated strings.Title
	if len(dominantCategory) > 0 {
		return fmt.Sprintf("%s%s Knowledge Cluster", 
			strings.ToUpper(dominantCategory[:1]), 
			dominantCategory[1:])
	}
	return "Knowledge Cluster"
}

func (le *learningEngine) generateClusterDescription(cluster Cluster, graph *KnowledgeGraph) string {
	return fmt.Sprintf("Cluster containing %d related knowledge artifacts with coherence score of %.2f",
		len(cluster.NodeIDs), cluster.Coherence)
}

func (le *learningEngine) areContradictory(a1, a2 *entities.LearningArtifact) bool {
	// Check for direct contradictions in rules/patterns
	if a1.Type == entities.LearningTypeRule && a2.Type == entities.LearningTypeRule {
		// Rules with opposite outcomes for same conditions
		return le.checkRuleContradiction(a1, a2)
	}
	
	// Check for conflicting constraints
	if a1.Type == entities.LearningTypeConstraint && a2.Type == entities.LearningTypeConstraint {
		return le.checkConstraintConflict(a1, a2)
	}
	
	// Check for temporal contradictions
	if a1.Category == a2.Category && a1.SourceID == a2.SourceID {
		return le.checkTemporalContradiction(a1, a2)
	}
	
	return false
}

func (le *learningEngine) checkRuleContradiction(a1, a2 *entities.LearningArtifact) bool {
	// This would parse rule metadata and check for conflicts
	// Simplified implementation
	return strings.Contains(a1.Description, "should") && 
	       strings.Contains(a2.Description, "should not") &&
	       le.calculateTextSimilarity(a1.Description, a2.Description) > 0.5
}

func (le *learningEngine) checkConstraintConflict(a1, a2 *entities.LearningArtifact) bool {
	// Check if constraints are mutually exclusive
	return a1.Category == a2.Category && 
	       math.Abs(a1.Confidence-a2.Confidence) < 0.1 &&
	       le.calculateTextSimilarity(a1.Description, a2.Description) > 0.6
}

func (le *learningEngine) checkTemporalContradiction(a1, a2 *entities.LearningArtifact) bool {
	// Check if newer artifact contradicts older one
	timeDiff := a2.CreatedAt.Sub(a1.CreatedAt)
	if timeDiff > 24*time.Hour {
		// Significant time difference suggests evolution rather than contradiction
		return false
	}
	
	return le.calculateTextSimilarity(a1.Description, a2.Description) > 0.7 &&
	       a1.Confidence > 0.6 && a2.Confidence > 0.6
}

func (le *learningEngine) classifyContradiction(a1, a2 *entities.LearningArtifact) string {
	if le.checkRuleContradiction(a1, a2) {
		return "direct_conflict"
	}
	
	if a1.Category != a2.Category {
		return "scope_overlap"
	}
	
	timeDiff := a2.CreatedAt.Sub(a1.CreatedAt)
	if timeDiff > 7*24*time.Hour {
		return "temporal_change"
	}
	
	return "general_inconsistency"
}

func (le *learningEngine) describeContradiction(a1, a2 *entities.LearningArtifact) string {
	return fmt.Sprintf("Artifact '%s' contradicts '%s': %s vs %s",
		a1.Title, a2.Title, a1.Description, a2.Description)
}

func (le *learningEngine) assessContradictionSeverity(a1, a2 *entities.LearningArtifact) string {
	// High severity if both have high confidence and impact
	if a1.Confidence > 0.8 && a2.Confidence > 0.8 &&
	   a1.ImpactScore > 0.7 && a2.ImpactScore > 0.7 {
		return "high"
	}
	
	// Medium severity if one has high confidence/impact
	if (a1.Confidence > 0.8 || a2.Confidence > 0.8) &&
	   (a1.ImpactScore > 0.6 || a2.ImpactScore > 0.6) {
		return "medium"
	}
	
	return "low"
}

func (le *learningEngine) gatherContradictionEvidence(a1, a2 *entities.LearningArtifact) []string {
	var evidence []string
	
	evidence = append(evidence, fmt.Sprintf("Artifact 1 created: %s", a1.CreatedAt.Format("2006-01-02")))
	evidence = append(evidence, fmt.Sprintf("Artifact 2 created: %s", a2.CreatedAt.Format("2006-01-02")))
	evidence = append(evidence, fmt.Sprintf("Confidence levels: %.2f vs %.2f", a1.Confidence, a2.Confidence))
	evidence = append(evidence, fmt.Sprintf("Impact scores: %.2f vs %.2f", a1.ImpactScore, a2.ImpactScore))
	
	return evidence
}

func (le *learningEngine) resolveByConfidenceAndRecency(a1, a2 *entities.LearningArtifact) *entities.LearningArtifact {
	// Weight confidence and recency
	score1 := a1.Confidence * 0.6 + (1.0 / (time.Since(a1.CreatedAt).Hours() + 1)) * 0.4
	score2 := a2.Confidence * 0.6 + (1.0 / (time.Since(a2.CreatedAt).Hours() + 1)) * 0.4
	
	if score1 > score2 {
		return a1
	}
	return a2
}

func (le *learningEngine) createScopeBoundaryArtifact(a1, a2 *entities.LearningArtifact) *entities.LearningArtifact {
	return &entities.LearningArtifact{
		Type:        entities.LearningTypeRule,
		Category:    "scope_definition",
		Title:       fmt.Sprintf("Scope boundary between %s and %s", a1.Category, a2.Category),
		Description: fmt.Sprintf("Defines when to apply '%s' vs '%s'", a1.Title, a2.Title),
		Confidence:  (a1.Confidence + a2.Confidence) / 2,
		ImpactScore: math.Max(a1.ImpactScore, a2.ImpactScore),
	}
}

func (le *learningEngine) createEvolutionArtifact(a1, a2 *entities.LearningArtifact) *entities.LearningArtifact {
	return &entities.LearningArtifact{
		Type:        entities.LearningTypePattern,
		Category:    "evolution",
		Title:       fmt.Sprintf("Evolution of %s understanding", a1.Category),
		Description: fmt.Sprintf("Knowledge evolved from '%s' to '%s'", a1.Description, a2.Description),
		Confidence:  a2.Confidence, // Use newer confidence
		ImpactScore: a2.ImpactScore,
	}
}

func (le *learningEngine) synthesizeArtifacts(a1, a2 *entities.LearningArtifact, contradiction *Contradiction) *entities.LearningArtifact {
	return &entities.LearningArtifact{
		Type:        entities.LearningTypeRelationship,
		Category:    "synthesis",
		Title:       "Synthesized understanding",
		Description: fmt.Sprintf("Synthesis resolving contradiction: %s", contradiction.Description),
		Confidence:  math.Min(a1.Confidence, a2.Confidence) * 0.9, // Slightly reduced confidence
		ImpactScore: (a1.ImpactScore + a2.ImpactScore) / 2,
	}
}

func (le *learningEngine) isSequentialPattern(points []DataPoint) bool {
	// Check if points follow a sequential pattern
	// Simplified implementation
	return len(points) >= 3
}
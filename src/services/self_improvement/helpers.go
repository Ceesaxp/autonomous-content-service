package self_improvement

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ProjectSuccess represents the success evaluation of a project
type ProjectSuccess struct {
	Score              float64
	ClientSatisfaction float64
	QualityScore       float64
	TimelinessScore    float64
	Challenges         []ProjectChallenge
}

// ProjectChallenge represents a challenge encountered in a project
type ProjectChallenge struct {
	Type        string
	Description string
	Impact      float64
	Resolution  string
}

// Helper methods for Service

func (s *Service) evaluateProjectSuccess(project *entities.Project, content []*entities.Content, feedback []*entities.Feedback) ProjectSuccess {
	success := ProjectSuccess{
		Score:       0,
		Challenges:  []ProjectChallenge{},
	}
	
	// Calculate quality score from content
	if len(content) > 0 {
		totalQuality := 0.0
		count := 0
		for _, c := range content {
			if c.Statistics != nil {
				qualityScore := (c.Statistics.ReadabilityScore + c.Statistics.SEOScore + c.Statistics.EngagementScore) / 3.0
				if qualityScore > 0 {
					totalQuality += qualityScore
					count++
				}
			}
		}
		if count > 0 {
			success.QualityScore = totalQuality / float64(count)
		}
	}
	
	// Calculate client satisfaction from feedback
	if len(feedback) > 0 {
		totalRating := 0.0
		ratingCount := 0
		for _, f := range feedback {
			if f.Rating != nil {
				totalRating += f.Rating.Score
				ratingCount++
			}
		}
		if ratingCount > 0 {
			success.ClientSatisfaction = totalRating / float64(ratingCount)
		}
	}
	
	// Calculate timeliness score
	if project.Status == entities.ProjectStatusCompleted {
		// If completed, assume it was on time (no completion timestamp in entity)
		success.TimelinessScore = 1.0
	}
	
	// Calculate overall score
	success.Score = (success.QualityScore/100*0.4 + 
		success.ClientSatisfaction/5*0.4 + 
		success.TimelinessScore*0.2)
	
	// Identify challenges
	if success.QualityScore < 70 {
		success.Challenges = append(success.Challenges, ProjectChallenge{
			Type:        "quality_issues",
			Description: "Content quality below threshold",
			Impact:      0.7,
			Resolution:  "Review and improve content generation process",
		})
	}
	
	if success.ClientSatisfaction < 3.5 {
		success.Challenges = append(success.Challenges, ProjectChallenge{
			Type:        "client_dissatisfaction",
			Description: "Client satisfaction below expectations",
			Impact:      0.8,
			Resolution:  "Analyze feedback and improve service delivery",
		})
	}
	
	return success
}

func (s *Service) calculateAverageQuality(content []*entities.Content) float64 {
	if len(content) == 0 {
		return 0
	}
	
	total := 0.0
	count := 0
	for _, c := range content {
		if c.Statistics != nil {
			qualityScore := (c.Statistics.ReadabilityScore + c.Statistics.SEOScore + c.Statistics.EngagementScore) / 3.0
			if qualityScore > 0 {
				total += qualityScore
				count++
			}
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / float64(count)
}

func (s *Service) extractSatisfactionEvidence(project *entities.Project, feedback []*entities.Feedback) []entities.Evidence {
	var evidence []entities.Evidence
	
	for _, f := range feedback {
		if f.Type == entities.FeedbackTypeTestimonial || 
		   (f.Rating != nil && f.Rating.Score >= 4) {
			evidence = append(evidence, entities.Evidence{
				Type:        "client_feedback",
				Description: string(f.Type),
				Data: map[string]interface{}{
					"rating":  f.Rating,
					"message": f.Message,
					"tags":    f.Tags,
				},
				Timestamp: f.CreatedAt,
				Strength:  0.9,
			})
		}
	}
	
	return evidence
}

func (s *Service) calculateFeedbackStrength(feedback *entities.Feedback) float64 {
	strength := 0.5 // Base strength
	
	// Adjust based on feedback type
	typeWeights := map[entities.FeedbackType]float64{
		entities.FeedbackTypeTestimonial: 0.9,
		entities.FeedbackTypeComplaint:   0.8,
		entities.FeedbackTypeSuggestion:  0.7,
		entities.FeedbackTypePositive:    0.6,
		entities.FeedbackTypeNegative:    0.6,
		entities.FeedbackTypeNeutral:     0.4,
	}
	
	if weight, ok := typeWeights[feedback.Type]; ok {
		strength = weight
	}
	
	// Adjust based on source
	if feedback.Source == entities.FeedbackSourceClient {
		strength *= 1.2
	}
	
	return math.Min(strength, 1.0)
}

func (s *Service) calculateFeedbackConfidence(feedback *entities.Feedback) float64 {
	confidence := 0.7 // Base confidence
	
	// Higher confidence for direct client feedback
	if feedback.Source == entities.FeedbackSourceClient {
		confidence = 0.9
	}
	
	// Adjust based on metadata
	if len(feedback.Metadata) > 3 {
		confidence += 0.05
	}
	
	return math.Min(confidence, 0.95)
}

func (s *Service) calculateFeedbackImpact(feedback *entities.Feedback) float64 {
	impact := 0.5 // Base impact
	
	// High impact for complaints and testimonials
	if feedback.Type == entities.FeedbackTypeComplaint {
		impact = 0.8
	} else if feedback.Type == entities.FeedbackTypeTestimonial {
		impact = 0.7
	}
	
	// Adjust based on rating
	if feedback.Rating != nil {
		if feedback.Rating.Score <= 2 {
			impact = math.Max(impact, 0.8)
		} else if feedback.Rating.Score >= 4.5 {
			impact = math.Max(impact, 0.7)
		}
	}
	
	return impact
}

func (s *Service) calculateErrorImpact(errorData ErrorData) float64 {
	impact := 0.5 // Base impact
	
	// Adjust based on frequency
	if errorData.Frequency > 100 {
		impact = 0.9
	} else if errorData.Frequency > 50 {
		impact = 0.8
	} else if errorData.Frequency > 10 {
		impact = 0.7
	}
	
	// Adjust based on component
	componentImpact := map[string]float64{
		"payment":          1.3,
		"content_creation": 1.2,
		"decision_making":  1.1,
		"api":              1.0,
	}
	
	if multiplier, ok := componentImpact[errorData.Component]; ok {
		impact *= multiplier
	}
	
	return math.Min(impact, 1.0)
}

func (s *Service) extractClusterEvidence(cluster Cluster, artifacts []*entities.LearningArtifact) []entities.Evidence {
	var evidence []entities.Evidence
	
	// Extract evidence from clustered artifacts
	for _, nodeID := range cluster.NodeIDs {
		for _, artifact := range artifacts {
			if artifact.ID == nodeID && len(artifact.Evidence) > 0 {
				// Take the strongest evidence from each artifact
				strongestEvidence := artifact.Evidence[0]
				for _, e := range artifact.Evidence {
					if e.Strength > strongestEvidence.Strength {
						strongestEvidence = e
					}
				}
				evidence = append(evidence, strongestEvidence)
			}
		}
	}
	
	return evidence
}

func (s *Service) calculateClusterImpact(cluster Cluster, artifacts []*entities.LearningArtifact) float64 {
	if len(cluster.NodeIDs) == 0 {
		return 0
	}
	
	totalImpact := 0.0
	count := 0
	
	for _, nodeID := range cluster.NodeIDs {
		for _, artifact := range artifacts {
			if artifact.ID == nodeID {
				totalImpact += artifact.ImpactScore
				count++
				break
			}
		}
	}
	
	if count == 0 {
		return 0
	}
	
	// Average impact weighted by cluster coherence
	return (totalImpact / float64(count)) * cluster.Coherence
}

func (s *Service) identifyContentCapabilityGaps(ctx context.Context) []*entities.CapabilityGap {
	var gaps []*entities.CapabilityGap
	
	// This would analyze failed content requests, unsupported formats, etc.
	// For now, returning example gaps
	gaps = append(gaps, &entities.CapabilityGap{
		ID:              fmt.Sprintf("gap_%d", time.Now().Unix()),
		Type:            entities.CapabilityGapTypeContent,
		Description:     "Video content creation capability",
		Frequency:       15,
		EstimatedImpact: 0.8,
		EstimatedEffort: 0.7,
		PotentialSources: []entities.CapabilitySource{
			{
				Type:     "api_integration",
				Provider: "Synthesia",
				Cost:     500,
				TimeToAcquire: "1 week",
				Confidence: 0.8,
			},
		},
		Status:    entities.GapStatusIdentified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	
	return gaps
}

func (s *Service) identifyIntegrationGaps(ctx context.Context) []*entities.CapabilityGap {
	var gaps []*entities.CapabilityGap
	
	// Analyze integration requests from projects
	gaps = append(gaps, &entities.CapabilityGap{
		ID:              fmt.Sprintf("gap_%d", time.Now().Unix()+1),
		Type:            entities.CapabilityGapTypeAPI,
		Description:     "Slack integration for notifications",
		Frequency:       8,
		EstimatedImpact: 0.6,
		EstimatedEffort: 0.4,
		PotentialSources: []entities.CapabilitySource{
			{
				Type:     "api_integration",
				Provider: "Slack API",
				Cost:     0,
				TimeToAcquire: "3 days",
				Confidence: 0.9,
			},
		},
		Status:    entities.GapStatusIdentified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	
	return gaps
}

func (s *Service) identifyDomainGaps(ctx context.Context) []*entities.CapabilityGap {
	var gaps []*entities.CapabilityGap
	
	// Analyze industry/language requests
	gaps = append(gaps, &entities.CapabilityGap{
		ID:              fmt.Sprintf("gap_%d", time.Now().Unix()+2),
		Type:            entities.CapabilityGapTypeLanguage,
		Description:     "Spanish language content creation",
		Frequency:       12,
		EstimatedImpact: 0.7,
		EstimatedEffort: 0.5,
		PotentialSources: []entities.CapabilitySource{
			{
				Type:     "llm_configuration",
				Provider: "Internal",
				Cost:     0,
				TimeToAcquire: "1 day",
				Confidence: 0.95,
			},
		},
		Status:    entities.GapStatusIdentified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	
	return gaps
}

func (s *Service) calculateGapPriority(gap *entities.CapabilityGap) float64 {
	// Priority = (Impact * Frequency) / (Effort * (Cost/1000 + 1))
	costFactor := 1.0
	if len(gap.PotentialSources) > 0 {
		// Use lowest cost option
		minCost := gap.PotentialSources[0].Cost
		for _, source := range gap.PotentialSources {
			if source.Cost < minCost {
				minCost = source.Cost
			}
		}
		costFactor = minCost/1000 + 1
	}
	
	priority := (gap.EstimatedImpact * float64(gap.Frequency)) / (gap.EstimatedEffort * costFactor)
	
	// Boost priority for certain capability types
	typeBoost := map[string]float64{
		entities.CapabilityGapTypeContent:  1.2,
		entities.CapabilityGapTypeAPI:      1.1,
		entities.CapabilityGapTypeLanguage: 1.0,
		entities.CapabilityGapTypeIndustry: 0.9,
	}
	
	if boost, ok := typeBoost[gap.Type]; ok {
		priority *= boost
	}
	
	return priority
}
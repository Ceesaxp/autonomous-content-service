package content_creation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// ResearchSource represents a source of information for research
type ResearchSource struct {
	Type        string `json:"type"`        // "web", "database", "knowledge_base"
	URL         string `json:"url,omitempty"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Credibility float64 `json:"credibility"` // 0-1 score
	Relevance   float64 `json:"relevance"`   // 0-1 score
	LastUpdated time.Time `json:"lastUpdated"`
}

// ResearchTopic represents a topic to research
type ResearchTopic struct {
	Topic       string   `json:"topic"`
	Keywords    []string `json:"keywords"`
	Priority    int      `json:"priority"` // 1-10
	MaxSources  int      `json:"maxSources"`
}

// ResearchOutput contains the results of research
type ResearchOutput struct {
	Topics    []ResearchTopic  `json:"topics"`
	Sources   []ResearchSource `json:"sources"`
	KeyFacts  []string         `json:"keyFacts"`
	References []string        `json:"references"`
	Summary   string           `json:"summary"`
}

// Researcher defines the interface for content research
type Researcher interface {
	// Research conducts research for content creation
	Research(ctx context.Context, content *entities.Content, requirements ResearchRequirements) (*ResearchOutput, error)
	
	// EvaluateSourceCredibility assesses the credibility of a source
	EvaluateSourceCredibility(ctx context.Context, source ResearchSource) (float64, error)
	
	// ExtractKeyFacts extracts key facts from research sources
	ExtractKeyFacts(ctx context.Context, sources []ResearchSource) ([]string, error)
}

// ResearchRequirements specifies what kind of research is needed
type ResearchRequirements struct {
	Topics          []string `json:"topics"`
	TargetAudience  string   `json:"targetAudience"`
	Industry        string   `json:"industry"`
	ContentPurpose  string   `json:"contentPurpose"`
	MaxSources      int      `json:"maxSources"`
	RequireRecent   bool     `json:"requireRecent"`   // Prefer recent sources
	RequireCredible bool     `json:"requireCredible"` // Only use credible sources
}

// LLMResearcher implements Researcher using LLM and search services
type LLMResearcher struct {
	llmClient     LLMClient
	searchService SearchService
}

// NewLLMResearcher creates a new LLM-based researcher
func NewLLMResearcher(llmClient LLMClient, searchService SearchService) *LLMResearcher {
	return &LLMResearcher{
		llmClient:     llmClient,
		searchService: searchService,
	}
}

// Research conducts comprehensive research for content creation
func (r *LLMResearcher) Research(ctx context.Context, content *entities.Content, requirements ResearchRequirements) (*ResearchOutput, error) {
	output := &ResearchOutput{
		Topics:     []ResearchTopic{},
		Sources:    []ResearchSource{},
		KeyFacts:   []string{},
		References: []string{},
	}

	// Step 1: Generate research topics using LLM
	topics, err := r.generateResearchTopics(ctx, content, requirements)
	if err != nil {
		return nil, fmt.Errorf("failed to generate research topics: %w", err)
	}
	output.Topics = topics

	// Step 2: Gather sources for each topic
	for _, topic := range topics {
		sources, err := r.gatherSources(ctx, topic, requirements)
		if err != nil {
			// Log error but continue with other topics
			fmt.Printf("Warning: Failed to gather sources for topic '%s': %v\n", topic.Topic, err)
			continue
		}
		output.Sources = append(output.Sources, sources...)
	}

	// Step 3: Evaluate source credibility
	for i := range output.Sources {
		credibility, err := r.EvaluateSourceCredibility(ctx, output.Sources[i])
		if err != nil {
			fmt.Printf("Warning: Failed to evaluate credibility for source '%s': %v\n", output.Sources[i].Title, err)
			credibility = 0.5 // Default moderate credibility
		}
		output.Sources[i].Credibility = credibility
	}

	// Step 4: Filter sources based on requirements
	if requirements.RequireCredible {
		filteredSources := []ResearchSource{}
		for _, source := range output.Sources {
			if source.Credibility >= 0.7 { // Minimum credibility threshold
				filteredSources = append(filteredSources, source)
			}
		}
		output.Sources = filteredSources
	}

	// Step 5: Extract key facts
	keyFacts, err := r.ExtractKeyFacts(ctx, output.Sources)
	if err != nil {
		fmt.Printf("Warning: Failed to extract key facts: %v\n", err)
	} else {
		output.KeyFacts = keyFacts
	}

	// Step 6: Generate research summary
	summary, err := r.generateResearchSummary(ctx, output)
	if err != nil {
		fmt.Printf("Warning: Failed to generate research summary: %v\n", err)
		summary = "Research completed with " + fmt.Sprintf("%d sources and %d key facts.", len(output.Sources), len(output.KeyFacts))
	}
	output.Summary = summary

	// Step 7: Generate references
	for _, source := range output.Sources {
		if source.URL != "" {
			output.References = append(output.References, fmt.Sprintf("%s - %s", source.Title, source.URL))
		}
	}

	return output, nil
}

// generateResearchTopics uses LLM to identify key research topics
func (r *LLMResearcher) generateResearchTopics(ctx context.Context, content *entities.Content, requirements ResearchRequirements) ([]ResearchTopic, error) {
	prompt := fmt.Sprintf(`
Generate research topics for creating content titled "%s" of type %s.

Content Purpose: %s
Target Audience: %s
Industry: %s
Suggested Topics: %s

Generate 3-5 specific research topics that will help create comprehensive content.
For each topic, provide:
1. The main topic title
2. 3-5 relevant keywords for searching
3. Priority level (1-10, where 10 is highest priority)

Format your response as a structured list with clear sections for each topic.
`,
		content.Title,
		content.Type,
		requirements.ContentPurpose,
		requirements.TargetAudience,
		requirements.Industry,
		strings.Join(requirements.Topics, ", "),
	)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM topic generation failed: %w", err)
	}

	return r.parseResearchTopics(response), nil
}

// parseResearchTopics extracts research topics from LLM response
func (r *LLMResearcher) parseResearchTopics(response string) []ResearchTopic {
	topics := []ResearchTopic{}
	lines := strings.Split(response, "\n")

	var currentTopic *ResearchTopic
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for topic headers
		if strings.Contains(strings.ToLower(line), "topic") && strings.Contains(line, ":") {
			if currentTopic != nil {
				topics = append(topics, *currentTopic)
			}
			
			topicName := extractAfterColon(line)
			currentTopic = &ResearchTopic{
				Topic:      topicName,
				Keywords:   []string{},
				Priority:   5, // Default priority
				MaxSources: 3,
			}
		} else if currentTopic != nil {
			// Look for keywords
			if strings.Contains(strings.ToLower(line), "keyword") {
				keywords := extractKeywordsFromLine(line)
				currentTopic.Keywords = append(currentTopic.Keywords, keywords...)
			}
			
			// Look for priority
			if strings.Contains(strings.ToLower(line), "priority") {
				priority := extractPriorityFromLine(line)
				if priority > 0 {
					currentTopic.Priority = priority
				}
			}
		}
	}

	// Add the last topic
	if currentTopic != nil {
		topics = append(topics, *currentTopic)
	}

	// If no topics were parsed, create default topics from content title
	if len(topics) == 0 {
		topics = append(topics, ResearchTopic{
			Topic:      "Main Topic Research",
			Keywords:   []string{"research", "information", "facts"},
			Priority:   8,
			MaxSources: 5,
		})
	}

	return topics
}

// gatherSources searches for sources related to a research topic
func (r *LLMResearcher) gatherSources(ctx context.Context, topic ResearchTopic, requirements ResearchRequirements) ([]ResearchSource, error) {
	sources := []ResearchSource{}

	// Search for each keyword
	for _, keyword := range topic.Keywords {
		searchQuery := keyword
		if requirements.Industry != "" {
			searchQuery += " " + requirements.Industry
		}

		// Perform web search
		searchResults, err := r.searchService.Search(ctx, searchQuery)
		if err != nil {
			fmt.Printf("Warning: Search failed for keyword '%s': %v\n", keyword, err)
			continue
		}

		// Convert search results to research sources
		for i, result := range searchResults {
			if i >= topic.MaxSources {
				break
			}

			// Fetch content for better analysis
			content, err := r.searchService.FetchContent(ctx, result.URL)
			if err != nil {
				fmt.Printf("Warning: Failed to fetch content from '%s': %v\n", result.URL, err)
				content = result.Snippet // Fallback to snippet
			}

			source := ResearchSource{
				Type:        "web",
				URL:         result.URL,
				Title:       result.Title,
				Content:     content,
				Credibility: 0.5, // Will be evaluated later
				Relevance:   float64(result.Relevance) / 10.0,
				LastUpdated: time.Now(), // We don't have this info from search
			}

			sources = append(sources, source)
		}
	}

	return sources, nil
}

// EvaluateSourceCredibility assesses the credibility of a source using LLM
func (r *LLMResearcher) EvaluateSourceCredibility(ctx context.Context, source ResearchSource) (float64, error) {
	prompt := fmt.Sprintf(`
Evaluate the credibility of this source on a scale of 0.0 to 1.0 (where 1.0 is highly credible):

Source Title: %s
Source URL: %s
Content Preview: %s

Consider factors such as:
- Domain authority and reputation
- Content quality and accuracy
- Author credentials
- Publication date relevance
- Source citation and references

Respond with only a decimal number between 0.0 and 1.0.
`,
		source.Title,
		source.URL,
		truncateString(source.Content, 500),
	)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return 0.5, fmt.Errorf("LLM credibility evaluation failed: %w", err)
	}

	// Extract credibility score from response
	scoreRegex := regexp.MustCompile(`(\d+\.?\d*)`)
	scoreMatch := scoreRegex.FindString(strings.TrimSpace(response))
	if scoreMatch == "" {
		return 0.5, nil // Default moderate credibility
	}

	var score float64
	fmt.Sscanf(scoreMatch, "%f", &score)
	
	// Ensure score is within valid range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// ExtractKeyFacts extracts important facts from research sources using LLM
func (r *LLMResearcher) ExtractKeyFacts(ctx context.Context, sources []ResearchSource) ([]string, error) {
	if len(sources) == 0 {
		return []string{}, nil
	}

	// Combine source content for analysis
	combinedContent := ""
	for i, source := range sources {
		if i >= 10 { // Limit to top 10 sources to avoid overwhelming the LLM
			break
		}
		combinedContent += fmt.Sprintf("Source: %s\n%s\n\n", source.Title, truncateString(source.Content, 300))
	}

	prompt := fmt.Sprintf(`
Extract 5-10 key facts from the following research sources. Focus on:
- Statistical data and numbers
- Important definitions and concepts
- Recent developments or trends
- Expert opinions and quotes
- Actionable insights

Format each fact as a single, clear sentence. Present facts as a numbered list.

Research Sources:
%s
`,
		combinedContent,
	)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return []string{}, fmt.Errorf("LLM fact extraction failed: %w", err)
	}

	return parseFactsList(response), nil
}

// generateResearchSummary creates a summary of the research findings
func (r *LLMResearcher) generateResearchSummary(ctx context.Context, research *ResearchOutput) (string, error) {
	prompt := fmt.Sprintf(`
Create a concise research summary based on the following findings:

Research Topics: %v
Number of Sources: %d
Key Facts:
%s

Write a 2-3 sentence summary highlighting the main research findings and their relevance for content creation.
`,
		extractTopicNames(research.Topics),
		len(research.Sources),
		strings.Join(research.KeyFacts, "\n"),
	)

	response, err := r.llmClient.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM summary generation failed: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// Helper functions

func extractAfterColon(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(line)
}

func extractKeywordsFromLine(line string) []string {
	// Extract text after colon and split by commas
	content := extractAfterColon(line)
	keywords := strings.Split(content, ",")
	
	result := []string{}
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		keyword = strings.Trim(keyword, `"'`)
		if keyword != "" {
			result = append(result, keyword)
		}
	}
	
	return result
}

func extractPriorityFromLine(line string) int {
	// Look for numbers in the line
	priorityRegex := regexp.MustCompile(`(\d+)`)
	match := priorityRegex.FindString(line)
	if match == "" {
		return 0
	}
	
	var priority int
	fmt.Sscanf(match, "%d", &priority)
	
	// Ensure priority is within valid range
	if priority < 1 {
		priority = 1
	} else if priority > 10 {
		priority = 10
	}
	
	return priority
}

func parseFactsList(response string) []string {
	facts := []string{}
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Remove numbering and bullet points
		cleanLine := regexp.MustCompile(`^\d+\.\s*|^-\s*|^\*\s*`).ReplaceAllString(line, "")
		cleanLine = strings.TrimSpace(cleanLine)
		
		if cleanLine != "" && len(cleanLine) > 10 { // Filter out very short lines
			facts = append(facts, cleanLine)
		}
	}
	
	return facts
}

func extractTopicNames(topics []ResearchTopic) []string {
	names := make([]string, len(topics))
	for i, topic := range topics {
		names[i] = topic.Topic
	}
	return names
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
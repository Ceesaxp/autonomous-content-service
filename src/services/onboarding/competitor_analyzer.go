package onboarding

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CompetitorAnalyzerImpl implements the CompetitorAnalyzer interface
type CompetitorAnalyzerImpl struct {
	httpClient *http.Client
	userAgent  string
}

// NewCompetitorAnalyzer creates a new competitor analyzer
func NewCompetitorAnalyzer() *CompetitorAnalyzerImpl {
	return &CompetitorAnalyzerImpl{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "ContentService-Bot/1.0",
	}
}

// AnalyzeWebsite analyzes a competitor website and extracts content strategy insights
func (c *CompetitorAnalyzerImpl) AnalyzeWebsite(websiteURL string) (*WebsiteAnalysis, error) {
	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		websiteURL = parsedURL.String()
	}
	
	// Fetch website content
	req, err := http.NewRequest("GET", websiteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch website: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("website returned status %d", resp.StatusCode)
	}
	
	// Read and analyze content
	var content strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			content.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	
	html := content.String()
	
	// Extract website information
	analysis := &WebsiteAnalysis{
		URL:             websiteURL,
		Title:           c.extractTitle(html),
		Description:     c.extractDescription(html),
		ContentTypes:    c.extractContentTypes(html),
		PublishingFreq:  c.estimatePublishingFrequency(html),
		SocialChannels:  c.extractSocialChannels(html),
		SEOScore:        c.calculateSEOScore(html),
		TechnicalStack:  c.identifyTechnicalStack(html),
		BrandVoice:      c.analyzeBrandVoice(html),
		ContentTopics:   c.extractContentTopics(html),
		Metadata: map[string]interface{}{
			"analyzed_at":    time.Now().Format(time.RFC3339),
			"response_time":  resp.Header.Get("Server"),
			"content_length": len(html),
		},
	}
	
	return analysis, nil
}

// ExtractContentStrategy analyzes multiple competitor websites and extracts content strategies
func (c *CompetitorAnalyzerImpl) ExtractContentStrategy(urls []string) (*ContentStrategy, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs provided")
	}
	
	analyses := make([]*WebsiteAnalysis, 0, len(urls))
	
	// Analyze each website
	for _, url := range urls {
		if url = strings.TrimSpace(url); url == "" {
			continue
		}
		
		analysis, err := c.AnalyzeWebsite(url)
		if err != nil {
			// Log error but continue with other URLs
			fmt.Printf("Failed to analyze %s: %v\n", url, err)
			continue
		}
		
		analyses = append(analyses, analysis)
	}
	
	if len(analyses) == 0 {
		return nil, fmt.Errorf("failed to analyze any of the provided URLs")
	}
	
	// Aggregate insights into content strategy
	strategy := c.aggregateContentStrategy(analyses)
	
	return strategy, nil
}

// IdentifyGaps identifies content and strategy gaps based on client goals and competitor analysis
func (c *CompetitorAnalyzerImpl) IdentifyGaps(clientGoals []string, competitorStrategies []*ContentStrategy) (*GapAnalysis, error) {
	if len(competitorStrategies) == 0 {
		return c.generateBasicGapAnalysis(clientGoals), nil
	}
	
	// Aggregate competitor data
	allContentTypes := make(map[string]int)
	allChannels := make(map[string]int)
	
	for _, strategy := range competitorStrategies {
		for _, contentType := range strategy.ContentFormats {
			allContentTypes[contentType]++
		}
		for _, channel := range strategy.Distribution {
			allChannels[channel]++
		}
		// Additional topic analysis would go here
	}
	
	// Identify gaps based on client goals
	gaps := &GapAnalysis{
		IdentifiedGaps:     []string{},
		Opportunities:      []string{},
		ContentSuggestions: []string{},
		ChannelGaps:        []string{},
		AudienceGaps:       []string{},
		RecommendedFocus:   []string{},
	}
	
	// Analyze content format gaps
	expectedFormats := c.getExpectedFormatsForGoals(clientGoals)
	for _, format := range expectedFormats {
		if allContentTypes[format] == 0 {
			gaps.IdentifiedGaps = append(gaps.IdentifiedGaps, fmt.Sprintf("No competitors using %s", format))
			gaps.Opportunities = append(gaps.Opportunities, fmt.Sprintf("First-mover advantage with %s", format))
		}
	}
	
	// Analyze channel gaps
	expectedChannels := c.getExpectedChannelsForGoals(clientGoals)
	for _, channel := range expectedChannels {
		if allChannels[channel] < len(competitorStrategies)/2 {
			gaps.ChannelGaps = append(gaps.ChannelGaps, channel)
		}
	}
	
	// Generate content suggestions based on goals and gaps
	gaps.ContentSuggestions = c.generateContentSuggestions(clientGoals, gaps.IdentifiedGaps)
	gaps.RecommendedFocus = c.generateRecommendedFocus(clientGoals, gaps)
	
	return gaps, nil
}

// Helper methods for website analysis

func (c *CompetitorAnalyzerImpl) extractTitle(html string) string {
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "No title found"
}

func (c *CompetitorAnalyzerImpl) extractDescription(html string) string {
	descRegex := regexp.MustCompile(`<meta[^>]*name=["\']description["\'][^>]*content=["\']([^"\']+)["\']`)
	matches := descRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "No description found"
}

func (c *CompetitorAnalyzerImpl) extractContentTypes(html string) []string {
	contentTypes := []string{}
	
	// Look for blog indicators
	if strings.Contains(strings.ToLower(html), "blog") || 
	   strings.Contains(strings.ToLower(html), "article") {
		contentTypes = append(contentTypes, "Blog Posts")
	}
	
	// Look for video content
	if strings.Contains(strings.ToLower(html), "video") || 
	   strings.Contains(strings.ToLower(html), "youtube") ||
	   strings.Contains(strings.ToLower(html), "vimeo") {
		contentTypes = append(contentTypes, "Video Content")
	}
	
	// Look for case studies
	if strings.Contains(strings.ToLower(html), "case study") || 
	   strings.Contains(strings.ToLower(html), "case-study") {
		contentTypes = append(contentTypes, "Case Studies")
	}
	
	// Look for whitepapers
	if strings.Contains(strings.ToLower(html), "whitepaper") || 
	   strings.Contains(strings.ToLower(html), "white paper") ||
	   strings.Contains(strings.ToLower(html), "ebook") {
		contentTypes = append(contentTypes, "Whitepapers")
	}
	
	// Look for webinars
	if strings.Contains(strings.ToLower(html), "webinar") || 
	   strings.Contains(strings.ToLower(html), "online event") {
		contentTypes = append(contentTypes, "Webinars")
	}
	
	// Look for newsletters
	if strings.Contains(strings.ToLower(html), "newsletter") || 
	   strings.Contains(strings.ToLower(html), "subscribe") {
		contentTypes = append(contentTypes, "Email Newsletter")
	}
	
	if len(contentTypes) == 0 {
		contentTypes = append(contentTypes, "Website Content")
	}
	
	return contentTypes
}

func (c *CompetitorAnalyzerImpl) estimatePublishingFrequency(html string) string {
	// Simple heuristic based on date patterns and content indicators
	dateRegex := regexp.MustCompile(`(20\d{2}[-/]\d{1,2}[-/]\d{1,2}|20\d{2})`)
	dates := dateRegex.FindAllString(html, -1)
	
	if len(dates) > 10 {
		return "High (Multiple posts per week)"
	} else if len(dates) > 5 {
		return "Medium (Weekly)"
	} else if len(dates) > 0 {
		return "Low (Monthly or less)"
	}
	
	return "Unknown"
}

func (c *CompetitorAnalyzerImpl) extractSocialChannels(html string) []string {
	channels := []string{}
	
	socialPlatforms := map[string]string{
		"linkedin.com":   "LinkedIn",
		"twitter.com":    "Twitter",
		"facebook.com":   "Facebook",
		"instagram.com":  "Instagram",
		"youtube.com":    "YouTube",
		"tiktok.com":     "TikTok",
		"github.com":     "GitHub",
		"medium.com":     "Medium",
	}
	
	lowerHTML := strings.ToLower(html)
	for domain, platform := range socialPlatforms {
		if strings.Contains(lowerHTML, domain) {
			channels = append(channels, platform)
		}
	}
	
	return channels
}

func (c *CompetitorAnalyzerImpl) calculateSEOScore(html string) float64 {
	score := 0.0
	maxScore := 10.0
	
	// Title tag
	if strings.Contains(html, "<title>") {
		score += 1.0
	}
	
	// Meta description
	if strings.Contains(html, `name="description"`) {
		score += 1.0
	}
	
	// H1 tags
	if strings.Contains(html, "<h1") {
		score += 1.0
	}
	
	// H2 tags
	if strings.Contains(html, "<h2") {
		score += 0.5
	}
	
	// Alt attributes
	if strings.Contains(html, `alt="`) {
		score += 0.5
	}
	
	// Structured data
	if strings.Contains(html, "application/ld+json") || 
	   strings.Contains(html, "schema.org") {
		score += 1.0
	}
	
	// Open Graph tags
	if strings.Contains(html, `property="og:`) {
		score += 1.0
	}
	
	// Meta viewport
	if strings.Contains(html, `name="viewport"`) {
		score += 0.5
	}
	
	// Canonical URL
	if strings.Contains(html, `rel="canonical"`) {
		score += 0.5
	}
	
	// SSL (assume HTTPS was used to fetch)
	score += 1.0
	
	// Additional checks could include:
	// - Page speed indicators
	// - Mobile responsiveness
	// - Internal linking
	// - XML sitemap references
	score += 2.0 // Base score for basic functionality
	
	return (score / maxScore) * 100
}

func (c *CompetitorAnalyzerImpl) identifyTechnicalStack(html string) []string {
	stack := []string{}
	
	techIndicators := map[string]string{
		"wp-content":       "WordPress",
		"shopify":          "Shopify",
		"squarespace":      "Squarespace",
		"wix":              "Wix",
		"react":            "React",
		"vue":              "Vue.js",
		"angular":          "Angular",
		"bootstrap":        "Bootstrap",
		"jquery":           "jQuery",
		"gtag":             "Google Analytics",
		"gtm":              "Google Tag Manager",
		"hubspot":          "HubSpot",
		"salesforce":       "Salesforce",
		"intercom":         "Intercom",
		"zendesk":          "Zendesk",
	}
	
	lowerHTML := strings.ToLower(html)
	for indicator, tech := range techIndicators {
		if strings.Contains(lowerHTML, indicator) {
			stack = append(stack, tech)
		}
	}
	
	return stack
}

func (c *CompetitorAnalyzerImpl) analyzeBrandVoice(html string) []string {
	voice := []string{}
	
	// Simple keyword analysis for brand voice
	voiceIndicators := map[string]string{
		"innovative":    "Innovative",
		"professional":  "Professional",
		"friendly":      "Friendly",
		"expert":        "Expert",
		"trusted":       "Trusted",
		"reliable":      "Reliable",
		"cutting-edge":  "Cutting-edge",
		"personalized":  "Personalized",
		"efficient":     "Efficient",
		"affordable":    "Affordable",
		"premium":       "Premium",
		"secure":        "Secure",
	}
	
	lowerHTML := strings.ToLower(html)
	for keyword, trait := range voiceIndicators {
		if strings.Contains(lowerHTML, keyword) {
			voice = append(voice, trait)
		}
	}
	
	// Limit to top 5 traits
	if len(voice) > 5 {
		voice = voice[:5]
	}
	
	return voice
}

func (c *CompetitorAnalyzerImpl) extractContentTopics(html string) []string {
	topics := []string{}
	
	// Extract topics from headings and content
	headingRegex := regexp.MustCompile(`<h[1-6][^>]*>([^<]+)</h[1-6]>`)
	headings := headingRegex.FindAllStringSubmatch(html, -1)
	
	topicMap := make(map[string]int)
	
	for _, heading := range headings {
		if len(heading) > 1 {
			words := strings.Fields(strings.ToLower(heading[1]))
			for _, word := range words {
				if len(word) > 4 { // Only consider longer words
					topicMap[word]++
				}
			}
		}
	}
	
	// Get most frequent topics
	for topic, count := range topicMap {
		if count >= 2 { // Appears at least twice
			// Simple title case - in production use golang.org/x/text/cases
			topics = append(topics, strings.ToUpper(topic[:1])+topic[1:])
		}
	}
	
	// Limit to top 10 topics
	if len(topics) > 10 {
		topics = topics[:10]
	}
	
	return topics
}

func (c *CompetitorAnalyzerImpl) aggregateContentStrategy(analyses []*WebsiteAnalysis) *ContentStrategy {
	if len(analyses) == 0 {
		return &ContentStrategy{}
	}
	
	// Aggregate data from all analyses
	allContentTypes := make(map[string]int)
	allChannels := make(map[string]int)
	allVoice := make(map[string]int)
	
	for _, analysis := range analyses {
		for _, contentType := range analysis.ContentTypes {
			allContentTypes[contentType]++
		}
		for _, channel := range analysis.SocialChannels {
			allChannels[channel]++
		}
		for _, voice := range analysis.BrandVoice {
			allVoice[voice]++
		}
	}
	
	// Extract most common elements
	contentFormats := c.getTopItems(allContentTypes, 5)
	distribution := c.getTopItems(allChannels, 5)
	contentPillars := []string{"Industry Insights", "Product Updates", "Customer Success"} // Default pillars
	
	strategy := &ContentStrategy{
		CompetitorName:  "Aggregated Competitors",
		ContentPillars:  contentPillars,
		PublishingPlan:  c.aggregatePublishingFrequency(analyses),
		TargetAudience:  "Inferred from content analysis",
		ContentFormats:  contentFormats,
		Distribution:    distribution,
		Engagement:      "Varies by platform",
		Strengths:       c.identifyStrengths(analyses),
		Weaknesses:      c.identifyWeaknesses(analyses),
	}
	
	return strategy
}

func (c *CompetitorAnalyzerImpl) getTopItems(itemMap map[string]int, limit int) []string {
	type item struct {
		name  string
		count int
	}
	
	items := make([]item, 0, len(itemMap))
	for name, count := range itemMap {
		items = append(items, item{name, count})
	}
	
	// Simple sorting by count (descending)
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].count < items[j+1].count {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
	
	result := make([]string, 0, limit)
	for i := 0; i < len(items) && i < limit; i++ {
		result = append(result, items[i].name)
	}
	
	return result
}

func (c *CompetitorAnalyzerImpl) aggregatePublishingFrequency(analyses []*WebsiteAnalysis) string {
	frequencies := make(map[string]int)
	
	for _, analysis := range analyses {
		frequencies[analysis.PublishingFreq]++
	}
	
	mostCommon := ""
	maxCount := 0
	
	for freq, count := range frequencies {
		if count > maxCount {
			maxCount = count
			mostCommon = freq
		}
	}
	
	return mostCommon
}

func (c *CompetitorAnalyzerImpl) identifyStrengths(analyses []*WebsiteAnalysis) []string {
	strengths := []string{}
	
	avgSEOScore := 0.0
	for _, analysis := range analyses {
		avgSEOScore += analysis.SEOScore
	}
	avgSEOScore /= float64(len(analyses))
	
	if avgSEOScore > 70 {
		strengths = append(strengths, "Strong SEO optimization")
	}
	
	// Aggregate social presence
	socialCount := 0
	for _, analysis := range analyses {
		socialCount += len(analysis.SocialChannels)
	}
	avgSocial := float64(socialCount) / float64(len(analyses))
	
	if avgSocial > 3 {
		strengths = append(strengths, "Strong social media presence")
	}
	
	// Check for content diversity
	contentTypeCount := 0
	for _, analysis := range analyses {
		contentTypeCount += len(analysis.ContentTypes)
	}
	avgContentTypes := float64(contentTypeCount) / float64(len(analyses))
	
	if avgContentTypes > 3 {
		strengths = append(strengths, "Diverse content formats")
	}
	
	if len(strengths) == 0 {
		strengths = append(strengths, "Established web presence")
	}
	
	return strengths
}

func (c *CompetitorAnalyzerImpl) identifyWeaknesses(analyses []*WebsiteAnalysis) []string {
	weaknesses := []string{}
	
	// Check for common weaknesses
	lowSEOCount := 0
	for _, analysis := range analyses {
		if analysis.SEOScore < 60 {
			lowSEOCount++
		}
	}
	
	if lowSEOCount > len(analyses)/2 {
		weaknesses = append(weaknesses, "Poor SEO optimization")
	}
	
	// Check for limited social presence
	noSocialCount := 0
	for _, analysis := range analyses {
		if len(analysis.SocialChannels) < 2 {
			noSocialCount++
		}
	}
	
	if noSocialCount > len(analyses)/2 {
		weaknesses = append(weaknesses, "Limited social media presence")
	}
	
	// Check for content variety
	limitedContentCount := 0
	for _, analysis := range analyses {
		if len(analysis.ContentTypes) < 2 {
			limitedContentCount++
		}
	}
	
	if limitedContentCount > len(analyses)/2 {
		weaknesses = append(weaknesses, "Limited content format variety")
	}
	
	if len(weaknesses) == 0 {
		weaknesses = append(weaknesses, "Opportunity for differentiation exists")
	}
	
	return weaknesses
}

// Helper methods for gap analysis

func (c *CompetitorAnalyzerImpl) getExpectedFormatsForGoals(goals []string) []string {
	formatMap := map[string][]string{
		"brand_awareness":     {"Blog Posts", "Social Media", "Video Content"},
		"lead_generation":     {"Whitepapers", "Case Studies", "Email Newsletter"},
		"thought_leadership":  {"Blog Posts", "Whitepapers", "Webinars"},
		"customer_retention":  {"Email Newsletter", "Case Studies", "Video Content"},
		"seo_traffic":         {"Blog Posts", "FAQ Content", "Long-form Articles"},
		"social_engagement":   {"Social Media", "Video Content", "Interactive Content"},
	}
	
	expectedFormats := make(map[string]bool)
	
	for _, goal := range goals {
		if formats, exists := formatMap[goal]; exists {
			for _, format := range formats {
				expectedFormats[format] = true
			}
		}
	}
	
	result := make([]string, 0, len(expectedFormats))
	for format := range expectedFormats {
		result = append(result, format)
	}
	
	return result
}

func (c *CompetitorAnalyzerImpl) getExpectedChannelsForGoals(goals []string) []string {
	channelMap := map[string][]string{
		"brand_awareness":    {"LinkedIn", "Twitter", "YouTube"},
		"lead_generation":    {"LinkedIn", "Email", "Website"},
		"thought_leadership": {"LinkedIn", "Medium", "Industry Publications"},
		"social_engagement":  {"LinkedIn", "Twitter", "Instagram", "Facebook"},
		"seo_traffic":        {"Website", "Blog", "YouTube"},
	}
	
	expectedChannels := make(map[string]bool)
	
	for _, goal := range goals {
		if channels, exists := channelMap[goal]; exists {
			for _, channel := range channels {
				expectedChannels[channel] = true
			}
		}
	}
	
	result := make([]string, 0, len(expectedChannels))
	for channel := range expectedChannels {
		result = append(result, channel)
	}
	
	return result
}

func (c *CompetitorAnalyzerImpl) generateContentSuggestions(goals []string, gaps []string) []string {
	suggestions := []string{}
	
	goalSuggestions := map[string][]string{
		"brand_awareness": {
			"Behind-the-scenes content",
			"Company culture stories",
			"Industry insights and trends",
		},
		"lead_generation": {
			"Free tools and calculators",
			"Industry benchmarking reports",
			"Problem-solving guides",
		},
		"thought_leadership": {
			"Original research and surveys",
			"Future predictions and analysis",
			"Expert interview series",
		},
	}
	
	for _, goal := range goals {
		if goalSuggestions, exists := goalSuggestions[goal]; exists {
			suggestions = append(suggestions, goalSuggestions...)
		}
	}
	
	// Add gap-specific suggestions
	if len(gaps) > 0 {
		suggestions = append(suggestions, "Leverage untapped content formats identified in competitive gaps")
	}
	
	return suggestions
}

func (c *CompetitorAnalyzerImpl) generateRecommendedFocus(goals []string, gaps *GapAnalysis) []string {
	focus := []string{}
	
	// Priority based on goals
	for _, goal := range goals {
		switch goal {
		case "brand_awareness":
			focus = append(focus, "Content marketing and storytelling")
		case "lead_generation":
			focus = append(focus, "Educational content and lead magnets")
		case "thought_leadership":
			focus = append(focus, "Original research and expert content")
		case "seo_traffic":
			focus = append(focus, "SEO-optimized content strategy")
		}
	}
	
	// Add gap-based recommendations
	if len(gaps.ChannelGaps) > 0 {
		focus = append(focus, "Multi-channel content distribution")
	}
	
	if len(gaps.IdentifiedGaps) > 0 {
		focus = append(focus, "First-mover advantage in underserved content areas")
	}
	
	// Remove duplicates and limit
	uniqueFocus := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, item := range focus {
		if !seen[item] && len(uniqueFocus) < 5 {
			uniqueFocus = append(uniqueFocus, item)
			seen[item] = true
		}
	}
	
	return uniqueFocus
}

func (c *CompetitorAnalyzerImpl) generateBasicGapAnalysis(goals []string) *GapAnalysis {
	return &GapAnalysis{
		IdentifiedGaps: []string{
			"Limited competitive analysis data available",
			"Opportunity for unique positioning",
		},
		Opportunities: []string{
			"First-mover advantage in identified niches",
			"Differentiation through unique content approach",
		},
		ContentSuggestions: c.generateContentSuggestions(goals, []string{}),
		ChannelGaps:        []string{"All channels available for exploration"},
		AudienceGaps:       []string{"Audience targeting opportunities"},
		RecommendedFocus:   c.generateRecommendedFocus(goals, &GapAnalysis{}),
	}
}
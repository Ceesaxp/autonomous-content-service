package content_creation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LLMClient defines the interface for interacting with an LLM service
type LLMClient interface {
	// Generate creates content using the LLM
	Generate(ctx context.Context, prompt interface{}) (string, error)
}

// ReadabilityScorer defines the interface for measuring content readability
type ReadabilityScorer interface {
	// AnalyzeReadability calculates readability metrics for content
	AnalyzeReadability(ctx context.Context, content string) (float64, []string, error)
}

// SEOAnalyzer defines the interface for SEO analysis
type SEOAnalyzer interface {
	// AnalyzeSEO evaluates content for search engine optimization
	AnalyzeSEO(ctx context.Context, title, content string) (float64, []string, []string, error)
}

// PlagiarismAPI defines the interface for plagiarism detection
type PlagiarismAPI interface {
	// CheckPlagiarism detects potential plagiarism in content
	CheckPlagiarism(ctx context.Context, content string) (float64, []PlagiarismDetail, error)
}

// PlagiarismDetail contains details about detected plagiarism
type PlagiarismDetail struct {
	Fragment   string  `json:"fragment"`
	Source     string  `json:"source"`
	Percentage float64 `json:"percentage"`
}

// SearchService defines the interface for web search and content retrieval
type SearchService interface {
	// Search performs a web search and returns results
	Search(ctx context.Context, query string) ([]SearchResult, error)

	// FetchContent retrieves content from a URL
	FetchContent(ctx context.Context, url string) (string, error)
}

// SearchResult represents a search result
type SearchResult struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Snippet       string `json:"snippet"`
	Authors       string `json:"authors"`
	PublishedDate string `json:"publishedDate"`
	Relevance     int    `json:"relevance"`
}

// OpenAIClient is a client for interacting with the OpenAI API
type OpenAIClient struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	HTTPClient  *http.Client
}

// NewOpenAIClient creates a new client for the OpenAI API
func NewOpenAIClient(apiKey, model string, maxTokens int, temperature float64) *OpenAIClient {
	return &OpenAIClient{
		APIKey:      apiKey,
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenAIMessage represents a message in the OpenAI conversation
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Generate creates content using the OpenAI API
func (c *OpenAIClient) Generate(ctx context.Context, prompt interface{}) (string, error) {
	// Handle different prompt types
	var messages []OpenAIMessage

	switch p := prompt.(type) {
	case string:
		// Simple string prompt
		messages = []OpenAIMessage{
			{Role: "system", Content: "You are a helpful content creation assistant."},
			{Role: "user", Content: p},
		}
	case []string:
		// Context with strings
		messages = []OpenAIMessage{
			{Role: "system", Content: "You are a helpful content creation assistant."},
		}

		// Context strings are added as separate messages
		for i, content := range p {
			role := "user"
			if i > 0 {
				role = "assistant"
				if i%2 == 0 {
					role = "user"
				}
			}
			messages = append(messages, OpenAIMessage{Role: role, Content: content})
		}
	default:
		return "", &InvalidPromptError{PromptType: prompt}
	}

	// Create the request body
	request := OpenAIRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   c.MaxTokens,
		Temperature: c.Temperature,
	}

	// Convert the request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.openai.com/v1/chat/completions",
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	// Parse the response
	var response OpenAIResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", err
	}

	// Extract the generated content
	if len(response.Choices) == 0 {
		return "", &EmptyResponseError{}
	}

	return response.Choices[0].Message.Content, nil
}

// WebSearchService implements the SearchService interface for web search
type WebSearchService struct {
	APIKey     string
	HTTPClient *http.Client
	SearchURL  string
}

// NewWebSearchService creates a new web search service
func NewWebSearchService(apiKey, searchURL string) *WebSearchService {
	return &WebSearchService{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		SearchURL: searchURL,
	}
}

// Search performs a web search and returns results
func (s *WebSearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	// Build the search URL
	params := url.Values{}
	params.Add("q", query)
	params.Add("key", s.APIKey)
	params.Add("num", "10") // Number of results

	searchURL := s.SearchURL + "?" + params.Encode()

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// Send the request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	// Parse the response (implementation depends on the search API used)
	results, err := parseSearchResults(respBody)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// FetchContent retrieves content from a URL
func (s *WebSearchService) FetchContent(ctx context.Context, urlStr string) (string, error) {
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	// Set a user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ContentCreationService/1.0)")

	// Send the request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to fetch content",
		}
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract text content from HTML (simplified; a real implementation would use a proper HTML parser)
	content := extractTextFromHTML(string(respBody))

	return content, nil
}

// parseSearchResults parses search results from API response
func parseSearchResults(respBody []byte) ([]SearchResult, error) {
	// This is a simplified implementation; the actual parsing depends on the search API used
	// For example, if using Google Custom Search API, it would parse the JSON response

	// Placeholder implementation
	var results []SearchResult

	// Parse JSON response (structure depends on the search API)
	// Example for a generic JSON response
	var response struct {
		Items []struct {
			Title       string `json:"title"`
			Link        string `json:"link"`
			Snippet     string `json:"snippet"`
			Authors     string `json:"authors,omitempty"`
			PublishedAt string `json:"publishedAt,omitempty"`
		} `json:"items"`
	}

	err := json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}

	// Convert API-specific format to SearchResult
	for i, item := range response.Items {
		results = append(results, SearchResult{
			Title:         item.Title,
			URL:           item.Link,
			Snippet:       item.Snippet,
			Authors:       item.Authors,
			PublishedDate: item.PublishedAt,
			Relevance:     len(response.Items) - i, // Higher relevance for earlier results
		})
	}

	return results, nil
}

// extractTextFromHTML extracts text content from HTML
func extractTextFromHTML(html string) string {
	// This is a simplified implementation; a real one would use a proper HTML parser
	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")

	// Remove extra whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Remove leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}

// Error types

// InvalidPromptError represents an error with an invalid prompt
type InvalidPromptError struct {
	PromptType interface{}
}

func (e *InvalidPromptError) Error() string {
	return fmt.Sprintf("invalid prompt type: %T", e.PromptType)
}

// APIError represents an error from an API call
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// EmptyResponseError represents an error when an API returns an empty response
type EmptyResponseError struct{}

func (e *EmptyResponseError) Error() string {
	return "API returned an empty response"
}

// BasicReadabilityScorer implements a simple readability scoring mechanism
type BasicReadabilityScorer struct{}

// NewBasicReadabilityScorer creates a new readability scorer
func NewBasicReadabilityScorer() *BasicReadabilityScorer {
	return &BasicReadabilityScorer{}
}

// AnalyzeReadability calculates readability metrics
func (s *BasicReadabilityScorer) AnalyzeReadability(ctx context.Context, content string) (float64, []string, error) {
	// Calculate basic readability metrics
	wordCount := countWords(content)
	sentenceCount := countSentences(content)
	syllableCount := estimateSyllables(content)

	// Calculate Flesch-Kincaid Grade Level
	var fkgl float64
	if sentenceCount > 0 {
		fkgl = 0.39*(float64(wordCount)/float64(sentenceCount)) +
			11.8*(float64(syllableCount)/float64(wordCount)) - 15.59
	}

	// Convert to a score out of 100 (higher is more readable)
	readabilityScore := 100 - (fkgl * 5)
	if readabilityScore > 100 {
		readabilityScore = 100
	} else if readabilityScore < 0 {
		readabilityScore = 0
	}

	// Generate suggestions based on metrics
	suggestions := []string{}

	// Sentence length suggestions
	avgSentenceLength := float64(wordCount) / float64(sentenceCount)
	if avgSentenceLength > 25 {
		suggestions = append(suggestions, "Consider using shorter sentences to improve readability.")
	}

	// Paragraph length suggestions
	paragraphs := strings.Split(content, "\n\n")
	avgParagraphWords := 0
	for _, paragraph := range paragraphs {
		avgParagraphWords += countWords(paragraph)
	}
	avgParagraphWords /= len(paragraphs)

	if avgParagraphWords > 100 {
		suggestions = append(suggestions, "Break down long paragraphs into smaller units for better readability.")
	}

	// Passive voice suggestion (simplified estimation)
	passiveVoiceCount := estimatePassiveVoice(content)
	passiveRatio := float64(passiveVoiceCount) / float64(sentenceCount)
	if passiveRatio > 0.2 {
		suggestions = append(suggestions, "Reduce the use of passive voice to make the content more engaging.")
	}

	// Transition words suggestion
	transitionWordsCount := countTransitionWords(content)
	transitionRatio := float64(transitionWordsCount) / float64(sentenceCount)
	if transitionRatio < 0.25 {
		suggestions = append(suggestions, "Add more transition words to improve the flow between sentences and paragraphs.")
	}

	return readabilityScore, suggestions, nil
}

// countWords counts the number of words in a text
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// countSentences counts the number of sentences in a text
func countSentences(text string) int {
	// Split on sentence ending punctuation
	sentenceRegex := regexp.MustCompile(`[.!?]+`)
	sentences := sentenceRegex.Split(text, -1)

	// Count non-empty sentences
	count := 0
	for _, sentence := range sentences {
		if strings.TrimSpace(sentence) != "" {
			count++
		}
	}

	// If no sentences were found but text exists, count it as one sentence
	if count == 0 && strings.TrimSpace(text) != "" {
		count = 1
	}

	return count
}

// estimateSyllables estimates the number of syllables in a text
func estimateSyllables(text string) int {
	// This is a simplified approximation; a real implementation would be more sophisticated
	words := strings.Fields(strings.ToLower(text))
	syllableCount := 0

	for _, word := range words {
		// Count vowel groups as syllables
		vowelGroups := regexp.MustCompile(`[aeiouy]+`).FindAllString(word, -1)
		count := len(vowelGroups)

		// Adjust for common patterns
		if count > 0 {
			// Silent e at the end
			if strings.HasSuffix(word, "e") && !strings.HasSuffix(word, "le") {
				count--
			}

			// Count at least one syllable per word
			if count == 0 {
				count = 1
			}
		}

		syllableCount += count
	}

	return syllableCount
}

// estimatePassiveVoice estimates the number of passive voice constructions
func estimatePassiveVoice(text string) int {
	// This is a simplified approximation; a real implementation would be more sophisticated
	// Look for common passive voice patterns
	patterns := []string{
		"is [\\w]+(ed|en)", "are [\\w]+(ed|en)",
		"was [\\w]+(ed|en)", "were [\\w]+(ed|en)",
		"be [\\w]+(ed|en)", "been [\\w]+(ed|en)",
		"being [\\w]+(ed|en)",
	}

	count := 0
	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
		count += len(matches)
	}

	return count
}

// countTransitionWords counts transition words in a text
func countTransitionWords(text string) int {
	// Common transition words and phrases
	transitionWords := []string{
		"additionally", "also", "furthermore", "moreover", "similarly",
		"however", "nevertheless", "on the other hand", "in contrast",
		"therefore", "thus", "consequently", "as a result",
		"for example", "for instance", "specifically",
		"first", "second", "third", "finally", "lastly",
		"in conclusion", "to summarize", "in summary",
	}

	count := 0
	lowerText := strings.ToLower(text)

	for _, word := range transitionWords {
		count += strings.Count(lowerText, word)
	}

	return count
}

// BasicSEOAnalyzer implements a simple SEO analysis mechanism
type BasicSEOAnalyzer struct{}

// NewBasicSEOAnalyzer creates a new SEO analyzer
func NewBasicSEOAnalyzer() *BasicSEOAnalyzer {
	return &BasicSEOAnalyzer{}
}

// AnalyzeSEO evaluates content for search engine optimization
func (s *BasicSEOAnalyzer) AnalyzeSEO(ctx context.Context, title, content string) (float64, []string, []string, error) {
	// Extract potential keywords from title and content
	titleWords := extractKeywordsFromText(title)
	//contentWords := extractKeywordsFromText(content)

	// Find most frequently used words in content
	wordFrequency := map[string]int{}
	for _, word := range strings.Fields(strings.ToLower(content)) {
		word = strings.Trim(word, `.,;:"'!?()-`)
		if word != "" {
			wordFrequency[word]++
		}
	}

	// Identify potential keywords based on frequency
	keywords := []string{}
	keywordScores := map[string]int{}

	// First, add title words as they are important
	for _, word := range titleWords {
		keywordScores[word] += 10
	}

	// Then add frequent content words
	for word, freq := range wordFrequency {
		// Skip common stop words
		if isStopWord(word) {
			continue
		}
		keywordScores[word] += freq
	}

	// Sort keywords by score
	type keywordScore struct {
		word  string
		score int
	}
	sortedKeywords := []keywordScore{}
	for word, score := range keywordScores {
		sortedKeywords = append(sortedKeywords, keywordScore{word, score})
	}
	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i].score > sortedKeywords[j].score
	})

	// Take top keywords
	for i, ks := range sortedKeywords {
		if i >= 10 { // Limit to top 10 keywords
			break
		}
		keywords = append(keywords, ks.word)
	}

	// Analyze SEO factors

	// 1. Title analysis
	titleLength := len(title)
	titleScore := 0.0
	if titleLength >= 30 && titleLength <= 60 {
		titleScore = 100.0
	} else if titleLength > 0 {
		titleScore = 60.0
	}

	// 2. Content length analysis
	contentWordCount := countWords(content)
	contentLengthScore := 0.0
	if contentWordCount >= 300 {
		contentLengthScore = 100.0
	} else if contentWordCount >= 100 {
		contentLengthScore = float64(contentWordCount) / 3.0
	}

	// 3. Keyword usage analysis
	keywordUsageScore := 0.0
	keywordDensity := 0.0
	if contentWordCount > 0 {
		keywordCount := 0
		for _, keyword := range keywords {
			keywordCount += strings.Count(strings.ToLower(content), strings.ToLower(keyword))
		}
		keywordDensity = float64(keywordCount) / float64(contentWordCount)
	}

	if keywordDensity >= 0.01 && keywordDensity <= 0.03 {
		keywordUsageScore = 100.0
	} else if keywordDensity > 0.03 {
		keywordUsageScore = 100.0 - ((keywordDensity - 0.03) * 1000.0)
	} else if keywordDensity > 0 {
		keywordUsageScore = keywordDensity * 3000.0
	}

	// 4. Heading analysis
	headingScore := 0.0
	headingCount := countHeadings(content)
	if headingCount > 0 {
		headingScore = 100.0
	}

	// 5. Internal and external links analysis
	linkScore := 0.0
	linkCount := countLinks(content)
	if linkCount > 0 {
		linkScore = 100.0
	}

	// Calculate overall SEO score
	seoScore := (titleScore * 0.2) +
		(contentLengthScore * 0.3) +
		(keywordUsageScore * 0.3) +
		(headingScore * 0.1) +
		(linkScore * 0.1)

	// Generate suggestions
	suggestions := []string{}

	if titleLength < 30 {
		suggestions = append(suggestions, "Make the title longer (aim for 30-60 characters) to improve SEO.")
	} else if titleLength > 60 {
		suggestions = append(suggestions, "Shorten the title (aim for 30-60 characters) for better SEO.")
	}

	if contentWordCount < 300 {
		suggestions = append(suggestions, "Increase content length to at least 300 words for better SEO.")
	}

	if keywordDensity < 0.01 {
		suggestions = append(suggestions, "Increase keyword usage (aim for 1-3% keyword density).")
	} else if keywordDensity > 0.03 {
		suggestions = append(suggestions, "Reduce keyword usage to avoid keyword stuffing (aim for 1-3% keyword density).")
	}

	if headingCount == 0 {
		suggestions = append(suggestions, "Add headings (H1, H2, H3) to structure content and improve SEO.")
	}

	if linkCount == 0 {
		suggestions = append(suggestions, "Add relevant internal or external links to improve SEO.")
	}

	return seoScore, keywords, suggestions, nil
}

// isStopWord checks if a word is a common stop word
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true,
		"is": true, "are": true, "was": true, "were": true,
		"and": true, "or": true, "but": true, "for": true,
		"in": true, "on": true, "at": true, "by": true, "to": true,
		"of": true, "with": true, "about": true,
	}

	return stopWords[word]
}

// countHeadings counts the number of headings in content
func countHeadings(content string) int {
	// This is a simplified implementation; assumes markdown-style headings
	headingRegex := regexp.MustCompile(`(?m)^#+\s.*$`)
	headings := headingRegex.FindAllString(content, -1)
	return len(headings)
}

// countLinks counts the number of links in content
func countLinks(content string) int {
	// This is a simplified implementation; assumes markdown-style links
	linkRegex := regexp.MustCompile(`\[.*?\]\(.*?\)`)
	links := linkRegex.FindAllString(content, -1)
	return len(links)
}

// SimplePlagiarismAPI implements a basic plagiarism detection mechanism
type SimplePlagiarismAPI struct{}

// NewSimplePlagiarismAPI creates a new plagiarism detection API
func NewSimplePlagiarismAPI() *SimplePlagiarismAPI {
	return &SimplePlagiarismAPI{}
}

// CheckPlagiarism detects potential plagiarism in content
func (p *SimplePlagiarismAPI) CheckPlagiarism(ctx context.Context, content string) (float64, []PlagiarismDetail, error) {
	// This is a simplified implementation; a real one would use a proper plagiarism detection service

	// For now, we'll just return a high originality score and no detected plagiarism
	return 0.95, []PlagiarismDetail{}, nil
}

// extractKeywordsFromText extracts keywords from text
func extractKeywordsFromText(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	keywords := []string{}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]")
		// Filter out stop words and short words
		if len(word) > 3 && !isStopWord(word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

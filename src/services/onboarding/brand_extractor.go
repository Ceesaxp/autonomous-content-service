package onboarding

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// BrandVoiceExtractorImpl implements the BrandVoiceExtractor interface
type BrandVoiceExtractorImpl struct {
	// In a real implementation, this would integrate with NLP services
	// like OpenAI, Google Cloud NLP, or custom ML models
}

// NewBrandVoiceExtractor creates a new brand voice extractor
func NewBrandVoiceExtractor() *BrandVoiceExtractorImpl {
	return &BrandVoiceExtractorImpl{}
}

// AnalyzeContent analyzes provided content and extracts brand guidelines
func (b *BrandVoiceExtractorImpl) AnalyzeContent(content []string) (*entities.BrandGuidelines, error) {
	if len(content) == 0 {
		return b.generateDefaultGuidelines(), nil
	}
	
	// Combine all content for analysis
	fullContent := strings.Join(content, "\n\n")
	
	// Perform brand voice analysis
	analysis := b.performBrandAnalysis(fullContent)
	
	// Convert analysis to brand guidelines
	guidelines := &entities.BrandGuidelines{
		Voice:       analysis.Voice,
		Tone:        analysis.ToneAttributes,
		Values:      analysis.Values,
		Personality: analysis.Personality,
		DoNotUse:    b.generateDoNotUse(analysis),
		Examples:    b.extractBestExamples(content, analysis),
	}
	
	return guidelines, nil
}

// ExtractTone analyzes a single piece of content and extracts tone attributes
func (b *BrandVoiceExtractorImpl) ExtractTone(content string) ([]string, error) {
	if content == "" {
		return []string{"professional", "neutral"}, nil
	}
	
	tones := []string{}
	
	// Analyze various tone indicators
	contentLower := strings.ToLower(content)
	
	// Tone analysis based on language patterns
	toneIndicators := map[string][]string{
		"professional": {
			"expertise", "experience", "professional", "industry", "standards",
			"quality", "excellence", "proven", "established", "certified",
		},
		"friendly": {
			"welcome", "thank you", "please", "happy", "excited", "love",
			"enjoy", "wonderful", "amazing", "fantastic", "great",
		},
		"authoritative": {
			"must", "should", "essential", "critical", "important", "required",
			"necessary", "fundamental", "key", "crucial", "vital",
		},
		"conversational": {
			"you", "your", "we", "us", "let's", "here's", "what's",
			"that's", "don't", "can't", "won't", "we're", "you're",
		},
		"technical": {
			"algorithm", "implementation", "framework", "methodology", "optimize",
			"configure", "integrate", "deploy", "architecture", "specification",
		},
		"inspiring": {
			"transform", "achieve", "success", "grow", "innovation", "breakthrough",
			"empower", "unlock", "potential", "vision", "dream", "aspire",
		},
		"educational": {
			"learn", "understand", "guide", "tutorial", "step", "how to",
			"explain", "demonstrate", "teach", "knowledge", "insight",
		},
		"casual": {
			"hey", "hi", "cool", "awesome", "super", "really", "pretty",
			"kinda", "sorta", "stuff", "things", "guys", "folks",
		},
	}
	
	// Count tone indicators
	toneScores := make(map[string]int)
	
	for tone, indicators := range toneIndicators {
		score := 0
		for _, indicator := range indicators {
			score += strings.Count(contentLower, indicator)
		}
		if score > 0 {
			toneScores[tone] = score
		}
	}
	
	// Extract top tones
	type toneScore struct {
		tone  string
		score int
	}
	
	scores := make([]toneScore, 0, len(toneScores))
	for tone, score := range toneScores {
		scores = append(scores, toneScore{tone, score})
	}
	
	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Take top 3-5 tones
	maxTones := 5
	for i, ts := range scores {
		if i >= maxTones {
			break
		}
		tones = append(tones, ts.tone)
	}
	
	// Ensure at least one tone
	if len(tones) == 0 {
		tones = append(tones, "professional")
	}
	
	return tones, nil
}

// IdentifyValues extracts brand values from content
func (b *BrandVoiceExtractorImpl) IdentifyValues(content []string) ([]string, error) {
	if len(content) == 0 {
		return []string{"quality", "customer-focus", "innovation"}, nil
	}
	
	values := []string{}
	fullContent := strings.ToLower(strings.Join(content, " "))
	
	// Value indicators
	valueIndicators := map[string][]string{
		"innovation": {
			"innovative", "innovation", "cutting-edge", "advanced", "revolutionary",
			"breakthrough", "pioneering", "modern", "future", "next-generation",
		},
		"quality": {
			"quality", "excellence", "premium", "superior", "best", "finest",
			"top-tier", "high-quality", "exceptional", "outstanding",
		},
		"integrity": {
			"honest", "integrity", "transparent", "trustworthy", "ethical",
			"reliable", "dependable", "authentic", "genuine", "principled",
		},
		"customer-focus": {
			"customer", "client", "service", "support", "satisfaction", "experience",
			"personalized", "tailored", "dedicated", "responsive",
		},
		"collaboration": {
			"together", "partnership", "collaboration", "teamwork", "community",
			"collective", "shared", "unified", "cooperative", "inclusive",
		},
		"sustainability": {
			"sustainable", "green", "eco", "environment", "responsibility",
			"conservation", "renewable", "carbon", "impact", "future",
		},
		"security": {
			"secure", "security", "protection", "privacy", "safe", "trusted",
			"compliance", "confidential", "encrypted", "verified",
		},
		"efficiency": {
			"efficient", "fast", "quick", "streamlined", "optimized", "effective",
			"productive", "simplified", "automated", "smart",
		},
		"accessibility": {
			"accessible", "inclusive", "everyone", "all", "universal", "easy",
			"simple", "user-friendly", "intuitive", "barrier-free",
		},
		"growth": {
			"growth", "scale", "expand", "develop", "progress", "improve",
			"advance", "evolve", "transform", "achieve",
		},
	}
	
	// Score values based on content
	valueScores := make(map[string]int)
	
	for value, indicators := range valueIndicators {
		score := 0
		for _, indicator := range indicators {
			score += strings.Count(fullContent, indicator)
		}
		if score > 0 {
			valueScores[value] = score
		}
	}
	
	// Extract top values
	type valueScore struct {
		value string
		score int
	}
	
	scores := make([]valueScore, 0, len(valueScores))
	for value, score := range valueScores {
		scores = append(scores, valueScore{value, score})
	}
	
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Take top 5 values
	maxValues := 5
	for i, vs := range scores {
		if i >= maxValues {
			break
		}
		values = append(values, vs.value)
	}
	
	// Default values if none found
	if len(values) == 0 {
		values = []string{"quality", "customer-focus", "innovation"}
	}
	
	return values, nil
}

// GenerateGuidelines creates comprehensive brand guidelines from analysis
func (b *BrandVoiceExtractorImpl) GenerateGuidelines(analysis *BrandAnalysis) (*entities.BrandGuidelines, error) {
	if analysis == nil {
		return b.generateDefaultGuidelines(), nil
	}
	
	guidelines := &entities.BrandGuidelines{
		Voice:       analysis.Voice,
		Tone:        analysis.ToneAttributes,
		Values:      analysis.Values,
		Personality: analysis.Personality,
		DoNotUse:    b.generateDoNotUse(analysis),
		Examples:    analysis.Examples,
	}
	
	return guidelines, nil
}

// performBrandAnalysis conducts comprehensive brand voice analysis
func (b *BrandVoiceExtractorImpl) performBrandAnalysis(content string) *BrandAnalysis {
	if content == "" {
		return b.generateDefaultAnalysis()
	}
	
	// Extract various brand elements
	tones, _ := b.ExtractTone(content)
	values, _ := b.IdentifyValues([]string{content})
	personality := b.extractPersonality(content)
	voice := b.determineOverallVoice(tones, personality)
	writingStyle := b.analyzeWritingStyle(content)
	vocabulary := b.extractKeyVocabulary(content)
	themes := b.extractMessageThemes(content)
	emotionalTone := b.analyzeEmotionalTone(content)
	examples := b.extractBestExamples([]string{content}, nil)
	recommendations := b.generateRecommendations(tones, values, personality)
	
	return &BrandAnalysis{
		Voice:           voice,
		ToneAttributes:  tones,
		Values:          values,
		Personality:     personality,
		WritingStyle:    writingStyle,
		Vocabulary:      vocabulary,
		MessageThemes:   themes,
		EmotionalTone:   emotionalTone,
		Examples:        examples,
		Recommendations: recommendations,
	}
}

func (b *BrandVoiceExtractorImpl) extractPersonality(content string) []string {
	personality := []string{}
	contentLower := strings.ToLower(content)
	
	personalityIndicators := map[string][]string{
		"approachable": {
			"friendly", "welcome", "easy", "simple", "comfortable", "warm",
			"inviting", "accessible", "open", "relaxed",
		},
		"confident": {
			"confident", "expert", "proven", "leader", "best", "superior",
			"trusted", "established", "experienced", "authoritative",
		},
		"innovative": {
			"innovative", "creative", "unique", "original", "revolutionary",
			"cutting-edge", "advanced", "modern", "fresh", "new",
		},
		"reliable": {
			"reliable", "dependable", "consistent", "stable", "secure",
			"trustworthy", "solid", "established", "proven", "guaranteed",
		},
		"passionate": {
			"passionate", "love", "excited", "enthusiastic", "dedicated",
			"committed", "devoted", "motivated", "inspired", "driven",
		},
		"sophisticated": {
			"sophisticated", "elegant", "refined", "premium", "luxury",
			"exclusive", "distinguished", "polished", "cultured", "discerning",
		},
		"helpful": {
			"help", "support", "assist", "guide", "service", "care",
			"solution", "answer", "resource", "useful",
		},
		"authentic": {
			"authentic", "genuine", "real", "honest", "transparent", "true",
			"sincere", "original", "natural", "unfiltered",
		},
	}
	
	// Score personality traits
	personalityScores := make(map[string]int)
	
	for trait, indicators := range personalityIndicators {
		score := 0
		for _, indicator := range indicators {
			score += strings.Count(contentLower, indicator)
		}
		if score > 0 {
			personalityScores[trait] = score
		}
	}
	
	// Extract top personality traits
	type personalityScore struct {
		trait string
		score int
	}
	
	scores := make([]personalityScore, 0, len(personalityScores))
	for trait, score := range personalityScores {
		scores = append(scores, personalityScore{trait, score})
	}
	
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Take top 4 traits
	maxTraits := 4
	for i, ps := range scores {
		if i >= maxTraits {
			break
		}
		personality = append(personality, ps.trait)
	}
	
	// Default personality if none found
	if len(personality) == 0 {
		personality = []string{"professional", "reliable", "helpful"}
	}
	
	return personality
}

func (b *BrandVoiceExtractorImpl) determineOverallVoice(tones, personality []string) string {
	// Combine tones and personality to determine overall voice
	allTraits := append(tones, personality...)
	
	voiceMapping := map[string]string{
		"professional": "Professional and authoritative",
		"friendly":     "Friendly and approachable", 
		"technical":    "Technical and expert-driven",
		"casual":       "Casual and conversational",
		"inspiring":    "Inspiring and motivational",
		"confident":    "Confident and assured",
		"innovative":   "Innovative and forward-thinking",
		"helpful":      "Helpful and supportive",
	}
	
	// Find the most prominent voice characteristic
	for _, trait := range allTraits {
		if voice, exists := voiceMapping[trait]; exists {
			return voice
		}
	}
	
	return "Professional and reliable"
}

func (b *BrandVoiceExtractorImpl) analyzeWritingStyle(content string) string {
	if content == "" {
		return "Professional business writing"
	}
	
	// Analyze sentence structure and complexity
	sentences := strings.Split(content, ".")
	avgSentenceLength := 0
	totalWords := 0
	
	for _, sentence := range sentences {
		words := strings.Fields(sentence)
		totalWords += len(words)
	}
	
	if len(sentences) > 0 {
		avgSentenceLength = totalWords / len(sentences)
	}
	
	// Determine writing style based on characteristics
	contentLower := strings.ToLower(content)
	
	if avgSentenceLength > 20 {
		return "Complex and detailed writing style"
	} else if avgSentenceLength < 10 {
		return "Concise and direct writing style"
	}
	
	if strings.Count(contentLower, "?") > strings.Count(contentLower, ".") {
		return "Inquisitive and engaging writing style"
	}
	
	if strings.Contains(contentLower, "we believe") || strings.Contains(contentLower, "our mission") {
		return "Mission-driven and purposeful writing style"
	}
	
	return "Clear and professional writing style"
}

func (b *BrandVoiceExtractorImpl) extractKeyVocabulary(content string) []string {
	if content == "" {
		return []string{"professional", "quality", "service", "solution"}
	}
	
	// Extract meaningful words (excluding common words)
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true,
		"on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"can": true, "must": true, "this": true, "that": true, "these": true,
		"those": true, "a": true, "an": true, "as": true, "if": true,
		"than": true, "when": true, "where": true, "why": true, "how": true,
	}
	
	// Extract words and count frequency
	words := regexp.MustCompile(`\b\w+\b`).FindAllString(strings.ToLower(content), -1)
	wordFreq := make(map[string]int)
	
	for _, word := range words {
		if len(word) > 3 && !stopWords[word] {
			wordFreq[word]++
		}
	}
	
	// Sort by frequency
	type wordScore struct {
		word  string
		score int
	}
	
	scores := make([]wordScore, 0, len(wordFreq))
	for word, freq := range wordFreq {
		scores = append(scores, wordScore{word, freq})
	}
	
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Return top 10 words
	vocabulary := make([]string, 0, 10)
	for i, ws := range scores {
		if i >= 10 {
			break
		}
		vocabulary = append(vocabulary, ws.word)
	}
	
	return vocabulary
}

func (b *BrandVoiceExtractorImpl) extractMessageThemes(content string) []string {
	if content == "" {
		return []string{"business excellence", "customer satisfaction", "quality service"}
	}
	
	themes := []string{}
	contentLower := strings.ToLower(content)
	
	themeIndicators := map[string][]string{
		"innovation and technology": {
			"innovation", "technology", "digital", "future", "advanced", "modern",
		},
		"customer success": {
			"customer", "success", "satisfaction", "experience", "service", "support",
		},
		"quality and excellence": {
			"quality", "excellence", "best", "superior", "premium", "outstanding",
		},
		"growth and transformation": {
			"growth", "transform", "scale", "expand", "develop", "improve",
		},
		"trust and reliability": {
			"trust", "reliable", "dependable", "secure", "proven", "established",
		},
		"partnership and collaboration": {
			"partner", "collaboration", "together", "team", "community", "relationship",
		},
	}
	
	// Score themes
	themeScores := make(map[string]int)
	
	for theme, indicators := range themeIndicators {
		score := 0
		for _, indicator := range indicators {
			score += strings.Count(contentLower, indicator)
		}
		if score > 0 {
			themeScores[theme] = score
		}
	}
	
	// Extract top themes
	type themeScore struct {
		theme string
		score int
	}
	
	scores := make([]themeScore, 0, len(themeScores))
	for theme, score := range themeScores {
		scores = append(scores, themeScore{theme, score})
	}
	
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Take top 5 themes
	for i, ts := range scores {
		if i >= 5 {
			break
		}
		themes = append(themes, ts.theme)
	}
	
	if len(themes) == 0 {
		themes = []string{"business excellence", "customer focus"}
	}
	
	return themes
}

func (b *BrandVoiceExtractorImpl) analyzeEmotionalTone(content string) string {
	if content == "" {
		return "neutral and professional"
	}
	
	contentLower := strings.ToLower(content)
	
	emotionalIndicators := map[string][]string{
		"positive and optimistic": {
			"great", "excellent", "amazing", "fantastic", "wonderful", "exciting",
			"success", "achieve", "growth", "opportunity", "bright", "positive",
		},
		"confident and assertive": {
			"confident", "proven", "guaranteed", "definitive", "certain", "assured",
			"expert", "leader", "best", "superior", "authority", "established",
		},
		"warm and personal": {
			"welcome", "thank", "appreciate", "personal", "individual", "care",
			"understand", "listen", "support", "family", "community", "together",
		},
		"urgent and action-oriented": {
			"now", "today", "immediately", "urgent", "quickly", "fast", "act",
			"hurry", "limited", "deadline", "soon", "instant", "rapid",
		},
		"calm and reassuring": {
			"calm", "peaceful", "stable", "secure", "safe", "protected", "comfort",
			"relax", "ease", "smooth", "gentle", "steady", "balanced",
		},
	}
	
	// Score emotional tones
	emotionScores := make(map[string]int)
	
	for emotion, indicators := range emotionalIndicators {
		score := 0
		for _, indicator := range indicators {
			score += strings.Count(contentLower, indicator)
		}
		if score > 0 {
			emotionScores[emotion] = score
		}
	}
	
	// Find dominant emotional tone
	maxScore := 0
	dominantEmotion := "neutral and professional"
	
	for emotion, score := range emotionScores {
		if score > maxScore {
			maxScore = score
			dominantEmotion = emotion
		}
	}
	
	return dominantEmotion
}

func (b *BrandVoiceExtractorImpl) extractBestExamples(content []string, analysis *BrandAnalysis) []string {
	if len(content) == 0 {
		return []string{
			"We deliver exceptional results through innovative solutions.",
			"Our team of experts is dedicated to your success.",
			"Let's work together to achieve your goals.",
		}
	}
	
	examples := []string{}
	
	// Extract good examples from content (sentences that are clear and on-brand)
	for _, text := range content {
		sentences := strings.Split(text, ".")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if len(sentence) > 20 && len(sentence) < 150 {
				// Simple quality check
				if strings.Contains(strings.ToLower(sentence), "we") ||
				   strings.Contains(strings.ToLower(sentence), "our") ||
				   strings.Contains(strings.ToLower(sentence), "you") {
					examples = append(examples, sentence+".")
				}
			}
		}
	}
	
	// Limit to best 5 examples
	if len(examples) > 5 {
		examples = examples[:5]
	}
	
	// Ensure we have some examples
	if len(examples) == 0 {
		examples = []string{
			"Our commitment to excellence drives everything we do.",
			"We believe in creating value for our customers.",
			"Together, we can achieve remarkable results.",
		}
	}
	
	return examples
}

func (b *BrandVoiceExtractorImpl) generateRecommendations(tones, values, personality []string) []string {
	recommendations := []string{}
	
	// Generate recommendations based on analysis
	if contains(tones, "professional") {
		recommendations = append(recommendations, "Maintain professional tone while adding warmth to connect with readers")
	}
	
	if contains(tones, "technical") {
		recommendations = append(recommendations, "Balance technical accuracy with accessibility for broader audiences")
	}
	
	if contains(values, "innovation") {
		recommendations = append(recommendations, "Emphasize forward-thinking and cutting-edge approaches in content")
	}
	
	if contains(personality, "helpful") {
		recommendations = append(recommendations, "Focus on providing actionable insights and practical solutions")
	}
	
	if len(recommendations) == 0 {
		recommendations = []string{
			"Maintain consistency across all communication channels",
			"Focus on clear, benefit-driven messaging",
			"Use active voice and concrete examples",
		}
	}
	
	return recommendations
}

func (b *BrandVoiceExtractorImpl) generateDoNotUse(analysis *BrandAnalysis) []string {
	doNotUse := []string{}
	
	// Generate "do not use" based on brand voice
	if contains(analysis.ToneAttributes, "professional") {
		doNotUse = append(doNotUse, "Overly casual language or slang")
	}
	
	if contains(analysis.ToneAttributes, "friendly") {
		doNotUse = append(doNotUse, "Cold, impersonal, or overly formal language")
	}
	
	if contains(analysis.Values, "integrity") {
		doNotUse = append(doNotUse, "Exaggerated claims or unrealistic promises")
	}
	
	if contains(analysis.Personality, "approachable") {
		doNotUse = append(doNotUse, "Complex jargon without explanation")
	}
	
	// Add universal don'ts
	doNotUse = append(doNotUse, 
		"Negative language about competitors",
		"Unclear or ambiguous statements",
		"Generic corporate buzzwords without substance",
	)
	
	return doNotUse
}

func (b *BrandVoiceExtractorImpl) generateDefaultGuidelines() *entities.BrandGuidelines {
	return &entities.BrandGuidelines{
		Voice:       "Professional and approachable",
		Tone:        []string{"professional", "helpful", "confident"},
		Values:      []string{"quality", "customer-focus", "innovation"},
		Personality: []string{"reliable", "expert", "supportive"},
		DoNotUse:    []string{"overly casual language", "negative competitor mentions", "unsubstantiated claims"},
		Examples: []string{
			"We help businesses achieve their content goals through strategic, data-driven approaches.",
			"Our expertise ensures your brand message resonates with your target audience.",
			"Let's create content that drives real results for your business.",
		},
	}
}

func (b *BrandVoiceExtractorImpl) generateDefaultAnalysis() *BrandAnalysis {
	return &BrandAnalysis{
		Voice:           "Professional and reliable",
		ToneAttributes:  []string{"professional", "helpful", "confident"},
		Values:          []string{"quality", "customer-focus", "innovation"},
		Personality:     []string{"reliable", "expert", "supportive"},
		WritingStyle:    "Clear and professional writing style",
		Vocabulary:      []string{"professional", "quality", "service", "solution"},
		MessageThemes:   []string{"business excellence", "customer success"},
		EmotionalTone:   "positive and confident",
		Examples:        []string{},
		Recommendations: []string{"Focus on clear, benefit-driven messaging", "Maintain consistency across channels"},
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
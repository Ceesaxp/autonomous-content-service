package self_improvement

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// capabilityAcquisition implements CapabilityAcquisition interface
type capabilityAcquisition struct {
	capabilityRepo repositories.CapabilityRepository
	httpClient     *http.Client
}

// NewCapabilityAcquisition creates a new capability acquisition service
func NewCapabilityAcquisition(capabilityRepo repositories.CapabilityRepository) CapabilityAcquisition {
	return &capabilityAcquisition{
		capabilityRepo: capabilityRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DiscoverAPIs discovers APIs that can provide a capability
func (ca *capabilityAcquisition) DiscoverAPIs(ctx context.Context, capability string) ([]*APIDiscovery, error) {
	var discoveries []*APIDiscovery
	
	// Search for APIs based on capability type
	switch {
	case strings.Contains(capability, "translation"):
		discoveries = append(discoveries, ca.discoverTranslationAPIs()...)
	case strings.Contains(capability, "image"):
		discoveries = append(discoveries, ca.discoverImageAPIs()...)
	case strings.Contains(capability, "video"):
		discoveries = append(discoveries, ca.discoverVideoAPIs()...)
	case strings.Contains(capability, "analytics"):
		discoveries = append(discoveries, ca.discoverAnalyticsAPIs()...)
	case strings.Contains(capability, "payment"):
		discoveries = append(discoveries, ca.discoverPaymentAPIs()...)
	case strings.Contains(capability, "notification"):
		discoveries = append(discoveries, ca.discoverNotificationAPIs()...)
	default:
		// Generic API search
		discoveries = append(discoveries, ca.searchGenericAPIs(capability)...)
	}
	
	return discoveries, nil
}

// EvaluateAPI evaluates an API for integration
func (ca *capabilityAcquisition) EvaluateAPI(ctx context.Context, api *APIDiscovery) (*APIEvaluation, error) {
	evaluation := &APIEvaluation{
		APIID:       api.Name,
		Performance: make(map[string]float64),
		Risks:       []Risk{},
	}
	
	// Evaluate cost-benefit
	monthlyCost := ca.calculateMonthlyCost(api.Pricing)
	expectedBenefit := ca.estimateBenefit(api.Capabilities)
	evaluation.CostBenefit = expectedBenefit / (monthlyCost + 1) // Avoid division by zero
	
	// Evaluate integration complexity
	integrationScore := ca.evaluateIntegrationComplexity(api)
	evaluation.IntegrationTime = ca.estimateIntegrationTime(integrationScore)
	
	// Evaluate compatibility
	evaluation.Compatibility = ca.evaluateCompatibility(api)
	
	// Evaluate reliability
	evaluation.Reliability = ca.evaluateReliability(api)
	
	// Performance metrics
	evaluation.Performance["rate_limit"] = float64(api.RateLimit)
	evaluation.Performance["response_time"] = ca.estimateResponseTime(api)
	evaluation.Performance["availability"] = ca.estimateAvailability(api)
	
	// Identify risks
	if api.RateLimit < 1000 {
		evaluation.Risks = append(evaluation.Risks, Risk{
			Type:        "rate_limit",
			Description: "Low rate limit may constrain usage",
			Probability: 0.6,
			Impact:      0.5,
			Score:       0.3,
		})
	}
	
	if monthlyCost > 1000 {
		evaluation.Risks = append(evaluation.Risks, Risk{
			Type:        "cost",
			Description: "High monthly cost",
			Probability: 1.0,
			Impact:      0.7,
			Score:       0.7,
		})
	}
	
	// Calculate overall score
	evaluation.Score = (evaluation.CostBenefit*0.3 + 
		evaluation.Compatibility*0.2 + 
		evaluation.Reliability*0.3 + 
		(1-integrationScore)*0.2) / 1.0
	
	// Generate recommendation
	if evaluation.Score > 0.7 {
		evaluation.Recommendation = "Highly recommended for integration"
	} else if evaluation.Score > 0.5 {
		evaluation.Recommendation = "Recommended with considerations"
	} else {
		evaluation.Recommendation = "Not recommended - explore alternatives"
	}
	
	return evaluation, nil
}

// IntegrateAPI integrates an API into the system
func (ca *capabilityAcquisition) IntegrateAPI(ctx context.Context, api *APIDiscovery) error {
	// Create integration configuration
	config := map[string]interface{}{
		"provider":      api.Provider,
		"api_name":      api.Name,
		"documentation": api.Documentation,
		"rate_limit":    api.RateLimit,
		"integrated_at": time.Now(),
	}
	
	// Generate integration code based on SDK availability
	if ca.hasSDKSupport(api) {
		// Use SDK
		config["integration_type"] = "sdk"
		config["sdk_language"] = ca.selectBestSDK(api.SDKLanguages)
	} else {
		// Use REST API
		config["integration_type"] = "rest"
		config["base_url"] = ca.extractBaseURL(api.Documentation)
	}
	
	// Store integration configuration
	// This would be stored in a configuration system
	
	// Update capability gap status
	gaps, err := ca.capabilityRepo.ListCapabilityGaps(ctx, entities.GapStatusApproved)
	if err != nil {
		return fmt.Errorf("listing gaps: %w", err)
	}
	
	// Find matching gap
	for _, gap := range gaps {
		if ca.apiMatchesGap(api, gap) {
			// Record resolution
			resolution := &entities.GapResolution{
				Method:        "api_integration",
				Source:        api.Provider,
				Cost:          ca.calculateMonthlyCost(api.Pricing),
				TimeToResolve: ca.estimateIntegrationTime(0.5),
				Effectiveness: 0.8,
				Details:       config,
				ResolvedAt:    time.Now(),
			}
			
			if err := ca.capabilityRepo.RecordGapResolution(ctx, gap.ID, resolution); err != nil {
				return fmt.Errorf("recording resolution: %w", err)
			}
			
			// Update gap status
			gap.Status = entities.GapStatusResolved
			gap.Resolution = resolution
			if err := ca.capabilityRepo.UpdateCapabilityGap(ctx, gap); err != nil {
				return fmt.Errorf("updating gap: %w", err)
			}
			
			break
		}
	}
	
	return nil
}

// TestIntegration tests an API integration
func (ca *capabilityAcquisition) TestIntegration(ctx context.Context, integrationID string) (*IntegrationTest, error) {
	test := &IntegrationTest{
		IntegrationID: integrationID,
		TestCases:     ca.generateTestCases(integrationID),
		Results:       []TestResult{},
		Performance:   make(map[string]float64),
		Issues:        []string{},
	}
	
	// Run test cases
	for _, testCase := range test.TestCases {
		result := ca.runTestCase(testCase)
		test.Results = append(test.Results, result)
		
		if !result.Passed {
			test.Issues = append(test.Issues, 
				fmt.Sprintf("Test '%s' failed: %s", testCase.Name, result.Error))
		}
	}
	
	// Calculate overall status
	passedCount := 0
	totalDuration := time.Duration(0)
	
	for _, result := range test.Results {
		if result.Passed {
			passedCount++
		}
		totalDuration += result.Duration
	}
	
	passRate := float64(passedCount) / float64(len(test.Results))
	if passRate >= 0.95 {
		test.OverallStatus = "passed"
	} else if passRate >= 0.7 {
		test.OverallStatus = "partial"
	} else {
		test.OverallStatus = "failed"
	}
	
	// Performance metrics
	test.Performance["pass_rate"] = passRate
	test.Performance["avg_duration_ms"] = float64(totalDuration.Milliseconds()) / float64(len(test.Results))
	
	return test, nil
}

// GenerateCapabilityScript generates a script to provide a capability
func (ca *capabilityAcquisition) GenerateCapabilityScript(ctx context.Context, capability string, language string) (*Script, error) {
	script := &Script{
		ID:          fmt.Sprintf("script_%s_%d", capability, time.Now().Unix()),
		Language:    language,
		Description: fmt.Sprintf("Script to provide %s capability", capability),
		Version:     "1.0.0",
	}
	
	// Generate script based on capability and language
	switch language {
	case "python":
		script.Code = ca.generatePythonScript(capability)
		script.Dependencies = ca.getPythonDependencies(capability)
	case "lua":
		script.Code = ca.generateLuaScript(capability)
		script.Dependencies = []string{} // Lua typically has fewer dependencies
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	
	// Define inputs and outputs
	script.Inputs = ca.defineScriptInputs(capability)
	script.Outputs = ca.defineScriptOutputs(capability)
	
	return script, nil
}

// ValidateScript validates a generated script
func (ca *capabilityAcquisition) ValidateScript(ctx context.Context, script *Script) (*ScriptValidation, error) {
	validation := &ScriptValidation{
		ScriptID: script.ID,
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}
	
	// Syntax validation
	syntaxErrors := ca.validateSyntax(script)
	if len(syntaxErrors) > 0 {
		validation.Valid = false
		validation.Errors = append(validation.Errors, syntaxErrors...)
	}
	
	// Security validation
	securityScore, securityIssues := ca.validateSecurity(script)
	validation.SecurityScore = securityScore
	if securityScore < 0.7 {
		validation.Warnings = append(validation.Warnings, securityIssues...)
	}
	
	// Performance estimation
	validation.Performance = ca.estimateScriptPerformance(script)
	
	// Dependency check
	missingDeps := ca.checkDependencies(script)
	if len(missingDeps) > 0 {
		validation.Warnings = append(validation.Warnings, 
			fmt.Sprintf("Missing dependencies: %v", missingDeps))
	}
	
	return validation, nil
}

// DeployScript deploys a validated script
func (ca *capabilityAcquisition) DeployScript(ctx context.Context, script *Script) error {
	// Create deployment package
	deployment := map[string]interface{}{
		"script_id":   script.ID,
		"language":    script.Language,
		"code":        script.Code,
		"deployed_at": time.Now(),
		"version":     script.Version,
		"inputs":      script.Inputs,
		"outputs":     script.Outputs,
	}
	
	// Deploy based on language
	switch script.Language {
	case "python":
		// Create Python execution environment
		deployment["executor"] = "python3"
		deployment["environment"] = ca.createPythonEnvironment(script.Dependencies)
	case "lua":
		// Create Lua execution environment
		deployment["executor"] = "lua"
		deployment["environment"] = map[string]interface{}{}
	}
	
	// Register the new capability
	// This would update the system's capability registry
	
	return nil
}

// Helper methods

func (ca *capabilityAcquisition) discoverTranslationAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "Google",
			Name:         "Google Translate API",
			Description:  "Neural machine translation service",
			Capabilities: []string{"text_translation", "language_detection", "batch_translation"},
			Pricing: map[string]float64{
				"per_character": 0.00002,
				"monthly_free":  500000,
			},
			RateLimit:     3000,
			Documentation: "https://cloud.google.com/translate/docs",
			SDKLanguages:  []string{"python", "go", "java", "node"},
			Requirements:  []string{"api_key", "project_id"},
		},
		{
			Provider:     "DeepL",
			Name:         "DeepL API",
			Description:  "High-quality neural translation",
			Capabilities: []string{"text_translation", "document_translation"},
			Pricing: map[string]float64{
				"per_character": 0.00003,
				"monthly_base":  7.49,
			},
			RateLimit:     1000,
			Documentation: "https://www.deepl.com/docs-api",
			SDKLanguages:  []string{"python", "node"},
			Requirements:  []string{"api_key"},
		},
	}
}

func (ca *capabilityAcquisition) discoverImageAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "OpenAI",
			Name:         "DALL-E API",
			Description:  "AI image generation",
			Capabilities: []string{"image_generation", "image_editing"},
			Pricing: map[string]float64{
				"per_image_1024": 0.02,
				"per_image_512":  0.018,
			},
			RateLimit:     50,
			Documentation: "https://platform.openai.com/docs/guides/images",
			SDKLanguages:  []string{"python", "node"},
			Requirements:  []string{"api_key"},
		},
		{
			Provider:     "Stability AI",
			Name:         "Stable Diffusion API",
			Description:  "Open source image generation",
			Capabilities: []string{"image_generation", "image_to_image"},
			Pricing: map[string]float64{
				"per_image": 0.01,
			},
			RateLimit:     150,
			Documentation: "https://api.stability.ai/docs",
			SDKLanguages:  []string{"python", "javascript"},
			Requirements:  []string{"api_key"},
		},
	}
}

func (ca *capabilityAcquisition) discoverVideoAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "Synthesia",
			Name:         "Synthesia API",
			Description:  "AI video generation with avatars",
			Capabilities: []string{"video_generation", "avatar_creation", "multilingual"},
			Pricing: map[string]float64{
				"per_minute":   2.5,
				"monthly_base": 89,
			},
			RateLimit:     10,
			Documentation: "https://docs.synthesia.io",
			SDKLanguages:  []string{"python", "node"},
			Requirements:  []string{"api_key", "workspace_id"},
		},
	}
}

func (ca *capabilityAcquisition) discoverAnalyticsAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "Google",
			Name:         "Google Analytics API",
			Description:  "Web and app analytics",
			Capabilities: []string{"traffic_analysis", "user_behavior", "conversion_tracking"},
			Pricing: map[string]float64{
				"monthly_free": 10000000,
			},
			RateLimit:     50000,
			Documentation: "https://developers.google.com/analytics",
			SDKLanguages:  []string{"python", "go", "java", "node"},
			Requirements:  []string{"oauth", "property_id"},
		},
	}
}

func (ca *capabilityAcquisition) discoverPaymentAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "PayPal",
			Name:         "PayPal API",
			Description:  "Payment processing and subscriptions",
			Capabilities: []string{"payment_processing", "subscriptions", "invoicing"},
			Pricing: map[string]float64{
				"transaction_percent": 0.029,
				"transaction_fixed":   0.30,
			},
			RateLimit:     10000,
			Documentation: "https://developer.paypal.com/docs",
			SDKLanguages:  []string{"python", "go", "java", "node", "php"},
			Requirements:  []string{"client_id", "client_secret"},
		},
	}
}

func (ca *capabilityAcquisition) discoverNotificationAPIs() []*APIDiscovery {
	return []*APIDiscovery{
		{
			Provider:     "Twilio",
			Name:         "Twilio API",
			Description:  "SMS, voice, and messaging",
			Capabilities: []string{"sms", "voice", "whatsapp", "email"},
			Pricing: map[string]float64{
				"per_sms":   0.0079,
				"per_minute": 0.014,
			},
			RateLimit:     1000,
			Documentation: "https://www.twilio.com/docs",
			SDKLanguages:  []string{"python", "go", "java", "node", "ruby"},
			Requirements:  []string{"account_sid", "auth_token"},
		},
		{
			Provider:     "SendGrid",
			Name:         "SendGrid API",
			Description:  "Email delivery service",
			Capabilities: []string{"transactional_email", "marketing_email", "analytics"},
			Pricing: map[string]float64{
				"monthly_free": 100,
				"per_email":    0.0008,
			},
			RateLimit:     3000,
			Documentation: "https://docs.sendgrid.com",
			SDKLanguages:  []string{"python", "go", "java", "node", "php"},
			Requirements:  []string{"api_key"},
		},
	}
}

func (ca *capabilityAcquisition) searchGenericAPIs(capability string) []*APIDiscovery {
	// This would search API directories or marketplaces
	// For now, return empty
	return []*APIDiscovery{}
}

func (ca *capabilityAcquisition) calculateMonthlyCost(pricing map[string]float64) float64 {
	// Estimate monthly cost based on pricing model
	if base, ok := pricing["monthly_base"]; ok {
		return base
	}
	
	// Estimate usage-based cost
	cost := 0.0
	
	if perChar, ok := pricing["per_character"]; ok {
		// Assume 10M characters/month
		cost += perChar * 10000000
	}
	
	if perImage, ok := pricing["per_image"]; ok {
		// Assume 1000 images/month
		cost += perImage * 1000
	}
	
	if perMin, ok := pricing["per_minute"]; ok {
		// Assume 100 minutes/month
		cost += perMin * 100
	}
	
	if perEmail, ok := pricing["per_email"]; ok {
		// Assume 10000 emails/month
		cost += perEmail * 10000
	}
	
	return cost
}

func (ca *capabilityAcquisition) estimateBenefit(capabilities []string) float64 {
	// Estimate benefit based on capabilities
	benefit := float64(len(capabilities)) * 1000 // Base benefit per capability
	
	// Adjust for high-value capabilities
	for _, cap := range capabilities {
		switch {
		case strings.Contains(cap, "translation"):
			benefit += 2000
		case strings.Contains(cap, "generation"):
			benefit += 3000
		case strings.Contains(cap, "analytics"):
			benefit += 1500
		case strings.Contains(cap, "payment"):
			benefit += 5000
		}
	}
	
	return benefit
}

func (ca *capabilityAcquisition) evaluateIntegrationComplexity(api *APIDiscovery) float64 {
	complexity := 0.0
	
	// Check SDK availability
	if len(api.SDKLanguages) == 0 {
		complexity += 0.3
	} else if !ca.hasPreferredSDK(api.SDKLanguages) {
		complexity += 0.1
	}
	
	// Check authentication complexity
	for _, req := range api.Requirements {
		switch req {
		case "oauth":
			complexity += 0.2
		case "api_key":
			complexity += 0.05
		case "certificate":
			complexity += 0.3
		}
	}
	
	// Check documentation quality (simplified)
	if api.Documentation == "" {
		complexity += 0.3
	}
	
	return math.Min(complexity, 1.0)
}

func (ca *capabilityAcquisition) estimateIntegrationTime(complexity float64) string {
	days := 1 + int(complexity*14) // 1-15 days based on complexity
	
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func (ca *capabilityAcquisition) evaluateCompatibility(api *APIDiscovery) float64 {
	compatibility := 1.0
	
	// Check language support
	if !ca.hasPreferredSDK(api.SDKLanguages) {
		compatibility -= 0.2
	}
	
	// Check rate limits
	if api.RateLimit < 100 {
		compatibility -= 0.3
	} else if api.RateLimit < 1000 {
		compatibility -= 0.1
	}
	
	return math.Max(compatibility, 0)
}

func (ca *capabilityAcquisition) evaluateReliability(api *APIDiscovery) float64 {
	// Base reliability on provider reputation
	providerReliability := map[string]float64{
		"Google":       0.95,
		"OpenAI":       0.9,
		"DeepL":        0.9,
		"Twilio":       0.9,
		"SendGrid":     0.85,
		"PayPal":       0.9,
		"Stability AI": 0.8,
		"Synthesia":    0.85,
	}
	
	if reliability, ok := providerReliability[api.Provider]; ok {
		return reliability
	}
	
	return 0.7 // Default reliability
}

func (ca *capabilityAcquisition) estimateResponseTime(api *APIDiscovery) float64 {
	// Estimate based on capability type
	for _, cap := range api.Capabilities {
		switch {
		case strings.Contains(cap, "generation"):
			return 5000 // 5 seconds for generation
		case strings.Contains(cap, "translation"):
			return 500 // 500ms for translation
		case strings.Contains(cap, "analytics"):
			return 1000 // 1 second for analytics
		}
	}
	
	return 200 // Default 200ms
}

func (ca *capabilityAcquisition) estimateAvailability(api *APIDiscovery) float64 {
	// Estimate based on provider
	return ca.evaluateReliability(api) + 0.04 // Most APIs have 99%+ availability
}

func (ca *capabilityAcquisition) hasSDKSupport(api *APIDiscovery) bool {
	return len(api.SDKLanguages) > 0
}

func (ca *capabilityAcquisition) selectBestSDK(languages []string) string {
	// Prefer Go, then Python
	preferred := []string{"go", "python", "node", "java"}
	
	for _, pref := range preferred {
		for _, lang := range languages {
			if lang == pref {
				return lang
			}
		}
	}
	
	return languages[0]
}

func (ca *capabilityAcquisition) hasPreferredSDK(languages []string) bool {
	preferred := []string{"go", "python"}
	
	for _, pref := range preferred {
		for _, lang := range languages {
			if lang == pref {
				return true
			}
		}
	}
	
	return false
}

func (ca *capabilityAcquisition) extractBaseURL(documentation string) string {
	// Extract API base URL from documentation
	// Simplified implementation
	if strings.Contains(documentation, "api.") {
		parts := strings.Split(documentation, "/")
		if len(parts) >= 3 {
			return parts[0] + "//" + parts[2]
		}
	}
	
	return ""
}

func (ca *capabilityAcquisition) apiMatchesGap(api *APIDiscovery, gap *entities.CapabilityGap) bool {
	// Check if API capabilities match gap description
	gapLower := strings.ToLower(gap.Description)
	
	for _, capability := range api.Capabilities {
		if strings.Contains(gapLower, strings.ToLower(capability)) {
			return true
		}
	}
	
	return false
}

func (ca *capabilityAcquisition) generateTestCases(integrationID string) []TestCase {
	// Generate standard test cases
	return []TestCase{
		{
			ID:   "auth_test",
			Name: "Authentication Test",
			Input: map[string]interface{}{
				"action": "authenticate",
			},
			Expected: map[string]interface{}{
				"status": "success",
			},
			Weight: 1.0,
		},
		{
			ID:   "basic_request",
			Name: "Basic API Request",
			Input: map[string]interface{}{
				"action": "test_request",
			},
			Expected: map[string]interface{}{
				"status": "success",
			},
			Weight: 1.0,
		},
		{
			ID:   "rate_limit",
			Name: "Rate Limit Test",
			Input: map[string]interface{}{
				"action": "burst_requests",
				"count":  10,
			},
			Expected: map[string]interface{}{
				"throttled": false,
			},
			Weight: 0.5,
		},
		{
			ID:   "error_handling",
			Name: "Error Handling Test",
			Input: map[string]interface{}{
				"action": "invalid_request",
			},
			Expected: map[string]interface{}{
				"error_handled": true,
			},
			Weight: 0.8,
		},
	}
}

func (ca *capabilityAcquisition) runTestCase(testCase TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		TestCaseID: testCase.ID,
		Passed:     false,
		Actual:     make(map[string]interface{}),
		Duration:   0,
	}
	
	// Simulate test execution
	// In real implementation, this would call the actual API
	
	switch testCase.ID {
	case "auth_test":
		result.Passed = true
		result.Actual["status"] = "success"
	case "basic_request":
		result.Passed = true
		result.Actual["status"] = "success"
	case "rate_limit":
		result.Passed = true
		result.Actual["throttled"] = false
	case "error_handling":
		result.Passed = true
		result.Actual["error_handled"] = true
	default:
		result.Error = "Unknown test case"
	}
	
	result.Duration = time.Since(start)
	
	// Calculate score
	if result.Passed {
		result.Score = 1.0
	}
	
	return result
}

func (ca *capabilityAcquisition) generatePythonScript(capability string) string {
	// Generate Python script based on capability
	switch {
	case strings.Contains(capability, "text_processing"):
		return ca.pythonTextProcessingScript()
	case strings.Contains(capability, "data_analysis"):
		return ca.pythonDataAnalysisScript()
	case strings.Contains(capability, "web_scraping"):
		return ca.pythonWebScrapingScript()
	default:
		return ca.pythonGenericScript(capability)
	}
}

func (ca *capabilityAcquisition) pythonTextProcessingScript() string {
	return `import re
import nltk
from typing import Dict, List

def process_text(input_data: Dict) -> Dict:
    """Process text with various NLP operations."""
    text = input_data.get('text', '')
    operations = input_data.get('operations', ['tokenize', 'clean'])
    
    result = {'original': text}
    
    if 'clean' in operations:
        # Remove special characters and normalize
        cleaned = re.sub(r'[^\w\s]', '', text)
        cleaned = ' '.join(cleaned.split())
        result['cleaned'] = cleaned
        text = cleaned
    
    if 'tokenize' in operations:
        # Tokenize into words
        tokens = text.split()
        result['tokens'] = tokens
        result['token_count'] = len(tokens)
    
    if 'sentiment' in operations:
        # Simple sentiment analysis
        positive_words = ['good', 'great', 'excellent', 'amazing', 'wonderful']
        negative_words = ['bad', 'terrible', 'awful', 'horrible', 'poor']
        
        text_lower = text.lower()
        pos_count = sum(1 for word in positive_words if word in text_lower)
        neg_count = sum(1 for word in negative_words if word in text_lower)
        
        if pos_count > neg_count:
            result['sentiment'] = 'positive'
        elif neg_count > pos_count:
            result['sentiment'] = 'negative'
        else:
            result['sentiment'] = 'neutral'
    
    return result

if __name__ == "__main__":
    # Test the function
    test_input = {
        'text': 'This is a great example of text processing!',
        'operations': ['clean', 'tokenize', 'sentiment']
    }
    print(process_text(test_input))
`
}

func (ca *capabilityAcquisition) pythonDataAnalysisScript() string {
	return `import pandas as pd
import numpy as np
from typing import Dict, List

def analyze_data(input_data: Dict) -> Dict:
    """Perform data analysis on provided dataset."""
    data = input_data.get('data', [])
    analysis_type = input_data.get('analysis_type', 'summary')
    
    if not data:
        return {'error': 'No data provided'}
    
    # Convert to DataFrame
    df = pd.DataFrame(data)
    result = {}
    
    if analysis_type == 'summary':
        # Basic statistical summary
        result['shape'] = df.shape
        result['columns'] = df.columns.tolist()
        result['dtypes'] = df.dtypes.to_dict()
        result['summary'] = df.describe().to_dict()
        result['null_counts'] = df.isnull().sum().to_dict()
    
    elif analysis_type == 'correlation':
        # Correlation analysis for numeric columns
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        if len(numeric_cols) > 1:
            result['correlation'] = df[numeric_cols].corr().to_dict()
    
    elif analysis_type == 'groupby':
        # Group by analysis
        group_col = input_data.get('group_column')
        agg_col = input_data.get('aggregate_column')
        if group_col and agg_col and group_col in df.columns and agg_col in df.columns:
            grouped = df.groupby(group_col)[agg_col].agg(['mean', 'sum', 'count'])
            result['grouped'] = grouped.to_dict()
    
    return result

if __name__ == "__main__":
    # Test the function
    test_input = {
        'data': [
            {'category': 'A', 'value': 10},
            {'category': 'B', 'value': 20},
            {'category': 'A', 'value': 15},
            {'category': 'B', 'value': 25}
        ],
        'analysis_type': 'summary'
    }
    print(analyze_data(test_input))
`
}

func (ca *capabilityAcquisition) pythonWebScrapingScript() string {
	return `import requests
from bs4 import BeautifulSoup
from typing import Dict, List
import time

def scrape_web(input_data: Dict) -> Dict:
    """Scrape web content based on provided parameters."""
    url = input_data.get('url', '')
    selectors = input_data.get('selectors', {})
    headers = input_data.get('headers', {
        'User-Agent': 'Mozilla/5.0 (compatible; AutoBot/1.0)'
    })
    
    if not url:
        return {'error': 'No URL provided'}
    
    try:
        # Make request
        response = requests.get(url, headers=headers, timeout=10)
        response.raise_for_status()
        
        # Parse HTML
        soup = BeautifulSoup(response.content, 'html.parser')
        result = {
            'url': url,
            'status_code': response.status_code,
            'content': {}
        }
        
        # Extract content based on selectors
        for key, selector in selectors.items():
            if selector.startswith('.'):
                # Class selector
                elements = soup.find_all(class_=selector[1:])
            elif selector.startswith('#'):
                # ID selector
                elements = [soup.find(id=selector[1:])]
            else:
                # Tag selector
                elements = soup.find_all(selector)
            
            # Extract text from elements
            result['content'][key] = [
                elem.get_text(strip=True) for elem in elements if elem
            ]
        
        # If no selectors, return page title and meta description
        if not selectors:
            title = soup.find('title')
            result['content']['title'] = title.get_text() if title else ''
            
            meta_desc = soup.find('meta', attrs={'name': 'description'})
            if meta_desc:
                result['content']['description'] = meta_desc.get('content', '')
        
        return result
        
    except requests.RequestException as e:
        return {'error': f'Request failed: {str(e)}'}
    except Exception as e:
        return {'error': f'Scraping failed: {str(e)}'}

if __name__ == "__main__":
    # Test the function
    test_input = {
        'url': 'https://example.com',
        'selectors': {
            'headings': 'h1',
            'paragraphs': 'p'
        }
    }
    print(scrape_web(test_input))
`
}

func (ca *capabilityAcquisition) pythonGenericScript(capability string) string {
	return fmt.Sprintf(`from typing import Dict, Any

def process_%s(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process %s capability."""
    # Input validation
    required_fields = ['data']
    for field in required_fields:
        if field not in input_data:
            return {'error': f'Missing required field: {field}'}
    
    data = input_data['data']
    options = input_data.get('options', {})
    
    # Process the capability
    result = {
        'status': 'success',
        'capability': '%s',
        'processed_data': data,
        'options_used': options
    }
    
    # Add capability-specific processing here
    
    return result

if __name__ == "__main__":
    # Test the function
    test_input = {
        'data': 'test data',
        'options': {'verbose': True}
    }
    print(process_%s(test_input))
`, strings.Replace(capability, " ", "_", -1), capability, capability, strings.Replace(capability, " ", "_", -1))
}

func (ca *capabilityAcquisition) generateLuaScript(capability string) string {
	// Generate Lua script based on capability
	return fmt.Sprintf(`-- %s capability script

function process_%s(input_data)
    -- Validate input
    if not input_data or not input_data.data then
        return {error = "Missing required field: data"}
    end
    
    local data = input_data.data
    local options = input_data.options or {}
    
    -- Process the capability
    local result = {
        status = "success",
        capability = "%s",
        processed_data = data,
        options_used = options
    }
    
    -- Add capability-specific processing here
    
    return result
end

-- Export the function
return {
    process = process_%s
}
`, capability, strings.Replace(capability, " ", "_", -1), capability, strings.Replace(capability, " ", "_", -1))
}

func (ca *capabilityAcquisition) getPythonDependencies(capability string) []string {
	deps := []string{"typing"}
	
	switch {
	case strings.Contains(capability, "text"):
		deps = append(deps, "nltk", "re")
	case strings.Contains(capability, "data"):
		deps = append(deps, "pandas", "numpy")
	case strings.Contains(capability, "web"):
		deps = append(deps, "requests", "beautifulsoup4")
	case strings.Contains(capability, "image"):
		deps = append(deps, "pillow", "opencv-python")
	}
	
	return deps
}

func (ca *capabilityAcquisition) defineScriptInputs(capability string) []ScriptInput {
	// Define common inputs
	inputs := []ScriptInput{
		{
			Name:        "data",
			Type:        "any",
			Description: "The data to process",
			Required:    true,
		},
		{
			Name:        "options",
			Type:        "object",
			Description: "Processing options",
			Required:    false,
			Default:     map[string]interface{}{},
		},
	}
	
	// Add capability-specific inputs
	switch {
	case strings.Contains(capability, "text"):
		inputs = append(inputs, ScriptInput{
			Name:        "operations",
			Type:        "array",
			Description: "Text operations to perform",
			Required:    false,
			Default:     []string{"clean", "tokenize"},
		})
	case strings.Contains(capability, "analysis"):
		inputs = append(inputs, ScriptInput{
			Name:        "analysis_type",
			Type:        "string",
			Description: "Type of analysis to perform",
			Required:    false,
			Default:     "summary",
		})
	}
	
	return inputs
}

func (ca *capabilityAcquisition) defineScriptOutputs(capability string) []ScriptOutput {
	// Define common outputs
	outputs := []ScriptOutput{
		{
			Name:        "status",
			Type:        "string",
			Description: "Processing status",
		},
		{
			Name:        "result",
			Type:        "object",
			Description: "Processing result",
		},
	}
	
	// Add capability-specific outputs
	switch {
	case strings.Contains(capability, "text"):
		outputs = append(outputs, ScriptOutput{
			Name:        "tokens",
			Type:        "array",
			Description: "Tokenized text",
		})
	case strings.Contains(capability, "analysis"):
		outputs = append(outputs, ScriptOutput{
			Name:        "statistics",
			Type:        "object",
			Description: "Statistical analysis results",
		})
	}
	
	return outputs
}

func (ca *capabilityAcquisition) validateSyntax(script *Script) []string {
	var errors []string
	
	// Basic syntax validation
	switch script.Language {
	case "python":
		// Check for basic Python syntax errors
		if !strings.Contains(script.Code, "def ") {
			errors = append(errors, "No function definition found")
		}
		if !strings.Contains(script.Code, "return ") {
			errors = append(errors, "No return statement found")
		}
	case "lua":
		// Check for basic Lua syntax errors
		if !strings.Contains(script.Code, "function ") {
			errors = append(errors, "No function definition found")
		}
		if !strings.Contains(script.Code, "return ") {
			errors = append(errors, "No return statement found")
		}
	}
	
	return errors
}

func (ca *capabilityAcquisition) validateSecurity(script *Script) (float64, []string) {
	score := 1.0
	var issues []string
	
	// Check for dangerous operations
	dangerous := []string{
		"exec", "eval", "compile", "__import__",
		"subprocess", "os.system", "open(",
	}
	
	for _, danger := range dangerous {
		if strings.Contains(script.Code, danger) {
			score -= 0.2
			issues = append(issues, fmt.Sprintf("Potentially dangerous operation: %s", danger))
		}
	}
	
	// Check for network operations
	if strings.Contains(script.Code, "requests") || strings.Contains(script.Code, "urllib") {
		score -= 0.1
		issues = append(issues, "Network operations detected")
	}
	
	// Check for file operations
	if strings.Contains(script.Code, "open(") || strings.Contains(script.Code, "file") {
		score -= 0.1
		issues = append(issues, "File operations detected")
	}
	
	return math.Max(score, 0), issues
}

func (ca *capabilityAcquisition) estimateScriptPerformance(script *Script) float64 {
	// Estimate based on script complexity
	lines := strings.Count(script.Code, "\n")
	
	// Base performance score
	performance := 1.0
	
	// Adjust based on lines of code
	if lines > 100 {
		performance -= 0.1
	}
	if lines > 500 {
		performance -= 0.2
	}
	
	// Adjust based on dependencies
	if len(script.Dependencies) > 5 {
		performance -= 0.1
	}
	
	// Check for performance-intensive operations
	if strings.Contains(script.Code, "for ") || strings.Contains(script.Code, "while ") {
		performance -= 0.05
	}
	
	return math.Max(performance, 0.1)
}

func (ca *capabilityAcquisition) checkDependencies(script *Script) []string {
	var missing []string
	
	// This would check against installed packages
	// For now, simulate some common missing deps
	for _, dep := range script.Dependencies {
		switch dep {
		case "typing", "re": // Built-in, always available
			continue
		case "pandas", "numpy", "nltk", "beautifulsoup4":
			// Simulate random missing deps
			if time.Now().Unix()%3 == 0 {
				missing = append(missing, dep)
			}
		}
	}
	
	return missing
}

func (ca *capabilityAcquisition) createPythonEnvironment(dependencies []string) map[string]interface{} {
	return map[string]interface{}{
		"python_version": "3.9",
		"dependencies":   dependencies,
		"virtual_env":    true,
		"pip_install":    strings.Join(dependencies, " "),
	}
}
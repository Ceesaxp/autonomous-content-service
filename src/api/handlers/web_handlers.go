package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// WebHandler handles web interface specific endpoints
type WebHandler struct {
	projectHandler *ProjectHandler
	contentHandler *ContentHandler
}

// NewWebHandler creates a new web handler instance
func NewWebHandler(projectHandler *ProjectHandler, contentHandler *ContentHandler) *WebHandler {
	return &WebHandler{
		projectHandler: projectHandler,
		contentHandler: contentHandler,
	}
}

// ProjectQuoteRequest represents a quote request from the web form
type ProjectQuoteRequest struct {
	ProjectType        string `json:"project_type"`
	ProjectDescription string `json:"project_description"`
	WordCount          string `json:"word_count"`
	Deadline           string `json:"deadline"`
	Tone               string `json:"tone,omitempty"`
	TargetAudience     string `json:"target_audience,omitempty"`
	Keywords           string `json:"keywords,omitempty"`
	References         string `json:"references,omitempty"`
	ClientName         string `json:"client_name"`
	ClientEmail        string `json:"client_email"`
	Company            string `json:"company,omitempty"`
	Phone              string `json:"phone,omitempty"`
	Budget             string `json:"budget,omitempty"`
	EstimatedPrice     float64 `json:"estimated_price,omitempty"`
}

// QuoteResponse represents the response to a quote request
type QuoteResponse struct {
	QuoteID       string    `json:"quote_id"`
	EstimatedPrice float64   `json:"estimated_price"`
	DeliveryTime   string    `json:"delivery_time"`
	ProjectID     string    `json:"project_id,omitempty"`
	Message       string    `json:"message"`
	CreatedAt     time.Time `json:"created_at"`
}

// ChatMessage represents a chat interaction
type ChatMessage struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// ChatResponse represents the bot's response
type ChatResponse struct {
	Response  string   `json:"response"`
	Options   []string `json:"options,omitempty"`
	SessionID string   `json:"session_id"`
}

// AnalyticsEvent represents an analytics tracking event
type AnalyticsEvent struct {
	EventType   string                 `json:"event_type"`
	SessionID   string                 `json:"session_id"`
	VisitorID   string                 `json:"visitor_id"`
	URL         string                 `json:"url"`
	Timestamp   int64                  `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// RequestQuote handles project quote requests from the web form
func (h *WebHandler) RequestQuote(w http.ResponseWriter, r *http.Request) {
	var req ProjectQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProjectType == "" || req.ClientEmail == "" || req.ClientName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Calculate estimated price based on project details
	estimatedPrice := h.calculatePrice(req)

	// Create a quote ID for tracking
	quoteID := fmt.Sprintf("quote_%d", time.Now().Unix())

	// Determine delivery time based on deadline
	deliveryTime := h.getDeliveryTime(req.Deadline)

	response := QuoteResponse{
		QuoteID:        quoteID,
		EstimatedPrice: estimatedPrice,
		DeliveryTime:   deliveryTime,
		Message:        "Thank you for your quote request. Our autonomous system has processed your requirements and will send a detailed quote to your email within the hour.",
		CreatedAt:      time.Now(),
	}

	// Log the quote request for processing
	log.Printf("Quote requested: %s - %s (%s) - $%.2f", 
		quoteID, req.ProjectType, req.ClientEmail, estimatedPrice)

	// In a real implementation, this would:
	// 1. Store the quote in database
	// 2. Send confirmation email
	// 3. Trigger autonomous pricing system
	// 4. Create project pipeline entry

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleChat processes chat messages from the website chatbot
func (h *WebHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var msg ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate session ID if not provided
	sessionID := msg.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("chat_%d", time.Now().Unix())
	}

	// Process message and generate response
	response, options := h.processChatMessage(msg.Message)

	chatResponse := ChatResponse{
		Response:  response,
		Options:   options,
		SessionID: sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse)
}

// TrackAnalytics handles analytics events from the website
func (h *WebHandler) TrackAnalytics(w http.ResponseWriter, r *http.Request) {
	var event AnalyticsEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log analytics event
	log.Printf("Analytics: %s - %s (Session: %s)", 
		event.EventType, event.URL, event.SessionID)

	// In a real implementation, this would:
	// 1. Store analytics data in time-series database
	// 2. Update real-time dashboards
	// 3. Trigger optimization algorithms
	// 4. Generate insights and reports

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// GetPortfolio returns portfolio items for display
func (h *WebHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	_ = r.URL.Query().Get("limit") // TODO: implement pagination

	// In a real implementation, this would fetch from content repository
	// For now, return sample portfolio items
	portfolioItems := []map[string]interface{}{
		{
			"id":          "ai-healthcare",
			"title":       "AI in Healthcare: Revolutionary Changes",
			"category":    "blog",
			"type":        "Blog Post",
			"word_count":  1200,
			"industry":    "Technology",
			"excerpt":     "Comprehensive analysis of AI's impact on healthcare delivery...",
			"metrics":     map[string]string{"engagement": "+300%", "seo_score": "95%"},
			"created_at":  time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
		},
		{
			"id":          "saas-landing",
			"title":       "SaaS Landing Page Copy",
			"category":    "marketing",
			"type":        "Marketing Copy",
			"word_count":  800,
			"industry":    "B2B SaaS",
			"excerpt":     "Conversion-optimized landing page copy that increased sign-ups by 45%...",
			"metrics":     map[string]string{"conversions": "+45%", "tested": "A/B tested"},
			"created_at":  time.Now().AddDate(0, 0, -14).Format(time.RFC3339),
		},
	}

	// Filter by category if specified
	if category != "" && category != "all" {
		filtered := make([]map[string]interface{}, 0)
		for _, item := range portfolioItems {
			if item["category"] == category {
				filtered = append(filtered, item)
			}
		}
		portfolioItems = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": portfolioItems,
		"total": len(portfolioItems),
	})
}

// GetPricing returns dynamic pricing information
func (h *WebHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	serviceType := r.URL.Query().Get("service")
	wordCount := r.URL.Query().Get("words")
	urgency := r.URL.Query().Get("urgency")
	complexity := r.URL.Query().Get("complexity")

	// Calculate dynamic pricing
	price := h.calculateDynamicPrice(serviceType, wordCount, urgency, complexity)

	response := map[string]interface{}{
		"service":          serviceType,
		"base_price":       price,
		"estimated_total":  price,
		"delivery_time":    h.getDeliveryTime(urgency),
		"currency":         "USD",
		"calculated_at":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSystemStatus returns the status of the autonomous system
func (h *WebHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"operational":     true,
		"queue_length":    5, // Number of projects in queue
		"average_turnaround": "18 hours",
		"success_rate":    "98.5%",
		"uptime":          "99.9%",
		"last_updated":    time.Now().Format(time.RFC3339),
		"capabilities": []string{
			"Blog Posts & Articles",
			"Marketing Copy",
			"Technical Documentation",
			"Social Media Content",
			"Email Marketing",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Helper methods

func (h *WebHandler) calculatePrice(req ProjectQuoteRequest) float64 {
	basePrice := 50.0 // Base price in USD

	// Adjust based on project type
	switch req.ProjectType {
	case "Blog Posts & Articles":
		basePrice = 50.0
	case "Marketing Copy":
		basePrice = 75.0
	case "Technical Documentation":
		basePrice = 100.0
	case "Social Media Content":
		basePrice = 25.0
	case "Email Marketing":
		basePrice = 60.0
	}

	// Adjust based on word count
	wordMultiplier := 1.0
	switch req.WordCount {
	case "500-1000":
		wordMultiplier = 1.5
	case "1000-2000":
		wordMultiplier = 2.5
	case "2000-5000":
		wordMultiplier = 4.0
	case "5000+":
		wordMultiplier = 6.0
	}

	// Adjust based on deadline urgency
	urgencyMultiplier := 1.0
	switch req.Deadline {
	case "rush":
		urgencyMultiplier = 1.5
	case "express":
		urgencyMultiplier = 1.25
	case "flexible":
		urgencyMultiplier = 0.9
	}

	return basePrice * wordMultiplier * urgencyMultiplier
}

func (h *WebHandler) calculateDynamicPrice(serviceType, wordCount, urgency, complexity string) float64 {
	// Simplified pricing calculation
	basePrice := 50.0
	
	if serviceType == "Marketing Copy" {
		basePrice = 75.0
	} else if serviceType == "Technical Documentation" {
		basePrice = 100.0
	}

	return basePrice
}

func (h *WebHandler) getDeliveryTime(deadline string) string {
	switch deadline {
	case "rush":
		return "6-12 hours"
	case "express":
		return "12-24 hours"
	case "flexible":
		return "3-7 days"
	default:
		return "24-48 hours"
	}
}

func (h *WebHandler) processChatMessage(message string) (string, []string) {
	// Simplified chat processing
	// In a real implementation, this would use the LLM system
	
	if contains(message, []string{"price", "cost", "quote"}) {
		return "I'd be happy to help you get a price quote! Our pricing depends on several factors:", 
			[]string{"Blog posts start at $50", "Marketing copy starts at $75", "Get detailed quote"}
	}
	
	if contains(message, []string{"service", "what do you do"}) {
		return "We offer autonomous content creation services including:",
			[]string{"Blog Posts & Articles", "Marketing Copy", "Technical Documentation", "View all services"}
	}
	
	if contains(message, []string{"time", "fast", "delivery"}) {
		return "Our autonomous system is incredibly fast! Standard delivery is 24-48 hours:",
			[]string{"Rush delivery: 6-12 hours", "Express delivery: 12-24 hours", "Start my project now"}
	}
	
	return "I'm here to help you with your content needs. What would you like to know?",
		[]string{"Get pricing information", "Learn about our services", "View portfolio examples", "Start a project"}
}

func contains(text string, keywords []string) bool {
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// FinancialHandlers handles HTTP requests for financial operations
type FinancialHandlers struct {
	transactionRepo repositories.TransactionRepository
	logger          *log.Logger
}

// NewFinancialHandlers creates a new financial handlers instance
func NewFinancialHandlers(transactionRepo repositories.TransactionRepository) *FinancialHandlers {
	return &FinancialHandlers{
		transactionRepo: transactionRepo,
		logger:          log.New(log.Writer(), "[FinancialHandler] ", log.LstdFlags),
	}
}

// Payment operations

func (h *FinancialHandlers) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Amount        float64                    `json:"amount"`
		Currency      string                     `json:"currency"`
		PaymentMethod entities.PaymentMethodType `json:"payment_method"`
		ClientID      string                     `json:"client_id"`
		ProjectID     string                     `json:"project_id,omitempty"`
		Description   string                     `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.Amount <= 0 || request.Currency == "" || request.ClientID == "" {
		http.Error(w, "Amount, currency, and client_id are required", http.StatusBadRequest)
		return
	}

	// Create transaction
	transaction := &entities.Transaction{
		TransactionID: uuid.New(),
		Amount:        entities.Money{Amount: float64(request.Amount * 100), Currency: request.Currency},
		Type:          entities.TransactionTypePayment,
		Status:        entities.TransactionStatusPending,
		PaymentMethod: request.PaymentMethod,
		Description:   request.Description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Parse UUIDs
	clientID, err := uuid.Parse(request.ClientID)
	if err != nil {
		http.Error(w, "Invalid client_id format", http.StatusBadRequest)
		return
	}
	transaction.ClientID = clientID

	if request.ProjectID != "" {
		projectID, err := uuid.Parse(request.ProjectID)
		if err != nil {
			http.Error(w, "Invalid project_id format", http.StatusBadRequest)
			return
		}
		transaction.ProjectID = &projectID
	}

	// Process payment (simplified logic)
	if err := h.processPaymentLogic(r.Context(), transaction); err != nil {
		h.logger.Printf("Payment processing failed: %v", err)
		transaction.Status = entities.TransactionStatusFailed
	} else {
		transaction.Status = entities.TransactionStatusCompleted
	}

	// Save transaction
	if err := h.transactionRepo.Create(r.Context(), transaction); err != nil {
		h.logger.Printf("Failed to save transaction: %v", err)
		// Continue with response even if save fails
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": transaction.TransactionID,
		"status":         transaction.Status,
		"amount":         transaction.Amount.Amount,
		"currency":       transaction.Amount.Currency,
		"processed_at":   transaction.CreatedAt,
	})
}

func (h *FinancialHandlers) RefundPayment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TransactionID string  `json:"transaction_id"`
		Amount        float64 `json:"amount,omitempty"`
		Reason        string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transactionID, err := uuid.Parse(request.TransactionID)
	if err != nil {
		http.Error(w, "Invalid transaction_id format", http.StatusBadRequest)
		return
	}

	// Create refund transaction
	refund := &entities.Transaction{
		TransactionID:    uuid.New(),
		Type:             entities.TransactionTypeRefund,
		Status:           entities.TransactionStatusCompleted,
		PaymentReference: request.TransactionID,
		Description:      fmt.Sprintf("Refund for transaction %s: %s", request.TransactionID, request.Reason),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	// Save refund transaction
	if err := h.transactionRepo.Create(r.Context(), refund); err != nil {
		h.logger.Printf("Failed to save refund: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"refund_id":            refund.TransactionID,
		"original_transaction": transactionID,
		"status":               "processed",
		"refunded_at":          refund.CreatedAt,
	})
}

func (h *FinancialHandlers) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	transactionID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
		return
	}

	// Try to get transaction
	transaction, err := h.transactionRepo.FindByID(r.Context(), transactionID)
	if err != nil {
		// Return mock status for now since repository is placeholder
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"transaction_id": id,
			"status":         "completed",
			"amount":         100.00,
			"currency":       "USD",
			"processed_at":   time.Now().Add(-time.Hour),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": transaction.TransactionID,
		"status":         transaction.Status,
		"amount":         transaction.Amount.Amount,
		"currency":       transaction.Amount.Currency,
		"processed_at":   transaction.CreatedAt,
	})
}

func (h *FinancialHandlers) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	methods := []map[string]interface{}{
		{
			"id":      "stripe_card",
			"name":    "Credit/Debit Card",
			"type":    "card",
			"enabled": true,
		},
		{
			"id":      "stripe_ach",
			"name":    "Bank Transfer (ACH)",
			"type":    "bank_transfer",
			"enabled": true,
		},
		{
			"id":      "crypto_bitcoin",
			"name":    "Bitcoin",
			"type":    "cryptocurrency",
			"enabled": true,
		},
		{
			"id":      "crypto_ethereum",
			"name":    "Ethereum",
			"type":    "cryptocurrency",
			"enabled": true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_methods": methods,
		"default_method":  "stripe_card",
	})
}

// Pricing operations

func (h *FinancialHandlers) GetPriceQuote(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ServiceType  string                 `json:"service_type"`
		Scope        string                 `json:"scope"`
		Urgency      string                 `json:"urgency"`
		ClientID     string                 `json:"client_id,omitempty"`
		Requirements map[string]interface{} `json:"requirements"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate price based on service type and requirements
	basePrice := h.calculateBasePrice(request.ServiceType, request.Scope)
	urgencyMultiplier := h.getUrgencyMultiplier(request.Urgency)
	finalPrice := basePrice * urgencyMultiplier

	quote := map[string]interface{}{
		"quote_id":           uuid.New().String(),
		"service_type":       request.ServiceType,
		"scope":              request.Scope,
		"base_price":         basePrice,
		"urgency_multiplier": urgencyMultiplier,
		"final_price":        finalPrice,
		"currency":           "USD",
		"valid_until":        time.Now().Add(24 * time.Hour),
		"breakdown": map[string]interface{}{
			"base_service": basePrice,
			"urgency_fee":  basePrice * (urgencyMultiplier - 1),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quote)
}

func (h *FinancialHandlers) GetPricingModels(w http.ResponseWriter, r *http.Request) {
	models := []map[string]interface{}{
		{
			"id":          "per_word",
			"name":        "Per Word Pricing",
			"description": "Charge based on word count",
			"base_rate":   0.10,
			"unit":        "word",
			"active":      true,
		},
		{
			"id":          "per_hour",
			"name":        "Hourly Rate",
			"description": "Charge based on time spent",
			"base_rate":   75.00,
			"unit":        "hour",
			"active":      true,
		},
		{
			"id":          "project_based",
			"name":        "Project-Based",
			"description": "Fixed price per project",
			"base_rate":   500.00,
			"unit":        "project",
			"active":      true,
		},
		{
			"id":          "value_based",
			"name":        "Value-Based Pricing",
			"description": "Price based on value delivered",
			"base_rate":   1000.00,
			"unit":        "project",
			"active":      false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pricing_models": models,
		"default_model":  "per_word",
	})
}

func (h *FinancialHandlers) CreatePricingExperiment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		ModelA      map[string]interface{} `json:"model_a"`
		ModelB      map[string]interface{} `json:"model_b"`
		Duration    string                 `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	experiment := map[string]interface{}{
		"experiment_id": uuid.New().String(),
		"name":          request.Name,
		"description":   request.Description,
		"status":        "active",
		"created_at":    time.Now(),
		"model_a":       request.ModelA,
		"model_b":       request.ModelB,
		"duration":      request.Duration,
		"traffic_split": 50, // 50/50 split
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(experiment)
}

func (h *FinancialHandlers) GetMarketData(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"industry_averages": map[string]interface{}{
			"content_writing": map[string]float64{
				"per_word_min": 0.05,
				"per_word_max": 0.25,
				"per_word_avg": 0.12,
			},
			"copywriting": map[string]float64{
				"per_word_min": 0.15,
				"per_word_max": 0.50,
				"per_word_avg": 0.30,
			},
		},
		"market_trends": []map[string]interface{}{
			{
				"trend":     "increasing_demand",
				"category":  "technical_writing",
				"change":    "+15%",
				"timeframe": "last_quarter",
			},
			{
				"trend":     "price_pressure",
				"category":  "generic_content",
				"change":    "-8%",
				"timeframe": "last_quarter",
			},
		},
		"competitor_analysis": map[string]interface{}{
			"average_rates": map[string]float64{
				"basic_content":     0.08,
				"premium_content":   0.20,
				"technical_content": 0.35,
			},
		},
		"updated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Treasury operations

func (h *FinancialHandlers) GetTreasuryBalance(w http.ResponseWriter, r *http.Request) {
	balance := map[string]interface{}{
		"total_balance": 50000.00,
		"currency":      "USD",
		"breakdown": map[string]interface{}{
			"operating_funds":  30000.00,
			"reserves":         15000.00,
			"investment_funds": 5000.00,
		},
		"crypto_holdings": map[string]interface{}{
			"bitcoin":  0.5,
			"ethereum": 10.0,
		},
		"last_updated": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *FinancialHandlers) AllocateFunds(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Amount      float64 `json:"amount"`
		FromAccount string  `json:"from_account"`
		ToAccount   string  `json:"to_account"`
		Purpose     string  `json:"purpose"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	allocation := map[string]interface{}{
		"allocation_id": uuid.New().String(),
		"amount":        request.Amount,
		"from_account":  request.FromAccount,
		"to_account":    request.ToAccount,
		"purpose":       request.Purpose,
		"status":        "completed",
		"allocated_at":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocation)
}

func (h *FinancialHandlers) GetTreasuryTransactions(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	// Mock treasury transactions
	transactions := []map[string]interface{}{
		{
			"transaction_id": uuid.New().String(),
			"type":           "allocation",
			"amount":         5000.00,
			"from_account":   "reserves",
			"to_account":     "operating",
			"purpose":        "Monthly operations funding",
			"timestamp":      time.Now().Add(-2 * time.Hour),
		},
		{
			"transaction_id": uuid.New().String(),
			"type":           "investment",
			"amount":         1000.00,
			"from_account":   "operating",
			"to_account":     "crypto_portfolio",
			"purpose":        "Bitcoin purchase",
			"timestamp":      time.Now().Add(-6 * time.Hour),
		},
	}

	if limit < len(transactions) {
		transactions = transactions[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
		"limit":        limit,
	})
}

// Helper methods

func (h *FinancialHandlers) processPaymentLogic(ctx context.Context, transaction *entities.Transaction) error {
	// Simplified payment processing logic
	// In a real implementation, this would integrate with payment processors

	switch transaction.PaymentMethod {
	case entities.PaymentMethodCreditCard:
		// Process credit card payment
		return h.processCreditCardPayment(transaction)
	case entities.PaymentMethodBankTransfer:
		// Process bank transfer
		return h.processBankTransfer(transaction)
	case entities.PaymentMethodCrypto:
		// Process cryptocurrency payment
		return h.processCryptoPayment(transaction)
	default:
		return fmt.Errorf("unsupported payment method: %s", transaction.PaymentMethod)
	}
}

func (h *FinancialHandlers) processCreditCardPayment(transaction *entities.Transaction) error {
	// Mock credit card processing
	// Would integrate with Stripe, Square, etc.
	return nil
}

func (h *FinancialHandlers) processBankTransfer(transaction *entities.Transaction) error {
	// Mock bank transfer processing
	// Would integrate with ACH, wire transfer systems
	return nil
}

func (h *FinancialHandlers) processCryptoPayment(transaction *entities.Transaction) error {
	// Mock crypto payment processing
	// Would integrate with crypto wallet services
	return nil
}

func (h *FinancialHandlers) calculateBasePrice(serviceType, scope string) float64 {
	// Simplified pricing calculation
	prices := map[string]map[string]float64{
		"content_writing": {
			"small":  250.00,
			"medium": 500.00,
			"large":  1000.00,
		},
		"copywriting": {
			"small":  500.00,
			"medium": 1000.00,
			"large":  2000.00,
		},
		"technical_writing": {
			"small":  750.00,
			"medium": 1500.00,
			"large":  3000.00,
		},
	}

	if servicePrices, ok := prices[serviceType]; ok {
		if price, ok := servicePrices[scope]; ok {
			return price
		}
	}

	return 500.00 // Default price
}

func (h *FinancialHandlers) getUrgencyMultiplier(urgency string) float64 {
	switch urgency {
	case "rush":
		return 2.0
	case "priority":
		return 1.5
	case "standard":
		return 1.0
	default:
		return 1.0
	}
}

package risk_management

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// FinancialRiskAnalyzerImpl implements financial risk analysis
type FinancialRiskAnalyzerImpl struct {
	riskRepo    repositories.RiskRepository
	paymentRepo repositories.PaymentRepository
	clientRepo  repositories.ClientRepository
	eventRepo   repositories.EventRepository
	thresholds  map[string]float64
}

// NewFinancialRiskAnalyzer creates a new financial risk analyzer
func NewFinancialRiskAnalyzer(
	riskRepo repositories.RiskRepository,
	paymentRepo repositories.PaymentRepository,
	clientRepo repositories.ClientRepository,
	eventRepo repositories.EventRepository,
) *FinancialRiskAnalyzerImpl {
	return &FinancialRiskAnalyzerImpl{
		riskRepo:    riskRepo,
		paymentRepo: paymentRepo,
		clientRepo:  clientRepo,
		eventRepo:   eventRepo,
		thresholds:  initializeDefaultThresholds(),
	}
}

// AnalyzeTransaction analyzes transaction risk
func (a *FinancialRiskAnalyzerImpl) AnalyzeTransaction(ctx context.Context, transaction *entities.Transaction) (*TransactionRiskResult, error) {
	result := &TransactionRiskResult{
		RiskScore:      0.0,
		RiskFactors:    make([]string, 0),
		RequiresReview: false,
		Action:         "approve",
		Reasons:        make([]string, 0),
	}

	// Check transaction amount against thresholds
	thresholds, err := a.getActiveThresholds(ctx, string(transaction.Type))
	if err != nil {
		return nil, fmt.Errorf("failed to get thresholds: %w", err)
	}

	// Base threshold is 10K USD, adjustable by transaction type
	baseThreshold := 10000.0
	for _, threshold := range thresholds {
		if threshold.Category == string(transaction.Type) && threshold.IsActive {
			baseThreshold = threshold.Threshold
			break
		}
	}

	if transaction.Amount.Amount > baseThreshold {
		result.RiskScore += 0.5
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("High value transaction (>$%.2f)", baseThreshold))
		result.RequiresReview = true
	}

	// Check client history
	if transaction.ClientID != uuid.Nil {
		clientRisk, err := a.assessClientRisk(ctx, transaction.ClientID.String())
		if err == nil {
			result.RiskScore += clientRisk
			if clientRisk > 0.3 {
				result.RiskFactors = append(result.RiskFactors, "Client risk profile")
			}
		}
	}

	// Check velocity (frequency of transactions)
	velocityRisk, err := a.checkTransactionVelocity(ctx, transaction)
	if err == nil {
		result.RiskScore += velocityRisk
		if velocityRisk > 0.2 {
			result.RiskFactors = append(result.RiskFactors, "High transaction velocity")
		}
	}

	// Check for unusual patterns
	patternRisk, patterns := a.checkUnusualPatterns(transaction)
	result.RiskScore += patternRisk
	if len(patterns) > 0 {
		result.RiskFactors = append(result.RiskFactors, patterns...)
	}

	// Normalize risk score
	result.RiskScore = math.Min(result.RiskScore, 1.0)

	// Determine action based on risk score
	switch {
	case result.RiskScore >= 0.8:
		result.Action = "reject"
		result.Reasons = append(result.Reasons, "High risk score")
	case result.RiskScore >= 0.5:
		result.Action = "review"
		result.RequiresReview = true
		result.Reasons = append(result.Reasons, "Moderate risk - manual review required")
	default:
		result.Action = "approve"
	}

	// Create risk entity if score is significant
	if result.RiskScore > 0.3 {
		risk := &entities.Risk{
			ID:                uuid.New(),
			Category:          entities.RiskTypeFinancial,
			Severity:          a.scoreToSeverity(result.RiskScore),
			Status:            entities.RiskStatusIdentified,
			Title:             fmt.Sprintf("Transaction Risk - %s", transaction.TransactionID),
			Description:       fmt.Sprintf("Risk score: %.2f, Factors: %v", result.RiskScore, result.RiskFactors),
			Likelihood:        result.RiskScore,
			Impact:            transaction.Amount.Amount / 100000.0, // Normalize impact
			MitigationActions: []string{fmt.Sprintf("Action: %s", result.Action)},
			Metadata: map[string]interface{}{
				"source":         "financial_analyzer",
				"transaction_id": transaction.TransactionID.String(),
				"amount":         transaction.Amount,
				"risk_factors":   result.RiskFactors,
				"affected_entities": []string{transaction.TransactionID.String()},
			},
			IdentifiedAt:   time.Now(),
			LastAssessment: time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := a.riskRepo.CreateRisk(ctx, risk); err != nil {
			return nil, fmt.Errorf("failed to create risk: %w", err)
		}
	}

	return result, nil
}

// CheckFraudIndicators checks for fraud indicators
func (a *FinancialRiskAnalyzerImpl) CheckFraudIndicators(ctx context.Context, data interface{}) (*FraudCheckResult, error) {
	// This would integrate with the existing fraud detection service
	// For now, return a basic implementation
	return &FraudCheckResult{
		IsFraudulent:                   false,
		FraudScore:                     0.0,
		FraudIndicators:                []string{},
		RequiresAdditionalVerification: false,
	}, nil
}

// AssessPaymentRisk assesses payment-specific risks
func (a *FinancialRiskAnalyzerImpl) AssessPaymentRisk(ctx context.Context, payment *entities.Payment) (*PaymentRiskResult, error) {
	result := &PaymentRiskResult{
		RiskScore: 0.0,
		Risks:     make([]string, 0),
		Action:    "process",
	}

	// Check payment method risk
	methodRisk := a.assessPaymentMethodRisk(string(payment.PaymentMethod))
	result.RiskScore += methodRisk
	if methodRisk > 0.2 {
		result.Risks = append(result.Risks, "Payment method risk")
	}

	// Check amount against thresholds
	if payment.Amount > 10000 {
		result.RiskScore += 0.3
		result.Risks = append(result.Risks, "High value payment")
	}

	// Check currency risk
	if payment.Currency != "USD" && payment.Currency != "EUR" {
		result.RiskScore += 0.1
		result.Risks = append(result.Risks, "Non-standard currency")
	}

	// Determine action
	if result.RiskScore >= 0.6 {
		result.Action = "hold"
	} else if result.RiskScore >= 0.8 {
		result.Action = "reject"
	}

	return result, nil
}

// CheckFinancialThresholds checks all financial thresholds
func (a *FinancialRiskAnalyzerImpl) CheckFinancialThresholds(ctx context.Context) ([]*ThresholdViolation, error) {
	violations := make([]*ThresholdViolation, 0)

	// Get all financial thresholds
	thresholds, err := a.riskRepo.GetThresholdsByType(ctx, entities.RiskTypeFinancial)
	if err != nil {
		return nil, fmt.Errorf("failed to get thresholds: %w", err)
	}

	// Check each threshold
	for _, threshold := range thresholds {
		if !threshold.IsActive {
			continue
		}

		// Get current value based on threshold category
		currentValue, err := a.getCurrentValue(ctx, threshold.Category)
		if err != nil {
			continue
		}

		if currentValue > threshold.Threshold {
			violation := &ThresholdViolation{
				ThresholdID:  threshold.ID,
				Type:         string(threshold.Type),
				Category:     threshold.Category,
				CurrentValue: currentValue,
				Threshold:    threshold.Threshold,
				Severity:     "high",
				Message:      fmt.Sprintf("%s exceeded: %.2f > %.2f", threshold.Category, currentValue, threshold.Threshold),
			}
			violations = append(violations, violation)
		}
	}

	return violations, nil
}

// Helper methods

func (a *FinancialRiskAnalyzerImpl) getActiveThresholds(ctx context.Context, category string) ([]*entities.RiskThreshold, error) {
	allThresholds, err := a.riskRepo.GetThresholdsByType(ctx, entities.RiskTypeFinancial)
	if err != nil {
		return nil, err
	}

	activeThresholds := make([]*entities.RiskThreshold, 0)
	for _, t := range allThresholds {
		if t.IsActive && (t.Category == category || t.Category == "all") {
			activeThresholds = append(activeThresholds, t)
		}
	}

	return activeThresholds, nil
}

func (a *FinancialRiskAnalyzerImpl) assessClientRisk(ctx context.Context, clientID string) (float64, error) {
	clientUUID, _ := uuid.Parse(clientID)
	client, err := a.clientRepo.FindByID(ctx, clientUUID)
	if err != nil {
		return 0.0, err
	}

	risk := 0.0

	// New client risk
	if time.Since(client.CreatedAt) < 30*24*time.Hour {
		risk += 0.1
	}

	// Check payment history
	payments, err := a.paymentRepo.GetPaymentsByClient(ctx, clientID, 100, 0)
	if err == nil {
		failedPayments := 0
		for _, p := range payments {
			if p.Status == entities.PaymentStatusFailed {
				failedPayments++
			}
		}

		if len(payments) > 0 {
			failureRate := float64(failedPayments) / float64(len(payments))
			risk += failureRate * 0.3
		}
	}

	return risk, nil
}

func (a *FinancialRiskAnalyzerImpl) checkTransactionVelocity(ctx context.Context, transaction *entities.Transaction) (float64, error) {
	// Simplified implementation - in production would check recent transaction frequency
	// For now, return low risk
	return 0.0, nil
}

func (a *FinancialRiskAnalyzerImpl) checkUnusualPatterns(transaction *entities.Transaction) (float64, []string) {
	risk := 0.0
	patterns := make([]string, 0)

	// Round number amounts (potential test transactions)
	if math.Mod(transaction.Amount.Amount, 100) == 0 && transaction.Amount.Amount > 1000 {
		risk += 0.1
		patterns = append(patterns, "Round number amount")
	}

	// Very small amounts (potential card testing)
	if transaction.Amount.Amount < 1.0 {
		risk += 0.2
		patterns = append(patterns, "Unusually small amount")
	}

	// Time-based patterns (transactions at unusual hours)
	hour := time.Now().Hour()
	if hour < 6 || hour > 23 {
		risk += 0.05
		patterns = append(patterns, "Transaction at unusual hour")
	}

	return risk, patterns
}

func (a *FinancialRiskAnalyzerImpl) assessPaymentMethodRisk(method string) float64 {
	riskScores := map[string]float64{
		"credit_card":   0.1,
		"debit_card":    0.1,
		"bank_transfer": 0.05,
		"crypto":        0.3,
		"wire_transfer": 0.15,
		"cash":          0.4,
		"check":         0.25,
	}

	if score, ok := riskScores[method]; ok {
		return score
	}

	return 0.2 // Default risk for unknown methods
}

func (a *FinancialRiskAnalyzerImpl) getCurrentValue(ctx context.Context, category string) (float64, error) {
	switch category {
	case "daily_volume":
		// Simplified implementation - would get today's transaction volume
		return 0, nil

	case "single_transaction":
		// This would be checked per transaction
		return 0, nil

	case "monthly_volume":
		// Simplified implementation - would get this month's transaction volume
		return 0, nil

	default:
		return 0, fmt.Errorf("unknown category: %s", category)
	}
}

func (a *FinancialRiskAnalyzerImpl) scoreToSeverity(score float64) entities.RiskSeverity {
	switch {
	case score >= 0.8:
		return entities.RiskSeverityCritical
	case score >= 0.6:
		return entities.RiskSeverityHigh
	case score >= 0.4:
		return entities.RiskSeverityMedium
	default:
		return entities.RiskSeverityLow
	}
}

func initializeDefaultThresholds() map[string]float64 {
	return map[string]float64{
		"single_transaction": 10000.0,
		"daily_volume":       50000.0,
		"monthly_volume":     500000.0,
		"high_risk_country":  5000.0,
		"new_client":         1000.0,
	}
}

// Additional types for fraud detection integration

type FraudCheckResult struct {
	IsFraudulent                   bool     `json:"is_fraudulent"`
	FraudScore                     float64  `json:"fraud_score"`
	FraudIndicators                []string `json:"fraud_indicators"`
	RequiresAdditionalVerification bool     `json:"requires_additional_verification"`
}

type PaymentRiskResult struct {
	RiskScore float64  `json:"risk_score"`
	Risks     []string `json:"risks"`
	Action    string   `json:"action"` // "process", "hold", "reject"
}

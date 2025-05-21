package payment

import (
	"context"
	"fmt"
	"math"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
)

// FraudDetectionImpl implements fraud detection for payments
type FraudDetectionImpl struct {
	paymentRepo    repositories.PaymentRepository
	config         *FraudConfig
	blacklistedIPs map[string]string // IP -> reason
	suspiciousUAs  []string          // Suspicious user agent patterns
}

// FraudConfig holds fraud detection configuration
type FraudConfig struct {
	MaxVelocityCount          int           // Max payments per time window
	VelocityWindow            time.Duration // Time window for velocity check
	UnusualAmountFactor       float64       // Factor for unusual amount detection
	MaxAmountThreshold        int64         // Max allowed amount
	MinAmountThreshold        int64         // Min allowed amount
	SuspiciousCountries       []string      // List of high-risk countries
	RequireVerificationAmount int64         // Amount above which verification is required
	BlockScore                float64       // Score above which to block
	ReviewScore               float64       // Score above which to review
	IPGeolocationAPI          string        // IP geolocation service URL
	DeviceFingerprintRequired bool          // Whether device fingerprint is required
}

// NewFraudDetectionService creates a new fraud detection service
func NewFraudDetectionService(
	paymentRepo repositories.PaymentRepository,
	config *FraudConfig,
) FraudDetectionService {
	return &FraudDetectionImpl{
		paymentRepo: paymentRepo,
		config:      config,
		blacklistedIPs: map[string]string{
			// Preloaded blacklisted IPs
			"10.0.0.1": "known_fraudster",
		},
		suspiciousUAs: []string{
			"bot", "crawler", "scraper", "automated", "python-requests",
			"curl", "wget", "postman", "insomnia",
		},
	}
}

// AnalyzePayment performs comprehensive fraud analysis on a payment
func (f *FraudDetectionImpl) AnalyzePayment(ctx context.Context, payment *entities.Payment) (*entities.FraudDetectionResult, error) {
	var flags []entities.FraudFlag
	var score float64
	var explanations []string

	// 1. Velocity check
	if hasVelocityFlag, s, explanation := f.checkVelocity(ctx, payment); hasVelocityFlag {
		flags = append(flags, entities.FraudFlagHighVelocity)
		score += s
		explanations = append(explanations, explanation)
	}

	// 2. Amount analysis
	if hasAmountFlag, s, explanation := f.checkUnusualAmount(ctx, payment); hasAmountFlag {
		flags = append(flags, entities.FraudFlagUnusualAmount)
		score += s
		explanations = append(explanations, explanation)
	}

	// 3. IP address analysis
	if hasIPFlag, s, explanation := f.checkSuspiciousIP(ctx, payment); hasIPFlag {
		flags = append(flags, entities.FraudFlagBlacklistedIP)
		score += s
		explanations = append(explanations, explanation)
	}

	// 4. Geographic location analysis
	if hasLocationFlag, s, explanation := f.checkSuspiciousLocation(ctx, payment); hasLocationFlag {
		flags = append(flags, entities.FraudFlagSuspiciousLocation)
		score += s
		explanations = append(explanations, explanation)
	}

	// 5. User agent analysis
	if hasUAFlag, s, explanation := f.checkSuspiciousUserAgent(ctx, payment); hasUAFlag {
		flags = append(flags, entities.FraudFlagSuspiciousUserAgent)
		score += s
		explanations = append(explanations, explanation)
	}

	// 6. Payment method analysis
	if hasMethodFlag, s, explanation := f.checkNewPaymentMethod(ctx, payment); hasMethodFlag {
		flags = append(flags, entities.FraudFlagNewPaymentMethod)
		score += s
		explanations = append(explanations, explanation)
	}

	// 7. Time pattern analysis
	if hasTimeFlag, s, explanation := f.checkTimePatterns(ctx, payment); hasTimeFlag {
		flags = append(flags, entities.FraudFlagTimePatterns)
		score += s
		explanations = append(explanations, explanation)
	}

	// 8. Failed verification check
	if hasVerificationFlag, s, explanation := f.checkFailedVerification(ctx, payment); hasVerificationFlag {
		flags = append(flags, entities.FraudFlagFailedVerification)
		score += s
		explanations = append(explanations, explanation)
	}

	// Normalize score (0.0 to 1.0)
	score = math.Min(score, 1.0)

	// Determine risk level
	var riskLevel entities.FraudRiskLevel
	switch {
	case score >= 0.8:
		riskLevel = entities.FraudRiskLevelCritical
	case score >= 0.6:
		riskLevel = entities.FraudRiskLevelHigh
	case score >= 0.3:
		riskLevel = entities.FraudRiskLevelMedium
	default:
		riskLevel = entities.FraudRiskLevelLow
	}

	// Determine action
	var action entities.FraudAction
	switch {
	case score >= f.config.BlockScore:
		action = entities.FraudActionBlock
	case score >= f.config.ReviewScore:
		action = entities.FraudActionReview
	case payment.Amount >= f.config.RequireVerificationAmount:
		action = entities.FraudActionRequireVerification
	default:
		action = entities.FraudActionAllow
	}

	// Calculate confidence
	confidence := f.calculateConfidence(score, len(flags))

	result := &entities.FraudDetectionResult{
		PaymentID:   payment.ID,
		FraudScore:  score,
		RiskLevel:   riskLevel,
		Flags:       flags,
		Explanation: strings.Join(explanations, "; "),
		Action:      action,
		Confidence:  confidence,
		ProcessedAt: time.Now(),
	}

	return result, nil
}

// UpdateWhitelist adds a client to the whitelist
func (f *FraudDetectionImpl) UpdateWhitelist(ctx context.Context, clientID string) error {
	// In a real implementation, this would update a whitelist database
	fmt.Printf("Adding client %s to whitelist\n", clientID)
	return nil
}

// UpdateBlacklist adds an IP to the blacklist
func (f *FraudDetectionImpl) UpdateBlacklist(ctx context.Context, ip string, reason string) error {
	f.blacklistedIPs[ip] = reason
	fmt.Printf("Added IP %s to blacklist with reason: %s\n", ip, reason)
	return nil
}

// GetRiskProfile retrieves the risk profile for a client
func (f *FraudDetectionImpl) GetRiskProfile(ctx context.Context, clientID string) (*RiskProfile, error) {
	// Get client's payment history
	payments, err := f.paymentRepo.GetPaymentsByClient(ctx, clientID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get client payments: %w", err)
	}

	var successfulPayments, failedPayments int64
	var totalAmount int64
	var lastPaymentAt *time.Time
	var preferredMethods []entities.PaymentMethod
	var locations []string
	var devices []string

	methodCounts := make(map[entities.PaymentMethod]int)
	locationCounts := make(map[string]int)

	for _, payment := range payments {
		if payment.IsCompleted() {
			successfulPayments++
			totalAmount += payment.NetAmount
		} else if payment.IsFailed() {
			failedPayments++
		}

		// Track last payment
		if lastPaymentAt == nil || payment.CreatedAt.After(*lastPaymentAt) {
			lastPaymentAt = &payment.CreatedAt
		}

		// Track payment methods
		methodCounts[payment.PaymentMethod]++

		// Track locations (from IP geolocation)
		if payment.IPAddress != nil {
			location := f.getLocationFromIP(*payment.IPAddress)
			if location != "" {
				locationCounts[location]++
			}
		}

		// Track device fingerprints (from user agent analysis)
		if payment.UserAgent != nil {
			device := f.extractDeviceFingerprint(*payment.UserAgent)
			if device != "" {
				devices = append(devices, device)
			}
		}
	}

	// Determine preferred payment methods (top 3)
	for method, count := range methodCounts {
		if count >= 2 { // Used at least twice
			preferredMethods = append(preferredMethods, method)
		}
	}

	// Determine frequent locations
	for location, count := range locationCounts {
		if count >= 2 { // Used at least twice
			locations = append(locations, location)
		}
	}

	// Calculate risk score
	var riskScore float64
	if successfulPayments+failedPayments > 0 {
		failureRate := float64(failedPayments) / float64(successfulPayments+failedPayments)
		riskScore = failureRate * 0.5 // Max 0.5 from failure rate

		// Add location-based risk
		for _, location := range locations {
			if f.isSuspiciousLocation(location) {
				riskScore += 0.2
			}
		}

		// Add device-based risk
		if len(devices) > 5 { // Too many different devices
			riskScore += 0.1
		}
	}

	riskScore = math.Min(riskScore, 1.0)

	var averageAmount float64
	if successfulPayments > 0 {
		averageAmount = float64(totalAmount) / float64(successfulPayments)
	}

	// Convert payment history to string slice (simplified)
	var paymentHistory []string
	for _, payment := range payments {
		paymentHistory = append(paymentHistory, payment.ID)
	}

	profile := &RiskProfile{
		ClientID:            clientID,
		RiskScore:           riskScore,
		PaymentHistory:      paymentHistory[:min(len(paymentHistory), 20)], // Last 20 payments
		SuccessfulPayments:  successfulPayments,
		FailedPayments:      failedPayments,
		LastPaymentAt:       lastPaymentAt,
		AverageAmount:       averageAmount,
		PreferredMethods:    preferredMethods,
		GeographicLocations: locations,
		DeviceFingerprints:  devices[:min(len(devices), 10)], // Last 10 devices
	}

	return profile, nil
}

// Fraud check implementations

func (f *FraudDetectionImpl) checkVelocity(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	// Get recent payments for this client
	//since := time.Now().Add(-f.config.VelocityWindow)

	// This would query payments in the time window
	// For demonstration, we'll simulate
	recentPaymentCount := 3 // Simulated count

	if recentPaymentCount >= f.config.MaxVelocityCount {
		return true, 0.4, fmt.Sprintf("High velocity: %d payments in %v", recentPaymentCount, f.config.VelocityWindow)
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkUnusualAmount(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	// Get client's historical payments to calculate average
	payments, err := f.paymentRepo.GetPaymentsByClient(ctx, payment.ClientID, 50, 0)
	if err != nil || len(payments) == 0 {
		// No history, consider large amounts suspicious
		if payment.Amount > f.config.MaxAmountThreshold {
			return true, 0.5, fmt.Sprintf("Large amount with no payment history: %d", payment.Amount)
		}
		return false, 0.0, ""
	}

	// Calculate average amount
	var totalAmount int64
	var validPayments int64
	for _, p := range payments {
		if p.IsCompleted() {
			totalAmount += p.Amount
			validPayments++
		}
	}

	if validPayments == 0 {
		return false, 0.0, ""
	}

	averageAmount := float64(totalAmount) / float64(validPayments)
	currentAmount := float64(payment.Amount)

	// Check if current amount is significantly different from average
	if currentAmount > averageAmount*f.config.UnusualAmountFactor ||
		currentAmount < averageAmount/f.config.UnusualAmountFactor {
		deviation := math.Abs(currentAmount-averageAmount) / averageAmount
		score := math.Min(deviation/5.0, 0.3) // Max 0.3 from amount
		return true, score, fmt.Sprintf("Unusual amount: %.2f vs average %.2f", currentAmount, averageAmount)
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkSuspiciousIP(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	if payment.IPAddress == nil {
		return false, 0.0, ""
	}

	ip := *payment.IPAddress

	// Check blacklist
	if reason, exists := f.blacklistedIPs[ip]; exists {
		return true, 0.8, fmt.Sprintf("Blacklisted IP: %s (%s)", ip, reason)
	}

	// Check if it's a known proxy/VPN
	if f.isProxyIP(ip) {
		return true, 0.4, fmt.Sprintf("Proxy/VPN IP detected: %s", ip)
	}

	// Check for private/invalid IPs
	if f.isPrivateIP(ip) {
		return true, 0.3, fmt.Sprintf("Private IP address: %s", ip)
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkSuspiciousLocation(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	if payment.IPAddress == nil {
		return false, 0.0, ""
	}

	location := f.getLocationFromIP(*payment.IPAddress)
	if location == "" {
		return false, 0.0, ""
	}

	// Check if location is in suspicious countries list
	for _, suspiciousCountry := range f.config.SuspiciousCountries {
		if strings.Contains(strings.ToLower(location), strings.ToLower(suspiciousCountry)) {
			return true, 0.3, fmt.Sprintf("Payment from high-risk location: %s", location)
		}
	}

	// Check if location is different from client's usual locations
	profile, err := f.GetRiskProfile(ctx, payment.ClientID)
	if err == nil && len(profile.GeographicLocations) > 0 {
		isKnownLocation := false
		for _, knownLocation := range profile.GeographicLocations {
			if strings.Contains(strings.ToLower(location), strings.ToLower(knownLocation)) {
				isKnownLocation = true
				break
			}
		}

		if !isKnownLocation {
			return true, 0.2, fmt.Sprintf("Payment from unusual location: %s", location)
		}
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkSuspiciousUserAgent(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	if payment.UserAgent == nil {
		return true, 0.1, "Missing user agent"
	}

	userAgent := strings.ToLower(*payment.UserAgent)

	// Check for suspicious patterns
	for _, pattern := range f.suspiciousUAs {
		if strings.Contains(userAgent, pattern) {
			return true, 0.3, fmt.Sprintf("Suspicious user agent: contains '%s'", pattern)
		}
	}

	// Check for very short or very long user agents
	if len(*payment.UserAgent) < 10 {
		return true, 0.2, "Unusually short user agent"
	}
	if len(*payment.UserAgent) > 1000 {
		return true, 0.2, "Unusually long user agent"
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkNewPaymentMethod(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	// Get client's payment method history
	payments, err := f.paymentRepo.GetPaymentsByClient(ctx, payment.ClientID, 20, 0)
	if err != nil {
		return false, 0.0, ""
	}

	// Check if this payment method has been used before
	for _, p := range payments {
		if p.PaymentMethod == payment.PaymentMethod && p.IsCompleted() {
			return false, 0.0, "" // Known payment method
		}
	}

	// New payment method with existing client history
	if len(payments) > 0 {
		return true, 0.15, fmt.Sprintf("First time using payment method: %s", payment.PaymentMethod)
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkTimePatterns(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	hour := payment.CreatedAt.Hour()

	// Flag payments during unusual hours (2 AM - 6 AM)
	if hour >= 2 && hour <= 6 {
		return true, 0.1, fmt.Sprintf("Payment during unusual hours: %02d:00", hour)
	}

	// Check for weekend large payments
	if (payment.CreatedAt.Weekday() == time.Saturday || payment.CreatedAt.Weekday() == time.Sunday) &&
		payment.Amount > f.config.MaxAmountThreshold/2 {
		return true, 0.1, "Large payment during weekend"
	}

	return false, 0.0, ""
}

func (f *FraudDetectionImpl) checkFailedVerification(ctx context.Context, payment *entities.Payment) (bool, float64, string) {
	// This would check if payment verification failed
	// For demonstration, we'll check certain criteria

	// Check if crypto payment has invalid wallet address
	if payment.IsCryptocurrency() && payment.WalletAddress != nil {
		if !f.isValidCryptoAddress(*payment.WalletAddress) {
			return true, 0.6, "Invalid cryptocurrency wallet address"
		}
	}

	// Check if required fields are missing
	if payment.IPAddress == nil && payment.Amount > f.config.RequireVerificationAmount {
		return true, 0.3, "Missing IP address for high-value payment"
	}

	return false, 0.0, ""
}

// Helper methods

func (f *FraudDetectionImpl) calculateConfidence(score float64, flagCount int) float64 {
	// Base confidence on score
	confidence := score

	// Increase confidence with more flags
	confidence += float64(flagCount) * 0.05

	// Cap at 1.0
	return math.Min(confidence, 1.0)
}

func (f *FraudDetectionImpl) isProxyIP(ip string) bool {
	// Simplified proxy detection
	// In reality, this would use a proxy detection service
	proxyPatterns := []string{"proxy", "vpn", "tor"}

	for _, pattern := range proxyPatterns {
		if strings.Contains(strings.ToLower(ip), pattern) {
			return true
		}
	}

	return false
}

func (f *FraudDetectionImpl) isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // Invalid IP
	}

	return parsedIP.IsPrivate() || parsedIP.IsLoopback()
}

func (f *FraudDetectionImpl) getLocationFromIP(ip string) string {
	// Simplified geolocation
	// In reality, this would use a geolocation service like MaxMind or IPStack
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "Private Network"
	}

	return "Unknown Location" // Placeholder
}

func (f *FraudDetectionImpl) isSuspiciousLocation(location string) bool {
	for _, suspicious := range f.config.SuspiciousCountries {
		if strings.Contains(strings.ToLower(location), strings.ToLower(suspicious)) {
			return true
		}
	}
	return false
}

func (f *FraudDetectionImpl) extractDeviceFingerprint(userAgent string) string {
	// Simplified device fingerprinting based on user agent
	// In reality, this would be much more sophisticated

	// Extract browser and OS info
	var browser, os string

	if strings.Contains(userAgent, "Chrome") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "Safari") {
		browser = "Safari"
	}

	if strings.Contains(userAgent, "Windows") {
		os = "Windows"
	} else if strings.Contains(userAgent, "Mac") {
		os = "Mac"
	} else if strings.Contains(userAgent, "Linux") {
		os = "Linux"
	}

	return fmt.Sprintf("%s_%s", browser, os)
}

func (f *FraudDetectionImpl) isValidCryptoAddress(address string) bool {
	// Basic Ethereum address validation
	if strings.HasPrefix(address, "0x") && len(address) == 42 {
		// Check if all characters after 0x are hex
		hexPart := address[2:]
		matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart)
		return matched
	}

	// Add validation for other crypto address formats as needed
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

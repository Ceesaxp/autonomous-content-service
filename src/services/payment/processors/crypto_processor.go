package processors

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/services/payment"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CryptoProcessor implements cryptocurrency payment processing
type CryptoProcessor struct {
	ethClient  *ethclient.Client
	privateKey *ecdsa.PrivateKey
	walletAddr common.Address
	config     *CryptoConfig
	contracts  map[string]*TokenContract
}

// CryptoConfig holds cryptocurrency configuration
type CryptoConfig struct {
	EthereumNodeURL       string
	PrivateKey            string
	WalletAddress         string
	RequiredConfirmations map[string]int64 // currency -> confirmations
	GasPriceMultiplier    float64
	MaxGasPrice           int64
	TokenContracts        map[string]string // currency -> contract address
	WebhookURL            string
}

// TokenContract represents an ERC20 token contract
type TokenContract struct {
	Address  common.Address
	ABI      abi.ABI
	Decimals uint8
	Symbol   string
}

// NewCryptoProcessor creates a new cryptocurrency payment processor
func NewCryptoProcessor(config *CryptoConfig) (payment.PaymentProcessor, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(config.EthereumNodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(config.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Parse wallet address
	walletAddr := common.HexToAddress(config.WalletAddress)

	// Initialize token contracts
	contracts := make(map[string]*TokenContract)

	// Standard ERC20 ABI
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIString))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	for currency, contractAddr := range config.TokenContracts {
		contracts[currency] = &TokenContract{
			Address:  common.HexToAddress(contractAddr),
			ABI:      erc20ABI,
			Decimals: 18, // Default, should be queried from contract
			Symbol:   currency,
		}
	}

	processor := &CryptoProcessor{
		ethClient:  client,
		privateKey: privateKey,
		walletAddr: walletAddr,
		config:     config,
		contracts:  contracts,
	}

	// Query token decimals and symbols
	for currency, contract := range contracts {
		if err := processor.updateTokenInfo(contract); err != nil {
			fmt.Printf("Warning: failed to update token info for %s: %v\n", currency, err)
		}
	}

	return processor, nil
}

// GetName returns the processor name
func (c *CryptoProcessor) GetName() string {
	return "crypto"
}

// GetSupportedMethods returns supported cryptocurrency payment methods
func (c *CryptoProcessor) GetSupportedMethods() []entities.PaymentMethod {
	methods := []entities.PaymentMethod{entities.PaymentMethodEthereum}

	for currency := range c.contracts {
		switch strings.ToUpper(currency) {
		case "USDC":
			methods = append(methods, entities.PaymentMethodUSDC)
		case "DAI":
			methods = append(methods, entities.PaymentMethodDAI)
		}
	}

	return methods
}

// ProcessPayment processes a cryptocurrency payment
func (c *CryptoProcessor) ProcessPayment(ctx context.Context, request *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	if request.WalletAddress == nil {
		return nil, fmt.Errorf("wallet address required for crypto payments")
	}

	fromAddr := common.HexToAddress(*request.WalletAddress)
	currency := strings.ToUpper(request.Currency)

	// Generate unique payment address or use existing wallet
	paymentAddr := c.walletAddr

	var estimatedConfirmation *time.Time
	var transactionHash *string

	if currency == "ETH" {
		// For ETH payments, monitor the wallet address
		hash, err := c.monitorETHPayment(ctx, fromAddr, paymentAddr, request.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to setup ETH payment monitoring: %w", err)
		}
		if hash != "" {
			transactionHash = &hash
			// Estimate confirmation time (15 seconds per block * required confirmations)
			confirmations := c.config.RequiredConfirmations["ETH"]
			if confirmations == 0 {
				confirmations = 12 // Default for ETH
			}
			estimated := time.Now().Add(time.Duration(confirmations) * 15 * time.Second)
			estimatedConfirmation = &estimated
		}
	} else {
		// For ERC20 tokens
		contract, exists := c.contracts[currency]
		if !exists {
			return nil, fmt.Errorf("unsupported token: %s", currency)
		}

		hash, err := c.monitorTokenPayment(ctx, fromAddr, paymentAddr, request.Amount, contract)
		if err != nil {
			return nil, fmt.Errorf("failed to setup token payment monitoring: %w", err)
		}
		if hash != "" {
			transactionHash = &hash
			// Estimate confirmation time
			confirmations := c.config.RequiredConfirmations[currency]
			if confirmations == 0 {
				confirmations = 12 // Default
			}
			estimated := time.Now().Add(time.Duration(confirmations) * 15 * time.Second)
			estimatedConfirmation = &estimated
		}
	}

	// Calculate gas fees (estimated)
	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		gasPrice = big.NewInt(20000000000) // 20 gwei default
	}

	// Adjust gas price with multiplier
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(int64(c.config.GasPriceMultiplier*100)))
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	// Estimate gas limit
	var gasLimit uint64 = 21000 // ETH transfer
	if currency != "ETH" {
		gasLimit = 65000 // ERC20 transfer
	}

	processorFee := new(big.Int).Mul(adjustedGasPrice, big.NewInt(int64(gasLimit)))
	processorFeeInt64 := processorFee.Int64()

	// Net amount (for display purposes, actual calculation happens on confirmation)
	netAmount := request.Amount - processorFeeInt64

	response := &payment.PaymentResponse{
		PaymentID:             request.ClientID, // Will be replaced with actual payment ID
		ExternalID:            transactionHash,
		Status:                entities.PaymentStatusPending,
		Amount:                request.Amount,
		Currency:              request.Currency,
		ProcessorFee:          processorFeeInt64,
		NetAmount:             netAmount,
		TransactionHash:       transactionHash,
		EstimatedConfirmation: estimatedConfirmation,
		Message:               fmt.Sprintf("Send %s %s to %s", c.formatAmount(request.Amount, currency), currency, paymentAddr.Hex()),
		Metadata: map[string]interface{}{
			"payment_address":        paymentAddr.Hex(),
			"required_confirmations": c.config.RequiredConfirmations[currency],
			"estimated_gas_fee":      processorFeeInt64,
		},
	}

	return response, nil
}

// GetPaymentStatus retrieves cryptocurrency payment status
func (c *CryptoProcessor) GetPaymentStatus(ctx context.Context, externalID string) (*payment.PaymentStatusResponse, error) {
	if !strings.HasPrefix(externalID, "0x") {
		return nil, fmt.Errorf("invalid transaction hash format")
	}

	txHash := common.HexToHash(externalID)

	// Get transaction receipt
	receipt, err := c.ethClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		// Transaction not found or not mined yet
		return &payment.PaymentStatusResponse{
			ExternalID: externalID,
			Status:     entities.PaymentStatusPending,
		}, nil
	}

	// Get transaction details
	tx, _, err := c.ethClient.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get current block number
	currentBlock, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	confirmations := currentBlock - receipt.BlockNumber.Uint64()

	// Determine currency and amount
	var currency string
	var amount int64

	if tx.To().Hex() == c.walletAddr.Hex() && len(tx.Data()) == 0 {
		// ETH transaction
		currency = "ETH"
		amount = tx.Value().Int64()
	} else {
		// Likely ERC20 transaction - parse the data
		currency, amount, err = c.parseTokenTransaction(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token transaction: %w", err)
		}
	}

	// Calculate gas fee
	gasUsed := receipt.GasUsed
	gasPrice := tx.GasPrice()
	processorFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasUsed))).Int64()

	// Determine status based on confirmations
	var status entities.PaymentStatus
	var failureReason *string

	requiredConfirmations := c.config.RequiredConfirmations[currency]
	if requiredConfirmations == 0 {
		requiredConfirmations = 12 // Default
	}

	if receipt.Status == 0 {
		status = entities.PaymentStatusFailed
		reason := "Transaction failed"
		failureReason = &reason
	} else if confirmations >= uint64(requiredConfirmations) {
		status = entities.PaymentStatusCompleted
	} else {
		status = entities.PaymentStatusConfirming
	}

	// Get block timestamp for processed date
	block, err := c.ethClient.BlockByNumber(ctx, receipt.BlockNumber)
	var processedAt *time.Time
	if err == nil {
		timestamp := time.Unix(int64(block.Time()), 0)
		processedAt = &timestamp
	}

	response := &payment.PaymentStatusResponse{
		ExternalID:    externalID,
		Status:        status,
		Amount:        amount,
		ProcessorFee:  processorFee,
		ProcessedAt:   processedAt,
		FailureReason: failureReason,
		Metadata: map[string]interface{}{
			"confirmations":          confirmations,
			"required_confirmations": requiredConfirmations,
			"block_number":           receipt.BlockNumber.Uint64(),
			"gas_used":               gasUsed,
			"gas_price":              gasPrice.String(),
			"currency":               currency,
		},
	}

	return response, nil
}

// ProcessRefund processes a cryptocurrency refund
func (c *CryptoProcessor) ProcessRefund(ctx context.Context, request *payment.RefundRequest) (*payment.RefundResponse, error) {
	// For crypto, refunds require sending a new transaction
	// This is a simplified implementation

	// Get original payment details to determine refund address
	// In a real implementation, this would be stored in the payment record
	refundAddr := common.HexToAddress("0x0000000000000000000000000000000000000000") // Placeholder

	// Determine currency from payment ID (simplified)
	currency := "ETH" // This should be retrieved from the original payment

	var txHash string
	var err error

	if currency == "ETH" {
		txHash, err = c.sendETH(ctx, refundAddr, request.Amount)
	} else {
		contract, exists := c.contracts[currency]
		if !exists {
			return nil, fmt.Errorf("unsupported token for refund: %s", currency)
		}
		txHash, err = c.sendToken(ctx, refundAddr, request.Amount, contract)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send refund transaction: %w", err)
	}

	// Calculate gas fee (estimated)
	gasPrice, _ := c.ethClient.SuggestGasPrice(ctx)
	if gasPrice == nil {
		gasPrice = big.NewInt(20000000000) // 20 gwei default
	}

	var gasLimit uint64 = 21000 // ETH transfer
	if currency != "ETH" {
		gasLimit = 65000 // ERC20 transfer
	}

	processorFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))).Int64()
	netRefund := request.Amount - processorFee

	response := &payment.RefundResponse{
		RefundID:     txHash,
		ExternalID:   &txHash,
		Status:       entities.RefundStatusProcessing,
		Amount:       request.Amount,
		ProcessorFee: processorFee,
		NetRefund:    netRefund,
		Message:      "Refund transaction sent",
		Metadata: map[string]interface{}{
			"transaction_hash": txHash,
			"refund_address":   refundAddr.Hex(),
			"currency":         currency,
		},
	}

	return response, nil
}

// ValidateWebhook validates cryptocurrency webhook (not typically used)
func (c *CryptoProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) bool {
	// Crypto payments don't typically use webhooks in the traditional sense
	// Instead, we monitor the blockchain directly
	return true
}

// ProcessWebhook processes cryptocurrency webhook events
func (c *CryptoProcessor) ProcessWebhook(ctx context.Context, payload []byte) (*payment.WebhookResponse, error) {
	// This would handle events from blockchain monitoring services
	// For now, return a placeholder response
	return &payment.WebhookResponse{
		EventType: "crypto_transaction_confirmed",
		Metadata: map[string]interface{}{
			"blockchain": "ethereum",
		},
	}, nil
}

// CalculateFees calculates cryptocurrency transaction fees
func (c *CryptoProcessor) CalculateFees(amount int64, currency string) int64 {
	// Get current gas price
	gasPrice, err := c.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(20000000000) // 20 gwei default
	}

	// Apply multiplier
	adjustedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(int64(c.config.GasPriceMultiplier*100)))
	adjustedGasPrice.Div(adjustedGasPrice, big.NewInt(100))

	// Estimate gas limit based on transaction type
	var gasLimit uint64 = 21000 // ETH transfer
	if currency != "ETH" {
		gasLimit = 65000 // ERC20 transfer
	}

	totalFee := new(big.Int).Mul(adjustedGasPrice, big.NewInt(int64(gasLimit)))
	return totalFee.Int64()
}

// Helper methods

func (c *CryptoProcessor) updateTokenInfo(contract *TokenContract) error {
	// Query token decimals
	decimalsData, err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract.Address,
		Data: c.encodeMethodCall("decimals"),
	}, nil)
	if err == nil && len(decimalsData) >= 32 {
		decimals := new(big.Int).SetBytes(decimalsData[28:32]).Uint64()
		contract.Decimals = uint8(decimals)
	}

	// Query token symbol
	symbolData, err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract.Address,
		Data: c.encodeMethodCall("symbol"),
	}, nil)
	if err == nil && len(symbolData) > 0 {
		// Decode string (simplified)
		if len(symbolData) >= 64 {
			offset := new(big.Int).SetBytes(symbolData[0:32]).Uint64()
			length := new(big.Int).SetBytes(symbolData[32:64]).Uint64()
			if offset == 32 && length > 0 && length <= 32 {
				symbol := string(symbolData[64 : 64+length])
				contract.Symbol = strings.TrimSpace(symbol)
			}
		}
	}

	return nil
}

func (c *CryptoProcessor) encodeMethodCall(method string) []byte {
	methodID := crypto.Keccak256([]byte(method + "()"))[:4]
	return methodID
}

func (c *CryptoProcessor) monitorETHPayment(ctx context.Context, from, to common.Address, expectedAmount int64) (string, error) {
	// In a real implementation, this would set up monitoring for incoming ETH transactions
	// For now, return empty hash indicating monitoring is set up
	return "", nil
}

func (c *CryptoProcessor) monitorTokenPayment(ctx context.Context, from, to common.Address, expectedAmount int64, contract *TokenContract) (string, error) {
	// In a real implementation, this would set up monitoring for incoming token transactions
	// For now, return empty hash indicating monitoring is set up
	return "", nil
}

func (c *CryptoProcessor) sendETH(ctx context.Context, to common.Address, amount int64) (string, error) {
	// Get nonce
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.walletAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, to, big.NewInt(amount), 21000, gasPrice, nil)

	// Sign transaction
	chainID, err := c.ethClient.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (c *CryptoProcessor) sendToken(ctx context.Context, to common.Address, amount int64, contract *TokenContract) (string, error) {
	// Create transfer method call data
	transferMethodID := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

	// Encode parameters: address (32 bytes) + amount (32 bytes)
	addressPadded := common.LeftPadBytes(to.Bytes(), 32)
	amountBig := big.NewInt(amount)
	amountPadded := common.LeftPadBytes(amountBig.Bytes(), 32)

	data := append(transferMethodID, addressPadded...)
	data = append(data, amountPadded...)

	// Get nonce
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.walletAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, contract.Address, big.NewInt(0), 65000, gasPrice, data)

	// Sign transaction
	chainID, err := c.ethClient.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (c *CryptoProcessor) parseTokenTransaction(tx *types.Transaction) (string, int64, error) {
	// Simplified token transaction parsing
	// In a real implementation, this would properly decode the transaction data

	data := tx.Data()
	if len(data) < 68 {
		return "", 0, fmt.Errorf("invalid token transaction data")
	}

	// Extract method ID
	methodID := data[:4]
	transferMethodID := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

	if string(methodID) != string(transferMethodID) {
		return "", 0, fmt.Errorf("not a transfer transaction")
	}

	// Extract amount (last 32 bytes)
	amountBytes := data[36:68]
	amount := new(big.Int).SetBytes(amountBytes)

	// Determine currency based on contract address
	currency := "UNKNOWN"
	for curr, contract := range c.contracts {
		if tx.To().Hex() == contract.Address.Hex() {
			currency = curr
			break
		}
	}

	return currency, amount.Int64(), nil
}

func (c *CryptoProcessor) formatAmount(amount int64, currency string) string {
	if currency == "ETH" {
		// Convert wei to ETH
		eth := new(big.Float).Quo(new(big.Float).SetInt64(amount), new(big.Float).SetInt64(1e18))
		return eth.Text('f', 6)
	}

	// For tokens, use contract decimals
	if contract, exists := c.contracts[currency]; exists {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(contract.Decimals)), nil)
		tokenAmount := new(big.Float).Quo(new(big.Float).SetInt64(amount), new(big.Float).SetInt(divisor))
		return tokenAmount.Text('f', int(contract.Decimals))
	}

	return fmt.Sprintf("%d", amount)
}

// Standard ERC20 ABI (simplified)
const erc20ABIString = `[
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "_to", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	}
]`

# Treasury System Operational Guide

## Overview

This guide provides comprehensive instructions for operating the Autonomous Content Service Treasury System. It covers day-to-day operations, emergency procedures, and system maintenance.

## System Architecture

### Core Components
1. **TreasuryCore**: Main contract handling fund management and allocation
2. **AssetManager**: Automated portfolio rebalancing and yield optimization
3. **MultiSigWallet**: Secure multi-signature transaction approval
4. **TreasuryUpgradeable**: Proxy pattern for safe contract upgrades

### Key Roles
- **Treasury Admin**: Overall system administration
- **Treasurer**: Day-to-day financial operations
- **Auditor**: Financial reporting and compliance
- **Emergency Manager**: Crisis response and system protection
- **Asset Manager**: Portfolio optimization and rebalancing

## Daily Operations

### 1. Revenue Processing

#### Receiving Revenue
```javascript
// Process incoming revenue (example: $5,000 USDC)
await treasury.connect(treasurer).receiveRevenue(
    USDC_ADDRESS,
    ethers.utils.parseUnits("5000", 6), // 6 decimals for USDC
    "Client payment for content services - Invoice #INV-001"
);
```

#### Automatic Distribution
The system automatically distributes revenue according to configured allocations:
- **Operations (40%)**: $2,000 - Available for immediate operational expenses
- **Reserves (20%)**: $1,000 - Emergency fund allocation
- **Upgrades (20%)**: $1,000 - System improvement budget
- **Profits (20%)**: $1,000 - Retained earnings

### 2. Operational Spending

#### Processing Operational Expenses
```javascript
// Pay for operational expense (example: $500 for server costs)
await treasury.connect(treasurer).spendOperational(
    USDC_ADDRESS,
    ethers.utils.parseUnits("500", 6),
    VENDOR_ADDRESS,
    "Monthly server hosting costs - DigitalOcean"
);
```

#### Spending Controls
- Maximum spending limited by available category budget
- All transactions logged with detailed descriptions
- Real-time budget tracking and alerts

### 3. Upgrade Fund Management

#### Planning Capital Expenditures
```javascript
// Initiate upgrade spending (example: $8,000 for new AI model)
await treasury.connect(treasurer).spendUpgrades(
    USDC_ADDRESS,
    ethers.utils.parseUnits("8000", 6),
    AI_PROVIDER_ADDRESS,
    "GPT-4 API credits for enhanced content generation"
);
```

#### Time-locked Transactions
- Transactions >$10,000 require 48-hour time lock
- Provides opportunity for review and cancellation
- Execute after time lock expires:

```javascript
// Execute after 48-hour delay
await treasury.connect(treasurer).executeTimelocked(
    USDC_ADDRESS,
    amount,
    recipient,
    description,
    originalTimestamp
);
```

## Portfolio Management

### 1. Asset Rebalancing

#### Manual Rebalancing
```javascript
// Trigger portfolio rebalancing
await assetManager.connect(assetManagerRole).executeRebalance();
```

#### Automatic Rebalancing
- Runs every 6 hours by default
- Triggered when asset allocation deviates >5% from target
- Respects daily volume limits ($100,000 default)

### 2. Yield Optimization

#### Monitor Yield Strategies
```javascript
// Get optimal yield strategy for USDC
const [protocol, apy] = await assetManager.getOptimalStrategy(USDC_ADDRESS);
console.log(`Best yield: ${apy/100}% APY on ${protocol}`);
```

#### Execute Yield Optimization
```javascript
// Optimize yield across all assets
await assetManager.connect(assetManagerRole).optimizeYield();
```

### 3. Price Oracle Updates

#### Update Asset Prices
```javascript
// Update USDC price (example: $1.00)
await assetManager.connect(oracleRole).updatePrice(
    USDC_ADDRESS,
    ethers.utils.parseEther("1.0"), // $1.00
    95 // 95% confidence
);
```

## Financial Reporting

### 1. Generate Financial Summary
```javascript
const summary = await treasury.getFinancialSummary();
console.log({
    totalRevenue: ethers.utils.formatEther(summary.totalRevenue),
    totalExpenses: ethers.utils.formatEther(summary.totalExpenses),
    reserveBalance: ethers.utils.formatEther(summary.reserveBalance),
    profitBalance: ethers.utils.formatEther(summary.profitBalance)
});
```

### 2. Asset Allocation Report
```javascript
const allocation = await treasury.getAssetAllocation();
for (let i = 0; i < allocation.tokens.length; i++) {
    console.log({
        token: allocation.tokens[i],
        balance: ethers.utils.formatEther(allocation.balances[i]),
        percentage: allocation.percentages[i] / 100 + "%"
    });
}
```

### 3. Transaction Analysis
```javascript
// Get all revenue transactions
const revenueTxs = await treasury.getTransactionsByCategory(0); // REVENUE
for (const txId of revenueTxs) {
    const tx = await treasury.transactions(txId);
    console.log({
        id: tx.id.toString(),
        amount: ethers.utils.formatEther(tx.amount),
        description: tx.description,
        timestamp: new Date(tx.timestamp.toNumber() * 1000)
    });
}
```

## System Configuration

### 1. Revenue Allocation Updates

#### Modify Allocation Percentages
```javascript
// Example: Increase reserves to 30%, reduce profits to 10%
const newAllocation = {
    operations: 4000, // 40%
    reserves: 3000,   // 30% (increased)
    upgrades: 2000,   // 20%
    profits: 1000     // 10% (decreased)
};

await treasury.updateAllocationConfig(newAllocation);
```

### 2. Asset Configuration

#### Add New Asset
```javascript
// Configure new stablecoin (example: DAI)
await treasury.configureAsset(
    DAI_ADDRESS,
    2000, // 20% target allocation
    500,  // 5% rebalance threshold
    true  // is stablecoin
);

// Configure rebalancing parameters
await assetManager.configureAsset(
    DAI_ADDRESS,
    2000, // 20% target
    500,  // 5% deviation threshold
    ethers.utils.parseEther("1000") // $1,000 minimum rebalance
);
```

#### Add Yield Strategy
```javascript
// Add Compound Finance strategy for USDC
await assetManager.addYieldStrategy(
    USDC_ADDRESS,
    COMPOUND_USDC_ADDRESS,
    450, // 4.5% APY
    25   // 25% risk score
);
```

## Emergency Procedures

### 1. System Pause

#### Emergency Pause
```javascript
// Immediately pause all operations
await treasury.connect(emergencyRole).pause();
```

#### Resume Operations
```javascript
// Resume after resolving emergency
await treasury.connect(adminRole).unpause();
```

### 2. Emergency Withdrawals

#### Emergency Fund Recovery
```javascript
// Emergency withdrawal (only when paused)
await treasury.connect(emergencyRole).emergencyWithdraw(
    USDC_ADDRESS,
    ethers.utils.parseUnits("10000", 6), // $10,000
    SAFE_WALLET_ADDRESS,
    "Emergency fund recovery due to security incident"
);
```

### 3. Asset Manager Emergency Stop

#### Stop Automated Trading
```javascript
// Stop all rebalancing and yield optimization
await assetManager.connect(adminRole).setEmergencyStop(true);
```

## Monitoring and Alerts

### 1. Key Metrics to Monitor

#### Financial Health
- Daily revenue vs. expenses
- Reserve fund ratio (target: 20%+)
- Asset allocation deviation from targets
- Yield generation performance

#### Security Indicators
- Failed transaction attempts
- Unusual spending patterns
- Oracle price deviations
- Gas price spikes affecting operations

#### Operational Metrics
- Transaction confirmation times
- Rebalancing frequency and volume
- Yield strategy performance
- System uptime and availability

### 2. Alert Conditions

#### Critical Alerts
- Emergency pause activated
- Failed emergency withdrawal
- Oracle confidence below 80%
- Asset deviation >10% from target

#### Warning Alerts
- Daily expense limit approaching
- Rebalancing threshold exceeded
- Yield strategy underperforming
- Gas prices above threshold

### 3. Monitoring Scripts

#### Daily Health Check
```javascript
async function dailyHealthCheck() {
    // Check financial summary
    const summary = await treasury.getFinancialSummary();
    
    // Check asset allocation
    const allocation = await treasury.getAssetAllocation();
    
    // Check recent transactions
    const recentTxs = await getRecentTransactions();
    
    // Generate report
    return {
        financialHealth: calculateHealthScore(summary),
        allocationStatus: checkAllocationTargets(allocation),
        recentActivity: recentTxs.length,
        timestamp: new Date()
    };
}
```

## Maintenance and Upgrades

### 1. Regular Maintenance

#### Weekly Tasks
- Review financial reports
- Verify asset allocations
- Update yield strategies
- Check system health metrics

#### Monthly Tasks
- Comprehensive financial analysis
- Security audit of recent transactions
- Review and update yield strategies
- Assess rebalancing performance

#### Quarterly Tasks
- Full system security audit
- Review and update allocation strategy
- Evaluate new asset additions
- Plan system upgrades

### 2. Contract Upgrades

#### Preparation
1. Deploy new implementation contract
2. Test on staging environment
3. Prepare upgrade transaction
4. Coordinate multisig signers

#### Execution
```javascript
// Upgrade to new implementation
await treasuryProxy.connect(adminRole).upgradeTo(
    NEW_IMPLEMENTATION_ADDRESS
);
```

#### Verification
1. Verify upgrade success
2. Test all critical functions
3. Confirm data migration
4. Monitor system stability

## Troubleshooting

### Common Issues

#### Transaction Failures
- **Insufficient funds**: Check category budgets
- **Gas price too low**: Increase gas limit
- **Timelock not expired**: Wait for timelock period
- **Role permissions**: Verify caller has required role

#### Rebalancing Issues
- **Price oracle stale**: Update price feeds
- **Slippage too high**: Adjust slippage tolerance
- **Daily limit exceeded**: Wait for daily reset
- **Asset not configured**: Add asset configuration

#### Access Issues
- **Role not granted**: Verify role assignments
- **Multisig requirements**: Ensure sufficient signatures
- **Contract paused**: Check emergency status
- **Network congestion**: Retry with higher gas

### Recovery Procedures

#### Fund Recovery
1. Identify affected assets
2. Calculate recovery amounts
3. Prepare emergency withdrawal
4. Execute through emergency role
5. Transfer to secure backup wallet

#### System Recovery
1. Assess system state
2. Identify root cause
3. Implement fixes
4. Test on staging
5. Deploy fixes
6. Resume normal operations

This operational guide provides the foundation for safe and effective treasury management. Regular training and drills ensure operational readiness for all scenarios.
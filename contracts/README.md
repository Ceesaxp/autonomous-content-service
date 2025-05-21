# Autonomous Treasury System Smart Contracts

## Overview

The Autonomous Treasury System is a comprehensive blockchain-based financial infrastructure designed to manage digital assets autonomously while maintaining the highest security standards. It provides automated revenue distribution, multi-signature security, portfolio rebalancing, and regulatory compliance.

## Features

### 🔐 Security
- **Multi-signature wallet** with tiered approval thresholds
- **Time-locked transactions** for high-value operations
- **Role-based access control** with granular permissions
- **Emergency pause functionality** with circuit breaker pattern
- **Reentrancy protection** and safe math operations

### 💰 Financial Management
- **Automated revenue distribution** based on configurable allocations
- **Category-based spending controls** with real-time budget tracking
- **Multi-asset portfolio support** (ETH, USDC, DAI, etc.)
- **Comprehensive audit trail** with cryptographic verification
- **Financial reporting** for regulatory compliance

### 📊 Portfolio Management
- **Automated rebalancing** to maintain target allocations
- **Yield optimization** with risk-adjusted strategy selection
- **Price oracle integration** with confidence scoring
- **Slippage protection** and liquidity management
- **Emergency asset recovery** mechanisms

### 🔧 Operational Features
- **Upgradeable contracts** with proxy pattern
- **Parameter configuration** through governance
- **Health monitoring** and alerting systems
- **Backup and recovery** procedures
- **Integration APIs** for external systems

## Architecture

```
├── TreasuryCore.sol           # Main treasury contract
├── AssetManager.sol           # Portfolio management
├── TreasuryUpgradeable.sol    # Upgradeable proxy
├── security/
│   └── MultiSigWallet.sol     # Multi-signature security
├── utils/
│   ├── ReentrancyGuard.sol    # Reentrancy protection
│   └── Pausable.sol           # Emergency controls
├── governance/
│   └── AccessControl.sol      # Role management
└── interfaces/
    └── ITreasury.sol          # Core interfaces
```

## Quick Start

### Prerequisites

- Node.js v16+
- Hardhat
- Git

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-org/autonomous-content-service.git
cd autonomous-content-service/contracts
```

2. **Install dependencies**
```bash
npm install
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Compile contracts**
```bash
npm run compile
```

5. **Run tests**
```bash
npm test
```

## Deployment

### Local Development

1. **Start local blockchain**
```bash
npm run node
```

2. **Deploy contracts**
```bash
npm run deploy:local
```

### Testnet Deployment

1. **Configure testnet settings**
```bash
# Update .env with testnet configuration
NETWORK=goerli
PRIVATE_KEY=your_private_key
INFURA_PROJECT_ID=your_project_id
```

2. **Deploy to testnet**
```bash
npm run deploy:goerli
```

3. **Verify contracts**
```bash
npm run verify:goerli
```

### Mainnet Deployment

⚠️ **WARNING**: Mainnet deployment requires extreme caution

1. **Security checklist**
   - [ ] External audit completed
   - [ ] Multi-signature setup verified
   - [ ] Emergency procedures tested
   - [ ] Backup wallet prepared

2. **Deploy to mainnet**
```bash
npm run deploy:mainnet
```

## Configuration

### Revenue Allocation

Configure revenue distribution percentages:

```javascript
const allocationConfig = {
    operations: 4000, // 40% - Operating expenses
    reserves: 2000,   // 20% - Emergency reserves
    upgrades: 2000,   // 20% - System improvements
    profits: 2000     // 20% - Retained earnings
};
```

### Asset Portfolio

Configure supported assets and target allocations:

```javascript
// Configure USDC (60% allocation)
await treasury.configureAsset(
    USDC_ADDRESS,
    6000, // 60% target
    500,  // 5% rebalance threshold
    true  // is stablecoin
);

// Configure ETH (20% allocation)
await treasury.configureAsset(
    ethers.constants.AddressZero, // ETH
    2000, // 20% target
    1000, // 10% rebalance threshold
    false // not stablecoin
);
```

### Multi-signature Setup

Configure multi-signature wallet:

```javascript
const multisigConfig = {
    owners: [
        "0x742d...", // Primary admin
        "0x8e14...", // Treasury manager
        "0x1a5c...", // Security officer
        "0x9f2b..."  // Backup signer
    ],
    requiredSignatures: 3,
    timelockThreshold: ethers.utils.parseEther("10000") // $10,000
};
```

## Usage Examples

### Processing Revenue

```javascript
// Receive $5,000 USDC revenue
await treasury.connect(treasurer).receiveRevenue(
    USDC_ADDRESS,
    ethers.utils.parseUnits("5000", 6),
    "Client payment - Invoice #INV-001"
);
```

### Operational Spending

```javascript
// Pay $500 for server costs
await treasury.connect(treasurer).spendOperational(
    USDC_ADDRESS,
    ethers.utils.parseUnits("500", 6),
    vendorAddress,
    "Monthly server hosting - AWS"
);
```

### Portfolio Rebalancing

```javascript
// Trigger automatic rebalancing
await assetManager.connect(assetManager).executeRebalance();

// Check rebalancing status
const allocation = await treasury.getAssetAllocation();
console.log("Current allocation:", allocation);
```

### Financial Reporting

```javascript
// Generate financial summary
const summary = await treasury.getFinancialSummary();
console.log({
    totalRevenue: ethers.utils.formatEther(summary.totalRevenue),
    totalExpenses: ethers.utils.formatEther(summary.totalExpenses),
    reserves: ethers.utils.formatEther(summary.reserveBalance),
    profits: ethers.utils.formatEther(summary.profitBalance)
});
```

## Security Considerations

### Multi-signature Requirements

Different transaction amounts require different approval levels:

- **$0 - $1,000**: 2 signatures
- **$1,000 - $10,000**: 3 signatures  
- **$10,000 - $100,000**: 4 signatures
- **$100,000+**: Full consensus + 48-hour timelock

### Emergency Procedures

In case of emergency:

1. **Immediate pause**
```javascript
await treasury.connect(emergencyRole).pause();
```

2. **Emergency withdrawal**
```javascript
await treasury.connect(emergencyRole).emergencyWithdraw(
    tokenAddress,
    amount,
    safeAddress
);
```

### Role Management

Key roles and responsibilities:

- **DEFAULT_ADMIN_ROLE**: System administration
- **TREASURER_ROLE**: Daily financial operations
- **AUDITOR_ROLE**: Financial reporting
- **EMERGENCY_ROLE**: Crisis response
- **ASSET_MANAGER_ROLE**: Portfolio management

## Testing

### Unit Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run gas analysis
npm run test:gas
```

### Integration Tests

```bash
# Test complete workflows
npm run test:integration

# Test emergency scenarios
npm run test:emergency
```

## Monitoring and Maintenance

### Health Checks

Monitor key metrics:

- Revenue vs. expenses
- Asset allocation drift
- Yield performance
- Security events

### Regular Maintenance

- **Daily**: Financial report review
- **Weekly**: Asset allocation analysis
- **Monthly**: Security audit
- **Quarterly**: System upgrade planning

## Documentation

- [🏗️ Architecture](./docs/TreasurySystemArchitecture.md)
- [🔒 Security Analysis](./docs/SecurityAnalysis.md)
- [📖 Operational Guide](./docs/OperationalGuide.md)
- [📊 API Reference](./docs/APIReference.md)

## Support

### Getting Help

- 📧 Email: support@autonomous-content-service.com
- 💬 Discord: [Join our community](https://discord.gg/...)
- 📚 Docs: [Documentation portal](https://docs.autonomous-content-service.com)

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Bug Reports

Report security issues privately to: security@autonomous-content-service.com

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

⚠️ **Important**: This software is provided as-is. Users are responsible for their own security and should conduct thorough audits before mainnet deployment. The autonomous nature of this system requires careful consideration of all operational parameters and emergency procedures.
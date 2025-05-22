# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Autonomous Content Creation Service is a digital-native business that functions without human intervention, leveraging Large Language Models (LLMs) for decision-making, content creation, and autonomous operations (including, but not limited to: publishing new content, engaging clients, hiring help from human and other autonomous agents, signing contracts, sending invoices, performing payments, submitting regulatory reports). It utilizes the **Autonomous Treasury System** to manage its crypto positions and smart contracts.

You keep project progress in TODO.md, updating once each step is completed.

## Documentation



## Project Architecture

The system consists of the following major components:

1. **Cognitive Engine**: Handles LLM integration, context management, and content creation
2. **API Layer**: REST API for service interaction
3. **Domain Model**: Core business entities and relationships
4. **Repository Layer**: Data access and persistence
5. **Services Layer**: Business logic implementation
6. **Financial Infrastructure**: Financial infrastructure necessary for autonomous business operations
7. **Operational Systems**: Service execution framework, quality control mechanisms, and resource allocation system
8. **Governance Structure**: Decision protocols, rules enforcement, risk management, regulatory reporting


## Development Setup

To work with this codebase, you'll need:

1. Go 1.23+
2. PostgreSQL database
3. Node.js v16+ and NPM
4. Hardhat for smart contract development
5. Ethereum development environment (Ganache/Hardhat network)
6. Environment variables (see .env.example)

### Smart Contract Development Setup

The project includes a comprehensive smart contract treasury system for autonomous financial operations:

1. **Install contract dependencies**:
   ```bash
   cd contracts/
   npm install
   ```

2. **Configure environment variables**:
   ```bash
   cp .env.example .env
   # Configure:
   # - PRIVATE_KEY: Your wallet private key
   # - INFURA_PROJECT_ID: Infura API key for network access
   # - ETHERSCAN_API_KEY: For contract verification
   ```

3. **Compile smart contracts**:
   ```bash
   npm run compile
   ```

4. **Run local blockchain**:
   ```bash
   npm run node  # Starts Hardhat local network
   ```

## Common Commands

### Building and Running

```bash
# Build the service
go build -o content-service ./src

# Run the service
./content-service

# Run with environment file
source .env && ./content-service
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./src/services/content_creation/...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Database Management

```bash
# Create database (PostgreSQL)
createdb contentservice

# Run migrations
psql contentservice < src/infrastructure/database/schema.sql
```

## Smart Contract Treasury

The project includes a comprehensive blockchain-based treasury system for autonomous financial operations:

### Key Features

- **Multi-signature Security**: Tiered approval thresholds for different transaction amounts
- **Automated Revenue Distribution**: Configurable allocation between operations, reserves, upgrades, and profits
- **Multi-asset Portfolio**: Support for ETH, USDC, DAI and other ERC20 tokens
- **Portfolio Rebalancing**: Automated asset allocation maintenance
- **Comprehensive Audit Trail**: All transactions cryptographically verified
- **Role-based Access Control**: Treasurer, Auditor, Emergency roles
- **Emergency Controls**: Pause functionality and emergency withdrawal procedures


```bash
# Compile smart contracts
cd contracts/ && npm run compile

# Run contract tests
npm test

# Deploy contracts locally
npm run deploy:local

# Deploy to testnets
npm run deploy:goerli

# Deploy to mainnet (use with extreme caution)
npm run deploy:mainnet

# Verify contracts on Etherscan
npm run verify:goerli

# Run gas analysis
npm run test:gas

# Generate coverage report
npm run test:coverage

# Lint Solidity code
npm run lint
```

### API Testing

The service exposes a REST API. You can test it with:

```bash
# Test content creation endpoint
curl -X POST http://localhost:8080/api/v1/projects/{project_id}/content -H "Content-Type: application/json" -d '{"title":"Sample Content","type":"blog_post","target_audience":"developers"}'
```

## Code Patterns and Conventions

### Directory Structure

- `/src`: Main Go source code
  - `/api`: API handlers and routing
  - `/config`: Configuration loading
  - `/domain`: Core domain model (entities, events)
  - `/infrastructure`: Database and external integrations
  - `/services`: Business logic implementation
- `/contracts`: Smart contract treasury system
  - `/src`: Solidity smart contracts
    - `TreasuryCore.sol`: Main treasury contract
    - `AssetManager.sol`: Portfolio management
    - `TreasuryUpgradeable.sol`: Upgradeable proxy
    - `/security`: Multi-signature and security contracts
    - `/governance`: Access control and governance
    - `/interfaces`: Contract interfaces
  - `/test`: Contract test suites
  - `/scripts`: Deployment scripts
  - `hardhat.config.js`: Hardhat configuration
  - `package.json`: NPM dependencies

### Interfaces

Key Go interfaces in the system:

- `LLMClient`: Interaction with language models
- `ContextManager`: Manages context for LLM prompts
- `QualityChecker`: Validates content quality
- `Repository`: Data access interfaces for each entity
- `PaymentProcessor`: Payment processing (Stripe, crypto)
- `CryptoWallet`: Cryptocurrency operations

Key Smart Contract interfaces:

- `ITreasury`: Core treasury operations
- `IAssetManager`: Portfolio management
- `IAccessControl`: Role-based permissions
- `IERC20`: Token standard interface

### Error Handling

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Configuration

The application uses environment variables for configuration with the following pattern:

```go
config.LoadConfig() // Loads from environment or .env file
```

### Dependency Injection

The codebase uses constructor-based dependency injection:

```go
func NewContentPipeline(repo Repository, client LLMClient, ...) *ContentPipeline {
    return &ContentPipeline{...}
}
```

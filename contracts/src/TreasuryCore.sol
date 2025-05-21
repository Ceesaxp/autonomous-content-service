// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/ITreasury.sol";
import "./security/MultiSigWallet.sol";
import "./utils/ReentrancyGuard.sol";
import "./utils/Pausable.sol";
import "./governance/AccessControl.sol";

/**
 * @title TreasuryCore
 * @dev Main treasury contract managing autonomous service funds
 * @author Autonomous Content Service
 */
contract TreasuryCore is ITreasury, MultiSigWallet, ReentrancyGuard, Pausable, AccessControl {
    using SafeMath for uint256;

    // Role definitions
    bytes32 public constant TREASURER_ROLE = keccak256("TREASURER_ROLE");
    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");
    bytes32 public constant EMERGENCY_ROLE = keccak256("EMERGENCY_ROLE");

    // Revenue allocation percentages (basis points, 10000 = 100%)
    struct AllocationConfig {
        uint256 operations;      // Operating expenses
        uint256 reserves;        // Emergency reserves
        uint256 upgrades;        // System improvements
        uint256 profits;         // Retained earnings
    }

    // Transaction categories for accounting
    enum TransactionCategory {
        REVENUE,
        OPERATING_EXPENSE,
        CAPITAL_EXPENSE,
        RESERVE_ALLOCATION,
        PROFIT_DISTRIBUTION,
        EMERGENCY_WITHDRAWAL
    }

    // Financial transaction record
    struct FinancialTransaction {
        uint256 id;
        address token;
        uint256 amount;
        TransactionCategory category;
        string description;
        uint256 timestamp;
        address initiator;
        bytes32 referenceHash;
    }

    // Asset management configuration
    struct AssetConfig {
        address token;
        uint256 targetPercentage;  // Target allocation percentage
        uint256 rebalanceThreshold; // Threshold for rebalancing
        bool isStablecoin;
        bool isActive;
    }

    // State variables
    AllocationConfig public allocationConfig;
    mapping(address => AssetConfig) public assetConfigs;
    address[] public managedAssets;
    
    mapping(uint256 => FinancialTransaction) public transactions;
    uint256 public transactionCounter;
    
    mapping(TransactionCategory => uint256) public categoryTotals;
    mapping(address => mapping(TransactionCategory => uint256)) public tokenCategoryTotals;

    // Rebalancing parameters
    uint256 public lastRebalanceTime;
    uint256 public rebalanceInterval = 24 hours;
    uint256 public emergencyReserveTarget = 2000; // 20% in basis points

    // Time lock for large transactions
    mapping(bytes32 => uint256) public timelocks;
    uint256 public constant TIMELOCK_DURATION = 48 hours;
    uint256 public timelockThreshold = 10000 * 10**18; // $10,000 equivalent

    // Events
    event RevenueReceived(address indexed token, uint256 amount, uint256 timestamp);
    event FundsDistributed(TransactionCategory category, address token, uint256 amount);
    event AssetRebalanced(address token, uint256 oldBalance, uint256 newBalance);
    event AllocationConfigUpdated(AllocationConfig newConfig);
    event EmergencyWithdrawal(address token, uint256 amount, address recipient);
    event TransactionRecorded(uint256 indexed txId, TransactionCategory category, uint256 amount);

    /**
     * @dev Constructor
     * @param _owners Initial multisig owners
     * @param _required Required signatures for multisig
     */
    constructor(
        address[] memory _owners,
        uint256 _required
    ) MultiSigWallet(_owners, _required) {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(TREASURER_ROLE, msg.sender);
        
        // Default allocation: 40% operations, 20% reserves, 20% upgrades, 20% profits
        allocationConfig = AllocationConfig({
            operations: 4000,
            reserves: 2000,
            upgrades: 2000,
            profits: 2000
        });
    }

    /**
     * @dev Receive revenue and automatically distribute according to allocation
     */
    function receiveRevenue(
        address token,
        uint256 amount,
        string calldata description
    ) external payable onlyRole(TREASURER_ROLE) nonReentrant whenNotPaused {
        require(amount > 0, "Amount must be positive");
        
        if (token == address(0)) {
            require(msg.value == amount, "ETH amount mismatch");
        } else {
            IERC20(token).transferFrom(msg.sender, address(this), amount);
        }

        _recordTransaction(token, amount, TransactionCategory.REVENUE, description);
        _distributeRevenue(token, amount);
        
        emit RevenueReceived(token, amount, block.timestamp);
    }

    /**
     * @dev Distribute revenue according to allocation configuration
     */
    function _distributeRevenue(address token, uint256 amount) internal {
        uint256 operationsAmount = amount.mul(allocationConfig.operations).div(10000);
        uint256 reservesAmount = amount.mul(allocationConfig.reserves).div(10000);
        uint256 upgradesAmount = amount.mul(allocationConfig.upgrades).div(10000);
        uint256 profitsAmount = amount.sub(operationsAmount).sub(reservesAmount).sub(upgradesAmount);

        categoryTotals[TransactionCategory.OPERATING_EXPENSE] = categoryTotals[TransactionCategory.OPERATING_EXPENSE].add(operationsAmount);
        categoryTotals[TransactionCategory.RESERVE_ALLOCATION] = categoryTotals[TransactionCategory.RESERVE_ALLOCATION].add(reservesAmount);
        categoryTotals[TransactionCategory.CAPITAL_EXPENSE] = categoryTotals[TransactionCategory.CAPITAL_EXPENSE].add(upgradesAmount);
        categoryTotals[TransactionCategory.PROFIT_DISTRIBUTION] = categoryTotals[TransactionCategory.PROFIT_DISTRIBUTION].add(profitsAmount);

        tokenCategoryTotals[token][TransactionCategory.OPERATING_EXPENSE] = tokenCategoryTotals[token][TransactionCategory.OPERATING_EXPENSE].add(operationsAmount);
        tokenCategoryTotals[token][TransactionCategory.RESERVE_ALLOCATION] = tokenCategoryTotals[token][TransactionCategory.RESERVE_ALLOCATION].add(reservesAmount);
        tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE] = tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE].add(upgradesAmount);
        tokenCategoryTotals[token][TransactionCategory.PROFIT_DISTRIBUTION] = tokenCategoryTotals[token][TransactionCategory.PROFIT_DISTRIBUTION].add(profitsAmount);

        emit FundsDistributed(TransactionCategory.OPERATING_EXPENSE, token, operationsAmount);
        emit FundsDistributed(TransactionCategory.RESERVE_ALLOCATION, token, reservesAmount);
        emit FundsDistributed(TransactionCategory.CAPITAL_EXPENSE, token, upgradesAmount);
        emit FundsDistributed(TransactionCategory.PROFIT_DISTRIBUTION, token, profitsAmount);
    }

    /**
     * @dev Spend funds for operational expenses
     */
    function spendOperational(
        address token,
        uint256 amount,
        address recipient,
        string calldata description
    ) external onlyRole(TREASURER_ROLE) nonReentrant whenNotPaused {
        require(tokenCategoryTotals[token][TransactionCategory.OPERATING_EXPENSE] >= amount, "Insufficient operational funds");
        
        tokenCategoryTotals[token][TransactionCategory.OPERATING_EXPENSE] = tokenCategoryTotals[token][TransactionCategory.OPERATING_EXPENSE].sub(amount);
        
        _recordTransaction(token, amount, TransactionCategory.OPERATING_EXPENSE, description);
        _transferFunds(token, recipient, amount);
    }

    /**
     * @dev Spend funds for upgrades and improvements
     */
    function spendUpgrades(
        address token,
        uint256 amount,
        address recipient,
        string calldata description
    ) external onlyRole(TREASURER_ROLE) nonReentrant whenNotPaused {
        require(tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE] >= amount, "Insufficient upgrade funds");
        
        if (amount > timelockThreshold) {
            bytes32 timelockId = keccak256(abi.encodePacked(token, amount, recipient, block.timestamp));
            timelocks[timelockId] = block.timestamp.add(TIMELOCK_DURATION);
            return;
        }
        
        tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE] = tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE].sub(amount);
        
        _recordTransaction(token, amount, TransactionCategory.CAPITAL_EXPENSE, description);
        _transferFunds(token, recipient, amount);
    }

    /**
     * @dev Execute timelocked transaction
     */
    function executeTimelocked(
        address token,
        uint256 amount,
        address recipient,
        string calldata description,
        uint256 timestamp
    ) external onlyRole(TREASURER_ROLE) {
        bytes32 timelockId = keccak256(abi.encodePacked(token, amount, recipient, timestamp));
        require(timelocks[timelockId] != 0, "Timelock not found");
        require(block.timestamp >= timelocks[timelockId], "Timelock not expired");
        
        delete timelocks[timelockId];
        
        tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE] = tokenCategoryTotals[token][TransactionCategory.CAPITAL_EXPENSE].sub(amount);
        _recordTransaction(token, amount, TransactionCategory.CAPITAL_EXPENSE, description);
        _transferFunds(token, recipient, amount);
    }

    /**
     * @dev Rebalance portfolio according to target allocations
     */
    function rebalancePortfolio() external onlyRole(TREASURER_ROLE) whenNotPaused {
        require(block.timestamp >= lastRebalanceTime.add(rebalanceInterval), "Rebalance too frequent");
        
        lastRebalanceTime = block.timestamp;
        
        uint256 totalValue = _calculateTotalPortfolioValue();
        
        for (uint256 i = 0; i < managedAssets.length; i++) {
            address token = managedAssets[i];
            AssetConfig storage config = assetConfigs[token];
            
            if (!config.isActive) continue;
            
            uint256 currentBalance = _getTokenBalance(token);
            uint256 targetBalance = totalValue.mul(config.targetPercentage).div(10000);
            
            if (_shouldRebalance(currentBalance, targetBalance, config.rebalanceThreshold)) {
                _rebalanceAsset(token, currentBalance, targetBalance);
                emit AssetRebalanced(token, currentBalance, targetBalance);
            }
        }
    }

    /**
     * @dev Emergency withdrawal function
     */
    function emergencyWithdraw(
        address token,
        uint256 amount,
        address recipient
    ) external onlyRole(EMERGENCY_ROLE) whenPaused {
        _recordTransaction(token, amount, TransactionCategory.EMERGENCY_WITHDRAWAL, "Emergency withdrawal");
        _transferFunds(token, recipient, amount);
        emit EmergencyWithdrawal(token, amount, recipient);
    }

    /**
     * @dev Update allocation configuration
     */
    function updateAllocationConfig(
        AllocationConfig calldata newConfig
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(
            newConfig.operations.add(newConfig.reserves).add(newConfig.upgrades).add(newConfig.profits) == 10000,
            "Allocations must sum to 100%"
        );
        
        allocationConfig = newConfig;
        emit AllocationConfigUpdated(newConfig);
    }

    /**
     * @dev Add or update asset configuration
     */
    function configureAsset(
        address token,
        uint256 targetPercentage,
        uint256 rebalanceThreshold,
        bool isStablecoin
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (!assetConfigs[token].isActive) {
            managedAssets.push(token);
        }
        
        assetConfigs[token] = AssetConfig({
            token: token,
            targetPercentage: targetPercentage,
            rebalanceThreshold: rebalanceThreshold,
            isStablecoin: isStablecoin,
            isActive: true
        });
    }

    // Internal helper functions
    function _recordTransaction(
        address token,
        uint256 amount,
        TransactionCategory category,
        string memory description
    ) internal {
        transactionCounter = transactionCounter.add(1);
        
        transactions[transactionCounter] = FinancialTransaction({
            id: transactionCounter,
            token: token,
            amount: amount,
            category: category,
            description: description,
            timestamp: block.timestamp,
            initiator: msg.sender,
            referenceHash: keccak256(abi.encodePacked(transactionCounter, token, amount, category, block.timestamp))
        });
        
        emit TransactionRecorded(transactionCounter, category, amount);
    }

    function _transferFunds(address token, address recipient, uint256 amount) internal {
        if (token == address(0)) {
            payable(recipient).transfer(amount);
        } else {
            IERC20(token).transfer(recipient, amount);
        }
    }

    function _getTokenBalance(address token) internal view returns (uint256) {
        if (token == address(0)) {
            return address(this).balance;
        } else {
            return IERC20(token).balanceOf(address(this));
        }
    }

    function _calculateTotalPortfolioValue() internal view returns (uint256) {
        // Implementation would require price oracles
        // Simplified for now
        uint256 total = 0;
        for (uint256 i = 0; i < managedAssets.length; i++) {
            total = total.add(_getTokenBalance(managedAssets[i]));
        }
        return total;
    }

    function _shouldRebalance(
        uint256 current,
        uint256 target,
        uint256 threshold
    ) internal pure returns (bool) {
        if (target == 0) return false;
        
        uint256 deviation = current > target ? current.sub(target) : target.sub(current);
        return deviation.mul(10000).div(target) > threshold;
    }

    function _rebalanceAsset(address token, uint256 current, uint256 target) internal {
        // Implementation would involve DEX integration for swapping
        // Placeholder for actual rebalancing logic
    }

    // View functions for financial reporting
    function getTransactionsByCategory(TransactionCategory category) external view returns (uint256[] memory) {
        uint256[] memory txIds = new uint256[](transactionCounter);
        uint256 count = 0;
        
        for (uint256 i = 1; i <= transactionCounter; i++) {
            if (transactions[i].category == category) {
                txIds[count] = i;
                count++;
            }
        }
        
        assembly {
            mstore(txIds, count)
        }
        
        return txIds;
    }

    function getFinancialSummary() external view returns (
        uint256 totalRevenue,
        uint256 totalExpenses,
        uint256 reserveBalance,
        uint256 profitBalance
    ) {
        totalRevenue = categoryTotals[TransactionCategory.REVENUE];
        totalExpenses = categoryTotals[TransactionCategory.OPERATING_EXPENSE].add(categoryTotals[TransactionCategory.CAPITAL_EXPENSE]);
        reserveBalance = categoryTotals[TransactionCategory.RESERVE_ALLOCATION];
        profitBalance = categoryTotals[TransactionCategory.PROFIT_DISTRIBUTION];
    }

    function getAssetAllocation() external view returns (
        address[] memory tokens,
        uint256[] memory balances,
        uint256[] memory percentages
    ) {
        tokens = new address[](managedAssets.length);
        balances = new uint256[](managedAssets.length);
        percentages = new uint256[](managedAssets.length);
        
        uint256 totalValue = _calculateTotalPortfolioValue();
        
        for (uint256 i = 0; i < managedAssets.length; i++) {
            tokens[i] = managedAssets[i];
            balances[i] = _getTokenBalance(managedAssets[i]);
            percentages[i] = totalValue > 0 ? balances[i].mul(10000).div(totalValue) : 0;
        }
    }

    // Emergency functions
    function pause() external onlyRole(EMERGENCY_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }
}
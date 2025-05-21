// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/ITreasury.sol";
import "./governance/AccessControl.sol";
import "./utils/ReentrancyGuard.sol";

/**
 * @title AssetManager
 * @dev Manages asset rebalancing and yield optimization
 */
contract AssetManager is AccessControl, ReentrancyGuard {
    using SafeMath for uint256;

    bytes32 public constant ASSET_MANAGER_ROLE = keccak256("ASSET_MANAGER_ROLE");
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");

    struct PriceData {
        uint256 price;
        uint256 timestamp;
        uint256 confidence;
    }

    struct YieldStrategy {
        address protocol;
        uint256 apy;
        uint256 tvl;
        uint256 riskScore; // 1-100, higher is riskier
        bool active;
    }

    struct RebalanceConfig {
        uint256 targetPercentage;
        uint256 maxDeviation;
        uint256 minRebalanceAmount;
        bool enabled;
    }

    // Price oracles
    mapping(address => PriceData) public priceData;
    mapping(address => address) public priceOracles;

    // Asset configurations
    mapping(address => RebalanceConfig) public rebalanceConfigs;
    address[] public supportedAssets;

    // Yield strategies
    mapping(address => YieldStrategy[]) public yieldStrategies;
    mapping(address => mapping(address => uint256)) public assetAllocations; // asset => protocol => amount

    // Rebalancing settings
    uint256 public rebalanceInterval = 6 hours;
    uint256 public lastRebalance;
    uint256 public slippageTolerance = 200; // 2% in basis points

    // Emergency settings
    bool public emergencyStop = false;
    uint256 public maxDailyRebalanceVolume = 100000 * 10**18; // $100k
    uint256 public dailyRebalanceVolume = 0;
    uint256 public lastResetTimestamp;

    event PriceUpdated(address indexed asset, uint256 price, uint256 timestamp);
    event AssetRebalanced(address indexed asset, uint256 fromAmount, uint256 toAmount);
    event YieldStrategyAdded(address indexed asset, address indexed protocol, uint256 apy);
    event YieldGenerated(address indexed asset, address indexed protocol, uint256 amount);
    event EmergencyStopToggled(bool enabled);

    modifier onlyWhenActive() {
        require(!emergencyStop, "Emergency stop active");
        _;
    }

    modifier validAsset(address asset) {
        require(rebalanceConfigs[asset].enabled, "Asset not supported");
        _;
    }

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(ASSET_MANAGER_ROLE, msg.sender);
        lastResetTimestamp = block.timestamp;
    }

    /**
     * @dev Update price data for an asset
     */
    function updatePrice(
        address asset,
        uint256 price,
        uint256 confidence
    ) external onlyRole(ORACLE_ROLE) {
        require(confidence >= 80, "Price confidence too low");
        
        priceData[asset] = PriceData({
            price: price,
            timestamp: block.timestamp,
            confidence: confidence
        });
        
        emit PriceUpdated(asset, price, block.timestamp);
    }

    /**
     * @dev Add or update rebalance configuration for an asset
     */
    function configureAsset(
        address asset,
        uint256 targetPercentage,
        uint256 maxDeviation,
        uint256 minRebalanceAmount
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (!rebalanceConfigs[asset].enabled) {
            supportedAssets.push(asset);
        }
        
        rebalanceConfigs[asset] = RebalanceConfig({
            targetPercentage: targetPercentage,
            maxDeviation: maxDeviation,
            minRebalanceAmount: minRebalanceAmount,
            enabled: true
        });
    }

    /**
     * @dev Add yield strategy for an asset
     */
    function addYieldStrategy(
        address asset,
        address protocol,
        uint256 apy,
        uint256 riskScore
    ) external onlyRole(ASSET_MANAGER_ROLE) validAsset(asset) {
        require(riskScore <= 100, "Invalid risk score");
        
        yieldStrategies[asset].push(YieldStrategy({
            protocol: protocol,
            apy: apy,
            tvl: 0,
            riskScore: riskScore,
            active: true
        }));
        
        emit YieldStrategyAdded(asset, protocol, apy);
    }

    /**
     * @dev Execute automatic rebalancing
     */
    function executeRebalance() external onlyRole(ASSET_MANAGER_ROLE) onlyWhenActive nonReentrant {
        require(
            block.timestamp >= lastRebalance.add(rebalanceInterval),
            "Rebalance too frequent"
        );
        
        _resetDailyLimits();
        
        uint256 totalPortfolioValue = _calculateTotalValue();
        
        for (uint256 i = 0; i < supportedAssets.length; i++) {
            address asset = supportedAssets[i];
            RebalanceConfig memory config = rebalanceConfigs[asset];
            
            if (!config.enabled) continue;
            
            uint256 currentValue = _getAssetValue(asset);
            uint256 targetValue = totalPortfolioValue.mul(config.targetPercentage).div(10000);
            
            if (_shouldRebalance(currentValue, targetValue, config.maxDeviation, config.minRebalanceAmount)) {
                _rebalanceAsset(asset, currentValue, targetValue);
            }
        }
        
        lastRebalance = block.timestamp;
    }

    /**
     * @dev Optimize yield across all assets
     */
    function optimizeYield() external onlyRole(ASSET_MANAGER_ROLE) onlyWhenActive {
        for (uint256 i = 0; i < supportedAssets.length; i++) {
            address asset = supportedAssets[i];
            _optimizeAssetYield(asset);
        }
    }

    /**
     * @dev Emergency stop functionality
     */
    function setEmergencyStop(bool enabled) external onlyRole(DEFAULT_ADMIN_ROLE) {
        emergencyStop = enabled;
        emit EmergencyStopToggled(enabled);
    }

    /**
     * @dev Get optimal yield strategy for an asset
     */
    function getOptimalStrategy(address asset) external view returns (address protocol, uint256 apy) {
        YieldStrategy[] memory strategies = yieldStrategies[asset];
        uint256 bestScore = 0;
        uint256 bestIndex = 0;
        
        for (uint256 i = 0; i < strategies.length; i++) {
            if (!strategies[i].active) continue;
            
            // Calculate risk-adjusted yield score
            uint256 score = strategies[i].apy.mul(100 - strategies[i].riskScore).div(100);
            
            if (score > bestScore) {
                bestScore = score;
                bestIndex = i;
            }
        }
        
        if (strategies.length > 0 && strategies[bestIndex].active) {
            return (strategies[bestIndex].protocol, strategies[bestIndex].apy);
        }
        
        return (address(0), 0);
    }

    /**
     * @dev Get current portfolio allocation
     */
    function getPortfolioAllocation() external view returns (
        address[] memory assets,
        uint256[] memory values,
        uint256[] memory percentages
    ) {
        assets = supportedAssets;
        values = new uint256[](supportedAssets.length);
        percentages = new uint256[](supportedAssets.length);
        
        uint256 totalValue = _calculateTotalValue();
        
        for (uint256 i = 0; i < supportedAssets.length; i++) {
            values[i] = _getAssetValue(supportedAssets[i]);
            percentages[i] = totalValue > 0 ? values[i].mul(10000).div(totalValue) : 0;
        }
    }

    // Internal functions
    function _calculateTotalValue() internal view returns (uint256 total) {
        for (uint256 i = 0; i < supportedAssets.length; i++) {
            total = total.add(_getAssetValue(supportedAssets[i]));
        }
    }

    function _getAssetValue(address asset) internal view returns (uint256) {
        uint256 balance = _getAssetBalance(asset);
        PriceData memory price = priceData[asset];
        
        require(block.timestamp - price.timestamp <= 1 hours, "Price data stale");
        
        return balance.mul(price.price).div(10**18);
    }

    function _getAssetBalance(address asset) internal view returns (uint256) {
        if (asset == address(0)) {
            return address(this).balance;
        } else {
            return IERC20(asset).balanceOf(address(this));
        }
    }

    function _shouldRebalance(
        uint256 current,
        uint256 target,
        uint256 maxDeviation,
        uint256 minAmount
    ) internal pure returns (bool) {
        if (target == 0) return false;
        
        uint256 deviation = current > target ? current.sub(target) : target.sub(current);
        uint256 deviationPercentage = deviation.mul(10000).div(target);
        
        return deviationPercentage > maxDeviation && deviation > minAmount;
    }

    function _rebalanceAsset(address asset, uint256 current, uint256 target) internal {
        require(
            dailyRebalanceVolume.add(current > target ? current.sub(target) : target.sub(current)) <= maxDailyRebalanceVolume,
            "Daily rebalance limit exceeded"
        );
        
        // Implementation would involve DEX integration
        // This is a placeholder for the actual rebalancing logic
        
        emit AssetRebalanced(asset, current, target);
        
        if (current > target) {
            dailyRebalanceVolume = dailyRebalanceVolume.add(current.sub(target));
        } else {
            dailyRebalanceVolume = dailyRebalanceVolume.add(target.sub(current));
        }
    }

    function _optimizeAssetYield(address asset) internal {
        (address optimalProtocol, uint256 apy) = this.getOptimalStrategy(asset);
        
        if (optimalProtocol != address(0)) {
            // Implementation would involve yield protocol integration
            // This is a placeholder for actual yield optimization
            emit YieldGenerated(asset, optimalProtocol, apy);
        }
    }

    function _resetDailyLimits() internal {
        if (block.timestamp >= lastResetTimestamp.add(24 hours)) {
            dailyRebalanceVolume = 0;
            lastResetTimestamp = block.timestamp;
        }
    }

    // Admin functions
    function updateRebalanceInterval(uint256 interval) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(interval >= 1 hours && interval <= 7 days, "Invalid interval");
        rebalanceInterval = interval;
    }

    function updateSlippageTolerance(uint256 tolerance) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(tolerance <= 1000, "Tolerance too high"); // Max 10%
        slippageTolerance = tolerance;
    }

    function updateDailyLimit(uint256 limit) external onlyRole(DEFAULT_ADMIN_ROLE) {
        maxDailyRebalanceVolume = limit;
    }
}
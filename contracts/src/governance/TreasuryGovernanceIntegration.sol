// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "../TreasuryCore.sol";
import "../AssetManager.sol";

/**
 * @title TreasuryGovernanceIntegration
 * @notice Integration contract between DAO governance and existing treasury system
 * @dev Allows governance-controlled treasury operations while maintaining security
 */
contract TreasuryGovernanceIntegration is AccessControl, ReentrancyGuard, Pausable {
    /// @notice Role for governance operations
    bytes32 public constant GOVERNANCE_ROLE = keccak256("GOVERNANCE_ROLE");
    
    /// @notice Role for emergency operations
    bytes32 public constant EMERGENCY_ROLE = keccak256("EMERGENCY_ROLE");
    
    /// @notice Role for treasury administrators
    bytes32 public constant TREASURY_ADMIN_ROLE = keccak256("TREASURY_ADMIN_ROLE");
    
    /// @notice Reference to the main treasury contract
    TreasuryCore public immutable treasury;
    
    /// @notice Reference to the asset manager contract
    AssetManager public immutable assetManager;
    
    /// @notice Maximum single transfer amount without additional approval
    uint256 public maxSingleTransfer;
    
    /// @notice Maximum daily transfer amount
    uint256 public maxDailyTransfer;
    
    /// @notice Tracking of daily transfers
    mapping(uint256 => uint256) public dailyTransfers; // day => amount
    
    /// @notice Mapping of approved allocations
    mapping(bytes32 => Allocation) public allocations;
    
    /// @notice Counter for allocation IDs
    uint256 public nextAllocationId;
    
    /// @notice Emergency withdrawal recipient
    address public emergencyRecipient;
    
    struct Allocation {
        bytes32 id;
        address recipient;
        uint256 amount;
        address asset;
        string purpose;
        uint256 approvedAt;
        uint256 expiresAt;
        bool executed;
        bool isBatch;
        uint256 installments;
        uint256 installmentAmount;
        uint256 nextInstallmentDate;
        uint256 installmentsPaid;
    }
    
    struct BatchTransfer {
        address[] recipients;
        uint256[] amounts;
        address asset;
        string purpose;
    }
    
    event AllocationApproved(
        bytes32 indexed id,
        address indexed recipient,
        uint256 amount,
        address asset,
        string purpose
    );
    
    event AllocationExecuted(
        bytes32 indexed id,
        address indexed recipient,
        uint256 amount,
        address asset
    );
    
    event InstallmentPaid(
        bytes32 indexed id,
        address indexed recipient,
        uint256 amount,
        uint256 installmentNumber
    );
    
    event BatchTransferExecuted(
        address[] recipients,
        uint256[] amounts,
        address asset,
        string purpose
    );
    
    event EmergencyWithdrawal(
        address indexed recipient,
        uint256 amount,
        address asset,
        string reason
    );
    
    event ParameterUpdated(string parameter, uint256 oldValue, uint256 newValue);
    event EmergencyRecipientUpdated(address oldRecipient, address newRecipient);
    
    error AllocationNotFound();
    error AllocationExpired();
    error AllocationAlreadyExecuted();
    error ExceedsMaxTransfer();
    error ExceedsDailyLimit();
    error InvalidAllocation();
    error InsufficientTreasuryBalance();
    error InvalidInstallmentSchedule();
    error InstallmentNotDue();
    
    /**
     * @notice Constructor to initialize treasury governance integration
     * @param _treasury Address of the treasury contract
     * @param _assetManager Address of the asset manager contract
     * @param _governance Address of the governance contract
     * @param _emergencyRecipient Address for emergency withdrawals
     */
    constructor(
        address _treasury,
        address _assetManager,
        address _governance,
        address _emergencyRecipient
    ) {
        require(_treasury != address(0), "Invalid treasury address");
        require(_assetManager != address(0), "Invalid asset manager address");
        require(_governance != address(0), "Invalid governance address");
        require(_emergencyRecipient != address(0), "Invalid emergency recipient");
        
        treasury = TreasuryCore(_treasury);
        assetManager = AssetManager(_assetManager);
        emergencyRecipient = _emergencyRecipient;
        
        // Set initial parameters
        maxSingleTransfer = 50000 * 10**18; // 50,000 USD equivalent
        maxDailyTransfer = 200000 * 10**18; // 200,000 USD equivalent
        
        // Setup roles
        _grantRole(DEFAULT_ADMIN_ROLE, _governance);
        _grantRole(GOVERNANCE_ROLE, _governance);
        _grantRole(TREASURY_ADMIN_ROLE, msg.sender);
        _grantRole(EMERGENCY_ROLE, _emergencyRecipient);
    }
    
    /**
     * @notice Approves a treasury allocation via governance
     * @param recipient Address to receive the allocation
     * @param amount Amount to allocate
     * @param asset Asset address (address(0) for ETH)
     * @param purpose Purpose of the allocation
     * @param expiresIn Time until allocation expires (in seconds)
     * @return allocationId Generated allocation ID
     */
    function approveAllocation(
        address recipient,
        uint256 amount,
        address asset,
        string memory purpose,
        uint256 expiresIn
    ) external onlyRole(GOVERNANCE_ROLE) returns (bytes32) {
        return _createAllocation(recipient, amount, asset, purpose, expiresIn, false, 0, 0, 0);
    }
    
    /**
     * @notice Approves a treasury allocation with installment schedule
     * @param recipient Address to receive the allocation
     * @param totalAmount Total amount to allocate
     * @param asset Asset address (address(0) for ETH)
     * @param purpose Purpose of the allocation
     * @param expiresIn Time until allocation expires (in seconds)
     * @param installments Number of installments
     * @param installmentAmount Amount per installment
     * @param installmentInterval Time between installments (in seconds)
     * @return allocationId Generated allocation ID
     */
    function approveInstallmentAllocation(
        address recipient,
        uint256 totalAmount,
        address asset,
        string memory purpose,
        uint256 expiresIn,
        uint256 installments,
        uint256 installmentAmount,
        uint256 installmentInterval
    ) external onlyRole(GOVERNANCE_ROLE) returns (bytes32) {
        if (installments == 0 || installmentAmount == 0) revert InvalidInstallmentSchedule();
        if (installments * installmentAmount != totalAmount) revert InvalidInstallmentSchedule();
        
        return _createAllocation(
            recipient,
            totalAmount,
            asset,
            purpose,
            expiresIn,
            true,
            installments,
            installmentAmount,
            installmentInterval
        );
    }
    
    /**
     * @notice Executes an approved allocation
     * @param allocationId ID of the allocation to execute
     */
    function executeAllocation(bytes32 allocationId) external nonReentrant whenNotPaused {
        Allocation storage allocation = allocations[allocationId];
        
        if (allocation.amount == 0) revert AllocationNotFound();
        if (allocation.executed && !allocation.isBatch) revert AllocationAlreadyExecuted();
        if (block.timestamp > allocation.expiresAt) revert AllocationExpired();
        
        if (allocation.isBatch) {
            _executeInstallment(allocationId);
        } else {
            _executeSingleAllocation(allocationId);
        }
    }
    
    /**
     * @notice Executes a batch transfer via governance
     * @param transfers Array of transfer details
     */
    function executeBatchTransfer(BatchTransfer[] memory transfers) 
        external 
        onlyRole(GOVERNANCE_ROLE) 
        nonReentrant 
        whenNotPaused 
    {
        for (uint256 i = 0; i < transfers.length; i++) {
            BatchTransfer memory transfer = transfers[i];
            
            if (transfer.recipients.length != transfer.amounts.length) revert InvalidAllocation();
            
            uint256 totalAmount = 0;
            for (uint256 j = 0; j < transfer.amounts.length; j++) {
                totalAmount += transfer.amounts[j];
            }
            
            _checkTransferLimits(totalAmount);
            _checkTreasuryBalance(transfer.asset, totalAmount);
            
            // Execute transfers through treasury
            for (uint256 j = 0; j < transfer.recipients.length; j++) {
                treasury.transferFunds(transfer.recipients[j], transfer.amounts[j], transfer.asset);
            }
            
            _updateDailyTransfers(totalAmount);
            
            emit BatchTransferExecuted(
                transfer.recipients,
                transfer.amounts,
                transfer.asset,
                transfer.purpose
            );
        }
    }
    
    /**
     * @notice Emergency withdrawal function
     * @param amount Amount to withdraw
     * @param asset Asset to withdraw
     * @param reason Reason for emergency withdrawal
     */
    function emergencyWithdraw(
        uint256 amount,
        address asset,
        string memory reason
    ) external onlyRole(EMERGENCY_ROLE) nonReentrant {
        _checkTreasuryBalance(asset, amount);
        
        treasury.transferFunds(emergencyRecipient, amount, asset);
        
        emit EmergencyWithdrawal(emergencyRecipient, amount, asset, reason);
    }
    
    /**
     * @notice Updates treasury parameters via governance
     * @param parameter Parameter name
     * @param value New parameter value
     */
    function updateParameter(string memory parameter, uint256 value) 
        external 
        onlyRole(GOVERNANCE_ROLE) 
    {
        uint256 oldValue;
        
        if (keccak256(bytes(parameter)) == keccak256(bytes("maxSingleTransfer"))) {
            oldValue = maxSingleTransfer;
            maxSingleTransfer = value;
        } else if (keccak256(bytes(parameter)) == keccak256(bytes("maxDailyTransfer"))) {
            oldValue = maxDailyTransfer;
            maxDailyTransfer = value;
        } else {
            revert("Invalid parameter");
        }
        
        emit ParameterUpdated(parameter, oldValue, value);
    }
    
    /**
     * @notice Updates emergency recipient
     * @param newRecipient New emergency recipient address
     */
    function updateEmergencyRecipient(address newRecipient) 
        external 
        onlyRole(GOVERNANCE_ROLE) 
    {
        require(newRecipient != address(0), "Invalid recipient");
        address oldRecipient = emergencyRecipient;
        emergencyRecipient = newRecipient;
        emit EmergencyRecipientUpdated(oldRecipient, newRecipient);
    }
    
    /**
     * @notice Pauses the contract
     */
    function pause() external onlyRole(EMERGENCY_ROLE) {
        _pause();
    }
    
    /**
     * @notice Unpauses the contract
     */
    function unpause() external onlyRole(GOVERNANCE_ROLE) {
        _unpause();
    }
    
    /**
     * @notice Gets allocation details
     * @param allocationId Allocation ID
     * @return Allocation struct
     */
    function getAllocation(bytes32 allocationId) external view returns (Allocation memory) {
        return allocations[allocationId];
    }
    
    /**
     * @notice Checks if an allocation is ready for execution
     * @param allocationId Allocation ID
     * @return Whether allocation can be executed
     */
    function isAllocationExecutable(bytes32 allocationId) external view returns (bool) {
        Allocation memory allocation = allocations[allocationId];
        
        if (allocation.amount == 0) return false;
        if (allocation.executed && !allocation.isBatch) return false;
        if (block.timestamp > allocation.expiresAt) return false;
        
        if (allocation.isBatch) {
            return block.timestamp >= allocation.nextInstallmentDate &&
                   allocation.installmentsPaid < allocation.installments;
        }
        
        return true;
    }
    
    /**
     * @notice Gets current day identifier for daily limits
     * @return Current day as timestamp / 86400
     */
    function getCurrentDay() public view returns (uint256) {
        return block.timestamp / 86400;
    }
    
    /**
     * @notice Gets remaining daily transfer allowance
     * @return Remaining amount that can be transferred today
     */
    function getRemainingDailyAllowance() external view returns (uint256) {
        uint256 today = getCurrentDay();
        uint256 transferred = dailyTransfers[today];
        
        if (transferred >= maxDailyTransfer) return 0;
        return maxDailyTransfer - transferred;
    }
    
    // Internal functions
    
    function _createAllocation(
        address recipient,
        uint256 amount,
        address asset,
        string memory purpose,
        uint256 expiresIn,
        bool isBatch,
        uint256 installments,
        uint256 installmentAmount,
        uint256 installmentInterval
    ) internal returns (bytes32) {
        if (recipient == address(0)) revert InvalidAllocation();
        if (amount == 0) revert InvalidAllocation();
        
        bytes32 allocationId = keccak256(abi.encodePacked(
            nextAllocationId++,
            recipient,
            amount,
            asset,
            block.timestamp
        ));
        
        uint256 nextInstallmentDate = isBatch ? block.timestamp + installmentInterval : 0;
        
        allocations[allocationId] = Allocation({
            id: allocationId,
            recipient: recipient,
            amount: amount,
            asset: asset,
            purpose: purpose,
            approvedAt: block.timestamp,
            expiresAt: block.timestamp + expiresIn,
            executed: false,
            isBatch: isBatch,
            installments: installments,
            installmentAmount: installmentAmount,
            nextInstallmentDate: nextInstallmentDate,
            installmentsPaid: 0
        });
        
        emit AllocationApproved(allocationId, recipient, amount, asset, purpose);
        
        return allocationId;
    }
    
    function _executeSingleAllocation(bytes32 allocationId) internal {
        Allocation storage allocation = allocations[allocationId];
        
        _checkTransferLimits(allocation.amount);
        _checkTreasuryBalance(allocation.asset, allocation.amount);
        
        allocation.executed = true;
        
        treasury.transferFunds(allocation.recipient, allocation.amount, allocation.asset);
        _updateDailyTransfers(allocation.amount);
        
        emit AllocationExecuted(allocationId, allocation.recipient, allocation.amount, allocation.asset);
    }
    
    function _executeInstallment(bytes32 allocationId) internal {
        Allocation storage allocation = allocations[allocationId];
        
        if (block.timestamp < allocation.nextInstallmentDate) revert InstallmentNotDue();
        if (allocation.installmentsPaid >= allocation.installments) revert AllocationAlreadyExecuted();
        
        _checkTransferLimits(allocation.installmentAmount);
        _checkTreasuryBalance(allocation.asset, allocation.installmentAmount);
        
        allocation.installmentsPaid++;
        allocation.nextInstallmentDate = block.timestamp + (allocation.expiresAt - allocation.approvedAt) / allocation.installments;
        
        if (allocation.installmentsPaid >= allocation.installments) {
            allocation.executed = true;
        }
        
        treasury.transferFunds(allocation.recipient, allocation.installmentAmount, allocation.asset);
        _updateDailyTransfers(allocation.installmentAmount);
        
        emit InstallmentPaid(allocationId, allocation.recipient, allocation.installmentAmount, allocation.installmentsPaid);
    }
    
    function _checkTransferLimits(uint256 amount) internal view {
        if (amount > maxSingleTransfer) revert ExceedsMaxTransfer();
        
        uint256 today = getCurrentDay();
        uint256 dailyTotal = dailyTransfers[today] + amount;
        if (dailyTotal > maxDailyTransfer) revert ExceedsDailyLimit();
    }
    
    function _checkTreasuryBalance(address asset, uint256 amount) internal view {
        uint256 balance;
        if (asset == address(0)) {
            balance = address(treasury).balance;
        } else {
            balance = IERC20(asset).balanceOf(address(treasury));
        }
        
        if (balance < amount) revert InsufficientTreasuryBalance();
    }
    
    function _updateDailyTransfers(uint256 amount) internal {
        uint256 today = getCurrentDay();
        dailyTransfers[today] += amount;
    }
}
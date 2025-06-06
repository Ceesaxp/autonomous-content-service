// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/governance/TimelockController.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title GovernanceTimelock
 * @notice Enhanced timelock controller for DAO governance with emergency controls
 * @dev Extends OpenZeppelin TimelockController with additional features
 */
contract GovernanceTimelock is TimelockController, ReentrancyGuard {
    /// @notice Minimum delay for normal operations (24 hours)
    uint256 public constant MIN_DELAY = 24 hours;
    
    /// @notice Maximum delay for operations (30 days)
    uint256 public constant MAX_DELAY = 30 days;
    
    /// @notice Emergency delay for critical operations (2 hours)
    uint256 public constant EMERGENCY_DELAY = 2 hours;
    
    /// @notice Role for emergency operations
    bytes32 public constant EMERGENCY_ROLE = keccak256("EMERGENCY_ROLE");
    
    /// @notice Role for guardian operations (can pause/cancel)
    bytes32 public constant GUARDIAN_ROLE = keccak256("GUARDIAN_ROLE");
    
    /// @notice Mapping of operation IDs to their emergency status
    mapping(bytes32 => bool) public emergencyOperations;
    
    /// @notice Mapping of operation IDs to their metadata
    mapping(bytes32 => OperationMetadata) public operationMetadata;
    
    /// @notice Whether the timelock is in emergency mode
    bool public emergencyMode;
    
    /// @notice Timestamp when emergency mode was activated
    uint256 public emergencyModeActivatedAt;
    
    /// @notice Maximum duration for emergency mode (7 days)
    uint256 public constant MAX_EMERGENCY_DURATION = 7 days;
    
    struct OperationMetadata {
        string description;
        string category;
        uint256 estimatedValue;
        address proposer;
        bool isCritical;
        uint256 createdAt;
    }
    
    event EmergencyModeActivated(address indexed activator, string reason);
    event EmergencyModeDeactivated(address indexed deactivator);
    event EmergencyOperationScheduled(bytes32 indexed id, string description);
    event OperationCancelledByGuardian(bytes32 indexed id, address indexed guardian, string reason);
    event DelayUpdated(uint256 oldDelay, uint256 newDelay);
    
    error EmergencyModeActive();
    error EmergencyModeInactive();
    error InvalidDelay();
    error OperationNotScheduled();
    error GuardianOnly();
    error EmergencyOnly();
    error EmergencyModeExpired();
    
    /**
     * @notice Constructor to initialize the governance timelock
     * @param minDelay Minimum delay for operations
     * @param proposers List of addresses that can propose operations
     * @param executors List of addresses that can execute operations
     * @param admin Admin address (typically the DAO governance contract)
     */
    constructor(
        uint256 minDelay,
        address[] memory proposers,
        address[] memory executors,
        address admin
    ) TimelockController(minDelay, proposers, executors, admin) {
        require(minDelay >= MIN_DELAY && minDelay <= MAX_DELAY, "GovernanceTimelock: invalid delay");
        
        // Grant emergency role to admin initially
        _grantRole(EMERGENCY_ROLE, admin);
        _grantRole(GUARDIAN_ROLE, admin);
    }
    
    /**
     * @notice Schedules an operation with metadata
     * @param target Target address for the operation
     * @param value ETH value for the operation
     * @param data Calldata for the operation
     * @param predecessor Operation that must be executed before this one
     * @param salt Salt for operation ID generation
     * @param delay Delay before operation can be executed
     * @param description Human-readable description
     * @param category Category of operation
     * @param estimatedValue Estimated value impact
     * @param isCritical Whether this is a critical operation
     */
    function scheduleWithMetadata(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt,
        uint256 delay,
        string memory description,
        string memory category,
        uint256 estimatedValue,
        bool isCritical
    ) external onlyRole(PROPOSER_ROLE) {
        if (emergencyMode && !hasRole(EMERGENCY_ROLE, msg.sender)) {
            revert EmergencyModeActive();
        }
        
        bytes32 id = hashOperation(target, value, data, predecessor, salt);
        
        // Store metadata
        operationMetadata[id] = OperationMetadata({
            description: description,
            category: category,
            estimatedValue: estimatedValue,
            proposer: msg.sender,
            isCritical: isCritical,
            createdAt: block.timestamp
        });
        
        schedule(target, value, data, predecessor, salt, delay);
    }
    
    /**
     * @notice Schedules a batch operation with metadata
     * @param targets Target addresses for the operations
     * @param values ETH values for the operations
     * @param payloads Calldata for the operations
     * @param predecessor Operation that must be executed before this batch
     * @param salt Salt for operation ID generation
     * @param delay Delay before operation can be executed
     * @param description Human-readable description
     * @param category Category of operation
     * @param estimatedValue Estimated total value impact
     * @param isCritical Whether this is a critical operation
     */
    function scheduleBatchWithMetadata(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata payloads,
        bytes32 predecessor,
        bytes32 salt,
        uint256 delay,
        string memory description,
        string memory category,
        uint256 estimatedValue,
        bool isCritical
    ) external onlyRole(PROPOSER_ROLE) {
        if (emergencyMode && !hasRole(EMERGENCY_ROLE, msg.sender)) {
            revert EmergencyModeActive();
        }
        
        bytes32 id = hashOperationBatch(targets, values, payloads, predecessor, salt);
        
        // Store metadata
        operationMetadata[id] = OperationMetadata({
            description: description,
            category: category,
            estimatedValue: estimatedValue,
            proposer: msg.sender,
            isCritical: isCritical,
            createdAt: block.timestamp
        });
        
        scheduleBatch(targets, values, payloads, predecessor, salt, delay);
    }
    
    /**
     * @notice Schedules an emergency operation with reduced delay
     * @param target Target address for the operation
     * @param value ETH value for the operation
     * @param data Calldata for the operation
     * @param predecessor Operation that must be executed before this one
     * @param salt Salt for operation ID generation
     * @param description Human-readable description
     */
    function scheduleEmergency(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt,
        string memory description
    ) external onlyRole(EMERGENCY_ROLE) nonReentrant {
        bytes32 id = hashOperation(target, value, data, predecessor, salt);
        
        // Mark as emergency operation
        emergencyOperations[id] = true;
        
        // Store metadata
        operationMetadata[id] = OperationMetadata({
            description: description,
            category: "Emergency",
            estimatedValue: value,
            proposer: msg.sender,
            isCritical: true,
            createdAt: block.timestamp
        });
        
        schedule(target, value, data, predecessor, salt, EMERGENCY_DELAY);
        
        emit EmergencyOperationScheduled(id, description);
    }
    
    /**
     * @notice Activates emergency mode
     * @param reason Reason for activating emergency mode
     */
    function activateEmergencyMode(string memory reason) external onlyRole(EMERGENCY_ROLE) {
        if (emergencyMode) revert EmergencyModeActive();
        
        emergencyMode = true;
        emergencyModeActivatedAt = block.timestamp;
        
        emit EmergencyModeActivated(msg.sender, reason);
    }
    
    /**
     * @notice Deactivates emergency mode
     */
    function deactivateEmergencyMode() external onlyRole(EMERGENCY_ROLE) {
        if (!emergencyMode) revert EmergencyModeInactive();
        
        emergencyMode = false;
        emergencyModeActivatedAt = 0;
        
        emit EmergencyModeDeactivated(msg.sender);
    }
    
    /**
     * @notice Cancels an operation (guardian function)
     * @param id Operation ID to cancel
     * @param reason Reason for cancellation
     */
    function cancelByGuardian(bytes32 id, string memory reason) external onlyRole(GUARDIAN_ROLE) {
        if (!isOperationPending(id)) revert OperationNotScheduled();
        
        cancel(id);
        
        emit OperationCancelledByGuardian(id, msg.sender, reason);
    }
    
    /**
     * @notice Updates the minimum delay for operations
     * @param newDelay New minimum delay
     */
    function updateDelay(uint256 newDelay) external {
        require(msg.sender == address(this), "GovernanceTimelock: caller must be timelock");
        if (newDelay < MIN_DELAY || newDelay > MAX_DELAY) revert InvalidDelay();
        
        uint256 oldDelay = getMinDelay();
        _setMinDelay(newDelay);
        
        emit DelayUpdated(oldDelay, newDelay);
    }
    
    /**
     * @notice Returns operation metadata
     * @param id Operation ID
     * @return Operation metadata struct
     */
    function getOperationMetadata(bytes32 id) external view returns (OperationMetadata memory) {
        return operationMetadata[id];
    }
    
    /**
     * @notice Checks if an operation is an emergency operation
     * @param id Operation ID
     * @return Whether the operation is marked as emergency
     */
    function isEmergencyOperation(bytes32 id) external view returns (bool) {
        return emergencyOperations[id];
    }
    
    /**
     * @notice Checks if emergency mode is still valid (not expired)
     * @return Whether emergency mode is valid
     */
    function isEmergencyModeValid() external view returns (bool) {
        if (!emergencyMode) return false;
        return block.timestamp <= emergencyModeActivatedAt + MAX_EMERGENCY_DURATION;
    }
    
    /**
     * @notice Gets the remaining time for emergency mode
     * @return Remaining time in seconds (0 if not in emergency mode or expired)
     */
    function getEmergencyModeTimeRemaining() external view returns (uint256) {
        if (!emergencyMode) return 0;
        
        uint256 expiry = emergencyModeActivatedAt + MAX_EMERGENCY_DURATION;
        if (block.timestamp >= expiry) return 0;
        
        return expiry - block.timestamp;
    }
    
    /**
     * @notice Lists pending operations with metadata
     * @param offset Starting offset for pagination
     * @param limit Maximum number of operations to return
     * @return operations Array of operation IDs
     * @return metadata Array of corresponding metadata
     */
    function getPendingOperations(uint256 offset, uint256 limit) 
        external 
        view 
        returns (bytes32[] memory operations, OperationMetadata[] memory metadata) 
    {
        // Note: This is a simplified implementation
        // In practice, you might want to maintain a separate array of pending operations
        // for more efficient querying
        operations = new bytes32[](limit);
        metadata = new OperationMetadata[](limit);
        
        // Implementation would require maintaining a list of pending operations
        // This is left as a placeholder for the actual implementation
    }
    
    /**
     * @notice Emergency function to automatically deactivate emergency mode if expired
     */
    function checkEmergencyModeExpiry() external {
        if (emergencyMode && block.timestamp > emergencyModeActivatedAt + MAX_EMERGENCY_DURATION) {
            emergencyMode = false;
            emergencyModeActivatedAt = 0;
            emit EmergencyModeDeactivated(address(this));
        }
    }
    
    // Override execute functions to check emergency mode expiry
    
    function execute(
        address target,
        uint256 value,
        bytes calldata payload,
        bytes32 predecessor,
        bytes32 salt
    ) public payable override nonReentrant {
        bytes32 id = hashOperation(target, value, payload, predecessor, salt);
        
        // Auto-deactivate expired emergency mode
        if (emergencyMode && block.timestamp > emergencyModeActivatedAt + MAX_EMERGENCY_DURATION) {
            emergencyMode = false;
            emergencyModeActivatedAt = 0;
            emit EmergencyModeDeactivated(address(this));
        }
        
        super.execute(target, value, payload, predecessor, salt);
    }
    
    function executeBatch(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata payloads,
        bytes32 predecessor,
        bytes32 salt
    ) public payable override nonReentrant {
        // Auto-deactivate expired emergency mode
        if (emergencyMode && block.timestamp > emergencyModeActivatedAt + MAX_EMERGENCY_DURATION) {
            emergencyMode = false;
            emergencyModeActivatedAt = 0;
            emit EmergencyModeDeactivated(address(this));
        }
        
        super.executeBatch(targets, values, payloads, predecessor, salt);
    }
    
    // Internal function to set minimum delay (for updateDelay function)
    function _setMinDelay(uint256 newDelay) internal {
        // This would need to be implemented based on the specific TimelockController version
        // For now, this is a placeholder
    }
}
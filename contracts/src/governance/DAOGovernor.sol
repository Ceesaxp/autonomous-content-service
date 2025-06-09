// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorTimelockControl.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title DAOGovernor
 * @notice Main governance contract for the DAO
 * @dev Implements OpenZeppelin Governor with timelock, quorum, and custom features
 */
contract DAOGovernor is 
    Governor,
    GovernorSettings,
    GovernorCountingSimple,
    GovernorVotes,
    GovernorVotesQuorumFraction,
    GovernorTimelockControl,
    Ownable,
    ReentrancyGuard
{
    /// @notice Minimum proposal threshold (0.1% of total supply)
    uint256 public constant MIN_PROPOSAL_THRESHOLD = 1000; // 0.1% in basis points
    
    /// @notice Maximum proposal threshold (5% of total supply)
    uint256 public constant MAX_PROPOSAL_THRESHOLD = 50000; // 5% in basis points
    
    /// @notice Minimum voting period (1 day)
    uint256 public constant MIN_VOTING_PERIOD = 1 days;
    
    /// @notice Maximum voting period (2 weeks)
    uint256 public constant MAX_VOTING_PERIOD = 14 days;
    
    /// @notice Minimum voting delay (1 hour)
    uint256 public constant MIN_VOTING_DELAY = 1 hours;
    
    /// @notice Maximum voting delay (1 week)
    uint256 public constant MAX_VOTING_DELAY = 7 days;
    
    /// @notice Emergency threshold for critical proposals (10% of total supply)
    uint256 public constant EMERGENCY_THRESHOLD = 100000; // 10% in basis points
    
    /// @notice Mapping of proposal types and their specific settings
    mapping(uint8 => ProposalTypeConfig) public proposalTypeConfigs;
    
    /// @notice Mapping of proposal IDs to their metadata
    mapping(uint256 => ProposalMetadata) public proposalMetadata;
    
    /// @notice Mapping of addresses authorized to create emergency proposals
    mapping(address => bool) public emergencyProposers;
    
    /// @notice Counter for proposal types
    uint8 public nextProposalType = 1;
    
    /// @notice Whether the governor is paused
    bool public paused;
    
    struct ProposalTypeConfig {
        string name;
        uint256 quorumFraction; // In basis points
        uint256 votingPeriod;
        uint256 proposalThreshold;
        bool requiresTimelock;
        bool emergencyBypass; // Can bypass normal delays in emergencies
    }
    
    struct ProposalMetadata {
        uint8 proposalType;
        string category;
        string discussionUrl;
        bytes32 ipfsHash;
        bool isEmergency;
        uint256 estimatedExecutionCost;
    }
    
    event ProposalTypeConfigured(
        uint8 indexed proposalType,
        string name,
        uint256 quorumFraction,
        uint256 votingPeriod,
        uint256 proposalThreshold
    );
    
    event EmergencyProposerUpdated(address indexed account, bool authorized);
    event GovernorPaused(bool paused);
    event ProposalCreatedWithMetadata(
        uint256 indexed proposalId,
        uint8 proposalType,
        string category,
        string discussionUrl,
        bool isEmergency
    );
    
    error GovernorIsPaused();
    error InvalidProposalType();
    error InvalidThreshold();
    error InvalidVotingPeriod();
    error InvalidVotingDelay();
    error UnauthorizedEmergencyProposal();
    error ProposalTypeNotConfigured();
    
    /**
     * @notice Constructor to initialize the DAO Governor
     * @param _token Address of the governance token
     * @param _timelock Address of the timelock controller
     * @param _quorumFraction Initial quorum fraction (in percentage, e.g., 4 for 4%)
     * @param _votingPeriod Initial voting period in seconds
     * @param _votingDelay Initial voting delay in seconds
     * @param _proposalThreshold Initial proposal threshold in tokens
     */
    constructor(
        IVotes _token,
        TimelockController _timelock,
        uint256 _quorumFraction,
        uint256 _votingPeriod,
        uint256 _votingDelay,
        uint256 _proposalThreshold
    )
        Governor("AutonomousContentDAO")
        GovernorSettings(_votingDelay, _votingPeriod, _proposalThreshold)
        GovernorVotes(_token)
        GovernorVotesQuorumFraction(_quorumFraction)
        GovernorTimelockControl(_timelock)
    {
        // Configure default proposal types
        _configureDefaultProposalTypes();
    }
    
    /**
     * @notice Creates a proposal with metadata
     * @param targets Target addresses for proposal actions
     * @param values ETH values for proposal actions
     * @param calldatas Calldata for proposal actions
     * @param description Proposal description
     * @param proposalType Type of proposal (0 = standard)
     * @param category Category of proposal
     * @param discussionUrl URL for proposal discussion
     * @param isEmergency Whether this is an emergency proposal
     * @return Proposal ID
     */
    function proposeWithMetadata(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description,
        uint8 proposalType,
        string memory category,
        string memory discussionUrl,
        bool isEmergency
    ) public returns (uint256) {
        if (paused) revert GovernorIsPaused();
        
        // Validate proposal type
        if (proposalType > 0 && proposalTypeConfigs[proposalType].proposalThreshold == 0) {
            revert ProposalTypeNotConfigured();
        }
        
        // Check emergency proposal authorization
        if (isEmergency && !emergencyProposers[msg.sender]) {
            revert UnauthorizedEmergencyProposal();
        }
        
        // Create the proposal
        uint256 proposalId = propose(targets, values, calldatas, description);
        
        // Store metadata
        proposalMetadata[proposalId] = ProposalMetadata({
            proposalType: proposalType,
            category: category,
            discussionUrl: discussionUrl,
            ipfsHash: bytes32(0), // Can be updated later
            isEmergency: isEmergency,
            estimatedExecutionCost: _estimateExecutionCost(targets, values, calldatas)
        });
        
        emit ProposalCreatedWithMetadata(proposalId, proposalType, category, discussionUrl, isEmergency);
        
        return proposalId;
    }
    
    /**
     * @notice Configures a proposal type
     * @param proposalType Type ID (0 is reserved for standard proposals)
     * @param name Human-readable name
     * @param quorumFraction Quorum requirement in basis points
     * @param votingPeriod Voting period in seconds
     * @param proposalThreshold Proposal threshold in tokens
     * @param requiresTimelock Whether proposals require timelock
     * @param emergencyBypass Whether emergency proposals can bypass delays
     */
    function configureProposalType(
        uint8 proposalType,
        string memory name,
        uint256 quorumFraction,
        uint256 votingPeriod,
        uint256 proposalThreshold,
        bool requiresTimelock,
        bool emergencyBypass
    ) external onlyGovernance {
        if (proposalType == 0) revert InvalidProposalType();
        if (votingPeriod < MIN_VOTING_PERIOD || votingPeriod > MAX_VOTING_PERIOD) {
            revert InvalidVotingPeriod();
        }
        
        proposalTypeConfigs[proposalType] = ProposalTypeConfig({
            name: name,
            quorumFraction: quorumFraction,
            votingPeriod: votingPeriod,
            proposalThreshold: proposalThreshold,
            requiresTimelock: requiresTimelock,
            emergencyBypass: emergencyBypass
        });
        
        emit ProposalTypeConfigured(proposalType, name, quorumFraction, votingPeriod, proposalThreshold);
    }
    
    /**
     * @notice Sets emergency proposer authorization
     * @param account Address to authorize/deauthorize
     * @param authorized Whether to authorize (true) or deauthorize (false)
     */
    function setEmergencyProposer(address account, bool authorized) external onlyGovernance {
        emergencyProposers[account] = authorized;
        emit EmergencyProposerUpdated(account, authorized);
    }
    
    /**
     * @notice Pauses or unpauses the governor
     * @param _paused Whether to pause (true) or unpause (false)
     */
    function setPaused(bool _paused) external onlyGovernance {
        paused = _paused;
        emit GovernorPaused(_paused);
    }
    
    /**
     * @notice Updates proposal metadata (IPFS hash, etc.)
     * @param proposalId Proposal ID
     * @param ipfsHash IPFS hash for proposal documents
     */
    function updateProposalMetadata(uint256 proposalId, bytes32 ipfsHash) external {
        require(
            msg.sender == proposalProposer(proposalId) || msg.sender == owner(),
            "DAOGovernor: unauthorized"
        );
        proposalMetadata[proposalId].ipfsHash = ipfsHash;
    }
    
    /**
     * @notice Returns proposal metadata
     * @param proposalId Proposal ID
     * @return Proposal metadata struct
     */
    function getProposalMetadata(uint256 proposalId) external view returns (ProposalMetadata memory) {
        return proposalMetadata[proposalId];
    }
    
    /**
     * @notice Returns proposal type configuration
     * @param proposalType Type ID
     * @return Proposal type configuration struct
     */
    function getProposalTypeConfig(uint8 proposalType) external view returns (ProposalTypeConfig memory) {
        return proposalTypeConfigs[proposalType];
    }
    
    // Override functions for proposal type-specific behavior
    
    /**
     * @notice Returns the quorum for a specific proposal
     * @param proposalId Proposal ID
     * @return Quorum required for the proposal
     */
    function quorum(uint256 proposalId) public view override(IGovernor, GovernorVotesQuorumFraction) returns (uint256) {
        uint8 proposalType = proposalMetadata[proposalId].proposalType;
        
        if (proposalType > 0 && proposalTypeConfigs[proposalType].quorumFraction > 0) {
            uint256 totalSupply = token.getPastTotalSupply(proposalSnapshot(proposalId));
            return (totalSupply * proposalTypeConfigs[proposalType].quorumFraction) / 10000;
        }
        
        return super.quorum(proposalId);
    }
    
    /**
     * @notice Returns the voting period for a specific proposal
     * @param proposalId Proposal ID
     * @return Voting period for the proposal
     */
    function proposalDeadline(uint256 proposalId) public view override(IGovernor, Governor) returns (uint256) {
        uint8 proposalType = proposalMetadata[proposalId].proposalType;
        
        if (proposalType > 0 && proposalTypeConfigs[proposalType].votingPeriod > 0) {
            return proposalSnapshot(proposalId) + proposalTypeConfigs[proposalType].votingPeriod;
        }
        
        return super.proposalDeadline(proposalId);
    }
    
    /**
     * @notice Returns the proposal threshold for a specific proposal type
     * @return Current proposal threshold
     */
    function proposalThreshold() public view override(Governor, GovernorSettings) returns (uint256) {
        return super.proposalThreshold();
    }
    
    /**
     * @notice Custom proposal threshold check for different proposal types
     * @param account Account to check threshold for
     * @param proposalType Type of proposal
     * @return Whether account meets threshold
     */
    function meetsProposalThreshold(address account, uint8 proposalType) public view returns (bool) {
        uint256 threshold;
        
        if (proposalType > 0 && proposalTypeConfigs[proposalType].proposalThreshold > 0) {
            threshold = proposalTypeConfigs[proposalType].proposalThreshold;
        } else {
            threshold = proposalThreshold();
        }
        
        return getVotes(account, clock() - 1) >= threshold;
    }
    
    // Required overrides for multiple inheritance
    
    function votingDelay() public view override(IGovernor, GovernorSettings) returns (uint256) {
        return super.votingDelay();
    }
    
    function votingPeriod() public view override(IGovernor, GovernorSettings) returns (uint256) {
        return super.votingPeriod();
    }
    
    function getVotes(address account, uint256 timepoint) public view override(IGovernor, Governor) returns (uint256) {
        return super.getVotes(account, timepoint);
    }
    
    function state(uint256 proposalId) public view override(Governor, GovernorTimelockControl) returns (ProposalState) {
        return super.state(proposalId);
    }
    
    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    ) public override(Governor, IGovernor) returns (uint256) {
        if (paused) revert GovernorIsPaused();
        return super.propose(targets, values, calldatas, description);
    }
    
    function _execute(
        uint256 proposalId,
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal override(Governor, GovernorTimelockControl) {
        super._execute(proposalId, targets, values, calldatas, descriptionHash);
    }
    
    function _cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal override(Governor, GovernorTimelockControl) returns (uint256) {
        return super._cancel(targets, values, calldatas, descriptionHash);
    }
    
    function _executor() internal view override(Governor, GovernorTimelockControl) returns (address) {
        return super._executor();
    }
    
    function supportsInterface(bytes4 interfaceId) public view override(Governor, GovernorTimelockControl) returns (bool) {
        return super.supportsInterface(interfaceId);
    }
    
    // Internal helper functions
    
    /**
     * @notice Configures default proposal types
     */
    function _configureDefaultProposalTypes() internal {
        // Standard Treasury Proposal
        proposalTypeConfigs[1] = ProposalTypeConfig({
            name: "Treasury",
            quorumFraction: 400, // 4%
            votingPeriod: 5 days,
            proposalThreshold: 1000 * 10**18, // 1000 tokens
            requiresTimelock: true,
            emergencyBypass: false
        });
        
        // Parameter Change Proposal
        proposalTypeConfigs[2] = ProposalTypeConfig({
            name: "Parameter",
            quorumFraction: 300, // 3%
            votingPeriod: 3 days,
            proposalThreshold: 500 * 10**18, // 500 tokens
            requiresTimelock: true,
            emergencyBypass: false
        });
        
        // Emergency Proposal
        proposalTypeConfigs[3] = ProposalTypeConfig({
            name: "Emergency",
            quorumFraction: 600, // 6%
            votingPeriod: 1 days,
            proposalThreshold: 5000 * 10**18, // 5000 tokens
            requiresTimelock: false,
            emergencyBypass: true
        });
        
        // Upgrade Proposal
        proposalTypeConfigs[4] = ProposalTypeConfig({
            name: "Upgrade",
            quorumFraction: 800, // 8%
            votingPeriod: 7 days,
            proposalThreshold: 2000 * 10**18, // 2000 tokens
            requiresTimelock: true,
            emergencyBypass: false
        });
    }
    
    /**
     * @notice Estimates the execution cost of a proposal
     * @param targets Target addresses
     * @param values ETH values
     * @param calldatas Calldata for each action
     * @return Estimated total cost in ETH
     */
    function _estimateExecutionCost(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas
    ) internal pure returns (uint256) {
        uint256 totalValue = 0;
        for (uint256 i = 0; i < values.length; i++) {
            totalValue += values[i];
        }
        return totalValue;
    }
}
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Votes.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title GovernanceToken
 * @notice ERC20 token with voting capabilities for DAO governance
 * @dev Implements ERC20Votes for vote delegation and snapshot functionality
 */
contract GovernanceToken is ERC20, ERC20Permit, ERC20Votes, Ownable, Pausable {
    /// @notice Maximum total supply of tokens (100 million tokens)
    uint256 public constant MAX_SUPPLY = 100_000_000 * 10**18;
    
    /// @notice Minimum time between minting operations (1 week)
    uint256 public constant MIN_MINT_INTERVAL = 7 days;
    
    /// @notice Maximum percentage of total supply that can be minted per operation (1%)
    uint256 public constant MAX_MINT_PERCENTAGE = 100; // 1% in basis points (100/10000)
    
    /// @notice Timestamp of the last mint operation
    uint256 public lastMintTimestamp;
    
    /// @notice Address authorized to mint new tokens (treasury or governance)
    address public minter;
    
    /// @notice Mapping of addresses that are blocked from transfers
    mapping(address => bool) public blockedAddresses;
    
    event MinterUpdated(address indexed oldMinter, address indexed newMinter);
    event AddressBlocked(address indexed account, bool blocked);
    event TokensMinted(address indexed to, uint256 amount, string reason);
    event TokensBurned(address indexed from, uint256 amount, string reason);
    
    error ExceedsMaxSupply();
    error MintingTooFrequent();
    error ExceedsMintLimit();
    error UnauthorizedMinter();
    error AddressBlocked();
    error InvalidAddress();
    
    /**
     * @notice Constructor to initialize the governance token
     * @param _name Token name
     * @param _symbol Token symbol
     * @param _initialSupply Initial supply to mint to deployer
     * @param _minter Address authorized to mint tokens
     */
    constructor(
        string memory _name,
        string memory _symbol,
        uint256 _initialSupply,
        address _minter
    ) ERC20(_name, _symbol) ERC20Permit(_name) {
        if (_initialSupply > MAX_SUPPLY) revert ExceedsMaxSupply();
        if (_minter == address(0)) revert InvalidAddress();
        
        minter = _minter;
        lastMintTimestamp = block.timestamp;
        
        // Mint initial supply to deployer
        if (_initialSupply > 0) {
            _mint(msg.sender, _initialSupply);
        }
    }
    
    /**
     * @notice Updates the authorized minter address
     * @param _newMinter New minter address
     */
    function updateMinter(address _newMinter) external onlyOwner {
        if (_newMinter == address(0)) revert InvalidAddress();
        address oldMinter = minter;
        minter = _newMinter;
        emit MinterUpdated(oldMinter, _newMinter);
    }
    
    /**
     * @notice Mints new tokens to specified address
     * @param _to Address to receive minted tokens
     * @param _amount Amount of tokens to mint
     * @param _reason Reason for minting (for transparency)
     */
    function mint(address _to, uint256 _amount, string calldata _reason) external {
        if (msg.sender != minter) revert UnauthorizedMinter();
        if (_to == address(0)) revert InvalidAddress();
        if (totalSupply() + _amount > MAX_SUPPLY) revert ExceedsMaxSupply();
        
        // Check minting frequency and limits
        if (block.timestamp < lastMintTimestamp + MIN_MINT_INTERVAL) {
            revert MintingTooFrequent();
        }
        
        // Check mint amount doesn't exceed percentage limit
        uint256 maxMintAmount = (totalSupply() * MAX_MINT_PERCENTAGE) / 10000;
        if (_amount > maxMintAmount) revert ExceedsMintLimit();
        
        lastMintTimestamp = block.timestamp;
        _mint(_to, _amount);
        
        emit TokensMinted(_to, _amount, _reason);
    }
    
    /**
     * @notice Burns tokens from sender's balance
     * @param _amount Amount of tokens to burn
     * @param _reason Reason for burning (for transparency)
     */
    function burn(uint256 _amount, string calldata _reason) external {
        _burn(msg.sender, _amount);
        emit TokensBurned(msg.sender, _amount, _reason);
    }
    
    /**
     * @notice Burns tokens from specified address (requires allowance)
     * @param _from Address to burn tokens from
     * @param _amount Amount of tokens to burn
     * @param _reason Reason for burning (for transparency)
     */
    function burnFrom(address _from, uint256 _amount, string calldata _reason) external {
        _spendAllowance(_from, msg.sender, _amount);
        _burn(_from, _amount);
        emit TokensBurned(_from, _amount, _reason);
    }
    
    /**
     * @notice Blocks or unblocks an address from transfers
     * @param _account Address to block/unblock
     * @param _blocked Whether to block (true) or unblock (false)
     */
    function setAddressBlocked(address _account, bool _blocked) external onlyOwner {
        if (_account == address(0)) revert InvalidAddress();
        blockedAddresses[_account] = _blocked;
        emit AddressBlocked(_account, _blocked);
    }
    
    /**
     * @notice Pauses all token transfers
     */
    function pause() external onlyOwner {
        _pause();
    }
    
    /**
     * @notice Unpauses token transfers
     */
    function unpause() external onlyOwner {
        _unpause();
    }
    
    /**
     * @notice Returns the current votes balance for `account`
     * @param account Address to check voting power for
     * @return Current voting power
     */
    function getVotes(address account) public view override returns (uint256) {
        return super.getVotes(account);
    }
    
    /**
     * @notice Returns the amount of votes `account` had at the end of a past block
     * @param account Address to check historical voting power for
     * @param blockNumber Block number to check votes at
     * @return Historical voting power
     */
    function getPastVotes(address account, uint256 blockNumber) public view override returns (uint256) {
        return super.getPastVotes(account, blockNumber);
    }
    
    /**
     * @notice Returns the total supply of votes available at the end of a past block
     * @param blockNumber Block number to check total votes at
     * @return Historical total voting power
     */
    function getPastTotalSupply(uint256 blockNumber) public view override returns (uint256) {
        return super.getPastTotalSupply(blockNumber);
    }
    
    /**
     * @notice Delegates voting power to another address
     * @param delegatee Address to delegate voting power to
     */
    function delegate(address delegatee) public override {
        if (blockedAddresses[msg.sender]) revert AddressBlocked();
        super.delegate(delegatee);
    }
    
    /**
     * @notice Delegates voting power using signature
     * @param delegatee Address to delegate voting power to
     * @param nonce Nonce for replay protection
     * @param expiry Signature expiration timestamp
     * @param v ECDSA signature v component
     * @param r ECDSA signature r component
     * @param s ECDSA signature s component
     */
    function delegateBySig(
        address delegatee,
        uint256 nonce,
        uint256 expiry,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public override {
        super.delegateBySig(delegatee, nonce, expiry, v, r, s);
    }
    
    // Required overrides for multiple inheritance
    
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 amount
    ) internal override whenNotPaused {
        if (blockedAddresses[from] || blockedAddresses[to]) {
            revert AddressBlocked();
        }
        super._beforeTokenTransfer(from, to, amount);
    }
    
    function _afterTokenTransfer(
        address from,
        address to,
        uint256 amount
    ) internal override(ERC20, ERC20Votes) {
        super._afterTokenTransfer(from, to, amount);
    }
    
    function _mint(address to, uint256 amount) internal override(ERC20, ERC20Votes) {
        super._mint(to, amount);
    }
    
    function _burn(address account, uint256 amount) internal override(ERC20, ERC20Votes) {
        super._burn(account, amount);
    }
    
    /**
     * @notice Returns the current block timestamp for voting checkpoints
     * @return Current block timestamp
     */
    function clock() public view override returns (uint48) {
        return uint48(block.timestamp);
    }
    
    /**
     * @notice Returns the clock mode for EIP-6372 compliance
     * @return Clock mode string
     */
    function CLOCK_MODE() public pure override returns (string memory) {
        return "mode=timestamp";
    }
}
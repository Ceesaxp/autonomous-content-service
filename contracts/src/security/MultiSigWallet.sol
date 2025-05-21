// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

library SafeMath {
    function add(uint256 a, uint256 b) internal pure returns (uint256 c) {
        c = a + b;
        assert(c >= a);
        return c;
    }

    function sub(uint256 a, uint256 b) internal pure returns (uint256 c) {
        c = a - b;
        assert(c <= a);
        return c;
    }
}

/**
 * @title MultiSigWallet
 * @dev Multi-signature wallet implementation for secure treasury operations
 */
contract MultiSigWallet {
    using SafeMath for uint256;

    struct Transaction {
        address destination;
        uint256 value;
        bytes data;
        bool executed;
        uint256 confirmations;
        mapping(address => bool) isConfirmed;
    }

    mapping(uint256 => Transaction) public transactions;
    mapping(address => bool) public isOwner;
    address[] public owners;
    uint256 public required;
    uint256 public transactionCount;

    // Tiered approval thresholds
    mapping(uint256 => uint256) public approvalThresholds;
    uint256[] public thresholdAmounts;

    event Deposit(address indexed sender, uint256 value);
    event SubmitTransaction(
        address indexed owner,
        uint256 indexed txIndex,
        address indexed to,
        uint256 value,
        bytes data
    );
    event ConfirmTransaction(address indexed owner, uint256 indexed txIndex);
    event RevokeConfirmation(address indexed owner, uint256 indexed txIndex);
    event ExecuteTransaction(address indexed owner, uint256 indexed txIndex);
    event OwnerAddition(address indexed owner);
    event OwnerRemoval(address indexed owner);
    event RequirementChange(uint256 required);

    modifier onlyWallet() {
        require(msg.sender == address(this), 'Only wallet can call');
        _;
    }

    modifier ownerDoesNotExist(address owner) {
        require(!isOwner[owner], 'Owner already exists');
        _;
    }

    modifier ownerExists(address owner) {
        require(isOwner[owner], 'Owner does not exist');
        _;
    }

    modifier transactionExists(uint256 transactionId) {
        require(
            transactions[transactionId].destination != address(0),
            'Transaction does not exist'
        );
        _;
    }

    modifier confirmed(uint256 transactionId, address owner) {
        require(
            transactions[transactionId].isConfirmed[owner],
            'Transaction not confirmed'
        );
        _;
    }

    modifier notConfirmed(uint256 transactionId, address owner) {
        require(
            !transactions[transactionId].isConfirmed[owner],
            'Transaction already confirmed'
        );
        _;
    }

    modifier notExecuted(uint256 transactionId) {
        require(
            !transactions[transactionId].executed,
            'Transaction already executed'
        );
        _;
    }

    modifier notNull(address _address) {
        require(_address != address(0), 'Address cannot be null');
        _;
    }

    modifier validRequirement(uint256 ownerCount, uint256 _required) {
        require(
            ownerCount <= 20 &&
                _required <= ownerCount &&
                _required != 0 &&
                ownerCount != 0,
            'Invalid requirement'
        );
        _;
    }

    constructor(
        address[] memory _owners,
        uint256 _required
    ) validRequirement(_owners.length, _required) {
        for (uint256 i = 0; i < _owners.length; i++) {
            require(
                !isOwner[_owners[i]] && _owners[i] != address(0),
                'Invalid owner'
            );
            isOwner[_owners[i]] = true;
        }
        owners = _owners;
        required = _required;

        // Set tiered approval thresholds
        thresholdAmounts = [
            1000 * 10 ** 18,
            10000 * 10 ** 18,
            100000 * 10 ** 18
        ]; // $1K, $10K, $100K
        approvalThresholds[1000 * 10 ** 18] = 2; // $1K requires 2 signatures
        approvalThresholds[10000 * 10 ** 18] = 3; // $10K requires 3 signatures
        approvalThresholds[100000 * 10 ** 18] = 4; // $100K requires 4 signatures
    }

    receive() external payable {
        if (msg.value > 0) {
            emit Deposit(msg.sender, msg.value);
        }
    }

    function submitTransaction(
        address destination,
        uint256 value,
        bytes memory data
    )
        public
        ownerExists(msg.sender)
        notNull(destination)
        returns (uint256 transactionId)
    {
        transactionId = addTransaction(destination, value, data);
        confirmTransaction(transactionId);
    }

    function confirmTransaction(
        uint256 transactionId
    )
        public
        ownerExists(msg.sender)
        transactionExists(transactionId)
        notConfirmed(transactionId, msg.sender)
    {
        transactions[transactionId].isConfirmed[msg.sender] = true;
        transactions[transactionId].confirmations = transactions[transactionId]
            .confirmations
            .add(1);

        emit ConfirmTransaction(msg.sender, transactionId);

        executeTransaction(transactionId);
    }

    function revokeConfirmation(
        uint256 transactionId
    )
        public
        ownerExists(msg.sender)
        confirmed(transactionId, msg.sender)
        notExecuted(transactionId)
    {
        transactions[transactionId].isConfirmed[msg.sender] = false;
        transactions[transactionId].confirmations = transactions[transactionId]
            .confirmations
            .sub(1);
        emit RevokeConfirmation(msg.sender, transactionId);
    }

    function executeTransaction(
        uint256 transactionId
    )
        public
        ownerExists(msg.sender)
        confirmed(transactionId, msg.sender)
        notExecuted(transactionId)
    {
        uint256 requiredSigs = getRequiredSignatures(
            transactions[transactionId].value
        );

        if (transactions[transactionId].confirmations >= requiredSigs) {
            transactions[transactionId].executed = true;

            (bool success, ) = transactions[transactionId].destination.call{
                value: transactions[transactionId].value
            }(transactions[transactionId].data);

            if (success) {
                emit ExecuteTransaction(msg.sender, transactionId);
            } else {
                transactions[transactionId].executed = false;
            }
        }
    }

    function getRequiredSignatures(
        uint256 value
    ) public view returns (uint256) {
        for (uint256 i = thresholdAmounts.length; i > 0; i--) {
            if (value >= thresholdAmounts[i - 1]) {
                return approvalThresholds[thresholdAmounts[i - 1]];
            }
        }
        return required; // Default requirement for smaller amounts
    }

    function addTransaction(
        address destination,
        uint256 value,
        bytes memory data
    ) internal notNull(destination) returns (uint256 transactionId) {
        transactionId = transactionCount;
        transactions[transactionId].destination = destination;
        transactions[transactionId].value = value;
        transactions[transactionId].data = data;
        transactions[transactionId].executed = false;
        transactions[transactionId].confirmations = 0;
        transactionCount = transactionCount.add(1);
        emit SubmitTransaction(
            msg.sender,
            transactionId,
            destination,
            value,
            data
        );
    }

    function getConfirmationCount(
        uint256 transactionId
    ) public view returns (uint256 count) {
        for (uint256 i = 0; i < owners.length; i++) {
            if (transactions[transactionId].isConfirmed[owners[i]]) {
                count = count.add(1);
            }
        }
    }

    function getTransactionCount(
        bool pending,
        bool executed
    ) public view returns (uint256 count) {
        for (uint256 i = 0; i < transactionCount; i++) {
            if (
                (pending && !transactions[i].executed) ||
                (executed && transactions[i].executed)
            ) {
                count = count.add(1);
            }
        }
    }

    function getOwners() public view returns (address[] memory) {
        return owners;
    }

    function getConfirmations(
        uint256 transactionId
    ) public view returns (address[] memory _confirmations) {
        address[] memory confirmationsTemp = new address[](owners.length);
        uint256 count = 0;
        uint256 i;

        for (i = 0; i < owners.length; i++) {
            if (transactions[transactionId].isConfirmed[owners[i]]) {
                confirmationsTemp[count] = owners[i];
                count = count.add(1);
            }
        }

        _confirmations = new address[](count);
        for (i = 0; i < count; i++) {
            _confirmations[i] = confirmationsTemp[i];
        }
    }

    function getTransactionIds(
        uint256 from,
        uint256 to,
        bool pending,
        bool executed
    ) public view returns (uint256[] memory _transactionIds) {
        uint256[] memory transactionIdsTemp = new uint256[](transactionCount);
        uint256 count = 0;
        uint256 i;

        for (i = 0; i < transactionCount; i++) {
            if (
                (pending && !transactions[i].executed) ||
                (executed && transactions[i].executed)
            ) {
                transactionIdsTemp[count] = i;
                count = count.add(1);
            }
        }

        _transactionIds = new uint256[](to - from);
        for (i = from; i < to; i++) {
            _transactionIds[i - from] = transactionIdsTemp[i];
        }
    }

    // Owner management functions (can only be called by the wallet itself)
    function addOwner(
        address owner
    )
        public
        onlyWallet
        ownerDoesNotExist(owner)
        notNull(owner)
        validRequirement(owners.length + 1, required)
    {
        isOwner[owner] = true;
        owners.push(owner);
        emit OwnerAddition(owner);
    }

    function removeOwner(address owner) public onlyWallet ownerExists(owner) {
        isOwner[owner] = false;
        for (uint256 i = 0; i < owners.length - 1; i++) {
            if (owners[i] == owner) {
                owners[i] = owners[owners.length - 1];
                break;
            }
        }
        owners.pop();

        if (required > owners.length) {
            changeRequirement(owners.length);
        }
        emit OwnerRemoval(owner);
    }

    function replaceOwner(
        address owner,
        address newOwner
    )
        public
        onlyWallet
        ownerExists(owner)
        ownerDoesNotExist(newOwner)
        notNull(newOwner)
    {
        for (uint256 i = 0; i < owners.length; i++) {
            if (owners[i] == owner) {
                owners[i] = newOwner;
                break;
            }
        }
        isOwner[owner] = false;
        isOwner[newOwner] = true;
        emit OwnerRemoval(owner);
        emit OwnerAddition(newOwner);
    }

    function changeRequirement(
        uint256 _required
    ) public onlyWallet validRequirement(owners.length, _required) {
        required = _required;
        emit RequirementChange(_required);
    }

    function updateApprovalThreshold(
        uint256 amount,
        uint256 requiredSigs
    ) public onlyWallet {
        approvalThresholds[amount] = requiredSigs;
    }
}

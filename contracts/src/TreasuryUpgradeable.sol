// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./TreasuryCore.sol";

/**
 * @title TreasuryUpgradeable
 * @dev Upgradeable proxy pattern implementation for treasury
 */
contract TreasuryUpgradeable {
    bytes32 internal constant _IMPLEMENTATION_SLOT = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc;
    bytes32 internal constant _ADMIN_SLOT = 0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103;

    event AdminChanged(address previousAdmin, address newAdmin);
    event Upgraded(address indexed implementation);

    modifier ifAdmin() {
        if (msg.sender == _getAdmin()) {
            _;
        } else {
            _fallback();
        }
    }

    constructor(address _logic, address _admin, bytes memory _data) payable {
        assert(_IMPLEMENTATION_SLOT == bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1));
        assert(_ADMIN_SLOT == bytes32(uint256(keccak256("eip1967.proxy.admin")) - 1));
        
        _changeAdmin(_admin);
        _upgrade(_logic);
        
        if(_data.length > 0) {
            (bool success,) = _logic.delegatecall(_data);
            require(success);
        }
    }

    fallback() external payable virtual {
        _fallback();
    }

    receive() external payable virtual {
        _fallback();
    }

    function admin() external ifAdmin returns (address admin_) {
        admin_ = _getAdmin();
    }

    function implementation() external ifAdmin returns (address implementation_) {
        implementation_ = _getImplementation();
    }

    function changeAdmin(address newAdmin) external virtual ifAdmin {
        _changeAdmin(newAdmin);
    }

    function upgradeTo(address newImplementation) external ifAdmin {
        _upgrade(newImplementation);
    }

    function upgradeToAndCall(address newImplementation, bytes calldata data) external payable ifAdmin {
        _upgrade(newImplementation);
        (bool success,) = newImplementation.delegatecall(data);
        require(success);
    }

    function _getAdmin() internal view returns (address) {
        return StorageSlot.getAddressSlot(_ADMIN_SLOT).value;
    }

    function _setAdmin(address newAdmin) private {
        StorageSlot.getAddressSlot(_ADMIN_SLOT).value = newAdmin;
    }

    function _changeAdmin(address newAdmin) internal {
        emit AdminChanged(_getAdmin(), newAdmin);
        _setAdmin(newAdmin);
    }

    function _getImplementation() internal view returns (address) {
        return StorageSlot.getAddressSlot(_IMPLEMENTATION_SLOT).value;
    }

    function _setImplementation(address newImplementation) private {
        StorageSlot.getAddressSlot(_IMPLEMENTATION_SLOT).value = newImplementation;
    }

    function _upgrade(address newImplementation) internal {
        _setImplementation(newImplementation);
        emit Upgraded(newImplementation);
    }

    function _fallback() internal virtual {
        _delegate(_getImplementation());
    }

    function _delegate(address implementation) internal virtual {
        assembly {
            calldatacopy(0, 0, calldatasize())
            let result := delegatecall(gas(), implementation, 0, calldatasize(), 0, 0)
            returndatacopy(0, 0, returndatasize())
            
            switch result
            case 0 {
                revert(0, returndatasize())
            }
            default {
                return(0, returndatasize())
            }
        }
    }
}

library StorageSlot {
    struct AddressSlot {
        address value;
    }

    function getAddressSlot(bytes32 slot) internal pure returns (AddressSlot storage r) {
        assembly {
            r.slot := slot
        }
    }
}
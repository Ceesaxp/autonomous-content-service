// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title ITreasury
 * @dev Interface for treasury operations
 */
interface ITreasury {
    function receiveRevenue(address token, uint256 amount, string calldata description) external payable;
    function spendOperational(address token, uint256 amount, address recipient, string calldata description) external;
    function spendUpgrades(address token, uint256 amount, address recipient, string calldata description) external;
    function rebalancePortfolio() external;
    function getFinancialSummary() external view returns (uint256, uint256, uint256, uint256);
}

/**
 * @title IERC20
 * @dev Standard ERC20 interface
 */
interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
    function totalSupply() external view returns (uint256);
    function allowance(address owner, address spender) external view returns (uint256);
    function approve(address spender, uint256 amount) external returns (bool);
}

/**
 * @title SafeMath
 * @dev Math operations with safety checks
 */
library SafeMath {
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a + b;
        require(c >= a, "SafeMath: addition overflow");
        return c;
    }

    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b <= a, "SafeMath: subtraction overflow");
        return a - b;
    }

    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        if (a == 0) return 0;
        uint256 c = a * b;
        require(c / a == b, "SafeMath: multiplication overflow");
        return c;
    }

    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b > 0, "SafeMath: division by zero");
        return a / b;
    }
}
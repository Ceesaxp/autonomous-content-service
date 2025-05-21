require("@nomiclabs/hardhat-waffle");
require("@nomiclabs/hardhat-ethers");
require("@openzeppelin/hardhat-upgrades");
require("hardhat-gas-reporter");
require("solidity-coverage");
require("dotenv").config();

const PRIVATE_KEY = process.env.PRIVATE_KEY || "0x0000000000000000000000000000000000000000000000000000000000000000";
const INFURA_PROJECT_ID = process.env.INFURA_PROJECT_ID || "";
const ETHERSCAN_API_KEY = process.env.ETHERSCAN_API_KEY || "";

module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    hardhat: {
      chainId: 31337,
      accounts: {
        count: 10,
        accountsBalance: "1000000000000000000000000" // 1M ETH
      }
    },
    localhost: {
      url: "http://127.0.0.1:8545",
      chainId: 31337
    },
    goerli: {
      url: `https://goerli.infura.io/v3/${INFURA_PROJECT_ID}`,
      accounts: [PRIVATE_KEY],
      chainId: 5,
      gasPrice: 20000000000, // 20 gwei
    },
    mainnet: {
      url: `https://mainnet.infura.io/v3/${INFURA_PROJECT_ID}`,
      accounts: [PRIVATE_KEY],
      chainId: 1,
      gasPrice: 30000000000, // 30 gwei
    },
    polygon: {
      url: "https://polygon-rpc.com/",
      accounts: [PRIVATE_KEY],
      chainId: 137,
      gasPrice: 30000000000, // 30 gwei
    },
    mumbai: {
      url: "https://rpc-mumbai.maticvigil.com/",
      accounts: [PRIVATE_KEY],
      chainId: 80001,
      gasPrice: 30000000000, // 30 gwei
    }
  },
  gasReporter: {
    enabled: process.env.REPORT_GAS !== undefined,
    currency: "USD",
    coinmarketcap: process.env.COINMARKETCAP_API_KEY
  },
  etherscan: {
    apiKey: ETHERSCAN_API_KEY
  },
  paths: {
    sources: "./src",
    tests: "./test",
    cache: "./cache",
    artifacts: "./artifacts"
  },
  mocha: {
    timeout: 40000
  }
};
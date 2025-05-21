const { ethers, upgrades } = require("hardhat");

async function main() {
    console.log("Starting Treasury System Deployment...");
    
    const [deployer, owner1, owner2, owner3] = await ethers.getSigners();
    console.log("Deploying contracts with account:", deployer.address);
    console.log("Account balance:", (await deployer.getBalance()).toString());

    // Deployment configuration
    const config = {
        multisigOwners: [deployer.address, owner1.address, owner2.address, owner3.address],
        requiredSignatures: 3,
        treasurerRole: deployer.address,
        auditorRole: owner1.address,
        emergencyRole: owner2.address,
        assetManagerRole: deployer.address,
        oracleRole: deployer.address,
    };

    console.log("\n=== Deploying Core Treasury System ===");

    // Deploy TreasuryCore implementation
    const TreasuryCore = await ethers.getContractFactory("TreasuryCore");
    console.log("Deploying TreasuryCore implementation...");
    
    const treasuryImpl = await TreasuryCore.deploy(
        config.multisigOwners,
        config.requiredSignatures
    );
    await treasuryImpl.deployed();
    console.log("TreasuryCore implementation deployed to:", treasuryImpl.address);

    // Deploy upgradeable proxy
    const TreasuryUpgradeable = await ethers.getContractFactory("TreasuryUpgradeable");
    console.log("Deploying upgradeable proxy...");
    
    const initData = TreasuryCore.interface.encodeFunctionData("initialize", []);
    const treasuryProxy = await TreasuryUpgradeable.deploy(
        treasuryImpl.address,
        deployer.address,
        initData
    );
    await treasuryProxy.deployed();
    console.log("Treasury proxy deployed to:", treasuryProxy.address);

    // Get treasury instance through proxy
    const treasury = TreasuryCore.attach(treasuryProxy.address);

    // Deploy AssetManager
    console.log("\n=== Deploying Asset Management System ===");
    const AssetManager = await ethers.getContractFactory("AssetManager");
    const assetManager = await AssetManager.deploy();
    await assetManager.deployed();
    console.log("AssetManager deployed to:", assetManager.address);

    console.log("\n=== Setting up Roles and Permissions ===");

    // Grant roles
    const TREASURER_ROLE = await treasury.TREASURER_ROLE();
    const AUDITOR_ROLE = await treasury.AUDITOR_ROLE();
    const EMERGENCY_ROLE = await treasury.EMERGENCY_ROLE();
    
    await treasury.grantRole(TREASURER_ROLE, config.treasurerRole);
    console.log("Granted TREASURER_ROLE to:", config.treasurerRole);
    
    await treasury.grantRole(AUDITOR_ROLE, config.auditorRole);
    console.log("Granted AUDITOR_ROLE to:", config.auditorRole);
    
    await treasury.grantRole(EMERGENCY_ROLE, config.emergencyRole);
    console.log("Granted EMERGENCY_ROLE to:", config.emergencyRole);

    // Setup AssetManager roles
    const ASSET_MANAGER_ROLE = await assetManager.ASSET_MANAGER_ROLE();
    const ORACLE_ROLE = await assetManager.ORACLE_ROLE();
    
    await assetManager.grantRole(ASSET_MANAGER_ROLE, config.assetManagerRole);
    console.log("Granted ASSET_MANAGER_ROLE to:", config.assetManagerRole);
    
    await assetManager.grantRole(ORACLE_ROLE, config.oracleRole);
    console.log("Granted ORACLE_ROLE to:", config.oracleRole);

    console.log("\n=== Configuring Initial Asset Allocations ===");

    // Configure revenue allocation (40% ops, 20% reserves, 20% upgrades, 20% profits)
    const allocationConfig = {
        operations: 4000,
        reserves: 2000,
        upgrades: 2000,
        profits: 2000
    };

    await treasury.updateAllocationConfig(allocationConfig);
    console.log("Updated revenue allocation configuration");

    // Configure stable assets (example addresses - replace with actual token addresses)
    const stablecoins = [
        {
            address: "0xA0b86a33E6417e8Ad7b68e77A88bF7Fa30E5e8C3", // Example USDC
            name: "USDC",
            targetPercentage: 6000, // 60%
            rebalanceThreshold: 500,  // 5%
            isStablecoin: true
        },
        {
            address: "0x6B175474E89094C44Da98b954EedeAC495271d0F", // Example DAI
            name: "DAI", 
            targetPercentage: 2000, // 20%
            rebalanceThreshold: 500,  // 5%
            isStablecoin: true
        }
    ];

    for (const stable of stablecoins) {
        await treasury.configureAsset(
            stable.address,
            stable.targetPercentage,
            stable.rebalanceThreshold,
            stable.isStablecoin
        );
        console.log(`Configured ${stable.name} asset allocation: ${stable.targetPercentage/100}%`);

        await assetManager.configureAsset(
            stable.address,
            stable.targetPercentage,
            stable.rebalanceThreshold,
            ethers.utils.parseEther("1000") // Min $1000 rebalance
        );
        console.log(`Configured ${stable.name} rebalancing parameters`);
    }

    // Configure ETH allocation
    const ethConfig = {
        address: ethers.constants.AddressZero, // ETH
        targetPercentage: 2000, // 20%
        rebalanceThreshold: 1000, // 10%
        isStablecoin: false
    };

    await treasury.configureAsset(
        ethConfig.address,
        ethConfig.targetPercentage,
        ethConfig.rebalanceThreshold,
        ethConfig.isStablecoin
    );
    console.log("Configured ETH asset allocation: 20%");

    await assetManager.configureAsset(
        ethConfig.address,
        ethConfig.targetPercentage,
        ethConfig.rebalanceThreshold,
        ethers.utils.parseEther("5") // Min 5 ETH rebalance
    );
    console.log("Configured ETH rebalancing parameters");

    console.log("\n=== Deployment Summary ===");
    console.log("Treasury Implementation:", treasuryImpl.address);
    console.log("Treasury Proxy:", treasuryProxy.address);
    console.log("Asset Manager:", assetManager.address);
    console.log("Multisig Owners:", config.multisigOwners);
    console.log("Required Signatures:", config.requiredSignatures);

    // Save deployment addresses
    const deployment = {
        network: "hardhat",
        treasuryImplementation: treasuryImpl.address,
        treasuryProxy: treasuryProxy.address,
        assetManager: assetManager.address,
        deployer: deployer.address,
        config: config,
        timestamp: new Date().toISOString()
    };

    console.log("\n=== Verifying Deployment ===");
    
    // Verify treasury setup
    const summary = await treasury.getFinancialSummary();
    console.log("Initial financial summary:", {
        totalRevenue: summary[0].toString(),
        totalExpenses: summary[1].toString(),
        reserveBalance: summary[2].toString(),
        profitBalance: summary[3].toString()
    });

    // Verify asset configuration
    const allocation = await treasury.getAssetAllocation();
    console.log("Initial asset allocation:", {
        tokens: allocation[0],
        balances: allocation[1].map(b => b.toString()),
        percentages: allocation[2].map(p => p.toString())
    });

    console.log("\n✅ Treasury System Deployment Complete!");
    
    return deployment;
}

main()
    .then((deployment) => {
        console.log("\nDeployment successful!");
        console.log("Save this deployment info:", JSON.stringify(deployment, null, 2));
        process.exit(0);
    })
    .catch((error) => {
        console.error("Deployment failed:", error);
        process.exit(1);
    });
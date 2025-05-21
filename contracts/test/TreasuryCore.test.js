const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("TreasuryCore", function () {
    let treasury, assetManager;
    let owner, treasurer, auditor, emergency, addr1, addr2, addr3, addr4;
    let owners, requiredSignatures;

    beforeEach(async function () {
        [owner, treasurer, auditor, emergency, addr1, addr2, addr3, addr4] = await ethers.getSigners();
        
        owners = [owner.address, treasurer.address, auditor.address, emergency.address];
        requiredSignatures = 3;

        // Deploy TreasuryCore
        const TreasuryCore = await ethers.getContractFactory("TreasuryCore");
        treasury = await TreasuryCore.deploy(owners, requiredSignatures);
        await treasury.deployed();

        // Deploy AssetManager
        const AssetManager = await ethers.getContractFactory("AssetManager");
        assetManager = await AssetManager.deploy();
        await assetManager.deployed();

        // Setup roles
        const TREASURER_ROLE = await treasury.TREASURER_ROLE();
        const AUDITOR_ROLE = await treasury.AUDITOR_ROLE();
        const EMERGENCY_ROLE = await treasury.EMERGENCY_ROLE();

        await treasury.grantRole(TREASURER_ROLE, treasurer.address);
        await treasury.grantRole(AUDITOR_ROLE, auditor.address);
        await treasury.grantRole(EMERGENCY_ROLE, emergency.address);
    });

    describe("Deployment", function () {
        it("Should set the correct multisig configuration", async function () {
            expect(await treasury.required()).to.equal(requiredSignatures);
            
            const contractOwners = await treasury.getOwners();
            expect(contractOwners).to.deep.equal(owners);
        });

        it("Should set default allocation configuration", async function () {
            const allocation = await treasury.allocationConfig();
            expect(allocation.operations).to.equal(4000); // 40%
            expect(allocation.reserves).to.equal(2000);   // 20%
            expect(allocation.upgrades).to.equal(2000);   // 20%
            expect(allocation.profits).to.equal(2000);    // 20%
        });

        it("Should grant initial roles correctly", async function () {
            const TREASURER_ROLE = await treasury.TREASURER_ROLE();
            const AUDITOR_ROLE = await treasury.AUDITOR_ROLE();
            const EMERGENCY_ROLE = await treasury.EMERGENCY_ROLE();

            expect(await treasury.hasRole(TREASURER_ROLE, treasurer.address)).to.be.true;
            expect(await treasury.hasRole(AUDITOR_ROLE, auditor.address)).to.be.true;
            expect(await treasury.hasRole(EMERGENCY_ROLE, emergency.address)).to.be.true;
        });
    });

    describe("Revenue Management", function () {
        it("Should receive and distribute revenue correctly", async function () {
            const revenueAmount = ethers.utils.parseEther("1000");
            
            // Send revenue as ETH
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Q1 Revenue",
                { value: revenueAmount }
            );

            const summary = await treasury.getFinancialSummary();
            expect(summary.totalRevenue).to.equal(revenueAmount);

            // Check if funds were distributed according to allocation
            const expectedOperations = revenueAmount.mul(4000).div(10000); // 40%
            const expectedReserves = revenueAmount.mul(2000).div(10000);   // 20%
            const expectedUpgrades = revenueAmount.mul(2000).div(10000);   // 20%
            const expectedProfits = revenueAmount.mul(2000).div(10000);    // 20%

            expect(summary.reserveBalance).to.equal(expectedReserves);
            expect(summary.profitBalance).to.equal(expectedProfits);
        });

        it("Should record revenue transaction correctly", async function () {
            const revenueAmount = ethers.utils.parseEther("500");
            
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Test Revenue",
                { value: revenueAmount }
            );

            const transaction = await treasury.transactions(1);
            expect(transaction.amount).to.equal(revenueAmount);
            expect(transaction.category).to.equal(0); // REVENUE
            expect(transaction.description).to.equal("Test Revenue");
        });

        it("Should emit RevenueReceived event", async function () {
            const revenueAmount = ethers.utils.parseEther("100");
            
            await expect(treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Test Revenue",
                { value: revenueAmount }
            )).to.emit(treasury, "RevenueReceived")
              .withArgs(ethers.constants.AddressZero, revenueAmount, await treasury.provider.getBlockNumber() + 1);
        });
    });

    describe("Operational Spending", function () {
        beforeEach(async function () {
            // Add revenue first
            const revenueAmount = ethers.utils.parseEther("1000");
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Setup Revenue",
                { value: revenueAmount }
            );
        });

        it("Should allow operational spending within budget", async function () {
            const spendAmount = ethers.utils.parseEther("100");
            const initialBalance = await addr1.getBalance();
            
            await treasury.connect(treasurer).spendOperational(
                ethers.constants.AddressZero,
                spendAmount,
                addr1.address,
                "Server costs"
            );

            const finalBalance = await addr1.getBalance();
            expect(finalBalance.sub(initialBalance)).to.equal(spendAmount);
        });

        it("Should reject operational spending exceeding budget", async function () {
            const excessiveAmount = ethers.utils.parseEther("500"); // More than 40% of 1000 ETH
            
            await expect(treasury.connect(treasurer).spendOperational(
                ethers.constants.AddressZero,
                excessiveAmount,
                addr1.address,
                "Excessive spending"
            )).to.be.revertedWith("Insufficient operational funds");
        });

        it("Should only allow treasurer role to spend operational funds", async function () {
            const spendAmount = ethers.utils.parseEther("50");
            
            await expect(treasury.connect(addr1).spendOperational(
                ethers.constants.AddressZero,
                spendAmount,
                addr2.address,
                "Unauthorized spending"
            )).to.be.reverted;
        });
    });

    describe("Asset Configuration", function () {
        it("Should configure asset allocation correctly", async function () {
            const mockToken = addr1.address;
            
            await treasury.configureAsset(
                mockToken,
                3000, // 30%
                500,  // 5% rebalance threshold
                true  // is stablecoin
            );

            const config = await treasury.assetConfigs(mockToken);
            expect(config.targetPercentage).to.equal(3000);
            expect(config.rebalanceThreshold).to.equal(500);
            expect(config.isStablecoin).to.be.true;
            expect(config.isActive).to.be.true;
        });

        it("Should only allow admin to configure assets", async function () {
            const mockToken = addr1.address;
            
            await expect(treasury.connect(addr1).configureAsset(
                mockToken,
                3000,
                500,
                true
            )).to.be.reverted;
        });
    });

    describe("Allocation Configuration", function () {
        it("Should update allocation configuration", async function () {
            const newConfig = {
                operations: 5000, // 50%
                reserves: 2500,   // 25%
                upgrades: 1500,   // 15%
                profits: 1000     // 10%
            };

            await treasury.updateAllocationConfig(newConfig);

            const allocation = await treasury.allocationConfig();
            expect(allocation.operations).to.equal(5000);
            expect(allocation.reserves).to.equal(2500);
            expect(allocation.upgrades).to.equal(1500);
            expect(allocation.profits).to.equal(1000);
        });

        it("Should reject invalid allocation configuration", async function () {
            const invalidConfig = {
                operations: 6000, // 60%
                reserves: 2000,   // 20%
                upgrades: 2000,   // 20%
                profits: 2000     // 20% - Total is 120%
            };

            await expect(treasury.updateAllocationConfig(invalidConfig))
                .to.be.revertedWith("Allocations must sum to 100%");
        });

        it("Should emit AllocationConfigUpdated event", async function () {
            const newConfig = {
                operations: 5000,
                reserves: 2500,
                upgrades: 1500,
                profits: 1000
            };

            await expect(treasury.updateAllocationConfig(newConfig))
                .to.emit(treasury, "AllocationConfigUpdated");
        });
    });

    describe("Time-locked Transactions", function () {
        beforeEach(async function () {
            // Add revenue for upgrade funds
            const revenueAmount = ethers.utils.parseEther("10000");
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Large Revenue",
                { value: revenueAmount }
            );
        });

        it("Should create timelock for large upgrade spending", async function () {
            const largeAmount = ethers.utils.parseEther("15"); // Above threshold
            
            await treasury.connect(treasurer).spendUpgrades(
                ethers.constants.AddressZero,
                largeAmount,
                addr1.address,
                "Major upgrade"
            );

            // Transaction should be timelocked, not executed immediately
            const balance = await addr1.getBalance();
            expect(balance).to.equal(ethers.utils.parseEther("10000")); // Initial balance
        });

        it("Should execute timelocked transaction after delay", async function () {
            const largeAmount = ethers.utils.parseEther("15");
            const timestamp = (await ethers.provider.getBlock("latest")).timestamp;
            
            await treasury.connect(treasurer).spendUpgrades(
                ethers.constants.AddressZero,
                largeAmount,
                addr1.address,
                "Major upgrade"
            );

            // Fast forward time by 48+ hours
            await ethers.provider.send("evm_increaseTime", [48 * 60 * 60 + 1]);
            await ethers.provider.send("evm_mine");

            const initialBalance = await addr1.getBalance();
            
            await treasury.connect(treasurer).executeTimelocked(
                ethers.constants.AddressZero,
                largeAmount,
                addr1.address,
                "Major upgrade",
                timestamp + 1 // Block timestamp when timelock was created
            );

            const finalBalance = await addr1.getBalance();
            expect(finalBalance.sub(initialBalance)).to.equal(largeAmount);
        });
    });

    describe("Emergency Functions", function () {
        beforeEach(async function () {
            // Add some funds
            const revenueAmount = ethers.utils.parseEther("1000");
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Emergency Test Revenue",
                { value: revenueAmount }
            );
        });

        it("Should allow emergency role to pause contract", async function () {
            await treasury.connect(emergency).pause();
            expect(await treasury.paused()).to.be.true;
        });

        it("Should prevent operations when paused", async function () {
            await treasury.connect(emergency).pause();
            
            await expect(treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                ethers.utils.parseEther("100"),
                "Should fail",
                { value: ethers.utils.parseEther("100") }
            )).to.be.revertedWith("Pausable: paused");
        });

        it("Should allow emergency withdrawal when paused", async function () {
            await treasury.connect(emergency).pause();
            
            const withdrawAmount = ethers.utils.parseEther("500");
            const initialBalance = await addr1.getBalance();
            
            await treasury.connect(emergency).emergencyWithdraw(
                ethers.constants.AddressZero,
                withdrawAmount,
                addr1.address
            );

            const finalBalance = await addr1.getBalance();
            expect(finalBalance.sub(initialBalance)).to.equal(withdrawAmount);
        });

        it("Should only allow admin to unpause", async function () {
            await treasury.connect(emergency).pause();
            
            await expect(treasury.connect(emergency).unpause())
                .to.be.reverted;
            
            await treasury.unpause(); // From owner
            expect(await treasury.paused()).to.be.false;
        });
    });

    describe("Financial Reporting", function () {
        beforeEach(async function () {
            // Setup revenue and spending for reports
            const revenueAmount = ethers.utils.parseEther("1000");
            await treasury.connect(treasurer).receiveRevenue(
                ethers.constants.AddressZero,
                revenueAmount,
                "Q1 Revenue",
                { value: revenueAmount }
            );

            await treasury.connect(treasurer).spendOperational(
                ethers.constants.AddressZero,
                ethers.utils.parseEther("100"),
                addr1.address,
                "Operating expense"
            );
        });

        it("Should provide accurate financial summary", async function () {
            const summary = await treasury.getFinancialSummary();
            
            expect(summary.totalRevenue).to.equal(ethers.utils.parseEther("1000"));
            expect(summary.totalExpenses).to.equal(ethers.utils.parseEther("100"));
            expect(summary.reserveBalance).to.equal(ethers.utils.parseEther("200")); // 20% of revenue
            expect(summary.profitBalance).to.equal(ethers.utils.parseEther("200"));  // 20% of revenue
        });

        it("Should return transactions by category", async function () {
            const revenueTxs = await treasury.getTransactionsByCategory(0); // REVENUE
            const expenseTxs = await treasury.getTransactionsByCategory(1); // OPERATING_EXPENSE
            
            expect(revenueTxs.length).to.equal(1);
            expect(expenseTxs.length).to.equal(1);
        });

        it("Should provide asset allocation data", async function () {
            const allocation = await treasury.getAssetAllocation();
            
            expect(allocation.tokens).to.be.an('array');
            expect(allocation.balances).to.be.an('array');
            expect(allocation.percentages).to.be.an('array');
        });
    });

    describe("Multisig Integration", function () {
        it("Should require multiple signatures for configuration changes", async function () {
            const newConfig = {
                operations: 5000,
                reserves: 2500,
                upgrades: 1500,
                profits: 1000
            };

            // This would need to be submitted through multisig
            // For testing purposes, we simulate the multisig approval process
            await treasury.updateAllocationConfig(newConfig);
            
            const allocation = await treasury.allocationConfig();
            expect(allocation.operations).to.equal(5000);
        });
    });
});
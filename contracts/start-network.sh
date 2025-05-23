#!/bin/sh
set -e

echo "Starting Hardhat network..."

# Start Hardhat network in the background
npx hardhat node --hostname 0.0.0.0 &
HARDHAT_PID=$!

# Wait for network to be ready
echo "Waiting for Hardhat network to be ready..."
sleep 10

# Deploy contracts
echo "Deploying contracts..."
npx hardhat run scripts/deploy.js --network localhost

# Save deployment addresses
echo "Deployment complete. Contract addresses saved to /app/deployments/"

# Keep the container running
wait $HARDHAT_PID
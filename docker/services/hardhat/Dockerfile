FROM node:18-alpine

# Install dependencies
RUN apk add --no-cache git python3 make g++

# Create app directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy contract source
COPY . ./

# Compile contracts
RUN npm run compile

# Copy and set up startup script
COPY start-network.sh /app/start-network.sh
RUN chmod +x /app/start-network.sh

# Expose Hardhat network port
EXPOSE 8545

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --retries=3 \
  CMD node -e "require('http').get('http://localhost:8545', (res) => process.exit(res.statusCode === 200 ? 0 : 1))"

# Run the local network
CMD ["/app/start-network.sh"]
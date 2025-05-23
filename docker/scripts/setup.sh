#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${GREEN}Setting up Autonomous Content Service Docker deployment...${NC}"

# Change to docker directory
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo ""
    echo -e "${YELLOW}IMPORTANT: Please edit docker/.env file with your configuration:${NC}"
    echo "  - Set a secure JWT_SECRET"
    echo "  - Set a secure DB_PASSWORD"
    echo "  - Add your LLM_API_KEY (OpenAI key)"
    echo "  - Configure other services as needed"
    echo ""
    echo "Then run: docker/scripts/deploy.sh"
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
fi

# Make scripts executable
chmod +x scripts/*.sh

echo -e "${GREEN}Setup complete!${NC}"
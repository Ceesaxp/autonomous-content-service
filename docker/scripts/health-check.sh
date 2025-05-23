#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

# Configuration
API_URL="${API_URL:-http://localhost/api}"
WEB_URL="${WEB_URL:-http://localhost}"

# Health check functions
check_service() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    echo -n "Checking $name... "
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_status"; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

check_docker_service() {
    local service=$1
    echo -n "Checking Docker service $service... "
    
    if docker-compose ps | grep -q "${service}.*Up.*healthy"; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# Main health checks
echo "Running health checks..."
echo ""

# Check Docker services
echo "Docker Services:"
check_docker_service "postgres"
check_docker_service "redis"
check_docker_service "hardhat"
check_docker_service "api"
check_docker_service "caddy"

echo ""
echo "HTTP Endpoints:"

# Check HTTP endpoints
check_service "Web Frontend" "$WEB_URL"
check_service "API Health" "$API_URL/health"
check_service "API Version" "$API_URL/v1/version"

echo ""
echo "Service Details:"

# Get service versions/info
echo -n "Database: "
docker-compose exec -T postgres psql -U contentservice -d contentservice -c "SELECT version();" 2>/dev/null | grep PostgreSQL | head -1 || echo "Unable to connect"

echo -n "Redis: "
docker-compose exec -T redis redis-cli INFO server 2>/dev/null | grep redis_version | cut -d: -f2 || echo "Unable to connect"

echo -n "API Version: "
curl -s "$API_URL/v1/version" 2>/dev/null | jq -r '.version' 2>/dev/null || echo "Unable to fetch"

echo ""
echo "Container Resource Usage:"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" $(docker-compose ps -q)
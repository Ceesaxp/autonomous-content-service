#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/docker"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Main deployment
cd "$DOCKER_DIR"

log_info "Starting Autonomous Content Service (simplified deployment)..."

# Check for .env file
if [ ! -f .env ]; then
    log_error ".env file not found. Run ./scripts/setup.sh first."
    exit 1
fi

# Build and start services
log_info "Building and starting services..."
docker-compose -f docker-compose.simple.yml up --build -d

# Wait for services to be ready
log_info "Waiting for services to start..."
sleep 20

# Check service health
log_info "Checking service health..."
docker-compose -f docker-compose.simple.yml ps

log_info "Deployment complete! 🚀"
echo ""
echo "Services are running at:"
echo "  - Web Interface: http://localhost"
echo "  - API Health: http://localhost/api/health"
echo ""
echo "To view logs: docker-compose -f docker-compose.simple.yml logs -f [service_name]"
echo "To stop: docker-compose -f docker-compose.simple.yml down"
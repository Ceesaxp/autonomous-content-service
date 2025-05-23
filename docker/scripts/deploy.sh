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

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_info "All dependencies are installed."
}

# Setup environment
setup_environment() {
    log_info "Setting up environment..."
    
    cd "$DOCKER_DIR"
    
    # Create .env file if it doesn't exist
    if [ ! -f .env ]; then
        log_info "Creating .env file from template..."
        cp .env.example .env
        log_warn "Please edit .env file with your configuration before proceeding."
        exit 1
    fi
    
    # Load environment variables
    source .env
}

# Build containers
build_containers() {
    log_info "Building Docker containers..."
    
    docker-compose build --no-cache
    
    log_info "Containers built successfully."
}

# Start services
start_services() {
    log_info "Starting services..."
    
    # Start infrastructure services first
    docker-compose up -d postgres redis
    
    # Wait for database to be ready
    log_info "Waiting for database to be ready..."
    sleep 10
    
    # Start blockchain
    docker-compose up -d hardhat
    
    # Wait for blockchain to be ready
    log_info "Waiting for blockchain to be ready..."
    sleep 15
    
    # Start API service
    docker-compose up -d api
    
    # Wait for API to be ready
    log_info "Waiting for API to be ready..."
    sleep 10
    
    # Start Caddy (web frontend and reverse proxy)
    docker-compose up -d caddy
    
    log_info "All services started successfully."
}

# Health check
health_check() {
    log_info "Running health checks..."
    
    # Check each service
    services=("postgres" "redis" "hardhat" "api" "caddy")
    
    for service in "${services[@]}"; do
        if docker-compose ps | grep -q "${service}.*Up.*healthy"; then
            log_info "$service is healthy"
        else
            log_error "$service is not healthy"
            docker-compose logs --tail=50 $service
            exit 1
        fi
    done
    
    log_info "All services are healthy!"
}

# Show status
show_status() {
    log_info "Deployment complete! 🚀"
    echo ""
    echo "Services are running at:"
    echo "  - Web Interface: http://localhost"
    echo "  - API: http://localhost/api"
    echo "  - Database: localhost:5432"
    echo "  - Redis: localhost:6379"
    echo "  - Blockchain: localhost:8545"
    echo ""
    echo "To view logs: docker-compose logs -f [service_name]"
    echo "To stop: docker-compose down"
    echo "To stop and remove data: docker-compose down -v"
}

# Main deployment flow
main() {
    log_info "Starting Autonomous Content Service deployment..."
    
    check_dependencies
    setup_environment
    build_containers
    start_services
    health_check
    show_status
}

# Handle command line arguments
case "${1:-}" in
    "stop")
        log_info "Stopping services..."
        cd "$DOCKER_DIR"
        docker-compose down
        ;;
    "clean")
        log_warn "Removing all services and data..."
        cd "$DOCKER_DIR"
        docker-compose down -v
        ;;
    "logs")
        cd "$DOCKER_DIR"
        docker-compose logs -f ${2:-}
        ;;
    "status")
        cd "$DOCKER_DIR"
        docker-compose ps
        ;;
    *)
        main
        ;;
esac
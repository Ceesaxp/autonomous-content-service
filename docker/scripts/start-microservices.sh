#!/bin/bash

# Autonomous Content Service - Microservices Startup Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker/docker-compose.microservices.yml"
ENV_FILE=".env"
PROJECT_NAME="acs-microservices"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Check environment file
check_environment() {
    log_info "Checking environment configuration..."
    
    if [ ! -f "$ENV_FILE" ]; then
        log_warning "Environment file $ENV_FILE not found"
        log_info "Creating environment file from template..."
        cp .env.microservices.example $ENV_FILE
        log_warning "Please edit $ENV_FILE with your actual configuration values"
        read -p "Press Enter to continue after updating the environment file..."
    fi
    
    # Check for required environment variables
    source $ENV_FILE
    
    if [ -z "$LLM_API_KEY" ] || [ "$LLM_API_KEY" = "your_openai_api_key_here" ]; then
        log_warning "LLM_API_KEY not configured. Content service functionality will be limited."
    fi
    
    log_success "Environment configuration checked"
}

# Build services
build_services() {
    log_info "Building microservices..."
    
    # Build all services
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build --parallel
    
    log_success "All microservices built successfully"
}

# Start infrastructure services first
start_infrastructure() {
    log_info "Starting infrastructure services..."
    
    # Start infrastructure services
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d postgres redis consul
    
    # Wait for services to be ready
    log_info "Waiting for infrastructure services to be ready..."
    
    # Wait for PostgreSQL
    log_info "Waiting for PostgreSQL..."
    until docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T postgres pg_isready -U postgres; do
        sleep 2
    done
    
    # Wait for Redis
    log_info "Waiting for Redis..."
    until docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T redis redis-cli ping; do
        sleep 2
    done
    
    # Wait for Consul
    log_info "Waiting for Consul..."
    until docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T consul consul members; do
        sleep 2
    done
    
    log_success "Infrastructure services are ready"
}

# Start business services
start_business_services() {
    log_info "Starting business microservices..."
    
    # Start all business services
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d \
        content-service \
        decision-service \
        financial-service \
        governance-service \
        hr-service \
        legal-service \
        risk-service \
        self-improvement-service
    
    log_info "Waiting for business services to be ready..."
    sleep 10
    
    log_success "Business microservices started"
}

# Start API Gateway
start_api_gateway() {
    log_info "Starting API Gateway..."
    
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d api-gateway
    
    log_info "Waiting for API Gateway to be ready..."
    sleep 5
    
    log_success "API Gateway started"
}

# Start monitoring services
start_monitoring() {
    log_info "Starting monitoring services..."
    
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d prometheus grafana
    
    log_success "Monitoring services started"
}

# Show service status
show_status() {
    log_info "Service Status:"
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
    
    echo ""
    log_info "Service URLs:"
    echo "  API Gateway:         http://localhost:8080"
    echo "  Content Service:     http://localhost:8081"
    echo "  Decision Service:    http://localhost:8082"
    echo "  Financial Service:   http://localhost:8083"
    echo "  Governance Service:  http://localhost:8084"
    echo "  HR Service:          http://localhost:8085"
    echo "  Legal Service:       http://localhost:8086"
    echo "  Risk Service:        http://localhost:8087"
    echo "  Self-Improvement:    http://localhost:8088"
    echo ""
    echo "  Consul UI:           http://localhost:8500"
    echo "  Prometheus:          http://localhost:9090"
    echo "  Grafana:             http://localhost:3000 (admin/admin)"
    echo ""
    echo "  PostgreSQL:          localhost:5432 (postgres/postgres)"
    echo "  Redis:               localhost:6379"
}

# Main execution
main() {
    log_info "Starting Autonomous Content Service Microservices..."
    
    check_prerequisites
    check_environment
    build_services
    start_infrastructure
    start_business_services
    start_api_gateway
    start_monitoring
    
    log_success "All services started successfully!"
    show_status
    
    log_info "To stop all services, run: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down"
    log_info "To view logs, run: docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f [service-name]"
}

# Execute main function
main "$@"
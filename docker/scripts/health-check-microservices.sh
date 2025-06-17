#!/bin/bash

# Autonomous Content Service - Health Check Script for Microservices
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service configurations
declare -A services=(
    ["api-gateway"]="8080"
    ["content-service"]="8081"
    ["decision-service"]="8082"
    ["financial-service"]="8083"
    ["governance-service"]="8084"
    ["hr-service"]="8085"
    ["legal-service"]="8086"
    ["risk-service"]="8087"
    ["self-improvement-service"]="8088"
)

declare -A infrastructure=(
    ["consul"]="8500"
    ["prometheus"]="9090"
    ["grafana"]="3000"
)

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

# Check service health
check_service_health() {
    local service=$1
    local port=$2
    local endpoint=${3:-"/health"}
    
    log_info "Checking $service on port $port..."
    
    # Check if service is responding
    if curl -f -s --connect-timeout 5 "http://localhost:$port$endpoint" > /dev/null; then
        log_success "$service is healthy"
        return 0
    else
        log_error "$service is unhealthy or not responding"
        return 1
    fi
}

# Check container status
check_container_status() {
    local container_name=$1
    
    if docker ps --filter "name=$container_name" --filter "status=running" | grep -q "$container_name"; then
        log_success "Container $container_name is running"
        return 0
    else
        log_error "Container $container_name is not running"
        return 1
    fi
}

# Check all microservices
check_microservices() {
    log_info "Checking microservices health..."
    
    local failed_services=0
    local total_services=${#services[@]}
    
    for service in "${!services[@]}"; do
        local port=${services[$service]}
        local container_name="acs-$service"
        
        echo ""
        log_info "=== Checking $service ==="
        
        # Check container status
        if ! check_container_status "$container_name"; then
            ((failed_services++))
            continue
        fi
        
        # Check health endpoint
        if ! check_service_health "$service" "$port"; then
            ((failed_services++))
        fi
    done
    
    echo ""
    if [ $failed_services -eq 0 ]; then
        log_success "All $total_services microservices are healthy"
    else
        log_error "$failed_services out of $total_services microservices have issues"
    fi
    
    return $failed_services
}

# Check infrastructure services
check_infrastructure() {
    log_info "Checking infrastructure services..."
    
    local failed_services=0
    
    # Check PostgreSQL
    echo ""
    log_info "=== Checking PostgreSQL ==="
    if check_container_status "acs-postgres"; then
        if docker exec acs-postgres pg_isready -U postgres > /dev/null 2>&1; then
            log_success "PostgreSQL is ready"
        else
            log_error "PostgreSQL is not ready"
            ((failed_services++))
        fi
    else
        ((failed_services++))
    fi
    
    # Check Redis
    echo ""
    log_info "=== Checking Redis ==="
    if check_container_status "acs-redis"; then
        if docker exec acs-redis redis-cli ping > /dev/null 2>&1; then
            log_success "Redis is ready"
        else
            log_error "Redis is not ready"
            ((failed_services++))
        fi
    else
        ((failed_services++))
    fi
    
    # Check infrastructure web services
    for service in "${!infrastructure[@]}"; do
        local port=${infrastructure[$service]}
        local container_name="acs-$service"
        
        echo ""
        log_info "=== Checking $service ==="
        
        if check_container_status "$container_name"; then
            check_service_health "$service" "$port" "/"
        else
            ((failed_services++))
        fi
    done
    
    echo ""
    if [ $failed_services -eq 0 ]; then
        log_success "All infrastructure services are healthy"
    else
        log_error "$failed_services infrastructure services have issues"
    fi
    
    return $failed_services
}

# Show service endpoints
show_endpoints() {
    echo ""
    log_info "=== Service Endpoints ==="
    echo ""
    
    echo "Business Services:"
    for service in "${!services[@]}"; do
        local port=${services[$service]}
        echo "  $service: http://localhost:$port/health"
    done
    
    echo ""
    echo "Infrastructure Services:"
    for service in "${!infrastructure[@]}"; do
        local port=${infrastructure[$service]}
        echo "  $service: http://localhost:$port"
    done
    
    echo ""
    echo "Database Services:"
    echo "  PostgreSQL: localhost:5432 (postgres/postgres)"
    echo "  Redis: localhost:6379"
}

# Show service logs
show_logs() {
    local service=$1
    local lines=${2:-50}
    
    if [ -z "$service" ]; then
        log_info "Available services for logs:"
        for service in "${!services[@]}"; do
            echo "  $service"
        done
        for service in "${!infrastructure[@]}"; do
            echo "  $service"
        done
        echo "  postgres"
        echo "  redis"
        return 0
    fi
    
    local container_name="acs-$service"
    
    if docker ps --filter "name=$container_name" | grep -q "$container_name"; then
        log_info "Showing last $lines lines for $service:"
        docker logs --tail "$lines" "$container_name"
    else
        log_error "Container $container_name not found"
        return 1
    fi
}

# Main function
main() {
    case "${1:-check}" in
        "check")
            log_info "Starting health check for Autonomous Content Service..."
            
            local infrastructure_issues=0
            local microservice_issues=0
            
            check_infrastructure
            infrastructure_issues=$?
            
            check_microservices
            microservice_issues=$?
            
            show_endpoints
            
            local total_issues=$((infrastructure_issues + microservice_issues))
            
            echo ""
            if [ $total_issues -eq 0 ]; then
                log_success "All services are healthy!"
                exit 0
            else
                log_error "Found $total_issues issues. Check the logs for more details."
                exit 1
            fi
            ;;
        "logs")
            show_logs "$2" "$3"
            ;;
        "endpoints")
            show_endpoints
            ;;
        *)
            echo "Usage: $0 [check|logs|endpoints]"
            echo ""
            echo "Commands:"
            echo "  check      - Check health of all services (default)"
            echo "  logs       - Show logs for a specific service"
            echo "  endpoints  - Show all service endpoints"
            echo ""
            echo "Examples:"
            echo "  $0 check"
            echo "  $0 logs content-service"
            echo "  $0 logs content-service 100"
            echo "  $0 endpoints"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
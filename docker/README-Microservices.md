# Autonomous Content Service - Microservices Architecture

This directory contains the Docker Compose configuration and support files for running the Autonomous Content Service as a microservices architecture.

## Architecture Overview

The system is decomposed into 9 independent microservices:

### Business Services
- **API Gateway** (Port 8080) - Central entry point and request routing
- **Content Service** (Port 8081) - Content creation and management ✅ *Full functionality*
- **Decision Service** (Port 8082) - Autonomous decision making
- **Financial Service** (Port 8083) - Payment processing and treasury management
- **Governance Service** (Port 8084) - System governance and compliance
- **HR Service** (Port 8085) - Human resource and talent management
- **Legal Service** (Port 8086) - Legal compliance and contract management
- **Risk Service** (Port 8087) - Risk assessment and management ✅ *Full functionality*
- **Self-Improvement Service** (Port 8088) - System learning and optimization ✅ *Full functionality*

### Infrastructure Services
- **PostgreSQL** (Port 5432) - Primary database
- **Redis** (Port 6379) - Caching and event streaming
- **Consul** (Port 8500) - Service discovery and configuration
- **Prometheus** (Port 9090) - Metrics collection and monitoring
- **Grafana** (Port 3000) - Metrics visualization and dashboards

## Quick Start

### Prerequisites
- Docker 20.0+ installed and running
- Docker Compose v2.0+ installed
- At least 4GB RAM available for containers
- Ports 8080-8088, 3000, 5432, 6379, 8500, 9090 available

### 1. Environment Setup

```bash
# Copy environment template
cp .env.microservices.example .env

# Edit with your configuration
nano .env
```

Required configuration:
- `LLM_API_KEY` - OpenAI API key for content generation
- `STRIPE_SECRET_KEY` - Stripe API key for payments (if using financial features)

### 2. Start All Services

```bash
# Run the startup script
./docker/scripts/start-microservices.sh
```

This script will:
1. Build all Docker images
2. Start infrastructure services (PostgreSQL, Redis, Consul)
3. Start business microservices
4. Start API Gateway
5. Start monitoring services (Prometheus, Grafana)

### 3. Verify Deployment

```bash
# Check service health
./docker/scripts/health-check-microservices.sh

# View service status
docker-compose -f docker/docker-compose.microservices.yml ps
```

## Service Access

### Web Interfaces
- **API Gateway**: http://localhost:8080
- **Consul UI**: http://localhost:8500
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### API Endpoints
Each service exposes its own API:
- Content: http://localhost:8081/api/v1/
- Decision: http://localhost:8082/api/v1/
- Financial: http://localhost:8083/api/v1/
- Governance: http://localhost:8084/api/v1/
- HR: http://localhost:8085/api/v1/
- Legal: http://localhost:8086/api/v1/
- Risk: http://localhost:8087/api/v1/
- Self-Improvement: http://localhost:8088/api/v1/

### Database Access
- **PostgreSQL**: localhost:5432 (postgres/postgres)
- **Redis**: localhost:6379

## Service Management

### View Logs
```bash
# All services
docker-compose -f docker/docker-compose.microservices.yml logs -f

# Specific service
docker-compose -f docker/docker-compose.microservices.yml logs -f content-service

# Using health check script
./docker/scripts/health-check-microservices.sh logs content-service
```

### Scale Services
```bash
# Scale content service to 3 instances
docker-compose -f docker/docker-compose.microservices.yml up -d --scale content-service=3

# Scale multiple services
docker-compose -f docker/docker-compose.microservices.yml up -d --scale content-service=2 --scale risk-service=2
```

### Stop Services
```bash
# Stop all services
docker-compose -f docker/docker-compose.microservices.yml down

# Stop and remove volumes (data loss!)
docker-compose -f docker/docker-compose.microservices.yml down -v
```

### Restart Single Service
```bash
# Restart content service
docker-compose -f docker/docker-compose.microservices.yml restart content-service

# Rebuild and restart
docker-compose -f docker/docker-compose.microservices.yml up -d --build content-service
```

## Service Development Status

### ✅ Fully Functional Services
These services have complete implementations with real business logic:

1. **Content Service** - Full content creation pipeline with LLM integration
2. **Risk Service** - Complete risk management and monitoring
3. **Self-Improvement Service** - Learning and optimization capabilities

### 🚧 Mock Implementation Services
These services compile and run but use basic mock handlers:

1. **Decision Service** - Basic decision endpoints
2. **Financial Service** - Basic payment endpoints
3. **Governance Service** - Basic governance endpoints
4. **HR Service** - Basic HR endpoints
5. **Legal Service** - Basic legal endpoints
6. **API Gateway** - Basic routing with real service integration

## Monitoring and Observability

### Prometheus Metrics
- Service-level metrics (requests, errors, latency)
- Business metrics (content quality, decision accuracy)
- Infrastructure metrics (database connections, memory usage)

### Grafana Dashboards
Access Grafana at http://localhost:3000 (admin/admin) for:
- Service health dashboards
- Performance monitoring
- Business KPI tracking

### Health Checks
All services provide `/health` endpoints:
```bash
curl http://localhost:8081/health  # Content service
curl http://localhost:8087/health  # Risk service
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check what's using a port
   lsof -i :8080
   
   # Kill process using port
   kill -9 $(lsof -t -i :8080)
   ```

2. **Database Connection Issues**
   ```bash
   # Check PostgreSQL logs
   docker-compose -f docker/docker-compose.microservices.yml logs postgres
   
   # Connect to database
   docker-compose -f docker/docker-compose.microservices.yml exec postgres psql -U postgres contentservice
   ```

3. **Service Discovery Issues**
   ```bash
   # Check Consul
   curl http://localhost:8500/v1/catalog/services
   
   # View Consul logs
   docker-compose -f docker/docker-compose.microservices.yml logs consul
   ```

4. **Memory Issues**
   ```bash
   # Check container memory usage
   docker stats
   
   # Increase Docker memory allocation in Docker Desktop
   ```

### Debug Mode

Enable debug logging by setting environment variables:
```bash
# In .env file
DEBUG=true
LOG_LEVEL=debug
```

### Service Dependencies

If services fail to start, check dependency order:
1. PostgreSQL + Redis + Consul (infrastructure)
2. Business microservices
3. API Gateway
4. Monitoring services

## Development Workflow

### Building Changes
```bash
# Rebuild specific service after code changes
docker-compose -f docker/docker-compose.microservices.yml build content-service
docker-compose -f docker/docker-compose.microservices.yml up -d content-service

# Rebuild all services
docker-compose -f docker/docker-compose.microservices.yml build
```

### Local Development
For active development, you can run services locally instead of in containers:

1. Start infrastructure services only:
   ```bash
   docker-compose -f docker/docker-compose.microservices.yml up -d postgres redis consul
   ```

2. Run your service locally:
   ```bash
   export DATABASE_URL=postgres://postgres:postgres@localhost:5432/contentservice?sslmode=disable
   export REDIS_URL=redis://localhost:6379
   export CONSUL_URL=localhost:8500
   go run ./src/cmd/content-service/
   ```

## Next Steps

1. **Complete Service Implementations** - Replace mock handlers with real business logic
2. **Add Circuit Breakers** - Implement resilience patterns between services  
3. **Service Mesh** - Consider Istio or Linkerd for advanced traffic management
4. **Kubernetes Migration** - Deploy to production Kubernetes cluster
5. **Automated Testing** - Add integration tests for service interactions

## Support

For issues or questions:
1. Check service logs using the health check script
2. Review Grafana dashboards for performance insights
3. Consult the main project README.md for architecture details
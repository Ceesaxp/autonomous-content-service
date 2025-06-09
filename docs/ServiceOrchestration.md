# Service Orchestration Architecture

## Overview

This document outlines the service orchestration approach for the Autonomous Content Service. We've chosen a **Docker Compose-based microservices architecture** that avoids the complexity of Kubernetes while providing all the benefits of service decomposition and orchestration.

## Architecture Philosophy

### Why Not Kubernetes?

**Complexity vs. Value Analysis:**
- **Operational overhead**: K8s requires cluster management, networking config, RBAC, secrets management
- **Development friction**: Local development becomes complex (minikube, kind, etc.)
- **Learning curve**: Requires K8s expertise that doesn't add value to our core business logic
- **Resource overhead**: K8s control plane consumes significant resources
- **Over-engineering**: We're building a single autonomous entity, not Netflix-scale infrastructure

**Our Use Case Reality:**
- Single autonomous service (not a massive distributed system)
- Likely running on single server or small cluster initially
- Docker Compose already provides what we need
- Focus should be on business logic, not infrastructure complexity

### Chosen Approach: Enhanced Docker Compose

**Core Principles:**
- **Simplicity**: Use Docker Compose as orchestration foundation
- **Targeted tools**: Add specific tools for specific capabilities
- **Service independence**: Each service can be developed, deployed, and scaled independently
- **Development-friendly**: Same tools for dev and production
- **Migration path**: Can upgrade to Docker Swarm or K8s later if truly needed

## Service Architecture

### Platform Services

| Service | Purpose | Technology |
|---------|---------|------------|
| `consul` | Service discovery & configuration management | HashiCorp Consul |
| `redis` | Caching & async message broker | Redis with Streams |
| `postgres` | Primary database | PostgreSQL 15 |
| `temporal` | Workflow orchestration | Temporal.io |
| `prometheus` | Metrics collection | Prometheus |
| `grafana` | Monitoring dashboards | Grafana |
| `api-gateway` | Entry point, routing, authentication | Custom Go service |

### Business Microservices

| Service | Responsibility | Port | Database |
|---------|----------------|------|----------|
| `content-service` | Content creation pipeline, LLM orchestration | 8081 | Shared |
| `decision-service` | AI decision making, policy enforcement | 8082 | Shared |
| `hr-service` | Talent management, workforce operations | 8083 | Shared |
| `financial-service` | Payments, pricing, treasury management | 8084 | Shared |
| `governance-service` | DAO governance, voting, proposals | 8085 | Shared |
| `legal-service` | Compliance, contracts, regulations | 8086 | Shared |
| `risk-service` | Risk assessment, incident management | 8087 | Shared |
| `self-improvement-service` | Learning, optimization, experimentation | 8088 | Shared |

## Communication Architecture

### Service Discovery
**Consul-based Discovery:**
```go
// Service registration
consul.RegisterService("content-service", "localhost:8081", "/health")

// Service discovery
endpoints := consul.DiscoverService("financial-service")
```

### Inter-Service Communication

**Synchronous Communication (REST):**
```yaml
# Through API Gateway
Client → API Gateway → Target Service

# Direct service-to-service (internal)
content-service → financial-service (pricing queries)
```

**Asynchronous Communication (Events):**
```yaml
# Redis Streams for event-driven patterns
Publisher → Redis Stream → Subscriber(s)

# Example: Content approval workflow
content-service → "content.approved" → [legal-service, financial-service]
```

### Event Schema
```json
{
  "event_type": "content.approved",
  "event_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z",
  "source_service": "content-service",
  "payload": {
    "content_id": "uuid",
    "project_id": "uuid",
    "client_id": "uuid",
    "metadata": {}
  }
}
```

## Configuration Management

### Multi-Layer Configuration
```yaml
# 1. Environment variables (Docker Compose)
environment:
  APP_ENV: production
  DB_HOST: postgres
  CONSUL_ADDR: consul:8500

# 2. Consul KV store (dynamic config)
/config/content-service/llm_model: "gpt-4"
/config/content-service/max_tokens: "2048"

# 3. Service-specific config files
./config/content-service.yml
```

### Configuration Hierarchy
1. **Environment variables** (highest priority)
2. **Consul KV store** (dynamic, hot-reloadable)
3. **Config files** (default values)
4. **Code defaults** (fallback)

## Workflow Orchestration

### Temporal.io for Complex Workflows
```go
// Business process workflow
type ClientOnboardingWorkflow struct {
    // Workflow definition
}

func (w *ClientOnboardingWorkflow) Execute(ctx workflow.Context, input OnboardingInput) error {
    // Step 1: Validate client data
    err := workflow.ExecuteActivity(ctx, ValidateClientActivity, input.ClientData)
    
    // Step 2: Create project
    project, err := workflow.ExecuteActivity(ctx, CreateProjectActivity, input.ProjectData)
    
    // Step 3: Initialize content pipeline
    err = workflow.ExecuteActivity(ctx, InitializeContentPipelineActivity, project.ID)
    
    return nil
}
```

### Redis Streams for Simple Event Chains
```go
// Simple event-driven workflows
type EventHandler struct {
    redis *redis.Client
}

func (h *EventHandler) HandleContentApproved(event ContentApprovedEvent) {
    // Trigger invoice generation
    h.redis.XAdd("financial.invoice.generate", event.ToMap())
    
    // Notify client
    h.redis.XAdd("notification.client.notify", event.ToMap())
}
```

## Resilience Patterns

### Circuit Breakers
```go
type CircuitBreaker struct {
    maxFailures   int
    resetTimeout  time.Duration
    state        State // Closed, Open, HalfOpen
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        return ErrCircuitOpen
    }
    
    err := fn()
    cb.handleResult(err)
    return err
}
```

### Retry with Exponential Backoff
```go
type RetryConfig struct {
    MaxRetries  int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func WithRetry(config RetryConfig, fn func() error) error {
    var lastErr error
    
    for i := 0; i <= config.MaxRetries; i++ {
        if i > 0 {
            delay := min(config.BaseDelay * time.Duration(math.Pow(config.Multiplier, float64(i-1))), config.MaxDelay)
            time.Sleep(delay)
        }
        
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
        }
    }
    
    return lastErr
}
```

### Health Checks
```go
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck func() error

func (hc *HealthChecker) Check() map[string]string {
    results := make(map[string]string)
    
    for name, check := range hc.checks {
        if err := check(); err != nil {
            results[name] = "unhealthy: " + err.Error()
        } else {
            results[name] = "healthy"
        }
    }
    
    return results
}
```

## Monitoring & Observability

### Metrics Collection (Prometheus)
```go
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"service", "method", "endpoint", "status"},
    )
    
    eventProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_processed_total",
            Help: "Total processed events",
        },
        []string{"service", "event_type", "status"},
    )
)
```

### Distributed Tracing
```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx, span := trace.StartSpan(r.Context(), "handle_request")
    defer span.End()
    
    // Add service context
    span.SetAttributes(
        attribute.String("service", "content-service"),
        attribute.String("operation", "create_content"),
    )
    
    // Propagate context to downstream services
    client := &http.Client{}
    req, _ := http.NewRequestWithContext(ctx, "POST", "http://financial-service/price", body)
}
```

### Centralized Logging
```yaml
# docker-compose.yml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
    labels: "service,environment"
```

## Directory Structure

```
src/
├── cmd/                                # Service entry points
│   ├── api-gateway/main.go
│   ├── content-service/main.go
│   ├── decision-service/main.go
│   ├── hr-service/main.go
│   ├── financial-service/main.go
│   ├── governance-service/main.go
│   ├── legal-service/main.go
│   ├── risk-service/main.go
│   └── self-improvement-service/main.go
├── services/                          # Business logic (existing)
│   ├── content_creation/
│   ├── decision_making/
│   ├── hr_management/
│   ├── payment/
│   ├── pricing/
│   ├── dao_governance/
│   ├── legal_compliance/
│   ├── risk_management/
│   └── self_improvement/
├── pkg/                               # Shared libraries
│   ├── discovery/                     # Consul service discovery
│   ├── events/                        # Event schemas and handling
│   ├── config/                        # Configuration management
│   ├── monitoring/                    # Observability utilities
│   ├── resilience/                    # Circuit breakers, retries
│   ├── auth/                          # Authentication utilities
│   └── middleware/                    # HTTP middleware
├── docker/
│   ├── docker-compose.microservices.yml
│   ├── docker-compose.monitoring.yml
│   ├── docker-compose.dev.yml
│   └── services/                      # Dockerfiles for each service
└── workflows/                         # Temporal workflow definitions
    ├── client_onboarding.go
    ├── content_creation.go
    └── financial_processing.go
```

## Development Workflow

### Local Development
```bash
# Start all services
docker-compose -f docker-compose.microservices.yml -f docker-compose.dev.yml up

# Start specific services
docker-compose up consul redis postgres content-service

# Scale services for testing
docker-compose up --scale content-service=3

# Monitor logs
docker-compose logs -f content-service
```

### Service Development
```bash
# Each service can be developed independently
cd src/cmd/content-service
go run main.go

# Or with hot reload
air # with .air.toml configuration
```

### Testing
```bash
# Unit tests per service
go test ./src/services/content_creation/...

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# End-to-end tests
./scripts/e2e-test.sh
```

## Deployment Strategies

### Environment Management
```yaml
# Development
docker-compose -f docker-compose.microservices.yml -f docker-compose.dev.yml up

# Staging  
APP_ENV=staging docker-compose -f docker-compose.microservices.yml up

# Production
APP_ENV=production docker-compose -f docker-compose.microservices.yml -f docker-compose.prod.yml up
```

### Rolling Updates
```bash
# Update specific service
docker-compose build content-service
docker-compose up -d --no-deps content-service

# Blue-green deployment simulation
docker-compose -p blue up -d
# Test blue environment
docker-compose -p green up -d
# Switch traffic
docker-compose -p blue down
```

## Migration Path

### Current State → Microservices
1. **Phase 1**: Extract service interfaces, create individual main packages
2. **Phase 2**: Set up Docker Compose microservices configuration
3. **Phase 3**: Implement service discovery and event-driven communication
4. **Phase 4**: Add monitoring and resilience patterns

### Future Scaling Options
1. **Docker Swarm**: Easy upgrade from Compose for multi-node clustering
2. **Nomad**: HashiCorp's simpler orchestrator (works with our Consul)
3. **Kubernetes**: Only if we truly need the ecosystem (service mesh, operators, etc.)

## Benefits Achieved

✅ **Service Independence**: Each service can be developed, deployed, and scaled independently  
✅ **Fault Isolation**: Failures in one service don't cascade to others  
✅ **Technology Flexibility**: Services can use different technologies as needed  
✅ **Team Autonomy**: Different teams can own different services  
✅ **Scalability**: Services can be scaled based on individual demand  
✅ **Development Simplicity**: Docker Compose for everything  
✅ **Production Readiness**: Can upgrade orchestration layer if needed  
✅ **Cost Efficiency**: No operational overhead of complex orchestration  

## Conclusion

This architecture provides all the benefits of microservices and service orchestration without the complexity and overhead of Kubernetes. It's optimized for our use case: a single autonomous entity that needs to be reliable, scalable, and maintainable, but doesn't require massive distributed systems infrastructure.

The approach is pragmatic, developer-friendly, and provides a clear migration path for future scaling needs.
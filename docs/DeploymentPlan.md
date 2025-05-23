# Deployment Plan: Dockerized Autonomous Content Service

## Architecture Overview

The system consists of these components:
- **Go API Service** (main application)
- **PostgreSQL Database** (data persistence)
- **Hardhat Node** (local Ethereum testnet)
- **Static Web Server** (frontend dashboard)
- **Redis** (caching and session management)

## Docker Deployment Strategy

### 1. Multi-Service Docker Compose Architecture

```yaml
# docker-compose.yml structure
services:
  postgres:     # Database service
  redis:        # Cache and sessions
  hardhat:      # Local Ethereum testnet
  api:          # Go backend service
  web:          # Static web frontend
  nginx:        # Reverse proxy and load balancer
```

### 2. Container Specifications

#### A. Go API Service Container
- **Base Image**: `golang:1.23-alpine`
- **Build Strategy**: Multi-stage build (compile → runtime)
- **Dependencies**: PostgreSQL client, crypto libraries
- **Environment**: Production-ready with minimal attack surface

#### B. Database Container
- **Image**: `postgres:15-alpine`
- **Features**: Auto-initialization with schema, persistent volumes
- **Security**: Custom user/database, encrypted connections

#### C. Hardhat Testnet Container
- **Base Image**: `node:18-alpine`
- **Purpose**: Local Ethereum network for testing
- **Features**: Pre-deployed contracts, funded accounts

#### D. Frontend Container
- **Base Image**: `nginx:alpine`
- **Content**: Static dashboard files, optimized assets
- **Security**: Security headers, rate limiting

#### E. Infrastructure Container
- **Redis**: `redis:7-alpine` for caching and sessions
- **Nginx**: Reverse proxy with SSL termination

### 3. Network Architecture

```
Internet → Nginx (443/80) → API (8080) → Database (5432)
                         → Web (80)    → Redis (6379)
                         → Hardhat (8545)
```

### 4. File Structure Plan

```
/
├── docker/
│   ├── docker-compose.yml           # Main orchestration
│   ├── docker-compose.dev.yml       # Development overrides
│   ├── .env.example                 # Environment template
│   └── services/
│       ├── api/
│       │   ├── Dockerfile           # Go API container
│       │   └── entrypoint.sh        # Startup script
│       ├── web/
│       │   ├── Dockerfile           # Web frontend container
│       │   └── nginx.conf           # Web server config
│       ├── hardhat/
│       │   ├── Dockerfile           # Hardhat testnet
│       │   └── start-network.sh     # Network startup
│       └── nginx/
│           ├── Dockerfile           # Reverse proxy
│           ├── nginx.conf           # Main config
│           └── ssl/                 # SSL certificates
├── scripts/
│   ├── deploy.sh                    # Main deployment script
│   ├── setup-env.sh                 # Environment setup
│   ├── init-db.sh                   # Database initialization
│   ├── deploy-contracts.sh          # Smart contract deployment
│   └── health-check.sh              # System health verification
└── config/
    ├── production.env               # Production environment
    ├── development.env              # Development environment
    └── test.env                     # Testing environment
```

### 5. Deployment Process

#### Phase 1: Environment Preparation
1. **Environment Configuration**
   - Generate secure secrets (JWT, DB passwords)
   - Configure LLM API keys
   - Set up monitoring credentials

2. **SSL Certificate Setup**
   - Generate self-signed certificates for development
   - Configure Let's Encrypt for production

#### Phase 2: Infrastructure Deployment
1. **Database Initialization**
   - Create database schema
   - Insert seed data
   - Set up backup procedures

2. **Network Setup**
   - Start Hardhat local testnet
   - Deploy smart contracts
   - Configure contract addresses

#### Phase 3: Application Deployment
1. **Service Orchestration**
   - Build all containers
   - Start services in dependency order
   - Verify inter-service communication

2. **Health Verification**
   - API endpoint tests
   - Database connectivity
   - Frontend accessibility
   - Smart contract interaction

### 6. Configuration Management

#### Environment Variables Structure
```bash
# Core Application
APP_ENV=development
PORT=8080
JWT_SECRET=<generated-secret>

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=contentservice
DB_PASSWORD=<generated-password>
DB_NAME=contentservice

# LLM Integration
LLM_API_KEY=<openai-key>
LLM_MODEL=gpt-4
LLM_MAX_TOKENS=2048

# Blockchain
HARDHAT_NETWORK_URL=http://hardhat:8545
TREASURY_CONTRACT_ADDRESS=<deployed-address>

# External Services
STRIPE_SECRET_KEY=<stripe-test-key>
REDIS_URL=redis://redis:6379
```

### 7. Security Considerations

#### Container Security
- Non-root user execution
- Minimal base images (Alpine Linux)
- Regular security updates
- Secrets management via Docker secrets

#### Network Security
- Internal service communication only
- Nginx reverse proxy with rate limiting
- SSL/TLS encryption
- Security headers

### 8. Development vs Production Differences

#### Development Features
- Hot reload for Go service
- Exposed database ports
- Debug logging enabled
- Test API keys

#### Production Optimizations
- Multi-stage builds for smaller images
- Health checks and restart policies
- Resource limits and reservations
- Production logging and monitoring

### 9. Monitoring and Observability

#### Health Checks
- API endpoint `/health`
- Database connection verification
- Smart contract connectivity
- Service dependency validation

#### Logging Strategy
- Structured JSON logging
- Container log aggregation
- Error tracking and alerting
- Performance metrics collection

### 10. Backup and Recovery

#### Data Persistence
- PostgreSQL data volumes
- Configuration file backups
- Smart contract deployment artifacts
- SSL certificates and keys

#### Recovery Procedures
- Database restoration scripts
- Container image rollback
- Configuration restoration
- Emergency contact procedures

## Implementation Components

### Required Docker Files
1. **docker-compose.yml** - Main orchestration file
2. **Dockerfile.api** - Go application container
3. **Dockerfile.web** - Frontend static files container
4. **Dockerfile.hardhat** - Hardhat testnet container
5. **nginx.conf** - Reverse proxy configuration

### Required Scripts
1. **deploy.sh** - Main deployment orchestration
2. **setup-env.sh** - Environment variable generation
3. **init-db.sh** - Database schema initialization
4. **deploy-contracts.sh** - Smart contract deployment
5. **health-check.sh** - System health verification

### Configuration Files
1. **.env.example** - Environment variable template
2. **development.env** - Development configuration
3. **production.env** - Production configuration
4. **nginx.conf** - Web server configuration

## Implementation Timeline

1. **Create Docker configurations** (2-3 hours)
2. **Implement deployment scripts** (1-2 hours)
3. **Configure environment management** (1 hour)
4. **Set up monitoring and health checks** (1 hour)
5. **Testing and validation** (1-2 hours)

**Total Estimated Time: 6-9 hours**

## Next Steps

1. Review and approve this deployment plan
2. Implement Docker configurations and scripts
3. Create environment templates and examples
4. Test deployment in development environment
5. Validate all components are working correctly

This plan provides a complete, production-ready deployment strategy that can scale from development testing to production deployment while maintaining security and reliability standards.
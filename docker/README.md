# Docker Deployment for Autonomous Content Service

This directory contains the Docker configuration for deploying the Autonomous Content Service.

## Quick Start

1. **Setup Environment**
   ```bash
   ./scripts/setup.sh
   ```

2. **Configure Environment Variables**
   Edit `docker/.env` file with your configuration:
   - `JWT_SECRET`: Set a secure secret key
   - `DB_PASSWORD`: Set a secure database password
   - `LLM_API_KEY`: Add your OpenAI API key
   - Other configuration as needed

3. **Deploy Services**
   ```bash
   ./scripts/deploy.sh
   ```

The development compose file mounts `.air.toml` and `Caddyfile.dev` for hot reload:
```bash
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

## Services

The deployment includes the following services:

- **PostgreSQL**: Primary database
- **Redis**: Caching and session management
- **Hardhat**: Local Ethereum testnet
- **API**: Go backend service
- **Caddy**: Web server and reverse proxy with automatic SSL

## Management Commands

```bash
# View service status
./scripts/deploy.sh status

# View logs
./scripts/deploy.sh logs [service_name]

# Stop services
./scripts/deploy.sh stop

# Remove all services and data
./scripts/deploy.sh clean

# Health check
./scripts/health-check.sh
```

## Development Mode

For development with hot reload:

```bash
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

## Ports

- Web Interface: http://localhost (ports 80/443)
- API: http://localhost/api
- PostgreSQL: localhost:5432 (dev only)
- Redis: localhost:6379 (dev only)
- Hardhat: localhost:8545

## Troubleshooting

1. **Services not starting**: Check logs with `./scripts/deploy.sh logs [service_name]`
2. **Database connection issues**: Ensure PostgreSQL is healthy before API starts
3. **Contract deployment fails**: Check Hardhat logs and ensure contracts compile
4. **SSL issues**: In production, ensure domain is correctly configured

## Production Deployment

For production deployment:

1. Update `.env` with production values
2. Set `APP_ENV=production`
3. Configure real domain in `DOMAIN` variable
4. Ensure firewall allows ports 80/443
5. Run deployment script
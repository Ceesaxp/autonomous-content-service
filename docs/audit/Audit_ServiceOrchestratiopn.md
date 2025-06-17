# Audit Report: Service Orchestration Implementation vs. Documentation

This audit report summarizes the alignment and gaps between the Docker Compose
implementation in the `docker/` folder and the Service Orchestration specification
in `docs/ServiceOrchestration.md`.

## 1. Core “microservices” Compose file

**Docs expect:**
```text
├── docker/
│   ├── docker-compose.microservices.yml
│   ├── docker-compose.monitoring.yml
│   ├── docker-compose.dev.yml
│   └── services/    # Dockerfiles for each service
```

**Implementation:**
- `docker/docker-compose.microservices.yml`: present, defines both infrastructure
  services (postgres, redis, consul, prometheus, grafana) and business services
  (api-gateway, content-service, decision-service, etc.) with Dockerfiles in
  `docker/services/`.
- `docker/services/`: contains Dockerfiles for every service listed in the docs.

**Gaps:**
- Missing `docker-compose.monitoring.yml` (monitoring services are embedded in the microservices file).
- Missing override files: `docker-compose.test.yml` and `docker-compose.prod.yml`.
- `temporal` container not defined in any Compose file.

## 2. Development override (`docker-compose.dev.yml`)

**Docs snippet:**
```bash
# Start all services
docker-compose -f docker-compose.microservices.yml -f docker-compose.dev.yml up
```

**Implementation:**
- `docker/docker-compose.dev.yml` configures hot reload for Go services and exposes
  dev ports for Postgres/Redis. Matches the docs for local development.

## 3. “Simple” / single-host Compose (undocumented)

**Implementation:**
- Files: `docker/docker-compose.simple.yml`, `docker/docker-compose.yml`, `docker/.env.example`,
  and helper scripts under `docker/scripts/`.
- `docker/README.md` describes a simple single-server variant.

**Gap:**
- Not mentioned in `docs/ServiceOrchestration.md`.

## 4. Environment variable templates

- `.env.microservices.example`: environment template for microservices stack (repo root).
- `docker/.env.example`: environment template for the simple deployment.

**Recommendation:**
- Consolidate or clarify the location of the microservices env template (update references or move file).

## 5. Monitoring configuration

- No standalone `docker-compose.monitoring.yml`; monitoring services are part of the microservices compose.

## 6. Missing test and prod overrides

**Docs reference:**
```bash
# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Production
APP_ENV=production docker-compose -f docker-compose.microservices.yml -f docker-compose.prod.yml up
```

**Gap:**
- Override files `docker-compose.test.yml` and `docker-compose.prod.yml` are not implemented.

## 7. Temporal server not defined

- `Temporal` is listed as a platform service in the docs but missing from all compose definitions.

## 8. Scripts for orchestration

- `docker/scripts/` contains helpers to start, health-check, deploy, and set up the stacks.
- These scripts align with docs’ workflow steps but are not mentioned in `docs/ServiceOrchestration.md`.

---

## Summary of Recommendations

1. Remove or reconcile references to missing Compose files in the docs.
2. Add a Temporal container to the microservices stack or remove it from the docs.
3. Clarify and consolidate environment-template file locations.
4. Document the simple stack in the Service Orchestration doc or reference `docker/README.md`.
5. Pin monitoring image versions and consider resource constraints.
6. Add missing test/prod override files if desired, or trim the documentation accordingly.
7. Consolidate the microservices README and the main Service Orchestration doc to avoid duplication.
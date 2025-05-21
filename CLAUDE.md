# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Autonomous Content Creation Service is a digital-native business that functions without human intervention, leveraging Large Language Models (LLMs) for decision-making, content creation, and autonomous operations.

## Project Architecture

The system consists of the following major components:

1. **Cognitive Engine**: Handles LLM integration, context management, and content creation
2. **API Layer**: REST API for service interaction
3. **Domain Model**: Core business entities and relationships
4. **Repository Layer**: Data access and persistence
5. **Services Layer**: Business logic implementation

## Development Setup

To work with this codebase, you'll need:

1. Go 1.18+
2. PostgreSQL database
3. Environment variables (see .env.example)

## Common Commands

### Building and Running

```bash
# Build the service
go build -o content-service ./src

# Run the service
./content-service

# Run with environment file
source .env && ./content-service
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./src/services/content_creation/...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Database Management

```bash
# Create database (PostgreSQL)
createdb contentservice

# Run migrations
psql contentservice < src/infrastructure/database/schema.sql
```

### API Testing

The service exposes a REST API. You can test it with:

```bash
# Test content creation endpoint
curl -X POST http://localhost:8080/api/v1/projects/{project_id}/content -H "Content-Type: application/json" -d '{"title":"Sample Content","type":"blog_post","target_audience":"developers"}'
```

## Code Patterns and Conventions

### Directory Structure

- `/src`: Main source code
  - `/api`: API handlers and routing
  - `/config`: Configuration loading
  - `/domain`: Core domain model (entities, events)
  - `/infrastructure`: Database and external integrations
  - `/services`: Business logic implementation

### Interfaces

Key interfaces in the system:

- `LLMClient`: Interaction with language models
- `ContextManager`: Manages context for LLM prompts
- `QualityChecker`: Validates content quality
- `Repository`: Data access interfaces for each entity

### Error Handling

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Configuration

The application uses environment variables for configuration with the following pattern:

```go
config.LoadConfig() // Loads from environment or .env file
```

### Dependency Injection

The codebase uses constructor-based dependency injection:

```go
func NewContentPipeline(repo Repository, client LLMClient, ...) *ContentPipeline {
    return &ContentPipeline{...}
}
```
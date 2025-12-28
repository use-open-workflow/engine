# Open Workflow Engine - Project Architecture

## Overview
Open Workflow is an open-source no-code workflow engine written in Go. It follows hexagonal (ports and adapters) architecture with DDD patterns.

## Technology Stack
- **Language**: Go 1.25.0
- **Web Framework**: Fiber v3 (gofiber/fiber/v3)
- **Database**: PostgreSQL (jackc/pgx/v5)
- **ID Generation**: ULID (oklog/ulid/v2)

## Project Structure

```
cmd/api/          - Application entry point (main.go)
api/              - HTTP handlers and routing
  router.go       - SetupRouter function, routes under /api/v1
  node/http/      - HTTP handlers per domain
di/               - Dependency injection container (Container struct)
internal/
  domain/         - Domain aggregates, events, and business logic
  port/           - Port interfaces (inbound services, outbound repositories)
  adapter/        - Adapter implementations
pkg/              - Shared packages (domain base types, ID generation)
migration/        - Database migrations
```

## Hexagonal Architecture Layers

### 1. Domain Layer (`internal/domain/`)
Contains pure business logic with no external dependencies.

**Current Domains:**
- `node/` - Node template domain
  - `aggregate/NodeTemplate` - Main aggregate with Name field, embeds BaseAggregate
  - `aggregate/NodeTemplateFactory` - Factory for creating NodeTemplate aggregates
  - `event/CreateNodeTemplate` - Domain event for creation
  - `event/UpdateNodeTemplate` - Domain event for updates

### 2. Port Layer (`internal/port/`)
Defines interfaces for both inbound (services) and outbound (repositories) operations.

**Inbound Ports** (`port/node/inbound/`):
- `NodeTemplateReadService` interface: List(), GetByID()
- `NodeTemplateWriteService` interface: Create(), Update(), Delete()
- `NodeTemplateDTO` - Data transfer object (ID, Name)
- Input structs: CreateNodeTemplateInput, UpdateNodeTemplateInput

**Outbound Ports** (`port/node/outbound/`):
- `NodeTemplateReadRepository` interface: FindMany(), FindByID()
- `NodeTemplateWriteRepository` interface: Save(), Update(), Delete()
- `NodeTemplateModel` - Database model (ID, Name, CreatedAt, UpdatedAt)
- Repository factories for dependency injection

**Shared Outbound Ports** (`port/outbound/`):
- `UnitOfWork` interface: Begin(), Commit(), Rollback(), RegisterNew/Dirty/Deleted(), Querier()
- `UnitOfWorkFactory` interface: Create()
- `Querier`, `Rows`, `Row`, `CommandTag` - Database abstraction interfaces
- Outbox pattern: OutboxMessage, OutboxProcessor, OutboxWriteRepository, OutboxReadRepository

### 3. Adapter Layer (`internal/adapter/`)
Implements port interfaces.

**Inbound Adapters** (`adapter/node/inbound/`):
- `NodeTemplateWriteService` struct - implements write operations with UoW pattern
- `NodeTemplateReadService` struct - implements read operations
- `NodeTemplateMapper` - converts between DTO and aggregate

**Outbound Adapters** (`adapter/node/outbound/`):
- `NodeTemplatePostgresWriteRepository` - PostgreSQL write repository
- `NodeTemplatePostgresReadRepository` - PostgreSQL read repository
- `NodeTemplateStaticReadRepository` / `NodeTemplateStaticWriteRepository` - in-memory implementations
- Repository factories for UoW-scoped repositories

**Shared Adapters** (`adapter/outbound/`):
- `UnitOfWorkPostgres` - PostgreSQL Unit of Work implementation with transaction management
- `UnitOfWorkPostgresFactory` - Creates UoW instances
- `OutboxProcessor` - Processes outbox messages
- Outbox repositories for event persistence

### 4. API Layer (`api/`)
HTTP handlers using Fiber framework.

- `SetupRouter()` - Creates Fiber app with middleware (recover, logger), routes under `/api/v1`
- `NodeTemplateHandler` - HTTP handler with List, GetByID, Create, Update, Delete methods

### 5. Dependency Injection (`di/`)
- `Container` struct - Holds all dependencies (Pool, services, OutboxProcessor)
- `NewContainer()` - Wires up all dependencies
- `Close()` - Cleanup resources

## Shared Packages (`pkg/`)

### Domain Base Types (`pkg/domain/`)
- **Interfaces**: Aggregate, Entity, Event, ValueObject
- **Base Structs**:
  - `BaseAggregate` - ID + events slice, methods: AddEvent(), Events(), ClearEvents()
  - `BaseEntity` - ID field
  - `BaseEvent` - id, aggregateID, aggregateType, eventType, occurredAt, version
  - `BaseValueObject` - marker interface implementation

### ID Generation (`pkg/id/`)
- `Factory` interface with New() method
- `ULIDFactory` - Generates ULIDs for aggregate/event IDs

## Key Patterns

### Unit of Work Pattern
- Services use `UnitOfWorkFactory` to create transaction-scoped UoW
- Repositories are scoped to UoW via factory pattern
- UoW tracks new/dirty/deleted aggregates
- Domain events are persisted to outbox on commit

### Outbox Pattern
- Domain events stored in outbox table during transaction
- `OutboxProcessor` publishes events asynchronously
- Ensures at-least-once delivery of domain events

### Repository Factory Pattern
- Repositories are created per-UoW for transaction scoping
- Factory interfaces: `NodeTemplateReadRepositoryFactory`, `NodeTemplateWriteRepositoryFactory`

## Commands
- `make test` - Run all tests
- `make build` - Build to `bin/api`
- `make run` - Run on port 3000
- `make fmt` - Format code
- `make clean` - Remove build artifacts

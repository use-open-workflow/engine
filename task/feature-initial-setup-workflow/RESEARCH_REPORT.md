# Research Report

## 1. Relevant Files

List files that will be modified or referenced, grouped by purpose:

```
Entry Points:
- api/router.go — add workflow routes registration
- api/workflow/http/workflow_handler.go — (new) HTTP handlers for workflow CRUD
- api/workflow/http/node_definition_handler.go — (new) HTTP handlers for node-definition CRUD
- api/workflow/http/edge_handler.go — (new) HTTP handlers for edge CRUD

Services/Logic:
- internal/domain/workflow/aggregate/workflow.go — (new) Workflow aggregate with NodeDefinition and Edge entities
- internal/domain/workflow/aggregate/workflow_factory.go — (new) Factory for Workflow creation
- internal/domain/workflow/entity/node_definition.go — (new) NodeDefinition entity
- internal/domain/workflow/entity/edge.go — (new) Edge entity
- internal/domain/workflow/event/*.go — (new) Domain events for workflow, node-definition, edge

Ports (Inbound):
- internal/port/workflow/inbound/workflow_dto.go — (new) DTOs for workflow, node-definition, edge
- internal/port/workflow/inbound/workflow_read_service.go — (new) Read service interface
- internal/port/workflow/inbound/workflow_write_service.go — (new) Write service interface
- internal/port/workflow/inbound/workflow_mapper.go — (new) Mapper interface

Ports (Outbound):
- internal/port/workflow/outbound/workflow_model.go — (new) Database models
- internal/port/workflow/outbound/workflow_read_repository.go — (new) Read repository interface
- internal/port/workflow/outbound/workflow_write_repository.go — (new) Write repository interface
- internal/port/workflow/outbound/*_factory.go — (new) Repository factories

Adapters (Inbound):
- internal/adapter/workflow/inbound/workflow_read_service.go — (new) Read service implementation
- internal/adapter/workflow/inbound/workflow_write_service.go — (new) Write service implementation
- internal/adapter/workflow/inbound/workflow_mapper.go — (new) Mapper implementation

Adapters (Outbound):
- internal/adapter/workflow/outbound/workflow_postgres_read_repository.go — (new) PostgreSQL read repo
- internal/adapter/workflow/outbound/workflow_postgres_write_repository.go — (new) PostgreSQL write repo
- internal/adapter/workflow/outbound/*_factory.go — (new) Repository factory implementations

Data Layer:
- migration/002_workflow_schema.sql — (new) Database schema for workflow, node_definition, edge

DI Container:
- di/container.go — register workflow services and repositories

Tests:
- internal/domain/workflow/aggregate/workflow_test.go — (new) Unit tests for aggregate
```

## 2. Dependencies & Integrations

- **Internal modules:**

  - `pkg/domain` — BaseAggregate, BaseEntity, BaseEvent base types
  - `pkg/id` — ULIDFactory for ID generation
  - `internal/port/outbound` — UnitOfWork, UnitOfWorkFactory, Querier interfaces
  - `internal/adapter/outbound` — UnitOfWorkPostgresFactory
  - `internal/domain/node/aggregate` — NodeTemplate (referenced by NodeDefinition)

- **External services/APIs:**

  - PostgreSQL via `jackc/pgx/v5`
  - Fiber v3 for HTTP handlers

- **Shared utilities:**
  - `pkg/id.Factory` — ULID generation
  - `pkg/domain.BaseAggregate` — embed in Workflow
  - `pkg/domain.BaseEntity` — embed in NodeDefinition, Edge
  - `pkg/domain.BaseEvent` — embed in domain events

## 3. Data Flow

```
HTTP Request → WorkflowHandler → WorkflowService → UnitOfWork → WorkflowRepository → PostgreSQL
                                      ↓
                              WorkflowFactory (create)
                              WorkflowAggregate (domain logic)
                                      ↓
                              Domain Events → Outbox (on commit)
```

Workflow aggregate owns NodeDefinition and Edge entities. NodeDefinition references NodeTemplate by ID. Edge connects two NodeDefinition IDs (from/to).

## 4. Impact Areas

**Direct modifications required:**

- `api/router.go` — add `registerWorkflowRoutes()` function
- `di/container.go` — add WorkflowReadService, WorkflowWriteService fields and wiring
- `migration/` — new migration file for workflow tables

**Indirect impacts:**

- Outbox pattern handles domain event publishing (no changes needed)
- NodeTemplate read is required when validating NodeDefinition creation

## 5. Implementation Constraints

**Coding patterns to follow:**

- Embed `domain.BaseAggregate` in Workflow, `domain.BaseEntity` in NodeDefinition/Edge
- Use Factory pattern for aggregate creation (see `NodeTemplateFactory`)
- Domain events created in aggregate methods, persisted via UoW outbox
- Repositories use UoW.Querier(ctx) for database access
- Repository factories create UoW-scoped repository instances
- Mappers convert between Aggregate ↔ DTO (inbound) and Aggregate ↔ Model (outbound)

**Validation/business rules:**

- NodeDefinition must reference a valid NodeTemplate ID
- Edge from/to must reference valid NodeDefinition IDs within the same Workflow
- Workflow name should be non-empty

**Testing requirements:**

- Unit tests for aggregate business logic
- Integration tests for repositories (optional for MVP)

**Database conventions:**

- Table names: snake_case singular (`workflow`, `node_definition`, `edge`)
- Primary keys: `VARCHAR(26)` for ULID
- Timestamps: `created_at`, `updated_at` with `TIMESTAMP WITH TIME ZONE`
- Foreign key constraints between tables

## 6. Reference Implementations

1. **NodeTemplate domain** — Full implementation of aggregate, factory, events, ports, adapters, and handler:

   - [internal/domain/node/aggregate/node_template.go](internal/domain/node/aggregate/node_template.go)
   - [internal/adapter/node/inbound/node_template_write_service.go](internal/adapter/node/inbound/node_template_write_service.go)
   - [api/node/http/node_template_handler.go](api/node/http/node_template_handler.go)

2. **Unit of Work pattern** — Transaction management with outbox:
   - [internal/adapter/outbound/unit_of_work_postgres.go](internal/adapter/outbound/unit_of_work_postgres.go)

# Research Report: Workflow Domain

## 1. Relevant Files

### Entry Points (to create)
- [api/workflow/http/workflow_handler.go](api/workflow/http/workflow_handler.go) — HTTP handlers for workflow, node-definition, edge
- [api/router.go](api/router.go) — Register new workflow routes (modify)

### Domain Layer (to create)
- `internal/domain/workflow/aggregate/workflow.go` — Workflow aggregate with NodeDefinitions and Edges
- `internal/domain/workflow/aggregate/workflow_factory.go` — Factory for creating Workflow
- `internal/domain/workflow/aggregate/node_definition.go` — NodeDefinition entity
- `internal/domain/workflow/aggregate/edge.go` — Edge entity
- `internal/domain/workflow/event/create_workflow.go` — Domain event for workflow creation
- `internal/domain/workflow/event/update_workflow.go` — Domain event for workflow updates

### Port Layer - Inbound (to create)
- `internal/port/workflow/inbound/workflow_dto.go` — DTO for Workflow, NodeDefinition, Edge
- `internal/port/workflow/inbound/workflow_read_service.go` — Read service interface
- `internal/port/workflow/inbound/workflow_write_service.go` — Write service interface with input structs
- `internal/port/workflow/inbound/workflow_mapper.go` — Mapper interface (aggregate → DTO)

### Port Layer - Outbound (to create)
- `internal/port/workflow/outbound/workflow_model.go` — Database models
- `internal/port/workflow/outbound/workflow_read_repository.go` — Read repository interface
- `internal/port/workflow/outbound/workflow_write_repository.go` — Write repository interface
- `internal/port/workflow/outbound/workflow_read_repository_factory.go` — Factory interface
- `internal/port/workflow/outbound/workflow_write_repository_factory.go` — Factory interface

### Adapter Layer - Inbound (to create)
- `internal/adapter/workflow/inbound/workflow_read_service.go` — Read service implementation
- `internal/adapter/workflow/inbound/workflow_write_service.go` — Write service implementation
- `internal/adapter/workflow/inbound/workflow_mapper.go` — Mapper implementation

### Adapter Layer - Outbound (to create)
- `internal/adapter/workflow/outbound/workflow_postgres_read_repository.go` — PostgreSQL read impl
- `internal/adapter/workflow/outbound/workflow_postgres_write_repository.go` — PostgreSQL write impl
- `internal/adapter/workflow/outbound/workflow_postgres_read_repository_factory.go` — Factory impl
- `internal/adapter/workflow/outbound/workflow_postgres_write_repository_factory.go` — Factory impl

### Infrastructure (to modify/create)
- [di/container.go](di/container.go) — Add workflow services to DI container
- `migration/002_workflow_schema.sql` — Database schema for workflow tables

### Reference Files
- [internal/domain/node/aggregate/node_template.go](internal/domain/node/aggregate/node_template.go) — Aggregate pattern reference
- [internal/adapter/node/inbound/node_template_write_service.go](internal/adapter/node/inbound/node_template_write_service.go) — Service pattern reference
- [api/node/http/node_template_handler.go](api/node/http/node_template_handler.go) — Handler pattern reference

---

## 2. Dependencies & Integrations

### Internal Modules
- `pkg/domain` — BaseAggregate, BaseEntity, BaseEvent, Event interface
- `pkg/id` — Factory interface (ULID generation)
- `internal/port/outbound` — UnitOfWork, UnitOfWorkFactory, Querier interfaces

### External Dependencies
- `github.com/gofiber/fiber/v3` — HTTP framework
- `github.com/jackc/pgx/v5/pgxpool` — PostgreSQL connection pool

### Shared Utilities
- `id.Factory` — Generate ULIDs for aggregate/entity/event IDs
- `outbound.UnitOfWork` — Transaction management
- `domain.BaseAggregate` — Embed in Workflow aggregate
- `domain.BaseEntity` — Embed in NodeDefinition and Edge entities

---

## 3. Data Flow

**Create Workflow Flow:**
1. HTTP handler receives POST `/api/v1/workflow` with JSON body
2. Handler calls `WriteService.Create()` with `CreateWorkflowInput`
3. Service creates UoW, begins transaction, uses factory to create Workflow aggregate
4. Repository saves Workflow + child entities (NodeDefinitions, Edges) within transaction
5. UoW commits, domain events go to outbox, DTO returned to handler

**Retrieve Workflow Flow:**
1. HTTP handler receives GET `/api/v1/workflow/:id`
2. Handler calls `ReadService.GetByID()`
3. Service creates UoW, repository reconstructs Workflow with child entities
4. Mapper converts aggregate to DTO, returned to handler

---

## 4. Impact Areas

### Direct Modifications
- [di/container.go](di/container.go) — Wire up workflow services, repositories, factories
- [api/router.go](api/router.go) — Register `/api/v1/workflow` routes

### Database Changes
- New tables: `workflows`, `node_definitions`, `edges`
- Foreign keys: `node_definitions.workflow_id`, `edges.workflow_id`
- Reference: `node_definitions.node_template_id` → `node_templates.id`

### Indirect Impacts
- Outbox table will receive workflow domain events
- Future: NodeTemplate may need validation when referenced by NodeDefinition

---

## 5. Implementation Constraints

### Coding Patterns (from existing code)
- **Aggregate**: Embed `domain.BaseAggregate`, use factory pattern, emit domain events
- **Entity**: Embed `domain.BaseEntity`, contained within aggregate
- **Service**: Use UoW pattern, create UoW per operation, use repository factories
- **Repository**: Accept UoW in constructor, use `uow.Querier(ctx)` for DB operations
- **Handler**: Use `c.Bind().JSON()`, return JSON with `c.JSON()`, use proper HTTP status codes

### Validation Rules
- Workflow must have a name
- NodeDefinition must reference a valid NodeTemplate ID
- Edge must reference valid from/to NodeDefinition IDs within same Workflow
- Edges should form a valid DAG (no cycles) — consider for future validation

### Database Conventions
- Table names: snake_case plural (`workflows`, `node_definitions`, `edges`)
- Primary keys: `VARCHAR(26)` for ULID
- Timestamps: `created_at`, `updated_at` with `TIMESTAMP WITH TIME ZONE`
- Use foreign key constraints

### Testing Requirements
- Unit tests for aggregate creation and business logic
- Integration tests for repositories (if applicable)
- Follow existing test patterns in `*_test.go` files

---

## 6. Reference Implementations

### Primary Reference: NodeTemplate Domain
The `node` domain provides a complete reference for implementing the `workflow` domain:

1. **Aggregate Pattern**: [internal/domain/node/aggregate/node_template.go](internal/domain/node/aggregate/node_template.go)
   - Shows BaseAggregate embedding, factory pattern, Reconstitute function, domain events

2. **Service Pattern**: [internal/adapter/node/inbound/node_template_write_service.go](internal/adapter/node/inbound/node_template_write_service.go)
   - Shows UoW usage, transaction handling, repository factory pattern

3. **Handler Pattern**: [api/node/http/node_template_handler.go](api/node/http/node_template_handler.go)
   - Shows Fiber v3 handler structure, JSON binding, error responses

### Key Differences for Workflow
- Workflow is an aggregate root containing child entities (NodeDefinition, Edge)
- Need to handle parent-child relationships in repository (save/load together)
- May need composite operations (add/remove nodes and edges)

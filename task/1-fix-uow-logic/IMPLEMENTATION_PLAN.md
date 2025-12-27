# Implementation Plan: Fix UoW Logic

## 1. Implementation Summary

The solution replaces the dual-UoW architecture with a single UoW instance per request by introducing a `WriteRepositoryFactory` pattern. The service creates a UoW, passes it to a factory that returns a repository bound to that UoW, ensuring all aggregate registrations and event persistence happen within the same transaction. This eliminates the data inconsistency bug where `sharedUow` holds aggregates but the service-level UoW commits the transaction.

---

## 2. Change Manifest

```
CREATE:
- internal/port/node/outbound/node_template_write_repository_factory.go — Factory interface for creating UoW-bound write repositories

MODIFY:
- internal/adapter/node/outbound/node_template_postgres_write_repository.go — Accept UoW as parameter, remove pool reference for writes
- internal/adapter/node/inbound/node_template_write_service.go — Use factory to create repository per operation
- internal/port/outbound/unit_of_work.go — Add Querier() method to interface for repository access
- internal/adapter/outbound/unit_of_work_postgres.go — Implement Querier() method
- di/container.go — Remove sharedUow, create write repository factory, update service wiring
```

---

## 3. Step-by-Step Plan

### Step 1: Extend UnitOfWork Interface with Querier Method

**File:** `internal/port/outbound/unit_of_work.go`

**Action:** MODIFY

**Rationale:** The write repository needs to execute SQL queries within the transaction managed by the UoW. By adding a `Querier()` method, the repository can access the transaction without needing direct knowledge of the UoW implementation.

**Pseudocode:**

```go
package outbound

import "context"

// Querier abstracts query execution for both pool and transaction
type Querier interface {
    Query(ctx context.Context, sql string, args ...any) (Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) Row
    Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

type UnitOfWork interface {
    Begin(ctx context.Context) (context.Context, error)
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    RegisterNew(aggregate any)
    RegisterDirty(aggregate any)
    RegisterDeleted(aggregate any)

    // NEW: Get the querier (tx if in transaction, or pool otherwise)
    Querier(ctx context.Context) Querier
}

type UnitOfWorkFactory interface {
    Create() UnitOfWork
}
```

**Dependencies:**
- `github.com/jackc/pgx/v5` (for Row, Rows, CommandTag types)

**Tests Required:**
- Test that Querier returns tx when transaction is active
- Test that Querier returns pool when no transaction

---

### Step 2: Implement Querier Method in UnitOfWorkPostgres

**File:** `internal/adapter/outbound/unit_of_work_postgres.go`

**Action:** MODIFY

**Rationale:** Implement the new interface method to provide the repository access to the transaction querier.

**Pseudocode:**

```go
// Add to UnitOfWorkPostgres

func (u *UnitOfWorkPostgres) Querier(ctx context.Context) querier {
    // 1. Check if transaction exists in context
    if tx, ok := u.GetTx(ctx); ok {
        // 2. Return transaction as querier
        return tx
    }
    // 3. Fallback to pool (for cases outside transaction, though shouldn't happen for writes)
    return u.pool
}

// Note: The querier interface already exists in this file (lines 36-40)
// We reuse it here
```

**Dependencies:**
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgxpool`

**Tests Required:**
- Test Querier returns transaction when Begin() was called
- Test Querier returns pool when no transaction in context

---

### Step 3: Create WriteRepositoryFactory Interface

**File:** `internal/port/node/outbound/node_template_write_repository_factory.go`

**Action:** CREATE

**Rationale:** Define the port interface for creating write repositories bound to a specific UoW. This follows the existing factory pattern and keeps the hexagonal architecture clean.

**Pseudocode:**

```go
package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// NodeTemplateWriteRepositoryFactory creates write repositories bound to a UoW
type NodeTemplateWriteRepositoryFactory interface {
    // Create returns a NodeTemplateWriteRepository that uses the given UoW
    // for transaction management and aggregate registration
    //
    // The returned repository:
    //   - Uses uow.Querier() for SQL execution
    //   - Calls uow.RegisterNew/RegisterDirty/RegisterDeleted
    //   - Does NOT manage transaction lifecycle (that's the service's job)
    Create(uow portOutbound.UnitOfWork) NodeTemplateWriteRepository
}
```

**Dependencies:**
- `internal/port/outbound` - for UnitOfWork interface
- `internal/port/node/outbound` - for NodeTemplateWriteRepository interface

**Tests Required:**
- Mock tests to verify factory creates repository with correct UoW

---

### Step 4: Refactor Write Repository to Accept UoW

**File:** `internal/adapter/node/outbound/node_template_postgres_write_repository.go`

**Action:** MODIFY

**Rationale:** Remove the stored UoW reference. Instead, the repository receives UoW at query time through a factory pattern. Also create the factory implementation in the same file.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/node/aggregate"
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

// NodeTemplatePostgresWriteRepository implements NodeTemplateWriteRepository
// It is bound to a specific UoW instance for the duration of a request
type NodeTemplatePostgresWriteRepository struct {
    uow portOutbound.UnitOfWork  // The UoW this repository is bound to
}

// NewNodeTemplatePostgresWriteRepository creates a repository bound to the given UoW
func NewNodeTemplatePostgresWriteRepository(
    uow portOutbound.UnitOfWork,
) *NodeTemplatePostgresWriteRepository {
    return &NodeTemplatePostgresWriteRepository{
        uow: uow,
    }
}

func (r *NodeTemplatePostgresWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    // 1. Get querier from UoW (will be the transaction)
    q := r.uow.Querier(ctx)

    // 2. Execute insert
    _, err := q.Exec(ctx, `
        INSERT INTO node_templates (id, name, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
    `, nodeTemplate.ID, nodeTemplate.Name)

    if err != nil {
        return fmt.Errorf("failed to save node template: %w", err)
    }

    // 3. Register with the SAME UoW that was passed in
    r.uow.RegisterNew(nodeTemplate)

    return nil
}

func (r *NodeTemplatePostgresWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    // 1. Get querier from UoW
    q := r.uow.Querier(ctx)

    // 2. Execute update
    _, err := q.Exec(ctx, `
        UPDATE node_templates
        SET name = $1, updated_at = NOW()
        WHERE id = $2
    `, nodeTemplate.Name, nodeTemplate.ID)

    if err != nil {
        return fmt.Errorf("failed to update node template: %w", err)
    }

    // 3. Register dirty with the same UoW
    r.uow.RegisterDirty(nodeTemplate)

    return nil
}

func (r *NodeTemplatePostgresWriteRepository) Delete(ctx context.Context, id string) error {
    // 1. Get querier from UoW
    q := r.uow.Querier(ctx)

    // 2. Execute delete
    _, err := q.Exec(ctx, `
        DELETE FROM node_templates
        WHERE id = $1
    `, id)

    if err != nil {
        return fmt.Errorf("failed to delete node template: %w", err)
    }

    // Note: RegisterDeleted not called here since we don't have the aggregate
    // If events are needed on delete, fetch the aggregate first in service

    return nil
}

// --- Factory Implementation ---

// NodeTemplatePostgresWriteRepositoryFactory creates Postgres write repositories
type NodeTemplatePostgresWriteRepositoryFactory struct {
    // No dependencies needed - the pool comes via UoW
}

func NewNodeTemplatePostgresWriteRepositoryFactory() *NodeTemplatePostgresWriteRepositoryFactory {
    return &NodeTemplatePostgresWriteRepositoryFactory{}
}

func (f *NodeTemplatePostgresWriteRepositoryFactory) Create(uow portOutbound.UnitOfWork) nodeOutbound.NodeTemplateWriteRepository {
    return NewNodeTemplatePostgresWriteRepository(uow)
}
```

**Dependencies:**
- `internal/port/outbound` - UnitOfWork interface
- `internal/port/node/outbound` - NodeTemplateWriteRepository interface
- `internal/domain/node/aggregate` - NodeTemplate aggregate

**Tests Required:**
- Test Save registers aggregate with provided UoW
- Test Update registers aggregate with provided UoW
- Test Delete executes within transaction
- Test factory creates repository with correct UoW binding

---

### Step 5: Update Write Service to Use Repository Factory

**File:** `internal/adapter/node/inbound/node_template_write_service.go`

**Action:** MODIFY

**Rationale:** Replace the static writeRepository with a writeRepositoryFactory. For each operation, create a UoW and then create a repository bound to that UoW.

**Pseudocode:**

```go
package inbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/node/aggregate"
    "use-open-workflow.io/engine/internal/port/node/inbound"
    nodeOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
    "use-open-workflow.io/engine/pkg/id"
)

type NodeTemplateWriteService struct {
    uowFactory             outbound.UnitOfWorkFactory
    writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory  // CHANGED: factory instead of instance
    factory                *aggregate.NodeTemplateFactory
    readRepository         nodeOutbound.NodeTemplateReadRepository
    mapper                 inbound.NodeTemplateMapper
    idFactory              id.Factory
}

func NewNodeTemplateWriteService(
    uowFactory outbound.UnitOfWorkFactory,
    writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory,  // CHANGED
    factory *aggregate.NodeTemplateFactory,
    readRepository nodeOutbound.NodeTemplateReadRepository,
    mapper inbound.NodeTemplateMapper,
    idFactory id.Factory,
) *NodeTemplateWriteService {
    return &NodeTemplateWriteService{
        uowFactory:             uowFactory,
        writeRepositoryFactory: writeRepositoryFactory,  // CHANGED
        factory:                factory,
        readRepository:         readRepository,
        mapper:                 mapper,
        idFactory:              idFactory,
    }
}

func (s *NodeTemplateWriteService) Create(ctx context.Context, input inbound.CreateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
    // 1. Create UoW for this request
    uow := s.uowFactory.Create()

    // 2. Create repository bound to THIS UoW (KEY CHANGE)
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // 3. Begin transaction
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

    // 4. Defer rollback on error
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 5. Create aggregate via factory
    nodeTemplate := s.factory.Make(input.Name)

    // 6. Save using the UoW-bound repository
    //    This registers the aggregate with the SAME UoW
    if err = writeRepo.Save(txCtx, nodeTemplate); err != nil {
        return nil, fmt.Errorf("failed to save node template: %w", err)
    }

    // 7. Commit - this will persist outbox events from registered aggregates
    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    // 8. Map and return
    return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Update(ctx context.Context, id string, input inbound.UpdateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
    // 1. Create UoW for this request
    uow := s.uowFactory.Create()

    // 2. Create repository bound to THIS UoW
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // 3. Begin transaction
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 4. Fetch existing aggregate
    nodeTemplate, err := s.readRepository.FindByID(txCtx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find node template: %w", err)
    }
    if nodeTemplate == nil {
        return nil, fmt.Errorf("node template not found: %s", id)
    }

    // 5. Apply domain logic (adds events to aggregate)
    nodeTemplate.UpdateName(s.idFactory, input.Name)

    // 6. Update using UoW-bound repository
    if err = writeRepo.Update(txCtx, nodeTemplate); err != nil {
        return nil, fmt.Errorf("failed to update node template: %w", err)
    }

    // 7. Commit
    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Delete(ctx context.Context, id string) error {
    // 1. Create UoW for this request
    uow := s.uowFactory.Create()

    // 2. Create repository bound to THIS UoW
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // 3. Begin transaction
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 4. Delete using UoW-bound repository
    if err = writeRepo.Delete(txCtx, id); err != nil {
        return fmt.Errorf("failed to delete node template: %w", err)
    }

    // 5. Commit
    if err = uow.Commit(txCtx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

**Dependencies:**
- `internal/port/outbound` - UnitOfWorkFactory
- `internal/port/node/outbound` - NodeTemplateWriteRepositoryFactory, NodeTemplateReadRepository
- `internal/domain/node/aggregate` - NodeTemplateFactory

**Tests Required:**
- Test Create uses single UoW for entire operation
- Test Update uses single UoW for entire operation
- Test Delete uses single UoW for entire operation
- Test rollback happens on repository failure
- Test domain events are persisted on commit

---

### Step 6: Update DI Container

**File:** `di/container.go`

**Action:** MODIFY

**Rationale:** Remove `sharedUow`, create the write repository factory, and update service wiring.

**Pseudocode:**

```go
package di

import (
    "context"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    nodeAdapterInbound "use-open-workflow.io/engine/internal/adapter/node/inbound"
    nodeAdapterOutbound "use-open-workflow.io/engine/internal/adapter/node/outbound"
    adapterOutbound "use-open-workflow.io/engine/internal/adapter/outbound"
    "use-open-workflow.io/engine/internal/domain/node/aggregate"
    "use-open-workflow.io/engine/internal/port/node/inbound"
    "use-open-workflow.io/engine/internal/port/outbound"
    "use-open-workflow.io/engine/pkg/id"
)

type Container struct {
    Pool                     *pgxpool.Pool
    NodeTemplateReadService  inbound.NodeTemplateReadService
    NodeTemplateWriteService inbound.NodeTemplateWriteService
    OutboxProcessor          outbound.OutboxProcessor
}

func NewContainer(ctx context.Context) (*Container, error) {
    // 1. Database connection (unchanged)
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        databaseURL = "postgres://postgres:postgres@localhost:5432/open_workflow?sslmode=disable"
    }

    pool, err := pgxpool.New(ctx, databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // 2. Shared dependencies (unchanged)
    idFactory := id.NewULIDFactory()

    // 3. Unit of Work Factory (unchanged)
    uowFactory := adapterOutbound.NewUnitOfWorkPostgresFactory(pool)

    // 4. Mappers (unchanged)
    nodeTemplateInboundMapper := nodeAdapterInbound.NewNodeTemplateMapper()

    // 5. Aggregate Factory (unchanged)
    nodeTemplateFactory := aggregate.NewNodeTemplateFactory(idFactory)

    // 6. Read Repository (unchanged)
    nodeTemplateReadRepository := nodeAdapterOutbound.NewNodeTemplatePostgresReadRepository(pool)

    // 7. REMOVED: sharedUow - no longer needed
    // 8. REMOVED: nodeTemplateWriteRepository - replaced by factory

    // 9. NEW: Write Repository Factory
    nodeTemplateWriteRepositoryFactory := nodeAdapterOutbound.NewNodeTemplatePostgresWriteRepositoryFactory()

    // 10. Services
    nodeTemplateReadService := nodeAdapterInbound.NewNodeTemplateReadService(
        nodeTemplateInboundMapper,
        nodeTemplateReadRepository,
    )

    // 11. UPDATED: Pass factory instead of repository
    nodeTemplateWriteService := nodeAdapterInbound.NewNodeTemplateWriteService(
        uowFactory,
        nodeTemplateWriteRepositoryFactory,  // CHANGED: factory instead of instance
        nodeTemplateFactory,
        nodeTemplateReadRepository,
        nodeTemplateInboundMapper,
        idFactory,
    )

    // 12. Outbox processor (unchanged)
    outboxReadRepository := adapterOutbound.NewOutboxPostgresReadRepository(pool)
    outboxWriteRepository := adapterOutbound.NewOutboxPostgresWriteRepository(pool)
    eventPublisher := adapterOutbound.NewOutboxNoopEventPublisher()
    outboxProcessor := adapterOutbound.NewOutboxProcessor(
        outboxReadRepository,
        outboxWriteRepository,
        eventPublisher,
        adapterOutbound.DefaultConfig(),
    )

    return &Container{
        Pool:                     pool,
        NodeTemplateReadService:  nodeTemplateReadService,
        NodeTemplateWriteService: nodeTemplateWriteService,
        OutboxProcessor:          outboxProcessor,
    }, nil
}

func (c *Container) Close() {
    if c.OutboxProcessor != nil {
        c.OutboxProcessor.Stop()
    }
    if c.Pool != nil {
        c.Pool.Close()
    }
}
```

**Dependencies:**
- All existing imports remain
- Remove direct write repository creation

**Tests Required:**
- Integration test that container creates working write service
- Test that write operations persist both entity and events atomically

---

## 4. Data Changes

**Schema/Model Updates:**

No database schema changes required. The existing tables (`node_templates`, `outbox`) remain unchanged.

**Migration Notes:**

- No migration needed
- Backward compatibility is maintained - the API behavior is identical, only the internal transaction management changes

---

## 5. Integration Points

| Service | Interaction | Error Handling |
|---------|-------------|----------------|
| PostgreSQL | Transaction management via UoW | Rollback on any error; errors propagate up |
| Outbox | Events persisted in same transaction as entity | If outbox insert fails, entire transaction rolls back |

---

## 6. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Transaction begin fails | Return error immediately, no cleanup needed |
| Repository operation fails | Defer calls Rollback(), error returned to caller |
| Commit fails | Transaction already rolled back by postgres, error returned |
| Concurrent updates | PostgreSQL handles via row-level locking |
| UoW created but never used | No side effects - no transaction started |
| Aggregate without events | No outbox entries created (normal behavior) |
| Delete without aggregate | No RegisterDeleted called (acceptable for hard delete) |

---

## 7. Testing Strategy

**Unit Tests:**

1. `unit_of_work_postgres_test.go`:
   - Test `Querier()` returns tx when transaction active
   - Test `Querier()` returns pool when no transaction
   - Test `RegisterNew/Dirty/Deleted` stores aggregates
   - Test `Commit()` persists events from registered aggregates
   - Test `Rollback()` clears registered aggregates

2. `node_template_postgres_write_repository_test.go`:
   - Test `Save()` executes insert and registers aggregate
   - Test `Update()` executes update and registers aggregate
   - Test `Delete()` executes delete

3. `node_template_write_service_test.go`:
   - Test `Create()` uses single UoW throughout
   - Test `Update()` uses single UoW throughout
   - Test `Delete()` uses single UoW throughout
   - Test rollback on error

**Integration Tests:**

1. `node_template_integration_test.go`:
   - Test create persists entity and outbox event atomically
   - Test update persists entity and outbox event atomically
   - Test failure mid-operation rolls back both entity and events
   - Verify outbox contains correct events after operations

**Manual Verification:**

1. Start the application: `make run`
2. Create a node template: `curl -X POST http://localhost:3000/node-templates -d '{"name":"test"}'`
3. Query outbox table: `SELECT * FROM outbox ORDER BY created_at DESC LIMIT 1;`
4. Verify the outbox event exists with correct aggregate_id
5. Update the node template and verify update event appears
6. Test failure scenario by modifying code to fail after save but before commit, verify no orphaned data

---

## 8. Implementation Order

Recommended sequence for implementation:

1. **Step 1: Extend UnitOfWork Interface** — Foundation for the pattern; no breaking changes
2. **Step 2: Implement Querier Method** — Implements the new interface method
3. **Step 3: Create WriteRepositoryFactory Interface** — Define the new port
4. **Step 4: Refactor Write Repository** — Implements new pattern; can be tested in isolation
5. **Step 5: Update Write Service** — Integrates new pattern; requires Step 3 and 4
6. **Step 6: Update DI Container** — Final wiring; requires all previous steps

This order ensures:
- Each step can be compiled and tested independently
- No intermediate broken states
- Clear rollback points if issues arise

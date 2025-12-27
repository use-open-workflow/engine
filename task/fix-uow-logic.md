# Fix Unit of Work Logic - Implementation Plan

## 1. Implementation Summary

The current implementation has a critical bug where the write repository uses a **shared UoW instance** while the write service creates a **separate UoW instance per operation** via the factory. This causes domain events registered in the repository to never be persisted to the outbox table because they're tracked on different UoW instances.

The fix introduces a **Repository Factory pattern** where repositories are created per-operation with the active UoW injected, ensuring the same UoW instance is used for both transaction management and aggregate registration.

## 2. Change Manifest

```
CREATE:
- internal/port/node/outbound/node_template_write_repository_factory.go — Repository factory interface
- internal/adapter/node/outbound/node_template_postgres_write_repository_factory.go — PostgreSQL factory implementation

MODIFY:
- internal/port/outbound/unit_of_work.go — Add Querier method to interface
- internal/adapter/outbound/unit_of_work_postgres.go — Implement Querier method, remove shared state issue
- internal/adapter/node/outbound/node_template_postgres_write_repository.go — Accept UoW via constructor, remove pool dependency
- internal/adapter/node/inbound/node_template_write_service.go — Use repository factory instead of repository instance
- di/container.go — Wire repository factory instead of shared repository
```

## 3. Step-by-Step Plan

---

### Step 1: Extend UnitOfWork Interface with Querier Method

**File:** `internal/port/outbound/unit_of_work.go`

**Action:** MODIFY

**Rationale:** The UoW needs to expose the active transaction/querier so repositories can execute queries within the transaction scope.

**Pseudocode:**

```go
package outbound

import "context"

// Querier abstracts database query execution (satisfied by pgx.Tx and pgxpool.Pool)
type Querier interface {
    Exec(ctx context.Context, sql string, args ...any) (any, error)
    Query(ctx context.Context, sql string, args ...any) (any, error)
    QueryRow(ctx context.Context, sql string, args ...any) any
}

type UnitOfWork interface {
    // Transaction lifecycle
    Begin(ctx context.Context) (context.Context, error)
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error

    // Aggregate tracking for event collection
    RegisterNew(aggregate any)
    RegisterDirty(aggregate any)
    RegisterDeleted(aggregate any)

    // NEW: Get the active querier (transaction if begun, otherwise pool)
    Querier(ctx context.Context) Querier
}

type UnitOfWorkFactory interface {
    Create() UnitOfWork
}
```

**Dependencies:** None

**Tests Required:**
- Test that Querier returns transaction after Begin
- Test that Querier behavior before Begin (should return pool or error)

---

### Step 2: Implement Querier Method in PostgreSQL UoW

**File:** `internal/adapter/outbound/unit_of_work_postgres.go`

**Action:** MODIFY

**Rationale:** Implement the new Querier method and ensure the UoW properly provides access to the active transaction.

**Pseudocode:**

```go
// Add pool reference for fallback (though typically tx should exist)
type UnitOfWorkPostgres struct {
    pool     *pgxpool.Pool
    newItems []domain.Aggregate
    dirty    []domain.Aggregate
    deleted  []domain.Aggregate
}

// Querier returns the active transaction from context, or pool as fallback
func (u *UnitOfWorkPostgres) Querier(ctx context.Context) outbound.Querier {
    // 1. Try to get transaction from context
    if tx, ok := u.GetTx(ctx); ok {
        return tx  // pgx.Tx implements Querier
    }

    // 2. Fallback to pool (for reads outside transaction)
    //    Note: This should rarely happen in write operations
    return u.pool
}

// GetTx remains for internal use
func (u *UnitOfWorkPostgres) GetTx(ctx context.Context) (pgx.Tx, bool) {
    tx, ok := ctx.Value(txKey).(pgx.Tx)
    return tx, ok
}
```

**Dependencies:**
- `internal/port/outbound` (Querier interface)

**Tests Required:**
- Test Querier returns tx after Begin is called
- Test Querier returns pool when no tx in context

---

### Step 3: Create Repository Factory Interface

**File:** `internal/port/node/outbound/node_template_write_repository_factory.go`

**Action:** CREATE

**Rationale:** Define the factory interface that creates repository instances bound to a specific UoW.

**Pseudocode:**

```go
package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// NodeTemplateWriteRepositoryFactory creates write repositories bound to a UnitOfWork
type NodeTemplateWriteRepositoryFactory interface {
    // Create returns a new repository instance that uses the provided UoW
    // for transaction management and aggregate registration
    Create(uow portOutbound.UnitOfWork) NodeTemplateWriteRepository
}
```

**Dependencies:**
- `internal/port/outbound` (UnitOfWork interface)
- `internal/port/node/outbound` (NodeTemplateWriteRepository interface)

**Tests Required:**
- Test that factory creates repository with correct UoW binding

---

### Step 4: Refactor Write Repository to Accept UoW

**File:** `internal/adapter/node/outbound/node_template_postgres_write_repository.go`

**Action:** MODIFY

**Rationale:** Repository should receive UoW at construction time, not as a shared singleton. This ensures all operations use the same UoW instance.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/node/aggregate"
    "use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplatePostgresWriteRepository struct {
    uow outbound.UnitOfWork  // Injected per-operation, not shared
}

// NewNodeTemplatePostgresWriteRepository creates a repository bound to the given UoW
func NewNodeTemplatePostgresWriteRepository(
    uow outbound.UnitOfWork,
) *NodeTemplatePostgresWriteRepository {
    return &NodeTemplatePostgresWriteRepository{
        uow: uow,
    }
}

func (r *NodeTemplatePostgresWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    // 1. Get querier from UoW (returns active transaction)
    q := r.uow.Querier(ctx)

    // 2. Execute INSERT within transaction
    _, err := q.Exec(ctx, `
        INSERT INTO node_templates (id, name, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
    `, nodeTemplate.ID, nodeTemplate.Name)

    if err != nil {
        return fmt.Errorf("failed to save node template: %w", err)
    }

    // 3. Register aggregate on the SAME UoW for event collection
    r.uow.RegisterNew(nodeTemplate)

    return nil
}

func (r *NodeTemplatePostgresWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    // 1. Get querier from UoW
    q := r.uow.Querier(ctx)

    // 2. Execute UPDATE
    _, err := q.Exec(ctx, `
        UPDATE node_templates
        SET name = $1, updated_at = NOW()
        WHERE id = $2
    `, nodeTemplate.Name, nodeTemplate.ID)

    if err != nil {
        return fmt.Errorf("failed to update node template: %w", err)
    }

    // 3. Register as dirty for event collection
    r.uow.RegisterDirty(nodeTemplate)

    return nil
}

func (r *NodeTemplatePostgresWriteRepository) Delete(ctx context.Context, id string) error {
    // 1. Get querier from UoW
    q := r.uow.Querier(ctx)

    // 2. Execute DELETE
    _, err := q.Exec(ctx, `
        DELETE FROM node_templates
        WHERE id = $1
    `, id)

    if err != nil {
        return fmt.Errorf("failed to delete node template: %w", err)
    }

    // Note: No RegisterDeleted here since we don't have the aggregate
    // If events needed for delete, service should load aggregate first

    return nil
}
```

**Dependencies:**
- `internal/port/outbound` (UnitOfWork interface)

**Tests Required:**
- Test Save registers aggregate as new
- Test Update registers aggregate as dirty
- Test all operations use UoW's querier

---

### Step 5: Create Repository Factory Implementation

**File:** `internal/adapter/node/outbound/node_template_postgres_write_repository_factory.go`

**Action:** CREATE

**Rationale:** Implement the factory that creates properly-bound repository instances.

**Pseudocode:**

```go
package outbound

import (
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
    nodePortOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
)

type NodeTemplatePostgresWriteRepositoryFactory struct{}

func NewNodeTemplatePostgresWriteRepositoryFactory() *NodeTemplatePostgresWriteRepositoryFactory {
    return &NodeTemplatePostgresWriteRepositoryFactory{}
}

// Create returns a new repository bound to the provided UoW
func (f *NodeTemplatePostgresWriteRepositoryFactory) Create(
    uow portOutbound.UnitOfWork,
) nodePortOutbound.NodeTemplateWriteRepository {
    return NewNodeTemplatePostgresWriteRepository(uow)
}
```

**Dependencies:**
- `internal/port/outbound` (UnitOfWork interface)
- `internal/port/node/outbound` (NodeTemplateWriteRepository, Factory interfaces)
- `internal/adapter/node/outbound` (repository implementation)

**Tests Required:**
- Test factory creates repository with correct UoW

---

### Step 6: Update Write Service to Use Repository Factory

**File:** `internal/adapter/node/inbound/node_template_write_service.go`

**Action:** MODIFY

**Rationale:** Service should create repositories per-operation using the factory, binding them to the active UoW.

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
    uowFactory           outbound.UnitOfWorkFactory
    factory              *aggregate.NodeTemplateFactory
    readRepository       nodeOutbound.NodeTemplateReadRepository
    writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory  // CHANGED: factory instead of instance
    mapper               inbound.NodeTemplateMapper
    idFactory            id.Factory
}

func NewNodeTemplateWriteService(
    uowFactory outbound.UnitOfWorkFactory,
    factory *aggregate.NodeTemplateFactory,
    readRepository nodeOutbound.NodeTemplateReadRepository,
    writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory,  // CHANGED
    mapper inbound.NodeTemplateMapper,
    idFactory id.Factory,
) *NodeTemplateWriteService {
    return &NodeTemplateWriteService{
        uowFactory:             uowFactory,
        factory:                factory,
        readRepository:         readRepository,
        writeRepositoryFactory: writeRepositoryFactory,  // CHANGED
        mapper:                 mapper,
        idFactory:              idFactory,
    }
}

func (s *NodeTemplateWriteService) Create(ctx context.Context, input inbound.CreateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
    // 1. Create UoW for this operation
    uow := s.uowFactory.Create()

    // 2. Begin transaction
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

    // 3. Ensure rollback on error
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 4. Create repository bound to THIS UoW
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // 5. Create aggregate (adds domain event internally)
    nodeTemplate := s.factory.Make(input.Name)

    // 6. Save via repository (registers on same UoW)
    if err = writeRepo.Save(txCtx, nodeTemplate); err != nil {
        return nil, fmt.Errorf("failed to save node template: %w", err)
    }

    // 7. Commit - persists data + outbox events atomically
    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Update(ctx context.Context, id string, input inbound.UpdateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
    uow := s.uowFactory.Create()

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // Create repository bound to this UoW
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // Load existing aggregate
    nodeTemplate, err := s.readRepository.FindByID(txCtx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find node template: %w", err)
    }
    if nodeTemplate == nil {
        return nil, fmt.Errorf("node template not found: %s", id)
    }

    // Update aggregate (raises domain event)
    nodeTemplate.UpdateName(s.idFactory, input.Name)

    // Persist update (registers on same UoW)
    if err = writeRepo.Update(txCtx, nodeTemplate); err != nil {
        return nil, fmt.Errorf("failed to update node template: %w", err)
    }

    // Commit - persists data + outbox events
    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Delete(ctx context.Context, id string) error {
    uow := s.uowFactory.Create()

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // Create repository bound to this UoW
    writeRepo := s.writeRepositoryFactory.Create(uow)

    // Delete
    if err = writeRepo.Delete(txCtx, id); err != nil {
        return fmt.Errorf("failed to delete node template: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

**Dependencies:**
- `internal/port/node/outbound` (NodeTemplateWriteRepositoryFactory)

**Tests Required:**
- Test Create uses factory to create repository with correct UoW
- Test Update uses factory to create repository with correct UoW
- Test Delete uses factory to create repository with correct UoW
- Test domain events are persisted to outbox on commit

---

### Step 7: Update DI Container Wiring

**File:** `di/container.go`

**Action:** MODIFY

**Rationale:** Wire the repository factory instead of a shared repository instance, removing the problematic shared UoW.

**Pseudocode:**

```go
func NewContainer(ctx context.Context) (*Container, error) {
    // ... database setup ...

    // Shared dependencies
    idFactory := id.NewULIDFactory()

    // Unit of Work Factory (unchanged)
    uowFactory := adapterOutbound.NewUnitOfWorkPostgresFactory(pool)

    // Mappers (unchanged)
    nodeTemplateInboundMapper := nodeAdapterInbound.NewNodeTemplateMapper()

    // Factory (unchanged)
    nodeTemplateFactory := aggregate.NewNodeTemplateFactory(idFactory)

    // Repositories
    nodeTemplateReadRepository := nodeAdapterOutbound.NewNodeTemplatePostgresReadRepository(pool)

    // CHANGED: Use repository factory instead of shared instance
    // REMOVED: sharedUow := adapterOutbound.NewUnitOfWorkPostgres(pool)
    // REMOVED: nodeTemplateWriteRepository := nodeAdapterOutbound.NewNodeTemplatePostgresWriteRepository(pool, sharedUow)
    nodeTemplateWriteRepositoryFactory := nodeAdapterOutbound.NewNodeTemplatePostgresWriteRepositoryFactory()

    // Services
    nodeTemplateReadService := nodeAdapterInbound.NewNodeTemplateReadService(
        nodeTemplateInboundMapper,
        nodeTemplateReadRepository,
    )

    nodeTemplateWriteService := nodeAdapterInbound.NewNodeTemplateWriteService(
        uowFactory,
        nodeTemplateFactory,
        nodeTemplateReadRepository,
        nodeTemplateWriteRepositoryFactory,  // CHANGED: factory instead of instance
        nodeTemplateInboundMapper,
        idFactory,
    )

    // ... rest unchanged ...
}
```

**Dependencies:**
- `internal/adapter/node/outbound` (repository factory)

**Tests Required:**
- Integration test: Create entity and verify outbox event exists
- Integration test: Update entity and verify outbox event exists

---

## 4. Data Changes

**No schema changes required.** The outbox table schema remains unchanged.

**Migration Notes:** None - this is a code-only fix.

---

## 5. Integration Points

| Service | Interaction | Error Handling |
|---------|-------------|----------------|
| PostgreSQL | Transaction management via UoW | Rollback on any error |
| Outbox table | Events persisted on UoW.Commit | Part of same transaction |
| OutboxProcessor | Reads from outbox table | No changes needed |

---

## 6. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Transaction begin fails | Return error, no cleanup needed |
| Repository save fails | Deferred rollback executes |
| Commit fails | Transaction already rolled back by postgres |
| Multiple aggregates in one transaction | All registered on same UoW, all events persisted |
| Concurrent requests | Each gets own UoW instance, no shared state |
| Querier called before Begin | Returns pool (fallback), but warns in logs |

---

## 7. Testing Strategy

**Unit Tests:**
- UoW.Querier returns tx after Begin
- UoW.Querier returns pool before Begin
- Repository factory creates instance with correct UoW
- Repository.Save calls RegisterNew on injected UoW
- Repository.Update calls RegisterDirty on injected UoW
- Service creates repository via factory for each operation

**Integration Tests:**
- Create aggregate → verify row in node_templates AND outbox
- Update aggregate → verify updated row AND new outbox event
- Delete aggregate → verify deleted row
- Rollback scenario → verify neither data nor outbox event exists

**Manual Verification:**
1. Start the application
2. Create a node template via API
3. Query database: `SELECT * FROM outbox WHERE aggregate_type = 'NodeTemplate'`
4. Verify event exists with correct payload
5. Check OutboxProcessor logs for event processing

---

## 8. Implementation Order

1. **Step 1: Extend UnitOfWork Interface** — Foundation for all other changes
2. **Step 2: Implement Querier in PostgreSQL UoW** — Enables repository changes
3. **Step 3: Create Repository Factory Interface** — Define contract
4. **Step 4: Refactor Write Repository** — Remove pool dependency, use UoW.Querier
5. **Step 5: Create Repository Factory Implementation** — Implement the factory
6. **Step 6: Update Write Service** — Use factory instead of instance
7. **Step 7: Update DI Container** — Wire everything together
8. **Run tests** — Verify fix works end-to-end

---

## 9. Verification Checklist

After implementation, verify:

- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] Create API call results in outbox event
- [ ] Update API call results in outbox event
- [ ] OutboxProcessor picks up and processes events
- [ ] No shared UoW instances exist in codebase
- [ ] Each write operation creates its own repository via factory

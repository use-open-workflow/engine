# Research Report: Fix UoW Logic

## Problem Analysis

The current implementation has **two separate UoW instances** for the same request:

1. **Shared UoW** (`sharedUow`) - Created once in `di/container.go:54` and passed to the write repository
2. **Service-level UoW** - Created via `uowFactory.Create()` in each service method (e.g., `Create`, `Update`, `Delete`)

This causes data inconsistency because:
- The write repository uses `sharedUow` to register aggregates and check for transactions
- The service creates a *new* UoW instance from the factory and manages the transaction lifecycle
- When commit happens, only the service-level UoW's transaction commits; the `sharedUow` holds the registered aggregates but doesn't commit its events

---

## 1. Relevant Files

**Entry Points:**
- `api/router.go` - Route setup, no UoW involvement
- `api/node/http/node_template_handler.go` - HTTP handlers, calls services

**Services/Logic:**
- `internal/adapter/node/inbound/node_template_write_service.go` - Creates UoW per operation, manages transaction lifecycle
- `internal/port/node/inbound/node_template_write_service.go` - Service interface

**Data Layer:**
- `internal/adapter/node/outbound/node_template_postgres_write_repository.go` - Uses `sharedUow` to get tx and register aggregates
- `internal/adapter/node/outbound/node_template_postgres_read_repository.go` - Uses pool directly (no UoW)
- `internal/port/node/outbound/node_template_write_repository.go` - Repository interface

**UoW Implementation:**
- `internal/port/outbound/unit_of_work.go` - UoW interface and factory interface
- `internal/adapter/outbound/unit_of_work_postgres.go` - UoW postgres implementation
- `internal/adapter/outbound/unit_of_work_postgres_factory.go` - Factory implementation

**DI Container:**
- `di/container.go` - Creates both `sharedUow` and `uowFactory`, wires dependencies

**Domain:**
- `pkg/domain/interface.go` - Aggregate interface (Events, ClearEvents)
- `pkg/domain/base_aggregate.go` - BaseAggregate with event handling

**Tests:**
- No test files exist currently

---

## 2. Dependencies & Integrations

**Internal Modules:**
- `pkg/domain` - Base aggregate with event tracking
- `pkg/id` - ULID factory for ID generation
- `internal/domain/node/aggregate` - NodeTemplate aggregate
- `internal/domain/node/event` - Domain events

**External Libraries:**
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/gofiber/fiber/v3` - HTTP framework

**Shared Utilities:**
- Transaction context key (`txKey`) for passing tx through context
- `querier` interface for abstracting pool/tx

---

## 3. Data Flow

Current (broken) flow:
```
Handler -> Service.Create() -> uowFactory.Create() [new UoW #2]
                            -> uow.Begin() [starts tx in UoW #2]
                            -> writeRepo.Save() -> sharedUow.RegisterNew() [registers in UoW #1]
                            -> uow.Commit() [commits UoW #2, but UoW #1 has the aggregates]
```

The write repository calls `r.uow.GetTx(ctx)` which looks for a transaction on the `sharedUow`, but the service created a different UoW instance. The transaction IS found (via context), but `persistOutboxEvents` uses the aggregates registered on `sharedUow`, not the service's UoW.

---

## 4. Impact Areas

**Direct Modifications Required:**
- `di/container.go` - Remove `sharedUow`, change how repositories are created
- `internal/adapter/node/outbound/node_template_postgres_write_repository.go` - Accept UoW per call or via factory
- `internal/port/node/outbound/node_template_write_repository.go` - May need interface change
- `internal/adapter/node/inbound/node_template_write_service.go` - May need to pass UoW to repository

**Indirect Impacts:**
- Outbox events persistence (currently in UoW commit)
- Any future repositories will need same pattern

---

## 5. Implementation Constraints

**Patterns to Follow:**
- Hexagonal architecture: ports define interfaces, adapters implement
- Repositories should not depend on concrete UoW implementation
- Transaction management at service layer

**Proposed Solution (from FEATURE.md):**
- Drop `sharedUow`
- Use repository factory that accepts UoW instance
- Service creates UoW, passes to repository factory to get repository bound to that UoW

**Business Rules:**
- All changes in a request must be atomic
- Domain events must be persisted in same transaction as aggregate changes

**Testing Requirements:**
- No existing tests; new implementation should be testable with mock UoW

---

## 6. Reference Implementations

**Current Pattern (to fix):**
- `internal/adapter/node/inbound/node_template_write_service.go:41-66` - Shows service creating UoW
- `internal/adapter/node/outbound/node_template_postgres_write_repository.go:29-34` - Shows how repository gets tx from UoW

**Recommended Pattern:**
The factory pattern is already partially implemented:
- `internal/port/outbound/unit_of_work.go:14-16` - `UnitOfWorkFactory` interface exists
- `internal/adapter/outbound/unit_of_work_postgres_factory.go` - Factory implementation exists

Extend this pattern to create a `WriteRepositoryFactory` that takes UoW and returns a repository bound to that UoW.

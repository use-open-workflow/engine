# Research Report: Add createdAt and updatedAt to Base Aggregate

## 1. Relevant Files

### Domain Layer (Core changes)
- [pkg/domain/base_aggregate.go](pkg/domain/base_aggregate.go) — Add `CreatedAt` and `UpdatedAt` fields to `BaseAggregate` struct; update `NewBaseAggregate` function
- [internal/domain/node/aggregate/node_template.go](internal/domain/node/aggregate/node_template.go) — Update `newNodeTemplate` and `ReconstituteNodeTemplate` to pass timestamps

### Port Layer (Interface contracts)
- [internal/port/node/inbound/node_template_dto.go](internal/port/node/inbound/node_template_dto.go) — Add `CreatedAt` and `UpdatedAt` fields to DTO with JSON tags
- [internal/port/node/outbound/node_template_model.go](internal/port/node/outbound/node_template_model.go) — Already has `CreatedAt` and `UpdatedAt` fields (no change needed)

### Adapter Layer (Mappers & Repositories)
- [internal/adapter/node/inbound/node_template_mapper.go](internal/adapter/node/inbound/node_template_mapper.go) — Map timestamps from aggregate to DTO
- [internal/adapter/node/outbound/node_template_mapper.go](internal/adapter/node/outbound/node_template_mapper.go) — Map timestamps between model and aggregate
- [internal/adapter/node/outbound/node_template_postgres_write_repository.go](internal/adapter/node/outbound/node_template_postgres_write_repository.go) — Replace `NOW()` with application-provided timestamps in `Save` and `Update` methods
- [internal/adapter/node/outbound/node_template_postgres_read_repository.go](internal/adapter/node/outbound/node_template_postgres_read_repository.go) — Pass `createdAt` and `updatedAt` to `ReconstituteNodeTemplate`

### Tests
- Test files in `internal/adapter/node/` (if any exist) — Update tests to include timestamp assertions

## 2. Dependencies & Integrations

### Internal Modules
- `pkg/domain` — Base types shared across all aggregates
- `pkg/id` — ULID factory (not needed for timestamps)
- `time` package — Standard library for `time.Time` type

### External Services/APIs
- None — timestamps are application-controlled

### Shared Utilities
- None specific to timestamps; use standard `time.Now().UTC()`

## 3. Data Flow

1. **Create flow**: Service calls factory → `NewBaseAggregate` sets `CreatedAt = time.Now().UTC()` and `UpdatedAt = time.Now().UTC()` → Repository `Save` uses aggregate's timestamp values in INSERT query → API returns DTO with timestamps

2. **Update flow**: Service loads aggregate via `ReconstituteNodeTemplate` (timestamps from DB) → Business logic updates fields → Service sets `UpdatedAt = time.Now().UTC()` before save → Repository `Update` uses aggregate's `UpdatedAt` in UPDATE query

3. **Read flow**: Repository queries DB → `ReconstituteNodeTemplate` receives timestamps → Mapper converts aggregate to DTO with timestamps → API returns DTO

## 4. Impact Areas

### Direct Modifications
- `BaseAggregate` struct (affects all future aggregates)
- `NodeTemplate` construction functions
- All mappers (inbound and outbound)
- Both repository methods (`Save`, `Update`)
- DTO structure (API response contract change)

### Indirect Impacts
- API consumers will receive new fields in responses (backward-compatible addition)
- Any code that constructs `BaseAggregate` directly needs updating

## 5. Implementation Constraints

### Coding Patterns
- Use `time.Time` type (consistent with existing `NodeTemplateModel`)
- Store all timestamps in UTC: `time.Now().UTC()`
- Use JSON tag format: `json:"createdAt"` (camelCase for API)
- Follow existing reconstitute pattern for hydrating from DB

### Validation/Business Rules
- `CreatedAt` is immutable after initial creation
- `UpdatedAt` must be set on every modification
- Both fields must be non-zero (never `time.Time{}`)

### Auth/Permission Requirements
- None — internal domain change

### Performance Considerations
- Minimal impact; `time.Now()` is inexpensive
- No additional DB queries required

### Testing Requirements
- Unit tests for `NewBaseAggregate` to verify timestamp initialization
- Integration tests to verify timestamps flow through to API responses
- Verify `UpdatedAt` changes on updates while `CreatedAt` remains unchanged

## 6. Reference Implementations

1. **NodeTemplateModel** ([internal/port/node/outbound/node_template_model.go:7-8](internal/port/node/outbound/node_template_model.go#L7-L8)) — Shows existing pattern for timestamp fields with `time.Time` type

2. **Outbox table schema** ([migration/001_initial_schema.sql:18](migration/001_initial_schema.sql#L18)) — Shows existing pattern for `created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()` but we'll replace SQL default with application-provided values

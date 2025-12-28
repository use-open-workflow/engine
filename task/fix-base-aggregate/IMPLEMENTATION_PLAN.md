# Implementation Plan: Add createdAt and updatedAt to Base Aggregate

## 1. Implementation Summary

This implementation adds `CreatedAt` and `UpdatedAt` timestamp fields to the `BaseAggregate` struct, making them available across all aggregates. The timestamps are managed by the application (not SQL functions), with `CreatedAt` set once during creation and `UpdatedAt` updated on every modification. The changes flow through the domain layer to the DTO layer, exposing timestamps in API responses.

## 2. Change Manifest

```
MODIFY:
- pkg/domain/base_aggregate.go — Add CreatedAt and UpdatedAt fields; update NewBaseAggregate and add ReconstituteBaseAggregate function
- internal/domain/node/aggregate/node_template.go — Update newNodeTemplate to use new BaseAggregate; update ReconstituteNodeTemplate to accept timestamps; add SetUpdatedAt method for update operations
- internal/port/node/inbound/node_template_dto.go — Add CreatedAt and UpdatedAt fields with JSON tags
- internal/adapter/node/inbound/node_template_mapper.go — Map timestamps from aggregate to DTO
- internal/adapter/node/outbound/node_template_mapper.go — Map timestamps between model and aggregate in both From and To methods
- internal/adapter/node/outbound/node_template_postgres_write_repository.go — Replace NOW() with application-provided timestamps in Save and Update methods
- internal/adapter/node/outbound/node_template_postgres_read_repository.go — Pass timestamps to ReconstituteNodeTemplate in FindMany and FindByID
```

## 3. Step-by-Step Plan

---

### Step 1: Update BaseAggregate Struct

**File:** `pkg/domain/base_aggregate.go`

**Action:** MODIFY

**Rationale:** The base aggregate needs to carry timestamp fields that all aggregates will inherit.

**Pseudocode:**

```go
import "time"

type BaseAggregate struct {
    ID        string
    CreatedAt time.Time
    UpdatedAt time.Time
    events    []Event
}

// NewBaseAggregate creates a new aggregate with current UTC timestamps
// Used when creating NEW domain entities
func NewBaseAggregate(id string) BaseAggregate {
    now := time.Now().UTC()
    return BaseAggregate{
        ID:        id,
        CreatedAt: now,
        UpdatedAt: now,
        events:    make([]Event, 0),
    }
}

// ReconstituteBaseAggregate recreates an aggregate from persisted data
// Used when loading existing entities from the database
// Does NOT initialize events (they are transient)
func ReconstituteBaseAggregate(id string, createdAt time.Time, updatedAt time.Time) BaseAggregate {
    return BaseAggregate{
        ID:        id,
        CreatedAt: createdAt,
        UpdatedAt: updatedAt,
        events:    make([]Event, 0),
    }
}

// SetUpdatedAt updates the UpdatedAt timestamp
// Should be called before any persistence operation that modifies the aggregate
func (a *BaseAggregate) SetUpdatedAt(t time.Time) {
    a.UpdatedAt = t
}
```

**Dependencies:**
- `time` package from standard library

**Tests Required:**
- Test `NewBaseAggregate` sets both timestamps to current UTC time
- Test `NewBaseAggregate` sets CreatedAt and UpdatedAt to the same value
- Test `ReconstituteBaseAggregate` preserves passed timestamps
- Test `SetUpdatedAt` updates only UpdatedAt field

---

### Step 2: Update NodeTemplate Aggregate

**File:** `internal/domain/node/aggregate/node_template.go`

**Action:** MODIFY

**Rationale:** The NodeTemplate must support reconstitution with timestamps and update the UpdatedAt field on modifications.

**Pseudocode:**

```go
import "time"

// newNodeTemplate creates a new NodeTemplate with auto-generated timestamps
// CreatedAt and UpdatedAt are set automatically by BaseAggregate
func newNodeTemplate(idFactory id.Factory, aggregateID string, name string) *NodeTemplate {
    nodeTemplate := &NodeTemplate{
        BaseAggregate: domain.NewBaseAggregate(aggregateID),  // timestamps auto-set
        Name:          name,
    }
    nodeTemplate.AddEvent(event.NewCreateNodeTemplate(idFactory, nodeTemplate.ID, name))
    return nodeTemplate
}

// ReconstituteNodeTemplate recreates a NodeTemplate from persisted data
// MUST receive timestamps from database
func ReconstituteNodeTemplate(aggregateID string, name string, createdAt time.Time, updatedAt time.Time) *NodeTemplate {
    return &NodeTemplate{
        BaseAggregate: domain.ReconstituteBaseAggregate(aggregateID, createdAt, updatedAt),
        Name:          name,
    }
}

// UpdateName updates the name field
// IMPORTANT: Sets UpdatedAt to current UTC time before persisting
func (n *NodeTemplate) UpdateName(idFactory id.Factory, name string) {
    n.Name = name
    n.SetUpdatedAt(time.Now().UTC())  // Update timestamp on modification
    n.AddEvent(event.NewUpdateNodeTemplate(idFactory, n.ID, name))
}
```

**Dependencies:**
- `time` package from standard library
- `pkg/domain` (already imported)

**Tests Required:**
- Test `newNodeTemplate` creates aggregate with current UTC timestamps
- Test `ReconstituteNodeTemplate` preserves passed timestamps
- Test `UpdateName` updates the UpdatedAt timestamp to current UTC time
- Test `UpdateName` does not modify CreatedAt

---

### Step 3: Update NodeTemplateDTO

**File:** `internal/port/node/inbound/node_template_dto.go`

**Action:** MODIFY

**Rationale:** The API response DTO needs to include timestamp fields for clients.

**Pseudocode:**

```go
import "time"

type NodeTemplateDTO struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**Dependencies:**
- `time` package from standard library

**Tests Required:**
- Test JSON serialization produces correct field names (camelCase)
- Test time format in JSON output is ISO 8601 / RFC 3339

---

### Step 4: Update Inbound Mapper (Aggregate → DTO)

**File:** `internal/adapter/node/inbound/node_template_mapper.go`

**Action:** MODIFY

**Rationale:** The mapper must copy timestamps from the aggregate to the DTO.

**Pseudocode:**

```go
func (m *NodeTemplateMapper) To(nodeTemplate *aggregate.NodeTemplate) (*inbound.NodeTemplateDTO, error) {
    return &inbound.NodeTemplateDTO{
        ID:        nodeTemplate.ID,
        Name:      nodeTemplate.Name,
        CreatedAt: nodeTemplate.CreatedAt,  // From BaseAggregate
        UpdatedAt: nodeTemplate.UpdatedAt,  // From BaseAggregate
    }, nil
}
```

**Dependencies:**
- No new imports required (aggregate already imported)

**Tests Required:**
- Test timestamps are correctly mapped from aggregate to DTO
- Test all fields are present in output DTO

---

### Step 5: Update Outbound Mapper (Model ↔ Aggregate)

**File:** `internal/adapter/node/outbound/node_template_mapper.go`

**Action:** MODIFY

**Rationale:** The mapper must transfer timestamps between the persistence model and the aggregate.

**Pseudocode:**

```go
// From converts a persistence model to a domain aggregate
// Timestamps from DB are passed to ReconstituteNodeTemplate
func (*NodeTemplateMapper) From(in *outbound.NodeTemplateModel) (*aggregate.NodeTemplate, error) {
    return aggregate.ReconstituteNodeTemplate(
        in.ID,
        in.Name,
        in.CreatedAt,  // From NodeTemplateModel
        in.UpdatedAt,  // From NodeTemplateModel
    ), nil
}

// To converts a domain aggregate to a persistence model
// Timestamps from aggregate are copied to model
func (*NodeTemplateMapper) To(in *aggregate.NodeTemplate) (*outbound.NodeTemplateModel, error) {
    return &outbound.NodeTemplateModel{
        ID:        in.ID,
        Name:      in.Name,
        CreatedAt: in.CreatedAt,  // From BaseAggregate
        UpdatedAt: in.UpdatedAt,  // From BaseAggregate
    }, nil
}
```

**Dependencies:**
- No new imports required

**Tests Required:**
- Test `From` correctly maps timestamps from model to aggregate
- Test `To` correctly maps timestamps from aggregate to model
- Test roundtrip: Model → Aggregate → Model preserves timestamps

---

### Step 6: Update Write Repository - Save Method

**File:** `internal/adapter/node/outbound/node_template_postgres_write_repository.go`

**Action:** MODIFY

**Rationale:** Replace SQL NOW() with application-provided timestamps for INSERT operations.

**Pseudocode:**

```go
func (r *NodeTemplatePostgresWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        INSERT INTO node_templates (id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
    `, nodeTemplate.ID, nodeTemplate.Name, nodeTemplate.CreatedAt, nodeTemplate.UpdatedAt)
    //                              ^^^^^^ Use aggregate's CreatedAt
    //                                                      ^^^^^^ Use aggregate's UpdatedAt

    if err != nil {
        return fmt.Errorf("failed to save node template: %w", err)
    }

    r.uow.RegisterNew(nodeTemplate)

    return nil
}
```

**Dependencies:**
- No new imports required

**Tests Required:**
- Test INSERT uses aggregate's timestamps, not NOW()
- Test saved record has correct CreatedAt and UpdatedAt values
- Integration test: create → read roundtrip preserves timestamps

---

### Step 7: Update Write Repository - Update Method

**File:** `internal/adapter/node/outbound/node_template_postgres_write_repository.go`

**Action:** MODIFY

**Rationale:** Replace SQL NOW() with application-provided UpdatedAt for UPDATE operations.

**Pseudocode:**

```go
func (r *NodeTemplatePostgresWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        UPDATE node_templates
        SET name = $1, updated_at = $2
        WHERE id = $3
    `, nodeTemplate.Name, nodeTemplate.UpdatedAt, nodeTemplate.ID)
    //                    ^^^^^^ Use aggregate's UpdatedAt

    if err != nil {
        return fmt.Errorf("failed to update node template: %w", err)
    }

    r.uow.RegisterDirty(nodeTemplate)

    return nil
}
```

**Dependencies:**
- No new imports required

**Tests Required:**
- Test UPDATE uses aggregate's UpdatedAt, not NOW()
- Test CreatedAt is not modified by UPDATE
- Integration test: update → read roundtrip preserves CreatedAt, updates UpdatedAt

---

### Step 8: Update Read Repository - FindByID Method

**File:** `internal/adapter/node/outbound/node_template_postgres_read_repository.go`

**Action:** MODIFY

**Rationale:** Pass timestamps from database to ReconstituteNodeTemplate.

**Pseudocode:**

```go
import "time"

func (r *NodeTemplatePostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.NodeTemplate, error) {
    q := r.uow.Querier(ctx)

    var name string
    var createdAt, updatedAt time.Time  // Change from interface{} to time.Time
    err := q.QueryRow(ctx, `
        SELECT id, name, created_at, updated_at
        FROM node_templates
        WHERE id = $1
    `, id).Scan(&id, &name, &createdAt, &updatedAt)

    if err != nil && err.Error() == "no rows in result set" {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query node template: %w", err)
    }

    // Pass timestamps to reconstitute function
    return aggregate.ReconstituteNodeTemplate(id, name, createdAt, updatedAt), nil
}
```

**Dependencies:**
- `time` package from standard library

**Tests Required:**
- Test timestamps are correctly scanned from database
- Test timestamps are passed to ReconstituteNodeTemplate
- Test returned aggregate has correct CreatedAt and UpdatedAt values

---

### Step 9: Update Read Repository - FindMany Method

**File:** `internal/adapter/node/outbound/node_template_postgres_read_repository.go`

**Action:** MODIFY

**Rationale:** Pass timestamps from database to ReconstituteNodeTemplate for each row.

**Pseudocode:**

```go
func (r *NodeTemplatePostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.NodeTemplate, error) {
    q := r.uow.Querier(ctx)

    rows, err := q.Query(ctx, `
        SELECT id, name, created_at, updated_at
        FROM node_templates
        ORDER BY created_at DESC
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to query node templates: %w", err)
    }
    defer rows.Close()

    var templates []*aggregate.NodeTemplate
    for rows.Next() {
        var id, name string
        var createdAt, updatedAt time.Time  // Change from interface{} to time.Time
        if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan node template: %w", err)
        }

        // Pass timestamps to reconstitute function
        template := aggregate.ReconstituteNodeTemplate(id, name, createdAt, updatedAt)
        templates = append(templates, template)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }

    return templates, nil
}
```

**Dependencies:**
- `time` package (may already be imported or needs to be added)

**Tests Required:**
- Test timestamps are correctly scanned from database for multiple rows
- Test each aggregate has correct CreatedAt and UpdatedAt values
- Test ordering by created_at DESC is preserved

---

## 4. Data Changes

**Schema/Model Updates:**

No database schema changes are required. The `node_templates` table already has `created_at` and `updated_at` columns with `TIMESTAMP WITH TIME ZONE` type.

**Migration Notes:**

- No migration required
- Existing data will continue to work since timestamps are already stored in the database
- The only change is moving timestamp control from SQL `NOW()` to application code

---

## 5. Integration Points

| Service | Interaction | Error Handling |
|---------|-------------|----------------|
| PostgreSQL | Timestamps stored as `TIMESTAMP WITH TIME ZONE` | Existing error handling is sufficient |
| JSON Serialization | `time.Time` serializes as RFC 3339 by default | No special handling needed |

---

## 6. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Zero-value timestamps on aggregate | Should never occur if proper factory functions are used; ReconstituteBaseAggregate requires timestamps |
| Timezone consistency | All timestamps use `time.Now().UTC()` to ensure UTC storage |
| Null timestamps from legacy data | Not applicable - schema has `NOT NULL` constraints |
| Time precision differences between Go and PostgreSQL | PostgreSQL `TIMESTAMP WITH TIME ZONE` has microsecond precision; Go `time.Time` has nanosecond precision. Sub-microsecond data may be truncated on roundtrip. This is acceptable. |

---

## 7. Testing Strategy

**Unit Tests:**

1. `pkg/domain/base_aggregate_test.go`:
   - `TestNewBaseAggregate_SetsTimestamps` - verify CreatedAt and UpdatedAt are set to current UTC
   - `TestNewBaseAggregate_TimestampsAreEqual` - verify both timestamps are identical on creation
   - `TestReconstituteBaseAggregate_PreservesTimestamps` - verify passed timestamps are preserved
   - `TestSetUpdatedAt_UpdatesOnlyUpdatedAt` - verify SetUpdatedAt only modifies UpdatedAt

2. `internal/domain/node/aggregate/node_template_test.go`:
   - `TestNewNodeTemplate_HasTimestamps` - verify new template has timestamps
   - `TestReconstituteNodeTemplate_PreservesTimestamps` - verify reconstituted template has passed timestamps
   - `TestUpdateName_UpdatesTimestamp` - verify UpdateName sets UpdatedAt to current time
   - `TestUpdateName_PreservesCreatedAt` - verify UpdateName does not modify CreatedAt

**Integration Tests:**

1. `internal/adapter/node/outbound/node_template_repository_test.go`:
   - `TestSave_PersistsTimestamps` - verify INSERT uses aggregate's timestamps
   - `TestUpdate_UpdatesOnlyUpdatedAt` - verify UPDATE modifies only UpdatedAt
   - `TestFindByID_ReturnsTimestamps` - verify read returns correct timestamps
   - `TestFindMany_ReturnsTimestamps` - verify list returns correct timestamps for all items
   - `TestRoundtrip_PreservesTimestamps` - verify create → read preserves timestamps

**Manual Verification:**

1. Start the application with `make run`
2. Create a new NodeTemplate via POST API
3. Verify response includes `createdAt` and `updatedAt` fields in ISO 8601 format
4. Verify `createdAt` and `updatedAt` have the same value
5. Update the NodeTemplate via PUT/PATCH API
6. Verify `updatedAt` is updated to a new value
7. Verify `createdAt` remains unchanged
8. Fetch the NodeTemplate via GET API
9. Verify timestamps are returned correctly

---

## 8. Implementation Order

Recommended sequence for implementation:

1. **Step 1: Update BaseAggregate** — Foundation that all other changes depend on
2. **Step 2: Update NodeTemplate aggregate** — Depends on BaseAggregate changes
3. **Step 3: Update NodeTemplateDTO** — Independent, but needed before mapper changes
4. **Step 5: Update Outbound Mapper** — Depends on NodeTemplate having timestamps
5. **Step 4: Update Inbound Mapper** — Depends on DTO having timestamp fields
6. **Step 8: Update FindByID** — Depends on ReconstituteNodeTemplate signature change
7. **Step 9: Update FindMany** — Depends on ReconstituteNodeTemplate signature change
8. **Step 6: Update Save method** — Depends on aggregate having timestamps
9. **Step 7: Update Update method** — Depends on aggregate having timestamps

**Rationale for ordering:**
- Domain layer changes (Steps 1-2) must come first as they define the data structures
- Port layer DTO (Step 3) can be done in parallel with domain changes
- Outbound mapper (Step 5) before inbound (Step 4) because read repository needs it first
- Read repositories (Steps 8-9) before write repositories (Steps 6-7) to ensure reconstitution works before we change how data is written
- Write repositories last because they depend on aggregates already having proper timestamp values

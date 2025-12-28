# Implementation Plan: Workflow Domain

## 1. Implementation Summary

This plan implements the `workflow` domain following the established hexagonal architecture with DDD patterns. The Workflow aggregate contains two child entities (NodeDefinition and Edge) that are managed together as a single transactional unit. The implementation follows the existing NodeTemplate domain patterns for services, repositories, and handlers, with additional complexity for parent-child entity relationships in database operations.

---

## 2. Change Manifest

```
CREATE:
- internal/domain/workflow/aggregate/workflow.go — Workflow aggregate root with child entities
- internal/domain/workflow/aggregate/workflow_factory.go — Factory for creating Workflow aggregates
- internal/domain/workflow/aggregate/node_definition.go — NodeDefinition entity
- internal/domain/workflow/aggregate/edge.go — Edge entity
- internal/domain/workflow/event/create_workflow.go — Domain event for workflow creation
- internal/domain/workflow/event/update_workflow.go — Domain event for workflow updates

- internal/port/workflow/inbound/workflow_dto.go — DTOs for Workflow, NodeDefinition, Edge
- internal/port/workflow/inbound/workflow_read_service.go — Read service interface
- internal/port/workflow/inbound/workflow_write_service.go — Write service interface with input structs
- internal/port/workflow/inbound/workflow_mapper.go — Mapper interface (aggregate → DTO)

- internal/port/workflow/outbound/workflow_model.go — Database models
- internal/port/workflow/outbound/workflow_read_repository.go — Read repository interface
- internal/port/workflow/outbound/workflow_write_repository.go — Write repository interface
- internal/port/workflow/outbound/workflow_read_repository_factory.go — Factory interface
- internal/port/workflow/outbound/workflow_write_repository_factory.go — Factory interface

- internal/adapter/workflow/inbound/workflow_read_service.go — Read service implementation
- internal/adapter/workflow/inbound/workflow_write_service.go — Write service implementation
- internal/adapter/workflow/inbound/workflow_mapper.go — Mapper implementation

- internal/adapter/workflow/outbound/workflow_postgres_read_repository.go — PostgreSQL read impl
- internal/adapter/workflow/outbound/workflow_postgres_write_repository.go — PostgreSQL write impl
- internal/adapter/workflow/outbound/workflow_postgres_read_repository_factory.go — Factory impl
- internal/adapter/workflow/outbound/workflow_postgres_write_repository_factory.go — Factory impl

- api/workflow/http/workflow_handler.go — HTTP handlers for workflow operations

- migration/002_workflow_schema.sql — Database schema for workflow tables

MODIFY:
- api/router.go — Register workflow routes
- di/container.go — Wire up workflow services and dependencies
```

---

## 3. Step-by-Step Plan

### Step 1: Create Database Migration

**File:** `migration/002_workflow_schema.sql`

**Action:** CREATE

**Rationale:** Database tables must exist before any domain code can persist data.

**Pseudocode:**

```sql
-- Workflow table (aggregate root)
CREATE TABLE workflow (
    id VARCHAR(26) PRIMARY KEY,           -- ULID
    name VARCHAR(255) NOT NULL,           -- Workflow name
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for listing workflows by creation date
CREATE INDEX idx_workflow_created_at ON workflow(created_at DESC);

-- NodeDefinition table (child entity of Workflow)
CREATE TABLE node_definition (
    id VARCHAR(26) PRIMARY KEY,           -- ULID
    workflow_id VARCHAR(26) NOT NULL,     -- Parent workflow reference
    node_template_id VARCHAR(26) NOT NULL,-- Reference to node_template
    name VARCHAR(255) NOT NULL,           -- Display name for this node instance
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,  -- Canvas X position
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,  -- Canvas Y position

    CONSTRAINT fk_node_definition_workflow
        FOREIGN KEY (workflow_id) REFERENCES workflow(id) ON DELETE CASCADE,
    CONSTRAINT fk_node_definition_node_template
        FOREIGN KEY (node_template_id) REFERENCES node_template(id)
);

-- Index for finding node definitions by workflow
CREATE INDEX idx_node_definition_workflow_id ON node_definition(workflow_id);

-- Edge table (child entity of Workflow, connects NodeDefinitions)
CREATE TABLE edge (
    id VARCHAR(26) PRIMARY KEY,           -- ULID
    workflow_id VARCHAR(26) NOT NULL,     -- Parent workflow reference
    from_node_id VARCHAR(26) NOT NULL,    -- Source NodeDefinition
    to_node_id VARCHAR(26) NOT NULL,      -- Target NodeDefinition

    CONSTRAINT fk_edge_workflow
        FOREIGN KEY (workflow_id) REFERENCES workflow(id) ON DELETE CASCADE,
    CONSTRAINT fk_edge_from_node
        FOREIGN KEY (from_node_id) REFERENCES node_definition(id) ON DELETE CASCADE,
    CONSTRAINT fk_edge_to_node
        FOREIGN KEY (to_node_id) REFERENCES node_definition(id) ON DELETE CASCADE,

    -- Prevent duplicate edges between same nodes
    CONSTRAINT uq_edge_from_to UNIQUE (workflow_id, from_node_id, to_node_id)
);

-- Index for finding edges by workflow
CREATE INDEX idx_edge_workflow_id ON edge(workflow_id);
```

**Dependencies:** PostgreSQL database, migration/001_initial_schema.sql must be applied first

**Tests Required:**
- Verify tables are created successfully
- Verify foreign key constraints work (cascade delete)
- Verify unique constraint on edges prevents duplicates

---

### Step 2: Create NodeDefinition Entity

**File:** `internal/domain/workflow/aggregate/node_definition.go`

**Action:** CREATE

**Rationale:** NodeDefinition is a child entity of Workflow representing an instance of a node template.

**Pseudocode:**

```go
package aggregate

import "use-open-workflow.io/engine/pkg/domain"

// NodeDefinition represents an instance of a NodeTemplate within a Workflow.
// It is a child entity owned by the Workflow aggregate.
type NodeDefinition struct {
    domain.BaseEntity
    WorkflowID     string   // Parent workflow ID
    NodeTemplateID string   // Reference to the template this node is based on
    Name           string   // Display name for this node instance
    PositionX      float64  // X position on the workflow canvas
    PositionY      float64  // Y position on the workflow canvas
}

// newNodeDefinition creates a new NodeDefinition entity.
// Called internally by Workflow.AddNodeDefinition()
func newNodeDefinition(id, workflowID, nodeTemplateID, name string, posX, posY float64) *NodeDefinition {
    return &NodeDefinition{
        BaseEntity:     domain.NewBaseEntity(id),
        WorkflowID:     workflowID,
        NodeTemplateID: nodeTemplateID,
        Name:           name,
        PositionX:      posX,
        PositionY:      posY,
    }
}

// ReconstituteNodeDefinition recreates a NodeDefinition from persistence.
// Used by repository when loading from database.
func ReconstituteNodeDefinition(id, workflowID, nodeTemplateID, name string, posX, posY float64) *NodeDefinition {
    return &NodeDefinition{
        BaseEntity:     domain.NewBaseEntity(id),
        WorkflowID:     workflowID,
        NodeTemplateID: nodeTemplateID,
        Name:           name,
        PositionX:      posX,
        PositionY:      posY,
    }
}

// UpdatePosition updates the node's canvas position
func (n *NodeDefinition) UpdatePosition(x, y float64) {
    n.PositionX = x
    n.PositionY = y
}

// UpdateName updates the node's display name
func (n *NodeDefinition) UpdateName(name string) {
    n.Name = name
}
```

**Dependencies:** `pkg/domain` (BaseEntity)

**Tests Required:**
- Test newNodeDefinition creates entity with correct fields
- Test ReconstituteNodeDefinition recreates entity from persistence data
- Test UpdatePosition modifies coordinates
- Test UpdateName modifies name

---

### Step 3: Create Edge Entity

**File:** `internal/domain/workflow/aggregate/edge.go`

**Action:** CREATE

**Rationale:** Edge is a child entity representing connections between NodeDefinitions within a Workflow.

**Pseudocode:**

```go
package aggregate

import "use-open-workflow.io/engine/pkg/domain"

// Edge represents a directed connection between two NodeDefinitions.
// It is a child entity owned by the Workflow aggregate.
type Edge struct {
    domain.BaseEntity
    WorkflowID string  // Parent workflow ID
    FromNodeID string  // Source NodeDefinition ID
    ToNodeID   string  // Target NodeDefinition ID
}

// newEdge creates a new Edge entity.
// Called internally by Workflow.AddEdge()
func newEdge(id, workflowID, fromNodeID, toNodeID string) *Edge {
    return &Edge{
        BaseEntity: domain.NewBaseEntity(id),
        WorkflowID: workflowID,
        FromNodeID: fromNodeID,
        ToNodeID:   toNodeID,
    }
}

// ReconstituteEdge recreates an Edge from persistence.
// Used by repository when loading from database.
func ReconstituteEdge(id, workflowID, fromNodeID, toNodeID string) *Edge {
    return &Edge{
        BaseEntity: domain.NewBaseEntity(id),
        WorkflowID: workflowID,
        FromNodeID: fromNodeID,
        ToNodeID:   toNodeID,
    }
}
```

**Dependencies:** `pkg/domain` (BaseEntity)

**Tests Required:**
- Test newEdge creates entity with correct fields
- Test ReconstituteEdge recreates entity from persistence data

---

### Step 4: Create Workflow Aggregate

**File:** `internal/domain/workflow/aggregate/workflow.go`

**Action:** CREATE

**Rationale:** Workflow is the aggregate root that owns NodeDefinitions and Edges.

**Pseudocode:**

```go
package aggregate

import (
    "errors"
    "time"

    "use-open-workflow.io/engine/internal/domain/workflow/event"
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

// Domain errors
var (
    ErrNodeDefinitionNotFound = errors.New("node definition not found in workflow")
    ErrEdgeAlreadyExists      = errors.New("edge already exists between these nodes")
    ErrSelfLoopNotAllowed     = errors.New("edge cannot connect a node to itself")
)

// Workflow is the aggregate root for the workflow domain.
// It contains NodeDefinitions and Edges as child entities.
type Workflow struct {
    domain.BaseAggregate
    Name            string
    NodeDefinitions []*NodeDefinition
    Edges           []*Edge
}

// newWorkflow creates a new Workflow aggregate.
// Private - use WorkflowFactory.Make() instead.
func newWorkflow(idFactory id.Factory, aggregateID, name string) *Workflow {
    w := &Workflow{
        BaseAggregate:   domain.NewBaseAggregate(aggregateID),
        Name:            name,
        NodeDefinitions: make([]*NodeDefinition, 0),
        Edges:           make([]*Edge, 0),
    }
    w.AddEvent(event.NewCreateWorkflow(idFactory, w.ID, name))
    return w
}

// ReconstituteWorkflow recreates a Workflow from persistence.
// NodeDefinitions and Edges should be set separately after reconstitution.
func ReconstituteWorkflow(
    aggregateID, name string,
    createdAt, updatedAt time.Time,
    nodeDefinitions []*NodeDefinition,
    edges []*Edge,
) *Workflow {
    return &Workflow{
        BaseAggregate:   domain.ReconstituteBaseAggregate(aggregateID, createdAt, updatedAt),
        Name:            name,
        NodeDefinitions: nodeDefinitions,
        Edges:           edges,
    }
}

// UpdateName updates the workflow name and emits an event.
func (w *Workflow) UpdateName(idFactory id.Factory, name string) {
    w.Name = name
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewUpdateWorkflow(idFactory, w.ID, name))
}

// AddNodeDefinition adds a new node definition to the workflow.
func (w *Workflow) AddNodeDefinition(idFactory id.Factory, nodeTemplateID, name string, posX, posY float64) *NodeDefinition {
    nodeID := idFactory.New()
    node := newNodeDefinition(nodeID, w.ID, nodeTemplateID, name, posX, posY)
    w.NodeDefinitions = append(w.NodeDefinitions, node)
    w.SetUpdatedAt(time.Now().UTC())
    return node
}

// RemoveNodeDefinition removes a node definition and all its connected edges.
func (w *Workflow) RemoveNodeDefinition(nodeID string) error {
    // 1. Find the node
    found := false
    newNodes := make([]*NodeDefinition, 0, len(w.NodeDefinitions))
    for _, node := range w.NodeDefinitions {
        if node.ID == nodeID {
            found = true
            continue // Skip this node (remove it)
        }
        newNodes = append(newNodes, node)
    }
    if !found {
        return ErrNodeDefinitionNotFound
    }

    // 2. Remove all edges connected to this node
    newEdges := make([]*Edge, 0, len(w.Edges))
    for _, edge := range w.Edges {
        if edge.FromNodeID == nodeID || edge.ToNodeID == nodeID {
            continue // Skip edges connected to removed node
        }
        newEdges = append(newEdges, edge)
    }

    w.NodeDefinitions = newNodes
    w.Edges = newEdges
    w.SetUpdatedAt(time.Now().UTC())
    return nil
}

// AddEdge adds a new edge connecting two node definitions.
func (w *Workflow) AddEdge(idFactory id.Factory, fromNodeID, toNodeID string) (*Edge, error) {
    // 1. Validate: no self-loops
    if fromNodeID == toNodeID {
        return nil, ErrSelfLoopNotAllowed
    }

    // 2. Validate: both nodes exist in this workflow
    fromExists := false
    toExists := false
    for _, node := range w.NodeDefinitions {
        if node.ID == fromNodeID {
            fromExists = true
        }
        if node.ID == toNodeID {
            toExists = true
        }
    }
    if !fromExists || !toExists {
        return nil, ErrNodeDefinitionNotFound
    }

    // 3. Validate: edge doesn't already exist
    for _, edge := range w.Edges {
        if edge.FromNodeID == fromNodeID && edge.ToNodeID == toNodeID {
            return nil, ErrEdgeAlreadyExists
        }
    }

    // 4. Create and add edge
    edgeID := idFactory.New()
    edge := newEdge(edgeID, w.ID, fromNodeID, toNodeID)
    w.Edges = append(w.Edges, edge)
    w.SetUpdatedAt(time.Now().UTC())
    return edge, nil
}

// RemoveEdge removes an edge by ID.
func (w *Workflow) RemoveEdge(edgeID string) error {
    found := false
    newEdges := make([]*Edge, 0, len(w.Edges))
    for _, edge := range w.Edges {
        if edge.ID == edgeID {
            found = true
            continue
        }
        newEdges = append(newEdges, edge)
    }
    if !found {
        return errors.New("edge not found")
    }
    w.Edges = newEdges
    w.SetUpdatedAt(time.Now().UTC())
    return nil
}

// FindNodeDefinition finds a node definition by ID.
func (w *Workflow) FindNodeDefinition(nodeID string) *NodeDefinition {
    for _, node := range w.NodeDefinitions {
        if node.ID == nodeID {
            return node
        }
    }
    return nil
}

// FindEdge finds an edge by ID.
func (w *Workflow) FindEdge(edgeID string) *Edge {
    for _, edge := range w.Edges {
        if edge.ID == edgeID {
            return edge
        }
    }
    return nil
}
```

**Dependencies:** `pkg/domain`, `pkg/id`, `internal/domain/workflow/event`

**Tests Required:**
- Test newWorkflow creates aggregate with event
- Test ReconstituteWorkflow recreates aggregate without events
- Test UpdateName modifies name and adds event
- Test AddNodeDefinition adds node to slice
- Test RemoveNodeDefinition removes node and cascades to edges
- Test AddEdge validates nodes exist and no duplicates
- Test AddEdge rejects self-loops
- Test RemoveEdge removes edge by ID
- Test FindNodeDefinition returns correct node or nil
- Test FindEdge returns correct edge or nil

---

### Step 5: Create Workflow Factory

**File:** `internal/domain/workflow/aggregate/workflow_factory.go`

**Action:** CREATE

**Rationale:** Factory encapsulates aggregate creation and ID generation.

**Pseudocode:**

```go
package aggregate

import "use-open-workflow.io/engine/pkg/id"

// WorkflowFactory creates new Workflow aggregates.
type WorkflowFactory struct {
    idFactory id.Factory
}

// NewWorkflowFactory creates a new WorkflowFactory.
func NewWorkflowFactory(idFactory id.Factory) *WorkflowFactory {
    return &WorkflowFactory{
        idFactory: idFactory,
    }
}

// Make creates a new Workflow aggregate with the given name.
func (f *WorkflowFactory) Make(name string) *Workflow {
    return newWorkflow(f.idFactory, f.idFactory.New(), name)
}

// IDFactory returns the ID factory for use in aggregate methods.
func (f *WorkflowFactory) IDFactory() id.Factory {
    return f.idFactory
}
```

**Dependencies:** `pkg/id`

**Tests Required:**
- Test NewWorkflowFactory creates factory
- Test Make creates workflow with generated ID

---

### Step 6: Create Domain Events

**File:** `internal/domain/workflow/event/create_workflow.go`

**Action:** CREATE

**Rationale:** Domain events enable event sourcing and async processing via outbox.

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

// CreateWorkflow is emitted when a new workflow is created.
type CreateWorkflow struct {
    domain.BaseEvent
    WorkflowID string `json:"workflow_id"`
    Name       string `json:"name"`
}

// NewCreateWorkflow creates a new CreateWorkflow event.
func NewCreateWorkflow(idFactory id.Factory, workflowID, name string) *CreateWorkflow {
    return &CreateWorkflow{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "CreateWorkflow",
        ),
        WorkflowID: workflowID,
        Name:       name,
    }
}
```

**File:** `internal/domain/workflow/event/update_workflow.go`

**Action:** CREATE

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

// UpdateWorkflow is emitted when a workflow is updated.
type UpdateWorkflow struct {
    domain.BaseEvent
    WorkflowID string `json:"workflow_id"`
    Name       string `json:"name"`
}

// NewUpdateWorkflow creates a new UpdateWorkflow event.
func NewUpdateWorkflow(idFactory id.Factory, workflowID, name string) *UpdateWorkflow {
    return &UpdateWorkflow{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "UpdateWorkflow",
        ),
        WorkflowID: workflowID,
        Name:       name,
    }
}
```

**Dependencies:** `pkg/domain`, `pkg/id`

**Tests Required:**
- Test NewCreateWorkflow creates event with correct fields
- Test NewUpdateWorkflow creates event with correct fields

---

### Step 7: Create Port Layer DTOs

**File:** `internal/port/workflow/inbound/workflow_dto.go`

**Action:** CREATE

**Rationale:** DTOs define the data contract between HTTP layer and services.

**Pseudocode:**

```go
package inbound

import "time"

// WorkflowDTO represents a Workflow for API responses.
type WorkflowDTO struct {
    ID              string              `json:"id"`
    Name            string              `json:"name"`
    NodeDefinitions []NodeDefinitionDTO `json:"nodeDefinitions"`
    Edges           []EdgeDTO           `json:"edges"`
    CreatedAt       time.Time           `json:"createdAt"`
    UpdatedAt       time.Time           `json:"updatedAt"`
}

// NodeDefinitionDTO represents a NodeDefinition for API responses.
type NodeDefinitionDTO struct {
    ID             string  `json:"id"`
    WorkflowID     string  `json:"workflowId"`
    NodeTemplateID string  `json:"nodeTemplateId"`
    Name           string  `json:"name"`
    PositionX      float64 `json:"positionX"`
    PositionY      float64 `json:"positionY"`
}

// EdgeDTO represents an Edge for API responses.
type EdgeDTO struct {
    ID         string `json:"id"`
    WorkflowID string `json:"workflowId"`
    FromNodeID string `json:"fromNodeId"`
    ToNodeID   string `json:"toNodeId"`
}
```

**Dependencies:** None

**Tests Required:** None (data structures only)

---

### Step 8: Create Port Layer Service Interfaces

**File:** `internal/port/workflow/inbound/workflow_read_service.go`

**Action:** CREATE

**Rationale:** Define read service contract for dependency inversion.

**Pseudocode:**

```go
package inbound

import "context"

// WorkflowReadService defines read operations for workflows.
type WorkflowReadService interface {
    // List returns all workflows (without child entities for performance)
    List(ctx context.Context) ([]*WorkflowDTO, error)

    // GetByID returns a workflow with all its node definitions and edges.
    GetByID(ctx context.Context, id string) (*WorkflowDTO, error)
}
```

**File:** `internal/port/workflow/inbound/workflow_write_service.go`

**Action:** CREATE

**Pseudocode:**

```go
package inbound

import "context"

// CreateWorkflowInput contains data for creating a new workflow.
type CreateWorkflowInput struct {
    Name string `json:"name"`
}

// UpdateWorkflowInput contains data for updating a workflow.
type UpdateWorkflowInput struct {
    Name string `json:"name"`
}

// AddNodeDefinitionInput contains data for adding a node definition.
type AddNodeDefinitionInput struct {
    NodeTemplateID string  `json:"nodeTemplateId"`
    Name           string  `json:"name"`
    PositionX      float64 `json:"positionX"`
    PositionY      float64 `json:"positionY"`
}

// AddEdgeInput contains data for adding an edge.
type AddEdgeInput struct {
    FromNodeID string `json:"fromNodeId"`
    ToNodeID   string `json:"toNodeId"`
}

// WorkflowWriteService defines write operations for workflows.
type WorkflowWriteService interface {
    // Create creates a new workflow.
    Create(ctx context.Context, input CreateWorkflowInput) (*WorkflowDTO, error)

    // Update updates an existing workflow's properties.
    Update(ctx context.Context, id string, input UpdateWorkflowInput) (*WorkflowDTO, error)

    // Delete deletes a workflow and all its child entities.
    Delete(ctx context.Context, id string) error

    // AddNodeDefinition adds a node definition to a workflow.
    AddNodeDefinition(ctx context.Context, workflowID string, input AddNodeDefinitionInput) (*WorkflowDTO, error)

    // RemoveNodeDefinition removes a node definition from a workflow.
    RemoveNodeDefinition(ctx context.Context, workflowID, nodeID string) (*WorkflowDTO, error)

    // AddEdge adds an edge between two node definitions.
    AddEdge(ctx context.Context, workflowID string, input AddEdgeInput) (*WorkflowDTO, error)

    // RemoveEdge removes an edge from a workflow.
    RemoveEdge(ctx context.Context, workflowID, edgeID string) (*WorkflowDTO, error)
}
```

**Dependencies:** None

**Tests Required:** None (interface definitions)

---

### Step 9: Create Port Layer Mapper Interface

**File:** `internal/port/workflow/inbound/workflow_mapper.go`

**Action:** CREATE

**Rationale:** Mapper interface allows different mapping implementations.

**Pseudocode:**

```go
package inbound

import "use-open-workflow.io/engine/internal/domain/workflow/aggregate"

// WorkflowMapper converts between aggregates and DTOs.
type WorkflowMapper interface {
    // To converts a Workflow aggregate to a WorkflowDTO.
    To(workflow *aggregate.Workflow) (*WorkflowDTO, error)

    // ToList converts a slice of Workflow aggregates to WorkflowDTOs.
    ToList(workflows []*aggregate.Workflow) ([]*WorkflowDTO, error)
}
```

**Dependencies:** `internal/domain/workflow/aggregate`

**Tests Required:** None (interface definition)

---

### Step 10: Create Port Layer Repository Interfaces

**File:** `internal/port/workflow/outbound/workflow_read_repository.go`

**Action:** CREATE

**Rationale:** Define repository contract for data access.

**Pseudocode:**

```go
package outbound

import (
    "context"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
)

// WorkflowReadRepository defines read operations for workflows.
type WorkflowReadRepository interface {
    // FindMany returns all workflows with their child entities.
    FindMany(ctx context.Context) ([]*aggregate.Workflow, error)

    // FindByID returns a workflow by ID with all child entities, or nil if not found.
    FindByID(ctx context.Context, id string) (*aggregate.Workflow, error)
}
```

**File:** `internal/port/workflow/outbound/workflow_write_repository.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import (
    "context"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
)

// WorkflowWriteRepository defines write operations for workflows.
type WorkflowWriteRepository interface {
    // Save persists a new workflow with all its child entities.
    Save(ctx context.Context, workflow *aggregate.Workflow) error

    // Update updates an existing workflow and syncs all child entities.
    // This performs a full sync: inserts new, updates existing, deletes removed.
    Update(ctx context.Context, workflow *aggregate.Workflow) error

    // Delete removes a workflow by ID (cascade deletes child entities via FK).
    Delete(ctx context.Context, id string) error
}
```

**Dependencies:** `internal/domain/workflow/aggregate`

**Tests Required:** None (interface definitions)

---

### Step 11: Create Port Layer Repository Factory Interfaces

**File:** `internal/port/workflow/outbound/workflow_read_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// WorkflowReadRepositoryFactory creates UoW-scoped read repositories.
type WorkflowReadRepositoryFactory interface {
    Create(uow portOutbound.UnitOfWork) WorkflowReadRepository
}
```

**File:** `internal/port/workflow/outbound/workflow_write_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// WorkflowWriteRepositoryFactory creates UoW-scoped write repositories.
type WorkflowWriteRepositoryFactory interface {
    Create(uow portOutbound.UnitOfWork) WorkflowWriteRepository
}
```

**Dependencies:** `internal/port/outbound`

**Tests Required:** None (interface definitions)

---

### Step 12: Create Port Layer Models

**File:** `internal/port/workflow/outbound/workflow_model.go`

**Action:** CREATE

**Rationale:** Models represent database row structures.

**Pseudocode:**

```go
package outbound

import "time"

// WorkflowModel represents a workflow database row.
type WorkflowModel struct {
    ID        string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NodeDefinitionModel represents a node_definition database row.
type NodeDefinitionModel struct {
    ID             string
    WorkflowID     string
    NodeTemplateID string
    Name           string
    PositionX      float64
    PositionY      float64
}

// EdgeModel represents an edge database row.
type EdgeModel struct {
    ID         string
    WorkflowID string
    FromNodeID string
    ToNodeID   string
}
```

**Dependencies:** None

**Tests Required:** None (data structures only)

---

### Step 13: Create Adapter Mapper Implementation

**File:** `internal/adapter/workflow/inbound/workflow_mapper.go`

**Action:** CREATE

**Rationale:** Concrete implementation of aggregate-to-DTO conversion.

**Pseudocode:**

```go
package inbound

import (
    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// WorkflowMapper implements the WorkflowMapper interface.
type WorkflowMapper struct{}

// NewWorkflowMapper creates a new WorkflowMapper.
func NewWorkflowMapper() *WorkflowMapper {
    return &WorkflowMapper{}
}

// To converts a Workflow aggregate to a WorkflowDTO.
func (m *WorkflowMapper) To(workflow *aggregate.Workflow) (*inbound.WorkflowDTO, error) {
    // 1. Convert NodeDefinitions
    nodeDefinitions := make([]inbound.NodeDefinitionDTO, len(workflow.NodeDefinitions))
    for i, nd := range workflow.NodeDefinitions {
        nodeDefinitions[i] = inbound.NodeDefinitionDTO{
            ID:             nd.ID,
            WorkflowID:     nd.WorkflowID,
            NodeTemplateID: nd.NodeTemplateID,
            Name:           nd.Name,
            PositionX:      nd.PositionX,
            PositionY:      nd.PositionY,
        }
    }

    // 2. Convert Edges
    edges := make([]inbound.EdgeDTO, len(workflow.Edges))
    for i, e := range workflow.Edges {
        edges[i] = inbound.EdgeDTO{
            ID:         e.ID,
            WorkflowID: e.WorkflowID,
            FromNodeID: e.FromNodeID,
            ToNodeID:   e.ToNodeID,
        }
    }

    // 3. Build and return DTO
    return &inbound.WorkflowDTO{
        ID:              workflow.ID,
        Name:            workflow.Name,
        NodeDefinitions: nodeDefinitions,
        Edges:           edges,
        CreatedAt:       workflow.CreatedAt,
        UpdatedAt:       workflow.UpdatedAt,
    }, nil
}

// ToList converts a slice of Workflow aggregates to WorkflowDTOs.
func (m *WorkflowMapper) ToList(workflows []*aggregate.Workflow) ([]*inbound.WorkflowDTO, error) {
    result := make([]*inbound.WorkflowDTO, len(workflows))
    for i, w := range workflows {
        dto, err := m.To(w)
        if err != nil {
            return nil, err
        }
        result[i] = dto
    }
    return result, nil
}
```

**Dependencies:** `internal/domain/workflow/aggregate`, `internal/port/workflow/inbound`

**Tests Required:**
- Test To converts aggregate to DTO correctly
- Test ToList converts slice correctly
- Test handles empty NodeDefinitions and Edges

---

### Step 14: Create Adapter Read Service Implementation

**File:** `internal/adapter/workflow/inbound/workflow_read_service.go`

**Action:** CREATE

**Rationale:** Implements read operations using UoW pattern.

**Pseudocode:**

```go
package inbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/port/outbound"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
)

// WorkflowReadService implements the WorkflowReadService interface.
type WorkflowReadService struct {
    uowFactory            outbound.UnitOfWorkFactory
    readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory
    mapper                inbound.WorkflowMapper
}

// NewWorkflowReadService creates a new WorkflowReadService.
func NewWorkflowReadService(
    uowFactory outbound.UnitOfWorkFactory,
    readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory,
    mapper inbound.WorkflowMapper,
) *WorkflowReadService {
    return &WorkflowReadService{
        uowFactory:            uowFactory,
        readRepositoryFactory: readRepositoryFactory,
        mapper:                mapper,
    }
}

// List returns all workflows.
func (s *WorkflowReadService) List(ctx context.Context) ([]*inbound.WorkflowDTO, error) {
    // 1. Create UoW and repository
    uow := s.uowFactory.Create()
    readRepo := s.readRepositoryFactory.Create(uow)

    // 2. Begin transaction (for consistency)
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer uow.Rollback(txCtx) // Read-only, rollback is fine

    // 3. Fetch workflows
    workflows, err := readRepo.FindMany(txCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflows: %w", err)
    }

    // 4. Convert to DTOs
    return s.mapper.ToList(workflows)
}

// GetByID returns a workflow by ID.
func (s *WorkflowReadService) GetByID(ctx context.Context, id string) (*inbound.WorkflowDTO, error) {
    // 1. Create UoW and repository
    uow := s.uowFactory.Create()
    readRepo := s.readRepositoryFactory.Create(uow)

    // 2. Begin transaction
    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer uow.Rollback(txCtx)

    // 3. Fetch workflow
    workflow, err := readRepo.FindByID(txCtx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, nil // Not found
    }

    // 4. Convert to DTO
    return s.mapper.To(workflow)
}
```

**Dependencies:** `internal/port/outbound`, `internal/port/workflow/inbound`, `internal/port/workflow/outbound`

**Tests Required:**
- Test List returns all workflows as DTOs
- Test GetByID returns workflow as DTO
- Test GetByID returns nil for not found

---

### Step 15: Create Adapter Write Service Implementation

**File:** `internal/adapter/workflow/inbound/workflow_write_service.go`

**Action:** CREATE

**Rationale:** Implements write operations with transaction management.

**Pseudocode:**

```go
package inbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/port/outbound"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/pkg/id"
)

// WorkflowWriteService implements the WorkflowWriteService interface.
type WorkflowWriteService struct {
    uowFactory             outbound.UnitOfWorkFactory
    writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory
    readRepositoryFactory  workflowOutbound.WorkflowReadRepositoryFactory
    factory                *aggregate.WorkflowFactory
    mapper                 inbound.WorkflowMapper
    idFactory              id.Factory
}

// NewWorkflowWriteService creates a new WorkflowWriteService.
func NewWorkflowWriteService(
    uowFactory outbound.UnitOfWorkFactory,
    writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory,
    readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory,
    factory *aggregate.WorkflowFactory,
    mapper inbound.WorkflowMapper,
    idFactory id.Factory,
) *WorkflowWriteService {
    return &WorkflowWriteService{
        uowFactory:             uowFactory,
        writeRepositoryFactory: writeRepositoryFactory,
        readRepositoryFactory:  readRepositoryFactory,
        factory:                factory,
        mapper:                 mapper,
        idFactory:              idFactory,
    }
}

// Create creates a new workflow.
func (s *WorkflowWriteService) Create(ctx context.Context, input inbound.CreateWorkflowInput) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Create workflow aggregate
    workflow := s.factory.Make(input.Name)

    // 2. Save to repository
    if err = writeRepo.Save(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to save workflow: %w", err)
    }

    // 3. Commit transaction
    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}

// Update updates an existing workflow.
func (s *WorkflowWriteService) Update(ctx context.Context, id string, input inbound.UpdateWorkflowInput) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Load existing workflow
    workflow, err := readRepo.FindByID(txCtx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", id)
    }

    // 2. Update aggregate
    workflow.UpdateName(s.idFactory, input.Name)

    // 3. Persist changes
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}

// Delete deletes a workflow.
func (s *WorkflowWriteService) Delete(ctx context.Context, id string) error {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    if err = writeRepo.Delete(txCtx, id); err != nil {
        return fmt.Errorf("failed to delete workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// AddNodeDefinition adds a node definition to a workflow.
func (s *WorkflowWriteService) AddNodeDefinition(ctx context.Context, workflowID string, input inbound.AddNodeDefinitionInput) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Load workflow
    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    // 2. Add node definition to aggregate
    workflow.AddNodeDefinition(s.idFactory, input.NodeTemplateID, input.Name, input.PositionX, input.PositionY)

    // 3. Persist changes (full sync of child entities)
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}

// RemoveNodeDefinition removes a node definition from a workflow.
func (s *WorkflowWriteService) RemoveNodeDefinition(ctx context.Context, workflowID, nodeID string) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Load workflow
    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    // 2. Remove node from aggregate (cascades to edges)
    if err = workflow.RemoveNodeDefinition(nodeID); err != nil {
        return nil, fmt.Errorf("failed to remove node definition: %w", err)
    }

    // 3. Persist changes
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}

// AddEdge adds an edge between two node definitions.
func (s *WorkflowWriteService) AddEdge(ctx context.Context, workflowID string, input inbound.AddEdgeInput) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Load workflow
    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    // 2. Add edge to aggregate (validates nodes exist)
    _, err = workflow.AddEdge(s.idFactory, input.FromNodeID, input.ToNodeID)
    if err != nil {
        return nil, fmt.Errorf("failed to add edge: %w", err)
    }

    // 3. Persist changes
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}

// RemoveEdge removes an edge from a workflow.
func (s *WorkflowWriteService) RemoveEdge(ctx context.Context, workflowID, edgeID string) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    // 1. Load workflow
    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    // 2. Remove edge from aggregate
    if err = workflow.RemoveEdge(edgeID); err != nil {
        return nil, fmt.Errorf("failed to remove edge: %w", err)
    }

    // 3. Persist changes
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.To(workflow)
}
```

**Dependencies:** All port layer interfaces, domain aggregate, id factory

**Tests Required:**
- Test Create creates and persists workflow
- Test Update loads, modifies, and persists
- Test Delete removes workflow
- Test AddNodeDefinition adds node to existing workflow
- Test RemoveNodeDefinition removes node and connected edges
- Test AddEdge validates and adds edge
- Test RemoveEdge removes edge
- Test error cases for not found scenarios

---

### Step 16: Create Adapter PostgreSQL Read Repository

**File:** `internal/adapter/workflow/outbound/workflow_postgres_read_repository.go`

**Action:** CREATE

**Rationale:** Implements database read operations with parent-child loading.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "fmt"
    "time"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresReadRepository implements WorkflowReadRepository.
type WorkflowPostgresReadRepository struct {
    uow portOutbound.UnitOfWork
}

// NewWorkflowPostgresReadRepository creates a new repository.
func NewWorkflowPostgresReadRepository(uow portOutbound.UnitOfWork) *WorkflowPostgresReadRepository {
    return &WorkflowPostgresReadRepository{uow: uow}
}

// FindMany returns all workflows with their child entities.
func (r *WorkflowPostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.Workflow, error) {
    q := r.uow.Querier(ctx)

    // 1. Query all workflows
    rows, err := q.Query(ctx, `
        SELECT id, name, created_at, updated_at
        FROM workflow
        ORDER BY created_at DESC
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to query workflows: %w", err)
    }
    defer rows.Close()

    // 2. Build workflow map
    workflowMap := make(map[string]*aggregate.Workflow)
    workflowIDs := make([]string, 0)

    for rows.Next() {
        var id, name string
        var createdAt, updatedAt time.Time
        if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan workflow: %w", err)
        }
        workflow := aggregate.ReconstituteWorkflow(id, name, createdAt, updatedAt, nil, nil)
        workflowMap[id] = workflow
        workflowIDs = append(workflowIDs, id)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }

    if len(workflowIDs) == 0 {
        return []*aggregate.Workflow{}, nil
    }

    // 3. Load all node definitions for these workflows
    if err := r.loadNodeDefinitions(ctx, workflowMap); err != nil {
        return nil, err
    }

    // 4. Load all edges for these workflows
    if err := r.loadEdges(ctx, workflowMap); err != nil {
        return nil, err
    }

    // 5. Preserve order
    result := make([]*aggregate.Workflow, len(workflowIDs))
    for i, id := range workflowIDs {
        result[i] = workflowMap[id]
    }

    return result, nil
}

// FindByID returns a workflow by ID with all child entities.
func (r *WorkflowPostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.Workflow, error) {
    q := r.uow.Querier(ctx)

    // 1. Query workflow
    var name string
    var createdAt, updatedAt time.Time
    err := q.QueryRow(ctx, `
        SELECT id, name, created_at, updated_at
        FROM workflow
        WHERE id = $1
    `, id).Scan(&id, &name, &createdAt, &updatedAt)

    if err != nil && err.Error() == "no rows in result set" {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query workflow: %w", err)
    }

    // 2. Query node definitions
    nodeRows, err := q.Query(ctx, `
        SELECT id, workflow_id, node_template_id, name, position_x, position_y
        FROM node_definition
        WHERE workflow_id = $1
    `, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query node definitions: %w", err)
    }
    defer nodeRows.Close()

    nodeDefinitions := make([]*aggregate.NodeDefinition, 0)
    for nodeRows.Next() {
        var ndID, wfID, ntID, ndName string
        var posX, posY float64
        if err := nodeRows.Scan(&ndID, &wfID, &ntID, &ndName, &posX, &posY); err != nil {
            return nil, fmt.Errorf("failed to scan node definition: %w", err)
        }
        nodeDefinitions = append(nodeDefinitions, aggregate.ReconstituteNodeDefinition(ndID, wfID, ntID, ndName, posX, posY))
    }
    if err := nodeRows.Err(); err != nil {
        return nil, fmt.Errorf("node row iteration error: %w", err)
    }

    // 3. Query edges
    edgeRows, err := q.Query(ctx, `
        SELECT id, workflow_id, from_node_id, to_node_id
        FROM edge
        WHERE workflow_id = $1
    `, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query edges: %w", err)
    }
    defer edgeRows.Close()

    edges := make([]*aggregate.Edge, 0)
    for edgeRows.Next() {
        var eID, wfID, fromID, toID string
        if err := edgeRows.Scan(&eID, &wfID, &fromID, &toID); err != nil {
            return nil, fmt.Errorf("failed to scan edge: %w", err)
        }
        edges = append(edges, aggregate.ReconstituteEdge(eID, wfID, fromID, toID))
    }
    if err := edgeRows.Err(); err != nil {
        return nil, fmt.Errorf("edge row iteration error: %w", err)
    }

    return aggregate.ReconstituteWorkflow(id, name, createdAt, updatedAt, nodeDefinitions, edges), nil
}

// loadNodeDefinitions loads node definitions for all workflows in map.
func (r *WorkflowPostgresReadRepository) loadNodeDefinitions(ctx context.Context, workflowMap map[string]*aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    rows, err := q.Query(ctx, `
        SELECT id, workflow_id, node_template_id, name, position_x, position_y
        FROM node_definition
        WHERE workflow_id = ANY($1)
    `, r.getWorkflowIDs(workflowMap))
    if err != nil {
        return fmt.Errorf("failed to query node definitions: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id, workflowID, nodeTemplateID, name string
        var posX, posY float64
        if err := rows.Scan(&id, &workflowID, &nodeTemplateID, &name, &posX, &posY); err != nil {
            return fmt.Errorf("failed to scan node definition: %w", err)
        }
        nd := aggregate.ReconstituteNodeDefinition(id, workflowID, nodeTemplateID, name, posX, posY)
        if wf, ok := workflowMap[workflowID]; ok {
            wf.NodeDefinitions = append(wf.NodeDefinitions, nd)
        }
    }
    return rows.Err()
}

// loadEdges loads edges for all workflows in map.
func (r *WorkflowPostgresReadRepository) loadEdges(ctx context.Context, workflowMap map[string]*aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    rows, err := q.Query(ctx, `
        SELECT id, workflow_id, from_node_id, to_node_id
        FROM edge
        WHERE workflow_id = ANY($1)
    `, r.getWorkflowIDs(workflowMap))
    if err != nil {
        return fmt.Errorf("failed to query edges: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id, workflowID, fromNodeID, toNodeID string
        if err := rows.Scan(&id, &workflowID, &fromNodeID, &toNodeID); err != nil {
            return fmt.Errorf("failed to scan edge: %w", err)
        }
        edge := aggregate.ReconstituteEdge(id, workflowID, fromNodeID, toNodeID)
        if wf, ok := workflowMap[workflowID]; ok {
            wf.Edges = append(wf.Edges, edge)
        }
    }
    return rows.Err()
}

func (r *WorkflowPostgresReadRepository) getWorkflowIDs(workflowMap map[string]*aggregate.Workflow) []string {
    ids := make([]string, 0, len(workflowMap))
    for id := range workflowMap {
        ids = append(ids, id)
    }
    return ids
}
```

**Dependencies:** `internal/domain/workflow/aggregate`, `internal/port/outbound`

**Tests Required:**
- Test FindMany returns all workflows with children
- Test FindMany returns empty slice when no workflows
- Test FindByID returns workflow with children
- Test FindByID returns nil for not found

---

### Step 17: Create Adapter PostgreSQL Write Repository

**File:** `internal/adapter/workflow/outbound/workflow_postgres_write_repository.go`

**Action:** CREATE

**Rationale:** Implements database write operations with full child entity sync.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresWriteRepository implements WorkflowWriteRepository.
type WorkflowPostgresWriteRepository struct {
    uow portOutbound.UnitOfWork
}

// NewWorkflowPostgresWriteRepository creates a new repository.
func NewWorkflowPostgresWriteRepository(uow portOutbound.UnitOfWork) *WorkflowPostgresWriteRepository {
    return &WorkflowPostgresWriteRepository{uow: uow}
}

// Save persists a new workflow with all its child entities.
func (r *WorkflowPostgresWriteRepository) Save(ctx context.Context, workflow *aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    // 1. Insert workflow
    _, err := q.Exec(ctx, `
        INSERT INTO workflow (id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
    `, workflow.ID, workflow.Name, workflow.CreatedAt, workflow.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to save workflow: %w", err)
    }

    // 2. Insert node definitions
    for _, nd := range workflow.NodeDefinitions {
        _, err := q.Exec(ctx, `
            INSERT INTO node_definition (id, workflow_id, node_template_id, name, position_x, position_y)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, nd.ID, nd.WorkflowID, nd.NodeTemplateID, nd.Name, nd.PositionX, nd.PositionY)
        if err != nil {
            return fmt.Errorf("failed to save node definition: %w", err)
        }
    }

    // 3. Insert edges
    for _, edge := range workflow.Edges {
        _, err := q.Exec(ctx, `
            INSERT INTO edge (id, workflow_id, from_node_id, to_node_id)
            VALUES ($1, $2, $3, $4)
        `, edge.ID, edge.WorkflowID, edge.FromNodeID, edge.ToNodeID)
        if err != nil {
            return fmt.Errorf("failed to save edge: %w", err)
        }
    }

    // 4. Register aggregate for event publishing
    r.uow.RegisterNew(workflow)

    return nil
}

// Update updates an existing workflow and syncs all child entities.
// Uses delete-and-insert strategy for simplicity.
func (r *WorkflowPostgresWriteRepository) Update(ctx context.Context, workflow *aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    // 1. Update workflow
    _, err := q.Exec(ctx, `
        UPDATE workflow
        SET name = $1, updated_at = $2
        WHERE id = $3
    `, workflow.Name, workflow.UpdatedAt, workflow.ID)
    if err != nil {
        return fmt.Errorf("failed to update workflow: %w", err)
    }

    // 2. Delete existing edges (must delete before nodes due to FK)
    _, err = q.Exec(ctx, `DELETE FROM edge WHERE workflow_id = $1`, workflow.ID)
    if err != nil {
        return fmt.Errorf("failed to delete existing edges: %w", err)
    }

    // 3. Delete existing node definitions
    _, err = q.Exec(ctx, `DELETE FROM node_definition WHERE workflow_id = $1`, workflow.ID)
    if err != nil {
        return fmt.Errorf("failed to delete existing node definitions: %w", err)
    }

    // 4. Re-insert node definitions
    for _, nd := range workflow.NodeDefinitions {
        _, err := q.Exec(ctx, `
            INSERT INTO node_definition (id, workflow_id, node_template_id, name, position_x, position_y)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, nd.ID, nd.WorkflowID, nd.NodeTemplateID, nd.Name, nd.PositionX, nd.PositionY)
        if err != nil {
            return fmt.Errorf("failed to insert node definition: %w", err)
        }
    }

    // 5. Re-insert edges
    for _, edge := range workflow.Edges {
        _, err := q.Exec(ctx, `
            INSERT INTO edge (id, workflow_id, from_node_id, to_node_id)
            VALUES ($1, $2, $3, $4)
        `, edge.ID, edge.WorkflowID, edge.FromNodeID, edge.ToNodeID)
        if err != nil {
            return fmt.Errorf("failed to insert edge: %w", err)
        }
    }

    // 6. Register aggregate for event publishing
    r.uow.RegisterDirty(workflow)

    return nil
}

// Delete removes a workflow by ID.
// Child entities are deleted via CASCADE.
func (r *WorkflowPostgresWriteRepository) Delete(ctx context.Context, id string) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `DELETE FROM workflow WHERE id = $1`, id)
    if err != nil {
        return fmt.Errorf("failed to delete workflow: %w", err)
    }

    return nil
}
```

**Dependencies:** `internal/domain/workflow/aggregate`, `internal/port/outbound`

**Tests Required:**
- Test Save persists workflow and children
- Test Update syncs all child entities
- Test Delete removes workflow (FK cascade removes children)

---

### Step 18: Create Repository Factories

**File:** `internal/adapter/workflow/outbound/workflow_postgres_read_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import (
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresReadRepositoryFactory creates read repositories.
type WorkflowPostgresReadRepositoryFactory struct{}

// NewWorkflowPostgresReadRepositoryFactory creates a new factory.
func NewWorkflowPostgresReadRepositoryFactory() *WorkflowPostgresReadRepositoryFactory {
    return &WorkflowPostgresReadRepositoryFactory{}
}

// Create creates a UoW-scoped read repository.
func (f *WorkflowPostgresReadRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowReadRepository {
    return NewWorkflowPostgresReadRepository(uow)
}
```

**File:** `internal/adapter/workflow/outbound/workflow_postgres_write_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import (
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresWriteRepositoryFactory creates write repositories.
type WorkflowPostgresWriteRepositoryFactory struct{}

// NewWorkflowPostgresWriteRepositoryFactory creates a new factory.
func NewWorkflowPostgresWriteRepositoryFactory() *WorkflowPostgresWriteRepositoryFactory {
    return &WorkflowPostgresWriteRepositoryFactory{}
}

// Create creates a UoW-scoped write repository.
func (f *WorkflowPostgresWriteRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowWriteRepository {
    return NewWorkflowPostgresWriteRepository(uow)
}
```

**Dependencies:** Port layer interfaces

**Tests Required:** None (simple factories)

---

### Step 19: Create HTTP Handler

**File:** `api/workflow/http/workflow_handler.go`

**Action:** CREATE

**Rationale:** Exposes workflow operations via REST API.

**Pseudocode:**

```go
package http

import (
    "github.com/gofiber/fiber/v3"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// WorkflowHandler handles HTTP requests for workflow operations.
type WorkflowHandler struct {
    readService  inbound.WorkflowReadService
    writeService inbound.WorkflowWriteService
}

// NewWorkflowHandler creates a new handler.
func NewWorkflowHandler(
    readService inbound.WorkflowReadService,
    writeService inbound.WorkflowWriteService,
) *WorkflowHandler {
    return &WorkflowHandler{
        readService:  readService,
        writeService: writeService,
    }
}

// List handles GET /api/v1/workflow
func (h *WorkflowHandler) List(c fiber.Ctx) error {
    workflows, err := h.readService.List(c.Context())
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    return c.JSON(workflows)
}

// GetByID handles GET /api/v1/workflow/:id
func (h *WorkflowHandler) GetByID(c fiber.Ctx) error {
    id := c.Params("id")
    workflow, err := h.readService.GetByID(c.Context(), id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    if workflow == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "workflow not found",
        })
    }
    return c.JSON(workflow)
}

// Create handles POST /api/v1/workflow
func (h *WorkflowHandler) Create(c fiber.Ctx) error {
    var input inbound.CreateWorkflowInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    workflow, err := h.writeService.Create(c.Context(), input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(workflow)
}

// Update handles PUT /api/v1/workflow/:id
func (h *WorkflowHandler) Update(c fiber.Ctx) error {
    id := c.Params("id")
    var input inbound.UpdateWorkflowInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    workflow, err := h.writeService.Update(c.Context(), id, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(workflow)
}

// Delete handles DELETE /api/v1/workflow/:id
func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
    id := c.Params("id")
    if err := h.writeService.Delete(c.Context(), id); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.SendStatus(fiber.StatusNoContent)
}

// AddNodeDefinition handles POST /api/v1/workflow/:id/node
func (h *WorkflowHandler) AddNodeDefinition(c fiber.Ctx) error {
    workflowID := c.Params("id")
    var input inbound.AddNodeDefinitionInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    workflow, err := h.writeService.AddNodeDefinition(c.Context(), workflowID, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(workflow)
}

// RemoveNodeDefinition handles DELETE /api/v1/workflow/:id/node/:nodeId
func (h *WorkflowHandler) RemoveNodeDefinition(c fiber.Ctx) error {
    workflowID := c.Params("id")
    nodeID := c.Params("nodeId")

    workflow, err := h.writeService.RemoveNodeDefinition(c.Context(), workflowID, nodeID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(workflow)
}

// AddEdge handles POST /api/v1/workflow/:id/edge
func (h *WorkflowHandler) AddEdge(c fiber.Ctx) error {
    workflowID := c.Params("id")
    var input inbound.AddEdgeInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    workflow, err := h.writeService.AddEdge(c.Context(), workflowID, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(workflow)
}

// RemoveEdge handles DELETE /api/v1/workflow/:id/edge/:edgeId
func (h *WorkflowHandler) RemoveEdge(c fiber.Ctx) error {
    workflowID := c.Params("id")
    edgeID := c.Params("edgeId")

    workflow, err := h.writeService.RemoveEdge(c.Context(), workflowID, edgeID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(workflow)
}
```

**Dependencies:** `github.com/gofiber/fiber/v3`, `internal/port/workflow/inbound`

**Tests Required:**
- Test each handler method returns correct HTTP status codes
- Test JSON binding for invalid input
- Test not found returns 404

---

### Step 20: Register Routes in Router

**File:** `api/router.go`

**Action:** MODIFY

**Rationale:** Make workflow endpoints available via the API.

**Pseudocode:**

```go
// Add import for workflow handler
import (
    // ... existing imports
    workflowHttp "use-open-workflow.io/engine/api/workflow/http"
)

func SetupRouter(c *di.Container) *fiber.App {
    // ... existing code ...

    api := app.Group("/api/v1")
    registerNodeTemplateRoutes(api, c)
    registerWorkflowRoutes(api, c)  // ADD THIS LINE

    return app
}

// ADD THIS FUNCTION
func registerWorkflowRoutes(router fiber.Router, c *di.Container) {
    workflowHandler := workflowHttp.NewWorkflowHandler(
        c.WorkflowReadService,
        c.WorkflowWriteService,
    )

    workflow := router.Group("/workflow")
    workflow.Get("/", workflowHandler.List)
    workflow.Get("/:id", workflowHandler.GetByID)
    workflow.Post("/", workflowHandler.Create)
    workflow.Put("/:id", workflowHandler.Update)
    workflow.Delete("/:id", workflowHandler.Delete)

    // Node definition routes
    workflow.Post("/:id/node", workflowHandler.AddNodeDefinition)
    workflow.Delete("/:id/node/:nodeId", workflowHandler.RemoveNodeDefinition)

    // Edge routes
    workflow.Post("/:id/edge", workflowHandler.AddEdge)
    workflow.Delete("/:id/edge/:edgeId", workflowHandler.RemoveEdge)
}
```

**Dependencies:** `api/workflow/http`, `di.Container` with workflow services

**Tests Required:**
- Verify routes are registered correctly

---

### Step 21: Wire Up Dependencies in Container

**File:** `di/container.go`

**Action:** MODIFY

**Rationale:** Inject all workflow dependencies for the application.

**Pseudocode:**

```go
// Add imports
import (
    // ... existing imports
    workflowAdapterInbound "use-open-workflow.io/engine/internal/adapter/workflow/inbound"
    workflowAdapterOutbound "use-open-workflow.io/engine/internal/adapter/workflow/outbound"
    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    workflowInbound "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// Update Container struct
type Container struct {
    Pool                     *pgxpool.Pool
    NodeTemplateReadService  inbound.NodeTemplateReadService
    NodeTemplateWriteService inbound.NodeTemplateWriteService
    WorkflowReadService      workflowInbound.WorkflowReadService   // ADD
    WorkflowWriteService     workflowInbound.WorkflowWriteService  // ADD
    OutboxProcessor          outbound.OutboxProcessor
}

// In NewContainer function, add after NodeTemplate wiring:

    // Workflow domain wiring

    // Mappers
    workflowInboundMapper := workflowAdapterInbound.NewWorkflowMapper()

    // Factory
    workflowFactory := aggregate.NewWorkflowFactory(idFactory)

    // Repository Factories
    workflowReadRepositoryFactory := workflowAdapterOutbound.NewWorkflowPostgresReadRepositoryFactory()
    workflowWriteRepositoryFactory := workflowAdapterOutbound.NewWorkflowPostgresWriteRepositoryFactory()

    // Services
    workflowReadService := workflowAdapterInbound.NewWorkflowReadService(
        uowFactory,
        workflowReadRepositoryFactory,
        workflowInboundMapper,
    )

    workflowWriteService := workflowAdapterInbound.NewWorkflowWriteService(
        uowFactory,
        workflowWriteRepositoryFactory,
        workflowReadRepositoryFactory,
        workflowFactory,
        workflowInboundMapper,
        idFactory,
    )

// Update return statement
    return &Container{
        Pool:                     pool,
        NodeTemplateReadService:  nodeTemplateReadService,
        NodeTemplateWriteService: nodeTemplateWriteService,
        WorkflowReadService:      workflowReadService,   // ADD
        WorkflowWriteService:     workflowWriteService,  // ADD
        OutboxProcessor:          outboxProcessor,
    }, nil
```

**Dependencies:** All workflow adapters and ports

**Tests Required:**
- Verify container builds successfully with all dependencies

---

## 4. Data Changes

### Schema/Model Updates

**New Tables:**

```sql
-- workflow: Aggregate root
workflow (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
)

-- node_definition: Child entity
node_definition (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    node_template_id VARCHAR(26) NOT NULL REFERENCES node_template(id),
    name VARCHAR(255) NOT NULL,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0
)

-- edge: Child entity
edge (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    from_node_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    to_node_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    UNIQUE (workflow_id, from_node_id, to_node_id)
)
```

### Migration Notes

- Migration must be applied after 001_initial_schema.sql
- Uses CASCADE delete to ensure child entities are cleaned up with parent
- Foreign key to node_template ensures referential integrity
- No backward compatibility concerns (new tables only)

---

## 5. Integration Points

| Service | Interaction | Error Handling |
|---------|-------------|----------------|
| NodeTemplate domain | NodeDefinition references node_template.id | Foreign key constraint returns error if template doesn't exist |
| Outbox pattern | Domain events persisted on commit | Transaction rollback if outbox write fails |
| UnitOfWork | Transaction management for all operations | Deferred rollback in service methods |

---

## 6. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Workflow not found on update/delete | Return error "workflow not found: {id}" |
| NodeDefinition references non-existent NodeTemplate | FK constraint error from database |
| Edge references non-existent NodeDefinition | Domain validation rejects with ErrNodeDefinitionNotFound |
| Self-loop edge (from == to) | Domain validation rejects with ErrSelfLoopNotAllowed |
| Duplicate edge between same nodes | Domain validation rejects with ErrEdgeAlreadyExists |
| Remove node that has connected edges | Domain method RemoveNodeDefinition cascades to remove edges |
| Empty workflow name | Allow (no validation currently, could add if needed) |
| Transaction failure mid-operation | UoW rollback ensures consistency |

---

## 7. Testing Strategy

### Unit Tests

- **Aggregate tests** (`workflow_test.go`):
  - Test Workflow creation with events
  - Test AddNodeDefinition adds to slice
  - Test RemoveNodeDefinition cascades to edges
  - Test AddEdge validates nodes and duplicates
  - Test RemoveEdge removes by ID
  - Test error cases return appropriate errors

- **Entity tests**:
  - Test NodeDefinition creation and reconstitution
  - Test Edge creation and reconstitution

- **Mapper tests**:
  - Test To converts aggregate with children to DTO
  - Test ToList handles empty and populated slices

### Integration Tests

- **Repository tests** (if applicable):
  - Test Save/FindByID roundtrip
  - Test Update syncs child entities correctly
  - Test Delete cascades to children
  - Test FindMany loads all children

### Manual Verification

1. Start application with `make run`
2. Apply migration: `psql -f migration/002_workflow_schema.sql`
3. Create workflow: `POST /api/v1/workflow` with `{"name": "Test Workflow"}`
4. Add node: `POST /api/v1/workflow/{id}/node` with node template reference
5. Add another node
6. Add edge: `POST /api/v1/workflow/{id}/edge` connecting the two nodes
7. Get workflow: `GET /api/v1/workflow/{id}` - verify all children returned
8. Remove edge: `DELETE /api/v1/workflow/{id}/edge/{edgeId}`
9. Remove node: `DELETE /api/v1/workflow/{id}/node/{nodeId}`
10. Delete workflow: `DELETE /api/v1/workflow/{id}`
11. Verify cascade delete removed children from database

---

## 8. Implementation Order

1. **Step 1: Migration** — Database tables must exist first
2. **Steps 2-6: Domain layer** — Core business logic, no external dependencies
3. **Steps 7-12: Port layer** — Define contracts before implementations
4. **Step 13: Mapper** — Needed by services
5. **Steps 14-18: Adapter layer services and repositories** — Implement contracts
6. **Step 19: HTTP handler** — Expose via API
7. **Steps 20-21: Wiring** — Connect everything in router and DI container

This order ensures each step builds on the previous, minimizing back-and-forth changes.

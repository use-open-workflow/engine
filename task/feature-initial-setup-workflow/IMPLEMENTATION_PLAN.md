# Implementation Plan

## 1. Implementation Summary

Implement the Workflow domain with three related components: Workflow (aggregate), NodeDefinition (entity), and Edge (entity). The Workflow aggregate owns both NodeDefinition and Edge collections. NodeDefinition references a NodeTemplate by ID, and Edge connects two NodeDefinition IDs (from/to) within the same Workflow. Following the established hexagonal architecture patterns, we will create domain models, ports (inbound/outbound), adapters, HTTP handlers, and wire everything through the DI container.

## 2. Change Manifest

```
CREATE:
- migration/002_workflow_schema.sql — Database schema for workflow, node_definition, edge tables

- internal/domain/workflow/aggregate/workflow.go — Workflow aggregate with embedded entities
- internal/domain/workflow/aggregate/workflow_factory.go — Factory for Workflow creation
- internal/domain/workflow/entity/node_definition.go — NodeDefinition entity
- internal/domain/workflow/entity/edge.go — Edge entity
- internal/domain/workflow/event/create_workflow.go — CreateWorkflow domain event
- internal/domain/workflow/event/update_workflow.go — UpdateWorkflow domain event
- internal/domain/workflow/event/add_node_definition.go — AddNodeDefinition domain event
- internal/domain/workflow/event/remove_node_definition.go — RemoveNodeDefinition domain event
- internal/domain/workflow/event/add_edge.go — AddEdge domain event
- internal/domain/workflow/event/remove_edge.go — RemoveEdge domain event

- internal/port/workflow/inbound/workflow_dto.go — DTOs for workflow, node_definition, edge
- internal/port/workflow/inbound/workflow_read_service.go — Read service interface
- internal/port/workflow/inbound/workflow_write_service.go — Write service interface
- internal/port/workflow/inbound/workflow_mapper.go — Mapper interface

- internal/port/workflow/outbound/workflow_model.go — Database models
- internal/port/workflow/outbound/workflow_read_repository.go — Read repository interface
- internal/port/workflow/outbound/workflow_write_repository.go — Write repository interface
- internal/port/workflow/outbound/workflow_read_repository_factory.go — Read repository factory interface
- internal/port/workflow/outbound/workflow_write_repository_factory.go — Write repository factory interface

- internal/adapter/workflow/inbound/workflow_read_service.go — Read service implementation
- internal/adapter/workflow/inbound/workflow_write_service.go — Write service implementation
- internal/adapter/workflow/inbound/workflow_mapper.go — Mapper implementation

- internal/adapter/workflow/outbound/workflow_postgres_read_repository.go — PostgreSQL read repository
- internal/adapter/workflow/outbound/workflow_postgres_write_repository.go — PostgreSQL write repository
- internal/adapter/workflow/outbound/workflow_postgres_read_repository_factory.go — Read repository factory
- internal/adapter/workflow/outbound/workflow_postgres_write_repository_factory.go — Write repository factory

- api/workflow/http/workflow_handler.go — HTTP handlers for workflow CRUD

MODIFY:
- api/router.go — Add registerWorkflowRoutes() function
- di/container.go — Add WorkflowReadService, WorkflowWriteService fields and wiring
```

## 3. Step-by-Step Plan

### Step 1: Create Database Migration

**File:** `migration/002_workflow_schema.sql`

**Action:** CREATE

**Rationale:** Define the database schema for workflow, node_definition, and edge tables with proper foreign key constraints.

**Pseudocode:**

```sql
-- Workflow table
CREATE TABLE IF NOT EXISTS workflow (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_created_at ON workflow(created_at DESC);

-- NodeDefinition table (belongs to Workflow)
CREATE TABLE IF NOT EXISTS node_definition (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    node_template_id VARCHAR(26) NOT NULL REFERENCES node_template(id),
    name VARCHAR(255) NOT NULL,
    config JSONB,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_node_definition_workflow_id ON node_definition(workflow_id);

-- Edge table (connects NodeDefinitions within a Workflow)
CREATE TABLE IF NOT EXISTS edge (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    from_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    to_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Prevent duplicate edges
    CONSTRAINT unique_edge UNIQUE (workflow_id, from_node_definition_id, to_node_definition_id),
    -- Prevent self-loops
    CONSTRAINT no_self_loop CHECK (from_node_definition_id != to_node_definition_id)
);

CREATE INDEX IF NOT EXISTS idx_edge_workflow_id ON edge(workflow_id);
CREATE INDEX IF NOT EXISTS idx_edge_from_node ON edge(from_node_definition_id);
CREATE INDEX IF NOT EXISTS idx_edge_to_node ON edge(to_node_definition_id);
```

**Dependencies:** None

**Tests Required:**
- Manual verification: Run migration against database

---

### Step 2: Create NodeDefinition Entity

**File:** `internal/domain/workflow/entity/node_definition.go`

**Action:** CREATE

**Rationale:** NodeDefinition entity represents a node instance within a workflow, referencing a NodeTemplate.

**Pseudocode:**

```go
package entity

import (
    "time"
    "use-open-workflow.io/engine/pkg/domain"
)

// NodeDefinition represents a node instance within a workflow
type NodeDefinition struct {
    domain.BaseEntity
    WorkflowID     string
    NodeTemplateID string
    Name           string
    Config         map[string]interface{}  // JSONB config
    PositionX      float64
    PositionY      float64
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// NewNodeDefinition creates a new NodeDefinition
// Parameters:
//   - id: ULID for this entity
//   - workflowID: parent workflow ID
//   - nodeTemplateID: reference to NodeTemplate
//   - name: display name for this node
//   - config: node-specific configuration (can be nil)
//   - positionX, positionY: visual position on canvas
func NewNodeDefinition(
    id string,
    workflowID string,
    nodeTemplateID string,
    name string,
    config map[string]interface{},
    positionX float64,
    positionY float64,
) *NodeDefinition {
    now := time.Now().UTC()
    return &NodeDefinition{
        BaseEntity:     domain.NewBaseEntity(id),
        WorkflowID:     workflowID,
        NodeTemplateID: nodeTemplateID,
        Name:           name,
        Config:         config,
        PositionX:      positionX,
        PositionY:      positionY,
        CreatedAt:      now,
        UpdatedAt:      now,
    }
}

// ReconstituteNodeDefinition recreates entity from database
func ReconstituteNodeDefinition(
    id string,
    workflowID string,
    nodeTemplateID string,
    name string,
    config map[string]interface{},
    positionX float64,
    positionY float64,
    createdAt time.Time,
    updatedAt time.Time,
) *NodeDefinition {
    return &NodeDefinition{
        BaseEntity:     domain.NewBaseEntity(id),
        WorkflowID:     workflowID,
        NodeTemplateID: nodeTemplateID,
        Name:           name,
        Config:         config,
        PositionX:      positionX,
        PositionY:      positionY,
        CreatedAt:      createdAt,
        UpdatedAt:      updatedAt,
    }
}

// UpdatePosition updates the visual position
func (n *NodeDefinition) UpdatePosition(x, y float64) {
    n.PositionX = x
    n.PositionY = y
    n.UpdatedAt = time.Now().UTC()
}

// UpdateConfig updates the node configuration
func (n *NodeDefinition) UpdateConfig(config map[string]interface{}) {
    n.Config = config
    n.UpdatedAt = time.Now().UTC()
}

// UpdateName updates the display name
func (n *NodeDefinition) UpdateName(name string) {
    n.Name = name
    n.UpdatedAt = time.Now().UTC()
}
```

**Dependencies:**
- `use-open-workflow.io/engine/pkg/domain`

**Tests Required:**
- Test NewNodeDefinition creates entity with correct timestamps
- Test ReconstituteNodeDefinition preserves all fields
- Test UpdatePosition updates position and timestamp
- Test UpdateConfig updates config and timestamp

---

### Step 3: Create Edge Entity

**File:** `internal/domain/workflow/entity/edge.go`

**Action:** CREATE

**Rationale:** Edge entity connects two NodeDefinition entities within a workflow.

**Pseudocode:**

```go
package entity

import (
    "time"
    "use-open-workflow.io/engine/pkg/domain"
)

// Edge connects two NodeDefinitions within a Workflow
type Edge struct {
    domain.BaseEntity
    WorkflowID           string
    FromNodeDefinitionID string
    ToNodeDefinitionID   string
    CreatedAt            time.Time
}

// NewEdge creates a new Edge
// Parameters:
//   - id: ULID for this entity
//   - workflowID: parent workflow ID
//   - fromNodeDefinitionID: source node definition
//   - toNodeDefinitionID: target node definition
func NewEdge(
    id string,
    workflowID string,
    fromNodeDefinitionID string,
    toNodeDefinitionID string,
) *Edge {
    return &Edge{
        BaseEntity:           domain.NewBaseEntity(id),
        WorkflowID:           workflowID,
        FromNodeDefinitionID: fromNodeDefinitionID,
        ToNodeDefinitionID:   toNodeDefinitionID,
        CreatedAt:            time.Now().UTC(),
    }
}

// ReconstituteEdge recreates entity from database
func ReconstituteEdge(
    id string,
    workflowID string,
    fromNodeDefinitionID string,
    toNodeDefinitionID string,
    createdAt time.Time,
) *Edge {
    return &Edge{
        BaseEntity:           domain.NewBaseEntity(id),
        WorkflowID:           workflowID,
        FromNodeDefinitionID: fromNodeDefinitionID,
        ToNodeDefinitionID:   toNodeDefinitionID,
        CreatedAt:            createdAt,
    }
}
```

**Dependencies:**
- `use-open-workflow.io/engine/pkg/domain`

**Tests Required:**
- Test NewEdge creates entity with correct fields
- Test ReconstituteEdge preserves all fields

---

### Step 4: Create Domain Events

**File:** `internal/domain/workflow/event/create_workflow.go`

**Action:** CREATE

**Rationale:** Domain event for workflow creation.

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

type CreateWorkflow struct {
    domain.BaseEvent
    WorkflowID  string `json:"workflow_id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

func NewCreateWorkflow(idFactory id.Factory, workflowID, name, description string) *CreateWorkflow {
    return &CreateWorkflow{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "CreateWorkflow",
        ),
        WorkflowID:  workflowID,
        Name:        name,
        Description: description,
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

type UpdateWorkflow struct {
    domain.BaseEvent
    WorkflowID  string `json:"workflow_id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

func NewUpdateWorkflow(idFactory id.Factory, workflowID, name, description string) *UpdateWorkflow {
    return &UpdateWorkflow{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "UpdateWorkflow",
        ),
        WorkflowID:  workflowID,
        Name:        name,
        Description: description,
    }
}
```

**File:** `internal/domain/workflow/event/add_node_definition.go`

**Action:** CREATE

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

type AddNodeDefinition struct {
    domain.BaseEvent
    WorkflowID       string `json:"workflow_id"`
    NodeDefinitionID string `json:"node_definition_id"`
    NodeTemplateID   string `json:"node_template_id"`
    Name             string `json:"name"`
}

func NewAddNodeDefinition(
    idFactory id.Factory,
    workflowID, nodeDefinitionID, nodeTemplateID, name string,
) *AddNodeDefinition {
    return &AddNodeDefinition{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "AddNodeDefinition",
        ),
        WorkflowID:       workflowID,
        NodeDefinitionID: nodeDefinitionID,
        NodeTemplateID:   nodeTemplateID,
        Name:             name,
    }
}
```

**File:** `internal/domain/workflow/event/remove_node_definition.go`

**Action:** CREATE

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

type RemoveNodeDefinition struct {
    domain.BaseEvent
    WorkflowID       string `json:"workflow_id"`
    NodeDefinitionID string `json:"node_definition_id"`
}

func NewRemoveNodeDefinition(
    idFactory id.Factory,
    workflowID, nodeDefinitionID string,
) *RemoveNodeDefinition {
    return &RemoveNodeDefinition{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "RemoveNodeDefinition",
        ),
        WorkflowID:       workflowID,
        NodeDefinitionID: nodeDefinitionID,
    }
}
```

**File:** `internal/domain/workflow/event/add_edge.go`

**Action:** CREATE

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

type AddEdge struct {
    domain.BaseEvent
    WorkflowID           string `json:"workflow_id"`
    EdgeID               string `json:"edge_id"`
    FromNodeDefinitionID string `json:"from_node_definition_id"`
    ToNodeDefinitionID   string `json:"to_node_definition_id"`
}

func NewAddEdge(
    idFactory id.Factory,
    workflowID, edgeID, fromNodeDefinitionID, toNodeDefinitionID string,
) *AddEdge {
    return &AddEdge{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "AddEdge",
        ),
        WorkflowID:           workflowID,
        EdgeID:               edgeID,
        FromNodeDefinitionID: fromNodeDefinitionID,
        ToNodeDefinitionID:   toNodeDefinitionID,
    }
}
```

**File:** `internal/domain/workflow/event/remove_edge.go`

**Action:** CREATE

**Pseudocode:**

```go
package event

import (
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

type RemoveEdge struct {
    domain.BaseEvent
    WorkflowID string `json:"workflow_id"`
    EdgeID     string `json:"edge_id"`
}

func NewRemoveEdge(idFactory id.Factory, workflowID, edgeID string) *RemoveEdge {
    return &RemoveEdge{
        BaseEvent: domain.NewBaseEvent(
            idFactory.New(),
            workflowID,
            "Workflow",
            "RemoveEdge",
        ),
        WorkflowID: workflowID,
        EdgeID:     edgeID,
    }
}
```

**Dependencies:**
- `use-open-workflow.io/engine/pkg/domain`
- `use-open-workflow.io/engine/pkg/id`

**Tests Required:**
- Test each event constructor creates event with correct fields

---

### Step 5: Create Workflow Aggregate

**File:** `internal/domain/workflow/aggregate/workflow.go`

**Action:** CREATE

**Rationale:** Workflow aggregate owns NodeDefinition and Edge collections and enforces business rules.

**Pseudocode:**

```go
package aggregate

import (
    "errors"
    "time"

    "use-open-workflow.io/engine/internal/domain/workflow/entity"
    "use-open-workflow.io/engine/internal/domain/workflow/event"
    "use-open-workflow.io/engine/pkg/domain"
    "use-open-workflow.io/engine/pkg/id"
)

var (
    ErrNodeDefinitionNotFound = errors.New("node definition not found")
    ErrEdgeNotFound           = errors.New("edge not found")
    ErrDuplicateEdge          = errors.New("edge already exists")
    ErrSelfLoop               = errors.New("edge cannot connect a node to itself")
)

// Workflow is the aggregate root containing NodeDefinitions and Edges
type Workflow struct {
    domain.BaseAggregate
    Name            string
    Description     string
    NodeDefinitions []*entity.NodeDefinition
    Edges           []*entity.Edge
}

// newWorkflow creates a new Workflow (internal constructor)
func newWorkflow(
    idFactory id.Factory,
    aggregateID string,
    name string,
    description string,
) *Workflow {
    workflow := &Workflow{
        BaseAggregate:   domain.NewBaseAggregate(aggregateID),
        Name:            name,
        Description:     description,
        NodeDefinitions: make([]*entity.NodeDefinition, 0),
        Edges:           make([]*entity.Edge, 0),
    }
    workflow.AddEvent(event.NewCreateWorkflow(idFactory, aggregateID, name, description))
    return workflow
}

// ReconstituteWorkflow recreates aggregate from database
// Note: NodeDefinitions and Edges should be added separately after reconstruction
func ReconstituteWorkflow(
    aggregateID string,
    name string,
    description string,
    createdAt time.Time,
    updatedAt time.Time,
) *Workflow {
    return &Workflow{
        BaseAggregate:   domain.ReconstituteBaseAggregate(aggregateID, createdAt, updatedAt),
        Name:            name,
        Description:     description,
        NodeDefinitions: make([]*entity.NodeDefinition, 0),
        Edges:           make([]*entity.Edge, 0),
    }
}

// UpdateName updates workflow name
func (w *Workflow) UpdateName(idFactory id.Factory, name string) {
    w.Name = name
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewUpdateWorkflow(idFactory, w.ID, w.Name, w.Description))
}

// UpdateDescription updates workflow description
func (w *Workflow) UpdateDescription(idFactory id.Factory, description string) {
    w.Description = description
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewUpdateWorkflow(idFactory, w.ID, w.Name, w.Description))
}

// AddNodeDefinition adds a new NodeDefinition to the workflow
// Returns the created NodeDefinition
func (w *Workflow) AddNodeDefinition(
    idFactory id.Factory,
    nodeTemplateID string,
    name string,
    config map[string]interface{},
    positionX float64,
    positionY float64,
) *entity.NodeDefinition {
    nodeDefID := idFactory.New()
    nodeDef := entity.NewNodeDefinition(
        nodeDefID,
        w.ID,
        nodeTemplateID,
        name,
        config,
        positionX,
        positionY,
    )
    w.NodeDefinitions = append(w.NodeDefinitions, nodeDef)
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewAddNodeDefinition(idFactory, w.ID, nodeDefID, nodeTemplateID, name))
    return nodeDef
}

// RemoveNodeDefinition removes a NodeDefinition by ID
// Also removes any edges connected to this node
func (w *Workflow) RemoveNodeDefinition(idFactory id.Factory, nodeDefID string) error {
    found := false
    newNodeDefs := make([]*entity.NodeDefinition, 0, len(w.NodeDefinitions))
    for _, nd := range w.NodeDefinitions {
        if nd.ID == nodeDefID {
            found = true
        } else {
            newNodeDefs = append(newNodeDefs, nd)
        }
    }
    if !found {
        return ErrNodeDefinitionNotFound
    }
    w.NodeDefinitions = newNodeDefs

    // Remove edges connected to this node
    newEdges := make([]*entity.Edge, 0, len(w.Edges))
    for _, e := range w.Edges {
        if e.FromNodeDefinitionID != nodeDefID && e.ToNodeDefinitionID != nodeDefID {
            newEdges = append(newEdges, e)
        }
    }
    w.Edges = newEdges

    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewRemoveNodeDefinition(idFactory, w.ID, nodeDefID))
    return nil
}

// GetNodeDefinition returns a NodeDefinition by ID
func (w *Workflow) GetNodeDefinition(nodeDefID string) *entity.NodeDefinition {
    for _, nd := range w.NodeDefinitions {
        if nd.ID == nodeDefID {
            return nd
        }
    }
    return nil
}

// AddEdge adds a new Edge connecting two NodeDefinitions
// Validates that both nodes exist and are different
func (w *Workflow) AddEdge(
    idFactory id.Factory,
    fromNodeDefID string,
    toNodeDefID string,
) (*entity.Edge, error) {
    // Validate no self-loop
    if fromNodeDefID == toNodeDefID {
        return nil, ErrSelfLoop
    }

    // Validate both nodes exist
    fromExists := false
    toExists := false
    for _, nd := range w.NodeDefinitions {
        if nd.ID == fromNodeDefID {
            fromExists = true
        }
        if nd.ID == toNodeDefID {
            toExists = true
        }
    }
    if !fromExists || !toExists {
        return nil, ErrNodeDefinitionNotFound
    }

    // Check for duplicate edge
    for _, e := range w.Edges {
        if e.FromNodeDefinitionID == fromNodeDefID && e.ToNodeDefinitionID == toNodeDefID {
            return nil, ErrDuplicateEdge
        }
    }

    edgeID := idFactory.New()
    edge := entity.NewEdge(edgeID, w.ID, fromNodeDefID, toNodeDefID)
    w.Edges = append(w.Edges, edge)
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewAddEdge(idFactory, w.ID, edgeID, fromNodeDefID, toNodeDefID))
    return edge, nil
}

// RemoveEdge removes an Edge by ID
func (w *Workflow) RemoveEdge(idFactory id.Factory, edgeID string) error {
    found := false
    newEdges := make([]*entity.Edge, 0, len(w.Edges))
    for _, e := range w.Edges {
        if e.ID == edgeID {
            found = true
        } else {
            newEdges = append(newEdges, e)
        }
    }
    if !found {
        return ErrEdgeNotFound
    }
    w.Edges = newEdges
    w.SetUpdatedAt(time.Now().UTC())
    w.AddEvent(event.NewRemoveEdge(idFactory, w.ID, edgeID))
    return nil
}

// GetEdge returns an Edge by ID
func (w *Workflow) GetEdge(edgeID string) *entity.Edge {
    for _, e := range w.Edges {
        if e.ID == edgeID {
            return e
        }
    }
    return nil
}

// SetNodeDefinitions sets the NodeDefinitions (used during reconstitution)
func (w *Workflow) SetNodeDefinitions(nodeDefs []*entity.NodeDefinition) {
    w.NodeDefinitions = nodeDefs
}

// SetEdges sets the Edges (used during reconstitution)
func (w *Workflow) SetEdges(edges []*entity.Edge) {
    w.Edges = edges
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/entity`
- `use-open-workflow.io/engine/internal/domain/workflow/event`
- `use-open-workflow.io/engine/pkg/domain`
- `use-open-workflow.io/engine/pkg/id`

**Tests Required:**
- Test newWorkflow creates aggregate with event
- Test ReconstituteWorkflow recreates aggregate without events
- Test UpdateName updates and emits event
- Test AddNodeDefinition adds node and emits event
- Test RemoveNodeDefinition removes node and connected edges
- Test AddEdge validates nodes exist and no self-loop
- Test AddEdge rejects duplicate edges
- Test RemoveEdge removes edge and emits event

---

### Step 6: Create Workflow Factory

**File:** `internal/domain/workflow/aggregate/workflow_factory.go`

**Action:** CREATE

**Rationale:** Factory pattern for creating Workflow aggregates.

**Pseudocode:**

```go
package aggregate

import (
    "use-open-workflow.io/engine/pkg/id"
)

type WorkflowFactory struct {
    idFactory id.Factory
}

func NewWorkflowFactory(idFactory id.Factory) *WorkflowFactory {
    return &WorkflowFactory{
        idFactory: idFactory,
    }
}

func (f *WorkflowFactory) Make(name string, description string) *Workflow {
    return newWorkflow(f.idFactory, f.idFactory.New(), name, description)
}
```

**Dependencies:**
- `use-open-workflow.io/engine/pkg/id`

**Tests Required:**
- Test Make creates workflow with correct name and description

---

### Step 7: Create Inbound Port DTOs

**File:** `internal/port/workflow/inbound/workflow_dto.go`

**Action:** CREATE

**Rationale:** DTOs for API layer communication.

**Pseudocode:**

```go
package inbound

import "time"

// WorkflowDTO represents a workflow for API responses
type WorkflowDTO struct {
    ID              string               `json:"id"`
    Name            string               `json:"name"`
    Description     string               `json:"description"`
    NodeDefinitions []*NodeDefinitionDTO `json:"nodeDefinitions"`
    Edges           []*EdgeDTO           `json:"edges"`
    CreatedAt       time.Time            `json:"createdAt"`
    UpdatedAt       time.Time            `json:"updatedAt"`
}

// NodeDefinitionDTO represents a node definition for API responses
type NodeDefinitionDTO struct {
    ID             string                 `json:"id"`
    WorkflowID     string                 `json:"workflowId"`
    NodeTemplateID string                 `json:"nodeTemplateId"`
    Name           string                 `json:"name"`
    Config         map[string]interface{} `json:"config,omitempty"`
    PositionX      float64                `json:"positionX"`
    PositionY      float64                `json:"positionY"`
    CreatedAt      time.Time              `json:"createdAt"`
    UpdatedAt      time.Time              `json:"updatedAt"`
}

// EdgeDTO represents an edge for API responses
type EdgeDTO struct {
    ID                   string    `json:"id"`
    WorkflowID           string    `json:"workflowId"`
    FromNodeDefinitionID string    `json:"fromNodeDefinitionId"`
    ToNodeDefinitionID   string    `json:"toNodeDefinitionId"`
    CreatedAt            time.Time `json:"createdAt"`
}

// Input structs for write operations
type CreateWorkflowInput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type UpdateWorkflowInput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type AddNodeDefinitionInput struct {
    NodeTemplateID string                 `json:"nodeTemplateId"`
    Name           string                 `json:"name"`
    Config         map[string]interface{} `json:"config,omitempty"`
    PositionX      float64                `json:"positionX"`
    PositionY      float64                `json:"positionY"`
}

type UpdateNodeDefinitionInput struct {
    Name      string                 `json:"name,omitempty"`
    Config    map[string]interface{} `json:"config,omitempty"`
    PositionX *float64               `json:"positionX,omitempty"`
    PositionY *float64               `json:"positionY,omitempty"`
}

type AddEdgeInput struct {
    FromNodeDefinitionID string `json:"fromNodeDefinitionId"`
    ToNodeDefinitionID   string `json:"toNodeDefinitionId"`
}
```

**Dependencies:** None

**Tests Required:**
- None (pure data structures)

---

### Step 8: Create Inbound Port Service Interfaces

**File:** `internal/port/workflow/inbound/workflow_read_service.go`

**Action:** CREATE

**Rationale:** Define read service interface.

**Pseudocode:**

```go
package inbound

import "context"

type WorkflowReadService interface {
    List(ctx context.Context) ([]*WorkflowDTO, error)
    GetByID(ctx context.Context, id string) (*WorkflowDTO, error)
}
```

**File:** `internal/port/workflow/inbound/workflow_write_service.go`

**Action:** CREATE

**Pseudocode:**

```go
package inbound

import "context"

type WorkflowWriteService interface {
    // Workflow CRUD
    Create(ctx context.Context, input CreateWorkflowInput) (*WorkflowDTO, error)
    Update(ctx context.Context, id string, input UpdateWorkflowInput) (*WorkflowDTO, error)
    Delete(ctx context.Context, id string) error

    // NodeDefinition operations (nested under Workflow)
    AddNodeDefinition(ctx context.Context, workflowID string, input AddNodeDefinitionInput) (*NodeDefinitionDTO, error)
    UpdateNodeDefinition(ctx context.Context, workflowID string, nodeDefID string, input UpdateNodeDefinitionInput) (*NodeDefinitionDTO, error)
    RemoveNodeDefinition(ctx context.Context, workflowID string, nodeDefID string) error

    // Edge operations (nested under Workflow)
    AddEdge(ctx context.Context, workflowID string, input AddEdgeInput) (*EdgeDTO, error)
    RemoveEdge(ctx context.Context, workflowID string, edgeID string) error
}
```

**Dependencies:** None

**Tests Required:**
- None (interfaces)

---

### Step 9: Create Inbound Port Mapper Interface

**File:** `internal/port/workflow/inbound/workflow_mapper.go`

**Action:** CREATE

**Rationale:** Define mapper interface for aggregate to DTO conversion.

**Pseudocode:**

```go
package inbound

import (
    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/domain/workflow/entity"
)

type WorkflowMapper interface {
    ToWorkflowDTO(workflow *aggregate.Workflow) (*WorkflowDTO, error)
    ToNodeDefinitionDTO(nodeDef *entity.NodeDefinition) (*NodeDefinitionDTO, error)
    ToEdgeDTO(edge *entity.Edge) (*EdgeDTO, error)
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/domain/workflow/entity`

**Tests Required:**
- None (interface)

---

### Step 10: Create Outbound Port Models

**File:** `internal/port/workflow/outbound/workflow_model.go`

**Action:** CREATE

**Rationale:** Database models for persistence layer.

**Pseudocode:**

```go
package outbound

import "time"

// WorkflowModel represents workflow in database
type WorkflowModel struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func NewWorkflowModel() *WorkflowModel {
    return &WorkflowModel{}
}

// NodeDefinitionModel represents node_definition in database
type NodeDefinitionModel struct {
    ID             string
    WorkflowID     string
    NodeTemplateID string
    Name           string
    Config         map[string]interface{}
    PositionX      float64
    PositionY      float64
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

func NewNodeDefinitionModel() *NodeDefinitionModel {
    return &NodeDefinitionModel{}
}

// EdgeModel represents edge in database
type EdgeModel struct {
    ID                   string
    WorkflowID           string
    FromNodeDefinitionID string
    ToNodeDefinitionID   string
    CreatedAt            time.Time
}

func NewEdgeModel() *EdgeModel {
    return &EdgeModel{}
}
```

**Dependencies:** None

**Tests Required:**
- None (pure data structures)

---

### Step 11: Create Outbound Port Repository Interfaces

**File:** `internal/port/workflow/outbound/workflow_read_repository.go`

**Action:** CREATE

**Rationale:** Define read repository interface.

**Pseudocode:**

```go
package outbound

import (
    "context"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
)

type WorkflowReadRepository interface {
    // FindMany returns all workflows (without nested entities for list view)
    FindMany(ctx context.Context) ([]*aggregate.Workflow, error)

    // FindByID returns workflow with all nested NodeDefinitions and Edges
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
    "use-open-workflow.io/engine/internal/domain/workflow/entity"
)

type WorkflowWriteRepository interface {
    // Workflow operations
    Save(ctx context.Context, workflow *aggregate.Workflow) error
    Update(ctx context.Context, workflow *aggregate.Workflow) error
    Delete(ctx context.Context, id string) error

    // NodeDefinition operations
    SaveNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error
    UpdateNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error
    DeleteNodeDefinition(ctx context.Context, id string) error

    // Edge operations
    SaveEdge(ctx context.Context, edge *entity.Edge) error
    DeleteEdge(ctx context.Context, id string) error
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/domain/workflow/entity`

**Tests Required:**
- None (interfaces)

---

### Step 12: Create Outbound Port Repository Factory Interfaces

**File:** `internal/port/workflow/outbound/workflow_read_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type WorkflowReadRepositoryFactory interface {
    Create(uow outbound.UnitOfWork) WorkflowReadRepository
}
```

**File:** `internal/port/workflow/outbound/workflow_write_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type WorkflowWriteRepositoryFactory interface {
    Create(uow outbound.UnitOfWork) WorkflowWriteRepository
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/port/outbound`

**Tests Required:**
- None (interfaces)

---

### Step 13: Create Outbound Adapter - PostgreSQL Read Repository

**File:** `internal/adapter/workflow/outbound/workflow_postgres_read_repository.go`

**Action:** CREATE

**Rationale:** PostgreSQL implementation of read repository.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/domain/workflow/entity"
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowPostgresReadRepository struct {
    uow portOutbound.UnitOfWork
}

func NewWorkflowPostgresReadRepository(
    uow portOutbound.UnitOfWork,
) *WorkflowPostgresReadRepository {
    return &WorkflowPostgresReadRepository{
        uow: uow,
    }
}

func (r *WorkflowPostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.Workflow, error) {
    q := r.uow.Querier(ctx)

    rows, err := q.Query(ctx, `
        SELECT id, name, description, created_at, updated_at
        FROM workflow
        ORDER BY created_at DESC
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to query workflows: %w", err)
    }
    defer rows.Close()

    var workflows []*aggregate.Workflow
    for rows.Next() {
        var id, name, description string
        var createdAt, updatedAt time.Time
        if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan workflow: %w", err)
        }

        workflow := aggregate.ReconstituteWorkflow(id, name, description, createdAt, updatedAt)
        workflows = append(workflows, workflow)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }

    return workflows, nil
}

func (r *WorkflowPostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.Workflow, error) {
    q := r.uow.Querier(ctx)

    // 1. Fetch workflow
    var name, description string
    var createdAt, updatedAt time.Time
    err := q.QueryRow(ctx, `
        SELECT id, name, description, created_at, updated_at
        FROM workflow
        WHERE id = $1
    `, id).Scan(&id, &name, &description, &createdAt, &updatedAt)

    if err != nil && err.Error() == "no rows in result set" {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query workflow: %w", err)
    }

    workflow := aggregate.ReconstituteWorkflow(id, name, description, createdAt, updatedAt)

    // 2. Fetch NodeDefinitions
    nodeRows, err := q.Query(ctx, `
        SELECT id, workflow_id, node_template_id, name, config, position_x, position_y, created_at, updated_at
        FROM node_definition
        WHERE workflow_id = $1
        ORDER BY created_at ASC
    `, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query node definitions: %w", err)
    }
    defer nodeRows.Close()

    var nodeDefs []*entity.NodeDefinition
    for nodeRows.Next() {
        var ndID, wfID, ntID, ndName string
        var configJSON []byte
        var posX, posY float64
        var ndCreatedAt, ndUpdatedAt time.Time

        if err := nodeRows.Scan(&ndID, &wfID, &ntID, &ndName, &configJSON, &posX, &posY, &ndCreatedAt, &ndUpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan node definition: %w", err)
        }

        var config map[string]interface{}
        if configJSON != nil {
            if err := json.Unmarshal(configJSON, &config); err != nil {
                return nil, fmt.Errorf("failed to unmarshal config: %w", err)
            }
        }

        nodeDef := entity.ReconstituteNodeDefinition(ndID, wfID, ntID, ndName, config, posX, posY, ndCreatedAt, ndUpdatedAt)
        nodeDefs = append(nodeDefs, nodeDef)
    }
    if err := nodeRows.Err(); err != nil {
        return nil, fmt.Errorf("node definition row iteration error: %w", err)
    }
    workflow.SetNodeDefinitions(nodeDefs)

    // 3. Fetch Edges
    edgeRows, err := q.Query(ctx, `
        SELECT id, workflow_id, from_node_definition_id, to_node_definition_id, created_at
        FROM edge
        WHERE workflow_id = $1
        ORDER BY created_at ASC
    `, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query edges: %w", err)
    }
    defer edgeRows.Close()

    var edges []*entity.Edge
    for edgeRows.Next() {
        var eID, eWfID, fromID, toID string
        var eCreatedAt time.Time

        if err := edgeRows.Scan(&eID, &eWfID, &fromID, &toID, &eCreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan edge: %w", err)
        }

        edge := entity.ReconstituteEdge(eID, eWfID, fromID, toID, eCreatedAt)
        edges = append(edges, edge)
    }
    if err := edgeRows.Err(); err != nil {
        return nil, fmt.Errorf("edge row iteration error: %w", err)
    }
    workflow.SetEdges(edges)

    return workflow, nil
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/domain/workflow/entity`
- `use-open-workflow.io/engine/internal/port/outbound`

**Tests Required:**
- Test FindMany returns workflows ordered by created_at
- Test FindByID returns nil for non-existent workflow
- Test FindByID returns workflow with NodeDefinitions and Edges

---

### Step 14: Create Outbound Adapter - PostgreSQL Write Repository

**File:** `internal/adapter/workflow/outbound/workflow_postgres_write_repository.go`

**Action:** CREATE

**Rationale:** PostgreSQL implementation of write repository.

**Pseudocode:**

```go
package outbound

import (
    "context"
    "encoding/json"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/domain/workflow/entity"
    portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowPostgresWriteRepository struct {
    uow portOutbound.UnitOfWork
}

func NewWorkflowPostgresWriteRepository(
    uow portOutbound.UnitOfWork,
) *WorkflowPostgresWriteRepository {
    return &WorkflowPostgresWriteRepository{
        uow: uow,
    }
}

func (r *WorkflowPostgresWriteRepository) Save(ctx context.Context, workflow *aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        INSERT INTO workflow (id, name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `, workflow.ID, workflow.Name, workflow.Description, workflow.CreatedAt, workflow.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to save workflow: %w", err)
    }

    r.uow.RegisterNew(workflow)
    return nil
}

func (r *WorkflowPostgresWriteRepository) Update(ctx context.Context, workflow *aggregate.Workflow) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        UPDATE workflow
        SET name = $1, description = $2, updated_at = $3
        WHERE id = $4
    `, workflow.Name, workflow.Description, workflow.UpdatedAt, workflow.ID)

    if err != nil {
        return fmt.Errorf("failed to update workflow: %w", err)
    }

    r.uow.RegisterDirty(workflow)
    return nil
}

func (r *WorkflowPostgresWriteRepository) Delete(ctx context.Context, id string) error {
    q := r.uow.Querier(ctx)

    // CASCADE will handle node_definition and edge deletion
    _, err := q.Exec(ctx, `
        DELETE FROM workflow
        WHERE id = $1
    `, id)

    if err != nil {
        return fmt.Errorf("failed to delete workflow: %w", err)
    }

    return nil
}

func (r *WorkflowPostgresWriteRepository) SaveNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error {
    q := r.uow.Querier(ctx)

    configJSON, err := json.Marshal(nodeDef.Config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    _, err = q.Exec(ctx, `
        INSERT INTO node_definition (id, workflow_id, node_template_id, name, config, position_x, position_y, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `, nodeDef.ID, nodeDef.WorkflowID, nodeDef.NodeTemplateID, nodeDef.Name, configJSON, nodeDef.PositionX, nodeDef.PositionY, nodeDef.CreatedAt, nodeDef.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to save node definition: %w", err)
    }

    return nil
}

func (r *WorkflowPostgresWriteRepository) UpdateNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error {
    q := r.uow.Querier(ctx)

    configJSON, err := json.Marshal(nodeDef.Config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    _, err = q.Exec(ctx, `
        UPDATE node_definition
        SET name = $1, config = $2, position_x = $3, position_y = $4, updated_at = $5
        WHERE id = $6
    `, nodeDef.Name, configJSON, nodeDef.PositionX, nodeDef.PositionY, nodeDef.UpdatedAt, nodeDef.ID)

    if err != nil {
        return fmt.Errorf("failed to update node definition: %w", err)
    }

    return nil
}

func (r *WorkflowPostgresWriteRepository) DeleteNodeDefinition(ctx context.Context, id string) error {
    q := r.uow.Querier(ctx)

    // CASCADE will handle edge deletion
    _, err := q.Exec(ctx, `
        DELETE FROM node_definition
        WHERE id = $1
    `, id)

    if err != nil {
        return fmt.Errorf("failed to delete node definition: %w", err)
    }

    return nil
}

func (r *WorkflowPostgresWriteRepository) SaveEdge(ctx context.Context, edge *entity.Edge) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        INSERT INTO edge (id, workflow_id, from_node_definition_id, to_node_definition_id, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, edge.ID, edge.WorkflowID, edge.FromNodeDefinitionID, edge.ToNodeDefinitionID, edge.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to save edge: %w", err)
    }

    return nil
}

func (r *WorkflowPostgresWriteRepository) DeleteEdge(ctx context.Context, id string) error {
    q := r.uow.Querier(ctx)

    _, err := q.Exec(ctx, `
        DELETE FROM edge
        WHERE id = $1
    `, id)

    if err != nil {
        return fmt.Errorf("failed to delete edge: %w", err)
    }

    return nil
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/domain/workflow/entity`
- `use-open-workflow.io/engine/internal/port/outbound`

**Tests Required:**
- Test Save persists workflow and registers with UoW
- Test Update modifies workflow and registers dirty
- Test Delete removes workflow (cascade deletes children)
- Test SaveNodeDefinition persists node definition with JSON config
- Test SaveEdge persists edge

---

### Step 15: Create Repository Factory Implementations

**File:** `internal/adapter/workflow/outbound/workflow_postgres_read_repository_factory.go`

**Action:** CREATE

**Pseudocode:**

```go
package outbound

import (
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowPostgresReadRepositoryFactory struct{}

func NewWorkflowPostgresReadRepositoryFactory() *WorkflowPostgresReadRepositoryFactory {
    return &WorkflowPostgresReadRepositoryFactory{}
}

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

type WorkflowPostgresWriteRepositoryFactory struct{}

func NewWorkflowPostgresWriteRepositoryFactory() *WorkflowPostgresWriteRepositoryFactory {
    return &WorkflowPostgresWriteRepositoryFactory{}
}

func (f *WorkflowPostgresWriteRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowWriteRepository {
    return NewWorkflowPostgresWriteRepository(uow)
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/port/workflow/outbound`
- `use-open-workflow.io/engine/internal/port/outbound`

**Tests Required:**
- None (simple factory delegation)

---

### Step 16: Create Inbound Adapter - Mapper Implementation

**File:** `internal/adapter/workflow/inbound/workflow_mapper.go`

**Action:** CREATE

**Rationale:** Convert between aggregates/entities and DTOs.

**Pseudocode:**

```go
package inbound

import (
    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/domain/workflow/entity"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

type WorkflowMapper struct{}

func NewWorkflowMapper() *WorkflowMapper {
    return &WorkflowMapper{}
}

func (m *WorkflowMapper) ToWorkflowDTO(workflow *aggregate.Workflow) (*inbound.WorkflowDTO, error) {
    nodeDefDTOs := make([]*inbound.NodeDefinitionDTO, len(workflow.NodeDefinitions))
    for i, nd := range workflow.NodeDefinitions {
        dto, err := m.ToNodeDefinitionDTO(nd)
        if err != nil {
            return nil, err
        }
        nodeDefDTOs[i] = dto
    }

    edgeDTOs := make([]*inbound.EdgeDTO, len(workflow.Edges))
    for i, e := range workflow.Edges {
        dto, err := m.ToEdgeDTO(e)
        if err != nil {
            return nil, err
        }
        edgeDTOs[i] = dto
    }

    return &inbound.WorkflowDTO{
        ID:              workflow.ID,
        Name:            workflow.Name,
        Description:     workflow.Description,
        NodeDefinitions: nodeDefDTOs,
        Edges:           edgeDTOs,
        CreatedAt:       workflow.CreatedAt,
        UpdatedAt:       workflow.UpdatedAt,
    }, nil
}

func (m *WorkflowMapper) ToNodeDefinitionDTO(nodeDef *entity.NodeDefinition) (*inbound.NodeDefinitionDTO, error) {
    return &inbound.NodeDefinitionDTO{
        ID:             nodeDef.ID,
        WorkflowID:     nodeDef.WorkflowID,
        NodeTemplateID: nodeDef.NodeTemplateID,
        Name:           nodeDef.Name,
        Config:         nodeDef.Config,
        PositionX:      nodeDef.PositionX,
        PositionY:      nodeDef.PositionY,
        CreatedAt:      nodeDef.CreatedAt,
        UpdatedAt:      nodeDef.UpdatedAt,
    }, nil
}

func (m *WorkflowMapper) ToEdgeDTO(edge *entity.Edge) (*inbound.EdgeDTO, error) {
    return &inbound.EdgeDTO{
        ID:                   edge.ID,
        WorkflowID:           edge.WorkflowID,
        FromNodeDefinitionID: edge.FromNodeDefinitionID,
        ToNodeDefinitionID:   edge.ToNodeDefinitionID,
        CreatedAt:            edge.CreatedAt,
    }, nil
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/domain/workflow/entity`
- `use-open-workflow.io/engine/internal/port/workflow/inbound`

**Tests Required:**
- Test ToWorkflowDTO converts all fields including nested entities
- Test ToNodeDefinitionDTO converts all fields
- Test ToEdgeDTO converts all fields

---

### Step 17: Create Inbound Adapter - Read Service Implementation

**File:** `internal/adapter/workflow/inbound/workflow_read_service.go`

**Action:** CREATE

**Rationale:** Implement read operations following existing patterns.

**Pseudocode:**

```go
package inbound

import (
    "context"

    "use-open-workflow.io/engine/internal/port/workflow/inbound"
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowReadService struct {
    uowFactory            outbound.UnitOfWorkFactory
    readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory
    mapper                inbound.WorkflowMapper
}

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

func (s *WorkflowReadService) List(ctx context.Context) ([]*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    readRepo := s.readRepositoryFactory.Create(uow)

    workflows, err := readRepo.FindMany(ctx)
    if err != nil {
        return nil, err
    }

    workflowDTOs := make([]*inbound.WorkflowDTO, len(workflows))
    for i, w := range workflows {
        dto, err := s.mapper.ToWorkflowDTO(w)
        if err != nil {
            return nil, err
        }
        workflowDTOs[i] = dto
    }

    return workflowDTOs, nil
}

func (s *WorkflowReadService) GetByID(ctx context.Context, id string) (*inbound.WorkflowDTO, error) {
    uow := s.uowFactory.Create()
    readRepo := s.readRepositoryFactory.Create(uow)

    workflow, err := readRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if workflow == nil {
        return nil, nil
    }

    return s.mapper.ToWorkflowDTO(workflow)
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/port/workflow/inbound`
- `use-open-workflow.io/engine/internal/port/workflow/outbound`
- `use-open-workflow.io/engine/internal/port/outbound`

**Tests Required:**
- Test List returns all workflows as DTOs
- Test GetByID returns workflow with nested entities
- Test GetByID returns nil for non-existent workflow

---

### Step 18: Create Inbound Adapter - Write Service Implementation

**File:** `internal/adapter/workflow/inbound/workflow_write_service.go`

**Action:** CREATE

**Rationale:** Implement write operations with UoW transaction pattern.

**Pseudocode:**

```go
package inbound

import (
    "context"
    "fmt"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
    workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
    "use-open-workflow.io/engine/internal/port/outbound"
    "use-open-workflow.io/engine/pkg/id"
)

type WorkflowWriteService struct {
    uowFactory             outbound.UnitOfWorkFactory
    writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory
    readRepositoryFactory  workflowOutbound.WorkflowReadRepositoryFactory
    factory                *aggregate.WorkflowFactory
    mapper                 inbound.WorkflowMapper
    idFactory              id.Factory
}

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

    workflow := s.factory.Make(input.Name, input.Description)

    if err = writeRepo.Save(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to save workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.ToWorkflowDTO(workflow)
}

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

    workflow, err := readRepo.FindByID(txCtx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", id)
    }

    workflow.UpdateName(s.idFactory, input.Name)
    workflow.UpdateDescription(s.idFactory, input.Description)

    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.ToWorkflowDTO(workflow)
}

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

func (s *WorkflowWriteService) AddNodeDefinition(ctx context.Context, workflowID string, input inbound.AddNodeDefinitionInput) (*inbound.NodeDefinitionDTO, error) {
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

    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    nodeDef := workflow.AddNodeDefinition(
        s.idFactory,
        input.NodeTemplateID,
        input.Name,
        input.Config,
        input.PositionX,
        input.PositionY,
    )

    if err = writeRepo.SaveNodeDefinition(txCtx, nodeDef); err != nil {
        return nil, fmt.Errorf("failed to save node definition: %w", err)
    }

    // Update workflow's updated_at
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.ToNodeDefinitionDTO(nodeDef)
}

func (s *WorkflowWriteService) UpdateNodeDefinition(ctx context.Context, workflowID string, nodeDefID string, input inbound.UpdateNodeDefinitionInput) (*inbound.NodeDefinitionDTO, error) {
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

    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    nodeDef := workflow.GetNodeDefinition(nodeDefID)
    if nodeDef == nil {
        return nil, fmt.Errorf("node definition not found: %s", nodeDefID)
    }

    // Apply updates
    if input.Name != "" {
        nodeDef.UpdateName(input.Name)
    }
    if input.Config != nil {
        nodeDef.UpdateConfig(input.Config)
    }
    if input.PositionX != nil && input.PositionY != nil {
        nodeDef.UpdatePosition(*input.PositionX, *input.PositionY)
    }

    if err = writeRepo.UpdateNodeDefinition(txCtx, nodeDef); err != nil {
        return nil, fmt.Errorf("failed to update node definition: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.ToNodeDefinitionDTO(nodeDef)
}

func (s *WorkflowWriteService) RemoveNodeDefinition(ctx context.Context, workflowID string, nodeDefID string) error {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return fmt.Errorf("workflow not found: %s", workflowID)
    }

    if err = workflow.RemoveNodeDefinition(s.idFactory, nodeDefID); err != nil {
        return fmt.Errorf("failed to remove node definition: %w", err)
    }

    // Delete from database (CASCADE handles edges)
    if err = writeRepo.DeleteNodeDefinition(txCtx, nodeDefID); err != nil {
        return fmt.Errorf("failed to delete node definition: %w", err)
    }

    // Update workflow's updated_at
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (s *WorkflowWriteService) AddEdge(ctx context.Context, workflowID string, input inbound.AddEdgeInput) (*inbound.EdgeDTO, error) {
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

    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return nil, fmt.Errorf("workflow not found: %s", workflowID)
    }

    edge, err := workflow.AddEdge(s.idFactory, input.FromNodeDefinitionID, input.ToNodeDefinitionID)
    if err != nil {
        return nil, fmt.Errorf("failed to add edge: %w", err)
    }

    if err = writeRepo.SaveEdge(txCtx, edge); err != nil {
        return nil, fmt.Errorf("failed to save edge: %w", err)
    }

    // Update workflow's updated_at
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return nil, fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return s.mapper.ToEdgeDTO(edge)
}

func (s *WorkflowWriteService) RemoveEdge(ctx context.Context, workflowID string, edgeID string) error {
    uow := s.uowFactory.Create()
    writeRepo := s.writeRepositoryFactory.Create(uow)
    readRepo := s.readRepositoryFactory.Create(uow)

    txCtx, err := uow.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            uow.Rollback(txCtx)
        }
    }()

    workflow, err := readRepo.FindByID(txCtx, workflowID)
    if err != nil {
        return fmt.Errorf("failed to find workflow: %w", err)
    }
    if workflow == nil {
        return fmt.Errorf("workflow not found: %s", workflowID)
    }

    if err = workflow.RemoveEdge(s.idFactory, edgeID); err != nil {
        return fmt.Errorf("failed to remove edge: %w", err)
    }

    if err = writeRepo.DeleteEdge(txCtx, edgeID); err != nil {
        return fmt.Errorf("failed to delete edge: %w", err)
    }

    // Update workflow's updated_at
    if err = writeRepo.Update(txCtx, workflow); err != nil {
        return fmt.Errorf("failed to update workflow: %w", err)
    }

    if err = uow.Commit(txCtx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

**Dependencies:**
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/internal/port/workflow/inbound`
- `use-open-workflow.io/engine/internal/port/workflow/outbound`
- `use-open-workflow.io/engine/internal/port/outbound`
- `use-open-workflow.io/engine/pkg/id`

**Tests Required:**
- Test Create creates workflow and commits
- Test Update modifies workflow
- Test Delete removes workflow
- Test AddNodeDefinition adds node to workflow
- Test RemoveNodeDefinition removes node and connected edges
- Test AddEdge validates nodes and creates edge
- Test AddEdge rejects self-loop
- Test AddEdge rejects duplicate
- Test RemoveEdge removes edge

---

### Step 19: Create HTTP Handler

**File:** `api/workflow/http/workflow_handler.go`

**Action:** CREATE

**Rationale:** HTTP handlers for workflow API endpoints.

**Pseudocode:**

```go
package http

import (
    "github.com/gofiber/fiber/v3"
    "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

type WorkflowHandler struct {
    readService  inbound.WorkflowReadService
    writeService inbound.WorkflowWriteService
}

func NewWorkflowHandler(
    readService inbound.WorkflowReadService,
    writeService inbound.WorkflowWriteService,
) *WorkflowHandler {
    return &WorkflowHandler{
        readService:  readService,
        writeService: writeService,
    }
}

// Workflow CRUD

func (h *WorkflowHandler) List(c fiber.Ctx) error {
    workflows, err := h.readService.List(c.Context())
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    return c.JSON(workflows)
}

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

func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
    id := c.Params("id")
    if err := h.writeService.Delete(c.Context(), id); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.SendStatus(fiber.StatusNoContent)
}

// NodeDefinition operations (nested under Workflow)

func (h *WorkflowHandler) AddNodeDefinition(c fiber.Ctx) error {
    workflowID := c.Params("id")
    var input inbound.AddNodeDefinitionInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    nodeDef, err := h.writeService.AddNodeDefinition(c.Context(), workflowID, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(nodeDef)
}

func (h *WorkflowHandler) UpdateNodeDefinition(c fiber.Ctx) error {
    workflowID := c.Params("id")
    nodeDefID := c.Params("nodeDefId")
    var input inbound.UpdateNodeDefinitionInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    nodeDef, err := h.writeService.UpdateNodeDefinition(c.Context(), workflowID, nodeDefID, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(nodeDef)
}

func (h *WorkflowHandler) RemoveNodeDefinition(c fiber.Ctx) error {
    workflowID := c.Params("id")
    nodeDefID := c.Params("nodeDefId")

    if err := h.writeService.RemoveNodeDefinition(c.Context(), workflowID, nodeDefID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.SendStatus(fiber.StatusNoContent)
}

// Edge operations (nested under Workflow)

func (h *WorkflowHandler) AddEdge(c fiber.Ctx) error {
    workflowID := c.Params("id")
    var input inbound.AddEdgeInput
    if err := c.Bind().JSON(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    edge, err := h.writeService.AddEdge(c.Context(), workflowID, input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(edge)
}

func (h *WorkflowHandler) RemoveEdge(c fiber.Ctx) error {
    workflowID := c.Params("id")
    edgeID := c.Params("edgeId")

    if err := h.writeService.RemoveEdge(c.Context(), workflowID, edgeID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.SendStatus(fiber.StatusNoContent)
}
```

**Dependencies:**
- `github.com/gofiber/fiber/v3`
- `use-open-workflow.io/engine/internal/port/workflow/inbound`

**Tests Required:**
- Test each handler endpoint returns correct status codes
- Test error handling returns proper JSON error responses

---

### Step 20: Modify Router to Register Workflow Routes

**File:** `api/router.go`

**Action:** MODIFY

**Rationale:** Add workflow routes registration.

**Pseudocode:**

```go
// Add import
import (
    // existing imports...
    workflowHttp "use-open-workflow.io/engine/api/workflow/http"
)

// In SetupRouter function, add after registerNodeTemplateRoutes:
func SetupRouter(c *di.Container) *fiber.App {
    // ... existing code ...

    api := app.Group("/api/v1")
    registerNodeTemplateRoutes(api, c)
    registerWorkflowRoutes(api, c)  // ADD THIS LINE

    return app
}

// Add new function:
func registerWorkflowRoutes(router fiber.Router, c *di.Container) {
    workflowHandler := workflowHttp.NewWorkflowHandler(
        c.WorkflowReadService,
        c.WorkflowWriteService,
    )

    workflow := router.Group("/workflow")

    // Workflow CRUD
    workflow.Get("/", workflowHandler.List)
    workflow.Get("/:id", workflowHandler.GetByID)
    workflow.Post("/", workflowHandler.Create)
    workflow.Put("/:id", workflowHandler.Update)
    workflow.Delete("/:id", workflowHandler.Delete)

    // NodeDefinition operations (nested)
    workflow.Post("/:id/node-definition", workflowHandler.AddNodeDefinition)
    workflow.Put("/:id/node-definition/:nodeDefId", workflowHandler.UpdateNodeDefinition)
    workflow.Delete("/:id/node-definition/:nodeDefId", workflowHandler.RemoveNodeDefinition)

    // Edge operations (nested)
    workflow.Post("/:id/edge", workflowHandler.AddEdge)
    workflow.Delete("/:id/edge/:edgeId", workflowHandler.RemoveEdge)
}
```

**Dependencies:**
- `use-open-workflow.io/engine/api/workflow/http`

**Tests Required:**
- Verify routes are registered correctly

---

### Step 21: Modify DI Container

**File:** `di/container.go`

**Action:** MODIFY

**Rationale:** Wire up all workflow dependencies.

**Pseudocode:**

```go
// Add imports:
import (
    // existing imports...
    workflowAggregate "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    workflowAdapterInbound "use-open-workflow.io/engine/internal/adapter/workflow/inbound"
    workflowAdapterOutbound "use-open-workflow.io/engine/internal/adapter/workflow/outbound"
    workflowInbound "use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// Modify Container struct - add fields:
type Container struct {
    Pool                     *pgxpool.Pool
    NodeTemplateReadService  inbound.NodeTemplateReadService
    NodeTemplateWriteService inbound.NodeTemplateWriteService
    WorkflowReadService      workflowInbound.WorkflowReadService   // ADD
    WorkflowWriteService     workflowInbound.WorkflowWriteService  // ADD
    OutboxProcessor          outbound.OutboxProcessor
}

// In NewContainer function, add after NodeTemplate wiring:
func NewContainer(ctx context.Context) (*Container, error) {
    // ... existing code ...

    // === WORKFLOW DOMAIN WIRING ===

    // Mapper
    workflowMapper := workflowAdapterInbound.NewWorkflowMapper()

    // Factory
    workflowFactory := workflowAggregate.NewWorkflowFactory(idFactory)

    // Repository Factories
    workflowReadRepositoryFactory := workflowAdapterOutbound.NewWorkflowPostgresReadRepositoryFactory()
    workflowWriteRepositoryFactory := workflowAdapterOutbound.NewWorkflowPostgresWriteRepositoryFactory()

    // Services
    workflowReadService := workflowAdapterInbound.NewWorkflowReadService(
        uowFactory,
        workflowReadRepositoryFactory,
        workflowMapper,
    )

    workflowWriteService := workflowAdapterInbound.NewWorkflowWriteService(
        uowFactory,
        workflowWriteRepositoryFactory,
        workflowReadRepositoryFactory,
        workflowFactory,
        workflowMapper,
        idFactory,
    )

    // ... existing outbox code ...

    return &Container{
        Pool:                     pool,
        NodeTemplateReadService:  nodeTemplateReadService,
        NodeTemplateWriteService: nodeTemplateWriteService,
        WorkflowReadService:      workflowReadService,   // ADD
        WorkflowWriteService:     workflowWriteService,  // ADD
        OutboxProcessor:          outboxProcessor,
    }, nil
}
```

**Dependencies:**
- All workflow domain, port, and adapter packages

**Tests Required:**
- Verify container creates successfully with all dependencies

---

### Step 22: Create Aggregate Unit Tests

**File:** `internal/domain/workflow/aggregate/workflow_test.go`

**Action:** CREATE

**Rationale:** Unit tests for Workflow aggregate business logic.

**Pseudocode:**

```go
package aggregate_test

import (
    "testing"

    "use-open-workflow.io/engine/internal/domain/workflow/aggregate"
    "use-open-workflow.io/engine/pkg/id"
)

func TestWorkflowFactory_Make(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)

    // Execute
    workflow := factory.Make("Test Workflow", "A test description")

    // Assert
    // - workflow.ID is not empty
    // - workflow.Name == "Test Workflow"
    // - workflow.Description == "A test description"
    // - workflow.NodeDefinitions is empty slice
    // - workflow.Edges is empty slice
    // - workflow.Events() has one CreateWorkflow event
}

func TestWorkflow_AddNodeDefinition(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)
    workflow := factory.Make("Test", "Desc")
    workflow.ClearEvents() // Clear creation event

    // Execute
    nodeDef := workflow.AddNodeDefinition(idFactory, "template-id", "Node 1", nil, 100, 200)

    // Assert
    // - nodeDef.ID is not empty
    // - nodeDef.WorkflowID == workflow.ID
    // - nodeDef.NodeTemplateID == "template-id"
    // - nodeDef.Name == "Node 1"
    // - nodeDef.PositionX == 100, nodeDef.PositionY == 200
    // - len(workflow.NodeDefinitions) == 1
    // - workflow.Events() has one AddNodeDefinition event
}

func TestWorkflow_RemoveNodeDefinition_RemovesConnectedEdges(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)
    workflow := factory.Make("Test", "Desc")

    node1 := workflow.AddNodeDefinition(idFactory, "t1", "Node 1", nil, 0, 0)
    node2 := workflow.AddNodeDefinition(idFactory, "t2", "Node 2", nil, 0, 0)
    workflow.AddEdge(idFactory, node1.ID, node2.ID)
    workflow.ClearEvents()

    // Execute
    err := workflow.RemoveNodeDefinition(idFactory, node1.ID)

    // Assert
    // - err is nil
    // - len(workflow.NodeDefinitions) == 1
    // - len(workflow.Edges) == 0 (edge was removed with node)
}

func TestWorkflow_AddEdge_ValidatesNoSelfLoop(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)
    workflow := factory.Make("Test", "Desc")
    node := workflow.AddNodeDefinition(idFactory, "t1", "Node 1", nil, 0, 0)

    // Execute
    _, err := workflow.AddEdge(idFactory, node.ID, node.ID)

    // Assert
    // - err == aggregate.ErrSelfLoop
}

func TestWorkflow_AddEdge_ValidatesNodesExist(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)
    workflow := factory.Make("Test", "Desc")
    node := workflow.AddNodeDefinition(idFactory, "t1", "Node 1", nil, 0, 0)

    // Execute
    _, err := workflow.AddEdge(idFactory, node.ID, "non-existent-id")

    // Assert
    // - err == aggregate.ErrNodeDefinitionNotFound
}

func TestWorkflow_AddEdge_RejectsDuplicate(t *testing.T) {
    // Setup
    idFactory := id.NewULIDFactory()
    factory := aggregate.NewWorkflowFactory(idFactory)
    workflow := factory.Make("Test", "Desc")
    node1 := workflow.AddNodeDefinition(idFactory, "t1", "Node 1", nil, 0, 0)
    node2 := workflow.AddNodeDefinition(idFactory, "t2", "Node 2", nil, 0, 0)
    workflow.AddEdge(idFactory, node1.ID, node2.ID)

    // Execute
    _, err := workflow.AddEdge(idFactory, node1.ID, node2.ID)

    // Assert
    // - err == aggregate.ErrDuplicateEdge
}
```

**Dependencies:**
- `testing`
- `use-open-workflow.io/engine/internal/domain/workflow/aggregate`
- `use-open-workflow.io/engine/pkg/id`

**Tests Required:**
- All test cases listed above

---

## 4. Data Changes

**Schema/Model Updates:**

```sql
-- workflow table
CREATE TABLE workflow (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- node_definition table
CREATE TABLE node_definition (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    node_template_id VARCHAR(26) NOT NULL REFERENCES node_template(id),
    name VARCHAR(255) NOT NULL,
    config JSONB,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- edge table
CREATE TABLE edge (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    from_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    to_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_edge UNIQUE (workflow_id, from_node_definition_id, to_node_definition_id),
    CONSTRAINT no_self_loop CHECK (from_node_definition_id != to_node_definition_id)
);
```

**Migration Notes:**

- Migration file `002_workflow_schema.sql` should be run after `001_initial_schema.sql`
- Foreign key to `node_template` requires the table to exist
- CASCADE delete ensures referential integrity when deleting workflows

## 5. Integration Points

- **NodeTemplate Domain:**
  - **Interaction:** NodeDefinition references NodeTemplate by ID
  - **Error Handling:** Foreign key constraint in database ensures NodeTemplate exists

- **Outbox Pattern:**
  - **Interaction:** Domain events (CreateWorkflow, AddNodeDefinition, etc.) are registered with UoW and persisted to outbox
  - **Error Handling:** Outbox processor handles retry on failure

## 6. Edge Cases & Error Handling

| Scenario                          | Handling                                              |
| --------------------------------- | ----------------------------------------------------- |
| Create workflow with empty name   | Database constraint or service validation (optional)  |
| Add edge with non-existent nodes  | Aggregate returns ErrNodeDefinitionNotFound           |
| Add edge creating self-loop       | Aggregate returns ErrSelfLoop                         |
| Add duplicate edge                | Aggregate returns ErrDuplicateEdge                    |
| Remove non-existent node          | Aggregate returns ErrNodeDefinitionNotFound           |
| Remove non-existent edge          | Aggregate returns ErrEdgeNotFound                     |
| Delete workflow with nodes/edges  | CASCADE delete removes all children automatically     |
| Invalid NodeTemplateID            | Foreign key constraint rejects save                   |
| Concurrent modification           | Transaction isolation handles conflicts               |

## 7. Testing Strategy

**Unit Tests:**

- Workflow aggregate creation and reconstitution
- AddNodeDefinition creates entity and emits event
- RemoveNodeDefinition removes node and connected edges
- AddEdge validates constraints (self-loop, duplicate, nodes exist)
- RemoveEdge removes edge and emits event
- Entity creation and update methods

**Integration Tests:**

- Repository FindMany and FindByID with nested entities
- Write repository Save/Update/Delete operations
- Transaction rollback on error
- Foreign key constraint enforcement

**Manual Verification:**

1. Run migration against database
2. Create a workflow via API
3. Add multiple node definitions
4. Add edges between nodes
5. Verify GET returns complete workflow with nested entities
6. Delete a node and verify connected edges are removed
7. Delete workflow and verify cascade deletion

## 8. Implementation Order

Recommended sequence for implementation:

1. **Step 1: Database Migration** — Schema must exist before any code runs
2. **Steps 2-3: Entities** — NodeDefinition and Edge are dependencies of Workflow
3. **Step 4: Domain Events** — Events are used by aggregate methods
4. **Steps 5-6: Aggregate and Factory** — Core domain logic
5. **Steps 7-9: Inbound Ports** — DTOs, service interfaces, mapper interface
6. **Steps 10-12: Outbound Ports** — Models, repository interfaces, factory interfaces
7. **Steps 13-15: Outbound Adapters** — PostgreSQL repositories and factories
8. **Steps 16-18: Inbound Adapters** — Mapper, read service, write service
9. **Step 19: HTTP Handler** — API layer
10. **Steps 20-21: Router and DI** — Wire everything together
11. **Step 22: Unit Tests** — Verify aggregate business logic

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

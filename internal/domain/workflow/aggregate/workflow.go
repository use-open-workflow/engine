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
	ErrEdgeNotFound           = errors.New("edge not found")
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
	if nodeDefinitions == nil {
		nodeDefinitions = make([]*NodeDefinition, 0)
	}
	if edges == nil {
		edges = make([]*Edge, 0)
	}
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
		return ErrEdgeNotFound
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

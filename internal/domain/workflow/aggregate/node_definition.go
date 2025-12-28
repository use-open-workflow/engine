package aggregate

import "use-open-workflow.io/engine/pkg/domain"

// NodeDefinition represents an instance of a NodeTemplate within a Workflow.
// It is a child entity owned by the Workflow aggregate.
type NodeDefinition struct {
	domain.BaseEntity
	WorkflowID     string
	NodeTemplateID string
	Name           string
	PositionX      float64
	PositionY      float64
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

// UpdatePosition updates the node's canvas position.
func (n *NodeDefinition) UpdatePosition(x, y float64) {
	n.PositionX = x
	n.PositionY = y
}

// UpdateName updates the node's display name.
func (n *NodeDefinition) UpdateName(name string) {
	n.Name = name
}

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

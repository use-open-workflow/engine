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

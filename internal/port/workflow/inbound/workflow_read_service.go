package inbound

import "context"

// WorkflowReadService defines read operations for workflows.
type WorkflowReadService interface {
	// List returns all workflows with their child entities.
	List(ctx context.Context) ([]*WorkflowDTO, error)

	// GetByID returns a workflow with all its node definitions and edges.
	GetByID(ctx context.Context, id string) (*WorkflowDTO, error)
}

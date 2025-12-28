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

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

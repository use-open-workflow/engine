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

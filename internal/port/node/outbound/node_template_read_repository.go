package outbound

import (
	"context"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplateReadRepository interface {
	FindMany(ctx context.Context) ([]*aggregate.NodeTemplate, error)
	FindByID(ctx context.Context, id string) (*aggregate.NodeTemplate, error)
}

package outbound

import (
	"context"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplateWriteRepository interface {
	Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error
	Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error
	Delete(ctx context.Context, id string) error
}

package outbound

import (
	"context"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplateStaticReadRepository struct{}

func NewNodeTemplateStaticReadRepository() *NodeTemplateStaticReadRepository {
	return &NodeTemplateStaticReadRepository{}
}

func (s *NodeTemplateStaticReadRepository) FindMany(ctx context.Context) ([]*aggregate.NodeTemplate, error) {
	return nil, nil
}

func (s *NodeTemplateStaticReadRepository) FindByID(ctx context.Context, id string) (*aggregate.NodeTemplate, error) {
	return nil, nil
}

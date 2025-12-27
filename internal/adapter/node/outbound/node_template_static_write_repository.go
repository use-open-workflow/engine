package outbound

import (
	"context"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplateStaticWriteRepository struct{}

func NewNodeTemplateStaticWriteRepository() *NodeTemplateStaticWriteRepository {
	return &NodeTemplateStaticWriteRepository{}
}

func (s *NodeTemplateStaticWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	return nil
}

func (s *NodeTemplateStaticWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	return nil
}

func (s *NodeTemplateStaticWriteRepository) Delete(ctx context.Context, id string) error {
	return nil
}

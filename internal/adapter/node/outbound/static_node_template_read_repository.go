package outbound

import (
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/outbound"
)

type StaticNodeTemplateReadRepository struct {
	nodeTemplateMapper outbound.NodeTemplateMapper
}

func NewStaticNodeTemplateReadRepository(
	nodeTemplateMapper outbound.NodeTemplateMapper,
) *StaticNodeTemplateReadRepository {
	return &StaticNodeTemplateReadRepository{
		nodeTemplateMapper: nodeTemplateMapper,
	}
}

func (s *StaticNodeTemplateReadRepository) FindMany() ([]*aggregate.NodeTemplate, error) {
	return nil, nil
}

func (s *StaticNodeTemplateReadRepository) FindByID(id string) (*aggregate.NodeTemplate, error) {
	return nil, nil
}

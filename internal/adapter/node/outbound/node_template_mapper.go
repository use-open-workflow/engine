package outbound

import (
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/outbound"
	"use-open-workflow.io/engine/pkg/domain"
)

type NodeTemplateMapper struct{}

func NewNodeTemplateMapper() *NodeTemplateMapper {
	return &NodeTemplateMapper{}
}

func (*NodeTemplateMapper) From(in *outbound.NodeTemplateModel) (*aggregate.NodeTemplate, error) {
	ret := &aggregate.NodeTemplate{
		BaseAggregate: domain.BaseAggregate{
			ID: in.ID,
		},
		Name: in.Name,
	}
	return ret, nil
}

func (*NodeTemplateMapper) To(in *aggregate.NodeTemplate) (*outbound.NodeTemplateModel, error) {
	ret := &outbound.NodeTemplateModel{
		ID:   in.ID,
		Name: in.Name,
	}
	return ret, nil
}

package outbound

import (
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/outbound"
)

type NodeTemplateMapper struct{}

func NewNodeTemplateMapper() *NodeTemplateMapper {
	return &NodeTemplateMapper{}
}

func (*NodeTemplateMapper) From(in *outbound.NodeTemplateModel) (*aggregate.NodeTemplate, error) {
	return aggregate.ReconstituteNodeTemplate(
		in.ID,
		in.Name,
		in.CreatedAt,
		in.UpdatedAt,
	), nil
}

func (*NodeTemplateMapper) To(in *aggregate.NodeTemplate) (*outbound.NodeTemplateModel, error) {
	return &outbound.NodeTemplateModel{
		ID:        in.ID,
		Name:      in.Name,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
	}, nil
}

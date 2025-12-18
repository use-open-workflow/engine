package inbound

import (
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/inbound"
)

type NodeTemplateMapper struct{}

func NewNodeTemplateMapper() *NodeTemplateMapper {
	return &NodeTemplateMapper{}
}

func (m *NodeTemplateMapper) To(nodeTemplate *aggregate.NodeTemplate) (*inbound.NodeTemplateDTO, error) {
	return &inbound.NodeTemplateDTO{
		ID:   nodeTemplate.ID,
		Name: nodeTemplate.Name,
	}, nil
}

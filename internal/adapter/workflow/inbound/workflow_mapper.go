package inbound

import (
	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/domain/workflow/entity"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
)

type WorkflowMapper struct{}

func NewWorkflowMapper() *WorkflowMapper {
	return &WorkflowMapper{}
}

func (m *WorkflowMapper) ToWorkflowDTO(workflow *aggregate.Workflow) (*inbound.WorkflowDTO, error) {
	nodeDefDTOs := make([]*inbound.NodeDefinitionDTO, len(workflow.NodeDefinitions))
	for i, nd := range workflow.NodeDefinitions {
		dto, err := m.ToNodeDefinitionDTO(nd)
		if err != nil {
			return nil, err
		}
		nodeDefDTOs[i] = dto
	}

	edgeDTOs := make([]*inbound.EdgeDTO, len(workflow.Edges))
	for i, e := range workflow.Edges {
		dto, err := m.ToEdgeDTO(e)
		if err != nil {
			return nil, err
		}
		edgeDTOs[i] = dto
	}

	return &inbound.WorkflowDTO{
		ID:              workflow.ID,
		Name:            workflow.Name,
		Description:     workflow.Description,
		NodeDefinitions: nodeDefDTOs,
		Edges:           edgeDTOs,
		CreatedAt:       workflow.CreatedAt,
		UpdatedAt:       workflow.UpdatedAt,
	}, nil
}

func (m *WorkflowMapper) ToNodeDefinitionDTO(nodeDef *entity.NodeDefinition) (*inbound.NodeDefinitionDTO, error) {
	return &inbound.NodeDefinitionDTO{
		ID:             nodeDef.ID,
		WorkflowID:     nodeDef.WorkflowID,
		NodeTemplateID: nodeDef.NodeTemplateID,
		Name:           nodeDef.Name,
		Config:         nodeDef.Config,
		PositionX:      nodeDef.PositionX,
		PositionY:      nodeDef.PositionY,
		CreatedAt:      nodeDef.CreatedAt,
		UpdatedAt:      nodeDef.UpdatedAt,
	}, nil
}

func (m *WorkflowMapper) ToEdgeDTO(edge *entity.Edge) (*inbound.EdgeDTO, error) {
	return &inbound.EdgeDTO{
		ID:                   edge.ID,
		WorkflowID:           edge.WorkflowID,
		FromNodeDefinitionID: edge.FromNodeDefinitionID,
		ToNodeDefinitionID:   edge.ToNodeDefinitionID,
		CreatedAt:            edge.CreatedAt,
	}, nil
}

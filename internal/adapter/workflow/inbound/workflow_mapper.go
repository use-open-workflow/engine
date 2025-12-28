package inbound

import (
	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// WorkflowMapper implements the WorkflowMapper interface.
type WorkflowMapper struct{}

// NewWorkflowMapper creates a new WorkflowMapper.
func NewWorkflowMapper() *WorkflowMapper {
	return &WorkflowMapper{}
}

// To converts a Workflow aggregate to a WorkflowDTO.
func (m *WorkflowMapper) To(workflow *aggregate.Workflow) (*inbound.WorkflowDTO, error) {
	// 1. Convert NodeDefinitions
	nodeDefinitions := make([]inbound.NodeDefinitionDTO, len(workflow.NodeDefinitions))
	for i, nd := range workflow.NodeDefinitions {
		nodeDefinitions[i] = inbound.NodeDefinitionDTO{
			ID:             nd.ID,
			WorkflowID:     nd.WorkflowID,
			NodeTemplateID: nd.NodeTemplateID,
			Name:           nd.Name,
			PositionX:      nd.PositionX,
			PositionY:      nd.PositionY,
		}
	}

	// 2. Convert Edges
	edges := make([]inbound.EdgeDTO, len(workflow.Edges))
	for i, e := range workflow.Edges {
		edges[i] = inbound.EdgeDTO{
			ID:         e.ID,
			WorkflowID: e.WorkflowID,
			FromNodeID: e.FromNodeID,
			ToNodeID:   e.ToNodeID,
		}
	}

	// 3. Build and return DTO
	return &inbound.WorkflowDTO{
		ID:              workflow.ID,
		Name:            workflow.Name,
		NodeDefinitions: nodeDefinitions,
		Edges:           edges,
		CreatedAt:       workflow.CreatedAt,
		UpdatedAt:       workflow.UpdatedAt,
	}, nil
}

// ToList converts a slice of Workflow aggregates to WorkflowDTOs.
func (m *WorkflowMapper) ToList(workflows []*aggregate.Workflow) ([]*inbound.WorkflowDTO, error) {
	result := make([]*inbound.WorkflowDTO, len(workflows))
	for i, w := range workflows {
		dto, err := m.To(w)
		if err != nil {
			return nil, err
		}
		result[i] = dto
	}
	return result, nil
}

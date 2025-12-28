package inbound

import (
	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/domain/workflow/entity"
)

type WorkflowMapper interface {
	ToWorkflowDTO(workflow *aggregate.Workflow) (*WorkflowDTO, error)
	ToNodeDefinitionDTO(nodeDef *entity.NodeDefinition) (*NodeDefinitionDTO, error)
	ToEdgeDTO(edge *entity.Edge) (*EdgeDTO, error)
}

package aggregate

import (
	"use-open-workflow.io/engine/internal/domain/node/event"
	"use-open-workflow.io/engine/pkg/domain"
)

type NodeTemplate struct {
	domain.BaseAggregate
	Name string
}

func newNodeTemplate(id string, name string) *NodeTemplate {
	nodeTemplate := &NodeTemplate{
		BaseAggregate: domain.BaseAggregate{
			ID: id,
		},
		Name: name,
	}
	nodeTemplate.AddEvent(event.NewCreateNodeTemplate(nodeTemplate.ID))
	return nodeTemplate
}

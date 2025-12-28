package aggregate

import (
	"time"

	"use-open-workflow.io/engine/internal/domain/node/event"
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type NodeTemplate struct {
	domain.BaseAggregate
	Name string
}

func newNodeTemplate(idFactory id.Factory, aggregateID string, name string) *NodeTemplate {
	nodeTemplate := &NodeTemplate{
		BaseAggregate: domain.NewBaseAggregate(aggregateID),
		Name:          name,
	}
	nodeTemplate.AddEvent(event.NewCreateNodeTemplate(idFactory, nodeTemplate.ID, name))
	return nodeTemplate
}

func ReconstituteNodeTemplate(aggregateID string, name string, createdAt time.Time, updatedAt time.Time) *NodeTemplate {
	return &NodeTemplate{
		BaseAggregate: domain.ReconstituteBaseAggregate(aggregateID, createdAt, updatedAt),
		Name:          name,
	}
}

func (n *NodeTemplate) UpdateName(idFactory id.Factory, name string) {
	n.Name = name
	n.SetUpdatedAt(time.Now().UTC())
	n.AddEvent(event.NewUpdateNodeTemplate(idFactory, n.ID, name))
}

package event

import "use-open-workflow.io/engine/pkg/domain"

type CreateNodeTemplate struct {
	domain.BaseEvent
	nodeTemplateID string
}

func NewCreateNodeTemplate(nodeTemplateID string) *CreateNodeTemplate {
	return &CreateNodeTemplate{
		BaseEvent:      domain.NewBaseEvent("ulid"),
		nodeTemplateID: nodeTemplateID,
	}
}

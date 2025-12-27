package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type UpdateNodeTemplate struct {
	domain.BaseEvent
	NodeTemplateID string `json:"node_template_id"`
	Name           string `json:"name"`
}

func NewUpdateNodeTemplate(idFactory id.Factory, nodeTemplateID, name string) *UpdateNodeTemplate {
	return &UpdateNodeTemplate{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			nodeTemplateID,
			"NodeTemplate",
			"UpdateNodeTemplate",
		),
		NodeTemplateID: nodeTemplateID,
		Name:           name,
	}
}

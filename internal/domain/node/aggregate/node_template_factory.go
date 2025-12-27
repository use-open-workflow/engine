package aggregate

import (
	"use-open-workflow.io/engine/pkg/id"
)

type NodeTemplateFactory struct {
	idFactory id.Factory
}

func NewNodeTemplateFactory(idFactory id.Factory) *NodeTemplateFactory {
	return &NodeTemplateFactory{
		idFactory: idFactory,
	}
}

func (s *NodeTemplateFactory) Make(name string) *NodeTemplate {
	return newNodeTemplate(s.idFactory, s.idFactory.New(), name)
}

package aggregate

import (
	"use-open-workflow.io/engine/pkg/id"
)

type WorkflowFactory struct {
	idFactory id.Factory
}

func NewWorkflowFactory(idFactory id.Factory) *WorkflowFactory {
	return &WorkflowFactory{
		idFactory: idFactory,
	}
}

func (f *WorkflowFactory) Make(name string, description string) *Workflow {
	return newWorkflow(f.idFactory, f.idFactory.New(), name, description)
}

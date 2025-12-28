package aggregate

import "use-open-workflow.io/engine/pkg/id"

// WorkflowFactory creates new Workflow aggregates.
type WorkflowFactory struct {
	idFactory id.Factory
}

// NewWorkflowFactory creates a new WorkflowFactory.
func NewWorkflowFactory(idFactory id.Factory) *WorkflowFactory {
	return &WorkflowFactory{
		idFactory: idFactory,
	}
}

// Make creates a new Workflow aggregate with the given name.
func (f *WorkflowFactory) Make(name string) *Workflow {
	return newWorkflow(f.idFactory, f.idFactory.New(), name)
}

// IDFactory returns the ID factory for use in aggregate methods.
func (f *WorkflowFactory) IDFactory() id.Factory {
	return f.idFactory
}

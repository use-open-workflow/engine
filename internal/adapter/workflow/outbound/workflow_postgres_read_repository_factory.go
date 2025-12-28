package outbound

import (
	"use-open-workflow.io/engine/internal/port/outbound"
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
)

type WorkflowPostgresReadRepositoryFactory struct{}

func NewWorkflowPostgresReadRepositoryFactory() *WorkflowPostgresReadRepositoryFactory {
	return &WorkflowPostgresReadRepositoryFactory{}
}

func (f *WorkflowPostgresReadRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowReadRepository {
	return NewWorkflowPostgresReadRepository(uow)
}

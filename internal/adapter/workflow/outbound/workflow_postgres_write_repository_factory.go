package outbound

import (
	"use-open-workflow.io/engine/internal/port/outbound"
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
)

type WorkflowPostgresWriteRepositoryFactory struct{}

func NewWorkflowPostgresWriteRepositoryFactory() *WorkflowPostgresWriteRepositoryFactory {
	return &WorkflowPostgresWriteRepositoryFactory{}
}

func (f *WorkflowPostgresWriteRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowWriteRepository {
	return NewWorkflowPostgresWriteRepository(uow)
}

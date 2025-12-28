package outbound

import (
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresWriteRepositoryFactory creates write repositories.
type WorkflowPostgresWriteRepositoryFactory struct{}

// NewWorkflowPostgresWriteRepositoryFactory creates a new factory.
func NewWorkflowPostgresWriteRepositoryFactory() *WorkflowPostgresWriteRepositoryFactory {
	return &WorkflowPostgresWriteRepositoryFactory{}
}

// Create creates a UoW-scoped write repository.
func (f *WorkflowPostgresWriteRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowWriteRepository {
	return NewWorkflowPostgresWriteRepository(uow)
}

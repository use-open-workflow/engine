package outbound

import (
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresReadRepositoryFactory creates read repositories.
type WorkflowPostgresReadRepositoryFactory struct{}

// NewWorkflowPostgresReadRepositoryFactory creates a new factory.
func NewWorkflowPostgresReadRepositoryFactory() *WorkflowPostgresReadRepositoryFactory {
	return &WorkflowPostgresReadRepositoryFactory{}
}

// Create creates a UoW-scoped read repository.
func (f *WorkflowPostgresReadRepositoryFactory) Create(uow outbound.UnitOfWork) workflowOutbound.WorkflowReadRepository {
	return NewWorkflowPostgresReadRepository(uow)
}

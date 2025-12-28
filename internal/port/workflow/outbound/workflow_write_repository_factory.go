package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// WorkflowWriteRepositoryFactory creates UoW-scoped write repositories.
type WorkflowWriteRepositoryFactory interface {
	Create(uow portOutbound.UnitOfWork) WorkflowWriteRepository
}

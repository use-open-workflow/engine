package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// WorkflowReadRepositoryFactory creates UoW-scoped read repositories.
type WorkflowReadRepositoryFactory interface {
	Create(uow portOutbound.UnitOfWork) WorkflowReadRepository
}

package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type WorkflowReadRepositoryFactory interface {
	Create(uow outbound.UnitOfWork) WorkflowReadRepository
}

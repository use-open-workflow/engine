package outbound

import portOutbound "use-open-workflow.io/engine/internal/port/outbound"

// NodeTemplateWriteRepositoryFactory creates write repositories bound to a UoW
type NodeTemplateWriteRepositoryFactory interface {
	// Create returns a NodeTemplateWriteRepository that uses the given UoW
	// for transaction management and aggregate registration
	//
	// The returned repository:
	//   - Uses uow.Querier() for SQL execution
	//   - Calls uow.RegisterNew/RegisterDirty/RegisterDeleted
	//   - Does NOT manage transaction lifecycle (that's the service's job)
	Create(uow portOutbound.UnitOfWork) NodeTemplateWriteRepository
}

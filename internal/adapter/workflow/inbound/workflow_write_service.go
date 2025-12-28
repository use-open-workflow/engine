package inbound

import (
	"context"
	"fmt"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
	"use-open-workflow.io/engine/pkg/id"
)

// WorkflowWriteService implements the WorkflowWriteService interface.
type WorkflowWriteService struct {
	uowFactory             outbound.UnitOfWorkFactory
	writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory
	readRepositoryFactory  workflowOutbound.WorkflowReadRepositoryFactory
	factory                *aggregate.WorkflowFactory
	mapper                 inbound.WorkflowMapper
	idFactory              id.Factory
}

// NewWorkflowWriteService creates a new WorkflowWriteService.
func NewWorkflowWriteService(
	uowFactory outbound.UnitOfWorkFactory,
	writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory,
	readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory,
	factory *aggregate.WorkflowFactory,
	mapper inbound.WorkflowMapper,
	idFactory id.Factory,
) *WorkflowWriteService {
	return &WorkflowWriteService{
		uowFactory:             uowFactory,
		writeRepositoryFactory: writeRepositoryFactory,
		readRepositoryFactory:  readRepositoryFactory,
		factory:                factory,
		mapper:                 mapper,
		idFactory:              idFactory,
	}
}

// Create creates a new workflow.
func (s *WorkflowWriteService) Create(ctx context.Context, input inbound.CreateWorkflowInput) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Create workflow aggregate
	workflow := s.factory.Make(input.Name)

	// 2. Save to repository
	if err = writeRepo.Save(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	// 3. Commit transaction
	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

// Update updates an existing workflow.
func (s *WorkflowWriteService) Update(ctx context.Context, id string, input inbound.UpdateWorkflowInput) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Load existing workflow
	workflow, err := readRepo.FindByID(txCtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}

	// 2. Update aggregate
	workflow.UpdateName(s.idFactory, input.Name)

	// 3. Persist changes
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

// Delete deletes a workflow.
func (s *WorkflowWriteService) Delete(ctx context.Context, id string) error {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	if err = writeRepo.Delete(txCtx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddNodeDefinition adds a node definition to a workflow.
func (s *WorkflowWriteService) AddNodeDefinition(ctx context.Context, workflowID string, input inbound.AddNodeDefinitionInput) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Load workflow
	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// 2. Add node definition to aggregate
	workflow.AddNodeDefinition(s.idFactory, input.NodeTemplateID, input.Name, input.PositionX, input.PositionY)

	// 3. Persist changes (full sync of child entities)
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

// RemoveNodeDefinition removes a node definition from a workflow.
func (s *WorkflowWriteService) RemoveNodeDefinition(ctx context.Context, workflowID, nodeID string) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Load workflow
	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// 2. Remove node from aggregate (cascades to edges)
	if err = workflow.RemoveNodeDefinition(nodeID); err != nil {
		return nil, fmt.Errorf("failed to remove node definition: %w", err)
	}

	// 3. Persist changes
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

// AddEdge adds an edge between two node definitions.
func (s *WorkflowWriteService) AddEdge(ctx context.Context, workflowID string, input inbound.AddEdgeInput) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Load workflow
	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// 2. Add edge to aggregate (validates nodes exist)
	_, err = workflow.AddEdge(s.idFactory, input.FromNodeID, input.ToNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to add edge: %w", err)
	}

	// 3. Persist changes
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

// RemoveEdge removes an edge from a workflow.
func (s *WorkflowWriteService) RemoveEdge(ctx context.Context, workflowID, edgeID string) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// 1. Load workflow
	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// 2. Remove edge from aggregate
	if err = workflow.RemoveEdge(edgeID); err != nil {
		return nil, fmt.Errorf("failed to remove edge: %w", err)
	}

	// 3. Persist changes
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(workflow)
}

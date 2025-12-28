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

type WorkflowWriteService struct {
	uowFactory             outbound.UnitOfWorkFactory
	writeRepositoryFactory workflowOutbound.WorkflowWriteRepositoryFactory
	readRepositoryFactory  workflowOutbound.WorkflowReadRepositoryFactory
	factory                *aggregate.WorkflowFactory
	mapper                 inbound.WorkflowMapper
	idFactory              id.Factory
}

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

	workflow := s.factory.Make(input.Name, input.Description)

	if err = writeRepo.Save(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.ToWorkflowDTO(workflow)
}

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

	workflow, err := readRepo.FindByID(txCtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}

	workflow.UpdateName(s.idFactory, input.Name)
	workflow.UpdateDescription(s.idFactory, input.Description)

	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.ToWorkflowDTO(workflow)
}

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

func (s *WorkflowWriteService) AddNodeDefinition(ctx context.Context, workflowID string, input inbound.AddNodeDefinitionInput) (*inbound.NodeDefinitionDTO, error) {
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

	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	nodeDef := workflow.AddNodeDefinition(
		s.idFactory,
		input.NodeTemplateID,
		input.Name,
		input.Config,
		input.PositionX,
		input.PositionY,
	)

	if err = writeRepo.SaveNodeDefinition(txCtx, nodeDef); err != nil {
		return nil, fmt.Errorf("failed to save node definition: %w", err)
	}

	// Update workflow's updated_at
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.ToNodeDefinitionDTO(nodeDef)
}

func (s *WorkflowWriteService) UpdateNodeDefinition(ctx context.Context, workflowID string, nodeDefID string, input inbound.UpdateNodeDefinitionInput) (*inbound.NodeDefinitionDTO, error) {
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

	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	nodeDef := workflow.GetNodeDefinition(nodeDefID)
	if nodeDef == nil {
		return nil, fmt.Errorf("node definition not found: %s", nodeDefID)
	}

	// Apply updates
	if input.Name != "" {
		nodeDef.UpdateName(input.Name)
	}
	if input.Config != nil {
		nodeDef.UpdateConfig(input.Config)
	}
	if input.PositionX != nil && input.PositionY != nil {
		nodeDef.UpdatePosition(*input.PositionX, *input.PositionY)
	}

	if err = writeRepo.UpdateNodeDefinition(txCtx, nodeDef); err != nil {
		return nil, fmt.Errorf("failed to update node definition: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.ToNodeDefinitionDTO(nodeDef)
}

func (s *WorkflowWriteService) RemoveNodeDefinition(ctx context.Context, workflowID string, nodeDefID string) error {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if err = workflow.RemoveNodeDefinition(s.idFactory, nodeDefID); err != nil {
		return fmt.Errorf("failed to remove node definition: %w", err)
	}

	// Delete from database (CASCADE handles edges)
	if err = writeRepo.DeleteNodeDefinition(txCtx, nodeDefID); err != nil {
		return fmt.Errorf("failed to delete node definition: %w", err)
	}

	// Update workflow's updated_at
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *WorkflowWriteService) AddEdge(ctx context.Context, workflowID string, input inbound.AddEdgeInput) (*inbound.EdgeDTO, error) {
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

	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	edge, err := workflow.AddEdge(s.idFactory, input.FromNodeDefinitionID, input.ToNodeDefinitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to add edge: %w", err)
	}

	if err = writeRepo.SaveEdge(txCtx, edge); err != nil {
		return nil, fmt.Errorf("failed to save edge: %w", err)
	}

	// Update workflow's updated_at
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.ToEdgeDTO(edge)
}

func (s *WorkflowWriteService) RemoveEdge(ctx context.Context, workflowID string, edgeID string) error {
	uow := s.uowFactory.Create()
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	workflow, err := readRepo.FindByID(txCtx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if err = workflow.RemoveEdge(s.idFactory, edgeID); err != nil {
		return fmt.Errorf("failed to remove edge: %w", err)
	}

	if err = writeRepo.DeleteEdge(txCtx, edgeID); err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	// Update workflow's updated_at
	if err = writeRepo.Update(txCtx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

package inbound

import (
	"context"

	"use-open-workflow.io/engine/internal/port/node/inbound"
	"use-open-workflow.io/engine/internal/port/node/outbound"
)

type NodeTemplateReadService struct {
	nodeTemplateMapper         inbound.NodeTemplateMapper
	nodeTemplateReadRepository outbound.NodeTemplateReadRepository
}

func NewNodeTemplateReadService(
	mapper inbound.NodeTemplateMapper,
	repository outbound.NodeTemplateReadRepository,
) *NodeTemplateReadService {
	return &NodeTemplateReadService{
		nodeTemplateMapper:         mapper,
		nodeTemplateReadRepository: repository,
	}
}

func (s *NodeTemplateReadService) List(ctx context.Context) ([]*inbound.NodeTemplateDTO, error) {
	nodeTemplates, err := s.nodeTemplateReadRepository.FindMany(ctx)
	if err != nil {
		return nil, err
	}

	nodeTemplateDTOs := make([]*inbound.NodeTemplateDTO, len(nodeTemplates))
	for i, v := range nodeTemplates {
		nodeTemplateDTO, err := s.nodeTemplateMapper.To(v)
		if err != nil {
			return nil, err
		}
		nodeTemplateDTOs[i] = nodeTemplateDTO
	}

	return nodeTemplateDTOs, nil
}

func (s *NodeTemplateReadService) GetByID(ctx context.Context, id string) (*inbound.NodeTemplateDTO, error) {
	nodeTemplate, err := s.nodeTemplateReadRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if nodeTemplate == nil {
		return nil, nil
	}

	return s.nodeTemplateMapper.To(nodeTemplate)
}

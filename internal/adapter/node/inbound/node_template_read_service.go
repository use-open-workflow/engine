package inbound

import (
	"use-open-workflow.io/engine/internal/port/node/inbound"
	"use-open-workflow.io/engine/internal/port/node/outbound"
)

type NodeTemplateReadService struct {
	nodeTemplateMapper         inbound.NodeTemplateMapper
	nodeTemplateReadRepository outbound.NodeTemplateRepository
}

func NewNodeTemplateReadService(
	mapper inbound.NodeTemplateMapper,
	repository outbound.NodeTemplateRepository,
) *NodeTemplateReadService {
	return &NodeTemplateReadService{
		nodeTemplateMapper:         mapper,
		nodeTemplateReadRepository: repository,
	}
}

func (s *NodeTemplateReadService) List() ([]*inbound.NodeTemplateDTO, error) {
	nodeTemplates, err := s.nodeTemplateReadRepository.FindMany()

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

package http

import (
	"github.com/gofiber/fiber/v3"
	"use-open-workflow.io/engine/internal/port/node/inbound"
)

type NodeTemplateHandler struct {
	service inbound.NodeTemplateReadService
}

func NewNodeTemplateHandler(service inbound.NodeTemplateReadService) *NodeTemplateHandler {
	return &NodeTemplateHandler{
		service: service,
	}
}

func (h *NodeTemplateHandler) List(c fiber.Ctx) error {
	nodeTemplates, err := h.service.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(nodeTemplates)
}

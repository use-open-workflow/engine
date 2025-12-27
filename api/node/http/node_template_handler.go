package http

import (
	"github.com/gofiber/fiber/v3"
	"use-open-workflow.io/engine/internal/port/node/inbound"
)

type NodeTemplateHandler struct {
	readService  inbound.NodeTemplateReadService
	writeService inbound.NodeTemplateWriteService
}

func NewNodeTemplateHandler(
	readService inbound.NodeTemplateReadService,
	writeService inbound.NodeTemplateWriteService,
) *NodeTemplateHandler {
	return &NodeTemplateHandler{
		readService:  readService,
		writeService: writeService,
	}
}

func (h *NodeTemplateHandler) List(c fiber.Ctx) error {
	nodeTemplates, err := h.readService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(nodeTemplates)
}

func (h *NodeTemplateHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	nodeTemplate, err := h.readService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if nodeTemplate == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "node template not found",
		})
	}
	return c.JSON(nodeTemplate)
}

func (h *NodeTemplateHandler) Create(c fiber.Ctx) error {
	var input inbound.CreateNodeTemplateInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	nodeTemplate, err := h.writeService.Create(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(nodeTemplate)
}

func (h *NodeTemplateHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var input inbound.UpdateNodeTemplateInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	nodeTemplate, err := h.writeService.Update(c.Context(), id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(nodeTemplate)
}

func (h *NodeTemplateHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.writeService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

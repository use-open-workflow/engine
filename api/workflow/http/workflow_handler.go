package http

import (
	"github.com/gofiber/fiber/v3"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
)

type WorkflowHandler struct {
	readService  inbound.WorkflowReadService
	writeService inbound.WorkflowWriteService
}

func NewWorkflowHandler(
	readService inbound.WorkflowReadService,
	writeService inbound.WorkflowWriteService,
) *WorkflowHandler {
	return &WorkflowHandler{
		readService:  readService,
		writeService: writeService,
	}
}

// Workflow CRUD

func (h *WorkflowHandler) List(c fiber.Ctx) error {
	workflows, err := h.readService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(workflows)
}

func (h *WorkflowHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	workflow, err := h.readService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if workflow == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	return c.JSON(workflow)
}

func (h *WorkflowHandler) Create(c fiber.Ctx) error {
	var input inbound.CreateWorkflowInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	workflow, err := h.writeService.Create(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workflow)
}

func (h *WorkflowHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var input inbound.UpdateWorkflowInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	workflow, err := h.writeService.Update(c.Context(), id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(workflow)
}

func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.writeService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// NodeDefinition operations (nested under Workflow)

func (h *WorkflowHandler) AddNodeDefinition(c fiber.Ctx) error {
	workflowID := c.Params("id")
	var input inbound.AddNodeDefinitionInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	nodeDef, err := h.writeService.AddNodeDefinition(c.Context(), workflowID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(nodeDef)
}

func (h *WorkflowHandler) UpdateNodeDefinition(c fiber.Ctx) error {
	workflowID := c.Params("id")
	nodeDefID := c.Params("nodeDefId")
	var input inbound.UpdateNodeDefinitionInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	nodeDef, err := h.writeService.UpdateNodeDefinition(c.Context(), workflowID, nodeDefID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(nodeDef)
}

func (h *WorkflowHandler) RemoveNodeDefinition(c fiber.Ctx) error {
	workflowID := c.Params("id")
	nodeDefID := c.Params("nodeDefId")

	if err := h.writeService.RemoveNodeDefinition(c.Context(), workflowID, nodeDefID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Edge operations (nested under Workflow)

func (h *WorkflowHandler) AddEdge(c fiber.Ctx) error {
	workflowID := c.Params("id")
	var input inbound.AddEdgeInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	edge, err := h.writeService.AddEdge(c.Context(), workflowID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(edge)
}

func (h *WorkflowHandler) RemoveEdge(c fiber.Ctx) error {
	workflowID := c.Params("id")
	edgeID := c.Params("edgeId")

	if err := h.writeService.RemoveEdge(c.Context(), workflowID, edgeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

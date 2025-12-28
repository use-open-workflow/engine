package http

import (
	"github.com/gofiber/fiber/v3"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
)

// WorkflowHandler handles HTTP requests for workflow operations.
type WorkflowHandler struct {
	readService  inbound.WorkflowReadService
	writeService inbound.WorkflowWriteService
}

// NewWorkflowHandler creates a new handler.
func NewWorkflowHandler(
	readService inbound.WorkflowReadService,
	writeService inbound.WorkflowWriteService,
) *WorkflowHandler {
	return &WorkflowHandler{
		readService:  readService,
		writeService: writeService,
	}
}

// List handles GET /api/v1/workflow
func (h *WorkflowHandler) List(c fiber.Ctx) error {
	workflows, err := h.readService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(workflows)
}

// GetByID handles GET /api/v1/workflow/:id
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

// Create handles POST /api/v1/workflow
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

// Update handles PUT /api/v1/workflow/:id
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

// Delete handles DELETE /api/v1/workflow/:id
func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.writeService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddNodeDefinition handles POST /api/v1/workflow/:id/node
func (h *WorkflowHandler) AddNodeDefinition(c fiber.Ctx) error {
	workflowID := c.Params("id")
	var input inbound.AddNodeDefinitionInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	workflow, err := h.writeService.AddNodeDefinition(c.Context(), workflowID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workflow)
}

// RemoveNodeDefinition handles DELETE /api/v1/workflow/:id/node/:nodeId
func (h *WorkflowHandler) RemoveNodeDefinition(c fiber.Ctx) error {
	workflowID := c.Params("id")
	nodeID := c.Params("nodeId")

	workflow, err := h.writeService.RemoveNodeDefinition(c.Context(), workflowID, nodeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(workflow)
}

// AddEdge handles POST /api/v1/workflow/:id/edge
func (h *WorkflowHandler) AddEdge(c fiber.Ctx) error {
	workflowID := c.Params("id")
	var input inbound.AddEdgeInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	workflow, err := h.writeService.AddEdge(c.Context(), workflowID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workflow)
}

// RemoveEdge handles DELETE /api/v1/workflow/:id/edge/:edgeId
func (h *WorkflowHandler) RemoveEdge(c fiber.Ctx) error {
	workflowID := c.Params("id")
	edgeID := c.Params("edgeId")

	workflow, err := h.writeService.RemoveEdge(c.Context(), workflowID, edgeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(workflow)
}

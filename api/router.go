package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"use-open-workflow.io/engine/api/node/http"
	workflowHttp "use-open-workflow.io/engine/api/workflow/http"
	"use-open-workflow.io/engine/di"
)

func SetupRouter(c *di.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Open Workflow API",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	api := app.Group("/api/v1")
	registerNodeTemplateRoutes(api, c)
	registerWorkflowRoutes(api, c)

	return app
}

func registerNodeTemplateRoutes(router fiber.Router, c *di.Container) {
	nodeTemplateHandler := http.NewNodeTemplateHandler(
		c.NodeTemplateReadService,
		c.NodeTemplateWriteService,
	)

	nodeTemplate := router.Group("/node-template")
	nodeTemplate.Get("/", nodeTemplateHandler.List)
	nodeTemplate.Get("/:id", nodeTemplateHandler.GetByID)
	nodeTemplate.Post("/", nodeTemplateHandler.Create)
	nodeTemplate.Put("/:id", nodeTemplateHandler.Update)
	nodeTemplate.Delete("/:id", nodeTemplateHandler.Delete)
}

func registerWorkflowRoutes(router fiber.Router, c *di.Container) {
	workflowHandler := workflowHttp.NewWorkflowHandler(
		c.WorkflowReadService,
		c.WorkflowWriteService,
	)

	workflow := router.Group("/workflow")

	// Workflow CRUD
	workflow.Get("/", workflowHandler.List)
	workflow.Get("/:id", workflowHandler.GetByID)
	workflow.Post("/", workflowHandler.Create)
	workflow.Put("/:id", workflowHandler.Update)
	workflow.Delete("/:id", workflowHandler.Delete)

	// NodeDefinition operations (nested)
	workflow.Post("/:id/node-definition", workflowHandler.AddNodeDefinition)
	workflow.Put("/:id/node-definition/:nodeDefId", workflowHandler.UpdateNodeDefinition)
	workflow.Delete("/:id/node-definition/:nodeDefId", workflowHandler.RemoveNodeDefinition)

	// Edge operations (nested)
	workflow.Post("/:id/edge", workflowHandler.AddEdge)
	workflow.Delete("/:id/edge/:edgeId", workflowHandler.RemoveEdge)
}

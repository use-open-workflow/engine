package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	nodeHttp "use-open-workflow.io/engine/api/node/http"
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
	nodeTemplateHandler := nodeHttp.NewNodeTemplateHandler(
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
	workflow.Get("/", workflowHandler.List)
	workflow.Get("/:id", workflowHandler.GetByID)
	workflow.Post("/", workflowHandler.Create)
	workflow.Put("/:id", workflowHandler.Update)
	workflow.Delete("/:id", workflowHandler.Delete)

	// Node definition routes
	workflow.Post("/:id/node", workflowHandler.AddNodeDefinition)
	workflow.Delete("/:id/node/:nodeId", workflowHandler.RemoveNodeDefinition)

	// Edge routes
	workflow.Post("/:id/edge", workflowHandler.AddEdge)
	workflow.Delete("/:id/edge/:edgeId", workflowHandler.RemoveEdge)
}

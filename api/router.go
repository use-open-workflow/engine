package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"use-open-workflow.io/engine/api/node/http"
	"use-open-workflow.io/engine/di"
)

func SetupRouter(c *di.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Open Workflow API",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	api := app.Group("/api/v1")
	registerNodeRoutes(api, c)

	return app
}

func registerNodeRoutes(router fiber.Router, c *di.Container) {
	nodeTemplateHandler := http.NewNodeTemplateHandler(c.NodeTemplateReadService)

	nodeTemplates := router.Group("/node-templates")
	nodeTemplates.Get("/", nodeTemplateHandler.List)
}

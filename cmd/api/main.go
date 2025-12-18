package main

import (
	"log"

	"use-open-workflow.io/engine/api"
	"use-open-workflow.io/engine/di"
)

func main() {
	c := di.NewContainer()

	app := api.SetupRouter(c)

	log.Fatal(app.Listen(":3000"))
}

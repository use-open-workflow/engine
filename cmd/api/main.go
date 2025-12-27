package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"use-open-workflow.io/engine/api"
	"use-open-workflow.io/engine/di"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := di.NewContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer c.Close()

	if err := c.OutboxProcessor.Start(ctx); err != nil {
		log.Fatalf("Failed to start outbox processor: %v", err)
	}

	app := api.SetupRouter(c)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down...")
		cancel()
		app.Shutdown()
	}()

	log.Println("Starting server on :3000")
	log.Fatal(app.Listen(":3000"))
}

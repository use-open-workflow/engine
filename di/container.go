package di

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	nodeAdapterInbound "use-open-workflow.io/engine/internal/adapter/node/inbound"
	nodeAdapterOutbound "use-open-workflow.io/engine/internal/adapter/node/outbound"
	adapterOutbound "use-open-workflow.io/engine/internal/adapter/outbound"
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/inbound"
	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/pkg/id"
)

type Container struct {
	Pool                     *pgxpool.Pool
	NodeTemplateReadService  inbound.NodeTemplateReadService
	NodeTemplateWriteService inbound.NodeTemplateWriteService
	OutboxProcessor          outbound.OutboxProcessor
}

func NewContainer(ctx context.Context) (*Container, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/open_workflow?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Shared dependencies
	idFactory := id.NewULIDFactory()

	// Unit of Work Factory
	uowFactory := adapterOutbound.NewUnitOfWorkPostgresFactory(pool)

	// Mappers
	nodeTemplateInboundMapper := nodeAdapterInbound.NewNodeTemplateMapper()

	// Factory
	nodeTemplateFactory := aggregate.NewNodeTemplateFactory(idFactory)

	// Repositories
	nodeTemplateReadRepository := nodeAdapterOutbound.NewNodeTemplatePostgresReadRepository(pool)

	// Write Repository Factory (replaces sharedUow + static repository)
	nodeTemplateWriteRepositoryFactory := nodeAdapterOutbound.NewNodeTemplatePostgresWriteRepositoryFactory()

	// Services
	nodeTemplateReadService := nodeAdapterInbound.NewNodeTemplateReadService(
		nodeTemplateInboundMapper,
		nodeTemplateReadRepository,
	)

	nodeTemplateWriteService := nodeAdapterInbound.NewNodeTemplateWriteService(
		uowFactory,
		nodeTemplateWriteRepositoryFactory,
		nodeTemplateFactory,
		nodeTemplateReadRepository,
		nodeTemplateInboundMapper,
		idFactory,
	)

	outboxReadRepository := adapterOutbound.NewOutboxPostgresReadRepository(pool)
	outboxWriteRepository := adapterOutbound.NewOutboxPostgresWriteRepository(pool)
	eventPublisher := adapterOutbound.NewOutboxNoopEventPublisher()
	outboxProcessor := adapterOutbound.NewOutboxProcessor(
		outboxReadRepository,
		outboxWriteRepository,
		eventPublisher,
		adapterOutbound.DefaultConfig(),
	)

	return &Container{
		Pool:                     pool,
		NodeTemplateReadService:  nodeTemplateReadService,
		NodeTemplateWriteService: nodeTemplateWriteService,
		OutboxProcessor:          outboxProcessor,
	}, nil
}

func (c *Container) Close() {
	if c.OutboxProcessor != nil {
		c.OutboxProcessor.Stop()
	}
	if c.Pool != nil {
		c.Pool.Close()
	}
}

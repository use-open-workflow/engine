package outbound

import (
	"context"
	"log"

	"use-open-workflow.io/engine/internal/port/outbound"
)

type OutboxNoopEventPublisher struct{}

func NewOutboxNoopEventPublisher() *OutboxNoopEventPublisher {
	return &OutboxNoopEventPublisher{}
}

func (p *OutboxNoopEventPublisher) Publish(ctx context.Context, msg *outbound.OutboxMessage) error {
	log.Printf("[OUTBOX] Event published: type=%s aggregate=%s/%s id=%s",
		msg.EventType,
		msg.AggregateType,
		msg.AggregateID,
		msg.ID,
	)
	return nil
}

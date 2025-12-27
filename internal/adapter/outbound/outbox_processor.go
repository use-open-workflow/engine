package outbound

import (
	"context"
	"log"
	"sync"
	"time"

	"use-open-workflow.io/engine/internal/port/outbound"
)

type Config struct {
	BatchSize       int
	PollInterval    time.Duration
	CleanupInterval time.Duration
	RetentionPeriod time.Duration
}

func DefaultConfig() Config {
	return Config{
		BatchSize:       100,
		PollInterval:    5 * time.Second,
		CleanupInterval: 1 * time.Hour,
		RetentionPeriod: 7 * 24 * time.Hour, // 7 days
	}
}

type OutboxProcessor struct {
	readRepository  outbound.OutboxReadRepository
	writeRepository outbound.OutboxWriteRepository
	eventPublisher  outbound.OutboxEventPublisher
	config          Config

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewOutboxProcessor(
	readRepository outbound.OutboxReadRepository,
	writeRepository outbound.OutboxWriteRepository,
	publisher outbound.OutboxEventPublisher,
	config Config,
) *OutboxProcessor {
	return &OutboxProcessor{
		readRepository:  readRepository,
		writeRepository: writeRepository,
		eventPublisher:  publisher,
		config:          config,
		stopCh:          make(chan struct{}),
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) error {
	p.wg.Add(2)

	go func() {
		defer p.wg.Done()
		p.processLoop(ctx)
	}()

	go func() {
		defer p.wg.Done()
		p.cleanupLoop(ctx)
	}()

	log.Println("Outbox processor started")
	return nil
}

func (p *OutboxProcessor) Stop() error {
	close(p.stopCh)
	p.wg.Wait()
	log.Println("Outbox processor stopped")
	return nil
}

func (p *OutboxProcessor) processLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				log.Printf("Error processing outbox batch: %v", err)
			}
		}
	}
}

func (p *OutboxProcessor) processBatch(ctx context.Context) error {
	messages, err := p.readRepository.FindUnprocessed(ctx, p.config.BatchSize)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if err := p.eventPublisher.Publish(ctx, msg); err != nil {
			log.Printf("Failed to publish message %s: %v", msg.ID, err)
			if err := p.writeRepository.IncrementRetry(ctx, msg.ID); err != nil {
				log.Printf("Failed to increment retry for %s: %v", msg.ID, err)
			}
			continue
		}

		if err := p.writeRepository.MarkProcessed(ctx, msg.ID); err != nil {
			log.Printf("Failed to mark message %s as processed: %v", msg.ID, err)
		}
	}

	return nil
}

func (p *OutboxProcessor) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.writeRepository.DeleteProcessed(ctx, p.config.RetentionPeriod); err != nil {
				log.Printf("Error cleaning up outbox: %v", err)
			}
		}
	}
}

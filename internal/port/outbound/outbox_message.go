package outbound

import "time"

type OutboxMessage struct {
	ID            string
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	RetryCount    int
}

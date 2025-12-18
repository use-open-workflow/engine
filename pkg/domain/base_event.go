package domain

type BaseEvent struct {
	ID string
}

func NewBaseEvent(id string) BaseEvent {
	return BaseEvent{
		ID: id,
	}
}

func (BaseEvent) IsEvent() bool {
	return true
}

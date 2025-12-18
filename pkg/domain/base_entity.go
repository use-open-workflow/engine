package domain

type BaseEntity struct {
	ID string
}

func NewBaseEntity(id string) BaseEntity {
	return BaseEntity{
		ID: id,
	}
}

func (BaseEntity) IsEntity() bool {
	return true
}

package domain

type BaseValueObject[T any] struct {
	Value T
}

func NewValueObject[T any](value T) BaseValueObject[T] {
	return BaseValueObject[T]{
		Value: value,
	}
}

func (BaseValueObject[any]) ValueObject() bool {
	return true
}

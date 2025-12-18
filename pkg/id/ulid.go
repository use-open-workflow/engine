package id

import "github.com/oklog/ulid/v2"

type ULIDFactory struct{}

func NewULIDFactory() *ULIDFactory {
	return &ULIDFactory{}
}

func (*ULIDFactory) New() string {
	return ulid.Make().String()
}

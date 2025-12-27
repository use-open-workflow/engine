package outbound

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"use-open-workflow.io/engine/internal/port/outbound"
)

type UnitOfWorkPostgresFactory struct {
	pool *pgxpool.Pool
}

func NewUnitOfWorkPostgresFactory(pool *pgxpool.Pool) *UnitOfWorkPostgresFactory {
	return &UnitOfWorkPostgresFactory{pool: pool}
}

func (f *UnitOfWorkPostgresFactory) Create() outbound.UnitOfWork {
	return NewUnitOfWorkPostgres(f.pool)
}

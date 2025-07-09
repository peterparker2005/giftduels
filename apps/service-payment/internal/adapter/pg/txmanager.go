package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

func NewPgxTxManager(pool *pgxpool.Pool) TxManager {
	return &pgxTxManager{pool}
}

type pgxTxManager struct {
	pool *pgxpool.Pool
}

func (m *pgxTxManager) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return m.pool.Begin(ctx)
}

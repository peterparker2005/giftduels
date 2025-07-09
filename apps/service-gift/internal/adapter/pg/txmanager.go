package pg

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type TxManager interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Sql() *sql.DB
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

func (m *pgxTxManager) Sql() *sql.DB {
	return stdlib.OpenDBFromPool(m.pool)
}

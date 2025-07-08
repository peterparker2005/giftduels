// internal/adapter/pg/ton_cursor_repo.go
package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
)

type tonCursorRepo struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// NewTonCursorRepo регистрирует реализацию CursorRepository
func NewTonCursorRepo(pool *pgxpool.Pool) ton.CursorRepository {
	return &tonCursorRepo{
		pool: pool,
		q:    sqlc.New(pool),
	}
}

func (r *tonCursorRepo) Get(ctx context.Context, network, walletAddress string) (uint64, error) {
	net, err := ToDBTonNetwork(network)
	if err != nil {
		return 0, fmt.Errorf("ton cursor get: %w", err)
	}
	lastLt, err := r.q.GetTonCursor(ctx, sqlc.GetTonCursorParams{
		Network:       net,
		WalletAddress: walletAddress,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("ton cursor get: %w", err)
	}
	return uint64(lastLt), nil
}

func (r *tonCursorRepo) Upsert(ctx context.Context, network, walletAddress string, lastLT uint64) error {
	net, err := ToDBTonNetwork(network)
	if err != nil {
		return fmt.Errorf("ton cursor upsert: %w", err)
	}
	return r.q.UpsertTonCursor(ctx, sqlc.UpsertTonCursorParams{
		Network:       net,
		WalletAddress: walletAddress,
		LastLt:        int64(lastLT),
	})
}

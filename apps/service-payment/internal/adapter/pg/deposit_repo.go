package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/ccoveille/go-safecast"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

type DepositRepository struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewDepositRepository(pool *pgxpool.Pool, logger *logger.Logger) ton.DepositRepository {
	return &DepositRepository{
		q:      sqlc.New(pool),
		pool:   pool,
		logger: logger,
	}
}

func (r *DepositRepository) WithTx(tx pgx.Tx) ton.DepositRepository {
	return &DepositRepository{
		q:      r.q.WithTx(tx),
		pool:   r.pool,
		logger: r.logger,
	}
}

func (r *DepositRepository) CreateDeposit(ctx context.Context, params *ton.CreateDepositParams) (*ton.Deposit, error) {
	amountNano, err := safecast.ToInt64(params.AmountNano)
	if err != nil {
		return nil, err
	}
	deposit, err := r.q.CreateDeposit(ctx, sqlc.CreateDepositParams{
		TelegramUserID: params.TelegramUserID,
		AmountNano:     amountNano,
		Payload:        params.Payload,
		ExpiresAt:      pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *DepositRepository) GetDepositByPayload(ctx context.Context, payload string) (*ton.Deposit, error) {
	deposit, err := r.q.GetDepositByPayload(ctx, payload)
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *DepositRepository) SetDepositTransaction(
	ctx context.Context,
	params *ton.SetDepositTransactionParams,
) (*ton.Deposit, error) {
	id, err := pgUUID(params.ID)
	if err != nil {
		return nil, err
	}

	txLtInt, err := safecast.ToInt64(params.TxLt)
	if err != nil {
		return nil, err
	}
	deposit, err := r.q.SetDepositTransaction(ctx, sqlc.SetDepositTransactionParams{
		ID:     id,
		TxHash: pgtype.Text{String: params.TxHash, Valid: true},
		TxLt:   pgtype.Int8{Int64: txLtInt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *DepositRepository) GetCursor(ctx context.Context, network, walletAddress string) (uint64, error) {
	net, err := ToDBTonNetwork(network)
	if err != nil {
		return 0, fmt.Errorf("ton cursor get: %w", err)
	}
	lastLt, err := r.q.GetTonCursor(ctx, sqlc.GetTonCursorParams{
		Network:       net,
		WalletAddress: walletAddress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("ton cursor get: %w", err)
	}
	lastLtUint, err := safecast.ToUint64(lastLt)
	if err != nil {
		return 0, err
	}
	return lastLtUint, nil
}

func (r *DepositRepository) UpsertCursor(ctx context.Context, network, walletAddress string, lastLT uint64) error {
	net, err := ToDBTonNetwork(network)
	if err != nil {
		return fmt.Errorf("ton cursor upsert: %w", err)
	}
	lastLtInt, err := safecast.ToInt64(lastLT)
	if err != nil {
		return err
	}
	return r.q.UpsertTonCursor(ctx, sqlc.UpsertTonCursorParams{
		Network:       net,
		WalletAddress: walletAddress,
		LastLt:        lastLtInt,
	})
}

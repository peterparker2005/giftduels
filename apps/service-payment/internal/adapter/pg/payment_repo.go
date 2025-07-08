package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	payment "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type repo struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewPaymentRepository(pool *pgxpool.Pool, logger *logger.Logger) payment.Repository {
	return &repo{
		q:      sqlc.New(pool),
		pool:   pool,
		logger: logger,
	}
}

func (r *repo) Create(ctx context.Context, params *payment.CreateBalanceParams) error {
	_, err := r.q.CreateUserBalance(ctx, sqlc.CreateUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) GetUserBalance(ctx context.Context, telegramUserID int64) (*payment.Balance, error) {
	balance, err := r.q.GetUserBalance(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	return &payment.Balance{
		ID:             balance.ID,
		TelegramUserID: balance.TelegramUserID,
		TonAmount:      balance.TonAmount,
		CreatedAt:      balance.CreatedAt.Time,
		UpdatedAt:      balance.UpdatedAt.Time,
	}, nil
}

func (r *repo) CreateTransaction(ctx context.Context, params *payment.CreateTransactionParams) error {
	_, err := r.q.CreateUserTransaction(ctx, sqlc.CreateUserTransactionParams{
		TelegramUserID: params.TelegramUserID,
		Amount:         params.Amount,
		Reason:         sqlc.TransactionReason(params.Reason),
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) AddUserBalance(ctx context.Context, params *payment.AddUserBalanceParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				r.logger.Error("rollback", zap.Error(err))
			}
		}
	}()

	qtx := r.q.WithTx(tx)

	err = qtx.AddUserBalance(ctx, sqlc.AddUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      params.Amount,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *repo) SpendUserBalance(ctx context.Context, params *payment.SpendUserBalanceParams) error {
	err := r.q.SpendUserBalance(ctx, sqlc.SpendUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      params.Amount,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreateDeposit(ctx context.Context, params *payment.CreateDepositParams) (*payment.Deposit, error) {
	deposit, err := r.q.CreateDeposit(ctx, sqlc.CreateDepositParams{
		TelegramUserID: params.TelegramUserID,
		AmountNano:     params.AmountNano,
		Payload:        params.Payload,
		ExpiresAt:      pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *repo) GetDepositByPayload(ctx context.Context, payload string) (*payment.Deposit, error) {
	deposit, err := r.q.GetDepositByPayload(ctx, payload)
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *repo) SetDepositTransaction(ctx context.Context, params *payment.SetDepositTransactionParams) (*payment.Deposit, error) {
	deposit, err := r.q.SetDepositTransaction(ctx, sqlc.SetDepositTransactionParams{
		ID:     pgtype.UUID{Bytes: uuid.MustParse(params.ID), Valid: true},
		TxHash: pgtype.Text{String: params.TxHash, Valid: true},
		TxLt:   pgtype.Int8{Int64: params.TxLt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

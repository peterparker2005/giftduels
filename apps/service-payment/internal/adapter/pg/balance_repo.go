package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	payment "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
)

type BalanceRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewBalanceRepository(pool *pgxpool.Pool) payment.Repository {
	return &BalanceRepository{
		q:    sqlc.New(pool),
		pool: pool,
	}
}

func (r *BalanceRepository) Create(ctx context.Context, params *payment.CreateBalanceParams) error {
	_, err := r.q.CreateUserBalance(ctx, sqlc.CreateUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) GetUserBalance(ctx context.Context, telegramUserID int64) (*payment.Balance, error) {
	balance, err := r.q.GetUserBalance(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	return &payment.Balance{
		ID:             balance.ID,
		TelegramUserID: balance.TelegramUserID,
		TonBalance:     balance.TonBalance,
		CreatedAt:      balance.CreatedAt.Time,
		UpdatedAt:      balance.UpdatedAt.Time,
	}, nil
}

func (r *BalanceRepository) CreateTransaction(ctx context.Context, params *payment.CreateTransactionParams) error {
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

func (r *BalanceRepository) AddUserBalance(ctx context.Context, params *payment.AddUserBalanceParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	err = qtx.AddUserBalance(ctx, sqlc.AddUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonBalance:     params.Amount,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) SpendUserBalance(ctx context.Context, params *payment.SpendUserBalanceParams) error {
	err := r.q.SpendUserBalance(ctx, sqlc.SpendUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonBalance:     params.Amount,
	})
	if err != nil {
		return err
	}
	return nil
}

package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	payment "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
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

func (r *repo) WithTx(tx pgx.Tx) payment.Repository {
	return &repo{
		q:      sqlc.New(tx),
		pool:   r.pool,
		logger: r.logger,
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
		ID:             balance.ID.String(),
		TelegramUserID: balance.TelegramUserID,
		TonAmount:      balance.TonAmount,
		CreatedAt:      balance.CreatedAt.Time,
		UpdatedAt:      balance.UpdatedAt.Time,
	}, nil
}

func (r *repo) CreateTransaction(ctx context.Context, params *payment.CreateTransactionParams) error {
	_, err := r.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		TelegramUserID: params.TelegramUserID,
		Amount:         params.Amount,
		Reason:         sqlc.TransactionReason(params.Reason),
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) AddUserBalance(ctx context.Context, params *payment.AddUserBalanceParams) (*payment.Balance, error) {
	balance, err := r.q.AddUserBalance(ctx, sqlc.AddUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      params.Amount,
	})
	if err != nil {
		return nil, err
	}

	return ToBalanceDomain(balance), nil
}

func (r *repo) SpendUserBalance(ctx context.Context, params *payment.SpendUserBalanceParams) (*payment.Balance, error) {
	// Сначала проверяем текущий баланс
	currentBalance, err := r.GetUserBalance(ctx, params.TelegramUserID)
	if err != nil {
		return nil, err
	}

	// Проверяем достаточность средств
	if currentBalance.TonAmount < params.Amount {
		return nil, errors.NewInsufficientTonError("insufficient balance for withdrawal")
	}

	// Списываем средства
	balance, err := r.q.SpendUserBalance(ctx, sqlc.SpendUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      params.Amount,
	})
	if err != nil {
		return nil, err
	}

	return ToBalanceDomain(balance), nil
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
		ID:     mustPgUUID(params.ID),
		TxHash: pgtype.Text{String: params.TxHash, Valid: true},
		TxLt:   pgtype.Int8{Int64: params.TxLt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return ToDepositDomain(deposit), nil
}

func (r *repo) DeleteTransaction(ctx context.Context, id string) error {
	err := r.q.DeleteTransaction(ctx, mustPgUUID(id))
	if err != nil {
		return err
	}
	return nil
}

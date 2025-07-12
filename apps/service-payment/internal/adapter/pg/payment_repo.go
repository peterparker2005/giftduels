package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	payment "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
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
		Metadata:       params.Metadata,
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

func (r *repo) DeleteTransaction(ctx context.Context, id string) error {
	err := r.q.DeleteTransaction(ctx, mustPgUUID(id))
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) GetUserTransactions(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) ([]*payment.Transaction, error) {
	transactions, err := r.q.GetUserTransactions(ctx, sqlc.GetUserTransactionsParams{
		TelegramUserID: telegramUserID,
		Limit:          int32(pagination.PageSize()),
		Offset:         int32(pagination.Offset()),
	})
	if err != nil {
		return nil, err
	}

	transactionsDomain := make([]*payment.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		transactionsDomain = append(transactionsDomain, ToTransactionDomain(transaction))
	}
	return transactionsDomain, nil
}

func (r *repo) GetUserTransactionsCount(ctx context.Context, telegramUserID int64) (int64, error) {
	count, err := r.q.GetUserTransactionsCount(ctx, telegramUserID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

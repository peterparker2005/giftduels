package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	payment "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	errorspkg "github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
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
	b, err := r.q.GetUserBalance(ctx, telegramUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zero := tonamount.Zero()
			return &payment.Balance{
				TelegramUserID: telegramUserID,
				TonAmount:      zero,
			}, nil
		}
		return nil, err
	}
	tonStr, err := fromPgNumeric(b.TonAmount)
	if err != nil {
		return nil, err
	}
	ta, err := tonamount.NewTonAmountFromString(tonStr)
	if err != nil {
		return nil, err
	}
	return &payment.Balance{
		ID:             b.ID.String(),
		TelegramUserID: b.TelegramUserID,
		TonAmount:      ta,
		CreatedAt:      b.CreatedAt.Time,
		UpdatedAt:      b.UpdatedAt.Time,
	}, nil
}

func (r *repo) CreateTransaction(
	ctx context.Context,
	params *payment.CreateTransactionParams,
) error {
	amount, err := pgNumeric(params.Amount.String())
	if err != nil {
		return err
	}

	_, err = r.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		TelegramUserID: params.TelegramUserID,
		Amount:         amount,
		Reason:         sqlc.TransactionReason(params.Reason),
		Metadata:       params.Metadata,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) AddUserBalance(
	ctx context.Context,
	params *payment.AddUserBalanceParams,
) (*payment.Balance, error) {
	amount, err := pgNumeric(params.Amount.String())
	if err != nil {
		return nil, err
	}
	b, err := r.q.UpsertUserBalance(ctx, sqlc.UpsertUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      amount,
	})
	if err != nil {
		return nil, err
	}
	return ToBalanceDomain(b), nil
}

func (r *repo) SpendUserBalance(
	ctx context.Context,
	params *payment.SpendUserBalanceParams,
) (*payment.Balance, error) {
	amount, err := pgNumeric(params.Amount.String())
	if err != nil {
		return nil, err
	}

	// Сначала проверяем текущий баланс
	currentBalance, err := r.GetUserBalance(ctx, params.TelegramUserID)
	if err != nil {
		return nil, err
	}

	// Проверяем достаточность средств
	if currentBalance.TonAmount.Decimal().Cmp(params.Amount.Decimal()) < 0 {
		return nil, errorspkg.NewInsufficientTonError("insufficient balance for withdrawal")
	}

	// Списываем средства
	balance, err := r.q.SpendUserBalance(ctx, sqlc.SpendUserBalanceParams{
		TelegramUserID: params.TelegramUserID,
		TonAmount:      amount,
	})
	if err != nil {
		return nil, err
	}

	return ToBalanceDomain(balance), nil
}

func (r *repo) DeleteTransaction(ctx context.Context, id string) error {
	idPg, err := pgUUID(id)
	if err != nil {
		return err
	}

	err = r.q.DeleteTransaction(ctx, idPg)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) GetUserTransactions(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) ([]*payment.Transaction, error) {
	transactions, err := r.q.GetUserTransactions(ctx, sqlc.GetUserTransactionsParams{
		TelegramUserID: telegramUserID,
		Limit:          pagination.PageSize(),
		Offset:         pagination.Offset(),
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

package payment

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type CreateBalanceParams struct {
	TelegramUserID int64
}

type CreateTransactionParams struct {
	TelegramUserID int64
	Amount         float64
	Reason         TransactionReason
	Metadata       []byte
}

type AddUserBalanceParams struct {
	TelegramUserID int64
	Amount         float64
}

type SpendUserBalanceParams struct {
	TelegramUserID int64
	Amount         float64
}

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	Create(ctx context.Context, params *CreateBalanceParams) error

	CreateTransaction(ctx context.Context, params *CreateTransactionParams) error
	DeleteTransaction(ctx context.Context, id string) error

	GetUserBalance(ctx context.Context, telegramUserID int64) (*Balance, error)
	AddUserBalance(ctx context.Context, params *AddUserBalanceParams) (*Balance, error)
	SpendUserBalance(ctx context.Context, params *SpendUserBalanceParams) (*Balance, error)

	GetUserTransactions(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) ([]*Transaction, error)
	GetUserTransactionsCount(ctx context.Context, telegramUserID int64) (int64, error)
}

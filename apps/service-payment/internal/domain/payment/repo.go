package payment

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type CreateBalanceParams struct {
	TelegramUserID int64
}

type CreateTransactionParams struct {
	TelegramUserID int64
	Amount         float64
	Reason         TransactionReason
}

type AddUserBalanceParams struct {
	TelegramUserID int64
	Amount         float64
}

type SpendUserBalanceParams struct {
	TelegramUserID int64
	Amount         float64
}

type SetDepositTransactionParams struct {
	ID     string
	TxHash string
	TxLt   int64
}

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	Create(ctx context.Context, params *CreateBalanceParams) error

	CreateTransaction(ctx context.Context, params *CreateTransactionParams) error
	DeleteTransaction(ctx context.Context, id int32) error

	GetUserBalance(ctx context.Context, telegramUserID int64) (*Balance, error)
	AddUserBalance(ctx context.Context, params *AddUserBalanceParams) (*Balance, error)
	SpendUserBalance(ctx context.Context, params *SpendUserBalanceParams) (*Balance, error)

	CreateDeposit(ctx context.Context, params *CreateDepositParams) (*Deposit, error)
	GetDepositByPayload(ctx context.Context, payload string) (*Deposit, error)
	SetDepositTransaction(ctx context.Context, params *SetDepositTransactionParams) (*Deposit, error)
}

package payment

import (
	"context"
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
	Create(ctx context.Context, params *CreateBalanceParams) error
	CreateTransaction(ctx context.Context, params *CreateTransactionParams) error
	AddUserBalance(ctx context.Context, params *AddUserBalanceParams) error
	GetUserBalance(ctx context.Context, telegramUserID int64) (*Balance, error)
	SpendUserBalance(ctx context.Context, params *SpendUserBalanceParams) error
	CreateDeposit(ctx context.Context, params *CreateDepositParams) (*Deposit, error)
	GetDepositByPayload(ctx context.Context, payload string) (*Deposit, error)
	SetDepositTransaction(ctx context.Context, params *SetDepositTransactionParams) (*Deposit, error)
}

package ton

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type SetDepositTransactionParams struct {
	ID     string
	TxHash string
	TxLt   uint64
}

// DepositRepository keeps and returns last_lt for given address and network.
type DepositRepository interface {
	WithTx(tx pgx.Tx) DepositRepository
	CreateDeposit(ctx context.Context, params *CreateDepositParams) (*Deposit, error)
	GetDepositByPayload(ctx context.Context, payload string) (*Deposit, error)
	SetDepositTransaction(ctx context.Context, params *SetDepositTransactionParams) (*Deposit, error)
	// Get returns saved lastLT for walletAddress and network.
	// If no record exists, returns 0 and nil-error.
	GetCursor(ctx context.Context, network, walletAddress string) (uint64, error)
	// UpsertCursor saves or updates cursor for walletAddress and network.
	UpsertCursor(ctx context.Context, network, walletAddress string, lastLT uint64) error
}

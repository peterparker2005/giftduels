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

// CursorRepository хранит и возвращает last_lt для заданного адреса и сети.
type DepositRepository interface {
	WithTx(tx pgx.Tx) DepositRepository
	CreateDeposit(ctx context.Context, params *CreateDepositParams) (*Deposit, error)
	GetDepositByPayload(ctx context.Context, payload string) (*Deposit, error)
	SetDepositTransaction(ctx context.Context, params *SetDepositTransactionParams) (*Deposit, error)
	// Get возвращает сохранённый lastLT для walletAddress и network.
	// Если записи нет, возвращает 0 и nil-ошибку.
	GetCursor(ctx context.Context, network, walletAddress string) (uint64, error)
	// UpsertCursor сохраняет или обновляет курсор для walletAddress и network.
	UpsertCursor(ctx context.Context, network, walletAddress string, lastLT uint64) error
}

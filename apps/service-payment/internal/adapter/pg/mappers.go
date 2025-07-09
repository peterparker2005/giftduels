package pg

import (
	"fmt"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
)

func ToBalanceDomain(b sqlc.UserBalance) *payment.Balance {
	return &payment.Balance{
		TelegramUserID: b.TelegramUserID,
		TonAmount:      b.TonAmount,
		ID:             b.ID,
		CreatedAt:      b.CreatedAt.Time,
		UpdatedAt:      b.UpdatedAt.Time,
	}
}

func ToDepositDomain(d sqlc.Deposit) *payment.Deposit {
	var txHash *string
	if d.TxHash.Valid {
		txHash = &d.TxHash.String
	}

	var txLt *uint64
	if d.TxLt.Valid {
		val := uint64(d.TxLt.Int64)
		txLt = &val
	}

	return &payment.Deposit{
		ID:             d.ID.Bytes,
		TelegramUserID: d.TelegramUserID,
		Status:         payment.DepositStatus(d.Status),
		AmountNano:     d.AmountNano,
		Payload:        d.Payload,
		ExpiresAt:      d.ExpiresAt.Time,
		TxHash:         txHash,
		TxLt:           txLt,
		CreatedAt:      d.CreatedAt.Time,
		UpdatedAt:      d.UpdatedAt.Time,
	}
}

func ToDBTonNetwork(n string) (sqlc.TonNetwork, error) {
	switch n {
	case "mainnet":
		return sqlc.TonNetworkMainnet, nil
	case "testnet":
		return sqlc.TonNetworkTestnet, nil
	}
	return sqlc.TonNetwork(""), fmt.Errorf("unknown ton network: %s", n)
}

package pg

import (
	"encoding/json"
	"fmt"

	"github.com/ccoveille/go-safecast"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

func ToBalanceDomain(b sqlc.UserBalance) *payment.Balance {
	tonAmountStr, err := fromPgNumeric(b.TonAmount)
	if err != nil {
		panic(err)
	}

	tonAmount, err := tonamount.NewTonAmountFromString(tonAmountStr)
	if err != nil {
		panic(err)
	}

	return &payment.Balance{
		ID:             b.ID.String(),
		TelegramUserID: b.TelegramUserID,
		TonAmount:      tonAmount,
		CreatedAt:      b.CreatedAt.Time,
		UpdatedAt:      b.UpdatedAt.Time,
	}
}

func ToDepositDomain(d sqlc.Deposit) *ton.Deposit {
	var txHash *string
	if d.TxHash.Valid {
		txHash = &d.TxHash.String
	}

	var txLt *uint64
	if d.TxLt.Valid {
		val, err := safecast.ToUint64(d.TxLt.Int64)
		if err != nil {
			panic(err)
		}
		txLt = &val
	}

	amountNano, err := safecast.ToUint64(d.AmountNano)
	if err != nil {
		panic(err)
	}

	return &ton.Deposit{
		ID:             d.ID.Bytes,
		TelegramUserID: d.TelegramUserID,
		Status:         ton.DepositStatus(d.Status),
		AmountNano:     amountNano,
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

func ToTransactionDomain(t sqlc.UserTransaction) *payment.Transaction {
	var metadata *payment.TransactionMetadata
	if t.Metadata != nil {
		// Создаём экземпляр, чтобы в него можно было прочитать данные
		m := &payment.TransactionMetadata{}
		// Безопасно парсим JSON
		if err := json.Unmarshal(t.Metadata, m); err == nil {
			metadata = m
		}
	}

	amountStr, err := fromPgNumeric(t.Amount)
	if err != nil {
		panic(err)
	}

	amount, err := tonamount.NewTonAmountFromString(amountStr)
	if err != nil {
		panic(err)
	}

	return &payment.Transaction{
		ID:             t.ID.String(),
		TelegramUserID: t.TelegramUserID,
		Amount:         amount,
		Reason:         payment.TransactionReason(t.Reason),
		CreatedAt:      t.CreatedAt.Time,
		Metadata:       metadata,
	}
}

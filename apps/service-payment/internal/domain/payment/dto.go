package payment

import "time"

type Balance struct {
	ID             string
	TelegramUserID int64
	TonAmount      float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Transaction struct {
	ID             string
	TelegramUserID int64
	Amount         float64
	Reason         TransactionReason
	Metadata       *TransactionMetadata
	CreatedAt      time.Time
}

type WithdrawOptions struct {
	TotalStarsFee uint32
	TotalTonFee   float64
}

type GiftFee struct {
	StarsFee uint32
	TonFee   float64
}

type TransactionReason string

const (
	TransactionReasonDeposit  TransactionReason = "deposit"
	TransactionReasonWithdraw TransactionReason = "withdraw"
	TransactionReasonRefund   TransactionReason = "refund"
)

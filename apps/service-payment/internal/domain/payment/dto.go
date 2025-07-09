package payment

import "time"

type Balance struct {
	ID             int32
	TelegramUserID int64
	TonAmount      float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Transaction struct {
	ID             int32
	TelegramUserID int64
	Amount         float64
	Reason         TransactionReason
	CreatedAt      time.Time
}

type WithdrawOptions struct {
	Fees          []GiftFee
	TotalStarsFee uint32
	TotalTonFee   float64
}

type GiftFee struct {
	GiftID   string
	StarsFee uint32
	TonFee   float64
}

type TransactionReason string

const (
	TransactionReasonWithdrawal TransactionReason = "withdrawal"
)

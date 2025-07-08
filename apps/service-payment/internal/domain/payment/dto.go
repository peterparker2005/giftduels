package payment

import "time"

type Balance struct {
	ID             int32
	TelegramUserID int64
	TonBalance     float64
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

type TransactionReason string

const (
	TransactionReasonWithdrawal TransactionReason = "withdrawal"
)

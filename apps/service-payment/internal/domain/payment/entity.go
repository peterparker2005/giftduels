package payment

import (
	"time"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type Balance struct {
	ID             string
	TelegramUserID int64
	TonAmount      *tonamount.TonAmount
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Transaction struct {
	ID             string
	TelegramUserID int64
	Amount         *tonamount.TonAmount
	Reason         TransactionReason
	Metadata       *TransactionMetadata
	CreatedAt      time.Time
}

type GiftWithdrawRequest struct {
	GiftID string
	Price  *tonamount.TonAmount
}

type WithdrawOptions struct {
	GiftFees      []*GiftFee
	TotalStarsFee uint32
	TotalTonFee   *tonamount.TonAmount
}

type GiftFee struct {
	GiftID   string
	StarsFee uint32
	TonFee   *tonamount.TonAmount
}

type TransactionReason string

const (
	TransactionReasonDeposit  TransactionReason = "deposit"
	TransactionReasonWithdraw TransactionReason = "withdraw"
	TransactionReasonRefund   TransactionReason = "refund"
)

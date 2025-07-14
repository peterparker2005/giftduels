package ton

import (
	"time"

	"github.com/google/uuid"
)

type DepositStatus string

const (
	DepositStatusPending   DepositStatus = "pending"
	DepositStatusReceived  DepositStatus = "received"
	DepositStatusConfirmed DepositStatus = "confirmed"
	DepositStatusExpired   DepositStatus = "expired"
)

type Deposit struct {
	ID             uuid.UUID
	TelegramUserID int64
	Status         DepositStatus
	AmountNano     uint64
	Payload        string
	ExpiresAt      time.Time
	TxHash         *string
	TxLt           *uint64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateDepositParams struct {
	TelegramUserID int64
	AmountNano     uint64
	Payload        string
	ExpiresAt      time.Time
}

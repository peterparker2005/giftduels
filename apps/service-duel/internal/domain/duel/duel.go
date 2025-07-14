package duel

import (
	"time"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type ID string

func (id ID) String() string {
	return string(id)
}

type TelegramUserID int64

type Status string

const (
	StatusWaitingForOpponent Status = "waiting_for_opponent"
	StatusInProgress         Status = "in_progress"
	StatusCompleted          Status = "completed"
	StatusCancelled          Status = "cancelled"
)

type Duel struct {
	ID               ID
	DisplayNumber    int64
	Params           Params
	WinnerID         *TelegramUserID
	NextRollDeadline *time.Time
	TotalStakeValue  *tonamount.TonAmount
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time

	Participants []Participant
	Stakes       []Stake
	Rounds       []Round
}

type Params struct {
	IsPrivate  bool
	MaxPlayers MaxPlayers
	MaxGifts   MaxGifts
}

type Participant struct {
	TelegramUserID TelegramUserID
	IsCreator      bool
}

type Stake struct {
	TelegramUserID TelegramUserID
	GiftID         string
	StakeValue     *tonamount.TonAmount
}

type Round struct {
	RoundNumber int
	Rolls       []Roll
}

type Roll struct {
	TelegramUserID TelegramUserID
	DiceValue      int
	RolledAt       time.Time
	IsAutoRolled   bool
}

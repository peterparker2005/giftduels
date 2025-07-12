package duel

import (
	"time"
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
	Params           DuelParams
	WinnerID         *TelegramUserID
	NextRollDeadline *time.Time
	TotalStakeValue  float64
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time

	Participants []Participant
	Stakes       []Stake
	Rounds       []Round
}

type DuelParams struct {
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
	StakeValue     float64
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

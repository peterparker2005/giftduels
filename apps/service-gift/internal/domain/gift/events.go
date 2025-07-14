package gift

import (
	"errors"
	"time"
)

type Event struct {
	ID             string
	GiftID         string
	TelegramUserID int64
	EventType      EventType
	RelatedGameID  *string
	OccurredAt     time.Time
}

type EventType string

const (
	EventTypeStake            EventType = "stake"
	EventTypeReturnFromGame   EventType = "return_from_game"
	EventTypeDeposit          EventType = "deposit"
	EventTypeWithdrawRequest  EventType = "withdraw_request"
	EventTypeWithdrawComplete EventType = "withdraw_complete"
	EventTypeWithdrawFail     EventType = "withdraw_fail"
)

func NewEvent(p *NewEventParams) (*Event, error) {
	if p.EventType == "" {
		return nil, errors.New("event type is required")
	}

	return &Event{
		ID:             p.ID,
		GiftID:         p.GiftID,
		TelegramUserID: p.TelegramUserID,
		EventType:      p.EventType,
		RelatedGameID:  p.RelatedGameID,
		OccurredAt:     p.OccurredAt,
	}, nil
}

type NewEventParams struct {
	ID             string
	GiftID         string
	TelegramUserID int64
	EventType      EventType
	RelatedGameID  *string
	OccurredAt     time.Time
}

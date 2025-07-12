package gift

import (
	"encoding/json"
	"time"
)

type Gift struct {
	ID               string
	OwnerTelegramID  int64
	Status           Status
	Price            float64
	WithdrawnAt      *time.Time
	Title, Slug      string
	CollectibleID    int32
	UpgradeMessageID int32
	TelegramGiftID   int64
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Gift attributes - direct references to lookup tables
	Collection Collection
	Model      Model
	Backdrop   Backdrop
	Symbol     Symbol
}

type Collection struct {
	ID        int32
	Name      string
	ShortName string
}

type Model struct {
	ID             int32
	CollectionID   int32
	Name           string
	ShortName      string
	RarityPerMille int32
}

type Symbol struct {
	ID             int32
	Name           string
	ShortName      string
	RarityPerMille int32
}

type Backdrop struct {
	ID             int32
	Name           string
	ShortName      string
	RarityPerMille int32
	CenterColor    *string
	EdgeColor      *string
	PatternColor   *string
	TextColor      *string
}

type GiftEvent struct {
	ID            string
	GiftID        string
	FromUserID    *int64
	ToUserID      *int64
	Action        string
	GameMode      *string
	RelatedGameID *string
	Description   *string
	Payload       json.RawMessage
	OccurredAt    time.Time
}

type Status string

const (
	StatusInGame          Status = "in_game"
	StatusWithdrawn       Status = "withdrawn"
	StatusWithdrawPending Status = "withdraw_pending"
	StatusOwned           Status = "owned"
)

type Attribute struct {
	Type           AttributeType
	Name           string
	RarityPerMille int32
}

type AttributeType string

const (
	AttributeTypeBackdrop AttributeType = "backdrop"
	AttributeTypeModel    AttributeType = "model"
	AttributeTypeSymbol   AttributeType = "symbol"
)

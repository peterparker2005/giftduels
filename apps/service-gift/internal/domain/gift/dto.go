package gift

import "time"

type Gift struct {
	ID               string
	OwnerTelegramID  int64
	Status           Status
	Price            float64
	EmojiID          int64
	WithdrawnAt      *time.Time
	Title, Slug      string
	CollectibleID    int64
	UpgradeMessageID int32
	TelegramGiftID   int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Attributes       []Attribute
}

type GiftEvent struct {
	ID         string
	FromUserID int64
	ToUserID   int64
	CreatedAt  time.Time
}

type Status string

const (
	StatusInGame    Status = "in_game"
	StatusWithdrawn Status = "withdrawn"
	StatusPending   Status = "pending"
	StatusOwned     Status = "owned"
)

type AttributeType string

const (
	AttributeTypeModel    AttributeType = "model"
	AttributeTypeSymbol   AttributeType = "symbol"
	AttributeTypeBackdrop AttributeType = "backdrop"
)

type Attribute struct {
	Type           AttributeType
	Name           string
	RarityPerMille int32
}

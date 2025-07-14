package gift

import (
	"strings"
	"time"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type Gift struct {
	ID               string
	OwnerTelegramID  int64
	Status           Status
	Price            *tonamount.TonAmount
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

	// metadata for game
	RelatedDuelID string
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

type NewGiftParams struct {
	ID               string
	OwnerTelegramID  int64
	Price            *tonamount.TonAmount
	Title, Slug      string
	CollectibleID    int32
	UpgradeMessageID int32
	TelegramGiftID   int64
	Attributes       AttributeData
	Now              time.Time
}

// NewGift is a constructor for Gift.
func NewGift(p NewGiftParams) (*Gift, error) {
	if p.Title == "" {
		return nil, ErrTitleRequired
	}
	if p.Slug == "" {
		return nil, ErrSlugRequired
	}

	return &Gift{
		ID:               p.ID,
		OwnerTelegramID:  p.OwnerTelegramID,
		Price:            p.Price,
		Title:            p.Title,
		Slug:             p.Slug,
		Status:           StatusOwned,
		CreatedAt:        p.Now,
		UpdatedAt:        p.Now,
		Collection:       Collection{ID: p.Attributes.CollectionID},
		Model:            Model{ID: p.Attributes.ModelID},
		Backdrop:         Backdrop{ID: p.Attributes.BackdropID},
		Symbol:           Symbol{ID: p.Attributes.SymbolID},
		TelegramGiftID:   p.TelegramGiftID,
		CollectibleID:    p.CollectibleID,
		UpgradeMessageID: p.UpgradeMessageID,
		WithdrawnAt:      nil,
	}, nil
}

func (g *Gift) Stake(telegramUserID int64) error {
	if !g.IsOwnedBy(telegramUserID) {
		return ErrGiftNotOwned
	}
	if g.Status != StatusOwned {
		return ErrGiftCannotStake
	}
	g.Status = StatusInGame
	g.OwnerTelegramID = telegramUserID
	return nil
}

// ChangeOwner is a method to change the owner of a gift.
func (g *Gift) ChangeOwner(ownerTelegramID int64) error {
	if g.Status != StatusOwned {
		return ErrGiftNotOwned
	}
	if g.OwnerTelegramID == ownerTelegramID {
		return ErrGiftAlreadyOwned
	}
	g.OwnerTelegramID = ownerTelegramID
	return nil
}

// MarkForWithdrawal marks a gift for withdrawal.
func (g *Gift) MarkForWithdrawal() error {
	if g.Status != StatusOwned {
		return ErrGiftNotOwned
	}
	g.Status = StatusWithdrawPending
	return nil
}

// CancelWithdrawal cancels the withdrawal of a gift.
func (g *Gift) CancelWithdrawal() error {
	if g.Status != StatusWithdrawPending {
		return ErrGiftNotWithdrawPending
	}
	g.Status = StatusOwned
	return nil
}

// CompleteWithdrawal completes the withdrawal of a gift.
func (g *Gift) CompleteWithdrawal(at time.Time) error {
	if g.Status != StatusWithdrawPending {
		return ErrGiftNotWithdrawPending
	}
	g.Status = StatusWithdrawn
	g.WithdrawnAt = &at
	return nil
}

// IsOwnedBy checks if the gift is owned by the specified user.
func (g *Gift) IsOwnedBy(telegramUserID int64) bool {
	return g.OwnerTelegramID == telegramUserID
}

// CanBeWithdrawn checks if the gift can be withdrawn.
func (g *Gift) CanBeWithdrawn() bool {
	return g.Status == StatusOwned
}

// CanBeWithdrawnBy checks if the gift can be withdrawn by the specified user.
func (g *Gift) CanBeWithdrawnBy(telegramUserID int64) bool {
	return g.IsOwnedBy(telegramUserID) && g.CanBeWithdrawn()
}

// SetRelatedDuelID sets the related duel ID for a gift.
func (g *Gift) SetRelatedDuelID(duelID string) {
	g.RelatedDuelID = duelID
}

// AttributeData represents the attribute data for a gift.
type AttributeData struct {
	CollectionID int32
	ModelID      int32
	BackdropID   int32
	SymbolID     int32
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ShortName(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

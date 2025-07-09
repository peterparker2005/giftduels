package gift

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type CreateGiftAttributeParams struct {
	GiftID                  string
	AttributeType           AttributeType
	AttributeName           string
	AttributeRarityPerMille int32
}

type CreateGiftParams struct {
	GiftID           string
	OwnerTelegramID  int64
	CollectibleID    int64
	UpgradeMessageID int32
	TelegramGiftID   int64
	Status           Status
	Price            float64
	Title            string
	Slug             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type GetUserGiftsResult struct {
	Gifts []*Gift
	Total int64
}

type GiftRepository interface {
	WithTx(tx pgx.Tx) GiftRepository
	GetGiftByID(ctx context.Context, id string) (*Gift, error)
	GetUserGifts(ctx context.Context, limit int32, offset int32, ownerTelegramID int64) (*GetUserGiftsResult, error)
	GetUserActiveGifts(ctx context.Context, limit int32, offset int32, ownerTelegramID int64) (*GetUserGiftsResult, error)
	StakeGiftForGame(ctx context.Context, id string) (*Gift, error)
	UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*Gift, error)
	MarkGiftForWithdrawal(ctx context.Context, id string) (*Gift, error)
	CancelGiftWithdrawal(ctx context.Context, id string) (*Gift, error)
	CompleteGiftWithdrawal(ctx context.Context, id string) (*Gift, error)
	CreateGiftWithDetails(ctx context.Context, gift *CreateGiftParams, attributes []CreateGiftAttributeParams) (*Gift, error)
	CreateGiftEvent(ctx context.Context, giftID string, fromUserID, toUserID int64) (*GiftEvent, error)
	GetGiftEvents(ctx context.Context, giftID string, limit int32, offset int32) ([]*GiftEvent, error)
	GetGiftsByIDs(ctx context.Context, ids []string) ([]*Gift, error)
}

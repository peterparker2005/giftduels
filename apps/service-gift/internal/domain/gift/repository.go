package gift

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
)

type Repository interface {
	GetGiftByID(ctx context.Context, id string) (*db.Gift, error)
	GetUserGifts(ctx context.Context, limit int32, offset int32, ownerTelegramID int64) ([]*db.Gift, error)
	StakeGiftForGame(ctx context.Context, id string) (*db.Gift, error)
	UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*db.Gift, error)
	MarkGiftForWithdrawal(ctx context.Context, id string) (*db.Gift, error)
	CompleteGiftWithdrawal(ctx context.Context, id string) (*db.Gift, error)
	CreateGiftEvent(ctx context.Context, giftID string, fromUserID, toUserID int64) (*db.GiftEvent, error)
	GetGiftEvents(ctx context.Context, giftID string, limit int32, offset int32) ([]*db.GiftEvent, error)
	CreateGift(ctx context.Context, gift *db.Gift) (*db.Gift, error)
}

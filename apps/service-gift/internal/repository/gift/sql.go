package gift

import (
	"context"
	"database/sql"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
)

func NewSQLRepository(db *db.Queries) gift.Repository {
	return &sqlRepository{
		db: db,
	}
}

type sqlRepository struct {
	db *db.Queries
}

func (r *sqlRepository) GetGiftByID(ctx context.Context, id string) (*db.Gift, error) {
	gift, err := r.db.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

func (r *sqlRepository) GetUserGifts(ctx context.Context, limit int32, offset int32, ownerTelegramID int64) ([]*db.Gift, error) {
	gifts, err := r.db.GetUserGifts(ctx, db.GetUserGiftsParams{Limit: limit, Offset: offset, OwnerTelegramID: ownerTelegramID})
	if err != nil {
		return nil, err
	}

	convertedGifts := make([]*db.Gift, len(gifts))
	for i, gift := range gifts {
		convertedGifts[i] = &gift
	}
	return convertedGifts, nil
}

func (r *sqlRepository) StakeGiftForGame(ctx context.Context, id string) (*db.Gift, error) {
	gift, err := r.db.StakeGiftForGame(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

func (r *sqlRepository) UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*db.Gift, error) {
	gift, err := r.db.UpdateGiftOwner(ctx, db.UpdateGiftOwnerParams{
		ID:              id,
		OwnerTelegramID: ownerTelegramID,
	})
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

func (r *sqlRepository) MarkGiftForWithdrawal(ctx context.Context, id string) (*db.Gift, error) {
	gift, err := r.db.MarkGiftForWithdrawal(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

func (r *sqlRepository) CompleteGiftWithdrawal(ctx context.Context, id string) (*db.Gift, error) {
	gift, err := r.db.CompleteGiftWithdrawal(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

// TODO: fix
func (r *sqlRepository) CreateGiftEvent(ctx context.Context, giftID string, fromUserID, toUserID int64) (*db.GiftEvent, error) {
	event, err := r.db.CreateGiftEvent(ctx, db.CreateGiftEventParams{
		GiftID:     sql.NullString{String: giftID, Valid: true},
		FromUserID: sql.NullInt64{Int64: fromUserID, Valid: true},
		ToUserID:   sql.NullInt64{Int64: toUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *sqlRepository) GetGiftEvents(ctx context.Context, giftID string, limit int32, offset int32) ([]*db.GiftEvent, error) {
	events, err := r.db.GetGiftEvents(ctx, db.GetGiftEventsParams{
		GiftID: sql.NullString{String: giftID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	convertedEvents := make([]*db.GiftEvent, len(events))
	for i, event := range events {
		convertedEvents[i] = &event
	}
	return convertedEvents, nil
}

func (r *sqlRepository) CreateGift(ctx context.Context, gift *db.Gift) (*db.Gift, error) {
	g, err := r.db.CreateGift(ctx, db.CreateGiftParams{
		ID:               gift.ID,
		OwnerTelegramID:  gift.OwnerTelegramID,
		UpgradeMessageID: gift.UpgradeMessageID,
		TonPrice:         gift.TonPrice,
		CollectibleID:    int32(gift.CollectibleID),
		Status:           gift.Status,
		TelegramGiftID:   gift.TelegramGiftID,
		Title:            gift.Title,
		Slug:             gift.Slug,
		ImageUrl:         gift.ImageUrl,
		CreatedAt:        gift.CreatedAt,
		UpdatedAt:        gift.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return &g, nil
}

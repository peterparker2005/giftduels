package gift

import (
	"database/sql"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
)

// ConvertDBGiftToDomain converts db.Gift to domain Gift
func ConvertDBGiftToDomain(dbGift *db.Gift) *gift.Gift {
	return &gift.Gift{
		ID:               dbGift.ID,
		OwnerTelegramID:  dbGift.OwnerTelegramID,
		CollectibleID:    int64(dbGift.CollectibleID),
		TelegramGiftID:   dbGift.TelegramGiftID,
		Status:           ConvertDBStatusToDomain(dbGift.Status),
		Price:            dbGift.TonPrice,
		WithdrawnAt:      convertDBTimeToPtr(dbGift.WithdrawnAt),
		UpgradeMessageID: dbGift.UpgradeMessageID,
		CreatedAt:        dbGift.CreatedAt,
		UpdatedAt:        dbGift.UpdatedAt,
		Title:            dbGift.Title,
		Slug:             dbGift.Slug,
		ImageUrl:         dbGift.ImageUrl.String,
		Attributes:       []gift.Attribute{},
	}
}

// ConvertDBStatusToDomain converts db.GiftStatus to domain Status
func ConvertDBStatusToDomain(dbStatus db.GiftStatus) gift.Status {
	switch dbStatus {
	case db.GiftStatusOwned:
		return gift.StatusOwned
	case db.GiftStatusWithdrawPending:
		return gift.StatusPending
	case db.GiftStatusWithdrawn:
		return gift.StatusWithdrawn
	case db.GiftStatusInGame:
		return gift.StatusInGame
	default:
		return gift.StatusPending
	}
}

// convertDBTimeToPtr converts sql.NullTime to *time.Time
func convertDBTimeToPtr(nullTime sql.NullTime) *time.Time {
	if nullTime.Valid {
		return &nullTime.Time
	}
	return nil
}

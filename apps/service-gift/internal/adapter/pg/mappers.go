package pg

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
)

func GiftToDomain(u sqlc.Gift) *gift.Gift {
	return &gift.Gift{
		ID:               u.ID.String(),
		OwnerTelegramID:  u.OwnerTelegramID,
		Status:           gift.Status(u.Status),
		Price:            u.TonPrice,
		WithdrawnAt:      &u.WithdrawnAt.Time,
		Title:            u.Title,
		Slug:             u.Slug,
		CollectibleID:    int64(u.CollectibleID),
		UpgradeMessageID: int32(u.UpgradeMessageID),
		TelegramGiftID:   u.TelegramGiftID,
		// Attributes: TODO:,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func CreateGiftParamsToDB(gift *gift.CreateGiftParams) sqlc.CreateGiftParams {
	return sqlc.CreateGiftParams{
		ID:               mustPgUUID(gift.GiftID),
		CollectibleID:    int32(gift.CollectibleID),
		TonPrice:         gift.Price,
		TelegramGiftID:   gift.TelegramGiftID,
		Title:            gift.Title,
		Slug:             gift.Slug,
		OwnerTelegramID:  gift.OwnerTelegramID,
		UpgradeMessageID: gift.UpgradeMessageID,
		Status:           sqlc.GiftStatus(gift.Status),
		CreatedAt:        pgtype.Timestamptz{Time: gift.CreatedAt, Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: gift.UpdatedAt, Valid: true},
	}
}

func CreateGiftAttributeParamsToDB(attr *gift.CreateGiftAttributeParams) sqlc.CreateGiftAttributeParams {
	return sqlc.CreateGiftAttributeParams{
		TelegramGiftID: attr.TelegramGiftID,
		Type:           sqlc.GiftAttributeType(attr.AttributeType),
		Name:           attr.AttributeName,
		Rarity:         attr.AttributeRarity,
	}
}

func AttributeToDomain(attr sqlc.GiftAttribute) *gift.Attribute {
	return &gift.Attribute{
		Type:   gift.AttributeType(attr.Type),
		Name:   attr.Name,
		Rarity: attr.Rarity,
	}
}

func AttributeToCreateParams(telegramGiftID int64, attr *gift.Attribute) sqlc.CreateGiftAttributeParams {
	return sqlc.CreateGiftAttributeParams{
		TelegramGiftID: telegramGiftID,
		Type:           sqlc.GiftAttributeType(attr.Type),
		Name:           attr.Name,
		Rarity:         attr.Rarity,
	}
}

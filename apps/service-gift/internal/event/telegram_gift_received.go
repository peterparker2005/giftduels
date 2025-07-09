package event

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	giftdomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TelegramGiftReceivedHandler struct {
	repo   giftdomain.GiftRepository
	logger *logger.Logger
}

func NewTelegramGiftReceivedHandler(repo giftdomain.GiftRepository, logger *logger.Logger) *TelegramGiftReceivedHandler {
	return &TelegramGiftReceivedHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *TelegramGiftReceivedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()
	h.logger.Info("Unmarshalling event copy me", zap.Any("payload", msg.Payload))

	var ev giftv1.TelegramGiftReceivedEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error("Failed to unmarshal event", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if ev.OwnerTelegramId == nil {
		h.logger.Error("Missing OwnerTelegramId in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing OwnerTelegramId in event")
	}

	id := uuid.New().String()

	floorPriceTON := 0.5 // random it for now

	createGift := &giftdomain.CreateGiftParams{
		GiftID:           id,
		CollectibleID:    int64(ev.CollectibleId),
		Price:            floorPriceTON,
		TelegramGiftID:   ev.TelegramGiftId.Value,
		Title:            ev.Title,
		Slug:             ev.Slug,
		OwnerTelegramID:  ev.OwnerTelegramId.Value,
		UpgradeMessageID: ev.UpgradeMessageId,
		Status:           giftdomain.StatusOwned,
		EmojiID:          ev.EmojiId,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	createAttributes := make([]giftdomain.CreateGiftAttributeParams, 0, len(ev.Attributes))
	for _, attr := range ev.Attributes {
		attrType, err := giftdomain.AttributeTypeFromProto(attr.Type)
		if err != nil {
			h.logger.Error("Failed to get attribute type", zap.Error(err), zap.String("message_id", msg.UUID))
			return err
		}
		createAttributes = append(createAttributes, giftdomain.CreateGiftAttributeParams{
			GiftID:                  id,
			AttributeType:           attrType,
			AttributeName:           attr.Name,
			AttributeRarityPerMille: attr.RarityPerMille,
		})
	}

	_, err := h.repo.CreateGiftWithDetails(ctx, createGift, createAttributes) // make it in transaction
	if err != nil {
		h.logger.Error("Failed to save gift", zap.Error(err), zap.String("message_id", msg.UUID))
		return err
	}

	h.logger.Info("Gift processed successfully",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", id),
		zap.Float64("price_ton", floorPriceTON))

	return nil
}

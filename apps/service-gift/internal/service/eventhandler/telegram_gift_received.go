package eventhandler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	giftdomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TelegramGiftReceivedHandler struct {
	repo      giftdomain.GiftRepository
	publisher message.Publisher
	logger    *logger.Logger
}

func NewTelegramGiftReceivedHandler(repo giftdomain.GiftRepository, publisher message.Publisher, logger *logger.Logger) *TelegramGiftReceivedHandler {
	return &TelegramGiftReceivedHandler{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
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

	// Extract and create/find attributes
	collectionID, modelID, backdropID, symbolID, err := h.processAttributes(ctx, &ev)
	if err != nil {
		h.logger.Error("Failed to process attributes", zap.Error(err), zap.String("message_id", msg.UUID))
		return err
	}

	createGift := &giftdomain.CreateGiftParams{
		GiftID:           id,
		CollectibleID:    ev.CollectibleId,
		Price:            floorPriceTON,
		TelegramGiftID:   ev.TelegramGiftId.Value,
		Title:            ev.Title,
		Slug:             ev.Slug,
		OwnerTelegramID:  ev.OwnerTelegramId.Value,
		UpgradeMessageID: ev.UpgradeMessageId,
		Status:           giftdomain.StatusOwned,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = h.repo.CreateGift(ctx, createGift, collectionID, modelID, backdropID, symbolID)
	if err != nil {
		h.logger.Error("Failed to save gift", zap.Error(err), zap.String("message_id", msg.UUID))
		return err
	}

	h.logger.Info("Gift processed successfully",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", id),
		zap.Float64("price_ton", floorPriceTON))

	importedEvent := &giftv1.GiftDepositedEvent{
		GiftId:          &sharedv1.GiftId{Value: id},
		OwnerTelegramId: ev.OwnerTelegramId,
		TelegramGiftId:  ev.TelegramGiftId,
		Title:           ev.Title,
		Slug:            ev.Slug,
		CollectibleId:   ev.CollectibleId,
	}

	payload, err := proto.Marshal(importedEvent)
	if err != nil {
		h.logger.Error("marshal event failed", zap.Error(err))
		return err
	}
	importedMsg := message.NewMessage(watermill.NewUUID(), payload)

	err = h.publisher.Publish(giftEvents.TopicGiftDeposited.String(), importedMsg)
	if err != nil {
		h.logger.Warn("publish event failed", zap.Error(err))
		return nil
	}

	return nil
}

// processAttributes extracts attributes from the event and creates or finds them in the database
func (h *TelegramGiftReceivedHandler) processAttributes(ctx context.Context, ev *giftv1.TelegramGiftReceivedEvent) (collectionID, modelID, backdropID, symbolID int32, err error) {
	// For now, we'll use default values. In a real implementation,
	// you would extract these from ev.AttributesBackdrop, ev.AttributesModel, ev.AttributesSymbol

	// Default collection
	collection, err := h.repo.FindCollectionByName(ctx, ev.GetTitle())
	if err != nil {
		// Create default collection if not found
		collection, err = h.repo.CreateCollection(ctx, &giftdomain.CreateCollectionParams{
			Name:      ev.GetTitle(),
			ShortName: shortName(ev.GetTitle()),
		})
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("create default collection: %w", err)
		}
	}
	collectionID = collection.ID

	// Default model
	model, err := h.repo.FindModelByName(ctx, ev.GetModel().GetName())
	if err != nil {
		// Create default model if not found
		model, err = h.repo.CreateModel(ctx, &giftdomain.CreateModelParams{
			CollectionID:   collectionID,
			Name:           ev.GetModel().GetName(),
			ShortName:      shortName(ev.GetModel().GetName()),
			RarityPerMille: ev.GetModel().GetRarityPerMille(), // 100%
		})
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("create default model: %w", err)
		}
	}
	modelID = model.ID

	// Default backdrop
	backdrop, err := h.repo.FindBackdropByName(ctx, ev.GetBackdrop().GetName())
	if err != nil {
		// Create default backdrop if not found
		backdrop, err = h.repo.CreateBackdrop(ctx, &giftdomain.CreateBackdropParams{
			Name:           ev.GetBackdrop().GetName(),
			ShortName:      shortName(ev.GetBackdrop().GetName()),
			RarityPerMille: ev.GetBackdrop().GetRarityPerMille(), // 100%
			CenterColor:    stringPtr(ev.GetBackdrop().GetCenterColor()),
			EdgeColor:      stringPtr(ev.GetBackdrop().GetEdgeColor()),
			PatternColor:   stringPtr(ev.GetBackdrop().GetPatternColor()),
			TextColor:      stringPtr(ev.GetBackdrop().GetTextColor()),
		})
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("create default backdrop: %w", err)
		}
	}
	backdropID = backdrop.ID

	// Default symbol
	symbol, err := h.repo.FindSymbolByName(ctx, ev.GetSymbol().GetName())
	if err != nil {
		// Create default symbol if not found
		symbol, err = h.repo.CreateSymbol(ctx, &giftdomain.CreateSymbolParams{
			Name:           ev.GetSymbol().GetName(),
			ShortName:      shortName(ev.GetSymbol().GetName()),
			RarityPerMille: ev.GetSymbol().GetRarityPerMille(), // 100%
		})
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("create default symbol: %w", err)
		}
	}
	symbolID = symbol.ID

	return collectionID, modelID, backdropID, symbolID, nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func shortName(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

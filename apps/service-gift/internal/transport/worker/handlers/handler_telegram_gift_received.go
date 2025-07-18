package workerhandlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/saga"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TelegramGiftReceivedHandler struct {
	withdrawalSaga   *saga.WithdrawalSaga
	giftEventHandler *command.GiftEventHandler
	priceService     pricing.PriceService
	publisher        message.Publisher
	logger           *logger.Logger
}

func NewTelegramGiftReceivedHandler(
	withdrawalSaga *saga.WithdrawalSaga,
	giftEventHandler *command.GiftEventHandler,
	priceService pricing.PriceService,
	publisher message.Publisher,
	logger *logger.Logger,
) *TelegramGiftReceivedHandler {
	return &TelegramGiftReceivedHandler{
		withdrawalSaga:   withdrawalSaga,
		giftEventHandler: giftEventHandler,
		priceService:     priceService,
		publisher:        publisher,
		logger:           logger,
	}
}

func (h *TelegramGiftReceivedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()
	h.logger.Info("Processing Telegram gift received event", zap.String("message_id", msg.UUID))

	// Parse event
	ev, err := h.parseEvent(msg)
	if err != nil {
		return fmt.Errorf("parse event: %w", err)
	}

	// Validate event
	if err = h.validateEvent(ev); err != nil {
		return fmt.Errorf("validate event: %w", err)
	}

	// Generate gift ID
	giftID := uuid.New().String()

	// Get price
	price, err := h.getGiftPrice(ctx, ev)
	if err != nil {
		return fmt.Errorf("get gift price: %w", err)
	}

	// Process attributes
	attrs, err := h.giftEventHandler.ProcessAttributesFromEvent(ctx, ev)
	if err != nil {
		return fmt.Errorf("process attributes: %w", err)
	}

	// Create gift
	if err = h.createGift(ctx, giftID, ev, price, attrs); err != nil {
		return fmt.Errorf("create gift: %w", err)
	}

	// Publish event
	if err = h.publishGiftDepositedEvent(ev, giftID); err != nil {
		h.logger.Warn("Failed to publish gift deposited event", zap.Error(err))
		// Don't return error here as the main operation succeeded
	}

	h.logger.Info("Gift processed successfully",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", giftID),
		zap.String("price_ton", price.String()))

	return nil
}

func (h *TelegramGiftReceivedHandler) parseEvent(
	msg *message.Message,
) (*giftv1.TelegramGiftReceivedEvent, error) {
	var ev giftv1.TelegramGiftReceivedEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error(
			"Failed to unmarshal event",
			zap.Error(err),
			zap.String("message_id", msg.UUID),
		)
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}
	return &ev, nil
}

func (h *TelegramGiftReceivedHandler) validateEvent(ev *giftv1.TelegramGiftReceivedEvent) error {
	if ev.GetOwnerTelegramId() == nil {
		return errors.New("missing OwnerTelegramId in event")
	}
	return nil
}

func (h *TelegramGiftReceivedHandler) getGiftPrice(
	ctx context.Context,
	ev *giftv1.TelegramGiftReceivedEvent,
) (*tonamount.TonAmount, error) {
	priceParams := &pricing.PriceServiceParams{
		Collection: ev.GetTitle(),
		Model:      ev.GetModel().GetName(),
		Symbol:     ev.GetSymbol().GetName(),
		Backdrop:   ev.GetBackdrop().GetName(),
	}
	priceResult, err := h.priceService.GetFloorPrice(ctx, priceParams)
	if err != nil {
		h.logger.Error(
			"Failed to get floor price",
			zap.Error(err),
		)
		// return nil, fmt.Errorf("get floor price: %w", err)
		amount, err := tonamount.NewTonAmountFromString("1")
		if err != nil {
			return nil, fmt.Errorf("new ton amount: %w", err)
		}
		return amount, nil
	}
	return tonamount.NewTonAmountFromString(priceResult.Price)
}

func (h *TelegramGiftReceivedHandler) createGift(
	ctx context.Context,
	giftID string,
	ev *giftv1.TelegramGiftReceivedEvent,
	price *tonamount.TonAmount,
	attrs *gift.AttributeData,
) error {
	_, err := h.giftEventHandler.CreateGiftFromEvent(ctx, giftID, ev, price, attrs)
	if err != nil {
		h.logger.Error("Failed to save gift", zap.Error(err))
		return fmt.Errorf("save gift: %w", err)
	}
	return nil
}

func (h *TelegramGiftReceivedHandler) publishGiftDepositedEvent(
	ev *giftv1.TelegramGiftReceivedEvent,
	giftID string,
) error {
	importedEvent := &giftv1.GiftDepositedEvent{
		GiftId:          &sharedv1.GiftId{Value: giftID},
		OwnerTelegramId: ev.GetOwnerTelegramId(),
		TelegramGiftId:  ev.GetTelegramGiftId(),
		Title:           ev.GetTitle(),
		Slug:            ev.GetSlug(),
		CollectibleId:   ev.GetCollectibleId(),
	}

	payload, err := proto.Marshal(importedEvent)
	if err != nil {
		h.logger.Error("Failed to marshal event", zap.Error(err))
		return fmt.Errorf("marshal event: %w", err)
	}

	importedMsg := message.NewMessage(watermill.NewUUID(), payload)
	return h.publisher.Publish(giftEvents.TopicGiftDeposited.String(), importedMsg)
}

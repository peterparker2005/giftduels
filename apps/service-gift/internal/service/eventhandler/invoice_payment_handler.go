package eventhandler

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type InvoicePaymentHandler struct {
	giftService *gift.Service
	logger      *logger.Logger
}

func NewInvoicePaymentHandler(giftService *gift.Service, logger *logger.Logger) *InvoicePaymentHandler {
	return &InvoicePaymentHandler{
		giftService: giftService,
		logger:      logger,
	}
}

func (h *InvoicePaymentHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var event telegrambotv1.InvoicePaymentEvent
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error("failed to unmarshal InvoicePaymentEvent", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if event.TelegramUserId == nil {
		h.logger.Error("missing TelegramUserId in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing TelegramUserId in event")
	}

	if event.StarsAmount == nil {
		h.logger.Error("missing StarsAmount in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing StarsAmount in event")
	}

	if event.Payload == nil {
		h.logger.Error("missing Payload in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing Payload in event")
	}

	telegramUserID := event.TelegramUserId.Value
	starsAmount := event.StarsAmount.Value

	log := h.logger.With(
		zap.Int64("telegram_user_id", telegramUserID),
		zap.Uint32("stars_amount", starsAmount),
		zap.String("invoice_id", event.InvoiceId),
		zap.String("message_id", msg.UUID),
	)

	log.Info("processing invoice payment event")

	// Проверяем что это оплата за gift withdrawal commission
	giftWithdrawCommission := event.Payload.GetGiftWithdrawCommission()
	if giftWithdrawCommission == nil {
		log.Warn("invoice payment is not for gift withdrawal commission, ignoring")
		return nil
	}

	// Извлекаем gift IDs из payload
	giftIDs := make([]string, len(giftWithdrawCommission.GiftCommissions))
	for i, commission := range giftWithdrawCommission.GiftCommissions {
		if commission.GiftId == nil {
			log.Error("missing GiftId in commission", zap.Int("commission_index", i))
			return fmt.Errorf("missing GiftId in commission at index %d", i)
		}
		giftIDs[i] = commission.GiftId.Value
	}

	log.Info("completing stars withdrawal", zap.Strings("gift_ids", giftIDs))

	// Завершаем Stars withdrawal
	gifts, err := h.giftService.CompleteStarsWithdrawal(ctx, telegramUserID, giftIDs, starsAmount)
	if err != nil {
		log.Error("failed to complete stars withdrawal", zap.Error(err))
		return fmt.Errorf("complete stars withdrawal: %w", err)
	}

	log.Info("stars withdrawal completed successfully",
		zap.Int("gifts_count", len(gifts)),
		zap.Strings("gift_ids", giftIDs))

	return nil
}

package eventhandler

import (
	"context"
	"errors"
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

func NewInvoicePaymentHandler(
	giftService *gift.Service,
	logger *logger.Logger,
) *InvoicePaymentHandler {
	return &InvoicePaymentHandler{
		giftService: giftService,
		logger:      logger,
	}
}

func (h *InvoicePaymentHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var event telegrambotv1.InvoicePaymentEvent
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error(
			"failed to unmarshal InvoicePaymentEvent",
			zap.Error(err),
			zap.String("message_id", msg.UUID),
		)
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if event.GetTelegramUserId() == nil {
		h.logger.Error("missing TelegramUserId in event", zap.String("message_id", msg.UUID))
		return errors.New("missing TelegramUserId in event")
	}

	if event.GetStarsAmount() == nil {
		h.logger.Error("missing StarsAmount in event", zap.String("message_id", msg.UUID))
		return errors.New("missing StarsAmount in event")
	}

	if event.GetPayload() == nil {
		h.logger.Error("missing Payload in event", zap.String("message_id", msg.UUID))
		return errors.New("missing Payload in event")
	}

	telegramUserID := event.GetTelegramUserId().GetValue()
	starsAmount := event.GetStarsAmount().GetValue()

	log := h.logger.With(
		zap.Int64("telegram_user_id", telegramUserID),
		zap.Uint32("stars_amount", starsAmount),
		zap.String("invoice_id", event.GetInvoiceId()),
		zap.String("message_id", msg.UUID),
	)

	log.Info("processing invoice payment event")

	// Проверяем что это оплата за gift withdrawal commission
	giftWithdrawCommission := event.GetPayload().GetGiftWithdrawCommission()
	if giftWithdrawCommission == nil {
		log.Warn("invoice payment is not for gift withdrawal commission, ignoring")
		return nil
	}

	// Извлекаем gift IDs из payload
	giftIDs := make([]string, len(giftWithdrawCommission.GetGiftCommissions()))
	for i, commission := range giftWithdrawCommission.GetGiftCommissions() {
		if commission.GetGiftId() == nil {
			log.Error("missing GiftId in commission", zap.Int("commission_index", i))
			return fmt.Errorf("missing GiftId in commission at index %d", i)
		}
		giftIDs[i] = commission.GetGiftId().GetValue()
	}

	log.Info("completing stars withdrawal", zap.Strings("gift_ids", giftIDs))

	// Завершаем Stars withdrawal
	gifts, err := h.giftService.CompleteStarsWithdrawal(ctx, giftIDs)
	if err != nil {
		log.Error("failed to complete stars withdrawal", zap.Error(err))
		return fmt.Errorf("complete stars withdrawal: %w", err)
	}

	log.Info("stars withdrawal completed successfully",
		zap.Int("gifts_count", len(gifts)),
		zap.Strings("gift_ids", giftIDs))

	return nil
}

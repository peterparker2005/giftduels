package workerhandlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	giftdomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type GiftWithdrawFailedHandler struct {
	repo   giftdomain.Repository
	logger *logger.Logger
}

func NewGiftWithdrawFailedHandler(
	repo giftdomain.Repository,
	logger *logger.Logger,
) *GiftWithdrawFailedHandler {
	return &GiftWithdrawFailedHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *GiftWithdrawFailedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var ev giftv1.GiftWithdrawFailedEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error(
			"Failed to unmarshal GiftWithdrawFailedEvent",
			zap.Error(err),
			zap.String("message_id", msg.UUID),
		)
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if ev.GetGiftId() == nil || ev.GetGiftId().GetValue() == "" {
		h.logger.Error("Missing GiftId in event", zap.String("message_id", msg.UUID))
		return errors.New("missing GiftId in event")
	}

	if ev.GetOwnerTelegramId() == nil {
		h.logger.Error("Missing OwnerTelegramId in event", zap.String("message_id", msg.UUID))
		return errors.New("missing OwnerTelegramId in event")
	}

	log := h.logger.With(
		zap.String("gift_id", ev.GetGiftId().GetValue()),
		zap.Int64("owner_telegram_id", ev.GetOwnerTelegramId().GetValue()),
		zap.String("error_reason", ev.GetErrorReason()),
		zap.Int32("attempts_made", ev.GetAttemptsMade()),
		zap.String("message_id", msg.UUID),
	)

	log.Info("Processing gift withdrawal failure, rolling back status")

	// Отменяем вывод подарка - возвращаем статус с withdraw_pending на owned
	gift, err := h.repo.CancelGiftWithdrawal(ctx, ev.GetGiftId().GetValue())
	if err != nil {
		log.Error("Failed to cancel gift withdrawal", zap.Error(err))

		// TODO: Опубликовать событие о неудачном rollback для дальнейшей компенсации
		// Это критическая ошибка - подарок остается в withdraw_pending, но комиссия может быть возвращена
		log.Error("CRITICAL: Gift status rollback failed - requires manual intervention",
			zap.String("gift_id", ev.GetGiftId().GetValue()),
			zap.Int64("owner_telegram_id", ev.GetOwnerTelegramId().GetValue()),
			zap.String("original_error", ev.GetErrorReason()),
		)

		return fmt.Errorf("cancel gift withdrawal: %w", err)
	}

	log.Info("Gift withdrawal successfully cancelled, status reverted to owned",
		zap.String("new_status", string(gift.Status)),
		zap.String("price", gift.Price.String()))

	return nil
}

package eventhandler

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	giftdomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type GiftWithdrawFailedHandler struct {
	repo   giftdomain.GiftRepository
	logger *logger.Logger
}

func NewGiftWithdrawFailedHandler(repo giftdomain.GiftRepository, logger *logger.Logger) *GiftWithdrawFailedHandler {
	return &GiftWithdrawFailedHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *GiftWithdrawFailedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var ev giftv1.GiftWithdrawFailedEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error("Failed to unmarshal GiftWithdrawFailedEvent", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if ev.GiftId == nil || ev.GiftId.Value == "" {
		h.logger.Error("Missing GiftId in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing GiftId in event")
	}

	if ev.OwnerTelegramId == nil {
		h.logger.Error("Missing OwnerTelegramId in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing OwnerTelegramId in event")
	}

	log := h.logger.With(
		zap.String("gift_id", ev.GiftId.Value),
		zap.Int64("owner_telegram_id", ev.OwnerTelegramId.Value),
		zap.String("error_reason", ev.ErrorReason),
		zap.Int32("attempts_made", ev.AttemptsMade),
		zap.String("message_id", msg.UUID),
	)

	log.Info("Processing gift withdrawal failure, rolling back status")

	// Отменяем вывод подарка - возвращаем статус с withdraw_pending на owned
	gift, err := h.repo.CancelGiftWithdrawal(ctx, ev.GiftId.Value)
	if err != nil {
		log.Error("Failed to cancel gift withdrawal", zap.Error(err))
		return fmt.Errorf("cancel gift withdrawal: %w", err)
	}

	log.Info("Gift withdrawal successfully cancelled, status reverted to owned",
		zap.String("new_status", string(gift.Status)),
		zap.Float64("price", gift.Price))

	return nil
}

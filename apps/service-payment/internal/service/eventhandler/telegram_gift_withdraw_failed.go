package eventhandler

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	paymentDomain "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TelegramGiftWithdrawFailedHandler struct {
	paymentService *payment.Service
	log            *logger.Logger
}

func NewTelegramGiftWithdrawFailedHandler(paymentService *payment.Service, log *logger.Logger) *TelegramGiftWithdrawFailedHandler {
	return &TelegramGiftWithdrawFailedHandler{
		paymentService: paymentService,
		log:            log,
	}
}

func (h *TelegramGiftWithdrawFailedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var event giftv1.GiftWithdrawFailedEvent
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.log.Error("failed to unmarshal GiftWithdrawFailedEvent", zap.Error(err))
		return err
	}

	telegramUserID := event.OwnerTelegramId.GetValue()
	commissionAmount := event.CommissionAmount.GetValue()

	h.log.Info("processing gift withdrawal failure, rolling back commission",
		zap.Int64("telegram_user_id", telegramUserID),
		zap.Float64("commission_amount", commissionAmount),
		zap.String("gift_id", event.GiftId.GetValue()),
		zap.String("error_reason", event.ErrorReason),
	)

	metadata := paymentDomain.TransactionMetadata{
		Gift: &paymentDomain.TransactionMetadata_GiftDetails{
			GiftID: event.GiftId.GetValue(),
			Title:  event.Title,
			Slug:   event.Slug,
		},
	}

	err := h.paymentService.RollbackWithdrawalCommission(ctx, telegramUserID, commissionAmount, metadata)
	if err != nil {
		h.log.Error("failed to rollback withdrawal commission",
			zap.Error(err),
			zap.Int64("telegram_user_id", telegramUserID),
			zap.Float64("commission_amount", commissionAmount),
		)

		// TODO: Опубликовать событие о неудачном rollback комиссии для компенсации
		// Это критическая ошибка - комиссия не возвращена, но статус подарка может быть сброшен
		h.log.Error("CRITICAL: Commission rollback failed - requires manual intervention",
			zap.Int64("telegram_user_id", telegramUserID),
			zap.Float64("commission_amount", commissionAmount),
			zap.String("gift_id", event.GiftId.GetValue()),
			zap.String("original_error", event.ErrorReason),
		)

		return err
	}

	h.log.Info("successfully rolled back withdrawal commission",
		zap.Int64("telegram_user_id", telegramUserID),
		zap.Float64("commission_amount", commissionAmount),
		zap.String("gift_id", event.GiftId.GetValue()),
	)

	return nil
}

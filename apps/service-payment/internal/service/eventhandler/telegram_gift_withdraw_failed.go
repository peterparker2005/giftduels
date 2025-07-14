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

func NewTelegramGiftWithdrawFailedHandler(
	paymentService *payment.Service,
	log *logger.Logger,
) *TelegramGiftWithdrawFailedHandler {
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

	telegramUserID := event.GetOwnerTelegramId().GetValue()
	commissionAmount := event.GetCommissionAmount().GetValue()

	h.log.Info("processing gift withdrawal failure, rolling back commission",
		zap.Int64("telegram_user_id", telegramUserID),
		zap.String("commission_amount", commissionAmount),
		zap.String("gift_id", event.GetGiftId().GetValue()),
		zap.String("error_reason", event.GetErrorReason()),
	)

	metadata := paymentDomain.TransactionMetadata{
		Gift: &paymentDomain.TransactionMetadataGiftDetails{
			GiftID: event.GetGiftId().GetValue(),
			Title:  event.GetTitle(),
			Slug:   event.GetSlug(),
		},
	}

	err := h.paymentService.RollbackWithdrawalCommission(
		ctx,
		telegramUserID,
		commissionAmount,
		metadata,
	)
	if err != nil {
		h.log.Error("failed to rollback withdrawal commission",
			zap.Error(err),
			zap.Int64("telegram_user_id", telegramUserID),
			zap.String("commission_amount", commissionAmount),
		)

		// TODO: Опубликовать событие о неудачном rollback комиссии для компенсации
		// Это критическая ошибка - комиссия не возвращена, но статус подарка может быть сброшен
		h.log.Error("CRITICAL: Commission rollback failed - requires manual intervention",
			zap.Int64("telegram_user_id", telegramUserID),
			zap.String("commission_amount", commissionAmount),
			zap.String("gift_id", event.GetGiftId().GetValue()),
			zap.String("original_error", event.GetErrorReason()),
		)

		return err
	}

	h.log.Info("successfully rolled back withdrawal commission",
		zap.Int64("telegram_user_id", telegramUserID),
		zap.String("commission_amount", commissionAmount),
		zap.String("gift_id", event.GetGiftId().GetValue()),
	)

	return nil
}

package eventhandler

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
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

	err := h.paymentService.RollbackWithdrawalCommission(ctx, telegramUserID, commissionAmount)
	if err != nil {
		h.log.Error("failed to rollback withdrawal commission",
			zap.Error(err),
			zap.Int64("telegram_user_id", telegramUserID),
			zap.Float64("commission_amount", commissionAmount),
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

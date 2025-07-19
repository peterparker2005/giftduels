package workerhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type GiftWithdrawnHandler struct {
	logger                    *logger.Logger
	giftCompleteWithdrawalCmd *command.GiftCompleteWithdrawalCommand
}

func NewGiftWithdrawnHandler(
	logger *logger.Logger,
	giftCompleteWithdrawalCmd *command.GiftCompleteWithdrawalCommand,
) *GiftWithdrawnHandler {
	return &GiftWithdrawnHandler{
		logger:                    logger,
		giftCompleteWithdrawalCmd: giftCompleteWithdrawalCmd,
	}
}

func (h *GiftWithdrawnHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var event giftv1.GiftWithdrawnEvent
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error("failed to unmarshal GiftWithdrawnEvent", zap.Error(err))
		return err
	}

	giftID := event.GetGiftId().GetValue()
	ownerTelegramID := event.GetOwnerTelegramId().GetValue()

	log := h.logger.With(
		zap.String("gift_id", giftID),
		zap.Int64("owner_telegram_id", ownerTelegramID),
		zap.String("message_id", msg.UUID),
	)

	log.Info("processing gift withdrawn event")

	// Завершаем вывод подарка
	completedGift, err := h.giftCompleteWithdrawalCmd.Execute(
		ctx,
		command.CompleteGiftWithdrawalParams{
			GiftID: giftID,
		},
	)
	if err != nil {
		log.Error("failed to complete gift withdrawal", zap.Error(err))
		return err
	}

	log.Info("gift withdrawal completed successfully",
		zap.String("gift_id", completedGift.ID),
		zap.String("status", string(completedGift.Status)),
		zap.Time("withdrawn_at", *completedGift.WithdrawnAt))

	return nil
}

package workerhandlers

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type GiftReturnedHandler struct {
	giftStakeCommand *command.GiftStakeCommand
	logger           *logger.Logger
}

func NewGiftReturnedHandler(
	giftStakeCommand *command.GiftStakeCommand,
	logger *logger.Logger,
) *GiftReturnedHandler {
	return &GiftReturnedHandler{
		giftStakeCommand: giftStakeCommand,
		logger:           logger,
	}
}

func (h *GiftReturnedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()
	giftID := string(msg.Payload)

	h.logger.Info("Processing gift returned event",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", giftID))

	// Return the gift from game status back to owned status
	_, err := h.giftStakeCommand.ReturnGiftFromGame(ctx, giftID)
	if err != nil {
		h.logger.Error("Failed to return gift from game",
			zap.String("message_id", msg.UUID),
			zap.String("gift_id", giftID),
			zap.Error(err))
		return fmt.Errorf("return gift from game: %w", err)
	}

	h.logger.Info("Gift returned from game successfully",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", giftID))

	return nil
}

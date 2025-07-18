package workerhandlers

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type DuelCompletedHandler struct {
	logger                    *logger.Logger
	giftReturnFromGameCommand *command.GiftReturnFromGameCommand
}

func NewDuelCompletedHandler(
	giftReturnFromGameCommand *command.GiftReturnFromGameCommand,
	logger *logger.Logger,
) *DuelCompletedHandler {
	return &DuelCompletedHandler{
		giftReturnFromGameCommand: giftReturnFromGameCommand,
		logger:                    logger,
	}
}

func (h *DuelCompletedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()

	var event duelv1.DuelCompletedEvent
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error("failed to unmarshal DuelCompletedEvent", zap.Error(err))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Get winner ID from the event
	winnerID := event.GetWinnerTelegramUserId().GetValue()
	duelID := event.GetDuelId().GetValue()

	h.logger.Info("processing duel completed event",
		zap.String("duelID", duelID),
		zap.Int64("winnerID", winnerID),
		zap.Int("stakesCount", len(event.GetStakes())))

	// Process all stakes and transfer gifts to the winner
	for _, stake := range event.GetStakes() {
		giftID := stake.GetGift().GetGiftId().GetValue()
		stakeOwnerID := stake.GetParticipantTelegramUserId().GetValue()

		// Skip if the gift already belongs to the winner
		if stakeOwnerID == winnerID {
			h.logger.Info("gift already belongs to winner, skipping transfer",
				zap.String("giftID", giftID),
				zap.Int64("winnerID", winnerID),
				zap.String("duelID", duelID))
			continue
		}

		h.logger.Info("transferring gift to winner",
			zap.String("giftID", giftID),
			zap.Int64("winnerID", winnerID),
			zap.Int64("previousOwnerID", stakeOwnerID),
			zap.String("duelID", duelID))

		// Execute the command to transfer gift ownership to the winner
		err := h.giftReturnFromGameCommand.Execute(ctx, command.GiftReturnFromGameCommandParams{
			GiftID:        giftID,
			OwnerID:       winnerID,
			RelatedGameID: duelID,
		})
		if err != nil {
			h.logger.Error("failed to transfer gift to winner",
				zap.String("giftID", giftID),
				zap.Int64("winnerID", winnerID),
				zap.Int64("previousOwnerID", stakeOwnerID),
				zap.String("duelID", duelID),
				zap.Error(err))
			return fmt.Errorf("transfer gift %s to winner %d: %w", giftID, winnerID, err)
		}

		h.logger.Info("successfully transferred gift to winner",
			zap.String("giftID", giftID),
			zap.Int64("winnerID", winnerID),
			zap.Int64("previousOwnerID", stakeOwnerID))
	}

	h.logger.Info("completed processing duel completed event",
		zap.String("duelID", duelID),
		zap.Int64("winnerID", winnerID),
		zap.Int("transferredGiftsCount", len(event.GetStakes())))

	return nil
}

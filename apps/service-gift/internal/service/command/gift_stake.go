package command

import (
	"context"

	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
)

type GiftStakeCommand struct {
	repo giftDomain.Repository
	log  *logger.Logger
}

func NewGiftStakeCommand(
	repo giftDomain.Repository,
	log *logger.Logger,
) *GiftStakeCommand {
	return &GiftStakeCommand{
		repo: repo,
		log:  log,
	}
}

type StakeGiftParams struct {
	GiftID         string
	TelegramUserID int64
	GameMetadata   *giftv1.StakeGiftRequest_DuelMetadata
}

func (c *GiftStakeCommand) StakeGift(
	ctx context.Context,
	params StakeGiftParams,
) (*giftDomain.Gift, error) {
	// First, try to get the gift to check if it exists and its current status
	gift, err := c.repo.GetGiftByID(ctx, params.GiftID)
	if err != nil {
		c.log.Error("failed to get gift for staking",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, giftDomain.ErrGiftNotFound
	}

	// Check if the gift can be staked by the user
	if !gift.IsOwnedBy(params.TelegramUserID) {
		c.log.Error("gift cannot be staked - not owned by user",
			zap.String("giftID", params.GiftID),
			zap.Int64("requestedByUser", params.TelegramUserID),
			zap.Int64("actualOwner", gift.OwnerTelegramID))
		return nil, giftDomain.ErrGiftNotOwned
	}

	// Check if the gift can be staked (status check)
	if gift.Status != giftDomain.StatusOwned {
		c.log.Error("gift cannot be staked - wrong status",
			zap.String("giftID", params.GiftID),
			zap.String("currentStatus", string(gift.Status)),
			zap.String("requiredStatus", string(giftDomain.StatusOwned)))
		return nil, giftDomain.ErrGiftNotOwned
	}

	// Now try to stake the gift
	stakedGift, err := c.repo.StakeGiftForGame(ctx, params.GiftID)
	if err != nil {
		c.log.Error("failed to stake gift for game",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, err
	}

	_, err = c.repo.CreateGiftEvent(ctx, giftDomain.CreateGiftEventParams{
		GiftID:         stakedGift.ID,
		RelatedGameID:  &params.GameMetadata.GetDuelId().Value,
		EventType:      giftDomain.EventTypeStake,
		TelegramUserID: params.TelegramUserID,
	})
	if err != nil {
		c.log.Error("failed to create gift event",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, err
	}

	return stakedGift, nil
}

// ReturnGiftFromGame returns a gift from in_game status back to owned status.
func (c *GiftStakeCommand) ReturnGiftFromGame(
	ctx context.Context,
	giftID string,
) (*giftDomain.Gift, error) {
	// First, try to get the gift to check if it exists and its current status
	gift, err := c.repo.GetGiftByID(ctx, giftID)
	if err != nil {
		c.log.Error("failed to get gift for returning from game",
			zap.String("giftID", giftID),
			zap.Error(err))
		return nil, giftDomain.ErrGiftNotFound
	}

	// Check if the gift is in game status
	if gift.Status != giftDomain.StatusInGame {
		c.log.Error("gift cannot be returned from game - wrong status",
			zap.String("giftID", giftID),
			zap.String("currentStatus", string(gift.Status)),
			zap.String("requiredStatus", string(giftDomain.StatusInGame)))
		return nil, giftDomain.ErrGiftNotOwned
	}

	// Return the gift to owned status
	returnedGift, err := c.repo.ReturnGiftFromGame(ctx, giftID)
	if err != nil {
		c.log.Error("failed to return gift from game",
			zap.String("giftID", giftID),
			zap.Error(err))
		return nil, err
	}
	return returnedGift, nil
}

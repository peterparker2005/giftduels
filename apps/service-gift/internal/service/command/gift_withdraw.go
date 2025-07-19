package command

import (
	"context"

	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type GiftWithdrawCommand struct {
	repo giftDomain.Repository
	log  *logger.Logger
}

func NewGiftWithdrawCommand(
	repo giftDomain.Repository,
	log *logger.Logger,
) *GiftWithdrawCommand {
	return &GiftWithdrawCommand{
		repo: repo,
		log:  log,
	}
}

func (c *GiftWithdrawCommand) ValidateGiftsOwnership(
	gifts []*giftDomain.Gift,
	telegramUserID int64,
) error {
	for _, g := range gifts {
		if !g.CanBeWithdrawnBy(telegramUserID) {
			c.log.Error("gift cannot be withdrawn by user",
				zap.String("giftID", g.ID),
				zap.Int64("userID", telegramUserID),
				zap.Int64("ownerID", g.OwnerTelegramID),
				zap.String("status", string(g.Status)),
			)
			return giftDomain.ErrGiftNotOwned
		}
	}
	return nil
}

func (c *GiftWithdrawCommand) MarkGiftForWithdrawal(
	ctx context.Context,
	giftID string,
) (*giftDomain.Gift, error) {
	gift, err := c.repo.MarkGiftForWithdrawal(ctx, giftID)
	if err != nil {
		c.log.Error("failed to mark gift for withdrawal", zap.Error(err))
		return nil, err
	}
	return gift, nil
}

func (c *GiftWithdrawCommand) GetAndValidateGiftsForWithdrawal(
	ctx context.Context,
	giftIDs []string,
	telegramUserID int64,
) ([]*giftDomain.Gift, error) {
	gifts, err := c.repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		c.log.Error("failed to get gifts by IDs", zap.Error(err))
		return nil, err
	}

	// Используем domain методы для валидации
	for _, g := range gifts {
		if !g.CanBeWithdrawnBy(telegramUserID) {
			c.log.Error("gift cannot be withdrawn by user",
				zap.String("giftID", g.ID),
				zap.Int64("userID", telegramUserID),
				zap.Int64("ownerID", g.OwnerTelegramID),
				zap.String("status", string(g.Status)),
			)
			return nil, giftDomain.ErrGiftNotOwned
		}
	}

	return gifts, nil
}

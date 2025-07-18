package command

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	giftdomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type GiftReturnFromGameCommand struct {
	logger    *logger.Logger
	repo      giftdomain.Repository
	txManager pg.TxManager
}

func NewGiftReturnFromGameCommand(
	logger *logger.Logger,
	repo giftdomain.Repository,
	txManager pg.TxManager,
) *GiftReturnFromGameCommand {
	return &GiftReturnFromGameCommand{
		logger:    logger,
		repo:      repo,
		txManager: txManager,
	}
}

type GiftReturnFromGameCommandParams struct {
	GiftID        string
	OwnerID       int64
	RelatedGameID string
}

func (c *GiftReturnFromGameCommand) Execute(
	ctx context.Context,
	params GiftReturnFromGameCommandParams,
) error {
	tx, err := c.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				c.logger.Error("failed to rollback tx", zap.Error(err))
			}
		}
	}()
	repo := c.repo.WithTx(tx)
	if err = repo.UpdateGiftOwner(ctx, params.GiftID, params.OwnerID); err != nil {
		c.logger.Error("failed to update gift owner", zap.Error(err))
		return err
	}
	_, err = repo.CreateGiftEvent(ctx, giftdomain.CreateGiftEventParams{
		GiftID:         params.GiftID,
		TelegramUserID: params.OwnerID,
		EventType:      giftdomain.EventTypeReturnFromGame,
		RelatedGameID:  &params.RelatedGameID,
	})
	if err != nil {
		c.logger.Error("failed to create gift event", zap.Error(err))
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		c.logger.Error("failed to commit tx", zap.Error(err))
		return err
	}
	return nil
}

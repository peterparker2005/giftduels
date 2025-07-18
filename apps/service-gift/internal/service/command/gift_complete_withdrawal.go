package command

import (
	"context"
	"time"

	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type GiftCompleteWithdrawalCommand struct {
	repo giftDomain.Repository
	log  *logger.Logger
}

func NewGiftCompleteWithdrawalCommand(
	repo giftDomain.Repository,
	log *logger.Logger,
) *GiftCompleteWithdrawalCommand {
	return &GiftCompleteWithdrawalCommand{
		repo: repo,
		log:  log,
	}
}

type CompleteGiftWithdrawalParams struct {
	GiftID string
}

func (c *GiftCompleteWithdrawalCommand) Execute(
	ctx context.Context,
	params CompleteGiftWithdrawalParams,
) (*giftDomain.Gift, error) {
	// Получаем подарок для валидации
	gift, err := c.repo.GetGiftByID(ctx, params.GiftID)
	if err != nil {
		c.log.Error("failed to get gift for withdrawal completion",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, errors.NewNotFoundError("gift not found: " + params.GiftID)
	}

	// Проверяем, что подарок находится в статусе withdraw_pending
	if gift.Status != giftDomain.StatusWithdrawPending {
		c.log.Error("gift cannot be completed - wrong status",
			zap.String("giftID", params.GiftID),
			zap.String("currentStatus", string(gift.Status)),
			zap.String("requiredStatus", string(giftDomain.StatusWithdrawPending)))
		return nil, errors.NewValidationError(
			"gift status",
			"gift must be in withdraw_pending status to complete withdrawal",
		)
	}

	// Используем domain метод для валидации
	if err := gift.CompleteWithdrawal(time.Now()); err != nil {
		c.log.Error("failed to complete withdrawal in domain",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, err
	}

	// Обновляем статус в базе данных
	completedGift, err := c.repo.CompleteGiftWithdrawal(ctx, params.GiftID)
	if err != nil {
		c.log.Error("failed to complete gift withdrawal in repository",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		return nil, err
	}

	// Создаем событие о завершении вывода
	_, err = c.repo.CreateGiftEvent(ctx, giftDomain.CreateGiftEventParams{
		GiftID:         completedGift.ID,
		TelegramUserID: completedGift.OwnerTelegramID,
		EventType:      giftDomain.EventTypeWithdrawComplete,
	})
	if err != nil {
		c.log.Error("failed to create gift withdrawal complete event",
			zap.String("giftID", params.GiftID),
			zap.Error(err))
		// Не возвращаем ошибку, так как основная операция завершена успешно
	}

	c.log.Info("gift withdrawal completed successfully",
		zap.String("giftID", params.GiftID),
		zap.String("newStatus", string(completedGift.Status)),
		zap.Time("withdrawnAt", *completedGift.WithdrawnAt))

	return completedGift, nil
}

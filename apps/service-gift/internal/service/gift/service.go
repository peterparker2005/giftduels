package gift

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type Service struct {
	repo gift.GiftRepository
}

func New(repo gift.GiftRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetGiftByID(ctx context.Context, id string) (*gift.Gift, error) {
	gift, err := s.repo.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

func (s *Service) GetUserGifts(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) ([]*gift.Gift, error) {
	gifts, err := s.repo.GetUserGifts(ctx, pagination.PageSize(), pagination.Offset(), telegramUserID)
	if err != nil {
		return nil, err
	}

	return gifts, nil
}

func (s *Service) StakeGift(ctx context.Context, giftID string) (*gift.Gift, error) {
	gift, err := s.repo.StakeGiftForGame(ctx, giftID)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

func (s *Service) TransferGiftToUser(ctx context.Context, giftID string, telegramUserID int64) (*gift.Gift, error) {
	// First update the gift owner
	gift, err := s.repo.UpdateGiftOwner(ctx, giftID, telegramUserID)
	if err != nil {
		return nil, err
	}

	// Create transfer record
	_, err = s.repo.CreateGiftEvent(ctx, giftID, gift.OwnerTelegramID, telegramUserID)
	if err != nil {
		return nil, err
	}

	return gift, nil
}

func (s *Service) WithdrawGift(ctx context.Context, giftID string) (*gift.Gift, error) {
	gift, err := s.repo.MarkGiftForWithdrawal(ctx, giftID)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

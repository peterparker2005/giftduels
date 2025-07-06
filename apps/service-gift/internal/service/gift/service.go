package gift

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
)

type Service struct {
	repo gift.Repository
}

func New(repo gift.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetGiftByID(ctx context.Context, id string) (*gift.Gift, error) {
	dbGift, err := s.repo.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ConvertDBGiftToDomain(dbGift), nil
}

func (s *Service) GetUserGifts(ctx context.Context, telegramUserID int64, limit int32, offset int32) ([]*gift.Gift, error) {
	dbGifts, err := s.repo.GetUserGifts(ctx, limit, offset, telegramUserID)
	if err != nil {
		return nil, err
	}

	gifts := make([]*gift.Gift, len(dbGifts))
	for i, dbGift := range dbGifts {
		gifts[i] = ConvertDBGiftToDomain(dbGift)
	}

	return gifts, nil
}

func (s *Service) StakeGift(ctx context.Context, giftID string) (*gift.Gift, error) {
	dbGift, err := s.repo.StakeGiftForGame(ctx, giftID)
	if err != nil {
		return nil, err
	}
	return ConvertDBGiftToDomain(dbGift), nil
}

func (s *Service) TransferGiftToUser(ctx context.Context, giftID string, telegramUserID int64) (*gift.Gift, error) {
	// First update the gift owner
	dbGift, err := s.repo.UpdateGiftOwner(ctx, giftID, telegramUserID)
	if err != nil {
		return nil, err
	}

	// Create transfer record
	_, err = s.repo.CreateGiftEvent(ctx, giftID, dbGift.OwnerTelegramID, telegramUserID)
	if err != nil {
		return nil, err
	}

	return ConvertDBGiftToDomain(dbGift), nil
}

func (s *Service) WithdrawGift(ctx context.Context, giftID string) (*gift.Gift, error) {
	dbGift, err := s.repo.MarkGiftForWithdrawal(ctx, giftID)
	if err != nil {
		return nil, err
	}
	return ConvertDBGiftToDomain(dbGift), nil
}

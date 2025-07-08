package user

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
)

type Service struct {
	repo user.UserRepository
}

func NewService(userRepo user.UserRepository) *Service {
	return &Service{repo: userRepo}
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	u, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) UpsertUser(ctx context.Context, params user.CreateUserParams) (*user.User, error) {
	u, err := s.repo.UpsertUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return u, nil
}

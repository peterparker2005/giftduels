package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
	identityEvents "github.com/peterparker2005/giftduels/packages/events/identity"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	repo      user.UserRepository
	publisher message.Publisher
	log       *logger.Logger
}

func NewService(userRepo user.UserRepository, publisher message.Publisher, log *logger.Logger) *Service {
	return &Service{
		repo:      userRepo,
		publisher: publisher,
		log:       log,
	}
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	u, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) UpsertUser(ctx context.Context, params *user.CreateUserParams) (*user.User, error) {
	u, created, err := s.repo.CreateOrUpdate(ctx, params)
	if err != nil {
		return nil, err
	}

	// Если пользователь был создан впервые, публикуем событие
	if created {
		if err = s.publishUserCreatedEvent(u); err != nil {
			// Логируем ошибку, но не возвращаем её, так как пользователь уже создан
			s.log.Error("Failed to publish user created event", zap.Error(err))
		}
	}

	return u, nil
}

func (s *Service) publishUserCreatedEvent(u *user.User) error {
	event := &identityv1.NewUserEvent{
		UserId: &sharedv1.UserId{
			Value: u.ID,
		},
		TelegramId: &sharedv1.TelegramUserId{
			Value: u.TelegramID,
		},
	}

	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user created event: %w", err)
	}

	msg := message.NewMessage(uuid.New().String(), payload)

	// Добавляем метаданные
	metadata := map[string]string{
		"event_type":       identityEvents.TopicUserCreated.String(),
		"telegram_user_id": strconv.FormatInt(u.TelegramID, 10),
		"user_id":          u.ID,
	}

	for key, value := range metadata {
		msg.Metadata.Set(key, value)
	}

	return s.publisher.Publish(identityEvents.TopicUserCreated.String(), msg)
}

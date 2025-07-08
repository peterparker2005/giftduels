package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type IdentityNewUserHandler struct {
	repo   payment.Repository
	logger *logger.Logger
}

func NewIdentityNewUserHandler(repo payment.Repository, logger *logger.Logger) *IdentityNewUserHandler {
	return &IdentityNewUserHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *IdentityNewUserHandler) Handle(msg *message.Message) error {
	ctx := context.Background()
	h.logger.Info("Unmarshalling event copy me", zap.Any("payload", msg.Payload))

	var ev identityv1.NewUserEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error("Failed to unmarshal event", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	userBalance, err := h.repo.GetUserBalance(ctx, ev.TelegramId.Value)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		h.logger.Error("Failed to get balance", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("get balance: %w", err)
	}

	if userBalance != nil {
		h.logger.Info("Balance already exists", zap.Int64("telegram_user_id", ev.TelegramId.Value))
		return nil
	}

	newBalance := &payment.CreateBalanceParams{
		TelegramUserID: ev.TelegramId.Value,
	}

	if err := h.repo.Create(ctx, newBalance); err != nil {
		h.logger.Error("Failed to create balance", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("create balance: %w", err)
	}

	return nil
}

package grpc

import (
	"context"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ identityv1.IdentityPublicServiceServer = (*IdentityPublicHandler)(nil)

type IdentityPublicHandler struct {
	identityv1.IdentityPublicServiceServer
	tokenSvc token.TokenService
	logger   *logger.Logger
	cfg      *config.Config
}

func NewIdentityPublicHandler(tokenSvc token.TokenService, logger *logger.Logger, cfg *config.Config) identityv1.IdentityPublicServiceServer {
	return &IdentityPublicHandler{
		tokenSvc: tokenSvc,
		logger:   logger,
		cfg:      cfg,
	}
}

// Authorize обрабатывает запрос авторизации, проверяет initData и генерирует сессионный token.
func (h *IdentityPublicHandler) Authorize(ctx context.Context, req *identityv1.AuthorizeRequest) (*identityv1.AuthorizeResponse, error) {
	if err := initdata.Validate(req.InitData, h.cfg.Telegram.BotToken, 24*time.Hour); err != nil {
		h.logger.Warn("invalid initData", zap.Error(err))
		return nil, errors.NewValidationError("init_data", "Invalid initialization data")
	}

	parsed, err := initdata.Parse(req.InitData)
	if err != nil {
		h.logger.Error("failed to parse initData", zap.Error(err))
		return nil, errors.NewInternalError("failed to parse initialization data")
	}

	telegramID := parsed.User.ID

	token, err := h.tokenSvc.Generate(telegramID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		return nil, errors.NewInternalError("failed to generate session token")
	}

	return &identityv1.AuthorizeResponse{
		Token: token,
	}, nil
}

func (h *IdentityPublicHandler) ValidateToken(ctx context.Context, req *identityv1.ValidateTokenRequest) (*identityv1.ValidateTokenResponse, error) {
	claims, err := h.tokenSvc.Validate(req.Token)
	if err != nil {
		h.logger.Warn("invalid token", zap.Error(err))
		return nil, errors.NewUnauthorizedError("invalid token")
	}

	return &identityv1.ValidateTokenResponse{
		TelegramUserId: &sharedv1.TelegramUserId{Value: claims.TelegramUserID},
	}, nil
}

func (h *IdentityPublicHandler) GetProfile(ctx context.Context, req *emptypb.Empty) (*identityv1.GetProfileResponse, error) {
	return nil, nil
}

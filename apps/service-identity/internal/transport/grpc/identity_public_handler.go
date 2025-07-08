package grpc

import (
	"context"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	userservice "github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ identityv1.IdentityPublicServiceServer = (*IdentityPublicHandler)(nil)

type IdentityPublicHandler struct {
	identityv1.IdentityPublicServiceServer
	tokenSvc token.TokenService
	userSvc  *userservice.Service
	logger   *logger.Logger
	cfg      *config.Config
}

func NewIdentityPublicHandler(tokenSvc token.TokenService, userSvc *userservice.Service, logger *logger.Logger, cfg *config.Config) identityv1.IdentityPublicServiceServer {
	return &IdentityPublicHandler{
		tokenSvc: tokenSvc,
		userSvc:  userSvc,
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

	// Upsert user data from Telegram initData
	userParams := &user.CreateUserParams{
		TelegramID:      telegramID,
		Username:        getStringOrEmpty(parsed.User.Username),
		FirstName:       parsed.User.FirstName,
		LastName:        getStringOrEmpty(parsed.User.LastName),
		PhotoUrl:        getStringOrEmpty(parsed.User.PhotoURL),
		LanguageCode:    getStringOrEmpty(parsed.User.LanguageCode),
		AllowsWriteToPm: parsed.User.AllowsWriteToPm,
		IsPremium:       parsed.User.IsPremium,
	}

	_, err = h.userSvc.UpsertUser(ctx, userParams)
	if err != nil {
		h.logger.Error("failed to upsert user", zap.Error(err), zap.Int64("telegram_id", telegramID))
		return nil, errors.NewInternalError("failed to process user data")
	}

	token, err := h.tokenSvc.Generate(telegramID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		return nil, errors.NewInternalError("failed to generate session token")
	}

	return &identityv1.AuthorizeResponse{
		Token: token,
	}, nil
}

func (h *IdentityPublicHandler) GetProfile(ctx context.Context, req *emptypb.Empty) (*identityv1.GetProfileResponse, error) {
	telegramID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	u, err := h.userSvc.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	return &identityv1.GetProfileResponse{
		Profile: toPBProfile(u),
	}, nil
}

// Helper functions to safely extract values from initData
func getStringOrEmpty(val string) string {
	return val
}

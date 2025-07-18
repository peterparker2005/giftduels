package grpchandlers

import (
	"context"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/proto"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	userservice "github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ identityv1.IdentityPublicServiceServer = (*IdentityPublicHandler)(nil)

type IdentityPublicHandler struct {
	identityv1.IdentityPublicServiceServer

	tokenSvc token.Service
	userSvc  *userservice.Service
	logger   *logger.Logger
	cfg      *config.Config
}

func NewIdentityPublicHandler(
	tokenSvc token.Service,
	userSvc *userservice.Service,
	logger *logger.Logger,
	cfg *config.Config,
) identityv1.IdentityPublicServiceServer {
	return &IdentityPublicHandler{
		tokenSvc: tokenSvc,
		userSvc:  userSvc,
		logger:   logger,
		cfg:      cfg,
	}
}

// Authorize обрабатывает запрос авторизации, проверяет initData и генерирует сессионный token.
func (h *IdentityPublicHandler) Authorize(
	ctx context.Context,
	req *identityv1.AuthorizeRequest,
) (*identityv1.AuthorizeResponse, error) {
	//nolint:mnd // 24 hours is a reasonable timeout
	if err := initdata.Validate(req.GetInitData(), h.cfg.Telegram.BotToken, 24*time.Hour); err != nil {
		h.logger.Warn("invalid initData", zap.Error(err))
		return nil, errors.NewError(
			errors.WithGRPCCode(codes.InvalidArgument),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			errors.WithMessage("invalid initData"),
		)
	}

	parsed, err := initdata.Parse(req.GetInitData())
	if err != nil {
		h.logger.Error("failed to parse initData", zap.Error(err))
		return nil, errors.NewError(
			errors.WithGRPCCode(codes.Internal),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			errors.WithMessage("failed to parse initData"),
		)
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
		h.logger.Error(
			"failed to upsert user",
			zap.Error(err),
			zap.Int64("telegram_id", telegramID),
		)
		return nil, errors.NewError(
			errors.WithGRPCCode(codes.Internal),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			errors.WithMessage("failed to process user data"),
		)
	}

	token, err := h.tokenSvc.Generate(telegramID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		return nil, errors.NewError(
			errors.WithGRPCCode(codes.Internal),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			errors.WithMessage("failed to generate session token"),
		)
	}

	return &identityv1.AuthorizeResponse{
		Token: token,
	}, nil
}

func (h *IdentityPublicHandler) GetProfile(
	ctx context.Context,
	_ *emptypb.Empty,
) (*identityv1.GetProfileResponse, error) {
	telegramID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	u, err := h.userSvc.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	return &identityv1.GetProfileResponse{
		Profile: proto.ToPBProfile(u),
	}, nil
}

// Helper functions to safely extract values from initData.
func getStringOrEmpty(val string) string {
	return val
}

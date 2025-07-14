package service

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	userservice "github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// Module предоставляет service зависимости

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("services",
	fx.Provide(
		// Предоставляем TokenService через JWTService
		func(cfg *config.Config, logger *logger.Logger) token.Service {
			return token.NewJWTService(&cfg.JWT, logger.Zap())
		},
		// Предоставляем UserService
		userservice.NewService,
		// Предоставляем *zap.Logger для обратной совместимости
		func(logger *logger.Logger) *zap.Logger {
			return logger.Zap()
		},
	),
)

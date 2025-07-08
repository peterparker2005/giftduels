package service

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	userservice "github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// Module предоставляет service зависимости
var Module = fx.Module("services",
	fx.Provide(
		// Предоставляем TokenService через JWTService
		func(cfg *config.Config, logger *logger.Logger) token.TokenService {
			return token.NewJWTService(&cfg.JWT, logger.Zap())
		},
		// Предоставляем UserService
		func(userRepo user.UserRepository) *userservice.Service {
			return userservice.NewService(userRepo)
		},
		// Предоставляем *zap.Logger для обратной совместимости
		func(logger *logger.Logger) *zap.Logger {
			return logger.Zap()
		},
	),
)

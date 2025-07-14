package service

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// Module предоставляет service зависимости

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("services",
	fx.Provide(
		func(cfg *config.Config, logger *logger.Logger) token.Service {
			return token.NewJWTService(&cfg.JWT, logger.Zap())
		},
		user.NewService,
	),
)

package app

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/version-go"
)

// LoggerModule предоставляет настроенный логгер через fx
var LoggerModule = fx.Module("logger",
	fx.Provide(
		func(cfg *config.Config) (*logger.Logger, error) {
			return logger.NewLogger(logger.Config{
				Service:     cfg.ServiceName,
				Level:       cfg.Logger.LogLevel,
				Pretty:      cfg.Logger.Pretty,
				Environment: cfg.Environment,
				Version:     version.Version,
			})
		},
	),
)

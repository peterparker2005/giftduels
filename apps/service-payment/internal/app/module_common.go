package app

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var moduleCommon = fx.Options(
	LoggerModule,
	fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
		return log.ToFxLogger()
	}),
	config.Module,
	pg.Module,
)

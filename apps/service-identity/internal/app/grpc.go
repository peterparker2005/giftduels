package app

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/event"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/transport"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

func NewGRPCApp() *fx.App {
	return fx.New(
		fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
			return log.ToFxLogger()
		}),
		config.Module,
		LoggerModule,
		pg.Module,
		event.Module,
		service.Module,
		transport.Module,
	)
}

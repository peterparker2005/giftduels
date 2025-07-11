package app

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/event"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/transport"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

func Run(cfg *config.Config) {
	fx.New(
		fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
			return log.ToFxLogger()
		}),
		fx.Provide(func() *config.Config {
			return cfg
		}),
		LoggerModule,
		pg.Module,
		event.Module,
		service.Module,
		transport.Module,

		// Lifecycle hooks
		fx.Invoke(registerHooks),
	).Run()
}

func registerHooks(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

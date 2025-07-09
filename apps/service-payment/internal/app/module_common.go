package app

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var moduleCommon = fx.Options(
	LoggerModule,
	fx.Provide(func(cfg *config.Config) (*clients.Clients, error) {
		return clients.NewClients(context.Background(), cfg.GRPC)
	}),
	fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
		return log.ToFxLogger()
	}),
	config.Module,
	pg.Module,
)

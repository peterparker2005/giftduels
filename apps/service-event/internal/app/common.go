package app

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-event/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/adapter/redis"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

//nolint:gochecknoglobals // fx module pattern
var CommonModule = fx.Options(
	config.Module,
	LoggerModule,
	fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
		return log.ToFxLogger()
	}),
	fx.Provide(func(cfg *config.Config) (*clients.Clients, error) {
		return clients.NewClients(context.Background(), cfg.GRPC)
	}),
	amqp.Module,
	redis.Module,
)

package app

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/transport"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"

	"github.com/peterparker2005/giftduels/packages/logger-go"
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
	pg.Module,
	service.Module,
)

func NewGRPCApp() *fx.App {
	return fx.New(
		CommonModule,
		transport.Module,
	)
}

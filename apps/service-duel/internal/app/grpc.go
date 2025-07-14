package app

import (
	"context"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	amqputil "github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/transport"
	duelEvents "github.com/peterparker2005/giftduels/packages/events/duel"
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
	pg.Module,
	service.Module,
	fx.Provide(
		amqputil.ProvideConnection,
		func(cfg *config.Config, c *amqp.ConnectionWrapper, l *logger.Logger) (message.Publisher, error) {
			return amqputil.ProvidePublisher(c, l, duelEvents.Config(cfg.ServiceName))
		},
	),
)

func NewGRPCApp() *fx.App {
	return fx.New(
		CommonModule,
		transport.Module,
	)
}

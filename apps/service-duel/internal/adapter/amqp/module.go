package amqp

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	duelevents "github.com/peterparker2005/giftduels/packages/events/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("amqp",
	fx.Provide(
		ProvideConnection,
		ProvideSubFactory,
		func(cfg *config.Config, c *amqp.ConnectionWrapper, l *logger.Logger) (message.Publisher, error) {
			return ProvidePublisher(c, l, duelevents.Config(cfg.ServiceName.String()))
		},
		ProvideRouter,
		ProvideOutbox,
	),
)

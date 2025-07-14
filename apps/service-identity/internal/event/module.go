package event

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	amqputil "github.com/peterparker2005/giftduels/apps/service-identity/internal/event/amqp"
	identityEvents "github.com/peterparker2005/giftduels/packages/events/identity"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	//-------------------------------- AMQP low-level -------------------------------
	fx.Provide(
		amqputil.ProvideConnection,

		func(cfg *config.Config, c *amqp.ConnectionWrapper, l *logger.Logger) (message.Publisher, error) {
			return amqputil.ProvidePublisher(c, l, identityEvents.Config(cfg.ServiceName.String()))
		},
	),
)

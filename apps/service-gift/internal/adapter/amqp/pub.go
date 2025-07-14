package amqp

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	giftevents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

func ProvidePublisher(
	conn *amqp.ConnectionWrapper,
	log *logger.Logger,
	cfg *config.Config,
) (message.Publisher, error) {
	amqpCfg := giftevents.Config(cfg.ServiceName.String())
	return amqp.NewPublisherWithConnection(
		amqpCfg.Build(),
		logger.NewWatermill(log),
		conn,
	)
}

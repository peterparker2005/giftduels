package amqp

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/packages/events"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// SubFactory — «дай exchange – получи Subscriber».
type SubFactory func(events.AMQPConfig) (message.Subscriber, error)

func ProvideSubFactory(
	conn *amqp.ConnectionWrapper,
	log *logger.Logger,
) SubFactory {
	return func(cfg events.AMQPConfig) (message.Subscriber, error) {
		return amqp.NewSubscriberWithConnection(
			cfg.Build(), logger.NewWatermill(log), conn,
		)
	}
}

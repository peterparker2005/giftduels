package amqp

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/packages/events"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

func ProvidePublisher(
	conn *amqp.ConnectionWrapper,
	log *logger.Logger,
	cfg events.AMQPConfig,
) (message.Publisher, error) {
	return amqp.NewPublisherWithConnection(
		cfg.Build(), logger.NewWatermill(log.Zap()), conn,
	)
}

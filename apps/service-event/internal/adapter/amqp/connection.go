package amqp

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// ProvideConnection — *одно* TCP-соединение на всё приложение.
func ProvideConnection(cfg *config.Config, log *logger.Logger) (*amqp.ConnectionWrapper, error) {
	return amqp.NewConnection(
		amqp.ConnectionConfig{AmqpURI: cfg.AMQP.Address()},
		logger.NewWatermill(log),
	)
}

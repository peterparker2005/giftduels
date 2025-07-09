package events

import (
	"time"

	watermillAmqp "github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPConfig struct {
	Service  string
	Exchange string
	Kind     string // fanout|topic|direct
	TTL      time.Duration
	Pool     int
}

func (c *AMQPConfig) Build() watermillAmqp.Config {
	qArgs := amqp.Table{
		"x-queue-type": "quorum",
	}
	if c.TTL > 0 {
		qArgs["x-message-ttl"] = int32(c.TTL.Milliseconds())
	}

	if c.Pool == 0 {
		c.Pool = 10
	}

	return watermillAmqp.Config{
		Exchange: watermillAmqp.ExchangeConfig{
			GenerateName: func(string) string { return c.Exchange },
			Type:         c.Kind,
			Durable:      true,
		},
		Publish: watermillAmqp.PublishConfig{
			ChannelPoolSize:    c.Pool,
			GenerateRoutingKey: func(t string) string { return t },
		},
		Queue: watermillAmqp.QueueConfig{
			GenerateName: func(t string) string { return c.Service + "." + t },
			Durable:      true,
			AutoDelete:   false,
			Arguments:    qArgs,
		},
		QueueBind: watermillAmqp.QueueBindConfig{
			GenerateRoutingKey: func(t string) string { return t },
		},
		Consume: watermillAmqp.ConsumeConfig{
			Qos: watermillAmqp.QosConfig{PrefetchCount: 1},
		},
		TopologyBuilder: &watermillAmqp.DefaultTopologyBuilder{},
		Marshaler: watermillAmqp.DefaultMarshaler{
			MessageUUIDHeaderKey: "x-message-id",
		},
	}
}

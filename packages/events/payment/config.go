package payment

import (
	"github.com/peterparker2005/giftduels/packages/events"
)

func Config(serviceName string) events.AMQPConfig {
	return events.AMQPConfig{
		Exchange: "payment.events",
		Kind:     "topic",
		Service:  serviceName,
		Pool:     10,
		TTL:      0,
	}
}

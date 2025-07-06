package telegram

import (
	"github.com/peterparker2005/giftduels/packages/events"
)

func Config(serviceName string) events.AMQPConfig {
	return events.AMQPConfig{
		Exchange: "telegram.events",
		Kind:     "topic",
		Service:  serviceName,
		Pool:     10,
		TTL:      0,
	}
}

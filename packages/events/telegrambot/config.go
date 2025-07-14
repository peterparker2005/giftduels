package telegrambot

import (
	"github.com/peterparker2005/giftduels/packages/events"
)

const (
	DefaultPool = 10
	DefaultTTL  = 0
)

func Config(serviceName string) events.AMQPConfig {
	return events.AMQPConfig{
		Exchange: "telegrambot.events",
		Kind:     "topic",
		Service:  serviceName,
		Pool:     DefaultPool,
		TTL:      DefaultTTL,
	}
}

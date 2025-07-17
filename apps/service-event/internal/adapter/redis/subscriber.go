package redis

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	"github.com/redis/go-redis/v9"
)

// ProvideRedisSubscriber ▶️ Redis Streams.
func ProvideRedisSubscriber(client *redis.Client, _ *config.Config) (message.Subscriber, error) {
	s, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:       client,
			Unmarshaller: redisstream.DefaultMarshallerUnmarshaller{},
			// ConsumerGroup:  cfg.ServiceName.String(),         // группа
			// Consumer:       cfg.ServiceName.String() + "_c1", // consumer id
			OldestId:       "$",
			FanOutOldestId: "$",
			//nolint:mnd // 2 seconds
			BlockTime: 2 * time.Second,
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

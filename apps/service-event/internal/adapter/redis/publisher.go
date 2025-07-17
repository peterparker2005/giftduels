package redis

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultMaxlen     = 10_000
	DefaultDuelMaxlen = 5_000
)

// ProvideRedisPublisher ▶️ Redis Streams.
func ProvideRedisPublisher(client *redis.Client) (message.Publisher, error) {
	// настраиваем Watermill’ный Publisher
	p, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:        client,
			Marshaller:    redisstream.DefaultMarshallerUnmarshaller{},
			DefaultMaxlen: DefaultMaxlen,
			Maxlens: map[string]int64{
				"duel:": DefaultDuelMaxlen,
			},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

package asynq

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewAsynqClient(redisClient *redis.Client) *asynq.Client {
	return asynq.NewClientFromRedisClient(redisClient)
}

func NewAsynqServer(redisClient *redis.Client) *asynq.Server {
	return asynq.NewServerFromRedisClient(
		redisClient,
		asynq.Config{
			Queues: map[string]int{
				queueName: 1,
			},
			//nolint:mnd // 10 is reasonable
			Concurrency: 10,
			RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second
			},
		},
	)
}

//nolint:gochecknoglobals // fx module pattern
var ConnectionModule = fx.Module("asynq.conn",
	fx.Provide(
		NewAsynqClient,
		NewAsynqServer,
		NewScheduler,
		NewSchedulerProvider,
	),
)

//nolint:gochecknoglobals // fx module pattern
var HandlerModule = fx.Module("asynq.handler",
	fx.Invoke(RegisterHandlers),
)

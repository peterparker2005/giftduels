package redis

import (
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(NewRedisClient),
	fx.Provide(
		ProvideRedisPublisher,
	),
	fx.Provide(
		ProvideRedisSubscriber,
	),
)

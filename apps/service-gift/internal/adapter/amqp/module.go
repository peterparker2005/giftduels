package amqp

import (
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("amqp",
	fx.Provide(
		ProvideConnection,
		ProvideSubFactory,
		ProvidePublisher,
		ProvideRouter,
		ProvideOutbox,
	),
)

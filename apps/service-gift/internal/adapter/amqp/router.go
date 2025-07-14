package amqp

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

const (
	defaultMaxRetries      = 5
	defaultHandlerTimeout  = 30 * time.Second
	defaultInitialInterval = time.Second
	defaultMultiplier      = 2
)

// ProvideRouter configures retry + poison.
func ProvideRouter(
	log *logger.Logger,
	pub message.Publisher,
) (*message.Router, error) {
	r, err := message.NewRouter(message.RouterConfig{}, logger.NewWatermill(log))
	if err != nil {
		return nil, err
	}

	// Increase retry for critical rollback operations
	retry := middleware.Retry{
		MaxRetries:      defaultMaxRetries,
		InitialInterval: defaultInitialInterval,
		Multiplier:      defaultMultiplier,
	}
	// TODO: add poison key to config?
	poison, err := middleware.PoisonQueue(pub, "gift.events.poison")
	if err != nil {
		return nil, err
	}

	r.AddMiddleware(
		middleware.CorrelationID,
		middleware.Timeout(defaultHandlerTimeout),
		middleware.Recoverer,
		retry.Middleware,
		poison,
	)

	return r, nil
}

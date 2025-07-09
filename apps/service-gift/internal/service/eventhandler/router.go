package eventhandler

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

// ProvideRouter настраивает retry + poison.
func ProvideRouter(
	log *logger.Logger,
	pub message.Publisher,
	poisonKey string, // gift.events.poison и т.п.
) (*message.Router, error) {
	r, err := message.NewRouter(message.RouterConfig{}, logger.NewWatermill(log))
	if err != nil {
		return nil, err
	}

	retry := middleware.Retry{MaxRetries: 3, InitialInterval: time.Second, Multiplier: 2}
	poison, err := middleware.PoisonQueue(pub, poisonKey)
	if err != nil {
		return nil, err
	}

	r.AddMiddleware(
		middleware.CorrelationID,
		middleware.Timeout(30*time.Second),
		middleware.Recoverer,
		retry.Middleware,
		poison,
	)

	return r, nil
}

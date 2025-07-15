package worker

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("worker",
	// провайдим хендлеры
	fx.Provide(
	// handlers
	),

	// инициализируем router + forwarder
	// fx.Invoke(registerHandlers),
)

func registerHandlers(
	lc fx.Lifecycle,
	cfg *config.Config,
	subFac amqp.SubFactory,
	fwd *forwarder.Forwarder,
	router *message.Router,
	log *logger.Logger,
) error {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := fwd.Run(context.Background()); err != nil &&
					!errors.Is(err, context.Canceled) {
					log.Fatal("forwarder stopped", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return fwd.Close()
		},
	})

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			runCtx, cancel := context.WithCancel(context.Background())

			// сохраняем cancel, чтобы вызвать его в OnStop
			lc.Append(fx.Hook{
				OnStop: func(_ context.Context) error {
					cancel() // остановить router.Run
					return router.Close()
				},
			})

			go func() {
				if err := router.Run(runCtx); err != nil &&
					!errors.Is(err, context.Canceled) {
					log.Fatal("router stopped", zap.Error(err))
				}
			}()
			return nil
		},
	})

	return nil
}

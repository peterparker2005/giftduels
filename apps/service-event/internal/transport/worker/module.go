package worker

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/transport/worker/handlers"
	duelevents "github.com/peterparker2005/giftduels/packages/events/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(
		handlers.NewDuelCreatedHandler,
	),
	// инициализируем router + forwarder
	fx.Invoke(registerHandlers),
)

func registerHandlers(
	lc fx.Lifecycle,
	cfg *config.Config,
	subFac amqp.SubFactory,
	router *message.Router,
	log *logger.Logger,
	duelCreatedHandler *handlers.DuelCreatedHandler,
) error {
	duelSub, err := subFac(duelevents.Config(cfg.ServiceName.String()))
	if err != nil {
		return err
	}

	router.AddNoPublisherHandler(
		"duel_created",
		duelevents.TopicDuelCreated.String(),
		duelSub,
		duelCreatedHandler.Handle,
	)

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
				if err = router.Run(runCtx); err != nil &&
					!errors.Is(err, context.Canceled) {
					log.Fatal("router stopped", zap.Error(err))
				}
			}()
			return nil
		},
	})

	return nil
}

package eventhandler

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	amqputil "github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	duelEvents "github.com/peterparker2005/giftduels/packages/events/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(
		amqputil.ProvideConnection,
		amqputil.ProvideSubFactory,

		func(cfg *config.Config, c *amqp.ConnectionWrapper, l *logger.Logger) (message.Publisher, error) {
			return amqputil.ProvidePublisher(c, l, duelEvents.Config(cfg.ServiceName.String()))
		},
	),

	fx.Invoke(func(
		cfg *config.Config,
		lc fx.Lifecycle,
		log *logger.Logger,
		subFac amqputil.SubFactory,
		pub message.Publisher,
		pool *pgxpool.Pool,
	) error {
		router, err := ProvideRouter(
			log,
			pub,
			duelEvents.Config(cfg.ServiceName.String()).Exchange+".poison",
		)
		if err != nil {
			return err
		}

		db := stdlib.OpenDBFromPool(pool)

		sqlSubscriber, err := sql.NewSubscriber(
			db,
			sql.SubscriberConfig{
				SchemaAdapter:    sql.DefaultPostgreSQLSchema{},
				OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
				InitializeSchema: true,
			},
			nil,
		)
		if err != nil {
			return err
		}

		fwd, err := forwarder.NewForwarder(
			sqlSubscriber,
			pub,
			logger.NewWatermill(log),
			forwarder.Config{
				ForwarderTopic: duelEvents.SQLOutboxTopic,
			},
		)
		if err != nil {
			return err
		}

		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err = fwd.Run(context.Background()); err != nil &&
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

				lc.Append(fx.Hook{
					OnStop: func(_ context.Context) error {
						cancel()
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
	}),
)

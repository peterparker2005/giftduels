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
	amqputil "github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/events/telegram"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	//-------------------------------- AMQP low-level -------------------------------
	fx.Provide(
		amqputil.ProvideConnection,
		amqputil.ProvideSubFactory,

		// publisher в gift.events, нужен для poison-queue
		func(cfg *config.Config, c *amqp.ConnectionWrapper, l *logger.Logger) (message.Publisher, error) {
			return amqputil.ProvidePublisher(c, l, giftEvents.Config(cfg.ServiceName))
		},
	),

	//-------------------------------- бизнес-логика --------------------------------
	fx.Provide(func(
		repo gift.GiftRepository,
		l *logger.Logger,
	) *TelegramGiftReceivedHandler {
		return NewTelegramGiftReceivedHandler(repo, l)
	}),

	fx.Provide(func(
		repo gift.GiftRepository,
		l *logger.Logger,
	) *GiftWithdrawFailedHandler {
		return NewGiftWithdrawFailedHandler(repo, l)
	}),

	//-------------------------------- router & lifecycle ---------------------------
	fx.Invoke(func(
		cfg *config.Config,
		lc fx.Lifecycle,
		log *logger.Logger,
		subFac amqputil.SubFactory,
		pub message.Publisher,
		giftReceivedHandler *TelegramGiftReceivedHandler,
		withdrawFailedHandler *GiftWithdrawFailedHandler,
		pool *pgxpool.Pool,
	) error {
		router, err := ProvideRouter(log, pub, giftEvents.Config(cfg.ServiceName).Exchange+".poison")
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
			// logger.NewWatermill(log),
		)
		if err != nil {
			return err
		}

		// ── подписчики ────────────────────────────────────────────────────────
		telegramSub, err := subFac(telegram.Config(cfg.ServiceName))
		if err != nil {
			return err
		}

		// ── хендлеры ──────────────────────────────────────────────────────────
		router.AddNoPublisherHandler("telegram_gift_received", telegram.TopicTelegramGiftReceived.String(), telegramSub, giftReceivedHandler.Handle)
		router.AddNoPublisherHandler("gift_withdraw_failed", giftEvents.TopicGiftWithdrawFailed.String(), telegramSub, withdrawFailedHandler.Handle)
		// router.AddNoPublisherHandler("tg_gift_poison", giftEvents.Config(cfg.ServiceName).Exchange+".poison", giftSub, func(m *message.Message) error {
		// 	log.Warn("💀 poison", zap.String("body", string(m.Payload)))
		// 	return nil
		// })

		fwd, err := forwarder.NewForwarder(sqlSubscriber, pub, logger.NewWatermill(log), forwarder.Config{
			ForwarderTopic: giftEvents.SqlOutboxTopic,
		})
		if err != nil {
			return err
		}

		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := fwd.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
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
	}),
)

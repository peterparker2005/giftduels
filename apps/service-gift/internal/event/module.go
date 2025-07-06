package event

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
	amqputil "github.com/peterparker2005/giftduels/apps/service-gift/internal/event/amqp"
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
		repo gift.Repository,
		pricingRepo pricing.Repository,
		l *logger.Logger,
	) *TelegramGiftReceivedHandler {
		return NewTelegramGiftReceivedHandler(repo, pricingRepo, l.Zap())
	}),

	//-------------------------------- router & lifecycle ---------------------------
	fx.Invoke(func(
		cfg *config.Config,
		lc fx.Lifecycle,
		log *logger.Logger,
		subFac amqputil.SubFactory,
		pub message.Publisher,
		handler *TelegramGiftReceivedHandler,
	) error {
		router, err := ProvideRouter(log, pub, giftEvents.Config(cfg.ServiceName).Exchange+".poison")
		if err != nil {
			return err
		}

		// ── подписчики ────────────────────────────────────────────────────────
		telegramSub, err := subFac(telegram.Config(cfg.ServiceName))
		if err != nil {
			return err
		}

		// giftSub, err := subFac(giftEvents.Config(cfg.ServiceName))
		// if err != nil {
		// 	return err
		// }

		// ── хендлеры ──────────────────────────────────────────────────────────
		router.AddNoPublisherHandler("tg_gift", telegram.TopicTelegramGiftReceived.String(), telegramSub, handler.Handle)
		// router.AddNoPublisherHandler("tg_gift_poison", giftEvents.Config(cfg.ServiceName).Exchange+".poison", giftSub, func(m *message.Message) error {
		// 	log.Warn("💀 poison", zap.String("body", string(m.Payload)))
		// 	return nil
		// })

		// ── fx-lifecycle ──────────────────────────────────────────────────────
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

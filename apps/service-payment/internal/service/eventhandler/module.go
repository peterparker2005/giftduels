package eventhandler

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	amqputil "github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/events/identity"
	paymentEvents "github.com/peterparker2005/giftduels/packages/events/payment"
	telegramEvents "github.com/peterparker2005/giftduels/packages/events/telegram"
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
			return amqputil.ProvidePublisher(c, l, paymentEvents.Config(cfg.ServiceName))
		},
	),

	fx.Provide(NewIdentityNewUserHandler),

	fx.Provide(NewTelegramGiftWithdrawFailedHandler),

	fx.Invoke(func(
		cfg *config.Config,
		lc fx.Lifecycle,
		log *logger.Logger,
		subFac amqputil.SubFactory,
		pub message.Publisher,
		newUserHandler *IdentityNewUserHandler,
		telegramGiftWithdrawFailHandler *TelegramGiftWithdrawFailedHandler,
	) error {
		router, err := ProvideRouter(log, pub, paymentEvents.Config(cfg.ServiceName).Exchange+".poison")
		if err != nil {
			return err
		}

		identitySub, err := subFac(identity.Config(cfg.ServiceName))
		if err != nil {
			return err
		}

		telegramSub, err := subFac(telegramEvents.Config(cfg.ServiceName))
		if err != nil {
			return err
		}

		router.AddNoPublisherHandler(
			"identity_new_user",
			identity.TopicUserCreated.String(),
			identitySub,
			newUserHandler.Handle,
		)
		router.AddNoPublisherHandler(
			"telegram_gift_withdraw_fail",
			giftEvents.TopicGiftWithdrawFailed.String(),
			telegramSub,
			telegramGiftWithdrawFailHandler.Handle,
		)
		// router.AddNoPublisherHandler("tg_gift_poison", giftEvents.Config(cfg.ServiceName).Exchange+".poison", giftSub, func(m *message.Message) error {
		// 	log.Warn("💀 poison", zap.String("body", string(m.Payload)))
		// 	return nil
		// })

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

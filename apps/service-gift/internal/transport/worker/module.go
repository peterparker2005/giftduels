package worker

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/amqp"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	workerhandlers "github.com/peterparker2005/giftduels/apps/service-gift/internal/transport/worker/handlers"
	duelevents "github.com/peterparker2005/giftduels/packages/events/duel"
	telegramEvents "github.com/peterparker2005/giftduels/packages/events/telegram"
	telegrambotEvents "github.com/peterparker2005/giftduels/packages/events/telegrambot"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	// провайдим инфраструктуру
	amqp.Module,

	// провайдим хендлеры
	fx.Provide(
		workerhandlers.NewTelegramGiftReceivedHandler,
		workerhandlers.NewGiftWithdrawFailedHandler,
		workerhandlers.NewInvoicePaymentHandler,
		workerhandlers.NewGiftReturnedHandler,
	),

	// инициализируем router + forwarder
	fx.Invoke(registerHandlers),
)

func registerHandlers(
	lc fx.Lifecycle,
	cfg *config.Config,
	subFac amqp.SubFactory,
	fwd *forwarder.Forwarder,
	tgHandler *workerhandlers.TelegramGiftReceivedHandler,
	failHandler *workerhandlers.GiftWithdrawFailedHandler,
	invHandler *workerhandlers.InvoicePaymentHandler,
	returnedHandler *workerhandlers.GiftReturnedHandler,
	router *message.Router,
	log *logger.Logger,
) error {
	// Создаем подписчиков для разных топиков
	telegramSub, err := subFac(telegramEvents.Config(cfg.ServiceName.String()))
	if err != nil {
		return err
	}

	telegrambotSub, err := subFac(telegrambotEvents.Config(cfg.ServiceName.String()))
	if err != nil {
		return err
	}

	duelSub, err := subFac(duelevents.Config(cfg.ServiceName.String()))
	if err != nil {
		return err
	}

	// регистрируем каждый
	router.AddNoPublisherHandler(
		"tg_gift_recv",
		telegramEvents.TopicTelegramGiftReceived.String(),
		telegramSub,
		tgHandler.Handle,
	)
	router.AddNoPublisherHandler(
		"gift_withdraw_fail",
		telegramEvents.TopicTelegramGiftWithdrawFailed.String(),
		telegramSub,
		failHandler.Handle,
	)
	router.AddNoPublisherHandler(
		"invoice_paid",
		telegrambotEvents.TopicInvoicePaymentCompleted.String(),
		telegrambotSub,
		invHandler.Handle,
	)
	router.AddNoPublisherHandler(
		"create_duel_fail",
		duelevents.TopicDuelCreateFailed.String(),
		duelSub,
		returnedHandler.Handle,
	)

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

package worker

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/amqp"
	workerhandlers "github.com/peterparker2005/giftduels/apps/service-gift/internal/transport/worker/handlers"
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
	),

	// инициализируем router + forwarder
	fx.Invoke(registerHandlers),
)

func registerHandlers(
	lc fx.Lifecycle,
	sub message.Subscriber,
	fwd *forwarder.Forwarder,
	tgHandler workerhandlers.TelegramGiftReceivedHandler,
	failHandler workerhandlers.GiftWithdrawFailedHandler,
	invHandler workerhandlers.InvoicePaymentHandler,
	router *message.Router,
	log *logger.Logger,
) {
	// регистрируем каждый
	router.AddNoPublisherHandler("tg_gift_recv", "telegram.gift.received", sub, tgHandler.Handle)
	router.AddNoPublisherHandler(
		"gift_withdraw_fail",
		"gift.withdraw.failed",
		sub,
		failHandler.Handle,
	)
	router.AddNoPublisherHandler(
		"invoice_paid",
		"invoice.payment.completed",
		sub,
		invHandler.Handle,
	)

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
}

package app

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/ton"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/tonworker"
	"go.uber.org/fx"
)

func NewWorkerTonApp() *fx.App {
	return fx.New(
		moduleCommon,
		service.Module,
		fx.Provide(
			ton.NewTonAPI,
			tonworker.NewProcessor,
		),
		fx.Invoke(func(
			processor *tonworker.Processor,
			lc fx.Lifecycle,
		) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					processor.Start()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return processor.Stop(ctx)
				},
			})
		}),
	)
}

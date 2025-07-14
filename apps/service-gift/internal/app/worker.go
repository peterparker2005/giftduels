package app

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/transport/worker"
	"go.uber.org/fx"
)

func NewWorkerApp() *fx.App {
	return fx.New(
		fx.Options(
			CommonModule,
			worker.Module,
		),
	)
}

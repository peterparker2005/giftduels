package app

import (
	"github.com/peterparker2005/giftduels/apps/service-event/internal/transport/worker"
	"go.uber.org/fx"
)

func NewWorkerApp() *fx.App {
	return fx.New(
		CommonModule,
		worker.Module,
	)
}

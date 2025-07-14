package app

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/eventhandler"
	"go.uber.org/fx"
)

func NewWorkerApp() *fx.App {
	return fx.New(
		fx.Options(
			CommonModule,
			eventhandler.Module,
		),
	)
}

package app

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/eventhandler"
	"go.uber.org/fx"
)

func NewWorkerApp() *fx.App {
	return fx.New(
		moduleCommon,
		service.Module,
		eventhandler.Module,
	)
}

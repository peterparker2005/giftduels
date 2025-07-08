package app

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/transport"
)

func NewGRPCApp() *fx.App {
	return fx.New(
		moduleCommon,
		service.Module,
		transport.Module,
	)
}

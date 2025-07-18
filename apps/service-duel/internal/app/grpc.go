package app

import (
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/asynq"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/transport"
	"go.uber.org/fx"
)

func NewGRPCApp() *fx.App {
	return fx.New(
		CommonModule,
		transport.Module,
		asynq.HandlerModule,
	)
}

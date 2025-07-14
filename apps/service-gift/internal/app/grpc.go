package app

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/transport/grpc"
)

func NewGRPCApp() *fx.App {
	return fx.New(
		CommonModule,
		grpc.Module,
	)
}

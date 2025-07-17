package app

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-event/internal/transport/grpc"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/transport/stream"
)

func NewGRPCApp() *fx.App {
	return fx.New(
		CommonModule,
		grpc.Module,
		stream.Module,
	)
}

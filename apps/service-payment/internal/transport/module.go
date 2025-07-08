package transport

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/transport/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("transport",
	grpc.Module,
)

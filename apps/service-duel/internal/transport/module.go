package transport

import (
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/transport/grpc"
	"go.uber.org/fx"
)

var Module = fx.Options(
	grpc.Module,
)

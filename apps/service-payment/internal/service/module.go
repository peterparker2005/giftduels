package service

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("service",
	fx.Provide(
		payment.NewService,
	),
)

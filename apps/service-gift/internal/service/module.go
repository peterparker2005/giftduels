package service

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/portals"
	giftService "github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	"go.uber.org/fx"
)

// Module предоставляет service зависимости
//
//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("services",
	fx.Provide(
		giftService.New,
		portals.NewPortalsPriceService,
	),
)
